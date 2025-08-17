package entity

import (
	"testing"
	"time"

	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
)

func TestNewRelationship(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		requesterID string
		receiverID  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "正常な友達関係作成",
			id:          "rel-001",
			requesterID: "user-001",
			receiverID:  "user-002",
			expectError: false,
		},
		{
			name:        "IDが空",
			id:          "",
			requesterID: "user-001",
			receiverID:  "user-002",
			expectError: true,
			errorMsg:    "関係IDは必須です",
		},
		{
			name:        "リクエスト送信者IDが空",
			id:          "rel-001",
			requesterID: "",
			receiverID:  "user-002",
			expectError: true,
			errorMsg:    "リクエスト送信者IDは必須です",
		},
		{
			name:        "リクエスト受信者IDが空",
			id:          "rel-001",
			requesterID: "user-001",
			receiverID:  "",
			expectError: true,
			errorMsg:    "リクエスト受信者IDは必須です",
		},
		{
			name:        "自分自身への友達リクエスト",
			id:          "rel-001",
			requesterID: "user-001",
			receiverID:  "user-001",
			expectError: true,
			errorMsg:    "自分自身に友達リクエストを送ることはできません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel, reason := NewRelationship(tt.id, tt.requesterID, tt.receiverID)

			if tt.expectError {
				if reason.IsOK() {
					t.Errorf("エラーが期待されたが、成功した")
				}
				if reason.Error() != tt.errorMsg {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMsg, reason.Error())
				}
				if rel != nil {
					t.Errorf("エラー時にはnilが期待されたが、関係が返された")
				}
			} else {
				if reason.IsNG() {
					t.Errorf("成功が期待されたが、エラーが発生: %s", reason.Error())
				}
				if rel == nil {
					t.Errorf("成功時には関係が期待されたが、nilが返された")
				} else {
					if rel.ID != tt.id {
						t.Errorf("ID: expected %s, got %s", tt.id, rel.ID)
					}
					if rel.RequesterID != tt.requesterID {
						t.Errorf("RequesterID: expected %s, got %s", tt.requesterID, rel.RequesterID)
					}
					if rel.ReceiverID != tt.receiverID {
						t.Errorf("ReceiverID: expected %s, got %s", tt.receiverID, rel.ReceiverID)
					}
					if rel.Status != valueobject.RelationshipStatusPending {
						t.Errorf("初期ステータスはPendingであるべき")
					}
				}
			}
		})
	}
}

func TestRelationship_UpdateStatus(t *testing.T) {
	tests := []struct {
		name        string
		fromStatus  valueobject.RelationshipStatus
		toStatus    valueobject.RelationshipStatus
		expectError bool
		errorMsg    string
	}{
		{
			name:        "承認待ち→承認済み",
			fromStatus:  valueobject.RelationshipStatusPending,
			toStatus:    valueobject.RelationshipStatusAccepted,
			expectError: false,
		},
		{
			name:        "承認待ち→拒否済み",
			fromStatus:  valueobject.RelationshipStatusPending,
			toStatus:    valueobject.RelationshipStatusRejected,
			expectError: false,
		},
		{
			name:        "承認待ち→ブロック",
			fromStatus:  valueobject.RelationshipStatusPending,
			toStatus:    valueobject.RelationshipStatusBlocked,
			expectError: false,
		},
		{
			name:        "承認済み→ブロック",
			fromStatus:  valueobject.RelationshipStatusAccepted,
			toStatus:    valueobject.RelationshipStatusBlocked,
			expectError: false,
		},
		{
			name:        "拒否済み→承認待ち（再送信）",
			fromStatus:  valueobject.RelationshipStatusRejected,
			toStatus:    valueobject.RelationshipStatusPending,
			expectError: false,
		},
		{
			name:        "ブロック→承認済み（不可）",
			fromStatus:  valueobject.RelationshipStatusBlocked,
			toStatus:    valueobject.RelationshipStatusAccepted,
			expectError: true,
			errorMsg:    "このステータスへの遷移はできません",
		},
		{
			name:        "承認済み→承認待ち（不可）",
			fromStatus:  valueobject.RelationshipStatusAccepted,
			toStatus:    valueobject.RelationshipStatusPending,
			expectError: true,
			errorMsg:    "このステータスへの遷移はできません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := &Relationship{
				ID:        "rel-001",
				Status:    tt.fromStatus,
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			}

			oldUpdatedAt := rel.UpdatedAt
			reason := rel.UpdateStatus(tt.toStatus)

			if tt.expectError {
				if reason.IsOK() {
					t.Errorf("エラーが期待されたが、成功した")
				}
				if reason.Error() != tt.errorMsg {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMsg, reason.Error())
				}
				if rel.Status != tt.fromStatus {
					t.Errorf("エラー時はステータスが変更されないべき")
				}
				if !rel.UpdatedAt.Equal(oldUpdatedAt) {
					t.Errorf("エラー時はUpdatedAtが更新されないべき")
				}
			} else {
				if reason.IsNG() {
					t.Errorf("成功が期待されたが、エラーが発生: %s", reason.Error())
				}
				if rel.Status != tt.toStatus {
					t.Errorf("ステータス: expected %s, got %s", tt.toStatus, rel.Status)
				}
				if rel.UpdatedAt.Equal(oldUpdatedAt) || rel.UpdatedAt.Before(oldUpdatedAt) {
					t.Errorf("成功時はUpdatedAtが更新されるべき")
				}
			}
		})
	}
}

