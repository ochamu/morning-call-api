package entity

import (
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// Relationship はユーザー間の友達関係を表すエンティティ
type Relationship struct {
	ID          string
	RequesterID string // 友達リクエストを送信したユーザー
	ReceiverID  string // 友達リクエストを受信したユーザー
	Status      valueobject.RelationshipStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewRelationship は新しい友達関係エンティティを作成する
func NewRelationship(id, requesterID, receiverID string) (*Relationship, valueobject.NGReason) {
	r := &Relationship{
		ID:          id,
		RequesterID: requesterID,
		ReceiverID:  receiverID,
		Status:      valueobject.RelationshipStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 検証
	if reason := r.Validate(); reason.IsNG() {
		return nil, reason
	}

	return r, valueobject.OK()
}

// Validate は友達関係エンティティの妥当性を検証する
func (r *Relationship) Validate() valueobject.NGReason {
	// ID検証
	if r.ID == "" {
		return valueobject.NG("関係IDは必須です")
	}

	// ユーザーID検証
	if reason := r.ValidateUsers(); reason.IsNG() {
		return reason
	}

	// ステータス検証
	if !r.Status.IsValid() {
		return valueobject.NG("無効なステータスです")
	}

	return valueobject.OK()
}

// ValidateUsers はリクエスター・レシーバーの妥当性を検証する
func (r *Relationship) ValidateUsers() valueobject.NGReason {
	if r.RequesterID == "" {
		return valueobject.NG("リクエスト送信者IDは必須です")
	}

	if r.ReceiverID == "" {
		return valueobject.NG("リクエスト受信者IDは必須です")
	}

	if r.RequesterID == r.ReceiverID {
		return valueobject.NG("自分自身に友達リクエストを送ることはできません")
	}

	return valueobject.OK()
}

// CanTransitionTo は指定されたステータスへの遷移が可能かを検証する
func (r *Relationship) CanTransitionTo(newStatus valueobject.RelationshipStatus) bool {
	return r.Status.CanTransitionTo(newStatus)
}

// UpdateStatus はステータスを更新する
func (r *Relationship) UpdateStatus(newStatus valueobject.RelationshipStatus) valueobject.NGReason {
	if !r.CanTransitionTo(newStatus) {
		return valueobject.NG("このステータスへの遷移はできません")
	}

	r.Status = newStatus
	r.UpdatedAt = time.Now()
	return valueobject.OK()
}

// Accept は友達リクエストを承認する
func (r *Relationship) Accept() valueobject.NGReason {
	if r.Status != valueobject.RelationshipStatusPending {
		return valueobject.NG("承認待ち状態のリクエストのみ承認できます")
	}
	return r.UpdateStatus(valueobject.RelationshipStatusAccepted)
}

// Reject は友達リクエストを拒否する
func (r *Relationship) Reject() valueobject.NGReason {
	if r.Status != valueobject.RelationshipStatusPending {
		return valueobject.NG("承認待ち状態のリクエストのみ拒否できます")
	}
	return r.UpdateStatus(valueobject.RelationshipStatusRejected)
}

// Block はユーザーをブロックする
func (r *Relationship) Block() valueobject.NGReason {
	// ブロックは承認待ち、承認済み、拒否済みから可能
	if r.Status == valueobject.RelationshipStatusBlocked {
		return valueobject.NG("既にブロック済みです")
	}
	return r.UpdateStatus(valueobject.RelationshipStatusBlocked)
}

// Resend は拒否済みの友達リクエストを再送信する
func (r *Relationship) Resend() valueobject.NGReason {
	if r.Status != valueobject.RelationshipStatusRejected {
		return valueobject.NG("拒否済みのリクエストのみ再送信できます")
	}
	return r.UpdateStatus(valueobject.RelationshipStatusPending)
}

// IsFriend は友達関係かを判定する
func (r *Relationship) IsFriend() bool {
	return r.Status.IsFriend()
}

// IsBlocked はブロック済みかを判定する
func (r *Relationship) IsBlocked() bool {
	return r.Status.IsBlocked()
}

// IsPending は承認待ちかを判定する
func (r *Relationship) IsPending() bool {
	return r.Status.IsPending()
}

// IsRejected は拒否済みかを判定する
func (r *Relationship) IsRejected() bool {
	return r.Status == valueobject.RelationshipStatusRejected
}

// InvolvesUser は指定されたユーザーが関係に含まれているかを判定する
func (r *Relationship) InvolvesUser(userID string) bool {
	return r.RequesterID == userID || r.ReceiverID == userID
}

// IsRequester は指定されたユーザーがリクエスト送信者かを判定する
func (r *Relationship) IsRequester(userID string) bool {
	return r.RequesterID == userID
}

// IsReceiver は指定されたユーザーがリクエスト受信者かを判定する
func (r *Relationship) IsReceiver(userID string) bool {
	return r.ReceiverID == userID
}

// GetOtherUserID は指定されたユーザーの相手のIDを返す
func (r *Relationship) GetOtherUserID(userID string) string {
	if r.RequesterID == userID {
		return r.ReceiverID
	}
	if r.ReceiverID == userID {
		return r.RequesterID
	}
	return "" // ユーザーが関係に含まれていない場合
}

// CanBeAcceptedBy は指定されたユーザーが承認可能かを判定する
func (r *Relationship) CanBeAcceptedBy(userID string) bool {
	// 受信者のみが承認可能
	return r.IsReceiver(userID) && r.IsPending()
}

// CanBeRejectedBy は指定されたユーザーが拒否可能かを判定する
func (r *Relationship) CanBeRejectedBy(userID string) bool {
	// 受信者のみが拒否可能
	return r.IsReceiver(userID) && r.IsPending()
}

// CanBeBlockedBy は指定されたユーザーがブロック可能かを判定する
func (r *Relationship) CanBeBlockedBy(userID string) bool {
	// 両者がブロック可能（既にブロック済みでない限り）
	return r.InvolvesUser(userID) && !r.IsBlocked()
}

// CanBeResendBy は指定されたユーザーが再送信可能かを判定する
func (r *Relationship) CanBeResendBy(userID string) bool {
	// リクエスト送信者のみが再送信可能
	return r.IsRequester(userID) && r.IsRejected()
}

// Equals は他の友達関係と同一かを判定する
func (r *Relationship) Equals(other *Relationship) bool {
	if other == nil {
		return false
	}
	return r.ID == other.ID
}
