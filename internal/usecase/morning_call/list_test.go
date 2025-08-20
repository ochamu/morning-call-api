package morning_call

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
	"github.com/ochamu/morning-call-api/internal/infrastructure/memory"
)

func TestNewListUseCase(t *testing.T) {
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	uc := NewListUseCase(morningCallRepo, userRepo)

	if uc == nil {
		t.Fatal("NewListUseCase returned nil")
	}
	if uc.morningCallRepo == nil {
		t.Error("morningCallRepo is nil")
	}
	if uc.userRepo == nil {
		t.Error("userRepo is nil")
	}
}

func TestListUseCase_Execute_InputValidation(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	// テスト用ユーザーを作成
	user1 := &entity.User{
		ID:           "user1",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}

	tests := []struct {
		name    string
		input   ListInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "ユーザーIDが空",
			input: ListInput{
				UserID:   "",
				ListType: ListTypeSent,
				Limit:    20,
			},
			wantErr: true,
			errMsg:  "ユーザーIDは必須です",
		},
		{
			name: "無効なリストタイプ",
			input: ListInput{
				UserID:   user1.ID,
				ListType: "invalid",
				Limit:    20,
			},
			wantErr: true,
			errMsg:  "一覧タイプは'sent'または'received'を指定してください",
		},
		{
			name: "存在しないユーザー",
			input: ListInput{
				UserID:   "nonexistent",
				ListType: ListTypeSent,
				Limit:    20,
			},
			wantErr: true,
			errMsg:  "ユーザーが見つかりません",
		},
		{
			name: "正常な入力（送信）",
			input: ListInput{
				UserID:   user1.ID,
				ListType: ListTypeSent,
				Limit:    20,
			},
			wantErr: false,
		},
		{
			name: "正常な入力（受信）",
			input: ListInput{
				UserID:   user1.ID,
				ListType: ListTypeReceived,
				Limit:    20,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewListUseCase(morningCallRepo, userRepo)
			output, err := uc.Execute(ctx, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message = %v, want contains %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if output == nil {
					t.Error("output is nil")
				}
			}
		})
	}
}

