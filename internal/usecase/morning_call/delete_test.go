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

func TestNewDeleteUseCase(t *testing.T) {
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()

	uc := NewDeleteUseCase(morningCallRepo, userRepo)

	if uc == nil {
		t.Fatal("NewDeleteUseCase returned nil")
	}
	if uc.morningCallRepo == nil {
		t.Error("morningCallRepo is nil")
	}
	if uc.userRepo == nil {
		t.Error("userRepo is nil")
	}
}

func TestDeleteUseCase_Execute(t *testing.T) {
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
	scheduledCall := &entity.MorningCall{
		ID:            "mc1",
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: futureTime,
		Message:       "おはよう！",
		Status:        valueobject.MorningCallStatusScheduled,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := morningCallRepo.Create(ctx, scheduledCall); err != nil {
		t.Fatalf("failed to create scheduled morning call: %v", err)
	}

	// キャンセル済みのモーニングコール
	cancelledCall := &entity.MorningCall{
		ID:            "mc2",
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: futureTime.Add(1 * time.Hour),
		Message:       "キャンセル済み",
		Status:        valueobject.MorningCallStatusCancelled,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := morningCallRepo.Create(ctx, cancelledCall); err != nil {
		t.Fatalf("failed to create cancelled morning call: %v", err)
	}

	// 配信済みのモーニングコール（削除不可）
	deliveredCall := &entity.MorningCall{
		ID:            "mc3",
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

	// 確認済みのモーニングコール（削除不可）
	confirmedCall := &entity.MorningCall{
		ID:            "mc4",
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: time.Now().Add(-2 * time.Hour),
		Message:       "確認済み",
		Status:        valueobject.MorningCallStatusConfirmed,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := morningCallRepo.Create(ctx, confirmedCall); err != nil {
		t.Fatalf("failed to create confirmed morning call: %v", err)
	}

	// 期限切れのモーニングコール（削除不可）
	expiredCall := &entity.MorningCall{
		ID:            "mc5",
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: time.Now().Add(-3 * time.Hour),
		Message:       "期限切れ",
		Status:        valueobject.MorningCallStatusExpired,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := morningCallRepo.Create(ctx, expiredCall); err != nil {
		t.Fatalf("failed to create expired morning call: %v", err)
	}

	// 送信者以外による削除テスト用のモーニングコール
	otherUserCall := &entity.MorningCall{
		ID:            "mc6",
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: futureTime.Add(2 * time.Hour),
		Message:       "別ユーザーテスト用",
		Status:        valueobject.MorningCallStatusScheduled,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := morningCallRepo.Create(ctx, otherUserCall); err != nil {
		t.Fatalf("failed to create other user test morning call: %v", err)
	}

	tests := []struct {
		name         string
		input        DeleteInput
		wantErr      bool
		errMsg       string
		checkDeleted bool // 削除されたことを確認するか
	}{
		{
			name: "IDが空",
			input: DeleteInput{
				ID:       "",
				SenderID: user1.ID,
			},
			wantErr: true,
			errMsg:  "モーニングコールIDは必須です",
		},
		{
			name: "送信者IDが空",
			input: DeleteInput{
				ID:       scheduledCall.ID,
				SenderID: "",
			},
			wantErr: true,
			errMsg:  "送信者IDは必須です",
		},
		{
			name: "存在しないモーニングコール",
			input: DeleteInput{
				ID:       "nonexistent",
				SenderID: user1.ID,
			},
			wantErr: true,
			errMsg:  "モーニングコールが見つかりません",
		},
		{
			name: "送信者以外による削除",
			input: DeleteInput{
				ID:       otherUserCall.ID,
				SenderID: user2.ID,
			},
			wantErr: true,
			errMsg:  "送信者のみがモーニングコールを削除できます",
		},
		{
			name: "配信済みモーニングコールの削除",
			input: DeleteInput{
				ID:       deliveredCall.ID,
				SenderID: user1.ID,
			},
			wantErr: true,
			errMsg:  "削除できるのはスケジュール済みまたはキャンセル済みのモーニングコールのみです",
		},
		{
			name: "確認済みモーニングコールの削除",
			input: DeleteInput{
				ID:       confirmedCall.ID,
				SenderID: user1.ID,
			},
			wantErr: true,
			errMsg:  "削除できるのはスケジュール済みまたはキャンセル済みのモーニングコールのみです",
		},
		{
			name: "期限切れモーニングコールの削除",
			input: DeleteInput{
				ID:       expiredCall.ID,
				SenderID: user1.ID,
			},
			wantErr: true,
			errMsg:  "削除できるのはスケジュール済みまたはキャンセル済みのモーニングコールのみです",
		},
		{
			name: "スケジュール済みの削除成功",
			input: DeleteInput{
				ID:       scheduledCall.ID,
				SenderID: user1.ID,
			},
			wantErr:      false,
			checkDeleted: true,
		},
		{
			name: "キャンセル済みの削除成功",
			input: DeleteInput{
				ID:       cancelledCall.ID,
				SenderID: user1.ID,
			},
			wantErr:      false,
			checkDeleted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewDeleteUseCase(morningCallRepo, userRepo)
			err := uc.Execute(ctx, tt.input)

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

				// 削除されたことを確認
				if tt.checkDeleted {
					_, err := morningCallRepo.FindByID(ctx, tt.input.ID)
					if err == nil {
						t.Error("expected morning call to be deleted but still exists")
					}
				}
			}
		})
	}
}

func TestDeleteUseCase_Execute_Authorization(t *testing.T) {
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

	// モーニングコールを作成（user1 -> user2）
	morningCall := &entity.MorningCall{
		ID:            "mc1",
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: time.Now().Add(24 * time.Hour),
		Message:       "テストメッセージ",
		Status:        valueobject.MorningCallStatusScheduled,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := morningCallRepo.Create(ctx, morningCall); err != nil {
		t.Fatalf("failed to create morning call: %v", err)
	}

	uc := NewDeleteUseCase(morningCallRepo, userRepo)

	// 送信者（user1）による削除は成功すべき
	err := uc.Execute(ctx, DeleteInput{
		ID:       morningCall.ID,
		SenderID: user1.ID,
	})
	if err != nil {
		t.Errorf("sender should be able to delete: %v", err)
	}

	// 削除されたことを確認
	_, err = morningCallRepo.FindByID(ctx, morningCall.ID)
	if err == nil {
		t.Error("morning call should be deleted")
	}

	// 新しいモーニングコールを作成してテストを続ける
	morningCall2 := &entity.MorningCall{
		ID:            "mc2",
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: time.Now().Add(25 * time.Hour),
		Message:       "テストメッセージ2",
		Status:        valueobject.MorningCallStatusScheduled,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := morningCallRepo.Create(ctx, morningCall2); err != nil {
		t.Fatalf("failed to create morning call 2: %v", err)
	}

	// 受信者（user2）による削除は失敗すべき
	err = uc.Execute(ctx, DeleteInput{
		ID:       morningCall2.ID,
		SenderID: user2.ID,
	})
	if err == nil {
		t.Error("receiver should not be able to delete")
	} else if !strings.Contains(err.Error(), "送信者のみがモーニングコールを削除できます") {
		t.Errorf("unexpected error message: %v", err.Error())
	}

	// 無関係なユーザー（user3）による削除も失敗すべき
	err = uc.Execute(ctx, DeleteInput{
		ID:       morningCall2.ID,
		SenderID: user3.ID,
	})
	if err == nil {
		t.Error("unrelated user should not be able to delete")
	} else if !strings.Contains(err.Error(), "送信者のみがモーニングコールを削除できます") {
		t.Errorf("unexpected error message: %v", err.Error())
	}
}

func TestDeleteUseCase_Execute_StatusTransitions(t *testing.T) {
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

	// 各ステータスのモーニングコールを作成
	statuses := []struct {
		status      valueobject.MorningCallStatus
		canDelete   bool
		description string
	}{
		{valueobject.MorningCallStatusScheduled, true, "スケジュール済み"},
		{valueobject.MorningCallStatusCancelled, true, "キャンセル済み"},
		{valueobject.MorningCallStatusDelivered, false, "配信済み"},
		{valueobject.MorningCallStatusConfirmed, false, "確認済み"},
		{valueobject.MorningCallStatusExpired, false, "期限切れ"},
	}

	uc := NewDeleteUseCase(morningCallRepo, userRepo)

	for i, s := range statuses {
		t.Run(s.description, func(t *testing.T) {
			// モーニングコールを作成
			mc := &entity.MorningCall{
				ID:            fmt.Sprintf("mc%d", i),
				SenderID:      user1.ID,
				ReceiverID:    user2.ID,
				ScheduledTime: time.Now().Add(time.Duration(i) * time.Hour),
				Message:       s.description,
				Status:        s.status,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			if err := morningCallRepo.Create(ctx, mc); err != nil {
				t.Fatalf("failed to create morning call: %v", err)
			}

			// 削除を試みる
			err := uc.Execute(ctx, DeleteInput{
				ID:       mc.ID,
				SenderID: user1.ID,
			})

			if s.canDelete {
				if err != nil {
					t.Errorf("should be able to delete %s morning call: %v", s.description, err)
				}
				// 削除されたことを確認
				_, err := morningCallRepo.FindByID(ctx, mc.ID)
				if err == nil {
					t.Errorf("%s morning call should be deleted", s.description)
				}
			} else {
				if err == nil {
					t.Errorf("should not be able to delete %s morning call", s.description)
				} else if !strings.Contains(err.Error(), "削除できるのはスケジュール済みまたはキャンセル済みのモーニングコールのみです") {
					t.Errorf("unexpected error message for %s: %v", s.description, err.Error())
				}
				// 削除されていないことを確認
				_, err := morningCallRepo.FindByID(ctx, mc.ID)
				if err != nil {
					t.Errorf("%s morning call should not be deleted", s.description)
				}
			}
		})
	}
}

func TestDeleteUseCase_Execute_MultipleDeletes(t *testing.T) {
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

	// 複数のモーニングコールを作成
	for i := 0; i < 5; i++ {
		mc := &entity.MorningCall{
			ID:            fmt.Sprintf("mc%d", i),
			SenderID:      user1.ID,
			ReceiverID:    user2.ID,
			ScheduledTime: time.Now().Add(time.Duration(i+1) * time.Hour),
			Message:       fmt.Sprintf("メッセージ%d", i),
			Status:        valueobject.MorningCallStatusScheduled,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := morningCallRepo.Create(ctx, mc); err != nil {
			t.Fatalf("failed to create morning call %d: %v", i, err)
		}
	}

	uc := NewDeleteUseCase(morningCallRepo, userRepo)

	// すべてのモーニングコールを順番に削除
	for i := 0; i < 5; i++ {
		err := uc.Execute(ctx, DeleteInput{
			ID:       fmt.Sprintf("mc%d", i),
			SenderID: user1.ID,
		})
		if err != nil {
			t.Errorf("failed to delete morning call %d: %v", i, err)
		}
	}

	// すべて削除されたことを確認
	for i := 0; i < 5; i++ {
		_, err := morningCallRepo.FindByID(ctx, fmt.Sprintf("mc%d", i))
		if err == nil {
			t.Errorf("morning call %d should be deleted", i)
		}
	}

	// 削除済みのモーニングコールを再度削除しようとする
	err := uc.Execute(ctx, DeleteInput{
		ID:       "mc0",
		SenderID: user1.ID,
	})
	if err == nil {
		t.Error("should not be able to delete already deleted morning call")
	} else if !strings.Contains(err.Error(), "モーニングコールが見つかりません") {
		t.Errorf("unexpected error message: %v", err.Error())
	}
}
