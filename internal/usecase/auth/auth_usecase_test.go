package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/infrastructure/auth"
	"github.com/ochamu/morning-call-api/internal/infrastructure/memory"
)

func TestNewAuthUseCase(t *testing.T) {
	userRepo := memory.NewUserRepository()
	passwordService := auth.NewPasswordService()

	uc := NewAuthUseCase(userRepo, passwordService)

	if uc == nil {
		t.Fatal("NewAuthUseCase returned nil")
	}
	if uc.userRepo == nil {
		t.Error("userRepo is nil")
	}
	if uc.passwordService == nil {
		t.Error("passwordService is nil")
	}
	if uc.sessions == nil {
		t.Error("sessions map is nil")
	}
	if uc.sessionTimeout != 24*time.Hour {
		t.Errorf("sessionTimeout = %v, want %v", uc.sessionTimeout, 24*time.Hour)
	}
}

func TestAuthUseCase_Login(t *testing.T) {
	ctx := context.Background()
	userRepo := memory.NewUserRepository()
	passwordService := auth.NewPasswordService()
	uc := NewAuthUseCase(userRepo, passwordService)

	// テスト用ユーザーを作成
	hashedPassword, err := passwordService.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	testUser := &entity.User{
		ID:           "user1",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		input   LoginInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功ケース",
			input: LoginInput{
				Username: "testuser",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "ユーザー名が空",
			input: LoginInput{
				Username: "",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "ユーザー名は必須です",
		},
		{
			name: "パスワードが空",
			input: LoginInput{
				Username: "testuser",
				Password: "",
			},
			wantErr: true,
			errMsg:  "パスワードは必須です",
		},
		{
			name: "存在しないユーザー",
			input: LoginInput{
				Username: "nonexistent",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "ユーザー名またはパスワードが間違っています",
		},
		{
			name: "パスワードが間違っている",
			input: LoginInput{
				Username: "testuser",
				Password: "wrongpassword",
			},
			wantErr: true,
			errMsg:  "ユーザー名またはパスワードが間違っています",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Login(ctx, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message = %v, want contains %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if output == nil {
					t.Error("output is nil")
				} else {
					if output.SessionID == "" {
						t.Error("SessionID is empty")
					}
					if output.User == nil {
						t.Error("User is nil")
					} else if output.User.Username != tt.input.Username {
						t.Errorf("User.Username = %v, want %v", output.User.Username, tt.input.Username)
					}
				}
			}
		})
	}
}

func TestAuthUseCase_Logout(t *testing.T) {
	ctx := context.Background()
	userRepo := memory.NewUserRepository()
	passwordService := auth.NewPasswordService()
	uc := NewAuthUseCase(userRepo, passwordService)

	// テスト用ユーザーとセッションを作成
	hashedPassword, err := passwordService.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	testUser := &entity.User{
		ID:           "user1",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// ログインしてセッションを作成
	loginOutput, err := uc.Login(ctx, LoginInput{
		Username: "testuser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	tests := []struct {
		name      string
		sessionID string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "成功ケース",
			sessionID: loginOutput.SessionID,
			wantErr:   false,
		},
		{
			name:      "セッションIDが空",
			sessionID: "",
			wantErr:   true,
			errMsg:    "セッションIDは必須です",
		},
		{
			name:      "無効なセッション",
			sessionID: "invalid-session-id",
			wantErr:   true,
			errMsg:    "無効なセッションです",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.Logout(ctx, tt.sessionID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message = %v, want contains %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// ログアウト後はセッションが無効になることを確認
				if _, err := uc.GetCurrentUser(ctx, tt.sessionID); err == nil {
					t.Error("expected error after logout but got nil")
				}
			}
		})
	}
}

func TestAuthUseCase_GetCurrentUser(t *testing.T) {
	ctx := context.Background()
	userRepo := memory.NewUserRepository()
	passwordService := auth.NewPasswordService()
	uc := NewAuthUseCase(userRepo, passwordService)

	// テスト用ユーザーを作成
	hashedPassword, err := passwordService.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	testUser := &entity.User{
		ID:           "user1",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// ログインしてセッションを作成
	loginOutput, err := uc.Login(ctx, LoginInput{
		Username: "testuser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	tests := []struct {
		name      string
		sessionID string
		wantErr   bool
		errMsg    string
		wantUser  bool
	}{
		{
			name:      "成功ケース",
			sessionID: loginOutput.SessionID,
			wantErr:   false,
			wantUser:  true,
		},
		{
			name:      "セッションIDが空",
			sessionID: "",
			wantErr:   true,
			errMsg:    "セッションIDは必須です",
			wantUser:  false,
		},
		{
			name:      "無効なセッション",
			sessionID: "invalid-session-id",
			wantErr:   true,
			errMsg:    "セッションが見つかりません",
			wantUser:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := uc.GetCurrentUser(ctx, tt.sessionID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message = %v, want contains %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.wantUser {
					if user == nil {
						t.Error("user is nil")
					} else if user.Username != "testuser" {
						t.Errorf("user.Username = %v, want %v", user.Username, "testuser")
					}
				}
			}
		})
	}
}

func TestAuthUseCase_GetCurrentUser_ExpiredSession(t *testing.T) {
	ctx := context.Background()
	userRepo := memory.NewUserRepository()
	passwordService := auth.NewPasswordService()
	uc := NewAuthUseCase(userRepo, passwordService)

	// セッションタイムアウトを短く設定
	uc.sessionTimeout = 1 * time.Millisecond

	// テスト用ユーザーを作成
	hashedPassword, err := passwordService.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	testUser := &entity.User{
		ID:           "user1",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// ログインしてセッションを作成
	loginOutput, err := uc.Login(ctx, LoginInput{
		Username: "testuser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	// セッションの有効期限が切れるまで待つ
	time.Sleep(2 * time.Millisecond)

	// 期限切れのセッションでユーザー取得を試みる
	_, err = uc.GetCurrentUser(ctx, loginOutput.SessionID)
	if err == nil {
		t.Error("expected error for expired session but got nil")
	} else if !strings.Contains(err.Error(), "セッションの有効期限が切れています") {
		t.Errorf("error message = %v, want contains '有効期限が切れています'", err.Error())
	}
}

func TestAuthUseCase_ValidateSession(t *testing.T) {
	ctx := context.Background()
	userRepo := memory.NewUserRepository()
	passwordService := auth.NewPasswordService()
	uc := NewAuthUseCase(userRepo, passwordService)

	// テスト用ユーザーを作成
	hashedPassword, err := passwordService.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	testUser := &entity.User{
		ID:           "user1",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// ログインしてセッションを作成
	loginOutput, err := uc.Login(ctx, LoginInput{
		Username: "testuser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	tests := []struct {
		name      string
		sessionID string
		want      bool
	}{
		{
			name:      "有効なセッション",
			sessionID: loginOutput.SessionID,
			want:      true,
		},
		{
			name:      "セッションIDが空",
			sessionID: "",
			want:      false,
		},
		{
			name:      "無効なセッション",
			sessionID: "invalid-session-id",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := uc.ValidateSession(ctx, tt.sessionID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if valid != tt.want {
				t.Errorf("ValidateSession() = %v, want %v", valid, tt.want)
			}
		})
	}
}

func TestAuthUseCase_CleanupExpiredSessions(t *testing.T) {
	userRepo := memory.NewUserRepository()
	passwordService := auth.NewPasswordService()
	uc := NewAuthUseCase(userRepo, passwordService)

	// セッションタイムアウトを短く設定
	uc.sessionTimeout = 1 * time.Millisecond

	// テスト用セッションを作成
	sessionID1, _ := uc.createSession("user1")
	sessionID2, _ := uc.createSession("user2")

	// セッションが存在することを確認
	if len(uc.sessions) != 2 {
		t.Errorf("sessions count = %d, want 2", len(uc.sessions))
	}

	// セッションの有効期限が切れるまで待つ
	time.Sleep(2 * time.Millisecond)

	// 期限切れセッションをクリーンアップ
	uc.CleanupExpiredSessions()

	// セッションが削除されたことを確認
	if len(uc.sessions) != 0 {
		t.Errorf("sessions count = %d, want 0", len(uc.sessions))
	}

	// 削除されたセッションIDで取得を試みる
	if _, err := uc.getSession(sessionID1); err == nil {
		t.Error("expected error for deleted session but got nil")
	}
	if _, err := uc.getSession(sessionID2); err == nil {
		t.Error("expected error for deleted session but got nil")
	}
}

func TestAuthUseCase_GetCurrentUser_DeletedUser(t *testing.T) {
	ctx := context.Background()
	userRepo := memory.NewUserRepository()
	passwordService := auth.NewPasswordService()
	uc := NewAuthUseCase(userRepo, passwordService)

	// テスト用ユーザーを作成
	hashedPassword, err := passwordService.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	testUser := &entity.User{
		ID:           "user1",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// ログインしてセッションを作成
	loginOutput, err := uc.Login(ctx, LoginInput{
		Username: "testuser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	// ユーザーを削除
	if err := userRepo.Delete(ctx, "user1"); err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	// 削除されたユーザーでセッションからユーザー取得を試みる
	_, err = uc.GetCurrentUser(ctx, loginOutput.SessionID)
	if err == nil {
		t.Error("expected error for deleted user but got nil")
	} else if !strings.Contains(err.Error(), "ユーザーが見つかりません") {
		t.Errorf("error message = %v, want contains 'ユーザーが見つかりません'", err.Error())
	}

	// セッションも削除されていることを確認
	if _, err := uc.getSession(loginOutput.SessionID); err == nil {
		t.Error("expected session to be deleted but it still exists")
	}
}
