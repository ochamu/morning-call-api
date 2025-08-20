package relationship

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// RejectFriendRequestUseCase は友達リクエスト拒否のユースケース
type RejectFriendRequestUseCase struct {
	relationshipRepo repository.RelationshipRepository
	userRepo         repository.UserRepository
}

// NewRejectFriendRequestUseCase は新しい友達リクエスト拒否ユースケースを作成する
func NewRejectFriendRequestUseCase(
	relationshipRepo repository.RelationshipRepository,
	userRepo repository.UserRepository,
) *RejectFriendRequestUseCase {
	return &RejectFriendRequestUseCase{
		relationshipRepo: relationshipRepo,
		userRepo:         userRepo,
	}
}

// RejectFriendRequestInput は友達リクエスト拒否の入力データ
type RejectFriendRequestInput struct {
	RelationshipID string // 拒否する関係ID
	ReceiverID     string // リクエスト受信者のユーザーID（拒否者）
}

// RejectFriendRequestOutput は友達リクエスト拒否の出力データ
type RejectFriendRequestOutput struct {
	Relationship *entity.Relationship
}

// Execute は友達リクエストを拒否する
func (uc *RejectFriendRequestUseCase) Execute(ctx context.Context, input RejectFriendRequestInput) (*RejectFriendRequestOutput, error) {
	// 入力値の基本検証
	if input.RelationshipID == "" {
		return nil, fmt.Errorf("関係IDは必須です")
	}
	if input.ReceiverID == "" {
		return nil, fmt.Errorf("拒否者IDは必須です")
	}

	// 拒否者の存在確認
	receiver, err := uc.userRepo.FindByID(ctx, input.ReceiverID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("拒否者が見つかりません")
		}
		return nil, fmt.Errorf("拒否者の確認中にエラーが発生しました: %w", err)
	}

	// 関係の取得
	relationship, err := uc.relationshipRepo.FindByID(ctx, input.RelationshipID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("友達リクエストが見つかりません")
		}
		return nil, fmt.Errorf("友達リクエストの取得中にエラーが発生しました: %w", err)
	}

	// 拒否権限の確認（リクエスト受信者のみが拒否可能）
	if relationship.ReceiverID != receiver.ID {
		return nil, fmt.Errorf("このリクエストを拒否する権限がありません")
	}

	// ステータスの確認
	switch relationship.Status {
	case valueobject.RelationshipStatusAccepted:
		return nil, fmt.Errorf("既に承認済みの友達リクエストは拒否できません")
	case valueobject.RelationshipStatusRejected:
		return nil, fmt.Errorf("既に拒否済みの友達リクエストです")
	case valueobject.RelationshipStatusBlocked:
		return nil, fmt.Errorf("ブロック関係のリクエストは拒否できません")
	case valueobject.RelationshipStatusPending:
		// 正常なケース - 拒否処理を続行
	default:
		return nil, fmt.Errorf("不正なステータスの友達リクエストです")
	}

	// リクエスト送信者の存在確認（データ整合性のため）
	requester, err := uc.userRepo.FindByID(ctx, relationship.RequesterID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("リクエスト送信者が見つかりません")
		}
		return nil, fmt.Errorf("リクエスト送信者の確認中にエラーが発生しました: %w", err)
	}

	// 拒否処理を実行
	if reason := relationship.Reject(); reason.IsNG() {
		return nil, fmt.Errorf("友達リクエストの拒否に失敗しました: %s", reason)
	}

	// 更新日時を設定
	relationship.UpdatedAt = time.Now()

	// リポジトリで更新
	if err := uc.relationshipRepo.Update(ctx, relationship); err != nil {
		return nil, fmt.Errorf("友達リクエストの拒否に失敗しました: %w", err)
	}

	// ログ出力（システムイベント）
	// 実際の実装では、ここで通知サービスを呼び出して
	// リクエスト送信者に拒否通知を送ることも考えられる
	// ただし、拒否の場合は通知しないという選択肢もある
	_ = requester // リクエスト送信者への通知用（将来の拡張用）

	return &RejectFriendRequestOutput{
		Relationship: relationship,
	}, nil
}
