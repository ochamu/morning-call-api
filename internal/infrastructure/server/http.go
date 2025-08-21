package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
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

	srv := &HTTPServer{
		router: router,
		config: cfg,
		deps:   deps,
	}

	// ルートの設定
	srv.setupRoutes()

	// HTTPサーバーの設定
	srv.server = &http.Server{
		Addr:           fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:        srv.applyMiddleware(router),
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	return srv
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
	authMiddleware := s.getAuthMiddleware()

	if authHandler == nil || userHandler == nil {
		log.Println("警告: ハンドラーが設定されていません")
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

	// TODO: 他のリソースのハンドラーを追加
	// Relationships
	// s.router.HandleFunc("/api/v1/relationships/request", deps.AuthMiddleware.Authenticate(relationshipHandler.HandleSendFriendRequest))
	// s.router.HandleFunc("/api/v1/relationships/friends", deps.AuthMiddleware.Authenticate(relationshipHandler.HandleListFriends))

	// Morning Calls
	// s.router.HandleFunc("/api/v1/morning-calls", deps.AuthMiddleware.Authenticate(morningCallHandler.HandleMorningCalls))
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
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]string{
						"code":    "INTERNAL_ERROR",
						"message": "内部エラーが発生しました",
					},
				})
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
