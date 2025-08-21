package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
)

// contextKey はコンテキストのキーの型
type contextKey string

const (
	// UserContextKey はコンテキストからユーザー情報を取得するためのキー
	UserContextKey contextKey = "user"
	// SessionIDContextKey はコンテキストからセッションIDを取得するためのキー
	SessionIDContextKey contextKey = "sessionID"
)

// BaseHandler はすべてのハンドラーの基底構造体
type BaseHandler struct {
	// 将来的に共通の依存性を追加する場合はここに定義
}

// ErrorResponse はエラーレスポンスの構造体
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail はエラーの詳細情報
type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details,omitempty"`
}

// ValidationError はバリデーションエラーの詳細
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// SuccessResponse は成功レスポンスの基本構造体
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// NewBaseHandler は新しいBaseHandlerを作成する
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{}
}

// SendJSON はJSONレスポンスを送信する
func (h *BaseHandler) SendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("JSONエンコードエラー: %v", err)
	}
}

// SendSuccess は成功レスポンスを送信する
func (h *BaseHandler) SendSuccess(w http.ResponseWriter, data interface{}, message string) {
	response := SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
	h.SendJSON(w, http.StatusOK, response)
}

// SendError はエラーレスポンスを送信する
func (h *BaseHandler) SendError(w http.ResponseWriter, status int, code string, message string, details []ValidationError) {
	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	h.SendJSON(w, status, response)
}

// SendValidationError はバリデーションエラーレスポンスを送信する
func (h *BaseHandler) SendValidationError(w http.ResponseWriter, errors []ValidationError) {
	h.SendError(w, http.StatusBadRequest, "VALIDATION_ERROR", "入力値が不正です", errors)
}

// SendAuthenticationError は認証エラーレスポンスを送信する
func (h *BaseHandler) SendAuthenticationError(w http.ResponseWriter) {
	h.SendError(w, http.StatusUnauthorized, "AUTHENTICATION_ERROR", "認証が必要です", nil)
}

// SendForbiddenError は権限エラーレスポンスを送信する
func (h *BaseHandler) SendForbiddenError(w http.ResponseWriter) {
	h.SendError(w, http.StatusForbidden, "FORBIDDEN", "この操作を実行する権限がありません", nil)
}

// SendNotFoundError はリソースが見つからないエラーレスポンスを送信する
func (h *BaseHandler) SendNotFoundError(w http.ResponseWriter, resource string) {
	message := fmt.Sprintf("%sが見つかりません", resource)
	h.SendError(w, http.StatusNotFound, "NOT_FOUND", message, nil)
}

// SendInternalServerError は内部サーバーエラーレスポンスを送信する
func (h *BaseHandler) SendInternalServerError(w http.ResponseWriter, err error) {
	log.Printf("内部サーバーエラー: %v", err)
	h.SendError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "サーバーエラーが発生しました", nil)
}

// ParseJSON はリクエストボディからJSONをパースする
func (h *BaseHandler) ParseJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("リクエストボディが空です")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // 未知のフィールドを許可しない

	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("JSONパースエラー: %w", err)
	}

	return nil
}

// GetUserFromContext はコンテキストからユーザー情報を取得する
func (h *BaseHandler) GetUserFromContext(ctx context.Context) (*entity.User, error) {
	user, ok := ctx.Value(UserContextKey).(*entity.User)
	if !ok || user == nil {
		return nil, fmt.Errorf("ユーザー情報がコンテキストに存在しません")
	}
	return user, nil
}

// GetSessionIDFromContext はコンテキストからセッションIDを取得する
func (h *BaseHandler) GetSessionIDFromContext(ctx context.Context) (string, error) {
	sessionID, ok := ctx.Value(SessionIDContextKey).(string)
	if !ok || sessionID == "" {
		return "", fmt.Errorf("セッションIDがコンテキストに存在しません")
	}
	return sessionID, nil
}

// RequireAuth は認証が必要なエンドポイントで使用するヘルパー
// ユーザー情報が取得できない場合は認証エラーを返す
func (h *BaseHandler) RequireAuth(w http.ResponseWriter, r *http.Request) (*entity.User, bool) {
	user, err := h.GetUserFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return nil, false
	}
	return user, true
}

// ValidateRequiredFields は必須フィールドのバリデーションを行う
func (h *BaseHandler) ValidateRequiredFields(fields map[string]string) []ValidationError {
	var errors []ValidationError

	for field, value := range fields {
		if value == "" {
			errors = append(errors, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("%sは必須です", field),
			})
		}
	}

	return errors
}

// GetQueryParam はクエリパラメータを取得する
func (h *BaseHandler) GetQueryParam(r *http.Request, key string, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetPathParam はパスパラメータを取得する（将来的な拡張用）
func (h *BaseHandler) GetPathParam(_ *http.Request, _ string) string {
	// 標準ライブラリでは直接パスパラメータを取得できないため、
	// ルーティング時に設定される値を使用する必要がある
	// 現在は空文字列を返すが、将来的に実装を追加
	return ""
}

// SetCookie はクッキーを設定する
func (h *BaseHandler) SetCookie(w http.ResponseWriter, name, value string, maxAge int, httpOnly bool, sameSite http.SameSite) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: httpOnly,
		SameSite: sameSite,
		Secure:   false, // HTTPSが有効な場合はtrueに設定
	}
	http.SetCookie(w, cookie)
}

// GetCookie はクッキーを取得する
func (h *BaseHandler) GetCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", fmt.Errorf("クッキーが見つかりません: %w", err)
	}
	return cookie.Value, nil
}

// DeleteCookie はクッキーを削除する
func (h *BaseHandler) DeleteCookie(w http.ResponseWriter, name string) {
	h.SetCookie(w, name, "", -1, true, http.SameSiteLaxMode)
}
