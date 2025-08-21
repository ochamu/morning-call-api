package response

import "time"

// MorningCallResponse はモーニングコールのレスポンス
type MorningCallResponse struct {
	ID            string     `json:"id"`
	SenderID      string     `json:"sender_id"`
	ReceiverID    string     `json:"receiver_id"`
	ScheduledTime time.Time  `json:"scheduled_time"`
	Message       string     `json:"message"`
	Status        string     `json:"status"`
	ConfirmedAt   *time.Time `json:"confirmed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// MorningCallListResponse はモーニングコール一覧のレスポンス
type MorningCallListResponse struct {
	MorningCalls []MorningCallResponse `json:"morning_calls"`
	Total        int                   `json:"total"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
}
