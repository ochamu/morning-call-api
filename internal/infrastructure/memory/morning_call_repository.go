package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// MorningCallRepository はメモリ内でモーニングコールエンティティを管理するリポジトリ実装
type MorningCallRepository struct {
	// メインストレージ（IDをキーとする）
	morningCalls map[string]*entity.MorningCall

	// インデックス（高速検索用）
	senderIndex   map[string][]string                        // senderID -> []morningCallID
	receiverIndex map[string][]string                        // receiverID -> []morningCallID
	statusIndex   map[valueobject.MorningCallStatus][]string // status -> []morningCallID
	userPairIndex map[string][]string                        // "senderID:receiverID" -> []morningCallID

	// 並行アクセス制御用
	mu sync.RWMutex
}

// NewMorningCallRepository は新しいメモリ内モーニングコールリポジトリを作成する
func NewMorningCallRepository() *MorningCallRepository {
	return &MorningCallRepository{
		morningCalls:  make(map[string]*entity.MorningCall),
		senderIndex:   make(map[string][]string),
		receiverIndex: make(map[string][]string),
		statusIndex:   make(map[valueobject.MorningCallStatus][]string),
		userPairIndex: make(map[string][]string),
	}
}

// Create は新しいモーニングコールを作成する
func (r *MorningCallRepository) Create(ctx context.Context, morningCall *entity.MorningCall) error {
	_ = ctx // 将来的なDB実装のために保持
	if morningCall == nil {
		return repository.ErrInvalidArgument
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 既存チェック
	if _, exists := r.morningCalls[morningCall.ID]; exists {
		return repository.ErrAlreadyExists
	}

	// モーニングコールのコピーを作成（外部からの変更を防ぐ）
	mcCopy := r.copyMorningCall(morningCall)

	// 保存
	r.morningCalls[mcCopy.ID] = mcCopy

	// インデックスを更新
	r.addToIndexes(mcCopy)

	return nil
}

// FindByID はIDでモーニングコールを検索する
func (r *MorningCallRepository) FindByID(ctx context.Context, id string) (*entity.MorningCall, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	morningCall, exists := r.morningCalls[id]
	if !exists {
		return nil, repository.ErrNotFound
	}

	return r.copyMorningCall(morningCall), nil
}

// Update はモーニングコール情報を更新する
func (r *MorningCallRepository) Update(ctx context.Context, morningCall *entity.MorningCall) error {
	_ = ctx // 将来的なDB実装のために保持
	if morningCall == nil {
		return repository.ErrInvalidArgument
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.morningCalls[morningCall.ID]
	if !exists {
		return repository.ErrNotFound
	}

	// 既存のインデックスから削除
	r.removeFromIndexes(existing)

	// モーニングコール情報を更新
	mcCopy := r.copyMorningCall(morningCall)
	r.morningCalls[mcCopy.ID] = mcCopy

	// 新しいインデックスに追加
	r.addToIndexes(mcCopy)

	return nil
}

// Delete はモーニングコールを削除する
func (r *MorningCallRepository) Delete(ctx context.Context, id string) error {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.Lock()
	defer r.mu.Unlock()

	morningCall, exists := r.morningCalls[id]
	if !exists {
		return repository.ErrNotFound
	}

	// インデックスから削除
	r.removeFromIndexes(morningCall)

	// モーニングコールを削除
	delete(r.morningCalls, id)

	return nil
}

// ExistsByID はIDでモーニングコールの存在を確認する
func (r *MorningCallRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.morningCalls[id]
	return exists, nil
}

// FindBySenderID は送信者IDでモーニングコールを検索する
func (r *MorningCallRepository) FindBySenderID(ctx context.Context, senderID string, offset, limit int) ([]*entity.MorningCall, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	if offset < 0 || limit < 0 {
		return nil, repository.ErrInvalidArgument
	}

	// limit が 0 の場合は空のスライスを返す
	if limit == 0 {
		return []*entity.MorningCall{}, nil
	}

	// インデックスから該当するIDを取得
	ids, exists := r.senderIndex[senderID]
	if !exists || len(ids) == 0 {
		return []*entity.MorningCall{}, nil
	}

	// モーニングコールを取得してスケジュール時刻でソート
	morningCalls := make([]*entity.MorningCall, 0, len(ids))
	for _, id := range ids {
		if mc, exists := r.morningCalls[id]; exists {
			morningCalls = append(morningCalls, r.copyMorningCall(mc))
		}
	}

	// スケジュール時刻でソート（降順：新しいものが先）
	sort.Slice(morningCalls, func(i, j int) bool {
		return morningCalls[i].ScheduledTime.After(morningCalls[j].ScheduledTime)
	})

	// ページネーション処理
	return r.paginate(morningCalls, offset, limit), nil
}

// FindByReceiverID は受信者IDでモーニングコールを検索する
func (r *MorningCallRepository) FindByReceiverID(ctx context.Context, receiverID string, offset, limit int) ([]*entity.MorningCall, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	if offset < 0 || limit < 0 {
		return nil, repository.ErrInvalidArgument
	}

	// limit が 0 の場合は空のスライスを返す
	if limit == 0 {
		return []*entity.MorningCall{}, nil
	}

	// インデックスから該当するIDを取得
	ids, exists := r.receiverIndex[receiverID]
	if !exists || len(ids) == 0 {
		return []*entity.MorningCall{}, nil
	}

	// モーニングコールを取得してスケジュール時刻でソート
	morningCalls := make([]*entity.MorningCall, 0, len(ids))
	for _, id := range ids {
		if mc, exists := r.morningCalls[id]; exists {
			morningCalls = append(morningCalls, r.copyMorningCall(mc))
		}
	}

	// スケジュール時刻でソート（昇順：直近のものが先）
	sort.Slice(morningCalls, func(i, j int) bool {
		return morningCalls[i].ScheduledTime.Before(morningCalls[j].ScheduledTime)
	})

	// ページネーション処理
	return r.paginate(morningCalls, offset, limit), nil
}

// FindByStatus はステータスでモーニングコールを検索する
func (r *MorningCallRepository) FindByStatus(ctx context.Context, status valueobject.MorningCallStatus, offset, limit int) ([]*entity.MorningCall, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	if offset < 0 || limit < 0 {
		return nil, repository.ErrInvalidArgument
	}

	// limit が 0 の場合は空のスライスを返す
	if limit == 0 {
		return []*entity.MorningCall{}, nil
	}

	// インデックスから該当するIDを取得
	ids, exists := r.statusIndex[status]
	if !exists || len(ids) == 0 {
		return []*entity.MorningCall{}, nil
	}

	// モーニングコールを取得してスケジュール時刻でソート
	morningCalls := make([]*entity.MorningCall, 0, len(ids))
	for _, id := range ids {
		if mc, exists := r.morningCalls[id]; exists {
			morningCalls = append(morningCalls, r.copyMorningCall(mc))
		}
	}

	// スケジュール時刻でソート（昇順：直近のものが先）
	sort.Slice(morningCalls, func(i, j int) bool {
		return morningCalls[i].ScheduledTime.Before(morningCalls[j].ScheduledTime)
	})

	// ページネーション処理
	return r.paginate(morningCalls, offset, limit), nil
}

// FindScheduledBefore は指定時刻より前にスケジュールされたモーニングコールを検索する
func (r *MorningCallRepository) FindScheduledBefore(ctx context.Context, t time.Time, offset, limit int) ([]*entity.MorningCall, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	if offset < 0 || limit < 0 {
		return nil, repository.ErrInvalidArgument
	}

	// limit が 0 の場合は空のスライスを返す
	if limit == 0 {
		return []*entity.MorningCall{}, nil
	}

	// 条件に該当するモーニングコールを収集
	morningCalls := make([]*entity.MorningCall, 0)
	for _, mc := range r.morningCalls {
		if mc.ScheduledTime.Before(t) {
			morningCalls = append(morningCalls, r.copyMorningCall(mc))
		}
	}

	// スケジュール時刻でソート（昇順：直近のものが先）
	sort.Slice(morningCalls, func(i, j int) bool {
		return morningCalls[i].ScheduledTime.Before(morningCalls[j].ScheduledTime)
	})

	// ページネーション処理
	return r.paginate(morningCalls, offset, limit), nil
}

// FindScheduledBetween は指定期間内にスケジュールされたモーニングコールを検索する
func (r *MorningCallRepository) FindScheduledBetween(ctx context.Context, start, end time.Time, offset, limit int) ([]*entity.MorningCall, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	if offset < 0 || limit < 0 {
		return nil, repository.ErrInvalidArgument
	}

	// limit が 0 の場合は空のスライスを返す
	if limit == 0 {
		return []*entity.MorningCall{}, nil
	}

	// start と end の妥当性チェック
	if start.After(end) {
		return nil, repository.ErrInvalidArgument
	}

	// 条件に該当するモーニングコールを収集
	morningCalls := make([]*entity.MorningCall, 0)
	for _, mc := range r.morningCalls {
		if (mc.ScheduledTime.Equal(start) || mc.ScheduledTime.After(start)) &&
			(mc.ScheduledTime.Equal(end) || mc.ScheduledTime.Before(end)) {
			morningCalls = append(morningCalls, r.copyMorningCall(mc))
		}
	}

	// スケジュール時刻でソート（昇順：直近のものが先）
	sort.Slice(morningCalls, func(i, j int) bool {
		return morningCalls[i].ScheduledTime.Before(morningCalls[j].ScheduledTime)
	})

	// ページネーション処理
	return r.paginate(morningCalls, offset, limit), nil
}

// FindActiveByUserPair は特定の送信者から受信者へのアクティブなモーニングコールを検索する
// 注意: モーニングコールには方向性があるため、senderIDとreceiverIDの順序は重要です
func (r *MorningCallRepository) FindActiveByUserPair(ctx context.Context, senderID, receiverID string) ([]*entity.MorningCall, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	// インデックスキーを生成
	pairKey := r.generateUserPairKey(senderID, receiverID)

	// インデックスから該当するIDを取得
	ids, exists := r.userPairIndex[pairKey]
	if !exists || len(ids) == 0 {
		return []*entity.MorningCall{}, nil
	}

	// アクティブなモーニングコールのみを収集
	morningCalls := make([]*entity.MorningCall, 0)
	for _, id := range ids {
		if mc, exists := r.morningCalls[id]; exists {
			// アクティブなステータスの判定（scheduled または delivered）
			if mc.Status == valueobject.MorningCallStatusScheduled ||
				mc.Status == valueobject.MorningCallStatusDelivered {
				morningCalls = append(morningCalls, r.copyMorningCall(mc))
			}
		}
	}

	// スケジュール時刻でソート（昇順：直近のものが先）
	sort.Slice(morningCalls, func(i, j int) bool {
		return morningCalls[i].ScheduledTime.Before(morningCalls[j].ScheduledTime)
	})

	return morningCalls, nil
}

// CountBySenderID は送信者IDでモーニングコール数を取得する
func (r *MorningCallRepository) CountBySenderID(ctx context.Context, senderID string) (int, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids, exists := r.senderIndex[senderID]
	if !exists {
		return 0, nil
	}

	return len(ids), nil
}

// CountByReceiverID は受信者IDでモーニングコール数を取得する
func (r *MorningCallRepository) CountByReceiverID(ctx context.Context, receiverID string) (int, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids, exists := r.receiverIndex[receiverID]
	if !exists {
		return 0, nil
	}

	return len(ids), nil
}

// CountByStatus はステータスごとのモーニングコール数を取得する
func (r *MorningCallRepository) CountByStatus(ctx context.Context, status valueobject.MorningCallStatus) (int, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids, exists := r.statusIndex[status]
	if !exists {
		return 0, nil
	}

	return len(ids), nil
}

// FindAll はすべてのモーニングコールを取得する（ページネーション対応）
func (r *MorningCallRepository) FindAll(ctx context.Context, offset, limit int) ([]*entity.MorningCall, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	if offset < 0 || limit < 0 {
		return nil, repository.ErrInvalidArgument
	}

	// limit が 0 の場合は空のスライスを返す
	if limit == 0 {
		return []*entity.MorningCall{}, nil
	}

	// すべてのモーニングコールをスライスに変換
	morningCalls := make([]*entity.MorningCall, 0, len(r.morningCalls))
	for _, mc := range r.morningCalls {
		morningCalls = append(morningCalls, r.copyMorningCall(mc))
	}

	// IDでソートして一貫した順序を保証
	sort.Slice(morningCalls, func(i, j int) bool {
		return morningCalls[i].ID < morningCalls[j].ID
	})

	// ページネーション処理
	return r.paginate(morningCalls, offset, limit), nil
}

// Count は総モーニングコール数を取得する
func (r *MorningCallRepository) Count(ctx context.Context) (int, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.morningCalls), nil
}

// copyMorningCall はモーニングコールエンティティのディープコピーを作成する
func (r *MorningCallRepository) copyMorningCall(mc *entity.MorningCall) *entity.MorningCall {
	return &entity.MorningCall{
		ID:            mc.ID,
		SenderID:      mc.SenderID,
		ReceiverID:    mc.ReceiverID,
		ScheduledTime: mc.ScheduledTime,
		Message:       mc.Message,
		Status:        mc.Status,
		CreatedAt:     mc.CreatedAt,
		UpdatedAt:     mc.UpdatedAt,
	}
}

// addToIndexes はモーニングコールを各インデックスに追加する
func (r *MorningCallRepository) addToIndexes(mc *entity.MorningCall) {
	// 送信者インデックス
	r.senderIndex[mc.SenderID] = append(r.senderIndex[mc.SenderID], mc.ID)

	// 受信者インデックス
	r.receiverIndex[mc.ReceiverID] = append(r.receiverIndex[mc.ReceiverID], mc.ID)

	// ステータスインデックス
	r.statusIndex[mc.Status] = append(r.statusIndex[mc.Status], mc.ID)

	// ユーザーペアインデックス
	pairKey := r.generateUserPairKey(mc.SenderID, mc.ReceiverID)
	r.userPairIndex[pairKey] = append(r.userPairIndex[pairKey], mc.ID)
}

// removeFromIndexes はモーニングコールを各インデックスから削除する
func (r *MorningCallRepository) removeFromIndexes(mc *entity.MorningCall) {
	// 送信者インデックスから削除
	r.senderIndex[mc.SenderID] = r.removeIDFromSlice(r.senderIndex[mc.SenderID], mc.ID)
	if len(r.senderIndex[mc.SenderID]) == 0 {
		delete(r.senderIndex, mc.SenderID)
	}

	// 受信者インデックスから削除
	r.receiverIndex[mc.ReceiverID] = r.removeIDFromSlice(r.receiverIndex[mc.ReceiverID], mc.ID)
	if len(r.receiverIndex[mc.ReceiverID]) == 0 {
		delete(r.receiverIndex, mc.ReceiverID)
	}

	// ステータスインデックスから削除
	r.statusIndex[mc.Status] = r.removeIDFromSlice(r.statusIndex[mc.Status], mc.ID)
	if len(r.statusIndex[mc.Status]) == 0 {
		delete(r.statusIndex, mc.Status)
	}

	// ユーザーペアインデックスから削除
	pairKey := r.generateUserPairKey(mc.SenderID, mc.ReceiverID)
	r.userPairIndex[pairKey] = r.removeIDFromSlice(r.userPairIndex[pairKey], mc.ID)
	if len(r.userPairIndex[pairKey]) == 0 {
		delete(r.userPairIndex, pairKey)
	}
}

// removeIDFromSlice はスライスから指定されたIDを削除する
func (r *MorningCallRepository) removeIDFromSlice(slice []string, id string) []string {
	result := make([]string, 0, len(slice))
	for _, v := range slice {
		if v != id {
			result = append(result, v)
		}
	}
	return result
}

// generateUserPairKey はユーザーペアのインデックスキーを生成する
// 注意: モーニングコールには送信者から受信者への方向性があるため、
// 引数の順序を保持します（正規化しません）
func (r *MorningCallRepository) generateUserPairKey(senderID, receiverID string) string {
	return senderID + ":" + receiverID
}

// paginate はモーニングコールのスライスにページネーションを適用する
func (r *MorningCallRepository) paginate(morningCalls []*entity.MorningCall, offset, limit int) []*entity.MorningCall {
	start := offset
	if start >= len(morningCalls) {
		return []*entity.MorningCall{}
	}

	end := start + limit
	if end > len(morningCalls) {
		end = len(morningCalls)
	}

	return morningCalls[start:end]
}
