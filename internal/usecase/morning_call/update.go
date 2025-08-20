package morning_call

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// UpdateUseCase はモーニングコール更新のユースケース
type UpdateUseCase struct {
	morningCallRepo repository.MorningCallRepository
	userRepo        repository.UserRepository
}

// NewUpdateUseCase は新しいモーニングコール更新ユースケースを作成する
func NewUpdateUseCase(
	morningCallRepo repository.MorningCallRepository,
	userRepo repository.UserRepository,
) *UpdateUseCase {
	return &UpdateUseCase{
		morningCallRepo: morningCallRepo,
		userRepo:        userRepo,
	}
}

// UpdateInput はモーニングコール更新の入力データ
type UpdateInput struct {
	ID            string
	SenderID      string // 更新権限確認用
	ScheduledTime *time.Time
	Message       *string
}

// UpdateOutput はモーニングコール更新の出力データ
type UpdateOutput struct {
	MorningCall *entity.MorningCall
}

// Execute はモーニングコールを更新する
func (uc *UpdateUseCase) Execute(ctx context.Context, input UpdateInput) (*UpdateOutput, error) {
	// 入力値の基本検証
	if input.ID == "" {
		return nil, fmt.Errorf("モーニングコールIDは必須です")
	}
	if input.SenderID == "" {
		return nil, fmt.Errorf("送信者IDは必須です")
	}
	if input.ScheduledTime == nil && input.Message == nil {
		return nil, fmt.Errorf("更新する項目を指定してください")
	}

	// モーニングコールの存在確認
	morningCall, err := uc.morningCallRepo.FindByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("モーニングコールが見つかりません")
		}
		return nil, fmt.Errorf("モーニングコールの取得中にエラーが発生しました: %w", err)
	}

	// 送信者の確認（送信者のみが更新可能）
	if morningCall.SenderID != input.SenderID {
		return nil, fmt.Errorf("送信者のみがモーニングコールを更新できます")
	}

	// ステータスの確認（スケジュール済みのもののみ更新可能）
	if morningCall.Status != valueobject.MorningCallStatusScheduled {
		return nil, fmt.Errorf("スケジュール済みのモーニングコールのみ更新できます")
	}

	// 時刻の更新
	if input.ScheduledTime != nil {
		// まず時刻の妥当性を検証（ドメインロジックでの検証）
		oldTime := morningCall.ScheduledTime
		morningCall.ScheduledTime = *input.ScheduledTime
		if reason := morningCall.ValidateScheduledTime(); reason != "" {
			morningCall.ScheduledTime = oldTime // ロールバック
			return nil, fmt.Errorf("%s", reason)
		}
		morningCall.ScheduledTime = oldTime // 一旦元に戻す

		// 時刻の重複チェック（自身を除く）
		activeCalls, err := uc.morningCallRepo.FindActiveByUserPair(ctx, morningCall.SenderID, morningCall.ReceiverID)
		if err != nil {
			return nil, fmt.Errorf("既存のモーニングコール確認中にエラーが発生しました: %w", err)
		}

		for _, call := range activeCalls {
			// 自身は除外
			if call.ID == morningCall.ID {
				continue
			}
			// 時刻が1分以内の場合は重複とみなす
			timeDiff := call.ScheduledTime.Sub(*input.ScheduledTime)
			if timeDiff < 0 {
				timeDiff = -timeDiff
			}
			if timeDiff < time.Minute {
				return nil, fmt.Errorf("同じ時刻付近に既にモーニングコールが設定されています")
			}
		}

		// 時刻を更新
		if reason := morningCall.UpdateScheduledTime(*input.ScheduledTime); reason != "" {
			return nil, fmt.Errorf("時刻の更新に失敗しました: %s", reason)
		}
	}

	// メッセージの更新
	if input.Message != nil {
		if reason := morningCall.UpdateMessage(*input.Message); reason != "" {
			return nil, fmt.Errorf("メッセージの更新に失敗しました: %s", reason)
		}
	}

	// リポジトリで更新
	if err := uc.morningCallRepo.Update(ctx, morningCall); err != nil {
		return nil, fmt.Errorf("モーニングコールの更新に失敗しました: %w", err)
	}

	return &UpdateOutput{
		MorningCall: morningCall,
	}, nil
}
