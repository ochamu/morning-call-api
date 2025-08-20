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

func TestNewListFriendRequestsUseCase(t *testing.T) {
	relationshipRepo := memory.NewRelationshipRepository()
	userRepo := memory.NewUserRepository()

	uc := NewListFriendRequestsUseCase(relationshipRepo, userRepo)

	if uc == nil {
		t.Fatal("NewListFriendRequestsUseCase returned nil")
	}
	if uc.relationshipRepo == nil {
		t.Error("relationshipRepo is nil")
	}
	if uc.userRepo == nil {
		t.Error("userRepo is nil")
	}
}

func TestListFriendRequestsUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	// テスト用ユーザーを作成
	mainUser := &entity.User{
		ID:           "main-user-id",
		Username:     "mainuser",
		Email:        "main@example.com",
		PasswordHash: "hashed",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	requester1 := &entity.User{
		ID:           "requester1-id",
		Username:     "requester1",
		Email:        "requester1@example.com",
		PasswordHash: "hashed",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	requester2 := &entity.User{
		ID:           "requester2-id",
		Username:     "requester2",
		Email:        "requester2@example.com",
		PasswordHash: "hashed",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	receiver1 := &entity.User{
		ID:           "receiver1-id",
		Username:     "receiver1",
		Email:        "receiver1@example.com",
		PasswordHash: "hashed",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	receiver2 := &entity.User{
		ID:           "receiver2-id",
		Username:     "receiver2",
		Email:        "receiver2@example.com",
		PasswordHash: "hashed",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	friend := &entity.User{
		ID:           "friend-id",
		Username:     "friend",
		Email:        "friend@example.com",
		PasswordHash: "hashed",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	tests := []struct {
		name      string
		input     ListFriendRequestsInput
		setup     func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository)
		wantErr   bool
		errMsg    string
		checkFunc func(t *testing.T, output *ListFriendRequestsOutput)
	}{
		{
			name: "成功ケース - 受信したリクエストを取得",
			input: ListFriendRequestsInput{
				UserID: mainUser.ID,
				Type:   "received",
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// ユーザーを作成
				if err := ur.Create(ctx, mainUser); err != nil {
					t.Fatalf("failed to create main user: %v", err)
				}
				if err := ur.Create(ctx, requester1); err != nil {
					t.Fatalf("failed to create requester1: %v", err)
				}
				if err := ur.Create(ctx, requester2); err != nil {
					t.Fatalf("failed to create requester2: %v", err)
				}

				// 受信したリクエストを作成
				rel1 := &entity.Relationship{
					ID:          "rel-1",
					RequesterID: requester1.ID,
					ReceiverID:  mainUser.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now().Add(-2 * time.Hour),
					UpdatedAt:   time.Now().Add(-2 * time.Hour),
				}
				if err := rr.Create(ctx, rel1); err != nil {
					t.Fatalf("failed to create relationship 1: %v", err)
				}

				rel2 := &entity.Relationship{
					ID:          "rel-2",
					RequesterID: requester2.ID,
					ReceiverID:  mainUser.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now().Add(-1 * time.Hour),
					UpdatedAt:   time.Now().Add(-1 * time.Hour),
				}
				if err := rr.Create(ctx, rel2); err != nil {
					t.Fatalf("failed to create relationship 2: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *ListFriendRequestsOutput) {
				if output == nil {
					t.Fatal("output is nil")
				}
				if output.TotalCount != 2 {
					t.Errorf("TotalCount = %d, want 2", output.TotalCount)
				}
				if len(output.Requests) != 2 {
					t.Errorf("len(Requests) = %d, want 2", len(output.Requests))
				}
				// 各リクエストの確認
				for _, req := range output.Requests {
					if req.Requester == nil {
						t.Error("Requester should not be nil for received requests")
					}
					if req.Receiver != nil {
						t.Error("Receiver should be nil for received requests")
					}
					if req.RequestedAt == "" {
						t.Error("RequestedAt should not be empty")
					}
				}
			},
		},
		{
			name: "成功ケース - 送信したリクエストを取得",
			input: ListFriendRequestsInput{
				UserID: mainUser.ID,
				Type:   "sent",
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// ユーザーを作成
				if err := ur.Create(ctx, mainUser); err != nil {
					t.Fatalf("failed to create main user: %v", err)
				}
				if err := ur.Create(ctx, receiver1); err != nil {
					t.Fatalf("failed to create receiver1: %v", err)
				}
				if err := ur.Create(ctx, receiver2); err != nil {
					t.Fatalf("failed to create receiver2: %v", err)
				}

				// 送信したリクエストを作成
				rel1 := &entity.Relationship{
					ID:          "rel-1",
					RequesterID: mainUser.ID,
					ReceiverID:  receiver1.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now().Add(-2 * time.Hour),
					UpdatedAt:   time.Now().Add(-2 * time.Hour),
				}
				if err := rr.Create(ctx, rel1); err != nil {
					t.Fatalf("failed to create relationship 1: %v", err)
				}

				rel2 := &entity.Relationship{
					ID:          "rel-2",
					RequesterID: mainUser.ID,
					ReceiverID:  receiver2.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now().Add(-1 * time.Hour),
					UpdatedAt:   time.Now().Add(-1 * time.Hour),
				}
				if err := rr.Create(ctx, rel2); err != nil {
					t.Fatalf("failed to create relationship 2: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *ListFriendRequestsOutput) {
				if output == nil {
					t.Fatal("output is nil")
				}
				if output.TotalCount != 2 {
					t.Errorf("TotalCount = %d, want 2", output.TotalCount)
				}
				if len(output.Requests) != 2 {
					t.Errorf("len(Requests) = %d, want 2", len(output.Requests))
				}
				// 各リクエストの確認
				for _, req := range output.Requests {
					if req.Requester != nil {
						t.Error("Requester should be nil for sent requests")
					}
					if req.Receiver == nil {
						t.Error("Receiver should not be nil for sent requests")
					}
					if req.RequestedAt == "" {
						t.Error("RequestedAt should not be empty")
					}
				}
			},
		},
		{
			name: "成功ケース - リクエストがない場合",
			input: ListFriendRequestsInput{
				UserID: mainUser.ID,
				Type:   "received",
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, mainUser); err != nil {
					t.Fatalf("failed to create main user: %v", err)
				}
				// リクエストは作成しない
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *ListFriendRequestsOutput) {
				if output == nil {
					t.Fatal("output is nil")
				}
				if output.TotalCount != 0 {
					t.Errorf("TotalCount = %d, want 0", output.TotalCount)
				}
				if len(output.Requests) != 0 {
					t.Errorf("len(Requests) = %d, want 0", len(output.Requests))
				}
			},
		},
		{
			name: "成功ケース - 異なるステータスが混在（Pendingのみ表示）",
			input: ListFriendRequestsInput{
				UserID: mainUser.ID,
				Type:   "received",
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// ユーザーを作成
				if err := ur.Create(ctx, mainUser); err != nil {
					t.Fatalf("failed to create main user: %v", err)
				}
				if err := ur.Create(ctx, requester1); err != nil {
					t.Fatalf("failed to create requester1: %v", err)
				}
				if err := ur.Create(ctx, friend); err != nil {
					t.Fatalf("failed to create friend: %v", err)
				}

				// Pendingリクエスト
				rel1 := &entity.Relationship{
					ID:          "rel-1",
					RequesterID: requester1.ID,
					ReceiverID:  mainUser.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, rel1); err != nil {
					t.Fatalf("failed to create pending relationship: %v", err)
				}

				// Accepted関係（友達）
				rel2 := &entity.Relationship{
					ID:          "rel-2",
					RequesterID: friend.ID,
					ReceiverID:  mainUser.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, rel2); err != nil {
					t.Fatalf("failed to create accepted relationship: %v", err)
				}

				// Rejected関係
				rel3 := &entity.Relationship{
					ID:          "rel-3",
					RequesterID: "rejected-user",
					ReceiverID:  mainUser.ID,
					Status:      valueobject.RelationshipStatusRejected,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, rel3); err != nil {
					t.Fatalf("failed to create rejected relationship: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *ListFriendRequestsOutput) {
				if output == nil {
					t.Fatal("output is nil")
				}
				// Pendingのリクエストのみが表示される
				if output.TotalCount != 1 {
					t.Errorf("TotalCount = %d, want 1", output.TotalCount)
				}
				if len(output.Requests) != 1 {
					t.Errorf("len(Requests) = %d, want 1", len(output.Requests))
				}
				if len(output.Requests) > 0 && output.Requests[0].Relationship.Status != valueobject.RelationshipStatusPending {
					t.Error("Only pending requests should be returned")
				}
			},
		},
		{
			name: "成功ケース - 削除されたユーザーからのリクエストをスキップ",
			input: ListFriendRequestsInput{
				UserID: mainUser.ID,
				Type:   "received",
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// メインユーザーと1人のリクエスターのみを作成（もう1人は作成しない＝削除済み扱い）
				if err := ur.Create(ctx, mainUser); err != nil {
					t.Fatalf("failed to create main user: %v", err)
				}
				if err := ur.Create(ctx, requester1); err != nil {
					t.Fatalf("failed to create requester1: %v", err)
				}

				// 存在するユーザーからのリクエスト
				rel1 := &entity.Relationship{
					ID:          "rel-1",
					RequesterID: requester1.ID,
					ReceiverID:  mainUser.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, rel1); err != nil {
					t.Fatalf("failed to create relationship 1: %v", err)
				}

				// 存在しないユーザーからのリクエスト
				rel2 := &entity.Relationship{
					ID:          "rel-2",
					RequesterID: "deleted-user-id",
					ReceiverID:  mainUser.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, rel2); err != nil {
					t.Fatalf("failed to create relationship 2: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *ListFriendRequestsOutput) {
				if output == nil {
					t.Fatal("output is nil")
				}
				// 削除されたユーザーからのリクエストはスキップされる
				if output.TotalCount != 1 {
					t.Errorf("TotalCount = %d, want 1", output.TotalCount)
				}
				if len(output.Requests) != 1 {
					t.Errorf("len(Requests) = %d, want 1", len(output.Requests))
				}
				if len(output.Requests) > 0 && output.Requests[0].Requester.ID != requester1.ID {
					t.Error("Only request from existing user should be returned")
				}
			},
		},
		{
			name: "成功ケース - 送信と受信の混在（typeで適切にフィルタ）",
			input: ListFriendRequestsInput{
				UserID: mainUser.ID,
				Type:   "received",
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// ユーザーを作成
				if err := ur.Create(ctx, mainUser); err != nil {
					t.Fatalf("failed to create main user: %v", err)
				}
				if err := ur.Create(ctx, requester1); err != nil {
					t.Fatalf("failed to create requester1: %v", err)
				}
				if err := ur.Create(ctx, receiver1); err != nil {
					t.Fatalf("failed to create receiver1: %v", err)
				}

				// 受信したリクエスト
				rel1 := &entity.Relationship{
					ID:          "rel-1",
					RequesterID: requester1.ID,
					ReceiverID:  mainUser.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, rel1); err != nil {
					t.Fatalf("failed to create received request: %v", err)
				}

				// 送信したリクエスト（表示されない）
				rel2 := &entity.Relationship{
					ID:          "rel-2",
					RequesterID: mainUser.ID,
					ReceiverID:  receiver1.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := rr.Create(ctx, rel2); err != nil {
					t.Fatalf("failed to create sent request: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *ListFriendRequestsOutput) {
				if output == nil {
					t.Fatal("output is nil")
				}
				// Type="received"なので受信したリクエストのみ表示
				if output.TotalCount != 1 {
					t.Errorf("TotalCount = %d, want 1", output.TotalCount)
				}
				if len(output.Requests) != 1 {
					t.Errorf("len(Requests) = %d, want 1", len(output.Requests))
				}
				if len(output.Requests) > 0 {
					req := output.Requests[0]
					if req.Relationship.ReceiverID != mainUser.ID {
						t.Error("Should only return received requests")
					}
					if req.Requester == nil {
						t.Error("Requester should not be nil for received requests")
					}
				}
			},
		},
		{
			name: "エラー - ユーザーIDが空",
			input: ListFriendRequestsInput{
				UserID: "",
				Type:   "received",
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// セットアップ不要
			},
			wantErr: true,
			errMsg:  "ユーザーIDは必須です",
		},
		{
			name: "エラー - 無効なタイプ",
			input: ListFriendRequestsInput{
				UserID: mainUser.ID,
				Type:   "invalid",
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// セットアップ不要
			},
			wantErr: true,
			errMsg:  "タイプは 'received' または 'sent' である必要があります",
		},
		{
			name: "エラー - ユーザーが存在しない",
			input: ListFriendRequestsInput{
				UserID: "nonexistent-user",
				Type:   "received",
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// ユーザーを作成しない
			},
			wantErr: true,
			errMsg:  "ユーザーが見つかりません",
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
			uc := NewListFriendRequestsUseCase(relationshipRepo, userRepo)
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
					tt.checkFunc(t, output)
				}
			}
		})
	}
}

