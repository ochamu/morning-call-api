package morning_call

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
	"github.com/ochamu/morning-call-api/internal/infrastructure/memory"
)

func TestNewUpdateUseCase(t *testing.T) {
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	uc := NewUpdateUseCase(morningCallRepo, userRepo)

	if uc == nil {
		t.Fatal("NewUpdateUseCase returned nil")
	}
	if uc.morningCallRepo == nil {
		t.Error("morningCallRepo is nil")
	}
	if uc.userRepo == nil {
		t.Error("userRepo is nil")
	}
}

func TestUpdateUseCase_Execute(t *testing.T) {
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

	// テスト用のモーニングコールを作成
	futureTime := time.Now().Add(24 * time.Hour)
	existingCall := &entity.MorningCall{
		ID:            "mc1",
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: futureTime,
		Message:       "おはよう！",
		Status:        valueobject.MorningCallStatusScheduled,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := morningCallRepo.Create(ctx, existingCall); err != nil {
		t.Fatalf("failed to create morning call: %v", err)
	}

	// 配信済みのモーニングコール（更新不可）
	deliveredCall := &entity.MorningCall{
		ID:            "mc2",
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: time.Now().Add(-1 * time.Hour),
		Message:       "配信済み",
		Status:        valueobject.MorningCallStatusDelivered,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := morningCallRepo.Create(ctx, deliveredCall); err != nil {
		t.Fatalf("failed to create delivered morning call: %v", err)
	}

	// 時刻変更用の値
	newTime := futureTime.Add(2 * time.Hour)
	newMessage := "更新されたメッセージ"

	tests := []struct {
		name    string
		input   UpdateInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "時刻のみ更新成功",
			input: UpdateInput{
				ID:            existingCall.ID,
				SenderID:      user1.ID,
				ScheduledTime: &newTime,
				Message:       nil,
			},
			wantErr: false,
		},
		{
			name: "メッセージのみ更新成功",
			input: UpdateInput{
				ID:            existingCall.ID,
				SenderID:      user1.ID,
				ScheduledTime: nil,
				Message:       &newMessage,
			},
			wantErr: false,
		},
		{
			name: "時刻とメッセージ両方更新成功",
			input: UpdateInput{
				ID:            existingCall.ID,
				SenderID:      user1.ID,
				ScheduledTime: &newTime,
				Message:       &newMessage,
			},
			wantErr: false,
		},
		{
			name: "IDが空",
			input: UpdateInput{
				ID:            "",
				SenderID:      user1.ID,
				ScheduledTime: &newTime,
				Message:       nil,
			},
			wantErr: true,
			errMsg:  "モーニングコールIDは必須です",
		},
		{
			name: "送信者IDが空",
			input: UpdateInput{
				ID:            existingCall.ID,
				SenderID:      "",
				ScheduledTime: &newTime,
				Message:       nil,
			},
			wantErr: true,
			errMsg:  "送信者IDは必須です",
		},
		{
			name: "更新項目が未指定",
			input: UpdateInput{
				ID:            existingCall.ID,
				SenderID:      user1.ID,
				ScheduledTime: nil,
				Message:       nil,
			},
			wantErr: true,
			errMsg:  "更新する項目を指定してください",
		},
		{
			name: "存在しないモーニングコール",
			input: UpdateInput{
				ID:            "nonexistent",
				SenderID:      user1.ID,
				ScheduledTime: &newTime,
				Message:       nil,
			},
			wantErr: true,
			errMsg:  "モーニングコールが見つかりません",
		},
		{
			name: "送信者以外による更新",
			input: UpdateInput{
				ID:            existingCall.ID,
				SenderID:      user2.ID,
				ScheduledTime: &newTime,
				Message:       nil,
			},
			wantErr: true,
			errMsg:  "送信者のみがモーニングコールを更新できます",
		},
		{
			name: "配信済みモーニングコールの更新",
			input: UpdateInput{
				ID:            deliveredCall.ID,
				SenderID:      user1.ID,
				ScheduledTime: &newTime,
				Message:       nil,
			},
			wantErr: true,
			errMsg:  "スケジュール済みのモーニングコールのみ更新できます",
		},
		{
			name: "過去の時刻への更新",
			input: UpdateInput{
				ID:            existingCall.ID,
				SenderID:      user1.ID,
				ScheduledTime: func() *time.Time { t := time.Now().Add(-1 * time.Hour); return &t }(),
				Message:       nil,
			},
			wantErr: true,
			errMsg:  "アラーム時刻は現在時刻より後である必要があります",
		},
		{
			name: "30日以上先の時刻への更新",
			input: UpdateInput{
				ID:            existingCall.ID,
				SenderID:      user1.ID,
				ScheduledTime: func() *time.Time { t := time.Now().Add(31 * 24 * time.Hour); return &t }(),
				Message:       nil,
			},
			wantErr: true,
			errMsg:  "アラーム時刻は30日以内で設定してください",
		},
		{
			name: "長すぎるメッセージへの更新",
			input: UpdateInput{
				ID:       existingCall.ID,
				SenderID: user1.ID,
				Message:  func() *string { s := strings.Repeat("あ", 501); return &s }(),
			},
			wantErr: true,
			errMsg:  "メッセージは500文字以内で入力してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUpdateUseCase(morningCallRepo, userRepo)
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
				} else {
					if output.MorningCall == nil {
						t.Error("MorningCall is nil")
					} else {
						mc := output.MorningCall
						if mc.ID != tt.input.ID {
							t.Errorf("MorningCall.ID = %v, want %v", mc.ID, tt.input.ID)
						}
						if tt.input.ScheduledTime != nil && !mc.ScheduledTime.Equal(*tt.input.ScheduledTime) {
							t.Errorf("MorningCall.ScheduledTime was not updated")
						}
						if tt.input.Message != nil && mc.Message != *tt.input.Message {
							t.Errorf("MorningCall.Message was not updated")
						}
					}
				}
			}
		})
	}
}

