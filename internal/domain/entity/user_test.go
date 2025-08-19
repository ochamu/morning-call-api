package entity

import (
	"strings"
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		username     string
		email        string
		passwordHash string
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "正常なユーザー作成",
			id:           "user-001",
			username:     "testuser",
			email:        "test@example.com",
			passwordHash: "hashedpassword",
			expectError:  false,
		},
		{
			name:         "IDが空",
			id:           "",
			username:     "testuser",
			email:        "test@example.com",
			passwordHash: "hashedpassword",
			expectError:  true,
			errorMsg:     "ユーザーIDは必須です",
		},
		{
			name:         "ユーザー名が空",
			id:           "user-001",
			username:     "",
			email:        "test@example.com",
			passwordHash: "hashedpassword",
			expectError:  true,
			errorMsg:     "ユーザー名は必須です",
		},
		{
			name:         "メールアドレスが空",
			id:           "user-001",
			username:     "testuser",
			email:        "",
			passwordHash: "hashedpassword",
			expectError:  true,
			errorMsg:     "メールアドレスは必須です",
		},
		{
			name:         "ユーザー名が短すぎる",
			id:           "user-001",
			username:     "ab",
			email:        "test@example.com",
			passwordHash: "hashedpassword",
			expectError:  true,
			errorMsg:     "ユーザー名は3文字以上である必要があります",
		},
		{
			name:         "ユーザー名が長すぎる",
			id:           "user-001",
			username:     strings.Repeat("a", 31),
			email:        "test@example.com",
			passwordHash: "hashedpassword",
			expectError:  true,
			errorMsg:     "ユーザー名は30文字以内である必要があります",
		},
		{
			name:         "ユーザー名に不正な文字",
			id:           "user-001",
			username:     "test@user",
			email:        "test@example.com",
			passwordHash: "hashedpassword",
			expectError:  true,
			errorMsg:     "ユーザー名には英数字、アンダースコア、ハイフンのみ使用できます",
		},
		{
			name:         "メールアドレスの形式が不正",
			id:           "user-001",
			username:     "testuser",
			email:        "invalid-email",
			passwordHash: "hashedpassword",
			expectError:  true,
			errorMsg:     "メールアドレスの形式が正しくありません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, reason := NewUser(tt.id, tt.username, tt.email, tt.passwordHash)

			if tt.expectError {
				if reason.IsOK() {
					t.Errorf("エラーが期待されたが、成功した")
				}
				if reason.Error() != tt.errorMsg {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMsg, reason.Error())
				}
				if user != nil {
					t.Errorf("エラー時にはnilが期待されたが、ユーザーが返された")
				}
			} else {
				if reason.IsNG() {
					t.Errorf("成功が期待されたが、エラーが発生: %s", reason.Error())
				}
				if user == nil {
					t.Errorf("成功時にはユーザーが期待されたが、nilが返された")
				} else {
					if user.ID != tt.id {
						t.Errorf("ID: expected %s, got %s", tt.id, user.ID)
					}
					if user.Username != tt.username {
						t.Errorf("Username: expected %s, got %s", tt.username, user.Username)
					}
					// メールアドレスは小文字に正規化される
					expectedEmail := strings.ToLower(tt.email)
					if user.Email != expectedEmail {
						t.Errorf("Email: expected %s, got %s", expectedEmail, user.Email)
					}
					if user.PasswordHash != tt.passwordHash {
						t.Errorf("PasswordHash: expected %s, got %s", tt.passwordHash, user.PasswordHash)
					}
				}
			}
		})
	}
}

