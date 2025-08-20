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

func TestNewBlockUserUseCase(t *testing.T) {
	relationshipRepo := memory.NewRelationshipRepository()
	userRepo := memory.NewUserRepository()

	uc := NewBlockUserUseCase(relationshipRepo, userRepo)

	if uc == nil {
		t.Fatal("NewBlockUserUseCase returned nil")
	}
	if uc.relationshipRepo == nil {
		t.Error("relationshipRepo is nil")
	}
	if uc.userRepo == nil {
		t.Error("userRepo is nil")
	}
}

func TestBlockUserUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	// テスト用ユーザーを作成
	blocker := &entity.User{
		ID:           "blocker-id",
		Username:     "blocker",
		Email:        "blocker@example.com",
		PasswordHash: "hashed",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	blocked := &entity.User{
		ID:           "blocked-id",
		Username:     "blocked",
		Email:        "blocked@example.com",
		PasswordHash: "hashed",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	thirdUser := &entity.User{
		ID:           "third-id",
		Username:     "third",
		Email:        "third@example.com",
		PasswordHash: "hashed",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	tests := []struct {
		name      string
		input     BlockUserInput
		setup     func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository)
		wantErr   bool
		errMsg    string
		checkFunc func(t *testing.T, output *BlockUserOutput, rr *memory.RelationshipRepository)
	}{
		{
			name: "成功ケース - 新規ブロック",
			input: BlockUserInput{
				BlockerID: blocker.ID,
				BlockedID: blocked.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, blocker); err != nil {
					t.Fatalf("failed to create blocker: %v", err)
				}
				if err := ur.Create(ctx, blocked); err != nil {
					t.Fatalf("failed to create blocked: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *BlockUserOutput, rr *memory.RelationshipRepository) {
				if output.Relationship == nil {
					t.Fatal("Relationship is nil")
				}
				if output.Relationship.Status != valueobject.RelationshipStatusBlocked {
					t.Errorf("Status = %v, want %v", output.Relationship.Status, valueobject.RelationshipStatusBlocked)
				}
			},
		},
		{
			name: "成功ケース - 友達関係からブロック",
			input: BlockUserInput{
				BlockerID: blocker.ID,
				BlockedID: blocked.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, blocker); err != nil {
					t.Fatalf("failed to create blocker: %v", err)
				}
				if err := ur.Create(ctx, blocked); err != nil {
					t.Fatalf("failed to create blocked: %v", err)
				}

				// 既存の友達関係を作成
				friendship := &entity.Relationship{
					ID:          "rel-1",
					RequesterID: blocker.ID,
					ReceiverID:  blocked.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, friendship); err != nil {
					t.Fatalf("failed to create friendship: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *BlockUserOutput, rr *memory.RelationshipRepository) {
				if output.Relationship == nil {
					t.Fatal("Relationship is nil")
				}
				if output.Relationship.Status != valueobject.RelationshipStatusBlocked {
					t.Errorf("Status = %v, want %v", output.Relationship.Status, valueobject.RelationshipStatusBlocked)
				}
				if output.Relationship.ID != "rel-1" {
					t.Error("Expected to update existing relationship, but got new one")
				}
			},
		},
		{
			name: "成功ケース - ペンディング状態からブロック",
			input: BlockUserInput{
				BlockerID: blocked.ID, // 受信者側からブロック
				BlockedID: blocker.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, blocker); err != nil {
					t.Fatalf("failed to create blocker: %v", err)
				}
				if err := ur.Create(ctx, blocked); err != nil {
					t.Fatalf("failed to create blocked: %v", err)
				}

				// ペンディング状態の関係を作成
				pending := &entity.Relationship{
					ID:          "rel-2",
					RequesterID: blocker.ID,
					ReceiverID:  blocked.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, pending); err != nil {
					t.Fatalf("failed to create pending relationship: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *BlockUserOutput, rr *memory.RelationshipRepository) {
				if output.Relationship == nil {
					t.Fatal("Relationship is nil")
				}
				if output.Relationship.Status != valueobject.RelationshipStatusBlocked {
					t.Errorf("Status = %v, want %v", output.Relationship.Status, valueobject.RelationshipStatusBlocked)
				}
			},
		},
		{
			name: "成功ケース - 拒否状態からブロック",
			input: BlockUserInput{
				BlockerID: blocked.ID, // 受信者側からブロック
				BlockedID: blocker.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, blocker); err != nil {
					t.Fatalf("failed to create blocker: %v", err)
				}
				if err := ur.Create(ctx, blocked); err != nil {
					t.Fatalf("failed to create blocked: %v", err)
				}

				// 拒否状態の関係を作成
				rejected := &entity.Relationship{
					ID:          "rel-3",
					RequesterID: blocker.ID,
					ReceiverID:  blocked.ID,
					Status:      valueobject.RelationshipStatusRejected,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, rejected); err != nil {
					t.Fatalf("failed to create rejected relationship: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *BlockUserOutput, rr *memory.RelationshipRepository) {
				if output.Relationship == nil {
					t.Fatal("Relationship is nil")
				}
				if output.Relationship.Status != valueobject.RelationshipStatusBlocked {
					t.Errorf("Status = %v, want %v", output.Relationship.Status, valueobject.RelationshipStatusBlocked)
				}
			},
		},
		{
			name: "エラー - 相手からブロックされている場合の処理",
			input: BlockUserInput{
				BlockerID: blocked.ID,
				BlockedID: blocker.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, blocker); err != nil {
					t.Fatalf("failed to create blocker: %v", err)
				}
				if err := ur.Create(ctx, blocked); err != nil {
					t.Fatalf("failed to create blocked: %v", err)
				}

				// 相手からのブロック関係を作成
				existingBlock := &entity.Relationship{
					ID:          "rel-4",
					RequesterID: blocker.ID,
					ReceiverID:  blocked.ID,
					Status:      valueobject.RelationshipStatusBlocked,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, existingBlock); err != nil {
					t.Fatalf("failed to create existing block: %v", err)
				}
			},
			wantErr: true,
			errMsg:  "相手にブロックされています",
		},
		{
			name: "エラー - ブロック実行者IDが空",
			input: BlockUserInput{
				BlockerID: "",
				BlockedID: blocked.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// セットアップ不要
			},
			wantErr: true,
			errMsg:  "ブロック実行者IDは必須です",
		},
		{
			name: "エラー - ブロック対象者IDが空",
			input: BlockUserInput{
				BlockerID: blocker.ID,
				BlockedID: "",
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// セットアップ不要
			},
			wantErr: true,
			errMsg:  "ブロック対象者IDは必須です",
		},
		{
			name: "エラー - 自分自身をブロック",
			input: BlockUserInput{
				BlockerID: blocker.ID,
				BlockedID: blocker.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// セットアップ不要
			},
			wantErr: true,
			errMsg:  "自分自身をブロックすることはできません",
		},
		{
			name: "エラー - ブロック実行者が存在しない",
			input: BlockUserInput{
				BlockerID: "nonexistent",
				BlockedID: blocked.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, blocked); err != nil {
					t.Fatalf("failed to create blocked: %v", err)
				}
			},
			wantErr: true,
			errMsg:  "ブロック実行者が見つかりません",
		},
		{
			name: "エラー - ブロック対象者が存在しない",
			input: BlockUserInput{
				BlockerID: blocker.ID,
				BlockedID: "nonexistent",
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, blocker); err != nil {
					t.Fatalf("failed to create blocker: %v", err)
				}
			},
			wantErr: true,
			errMsg:  "ブロック対象者が見つかりません",
		},
		{
			name: "エラー - 既にブロック済み",
			input: BlockUserInput{
				BlockerID: blocker.ID,
				BlockedID: blocked.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, blocker); err != nil {
					t.Fatalf("failed to create blocker: %v", err)
				}
				if err := ur.Create(ctx, blocked); err != nil {
					t.Fatalf("failed to create blocked: %v", err)
				}

				// 既にブロック済みの関係を作成
				blocked := &entity.Relationship{
					ID:          "rel-5",
					RequesterID: blocker.ID,
					ReceiverID:  blocked.ID,
					Status:      valueobject.RelationshipStatusBlocked,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, blocked); err != nil {
					t.Fatalf("failed to create blocked relationship: %v", err)
				}
			},
			wantErr: true,
			errMsg:  "既にこのユーザーをブロックしています",
		},
		{
			name: "成功ケース - 第三者の関係は影響しない",
			input: BlockUserInput{
				BlockerID: blocker.ID,
				BlockedID: blocked.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, blocker); err != nil {
					t.Fatalf("failed to create blocker: %v", err)
				}
				if err := ur.Create(ctx, blocked); err != nil {
					t.Fatalf("failed to create blocked: %v", err)
				}
				if err := ur.Create(ctx, thirdUser); err != nil {
					t.Fatalf("failed to create third user: %v", err)
				}

				// 第三者との関係を作成
				thirdRelation := &entity.Relationship{
					ID:          "rel-6",
					RequesterID: thirdUser.ID,
					ReceiverID:  blocked.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, thirdRelation); err != nil {
					t.Fatalf("failed to create third relation: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *BlockUserOutput, rr *memory.RelationshipRepository) {
				if output.Relationship == nil {
					t.Fatal("Relationship is nil")
				}
				// 第三者の関係は影響を受けないことを確認
				thirdRel, err := rr.FindByID(ctx, "rel-6")
				if err != nil {
					t.Errorf("Failed to find third relation: %v", err)
				}
				if thirdRel.Status != valueobject.RelationshipStatusAccepted {
					t.Error("Third user's relationship should not be affected")
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
			uc := NewBlockUserUseCase(relationshipRepo, userRepo)
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

