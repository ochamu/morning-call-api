package entity

import (
	"strings"
	"testing"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

func TestNewMorningCall(t *testing.T) {
	futureTime := time.Now().Add(1 * time.Hour)
	pastTime := time.Now().Add(-1 * time.Hour)
	farFutureTime := time.Now().Add(31 * 24 * time.Hour)

	tests := []struct {
		name          string
		id            string
		senderID      string
		receiverID    string
		scheduledTime time.Time
		message       string
		expectError   bool
		errorMsg      string
	}{
		{
			name:          "正常なモーニングコール作成",
			id:            "mc-001",
			senderID:      "user-001",
			receiverID:    "user-002",
			scheduledTime: futureTime,
			message:       "おはよう！今日も頑張ろう！",
			expectError:   false,
		},
		{
			name:          "メッセージなしでも作成可能",
			id:            "mc-001",
			senderID:      "user-001",
			receiverID:    "user-002",
			scheduledTime: futureTime,
			message:       "",
			expectError:   false,
		},
		{
			name:          "IDが空",
			id:            "",
			senderID:      "user-001",
			receiverID:    "user-002",
			scheduledTime: futureTime,
			message:       "test",
			expectError:   true,
			errorMsg:      "モーニングコールIDは必須です",
		},
		{
			name:          "送信者IDが空",
			id:            "mc-001",
			senderID:      "",
			receiverID:    "user-002",
			scheduledTime: futureTime,
			message:       "test",
			expectError:   true,
			errorMsg:      "送信者IDは必須です",
		},
		{
			name:          "受信者IDが空",
			id:            "mc-001",
			senderID:      "user-001",
			receiverID:    "",
			scheduledTime: futureTime,
			message:       "test",
			expectError:   true,
			errorMsg:      "受信者IDは必須です",
		},
		{
			name:          "自分自身への送信",
			id:            "mc-001",
			senderID:      "user-001",
			receiverID:    "user-001",
			scheduledTime: futureTime,
			message:       "test",
			expectError:   true,
			errorMsg:      "自分自身にモーニングコールを設定することはできません",
		},
		{
			name:          "過去の時刻",
			id:            "mc-001",
			senderID:      "user-001",
			receiverID:    "user-002",
			scheduledTime: pastTime,
			message:       "test",
			expectError:   true,
			errorMsg:      "アラーム時刻は現在時刻より後である必要があります",
		},
		{
			name:          "30日以上先の時刻",
			id:            "mc-001",
			senderID:      "user-001",
			receiverID:    "user-002",
			scheduledTime: farFutureTime,
			message:       "test",
			expectError:   true,
			errorMsg:      "アラーム時刻は30日以内で設定してください",
		},
		{
			name:          "メッセージが長すぎる",
			id:            "mc-001",
			senderID:      "user-001",
			receiverID:    "user-002",
			scheduledTime: futureTime,
			message:       strings.Repeat("あ", 501),
			expectError:   true,
			errorMsg:      "メッセージは500文字以内で入力してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc, reason := NewMorningCall(tt.id, tt.senderID, tt.receiverID, tt.scheduledTime, tt.message)

			if tt.expectError {
				if reason.IsOK() {
					t.Errorf("エラーが期待されたが、成功した")
				}
				if reason.Error() != tt.errorMsg {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMsg, reason.Error())
				}
				if mc != nil {
					t.Errorf("エラー時にはnilが期待されたが、モーニングコールが返された")
				}
			} else {
				if reason.IsNG() {
					t.Errorf("成功が期待されたが、エラーが発生: %s", reason.Error())
				}
				if mc == nil {
					t.Errorf("成功時にはモーニングコールが期待されたが、nilが返された")
				} else {
					if mc.ID != tt.id {
						t.Errorf("ID: expected %s, got %s", tt.id, mc.ID)
					}
					if mc.SenderID != tt.senderID {
						t.Errorf("SenderID: expected %s, got %s", tt.senderID, mc.SenderID)
					}
					if mc.ReceiverID != tt.receiverID {
						t.Errorf("ReceiverID: expected %s, got %s", tt.receiverID, mc.ReceiverID)
					}
					if mc.Message != tt.message {
						t.Errorf("Message: expected %s, got %s", tt.message, mc.Message)
					}
					if mc.Status != valueobject.MorningCallStatusScheduled {
						t.Errorf("初期ステータスはScheduledであるべき")
					}
				}
			}
		})
	}
}

