package relationship

import (
	"context"
	"errors"
	"fmt"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// ListFriendRequestsUseCase は友達リクエスト一覧取得のユースケース
type ListFriendRequestsUseCase struct {
	relationshipRepo repository.RelationshipRepository
	userRepo         repository.UserRepository
}

// NewListFriendRequestsUseCase は新しい友達リクエスト一覧取得ユースケースを作成する
func NewListFriendRequestsUseCase(
	relationshipRepo repository.RelationshipRepository,
	userRepo repository.UserRepository,
) *ListFriendRequestsUseCase {
	return &ListFriendRequestsUseCase{
		relationshipRepo: relationshipRepo,
		userRepo:         userRepo,
	}
}

// ListFriendRequestsInput は友達リクエスト一覧取得の入力データ
type ListFriendRequestsInput struct {
	UserID string // 友達リクエストを取得するユーザーID
	Type   string // "received" (受信したリクエスト) または "sent" (送信したリクエスト)
}

// FriendRequestInfo は友達リクエスト情報
type FriendRequestInfo struct {
	Relationship *entity.Relationship // 関係情報
	Requester    *entity.User         // リクエスト送信者の情報（受信リクエストの場合）
	Receiver     *entity.User         // リクエスト受信者の情報（送信リクエストの場合）
	RequestedAt  string               // リクエスト日時（文字列表現）
}

// ListFriendRequestsOutput は友達リクエスト一覧取得の出力データ
type ListFriendRequestsOutput struct {
	Requests   []FriendRequestInfo // 友達リクエスト一覧
	TotalCount int                 // 総リクエスト数
}

// Execute は友達リクエスト一覧を取得する
func (uc *ListFriendRequestsUseCase) Execute(ctx context.Context, input ListFriendRequestsInput) (*ListFriendRequestsOutput, error) {
	// 入力値の基本検証
	if input.UserID == "" {
		return nil, fmt.Errorf("ユーザーIDは必須です")
	}
	if input.Type != "received" && input.Type != "sent" {
		return nil, fmt.Errorf("タイプは 'received' または 'sent' である必要があります")
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
			return &ListFriendRequestsOutput{
				Requests:   []FriendRequestInfo{},
				TotalCount: 0,
			}, nil
		}
		return nil, fmt.Errorf("関係の取得中にエラーが発生しました: %w", err)
	}

	// ペンディング中のリクエストのみをフィルタリング
	var requests []FriendRequestInfo
	for _, rel := range relationships {
		// Pendingステータスのみを対象とする
		if rel.Status != valueobject.RelationshipStatusPending {
			continue
		}

		// タイプに応じてフィルタリング
		var shouldInclude bool
		var otherUserID string
		var otherUser *entity.User

		if input.Type == "received" {
			// 受信したリクエスト：自分がReceiverの場合
			if rel.ReceiverID == user.ID {
				shouldInclude = true
				otherUserID = rel.RequesterID
			}
		} else { // input.Type == "sent"
			// 送信したリクエスト：自分がRequesterの場合
			if rel.RequesterID == user.ID {
				shouldInclude = true
				otherUserID = rel.ReceiverID
			}
		}

		if !shouldInclude {
			continue
		}

		// 相手ユーザーの情報を取得
		otherUser, err = uc.userRepo.FindByID(ctx, otherUserID)
		if err != nil {
			// ユーザーが削除されている場合はスキップ
			if errors.Is(err, repository.ErrNotFound) {
				// 削除されたユーザーとのリクエストは表示しない
				// ただし、データクリーンアップの観点から、
				// 将来的には削除されたユーザーとの関係も削除する処理を検討
				continue
			}
			return nil, fmt.Errorf("ユーザー情報の取得中にエラーが発生しました: %w", err)
		}

		// リクエスト情報を構築
		requestInfo := FriendRequestInfo{
			Relationship: rel,
			RequestedAt:  rel.CreatedAt.Format("2006-01-02 15:04:05"), // リクエスト日時
		}

		// タイプに応じて適切な情報を設定
		if input.Type == "received" {
			requestInfo.Requester = otherUser // リクエスト送信者
		} else {
			requestInfo.Receiver = otherUser // リクエスト受信者
		}

		requests = append(requests, requestInfo)
	}

	// リクエストを日時の新しい順にソート
	// 実装の簡略化のため、現在はリポジトリから返される順序のまま
	// 将来的には、CreatedAtでソートする処理を追加することも検討

	return &ListFriendRequestsOutput{
		Requests:   requests,
		TotalCount: len(requests),
	}, nil
}
