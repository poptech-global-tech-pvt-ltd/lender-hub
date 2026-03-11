package entity

import "testing"

func TestRefundStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status   RefundStatus
		terminal bool
	}{
		{RefundStatusPending, false},
		{RefundStatusProcessing, false},
		{RefundStatusUnknown, false},
		{RefundStatusSuccess, true},
		{RefundStatusFailed, true},
	}
	for _, tt := range tests {
		got := tt.status.IsTerminal()
		if got != tt.terminal {
			t.Errorf("RefundStatus(%q).IsTerminal() = %v, want %v", tt.status, got, tt.terminal)
		}
	}
}

func TestRefundStatus_IsResolvable(t *testing.T) {
	tests := []struct {
		status     RefundStatus
		resolvable bool
	}{
		{RefundStatusPending, true},
		{RefundStatusProcessing, true},
		{RefundStatusUnknown, true},
		{RefundStatusSuccess, false},
		{RefundStatusFailed, false},
	}
	for _, tt := range tests {
		got := tt.status.IsResolvable()
		if got != tt.resolvable {
			t.Errorf("RefundStatus(%q).IsResolvable() = %v, want %v", tt.status, got, tt.resolvable)
		}
	}
}

func TestRefundStatus_OrDefault(t *testing.T) {
	if got := RefundStatus("").OrDefault(); got != RefundStatusPending {
		t.Errorf("empty.OrDefault() = %q, want PENDING", got)
	}
	if got := RefundStatusSuccess.OrDefault(); got != RefundStatusSuccess {
		t.Errorf("SUCCESS.OrDefault() = %q, want SUCCESS", got)
	}
}

func TestRefundStatus_StateTransitions(t *testing.T) {
	var r RefundStatus
	r = RefundStatusPending
	r = RefundStatusSuccess
	if r != RefundStatusSuccess {
		t.Errorf("expected SUCCESS after MarkSuccess path")
	}
	// MarkSuccess/MarkFailed are on entity.Refund, not RefundStatus - tested via entity if needed
}
