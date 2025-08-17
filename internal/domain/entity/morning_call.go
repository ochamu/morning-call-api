package entity

import (
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

// MorningCall は一人のユーザーが別のユーザーに設定するアラームを表すエンティティ
type MorningCall struct {
	ID            string
	SenderID      string
	ReceiverID    string
	ScheduledTime time.Time
	Message       string
	Status        valueobject.MorningCallStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewMorningCall は新しいモーニングコールエンティティを作成する
func NewMorningCall(id, senderID, receiverID string, scheduledTime time.Time, message string) (*MorningCall, valueobject.NGReason) {
	mc := &MorningCall{
		ID:            id,
		SenderID:      senderID,
		ReceiverID:    receiverID,
		ScheduledTime: scheduledTime,
		Message:       message,
		Status:        valueobject.MorningCallStatusScheduled,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 検証
	if reason := mc.Validate(); reason.IsNG() {
		return nil, reason
	}

	return mc, valueobject.OK()
}

// Validate はモーニングコールエンティティの妥当性を検証する
func (mc *MorningCall) Validate() valueobject.NGReason {
	// ID検証
	if mc.ID == "" {
		return valueobject.NG("モーニングコールIDは必須です")
	}

	// 送信者・受信者検証
	if reason := mc.ValidateSenderReceiver(); reason.IsNG() {
		return reason
	}

	// 時刻検証
	if reason := mc.ValidateScheduledTime(); reason.IsNG() {
		return reason
	}

	// メッセージ検証
	if reason := mc.ValidateMessage(); reason.IsNG() {
		return reason
	}

	// ステータス検証
	if !mc.Status.IsValid() {
		return valueobject.NG("無効なステータスです")
	}

	return valueobject.OK()
}

// ValidateSenderReceiver は送信者と受信者の妥当性を検証する
func (mc *MorningCall) ValidateSenderReceiver() valueobject.NGReason {
	if mc.SenderID == "" {
		return valueobject.NG("送信者IDは必須です")
	}

	if mc.ReceiverID == "" {
		return valueobject.NG("受信者IDは必須です")
	}

	if mc.SenderID == mc.ReceiverID {
		return valueobject.NG("自分自身にモーニングコールを設定することはできません")
	}

	return valueobject.OK()
}

// ValidateScheduledTime はアラーム時刻の妥当性を検証する
func (mc *MorningCall) ValidateScheduledTime() valueobject.NGReason {
	now := time.Now()

	// 過去の時刻は許可しない（作成時のみ。既存のものは過去になる可能性がある）
	if mc.Status == valueobject.MorningCallStatusScheduled && mc.ScheduledTime.Before(now) {
		return valueobject.NG("アラーム時刻は現在時刻より後である必要があります")
	}

	// 30日以内の制限
	maxTime := now.Add(30 * 24 * time.Hour)
	if mc.ScheduledTime.After(maxTime) {
		return valueobject.NG("アラーム時刻は30日以内で設定してください")
	}

	return valueobject.OK()
}

// ValidateMessage はメッセージの妥当性を検証する
func (mc *MorningCall) ValidateMessage() valueobject.NGReason {
	// メッセージは任意（空でもOK）
	// rune（文字）単位でカウント
	messageLength := len([]rune(mc.Message))
	if messageLength > 500 {
		return valueobject.NG("メッセージは500文字以内で入力してください")
	}

	return valueobject.OK()
}

// CanTransitionTo は指定されたステータスへの遷移が可能かを検証する
func (mc *MorningCall) CanTransitionTo(newStatus valueobject.MorningCallStatus) bool {
	return mc.Status.CanTransitionTo(newStatus)
}

// UpdateStatus はステータスを更新する
func (mc *MorningCall) UpdateStatus(newStatus valueobject.MorningCallStatus) valueobject.NGReason {
	if !mc.CanTransitionTo(newStatus) {
		return valueobject.NG("このステータスへの遷移はできません")
	}

	mc.Status = newStatus
	mc.UpdatedAt = time.Now()
	return valueobject.OK()
}

// Cancel はモーニングコールをキャンセルする
func (mc *MorningCall) Cancel() valueobject.NGReason {
	return mc.UpdateStatus(valueobject.MorningCallStatusCancelled)
}

// MarkAsDelivered はモーニングコールを配信済みにする
func (mc *MorningCall) MarkAsDelivered() valueobject.NGReason {
	return mc.UpdateStatus(valueobject.MorningCallStatusDelivered)
}

// ConfirmWakeUp は起床確認を記録する
func (mc *MorningCall) ConfirmWakeUp() valueobject.NGReason {
	return mc.UpdateStatus(valueobject.MorningCallStatusConfirmed)
}

// MarkAsExpired はモーニングコールを期限切れにする
func (mc *MorningCall) MarkAsExpired() valueobject.NGReason {
	return mc.UpdateStatus(valueobject.MorningCallStatusExpired)
}

// UpdateMessage はメッセージを更新する（スケジュール済みの場合のみ）
func (mc *MorningCall) UpdateMessage(newMessage string) valueobject.NGReason {
	if mc.Status != valueobject.MorningCallStatusScheduled {
		return valueobject.NG("スケジュール済みのモーニングコールのみ更新できます")
	}

	oldMessage := mc.Message
	mc.Message = newMessage

	if reason := mc.ValidateMessage(); reason.IsNG() {
		mc.Message = oldMessage // ロールバック
		return reason
	}

	mc.UpdatedAt = time.Now()
	return valueobject.OK()
}

// UpdateScheduledTime はアラーム時刻を更新する（スケジュール済みの場合のみ）
func (mc *MorningCall) UpdateScheduledTime(newTime time.Time) valueobject.NGReason {
	if mc.Status != valueobject.MorningCallStatusScheduled {
		return valueobject.NG("スケジュール済みのモーニングコールのみ更新できます")
	}

	oldTime := mc.ScheduledTime
	mc.ScheduledTime = newTime

	if reason := mc.ValidateScheduledTime(); reason.IsNG() {
		mc.ScheduledTime = oldTime // ロールバック
		return reason
	}

	mc.UpdatedAt = time.Now()
	return valueobject.OK()
}

// IsActive はモーニングコールが有効（配信待ちまたは配信済み）かを判定する
func (mc *MorningCall) IsActive() bool {
	return mc.Status == valueobject.MorningCallStatusScheduled ||
		mc.Status == valueobject.MorningCallStatusDelivered
}

// IsPast はアラーム時刻が過去かを判定する
func (mc *MorningCall) IsPast() bool {
	return mc.ScheduledTime.Before(time.Now())
}

// ShouldDeliver は配信すべきかを判定する
func (mc *MorningCall) ShouldDeliver() bool {
	return mc.Status == valueobject.MorningCallStatusScheduled && mc.IsPast()
}

// Equals は他のモーニングコールと同一かを判定する
func (mc *MorningCall) Equals(other *MorningCall) bool {
	if other == nil {
		return false
	}
	return mc.ID == other.ID
}
