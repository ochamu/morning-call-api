package response

import (
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// RelationshipResponse はRelationshipのレスポンス
type RelationshipResponse struct {
	ID          string                         `json:"id"`
	RequesterID string                         `json:"requester_id"`
	ReceiverID  string                         `json:"receiver_id"`
	Status      valueobject.RelationshipStatus `json:"status"`
	CreatedAt   time.Time                      `json:"created_at"`
	UpdatedAt   time.Time                      `json:"updated_at"`
}

// NewRelationshipResponse はentityからレスポンスを作成
func NewRelationshipResponse(r *entity.Relationship) *RelationshipResponse {
	return &RelationshipResponse{
		ID:          r.ID,
		RequesterID: r.RequesterID,
		ReceiverID:  r.ReceiverID,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// RelationshipListResponse は関係一覧のレスポンス
type RelationshipListResponse struct {
	Relationships []*RelationshipResponse `json:"relationships"`
	Total         int                     `json:"total"`
}

// NewRelationshipListResponse はentityのスライスからレスポンスを作成
func NewRelationshipListResponse(relationships []*entity.Relationship) *RelationshipListResponse {
	res := &RelationshipListResponse{
		Relationships: make([]*RelationshipResponse, 0, len(relationships)),
		Total:         len(relationships),
	}
	for _, r := range relationships {
		res.Relationships = append(res.Relationships, NewRelationshipResponse(r))
	}
	return res
}

// FriendResponse は友達情報のレスポンス
type FriendResponse struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	FriendSince time.Time `json:"friend_since"`
}

// FriendListResponse は友達一覧のレスポンス
type FriendListResponse struct {
	Friends []*FriendResponse `json:"friends"`
	Total   int               `json:"total"`
}
