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
)

// HTTPServer はHTTPサーバーの構造体です
type HTTPServer struct {
	server *http.Server
	router *http.ServeMux
	config *config.Config
	deps   interface{} // 依存性コンテナ（main.goのDependencies）
}

// NewHTTPServer は新しいHTTPサーバーを作成します
func NewHTTPServer(cfg *config.Config, deps interface{}) *HTTPServer {
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

	// TODO: 各リソースのハンドラーを追加
	// Authentication
	// s.router.HandleFunc("/api/v1/auth/login", s.handleLogin)
	// s.router.HandleFunc("/api/v1/auth/logout", s.handleLogout)

	// Users
	// s.router.HandleFunc("/api/v1/users/register", s.handleUserRegister)
	// s.router.HandleFunc("/api/v1/users/me", s.handleGetCurrentUser)

	// Relationships
	// s.router.HandleFunc("/api/v1/relationships/request", s.handleSendFriendRequest)
	// s.router.HandleFunc("/api/v1/relationships/friends", s.handleListFriends)

	// Morning Calls
	// s.router.HandleFunc("/api/v1/morning-calls", s.handleMorningCalls)
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
