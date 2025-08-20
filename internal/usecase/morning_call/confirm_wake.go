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

// ConfirmWakeUseCase は起床確認のユースケース
type ConfirmWakeUseCase struct {
	morningCallRepo repository.MorningCallRepository
	userRepo        repository.UserRepository
}

// NewConfirmWakeUseCase は新しい起床確認ユースケースを作成する
func NewConfirmWakeUseCase(
	morningCallRepo repository.MorningCallRepository,
	userRepo repository.UserRepository,
) *ConfirmWakeUseCase {
	return &ConfirmWakeUseCase{
		morningCallRepo: morningCallRepo,
		userRepo:        userRepo,
	}
}

// ConfirmWakeInput は起床確認の入力データ
type ConfirmWakeInput struct {
	MorningCallID string
	ReceiverID    string // 起床確認をする受信者のID
}

// ConfirmWakeOutput は起床確認の出力データ
type ConfirmWakeOutput struct {
	MorningCall *entity.MorningCall
	ConfirmedAt time.Time
}

// Execute は起床確認を実行する
func (uc *ConfirmWakeUseCase) Execute(ctx context.Context, input ConfirmWakeInput) (*ConfirmWakeOutput, error) {
	// 入力値の基本検証
	if input.MorningCallID == "" {
		return nil, fmt.Errorf("モーニングコールIDは必須です")
	}
	if input.ReceiverID == "" {
		return nil, fmt.Errorf("受信者IDは必須です")
	}

	// 受信者の存在確認
	receiver, err := uc.userRepo.FindByID(ctx, input.ReceiverID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("受信者が見つかりません")
		}
		return nil, fmt.Errorf("受信者の確認中にエラーが発生しました: %w", err)
	}

	// モーニングコールの取得
	morningCall, err := uc.morningCallRepo.FindByID(ctx, input.MorningCallID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("モーニングコールが見つかりません")
		}
		return nil, fmt.Errorf("モーニングコールの取得中にエラーが発生しました: %w", err)
	}

	// 受信者の確認（受信者本人のみ起床確認可能）
	if morningCall.ReceiverID != receiver.ID {
		return nil, fmt.Errorf("受信者のみが起床確認を行えます")
	}

	// ステータスの確認（配信済みのもののみ起床確認可能）
	if morningCall.Status != valueobject.MorningCallStatusDelivered {
		switch morningCall.Status {
		case valueobject.MorningCallStatusScheduled:
			return nil, fmt.Errorf("まだ配信されていないモーニングコールは起床確認できません")
		case valueobject.MorningCallStatusConfirmed:
			return nil, fmt.Errorf("すでに起床確認済みです")
		case valueobject.MorningCallStatusCancelled:
			return nil, fmt.Errorf("キャンセル済みのモーニングコールは起床確認できません")
		case valueobject.MorningCallStatusExpired:
			return nil, fmt.Errorf("期限切れのモーニングコールは起床確認できません")
		default:
			return nil, fmt.Errorf("このステータスのモーニングコールは起床確認できません")
		}
	}

	// 起床確認を記録
	if reason := morningCall.ConfirmWakeUp(); reason.IsNG() {
		return nil, fmt.Errorf("起床確認の記録に失敗しました: %s", string(reason))
	}

	// リポジトリに保存
	if err := uc.morningCallRepo.Update(ctx, morningCall); err != nil {
		return nil, fmt.Errorf("起床確認の保存に失敗しました: %w", err)
	}

	// 結果を返す
	return &ConfirmWakeOutput{
		MorningCall: morningCall,
		ConfirmedAt: morningCall.UpdatedAt,
	}, nil
}
