package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// TestRelationshipRepository_Create は友達関係作成のテスト
func TestRelationshipRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := NewRelationshipRepository()

	tests := []struct {
		name         string
		relationship *entity.Relationship
		wantErr      error
	}{
		{
			name: "正常な友達関係を作成できる",
			relationship: &entity.Relationship{
				ID:          "rel1",
				RequesterID: "user1",
				ReceiverID:  "user2",
				Status:      valueobject.RelationshipStatusPending,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: nil,
		},
		{
			name:         "nilの関係は作成できない",
			relationship: nil,
			wantErr:      repository.ErrInvalidArgument,
		},
		{
			name: "同じIDの関係は作成できない",
			relationship: &entity.Relationship{
				ID:          "rel1",
				RequesterID: "user3",
				ReceiverID:  "user4",
				Status:      valueobject.RelationshipStatusPending,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: repository.ErrAlreadyExists,
		},
		{
			name: "同じユーザーペア間の関係は作成できない",
			relationship: &entity.Relationship{
				ID:          "rel2",
				RequesterID: "user1",
				ReceiverID:  "user2",
				Status:      valueobject.RelationshipStatusPending,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: repository.ErrAlreadyExists,
		},
		{
			name: "逆方向の同じユーザーペアも作成できない",
			relationship: &entity.Relationship{
				ID:          "rel3",
				RequesterID: "user2",
				ReceiverID:  "user1",
				Status:      valueobject.RelationshipStatusPending,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: repository.ErrAlreadyExists,
		},
	}

	// 最初の関係を作成
	firstRel := tests[0].relationship
	if err := repo.Create(ctx, firstRel); err != nil {
		t.Fatalf("初期データの作成に失敗: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "正常な友達関係を作成できる" {
				return // 既に作成済み
			}

			err := repo.Create(ctx, tt.relationship)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRelationshipRepository_FindByID はID検索のテスト
func TestRelationshipRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	repo := NewRelationshipRepository()

	// テストデータ作成
	rel := &entity.Relationship{
		ID:          "rel1",
		RequesterID: "user1",
		ReceiverID:  "user2",
		Status:      valueobject.RelationshipStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repo.Create(ctx, rel); err != nil {
		t.Fatalf("テストデータの作成に失敗: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "存在する関係を取得できる",
			id:      "rel1",
			wantErr: nil,
		},
		{
			name:    "存在しない関係はエラー",
			id:      "nonexistent",
			wantErr: repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindByID(ctx, tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got.ID != tt.id {
				t.Errorf("FindByID() got ID = %v, want %v", got.ID, tt.id)
			}
		})
	}
}

// TestRelationshipRepository_Update は更新のテスト
func TestRelationshipRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := NewRelationshipRepository()

	// テストデータ作成
	rel := &entity.Relationship{
		ID:          "rel1",
		RequesterID: "user1",
		ReceiverID:  "user2",
		Status:      valueobject.RelationshipStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repo.Create(ctx, rel); err != nil {
		t.Fatalf("テストデータの作成に失敗: %v", err)
	}

	// ユーザーペア重複チェック用の関係を追加
	rel2 := &entity.Relationship{
		ID:          "rel2",
		RequesterID: "user3",
		ReceiverID:  "user4",
		Status:      valueobject.RelationshipStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repo.Create(ctx, rel2); err != nil {
		t.Fatalf("テストデータ2の作成に失敗: %v", err)
	}

	tests := []struct {
		name         string
		relationship *entity.Relationship
		wantErr      error
	}{
		{
			name: "存在する関係を更新できる",
			relationship: &entity.Relationship{
				ID:          "rel1",
				RequesterID: "user1",
				ReceiverID:  "user2",
				Status:      valueobject.RelationshipStatusAccepted,
				CreatedAt:   rel.CreatedAt,
				UpdatedAt:   time.Now(),
			},
			wantErr: nil,
		},
		{
			name:         "nilの関係は更新できない",
			relationship: nil,
			wantErr:      repository.ErrInvalidArgument,
		},
		{
			name: "存在しない関係は更新できない",
			relationship: &entity.Relationship{
				ID:          "nonexistent",
				RequesterID: "user5",
				ReceiverID:  "user6",
				Status:      valueobject.RelationshipStatusPending,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "ユーザーペアを他の関係と重複するように更新できない",
			relationship: &entity.Relationship{
				ID:          "rel2",
				RequesterID: "user1", // user3 -> user1に変更（rel1と重複）
				ReceiverID:  "user2", // user4 -> user2に変更（rel1と重複）
				Status:      valueobject.RelationshipStatusPending,
				CreatedAt:   rel2.CreatedAt,
				UpdatedAt:   time.Now(),
			},
			wantErr: repository.ErrAlreadyExists,
		},
		{
			name: "ユーザーペアを他の関係と逆順で重複するように更新できない",
			relationship: &entity.Relationship{
				ID:          "rel2",
				RequesterID: "user2", // user3 -> user2に変更（rel1の逆順）
				ReceiverID:  "user1", // user4 -> user1に変更（rel1の逆順）
				Status:      valueobject.RelationshipStatusPending,
				CreatedAt:   rel2.CreatedAt,
				UpdatedAt:   time.Now(),
			},
			wantErr: repository.ErrAlreadyExists,
		},
		{
			name: "ユーザーペアを異なるペアに更新できる",
			relationship: &entity.Relationship{
				ID:          "rel2",
				RequesterID: "user7", // 重複しない新しいペア
				ReceiverID:  "user8",
				Status:      valueobject.RelationshipStatusPending,
				CreatedAt:   rel2.CreatedAt,
				UpdatedAt:   time.Now(),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(ctx, tt.relationship)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 更新が成功した場合、変更が反映されているか確認
			if err == nil {
				updated, _ := repo.FindByID(ctx, tt.relationship.ID)
				if updated.Status != tt.relationship.Status {
					t.Errorf("Update() status not updated: got %v, want %v",
						updated.Status, tt.relationship.Status)
				}
			}
		})
	}
}

// TestRelationshipRepository_Delete は削除のテスト
func TestRelationshipRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := NewRelationshipRepository()

	// テストデータ作成
	rel := &entity.Relationship{
		ID:          "rel1",
		RequesterID: "user1",
		ReceiverID:  "user2",
		Status:      valueobject.RelationshipStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repo.Create(ctx, rel); err != nil {
		t.Fatalf("テストデータの作成に失敗: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "存在する関係を削除できる",
			id:      "rel1",
			wantErr: nil,
		},
		{
			name:    "存在しない関係の削除はエラー",
			id:      "nonexistent",
			wantErr: repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 削除が成功した場合、取得できないことを確認
			if err == nil {
				_, err := repo.FindByID(ctx, tt.id)
				if !errors.Is(err, repository.ErrNotFound) {
					t.Errorf("Delete() relationship still exists after deletion")
				}
			}
		})
	}
}

// TestRelationshipRepository_FindByUserPair はユーザーペア検索のテスト
func TestRelationshipRepository_FindByUserPair(t *testing.T) {
	ctx := context.Background()
	repo := NewRelationshipRepository()

	// テストデータ作成
	rel := &entity.Relationship{
		ID:          "rel1",
		RequesterID: "user1",
		ReceiverID:  "user2",
		Status:      valueobject.RelationshipStatusAccepted,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repo.Create(ctx, rel); err != nil {
		t.Fatalf("テストデータの作成に失敗: %v", err)
	}

	tests := []struct {
		name    string
		userID1 string
		userID2 string
		wantID  string
		wantErr error
	}{
		{
			name:    "正順で検索できる",
			userID1: "user1",
			userID2: "user2",
			wantID:  "rel1",
			wantErr: nil,
		},
		{
			name:    "逆順でも検索できる",
			userID1: "user2",
			userID2: "user1",
			wantID:  "rel1",
			wantErr: nil,
		},
		{
			name:    "存在しないペアはエラー",
			userID1: "user3",
			userID2: "user4",
			wantID:  "",
			wantErr: repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindByUserPair(ctx, tt.userID1, tt.userID2)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindByUserPair() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got.ID != tt.wantID {
				t.Errorf("FindByUserPair() got ID = %v, want %v", got.ID, tt.wantID)
			}
		})
	}
}

// TestRelationshipRepository_AreFriends は友達確認のテスト
func TestRelationshipRepository_AreFriends(t *testing.T) {
	ctx := context.Background()
	repo := NewRelationshipRepository()

	// テストデータ作成
	relationships := []*entity.Relationship{
		{
			ID:          "rel1",
			RequesterID: "user1",
			ReceiverID:  "user2",
			Status:      valueobject.RelationshipStatusAccepted,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rel2",
			RequesterID: "user3",
			ReceiverID:  "user4",
			Status:      valueobject.RelationshipStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rel3",
			RequesterID: "user5",
			ReceiverID:  "user6",
			Status:      valueobject.RelationshipStatusBlocked,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, rel := range relationships {
		if err := repo.Create(ctx, rel); err != nil {
			t.Fatalf("テストデータの作成に失敗: %v", err)
		}
	}

	tests := []struct {
		name       string
		userID1    string
		userID2    string
		wantResult bool
	}{
		{
			name:       "承認済みの関係は友達",
			userID1:    "user1",
			userID2:    "user2",
			wantResult: true,
		},
		{
			name:       "承認待ちの関係は友達ではない",
			userID1:    "user3",
			userID2:    "user4",
			wantResult: false,
		},
		{
			name:       "ブロック済みの関係は友達ではない",
			userID1:    "user5",
			userID2:    "user6",
			wantResult: false,
		},
		{
			name:       "関係が存在しない場合は友達ではない",
			userID1:    "user7",
			userID2:    "user8",
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.AreFriends(ctx, tt.userID1, tt.userID2)
			if err != nil {
				t.Errorf("AreFriends() error = %v", err)
				return
			}
			if got != tt.wantResult {
				t.Errorf("AreFriends() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

// TestRelationshipRepository_FindFriendsByUserID は友達検索のテスト
func TestRelationshipRepository_FindFriendsByUserID(t *testing.T) {
	ctx := context.Background()
	repo := NewRelationshipRepository()

	// テストデータ作成
	relationships := []*entity.Relationship{
		{
			ID:          "rel1",
			RequesterID: "user1",
			ReceiverID:  "user2",
			Status:      valueobject.RelationshipStatusAccepted,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rel2",
			RequesterID: "user1",
			ReceiverID:  "user3",
			Status:      valueobject.RelationshipStatusAccepted,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rel3",
			RequesterID: "user4",
			ReceiverID:  "user1",
			Status:      valueobject.RelationshipStatusAccepted,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rel4",
			RequesterID: "user1",
			ReceiverID:  "user5",
			Status:      valueobject.RelationshipStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rel5",
			RequesterID: "user1",
			ReceiverID:  "user6",
			Status:      valueobject.RelationshipStatusBlocked,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, rel := range relationships {
		if err := repo.Create(ctx, rel); err != nil {
			t.Fatalf("テストデータの作成に失敗: %v", err)
		}
	}

	// user1の友達を検索
	friends, err := repo.FindFriendsByUserID(ctx, "user1", 0, 10)
	if err != nil {
		t.Fatalf("FindFriendsByUserID() error = %v", err)
	}

	// user1は3人と友達（user2, user3, user4）
	if len(friends) != 3 {
		t.Errorf("FindFriendsByUserID() got %d friends, want 3", len(friends))
	}

	// すべての関係がAcceptedステータスであることを確認
	for _, f := range friends {
		if f.Status != valueobject.RelationshipStatusAccepted {
			t.Errorf("FindFriendsByUserID() got non-accepted status: %v", f.Status)
		}
		// user1が関係に含まれていることを確認
		if !f.InvolvesUser("user1") {
			t.Errorf("FindFriendsByUserID() returned relationship not involving user1")
		}
	}
}

// TestRelationshipRepository_FindPendingRequestsByReceiverID は承認待ちリクエスト検索のテスト
func TestRelationshipRepository_FindPendingRequestsByReceiverID(t *testing.T) {
	ctx := context.Background()
	repo := NewRelationshipRepository()

	// テストデータ作成
	relationships := []*entity.Relationship{
		{
			ID:          "rel1",
			RequesterID: "user2",
			ReceiverID:  "user1",
			Status:      valueobject.RelationshipStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rel2",
			RequesterID: "user3",
			ReceiverID:  "user1",
			Status:      valueobject.RelationshipStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rel3",
			RequesterID: "user1",
			ReceiverID:  "user4",
			Status:      valueobject.RelationshipStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rel4",
			RequesterID: "user5",
			ReceiverID:  "user1",
			Status:      valueobject.RelationshipStatusAccepted,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, rel := range relationships {
		if err := repo.Create(ctx, rel); err != nil {
			t.Fatalf("テストデータの作成に失敗: %v", err)
		}
	}

	// user1が受信した承認待ちリクエストを検索
	requests, err := repo.FindPendingRequestsByReceiverID(ctx, "user1", 0, 10)
	if err != nil {
		t.Fatalf("FindPendingRequestsByReceiverID() error = %v", err)
	}

	// user1は2つの承認待ちリクエストを受信（user2, user3から）
	if len(requests) != 2 {
		t.Errorf("FindPendingRequestsByReceiverID() got %d requests, want 2", len(requests))
	}

	// すべてがPendingステータスでuser1が受信者であることを確認
	for _, req := range requests {
		if req.Status != valueobject.RelationshipStatusPending {
			t.Errorf("FindPendingRequestsByReceiverID() got non-pending status: %v", req.Status)
		}
		if req.ReceiverID != "user1" {
			t.Errorf("FindPendingRequestsByReceiverID() got wrong receiver: %v", req.ReceiverID)
		}
	}
}

// TestRelationshipRepository_CountFriendsByUserID は友達数カウントのテスト
func TestRelationshipRepository_CountFriendsByUserID(t *testing.T) {
	ctx := context.Background()
	repo := NewRelationshipRepository()

	// テストデータ作成（前のテストと同じ）
	relationships := []*entity.Relationship{
		{
			ID:          "rel1",
			RequesterID: "user1",
			ReceiverID:  "user2",
			Status:      valueobject.RelationshipStatusAccepted,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rel2",
			RequesterID: "user3",
			ReceiverID:  "user1",
			Status:      valueobject.RelationshipStatusAccepted,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rel3",
			RequesterID: "user1",
			ReceiverID:  "user4",
			Status:      valueobject.RelationshipStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, rel := range relationships {
		if err := repo.Create(ctx, rel); err != nil {
			t.Fatalf("テストデータの作成に失敗: %v", err)
		}
	}

	tests := []struct {
		name      string
		userID    string
		wantCount int
	}{
		{
			name:      "user1の友達数",
			userID:    "user1",
			wantCount: 2,
		},
		{
			name:      "user2の友達数",
			userID:    "user2",
			wantCount: 1,
		},
		{
			name:      "友達がいないユーザー",
			userID:    "user5",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := repo.CountFriendsByUserID(ctx, tt.userID)
			if err != nil {
				t.Errorf("CountFriendsByUserID() error = %v", err)
				return
			}
			if count != tt.wantCount {
				t.Errorf("CountFriendsByUserID() = %d, want %d", count, tt.wantCount)
			}
		})
	}
}

// TestRelationshipRepository_Pagination はページネーションのテスト
func TestRelationshipRepository_Pagination(t *testing.T) {
	ctx := context.Background()
	repo := NewRelationshipRepository()

	// 10個のテストデータを作成
	for i := 0; i < 10; i++ {
		rel := &entity.Relationship{
			ID:          generateTestRelationshipID(i),
			RequesterID: "user1",
			ReceiverID:  generateTestUserID(i + 2),
			Status:      valueobject.RelationshipStatusAccepted,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := repo.Create(ctx, rel); err != nil {
			t.Fatalf("テストデータの作成に失敗: %v", err)
		}
	}

	tests := []struct {
		name       string
		offset     int
		limit      int
		wantLength int
		wantErr    error
	}{
		{
			name:       "最初の5件",
			offset:     0,
			limit:      5,
			wantLength: 5,
			wantErr:    nil,
		},
		{
			name:       "次の5件",
			offset:     5,
			limit:      5,
			wantLength: 5,
			wantErr:    nil,
		},
		{
			name:       "最後の3件",
			offset:     7,
			limit:      5,
			wantLength: 3,
			wantErr:    nil,
		},
		{
			name:       "範囲外のオフセット",
			offset:     15,
			limit:      5,
			wantLength: 0,
			wantErr:    nil,
		},
		{
			name:       "limitが0の場合は空のスライス",
			offset:     0,
			limit:      0,
			wantLength: 0,
			wantErr:    nil,
		},
		{
			name:       "負のoffsetはエラー",
			offset:     -1,
			limit:      5,
			wantLength: 0,
			wantErr:    repository.ErrInvalidArgument,
		},
		{
			name:       "負のlimitはエラー",
			offset:     0,
			limit:      -1,
			wantLength: 0,
			wantErr:    repository.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.FindByRequesterID(ctx, "user1", tt.offset, tt.limit)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindByRequesterID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(results) != tt.wantLength {
				t.Errorf("FindByRequesterID() returned %d items, want %d", len(results), tt.wantLength)
			}
		})
	}
}

// TestRelationshipRepository_ConcurrentAccess は並行アクセスのテスト
func TestRelationshipRepository_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	repo := NewRelationshipRepository()

	// 並行して複数の関係を作成
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			rel := &entity.Relationship{
				ID:          generateTestRelationshipID(index),
				RequesterID: generateTestUserID(index),
				ReceiverID:  generateTestUserID(index + 100),
				Status:      valueobject.RelationshipStatusPending,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if err := repo.Create(ctx, rel); err != nil {
				t.Errorf("並行Create失敗: %v", err)
			}
			done <- true
		}(i)
	}

	// すべての処理が完了するまで待つ
	for i := 0; i < 10; i++ {
		<-done
	}

	// 作成された関係の数を確認
	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 10 {
		t.Errorf("並行作成後のカウント = %d, want 10", count)
	}

	// 並行して読み取りと更新を実行
	for i := 0; i < 10; i++ {
		go func(index int) {
			relID := generateTestRelationshipID(index)

			// 読み取り
			rel, err := repo.FindByID(ctx, relID)
			if err != nil {
				t.Errorf("並行FindByID失敗: %v", err)
				done <- true
				return
			}

			// ステータス更新
			rel.Status = valueobject.RelationshipStatusAccepted
			if err := repo.Update(ctx, rel); err != nil {
				t.Errorf("並行Update失敗: %v", err)
			}
			done <- true
		}(i)
	}

	// すべての処理が完了するまで待つ
	for i := 0; i < 10; i++ {
		<-done
	}

	// すべての関係がAcceptedになっていることを確認
	allRels, err := repo.FindAll(ctx, 0, 20) // 全件取得するため十分な数を指定
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	for _, rel := range allRels {
		if rel.Status != valueobject.RelationshipStatusAccepted {
			t.Errorf("並行更新後のステータスが不正: got %v, want Accepted", rel.Status)
		}
	}
}
