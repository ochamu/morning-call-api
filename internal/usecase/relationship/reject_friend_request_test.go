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

func TestNewRejectFriendRequestUseCase(t *testing.T) {
	relationshipRepo := memory.NewRelationshipRepository()
	userRepo := memory.NewUserRepository()

	uc := NewRejectFriendRequestUseCase(relationshipRepo, userRepo)

	if uc == nil {
		t.Fatal("NewRejectFriendRequestUseCase returned nil")
	}
	if uc.relationshipRepo == nil {
		t.Error("relationshipRepo is nil")
	}
	if uc.userRepo == nil {
		t.Error("userRepo is nil")
	}
}

func TestRejectFriendRequestUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		setup     func(*testing.T, *memory.RelationshipRepository, *memory.UserRepository)
		input     RejectFriendRequestInput
		wantErr   bool
		errMsg    string
		checkFunc func(*testing.T, *RejectFriendRequestOutput)
	}{
		{
			name: "成功ケース - Pending状態のリクエストを拒否",
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// ユーザーを作成
				requester := &entity.User{
					ID:           "requester1",
					Username:     "alice",
					Email:        "alice@example.com",
					PasswordHash: "hash",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				receiver := &entity.User{
					ID:           "receiver1",
					Username:     "bob",
					Email:        "bob@example.com",
					PasswordHash: "hash",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				if err := ur.Create(ctx, requester); err != nil {
					t.Fatalf("failed to create requester: %v", err)
				}
				if err := ur.Create(ctx, receiver); err != nil {
					t.Fatalf("failed to create receiver: %v", err)
				}

				// Pending状態の関係を作成
				relationship := &entity.Relationship{
					ID:          "rel1",
					RequesterID: requester.ID,
					ReceiverID:  receiver.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, relationship); err != nil {
					t.Fatalf("failed to create relationship: %v", err)
				}
			},
			input: RejectFriendRequestInput{
				RelationshipID: "rel1",
				ReceiverID:     "receiver1",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *RejectFriendRequestOutput) {
				if output == nil {
					t.Fatal("output is nil")
				}
				if output.Relationship == nil {
					t.Fatal("Relationship is nil")
				}
				if output.Relationship.Status != valueobject.RelationshipStatusRejected {
					t.Errorf("Status = %v, want %v", output.Relationship.Status, valueobject.RelationshipStatusRejected)
				}
			},
		},
		{
			name:  "エラー - 関係IDが空",
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {},
			input: RejectFriendRequestInput{
				RelationshipID: "",
				ReceiverID:     "receiver1",
			},
			wantErr: true,
			errMsg:  "関係IDは必須です",
		},
		{
			name:  "エラー - 拒否者IDが空",
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {},
			input: RejectFriendRequestInput{
				RelationshipID: "rel1",
				ReceiverID:     "",
			},
			wantErr: true,
			errMsg:  "拒否者IDは必須です",
		},
		{
			name: "エラー - 拒否者が存在しない",
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				relationship := &entity.Relationship{
					ID:          "rel1",
					RequesterID: "requester1",
					ReceiverID:  "receiver1",
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, relationship); err != nil {
					t.Fatalf("failed to create relationship: %v", err)
				}
			},
			input: RejectFriendRequestInput{
				RelationshipID: "rel1",
				ReceiverID:     "nonexistent",
			},
			wantErr: true,
			errMsg:  "拒否者が見つかりません",
		},
		{
			name: "エラー - 関係が存在しない",
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				receiver := &entity.User{
					ID:           "receiver1",
					Username:     "bob",
					Email:        "bob@example.com",
					PasswordHash: "hash",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				if err := ur.Create(ctx, receiver); err != nil {
					t.Fatalf("failed to create receiver: %v", err)
				}
			},
			input: RejectFriendRequestInput{
				RelationshipID: "nonexistent",
				ReceiverID:     "receiver1",
			},
			wantErr: true,
			errMsg:  "友達リクエストが見つかりません",
		},
		{
			name: "エラー - 拒否権限がない（リクエスト送信者が拒否しようとする）",
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				requester := &entity.User{
					ID:           "requester1",
					Username:     "alice",
					Email:        "alice@example.com",
					PasswordHash: "hash",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				receiver := &entity.User{
					ID:           "receiver1",
					Username:     "bob",
					Email:        "bob@example.com",
					PasswordHash: "hash",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				if err := ur.Create(ctx, requester); err != nil {
					t.Fatalf("failed to create requester: %v", err)
				}
				if err := ur.Create(ctx, receiver); err != nil {
					t.Fatalf("failed to create receiver: %v", err)
				}

				relationship := &entity.Relationship{
					ID:          "rel1",
					RequesterID: requester.ID,
					ReceiverID:  receiver.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, relationship); err != nil {
					t.Fatalf("failed to create relationship: %v", err)
				}
			},
			input: RejectFriendRequestInput{
				RelationshipID: "rel1",
				ReceiverID:     "requester1", // 送信者が拒否しようとしている
			},
			wantErr: true,
			errMsg:  "このリクエストを拒否する権限がありません",
		},
		{
			name: "エラー - 既に承認済みのリクエスト",
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				requester := &entity.User{
					ID:           "requester1",
					Username:     "alice",
					Email:        "alice@example.com",
					PasswordHash: "hash",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				receiver := &entity.User{
					ID:           "receiver1",
					Username:     "bob",
					Email:        "bob@example.com",
					PasswordHash: "hash",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				if err := ur.Create(ctx, requester); err != nil {
					t.Fatalf("failed to create requester: %v", err)
				}
				if err := ur.Create(ctx, receiver); err != nil {
					t.Fatalf("failed to create receiver: %v", err)
				}

				relationship := &entity.Relationship{
					ID:          "rel1",
					RequesterID: requester.ID,
					ReceiverID:  receiver.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, relationship); err != nil {
					t.Fatalf("failed to create relationship: %v", err)
				}
			},
			input: RejectFriendRequestInput{
				RelationshipID: "rel1",
				ReceiverID:     "receiver1",
			},
			wantErr: true,
			errMsg:  "既に承認済みの友達リクエストは拒否できません",
		},
		{
			name: "エラー - 既に拒否済みのリクエスト",
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				requester := &entity.User{
					ID:           "requester1",
					Username:     "alice",
					Email:        "alice@example.com",
					PasswordHash: "hash",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				receiver := &entity.User{
					ID:           "receiver1",
					Username:     "bob",
					Email:        "bob@example.com",
					PasswordHash: "hash",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				if err := ur.Create(ctx, requester); err != nil {
					t.Fatalf("failed to create requester: %v", err)
				}
				if err := ur.Create(ctx, receiver); err != nil {
					t.Fatalf("failed to create receiver: %v", err)
				}

				relationship := &entity.Relationship{
					ID:          "rel1",
					RequesterID: requester.ID,
					ReceiverID:  receiver.ID,
					Status:      valueobject.RelationshipStatusRejected,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, relationship); err != nil {
					t.Fatalf("failed to create relationship: %v", err)
				}
			},
			input: RejectFriendRequestInput{
				RelationshipID: "rel1",
				ReceiverID:     "receiver1",
			},
			wantErr: true,
			errMsg:  "既に拒否済みの友達リクエストです",
		},
		{
			name: "エラー - ブロック関係のリクエスト",
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				requester := &entity.User{
					ID:           "requester1",
					Username:     "alice",
					Email:        "alice@example.com",
					PasswordHash: "hash",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				receiver := &entity.User{
					ID:           "receiver1",
					Username:     "bob",
					Email:        "bob@example.com",
					PasswordHash: "hash",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				if err := ur.Create(ctx, requester); err != nil {
					t.Fatalf("failed to create requester: %v", err)
				}
				if err := ur.Create(ctx, receiver); err != nil {
					t.Fatalf("failed to create receiver: %v", err)
				}

				relationship := &entity.Relationship{
					ID:          "rel1",
					RequesterID: requester.ID,
					ReceiverID:  receiver.ID,
					Status:      valueobject.RelationshipStatusBlocked,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, relationship); err != nil {
					t.Fatalf("failed to create relationship: %v", err)
				}
			},
			input: RejectFriendRequestInput{
				RelationshipID: "rel1",
				ReceiverID:     "receiver1",
			},
			wantErr: true,
			errMsg:  "ブロック関係のリクエストは拒否できません",
		},
		{
			name: "エラー - リクエスト送信者が存在しない（データ不整合）",
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				receiver := &entity.User{
					ID:           "receiver1",
					Username:     "bob",
					Email:        "bob@example.com",
					PasswordHash: "hash",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				if err := ur.Create(ctx, receiver); err != nil {
					t.Fatalf("failed to create receiver: %v", err)
				}

				// 存在しない送信者IDで関係を作成（データ不整合の状態）
				relationship := &entity.Relationship{
					ID:          "rel1",
					RequesterID: "nonexistent_requester",
					ReceiverID:  receiver.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, relationship); err != nil {
					t.Fatalf("failed to create relationship: %v", err)
				}
			},
			input: RejectFriendRequestInput{
				RelationshipID: "rel1",
				ReceiverID:     "receiver1",
			},
			wantErr: true,
			errMsg:  "リクエスト送信者が見つかりません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// リポジトリの初期化
			relationshipRepo := memory.NewRelationshipRepository()
			userRepo := memory.NewUserRepository()

			// セットアップ
			tt.setup(t, relationshipRepo, userRepo)

			// UseCase作成
			uc := NewRejectFriendRequestUseCase(relationshipRepo, userRepo)

			// 実行
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
				if tt.checkFunc != nil {
					tt.checkFunc(t, output)
				}
			}
		})
	}
}

func TestRejectFriendRequestUseCase_Execute_UpdateTimestamp(t *testing.T) {
	ctx := context.Background()

	// リポジトリの初期化
	relationshipRepo := memory.NewRelationshipRepository()
	userRepo := memory.NewUserRepository()

	// ユーザーを作成
	requester := &entity.User{
		ID:           "requester1",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	receiver := &entity.User{
		ID:           "receiver1",
		Username:     "bob",
		Email:        "bob@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := userRepo.Create(ctx, requester); err != nil {
		t.Fatalf("failed to create requester: %v", err)
	}
	if err := userRepo.Create(ctx, receiver); err != nil {
		t.Fatalf("failed to create receiver: %v", err)
	}

	// 古い更新日時で関係を作成
	oldTime := time.Now().Add(-24 * time.Hour)
	relationship := &entity.Relationship{
		ID:          "rel1",
		RequesterID: requester.ID,
		ReceiverID:  receiver.ID,
		Status:      valueobject.RelationshipStatusPending,
		CreatedAt:   oldTime,
		UpdatedAt:   oldTime,
	}
	if err := relationshipRepo.Create(ctx, relationship); err != nil {
		t.Fatalf("failed to create relationship: %v", err)
	}

	// UseCase作成と実行
	uc := NewRejectFriendRequestUseCase(relationshipRepo, userRepo)
	input := RejectFriendRequestInput{
		RelationshipID: "rel1",
		ReceiverID:     receiver.ID,
	}

	beforeTime := time.Now()
	output, err := uc.Execute(ctx, input)
	afterTime := time.Now()

	// エラーチェック
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 更新日時のチェック
	if output.Relationship.UpdatedAt.Before(beforeTime) || output.Relationship.UpdatedAt.After(afterTime) {
		t.Errorf("UpdatedAt = %v, want between %v and %v", output.Relationship.UpdatedAt, beforeTime, afterTime)
	}

	// 作成日時は変更されていないことを確認
	if !output.Relationship.CreatedAt.Equal(oldTime) {
		t.Errorf("CreatedAt was modified: got %v, want %v", output.Relationship.CreatedAt, oldTime)
	}
}
