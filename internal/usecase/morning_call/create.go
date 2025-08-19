package morning_call

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

// CreateUseCase はモーニングコール作成のユースケース
type CreateUseCase struct {
	morningCallRepo  repository.MorningCallRepository
	userRepo         repository.UserRepository
	relationshipRepo repository.RelationshipRepository
}

// NewCreateUseCase は新しいモーニングコール作成ユースケースを作成する
func NewCreateUseCase(
	morningCallRepo repository.MorningCallRepository,
	userRepo repository.UserRepository,
	relationshipRepo repository.RelationshipRepository,
) *CreateUseCase {
	return &CreateUseCase{
		morningCallRepo:  morningCallRepo,
		userRepo:         userRepo,
		relationshipRepo: relationshipRepo,
	}
}

// CreateInput はモーニングコール作成の入力データ
type CreateInput struct {
	SenderID      string
	ReceiverID    string
	ScheduledTime time.Time
	Message       string
}

// CreateOutput はモーニングコール作成の出力データ
type CreateOutput struct {
	MorningCall *entity.MorningCall
}

// Execute はモーニングコールを作成する
func (uc *CreateUseCase) Execute(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	// 入力値の基本検証
	if input.SenderID == "" {
		return nil, fmt.Errorf("送信者IDは必須です")
	}
	if input.ReceiverID == "" {
		return nil, fmt.Errorf("受信者IDは必須です")
	}
	if input.SenderID == input.ReceiverID {
		return nil, fmt.Errorf("自分自身にモーニングコールを設定することはできません")
	}
	if input.ScheduledTime.IsZero() {
		return nil, fmt.Errorf("スケジュール時刻は必須です")
	}

	// 送信者の存在確認
	sender, err := uc.userRepo.FindByID(ctx, input.SenderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("送信者が見つかりません")
		}
		return nil, fmt.Errorf("送信者の確認中にエラーが発生しました: %w", err)
	}

	// 受信者の存在確認
	receiver, err := uc.userRepo.FindByID(ctx, input.ReceiverID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("受信者が見つかりません")
		}
		return nil, fmt.Errorf("受信者の確認中にエラーが発生しました: %w", err)
	}

	// 友達関係の確認
	areFriends, err := uc.relationshipRepo.AreFriends(ctx, input.SenderID, input.ReceiverID)
	if err != nil {
		return nil, fmt.Errorf("友達関係の確認中にエラーが発生しました: %w", err)
	}
	if !areFriends {
		return nil, fmt.Errorf("友達関係にないユーザーにはモーニングコールを設定できません")
	}

	// ブロック状態の確認
	isBlocked, err := uc.relationshipRepo.IsBlocked(ctx, input.SenderID, input.ReceiverID)
	if err != nil {
		return nil, fmt.Errorf("ブロック状態の確認中にエラーが発生しました: %w", err)
	}
	if isBlocked {
		return nil, fmt.Errorf("ブロックされているユーザーにはモーニングコールを設定できません")
	}

	// 同じユーザーペアで既にアクティブなモーニングコールがないか確認
	activeCalls, err := uc.morningCallRepo.FindActiveByUserPair(ctx, input.SenderID, input.ReceiverID)
	if err != nil {
		return nil, fmt.Errorf("既存のモーニングコール確認中にエラーが発生しました: %w", err)
	}

	// 同じ時刻に既にモーニングコールが設定されていないか確認
	for _, call := range activeCalls {
		// 時刻が1分以内の場合は重複とみなす
		timeDiff := call.ScheduledTime.Sub(input.ScheduledTime)
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}
		if timeDiff < time.Minute {
			return nil, fmt.Errorf("同じ時刻付近に既にモーニングコールが設定されています")
		}
	}

	// UUIDを生成
	id, err := utils.GenerateUUID()
	if err != nil {
		return nil, fmt.Errorf("ID生成に失敗しました: %w", err)
	}

	// モーニングコールエンティティを作成
	now := time.Now()
	morningCall := &entity.MorningCall{
		ID:            id,
		SenderID:      sender.ID,
		ReceiverID:    receiver.ID,
		ScheduledTime: input.ScheduledTime,
		Message:       input.Message,
		Status:        valueobject.MorningCallStatusScheduled,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// ドメイン検証
	if reason := morningCall.Validate(); reason != "" {
		return nil, fmt.Errorf("モーニングコールの検証に失敗しました: %s", reason)
	}

	// リポジトリに保存
	if err := uc.morningCallRepo.Create(ctx, morningCall); err != nil {
		return nil, fmt.Errorf("モーニングコールの作成に失敗しました: %w", err)
	}

	return &CreateOutput{
		MorningCall: morningCall,
	}, nil
}
