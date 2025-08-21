package handler

import (
	"errors"
	"net/http"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/handler/dto/request"
	"github.com/ochamu/morning-call-api/internal/handler/dto/response"
	"github.com/ochamu/morning-call-api/internal/infrastructure/auth"
	authUC "github.com/ochamu/morning-call-api/internal/usecase/auth"
)

// AuthHandler は認証関連のハンドラー
type AuthHandler struct {
	*BaseHandler
	authUseCase    *authUC.AuthUseCase
	sessionManager *auth.SessionManager
}

// NewAuthHandler は新しい認証ハンドラーを作成する
func NewAuthHandler(authUseCase *authUC.AuthUseCase, sessionManager *auth.SessionManager) *AuthHandler {
	return &AuthHandler{
		BaseHandler:    NewBaseHandler(),
		authUseCase:    authUseCase,
		sessionManager: sessionManager,
	}
}

// HandleLogin はログインリクエストを処理する
// POST /api/v1/auth/login
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// POSTメソッドのみ許可
	if r.Method != http.MethodPost {
		h.SendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POSTメソッドのみ許可されています", nil)
		return
	}

	// リクエストボディをパース
	var req request.LoginRequest
	if err := h.ParseJSON(r, &req); err != nil {
		h.SendError(w, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です", nil)
		return
	}

	// バリデーション
	if validationErrs := req.Validate(); len(validationErrs) > 0 {
		var validationErrors []ValidationError
		for field, message := range validationErrs {
			validationErrors = append(validationErrors, ValidationError{
				Field:   field,
				Message: message,
			})
		}
		h.SendValidationError(w, validationErrors)
		return
	}

	// ログイン処理を実行
	loginInput := authUC.LoginInput{
		Username: req.Username,
		Password: req.Password,
	}

	loginOutput, err := h.authUseCase.Login(r.Context(), loginInput)
	if err != nil {
		// 認証エラーの場合
		if errors.Is(err, repository.ErrNotFound) || err.Error() == "ユーザー名またはパスワードが間違っています" {
			h.SendError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "ユーザー名またはパスワードが間違っています", nil)
			return
		}
		// その他のエラー
		h.SendInternalServerError(w, err)
		return
	}

	// セッションを作成（AuthUseCaseが既にセッションを作成しているため、ここでは取得のみ）
	// 将来的にはセッションマネージャーに統一する
	session, err := h.sessionManager.CreateSession(loginOutput.User.ID)
	if err != nil {
		h.SendInternalServerError(w, err)
		return
	}

	// Cookieにセッションを設定
	h.SetCookie(w, "session_id", session.ID, 86400, true, http.SameSiteLaxMode) // 24時間有効

	// レスポンスを返す
	resp := response.LoginResponse{
		SessionID: session.ID,
		User:      h.convertToUserDTO(loginOutput.User),
		ExpiresAt: session.ExpiresAt,
	}

	h.SendJSON(w, http.StatusOK, resp)
}

// HandleLogout はログアウトリクエストを処理する
// POST /api/v1/auth/logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// POSTメソッドのみ許可
	if r.Method != http.MethodPost {
		h.SendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POSTメソッドのみ許可されています", nil)
		return
	}

	// セッションIDを取得
	sessionID, err := h.GetSessionIDFromContext(r.Context())
	if err != nil {
		// セッションがない場合でもログアウトは成功として扱う
		h.DeleteCookie(w, "session_id")
		resp := response.LogoutResponse{
			Success: true,
			Message: "ログアウトしました",
		}
		h.SendJSON(w, http.StatusOK, resp)
		return
	}

	// セッションを削除
	_ = h.sessionManager.DeleteSession(sessionID)
	_ = h.authUseCase.Logout(r.Context(), sessionID) // 既存のAuthUseCaseのセッションも削除

	// Cookieを削除
	h.DeleteCookie(w, "session_id")

	// レスポンスを返す
	resp := response.LogoutResponse{
		Success: true,
		Message: "ログアウトしました",
	}

	h.SendJSON(w, http.StatusOK, resp)
}

// HandleGetCurrentUser は現在のユーザー情報を取得する
// GET /api/v1/auth/me
func (h *AuthHandler) HandleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// GETメソッドのみ許可
	if r.Method != http.MethodGet {
		h.SendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GETメソッドのみ許可されています", nil)
		return
	}

	// 認証が必要（ミドルウェアで処理済み）
	user, ok := h.RequireAuth(w, r)
	if !ok {
		return
	}

	// レスポンスを返す
	resp := response.CurrentUserResponse{
		User: h.convertToUserDTO(user),
	}

	h.SendJSON(w, http.StatusOK, resp)
}

// HandleValidateSession はセッションの有効性を確認する
// GET /api/v1/auth/validate
func (h *AuthHandler) HandleValidateSession(w http.ResponseWriter, r *http.Request) {
	// GETメソッドのみ許可
	if r.Method != http.MethodGet {
		h.SendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GETメソッドのみ許可されています", nil)
		return
	}

	// セッションIDを取得
	sessionID, err := h.GetSessionIDFromContext(r.Context())
	if err != nil {
		h.SendJSON(w, http.StatusOK, map[string]interface{}{
			"valid": false,
		})
		return
	}

	// セッションの有効性を確認
	valid, _ := h.sessionManager.ValidateSession(sessionID)

	// レスポンスを返す
	h.SendJSON(w, http.StatusOK, map[string]interface{}{
		"valid": valid,
	})
}

// HandleRefreshSession はセッションの有効期限を延長する
// POST /api/v1/auth/refresh
func (h *AuthHandler) HandleRefreshSession(w http.ResponseWriter, r *http.Request) {
	// POSTメソッドのみ許可
	if r.Method != http.MethodPost {
		h.SendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POSTメソッドのみ許可されています", nil)
		return
	}

	// セッションIDを取得
	sessionID, err := h.GetSessionIDFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// セッションの有効期限を延長
	if err := h.sessionManager.ExtendSession(sessionID, 86400); err != nil { // 24時間延長
		h.SendAuthenticationError(w)
		return
	}

	// 新しいセッション情報を取得
	session, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// レスポンスを返す
	h.SendJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"expires_at": session.ExpiresAt,
		"message":    "セッションの有効期限を延長しました",
	})
}

// convertToUserDTO はエンティティをDTOに変換する
func (h *AuthHandler) convertToUserDTO(user *entity.User) response.UserDTO {
	return response.UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
