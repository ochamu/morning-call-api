package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/handler"
	"github.com/ochamu/morning-call-api/internal/infrastructure/auth"
)

// AuthMiddleware は認証ミドルウェア
type AuthMiddleware struct {
	sessionManager *auth.SessionManager
	userRepo       repository.UserRepository
	baseHandler    *handler.BaseHandler
}

// NewAuthMiddleware は新しい認証ミドルウェアを作成する
func NewAuthMiddleware(sessionManager *auth.SessionManager, userRepo repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		sessionManager: sessionManager,
		userRepo:       userRepo,
		baseHandler:    handler.NewBaseHandler(),
	}
}

// Authenticate は認証が必要なエンドポイントに適用するミドルウェア
func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// セッションIDを取得（Cookieまたはヘッダーから）
		sessionID := m.getSessionID(r)
		if sessionID == "" {
			m.baseHandler.SendAuthenticationError(w)
			return
		}

		// セッションの検証
		valid, err := m.sessionManager.ValidateSession(sessionID)
		if err != nil || !valid {
			m.baseHandler.SendAuthenticationError(w)
			return
		}

		// セッションからユーザーIDを取得
		userID, err := m.sessionManager.GetUserIDFromSession(sessionID)
		if err != nil {
			m.baseHandler.SendAuthenticationError(w)
			return
		}

		// ユーザー情報を取得
		user, err := m.userRepo.FindByID(r.Context(), userID)
		if err != nil {
			m.baseHandler.SendAuthenticationError(w)
			return
		}

		// コンテキストにユーザー情報とセッションIDを設定
		ctx := context.WithValue(r.Context(), handler.UserContextKey, user)
		ctx = context.WithValue(ctx, handler.SessionIDContextKey, sessionID)

		// 次のハンドラーを実行
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// OptionalAuth は認証が任意のエンドポイントに適用するミドルウェア
// 認証情報があればコンテキストに設定し、なければそのまま処理を続行する
func (m *AuthMiddleware) OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// セッションIDを取得（Cookieまたはヘッダーから）
		sessionID := m.getSessionID(r)
		if sessionID != "" {
			// セッションの検証
			valid, err := m.sessionManager.ValidateSession(sessionID)
			if err == nil && valid {
				// セッションからユーザーIDを取得
				userID, err := m.sessionManager.GetUserIDFromSession(sessionID)
				if err == nil {
					// ユーザー情報を取得
					user, err := m.userRepo.FindByID(r.Context(), userID)
					if err == nil {
						// コンテキストにユーザー情報とセッションIDを設定
						ctx := context.WithValue(r.Context(), handler.UserContextKey, user)
						ctx = context.WithValue(ctx, handler.SessionIDContextKey, sessionID)
						r = r.WithContext(ctx)
					}
				}
			}
		}

		// 次のハンドラーを実行
		next.ServeHTTP(w, r)
	}
}

// RequireAdmin は管理者権限が必要なエンドポイントに適用するミドルウェア
func (m *AuthMiddleware) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return m.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		// ユーザー情報を取得
		user, err := m.baseHandler.GetUserFromContext(r.Context())
		if err != nil {
			m.baseHandler.SendAuthenticationError(w)
			return
		}

		// 管理者権限のチェック（現在は実装されていないため、将来的に追加）
		// if !user.IsAdmin {
		//     m.baseHandler.SendForbiddenError(w)
		//     return
		// }
		_ = user // 未使用変数の警告を回避

		// 次のハンドラーを実行
		next.ServeHTTP(w, r)
	})
}

// getSessionID はリクエストからセッションIDを取得する
func (m *AuthMiddleware) getSessionID(r *http.Request) string {
	// 1. Authorizationヘッダーから取得を試みる
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// "Bearer <session_id>" 形式を想定
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
		// "Session <session_id>" 形式も許可
		if len(parts) == 2 && strings.ToLower(parts[0]) == "session" {
			return parts[1]
		}
	}

	// 2. X-Session-IDヘッダーから取得を試みる
	sessionHeader := r.Header.Get("X-Session-ID")
	if sessionHeader != "" {
		return sessionHeader
	}

	// 3. Cookieから取得を試みる
	cookie, err := r.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

// ExtendSession はセッションの有効期限を延長する
func (m *AuthMiddleware) ExtendSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// セッションIDを取得
		sessionID, _ := m.baseHandler.GetSessionIDFromContext(r.Context())
		if sessionID != "" {
			// セッションの有効期限を延長（エラーは無視）
			_ = m.sessionManager.ExtendSession(sessionID, 24*60*60) // 24時間延長
		}

		// 次のハンドラーを実行
		next.ServeHTTP(w, r)
	}
}
