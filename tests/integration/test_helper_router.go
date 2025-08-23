package integration

import (
	"context"
	"net/http"
	"strings"

	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/handler"
	"github.com/ochamu/morning-call-api/internal/infrastructure/auth"
)

// SetupTestRouter はテスト用のルーターをセットアップします
func SetupTestRouter(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	morningCallHandler *handler.MorningCallHandler,
	relationshipHandler *handler.RelationshipHandler,
	sessionManager *auth.SessionManager,
	userRepo repository.UserRepository,
) http.Handler {
	router := http.NewServeMux()

	// 認証ミドルウェアの初期化（簡略版）
	authMiddleware := &testAuthMiddleware{
		sessionManager: sessionManager,
		userRepo:       userRepo,
	}

	// ヘルスチェックエンドポイント
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"morning-call-api","version":"1.0.0","timestamp":1755799593}`))
	})

	// 認証エンドポイント（認証不要）
	router.HandleFunc("/api/v1/auth/login", authHandler.HandleLogin)
	router.HandleFunc("/api/v1/users/register", userHandler.HandleRegister)

	// 認証が必要なエンドポイント
	router.HandleFunc("/api/v1/auth/logout", authMiddleware.Authenticate(authHandler.HandleLogout))
	router.HandleFunc("/api/v1/users/me", authMiddleware.Authenticate(userHandler.HandleGetProfile))
	router.HandleFunc("/api/v1/users/search", authMiddleware.Authenticate(userHandler.HandleSearchUsers))

	// Special morning call endpoints (これらを先に登録)
	router.HandleFunc("/api/v1/morning-calls/sent", authMiddleware.Authenticate(morningCallHandler.HandleListSent))
	router.HandleFunc("/api/v1/morning-calls/received", authMiddleware.Authenticate(morningCallHandler.HandleListReceived))

	// MorningCallエンドポイント
	router.HandleFunc("/api/v1/morning-calls", authMiddleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			morningCallHandler.HandleCreate(w, r)
		case http.MethodGet:
			// クエリパラメータで判定
			if r.URL.Query().Get("type") == "sent" {
				morningCallHandler.HandleListSent(w, r)
			} else if r.URL.Query().Get("type") == "received" {
				morningCallHandler.HandleListReceived(w, r)
			} else {
				http.Error(w, "Query parameter 'type' is required (sent or received)", http.StatusBadRequest)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// MorningCall ID based endpoints - カスタムマッチング
	// http.ServeMuxの制限により、個別のハンドラーを使用
	router.Handle("/api/v1/morning-calls/", authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path
		path := r.URL.Path
		if !strings.HasPrefix(path, "/api/v1/morning-calls/") || len(path) <= len("/api/v1/morning-calls/") {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		
		idPart := path[len("/api/v1/morning-calls/"):]
		
		// Extract morning call ID
		morningCallID := idPart
		if idx := strings.Index(idPart, "/"); idx != -1 {
			morningCallID = idPart[:idx]
		}
		
		// Add morning call ID to context
		ctx := context.WithValue(r.Context(), "morningCallID", morningCallID)
		r = r.WithContext(ctx)
		
		// Check for specific endpoints
		if strings.HasSuffix(idPart, "/confirm") {
			if r.Method != http.MethodPut {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			morningCallHandler.HandleConfirmWake(w, r)
			return
		}
		
		// Regular CRUD operations
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
	})))


	// Relationshipエンドポイント
	router.HandleFunc("/api/v1/relationships/request", authMiddleware.Authenticate(relationshipHandler.HandleSendFriendRequest))
	router.HandleFunc("/api/v1/relationships/friends", authMiddleware.Authenticate(relationshipHandler.HandleListFriends))
	router.HandleFunc("/api/v1/relationships/requests", authMiddleware.Authenticate(relationshipHandler.HandleListFriendRequests))

	// Relationship ID based endpoints
	router.HandleFunc("/api/v1/relationships/", authMiddleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > len("/api/v1/relationships/") {
			idPart := path[len("/api/v1/relationships/"):]
			
			// Check for specific action endpoints
			if strings.HasSuffix(idPart, "/accept") {
				// パスはそのまま維持（ハンドラーが期待する形式で渡す）
				relationshipHandler.HandleAcceptFriendRequest(w, r)
				return
			}
			if strings.HasSuffix(idPart, "/block") {
				// パスはそのまま維持
				relationshipHandler.HandleBlockUser(w, r)
				return
			}
			
			// DELETE endpoint
			if r.Method == http.MethodDelete {
				relationshipHandler.HandleRemoveRelationship(w, r)
				return
			}
			
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// CORSミドルウェアを適用
	return applyCORS(router)
}

// testAuthMiddleware はテスト用の簡易認証ミドルウェア
type testAuthMiddleware struct {
	sessionManager *auth.SessionManager
	userRepo       repository.UserRepository
}

// Authenticate はハンドラーに認証を適用します
func (m *testAuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// セッションIDをCookieから取得
		cookie, err := r.Cookie("session_id")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":{"code":"AUTHENTICATION_ERROR","message":"認証が必要です"}}`))
			return
		}

		// セッションの検証
		session, err := m.sessionManager.GetSession(cookie.Value)
		if err != nil || session == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":{"code":"AUTHENTICATION_ERROR","message":"認証が必要です"}}`))
			return
		}

		// セッションの有効期限チェック
		if session.IsExpired() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":{"code":"AUTHENTICATION_ERROR","message":"認証が必要です"}}`))
			return
		}

		// ユーザー情報を取得
		user, err := m.userRepo.FindByID(r.Context(), session.UserID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":{"code":"AUTHENTICATION_ERROR","message":"認証が必要です"}}`))
			return
		}

		// コンテキストにユーザー情報とセッションIDを設定
		ctx := context.WithValue(r.Context(), handler.UserContextKey, user)
		ctx = context.WithValue(ctx, handler.SessionIDContextKey, cookie.Value)
		next(w, r.WithContext(ctx))
	}
}

// applyCORS はCORSヘッダーを適用します
func applyCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}