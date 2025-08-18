package memory

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// テスト用のヘルパー関数

func createTestMorningCall(id, senderID, receiverID string, scheduledTime time.Time, status valueobject.MorningCallStatus) *entity.MorningCall {
	return &entity.MorningCall{
		ID:            id,
		SenderID:      senderID,
		ReceiverID:    receiverID,
		ScheduledTime: scheduledTime,
		Message:       "Test message",
		Status:        status,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func TestNewMorningCallRepository(t *testing.T) {
	repo := NewMorningCallRepository()

	if repo == nil {
		t.Fatal("NewMorningCallRepository returned nil")
	}

	if repo.morningCalls == nil {
		t.Error("morningCalls map is nil")
	}

	if repo.senderIndex == nil {
		t.Error("senderIndex map is nil")
	}

	if repo.receiverIndex == nil {
		t.Error("receiverIndex map is nil")
	}

	if repo.statusIndex == nil {
		t.Error("statusIndex map is nil")
	}

	if repo.userPairIndex == nil {
		t.Error("userPairIndex map is nil")
	}
}

func TestMorningCallRepository_Create(t *testing.T) {
	tests := []struct {
		name        string
		morningCall *entity.MorningCall
		setupFunc   func(*MorningCallRepository)
		wantErr     error
	}{
		{
			name: "正常なモーニングコール作成",
			morningCall: createTestMorningCall(
				"mc1", "user1", "user2",
				time.Now().Add(1*time.Hour),
				valueobject.MorningCallStatusScheduled,
			),
			setupFunc: func(r *MorningCallRepository) {},
			wantErr:   nil,
		},
		{
			name:        "nilのモーニングコール",
			morningCall: nil,
			setupFunc:   func(r *MorningCallRepository) {},
			wantErr:     repository.ErrInvalidArgument,
		},
		{
			name: "既存のIDでモーニングコール作成",
			morningCall: createTestMorningCall(
				"mc1", "user1", "user2",
				time.Now().Add(1*time.Hour),
				valueobject.MorningCallStatusScheduled,
			),
			setupFunc: func(r *MorningCallRepository) {
				// 事前に同じIDのモーニングコールを作成（Createメソッド経由で一貫性を保つ）
				ctx := context.Background()
				if err := r.Create(ctx, createTestMorningCall(
					"mc1", "user3", "user4",
					time.Now().Add(2*time.Hour),
					valueobject.MorningCallStatusScheduled,
				)); err != nil {
					// セットアップエラーはテスト失敗として扱う
					t.Fatalf("Failed to setup test: %v", err)
				}
			},
			wantErr: repository.ErrAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			err := repo.Create(context.Background(), tt.morningCall)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				// 作成成功時の検証
				// メインストレージに存在確認
				if _, exists := repo.morningCalls[tt.morningCall.ID]; !exists {
					t.Error("MorningCall not found in main storage")
				}

				// インデックスの確認
				if !containsString(repo.senderIndex[tt.morningCall.SenderID], tt.morningCall.ID) {
					t.Error("MorningCall ID not found in sender index")
				}

				if !containsString(repo.receiverIndex[tt.morningCall.ReceiverID], tt.morningCall.ID) {
					t.Error("MorningCall ID not found in receiver index")
				}

				if !containsString(repo.statusIndex[tt.morningCall.Status], tt.morningCall.ID) {
					t.Error("MorningCall ID not found in status index")
				}

				pairKey := repo.generateUserPairKey(tt.morningCall.SenderID, tt.morningCall.ReceiverID)
				if !containsString(repo.userPairIndex[pairKey], tt.morningCall.ID) {
					t.Error("MorningCall ID not found in user pair index")
				}

				// オリジナルの変更が保存されたデータに影響しないことを確認
				tt.morningCall.Message = "Modified"
				if repo.morningCalls[tt.morningCall.ID].Message == "Modified" {
					t.Error("Original modification affected stored data")
				}
			}
		})
	}
}

