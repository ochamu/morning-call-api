package relationship

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
	"github.com/ochamu/morning-call-api/pkg/utils"
)

// SendFriendRequestUseCase は友達リクエスト送信のユースケース
type SendFriendRequestUseCase struct {
	relationshipRepo repository.RelationshipRepository
	userRepo         repository.UserRepository
}

// NewSendFriendRequestUseCase は新しい友達リクエスト送信ユースケースを作成する
func NewSendFriendRequestUseCase(
	relationshipRepo repository.RelationshipRepository,
	userRepo repository.UserRepository,
) *SendFriendRequestUseCase {
	return &SendFriendRequestUseCase{
		relationshipRepo: relationshipRepo,
		userRepo:         userRepo,
	}
}

// SendFriendRequestInput は友達リクエスト送信の入力データ
type SendFriendRequestInput struct {
	RequesterID string // リクエスト送信者のユーザーID
	ReceiverID  string // リクエスト受信者のユーザーID
}

// SendFriendRequestOutput は友達リクエスト送信の出力データ
type SendFriendRequestOutput struct {
	Relationship *entity.Relationship
}

// Execute は友達リクエストを送信する
func (uc *SendFriendRequestUseCase) Execute(ctx context.Context, input SendFriendRequestInput) (*SendFriendRequestOutput, error) {
	// 入力値の基本検証
	if input.RequesterID == "" {
		return nil, fmt.Errorf("リクエスト送信者IDは必須です")
	}
	if input.ReceiverID == "" {
		return nil, fmt.Errorf("リクエスト受信者IDは必須です")
	}
	if input.RequesterID == input.ReceiverID {
		return nil, fmt.Errorf("自分自身に友達リクエストを送ることはできません")
	}

	// リクエスト送信者の存在確認
	requester, err := uc.userRepo.FindByID(ctx, input.RequesterID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("リクエスト送信者が見つかりません")
		}
		return nil, fmt.Errorf("リクエスト送信者の確認中にエラーが発生しました: %w", err)
	}

	// リクエスト受信者の存在確認
	receiver, err := uc.userRepo.FindByID(ctx, input.ReceiverID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("リクエスト受信者が見つかりません")
		}
		return nil, fmt.Errorf("リクエスト受信者の確認中にエラーが発生しました: %w", err)
	}

	// 既存の関係を確認
	existingRelationship, err := uc.relationshipRepo.FindByUserPair(ctx, input.RequesterID, input.ReceiverID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("既存の関係確認中にエラーが発生しました: %w", err)
	}

	// 既存の関係がある場合の処理
	if existingRelationship != nil {
		switch existingRelationship.Status {
		case valueobject.RelationshipStatusAccepted:
			return nil, fmt.Errorf("既に友達関係です")
		case valueobject.RelationshipStatusPending:
			// 既に承認待ちのリクエストがある場合
			if existingRelationship.RequesterID == input.RequesterID {
				return nil, fmt.Errorf("既に友達リクエストを送信済みです")
			}
			// 相手から既にリクエストが来ている場合
			return nil, fmt.Errorf("相手から既に友達リクエストが送信されています。リクエストを承認してください")
		case valueobject.RelationshipStatusBlocked:
			// どちらかがブロックしている場合
			if existingRelationship.RequesterID == input.RequesterID {
				return nil, fmt.Errorf("相手をブロックしているため、友達リクエストを送信できません")
			}
			return nil, fmt.Errorf("相手にブロックされているため、友達リクエストを送信できません")
		case valueobject.RelationshipStatusRejected:
			// 以前に拒否されたリクエストの場合
			if existingRelationship.RequesterID == input.RequesterID {
				// 同じ方向のリクエストで拒否済みの場合、再送信を試みる
				now := time.Now()
				// 拒否から24時間経過していない場合はエラー
				if existingRelationship.UpdatedAt.Add(24 * time.Hour).After(now) {
					return nil, fmt.Errorf("友達リクエストが拒否されました。24時間後に再送信できます")
				}
				// 24時間経過している場合は再送信
				if reason := existingRelationship.Resend(); reason.IsNG() {
					return nil, fmt.Errorf("友達リクエストの再送信に失敗しました: %s", reason)
				}
				// リポジトリで更新
				if err := uc.relationshipRepo.Update(ctx, existingRelationship); err != nil {
					return nil, fmt.Errorf("友達リクエストの再送信に失敗しました: %w", err)
				}
				return &SendFriendRequestOutput{
					Relationship: existingRelationship,
				}, nil
			}
			// 逆方向のリクエストが拒否されている場合は新規作成を許可
		}
	}

	// UUIDを生成
	id, err := utils.GenerateUUID()
	if err != nil {
		return nil, fmt.Errorf("ID生成に失敗しました: %w", err)
	}

	// 友達関係エンティティを作成
	relationship, reason := entity.NewRelationship(id, requester.ID, receiver.ID)
	if reason.IsNG() {
		return nil, fmt.Errorf("友達リクエストの作成に失敗しました: %s", reason)
	}

	// リポジトリに保存
	if err := uc.relationshipRepo.Create(ctx, relationship); err != nil {
		// 重複エラーの場合は分かりやすいメッセージにする
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, fmt.Errorf("既に友達関係が存在します")
		}
		return nil, fmt.Errorf("友達リクエストの送信に失敗しました: %w", err)
	}

	return &SendFriendRequestOutput{
		Relationship: relationship,
	}, nil
}
