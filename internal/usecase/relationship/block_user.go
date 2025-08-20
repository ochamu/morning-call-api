package relationship

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
	"github.com/ochamu/morning-call-api/pkg/utils"
)

// BlockUserUseCase はユーザーブロックのユースケース
type BlockUserUseCase struct {
	relationshipRepo repository.RelationshipRepository
	userRepo         repository.UserRepository
}

// NewBlockUserUseCase は新しいユーザーブロックユースケースを作成する
func NewBlockUserUseCase(
	relationshipRepo repository.RelationshipRepository,
	userRepo repository.UserRepository,
) *BlockUserUseCase {
	return &BlockUserUseCase{
		relationshipRepo: relationshipRepo,
		userRepo:         userRepo,
	}
}

// BlockUserInput はユーザーブロックの入力データ
type BlockUserInput struct {
	BlockerID string // ブロックする側のユーザーID
	BlockedID string // ブロックされる側のユーザーID
}

// BlockUserOutput はユーザーブロックの出力データ
type BlockUserOutput struct {
	Relationship *entity.Relationship
}

// Execute はユーザーをブロックする
func (uc *BlockUserUseCase) Execute(ctx context.Context, input BlockUserInput) (*BlockUserOutput, error) {
	// 入力値の基本検証
	if input.BlockerID == "" {
		return nil, fmt.Errorf("ブロック実行者IDは必須です")
	}
	if input.BlockedID == "" {
		return nil, fmt.Errorf("ブロック対象者IDは必須です")
	}
	if input.BlockerID == input.BlockedID {
		return nil, fmt.Errorf("自分自身をブロックすることはできません")
	}

	// ブロック実行者の存在確認
	blocker, err := uc.userRepo.FindByID(ctx, input.BlockerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("ブロック実行者が見つかりません")
		}
		return nil, fmt.Errorf("ブロック実行者の確認中にエラーが発生しました: %w", err)
	}

	// ブロック対象者の存在確認
	blocked, err := uc.userRepo.FindByID(ctx, input.BlockedID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("ブロック対象者が見つかりません")
		}
		return nil, fmt.Errorf("ブロック対象者の確認中にエラーが発生しました: %w", err)
	}

	// 既存の関係を確認
	existingRelationship, err := uc.relationshipRepo.FindByUserPair(ctx, input.BlockerID, input.BlockedID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("既存の関係確認中にエラーが発生しました: %w", err)
	}

	var relationship *entity.Relationship

	// 既存の関係がある場合
	if existingRelationship != nil {
		// 既にブロック済みかどうか確認
		if existingRelationship.Status == valueobject.RelationshipStatusBlocked {
			// ブロック実行者が同じ場合（自分がRequesterとしてブロック済み）
			if existingRelationship.RequesterID == input.BlockerID {
				return nil, fmt.Errorf("既にこのユーザーをブロックしています")
			}
			// 相手からブロックされている場合
			// メモリリポジトリの制約上、同じユーザーペアで複数の関係を持てないため、
			// 相手からブロックされている場合は新規作成できない
			return nil, fmt.Errorf("相手にブロックされています")
		} else {
			// ブロック以外の状態の場合、既存の関係を更新する
			// ブロック実行者が関係の所有者（RequesterまたはReceiver）である必要がある
			if existingRelationship.RequesterID == input.BlockerID || existingRelationship.ReceiverID == input.BlockerID {
				// ブロック処理を実行
				if reason := existingRelationship.Block(); reason.IsNG() {
					return nil, fmt.Errorf("ユーザーのブロックに失敗しました: %s", reason)
				}

				// 更新日時を設定
				existingRelationship.UpdatedAt = time.Now()

				// リポジトリで更新
				if err := uc.relationshipRepo.Update(ctx, existingRelationship); err != nil {
					return nil, fmt.Errorf("ユーザーのブロックに失敗しました: %w", err)
				}

				relationship = existingRelationship
			} else {
				// 既存の関係があるが、ブロック実行者が関係に関与していない場合
				// この場合はエラーとして処理
				return nil, fmt.Errorf("この関係をブロックする権限がありません")
			}
		}
	}

	// 新規にブロック関係を作成する必要がある場合
	if relationship == nil {
		// UUIDを生成
		id, err := utils.GenerateUUID()
		if err != nil {
			return nil, fmt.Errorf("ID生成に失敗しました: %w", err)
		}

		// ブロック関係エンティティを作成
		var reason valueobject.NGReason
		relationship, reason = entity.NewRelationship(id, blocker.ID, blocked.ID)
		if reason.IsNG() {
			return nil, fmt.Errorf("ブロック関係の作成に失敗しました: %s", reason)
		}

		// nilチェック（念のため）
		if relationship == nil {
			return nil, fmt.Errorf("ブロック関係の作成に失敗しました: エンティティがnilです")
		}

		// 即座にブロック状態に設定
		if reason := relationship.Block(); reason.IsNG() {
			return nil, fmt.Errorf("ブロック関係の設定に失敗しました: %s", reason)
		}

		// リポジトリに保存
		if err := uc.relationshipRepo.Create(ctx, relationship); err != nil {
			// 重複エラーの場合
			if errors.Is(err, repository.ErrAlreadyExists) {
				return nil, fmt.Errorf("既にブロック関係が存在します")
			}
			return nil, fmt.Errorf("ブロック関係の作成に失敗しました: %w", err)
		}
	}

	// ログ出力（システムイベント）
	// 実際の実装では、ここで関連するモーニングコールの削除や
	// 通知サービスの呼び出しを行うことも考えられる
	// ただし、ブロックの場合は通知しないという選択肢もある
	_ = blocker // ブロック実行者のログ用（将来の拡張用）
	_ = blocked // ブロック対象者のログ用（将来の拡張用）

	return &BlockUserOutput{
		Relationship: relationship,
	}, nil
}
