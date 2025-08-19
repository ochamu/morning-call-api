package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// RelationshipRepository はメモリ内で友達関係エンティティを管理するリポジトリ実装
type RelationshipRepository struct {
	// メインストレージ（IDをキーとする）
	relationships map[string]*entity.Relationship

	// インデックス（高速検索用）
	requesterIndex  map[string][]string                                    // requesterID -> []relationshipID
	receiverIndex   map[string][]string                                    // receiverID -> []relationshipID
	userPairIndex   map[string]string                                      // "userID1:userID2" -> relationshipID（小さいID:大きいID）
	statusIndex     map[valueobject.RelationshipStatus][]string            // status -> []relationshipID
	userStatusIndex map[string]map[valueobject.RelationshipStatus][]string // userID -> status -> []relationshipID

	// 並行アクセス制御用
	mu sync.RWMutex
}

// NewRelationshipRepository は新しいメモリ内友達関係リポジトリを作成する
func NewRelationshipRepository() *RelationshipRepository {
	return &RelationshipRepository{
		relationships:   make(map[string]*entity.Relationship),
		requesterIndex:  make(map[string][]string),
		receiverIndex:   make(map[string][]string),
		userPairIndex:   make(map[string]string),
		statusIndex:     make(map[valueobject.RelationshipStatus][]string),
		userStatusIndex: make(map[string]map[valueobject.RelationshipStatus][]string),
	}
}

