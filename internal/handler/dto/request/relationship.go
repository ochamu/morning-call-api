package request

// SendFriendRequest は友達リクエスト送信のリクエスト
type SendFriendRequest struct {
	ReceiverID string `json:"receiver_id"`
}

// AcceptFriendRequest は友達リクエスト承認のリクエスト
type AcceptFriendRequest struct {
	RelationshipID string `json:"relationship_id"`
}

// RejectFriendRequest は友達リクエスト拒否のリクエスト
type RejectFriendRequest struct {
	RelationshipID string `json:"relationship_id"`
}

// BlockUserRequest はユーザーブロックのリクエスト
type BlockUserRequest struct {
	UserID string `json:"user_id"`
}

// RemoveRelationshipRequest は関係削除のリクエスト
type RemoveRelationshipRequest struct {
	RelationshipID string `json:"relationship_id"`
}
