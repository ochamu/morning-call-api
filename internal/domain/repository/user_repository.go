package repository

import (
	"context"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
)

// UserRepository はユーザーエンティティの永続化を担うリポジトリインターフェース
type UserRepository interface {
	// Create は新しいユーザーを作成する
	Create(ctx context.Context, user *entity.User) error

	// FindByID はIDでユーザーを検索する
	FindByID(ctx context.Context, id string) (*entity.User, error)

	// FindByUsername はユーザー名でユーザーを検索する
	FindByUsername(ctx context.Context, username string) (*entity.User, error)

	// FindByEmail はメールアドレスでユーザーを検索する
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	// Update はユーザー情報を更新する
	Update(ctx context.Context, user *entity.User) error

	// Delete はユーザーを削除する
	Delete(ctx context.Context, id string) error

	// ExistsByID はIDでユーザーの存在を確認する
	ExistsByID(ctx context.Context, id string) (bool, error)

	// ExistsByUsername はユーザー名でユーザーの存在を確認する
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// ExistsByEmail はメールアドレスでユーザーの存在を確認する
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// FindAll はすべてのユーザーを取得する（ページネーション対応）
	FindAll(ctx context.Context, offset, limit int) ([]*entity.User, error)

	// Count は総ユーザー数を取得する
	Count(ctx context.Context) (int, error)
}
