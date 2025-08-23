package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ochamu/morning-call-api/internal/config"
	"github.com/ochamu/morning-call-api/internal/handler"
	"github.com/ochamu/morning-call-api/internal/handler/middleware"
)

// HTTPServer はHTTPサーバーの構造体です
type HTTPServer struct {
	server *http.Server
	router *http.ServeMux
	config *config.Config
	deps   *Dependencies // 依存性コンテナ
}

// NewHTTPServer は新しいHTTPサーバーを作成します
func NewHTTPServer(cfg *config.Config, deps *Dependencies) *HTTPServer {
	router := http.NewServeMux()
	
	// ミドルウェアを作成
	authMiddleware := deps.AuthMiddleware
	
	// ヘルスチェック
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})
	
	// API情報
	router.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"Morning Call API","version":"v1"}`))
	})
	
	// 認証エンドポイント
	router.HandleFunc("/api/v1/auth/login", deps.Handlers.Auth.HandleLogin)
	router.HandleFunc("/api/v1/auth/logout", authMiddleware.Authenticate(deps.Handlers.Auth.HandleLogout))
	
	// ユーザーエンドポイント
	router.HandleFunc("/api/v1/users/register", deps.Handlers.User.HandleRegister)
	router.HandleFunc("/api/v1/users/me", authMiddleware.Authenticate(deps.Handlers.User.HandleGetProfile))
	router.HandleFunc("/api/v1/users/search", authMiddleware.Authenticate(deps.Handlers.User.HandleSearchUsers))
	
	// リレーションシップエンドポイント
	router.HandleFunc("/api/v1/relationships/request", authMiddleware.Authenticate(deps.Handlers.Relationship.HandleSendFriendRequest))
	router.HandleFunc("/api/v1/relationships/", authMiddleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		// /api/v1/relationships/{id}/* のパターンを処理
		path := r.URL.Path
		parts := strings.Split(strings.TrimPrefix(path, "/api/v1/relationships/"), "/")
		
		if len(parts) < 2 || parts[0] == "" {
			http.Error(w, "Invalid relationship ID", http.StatusBadRequest)
			return
		}
		
		relationshipID := parts[0]
		action := parts[1]
		
		switch action {
		case "accept":
			if r.Method == http.MethodPut {
				// relationshipIDをコンテキストに設定
				ctx := context.WithValue(r.Context(), "relationshipID", relationshipID)
				deps.Handlers.Relationship.HandleAcceptFriendRequest(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		case "reject":
			if r.Method == http.MethodPut {
				ctx := context.WithValue(r.Context(), "relationshipID", relationshipID)
				deps.Handlers.Relationship.HandleRejectFriendRequest(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		case "block":
			if r.Method == http.MethodPut {
				ctx := context.WithValue(r.Context(), "relationshipID", relationshipID)
				deps.Handlers.Relationship.HandleBlockUser(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		default:
			// DELETE /api/v1/relationships/{id}
			if r.Method == http.MethodDelete && action == "" {
				ctx := context.WithValue(r.Context(), "relationshipID", relationshipID)
				deps.Handlers.Relationship.HandleRemoveRelationship(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Not found", http.StatusNotFound)
			}
		}
	}))
	router.HandleFunc("/api/v1/relationships/friends", authMiddleware.Authenticate(deps.Handlers.Relationship.HandleListFriends))
	router.HandleFunc("/api/v1/relationships/requests", authMiddleware.Authenticate(deps.Handlers.Relationship.HandleListFriendRequests))
	
	// モーニングコールエンドポイント
	router.HandleFunc("/api/v1/morning-calls", authMiddleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			deps.Handlers.MorningCall.HandleCreate(w, r)
		case http.MethodGet:
			// クエリパラメータで判定
			if r.URL.Query().Get("type") == "sent" {
				deps.Handlers.MorningCall.HandleListSent(w, r)
			} else if r.URL.Query().Get("type") == "received" {
				deps.Handlers.MorningCall.HandleListReceived(w, r)
			} else {
				http.Error(w, "Query parameter 'type' is required (sent or received)", http.StatusBadRequest)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	
	router.HandleFunc("/api/v1/morning-calls/sent", authMiddleware.Authenticate(deps.Handlers.MorningCall.HandleListSent))
	router.HandleFunc("/api/v1/morning-calls/received", authMiddleware.Authenticate(deps.Handlers.MorningCall.HandleListReceived))
	
	// パスが/api/v1/morning-calls/で始まる全てのリクエストを処理
	// Go標準のServeMuxは末尾スラッシュがある場合、そのプレフィックスで始まる全パスをマッチする
	router.HandleFunc("/api/v1/morning-calls/", authMiddleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		// /api/v1/morning-calls/{id}/* のパターンを処理
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/morning-calls/")
		
		// 空の場合は別のハンドラーで処理されるべき
		if path == "" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		
		// pathをスラッシュで分割
		parts := strings.Split(path, "/")
		morningCallID := parts[0]
		
		if morningCallID == "" {
			http.Error(w, "Invalid morning call ID", http.StatusBadRequest)
			return
		}
		
		// /api/v1/morning-calls/{id}/confirm
		if len(parts) > 1 && parts[1] == "confirm" {
			if r.Method == http.MethodPut {
				ctx := context.WithValue(r.Context(), "morningCallID", morningCallID)
				deps.Handlers.MorningCall.HandleConfirmWake(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		
		// /api/v1/morning-calls/{id}
		switch r.Method {
		case http.MethodGet:
			ctx := context.WithValue(r.Context(), "morningCallID", morningCallID)
			deps.Handlers.MorningCall.HandleGet(w, r.WithContext(ctx))
		case http.MethodPut:
			ctx := context.WithValue(r.Context(), "morningCallID", morningCallID)
			deps.Handlers.MorningCall.HandleUpdate(w, r.WithContext(ctx))
		case http.MethodDelete:
			ctx := context.WithValue(r.Context(), "morningCallID", morningCallID)
			deps.Handlers.MorningCall.HandleDelete(w, r.WithContext(ctx))
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	
	// HTTPサーバーを作成
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
	
	return &HTTPServer{
		server: server,
		router: router,
		config: cfg,
		deps:   deps,
	}
}

// setupRoutes はルーティングを設定します
func (s *HTTPServer) setupRoutes() {
	// ヘルスチェックエンドポイント
	s.router.HandleFunc("/health", s.handleHealth)

	// APIバージョン情報
	s.router.HandleFunc("/api/v1", s.handleAPIInfo)

	// 依存性を取得
	if s.deps == nil {
		log.Println("警告: 依存性が設定されていません")
		return
	}

	// ハンドラーを取得
	authHandler := s.getAuthHandler()
	userHandler := s.getUserHandler()
	morningCallHandler := s.getMorningCallHandler()
	relationshipHandler := s.getRelationshipHandler()
	authMiddleware := s.getAuthMiddleware()

	if authHandler == nil || userHandler == nil {
		log.Println("警告: 必須ハンドラーが設定されていません")
		return
	}

	// 認証エンドポイント（認証不要）
	s.router.HandleFunc("/api/v1/auth/login", authHandler.HandleLogin)
	s.router.HandleFunc("/api/v1/auth/validate", authHandler.HandleValidateSession)
	s.router.HandleFunc("/api/v1/users/register", userHandler.HandleRegister)

	// 認証が必要なエンドポイント
	if authMiddleware != nil {
		// 認証エンドポイント
		s.router.HandleFunc("/api/v1/auth/logout", authMiddleware.Authenticate(authHandler.HandleLogout))
		s.router.HandleFunc("/api/v1/auth/me", authMiddleware.Authenticate(authHandler.HandleGetCurrentUser))
		s.router.HandleFunc("/api/v1/auth/refresh", authMiddleware.Authenticate(authHandler.HandleRefreshSession))

		// ユーザーエンドポイント
		s.router.HandleFunc("/api/v1/users/profile", authMiddleware.Authenticate(userHandler.HandleGetProfile))
		s.router.HandleFunc("/api/v1/users/search", authMiddleware.Authenticate(userHandler.HandleSearchUsers))
		// ユーザーIDによる取得（パスパラメータ対応）
		s.router.HandleFunc("/api/v1/users/", authMiddleware.Authenticate(userHandler.HandleGetUserByID))
	}

	// Relationshipsエンドポイント
	if relationshipHandler != nil && authMiddleware != nil {
		s.router.HandleFunc("/api/v1/relationships/request", authMiddleware.Authenticate(relationshipHandler.HandleSendFriendRequest))
		s.router.HandleFunc("/api/v1/relationships/friends", authMiddleware.Authenticate(relationshipHandler.HandleListFriends))
		s.router.HandleFunc("/api/v1/relationships/requests", authMiddleware.Authenticate(relationshipHandler.HandleListFriendRequests))
		// IDを含むエンドポイント
		s.router.HandleFunc("/api/v1/relationships/", authMiddleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			// パスからIDを抽出
			prefix := "/api/v1/relationships/"
			if !strings.HasPrefix(path, prefix) {
				http.Error(w, "Invalid path", http.StatusBadRequest)
				return
			}
			
			idPart := strings.TrimPrefix(path, prefix)
			if idPart == "" {
				http.Error(w, "Relationship ID is required", http.StatusBadRequest)
				return
			}
			
			// IDとサブパスを分離
			relationshipID := idPart
			if idx := strings.Index(idPart, "/"); idx != -1 {
				relationshipID = idPart[:idx]
			}
			
			// コンテキストにIDを追加
			ctx := context.WithValue(r.Context(), "relationshipID", relationshipID)
			r = r.WithContext(ctx)
			
			if strings.HasSuffix(path, "/accept") {
				if r.Method != http.MethodPut {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}
				relationshipHandler.HandleAcceptFriendRequest(w, r)
			} else if strings.HasSuffix(path, "/reject") {
				if r.Method != http.MethodPut {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}
				relationshipHandler.HandleRejectFriendRequest(w, r)
			} else if strings.HasSuffix(path, "/block") {
				if r.Method != http.MethodPut {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}
				relationshipHandler.HandleBlockUser(w, r)
			} else if r.Method == http.MethodDelete {
				relationshipHandler.HandleRemoveRelationship(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}))
	}

	// Morning Callsエンドポイント
	if morningCallHandler != nil && authMiddleware != nil {
		// 一覧系
		s.router.HandleFunc("/api/v1/morning-calls/sent", authMiddleware.Authenticate(morningCallHandler.HandleListSent))
		s.router.HandleFunc("/api/v1/morning-calls/received", authMiddleware.Authenticate(morningCallHandler.HandleListReceived))

		// CRUD操作
		s.router.HandleFunc("/api/v1/morning-calls", authMiddleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				morningCallHandler.HandleCreate(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}))

		// IDを含むエンドポイント
		s.router.HandleFunc("/api/v1/morning-calls/", authMiddleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			// パスからIDを抽出
			prefix := "/api/v1/morning-calls/"
			if !strings.HasPrefix(path, prefix) {
				http.Error(w, "Invalid path", http.StatusBadRequest)
				return
			}
			
			idPart := strings.TrimPrefix(path, prefix)
			if idPart == "" {
				http.Error(w, "Morning call ID is required", http.StatusBadRequest)
				return
			}
			
			// IDとサブパスを分離
			morningCallID := idPart
			if idx := strings.Index(idPart, "/"); idx != -1 {
				morningCallID = idPart[:idx]
			}
			
			// コンテキストにIDを追加
			ctx := context.WithValue(r.Context(), "morningCallID", morningCallID)
			r = r.WithContext(ctx)
			
			if strings.HasSuffix(path, "/confirm") {
				if r.Method != http.MethodPut {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}
				morningCallHandler.HandleConfirmWake(w, r)
			} else {
				switch r.Method {
				case http.MethodGet:
					morningCallHandler.HandleGet(w, r)
				case http.MethodPut:
					morningCallHandler.HandleUpdate(w, r)
				case http.MethodDelete:
					morningCallHandler.HandleDelete(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}
		}))
	}
}

// applyMiddleware はミドルウェアを適用します
func (s *HTTPServer) applyMiddleware(handler http.Handler) http.Handler {
	// ミドルウェアチェーンの構築
	handler = s.recoveryMiddleware(handler)
	handler = s.loggingMiddleware(handler)
	handler = s.corsMiddleware(handler)

	return handler
}

// loggingMiddleware はリクエストログを記録するミドルウェアです
func (s *HTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// レスポンスライターのラッパー
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// 次のハンドラーを実行
		next.ServeHTTP(lrw, r)

		// アクセスログ出力
		duration := time.Since(start)
		log.Printf(
			"[%s] %s %s %d %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			lrw.statusCode,
			duration,
		)
	})
}

// recoveryMiddleware はパニックから回復するミドルウェアです
func (s *HTTPServer) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("パニックが発生しました: %v\n%s", err, debug.Stack())

				// エラーレスポンスを返す
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]string{
						"code":    "INTERNAL_ERROR",
						"message": "内部エラーが発生しました",
					},
				}); err != nil {
					log.Printf("パニックリカバリー時のエラーレスポンス送信に失敗しました: %v", err)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// corsMiddleware はCORSヘッダーを設定するミドルウェアです
func (s *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS ヘッダーの設定
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// プリフライトリクエストの処理
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleHealth はヘルスチェックエンドポイントのハンドラーです
func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "morning-call-api",
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(health); err != nil {
		log.Printf("ヘルスチェックレスポンスの送信に失敗しました: %v", err)
	}
}

// handleAPIInfo はAPI情報エンドポイントのハンドラーです
func (s *HTTPServer) handleAPIInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := map[string]interface{}{
		"name":        "Morning Call API",
		"version":     "1.0.0",
		"description": "友達にアラームを設定できるAPIサービス",
		"endpoints": map[string]string{
			"health":        "/health",
			"auth":          "/api/v1/auth",
			"users":         "/api/v1/users",
			"relationships": "/api/v1/relationships",
			"morning_calls": "/api/v1/morning-calls",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(info); err != nil {
		log.Printf("API情報レスポンスの送信に失敗しました: %v", err)
	}
}

// ListenAndServe はHTTPサーバーを起動します
func (s *HTTPServer) ListenAndServe() error {
	return s.server.ListenAndServe()
}

// Shutdown はHTTPサーバーをグレースフルにシャットダウンします
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// loggingResponseWriter はレスポンスのステータスコードを記録するためのラッパーです
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader はステータスコードを記録してから書き込みます
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// ErrorResponse はエラーレスポンスの構造体です
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail はエラーの詳細情報です
type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details,omitempty"`
}

