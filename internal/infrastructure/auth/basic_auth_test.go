package auth

import (
	"testing"
)

func TestPasswordService_HashPassword(t *testing.T) {
	service := NewPasswordService()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "正常なパスワード",
			password: "Password123!",
			wantErr:  false,
		},
		{
			name:     "長いパスワード",
			password: "ThisIsAVeryLongPasswordWith123!@#SpecialCharacters",
			wantErr:  false,
		},
		{
			name:     "72バイトを超えるパスワード",
			password: "ThisIsAVeryLongPasswordThatExceedsTheSeventyTwoByteLimitOf_Bcrypt_1234567890!@#$%^&*()",
			wantErr:  false,
		},
		{
			name:     "空のパスワード",
			password: "",
			wantErr:  true,
		},
		{
			name:     "特殊文字を含むパスワード",
			password: "P@ssw0rd!#$%^&*()",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := service.HashPassword(tt.password)

			if tt.wantErr {
				if err == nil {
					t.Error("HashPassword() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Errorf("HashPassword() error = %v, want nil", err)
				return
			}

			if hash == "" {
				t.Error("HashPassword() hash is empty")
			}

			if hash == tt.password {
				t.Error("HashPassword() returned plain password, not hashed")
			}

			// bcryptはソルトがランダムなため、同じパスワードでも異なるハッシュが生成される
			// 代わりに、生成されたハッシュで元のパスワードが検証できることを確認
			valid, err := service.VerifyPassword(tt.password, hash)
			if err != nil {
				t.Errorf("VerifyPassword() error = %v", err)
				return
			}
			if !valid {
				t.Error("VerifyPassword() failed to verify the password with its own hash")
			}
		})
	}
}

func TestPasswordService_VerifyPassword(t *testing.T) {
	service := NewPasswordService()

	// テスト用のパスワードとハッシュを準備
	testPassword := "TestPassword123!"
	testHash, err := service.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("Failed to create test hash: %v", err)
	}

	tests := []struct {
		name         string
		password     string
		passwordHash string
		want         bool
		wantErr      bool
	}{
		{
			name:         "正しいパスワード",
			password:     testPassword,
			passwordHash: testHash,
			want:         true,
			wantErr:      false,
		},
		{
			name:         "間違ったパスワード",
			password:     "WrongPassword123!",
			passwordHash: testHash,
			want:         false,
			wantErr:      false,
		},
		{
			name:         "空のパスワード",
			password:     "",
			passwordHash: testHash,
			want:         false,
			wantErr:      true,
		},
		{
			name:         "空のハッシュ",
			password:     testPassword,
			passwordHash: "",
			want:         false,
			wantErr:      true,
		},
		{
			name:         "両方空",
			password:     "",
			passwordHash: "",
			want:         false,
			wantErr:      true,
		},
		{
			name:         "大文字小文字の違い",
			password:     "testpassword123!",
			passwordHash: testHash,
			want:         false,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.VerifyPassword(tt.password, tt.passwordHash)

			if tt.wantErr {
				if err == nil {
					t.Error("VerifyPassword() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Errorf("VerifyPassword() error = %v, want nil", err)
				return
			}

			if got != tt.want {
				t.Errorf("VerifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordService_DifferentPasswords(t *testing.T) {
	service := NewPasswordService()

	passwords := []string{
		"Password123!",
		"DifferentPass456@",
		"AnotherOne789#",
	}

	hashes := make(map[string]string)

	// 各パスワードのハッシュを生成
	for _, pwd := range passwords {
		hash, err := service.HashPassword(pwd)
		if err != nil {
			t.Errorf("Failed to hash password %s: %v", pwd, err)
			continue
		}
		hashes[pwd] = hash
	}

	// 異なるパスワードが異なるハッシュを生成することを確認
	for i, pwd1 := range passwords {
		for j, pwd2 := range passwords {
			if i != j {
				if hashes[pwd1] == hashes[pwd2] {
					t.Errorf("Different passwords produced same hash: %s and %s", pwd1, pwd2)
				}

				// 異なるパスワードのハッシュで検証が失敗することを確認
				valid, err := service.VerifyPassword(pwd1, hashes[pwd2])
				if err != nil {
					t.Errorf("VerifyPassword() error = %v", err)
					continue
				}
				if valid {
					t.Errorf("VerifyPassword() accepted wrong password: %s for hash of %s", pwd1, pwd2)
				}
			}
		}
	}
}