func TestUser_ValidateUsername(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "有効なユーザー名",
			username:    "validUser123",
			expectError: false,
		},
		{
			name:        "アンダースコアを含む",
			username:    "valid_user",
			expectError: false,
		},
		{
			name:        "ハイフンを含む",
			username:    "valid-user",
			expectError: false,
		},
		{
			name:        "空のユーザー名",
			username:    "",
			expectError: true,
			errorMsg:    "ユーザー名は必須です",
		},
		{
			name:        "短すぎるユーザー名",
			username:    "ab",
			expectError: true,
			errorMsg:    "ユーザー名は3文字以上である必要があります",
		},
		{
			name:        "長すぎるユーザー名",
			username:    strings.Repeat("a", 31),
			expectError: true,
			errorMsg:    "ユーザー名は30文字以内である必要があります",
		},
		{
			name:        "日本語を含む",
			username:    "ユーザー",
			expectError: true,
			errorMsg:    "ユーザー名には英数字、アンダースコア、ハイフンのみ使用できます",
		},
		{
			name:        "特殊文字を含む",
			username:    "user@123",
			expectError: true,
			errorMsg:    "ユーザー名には英数字、アンダースコア、ハイフンのみ使用できます",
		},
		{
			name:        "スペースを含む",
			username:    "user name",
			expectError: true,
			errorMsg:    "ユーザー名には英数字、アンダースコア、ハイフンのみ使用できます",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				Username: tt.username,
			}
			reason := user.ValidateUsername()

			if tt.expectError {
				if reason.IsOK() {
					t.Errorf("エラーが期待されたが、成功した")
				}
				if reason.Error() != tt.errorMsg {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMsg, reason.Error())
				}
			} else {
				if reason.IsNG() {
					t.Errorf("成功が期待されたが、エラーが発生: %s", reason.Error())
				}
			}
		})
	}
}

func TestUser_ValidateEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		expectedEmail string // 正規化後の期待値
		expectError   bool
		errorMsg      string
	}{
		{
			name:          "有効なメールアドレス",
			email:         "test@example.com",
			expectedEmail: "test@example.com",
			expectError:   false,
		},
		{
			name:          "大文字を含むメールアドレス",
			email:         "Test@Example.COM",
			expectedEmail: "test@example.com",
			expectError:   false,
		},
		{
			name:          "サブドメインを含む",
			email:         "test@sub.example.com",
			expectedEmail: "test@sub.example.com",
			expectError:   false,
		},
		{
			name:          "プラス記号を含む",
			email:         "test+tag@example.com",
			expectedEmail: "test+tag@example.com",
			expectError:   false,
		},
		{
			name:        "空のメールアドレス",
			email:       "",
			expectError: true,
			errorMsg:    "メールアドレスは必須です",
		},
		{
			name:        "@がない",
			email:       "testexample.com",
			expectError: true,
			errorMsg:    "メールアドレスの形式が正しくありません",
		},
		{
			name:        "ドメインがない",
			email:       "test@",
			expectError: true,
			errorMsg:    "メールアドレスの形式が正しくありません",
		},
		{
			name:        "ローカル部がない",
			email:       "@example.com",
			expectError: true,
			errorMsg:    "メールアドレスの形式が正しくありません",
		},
		{
			name:        "TLDがない",
			email:       "test@example",
			expectError: true,
			errorMsg:    "メールアドレスの形式が正しくありません",
		},
		{
			name:        "長すぎるメールアドレス",
			email:       strings.Repeat("a", 245) + "@example.com",
			expectError: true,
			errorMsg:    "メールアドレスは255文字以内である必要があります",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				Email: tt.email,
			}
			reason := user.ValidateEmail()

			if tt.expectError {
				if reason.IsOK() {
					t.Errorf("エラーが期待されたが、成功した")
				}
				if reason.Error() != tt.errorMsg {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMsg, reason.Error())
				}
			} else {
				if reason.IsNG() {
					t.Errorf("成功が期待されたが、エラーが発生: %s", reason.Error())
				}
				// 正規化の確認
				if user.Email != tt.expectedEmail {
					t.Errorf("期待されたメールアドレス: %s, 実際: %s", tt.expectedEmail, user.Email)
				}
			}
		})
	}
}