func TestUpdateUseCase_Execute_DuplicateTime(t *testing.T) {
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

	// 既存のモーニングコール1
	scheduledTime1 := time.Now().Add(24 * time.Hour)
	existingCall1 := &entity.MorningCall{
		ID:            "mc1",
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: scheduledTime1,
		Message:       "モーニングコール1",
		Status:        valueobject.MorningCallStatusScheduled,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := morningCallRepo.Create(ctx, existingCall1); err != nil {
		t.Fatalf("failed to create morning call 1: %v", err)
	}

	// 既存のモーニングコール2
	scheduledTime2 := scheduledTime1.Add(2 * time.Hour)
	existingCall2 := &entity.MorningCall{
		ID:            "mc2",
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: scheduledTime2,
		Message:       "モーニングコール2",
		Status:        valueobject.MorningCallStatusScheduled,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := morningCallRepo.Create(ctx, existingCall2); err != nil {
		t.Fatalf("failed to create morning call 2: %v", err)
	}

	uc := NewUpdateUseCase(morningCallRepo, userRepo)

	// モーニングコール2を1と同じ時刻付近に更新しようとする（30秒差）
	duplicateTime := scheduledTime1.Add(30 * time.Second)
	_, err := uc.Execute(ctx, UpdateInput{
		ID:            existingCall2.ID,
		SenderID:      user1.ID,
		ScheduledTime: &duplicateTime,
		Message:       nil,
	})

	if err == nil {
		t.Error("expected error for duplicate time but got nil")
	} else if !strings.Contains(err.Error(), "同じ時刻付近に既にモーニングコールが設定されています") {
		t.Errorf("error message = %v, want contains '同じ時刻付近に既にモーニングコールが設定されています'", err.Error())
	}

	// 1分以上離れた時刻なら成功する
	validTime := scheduledTime1.Add(90 * time.Second)
	output, err := uc.Execute(ctx, UpdateInput{
		ID:            existingCall2.ID,
		SenderID:      user1.ID,
		ScheduledTime: &validTime,
		Message:       nil,
	})

	if err != nil {
		t.Errorf("unexpected error for valid time update: %v", err)
	}
	if output == nil || output.MorningCall == nil {
		t.Error("expected successful update but got nil output")
	}
}

func TestUpdateUseCase_Execute_SelfUpdate(t *testing.T) {
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

	// モーニングコールを作成
	scheduledTime := time.Now().Add(24 * time.Hour)
	morningCall := &entity.MorningCall{
		ID:            "mc1",
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: scheduledTime,
		Message:       "元のメッセージ",
		Status:        valueobject.MorningCallStatusScheduled,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := morningCallRepo.Create(ctx, morningCall); err != nil {
		t.Fatalf("failed to create morning call: %v", err)
	}

	uc := NewUpdateUseCase(morningCallRepo, userRepo)

	// 同じ時刻への更新（自身なので重複チェックに引っかからない）
	sameTime := scheduledTime
	output, err := uc.Execute(ctx, UpdateInput{
		ID:            morningCall.ID,
		SenderID:      user1.ID,
		ScheduledTime: &sameTime,
		Message:       nil,
	})

	if err != nil {
		t.Errorf("unexpected error for self time update: %v", err)
	}
	if output == nil || output.MorningCall == nil {
		t.Error("expected successful update but got nil output")
	}
}