func TestMorningCallRepository_FindByID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupFunc func(*MorningCallRepository)
		want      *entity.MorningCall
		wantErr   error
	}{
		{
			name: "存在するモーニングコールの検索",
			id:   "mc1",
			setupFunc: func(r *MorningCallRepository) {
				ctx := context.Background()
				mc := createTestMorningCall(
					"mc1", "user1", "user2",
					time.Now().Add(1*time.Hour),
					valueobject.MorningCallStatusScheduled,
				)
				if err := r.Create(ctx, mc); err != nil {
					t.Fatalf("Failed to setup test: %v", err)
				}
			},
			want: createTestMorningCall(
				"mc1", "user1", "user2",
				time.Now().Add(1*time.Hour),
				valueobject.MorningCallStatusScheduled,
			),
			wantErr: nil,
		},
		{
			name:      "存在しないモーニングコールの検索",
			id:        "nonexistent",
			setupFunc: func(r *MorningCallRepository) {},
			want:      nil,
			wantErr:   repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			got, err := repo.FindByID(context.Background(), tt.id)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if got.ID != tt.want.ID {
					t.Errorf("FindByID() got ID = %v, want %v", got.ID, tt.want.ID)
				}

				// 返されたデータの変更が元のデータに影響しないことを確認
				originalMessage := repo.morningCalls[tt.id].Message
				got.Message = "Modified"
				if repo.morningCalls[tt.id].Message != originalMessage {
					t.Error("Returned data modification affected stored data")
				}
			}
		})
	}
}

func TestMorningCallRepository_Update(t *testing.T) {
	tests := []struct {
		name        string
		morningCall *entity.MorningCall
		setupFunc   func(*MorningCallRepository)
		wantErr     error
	}{
		{
			name: "正常な更新",
			morningCall: createTestMorningCall(
				"mc1", "user1", "user2",
				time.Now().Add(2*time.Hour),
				valueobject.MorningCallStatusDelivered,
			),
			setupFunc: func(r *MorningCallRepository) {
				ctx := context.Background()
				mc := createTestMorningCall(
					"mc1", "user1", "user2",
					time.Now().Add(1*time.Hour),
					valueobject.MorningCallStatusScheduled,
				)
				if err := r.Create(ctx, mc); err != nil {
					t.Fatalf("Failed to setup test: %v", err)
				}
			},
			wantErr: nil,
		},
		{
			name:        "nilのモーニングコール",
			morningCall: nil,
			setupFunc:   func(r *MorningCallRepository) {},
			wantErr:     repository.ErrInvalidArgument,
		},
		{
			name: "存在しないモーニングコールの更新",
			morningCall: createTestMorningCall(
				"nonexistent", "user1", "user2",
				time.Now().Add(1*time.Hour),
				valueobject.MorningCallStatusScheduled,
			),
			setupFunc: func(r *MorningCallRepository) {},
			wantErr:   repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			err := repo.Update(context.Background(), tt.morningCall)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				// 更新成功時の検証
				stored := repo.morningCalls[tt.morningCall.ID]
				if stored.Status != tt.morningCall.Status {
					t.Errorf("Status not updated: got %v, want %v", stored.Status, tt.morningCall.Status)
				}

				// インデックスの更新確認
				if !containsString(repo.statusIndex[tt.morningCall.Status], tt.morningCall.ID) {
					t.Error("Status index not updated")
				}
			}
		})
	}
}

