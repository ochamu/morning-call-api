package repository

import (
	"context"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// MorningCallRepository はモーニングコールエンティティの永続化を担うリポジトリインターフェース
type MorningCallRepository interface {
	// Create は新しいモーニングコールを作成する
	Create(ctx context.Context, morningCall *entity.MorningCall) error

	// FindByID はIDでモーニングコールを検索する
	FindByID(ctx context.Context, id string) (*entity.MorningCall, error)

	// Update はモーニングコール情報を更新する
	Update(ctx context.Context, morningCall *entity.MorningCall) error

	// Delete はモーニングコールを削除する
	Delete(ctx context.Context, id string) error

	// ExistsByID はIDでモーニングコールの存在を確認する
	ExistsByID(ctx context.Context, id string) (bool, error)

	// FindBySenderID は送信者IDでモーニングコールを検索する
	FindBySenderID(ctx context.Context, senderID string, offset, limit int) ([]*entity.MorningCall, error)

	// FindByReceiverID は受信者IDでモーニングコールを検索する
	FindByReceiverID(ctx context.Context, receiverID string, offset, limit int) ([]*entity.MorningCall, error)

	// FindByStatus はステータスでモーニングコールを検索する
	FindByStatus(ctx context.Context, status valueobject.MorningCallStatus, offset, limit int) ([]*entity.MorningCall, error)

	// FindScheduledBefore は指定時刻より前にスケジュールされたモーニングコールを検索する
	FindScheduledBefore(ctx context.Context, time time.Time, offset, limit int) ([]*entity.MorningCall, error)

	// FindScheduledBetween は指定期間内にスケジュールされたモーニングコールを検索する
	FindScheduledBetween(ctx context.Context, start, end time.Time, offset, limit int) ([]*entity.MorningCall, error)

	// FindActiveByUserPair は特定のユーザーペア間のアクティブなモーニングコールを検索する
	FindActiveByUserPair(ctx context.Context, senderID, receiverID string) ([]*entity.MorningCall, error)

	// CountBySenderID は送信者IDでモーニングコール数を取得する
	CountBySenderID(ctx context.Context, senderID string) (int, error)

	// CountByReceiverID は受信者IDでモーニングコール数を取得する
	CountByReceiverID(ctx context.Context, receiverID string) (int, error)

	// CountByStatus はステータスごとのモーニングコール数を取得する
	CountByStatus(ctx context.Context, status valueobject.MorningCallStatus) (int, error)

	// FindAll はすべてのモーニングコールを取得する（ページネーション対応）
	FindAll(ctx context.Context, offset, limit int) ([]*entity.MorningCall, error)

	// Count は総モーニングコール数を取得する
	Count(ctx context.Context) (int, error)
}
