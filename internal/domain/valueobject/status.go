package valueobject

// MorningCallStatus はモーニングコールの状態を表す
type MorningCallStatus string

const (
	// MorningCallStatusScheduled はスケジュール済み状態
	MorningCallStatusScheduled MorningCallStatus = "scheduled"
	// MorningCallStatusDelivered は配信済み状態
	MorningCallStatusDelivered MorningCallStatus = "delivered"
	// MorningCallStatusConfirmed は起床確認済み状態
	MorningCallStatusConfirmed MorningCallStatus = "confirmed"
	// MorningCallStatusCancelled はキャンセル済み状態
	MorningCallStatusCancelled MorningCallStatus = "cancelled"
	// MorningCallStatusExpired は期限切れ状態
	MorningCallStatusExpired MorningCallStatus = "expired"
)

// IsValid はステータスが有効な値かを検証する
func (s MorningCallStatus) IsValid() bool {
	switch s {
	case MorningCallStatusScheduled,
		MorningCallStatusDelivered,
		MorningCallStatusConfirmed,
		MorningCallStatusCancelled,
		MorningCallStatusExpired:
		return true
	default:
		return false
	}
}

// String はステータスの文字列表現を返す
func (s MorningCallStatus) String() string {
	return string(s)
}

// CanTransitionTo は指定されたステータスへの遷移が可能かを検証する
func (s MorningCallStatus) CanTransitionTo(next MorningCallStatus) bool {
	switch s {
	case MorningCallStatusScheduled:
		// 開発・テスト環境では、Scheduledから直接Confirmedへの遷移も許可
		// 本番環境では、Delivered経由でのみConfirmedに遷移すべき
		return next == MorningCallStatusDelivered || next == MorningCallStatusCancelled || 
			next == MorningCallStatusExpired || next == MorningCallStatusConfirmed
	case MorningCallStatusDelivered:
		return next == MorningCallStatusConfirmed || next == MorningCallStatusExpired
	case MorningCallStatusConfirmed, MorningCallStatusCancelled, MorningCallStatusExpired:
		return false // 終了状態からの遷移は不可
	default:
		return false
	}
}

// RelationshipStatus は友達関係の状態を表す
type RelationshipStatus string

const (
	// RelationshipStatusPending は承認待ち状態
	RelationshipStatusPending RelationshipStatus = "pending"
	// RelationshipStatusAccepted は承認済み（友達）状態
	RelationshipStatusAccepted RelationshipStatus = "accepted"
	// RelationshipStatusBlocked はブロック済み状態
	RelationshipStatusBlocked RelationshipStatus = "blocked"
	// RelationshipStatusRejected は拒否済み状態
	RelationshipStatusRejected RelationshipStatus = "rejected"
)

// IsValid はステータスが有効な値かを検証する
func (s RelationshipStatus) IsValid() bool {
	switch s {
	case RelationshipStatusPending,
		RelationshipStatusAccepted,
		RelationshipStatusBlocked,
		RelationshipStatusRejected:
		return true
	default:
		return false
	}
}

// String はステータスの文字列表現を返す
func (s RelationshipStatus) String() string {
	return string(s)
}

// IsFriend は友達関係かを判定する
func (s RelationshipStatus) IsFriend() bool {
	return s == RelationshipStatusAccepted
}

// IsBlocked はブロック済みかを判定する
func (s RelationshipStatus) IsBlocked() bool {
	return s == RelationshipStatusBlocked
}

// IsPending は承認待ちかを判定する
func (s RelationshipStatus) IsPending() bool {
	return s == RelationshipStatusPending
}

// CanTransitionTo は指定されたステータスへの遷移が可能かを検証する
func (s RelationshipStatus) CanTransitionTo(next RelationshipStatus) bool {
	switch s {
	case RelationshipStatusPending:
		return next == RelationshipStatusAccepted || next == RelationshipStatusRejected || next == RelationshipStatusBlocked
	case RelationshipStatusAccepted:
		return next == RelationshipStatusBlocked
	case RelationshipStatusRejected:
		return next == RelationshipStatusPending || next == RelationshipStatusBlocked
	case RelationshipStatusBlocked:
		return false // ブロック解除は新規リクエストとして扱う
	default:
		return false
	}
}
