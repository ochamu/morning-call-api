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

// BlockRelationshipUseCase は関係IDを使用したブロックのユースケース
type BlockRelationshipUseCase struct {
	relationshipRepo repository.RelationshipRepository
	userRepo         repository.UserRepository
}

// NewBlockRelationshipUseCase は新しい関係ブロックユースケースを作成する
func NewBlockRelationshipUseCase(
	relationshipRepo repository.RelationshipRepository,
	userRepo repository.UserRepository,
) *BlockRelationshipUseCase {
	return &BlockRelationshipUseCase{
		relationshipRepo: relationshipRepo,
		userRepo:         userRepo,
	}
}

// BlockRelationshipInput は関係ブロックの入力データ
type BlockRelationshipInput struct {
	RelationshipID string // ブロック対象の関係ID
	BlockerID      string // ブロックする側のユーザーID
}

// BlockRelationshipOutput は関係ブロックの出力データ
type BlockRelationshipOutput struct {
	Relationship *entity.Relationship
}

// Execute は関係をブロックする
func (uc *BlockRelationshipUseCase) Execute(ctx context.Context, input BlockRelationshipInput) (*BlockRelationshipOutput, error) {
	// 入力値の基本検証
	if input.RelationshipID == "" {
		return nil, fmt.Errorf("関係IDは必須です")
	}
	if input.BlockerID == "" {
		return nil, fmt.Errorf("ブロック実行者IDは必須です")
	}

	// 関係を取得
	relationship, err := uc.relationshipRepo.FindByID(ctx, input.RelationshipID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("関係が見つかりません")
		}
		return nil, fmt.Errorf("関係の取得中にエラーが発生しました: %w", err)
	}

	// ブロック実行者が関係に関与しているか確認
	if relationship.RequesterID != input.BlockerID && relationship.ReceiverID != input.BlockerID {
		return nil, fmt.Errorf("この関係をブロックする権限がありません")
	}

	// 既にブロック済みかどうか確認
	if relationship.Status == valueobject.RelationshipStatusBlocked {
		return nil, fmt.Errorf("既にブロックされています")
	}

	// ブロック処理を実行
	if reason := relationship.Block(); reason.IsNG() {
		return nil, fmt.Errorf("関係のブロックに失敗しました: %s", reason)
	}

	// 更新日時を設定
	relationship.UpdatedAt = time.Now()

	// リポジトリで更新
	if err := uc.relationshipRepo.Update(ctx, relationship); err != nil {
		return nil, fmt.Errorf("関係のブロックに失敗しました: %w", err)
	}

	return &BlockRelationshipOutput{
		Relationship: relationship,
	}, nil
}