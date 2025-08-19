package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/ochamu/morning-call-api/internal/domain/service"
)

// インターフェースの実装を保証
var _ service.PasswordService = (*PasswordService)(nil)

// PasswordService はパスワードのハッシュ化と検証を行うサービス
type PasswordService struct {
	// salt はパスワードハッシュ化に使用するソルト
	// 本番環境では環境変数から取得するべきだが、現時点では固定値を使用
	salt string
}

// NewPasswordService は新しいPasswordServiceを作成する
func NewPasswordService() *PasswordService {
	return &PasswordService{
		// TODO: 本番環境では環境変数から取得する
		salt: "morning-call-api-salt-2024",
	}
}

// HashPassword はパスワードをハッシュ化する
func (s *PasswordService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password is required")
	}

	// SHA256でハッシュ化（salt付き）
	hasher := sha256.New()
	hasher.Write([]byte(password + s.salt))
	hashBytes := hasher.Sum(nil)

	// hex文字列に変換
	hashString := hex.EncodeToString(hashBytes)

	return hashString, nil
}

// VerifyPassword はパスワードとハッシュを検証する
func (s *PasswordService) VerifyPassword(password, passwordHash string) (bool, error) {
	if password == "" || passwordHash == "" {
		return false, fmt.Errorf("password and passwordHash are required")
	}

	// 入力されたパスワードをハッシュ化
	hashedInput, err := s.HashPassword(password)
	if err != nil {
		return false, fmt.Errorf("failed to hash password: %w", err)
	}

	// ハッシュ値を比較
	return hashedInput == passwordHash, nil
}
