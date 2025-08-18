package memory

import (
	"context"
	"sync"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
)

// UserRepository はメモリ内でユーザーエンティティを管理するリポジトリ実装
type UserRepository struct {
	// メインストレージ（IDをキーとする）
	users map[string]*entity.User

	// インデックス（高速検索用）
	usernameIndex map[string]string // username -> ID
	emailIndex    map[string]string // email -> ID

	// 並行アクセス制御用
	mu sync.RWMutex
}

// NewUserRepository は新しいメモリ内ユーザーリポジトリを作成する
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:         make(map[string]*entity.User),
		usernameIndex: make(map[string]string),
		emailIndex:    make(map[string]string),
	}
}

// Create は新しいユーザーを作成する
func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	if user == nil {
		return repository.ErrInvalidArgument
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 既存チェック
	if _, exists := r.users[user.ID]; exists {
		return repository.ErrAlreadyExists
	}

	// ユーザー名の重複チェック
	if _, exists := r.usernameIndex[user.Username]; exists {
		return repository.ErrAlreadyExists
	}

	// メールアドレスの重複チェック
	if _, exists := r.emailIndex[user.Email]; exists {
		return repository.ErrAlreadyExists
	}

	// ユーザーのコピーを作成（外部からの変更を防ぐ）
	userCopy := r.copyUser(user)

	// 保存
	r.users[userCopy.ID] = userCopy
	r.usernameIndex[userCopy.Username] = userCopy.ID
	r.emailIndex[userCopy.Email] = userCopy.ID

	return nil
}

// FindByID はIDでユーザーを検索する
func (r *UserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, repository.ErrNotFound
	}

	return r.copyUser(user), nil
}

// FindByUsername はユーザー名でユーザーを検索する
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.usernameIndex[username]
	if !exists {
		return nil, repository.ErrNotFound
	}

	user := r.users[id]
	return r.copyUser(user), nil
}

// FindByEmail はメールアドレスでユーザーを検索する
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.emailIndex[email]
	if !exists {
		return nil, repository.ErrNotFound
	}

	user := r.users[id]
	return r.copyUser(user), nil
}

// Update はユーザー情報を更新する
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	if user == nil {
		return repository.ErrInvalidArgument
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.users[user.ID]
	if !exists {
		return repository.ErrNotFound
	}

	// ユーザー名が変更された場合のインデックス更新
	if existing.Username != user.Username {
		// 新しいユーザー名が既に使用されていないか確認
		if _, exists := r.usernameIndex[user.Username]; exists {
			return repository.ErrAlreadyExists
		}
		// 古いインデックスを削除
		delete(r.usernameIndex, existing.Username)
		// 新しいインデックスを追加
		r.usernameIndex[user.Username] = user.ID
	}

	// メールアドレスが変更された場合のインデックス更新
	if existing.Email != user.Email {
		// 新しいメールアドレスが既に使用されていないか確認
		if _, exists := r.emailIndex[user.Email]; exists {
			return repository.ErrAlreadyExists
		}
		// 古いインデックスを削除
		delete(r.emailIndex, existing.Email)
		// 新しいインデックスを追加
		r.emailIndex[user.Email] = user.ID
	}

	// ユーザー情報を更新
	userCopy := r.copyUser(user)
	userCopy.UpdatedAt = time.Now()
	r.users[user.ID] = userCopy

	return nil
}

// Delete はユーザーを削除する
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return repository.ErrNotFound
	}

	// インデックスから削除
	delete(r.usernameIndex, user.Username)
	delete(r.emailIndex, user.Email)

	// ユーザーを削除
	delete(r.users, id)

	return nil
}

// ExistsByID はIDでユーザーの存在を確認する
func (r *UserRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.users[id]
	return exists, nil
}

// ExistsByUsername はユーザー名でユーザーの存在を確認する
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.usernameIndex[username]
	return exists, nil
}

// ExistsByEmail はメールアドレスでユーザーの存在を確認する
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.emailIndex[email]
	return exists, nil
}

// FindAll はすべてのユーザーを取得する（ページネーション対応）
func (r *UserRepository) FindAll(ctx context.Context, offset, limit int) ([]*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if offset < 0 || limit < 0 {
		return nil, repository.ErrInvalidArgument
	}

	// すべてのユーザーをスライスに変換
	allUsers := make([]*entity.User, 0, len(r.users))
	for _, user := range r.users {
		allUsers = append(allUsers, r.copyUser(user))
	}

	// ページネーション処理
	start := offset
	if start > len(allUsers) {
		return []*entity.User{}, nil
	}

	end := start + limit
	if end > len(allUsers) {
		end = len(allUsers)
	}

	return allUsers[start:end], nil
}

// Count は総ユーザー数を取得する
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.users), nil
}

// copyUser はユーザーエンティティのディープコピーを作成する
func (r *UserRepository) copyUser(user *entity.User) *entity.User {
	return &entity.User{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