// ValidationError はバリデーションエラーの詳細です
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// TODO: 以下のヘルパー関数は将来のハンドラー実装で使用予定
// sendJSONResponse はJSONレスポンスを送信するヘルパー関数です
// func sendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(status)
// 	if err := json.NewEncoder(w).Encode(data); err != nil {
// 		log.Printf("JSONレスポンスの送信に失敗しました: %v", err)
// 	}
// }

// sendErrorResponse はエラーレスポンスを送信するヘルパー関数です
// func sendErrorResponse(w http.ResponseWriter, status int, code, message string, details []ValidationError) {
// 	response := ErrorResponse{
// 		Error: ErrorDetail{
// 			Code:    code,
// 			Message: message,
// 			Details: details,
// 		},
// 	}
// 	sendJSONResponse(w, status, response)
// }

// getAuthHandler は依存性からAuthハンドラーを取得する
func (s *HTTPServer) getAuthHandler() *handler.AuthHandler {
	if s.deps == nil || s.deps.Handlers.Auth == nil {
		return nil
	}
	return s.deps.Handlers.Auth
}

// getUserHandler は依存性からUserハンドラーを取得する
func (s *HTTPServer) getUserHandler() *handler.UserHandler {
	if s.deps == nil || s.deps.Handlers.User == nil {
		return nil
	}
	return s.deps.Handlers.User
}

// getAuthMiddleware は依存性から認証ミドルウェアを取得する
func (s *HTTPServer) getAuthMiddleware() *middleware.AuthMiddleware {
	if s.deps == nil || s.deps.AuthMiddleware == nil {
		return nil
	}
	return s.deps.AuthMiddleware
}

// getMorningCallHandler は依存性からMorningCallハンドラーを取得する
func (s *HTTPServer) getMorningCallHandler() *handler.MorningCallHandler {
	if s.deps == nil || s.deps.Handlers.MorningCall == nil {
		return nil
	}
	return s.deps.Handlers.MorningCall
}

// getRelationshipHandler は依存性からRelationshipハンドラーを取得する
func (s *HTTPServer) getRelationshipHandler() *handler.RelationshipHandler {
	if s.deps == nil || s.deps.Handlers.Relationship == nil {
		return nil
	}
	return s.deps.Handlers.Relationship
}