// Create は新しい友達関係を作成する
func (r *RelationshipRepository) Create(ctx context.Context, relationship *entity.Relationship) error {
	_ = ctx // 将来的なDB実装のために保持
	if relationship == nil {
		return repository.ErrInvalidArgument
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 既存チェック
	if _, exists := r.relationships[relationship.ID]; exists {
		return repository.ErrAlreadyExists
	}

	// ユーザーペアの既存関係チェック
	pairKey := r.createUserPairKey(relationship.RequesterID, relationship.ReceiverID)
	if _, exists := r.userPairIndex[pairKey]; exists {
		return repository.ErrAlreadyExists
	}

	// 関係のコピーを作成（外部からの変更を防ぐ）
	relationshipCopy := r.copyRelationship(relationship)

	// 保存
	r.relationships[relationshipCopy.ID] = relationshipCopy
	r.addToIndexes(relationshipCopy)

	return nil
}

// FindByID はIDで友達関係を検索する
func (r *RelationshipRepository) FindByID(ctx context.Context, id string) (*entity.Relationship, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	relationship, exists := r.relationships[id]
	if !exists {
		return nil, repository.ErrNotFound
	}

	return r.copyRelationship(relationship), nil
}

// Update は友達関係情報を更新する
func (r *RelationshipRepository) Update(ctx context.Context, relationship *entity.Relationship) error {
	_ = ctx // 将来的なDB実装のために保持
	if relationship == nil {
		return repository.ErrInvalidArgument
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	oldRelationship, exists := r.relationships[relationship.ID]
	if !exists {
		return repository.ErrNotFound
	}

	// ユーザーペアが変更された場合の重複チェック
	if oldRelationship.RequesterID != relationship.RequesterID ||
		oldRelationship.ReceiverID != relationship.ReceiverID {
		newPairKey := r.createUserPairKey(relationship.RequesterID, relationship.ReceiverID)
		if existingID, exists := r.userPairIndex[newPairKey]; exists && existingID != relationship.ID {
			// 別の関係が既に同じユーザーペアを持っている
			return repository.ErrAlreadyExists
		}
	}

	// インデックスから古い情報を削除
	r.removeFromIndexes(oldRelationship)

	// 関係のコピーを作成して保存
	relationshipCopy := r.copyRelationship(relationship)
	r.relationships[relationshipCopy.ID] = relationshipCopy

	// 新しい情報でインデックスを更新
	r.addToIndexes(relationshipCopy)

	return nil
}

// Delete は友達関係を削除する
func (r *RelationshipRepository) Delete(ctx context.Context, id string) error {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.Lock()
	defer r.mu.Unlock()

	relationship, exists := r.relationships[id]
	if !exists {
		return repository.ErrNotFound
	}

	// インデックスから削除
	r.removeFromIndexes(relationship)

	// メインストレージから削除
	delete(r.relationships, id)

	return nil
}

// ExistsByID はIDで友達関係の存在を確認する
func (r *RelationshipRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.relationships[id]
	return exists, nil
}

// FindByUserPair は特定のユーザーペア間の関係を検索する
func (r *RelationshipRepository) FindByUserPair(ctx context.Context, userID1, userID2 string) (*entity.Relationship, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	pairKey := r.createUserPairKey(userID1, userID2)
	relationshipID, exists := r.userPairIndex[pairKey]
	if !exists {
		return nil, repository.ErrNotFound
	}

	relationship := r.relationships[relationshipID]
	return r.copyRelationship(relationship), nil
}

// FindByRequesterID はリクエスト送信者IDで友達関係を検索する
func (r *RelationshipRepository) FindByRequesterID(ctx context.Context, requesterID string, offset, limit int) ([]*entity.Relationship, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	relationshipIDs, exists := r.requesterIndex[requesterID]
	if !exists || len(relationshipIDs) == 0 {
		return []*entity.Relationship{}, nil
	}

	return r.getRelationshipsWithPagination(relationshipIDs, offset, limit)
}

// FindByReceiverID はリクエスト受信者IDで友達関係を検索する
func (r *RelationshipRepository) FindByReceiverID(ctx context.Context, receiverID string, offset, limit int) ([]*entity.Relationship, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	relationshipIDs, exists := r.receiverIndex[receiverID]
	if !exists || len(relationshipIDs) == 0 {
		return []*entity.Relationship{}, nil
	}

	return r.getRelationshipsWithPagination(relationshipIDs, offset, limit)
}

// FindByUserID はユーザーIDで友達関係を検索する（送信者・受信者両方）
func (r *RelationshipRepository) FindByUserID(ctx context.Context, userID string, offset, limit int) ([]*entity.Relationship, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	// リクエスト送信者と受信者の両方から検索
	relationshipIDMap := make(map[string]bool)

	if ids, exists := r.requesterIndex[userID]; exists {
		for _, id := range ids {
			relationshipIDMap[id] = true
		}
	}

	if ids, exists := r.receiverIndex[userID]; exists {
		for _, id := range ids {
			relationshipIDMap[id] = true
		}
	}

	// マップからIDリストを作成
	var relationshipIDs []string
	for id := range relationshipIDMap {
		relationshipIDs = append(relationshipIDs, id)
	}

	if len(relationshipIDs) == 0 {
		return []*entity.Relationship{}, nil
	}

	return r.getRelationshipsWithPagination(relationshipIDs, offset, limit)
}

// FindByStatus はステータスで友達関係を検索する
func (r *RelationshipRepository) FindByStatus(ctx context.Context, status valueobject.RelationshipStatus, offset, limit int) ([]*entity.Relationship, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	relationshipIDs, exists := r.statusIndex[status]
	if !exists || len(relationshipIDs) == 0 {
		return []*entity.Relationship{}, nil
	}

	return r.getRelationshipsWithPagination(relationshipIDs, offset, limit)
}

// FindFriendsByUserID はユーザーIDで友達（承認済み）関係を検索する
func (r *RelationshipRepository) FindFriendsByUserID(ctx context.Context, userID string, offset, limit int) ([]*entity.Relationship, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	userStatusMap, exists := r.userStatusIndex[userID]
	if !exists {
		return []*entity.Relationship{}, nil
	}

	relationshipIDs, exists := userStatusMap[valueobject.RelationshipStatusAccepted]
	if !exists || len(relationshipIDs) == 0 {
		return []*entity.Relationship{}, nil
	}

	return r.getRelationshipsWithPagination(relationshipIDs, offset, limit)
}

// FindPendingRequestsByReceiverID は受信者IDで承認待ちリクエストを検索する
func (r *RelationshipRepository) FindPendingRequestsByReceiverID(ctx context.Context, receiverID string, offset, limit int) ([]*entity.Relationship, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	var pendingRelationships []string

	if ids, exists := r.receiverIndex[receiverID]; exists {
		for _, id := range ids {
			if rel := r.relationships[id]; rel != nil && rel.Status == valueobject.RelationshipStatusPending {
				pendingRelationships = append(pendingRelationships, id)
			}
		}
	}

	if len(pendingRelationships) == 0 {
		return []*entity.Relationship{}, nil
	}

	return r.getRelationshipsWithPagination(pendingRelationships, offset, limit)
}

// FindPendingRequestsByRequesterID は送信者IDで承認待ちリクエストを検索する
func (r *RelationshipRepository) FindPendingRequestsByRequesterID(ctx context.Context, requesterID string, offset, limit int) ([]*entity.Relationship, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	var pendingRelationships []string

	if ids, exists := r.requesterIndex[requesterID]; exists {
		for _, id := range ids {
			if rel := r.relationships[id]; rel != nil && rel.Status == valueobject.RelationshipStatusPending {
				pendingRelationships = append(pendingRelationships, id)
			}
		}
	}

	if len(pendingRelationships) == 0 {
		return []*entity.Relationship{}, nil
	}

	return r.getRelationshipsWithPagination(pendingRelationships, offset, limit)
}

// FindBlockedRelationshipsByUserID はユーザーIDでブロック関係を検索する
func (r *RelationshipRepository) FindBlockedRelationshipsByUserID(ctx context.Context, userID string, offset, limit int) ([]*entity.Relationship, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	userStatusMap, exists := r.userStatusIndex[userID]
	if !exists {
		return []*entity.Relationship{}, nil
	}

	relationshipIDs, exists := userStatusMap[valueobject.RelationshipStatusBlocked]
	if !exists || len(relationshipIDs) == 0 {
		return []*entity.Relationship{}, nil
	}

	return r.getRelationshipsWithPagination(relationshipIDs, offset, limit)
}

// ExistsByUserPair は特定のユーザーペア間の関係の存在を確認する
func (r *RelationshipRepository) ExistsByUserPair(ctx context.Context, userID1, userID2 string) (bool, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	pairKey := r.createUserPairKey(userID1, userID2)
	_, exists := r.userPairIndex[pairKey]
	return exists, nil
}

// AreFriends は2人のユーザーが友達関係かを確認する
func (r *RelationshipRepository) AreFriends(ctx context.Context, userID1, userID2 string) (bool, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	pairKey := r.createUserPairKey(userID1, userID2)
	relationshipID, exists := r.userPairIndex[pairKey]
	if !exists {
		return false, nil
	}

	relationship := r.relationships[relationshipID]
	return relationship.Status == valueobject.RelationshipStatusAccepted, nil
}

// IsBlocked は2人のユーザー間にブロック関係が存在するかを確認する
// 注: 現在のエンティティ設計では誰がブロックしたかは記録されないため、
// 単にブロック関係の存在のみを確認します
func (r *RelationshipRepository) IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	pairKey := r.createUserPairKey(blockerID, blockedID)
	relationshipID, exists := r.userPairIndex[pairKey]
	if !exists {
		return false, nil
	}

	relationship := r.relationships[relationshipID]
	// ブロック関係が存在するかを確認
	// 注: 将来的にブロックした人を特定する必要がある場合は、
	// エンティティにBlockerIDフィールドを追加する必要があります
	return relationship.Status == valueobject.RelationshipStatusBlocked, nil
}

// CountFriendsByUserID はユーザーIDで友達数を取得する
func (r *RelationshipRepository) CountFriendsByUserID(ctx context.Context, userID string) (int, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	userStatusMap, exists := r.userStatusIndex[userID]
	if !exists {
		return 0, nil
	}

	relationshipIDs, exists := userStatusMap[valueobject.RelationshipStatusAccepted]
	if !exists {
		return 0, nil
	}

	return len(relationshipIDs), nil
}

// CountPendingRequestsByReceiverID は受信者IDで承認待ちリクエスト数を取得する
func (r *RelationshipRepository) CountPendingRequestsByReceiverID(ctx context.Context, receiverID string) (int, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	if ids, exists := r.receiverIndex[receiverID]; exists {
		for _, id := range ids {
			if rel := r.relationships[id]; rel != nil && rel.Status == valueobject.RelationshipStatusPending {
				count++
			}
		}
	}

	return count, nil
}

// CountByStatus はステータスごとの関係数を取得する
func (r *RelationshipRepository) CountByStatus(ctx context.Context, status valueobject.RelationshipStatus) (int, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	relationshipIDs, exists := r.statusIndex[status]
	if !exists {
		return 0, nil
	}

	return len(relationshipIDs), nil
}

// FindAll はすべての友達関係を取得する（ページネーション対応）
func (r *RelationshipRepository) FindAll(ctx context.Context, offset, limit int) ([]*entity.Relationship, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	// すべてのIDを収集
	var allIDs []string
	for id := range r.relationships {
		allIDs = append(allIDs, id)
	}

	// ソート（一貫性のため）
	sort.Strings(allIDs)

	return r.getRelationshipsWithPagination(allIDs, offset, limit)
}

// Count は総関係数を取得する
func (r *RelationshipRepository) Count(ctx context.Context) (int, error) {
	_ = ctx // 将来的なDB実装のために保持
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.relationships), nil
}

