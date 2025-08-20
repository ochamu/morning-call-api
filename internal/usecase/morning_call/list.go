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
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("ユーザーが見つかりません")
		}
		return nil, fmt.Errorf("ユーザーの確認中にエラーが発生しました: %w", err)
	}

	// 共通ロジックでリスト取得
	morningCalls, totalCount, err := uc.listCallsWithFilters(ctx, input)
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

// listCallsWithFilters は共通のフィルタリングロジックでモーニングコール一覧を取得する
func (uc *ListUseCase) listCallsWithFilters(ctx context.Context, input ListInput) ([]*entity.MorningCall, int, error) {
	// 期間フィルタがある場合
	if input.StartTime != nil && input.EndTime != nil {
		return uc.listCallsWithTimeRange(ctx, input)
	}

	// 期間フィルタがない場合
	return uc.listCallsWithoutTimeRange(ctx, input)
}

// listCallsWithTimeRange は期間フィルタを適用してモーニングコール一覧を取得する
func (uc *ListUseCase) listCallsWithTimeRange(ctx context.Context, input ListInput) ([]*entity.MorningCall, int, error) {
	// 期間の妥当性チェック
	if input.StartTime.After(*input.EndTime) {
		return nil, 0, fmt.Errorf("開始時刻は終了時刻より前である必要があります")
	}

	// TODO: 将来的にはリポジトリレベルでユーザーIDフィルタを適用して
	// パフォーマンスを改善する必要がある。現在は暫定的に10,000件の制限を設ける。
	// 期間内のモーニングコールを取得
	allCalls, err := uc.morningCallRepo.FindScheduledBetween(ctx, *input.StartTime, *input.EndTime, 0, 10000)
	if err != nil {
		return nil, 0, fmt.Errorf("モーニングコールの取得中にエラーが発生しました: %w", err)
	}

	// ユーザーIDとステータスでフィルタリング
	filteredCalls := uc.filterCalls(allCalls, input)

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

// listCallsWithoutTimeRange は期間フィルタなしでモーニングコール一覧を取得する
func (uc *ListUseCase) listCallsWithoutTimeRange(ctx context.Context, input ListInput) ([]*entity.MorningCall, int, error) {
	var morningCalls []*entity.MorningCall
	var allCalls []*entity.MorningCall
	var err error

	// ステータスフィルタがある場合は、正確な総件数のため全件取得が必要
	if input.Status != nil {
		// 全件取得してフィルタリング（ページネーションは後で適用）
		if input.ListType == ListTypeSent {
			allCalls, err = uc.morningCallRepo.FindBySenderID(ctx, input.UserID, 0, 10000)
		} else {
			allCalls, err = uc.morningCallRepo.FindByReceiverID(ctx, input.UserID, 0, 10000)
		}
		if err != nil {
			return nil, 0, fmt.Errorf("モーニングコールの取得中にエラーが発生しました: %w", err)
		}

		// ステータスでフィルタリング
		var filteredCalls []*entity.MorningCall
		for _, call := range allCalls {
			if call.Status == *input.Status {
				filteredCalls = append(filteredCalls, call)
			}
		}

		// フィルタ適用後の総件数
		totalCount := len(filteredCalls)

		// ページネーション適用
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

	// ステータスフィルタがない場合は通常のページネーション
	if input.ListType == ListTypeSent {
		morningCalls, err = uc.morningCallRepo.FindBySenderID(ctx, input.UserID, input.Offset, input.Limit)
		if err != nil {
			return nil, 0, fmt.Errorf("送信モーニングコールの取得中にエラーが発生しました: %w", err)
		}
		totalCount, err := uc.morningCallRepo.CountBySenderID(ctx, input.UserID)
		if err != nil {
			return nil, 0, fmt.Errorf("送信モーニングコール数の取得中にエラーが発生しました: %w", err)
		}
		return morningCalls, totalCount, nil
	}

	// 受信リストの場合
	morningCalls, err = uc.morningCallRepo.FindByReceiverID(ctx, input.UserID, input.Offset, input.Limit)
	if err != nil {
		return nil, 0, fmt.Errorf("受信モーニングコールの取得中にエラーが発生しました: %w", err)
	}
	totalCount, err := uc.morningCallRepo.CountByReceiverID(ctx, input.UserID)
	if err != nil {
		return nil, 0, fmt.Errorf("受信モーニングコール数の取得中にエラーが発生しました: %w", err)
	}
	return morningCalls, totalCount, nil
}

// filterCalls はモーニングコールリストにフィルタを適用する
func (uc *ListUseCase) filterCalls(calls []*entity.MorningCall, input ListInput) []*entity.MorningCall {
	var filteredCalls []*entity.MorningCall

	for _, call := range calls {
		// ユーザーIDでフィルタリング
		if input.ListType == ListTypeSent && call.SenderID != input.UserID {
			continue
		}
		if input.ListType == ListTypeReceived && call.ReceiverID != input.UserID {
			continue
		}

		// ステータスでフィルタリング
		if input.Status != nil && call.Status != *input.Status {
			continue
		}

		filteredCalls = append(filteredCalls, call)
	}

	return filteredCalls
}
