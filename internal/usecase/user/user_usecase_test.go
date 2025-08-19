package user

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
)

// mockPasswordService はテスト用のモックパスワードサービス
type mockPasswordService struct {
	shouldFailHash   bool
	shouldFailVerify bool
}

func (m *mockPasswordService) HashPassword(password string) (string, error) {
	if m.shouldFailHash {
		return "", fmt.Errorf("hash failed")
	}
	// テスト用の簡単なハッシュ（実際のハッシュではない）
	return "hashed_" + password, nil
}

func (m *mockPasswordService) VerifyPassword(password, passwordHash string) (bool, error) {
	if m.shouldFailVerify {
		return false, fmt.Errorf("verify failed")
	}
	// テスト用の簡単な検証
	return passwordHash == "hashed_"+password, nil
}

// mockUserRepository はテスト用のモックリポジトリ
type mockUserRepository struct {
	users            map[string]*entity.User
	usersByUsername  map[string]*entity.User
	usersByEmail     map[string]*entity.User
	shouldFailCreate bool
	shouldFailFind   bool
	shouldFailExists bool
}

// newMockUserRepository は新しいモックリポジトリを作成する
func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:           make(map[string]*entity.User),
		usersByUsername: make(map[string]*entity.User),
		usersByEmail:    make(map[string]*entity.User),
	}
}

func (r *mockUserRepository) Create(ctx context.Context, user *entity.User) error {
	_ = ctx // テスト用モックのため未使用
	if r.shouldFailCreate {
		return repository.ErrConnectionFailed
	}

	// 重複チェック
	if _, exists := r.users[user.ID]; exists {
		return repository.ErrAlreadyExists
	}
	if _, exists := r.usersByUsername[strings.ToLower(user.Username)]; exists {
		return repository.ErrAlreadyExists
	}
	if _, exists := r.usersByEmail[strings.ToLower(user.Email)]; exists {
		return repository.ErrAlreadyExists
	}

	r.users[user.ID] = user
	r.usersByUsername[strings.ToLower(user.Username)] = user
	r.usersByEmail[strings.ToLower(user.Email)] = user
	return nil
}

func (r *mockUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	_ = ctx // テスト用モックのため未使用
	if r.shouldFailFind {
		return nil, repository.ErrConnectionFailed
	}

	user, exists := r.users[id]
	if !exists {
		return nil, repository.ErrNotFound
	}
	return user, nil
}

func (r *mockUserRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	_ = ctx // テスト用モックのため未使用
	if r.shouldFailFind {
		return nil, repository.ErrConnectionFailed
	}

	user, exists := r.usersByUsername[strings.ToLower(username)]
	if !exists {
		return nil, repository.ErrNotFound
	}
	return user, nil
}

func (r *mockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	_ = ctx // テスト用モックのため未使用
	if r.shouldFailFind {
		return nil, repository.ErrConnectionFailed
	}

	user, exists := r.usersByEmail[strings.ToLower(email)]
	if !exists {
		return nil, repository.ErrNotFound
	}
	return user, nil
}

func (r *mockUserRepository) Update(ctx context.Context, user *entity.User) error {
	_ = ctx // テスト用モックのため未使用
	if _, exists := r.users[user.ID]; !exists {
		return repository.ErrNotFound
	}
	r.users[user.ID] = user
	return nil
}

func (r *mockUserRepository) Delete(ctx context.Context, id string) error {
	_ = ctx // テスト用モックのため未使用
	user, exists := r.users[id]
	if !exists {
		return repository.ErrNotFound
	}
	delete(r.users, id)
	delete(r.usersByUsername, strings.ToLower(user.Username))
	delete(r.usersByEmail, strings.ToLower(user.Email))
	return nil
}

func (r *mockUserRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	_ = ctx // テスト用モックのため未使用
	if r.shouldFailExists {
		return false, repository.ErrConnectionFailed
	}
	_, exists := r.users[id]
	return exists, nil
}

func (r *mockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	_ = ctx // テスト用モックのため未使用
	if r.shouldFailExists {
		return false, repository.ErrConnectionFailed
	}
	_, exists := r.usersByUsername[strings.ToLower(username)]
	return exists, nil
}

func (r *mockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_ = ctx // テスト用モックのため未使用
	if r.shouldFailExists {
		return false, repository.ErrConnectionFailed
	}
	_, exists := r.usersByEmail[strings.ToLower(email)]
	return exists, nil
}

