package response

import "time"

// LoginResponse はログインレスポンスのDTO
type LoginResponse struct {
	SessionID string    `json:"session_id"`
	User      UserDTO   `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
}

// LogoutResponse はログアウトレスポンスのDTO
type LogoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RegisterResponse はユーザー登録レスポンスのDTO
type RegisterResponse struct {
	Success bool    `json:"success"`
	User    UserDTO `json:"user"`
	Message string  `json:"message"`
}

// UserDTO はユーザー情報のDTO
type UserDTO struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SessionInfo はセッション情報のDTO
type SessionInfo struct {
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CurrentUserResponse は現在のユーザー情報レスポンスのDTO
type CurrentUserResponse struct {
	User UserDTO `json:"user"`
}
