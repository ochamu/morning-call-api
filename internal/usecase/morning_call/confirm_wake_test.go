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

func TestNewConfirmWakeUseCase(t *testing.T) {
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	uc := NewConfirmWakeUseCase(morningCallRepo, userRepo)

	if uc == nil {
		t.Fatal("NewConfirmWakeUseCase returned nil")
	}
	if uc.morningCallRepo == nil {
		t.Error("morningCallRepo is nil")
	}
	if uc.userRepo == nil {
		t.Error("userRepo is nil")
	}
}

func TestConfirmWakeUseCase_Execute_InputValidation(t *testing.T) {
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
		input   ConfirmWakeInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "モーニングコールIDが空",
			input: ConfirmWakeInput{
				MorningCallID: "",
				ReceiverID:    user1.ID,
			},
			wantErr: true,
			errMsg:  "モーニングコールIDは必須です",
		},
		{
			name: "受信者IDが空",
			input: ConfirmWakeInput{
				MorningCallID: "mc1",
				ReceiverID:    "",
			},
			wantErr: true,
			errMsg:  "受信者IDは必須です",
		},
		{
			name: "存在しない受信者",
			input: ConfirmWakeInput{
				MorningCallID: "mc1",
				ReceiverID:    "nonexistent",
			},
			wantErr: true,
			errMsg:  "受信者が見つかりません",
		},
		{
			name: "存在しないモーニングコール",
			input: ConfirmWakeInput{
				MorningCallID: "nonexistent",
				ReceiverID:    user1.ID,
			},
			wantErr: true,
			errMsg:  "モーニングコールが見つかりません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewConfirmWakeUseCase(morningCallRepo, userRepo)
			output, err := uc.Execute(ctx, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message = %v, want contains %v", err.Error(), tt.errMsg)
				}
				if output != nil {
					t.Error("expected output to be nil on error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if output == nil {
					t.Error("expected output but got nil")
				}
			}
		})
	}
}