func TestMorningCall_ValidateScheduledTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		scheduledTime time.Time
		status        valueobject.MorningCallStatus
		expectError   bool
		errorMsg      string
	}{
		{
			name:          "未来の時刻（1時間後）",
			scheduledTime: now.Add(1 * time.Hour),
			status:        valueobject.MorningCallStatusScheduled,
			expectError:   false,
		},
		{
			name:          "未来の時刻（29日後）",
			scheduledTime: now.Add(29 * 24 * time.Hour),
			status:        valueobject.MorningCallStatusScheduled,
			expectError:   false,
		},
		{
			name:          "過去の時刻（スケジュール済み）",
			scheduledTime: now.Add(-1 * time.Hour),
			status:        valueobject.MorningCallStatusScheduled,
			expectError:   true,
			errorMsg:      "アラーム時刻は現在時刻より後である必要があります",
		},
		{
			name:          "過去の時刻（配信済み）",
			scheduledTime: now.Add(-1 * time.Hour),
			status:        valueobject.MorningCallStatusDelivered,
			expectError:   false, // 配信済みなら過去でもOK
		},
		{
			name:          "31日後の時刻",
			scheduledTime: now.Add(31 * 24 * time.Hour),
			status:        valueobject.MorningCallStatusScheduled,
			expectError:   true,
			errorMsg:      "アラーム時刻は30日以内で設定してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MorningCall{
				ScheduledTime: tt.scheduledTime,
				Status:        tt.status,
			}
			reason := mc.ValidateScheduledTime()

			if tt.expectError {
				if reason.IsOK() {
					t.Errorf("エラーが期待されたが、成功した")
				}
				if reason.Error() != tt.errorMsg {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMsg, reason.Error())
				}
			} else {
				if reason.IsNG() {
					t.Errorf("成功が期待されたが、エラーが発生: %s", reason.Error())
				}
			}
		})
	}
}

func TestMorningCall_ValidateMessage(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "通常のメッセージ",
			message:     "おはよう！今日も頑張ろう！",
			expectError: false,
		},
		{
			name:        "空のメッセージ",
			message:     "",
			expectError: false,
		},
		{
			name:        "500文字ちょうど",
			message:     strings.Repeat("あ", 500),
			expectError: false,
		},
		{
			name:        "501文字",
			message:     strings.Repeat("あ", 501),
			expectError: true,
			errorMsg:    "メッセージは500文字以内で入力してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MorningCall{
				Message: tt.message,
			}
			reason := mc.ValidateMessage()

			if tt.expectError {
				if reason.IsOK() {
					t.Errorf("エラーが期待されたが、成功した")
				}
				if reason.Error() != tt.errorMsg {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMsg, reason.Error())
				}
			} else {
				if reason.IsNG() {
					t.Errorf("成功が期待されたが、エラーが発生: %s", reason.Error())
				}
			}
		})
	}
}

