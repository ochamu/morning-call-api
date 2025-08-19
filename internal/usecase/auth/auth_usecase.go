package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/service"
)

// Session はユーザーセッション情報を表す
type Session struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// AuthUseCase は認証に関するユースケースを実装する
type AuthUseCase struct {
	userRepo        repository.UserRepository
	passwordService service.PasswordService
	sessions        map[string]*Session
	sessionMutex    sync.RWMutex
	sessionTimeout  time.Duration
}

// NewAuthUseCase は新しい認証ユースケースを作成する
func NewAuthUseCase(userRepo repository.UserRepository, passwordService service.PasswordService) *AuthUseCase {
	return &AuthUseCase{
		userRepo:        userRepo,
		passwordService: passwordService,
		sessions:        make(map[string]*Session),
		sessionTimeout:  24 * time.Hour, // デフォルトで24時間のセッション有効期限
	}
}

// LoginInput はログイン時の入力データ
type LoginInput struct {
	Username string
	Password string
}

// LoginOutput はログイン時の出力データ
type LoginOutput struct {
	SessionID string
	User      *entity.User
}

// Login はユーザー名とパスワードで認証を行う
func (u *AuthUseCase) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	// 入力値の検証
	if input.Username == "" {
		return nil, fmt.Errorf("ユーザー名は必須です")
	}
	if input.Password == "" {
		return nil, fmt.Errorf("パスワードは必須です")
	}

	// ユーザーを取得
	user, err := u.userRepo.FindByUsername(ctx, input.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("ユーザー名またはパスワードが間違っています")
		}
		return nil, fmt.Errorf("ログイン処理中にエラーが発生しました: %w", err)
	}

	// パスワードを検証
	valid, err := u.passwordService.VerifyPassword(input.Password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("パスワード検証中にエラーが発生しました: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("ユーザー名またはパスワードが間違っています")
	}

	// セッションを作成
	sessionID, err := u.createSession(user.ID)
	if err != nil {
		return nil, fmt.Errorf("セッション作成に失敗しました: %w", err)
	}

	return &LoginOutput{
		SessionID: sessionID,
		User:      user,
	}, nil
}

// Logout はセッションを削除してログアウトする
func (u *AuthUseCase) Logout(_ context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("セッションIDは必須です")
	}

	u.sessionMutex.Lock()
	defer u.sessionMutex.Unlock()

	if _, exists := u.sessions[sessionID]; !exists {
		return fmt.Errorf("無効なセッションです")
	}

	delete(u.sessions, sessionID)
	return nil
}

// GetCurrentUser はセッションIDから現在のユーザー情報を取得する
func (u *AuthUseCase) GetCurrentUser(ctx context.Context, sessionID string) (*entity.User, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("セッションIDは必須です")
	}

	// セッションを取得
	session, err := u.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	// セッションの有効期限を確認
	if time.Now().After(session.ExpiresAt) {
		// 期限切れのセッションを削除
		u.sessionMutex.Lock()
		delete(u.sessions, sessionID)
		u.sessionMutex.Unlock()
		return nil, fmt.Errorf("セッションの有効期限が切れています")
	}

	// ユーザー情報を取得
	user, err := u.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// ユーザーが削除された場合はセッションも削除
			u.sessionMutex.Lock()
			delete(u.sessions, sessionID)
			u.sessionMutex.Unlock()
			return nil, fmt.Errorf("ユーザーが見つかりません")
		}
		return nil, fmt.Errorf("ユーザー情報の取得に失敗しました: %w", err)
	}

	return user, nil
}

// ValidateSession はセッションの有効性を検証する
func (u *AuthUseCase) ValidateSession(_ context.Context, sessionID string) (bool, error) {
	if sessionID == "" {
		return false, nil
	}

	session, err := u.getSession(sessionID)
	if err != nil {
		return false, nil
	}

	// 有効期限のチェック
	if time.Now().After(session.ExpiresAt) {
		// 期限切れのセッションを削除
		u.sessionMutex.Lock()
		delete(u.sessions, sessionID)
		u.sessionMutex.Unlock()
		return false, nil
	}

	return true, nil
}

// createSession は新しいセッションを作成する
func (u *AuthUseCase) createSession(userID string) (string, error) {
	// セッションIDを生成
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("セッションID生成に失敗しました: %w", err)
	}
	sessionID := hex.EncodeToString(b)

	// セッションを保存
	u.sessionMutex.Lock()
	defer u.sessionMutex.Unlock()

	now := time.Now()
	u.sessions[sessionID] = &Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(u.sessionTimeout),
	}

	return sessionID, nil
}

// getSession はセッションを取得する
func (u *AuthUseCase) getSession(sessionID string) (*Session, error) {
	u.sessionMutex.RLock()
	defer u.sessionMutex.RUnlock()

	session, exists := u.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("セッションが見つかりません")
	}

	return session, nil
}

// CleanupExpiredSessions は期限切れのセッションを削除する（定期実行用）
func (u *AuthUseCase) CleanupExpiredSessions() {
	u.sessionMutex.Lock()
	defer u.sessionMutex.Unlock()

	now := time.Now()
	for id, session := range u.sessions {
		if now.After(session.ExpiresAt) {
			delete(u.sessions, id)
		}
	}
}
