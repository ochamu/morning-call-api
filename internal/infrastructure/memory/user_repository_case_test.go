package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
)

// TestUserRepository_CaseInsensitive は大小文字を区別しない検索のテスト
func TestUserRepository_CaseInsensitive(t *testing.T) {
	ctx := context.Background()

	t.Run("ユーザー名の大小文字を区別しない重複チェック", func(t *testing.T) {
		repo := NewUserRepository()

		// 最初のユーザーを作成
		user1 := &entity.User{
			ID:           "user1",
			Username:     "TestUser",
			Email:        "test@example.com",
			PasswordHash: "hash1",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Create(ctx, user1)
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		// 大小文字違いの同じユーザー名で作成試行
		testCases := []struct {
			name     string
			username string
			email    string
		}{
			{"全て小文字", "testuser", "test2@example.com"},
			{"全て大文字", "TESTUSER", "test3@example.com"},
			{"混在パターン1", "TestUser", "test4@example.com"},
			{"混在パターン2", "tEsTuSeR", "test5@example.com"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				user := &entity.User{
					ID:           "user_" + tc.name,
					Username:     tc.username,
					Email:        tc.email,
					PasswordHash: "hash",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}

				err := repo.Create(ctx, user)
				if !errors.Is(err, repository.ErrAlreadyExists) {
					t.Errorf("Expected ErrAlreadyExists for username %s, got %v", tc.username, err)
				}
			})
		}
	})

	t.Run("メールアドレスの大小文字を区別しない重複チェック", func(t *testing.T) {
		repo := NewUserRepository()

		// 最初のユーザーを作成
		user1 := &entity.User{
			ID:           "user1",
			Username:     "user1",
			Email:        "Test@Example.com",
			PasswordHash: "hash1",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Create(ctx, user1)
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		// 大小文字違いの同じメールアドレスで作成試行
		testCases := []struct {
			name     string
			username string
			email    string
		}{
			{"全て小文字", "user2", "test@example.com"},
			{"全て大文字", "user3", "TEST@EXAMPLE.COM"},
			{"混在パターン1", "user4", "Test@Example.com"},
			{"混在パターン2", "user5", "tEsT@eXaMpLe.CoM"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				user := &entity.User{
					ID:           "user_" + tc.name,
					Username:     tc.username,
					Email:        tc.email,
					PasswordHash: "hash",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}

				err := repo.Create(ctx, user)
				if !errors.Is(err, repository.ErrAlreadyExists) {
					t.Errorf("Expected ErrAlreadyExists for email %s, got %v", tc.email, err)
				}
			})
		}
	})

	t.Run("ユーザー名の大小文字を区別しない検索", func(t *testing.T) {
		repo := NewUserRepository()

		// ユーザーを作成
		user := &entity.User{
			ID:           "user1",
			Username:     "TestUser",
			Email:        "test@example.com",
			PasswordHash: "hash",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// 大小文字違いで検索
		testCases := []string{
			"testuser",
			"TESTUSER",
			"TestUser",
			"tEsTuSeR",
		}

		for _, username := range testCases {
			t.Run(username, func(t *testing.T) {
				// FindByUsernameテスト
				found, err := repo.FindByUsername(ctx, username)
				if err != nil {
					t.Errorf("FindByUsername(%s) failed: %v", username, err)
				}
				if found == nil {
					t.Errorf("FindByUsername(%s) returned nil", username)
				} else if found.ID != user.ID {
					t.Errorf("FindByUsername(%s) returned wrong user: got ID %s, want %s", username, found.ID, user.ID)
				}

				// ExistsByUsernameテスト
				exists, err := repo.ExistsByUsername(ctx, username)
				if err != nil {
					t.Errorf("ExistsByUsername(%s) failed: %v", username, err)
				}
				if !exists {
					t.Errorf("ExistsByUsername(%s) returned false, want true", username)
				}
			})
		}
	})

	t.Run("メールアドレスの大小文字を区別しない検索", func(t *testing.T) {
		repo := NewUserRepository()

		// ユーザーを作成
		user := &entity.User{
			ID:           "user1",
			Username:     "testuser",
			Email:        "Test@Example.com",
			PasswordHash: "hash",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// 大小文字違いで検索
		testCases := []string{
			"test@example.com",
			"TEST@EXAMPLE.COM",
			"Test@Example.com",
			"tEsT@eXaMpLe.CoM",
		}

		for _, email := range testCases {
			t.Run(email, func(t *testing.T) {
				// FindByEmailテスト
				found, err := repo.FindByEmail(ctx, email)
				if err != nil {
					t.Errorf("FindByEmail(%s) failed: %v", email, err)
				}
				if found == nil {
					t.Errorf("FindByEmail(%s) returned nil", email)
				} else if found.ID != user.ID {
					t.Errorf("FindByEmail(%s) returned wrong user: got ID %s, want %s", email, found.ID, user.ID)
				}

				// ExistsByEmailテスト
				exists, err := repo.ExistsByEmail(ctx, email)
				if err != nil {
					t.Errorf("ExistsByEmail(%s) failed: %v", email, err)
				}
				if !exists {
					t.Errorf("ExistsByEmail(%s) returned false, want true", email)
				}
			})
		}
	})

	t.Run("更新時の大小文字処理", func(t *testing.T) {
		repo := NewUserRepository()

		// 2つのユーザーを作成
		user1 := &entity.User{
			ID:           "user1",
			Username:     "User1",
			Email:        "user1@example.com",
			PasswordHash: "hash1",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		user2 := &entity.User{
			ID:           "user2",
			Username:     "User2",
			Email:        "user2@example.com",
			PasswordHash: "hash2",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := repo.Create(ctx, user1); err != nil {
			t.Fatalf("Failed to create user1: %v", err)
		}
		if err := repo.Create(ctx, user2); err != nil {
			t.Fatalf("Failed to create user2: %v", err)
		}

		// user1のユーザー名を大小文字違いのuser2に更新試行
		user1.Username = "USER2"
		err := repo.Update(ctx, user1)
		if !errors.Is(err, repository.ErrAlreadyExists) {
			t.Errorf("Expected ErrAlreadyExists when updating to existing username with different case, got %v", err)
		}

		// user1のメールを大小文字違いのuser2に更新試行
		user1.Username = "User1" // 元に戻す
		user1.Email = "USER2@EXAMPLE.COM"
		err = repo.Update(ctx, user1)
		if !errors.Is(err, repository.ErrAlreadyExists) {
			t.Errorf("Expected ErrAlreadyExists when updating to existing email with different case, got %v", err)
		}
	})

	t.Run("削除時の大小文字処理", func(t *testing.T) {
		repo := NewUserRepository()

		// ユーザーを作成
		user := &entity.User{
			ID:           "user1",
			Username:     "TestUser",
			Email:        "Test@Example.com",
			PasswordHash: "hash",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := repo.Create(ctx, user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// 削除
		if err := repo.Delete(ctx, user.ID); err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		// 大小文字違いで検索して存在しないことを確認
		testCases := []struct {
			name    string
			search  string
			isEmail bool
		}{
			{"小文字ユーザー名", "testuser", false},
			{"大文字ユーザー名", "TESTUSER", false},
			{"小文字メール", "test@example.com", true},
			{"大文字メール", "TEST@EXAMPLE.COM", true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if tc.isEmail {
					exists, err := repo.ExistsByEmail(ctx, tc.search)
					if err != nil {
						t.Errorf("ExistsByEmail(%s) failed: %v", tc.search, err)
					}
					if exists {
						t.Errorf("ExistsByEmail(%s) returned true after deletion, want false", tc.search)
					}
				} else {
					exists, err := repo.ExistsByUsername(ctx, tc.search)
					if err != nil {
						t.Errorf("ExistsByUsername(%s) failed: %v", tc.search, err)
					}
					if exists {
						t.Errorf("ExistsByUsername(%s) returned true after deletion, want false", tc.search)
					}
				}
			})
		}
	})
}
