package request

import "time"

// CreateMorningCallRequest はモーニングコール作成リクエスト
type CreateMorningCallRequest struct {
	ReceiverID    string    `json:"receiver_id"`
	ScheduledTime time.Time `json:"scheduled_time"`
	Message       string    `json:"message"`
}

// UpdateMorningCallRequest はモーニングコール更新リクエスト
type UpdateMorningCallRequest struct {
	ScheduledTime time.Time `json:"scheduled_time"`
	Message       string    `json:"message"`
}

// ListMorningCallsRequest はモーニングコール一覧取得リクエスト
type ListMorningCallsRequest struct {
	Status string `json:"status,omitempty"` // pending, sent, confirmed
	Type   string `json:"type,omitempty"`   // sent, received
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}
