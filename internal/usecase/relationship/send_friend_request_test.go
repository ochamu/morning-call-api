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

func TestNewSendFriendRequestUseCase(t *testing.T) {
	relationshipRepo := memory.NewRelationshipRepository()
	userRepo := memory.NewUserRepository()

	uc := NewSendFriendRequestUseCase(relationshipRepo, userRepo)

	if uc == nil {
		t.Fatal("NewSendFriendRequestUseCase returned nil")
	}
	if uc.relationshipRepo == nil {
		t.Error("relationshipRepo is nil")
	}
	if uc.userRepo == nil {
		t.Error("userRepo is nil")
	}
}

func TestSendFriendRequestUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	relationshipRepo := memory.NewRelationshipRepository()
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
	user4 := &entity.User{
		ID:           "user4",
		Username:     "david",
		Email:        "david@example.com",
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
	if err := userRepo.Create(ctx, user4); err != nil {
		t.Fatalf("failed to create user4: %v", err)
	}

	// user1とuser3は既に友達
	existingFriendship := &entity.Relationship{
		ID:          "rel1",
		RequesterID: user1.ID,
		ReceiverID:  user3.ID,
		Status:      valueobject.RelationshipStatusAccepted,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := relationshipRepo.Create(ctx, existingFriendship); err != nil {
		t.Fatalf("failed to create existing friendship: %v", err)
	}

	// user1はuser4をブロック
	blockedRelation := &entity.Relationship{
		ID:          "rel2",
		RequesterID: user1.ID,
		ReceiverID:  user4.ID,
		Status:      valueobject.RelationshipStatusBlocked,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := relationshipRepo.Create(ctx, blockedRelation); err != nil {
		t.Fatalf("failed to create blocked relation: %v", err)
	}

	// UseCaseを作成
	uc := NewSendFriendRequestUseCase(relationshipRepo, userRepo)

	tests := []struct {
		name        string
		input       SendFriendRequestInput
		wantErr     bool
		errContains string
	}{
		{
			name: "正常: 新規友達リクエスト送信",
			input: SendFriendRequestInput{
				RequesterID: user1.ID,
				ReceiverID:  user2.ID,
			},
			wantErr: false,
		},
		{
			name: "エラー: リクエスト送信者IDが空",
			input: SendFriendRequestInput{
				RequesterID: "",
				ReceiverID:  user2.ID,
			},
			wantErr:     true,
			errContains: "リクエスト送信者IDは必須です",
		},
		{
			name: "エラー: リクエスト受信者IDが空",
			input: SendFriendRequestInput{
				RequesterID: user1.ID,
				ReceiverID:  "",
			},
			wantErr:     true,
			errContains: "リクエスト受信者IDは必須です",
		},
		{
			name: "エラー: 自分自身に友達リクエスト",
			input: SendFriendRequestInput{
				RequesterID: user1.ID,
				ReceiverID:  user1.ID,
			},
			wantErr:     true,
			errContains: "自分自身に友達リクエストを送ることはできません",
		},
		{
			name: "エラー: リクエスト送信者が存在しない",
			input: SendFriendRequestInput{
				RequesterID: "non_existent_user",
				ReceiverID:  user2.ID,
			},
			wantErr:     true,
			errContains: "リクエスト送信者が見つかりません",
		},
		{
			name: "エラー: リクエスト受信者が存在しない",
			input: SendFriendRequestInput{
				RequesterID: user1.ID,
				ReceiverID:  "non_existent_user",
			},
			wantErr:     true,
			errContains: "リクエスト受信者が見つかりません",
		},
		{
			name: "エラー: 既に友達関係",
			input: SendFriendRequestInput{
				RequesterID: user1.ID,
				ReceiverID:  user3.ID,
			},
			wantErr:     true,
			errContains: "既に友達関係です",
		},
		{
			name: "エラー: ブロックしている相手にリクエスト",
			input: SendFriendRequestInput{
				RequesterID: user1.ID,
				ReceiverID:  user4.ID,
			},
			wantErr:     true,
			errContains: "相手をブロックしているため、友達リクエストを送信できません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(ctx, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error message = %q, want contains %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if output == nil {
					t.Error("output is nil")
					return
				}
				if output.Relationship == nil {
					t.Error("output.Relationship is nil")
					return
				}
				// 関係の検証
				if output.Relationship.RequesterID != tt.input.RequesterID {
					t.Errorf("RequesterID = %q, want %q", output.Relationship.RequesterID, tt.input.RequesterID)
				}
				if output.Relationship.ReceiverID != tt.input.ReceiverID {
					t.Errorf("ReceiverID = %q, want %q", output.Relationship.ReceiverID, tt.input.ReceiverID)
				}
				if output.Relationship.Status != valueobject.RelationshipStatusPending {
					t.Errorf("Status = %v, want %v", output.Relationship.Status, valueobject.RelationshipStatusPending)
				}
			}
		})
	}
}