func TestMorningCall_UpdateStatus(t *testing.T) {
	tests := []struct {
		name        string
		fromStatus  valueobject.MorningCallStatus
		toStatus    valueobject.MorningCallStatus
		expectError bool
		errorMsg    string
	}{
		{
			name:        "スケジュール済み→配信済み",
			fromStatus:  valueobject.MorningCallStatusScheduled,
			toStatus:    valueobject.MorningCallStatusDelivered,
			expectError: false,
		},
		{
			name:        "スケジュール済み→キャンセル",
			fromStatus:  valueobject.MorningCallStatusScheduled,
			toStatus:    valueobject.MorningCallStatusCancelled,
			expectError: false,
		},
		{
			name:        "配信済み→確認済み",
			fromStatus:  valueobject.MorningCallStatusDelivered,
			toStatus:    valueobject.MorningCallStatusConfirmed,
			expectError: false,
		},
		{
			name:        "確認済み→スケジュール済み（不可）",
			fromStatus:  valueobject.MorningCallStatusConfirmed,
			toStatus:    valueobject.MorningCallStatusScheduled,
			expectError: true,
			errorMsg:    "このステータスへの遷移はできません",
		},
		{
			name:        "キャンセル済み→配信済み（不可）",
			fromStatus:  valueobject.MorningCallStatusCancelled,
			toStatus:    valueobject.MorningCallStatusDelivered,
			expectError: true,
			errorMsg:    "このステータスへの遷移はできません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MorningCall{
				ID:        "mc-001",
				Status:    tt.fromStatus,
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			}

			oldUpdatedAt := mc.UpdatedAt
			reason := mc.UpdateStatus(tt.toStatus)

			if tt.expectError {
				if reason.IsOK() {
					t.Errorf("エラーが期待されたが、成功した")
				}
				if reason.Error() != tt.errorMsg {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMsg, reason.Error())
				}
				if mc.Status != tt.fromStatus {
					t.Errorf("エラー時はステータスが変更されないべき")
				}
				if !mc.UpdatedAt.Equal(oldUpdatedAt) {
					t.Errorf("エラー時はUpdatedAtが更新されないべき")
				}
			} else {
				if reason.IsNG() {
					t.Errorf("成功が期待されたが、エラーが発生: %s", reason.Error())
				}
				if mc.Status != tt.toStatus {
					t.Errorf("ステータス: expected %s, got %s", tt.toStatus, mc.Status)
				}
				if mc.UpdatedAt.Equal(oldUpdatedAt) || mc.UpdatedAt.Before(oldUpdatedAt) {
					t.Errorf("成功時はUpdatedAtが更新されるべき")
				}
			}
		})
	}
}

func TestMorningCall_StatusTransitionMethods(t *testing.T) {
	t.Run("Cancel", func(t *testing.T) {
		mc := &MorningCall{
			Status: valueobject.MorningCallStatusScheduled,
		}
		reason := mc.Cancel()
		if reason.IsNG() {
			t.Errorf("キャンセルに失敗: %s", reason.Error())
		}
		if mc.Status != valueobject.MorningCallStatusCancelled {
			t.Errorf("ステータスがCancelledになるべき")
		}
	})

	t.Run("MarkAsDelivered", func(t *testing.T) {
		mc := &MorningCall{
			Status: valueobject.MorningCallStatusScheduled,
		}
		reason := mc.MarkAsDelivered()
		if reason.IsNG() {
			t.Errorf("配信済みへの変更に失敗: %s", reason.Error())
		}
		if mc.Status != valueobject.MorningCallStatusDelivered {
			t.Errorf("ステータスがDeliveredになるべき")
		}
	})

	t.Run("ConfirmWakeUp", func(t *testing.T) {
		mc := &MorningCall{
			Status: valueobject.MorningCallStatusDelivered,
		}
		reason := mc.ConfirmWakeUp()
		if reason.IsNG() {
			t.Errorf("起床確認に失敗: %s", reason.Error())
		}
		if mc.Status != valueobject.MorningCallStatusConfirmed {
			t.Errorf("ステータスがConfirmedになるべき")
		}
	})

	t.Run("MarkAsExpired", func(t *testing.T) {
		mc := &MorningCall{
			Status: valueobject.MorningCallStatusScheduled,
		}
		reason := mc.MarkAsExpired()
		if reason.IsNG() {
			t.Errorf("期限切れへの変更に失敗: %s", reason.Error())
		}
		if mc.Status != valueobject.MorningCallStatusExpired {
			t.Errorf("ステータスがExpiredになるべき")
		}
	})
}