func TestMorningCallRepository_Delete(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupFunc func(*MorningCallRepository)
		wantErr   error
	}{
		{
			name: "正常な削除",
			id:   "mc1",
			setupFunc: func(r *MorningCallRepository) {
				ctx := context.Background()
				mc := createTestMorningCall(
					"mc1", "user1", "user2",
					time.Now().Add(1*time.Hour),
					valueobject.MorningCallStatusScheduled,
				)
				if err := r.Create(ctx, mc); err != nil {
					t.Fatalf("Failed to setup test: %v", err)
				}
			},
			wantErr: nil,
		},
		{
			name:      "存在しないモーニングコールの削除",
			id:        "nonexistent",
			setupFunc: func(r *MorningCallRepository) {},
			wantErr:   repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			err := repo.Delete(context.Background(), tt.id)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				// 削除成功時の検証
				if _, exists := repo.morningCalls[tt.id]; exists {
					t.Error("MorningCall still exists after deletion")
				}

				// インデックスからも削除されていることを確認
				for _, ids := range repo.senderIndex {
					if containsString(ids, tt.id) {
						t.Error("MorningCall ID still in sender index")
					}
				}

				for _, ids := range repo.receiverIndex {
					if containsString(ids, tt.id) {
						t.Error("MorningCall ID still in receiver index")
					}
				}

				for _, ids := range repo.statusIndex {
					if containsString(ids, tt.id) {
						t.Error("MorningCall ID still in status index")
					}
				}

				for _, ids := range repo.userPairIndex {
					if containsString(ids, tt.id) {
						t.Error("MorningCall ID still in user pair index")
					}
				}
			}
		})
	}
}

func TestMorningCallRepository_ExistsByID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupFunc func(*MorningCallRepository)
		want      bool
	}{
		{
			name: "存在するモーニングコール",
			id:   "mc1",
			setupFunc: func(r *MorningCallRepository) {
				ctx := context.Background()
				mc := createTestMorningCall(
					"mc1", "user1", "user2",
					time.Now().Add(1*time.Hour),
					valueobject.MorningCallStatusScheduled,
				)
				if err := r.Create(ctx, mc); err != nil {
					t.Fatalf("Failed to setup test: %v", err)
				}
			},
			want: true,
		},
		{
			name:      "存在しないモーニングコール",
			id:        "nonexistent",
			setupFunc: func(r *MorningCallRepository) {},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			got, err := repo.ExistsByID(context.Background(), tt.id)
			if err != nil {
				t.Fatalf("ExistsByID() unexpected error = %v", err)
			}

			if got != tt.want {
				t.Errorf("ExistsByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMorningCallRepository_FindBySenderID(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name      string
		senderID  string
		offset    int
		limit     int
		setupFunc func(*MorningCallRepository)
		wantCount int
		wantErr   error
	}{
		{
			name:     "送信者のモーニングコール取得",
			senderID: "user1",
			offset:   0,
			limit:    10,
			setupFunc: func(r *MorningCallRepository) {
				mcs := []*entity.MorningCall{
					createTestMorningCall("mc1", "user1", "user2", baseTime.Add(1*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc2", "user1", "user3", baseTime.Add(2*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc3", "user2", "user1", baseTime.Add(3*time.Hour), valueobject.MorningCallStatusScheduled),
				}
				for _, mc := range mcs {
					r.morningCalls[mc.ID] = mc
					r.addToIndexes(mc)
				}
			},
			wantCount: 2,
			wantErr:   nil,
		},
		{
			name:      "存在しない送信者",
			senderID:  "nonexistent",
			offset:    0,
			limit:     10,
			setupFunc: func(r *MorningCallRepository) {},
			wantCount: 0,
			wantErr:   nil,
		},
		{
			name:     "ページネーション",
			senderID: "user1",
			offset:   1,
			limit:    1,
			setupFunc: func(r *MorningCallRepository) {
				mcs := []*entity.MorningCall{
					createTestMorningCall("mc1", "user1", "user2", baseTime.Add(1*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc2", "user1", "user3", baseTime.Add(2*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc3", "user1", "user4", baseTime.Add(3*time.Hour), valueobject.MorningCallStatusScheduled),
				}
				for _, mc := range mcs {
					r.morningCalls[mc.ID] = mc
					r.addToIndexes(mc)
				}
			},
			wantCount: 1,
			wantErr:   nil,
		},
		{
			name:      "不正なoffset",
			senderID:  "user1",
			offset:    -1,
			limit:     10,
			setupFunc: func(r *MorningCallRepository) {},
			wantCount: 0,
			wantErr:   repository.ErrInvalidArgument,
		},
		{
			name:      "不正なlimit",
			senderID:  "user1",
			offset:    0,
			limit:     -1,
			setupFunc: func(r *MorningCallRepository) {},
			wantCount: 0,
			wantErr:   repository.ErrInvalidArgument,
		},
		{
			name:      "limit が 0",
			senderID:  "user1",
			offset:    0,
			limit:     0,
			setupFunc: func(r *MorningCallRepository) {},
			wantCount: 0,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			got, err := repo.FindBySenderID(context.Background(), tt.senderID, tt.offset, tt.limit)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindBySenderID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(got) != tt.wantCount {
					t.Errorf("FindBySenderID() returned %d items, want %d", len(got), tt.wantCount)
				}

				// 送信者IDが正しいことを確認
				for _, mc := range got {
					if mc.SenderID != tt.senderID && tt.wantCount > 0 {
						t.Errorf("MorningCall has wrong SenderID: got %v, want %v", mc.SenderID, tt.senderID)
					}
				}

				// ソート順の確認（降順：新しいものが先）
				for i := 1; i < len(got); i++ {
					if got[i-1].ScheduledTime.Before(got[i].ScheduledTime) {
						t.Error("MorningCalls are not sorted in descending order by ScheduledTime")
					}
				}
			}
		})
	}
}

func TestMorningCallRepository_FindByReceiverID(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name       string
		receiverID string
		offset     int
		limit      int
		setupFunc  func(*MorningCallRepository)
		wantCount  int
		wantErr    error
	}{
		{
			name:       "受信者のモーニングコール取得",
			receiverID: "user2",
			offset:     0,
			limit:      10,
			setupFunc: func(r *MorningCallRepository) {
				mcs := []*entity.MorningCall{
					createTestMorningCall("mc1", "user1", "user2", baseTime.Add(1*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc2", "user3", "user2", baseTime.Add(2*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc3", "user1", "user3", baseTime.Add(3*time.Hour), valueobject.MorningCallStatusScheduled),
				}
				for _, mc := range mcs {
					r.morningCalls[mc.ID] = mc
					r.addToIndexes(mc)
				}
			},
			wantCount: 2,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			got, err := repo.FindByReceiverID(context.Background(), tt.receiverID, tt.offset, tt.limit)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindByReceiverID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(got) != tt.wantCount {
					t.Errorf("FindByReceiverID() returned %d items, want %d", len(got), tt.wantCount)
				}

				// 受信者IDが正しいことを確認
				for _, mc := range got {
					if mc.ReceiverID != tt.receiverID && tt.wantCount > 0 {
						t.Errorf("MorningCall has wrong ReceiverID: got %v, want %v", mc.ReceiverID, tt.receiverID)
					}
				}

				// ソート順の確認（昇順：直近のものが先）
				for i := 1; i < len(got); i++ {
					if got[i-1].ScheduledTime.After(got[i].ScheduledTime) {
						t.Error("MorningCalls are not sorted in ascending order by ScheduledTime")
					}
				}
			}
		})
	}
}

