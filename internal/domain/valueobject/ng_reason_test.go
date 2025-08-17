package valueobject

import "testing"

func TestNGReason_IsOK(t *testing.T) {
	tests := []struct {
		name     string
		reason   NGReason
		expected bool
	}{
		{
			name:     "空文字列はOK",
			reason:   NGReason(""),
			expected: true,
		},
		{
			name:     "OK関数で生成した値はOK",
			reason:   OK(),
			expected: true,
		},
		{
			name:     "エラーメッセージありはNG",
			reason:   NGReason("エラー"),
			expected: false,
		},
		{
			name:     "NG関数で生成した値はNG",
			reason:   NG("エラー"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.reason.IsOK(); got != tt.expected {
				t.Errorf("IsOK() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNGReason_IsNG(t *testing.T) {
	tests := []struct {
		name     string
		reason   NGReason
		expected bool
	}{
		{
			name:     "空文字列はNG",
			reason:   NGReason(""),
			expected: false,
		},
		{
			name:     "OK関数で生成した値はNG",
			reason:   OK(),
			expected: false,
		},
		{
			name:     "エラーメッセージありはNG",
			reason:   NGReason("エラー"),
			expected: true,
		},
		{
			name:     "NG関数で生成した値はNG",
			reason:   NG("エラー"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.reason.IsNG(); got != tt.expected {
				t.Errorf("IsNG() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNGReason_Error(t *testing.T) {
	tests := []struct {
		name     string
		reason   NGReason
		expected string
	}{
		{
			name:     "空文字列",
			reason:   NGReason(""),
			expected: "",
		},
		{
			name:     "エラーメッセージ",
			reason:   NGReason("検証エラー"),
			expected: "検証エラー",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.reason.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}
