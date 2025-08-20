package morning_call

import (
	"context"
	"fmt"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// ListUseCase はモーニングコール一覧取得のユースケース
type ListUseCase struct {
	morningCallRepo repository.MorningCallRepository
	userRepo        repository.UserRepository
}

// NewListUseCase は新しいモーニングコール一覧取得ユースケースを作成する
func NewListUseCase(
	morningCallRepo repository.MorningCallRepository,
	userRepo repository.UserRepository,
) *ListUseCase {
	return &ListUseCase{
		morningCallRepo: morningCallRepo,
		userRepo:        userRepo,
	}
}

// ListInput はモーニングコール一覧取得の入力データ
type ListInput struct {
	UserID    string                         // 必須：リクエストユーザーのID
	ListType  ListType                       // 必須：一覧の種類（送信/受信）
	Status    *valueobject.MorningCallStatus // オプション：ステータスでフィルタ
	StartTime *time.Time                     // オプション：開始時刻でフィルタ
	EndTime   *time.Time                     // オプション：終了時刻でフィルタ
	Offset    int                            // ページネーション：開始位置
	Limit     int                            // ページネーション：取得件数
}

// ListType は一覧の種類を表す
type ListType string

const (
	ListTypeSent     ListType = "sent"     // 送信したモーニングコール
	ListTypeReceived ListType = "received" // 受信したモーニングコール
)

// ListOutput はモーニングコール一覧取得の出力データ
type ListOutput struct {
	MorningCalls []*entity.MorningCall
	TotalCount   int  // フィルタ適用後の総件数
	HasNext      bool // 次のページがあるか
}

// Execute はモーニングコール一覧を取得する
func (uc *ListUseCase) Execute(ctx context.Context, input ListInput) (*ListOutput, error) {
	// 入力値の基本検証
	if input.UserID == "" {
		return nil, fmt.Errorf("ユーザーIDは必須です")
	}
	if input.ListType != ListTypeSent && input.ListType != ListTypeReceived {
		return nil, fmt.Errorf("一覧タイプは'sent'または'received'を指定してください")
	}
	if input.Limit <= 0 {
		input.Limit = 20 // デフォルト値
	}
	if input.Limit > 100 {
		input.Limit = 100 // 最大値制限
	}
	if input.Offset < 0 {
		input.Offset = 0
	}

	// ユーザーの存在確認
	_, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, fmt.Errorf("ユーザーが見つかりません")
		}
		return nil, fmt.Errorf("ユーザーの確認中にエラーが発生しました: %w", err)
	}

	var morningCalls []*entity.MorningCall
	var totalCount int

	// リストタイプに応じて取得
	if input.ListType == ListTypeSent {
		// 送信したモーニングコール一覧
		morningCalls, totalCount, err = uc.listSentCalls(ctx, input)
	} else {
		// 受信したモーニングコール一覧
		morningCalls, totalCount, err = uc.listReceivedCalls(ctx, input)
	}

	if err != nil {
		return nil, err
	}

	// 次のページがあるか判定
	hasNext := (input.Offset + len(morningCalls)) < totalCount

	return &ListOutput{
		MorningCalls: morningCalls,
		TotalCount:   totalCount,
		HasNext:      hasNext,
	}, nil
}