func TestUser_CanSendMorningCallTo(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		receiverID  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "他のユーザーへの送信",
			userID:      "user-001",
			receiverID:  "user-002",
			expectError: false,
		},
		{
			name:        "自分自身への送信",
			userID:      "user-001",
			receiverID:  "user-001",
			expectError: true,
			errorMsg:    "自分自身にモーニングコールを設定することはできません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				ID: tt.userID,
			}
			reason := user.CanSendMorningCallTo(tt.receiverID)

			if tt.expectError {
				if reason.IsOK() {
					t.Errorf("エラーが期待されたが、成功した")
				}
				if reason.Error() != tt.errorMsg {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMsg, reason.Error())
				}
			} else {
				if reason.IsNG() {
					t.Errorf("成功が期待されたが、エラーが発生: %s", reason.Error())
				}
			}
		})
	}
}

func TestUser_UpdateUsername(t *testing.T) {
	tests := []struct {
		name         string
		initialName  string
		newName      string
		expectError  bool
		expectedName string
		errorMsg     string
	}{
		{
			name:         "正常な更新",
			initialName:  "olduser",
			newName:      "newuser",
			expectError:  false,
			expectedName: "newuser",
		},
		{
			name:         "無効な新しい名前",
			initialName:  "olduser",
			newName:      "ab",
			expectError:  true,
			expectedName: "olduser", // ロールバックされる
			errorMsg:     "ユーザー名は3文字以上である必要があります",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				ID:        "user-001",
				Username:  tt.initialName,
				Email:     "test@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			}

			oldUpdatedAt := user.UpdatedAt
			reason := user.UpdateUsername(tt.newName)

			if tt.expectError {
				if reason.IsOK() {
					t.Errorf("エラーが期待されたが、成功した")
				}
				if reason.Error() != tt.errorMsg {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMsg, reason.Error())
				}
				if user.Username != tt.expectedName {
					t.Errorf("ロールバック後のユーザー名: expected %s, got %s", tt.expectedName, user.Username)
				}
				if !user.UpdatedAt.Equal(oldUpdatedAt) {
					t.Errorf("エラー時はUpdatedAtが更新されないべき")
				}
			} else {
				if reason.IsNG() {
					t.Errorf("成功が期待されたが、エラーが発生: %s", reason.Error())
				}
				if user.Username != tt.expectedName {
					t.Errorf("更新後のユーザー名: expected %s, got %s", tt.expectedName, user.Username)
				}
				if user.UpdatedAt.Equal(oldUpdatedAt) || user.UpdatedAt.Before(oldUpdatedAt) {
					t.Errorf("成功時はUpdatedAtが更新されるべき")
				}
			}
		})
	}
}

func TestUser_UpdateEmail(t *testing.T) {
	tests := []struct {
		name          string
		initialEmail  string
		newEmail      string
		expectError   bool
		expectedEmail string
		errorMsg      string
	}{
		{
			name:          "正常な更新",
			initialEmail:  "old@example.com",
			newEmail:      "new@example.com",
			expectError:   false,
			expectedEmail: "new@example.com",
		},
		{
			name:          "大文字を含むメールアドレス",
			initialEmail:  "old@example.com",
			newEmail:      "NEW@EXAMPLE.COM",
			expectError:   false,
			expectedEmail: "new@example.com", // 小文字に正規化
		},
		{
			name:          "無効な新しいメールアドレス",
			initialEmail:  "old@example.com",
			newEmail:      "invalid-email",
			expectError:   true,
			expectedEmail: "old@example.com", // ロールバックされる
			errorMsg:      "メールアドレスの形式が正しくありません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				ID:        "user-001",
				Username:  "testuser",
				Email:     tt.initialEmail,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			}

			oldUpdatedAt := user.UpdatedAt
			reason := user.UpdateEmail(tt.newEmail)

			if tt.expectError {
				if reason.IsOK() {
					t.Errorf("エラーが期待されたが、成功した")
				}
				if reason.Error() != tt.errorMsg {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMsg, reason.Error())
				}
				if user.Email != tt.expectedEmail {
					t.Errorf("ロールバック後のメールアドレス: expected %s, got %s", tt.expectedEmail, user.Email)
				}
				if !user.UpdatedAt.Equal(oldUpdatedAt) {
					t.Errorf("エラー時はUpdatedAtが更新されないべき")
				}
			} else {
				if reason.IsNG() {
					t.Errorf("成功が期待されたが、エラーが発生: %s", reason.Error())
				}
				if user.Email != tt.expectedEmail {
					t.Errorf("更新後のメールアドレス: expected %s, got %s", tt.expectedEmail, user.Email)
				}
				if user.UpdatedAt.Equal(oldUpdatedAt) || user.UpdatedAt.Before(oldUpdatedAt) {
					t.Errorf("成功時はUpdatedAtが更新されるべき")
				}
			}
		})
	}
}

