package relationship

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
	"github.com/ochamu/morning-call-api/internal/infrastructure/memory"
)

func TestNewRemoveRelationshipUseCase(t *testing.T) {
	relationshipRepo := memory.NewRelationshipRepository()
	userRepo := memory.NewUserRepository()

	uc := NewRemoveRelationshipUseCase(relationshipRepo, userRepo)

	if uc == nil {
		t.Fatal("NewRemoveRelationshipUseCase returned nil")
	}
	if uc.relationshipRepo == nil {
		t.Error("relationshipRepo is nil")
	}
	if uc.userRepo == nil {
		t.Error("userRepo is nil")
	}
}

func TestRemoveRelationshipUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	// テスト用ユーザーを作成
	user1 := &entity.User{
		ID:           "user1-id",
		Username:     "user1",
		Email:        "user1@example.com",
		PasswordHash: "hashed",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	user2 := &entity.User{
		ID:           "user2-id",
		Username:     "user2",
		Email:        "user2@example.com",
		PasswordHash: "hashed",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	user3 := &entity.User{
		ID:           "user3-id",
		Username:     "user3",
		Email:        "user3@example.com",
		PasswordHash: "hashed",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	tests := []struct {
		name      string
		input     RemoveRelationshipInput
		setup     func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository)
		wantErr   bool
		errMsg    string
		checkFunc func(t *testing.T, output *RemoveRelationshipOutput, rr *memory.RelationshipRepository)
	}{
		{
			name: "成功ケース - 友達関係を解除（リクエスト送信者側から）",
			input: RemoveRelationshipInput{
				RelationshipID: "rel-1",
				UserID:         user1.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
				if err := ur.Create(ctx, user2); err != nil {
					t.Fatalf("failed to create user2: %v", err)
				}

				// 承認済みの友達関係を作成
				friendship := &entity.Relationship{
					ID:          "rel-1",
					RequesterID: user1.ID,
					ReceiverID:  user2.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, friendship); err != nil {
					t.Fatalf("failed to create friendship: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *RemoveRelationshipOutput, rr *memory.RelationshipRepository) {
				if !output.Success {
					t.Error("Expected success to be true")
				}
				if output.Message != "友達関係を解除しました" {
					t.Errorf("Message = %v, want %v", output.Message, "友達関係を解除しました")
				}
				// 関係が削除されていることを確認
				if _, err := rr.FindByID(ctx, "rel-1"); err == nil {
					t.Error("Relationship should have been deleted")
				}
			},
		},
		{
			name: "成功ケース - 友達関係を解除（リクエスト受信者側から）",
			input: RemoveRelationshipInput{
				RelationshipID: "rel-2",
				UserID:         user2.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
				if err := ur.Create(ctx, user2); err != nil {
					t.Fatalf("failed to create user2: %v", err)
				}

				// 承認済みの友達関係を作成
				friendship := &entity.Relationship{
					ID:          "rel-2",
					RequesterID: user1.ID,
					ReceiverID:  user2.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, friendship); err != nil {
					t.Fatalf("failed to create friendship: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *RemoveRelationshipOutput, rr *memory.RelationshipRepository) {
				if !output.Success {
					t.Error("Expected success to be true")
				}
				if output.Message != "友達関係を解除しました" {
					t.Errorf("Message = %v, want %v", output.Message, "友達関係を解除しました")
				}
			},
		},
		{
			name: "成功ケース - ペンディングリクエストを取り下げ（送信者）",
			input: RemoveRelationshipInput{
				RelationshipID: "rel-3",
				UserID:         user1.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
				if err := ur.Create(ctx, user2); err != nil {
					t.Fatalf("failed to create user2: %v", err)
				}

				// ペンディング状態の関係を作成
				pending := &entity.Relationship{
					ID:          "rel-3",
					RequesterID: user1.ID,
					ReceiverID:  user2.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, pending); err != nil {
					t.Fatalf("failed to create pending relationship: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *RemoveRelationshipOutput, rr *memory.RelationshipRepository) {
				if !output.Success {
					t.Error("Expected success to be true")
				}
				if output.Message != "友達リクエストを取り下げました" {
					t.Errorf("Message = %v, want %v", output.Message, "友達リクエストを取り下げました")
				}
			},
		},
		{
			name: "成功ケース - ペンディングリクエストを削除（受信者）",
			input: RemoveRelationshipInput{
				RelationshipID: "rel-4",
				UserID:         user2.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
				if err := ur.Create(ctx, user2); err != nil {
					t.Fatalf("failed to create user2: %v", err)
				}

				// ペンディング状態の関係を作成
				pending := &entity.Relationship{
					ID:          "rel-4",
					RequesterID: user1.ID,
					ReceiverID:  user2.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, pending); err != nil {
					t.Fatalf("failed to create pending relationship: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *RemoveRelationshipOutput, rr *memory.RelationshipRepository) {
				if !output.Success {
					t.Error("Expected success to be true")
				}
				if output.Message != "友達リクエストを削除しました" {
					t.Errorf("Message = %v, want %v", output.Message, "友達リクエストを削除しました")
				}
			},
		},
		{
			name: "成功ケース - 拒否済みリクエストを削除",
			input: RemoveRelationshipInput{
				RelationshipID: "rel-5",
				UserID:         user1.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
				if err := ur.Create(ctx, user2); err != nil {
					t.Fatalf("failed to create user2: %v", err)
				}

				// 拒否済み状態の関係を作成
				rejected := &entity.Relationship{
					ID:          "rel-5",
					RequesterID: user1.ID,
					ReceiverID:  user2.ID,
					Status:      valueobject.RelationshipStatusRejected,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, rejected); err != nil {
					t.Fatalf("failed to create rejected relationship: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *RemoveRelationshipOutput, rr *memory.RelationshipRepository) {
				if !output.Success {
					t.Error("Expected success to be true")
				}
				if output.Message != "拒否済みの友達リクエストを削除しました" {
					t.Errorf("Message = %v, want %v", output.Message, "拒否済みの友達リクエストを削除しました")
				}
			},
		},
		{
			name: "成功ケース - 相手ユーザーが削除済みでも削除可能",
			input: RemoveRelationshipInput{
				RelationshipID: "rel-6",
				UserID:         user1.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
				// user2は作成しない（削除済みユーザーをシミュレート）

				// 友達関係を作成
				friendship := &entity.Relationship{
					ID:          "rel-6",
					RequesterID: user1.ID,
					ReceiverID:  "deleted-user-id",
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, friendship); err != nil {
					t.Fatalf("failed to create friendship: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *RemoveRelationshipOutput, rr *memory.RelationshipRepository) {
				if !output.Success {
					t.Error("Expected success to be true")
				}
			},
		},
		{
			name: "エラー - ブロック関係は削除不可",
			input: RemoveRelationshipInput{
				RelationshipID: "rel-7",
				UserID:         user1.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
				if err := ur.Create(ctx, user2); err != nil {
					t.Fatalf("failed to create user2: %v", err)
				}

				// ブロック関係を作成
				blocked := &entity.Relationship{
					ID:          "rel-7",
					RequesterID: user1.ID,
					ReceiverID:  user2.ID,
					Status:      valueobject.RelationshipStatusBlocked,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, blocked); err != nil {
					t.Fatalf("failed to create blocked relationship: %v", err)
				}
			},
			wantErr: true,
			errMsg:  "ブロック関係は削除できません。先にブロックを解除してください",
		},
		{
			name: "エラー - 関係IDが空",
			input: RemoveRelationshipInput{
				RelationshipID: "",
				UserID:         user1.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// セットアップ不要
			},
			wantErr: true,
			errMsg:  "関係IDは必須です",
		},
		{
			name: "エラー - ユーザーIDが空",
			input: RemoveRelationshipInput{
				RelationshipID: "rel-8",
				UserID:         "",
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// セットアップ不要
			},
			wantErr: true,
			errMsg:  "ユーザーIDは必須です",
		},
		{
			name: "エラー - ユーザーが存在しない",
			input: RemoveRelationshipInput{
				RelationshipID: "rel-9",
				UserID:         "nonexistent",
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// ユーザーを作成しない
			},
			wantErr: true,
			errMsg:  "ユーザーが見つかりません",
		},
		{
			name: "エラー - 関係が存在しない",
			input: RemoveRelationshipInput{
				RelationshipID: "nonexistent",
				UserID:         user1.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
			},
			wantErr: true,
			errMsg:  "関係が見つかりません",
		},
		{
			name: "エラー - 削除権限がない（第三者）",
			input: RemoveRelationshipInput{
				RelationshipID: "rel-10",
				UserID:         user3.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
				if err := ur.Create(ctx, user2); err != nil {
					t.Fatalf("failed to create user2: %v", err)
				}
				if err := ur.Create(ctx, user3); err != nil {
					t.Fatalf("failed to create user3: %v", err)
				}

				// user1とuser2の関係を作成
				friendship := &entity.Relationship{
					ID:          "rel-10",
					RequesterID: user1.ID,
					ReceiverID:  user2.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, friendship); err != nil {
					t.Fatalf("failed to create friendship: %v", err)
				}
			},
			wantErr: true,
			errMsg:  "この関係を削除する権限がありません",
		},
		{
			name: "成功ケース - 第三者の関係は影響しない",
			input: RemoveRelationshipInput{
				RelationshipID: "rel-11",
				UserID:         user1.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
				if err := ur.Create(ctx, user2); err != nil {
					t.Fatalf("failed to create user2: %v", err)
				}
				if err := ur.Create(ctx, user3); err != nil {
					t.Fatalf("failed to create user3: %v", err)
				}

				// user1とuser2の関係
				rel1 := &entity.Relationship{
					ID:          "rel-11",
					RequesterID: user1.ID,
					ReceiverID:  user2.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, rel1); err != nil {
					t.Fatalf("failed to create rel1: %v", err)
				}

				// user2とuser3の関係（影響を受けない）
				rel2 := &entity.Relationship{
					ID:          "rel-12",
					RequesterID: user2.ID,
					ReceiverID:  user3.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, rel2); err != nil {
					t.Fatalf("failed to create rel2: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *RemoveRelationshipOutput, rr *memory.RelationshipRepository) {
				// rel-11は削除されている
				if _, err := rr.FindByID(ctx, "rel-11"); err == nil {
					t.Error("rel-11 should have been deleted")
				}
				// rel-12は影響を受けない
				rel2, err := rr.FindByID(ctx, "rel-12")
				if err != nil {
					t.Errorf("rel-12 should still exist: %v", err)
				}
				if rel2.Status != valueobject.RelationshipStatusAccepted {
					t.Error("rel-12 status should not be affected")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各テストケースで新しいリポジトリを作成
			relationshipRepo := memory.NewRelationshipRepository()
			userRepo := memory.NewUserRepository()

			// セットアップを実行
			if tt.setup != nil {
				tt.setup(t, relationshipRepo, userRepo)
			}

			// UseCaseを作成して実行
			uc := NewRemoveRelationshipUseCase(relationshipRepo, userRepo)
			output, err := uc.Execute(ctx, tt.input)

			// エラーチェック
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
				// 追加のチェック
				if tt.checkFunc != nil {
					tt.checkFunc(t, output, relationshipRepo)
				}
			}
		})
	}
}