func TestConfirmWakeUseCase_Execute_Authorization(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	// テスト用ユーザーを作成
	sender := &entity.User{
		ID:           "sender",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	receiver := &entity.User{
		ID:           "receiver",
		Username:     "bob",
		Email:        "bob@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	other := &entity.User{
		ID:           "other",
		Username:     "charlie",
		Email:        "charlie@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// ユーザーをリポジトリに追加
	if err := userRepo.Create(ctx, sender); err != nil {
		t.Fatalf("failed to create sender: %v", err)
	}
	if err := userRepo.Create(ctx, receiver); err != nil {
		t.Fatalf("failed to create receiver: %v", err)
	}
	if err := userRepo.Create(ctx, other); err != nil {
		t.Fatalf("failed to create other: %v", err)
	}

	// 配信済みのモーニングコールを作成
	morningCall := &entity.MorningCall{
		ID:            "mc1",
		SenderID:      sender.ID,
		ReceiverID:    receiver.ID,
		ScheduledTime: time.Now().Add(-1 * time.Hour),
		Message:       "おはよう！",
		Status:        valueobject.MorningCallStatusDelivered,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := morningCallRepo.Create(ctx, morningCall); err != nil {
		t.Fatalf("failed to create morning call: %v", err)
	}

	uc := NewConfirmWakeUseCase(morningCallRepo, userRepo)

	// 送信者による起床確認（失敗すべき）
	output, err := uc.Execute(ctx, ConfirmWakeInput{
		MorningCallID: morningCall.ID,
		ReceiverID:    sender.ID,
	})
	if err == nil {
		t.Error("sender should not be able to confirm wake")
	} else if !strings.Contains(err.Error(), "受信者のみが起床確認を行えます") {
		t.Errorf("unexpected error message: %v", err.Error())
	}
	if output != nil {
		t.Error("expected output to be nil on error")
	}

	// 無関係なユーザーによる起床確認（失敗すべき）
	output2, err := uc.Execute(ctx, ConfirmWakeInput{
		MorningCallID: morningCall.ID,
		ReceiverID:    other.ID,
	})
	if err == nil {
		t.Error("unrelated user should not be able to confirm wake")
	} else if !strings.Contains(err.Error(), "受信者のみが起床確認を行えます") {
		t.Errorf("unexpected error message: %v", err.Error())
	}
	if output2 != nil {
		t.Error("expected output to be nil on error")
	}

	// 受信者による起床確認（成功すべき）
	output3, err := uc.Execute(ctx, ConfirmWakeInput{
		MorningCallID: morningCall.ID,
		ReceiverID:    receiver.ID,
	})
	if err != nil {
		t.Errorf("receiver should be able to confirm wake: %v", err)
	}
	if output3 == nil {
		t.Error("expected output but got nil")
	} else {
		if output3.MorningCall == nil {
			t.Error("expected MorningCall but got nil")
		} else if output3.MorningCall.Status != valueobject.MorningCallStatusConfirmed {
			t.Errorf("expected status to be Confirmed, got %v", output3.MorningCall.Status)
		}
		if output3.ConfirmedAt.IsZero() {
			t.Error("expected ConfirmedAt to be set")
		}
	}
}

func TestConfirmWakeUseCase_Execute_StatusTransitions(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	// テスト用ユーザーを作成
	sender := &entity.User{
		ID:           "sender",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	receiver := &entity.User{
		ID:           "receiver",
		Username:     "bob",
		Email:        "bob@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// ユーザーをリポジトリに追加
	if err := userRepo.Create(ctx, sender); err != nil {
		t.Fatalf("failed to create sender: %v", err)
	}
	if err := userRepo.Create(ctx, receiver); err != nil {
		t.Fatalf("failed to create receiver: %v", err)
	}

	// 各ステータスのモーニングコールを作成してテスト
	testCases := []struct {
		name       string
		status     valueobject.MorningCallStatus
		canConfirm bool
		errMsg     string
	}{
		{
			name:       "スケジュール済み",
			status:     valueobject.MorningCallStatusScheduled,
			canConfirm: false,
			errMsg:     "まだ配信されていないモーニングコールは起床確認できません",
		},
		{
			name:       "配信済み",
			status:     valueobject.MorningCallStatusDelivered,
			canConfirm: true,
		},
		{
			name:       "確認済み",
			status:     valueobject.MorningCallStatusConfirmed,
			canConfirm: false,
			errMsg:     "すでに起床確認済みです",
		},
		{
			name:       "キャンセル済み",
			status:     valueobject.MorningCallStatusCancelled,
			canConfirm: false,
			errMsg:     "キャンセル済みのモーニングコールは起床確認できません",
		},
		{
			name:       "期限切れ",
			status:     valueobject.MorningCallStatusExpired,
			canConfirm: false,
			errMsg:     "期限切れのモーニングコールは起床確認できません",
		},
	}

	uc := NewConfirmWakeUseCase(morningCallRepo, userRepo)

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モーニングコールを作成
			morningCall := &entity.MorningCall{
				ID:            fmt.Sprintf("mc_%d", i),
				SenderID:      sender.ID,
				ReceiverID:    receiver.ID,
				ScheduledTime: time.Now().Add(-1 * time.Hour),
				Message:       fmt.Sprintf("テスト%d", i),
				Status:        tc.status,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			if err := morningCallRepo.Create(ctx, morningCall); err != nil {
				t.Fatalf("failed to create morning call: %v", err)
			}

			// 起床確認を試みる
			output, err := uc.Execute(ctx, ConfirmWakeInput{
				MorningCallID: morningCall.ID,
				ReceiverID:    receiver.ID,
			})

			if tc.canConfirm {
				if err != nil {
					t.Errorf("should be able to confirm wake for %s status: %v", tc.name, err)
				}
				if output == nil {
					t.Error("expected output but got nil")
				} else {
					if output.MorningCall == nil {
						t.Error("expected MorningCall but got nil")
					} else if output.MorningCall.Status != valueobject.MorningCallStatusConfirmed {
						t.Errorf("expected status to be Confirmed, got %v", output.MorningCall.Status)
					}
					if output.ConfirmedAt.IsZero() {
						t.Error("expected ConfirmedAt to be set")
					}
				}

				// リポジトリから再取得して確認
				updatedMC, err := morningCallRepo.FindByID(ctx, morningCall.ID)
				if err != nil {
					t.Errorf("failed to get updated morning call: %v", err)
				}
				if updatedMC.Status != valueobject.MorningCallStatusConfirmed {
					t.Errorf("status not updated in repository, got %v", updatedMC.Status)
				}
			} else {
				if err == nil {
					t.Errorf("should not be able to confirm wake for %s status", tc.name)
				} else if tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("unexpected error message: got %v, want contains %v", err.Error(), tc.errMsg)
				}
				if output != nil {
					t.Error("expected output to be nil on error")
				}

				// ステータスが変更されていないことを確認
				unchangedMC, err := morningCallRepo.FindByID(ctx, morningCall.ID)
				if err != nil {
					t.Errorf("failed to get morning call: %v", err)
				}
				if unchangedMC.Status != tc.status {
					t.Errorf("status should not change, expected %v, got %v", tc.status, unchangedMC.Status)
				}
			}
		})
	}
}

func TestConfirmWakeUseCase_Execute_SuccessFlow(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	// テスト用ユーザーを作成
	sender := &entity.User{
		ID:           "sender",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	receiver := &entity.User{
		ID:           "receiver",
		Username:     "bob",
		Email:        "bob@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// ユーザーをリポジトリに追加
	if err := userRepo.Create(ctx, sender); err != nil {
		t.Fatalf("failed to create sender: %v", err)
	}
	if err := userRepo.Create(ctx, receiver); err != nil {
		t.Fatalf("failed to create receiver: %v", err)
	}

	// 配信済みのモーニングコールを作成
	morningCall := &entity.MorningCall{
		ID:            "mc1",
		SenderID:      sender.ID,
		ReceiverID:    receiver.ID,
		ScheduledTime: time.Now().Add(-2 * time.Hour),
		Message:       "おはようございます！今日も一日頑張りましょう！",
		Status:        valueobject.MorningCallStatusDelivered,
		CreatedAt:     time.Now().Add(-3 * time.Hour),
		UpdatedAt:     time.Now().Add(-2 * time.Hour),
	}
	if err := morningCallRepo.Create(ctx, morningCall); err != nil {
		t.Fatalf("failed to create morning call: %v", err)
	}

	uc := NewConfirmWakeUseCase(morningCallRepo, userRepo)

	// 起床確認を実行
	beforeConfirm := time.Now()
	output, err := uc.Execute(ctx, ConfirmWakeInput{
		MorningCallID: morningCall.ID,
		ReceiverID:    receiver.ID,
	})
	afterConfirm := time.Now()

	// エラーチェック
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 出力の検証
	if output == nil {
		t.Fatal("expected output but got nil")
	}
	if output.MorningCall == nil {
		t.Fatal("expected MorningCall but got nil")
	}

	// モーニングコールの内容検証
	mc := output.MorningCall
	if mc.ID != morningCall.ID {
		t.Errorf("ID = %v, want %v", mc.ID, morningCall.ID)
	}
	if mc.SenderID != sender.ID {
		t.Errorf("SenderID = %v, want %v", mc.SenderID, sender.ID)
	}
	if mc.ReceiverID != receiver.ID {
		t.Errorf("ReceiverID = %v, want %v", mc.ReceiverID, receiver.ID)
	}
	if mc.Status != valueobject.MorningCallStatusConfirmed {
		t.Errorf("Status = %v, want %v", mc.Status, valueobject.MorningCallStatusConfirmed)
	}
	if mc.Message != morningCall.Message {
		t.Errorf("Message = %v, want %v", mc.Message, morningCall.Message)
	}

	// 確認時刻の検証
	if output.ConfirmedAt.Before(beforeConfirm) || output.ConfirmedAt.After(afterConfirm) {
		t.Errorf("ConfirmedAt = %v, want between %v and %v", output.ConfirmedAt, beforeConfirm, afterConfirm)
	}
	if !mc.UpdatedAt.Equal(output.ConfirmedAt) {
		t.Errorf("UpdatedAt should equal ConfirmedAt, got UpdatedAt = %v, ConfirmedAt = %v", mc.UpdatedAt, output.ConfirmedAt)
	}

	// リポジトリから再取得して永続化を確認
	persistedMC, err := morningCallRepo.FindByID(ctx, morningCall.ID)
	if err != nil {
		t.Fatalf("failed to get persisted morning call: %v", err)
	}
	if persistedMC.Status != valueobject.MorningCallStatusConfirmed {
		t.Errorf("persisted Status = %v, want %v", persistedMC.Status, valueobject.MorningCallStatusConfirmed)
	}
}

func TestConfirmWakeUseCase_Execute_MultipleConfirmAttempts(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	// テスト用ユーザーを作成
	sender := &entity.User{
		ID:           "sender",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	receiver := &entity.User{
		ID:           "receiver",
		Username:     "bob",
		Email:        "bob@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// ユーザーをリポジトリに追加
	if err := userRepo.Create(ctx, sender); err != nil {
		t.Fatalf("failed to create sender: %v", err)
	}
	if err := userRepo.Create(ctx, receiver); err != nil {
		t.Fatalf("failed to create receiver: %v", err)
	}

	// 配信済みのモーニングコールを作成
	morningCall := &entity.MorningCall{
		ID:            "mc1",
		SenderID:      sender.ID,
		ReceiverID:    receiver.ID,
		ScheduledTime: time.Now().Add(-1 * time.Hour),
		Message:       "おはよう！",
		Status:        valueobject.MorningCallStatusDelivered,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := morningCallRepo.Create(ctx, morningCall); err != nil {
		t.Fatalf("failed to create morning call: %v", err)
	}

	uc := NewConfirmWakeUseCase(morningCallRepo, userRepo)

	// 1回目の起床確認（成功すべき）
	output1, err := uc.Execute(ctx, ConfirmWakeInput{
		MorningCallID: morningCall.ID,
		ReceiverID:    receiver.ID,
	})
	if err != nil {
		t.Errorf("first confirm should succeed: %v", err)
	}
	if output1 == nil || output1.MorningCall == nil {
		t.Error("expected output with MorningCall")
	} else if output1.MorningCall.Status != valueobject.MorningCallStatusConfirmed {
		t.Errorf("expected status to be Confirmed, got %v", output1.MorningCall.Status)
	}

	// 2回目の起床確認（失敗すべき - すでに確認済み）
	output2, err := uc.Execute(ctx, ConfirmWakeInput{
		MorningCallID: morningCall.ID,
		ReceiverID:    receiver.ID,
	})
	if err == nil {
		t.Error("second confirm should fail")
	} else if !strings.Contains(err.Error(), "すでに起床確認済みです") {
		t.Errorf("unexpected error message: %v", err.Error())
	}
	if output2 != nil {
		t.Error("expected output to be nil on error")
	}
}

func TestConfirmWakeUseCase_Execute_MultipleMorningCalls(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	// テスト用ユーザーを作成
	sender := &entity.User{
		ID:           "sender",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	receiver := &entity.User{
		ID:           "receiver",
		Username:     "bob",
		Email:        "bob@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// ユーザーをリポジトリに追加
	if err := userRepo.Create(ctx, sender); err != nil {
		t.Fatalf("failed to create sender: %v", err)
	}
	if err := userRepo.Create(ctx, receiver); err != nil {
		t.Fatalf("failed to create receiver: %v", err)
	}

	// 複数の配信済みモーニングコールを作成
	morningCalls := []*entity.MorningCall{
		{
			ID:            "mc1",
			SenderID:      sender.ID,
			ReceiverID:    receiver.ID,
			ScheduledTime: time.Now().Add(-3 * time.Hour),
			Message:       "朝6時のアラーム",
			Status:        valueobject.MorningCallStatusDelivered,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            "mc2",
			SenderID:      sender.ID,
			ReceiverID:    receiver.ID,
			ScheduledTime: time.Now().Add(-2 * time.Hour),
			Message:       "朝7時のアラーム",
			Status:        valueobject.MorningCallStatusDelivered,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            "mc3",
			SenderID:      sender.ID,
			ReceiverID:    receiver.ID,
			ScheduledTime: time.Now().Add(-1 * time.Hour),
			Message:       "朝8時のアラーム",
			Status:        valueobject.MorningCallStatusDelivered,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	for _, mc := range morningCalls {
		if err := morningCallRepo.Create(ctx, mc); err != nil {
			t.Fatalf("failed to create morning call %s: %v", mc.ID, err)
		}
	}

	uc := NewConfirmWakeUseCase(morningCallRepo, userRepo)

	// 各モーニングコールを個別に確認
	for _, mc := range morningCalls {
		output, err := uc.Execute(ctx, ConfirmWakeInput{
			MorningCallID: mc.ID,
			ReceiverID:    receiver.ID,
		})

		if err != nil {
			t.Errorf("failed to confirm morning call %s: %v", mc.ID, err)
			continue
		}

		if output == nil || output.MorningCall == nil {
			t.Errorf("expected output for morning call %s", mc.ID)
			continue
		}

		if output.MorningCall.Status != valueobject.MorningCallStatusConfirmed {
			t.Errorf("morning call %s: expected status Confirmed, got %v", mc.ID, output.MorningCall.Status)
		}

		// 他のモーニングコールが影響を受けていないことを確認
		for _, otherMC := range morningCalls {
			if otherMC.ID == mc.ID {
				continue
			}

			checkMC, err := morningCallRepo.FindByID(ctx, otherMC.ID)
			if err != nil {
				t.Errorf("failed to get morning call %s: %v", otherMC.ID, err)
				continue
			}

			// まだ確認していないものは配信済みのまま、確認済みのものは確認済みのまま
			if checkMC.Status != valueobject.MorningCallStatusDelivered &&
				checkMC.Status != valueobject.MorningCallStatusConfirmed {
				t.Errorf("morning call %s has unexpected status %v", otherMC.ID, checkMC.Status)
			}
		}
	}

	// すべてが確認済みになったことを最終確認
	for _, mc := range morningCalls {
		checkMC, err := morningCallRepo.FindByID(ctx, mc.ID)
		if err != nil {
			t.Errorf("failed to get morning call %s: %v", mc.ID, err)
			continue
		}
		if checkMC.Status != valueobject.MorningCallStatusConfirmed {
			t.Errorf("morning call %s should be confirmed, got %v", mc.ID, checkMC.Status)
		}
	}
}
