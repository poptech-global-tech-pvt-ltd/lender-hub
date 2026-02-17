package idgen

import (
	"strings"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	gen := NewIDGenerator()

	tests := []struct {
		name   string
		prefix string
	}{
		{"payment", PrefixPayment},
		{"refund", PrefixRefund},
		{"onboarding", PrefixOnboarding},
		{"request", PrefixRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := gen.Generate(tt.prefix)

			// Check format: PREFIX_ULID
			if !strings.HasPrefix(id, tt.prefix+"_") {
				t.Errorf("ID %s does not start with prefix %s_", id, tt.prefix)
			}

			// Check total length: prefix + _ + 26 ULID chars
			expectedLen := len(tt.prefix) + 1 + 26
			if len(id) != expectedLen {
				t.Errorf("ID %s has length %d, expected %d", id, len(id), expectedLen)
			}

			// Validate format
			if !gen.Validate(id, tt.prefix) {
				t.Errorf("Generated ID %s failed validation", id)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	gen := NewIDGenerator()

	tests := []struct {
		name     string
		id       string
		prefix   string
		expected bool
	}{
		{"valid payment", "PAY_01ARZ3NDEKTSV4RRFFQ69G5FAV", PrefixPayment, true},
		{"valid refund", "REF_01ARZ3NDEKTSV4RRFFQ69G5FAV", PrefixRefund, true},
		{"wrong prefix", "PAY_01ARZ3NDEKTSV4RRFFQ69G5FAV", PrefixRefund, false},
		{"no separator", "PAY01ARZ3NDEKTSV4RRFFQ69G5FAV", PrefixPayment, false},
		{"short ULID", "PAY_01ARZ3NDEK", PrefixPayment, false},
		{"invalid chars", "PAY_01ARZ3NDEKTSV4RRFFQ69G5FA!", PrefixPayment, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.Validate(tt.id, tt.prefix)
			if result != tt.expected {
				t.Errorf("Validate(%s, %s) = %v, expected %v", tt.id, tt.prefix, result, tt.expected)
			}
		})
	}
}

func TestExtractTimestamp(t *testing.T) {
	gen := NewIDGenerator()

	now := time.Now().Truncate(time.Millisecond)
	id := gen.GenerateWithTime(PrefixPayment, now)

	extracted, err := gen.ExtractTimestamp(id)
	if err != nil {
		t.Fatalf("ExtractTimestamp failed: %v", err)
	}

	// ULID timestamp has millisecond precision
	if !extracted.Equal(now) {
		t.Errorf("Extracted timestamp %v != expected %v", extracted, now)
	}
}

func TestUniqueness(t *testing.T) {
	gen := NewIDGenerator()

	// Generate 1000 IDs and check for collisions
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := gen.PaymentID()
		if seen[id] {
			t.Fatalf("Collision detected: %s", id)
		}
		seen[id] = true
	}
}

func TestConvenienceMethods(t *testing.T) {
	gen := NewIDGenerator()

	tests := []struct {
		name     string
		generate func() string
		prefix   string
	}{
		{"PaymentID", gen.PaymentID, PrefixPayment},
		{"RefundID", gen.RefundID, PrefixRefund},
		{"OnboardingID", gen.OnboardingID, PrefixOnboarding},
		{"RequestID", gen.RequestID, PrefixRequest},
		{"WebhookID", gen.WebhookID, PrefixWebhook},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.generate()
			if !strings.HasPrefix(id, tt.prefix+"_") {
				t.Errorf("%s() = %s, expected prefix %s_", tt.name, id, tt.prefix)
			}
		})
	}
}
