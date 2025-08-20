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

func TestNewAcceptFriendRequestUseCase(t *testing.T) {
	relationshipRepo := memory.NewRelationshipRepository()
	userRepo := memory.NewUserRepository()

	uc := NewAcceptFriendRequestUseCase(relationshipRepo, userRepo)

	if uc == nil {
		t.Fatal("NewAcceptFriendRequestUseCase returned nil")
	}
	if uc.relationshipRepo == nil {
		t.Error("relationshipRepo is nil")
	}
	if uc.userRepo == nil {
		t.Error("userRepo is nil")
	}
}

func TestAcceptFriendRequestUseCase_Execute_Success(t *testing.T) {
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

	// 承認待ちの友達リクエストを作成
	pendingRequest := &entity.Relationship{
		ID:          "rel1",
		RequesterID: user1.ID,
		ReceiverID:  user2.ID,
		Status:      valueobject.RelationshipStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := relationshipRepo.Create(ctx, pendingRequest); err != nil {
		t.Fatalf("failed to create pending request: %v", err)
	}

	// UseCaseを作成して実行
	uc := NewAcceptFriendRequestUseCase(relationshipRepo, userRepo)
	input := AcceptFriendRequestInput{
		RelationshipID: pendingRequest.ID,
		ReceiverID:     user2.ID,
	}

	output, err := uc.Execute(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 出力の検証
	if output == nil {
		t.Fatal("output is nil")
	}
	if output.Relationship == nil {
		t.Fatal("Relationship is nil")
	}
	if output.Relationship.Status != valueobject.RelationshipStatusAccepted {
		t.Errorf("Status = %v, want %v", output.Relationship.Status, valueobject.RelationshipStatusAccepted)
	}
	if output.Relationship.RequesterID != user1.ID {
		t.Errorf("RequesterID = %v, want %v", output.Relationship.RequesterID, user1.ID)
	}
	if output.Relationship.ReceiverID != user2.ID {
		t.Errorf("ReceiverID = %v, want %v", output.Relationship.ReceiverID, user2.ID)
	}

	// リポジトリから再取得して確認
	updatedRelationship, err := relationshipRepo.FindByID(ctx, pendingRequest.ID)
	if err != nil {
		t.Fatalf("failed to find updated relationship: %v", err)
	}
	if updatedRelationship.Status != valueobject.RelationshipStatusAccepted {
		t.Errorf("Updated status = %v, want %v", updatedRelationship.Status, valueobject.RelationshipStatusAccepted)
	}
}

func TestAcceptFriendRequestUseCase_Execute_EmptyRelationshipID(t *testing.T) {
	ctx := context.Background()

	relationshipRepo := memory.NewRelationshipRepository()
	userRepo := memory.NewUserRepository()

	// テスト用ユーザーを作成
	user := &entity.User{
		ID:           "user1",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	uc := NewAcceptFriendRequestUseCase(relationshipRepo, userRepo)
	input := AcceptFriendRequestInput{
		RelationshipID: "",
		ReceiverID:     user.ID,
	}

	_, err := uc.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for empty relationship ID but got nil")
		return
	}
	if !strings.Contains(err.Error(), "関係IDは必須です") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAcceptFriendRequestUseCase_Execute_EmptyReceiverID(t *testing.T) {
	ctx := context.Background()

	relationshipRepo := memory.NewRelationshipRepository()
	userRepo := memory.NewUserRepository()

	uc := NewAcceptFriendRequestUseCase(relationshipRepo, userRepo)
	input := AcceptFriendRequestInput{
		RelationshipID: "rel1",
		ReceiverID:     "",
	}

	_, err := uc.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for empty receiver ID but got nil")
		return
	}
	if !strings.Contains(err.Error(), "承認者IDは必須です") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAcceptFriendRequestUseCase_Execute_NonExistentReceiver(t *testing.T) {
	ctx := context.Background()

	relationshipRepo := memory.NewRelationshipRepository()
	userRepo := memory.NewUserRepository()

	uc := NewAcceptFriendRequestUseCase(relationshipRepo, userRepo)
	input := AcceptFriendRequestInput{
		RelationshipID: "rel1",
		ReceiverID:     "nonexistent",
	}

	_, err := uc.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for non-existent receiver but got nil")
		return
	}
	if !strings.Contains(err.Error(), "承認者が見つかりません") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAcceptFriendRequestUseCase_Execute_NonExistentRelationship(t *testing.T) {
	ctx := context.Background()

	relationshipRepo := memory.NewRelationshipRepository()
	userRepo := memory.NewUserRepository()

	// テスト用ユーザーを作成
	user := &entity.User{
		ID:           "user1",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	uc := NewAcceptFriendRequestUseCase(relationshipRepo, userRepo)
	input := AcceptFriendRequestInput{
		RelationshipID: "nonexistent",
		ReceiverID:     user.ID,
	}

	_, err := uc.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for non-existent relationship but got nil")
		return
	}
	if !strings.Contains(err.Error(), "友達リクエストが見つかりません") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAcceptFriendRequestUseCase_Execute_UnauthorizedUser(t *testing.T) {
	ctx := context.Background()

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

	// user1からuser2への友達リクエスト
	pendingRequest := &entity.Relationship{
		ID:          "rel1",
		RequesterID: user1.ID,
		ReceiverID:  user2.ID,
		Status:      valueobject.RelationshipStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := relationshipRepo.Create(ctx, pendingRequest); err != nil {
		t.Fatalf("failed to create pending request: %v", err)
	}

	// user3（関係のない第三者）が承認を試みる
	uc := NewAcceptFriendRequestUseCase(relationshipRepo, userRepo)
	input := AcceptFriendRequestInput{
		RelationshipID: pendingRequest.ID,
		ReceiverID:     user3.ID,
	}

	_, err := uc.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for unauthorized user but got nil")
		return
	}
	if !strings.Contains(err.Error(), "このリクエストを承認する権限がありません") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAcceptFriendRequestUseCase_Execute_AlreadyAccepted(t *testing.T) {
	ctx := context.Background()

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

	// 既に承認済みの友達リクエスト
	acceptedRequest := &entity.Relationship{
		ID:          "rel1",
		RequesterID: user1.ID,
		ReceiverID:  user2.ID,
		Status:      valueobject.RelationshipStatusAccepted,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := relationshipRepo.Create(ctx, acceptedRequest); err != nil {
		t.Fatalf("failed to create accepted request: %v", err)
	}

	uc := NewAcceptFriendRequestUseCase(relationshipRepo, userRepo)
	input := AcceptFriendRequestInput{
		RelationshipID: acceptedRequest.ID,
		ReceiverID:     user2.ID,
	}

	_, err := uc.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for already accepted request but got nil")
		return
	}
	if !strings.Contains(err.Error(), "既に承認済みの友達リクエストです") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAcceptFriendRequestUseCase_Execute_RejectedRequest(t *testing.T) {
	ctx := context.Background()

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

	// 拒否済みの友達リクエスト
	rejectedRequest := &entity.Relationship{
		ID:          "rel1",
		RequesterID: user1.ID,
		ReceiverID:  user2.ID,
		Status:      valueobject.RelationshipStatusRejected,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := relationshipRepo.Create(ctx, rejectedRequest); err != nil {
		t.Fatalf("failed to create rejected request: %v", err)
	}

	uc := NewAcceptFriendRequestUseCase(relationshipRepo, userRepo)
	input := AcceptFriendRequestInput{
		RelationshipID: rejectedRequest.ID,
		ReceiverID:     user2.ID,
	}

	_, err := uc.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for rejected request but got nil")
		return
	}
	if !strings.Contains(err.Error(), "拒否済みの友達リクエストは承認できません") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAcceptFriendRequestUseCase_Execute_BlockedRelationship(t *testing.T) {
	ctx := context.Background()

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

	// ブロック関係
	blockedRelationship := &entity.Relationship{
		ID:          "rel1",
		RequesterID: user1.ID,
		ReceiverID:  user2.ID,
		Status:      valueobject.RelationshipStatusBlocked,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := relationshipRepo.Create(ctx, blockedRelationship); err != nil {
		t.Fatalf("failed to create blocked relationship: %v", err)
	}

	uc := NewAcceptFriendRequestUseCase(relationshipRepo, userRepo)
	input := AcceptFriendRequestInput{
		RelationshipID: blockedRelationship.ID,
		ReceiverID:     user2.ID,
	}

	_, err := uc.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for blocked relationship but got nil")
		return
	}
	if !strings.Contains(err.Error(), "ブロック関係のリクエストは承認できません") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAcceptFriendRequestUseCase_Execute_RequesterNotFound(t *testing.T) {
	ctx := context.Background()

	relationshipRepo := memory.NewRelationshipRepository()
	userRepo := memory.NewUserRepository()

	// 受信者のみを作成（送信者は存在しない）
	receiver := &entity.User{
		ID:           "user2",
		Username:     "bob",
		Email:        "bob@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := userRepo.Create(ctx, receiver); err != nil {
		t.Fatalf("failed to create receiver: %v", err)
	}

	// 送信者が存在しないリクエスト（データ不整合のシミュレーション）
	orphanRequest := &entity.Relationship{
		ID:          "rel1",
		RequesterID: "nonexistent_user",
		ReceiverID:  receiver.ID,
		Status:      valueobject.RelationshipStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := relationshipRepo.Create(ctx, orphanRequest); err != nil {
		t.Fatalf("failed to create orphan request: %v", err)
	}

	uc := NewAcceptFriendRequestUseCase(relationshipRepo, userRepo)
	input := AcceptFriendRequestInput{
		RelationshipID: orphanRequest.ID,
		ReceiverID:     receiver.ID,
	}

	_, err := uc.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for non-existent requester but got nil")
		return
	}
	if !strings.Contains(err.Error(), "リクエスト送信者が見つかりません") {
		t.Errorf("unexpected error message: %v", err)
	}
}
