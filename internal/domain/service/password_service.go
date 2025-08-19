package service

// PasswordService はパスワードのハッシュ化と検証を行うサービスのインターフェース
type PasswordService interface {
	// HashPassword はパスワードをハッシュ化する
	HashPassword(password string) (string, error)

	// VerifyPassword はパスワードとハッシュ値を比較検証する
	VerifyPassword(password, passwordHash string) (bool, error)
}