func TestSendFriendRequestUseCase_Execute_DuplicateRequest(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	relationshipRepo := memory.NewRelationshipRepository()
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

	// UseCaseを作成
	uc := NewSendFriendRequestUseCase(relationshipRepo, userRepo)

	// 1回目のリクエスト送信
	input := SendFriendRequestInput{
		RequesterID: user1.ID,
		ReceiverID:  user2.ID,
	}
	output1, err := uc.Execute(ctx, input)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	if output1 == nil || output1.Relationship == nil {
		t.Fatal("first request returned nil")
	}

	// 2回目の同じリクエスト送信（エラーになるはず）
	output2, err := uc.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for duplicate request but got nil")
		return
	}
	if output2 != nil {
		t.Error("expected nil output for duplicate request")
	}
	if !strings.Contains(err.Error(), "既に友達リクエストを送信済みです") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSendFriendRequestUseCase_Execute_ReverseRequest(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	relationshipRepo := memory.NewRelationshipRepository()
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

	// user2からuser1への既存のリクエスト
	existingRequest := &entity.Relationship{
		ID:          "rel1",
		RequesterID: user2.ID,
		ReceiverID:  user1.ID,
		Status:      valueobject.RelationshipStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := relationshipRepo.Create(ctx, existingRequest); err != nil {
		t.Fatalf("failed to create existing request: %v", err)
	}

	// UseCaseを作成
	uc := NewSendFriendRequestUseCase(relationshipRepo, userRepo)

	// user1からuser2へのリクエスト（逆方向）
	input := SendFriendRequestInput{
		RequesterID: user1.ID,
		ReceiverID:  user2.ID,
	}
	output, err := uc.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for reverse request but got nil")
		return
	}
	if output != nil {
		t.Error("expected nil output for reverse request")
	}
	if !strings.Contains(err.Error(), "相手から既に友達リクエストが送信されています") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSendFriendRequestUseCase_Execute_ResendAfterRejection(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	relationshipRepo := memory.NewRelationshipRepository()
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

	// 拒否されたリクエスト（25時間前）
	rejectedRequest := &entity.Relationship{
		ID:          "rel1",
		RequesterID: user1.ID,
		ReceiverID:  user2.ID,
		Status:      valueobject.RelationshipStatusRejected,
		CreatedAt:   time.Now().Add(-25 * time.Hour),
		UpdatedAt:   time.Now().Add(-25 * time.Hour),
	}
	if err := relationshipRepo.Create(ctx, rejectedRequest); err != nil {
		t.Fatalf("failed to create rejected request: %v", err)
	}

	// UseCaseを作成
	uc := NewSendFriendRequestUseCase(relationshipRepo, userRepo)

	// 24時間後の再送信
	input := SendFriendRequestInput{
		RequesterID: user1.ID,
		ReceiverID:  user2.ID,
	}
	output, err := uc.Execute(ctx, input)
	if err != nil {
		t.Errorf("unexpected error for resend after 24 hours: %v", err)
	}
	if output == nil || output.Relationship == nil {
		t.Fatal("resend returned nil")
	}
	if output.Relationship.Status != valueobject.RelationshipStatusPending {
		t.Errorf("Status = %v, want %v", output.Relationship.Status, valueobject.RelationshipStatusPending)
	}
}

func TestSendFriendRequestUseCase_Execute_ResendTooSoon(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	relationshipRepo := memory.NewRelationshipRepository()
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

	// 拒否されたリクエスト（1時間前）
	rejectedRequest := &entity.Relationship{
		ID:          "rel1",
		RequesterID: user1.ID,
		ReceiverID:  user2.ID,
		Status:      valueobject.RelationshipStatusRejected,
		CreatedAt:   time.Now().Add(-1 * time.Hour),
		UpdatedAt:   time.Now().Add(-1 * time.Hour),
	}
	if err := relationshipRepo.Create(ctx, rejectedRequest); err != nil {
		t.Fatalf("failed to create rejected request: %v", err)
	}

	// UseCaseを作成
	uc := NewSendFriendRequestUseCase(relationshipRepo, userRepo)

	// 24時間以内の再送信（エラーになるはず）
	input := SendFriendRequestInput{
		RequesterID: user1.ID,
		ReceiverID:  user2.ID,
	}
	output, err := uc.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for resend too soon but got nil")
		return
	}
	if output != nil {
		t.Error("expected nil output for resend too soon")
	}
	if !strings.Contains(err.Error(), "24時間後に再送信できます") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSendFriendRequestUseCase_Execute_BlockedByReceiver(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	relationshipRepo := memory.NewRelationshipRepository()
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

	// user2がuser1をブロック
	blockedRelation := &entity.Relationship{
		ID:          "rel1",
		RequesterID: user2.ID,
		ReceiverID:  user1.ID,
		Status:      valueobject.RelationshipStatusBlocked,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := relationshipRepo.Create(ctx, blockedRelation); err != nil {
		t.Fatalf("failed to create blocked relation: %v", err)
	}

	// UseCaseを作成
	uc := NewSendFriendRequestUseCase(relationshipRepo, userRepo)

	// user1からuser2へのリクエスト（ブロックされている）
	input := SendFriendRequestInput{
		RequesterID: user1.ID,
		ReceiverID:  user2.ID,
	}
	output, err := uc.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for blocked by receiver but got nil")
		return
	}
	if output != nil {
		t.Error("expected nil output for blocked by receiver")
	}
	if !strings.Contains(err.Error(), "相手にブロックされているため、友達リクエストを送信できません") {
		t.Errorf("unexpected error message: %v", err)
	}
}