func TestMorningCallRepository_FindByStatus(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name      string
		status    valueobject.MorningCallStatus
		offset    int
		limit     int
		setupFunc func(*MorningCallRepository)
		wantCount int
		wantErr   error
	}{
		{
			name:   "ステータスでモーニングコール取得",
			status: valueobject.MorningCallStatusScheduled,
			offset: 0,
			limit:  10,
			setupFunc: func(r *MorningCallRepository) {
				mcs := []*entity.MorningCall{
					createTestMorningCall("mc1", "user1", "user2", baseTime.Add(1*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc2", "user1", "user3", baseTime.Add(2*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc3", "user2", "user1", baseTime.Add(3*time.Hour), valueobject.MorningCallStatusDelivered),
					createTestMorningCall("mc4", "user3", "user2", baseTime.Add(4*time.Hour), valueobject.MorningCallStatusConfirmed),
				}
				for _, mc := range mcs {
					r.morningCalls[mc.ID] = mc
					r.addToIndexes(mc)
				}
			},
			wantCount: 2,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			got, err := repo.FindByStatus(context.Background(), tt.status, tt.offset, tt.limit)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindByStatus() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(got) != tt.wantCount {
					t.Errorf("FindByStatus() returned %d items, want %d", len(got), tt.wantCount)
				}

				// ステータスが正しいことを確認
				for _, mc := range got {
					if mc.Status != tt.status && tt.wantCount > 0 {
						t.Errorf("MorningCall has wrong Status: got %v, want %v", mc.Status, tt.status)
					}
				}
			}
		})
	}
}

