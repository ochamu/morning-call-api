package morning_call

import (
	"context"
	"errors"
	"fmt"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// DeleteUseCase はモーニングコール削除のユースケース
type DeleteUseCase struct {
	morningCallRepo repository.MorningCallRepository
}

// NewDeleteUseCase は新しいモーニングコール削除ユースケースを作成する
func NewDeleteUseCase(
	morningCallRepo repository.MorningCallRepository,
) *DeleteUseCase {
	return &DeleteUseCase{
		morningCallRepo: morningCallRepo,
	}
}

// DeleteInput はモーニングコール削除の入力データ
type DeleteInput struct {
	ID       string
	SenderID string // 削除権限確認用
}

// DeleteOutput はモーニングコール削除の出力データ
type DeleteOutput struct {
	DeletedMorningCall *entity.MorningCall // 削除されたモーニングコールの情報
}

// Execute はモーニングコールを削除する
func (uc *DeleteUseCase) Execute(ctx context.Context, input DeleteInput) (*DeleteOutput, error) {
	// 入力値の基本検証
	if input.ID == "" {
		return nil, fmt.Errorf("モーニングコールIDは必須です")
	}
	if input.SenderID == "" {
		return nil, fmt.Errorf("送信者IDは必須です")
	}

	// モーニングコールの存在確認
	morningCall, err := uc.morningCallRepo.FindByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("モーニングコールが見つかりません")
		}
		return nil, fmt.Errorf("モーニングコールの取得中にエラーが発生しました: %w", err)
	}

	// 送信者の確認（送信者のみが削除可能）
	if morningCall.SenderID != input.SenderID {
		return nil, fmt.Errorf("送信者のみがモーニングコールを削除できます")
	}

	// ステータスの確認（スケジュール済みまたはキャンセル済みのみ削除可能）
	// 配信済みや確認済みのものは履歴として残す必要があるため削除不可
	if morningCall.Status != valueobject.MorningCallStatusScheduled &&
		morningCall.Status != valueobject.MorningCallStatusCancelled {
		return nil, fmt.Errorf("削除できるのはスケジュール済みまたはキャンセル済みのモーニングコールのみです")
	}

	// リポジトリから削除
	if err := uc.morningCallRepo.Delete(ctx, input.ID); err != nil {
		return nil, fmt.Errorf("モーニングコールの削除に失敗しました: %w", err)
	}

	// 削除されたモーニングコールの情報を返す
	return &DeleteOutput{
		DeletedMorningCall: morningCall,
	}, nil
}