func TestRelationship_Accept(t *testing.T) {
	tests := []struct {
		name        string
		status      valueobject.RelationshipStatus
		expectError bool
		errorMsg    string
	}{
		{
			name:        "承認待ちから承認",
			status:      valueobject.RelationshipStatusPending,
			expectError: false,
		},
		{
			name:        "承認済みから承認（不可）",
			status:      valueobject.RelationshipStatusAccepted,
			expectError: true,
			errorMsg:    "承認待ち状態のリクエストのみ承認できます",
		},
		{
			name:        "拒否済みから承認（不可）",
			status:      valueobject.RelationshipStatusRejected,
			expectError: true,
			errorMsg:    "承認待ち状態のリクエストのみ承認できます",
		},
		{
			name:        "ブロック済みから承認（不可）",
			status:      valueobject.RelationshipStatusBlocked,
			expectError: true,
			errorMsg:    "承認待ち状態のリクエストのみ承認できます",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := &Relationship{
				Status: tt.status,
			}
			reason := rel.Accept()

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
				if rel.Status != valueobject.RelationshipStatusAccepted {
					t.Errorf("ステータスがAcceptedになるべき")
				}
			}
		})
	}
}

func TestRelationship_Reject(t *testing.T) {
	tests := []struct {
		name        string
		status      valueobject.RelationshipStatus
		expectError bool
		errorMsg    string
	}{
		{
			name:        "承認待ちから拒否",
			status:      valueobject.RelationshipStatusPending,
			expectError: false,
		},
		{
			name:        "承認済みから拒否（不可）",
			status:      valueobject.RelationshipStatusAccepted,
			expectError: true,
			errorMsg:    "承認待ち状態のリクエストのみ拒否できます",
		},
		{
			name:        "拒否済みから拒否（不可）",
			status:      valueobject.RelationshipStatusRejected,
			expectError: true,
			errorMsg:    "承認待ち状態のリクエストのみ拒否できます",
		},
		{
			name:        "ブロック済みから拒否（不可）",
			status:      valueobject.RelationshipStatusBlocked,
			expectError: true,
			errorMsg:    "承認待ち状態のリクエストのみ拒否できます",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := &Relationship{
				Status: tt.status,
			}
			reason := rel.Reject()

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
				if rel.Status != valueobject.RelationshipStatusRejected {
					t.Errorf("ステータスがRejectedになるべき")
				}
			}
		})
	}
}