func TestMorningCallRepository_FindScheduledBefore(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name      string
		before    time.Time
		offset    int
		limit     int
		setupFunc func(*MorningCallRepository)
		wantCount int
		wantErr   error
	}{
		{
			name:   "指定時刻前のモーニングコール取得",
			before: baseTime.Add(3 * time.Hour),
			offset: 0,
			limit:  10,
			setupFunc: func(r *MorningCallRepository) {
				mcs := []*entity.MorningCall{
					createTestMorningCall("mc1", "user1", "user2", baseTime.Add(1*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc2", "user1", "user3", baseTime.Add(2*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc3", "user2", "user1", baseTime.Add(4*time.Hour), valueobject.MorningCallStatusScheduled),
				}
				for _, mc := range mcs {
					r.morningCalls[mc.ID] = mc
				}
			},
			wantCount: 2,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			got, err := repo.FindScheduledBefore(context.Background(), tt.before, tt.offset, tt.limit)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindScheduledBefore() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(got) != tt.wantCount {
					t.Errorf("FindScheduledBefore() returned %d items, want %d", len(got), tt.wantCount)
				}

				// 時刻が条件を満たすことを確認
				for _, mc := range got {
					if !mc.ScheduledTime.Before(tt.before) {
						t.Errorf("MorningCall scheduled at %v is not before %v", mc.ScheduledTime, tt.before)
					}
				}

				// ソート順の確認（昇順）
				for i := 1; i < len(got); i++ {
					if got[i-1].ScheduledTime.After(got[i].ScheduledTime) {
						t.Error("MorningCalls are not sorted in ascending order")
					}
				}
			}
		})
	}
}

func TestMorningCallRepository_FindScheduledBetween(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name      string
		start     time.Time
		end       time.Time
		offset    int
		limit     int
		setupFunc func(*MorningCallRepository)
		wantCount int
		wantErr   error
	}{
		{
			name:   "期間内のモーニングコール取得",
			start:  baseTime.Add(2 * time.Hour),
			end:    baseTime.Add(4 * time.Hour),
			offset: 0,
			limit:  10,
			setupFunc: func(r *MorningCallRepository) {
				mcs := []*entity.MorningCall{
					createTestMorningCall("mc1", "user1", "user2", baseTime.Add(1*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc2", "user1", "user3", baseTime.Add(2*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc3", "user2", "user1", baseTime.Add(3*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc4", "user3", "user2", baseTime.Add(4*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc5", "user1", "user2", baseTime.Add(5*time.Hour), valueobject.MorningCallStatusScheduled),
				}
				for _, mc := range mcs {
					r.morningCalls[mc.ID] = mc
				}
			},
			wantCount: 3, // mc2, mc3, mc4
			wantErr:   nil,
		},
		{
			name:      "開始時刻が終了時刻より後",
			start:     baseTime.Add(4 * time.Hour),
			end:       baseTime.Add(2 * time.Hour),
			offset:    0,
			limit:     10,
			setupFunc: func(r *MorningCallRepository) {},
			wantCount: 0,
			wantErr:   repository.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			got, err := repo.FindScheduledBetween(context.Background(), tt.start, tt.end, tt.offset, tt.limit)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindScheduledBetween() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(got) != tt.wantCount {
					t.Errorf("FindScheduledBetween() returned %d items, want %d", len(got), tt.wantCount)
				}

				// 時刻が条件を満たすことを確認
				for _, mc := range got {
					if mc.ScheduledTime.Before(tt.start) || mc.ScheduledTime.After(tt.end) {
						t.Errorf("MorningCall scheduled at %v is not between %v and %v", mc.ScheduledTime, tt.start, tt.end)
					}
				}
			}
		})
	}
}

