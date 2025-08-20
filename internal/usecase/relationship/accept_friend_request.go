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

// AcceptFriendRequestUseCase は友達リクエスト承認のユースケース
type AcceptFriendRequestUseCase struct {
	relationshipRepo repository.RelationshipRepository
	userRepo         repository.UserRepository
}

// NewAcceptFriendRequestUseCase は新しい友達リクエスト承認ユースケースを作成する
func NewAcceptFriendRequestUseCase(
	relationshipRepo repository.RelationshipRepository,
	userRepo repository.UserRepository,
) *AcceptFriendRequestUseCase {
	return &AcceptFriendRequestUseCase{
		relationshipRepo: relationshipRepo,
		userRepo:         userRepo,
	}
}

// AcceptFriendRequestInput は友達リクエスト承認の入力データ
type AcceptFriendRequestInput struct {
	RelationshipID string // 承認する関係ID
	ReceiverID     string // リクエスト受信者のユーザーID（承認者）
}

// AcceptFriendRequestOutput は友達リクエスト承認の出力データ
type AcceptFriendRequestOutput struct {
	Relationship *entity.Relationship
}

// Execute は友達リクエストを承認する
func (uc *AcceptFriendRequestUseCase) Execute(ctx context.Context, input AcceptFriendRequestInput) (*AcceptFriendRequestOutput, error) {
	// 入力値の基本検証
	if input.RelationshipID == "" {
		return nil, fmt.Errorf("関係IDは必須です")
	}
	if input.ReceiverID == "" {
		return nil, fmt.Errorf("承認者IDは必須です")
	}

	// 承認者の存在確認
	receiver, err := uc.userRepo.FindByID(ctx, input.ReceiverID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("承認者が見つかりません")
		}
		return nil, fmt.Errorf("承認者の確認中にエラーが発生しました: %w", err)
	}

	// 関係の取得
	relationship, err := uc.relationshipRepo.FindByID(ctx, input.RelationshipID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("友達リクエストが見つかりません")
		}
		return nil, fmt.Errorf("友達リクエストの取得中にエラーが発生しました: %w", err)
	}

	// 承認権限の確認（リクエスト受信者のみが承認可能）
	if relationship.ReceiverID != receiver.ID {
		return nil, fmt.Errorf("このリクエストを承認する権限がありません")
	}

	// ステータスの確認
	switch relationship.Status {
	case valueobject.RelationshipStatusAccepted:
		return nil, fmt.Errorf("既に承認済みの友達リクエストです")
	case valueobject.RelationshipStatusRejected:
		return nil, fmt.Errorf("拒否済みの友達リクエストは承認できません")
	case valueobject.RelationshipStatusBlocked:
		return nil, fmt.Errorf("ブロック関係のリクエストは承認できません")
	case valueobject.RelationshipStatusPending:
		// 正常なケース - 承認処理を続行
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

	// 承認処理を実行
	if reason := relationship.Accept(); reason.IsNG() {
		return nil, fmt.Errorf("友達リクエストの承認に失敗しました: %s", reason)
	}

	// 更新日時を設定
	relationship.UpdatedAt = time.Now()

	// リポジトリで更新
	if err := uc.relationshipRepo.Update(ctx, relationship); err != nil {
		return nil, fmt.Errorf("友達リクエストの承認に失敗しました: %w", err)
	}

	// ログ出力（システムイベント）
	// 実際の実装では、ここで通知サービスを呼び出して
	// リクエスト送信者に承認通知を送ることも考えられる
	_ = requester // リクエスト送信者への通知用（将来の拡張用）

	return &AcceptFriendRequestOutput{
		Relationship: relationship,
	}, nil
}