// ===== ヘルパーメソッド =====

// copyRelationship は友達関係のディープコピーを作成する
func (r *RelationshipRepository) copyRelationship(rel *entity.Relationship) *entity.Relationship {
	if rel == nil {
		return nil
	}
	relCopy := *rel
	return &relCopy
}

// createUserPairKey はユーザーペアのキーを作成する（小さいID:大きいID）
func (r *RelationshipRepository) createUserPairKey(userID1, userID2 string) string {
	if userID1 < userID2 {
		return userID1 + ":" + userID2
	}
	return userID2 + ":" + userID1
}

// addToIndexes は関係をインデックスに追加する
func (r *RelationshipRepository) addToIndexes(rel *entity.Relationship) {
	// RequesterIndexに追加
	r.requesterIndex[rel.RequesterID] = append(r.requesterIndex[rel.RequesterID], rel.ID)

	// ReceiverIndexに追加
	r.receiverIndex[rel.ReceiverID] = append(r.receiverIndex[rel.ReceiverID], rel.ID)

	// UserPairIndexに追加
	pairKey := r.createUserPairKey(rel.RequesterID, rel.ReceiverID)
	r.userPairIndex[pairKey] = rel.ID

	// StatusIndexに追加
	r.statusIndex[rel.Status] = append(r.statusIndex[rel.Status], rel.ID)

	// UserStatusIndexに追加（リクエスター）
	if r.userStatusIndex[rel.RequesterID] == nil {
		r.userStatusIndex[rel.RequesterID] = make(map[valueobject.RelationshipStatus][]string)
	}
	r.userStatusIndex[rel.RequesterID][rel.Status] = append(r.userStatusIndex[rel.RequesterID][rel.Status], rel.ID)

	// UserStatusIndexに追加（レシーバー）
	if r.userStatusIndex[rel.ReceiverID] == nil {
		r.userStatusIndex[rel.ReceiverID] = make(map[valueobject.RelationshipStatus][]string)
	}
	r.userStatusIndex[rel.ReceiverID][rel.Status] = append(r.userStatusIndex[rel.ReceiverID][rel.Status], rel.ID)
}

