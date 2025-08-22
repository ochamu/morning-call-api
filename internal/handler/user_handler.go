package handler

import (
	"errors"
	"net/http"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/handler/dto/request"
	"github.com/ochamu/morning-call-api/internal/handler/dto/response"
	"github.com/ochamu/morning-call-api/internal/infrastructure/auth"
	"github.com/ochamu/morning-call-api/internal/usecase/user"
)

// UserHandler はユーザー関連のハンドラー
type UserHandler struct {
	*BaseHandler
	userUseCase    *user.UserUseCase
	sessionManager *auth.SessionManager
}

// NewUserHandler は新しいユーザーハンドラーを作成する
func NewUserHandler(userUseCase *user.UserUseCase, sessionManager *auth.SessionManager) *UserHandler {
	return &UserHandler{
		BaseHandler:    NewBaseHandler(),
		userUseCase:    userUseCase,
		sessionManager: sessionManager,
	}
}

// HandleRegister はユーザー登録リクエストを処理する
// POST /api/v1/users/register
func (h *UserHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	// POSTメソッドのみ許可
	if r.Method != http.MethodPost {
		h.SendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POSTメソッドのみ許可されています", nil)
		return
	}

	// リクエストボディをパース
	var req request.RegisterRequest
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

	// ユーザー登録処理を実行
	registerInput := user.RegisterInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	registerOutput, err := h.userUseCase.Register(r.Context(), registerInput)
	if err != nil {
		// ユーザー名またはメールアドレスが既に存在する場合
		if errors.Is(err, repository.ErrAlreadyExists) {
			h.SendError(w, http.StatusConflict, "ALREADY_EXISTS", "ユーザー名またはメールアドレスが既に使用されています", nil)
			return
		}
		// バリデーションエラーの場合
		if err.Error() == "ユーザー名は必須です" || err.Error() == "メールアドレスは必須です" || err.Error() == "パスワードは必須です" {
			h.SendError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		// その他のエラー
		h.SendInternalServerError(w, err)
		return
	}

	// 自動ログイン（オプション）
	// セッションを作成
	session, err := h.sessionManager.CreateSession(registerOutput.User.ID)
	if err != nil {
		// セッション作成に失敗しても登録は成功として扱う
		resp := response.RegisterResponse{
			Success: true,
			User:    h.convertToUserDTO(registerOutput.User),
			Message: "ユーザー登録が完了しました。ログインしてください。",
		}
		h.SendJSON(w, http.StatusCreated, resp)
		return
	}

	// Cookieにセッションを設定
	h.SetCookie(w, "session_id", session.ID, 86400, true, http.SameSiteLaxMode) // 24時間有効

	// レスポンスを返す
	resp := response.RegisterResponse{
		Success: true,
		User:    h.convertToUserDTO(registerOutput.User),
		Message: "ユーザー登録が完了しました",
	}

	h.SendJSON(w, http.StatusCreated, resp)
}

// HandleGetProfile はユーザープロフィールを取得する
// GET /api/v1/users/profile
func (h *UserHandler) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	// GETメソッドのみ許可
	if r.Method != http.MethodGet {
		h.SendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GETメソッドのみ許可されています", nil)
		return
	}

	// 認証が必要
	currentUser, ok := h.RequireAuth(w, r)
	if !ok {
		return
	}

	// レスポンスを返す
	h.SendJSON(w, http.StatusOK, map[string]interface{}{
		"user": h.convertToUserDTO(currentUser),
	})
}

// HandleSearchUsers はユーザーを検索する
// GET /api/v1/users/search?query=xxx
func (h *UserHandler) HandleSearchUsers(w http.ResponseWriter, r *http.Request) {
	// GETメソッドのみ許可
	if r.Method != http.MethodGet {
		h.SendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GETメソッドのみ許可されています", nil)
		return
	}

	// 認証が必要
	currentUser, ok := h.RequireAuth(w, r)
	if !ok {
		return
	}

	// クエリパラメータを取得
	query := h.GetQueryParam(r, "query", "")
	if query == "" {
		h.SendError(w, http.StatusBadRequest, "INVALID_REQUEST", "検索クエリが指定されていません", nil)
		return
	}

	// ユーザー検索を実行
	searchOutput, err := h.userUseCase.SearchUsers(r.Context(), user.SearchUsersInput{
		Query:     query,
		ExcludeID: currentUser.ID, // 自分自身を除外
		Limit:     100,
	})
	if err != nil {
		h.SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "ユーザー検索に失敗しました", nil)
		return
	}

	// DTOに変換
	var users []response.UserDTO
	for _, u := range searchOutput.Users {
		users = append(users, h.convertToUserDTO(u))
	}

	// レスポンスを返す
	h.SendJSON(w, http.StatusOK, map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

// HandleGetUserByID は指定したIDのユーザー情報を取得する
// GET /api/v1/users/{id}
func (h *UserHandler) HandleGetUserByID(w http.ResponseWriter, r *http.Request) {
	// GETメソッドのみ許可
	if r.Method != http.MethodGet {
		h.SendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GETメソッドのみ許可されています", nil)
		return
	}

	// 認証が必要
	_, ok := h.RequireAuth(w, r)
	if !ok {
		return
	}

	// パスからユーザーIDを取得
	// 標準ライブラリでは直接パスパラメータを取得できないため、
	// URLパスから手動で抽出する必要がある
	// 例: /api/v1/users/123 から "123" を取得
	path := r.URL.Path
	prefix := "/api/v1/users/"
	if len(path) <= len(prefix) {
		h.SendError(w, http.StatusBadRequest, "INVALID_REQUEST", "ユーザーIDが指定されていません", nil)
		return
	}

	userID := path[len(prefix):]
	if userID == "" {
		h.SendError(w, http.StatusBadRequest, "INVALID_REQUEST", "ユーザーIDが指定されていません", nil)
		return
	}

	// ユーザー情報を取得
	// UserUseCaseのGetByIDメソッドを使用
	foundUser, err := h.userUseCase.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.SendNotFoundError(w, "ユーザー")
			return
		}
		h.SendInternalServerError(w, err)
		return
	}

	// レスポンスを返す
	h.SendJSON(w, http.StatusOK, h.convertToUserDTO(foundUser))
}

// convertToUserDTO はエンティティをDTOに変換する
func (h *UserHandler) convertToUserDTO(u *entity.User) response.UserDTO {
	return response.UserDTO{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
