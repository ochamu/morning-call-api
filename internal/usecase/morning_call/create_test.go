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

func TestNewCreateUseCase(t *testing.T) {
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()
	relationshipRepo := memory.NewRelationshipRepository()

	uc := NewCreateUseCase(morningCallRepo, userRepo, relationshipRepo)

	if uc == nil {
		t.Fatal("NewCreateUseCase returned nil")
	}
	if uc.morningCallRepo == nil {
		t.Error("morningCallRepo is nil")
	}
	if uc.userRepo == nil {
		t.Error("userRepo is nil")
	}
	if uc.relationshipRepo == nil {
		t.Error("relationshipRepo is nil")
	}
}

func TestCreateUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()
	relationshipRepo := memory.NewRelationshipRepository()

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

	// user1とuser2を友達関係にする
	friendship := &entity.Relationship{
		ID:          "rel1",
		RequesterID: user1.ID,
		ReceiverID:  user2.ID,
		Status:      valueobject.RelationshipStatusAccepted,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := relationshipRepo.Create(ctx, friendship); err != nil {
		t.Fatalf("failed to create friendship: %v", err)
	}

	// user1とuser3をブロック関係にする
	blockedRelation := &entity.Relationship{
		ID:          "rel2",
		RequesterID: user1.ID,
		ReceiverID:  user3.ID,
		Status:      valueobject.RelationshipStatusBlocked,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := relationshipRepo.Create(ctx, blockedRelation); err != nil {
		t.Fatalf("failed to create blocked relation: %v", err)
	}

	// 将来の時刻を設定
	futureTime := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name    string
		input   CreateInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功ケース",
			input: CreateInput{
				SenderID:      user1.ID,
				ReceiverID:    user2.ID,
				ScheduledTime: futureTime,
				Message:       "おはよう！今日も頑張ろう！",
			},
			wantErr: false,
		},
		{
			name: "送信者IDが空",
			input: CreateInput{
				SenderID:      "",
				ReceiverID:    user2.ID,
				ScheduledTime: futureTime,
				Message:       "テストメッセージ",
			},
			wantErr: true,
			errMsg:  "送信者IDは必須です",
		},
		{
			name: "受信者IDが空",
			input: CreateInput{
				SenderID:      user1.ID,
				ReceiverID:    "",
				ScheduledTime: futureTime,
				Message:       "テストメッセージ",
			},
			wantErr: true,
			errMsg:  "受信者IDは必須です",
		},
		{
			name: "自分自身への送信",
			input: CreateInput{
				SenderID:      user1.ID,
				ReceiverID:    user1.ID,
				ScheduledTime: futureTime,
				Message:       "テストメッセージ",
			},
			wantErr: true,
			errMsg:  "自分自身にモーニングコールを設定することはできません",
		},
		{
			name: "スケジュール時刻が未設定",
			input: CreateInput{
				SenderID:   user1.ID,
				ReceiverID: user2.ID,
				Message:    "テストメッセージ",
			},
			wantErr: true,
			errMsg:  "スケジュール時刻は必須です",
		},
		{
			name: "存在しない送信者",
			input: CreateInput{
				SenderID:      "nonexistent",
				ReceiverID:    user2.ID,
				ScheduledTime: futureTime,
				Message:       "テストメッセージ",
			},
			wantErr: true,
			errMsg:  "送信者が見つかりません",
		},
		{
			name: "存在しない受信者",
			input: CreateInput{
				SenderID:      user1.ID,
				ReceiverID:    "nonexistent",
				ScheduledTime: futureTime,
				Message:       "テストメッセージ",
			},
			wantErr: true,
			errMsg:  "受信者が見つかりません",
		},
		{
			name: "友達関係にないユーザー",
			input: CreateInput{
				SenderID:      user2.ID,
				ReceiverID:    user3.ID,
				ScheduledTime: futureTime,
				Message:       "テストメッセージ",
			},
			wantErr: true,
			errMsg:  "友達関係にないユーザーにはモーニングコールを設定できません",
		},
		{
			name: "ブロックされているユーザー",
			input: CreateInput{
				SenderID:      user1.ID,
				ReceiverID:    user3.ID,
				ScheduledTime: futureTime,
				Message:       "テストメッセージ",
			},
			wantErr: true,
			errMsg:  "友達関係にないユーザーにはモーニングコールを設定できません", // ブロック関係はAreFriendsでfalseを返すため
		},
		{
			name: "過去の時刻でのスケジュール",
			input: CreateInput{
				SenderID:      user1.ID,
				ReceiverID:    user2.ID,
				ScheduledTime: time.Now().Add(-1 * time.Hour),
				Message:       "テストメッセージ",
			},
			wantErr: true,
			errMsg:  "アラーム時刻は現在時刻より後である必要があります",
		},
		{
			name: "30日以上先の時刻でのスケジュール",
			input: CreateInput{
				SenderID:      user1.ID,
				ReceiverID:    user2.ID,
				ScheduledTime: time.Now().Add(31 * 24 * time.Hour),
				Message:       "テストメッセージ",
			},
			wantErr: true,
			errMsg:  "アラーム時刻は30日以内で設定してください",
		},
		{
			name: "メッセージが長すぎる",
			input: CreateInput{
				SenderID:      user1.ID,
				ReceiverID:    user2.ID,
				ScheduledTime: futureTime.Add(10 * time.Hour), // 他のテストケースと時刻をずらす
				Message:       strings.Repeat("あ", 501),       // 501文字
			},
			wantErr: true,
			errMsg:  "メッセージは500文字以内で入力してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各テストケースで新しいUseCaseインスタンスを作成
			uc := NewCreateUseCase(morningCallRepo, userRepo, relationshipRepo)
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
						// 作成されたモーニングコールの検証
						mc := output.MorningCall
						if mc.ID == "" {
							t.Error("MorningCall.ID is empty")
						}
						if mc.SenderID != tt.input.SenderID {
							t.Errorf("MorningCall.SenderID = %v, want %v", mc.SenderID, tt.input.SenderID)
						}
						if mc.ReceiverID != tt.input.ReceiverID {
							t.Errorf("MorningCall.ReceiverID = %v, want %v", mc.ReceiverID, tt.input.ReceiverID)
						}
						if !mc.ScheduledTime.Equal(tt.input.ScheduledTime) {
							t.Errorf("MorningCall.ScheduledTime = %v, want %v", mc.ScheduledTime, tt.input.ScheduledTime)
						}
						if mc.Message != tt.input.Message {
							t.Errorf("MorningCall.Message = %v, want %v", mc.Message, tt.input.Message)
						}
						if mc.Status != valueobject.MorningCallStatusScheduled {
							t.Errorf("MorningCall.Status = %v, want %v", mc.Status, valueobject.MorningCallStatusScheduled)
						}
					}
				}
			}
		})
	}
}