func (r *mockUserRepository) FindAll(ctx context.Context, offset, limit int) ([]*entity.User, error) {
	_ = ctx // テスト用モックのため未使用
	users := make([]*entity.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	// 簡単なページネーション実装
	start := offset
	if start > len(users) {
		return []*entity.User{}, nil
	}

	end := offset + limit
	if end > len(users) {
		end = len(users)
	}

	return users[start:end], nil
}

func (r *mockUserRepository) Count(ctx context.Context) (int, error) {
	_ = ctx // テスト用モックのため未使用
	return len(r.users), nil
}

// TestRegister_Success はユーザー登録の成功ケースをテストする
func TestRegister_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    RegisterInput
		wantUser bool
	}{
		{
			name: "正常なユーザー登録",
			input: RegisterInput{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "Password123!",
			},
			wantUser: true,
		},
		{
			name: "複雑なパスワードでの登録",
			input: RegisterInput{
				Username: "complexuser",
				Email:    "complex@example.com",
				Password: "C0mpl3x!P@ssw0rd#2024",
			},
			wantUser: true,
		},
		{
			name: "最小文字数のユーザー名",
			input: RegisterInput{
				Username: "abc",
				Email:    "abc@example.com",
				Password: "Abc123!@",
			},
			wantUser: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := newMockUserRepository()
			passwordService := &mockPasswordService{}
			uc := NewUserUseCase(repo, passwordService)
			ctx := context.Background()

			// Act
			output, err := uc.Register(ctx, tt.input)

			// Assert
			if err != nil {
				t.Errorf("Register() error = %v, want nil", err)
				return
			}

			if tt.wantUser && output.User == nil {
				t.Error("Register() User is nil, want user")
				return
			}

			if output.User != nil {
				if output.User.Username != tt.input.Username {
					t.Errorf("Register() Username = %v, want %v", output.User.Username, tt.input.Username)
				}
				if output.User.Email != strings.ToLower(tt.input.Email) {
					t.Errorf("Register() Email = %v, want %v", output.User.Email, strings.ToLower(tt.input.Email))
				}
				if output.User.PasswordHash == "" {
					t.Error("Register() PasswordHash is empty")
				}
				if output.User.PasswordHash == tt.input.Password {
					t.Error("Register() Password was not hashed")
				}
			}
		})
	}
}

// TestRegister_ValidationErrors はバリデーションエラーのテストケース
func TestRegister_ValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		input      RegisterInput
		wantErrMsg string
	}{
		{
			name: "空のユーザー名",
			input: RegisterInput{
				Username: "",
				Email:    "test@example.com",
				Password: "Password123!",
			},
			wantErrMsg: "ユーザー名、メールアドレス、パスワードは必須です",
		},
		{
			name: "空のメールアドレス",
			input: RegisterInput{
				Username: "testuser",
				Email:    "",
				Password: "Password123!",
			},
			wantErrMsg: "ユーザー名、メールアドレス、パスワードは必須です",
		},
		{
			name: "空のパスワード",
			input: RegisterInput{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "",
			},
			wantErrMsg: "ユーザー名、メールアドレス、パスワードは必須です",
		},
		{
			name: "短すぎるパスワード",
			input: RegisterInput{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "Pass1!",
			},
			wantErrMsg: "パスワードは8文字以上である必要があります",
		},
		{
			name: "大文字がないパスワード",
			input: RegisterInput{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123!",
			},
			wantErrMsg: "パスワードは大文字、小文字、数字、特殊文字をそれぞれ1文字以上含む必要があります",
		},
		{
			name: "小文字がないパスワード",
			input: RegisterInput{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "PASSWORD123!",
			},
			wantErrMsg: "パスワードは大文字、小文字、数字、特殊文字をそれぞれ1文字以上含む必要があります",
		},
		{
			name: "数字がないパスワード",
			input: RegisterInput{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "Password!",
			},
			wantErrMsg: "パスワードは大文字、小文字、数字、特殊文字をそれぞれ1文字以上含む必要があります",
		},
		{
			name: "特殊文字がないパスワード",
			input: RegisterInput{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "Password123",
			},
			wantErrMsg: "パスワードは大文字、小文字、数字、特殊文字をそれぞれ1文字以上含む必要があります",
		},
		{
			name: "短すぎるユーザー名",
			input: RegisterInput{
				Username: "ab",
				Email:    "test@example.com",
				Password: "Password123!",
			},
			wantErrMsg: "ユーザー名は3文字以上である必要があります",
		},
		{
			name: "不正な文字を含むユーザー名",
			input: RegisterInput{
				Username: "test@user",
				Email:    "test@example.com",
				Password: "Password123!",
			},
			wantErrMsg: "ユーザー名には英数字、アンダースコア、ハイフンのみ使用できます",
		},
		{
			name: "不正なメールアドレス形式",
			input: RegisterInput{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "Password123!",
			},
			wantErrMsg: "メールアドレスの形式が正しくありません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := newMockUserRepository()
			passwordService := &mockPasswordService{}
			uc := NewUserUseCase(repo, passwordService)
			ctx := context.Background()

			// Act
			_, err := uc.Register(ctx, tt.input)

			// Assert
			if err == nil {
				t.Error("Register() error = nil, want error")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("Register() error = %v, want error containing %v", err, tt.wantErrMsg)
			}
		})
	}
}