func TestUser_Equals(t *testing.T) {
	user1 := &User{
		ID:       "user-001",
		Username: "user1",
		Email:    "user1@example.com",
	}

	user2 := &User{
		ID:       "user-001",
		Username: "different",
		Email:    "different@example.com",
	}

	user3 := &User{
		ID:       "user-002",
		Username: "user1",
		Email:    "user1@example.com",
	}

	tests := []struct {
		name     string
		user     *User
		other    *User
		expected bool
	}{
		{
			name:     "同じIDのユーザー",
			user:     user1,
			other:    user2,
			expected: true,
		},
		{
			name:     "異なるIDのユーザー",
			user:     user1,
			other:    user3,
			expected: false,
		},
		{
			name:     "nilとの比較",
			user:     user1,
			other:    nil,
			expected: false,
		},
		{
			name:     "自分自身との比較",
			user:     user1,
			other:    user1,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.Equals(tt.other)
			if result != tt.expected {
				t.Errorf("Equals() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "有効なパスワード",
			password:    "Password123!",
			expectError: false,
		},
		{
			name:        "複雑なパスワード",
			password:    "C0mpl3x!P@ssw0rd#2024",
			expectError: false,
		},
		{
			name:        "最小要件を満たすパスワード",
			password:    "Abcd123!",
			expectError: false,
		},
		{
			name:        "空のパスワード",
			password:    "",
			expectError: true,
			errorMsg:    "パスワードは必須です",
		},
		{
			name:        "短すぎるパスワード",
			password:    "Pass1!",
			expectError: true,
			errorMsg:    "パスワードは8文字以上である必要があります",
		},
		{
			name:        "長すぎるパスワード",
			password:    strings.Repeat("a", 95) + "A1!",
			expectError: false, // 100文字はOK
		},
		{
			name:        "100文字を超えるパスワード",
			password:    strings.Repeat("a", 98) + "A1!",
			expectError: true,
			errorMsg:    "パスワードは100文字以内である必要があります",
		},
		{
			name:        "大文字がない",
			password:    "password123!",
			expectError: true,
			errorMsg:    "パスワードは大文字、小文字、数字、特殊文字をそれぞれ1文字以上含む必要があります",
		},
		{
			name:        "小文字がない",
			password:    "PASSWORD123!",
			expectError: true,
			errorMsg:    "パスワードは大文字、小文字、数字、特殊文字をそれぞれ1文字以上含む必要があります",
		},
		{
			name:        "数字がない",
			password:    "Password!",
			expectError: true,
			errorMsg:    "パスワードは大文字、小文字、数字、特殊文字をそれぞれ1文字以上含む必要があります",
		},
		{
			name:        "特殊文字がない",
			password:    "Password123",
			expectError: true,
			errorMsg:    "パスワードは大文字、小文字、数字、特殊文字をそれぞれ1文字以上含む必要があります",
		},
		{
			name:        "様々な特殊文字を含む",
			password:    "Pass123@#$%",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := ValidatePassword(tt.password)

			if tt.expectError {
				if reason.IsOK() {
					t.Errorf("エラーが期待されたが、成功した")
				}
				if reason.Error() != tt.errorMsg {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMsg, reason.Error())
				}
			} else {
				if reason.IsNG() {
					t.Errorf("成功が期待されたが、エラーが発生: %s", reason.Error())
				}
			}
		})
	}
}