// listSentCalls は送信したモーニングコール一覧を取得する
func (uc *ListUseCase) listSentCalls(ctx context.Context, input ListInput) ([]*entity.MorningCall, int, error) {
	// 期間フィルタがある場合
	if input.StartTime != nil && input.EndTime != nil {
		// 期間の妥当性チェック
		if input.StartTime.After(*input.EndTime) {
			return nil, 0, fmt.Errorf("開始時刻は終了時刻より前である必要があります")
		}

		// 期間内のモーニングコールを取得
		allCalls, err := uc.morningCallRepo.FindScheduledBetween(ctx, *input.StartTime, *input.EndTime, 0, 10000)
		if err != nil {
			return nil, 0, fmt.Errorf("モーニングコールの取得中にエラーが発生しました: %w", err)
		}

		// 送信者でフィルタリング
		var filteredCalls []*entity.MorningCall
		for _, call := range allCalls {
			if call.SenderID == input.UserID {
				// ステータスフィルタの適用
				if input.Status == nil || call.Status == *input.Status {
					filteredCalls = append(filteredCalls, call)
				}
			}
		}

		// ページネーション適用
		totalCount := len(filteredCalls)
		start := input.Offset
		end := input.Offset + input.Limit
		if start > totalCount {
			return []*entity.MorningCall{}, totalCount, nil
		}
		if end > totalCount {
			end = totalCount
		}

		return filteredCalls[start:end], totalCount, nil
	}

	// 期間フィルタがない場合は通常の取得
	morningCalls, err := uc.morningCallRepo.FindBySenderID(ctx, input.UserID, input.Offset, input.Limit)
	if err != nil {
		return nil, 0, fmt.Errorf("送信モーニングコールの取得中にエラーが発生しました: %w", err)
	}

	// ステータスフィルタの適用
	if input.Status != nil {
		var filteredCalls []*entity.MorningCall
		for _, call := range morningCalls {
			if call.Status == *input.Status {
				filteredCalls = append(filteredCalls, call)
			}
		}
		morningCalls = filteredCalls
	}

	// 総件数を取得
	totalCount, err := uc.morningCallRepo.CountBySenderID(ctx, input.UserID)
	if err != nil {
		return nil, 0, fmt.Errorf("送信モーニングコール数の取得中にエラーが発生しました: %w", err)
	}

	return morningCalls, totalCount, nil
}

// listReceivedCalls は受信したモーニングコール一覧を取得する
func (uc *ListUseCase) listReceivedCalls(ctx context.Context, input ListInput) ([]*entity.MorningCall, int, error) {
	// 期間フィルタがある場合
	if input.StartTime != nil && input.EndTime != nil {
		// 期間の妥当性チェック
		if input.StartTime.After(*input.EndTime) {
			return nil, 0, fmt.Errorf("開始時刻は終了時刻より前である必要があります")
		}

		// 期間内のモーニングコールを取得
		allCalls, err := uc.morningCallRepo.FindScheduledBetween(ctx, *input.StartTime, *input.EndTime, 0, 10000)
		if err != nil {
			return nil, 0, fmt.Errorf("モーニングコールの取得中にエラーが発生しました: %w", err)
		}

		// 受信者でフィルタリング
		var filteredCalls []*entity.MorningCall
		for _, call := range allCalls {
			if call.ReceiverID == input.UserID {
				// ステータスフィルタの適用
				if input.Status == nil || call.Status == *input.Status {
					filteredCalls = append(filteredCalls, call)
				}
			}
		}

		// ページネーション適用
		totalCount := len(filteredCalls)
		start := input.Offset
		end := input.Offset + input.Limit
		if start > totalCount {
			return []*entity.MorningCall{}, totalCount, nil
		}
		if end > totalCount {
			end = totalCount
		}

		return filteredCalls[start:end], totalCount, nil
	}

	// 期間フィルタがない場合は通常の取得
	morningCalls, err := uc.morningCallRepo.FindByReceiverID(ctx, input.UserID, input.Offset, input.Limit)
	if err != nil {
		return nil, 0, fmt.Errorf("受信モーニングコールの取得中にエラーが発生しました: %w", err)
	}

	// ステータスフィルタの適用
	if input.Status != nil {
		var filteredCalls []*entity.MorningCall
		for _, call := range morningCalls {
			if call.Status == *input.Status {
				filteredCalls = append(filteredCalls, call)
			}
		}
		morningCalls = filteredCalls
	}

	// 総件数を取得
	totalCount, err := uc.morningCallRepo.CountByReceiverID(ctx, input.UserID)
	if err != nil {
		return nil, 0, fmt.Errorf("受信モーニングコール数の取得中にエラーが発生しました: %w", err)
	}

	return morningCalls, totalCount, nil
}
