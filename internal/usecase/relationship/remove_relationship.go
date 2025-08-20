package relationship

import (
	"context"
	"errors"
	"fmt"

	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// RemoveRelationshipUseCase は関係削除のユースケース
type RemoveRelationshipUseCase struct {
	relationshipRepo repository.RelationshipRepository
	userRepo         repository.UserRepository
}

// NewRemoveRelationshipUseCase は新しい関係削除ユースケースを作成する
func NewRemoveRelationshipUseCase(
	relationshipRepo repository.RelationshipRepository,
	userRepo repository.UserRepository,
) *RemoveRelationshipUseCase {
	return &RemoveRelationshipUseCase{
		relationshipRepo: relationshipRepo,
		userRepo:         userRepo,
	}
}

// RemoveRelationshipInput は関係削除の入力データ
type RemoveRelationshipInput struct {
	RelationshipID string // 削除する関係ID
	UserID         string // 削除を実行するユーザーID
}

// RemoveRelationshipOutput は関係削除の出力データ
type RemoveRelationshipOutput struct {
	Success bool
	Message string
}

// Execute は関係を削除する
func (uc *RemoveRelationshipUseCase) Execute(ctx context.Context, input RemoveRelationshipInput) (*RemoveRelationshipOutput, error) {
	// 入力値の基本検証
	if input.RelationshipID == "" {
		return nil, fmt.Errorf("関係IDは必須です")
	}
	if input.UserID == "" {
		return nil, fmt.Errorf("ユーザーIDは必須です")
	}

	// 削除実行者の存在確認
	user, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("ユーザーが見つかりません")
		}
		return nil, fmt.Errorf("ユーザーの確認中にエラーが発生しました: %w", err)
	}

	// 関係の取得
	relationship, err := uc.relationshipRepo.FindByID(ctx, input.RelationshipID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("関係が見つかりません")
		}
		return nil, fmt.Errorf("関係の取得中にエラーが発生しました: %w", err)
	}

	// 削除権限の確認
	// 関係に含まれるユーザーのみが削除可能
	if !relationship.InvolvesUser(user.ID) {
		return nil, fmt.Errorf("この関係を削除する権限がありません")
	}

	// ステータスに基づく削除可否の判定
	switch relationship.Status {
	case valueobject.RelationshipStatusBlocked:
		// ブロック関係は削除不可
		// ブロック解除は別のユースケース（UnblockUser）で処理すべき
		return nil, fmt.Errorf("ブロック関係は削除できません。先にブロックを解除してください")
	case valueobject.RelationshipStatusAccepted:
		// 友達関係の解除
		// 両者が削除可能
		if !relationship.InvolvesUser(user.ID) {
			return nil, fmt.Errorf("友達関係の解除は当事者のみが実行できます")
		}
		// 削除処理を続行
	case valueobject.RelationshipStatusPending:
		// ペンディング中のリクエスト
		// 送信者は取り下げ、受信者は削除（拒否とは異なる）として処理
		if !relationship.InvolvesUser(user.ID) {
			return nil, fmt.Errorf("友達リクエストの削除は当事者のみが実行できます")
		}
		// 削除処理を続行
	case valueobject.RelationshipStatusRejected:
		// 拒否済みのリクエスト
		// 両者が削除可能（履歴のクリーンアップ）
		if !relationship.InvolvesUser(user.ID) {
			return nil, fmt.Errorf("拒否済みリクエストの削除は当事者のみが実行できます")
		}
		// 削除処理を続行
	default:
		return nil, fmt.Errorf("不正なステータスの関係です")
	}

	// 関係の相手ユーザーの存在確認（データ整合性のため）
	otherUserID := relationship.GetOtherUserID(user.ID)
	if otherUserID == "" {
		return nil, fmt.Errorf("関係の相手ユーザーが特定できません")
	}

	otherUser, err := uc.userRepo.FindByID(ctx, otherUserID)
	if err != nil {
		// 相手ユーザーが存在しない場合でも削除は許可する（データクリーンアップのため）
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("相手ユーザーの確認中にエラーが発生しました: %w", err)
		}
		// 相手ユーザーが削除されている場合のログ用
		_ = otherUser
	}

	// リポジトリから削除
	if err := uc.relationshipRepo.Delete(ctx, relationship.ID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("削除対象の関係が見つかりません")
		}
		return nil, fmt.Errorf("関係の削除に失敗しました: %w", err)
	}

	// 削除成功メッセージの生成
	var message string
	switch relationship.Status {
	case valueobject.RelationshipStatusAccepted:
		message = "友達関係を解除しました"
	case valueobject.RelationshipStatusPending:
		if relationship.IsRequester(user.ID) {
			message = "友達リクエストを取り下げました"
		} else {
			message = "友達リクエストを削除しました"
		}
	case valueobject.RelationshipStatusRejected:
		message = "拒否済みの友達リクエストを削除しました"
	default:
		message = "関係を削除しました"
	}

	// ログ出力（システムイベント）
	// 実際の実装では、ここで関連するモーニングコールの削除や
	// 通知サービスの呼び出しを行うことも考えられる
	_ = user      // 削除実行者のログ用（将来の拡張用）
	_ = otherUser // 相手ユーザーのログ用（将来の拡張用）

	return &RemoveRelationshipOutput{
		Success: true,
		Message: message,
	}, nil
}