func TestRelationship_Block(t *testing.T) {
	tests := []struct {
		name        string
		status      valueobject.RelationshipStatus
		expectError bool
		errorMsg    string
	}{
		{
			name:        "承認待ちからブロック",
			status:      valueobject.RelationshipStatusPending,
			expectError: false,
		},
		{
			name:        "承認済みからブロック",
			status:      valueobject.RelationshipStatusAccepted,
			expectError: false,
		},
		{
			name:        "拒否済みからブロック",
			status:      valueobject.RelationshipStatusRejected,
			expectError: false,
		},
		{
			name:        "ブロック済みからブロック（不可）",
			status:      valueobject.RelationshipStatusBlocked,
			expectError: true,
			errorMsg:    "既にブロック済みです",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := &Relationship{
				Status: tt.status,
			}
			reason := rel.Block()

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
				if rel.Status != valueobject.RelationshipStatusBlocked {
					t.Errorf("ステータスがBlockedになるべき")
				}
			}
		})
	}
}

func TestRelationship_Resend(t *testing.T) {
	tests := []struct {
		name        string
		status      valueobject.RelationshipStatus
		expectError bool
		errorMsg    string
	}{
		{
			name:        "拒否済みから再送信",
			status:      valueobject.RelationshipStatusRejected,
			expectError: false,
		},
		{
			name:        "承認待ちから再送信（不可）",
			status:      valueobject.RelationshipStatusPending,
			expectError: true,
			errorMsg:    "拒否済みのリクエストのみ再送信できます",
		},
		{
			name:        "承認済みから再送信（不可）",
			status:      valueobject.RelationshipStatusAccepted,
			expectError: true,
			errorMsg:    "拒否済みのリクエストのみ再送信できます",
		},
		{
			name:        "ブロック済みから再送信（不可）",
			status:      valueobject.RelationshipStatusBlocked,
			expectError: true,
			errorMsg:    "拒否済みのリクエストのみ再送信できます",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := &Relationship{
				Status: tt.status,
			}
			reason := rel.Resend()

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
				if rel.Status != valueobject.RelationshipStatusPending {
					t.Errorf("ステータスがPendingになるべき")
				}
			}
		})
	}
}

func TestRelationship_StatusChecks(t *testing.T) {
	tests := []struct {
		name       string
		status     valueobject.RelationshipStatus
		isFriend   bool
		isBlocked  bool
		isPending  bool
		isRejected bool
	}{
		{
			name:       "承認待ち",
			status:     valueobject.RelationshipStatusPending,
			isFriend:   false,
			isBlocked:  false,
			isPending:  true,
			isRejected: false,
		},
		{
			name:       "承認済み（友達）",
			status:     valueobject.RelationshipStatusAccepted,
			isFriend:   true,
			isBlocked:  false,
			isPending:  false,
			isRejected: false,
		},
		{
			name:       "ブロック済み",
			status:     valueobject.RelationshipStatusBlocked,
			isFriend:   false,
			isBlocked:  true,
			isPending:  false,
			isRejected: false,
		},
		{
			name:       "拒否済み",
			status:     valueobject.RelationshipStatusRejected,
			isFriend:   false,
			isBlocked:  false,
			isPending:  false,
			isRejected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := &Relationship{
				Status: tt.status,
			}

			if got := rel.IsFriend(); got != tt.isFriend {
				t.Errorf("IsFriend() = %v, expected %v", got, tt.isFriend)
			}
			if got := rel.IsBlocked(); got != tt.isBlocked {
				t.Errorf("IsBlocked() = %v, expected %v", got, tt.isBlocked)
			}
			if got := rel.IsPending(); got != tt.isPending {
				t.Errorf("IsPending() = %v, expected %v", got, tt.isPending)
			}
			if got := rel.IsRejected(); got != tt.isRejected {
				t.Errorf("IsRejected() = %v, expected %v", got, tt.isRejected)
			}
		})
	}
}

