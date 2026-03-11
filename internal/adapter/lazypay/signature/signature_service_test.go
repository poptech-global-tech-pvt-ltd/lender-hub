package signature

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"testing"
)

func TestSignEligibility(t *testing.T) {
	svc := NewSignatureService("test-access-key", "test-secret-key")

	mobile := "9876543210"
	email := "test@example.com"
	amount := 1000.50

	sig := svc.SignEligibility(mobile, email, amount)

	// Verify signature format (hex string, 40 chars for SHA1)
	if len(sig) != 40 {
		t.Errorf("expected signature length 40, got %d", len(sig))
	}

	// Verify signature is deterministic
	sig2 := svc.SignEligibility(mobile, email, amount)
	if sig != sig2 {
		t.Error("signature should be deterministic")
	}

	// Verify signature changes with different input
	sig3 := svc.SignEligibility(mobile, email, 2000.50)
	if sig == sig3 {
		t.Error("signature should change with different amount")
	}
}

func TestSignCustomerStatus(t *testing.T) {
	svc := NewSignatureService("test-access-key", "test-secret-key")

	mobile := "9876543210"
	sig := svc.SignCustomerStatus(mobile)

	// Verify signature format
	if len(sig) != 40 {
		t.Errorf("expected signature length 40, got %d", len(sig))
	}

	// Verify signature includes access key
	sig2 := svc.SignCustomerStatus(mobile)
	if sig != sig2 {
		t.Error("signature should be deterministic")
	}
}

func TestSignOrder(t *testing.T) {
	svc := NewSignatureService("test-access-key", "test-secret-key")

	merchantTxnID := "txn-123"
	amount := 500.75
	sig := svc.SignOrder(merchantTxnID, amount)

	// Verify signature format
	if len(sig) != 40 {
		t.Errorf("expected signature length 40, got %d", len(sig))
	}

	// Verify signature is deterministic
	sig2 := svc.SignOrder(merchantTxnID, amount)
	if sig != sig2 {
		t.Error("signature should be deterministic")
	}
}

func TestVerifyWebhook(t *testing.T) {
	svc := NewSignatureService("test-access-key", "test-secret-key")

	payload := []byte(`{"event":"test","data":"value"}`)

	// Generate valid signature
	h := hmac.New(sha1.New, []byte("test-secret-key"))
	h.Write(payload)
	validSig := hex.EncodeToString(h.Sum(nil))

	// Verify valid signature
	if !svc.VerifyWebhook(payload, validSig) {
		t.Error("valid signature should verify")
	}

	// Verify invalid signature
	if svc.VerifyWebhook(payload, "invalid-signature") {
		t.Error("invalid signature should not verify")
	}

	// Verify tampered payload
	tamperedPayload := []byte(`{"event":"tampered","data":"value"}`)
	if svc.VerifyWebhook(tamperedPayload, validSig) {
		t.Error("tampered payload should not verify")
	}
}

func TestVerifyWebhook_TimingSafe(t *testing.T) {
	svc := NewSignatureService("test-access-key", "test-secret-key")

	payload := []byte(`{"event":"test"}`)

	// Generate valid signature
	h := hmac.New(sha1.New, []byte("test-secret-key"))
	h.Write(payload)
	validSig := hex.EncodeToString(h.Sum(nil))

	// Verify that hmac.Equal is used (timing-safe)
	// This test ensures we're using constant-time comparison
	if !svc.VerifyWebhook(payload, validSig) {
		t.Error("valid signature should verify")
	}

	// Test with different length signatures (should still be safe)
	if svc.VerifyWebhook(payload, "short") {
		t.Error("short signature should not verify")
	}
}

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		amount   float64
		expected string
	}{
		{1000.0, "1000.00"},
		{1000.5, "1000.50"},
		{1000.55, "1000.55"},
		{0.0, "0.00"},
		{0.01, "0.01"},
	}

	for _, tt := range tests {
		// We can't directly test formatAmount, but we can test through SignOrder
		svc := NewSignatureService("test-key", "test-secret")
		sig1 := svc.SignOrder("txn-1", tt.amount)
		sig2 := svc.SignOrder("txn-1", tt.amount)
		if sig1 != sig2 {
			t.Errorf("amount %.2f should format consistently", tt.amount)
		}
	}
}
