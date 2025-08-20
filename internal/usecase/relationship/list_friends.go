package relationship

import (
	"context"
	"errors"
	"fmt"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// ListFriendsUseCase は友達リスト取得のユースケース
type ListFriendsUseCase struct {
	relationshipRepo repository.RelationshipRepository
	userRepo         repository.UserRepository
}

// NewListFriendsUseCase は新しい友達リスト取得ユースケースを作成する
func NewListFriendsUseCase(
	relationshipRepo repository.RelationshipRepository,
	userRepo repository.UserRepository,
) *ListFriendsUseCase {
	return &ListFriendsUseCase{
		relationshipRepo: relationshipRepo,
		userRepo:         userRepo,
	}
}

// ListFriendsInput は友達リスト取得の入力データ
type ListFriendsInput struct {
	UserID string // 友達リストを取得するユーザーID
}

// FriendInfo は友達情報
type FriendInfo struct {
	User         *entity.User         // 友達のユーザー情報
	Relationship *entity.Relationship // 関係情報
	IsRequester  bool                 // 自分がリクエスト送信者かどうか
	FriendSince  string               // 友達になった日時（文字列表現）
}

// ListFriendsOutput は友達リスト取得の出力データ
type ListFriendsOutput struct {
	Friends    []FriendInfo // 友達リスト
	TotalCount int          // 総友達数
}

// Execute は友達リストを取得する
func (uc *ListFriendsUseCase) Execute(ctx context.Context, input ListFriendsInput) (*ListFriendsOutput, error) {
	// 入力値の基本検証
	if input.UserID == "" {
		return nil, fmt.Errorf("ユーザーIDは必須です")
	}

	// ユーザーの存在確認
	user, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("ユーザーが見つかりません")
		}
		return nil, fmt.Errorf("ユーザーの確認中にエラーが発生しました: %w", err)
	}

	// ユーザーに関連する関係をすべて取得
	// 現時点では全件取得（offset: 0, limit: 1000）
	// 将来的にはページネーションパラメータを入力に追加することも検討
	relationships, err := uc.relationshipRepo.FindByUserID(ctx, user.ID, 0, 1000)
	if err != nil {
		// NotFoundの場合は空リストを返す
		if errors.Is(err, repository.ErrNotFound) {
			return &ListFriendsOutput{
				Friends:    []FriendInfo{},
				TotalCount: 0,
			}, nil
		}
		return nil, fmt.Errorf("関係の取得中にエラーが発生しました: %w", err)
	}

	// 友達（Acceptedステータス）のみをフィルタリング
	var friends []FriendInfo
	for _, rel := range relationships {
		// Acceptedステータスのみを対象とする
		if rel.Status != valueobject.RelationshipStatusAccepted {
			continue
		}

		// 友達のユーザーIDを特定
		var friendID string
		var isRequester bool
		if rel.RequesterID == user.ID {
			friendID = rel.ReceiverID
			isRequester = true
		} else {
			friendID = rel.RequesterID
			isRequester = false
		}

		// 友達のユーザー情報を取得
		friendUser, err := uc.userRepo.FindByID(ctx, friendID)
		if err != nil {
			// ユーザーが削除されている場合はスキップ
			if errors.Is(err, repository.ErrNotFound) {
				// 削除されたユーザーとの友達関係は表示しない
				// ただし、データクリーンアップの観点から、
				// 将来的には削除されたユーザーとの関係も削除する処理を検討
				continue
			}
			return nil, fmt.Errorf("友達情報の取得中にエラーが発生しました: %w", err)
		}

		// 友達情報を構築
		friendInfo := FriendInfo{
			User:         friendUser,
			Relationship: rel,
			IsRequester:  isRequester,
			FriendSince:  rel.UpdatedAt.Format("2006-01-02 15:04:05"), // 承認日時（UpdatedAt）を友達になった日時とする
		}

		friends = append(friends, friendInfo)
	}

	// 友達リストを友達になった日時の新しい順にソート
	// 実装の簡略化のため、現在はリポジトリから返される順序のまま
	// 将来的には、UpdatedAtでソートする処理を追加することも検討

	return &ListFriendsOutput{
		Friends:    friends,
		TotalCount: len(friends),
	}, nil
}