func TestRelationship_UserRelatedMethods(t *testing.T) {
	rel := &Relationship{
		RequesterID: "user-001",
		ReceiverID:  "user-002",
	}

	t.Run("InvolvesUser", func(t *testing.T) {
		tests := []struct {
			userID   string
			expected bool
		}{
			{"user-001", true},  // リクエスター
			{"user-002", true},  // レシーバー
			{"user-003", false}, // 無関係
		}

		for _, tt := range tests {
			if got := rel.InvolvesUser(tt.userID); got != tt.expected {
				t.Errorf("InvolvesUser(%s) = %v, expected %v", tt.userID, got, tt.expected)
			}
		}
	})

	t.Run("IsRequester", func(t *testing.T) {
		tests := []struct {
			userID   string
			expected bool
		}{
			{"user-001", true},
			{"user-002", false},
			{"user-003", false},
		}

		for _, tt := range tests {
			if got := rel.IsRequester(tt.userID); got != tt.expected {
				t.Errorf("IsRequester(%s) = %v, expected %v", tt.userID, got, tt.expected)
			}
		}
	})

	t.Run("IsReceiver", func(t *testing.T) {
		tests := []struct {
			userID   string
			expected bool
		}{
			{"user-001", false},
			{"user-002", true},
			{"user-003", false},
		}

		for _, tt := range tests {
			if got := rel.IsReceiver(tt.userID); got != tt.expected {
				t.Errorf("IsReceiver(%s) = %v, expected %v", tt.userID, got, tt.expected)
			}
		}
	})

	t.Run("GetOtherUserID", func(t *testing.T) {
		tests := []struct {
			userID   string
			expected string
		}{
			{"user-001", "user-002"}, // リクエスターから見た相手
			{"user-002", "user-001"}, // レシーバーから見た相手
			{"user-003", ""},         // 無関係な場合は空文字
		}

		for _, tt := range tests {
			if got := rel.GetOtherUserID(tt.userID); got != tt.expected {
				t.Errorf("GetOtherUserID(%s) = %v, expected %v", tt.userID, got, tt.expected)
			}
		}
	})
}

