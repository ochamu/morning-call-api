package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

func TestUserRepository_Create(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func(*UserRepository)
		user    *entity.User
		wantErr error
	}{
		{
			name:    "新規ユーザーの作成成功",
			user:    createTestUser("user1", "testuser", "test@example.com"),
			wantErr: nil,
		},
		{
			name: "同一IDのユーザーが既に存在する場合",
			setup: func(r *UserRepository) {
				r.Create(ctx, createTestUser("user1", "existing", "existing@example.com"))
			},
			user:    createTestUser("user1", "newuser", "new@example.com"),
			wantErr: repository.ErrAlreadyExists,
		},
		{
			name: "同一ユーザー名が既に存在する場合",
			setup: func(r *UserRepository) {
				r.Create(ctx, createTestUser("user1", "testuser", "existing@example.com"))
			},
			user:    createTestUser("user2", "testuser", "new@example.com"),
			wantErr: repository.ErrAlreadyExists,
		},
		{
			name: "同一メールアドレスが既に存在する場合",
			setup: func(r *UserRepository) {
				r.Create(ctx, createTestUser("user1", "existing", "test@example.com"))
			},
			user:    createTestUser("user2", "newuser", "test@example.com"),
			wantErr: repository.ErrAlreadyExists,
		},
		{
			name:    "nilユーザーの場合",
			user:    nil,
			wantErr: repository.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewUserRepository()
			if tt.setup != nil {
				tt.setup(repo)
			}

			err := repo.Create(ctx, tt.user)
			if err != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository()

	// テスト用ユーザーを作成
	user := createTestUser("user1", "testuser", "test@example.com")
	repo.Create(ctx, user)

	tests := []struct {
		name    string
		id      string
		want    *entity.User
		wantErr error
	}{
		{
			name:    "存在するユーザーの検索",
			id:      "user1",
			want:    user,
			wantErr: nil,
		},
		{
			name:    "存在しないユーザーの検索",
			id:      "nonexistent",
			want:    nil,
			wantErr: repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindByID(ctx, tt.id)
			if err != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && got != nil {
				if got.ID != tt.want.ID || got.Username != tt.want.Username || got.Email != tt.want.Email {
					t.Errorf("FindByID() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestUserRepository_FindByUsername(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository()

	// テスト用ユーザーを作成
	user := createTestUser("user1", "testuser", "test@example.com")
	repo.Create(ctx, user)

	tests := []struct {
		name     string
		username string
		want     *entity.User
		wantErr  error
	}{
		{
			name:     "存在するユーザー名での検索",
			username: "testuser",
			want:     user,
			wantErr:  nil,
		},
		{
			name:     "存在しないユーザー名での検索",
			username: "nonexistent",
			want:     nil,
			wantErr:  repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindByUsername(ctx, tt.username)
			if err != tt.wantErr {
				t.Errorf("FindByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && got != nil {
				if got.ID != tt.want.ID || got.Username != tt.want.Username {
					t.Errorf("FindByUsername() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository()

	// テスト用ユーザーを作成
	user := createTestUser("user1", "testuser", "test@example.com")
	repo.Create(ctx, user)

	tests := []struct {
		name    string
		email   string
		want    *entity.User
		wantErr error
	}{
		{
			name:    "存在するメールアドレスでの検索",
			email:   "test@example.com",
			want:    user,
			wantErr: nil,
		},
		{
			name:    "存在しないメールアドレスでの検索",
			email:   "nonexistent@example.com",
			want:    nil,
			wantErr: repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindByEmail(ctx, tt.email)
			if err != tt.wantErr {
				t.Errorf("FindByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && got != nil {
				if got.ID != tt.want.ID || got.Email != tt.want.Email {
					t.Errorf("FindByEmail() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func(*UserRepository)
		update  *entity.User
		wantErr error
	}{
		{
			name: "ユーザー情報の更新成功",
			setup: func(r *UserRepository) {
				r.Create(ctx, createTestUser("user1", "oldname", "old@example.com"))
			},
			update:  createTestUser("user1", "newname", "new@example.com"),
			wantErr: nil,
		},
		{
			name:    "存在しないユーザーの更新",
			update:  createTestUser("nonexistent", "name", "email@example.com"),
			wantErr: repository.ErrNotFound,
		},
		{
			name: "ユーザー名が他のユーザーと重複",
			setup: func(r *UserRepository) {
				r.Create(ctx, createTestUser("user1", "user1", "user1@example.com"))
				r.Create(ctx, createTestUser("user2", "user2", "user2@example.com"))
			},
			update:  createTestUser("user1", "user2", "newemail@example.com"),
			wantErr: repository.ErrAlreadyExists,
		},
		{
			name: "メールアドレスが他のユーザーと重複",
			setup: func(r *UserRepository) {
				r.Create(ctx, createTestUser("user1", "user1", "user1@example.com"))
				r.Create(ctx, createTestUser("user2", "user2", "user2@example.com"))
			},
			update:  createTestUser("user1", "newname", "user2@example.com"),
			wantErr: repository.ErrAlreadyExists,
		},
		{
			name:    "nilユーザーの更新",
			update:  nil,
			wantErr: repository.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewUserRepository()
			if tt.setup != nil {
				tt.setup(repo)
			}

			err := repo.Update(ctx, tt.update)
			if err != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 更新が成功した場合、インデックスが正しく更新されているか確認
			if err == nil && tt.update != nil {
				// ユーザー名で検索できるか
				found, _ := repo.FindByUsername(ctx, tt.update.Username)
				if found == nil || found.ID != tt.update.ID {
					t.Errorf("Updated user not found by username")
				}

				// メールアドレスで検索できるか
				found, _ = repo.FindByEmail(ctx, tt.update.Email)
				if found == nil || found.ID != tt.update.ID {
					t.Errorf("Updated user not found by email")
				}
			}
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func(*UserRepository)
		id      string
		wantErr error
	}{
		{
			name: "ユーザーの削除成功",
			setup: func(r *UserRepository) {
				r.Create(ctx, createTestUser("user1", "testuser", "test@example.com"))
			},
			id:      "user1",
			wantErr: nil,
		},
		{
			name:    "存在しないユーザーの削除",
			id:      "nonexistent",
			wantErr: repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewUserRepository()
			if tt.setup != nil {
				tt.setup(repo)
			}

			err := repo.Delete(ctx, tt.id)
			if err != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 削除が成功した場合、ユーザーが本当に削除されているか確認
			if err == nil {
				_, findErr := repo.FindByID(ctx, tt.id)
				if findErr != repository.ErrNotFound {
					t.Errorf("Deleted user still exists")
				}
			}
		})
	}
}

func TestUserRepository_ExistsByID(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository()

	// テスト用ユーザーを作成
	repo.Create(ctx, createTestUser("user1", "testuser", "test@example.com"))

	tests := []struct {
		name    string
		id      string
		want    bool
		wantErr error
	}{
		{
			name:    "存在するユーザーID",
			id:      "user1",
			want:    true,
			wantErr: nil,
		},
		{
			name:    "存在しないユーザーID",
			id:      "nonexistent",
			want:    false,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.ExistsByID(ctx, tt.id)
			if err != tt.wantErr {
				t.Errorf("ExistsByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExistsByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserRepository_FindAll(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository()

	// テスト用ユーザーを複数作成
	for i := 1; i <= 5; i++ {
		user := createTestUser(
			"user"+string(rune('0'+i)),
			"user"+string(rune('0'+i)),
			"user"+string(rune('0'+i))+"@example.com",
		)
		repo.Create(ctx, user)
	}

	tests := []struct {
		name      string
		offset    int
		limit     int
		wantCount int
		wantErr   error
	}{
		{
			name:      "全ユーザー取得",
			offset:    0,
			limit:     10,
			wantCount: 5,
			wantErr:   nil,
		},
		{
			name:      "ページネーション（最初の3件）",
			offset:    0,
			limit:     3,
			wantCount: 3,
			wantErr:   nil,
		},
		{
			name:      "ページネーション（2件目から2件）",
			offset:    2,
			limit:     2,
			wantCount: 2,
			wantErr:   nil,
		},
		{
			name:      "offsetが総数を超える場合",
			offset:    10,
			limit:     5,
			wantCount: 0,
			wantErr:   nil,
		},
		{
			name:      "負のoffset",
			offset:    -1,
			limit:     5,
			wantCount: 0,
			wantErr:   repository.ErrInvalidArgument,
		},
		{
			name:      "負のlimit",
			offset:    0,
			limit:     -1,
			wantCount: 0,
			wantErr:   repository.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindAll(ctx, tt.offset, tt.limit)
			if err != tt.wantErr {
				t.Errorf("FindAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(got) != tt.wantCount {
				t.Errorf("FindAll() returned %d users, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestUserRepository_Count(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository()

	// 最初は0件
	count, err := repo.Count(ctx)
	if err != nil {
		t.Errorf("Count() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Count() = %d, want 0", count)
	}

	// ユーザーを追加
	for i := 1; i <= 3; i++ {
		user := createTestUser(
			"user"+string(rune('0'+i)),
			"user"+string(rune('0'+i)),
			"user"+string(rune('0'+i))+"@example.com",
		)
		repo.Create(ctx, user)
	}

	// 3件になっているか確認
	count, err = repo.Count(ctx)
	if err != nil {
		t.Errorf("Count() error = %v", err)
	}
	if count != 3 {
		t.Errorf("Count() = %d, want 3", count)
	}

	// 1件削除
	repo.Delete(ctx, "user1")

	// 2件になっているか確認
	count, err = repo.Count(ctx)
	if err != nil {
		t.Errorf("Count() error = %v", err)
	}
	if count != 2 {
		t.Errorf("Count() = %d, want 2", count)
	}
}

func TestUserRepository_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository()

	// 並行でユーザーを作成
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			user := createTestUser(
				"user"+string(rune('0'+idx)),
				"user"+string(rune('0'+idx)),
				"user"+string(rune('0'+idx))+"@example.com",
			)

			// Create
			err := repo.Create(ctx, user)
			if err != nil && err != repository.ErrAlreadyExists {
				t.Errorf("Concurrent Create() error = %v", err)
			}

			// Read
			_, err = repo.FindByID(ctx, user.ID)
			if err != nil && err != repository.ErrNotFound {
				t.Errorf("Concurrent FindByID() error = %v", err)
			}

			// Update
			user.Username = "updated" + user.Username
			err = repo.Update(ctx, user)
			if err != nil && err != repository.ErrNotFound && err != repository.ErrAlreadyExists {
				t.Errorf("Concurrent Update() error = %v", err)
			}
		}(i)
	}

	wg.Wait()

	// 最終的なユーザー数を確認
	count, err := repo.Count(ctx)
	if err != nil {
		t.Errorf("Count() after concurrent access error = %v", err)
	}
	if count > numGoroutines {
		t.Errorf("Count() = %d, want <= %d", count, numGoroutines)
	}
}

func TestUserRepository_DataIsolation(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository()

	// ユーザーを作成
	original := createTestUser("user1", "original", "original@example.com")
	repo.Create(ctx, original)

	// FindByIDで取得
	retrieved, _ := repo.FindByID(ctx, "user1")

	// 取得したユーザーを変更
	retrieved.Username = "modified"
	retrieved.Email = "modified@example.com"

	// 再度FindByIDで取得
	unchanged, _ := repo.FindByID(ctx, "user1")

	// リポジトリ内のデータが変更されていないことを確認
	if unchanged.Username != "original" {
		t.Errorf("Repository data was modified externally: username = %s, want original", unchanged.Username)
	}
	if unchanged.Email != "original@example.com" {
		t.Errorf("Repository data was modified externally: email = %s, want original@example.com", unchanged.Email)
	}
}

// ヘルパー関数：テスト用ユーザーを作成
func createTestUser(id, username, email string) *entity.User {
	user, reason := entity.NewUser(id, username, email)
	if reason != valueobject.OK() {
		// テスト用なので、エラーの場合は直接構造体を作成
		return &entity.User{
			ID:        id,
			Username:  username,
			Email:     email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}
	return user
}