func TestMorningCall_UpdateMessage(t *testing.T) {
	tests := []struct {
		name           string
		status         valueobject.MorningCallStatus
		initialMessage string
		newMessage     string
		expectError    bool
		expectedMsg    string
		errorMsg       string
	}{
		{
			name:           "スケジュール済みの場合の更新",
			status:         valueobject.MorningCallStatusScheduled,
			initialMessage: "old message",
			newMessage:     "new message",
			expectError:    false,
			expectedMsg:    "new message",
		},
		{
			name:           "配信済みの場合の更新（不可）",
			status:         valueobject.MorningCallStatusDelivered,
			initialMessage: "old message",
			newMessage:     "new message",
			expectError:    true,
			expectedMsg:    "old message",
			errorMsg:       "スケジュール済みのモーニングコールのみ更新できます",
		},
		{
			name:           "長すぎるメッセージ",
			status:         valueobject.MorningCallStatusScheduled,
			initialMessage: "old message",
			newMessage:     strings.Repeat("あ", 501),
			expectError:    true,
			expectedMsg:    "old message",
			errorMsg:       "メッセージは500文字以内で入力してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MorningCall{
				Status:    tt.status,
				Message:   tt.initialMessage,
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			}

			oldUpdatedAt := mc.UpdatedAt
			reason := mc.UpdateMessage(tt.newMessage)

			if tt.expectError {
				if reason.IsOK() {
					t.Errorf("エラーが期待されたが、成功した")
				}
				if reason.Error() != tt.errorMsg {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMsg, reason.Error())
				}
				if mc.Message != tt.expectedMsg {
					t.Errorf("メッセージ: expected %s, got %s", tt.expectedMsg, mc.Message)
				}
				if !mc.UpdatedAt.Equal(oldUpdatedAt) {
					t.Errorf("エラー時はUpdatedAtが更新されないべき")
				}
			} else {
				if reason.IsNG() {
					t.Errorf("成功が期待されたが、エラーが発生: %s", reason.Error())
				}
				if mc.Message != tt.expectedMsg {
					t.Errorf("メッセージ: expected %s, got %s", tt.expectedMsg, mc.Message)
				}
				if mc.UpdatedAt.Equal(oldUpdatedAt) || mc.UpdatedAt.Before(oldUpdatedAt) {
					t.Errorf("成功時はUpdatedAtが更新されるべき")
				}
			}
		})
	}
}

func TestMorningCall_UpdateScheduledTime(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(2 * time.Hour)
	pastTime := now.Add(-1 * time.Hour)
	farFutureTime := now.Add(31 * 24 * time.Hour)

	tests := []struct {
		name        string
		status      valueobject.MorningCallStatus
		initialTime time.Time
		newTime     time.Time
		expectError bool
		errorMsg    string
	}{
		{
			name:        "スケジュール済みの場合の更新",
			status:      valueobject.MorningCallStatusScheduled,
			initialTime: now.Add(1 * time.Hour),
			newTime:     futureTime,
			expectError: false,
		},
		{
			name:        "配信済みの場合の更新（不可）",
			status:      valueobject.MorningCallStatusDelivered,
			initialTime: now.Add(1 * time.Hour),
			newTime:     futureTime,
			expectError: true,
			errorMsg:    "スケジュール済みのモーニングコールのみ更新できます",
		},
		{
			name:        "過去の時刻への更新",
			status:      valueobject.MorningCallStatusScheduled,
			initialTime: now.Add(1 * time.Hour),
			newTime:     pastTime,
			expectError: true,
			errorMsg:    "アラーム時刻は現在時刻より後である必要があります",
		},
		{
			name:        "30日以上先への更新",
			status:      valueobject.MorningCallStatusScheduled,
			initialTime: now.Add(1 * time.Hour),
			newTime:     farFutureTime,
			expectError: true,
			errorMsg:    "アラーム時刻は30日以内で設定してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MorningCall{
				Status:        tt.status,
				ScheduledTime: tt.initialTime,
				UpdatedAt:     now.Add(-1 * time.Hour),
			}

			oldUpdatedAt := mc.UpdatedAt
			reason := mc.UpdateScheduledTime(tt.newTime)

			if tt.expectError {
				if reason.IsOK() {
					t.Errorf("エラーが期待されたが、成功した")
				}
				if reason.Error() != tt.errorMsg {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMsg, reason.Error())
				}
				if !mc.ScheduledTime.Equal(tt.initialTime) {
					t.Errorf("エラー時は時刻が変更されないべき")
				}
				if !mc.UpdatedAt.Equal(oldUpdatedAt) {
					t.Errorf("エラー時はUpdatedAtが更新されないべき")
				}
			} else {
				if reason.IsNG() {
					t.Errorf("成功が期待されたが、エラーが発生: %s", reason.Error())
				}
				if !mc.ScheduledTime.Equal(tt.newTime) {
					t.Errorf("時刻が更新されるべき")
				}
				if mc.UpdatedAt.Equal(oldUpdatedAt) || mc.UpdatedAt.Before(oldUpdatedAt) {
					t.Errorf("成功時はUpdatedAtが更新されるべき")
				}
			}
		})
	}
}

