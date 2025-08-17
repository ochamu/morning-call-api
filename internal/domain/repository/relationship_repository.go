package repository

import (
	"context"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// RelationshipRepository は友達関係エンティティの永続化を担うリポジトリインターフェース
type RelationshipRepository interface {
	// Create は新しい友達関係を作成する
	Create(ctx context.Context, relationship *entity.Relationship) error

	// FindByID はIDで友達関係を検索する
	FindByID(ctx context.Context, id string) (*entity.Relationship, error)

	// Update は友達関係情報を更新する
	Update(ctx context.Context, relationship *entity.Relationship) error

	// Delete は友達関係を削除する
	Delete(ctx context.Context, id string) error

	// ExistsByID はIDで友達関係の存在を確認する
	ExistsByID(ctx context.Context, id string) (bool, error)

	// FindByUserPair は特定のユーザーペア間の関係を検索する
	FindByUserPair(ctx context.Context, userID1, userID2 string) (*entity.Relationship, error)

	// FindByRequesterID はリクエスト送信者IDで友達関係を検索する
	FindByRequesterID(ctx context.Context, requesterID string, offset, limit int) ([]*entity.Relationship, error)

	// FindByReceiverID はリクエスト受信者IDで友達関係を検索する
	FindByReceiverID(ctx context.Context, receiverID string, offset, limit int) ([]*entity.Relationship, error)

	// FindByUserID はユーザーIDで友達関係を検索する（送信者・受信者両方）
	FindByUserID(ctx context.Context, userID string, offset, limit int) ([]*entity.Relationship, error)

	// FindByStatus はステータスで友達関係を検索する
	FindByStatus(ctx context.Context, status valueobject.RelationshipStatus, offset, limit int) ([]*entity.Relationship, error)

	// FindFriendsByUserID はユーザーIDで友達（承認済み）関係を検索する
	FindFriendsByUserID(ctx context.Context, userID string, offset, limit int) ([]*entity.Relationship, error)

	// FindPendingRequestsByReceiverID は受信者IDで承認待ちリクエストを検索する
	FindPendingRequestsByReceiverID(ctx context.Context, receiverID string, offset, limit int) ([]*entity.Relationship, error)

	// FindPendingRequestsByRequesterID は送信者IDで承認待ちリクエストを検索する
	FindPendingRequestsByRequesterID(ctx context.Context, requesterID string, offset, limit int) ([]*entity.Relationship, error)

	// FindBlockedRelationshipsByUserID はユーザーIDでブロック関係を検索する
	FindBlockedRelationshipsByUserID(ctx context.Context, userID string, offset, limit int) ([]*entity.Relationship, error)

	// ExistsByUserPair は特定のユーザーペア間の関係の存在を確認する
	ExistsByUserPair(ctx context.Context, userID1, userID2 string) (bool, error)

	// AreFriends は2人のユーザーが友達関係かを確認する
	AreFriends(ctx context.Context, userID1, userID2 string) (bool, error)

	// IsBlocked は指定ユーザーがブロックされているかを確認する
	IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error)

	// CountFriendsByUserID はユーザーIDで友達数を取得する
	CountFriendsByUserID(ctx context.Context, userID string) (int, error)

	// CountPendingRequestsByReceiverID は受信者IDで承認待ちリクエスト数を取得する
	CountPendingRequestsByReceiverID(ctx context.Context, receiverID string) (int, error)

	// CountByStatus はステータスごとの関係数を取得する
	CountByStatus(ctx context.Context, status valueobject.RelationshipStatus) (int, error)

	// FindAll はすべての友達関係を取得する（ページネーション対応）
	FindAll(ctx context.Context, offset, limit int) ([]*entity.Relationship, error)

	// Count は総関係数を取得する
	Count(ctx context.Context) (int, error)
}