// TestRegister_DuplicateErrors は重複エラーのテストケース
func TestRegister_DuplicateErrors(t *testing.T) {
	tests := []struct {
		name         string
		existingUser *entity.User
		input        RegisterInput
		wantErrMsg   string
	}{
		{
			name: "既存のユーザー名",
			existingUser: &entity.User{
				ID:           "existing-id",
				Username:     "existinguser",
				Email:        "existing@example.com",
				PasswordHash: "hash",
			},
			input: RegisterInput{
				Username: "existinguser",
				Email:    "new@example.com",
				Password: "Password123!",
			},
			wantErrMsg: "ユーザー名 'existinguser' は既に使用されています",
		},
		{
			name: "既存のメールアドレス",
			existingUser: &entity.User{
				ID:           "existing-id",
				Username:     "existinguser",
				Email:        "existing@example.com",
				PasswordHash: "hash",
			},
			input: RegisterInput{
				Username: "newuser",
				Email:    "existing@example.com",
				Password: "Password123!",
			},
			wantErrMsg: "メールアドレス 'existing@example.com' は既に登録されています",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := newMockUserRepository()
			if tt.existingUser != nil {
				repo.users[tt.existingUser.ID] = tt.existingUser
				repo.usersByUsername[strings.ToLower(tt.existingUser.Username)] = tt.existingUser
				repo.usersByEmail[strings.ToLower(tt.existingUser.Email)] = tt.existingUser
			}
			passwordService := &mockPasswordService{}
			uc := NewUserUseCase(repo, passwordService)
			ctx := context.Background()

			// Act
			_, err := uc.Register(ctx, tt.input)

			// Assert
			if err == nil {
				t.Error("Register() error = nil, want error")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("Register() error = %v, want error containing %v", err, tt.wantErrMsg)
			}
		})
	}
}

// TestGetByID はIDによるユーザー取得のテスト
func TestGetByID(t *testing.T) {
	// Arrange
	repo := newMockUserRepository()
	passwordService := &mockPasswordService{}
	uc := NewUserUseCase(repo, passwordService)
	ctx := context.Background()

	// 既存ユーザーを作成
	existingUser := &entity.User{
		ID:           "test-id",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash",
	}
	repo.users[existingUser.ID] = existingUser

	tests := []struct {
		name       string
		userID     string
		wantUser   bool
		wantErrMsg string
	}{
		{
			name:     "存在するユーザー",
			userID:   "test-id",
			wantUser: true,
		},
		{
			name:       "存在しないユーザー",
			userID:     "non-existent",
			wantUser:   false,
			wantErrMsg: "ユーザーが見つかりません",
		},
		{
			name:       "空のID",
			userID:     "",
			wantUser:   false,
			wantErrMsg: "ユーザーIDは必須です",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			user, err := uc.GetByID(ctx, tt.userID)

			// Assert
			if tt.wantUser {
				if err != nil {
					t.Errorf("GetByID() error = %v, want nil", err)
				}
				if user == nil {
					t.Error("GetByID() user is nil, want user")
				}
			} else {
				if err == nil {
					t.Error("GetByID() error = nil, want error")
				} else if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("GetByID() error = %v, want error containing %v", err, tt.wantErrMsg)
				}
			}
		})
	}
}

// TestVerifyPassword はパスワード検証のテスト
func TestVerifyPassword(t *testing.T) {
	// Arrange
	repo := newMockUserRepository()
	passwordService := &mockPasswordService{}
	uc := NewUserUseCase(repo, passwordService)
	ctx := context.Background()

	// テストユーザーを登録
	input := RegisterInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "Password123!",
	}
	output, err := uc.Register(ctx, input)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	tests := []struct {
		name       string
		username   string
		password   string
		wantUser   bool
		wantErrMsg string
	}{
		{
			name:     "正しいパスワード",
			username: "testuser",
			password: "Password123!",
			wantUser: true,
		},
		{
			name:       "間違ったパスワード",
			username:   "testuser",
			password:   "WrongPassword123!",
			wantUser:   false,
			wantErrMsg: "パスワードが正しくありません",
		},
		{
			name:       "存在しないユーザー",
			username:   "nonexistent",
			password:   "Password123!",
			wantUser:   false,
			wantErrMsg: "ユーザーが見つかりません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			user, err := uc.VerifyPassword(ctx, tt.username, tt.password)

			// Assert
			if tt.wantUser {
				if err != nil {
					t.Errorf("VerifyPassword() error = %v, want nil", err)
				}
				if user == nil {
					t.Error("VerifyPassword() user is nil, want user")
				} else if user.ID != output.User.ID {
					t.Errorf("VerifyPassword() user.ID = %v, want %v", user.ID, output.User.ID)
				}
			} else {
				if err == nil {
					t.Error("VerifyPassword() error = nil, want error")
				} else if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("VerifyPassword() error = %v, want error containing %v", err, tt.wantErrMsg)
				}
			}
		})
	}
}