func TestCreateUseCase_Execute_DuplicateMorningCall(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()
	relationshipRepo := memory.NewRelationshipRepository()

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

	// ユーザーと友達関係をリポジトリに追加
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	if err := userRepo.Create(ctx, user2); err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	friendship := &entity.Relationship{
		ID:          "rel1",
		RequesterID: user1.ID,
		ReceiverID:  user2.ID,
		Status:      valueobject.RelationshipStatusAccepted,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := relationshipRepo.Create(ctx, friendship); err != nil {
		t.Fatalf("failed to create friendship: %v", err)
	}

	// 既存のアクティブなモーニングコールを作成
	scheduledTime := time.Now().Add(24 * time.Hour)
	existingCall := &entity.MorningCall{
		ID:            "mc1",
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: scheduledTime,
		Message:       "既存のモーニングコール",
		Status:        valueobject.MorningCallStatusScheduled,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := morningCallRepo.Create(ctx, existingCall); err != nil {
		t.Fatalf("failed to create existing morning call: %v", err)
	}

	uc := NewCreateUseCase(morningCallRepo, userRepo, relationshipRepo)

	// 同じ時刻付近（30秒後）に新しいモーニングコールを作成しようとする
	input := CreateInput{
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: scheduledTime.Add(30 * time.Second),
		Message:       "新しいモーニングコール",
	}

	_, err := uc.Execute(ctx, input)
	if err == nil {
		t.Error("expected error for duplicate morning call but got nil")
	} else if !strings.Contains(err.Error(), "同じ時刻付近に既にモーニングコールが設定されています") {
		t.Errorf("error message = %v, want contains '同じ時刻付近に既にモーニングコールが設定されています'", err.Error())
	}

	// 1分以上離れた時刻なら成功する
	input2 := CreateInput{
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: scheduledTime.Add(2 * time.Hour),
		Message:       "新しいモーニングコール",
	}

	output, err := uc.Execute(ctx, input2)
	if err != nil {
		t.Errorf("unexpected error for non-duplicate morning call: %v", err)
	}
	if output == nil || output.MorningCall == nil {
		t.Error("expected successful creation but got nil output")
	}
}

func TestCreateUseCase_Execute_BidirectionalFriendship(t *testing.T) {
	ctx := context.Background()

	// テスト用のリポジトリを作成
	morningCallRepo := memory.NewMorningCallRepository()
	userRepo := memory.NewUserRepository()
	relationshipRepo := memory.NewRelationshipRepository()

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

	// user2からuser1への友達関係（逆方向）
	friendship := &entity.Relationship{
		ID:          "rel1",
		RequesterID: user2.ID,
		ReceiverID:  user1.ID,
		Status:      valueobject.RelationshipStatusAccepted,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := relationshipRepo.Create(ctx, friendship); err != nil {
		t.Fatalf("failed to create friendship: %v", err)
	}

	uc := NewCreateUseCase(morningCallRepo, userRepo, relationshipRepo)

	// user1からuser2へのモーニングコール（友達関係は逆方向だが、双方向として扱われるべき）
	input := CreateInput{
		SenderID:      user1.ID,
		ReceiverID:    user2.ID,
		ScheduledTime: time.Now().Add(24 * time.Hour),
		Message:       "おはよう！",
	}

	output, err := uc.Execute(ctx, input)
	if err != nil {
		t.Errorf("unexpected error for bidirectional friendship: %v", err)
	}
	if output == nil || output.MorningCall == nil {
		t.Error("expected successful creation with bidirectional friendship")
	}
}
