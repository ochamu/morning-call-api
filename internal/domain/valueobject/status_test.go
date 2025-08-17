package valueobject

import "testing"

func TestMorningCallStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   MorningCallStatus
		expected bool
	}{
		{
			name:     "スケジュール済みは有効",
			status:   MorningCallStatusScheduled,
			expected: true,
		},
		{
			name:     "配信済みは有効",
			status:   MorningCallStatusDelivered,
			expected: true,
		},
		{
			name:     "確認済みは有効",
			status:   MorningCallStatusConfirmed,
			expected: true,
		},
		{
			name:     "キャンセル済みは有効",
			status:   MorningCallStatusCancelled,
			expected: true,
		},
		{
			name:     "期限切れは有効",
			status:   MorningCallStatusExpired,
			expected: true,
		},
		{
			name:     "不明なステータスは無効",
			status:   MorningCallStatus("unknown"),
			expected: false,
		},
		{
			name:     "空文字列は無効",
			status:   MorningCallStatus(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.expected {
				t.Errorf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMorningCallStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     MorningCallStatus
		to       MorningCallStatus
		expected bool
	}{
		// Scheduled からの遷移
		{
			name:     "スケジュール済み→配信済み",
			from:     MorningCallStatusScheduled,
			to:       MorningCallStatusDelivered,
			expected: true,
		},
		{
			name:     "スケジュール済み→キャンセル",
			from:     MorningCallStatusScheduled,
			to:       MorningCallStatusCancelled,
			expected: true,
		},
		{
			name:     "スケジュール済み→期限切れ",
			from:     MorningCallStatusScheduled,
			to:       MorningCallStatusExpired,
			expected: true,
		},
		{
			name:     "スケジュール済み→確認済み（直接遷移不可）",
			from:     MorningCallStatusScheduled,
			to:       MorningCallStatusConfirmed,
			expected: false,
		},
		// Delivered からの遷移
		{
			name:     "配信済み→確認済み",
			from:     MorningCallStatusDelivered,
			to:       MorningCallStatusConfirmed,
			expected: true,
		},
		{
			name:     "配信済み→期限切れ",
			from:     MorningCallStatusDelivered,
			to:       MorningCallStatusExpired,
			expected: true,
		},
		{
			name:     "配信済み→キャンセル（不可）",
			from:     MorningCallStatusDelivered,
			to:       MorningCallStatusCancelled,
			expected: false,
		},
		// 終了状態からの遷移
		{
			name:     "確認済み→他の状態（不可）",
			from:     MorningCallStatusConfirmed,
			to:       MorningCallStatusScheduled,
			expected: false,
		},
		{
			name:     "キャンセル済み→他の状態（不可）",
			from:     MorningCallStatusCancelled,
			to:       MorningCallStatusScheduled,
			expected: false,
		},
		{
			name:     "期限切れ→他の状態（不可）",
			from:     MorningCallStatusExpired,
			to:       MorningCallStatusScheduled,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.expected {
				t.Errorf("CanTransitionTo() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRelationshipStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   RelationshipStatus
		expected bool
	}{
		{
			name:     "承認待ちは有効",
			status:   RelationshipStatusPending,
			expected: true,
		},
		{
			name:     "承認済みは有効",
			status:   RelationshipStatusAccepted,
			expected: true,
		},
		{
			name:     "ブロック済みは有効",
			status:   RelationshipStatusBlocked,
			expected: true,
		},
		{
			name:     "拒否済みは有効",
			status:   RelationshipStatusRejected,
			expected: true,
		},
		{
			name:     "不明なステータスは無効",
			status:   RelationshipStatus("unknown"),
			expected: false,
		},
		{
			name:     "空文字列は無効",
			status:   RelationshipStatus(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.expected {
				t.Errorf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRelationshipStatus_IsFriend(t *testing.T) {
	tests := []struct {
		name     string
		status   RelationshipStatus
		expected bool
	}{
		{
			name:     "承認済みは友達",
			status:   RelationshipStatusAccepted,
			expected: true,
		},
		{
			name:     "承認待ちは友達ではない",
			status:   RelationshipStatusPending,
			expected: false,
		},
		{
			name:     "ブロック済みは友達ではない",
			status:   RelationshipStatusBlocked,
			expected: false,
		},
		{
			name:     "拒否済みは友達ではない",
			status:   RelationshipStatusRejected,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsFriend(); got != tt.expected {
				t.Errorf("IsFriend() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRelationshipStatus_IsBlocked(t *testing.T) {
	tests := []struct {
		name     string
		status   RelationshipStatus
		expected bool
	}{
		{
			name:     "ブロック済みはブロック状態",
			status:   RelationshipStatusBlocked,
			expected: true,
		},
		{
			name:     "承認済みはブロック状態ではない",
			status:   RelationshipStatusAccepted,
			expected: false,
		},
		{
			name:     "承認待ちはブロック状態ではない",
			status:   RelationshipStatusPending,
			expected: false,
		},
		{
			name:     "拒否済みはブロック状態ではない",
			status:   RelationshipStatusRejected,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsBlocked(); got != tt.expected {
				t.Errorf("IsBlocked() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRelationshipStatus_IsPending(t *testing.T) {
	tests := []struct {
		name     string
		status   RelationshipStatus
		expected bool
	}{
		{
			name:     "承認待ちは承認待ち状態",
			status:   RelationshipStatusPending,
			expected: true,
		},
		{
			name:     "承認済みは承認待ち状態ではない",
			status:   RelationshipStatusAccepted,
			expected: false,
		},
		{
			name:     "ブロック済みは承認待ち状態ではない",
			status:   RelationshipStatusBlocked,
			expected: false,
		},
		{
			name:     "拒否済みは承認待ち状態ではない",
			status:   RelationshipStatusRejected,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsPending(); got != tt.expected {
				t.Errorf("IsPending() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRelationshipStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     RelationshipStatus
		to       RelationshipStatus
		expected bool
	}{
		// Pending からの遷移
		{
			name:     "承認待ち→承認済み",
			from:     RelationshipStatusPending,
			to:       RelationshipStatusAccepted,
			expected: true,
		},
		{
			name:     "承認待ち→拒否済み",
			from:     RelationshipStatusPending,
			to:       RelationshipStatusRejected,
			expected: true,
		},
		{
			name:     "承認待ち→ブロック",
			from:     RelationshipStatusPending,
			to:       RelationshipStatusBlocked,
			expected: true,
		},
		// Accepted からの遷移
		{
			name:     "承認済み→ブロック",
			from:     RelationshipStatusAccepted,
			to:       RelationshipStatusBlocked,
			expected: true,
		},
		{
			name:     "承認済み→拒否済み（不可）",
			from:     RelationshipStatusAccepted,
			to:       RelationshipStatusRejected,
			expected: false,
		},
		{
			name:     "承認済み→承認待ち（不可）",
			from:     RelationshipStatusAccepted,
			to:       RelationshipStatusPending,
			expected: false,
		},
		// Rejected からの遷移
		{
			name:     "拒否済み→承認待ち（再リクエスト）",
			from:     RelationshipStatusRejected,
			to:       RelationshipStatusPending,
			expected: true,
		},
		{
			name:     "拒否済み→ブロック",
			from:     RelationshipStatusRejected,
			to:       RelationshipStatusBlocked,
			expected: true,
		},
		{
			name:     "拒否済み→承認済み（直接遷移不可）",
			from:     RelationshipStatusRejected,
			to:       RelationshipStatusAccepted,
			expected: false,
		},
		// Blocked からの遷移
		{
			name:     "ブロック→承認待ち（不可）",
			from:     RelationshipStatusBlocked,
			to:       RelationshipStatusPending,
			expected: false,
		},
		{
			name:     "ブロック→承認済み（不可）",
			from:     RelationshipStatusBlocked,
			to:       RelationshipStatusAccepted,
			expected: false,
		},
		{
			name:     "ブロック→拒否済み（不可）",
			from:     RelationshipStatusBlocked,
			to:       RelationshipStatusRejected,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.expected {
				t.Errorf("CanTransitionTo() = %v, want %v", got, tt.expected)
			}
		})
	}
}