func TestMorningCall_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   valueobject.MorningCallStatus
		expected bool
	}{
		{
			name:     "スケジュール済みはアクティブ",
			status:   valueobject.MorningCallStatusScheduled,
			expected: true,
		},
		{
			name:     "配信済みはアクティブ",
			status:   valueobject.MorningCallStatusDelivered,
			expected: true,
		},
		{
			name:     "確認済みは非アクティブ",
			status:   valueobject.MorningCallStatusConfirmed,
			expected: false,
		},
		{
			name:     "キャンセル済みは非アクティブ",
			status:   valueobject.MorningCallStatusCancelled,
			expected: false,
		},
		{
			name:     "期限切れは非アクティブ",
			status:   valueobject.MorningCallStatusExpired,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MorningCall{
				Status: tt.status,
			}
			if got := mc.IsActive(); got != tt.expected {
				t.Errorf("IsActive() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestMorningCall_IsPast(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		scheduledTime time.Time
		expected      bool
	}{
		{
			name:          "過去の時刻",
			scheduledTime: now.Add(-1 * time.Hour),
			expected:      true,
		},
		{
			name:          "未来の時刻",
			scheduledTime: now.Add(1 * time.Hour),
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MorningCall{
				ScheduledTime: tt.scheduledTime,
			}
			if got := mc.IsPast(); got != tt.expected {
				t.Errorf("IsPast() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestMorningCall_ShouldDeliver(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		status        valueobject.MorningCallStatus
		scheduledTime time.Time
		expected      bool
	}{
		{
			name:          "スケジュール済み＆過去",
			status:        valueobject.MorningCallStatusScheduled,
			scheduledTime: now.Add(-1 * time.Hour),
			expected:      true,
		},
		{
			name:          "スケジュール済み＆未来",
			status:        valueobject.MorningCallStatusScheduled,
			scheduledTime: now.Add(1 * time.Hour),
			expected:      false,
		},
		{
			name:          "配信済み＆過去",
			status:        valueobject.MorningCallStatusDelivered,
			scheduledTime: now.Add(-1 * time.Hour),
			expected:      false,
		},
		{
			name:          "キャンセル済み＆過去",
			status:        valueobject.MorningCallStatusCancelled,
			scheduledTime: now.Add(-1 * time.Hour),
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MorningCall{
				Status:        tt.status,
				ScheduledTime: tt.scheduledTime,
			}
			if got := mc.ShouldDeliver(); got != tt.expected {
				t.Errorf("ShouldDeliver() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestMorningCall_Equals(t *testing.T) {
	mc1 := &MorningCall{
		ID:         "mc-001",
		SenderID:   "user-001",
		ReceiverID: "user-002",
	}

	mc2 := &MorningCall{
		ID:         "mc-001",
		SenderID:   "user-003",
		ReceiverID: "user-004",
	}

	mc3 := &MorningCall{
		ID:         "mc-002",
		SenderID:   "user-001",
		ReceiverID: "user-002",
	}

	tests := []struct {
		name     string
		mc       *MorningCall
		other    *MorningCall
		expected bool
	}{
		{
			name:     "同じIDのモーニングコール",
			mc:       mc1,
			other:    mc2,
			expected: true,
		},
		{
			name:     "異なるIDのモーニングコール",
			mc:       mc1,
			other:    mc3,
			expected: false,
		},
		{
			name:     "nilとの比較",
			mc:       mc1,
			other:    nil,
			expected: false,
		},
		{
			name:     "自分自身との比較",
			mc:       mc1,
			other:    mc1,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mc.Equals(tt.other)
			if result != tt.expected {
				t.Errorf("Equals() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
