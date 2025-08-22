package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/ochamu/morning-call-api/internal/domain/service"
)

// インターフェースの実装を保証
var _ service.PasswordService = (*PasswordService)(nil)

// PasswordService はパスワードのハッシュ化と検証を行うサービス
type PasswordService struct {
	// bcryptのコスト係数（10-12が推奨）
	cost int
}

// NewPasswordService は新しいPasswordServiceを作成する
func NewPasswordService() *PasswordService {
	return &PasswordService{
		cost: 10, // bcryptの推奨コスト係数
	}
}

// HashPassword はパスワードをハッシュ化する
func (s *PasswordService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password is required")
	}

	// bcryptは72バイトまでしか処理できないため、長いパスワードは事前にSHA256でハッシュ化
	passwordBytes := []byte(password)
	if len(passwordBytes) > 72 {
		// SHA256でプリハッシュ
		hasher := sha256.New()
		hasher.Write(passwordBytes)
		hashBytes := hasher.Sum(nil)
		// hex文字列に変換（64文字 = 32バイト * 2）
		passwordBytes = []byte(hex.EncodeToString(hashBytes))
	}

	// bcryptでハッシュ化
	hash, err := bcrypt.GenerateFromPassword(passwordBytes, s.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword はパスワードとハッシュを検証する
func (s *PasswordService) VerifyPassword(password, passwordHash string) (bool, error) {
	if password == "" || passwordHash == "" {
		return false, fmt.Errorf("password and passwordHash are required")
	}

	// bcryptは72バイトまでしか処理できないため、長いパスワードは事前にSHA256でハッシュ化
	passwordBytes := []byte(password)
	if len(passwordBytes) > 72 {
		// SHA256でプリハッシュ
		hasher := sha256.New()
		hasher.Write(passwordBytes)
		hashBytes := hasher.Sum(nil)
		// hex文字列に変換
		passwordBytes = []byte(hex.EncodeToString(hashBytes))
	}

	// bcryptで検証
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), passwordBytes)
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, fmt.Errorf("failed to verify password: %w", err)
	}

	return true, nil
}