// removeFromIndexes は関係をインデックスから削除する
func (r *RelationshipRepository) removeFromIndexes(rel *entity.Relationship) {
	// RequesterIndexから削除
	r.requesterIndex[rel.RequesterID] = r.removeFromSlice(r.requesterIndex[rel.RequesterID], rel.ID)
	if len(r.requesterIndex[rel.RequesterID]) == 0 {
		delete(r.requesterIndex, rel.RequesterID)
	}

	// ReceiverIndexから削除
	r.receiverIndex[rel.ReceiverID] = r.removeFromSlice(r.receiverIndex[rel.ReceiverID], rel.ID)
	if len(r.receiverIndex[rel.ReceiverID]) == 0 {
		delete(r.receiverIndex, rel.ReceiverID)
	}

	// UserPairIndexから削除
	pairKey := r.createUserPairKey(rel.RequesterID, rel.ReceiverID)
	delete(r.userPairIndex, pairKey)

	// StatusIndexから削除
	r.statusIndex[rel.Status] = r.removeFromSlice(r.statusIndex[rel.Status], rel.ID)
	if len(r.statusIndex[rel.Status]) == 0 {
		delete(r.statusIndex, rel.Status)
	}

	// UserStatusIndexから削除（リクエスター）
	if statusMap, exists := r.userStatusIndex[rel.RequesterID]; exists {
		statusMap[rel.Status] = r.removeFromSlice(statusMap[rel.Status], rel.ID)
		if len(statusMap[rel.Status]) == 0 {
			delete(statusMap, rel.Status)
		}
		if len(statusMap) == 0 {
			delete(r.userStatusIndex, rel.RequesterID)
		}
	}

	// UserStatusIndexから削除（レシーバー）
	if statusMap, exists := r.userStatusIndex[rel.ReceiverID]; exists {
		statusMap[rel.Status] = r.removeFromSlice(statusMap[rel.Status], rel.ID)
		if len(statusMap[rel.Status]) == 0 {
			delete(statusMap, rel.Status)
		}
		if len(statusMap) == 0 {
			delete(r.userStatusIndex, rel.ReceiverID)
		}
	}
}

// removeFromSlice はスライスから指定の要素を削除する
// 注: 順序を保持しない高速版。インデックス用途では順序は重要でないため問題なし
func (r *RelationshipRepository) removeFromSlice(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			// 最後の要素を現在位置に移動して、スライスを縮小
			// これにより O(n) から O(1) の削除が可能
			lastIdx := len(slice) - 1
			if i != lastIdx {
				slice[i] = slice[lastIdx]
			}
			return slice[:lastIdx]
		}
	}
	return slice
}

// getRelationshipsWithPagination はページネーション付きで関係を取得する
func (r *RelationshipRepository) getRelationshipsWithPagination(ids []string, offset, limit int) ([]*entity.Relationship, error) {
	// パラメータバリデーション（MorningCallRepositoryと同じ仕様）
	if offset < 0 || limit < 0 {
		return nil, repository.ErrInvalidArgument
	}

	// limit が 0 の場合は空のスライスを返す
	if limit == 0 {
		return []*entity.Relationship{}, nil
	}

	// オフセットが範囲外の場合は空のスライスを返す
	if offset >= len(ids) {
		return []*entity.Relationship{}, nil
	}

	// 終了位置の計算
	end := offset + limit
	if end > len(ids) {
		end = len(ids)
	}

	// 指定範囲の関係を収集
	result := make([]*entity.Relationship, 0, end-offset)
	for i := offset; i < end; i++ {
		if rel, exists := r.relationships[ids[i]]; exists {
			result = append(result, r.copyRelationship(rel))
		}
	}

	return result, nil
}