func TestListUseCase_Execute_SentList(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	// テスト用ユーザーを作成
	user1 := &entity.User{
		ID:           "user1",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	user2 := &entity.User{
		ID:           "user2",
		Username:     "bob",
		Email:        "bob@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	user3 := &entity.User{
		ID:           "user3",
		Username:     "charlie",
		Email:        "charlie@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// ユーザーをリポジトリに追加
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	if err := userRepo.Create(ctx, user2); err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}
	if err := userRepo.Create(ctx, user3); err != nil {
		t.Fatalf("failed to create user3: %v", err)
	}

	// user1が送信したモーニングコールを複数作成
	now := time.Now()
	for i := 0; i < 5; i++ {
		mc := &entity.MorningCall{
			ID:            fmt.Sprintf("mc_sent_%d", i),
			SenderID:      user1.ID,
			ReceiverID:    user2.ID,
			ScheduledTime: now.Add(time.Duration(i+1) * time.Hour),
			Message:       fmt.Sprintf("送信メッセージ%d", i),
			Status:        valueobject.MorningCallStatusScheduled,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := morningCallRepo.Create(ctx, mc); err != nil {
			t.Fatalf("failed to create morning call %d: %v", i, err)
		}
	}

	// user2が送信したモーニングコール（user1の送信リストには含まれない）
	for i := 0; i < 3; i++ {
		mc := &entity.MorningCall{
			ID:            fmt.Sprintf("mc_other_%d", i),
			SenderID:      user2.ID,
			ReceiverID:    user1.ID,
			ScheduledTime: now.Add(time.Duration(i+10) * time.Hour),
			Message:       fmt.Sprintf("他ユーザーメッセージ%d", i),
			Status:        valueobject.MorningCallStatusScheduled,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := morningCallRepo.Create(ctx, mc); err != nil {
			t.Fatalf("failed to create other user morning call %d: %v", i, err)
		}
	}

	uc := NewListUseCase(morningCallRepo, userRepo)

	// user1の送信リストを取得
	output, err := uc.Execute(ctx, ListInput{
		UserID:   user1.ID,
		ListType: ListTypeSent,
		Offset:   0,
		Limit:    10,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 結果の検証
	if len(output.MorningCalls) != 5 {
		t.Errorf("expected 5 sent morning calls, got %d", len(output.MorningCalls))
	}

	// すべてのモーニングコールがuser1が送信したものであることを確認
	for _, mc := range output.MorningCalls {
		if mc.SenderID != user1.ID {
			t.Errorf("expected all morning calls to be sent by user1, got sender %s", mc.SenderID)
		}
	}

	if output.TotalCount != 5 {
		t.Errorf("expected total count 5, got %d", output.TotalCount)
	}

	if output.HasNext {
		t.Error("expected HasNext to be false")
	}
}

func TestListUseCase_Execute_ReceivedList(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	// テスト用ユーザーを作成
	user1 := &entity.User{
		ID:           "user1",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	user2 := &entity.User{
		ID:           "user2",
		Username:     "bob",
		Email:        "bob@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// ユーザーをリポジトリに追加
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	if err := userRepo.Create(ctx, user2); err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	// user1が受信するモーニングコールを複数作成
	now := time.Now()
	for i := 0; i < 7; i++ {
		mc := &entity.MorningCall{
			ID:            fmt.Sprintf("mc_received_%d", i),
			SenderID:      user2.ID,
			ReceiverID:    user1.ID,
			ScheduledTime: now.Add(time.Duration(i+1) * time.Hour),
			Message:       fmt.Sprintf("受信メッセージ%d", i),
			Status:        valueobject.MorningCallStatusScheduled,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := morningCallRepo.Create(ctx, mc); err != nil {
			t.Fatalf("failed to create morning call %d: %v", i, err)
		}
	}

	// user2が受信するモーニングコール（user1の受信リストには含まれない）
	for i := 0; i < 3; i++ {
		mc := &entity.MorningCall{
			ID:            fmt.Sprintf("mc_other_%d", i),
			SenderID:      user1.ID,
			ReceiverID:    user2.ID,
			ScheduledTime: now.Add(time.Duration(i+10) * time.Hour),
			Message:       fmt.Sprintf("他ユーザー受信メッセージ%d", i),
			Status:        valueobject.MorningCallStatusScheduled,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := morningCallRepo.Create(ctx, mc); err != nil {
			t.Fatalf("failed to create other user morning call %d: %v", i, err)
		}
	}

	uc := NewListUseCase(morningCallRepo, userRepo)

	// user1の受信リストを取得
	output, err := uc.Execute(ctx, ListInput{
		UserID:   user1.ID,
		ListType: ListTypeReceived,
		Offset:   0,
		Limit:    10,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 結果の検証
	if len(output.MorningCalls) != 7 {
		t.Errorf("expected 7 received morning calls, got %d", len(output.MorningCalls))
	}

	// すべてのモーニングコールがuser1が受信したものであることを確認
	for _, mc := range output.MorningCalls {
		if mc.ReceiverID != user1.ID {
			t.Errorf("expected all morning calls to be received by user1, got receiver %s", mc.ReceiverID)
		}
	}

	if output.TotalCount != 7 {
		t.Errorf("expected total count 7, got %d", output.TotalCount)
	}

	if output.HasNext {
		t.Error("expected HasNext to be false")
	}
}

func TestListUseCase_Execute_Pagination(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	// テスト用ユーザーを作成
	user1 := &entity.User{
		ID:           "user1",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	user2 := &entity.User{
		ID:           "user2",
		Username:     "bob",
		Email:        "bob@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// ユーザーをリポジトリに追加
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	if err := userRepo.Create(ctx, user2); err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	// user1が送信したモーニングコールを25個作成
	now := time.Now()
	for i := 0; i < 25; i++ {
		mc := &entity.MorningCall{
			ID:            fmt.Sprintf("mc_%02d", i),
			SenderID:      user1.ID,
			ReceiverID:    user2.ID,
			ScheduledTime: now.Add(time.Duration(i+1) * time.Hour),
			Message:       fmt.Sprintf("メッセージ%d", i),
			Status:        valueobject.MorningCallStatusScheduled,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := morningCallRepo.Create(ctx, mc); err != nil {
			t.Fatalf("failed to create morning call %d: %v", i, err)
		}
	}

	uc := NewListUseCase(morningCallRepo, userRepo)

	// 1ページ目を取得（10件）
	output1, err := uc.Execute(ctx, ListInput{
		UserID:   user1.ID,
		ListType: ListTypeSent,
		Offset:   0,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error on page 1: %v", err)
	}

	if len(output1.MorningCalls) != 10 {
		t.Errorf("expected 10 morning calls on page 1, got %d", len(output1.MorningCalls))
	}
	if output1.TotalCount != 25 {
		t.Errorf("expected total count 25, got %d", output1.TotalCount)
	}
	if !output1.HasNext {
		t.Error("expected HasNext to be true on page 1")
	}

	// 2ページ目を取得（10件）
	output2, err := uc.Execute(ctx, ListInput{
		UserID:   user1.ID,
		ListType: ListTypeSent,
		Offset:   10,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error on page 2: %v", err)
	}

	if len(output2.MorningCalls) != 10 {
		t.Errorf("expected 10 morning calls on page 2, got %d", len(output2.MorningCalls))
	}
	if !output2.HasNext {
		t.Error("expected HasNext to be true on page 2")
	}

	// 3ページ目を取得（5件）
	output3, err := uc.Execute(ctx, ListInput{
		UserID:   user1.ID,
		ListType: ListTypeSent,
		Offset:   20,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error on page 3: %v", err)
	}

	if len(output3.MorningCalls) != 5 {
		t.Errorf("expected 5 morning calls on page 3, got %d", len(output3.MorningCalls))
	}
	if output3.HasNext {
		t.Error("expected HasNext to be false on page 3")
	}

	// 範囲外のページを取得
	output4, err := uc.Execute(ctx, ListInput{
		UserID:   user1.ID,
		ListType: ListTypeSent,
		Offset:   30,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error on out of range page: %v", err)
	}

	if len(output4.MorningCalls) != 0 {
		t.Errorf("expected 0 morning calls for out of range page, got %d", len(output4.MorningCalls))
	}
	if output4.HasNext {
		t.Error("expected HasNext to be false for out of range page")
	}
}

func TestListUseCase_Execute_StatusFilter(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	// テスト用ユーザーを作成
	user1 := &entity.User{
		ID:           "user1",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	user2 := &entity.User{
		ID:           "user2",
		Username:     "bob",
		Email:        "bob@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// ユーザーをリポジトリに追加
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	if err := userRepo.Create(ctx, user2); err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	// 異なるステータスのモーニングコールを作成
	now := time.Now()
	statuses := []valueobject.MorningCallStatus{
		valueobject.MorningCallStatusScheduled,
		valueobject.MorningCallStatusScheduled,
		valueobject.MorningCallStatusScheduled,
		valueobject.MorningCallStatusDelivered,
		valueobject.MorningCallStatusDelivered,
		valueobject.MorningCallStatusConfirmed,
		valueobject.MorningCallStatusCancelled,
		valueobject.MorningCallStatusExpired,
	}

	for i, status := range statuses {
		mc := &entity.MorningCall{
			ID:            fmt.Sprintf("mc_%d", i),
			SenderID:      user1.ID,
			ReceiverID:    user2.ID,
			ScheduledTime: now.Add(time.Duration(i+1) * time.Hour),
			Message:       fmt.Sprintf("メッセージ%d", i),
			Status:        status,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := morningCallRepo.Create(ctx, mc); err != nil {
			t.Fatalf("failed to create morning call %d: %v", i, err)
		}
	}

	uc := NewListUseCase(morningCallRepo, userRepo)

	// スケジュール済みのみでフィルタ
	statusScheduled := valueobject.MorningCallStatusScheduled
	output, err := uc.Execute(ctx, ListInput{
		UserID:   user1.ID,
		ListType: ListTypeSent,
		Status:   &statusScheduled,
		Offset:   0,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.MorningCalls) != 3 {
		t.Errorf("expected 3 scheduled morning calls, got %d", len(output.MorningCalls))
	}

	// すべてスケジュール済みであることを確認
	for _, mc := range output.MorningCalls {
		if mc.Status != valueobject.MorningCallStatusScheduled {
			t.Errorf("expected status Scheduled, got %v", mc.Status)
		}
	}

	// 配信済みのみでフィルタ
	statusDelivered := valueobject.MorningCallStatusDelivered
	output2, err := uc.Execute(ctx, ListInput{
		UserID:   user1.ID,
		ListType: ListTypeSent,
		Status:   &statusDelivered,
		Offset:   0,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output2.MorningCalls) != 2 {
		t.Errorf("expected 2 delivered morning calls, got %d", len(output2.MorningCalls))
	}
}

func TestListUseCase_Execute_TimeRangeFilter(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	// テスト用ユーザーを作成
	user1 := &entity.User{
		ID:           "user1",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	user2 := &entity.User{
		ID:           "user2",
		Username:     "bob",
		Email:        "bob@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// ユーザーをリポジトリに追加
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	if err := userRepo.Create(ctx, user2); err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	// 異なる時刻のモーニングコールを作成
	baseTime := time.Now()
	times := []time.Duration{
		1 * time.Hour,
		5 * time.Hour,
		10 * time.Hour,
		15 * time.Hour,
		20 * time.Hour,
		25 * time.Hour,
		30 * time.Hour,
	}

	for i, duration := range times {
		mc := &entity.MorningCall{
			ID:            fmt.Sprintf("mc_%d", i),
			SenderID:      user1.ID,
			ReceiverID:    user2.ID,
			ScheduledTime: baseTime.Add(duration),
			Message:       fmt.Sprintf("メッセージ%d", i),
			Status:        valueobject.MorningCallStatusScheduled,
			CreatedAt:     baseTime,
			UpdatedAt:     baseTime,
		}
		if err := morningCallRepo.Create(ctx, mc); err != nil {
			t.Fatalf("failed to create morning call %d: %v", i, err)
		}
	}

	uc := NewListUseCase(morningCallRepo, userRepo)

	// 5時間後から20時間後までの範囲でフィルタ
	startTime := baseTime.Add(5 * time.Hour)
	endTime := baseTime.Add(20 * time.Hour)
	output, err := uc.Execute(ctx, ListInput{
		UserID:    user1.ID,
		ListType:  ListTypeSent,
		StartTime: &startTime,
		EndTime:   &endTime,
		Offset:    0,
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 5時間後、10時間後、15時間後、20時間後の4件が含まれるはず
	if len(output.MorningCalls) != 4 {
		t.Errorf("expected 4 morning calls in time range, got %d", len(output.MorningCalls))
	}

	// 時刻が範囲内であることを確認
	for _, mc := range output.MorningCalls {
		if mc.ScheduledTime.Before(startTime) || mc.ScheduledTime.After(endTime) {
			t.Errorf("morning call scheduled at %v is outside range [%v, %v]",
				mc.ScheduledTime, startTime, endTime)
		}
	}

	// 開始時刻が終了時刻より後の場合のエラーテスト
	invalidStartTime := baseTime.Add(25 * time.Hour)
	invalidEndTime := baseTime.Add(5 * time.Hour)
	_, err = uc.Execute(ctx, ListInput{
		UserID:    user1.ID,
		ListType:  ListTypeSent,
		StartTime: &invalidStartTime,
		EndTime:   &invalidEndTime,
		Offset:    0,
		Limit:     10,
	})
	if err == nil {
		t.Error("expected error for invalid time range but got nil")
	}
	if !strings.Contains(err.Error(), "開始時刻は終了時刻より前である必要があります") {
		t.Errorf("unexpected error message: %v", err.Error())
	}
}

func TestListUseCase_Execute_DefaultValues(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	// テスト用ユーザーを作成
	user1 := &entity.User{
		ID:           "user1",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}

	uc := NewListUseCase(morningCallRepo, userRepo)

	// Limitが0の場合、デフォルト値（20）が適用される
	output, err := uc.Execute(ctx, ListInput{
		UserID:   user1.ID,
		ListType: ListTypeSent,
		Limit:    0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output == nil {
		t.Fatal("output is nil")
	}

	// Limitが100を超える場合、100に制限される
	output2, err := uc.Execute(ctx, ListInput{
		UserID:   user1.ID,
		ListType: ListTypeSent,
		Limit:    200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output2 == nil {
		t.Fatal("output is nil")
	}

	// Offsetが負の場合、0に修正される
	output3, err := uc.Execute(ctx, ListInput{
		UserID:   user1.ID,
		ListType: ListTypeSent,
		Offset:   -10,
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output3 == nil {
		t.Fatal("output is nil")
	}
}
