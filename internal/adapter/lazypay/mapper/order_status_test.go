package mapper

import (
	"testing"

	"lending-hub-service/internal/domain/order/entity"
)

// TestMapLPOrderStatusToCanonical verifies LP order/sale status mapping to canonical OrderStatus.
// The mapping is: OrderStatus(lpOrderStatus).OrDefault().NormalizeForDB()
func TestMapLPOrderStatusToCanonical(t *testing.T) {
	mapStatus := func(lpOrderStatus, saleTxnStatus string) entity.OrderStatus {
		// Enquiry returns order-level status; sale txn may differ. Use order status as primary.
		s := entity.OrderStatus(lpOrderStatus)
		if s == "" {
			s = entity.OrderStatus(saleTxnStatus)
		}
		return s.OrDefault().NormalizeForDB()
	}

	tests := []struct {
		lpOrderStatus string
		saleTxnStatus string
		expected      entity.OrderStatus
	}{
		{"COMPLETE", "SUCCESS", entity.OrderSuccess},
		{"COMPLETE", "FAILED", entity.OrderSuccess},
		{"FAILED", "", entity.OrderFailed},
		{"FAILED", "SUCCESS", entity.OrderFailed},
		{"EXPIRED", "", entity.OrderExpired},
		{"EXPIRED", "SUCCESS", entity.OrderExpired},
		{"CANCELLED", "", entity.OrderCancelled},
		{"IN_PROGRESS", "", entity.OrderStatus("IN_PROGRESS")}, // non-empty stays as-is per OrDefault
		{"", "", entity.OrderPending},
		{"COMPLETE", "PROCESSING", entity.OrderSuccess},
		{"unknown_value", "unknown_value", entity.OrderStatus("unknown_value")}, // non-empty stays as-is
		{"PENDING", "", entity.OrderPending},
		{"SUCCESS", "SUCCESS", entity.OrderSuccess},
	}

	for _, tt := range tests {
		got := mapStatus(tt.lpOrderStatus, tt.saleTxnStatus)
		if got != tt.expected {
			t.Errorf("mapStatus(%q, %q) = %q, want %q", tt.lpOrderStatus, tt.saleTxnStatus, got, tt.expected)
		}
	}
}
