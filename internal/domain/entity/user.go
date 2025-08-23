package entity

import (
	"regexp"
	"strings"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// User はシステムのユーザーを表すエンティティ
type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string // ハッシュ化されたパスワード
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// emailRegex はメールアドレスの簡易的な検証用正規表現
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// NewUser は新しいユーザーエンティティを作成する
func NewUser(id, username, email, passwordHash string) (*User, valueobject.NGReason) {
	user := &User{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 検証
	if reason := user.Validate(); reason.IsNG() {
		return nil, reason
	}

	return user, valueobject.OK()
}

// Validate はユーザーエンティティの妥当性を検証する
func (u *User) Validate() valueobject.NGReason {
	// ID検証
	if u.ID == "" {
		return valueobject.NG("ユーザーIDは必須です")
	}

	// ユーザー名検証
	if reason := u.ValidateUsername(); reason.IsNG() {
		return reason
	}

	// メールアドレス検証
	if reason := u.ValidateEmail(); reason.IsNG() {
		return reason
	}

	return valueobject.OK()
}

// ValidateUsername はユーザー名の妥当性を検証する
func (u *User) ValidateUsername() valueobject.NGReason {
	if u.Username == "" {
		return valueobject.NG("ユーザー名は必須です")
	}

	if len(u.Username) < 3 {
		return valueobject.NG("ユーザー名は3文字以上である必要があります")
	}

	if len(u.Username) > 30 {
		return valueobject.NG("ユーザー名は30文字以内である必要があります")
	}

	// ユーザー名に使用可能な文字のチェック（英数字、アンダースコア、ハイフン）
	for _, r := range u.Username {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return valueobject.NG("ユーザー名には英数字、アンダースコア、ハイフンのみ使用できます")
		}
	}

	return valueobject.OK()
}

// ValidateEmail はメールアドレスの妥当性を検証する
func (u *User) ValidateEmail() valueobject.NGReason {
	if u.Email == "" {
		return valueobject.NG("メールアドレスは必須です")
	}

	// 小文字に正規化
	u.Email = strings.ToLower(u.Email)

	if !emailRegex.MatchString(u.Email) {
		return valueobject.NG("メールアドレスの形式が正しくありません")
	}

	if len(u.Email) > 255 {
		return valueobject.NG("メールアドレスは255文字以内である必要があります")
	}

	return valueobject.OK()
}

// ValidatePassword はパスワードの妥当性を検証する（平文パスワード用）
func ValidatePassword(password string) valueobject.NGReason {
	if password == "" {
		return valueobject.NG("パスワードは必須です")
	}

	if len(password) < 8 {
		return valueobject.NG("パスワードは8文字以上である必要があります")
	}

	// bcryptの制限（72バイト）を考慮
	if len(password) > 72 {
		return valueobject.NG("パスワードは72文字以内である必要があります")
	}

	// パスワード強度の基本的なチェック
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, r := range password {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?", r):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return valueobject.NG("パスワードは大文字、小文字、数字、特殊文字をそれぞれ1文字以上含む必要があります")
	}

	return valueobject.OK()
}

// CanSendMorningCallTo は指定したユーザーにモーニングコールを送信可能か検証する
// 友達関係の確認は別レイヤーで行うため、ここでは自己送信のチェックのみ
func (u *User) CanSendMorningCallTo(receiverID string) valueobject.NGReason {
	if u.ID == receiverID {
		return valueobject.NG("自分自身にモーニングコールを設定することはできません")
	}
	return valueobject.OK()
}

// UpdateUsername はユーザー名を更新する
func (u *User) UpdateUsername(newUsername string) valueobject.NGReason {
	oldUsername := u.Username
	u.Username = newUsername

	if reason := u.ValidateUsername(); reason.IsNG() {
		u.Username = oldUsername // ロールバック
		return reason
	}

	u.UpdatedAt = time.Now()
	return valueobject.OK()
}

// UpdateEmail はメールアドレスを更新する
func (u *User) UpdateEmail(newEmail string) valueobject.NGReason {
	oldEmail := u.Email
	u.Email = newEmail

	if reason := u.ValidateEmail(); reason.IsNG() {
		u.Email = oldEmail // ロールバック
		return reason
	}

	u.UpdatedAt = time.Now()
	return valueobject.OK()
}

// Equals は他のユーザーと同一かを判定する
func (u *User) Equals(other *User) bool {
	if other == nil {
		return false
	}
	return u.ID == other.ID
}