func TestMorningCallRepository_FindActiveByUserPair(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name       string
		senderID   string
		receiverID string
		setupFunc  func(*MorningCallRepository)
		wantCount  int
		wantErr    error
	}{
		{
			name:       "ユーザーペア間のアクティブなモーニングコール",
			senderID:   "user1",
			receiverID: "user2",
			setupFunc: func(r *MorningCallRepository) {
				mcs := []*entity.MorningCall{
					createTestMorningCall("mc1", "user1", "user2", baseTime.Add(1*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc2", "user1", "user2", baseTime.Add(2*time.Hour), valueobject.MorningCallStatusDelivered),
					createTestMorningCall("mc3", "user1", "user2", baseTime.Add(3*time.Hour), valueobject.MorningCallStatusConfirmed),
					createTestMorningCall("mc4", "user1", "user3", baseTime.Add(4*time.Hour), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc5", "user2", "user1", baseTime.Add(5*time.Hour), valueobject.MorningCallStatusScheduled),
				}
				for _, mc := range mcs {
					r.morningCalls[mc.ID] = mc
					r.addToIndexes(mc)
				}
			},
			wantCount: 2, // mc1, mc2 (scheduled と delivered のみ)
			wantErr:   nil,
		},
		{
			name:       "アクティブなモーニングコールがない",
			senderID:   "user1",
			receiverID: "user2",
			setupFunc: func(r *MorningCallRepository) {
				mcs := []*entity.MorningCall{
					createTestMorningCall("mc1", "user1", "user2", baseTime.Add(1*time.Hour), valueobject.MorningCallStatusConfirmed),
					createTestMorningCall("mc2", "user1", "user2", baseTime.Add(2*time.Hour), valueobject.MorningCallStatusCancelled),
				}
				for _, mc := range mcs {
					r.morningCalls[mc.ID] = mc
					r.addToIndexes(mc)
				}
			},
			wantCount: 0,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			got, err := repo.FindActiveByUserPair(context.Background(), tt.senderID, tt.receiverID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindActiveByUserPair() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(got) != tt.wantCount {
					t.Errorf("FindActiveByUserPair() returned %d items, want %d", len(got), tt.wantCount)
				}

				// ユーザーペアとステータスが正しいことを確認
				for _, mc := range got {
					if mc.SenderID != tt.senderID || mc.ReceiverID != tt.receiverID {
						t.Errorf("Wrong user pair: got sender=%v receiver=%v, want sender=%v receiver=%v",
							mc.SenderID, mc.ReceiverID, tt.senderID, tt.receiverID)
					}

					if mc.Status != valueobject.MorningCallStatusScheduled &&
						mc.Status != valueobject.MorningCallStatusDelivered {
						t.Errorf("Non-active status: %v", mc.Status)
					}
				}
			}
		})
	}
}

func TestMorningCallRepository_Count(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*MorningCallRepository)
		want      int
	}{
		{
			name: "複数のモーニングコール",
			setupFunc: func(r *MorningCallRepository) {
				for i := 0; i < 5; i++ {
					id := fmt.Sprintf("mc%d", i)
					r.morningCalls[id] = createTestMorningCall(
						id, "user1", "user2",
						time.Now().Add(time.Duration(i)*time.Hour),
						valueobject.MorningCallStatusScheduled,
					)
				}
			},
			want: 5,
		},
		{
			name:      "モーニングコールがない",
			setupFunc: func(r *MorningCallRepository) {},
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			got, err := repo.Count(context.Background())
			if err != nil {
				t.Fatalf("Count() unexpected error = %v", err)
			}

			if got != tt.want {
				t.Errorf("Count() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMorningCallRepository_CountBySenderID(t *testing.T) {
	tests := []struct {
		name      string
		senderID  string
		setupFunc func(*MorningCallRepository)
		want      int
	}{
		{
			name:     "送信者のモーニングコール数",
			senderID: "user1",
			setupFunc: func(r *MorningCallRepository) {
				mcs := []*entity.MorningCall{
					createTestMorningCall("mc1", "user1", "user2", time.Now(), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc2", "user1", "user3", time.Now(), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc3", "user2", "user1", time.Now(), valueobject.MorningCallStatusScheduled),
				}
				for _, mc := range mcs {
					r.morningCalls[mc.ID] = mc
					r.addToIndexes(mc)
				}
			},
			want: 2,
		},
		{
			name:      "存在しない送信者",
			senderID:  "nonexistent",
			setupFunc: func(r *MorningCallRepository) {},
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			got, err := repo.CountBySenderID(context.Background(), tt.senderID)
			if err != nil {
				t.Fatalf("CountBySenderID() unexpected error = %v", err)
			}

			if got != tt.want {
				t.Errorf("CountBySenderID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMorningCallRepository_CountByReceiverID(t *testing.T) {
	tests := []struct {
		name       string
		receiverID string
		setupFunc  func(*MorningCallRepository)
		want       int
	}{
		{
			name:       "受信者のモーニングコール数",
			receiverID: "user2",
			setupFunc: func(r *MorningCallRepository) {
				mcs := []*entity.MorningCall{
					createTestMorningCall("mc1", "user1", "user2", time.Now(), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc2", "user3", "user2", time.Now(), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc3", "user2", "user1", time.Now(), valueobject.MorningCallStatusScheduled),
				}
				for _, mc := range mcs {
					r.morningCalls[mc.ID] = mc
					r.addToIndexes(mc)
				}
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			got, err := repo.CountByReceiverID(context.Background(), tt.receiverID)
			if err != nil {
				t.Fatalf("CountByReceiverID() unexpected error = %v", err)
			}

			if got != tt.want {
				t.Errorf("CountByReceiverID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMorningCallRepository_CountByStatus(t *testing.T) {
	tests := []struct {
		name      string
		status    valueobject.MorningCallStatus
		setupFunc func(*MorningCallRepository)
		want      int
	}{
		{
			name:   "ステータスごとのモーニングコール数",
			status: valueobject.MorningCallStatusScheduled,
			setupFunc: func(r *MorningCallRepository) {
				mcs := []*entity.MorningCall{
					createTestMorningCall("mc1", "user1", "user2", time.Now(), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc2", "user1", "user3", time.Now(), valueobject.MorningCallStatusScheduled),
					createTestMorningCall("mc3", "user2", "user1", time.Now(), valueobject.MorningCallStatusDelivered),
				}
				for _, mc := range mcs {
					r.morningCalls[mc.ID] = mc
					r.addToIndexes(mc)
				}
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			got, err := repo.CountByStatus(context.Background(), tt.status)
			if err != nil {
				t.Fatalf("CountByStatus() unexpected error = %v", err)
			}

			if got != tt.want {
				t.Errorf("CountByStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMorningCallRepository_FindAll(t *testing.T) {
	tests := []struct {
		name      string
		offset    int
		limit     int
		setupFunc func(*MorningCallRepository)
		wantCount int
		wantErr   error
	}{
		{
			name:   "すべてのモーニングコール取得",
			offset: 0,
			limit:  10,
			setupFunc: func(r *MorningCallRepository) {
				for i := 0; i < 3; i++ {
					mc := createTestMorningCall(
						fmt.Sprintf("mc%d", i), "user1", "user2",
						time.Now().Add(time.Duration(i)*time.Hour),
						valueobject.MorningCallStatusScheduled,
					)
					r.morningCalls[mc.ID] = mc
				}
			},
			wantCount: 3,
			wantErr:   nil,
		},
		{
			name:   "ページネーション",
			offset: 1,
			limit:  2,
			setupFunc: func(r *MorningCallRepository) {
				for i := 0; i < 5; i++ {
					mc := createTestMorningCall(
						fmt.Sprintf("mc%d", i), "user1", "user2",
						time.Now().Add(time.Duration(i)*time.Hour),
						valueobject.MorningCallStatusScheduled,
					)
					r.morningCalls[mc.ID] = mc
				}
			},
			wantCount: 2,
			wantErr:   nil,
		},
		{
			name:      "範囲外のoffset",
			offset:    10,
			limit:     5,
			setupFunc: func(r *MorningCallRepository) {},
			wantCount: 0,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMorningCallRepository()
			tt.setupFunc(repo)

			got, err := repo.FindAll(context.Background(), tt.offset, tt.limit)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindAll() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(got) != tt.wantCount {
					t.Errorf("FindAll() returned %d items, want %d", len(got), tt.wantCount)
				}

				// IDでソートされていることを確認
				sorted := sort.SliceIsSorted(got, func(i, j int) bool {
					return got[i].ID < got[j].ID
				})
				if !sorted && len(got) > 1 {
					t.Error("MorningCalls are not sorted by ID")
				}
			}
		})
	}
}

func TestMorningCallRepository_ConcurrentAccess(t *testing.T) {
	repo := NewMorningCallRepository()
	ctx := context.Background()

	// 並行書き込みテスト
	t.Run("concurrent writes", func(t *testing.T) {
		const numGoroutines = 10
		const itemsPerGoroutine = 10

		errChan := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				var lastErr error
				for j := 0; j < itemsPerGoroutine; j++ {
					mc := createTestMorningCall(
						fmt.Sprintf("g%d_mc%d", goroutineID, j),
						"user1", "user2",
						time.Now().Add(time.Duration(j)*time.Hour),
						valueobject.MorningCallStatusScheduled,
					)
					if err := repo.Create(ctx, mc); err != nil {
						lastErr = err
					}
				}
				errChan <- lastErr
			}(i)
		}

		for i := 0; i < numGoroutines; i++ {
			if err := <-errChan; err != nil {
				t.Errorf("Concurrent write error: %v", err)
			}
		}

		count, _ := repo.Count(ctx)
		expectedCount := numGoroutines * itemsPerGoroutine
		if count != expectedCount {
			t.Errorf("Expected %d items, got %d", expectedCount, count)
		}
	})

	// 並行読み書きテスト
	t.Run("concurrent read and write", func(t *testing.T) {
		repo := NewMorningCallRepository()

		// 初期データを作成
		for i := 0; i < 10; i++ {
			mc := createTestMorningCall(
				fmt.Sprintf("initial_mc%d", i),
				"user1", "user2",
				time.Now().Add(time.Duration(i)*time.Hour),
				valueobject.MorningCallStatusScheduled,
			)
			if err := repo.Create(ctx, mc); err != nil {
				t.Fatalf("Failed to create initial data: %v", err)
			}
		}

		done := make(chan bool)

		// 読み込みゴルーチン
		go func() {
			for i := 0; i < 100; i++ {
				_, err := repo.FindAll(ctx, 0, 100)
				if err != nil {
					t.Errorf("FindAll failed: %v", err)
				}
				_, err = repo.Count(ctx)
				if err != nil {
					t.Errorf("Count failed: %v", err)
				}
				_, err = repo.FindBySenderID(ctx, "user1", 0, 10)
				if err != nil {
					t.Errorf("FindBySenderID failed: %v", err)
				}
			}
			done <- true
		}()

		// 書き込みゴルーチン
		go func() {
			for i := 0; i < 10; i++ {
				mc := createTestMorningCall(
					fmt.Sprintf("concurrent_mc%d", i),
					"user1", "user2",
					time.Now().Add(time.Duration(i)*time.Hour),
					valueobject.MorningCallStatusScheduled,
				)
				if err := repo.Create(ctx, mc); err != nil {
					t.Errorf("Create failed in concurrent write: %v", err)
				}
			}
			done <- true
		}()

		// 両方のゴルーチンが完了するまで待つ
		<-done
		<-done

		// データの整合性を確認
		count, _ := repo.Count(ctx)
		if count != 20 {
			t.Errorf("Expected 20 items after concurrent operations, got %d", count)
		}
	})
}

// Helper function
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
