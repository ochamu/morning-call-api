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

func TestNewListFriendsUseCase(t *testing.T) {
	relationshipRepo := memory.NewRelationshipRepository()
	userRepo := memory.NewUserRepository()

	uc := NewListFriendsUseCase(relationshipRepo, userRepo)

	if uc == nil {
		t.Fatal("NewListFriendsUseCase returned nil")
	}
	if uc.relationshipRepo == nil {
		t.Error("relationshipRepo is nil")
	}
	if uc.userRepo == nil {
		t.Error("userRepo is nil")
	}
}

func TestListFriendsUseCase_Execute(t *testing.T) {
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

	user4 := &entity.User{
		ID:           "user4-id",
		Username:     "user4",
		Email:        "user4@example.com",
		PasswordHash: "hashed",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	tests := []struct {
		name      string
		input     ListFriendsInput
		setup     func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository)
		wantErr   bool
		errMsg    string
		checkFunc func(t *testing.T, output *ListFriendsOutput)
	}{
		{
			name: "成功ケース - 複数の友達",
			input: ListFriendsInput{
				UserID: user1.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// ユーザーを作成
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
				if err := ur.Create(ctx, user2); err != nil {
					t.Fatalf("failed to create user2: %v", err)
				}
				if err := ur.Create(ctx, user3); err != nil {
					t.Fatalf("failed to create user3: %v", err)
				}

				// user1がuser2にリクエストを送信して承認済み
				rel1 := &entity.Relationship{
					ID:          "rel-1",
					RequesterID: user1.ID,
					ReceiverID:  user2.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now().Add(-48 * time.Hour),
					UpdatedAt:   time.Now().Add(-24 * time.Hour),
				}
				if err := rr.Create(ctx, rel1); err != nil {
					t.Fatalf("failed to create relationship 1: %v", err)
				}

				// user3がuser1にリクエストを送信して承認済み
				rel2 := &entity.Relationship{
					ID:          "rel-2",
					RequesterID: user3.ID,
					ReceiverID:  user1.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now().Add(-36 * time.Hour),
					UpdatedAt:   time.Now().Add(-12 * time.Hour),
				}
				if err := rr.Create(ctx, rel2); err != nil {
					t.Fatalf("failed to create relationship 2: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *ListFriendsOutput) {
				if output == nil {
					t.Fatal("output is nil")
				}
				if output.TotalCount != 2 {
					t.Errorf("TotalCount = %d, want 2", output.TotalCount)
				}
				if len(output.Friends) != 2 {
					t.Errorf("Friends count = %d, want 2", len(output.Friends))
				}

				// 友達のユーザーIDを確認
				friendIDs := make(map[string]bool)
				for _, friend := range output.Friends {
					friendIDs[friend.User.ID] = true
				}
				if !friendIDs[user2.ID] {
					t.Error("user2 should be in friends list")
				}
				if !friendIDs[user3.ID] {
					t.Error("user3 should be in friends list")
				}

				// IsRequesterフラグの確認
				for _, friend := range output.Friends {
					if friend.User.ID == user2.ID {
						if !friend.IsRequester {
							t.Error("user1 should be requester for user2")
						}
					}
					if friend.User.ID == user3.ID {
						if friend.IsRequester {
							t.Error("user1 should not be requester for user3")
						}
					}
				}
			},
		},
		{
			name: "成功ケース - 友達がいない",
			input: ListFriendsInput{
				UserID: user1.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *ListFriendsOutput) {
				if output == nil {
					t.Fatal("output is nil")
				}
				if output.TotalCount != 0 {
					t.Errorf("TotalCount = %d, want 0", output.TotalCount)
				}
				if len(output.Friends) != 0 {
					t.Errorf("Friends count = %d, want 0", len(output.Friends))
				}
			},
		},
		{
			name: "成功ケース - ペンディング、拒否、ブロックは含まれない",
			input: ListFriendsInput{
				UserID: user1.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// ユーザーを作成
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
				if err := ur.Create(ctx, user2); err != nil {
					t.Fatalf("failed to create user2: %v", err)
				}
				if err := ur.Create(ctx, user3); err != nil {
					t.Fatalf("failed to create user3: %v", err)
				}
				if err := ur.Create(ctx, user4); err != nil {
					t.Fatalf("failed to create user4: %v", err)
				}

				// user1とuser2は友達（承認済み）
				rel1 := &entity.Relationship{
					ID:          "rel-1",
					RequesterID: user1.ID,
					ReceiverID:  user2.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now().Add(-24 * time.Hour),
					UpdatedAt:   time.Now().Add(-12 * time.Hour),
				}
				if err := rr.Create(ctx, rel1); err != nil {
					t.Fatalf("failed to create accepted relationship: %v", err)
				}

				// user1からuser3へのペンディングリクエスト
				rel2 := &entity.Relationship{
					ID:          "rel-2",
					RequesterID: user1.ID,
					ReceiverID:  user3.ID,
					Status:      valueobject.RelationshipStatusPending,
					CreatedAt:   time.Now().Add(-6 * time.Hour),
					UpdatedAt:   time.Now().Add(-6 * time.Hour),
				}
				if err := rr.Create(ctx, rel2); err != nil {
					t.Fatalf("failed to create pending relationship: %v", err)
				}

				// user4からuser1への拒否済みリクエスト
				rel3 := &entity.Relationship{
					ID:          "rel-3",
					RequesterID: user4.ID,
					ReceiverID:  user1.ID,
					Status:      valueobject.RelationshipStatusRejected,
					CreatedAt:   time.Now().Add(-3 * time.Hour),
					UpdatedAt:   time.Now().Add(-1 * time.Hour),
				}
				if err := rr.Create(ctx, rel3); err != nil {
					t.Fatalf("failed to create rejected relationship: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *ListFriendsOutput) {
				if output == nil {
					t.Fatal("output is nil")
				}
				// 承認済みの友達のみが含まれることを確認
				if output.TotalCount != 1 {
					t.Errorf("TotalCount = %d, want 1", output.TotalCount)
				}
				if len(output.Friends) != 1 {
					t.Errorf("Friends count = %d, want 1", len(output.Friends))
				}
				if output.Friends[0].User.ID != user2.ID {
					t.Errorf("Friend ID = %s, want %s", output.Friends[0].User.ID, user2.ID)
				}
			},
		},
		{
			name: "成功ケース - 削除されたユーザーとの友達関係は表示されない",
			input: ListFriendsInput{
				UserID: user1.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// user1とuser2を作成
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
				if err := ur.Create(ctx, user2); err != nil {
					t.Fatalf("failed to create user2: %v", err)
				}
				if err := ur.Create(ctx, user3); err != nil {
					t.Fatalf("failed to create user3: %v", err)
				}

				// user1とuser2は友達
				rel1 := &entity.Relationship{
					ID:          "rel-1",
					RequesterID: user1.ID,
					ReceiverID:  user2.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now().Add(-24 * time.Hour),
					UpdatedAt:   time.Now().Add(-12 * time.Hour),
				}
				if err := rr.Create(ctx, rel1); err != nil {
					t.Fatalf("failed to create relationship 1: %v", err)
				}

				// user1とuser3も友達
				rel2 := &entity.Relationship{
					ID:          "rel-2",
					RequesterID: user1.ID,
					ReceiverID:  user3.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now().Add(-24 * time.Hour),
					UpdatedAt:   time.Now().Add(-12 * time.Hour),
				}
				if err := rr.Create(ctx, rel2); err != nil {
					t.Fatalf("failed to create relationship 2: %v", err)
				}

				// user3を削除
				if err := ur.Delete(ctx, user3.ID); err != nil {
					t.Fatalf("failed to delete user3: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *ListFriendsOutput) {
				if output == nil {
					t.Fatal("output is nil")
				}
				// 削除されたuser3は表示されない
				if output.TotalCount != 1 {
					t.Errorf("TotalCount = %d, want 1", output.TotalCount)
				}
				if len(output.Friends) != 1 {
					t.Errorf("Friends count = %d, want 1", len(output.Friends))
				}
				if output.Friends[0].User.ID != user2.ID {
					t.Errorf("Friend ID = %s, want %s", output.Friends[0].User.ID, user2.ID)
				}
			},
		},
		{
			name: "エラー - ユーザーIDが空",
			input: ListFriendsInput{
				UserID: "",
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// セットアップ不要
			},
			wantErr: true,
			errMsg:  "ユーザーIDは必須です",
		},
		{
			name: "エラー - ユーザーが存在しない",
			input: ListFriendsInput{
				UserID: "nonexistent",
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// セットアップ不要
			},
			wantErr: true,
			errMsg:  "ユーザーが見つかりません",
		},
		{
			name: "成功ケース - ブロック関係は表示されない",
			input: ListFriendsInput{
				UserID: user1.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				// ユーザーを作成
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
				if err := ur.Create(ctx, user2); err != nil {
					t.Fatalf("failed to create user2: %v", err)
				}
				if err := ur.Create(ctx, user3); err != nil {
					t.Fatalf("failed to create user3: %v", err)
				}

				// user1とuser2は友達
				rel1 := &entity.Relationship{
					ID:          "rel-1",
					RequesterID: user1.ID,
					ReceiverID:  user2.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   time.Now().Add(-24 * time.Hour),
					UpdatedAt:   time.Now().Add(-12 * time.Hour),
				}
				if err := rr.Create(ctx, rel1); err != nil {
					t.Fatalf("failed to create accepted relationship: %v", err)
				}

				// user1がuser3をブロック
				rel2 := &entity.Relationship{
					ID:          "rel-2",
					RequesterID: user1.ID,
					ReceiverID:  user3.ID,
					Status:      valueobject.RelationshipStatusBlocked,
					CreatedAt:   time.Now().Add(-6 * time.Hour),
					UpdatedAt:   time.Now().Add(-6 * time.Hour),
				}
				if err := rr.Create(ctx, rel2); err != nil {
					t.Fatalf("failed to create blocked relationship: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *ListFriendsOutput) {
				if output == nil {
					t.Fatal("output is nil")
				}
				// ブロック関係は表示されない
				if output.TotalCount != 1 {
					t.Errorf("TotalCount = %d, want 1", output.TotalCount)
				}
				if len(output.Friends) != 1 {
					t.Errorf("Friends count = %d, want 1", len(output.Friends))
				}
				if output.Friends[0].User.ID != user2.ID {
					t.Errorf("Friend ID = %s, want %s", output.Friends[0].User.ID, user2.ID)
				}
			},
		},
		{
			name: "成功ケース - 友達になった日時フォーマットの確認",
			input: ListFriendsInput{
				UserID: user1.ID,
			},
			setup: func(t *testing.T, rr *memory.RelationshipRepository, ur *memory.UserRepository) {
				if err := ur.Create(ctx, user1); err != nil {
					t.Fatalf("failed to create user1: %v", err)
				}
				if err := ur.Create(ctx, user2); err != nil {
					t.Fatalf("failed to create user2: %v", err)
				}

				// 特定の日時で友達関係を作成
				specificTime := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
				rel := &entity.Relationship{
					ID:          "rel-1",
					RequesterID: user1.ID,
					ReceiverID:  user2.ID,
					Status:      valueobject.RelationshipStatusAccepted,
					CreatedAt:   specificTime.Add(-24 * time.Hour),
					UpdatedAt:   specificTime,
				}
				if err := rr.Create(ctx, rel); err != nil {
					t.Fatalf("failed to create relationship: %v", err)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *ListFriendsOutput) {
				if output == nil {
					t.Fatal("output is nil")
				}
				if len(output.Friends) != 1 {
					t.Fatalf("Friends count = %d, want 1", len(output.Friends))
				}
				// FriendSinceのフォーマットを確認
				friendSince := output.Friends[0].FriendSince
				// YYYY-MM-DD HH:MM:SS形式であることを確認
				if !strings.Contains(friendSince, "2024-01-15") {
					t.Errorf("FriendSince does not contain expected date: %s", friendSince)
				}
				if !strings.Contains(friendSince, ":") {
					t.Errorf("FriendSince does not contain time separator: %s", friendSince)
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
			uc := NewListFriendsUseCase(relationshipRepo, userRepo)
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
