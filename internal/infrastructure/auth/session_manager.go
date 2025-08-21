package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Session はユーザーセッション情報を表す
type Session struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
	Data      map[string]interface{} // 追加のセッションデータ
}

// SessionManager はセッション管理を行う
type SessionManager struct {
	sessions       map[string]*Session
	mutex          sync.RWMutex
	defaultTimeout time.Duration
	// クリーンアップ用のチャネル
	cleanupTicker *time.Ticker
	stopCleanup   chan bool
}

// NewSessionManager は新しいセッションマネージャーを作成する
func NewSessionManager(timeout time.Duration) *SessionManager {
	sm := &SessionManager{
		sessions:       make(map[string]*Session),
		defaultTimeout: timeout,
		stopCleanup:    make(chan bool),
	}

	// 期限切れセッションの自動クリーンアップを開始
	sm.startCleanupRoutine()

	return sm
}

// CreateSession は新しいセッションを作成する
func (sm *SessionManager) CreateSession(userID string) (*Session, error) {
	if userID == "" {
		return nil, fmt.Errorf("ユーザーIDは必須です")
	}

	// セッションIDを生成
	sessionID, err := sm.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("セッションID生成に失敗しました: %w", err)
	}

	now := time.Now()
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(sm.defaultTimeout),
		Data:      make(map[string]interface{}),
	}

	// セッションを保存
	sm.mutex.Lock()
	sm.sessions[sessionID] = session
	sm.mutex.Unlock()

	return session, nil
}

// GetSession はセッションを取得する
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("セッションIDは必須です")
	}

	sm.mutex.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("セッションが見つかりません")
	}

	// 有効期限の確認
	if time.Now().After(session.ExpiresAt) {
		// 期限切れセッションを削除
		sm.DeleteSession(sessionID)
		return nil, fmt.Errorf("セッションの有効期限が切れています")
	}

	return session, nil
}

// UpdateSession はセッションを更新する
func (sm *SessionManager) UpdateSession(sessionID string, data map[string]interface{}) error {
	if sessionID == "" {
		return fmt.Errorf("セッションIDは必須です")
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("セッションが見つかりません")
	}

	// データを更新
	for key, value := range data {
		session.Data[key] = value
	}

	return nil
}

// ExtendSession はセッションの有効期限を延長する
func (sm *SessionManager) ExtendSession(sessionID string, duration time.Duration) error {
	if sessionID == "" {
		return fmt.Errorf("セッションIDは必須です")
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("セッションが見つかりません")
	}

	// 有効期限を延長
	session.ExpiresAt = time.Now().Add(duration)

	return nil
}

// DeleteSession はセッションを削除する
func (sm *SessionManager) DeleteSession(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("セッションIDは必須です")
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.sessions[sessionID]; !exists {
		return fmt.Errorf("セッションが見つかりません")
	}

	delete(sm.sessions, sessionID)
	return nil
}

// ValidateSession はセッションの有効性を検証する
func (sm *SessionManager) ValidateSession(sessionID string) (bool, error) {
	if sessionID == "" {
		return false, nil
	}

	session, err := sm.GetSession(sessionID)
	if err != nil {
		return false, nil
	}

	// 有効期限のチェック
	if time.Now().After(session.ExpiresAt) {
		sm.DeleteSession(sessionID)
		return false, nil
	}

	return true, nil
}

// GetUserIDFromSession はセッションからユーザーIDを取得する
func (sm *SessionManager) GetUserIDFromSession(sessionID string) (string, error) {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return "", err
	}

	return session.UserID, nil
}

// CleanupExpiredSessions は期限切れのセッションを削除する
func (sm *SessionManager) CleanupExpiredSessions() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	for id, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			delete(sm.sessions, id)
		}
	}
}

// GetActiveSessionCount はアクティブなセッション数を取得する
func (sm *SessionManager) GetActiveSessionCount() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	count := 0
	now := time.Now()
	for _, session := range sm.sessions {
		if now.Before(session.ExpiresAt) {
			count++
		}
	}

	return count
}

// GetSessionsByUserID は特定のユーザーのすべてのセッションを取得する
func (sm *SessionManager) GetSessionsByUserID(userID string) []*Session {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var sessions []*Session
	now := time.Now()
	for _, session := range sm.sessions {
		if session.UserID == userID && now.Before(session.ExpiresAt) {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// InvalidateUserSessions は特定のユーザーのすべてのセッションを無効化する
func (sm *SessionManager) InvalidateUserSessions(userID string) error {
	if userID == "" {
		return fmt.Errorf("ユーザーIDは必須です")
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for id, session := range sm.sessions {
		if session.UserID == userID {
			delete(sm.sessions, id)
		}
	}

	return nil
}

// Stop はセッションマネージャーを停止する（クリーンアップルーチンを停止）
func (sm *SessionManager) Stop() {
	if sm.cleanupTicker != nil {
		sm.cleanupTicker.Stop()
		sm.stopCleanup <- true
	}
}

// generateSessionID はセキュアなセッションIDを生成する
func (sm *SessionManager) generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("ランダムバイト生成に失敗しました: %w", err)
	}

	return hex.EncodeToString(b), nil
}

// startCleanupRoutine は定期的に期限切れセッションをクリーンアップするルーチンを開始する
func (sm *SessionManager) startCleanupRoutine() {
	// 5分ごとにクリーンアップを実行
	sm.cleanupTicker = time.NewTicker(5 * time.Minute)

	go func() {
		for {
			select {
			case <-sm.cleanupTicker.C:
				sm.CleanupExpiredSessions()
			case <-sm.stopCleanup:
				return
			}
		}
	}()
}