func TestRelationship_PermissionMethods(t *testing.T) {
	t.Run("CanBeAcceptedBy", func(t *testing.T) {
		tests := []struct {
			name        string
			requesterID string
			receiverID  string
			status      valueobject.RelationshipStatus
			userID      string
			expected    bool
		}{
			{
				name:        "受信者が承認待ちを承認可能",
				requesterID: "user-001",
				receiverID:  "user-002",
				status:      valueobject.RelationshipStatusPending,
				userID:      "user-002",
				expected:    true,
			},
			{
				name:        "送信者は承認不可",
				requesterID: "user-001",
				receiverID:  "user-002",
				status:      valueobject.RelationshipStatusPending,
				userID:      "user-001",
				expected:    false,
			},
			{
				name:        "承認済みは承認不可",
				requesterID: "user-001",
				receiverID:  "user-002",
				status:      valueobject.RelationshipStatusAccepted,
				userID:      "user-002",
				expected:    false,
			},
			{
				name:        "無関係なユーザーは承認不可",
				requesterID: "user-001",
				receiverID:  "user-002",
				status:      valueobject.RelationshipStatusPending,
				userID:      "user-003",
				expected:    false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				rel := &Relationship{
					RequesterID: tt.requesterID,
					ReceiverID:  tt.receiverID,
					Status:      tt.status,
				}
				if got := rel.CanBeAcceptedBy(tt.userID); got != tt.expected {
					t.Errorf("CanBeAcceptedBy(%s) = %v, expected %v", tt.userID, got, tt.expected)
				}
			})
		}
	})

	t.Run("CanBeRejectedBy", func(t *testing.T) {
		tests := []struct {
			name        string
			requesterID string
			receiverID  string
			status      valueobject.RelationshipStatus
			userID      string
			expected    bool
		}{
			{
				name:        "受信者が承認待ちを拒否可能",
				requesterID: "user-001",
				receiverID:  "user-002",
				status:      valueobject.RelationshipStatusPending,
				userID:      "user-002",
				expected:    true,
			},
			{
				name:        "送信者は拒否不可",
				requesterID: "user-001",
				receiverID:  "user-002",
				status:      valueobject.RelationshipStatusPending,
				userID:      "user-001",
				expected:    false,
			},
			{
				name:        "承認済みは拒否不可",
				requesterID: "user-001",
				receiverID:  "user-002",
				status:      valueobject.RelationshipStatusAccepted,
				userID:      "user-002",
				expected:    false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				rel := &Relationship{
					RequesterID: tt.requesterID,
					ReceiverID:  tt.receiverID,
					Status:      tt.status,
				}
				if got := rel.CanBeRejectedBy(tt.userID); got != tt.expected {
					t.Errorf("CanBeRejectedBy(%s) = %v, expected %v", tt.userID, got, tt.expected)
				}
			})
		}
	})

	t.Run("CanBeBlockedBy", func(t *testing.T) {
		tests := []struct {
			name        string
			requesterID string
			receiverID  string
			status      valueobject.RelationshipStatus
			userID      string
			expected    bool
		}{
			{
				name:        "送信者がブロック可能",
				requesterID: "user-001",
				receiverID:  "user-002",
				status:      valueobject.RelationshipStatusPending,
				userID:      "user-001",
				expected:    true,
			},
			{
				name:        "受信者がブロック可能",
				requesterID: "user-001",
				receiverID:  "user-002",
				status:      valueobject.RelationshipStatusPending,
				userID:      "user-002",
				expected:    true,
			},
			{
				name:        "既にブロック済みはブロック不可",
				requesterID: "user-001",
				receiverID:  "user-002",
				status:      valueobject.RelationshipStatusBlocked,
				userID:      "user-001",
				expected:    false,
			},
			{
				name:        "無関係なユーザーはブロック不可",
				requesterID: "user-001",
				receiverID:  "user-002",
				status:      valueobject.RelationshipStatusPending,
				userID:      "user-003",
				expected:    false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				rel := &Relationship{
					RequesterID: tt.requesterID,
					ReceiverID:  tt.receiverID,
					Status:      tt.status,
				}
				if got := rel.CanBeBlockedBy(tt.userID); got != tt.expected {
					t.Errorf("CanBeBlockedBy(%s) = %v, expected %v", tt.userID, got, tt.expected)
				}
			})
		}
	})

	t.Run("CanBeResendBy", func(t *testing.T) {
		tests := []struct {
			name        string
			requesterID string
			receiverID  string
			status      valueobject.RelationshipStatus
			userID      string
			expected    bool
		}{
			{
				name:        "送信者が拒否済みを再送信可能",
				requesterID: "user-001",
				receiverID:  "user-002",
				status:      valueobject.RelationshipStatusRejected,
				userID:      "user-001",
				expected:    true,
			},
			{
				name:        "受信者は再送信不可",
				requesterID: "user-001",
				receiverID:  "user-002",
				status:      valueobject.RelationshipStatusRejected,
				userID:      "user-002",
				expected:    false,
			},
			{
				name:        "承認待ちは再送信不可",
				requesterID: "user-001",
				receiverID:  "user-002",
				status:      valueobject.RelationshipStatusPending,
				userID:      "user-001",
				expected:    false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				rel := &Relationship{
					RequesterID: tt.requesterID,
					ReceiverID:  tt.receiverID,
					Status:      tt.status,
				}
				if got := rel.CanBeResendBy(tt.userID); got != tt.expected {
					t.Errorf("CanBeResendBy(%s) = %v, expected %v", tt.userID, got, tt.expected)
				}
			})
		}
	})
}

func TestRelationship_Equals(t *testing.T) {
	rel1 := &Relationship{
		ID:          "rel-001",
		RequesterID: "user-001",
		ReceiverID:  "user-002",
	}

	rel2 := &Relationship{
		ID:          "rel-001",
		RequesterID: "user-003",
		ReceiverID:  "user-004",
	}

	rel3 := &Relationship{
		ID:          "rel-002",
		RequesterID: "user-001",
		ReceiverID:  "user-002",
	}

	tests := []struct {
		name     string
		rel      *Relationship
		other    *Relationship
		expected bool
	}{
		{
			name:     "同じIDの関係",
			rel:      rel1,
			other:    rel2,
			expected: true,
		},
		{
			name:     "異なるIDの関係",
			rel:      rel1,
			other:    rel3,
			expected: false,
		},
		{
			name:     "nilとの比較",
			rel:      rel1,
			other:    nil,
			expected: false,
		},
		{
			name:     "自分自身との比較",
			rel:      rel1,
			other:    rel1,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rel.Equals(tt.other)
			if result != tt.expected {
				t.Errorf("Equals() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
