package idgen

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	gen := New()

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
		})
	}
}

func TestConvenienceMethods(t *testing.T) {
	gen := New()

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
		{"UserProfileID", gen.UserProfileID, PrefixUserProfile},
		{"IdempotencyKey", gen.IdempotencyKey, PrefixIdempotency},
		{"LockingID", gen.LockingID, PrefixLocking},
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

func TestUniqueness(t *testing.T) {
	gen := New()

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
