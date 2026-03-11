package entity

import "testing"

func TestOrderStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status   OrderStatus
		terminal bool
	}{
		{OrderPending, false},
		{OrderSuccess, true},
		{OrderComplete, true},
		{OrderFailed, true},
		{OrderRefunded, true},
		{OrderExpired, true},
		{OrderCancelled, true},
		{"", false},
		{"IN_PROGRESS", false},
		{"unknown", false},
	}
	for _, tt := range tests {
		got := tt.status.IsTerminal()
		if got != tt.terminal {
			t.Errorf("OrderStatus(%q).IsTerminal() = %v, want %v", tt.status, got, tt.terminal)
		}
	}
}

func TestOrderStatus_OrDefault(t *testing.T) {
	tests := []struct {
		status OrderStatus
		want   OrderStatus
	}{
		{"", OrderPending},
		{"SUCCESS", OrderSuccess},
		{OrderPending, OrderPending},
		{OrderFailed, OrderFailed},
	}
	for _, tt := range tests {
		got := tt.status.OrDefault()
		if got != tt.want {
			t.Errorf("OrderStatus(%q).OrDefault() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestOrderStatus_NormalizeForDB(t *testing.T) {
	if got := OrderComplete.NormalizeForDB(); got != OrderSuccess {
		t.Errorf("COMPLETE.NormalizeForDB() = %q, want SUCCESS", got)
	}
	if got := OrderSuccess.NormalizeForDB(); got != OrderSuccess {
		t.Errorf("SUCCESS.NormalizeForDB() = %q, want SUCCESS", got)
	}
	if got := OrderPending.NormalizeForDB(); got != OrderPending {
		t.Errorf("PENDING.NormalizeForDB() = %q, want PENDING", got)
	}
}
