package signature

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

// SignatureService handles HMAC-SHA1 signature generation and verification
type SignatureService struct {
	secretKey string
	accessKey string
}

// NewSignatureService creates a new signature service
func NewSignatureService(accessKey, secretKey string) *SignatureService {
	return &SignatureService{
		accessKey: accessKey,
		secretKey: secretKey,
	}
}

// hmacSHA1 generates HMAC-SHA1 hex digest
func (s *SignatureService) hmacSHA1(data string) string {
	h := hmac.New(sha1.New, []byte(s.secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// SignEligibility signs eligibility request
// data = mobile + email + orderAmount + "INR"
func (s *SignatureService) SignEligibility(mobile, email string, amount float64) string {
	data := mobile + email + formatAmount(amount) + "INR"
	return s.hmacSHA1(data)
}

// SignCustomerStatus signs customer status request
// data = accessKey + mobile
func (s *SignatureService) SignCustomerStatus(mobile string) string {
	data := s.accessKey + mobile
	return s.hmacSHA1(data)
}

// SignOrder signs order creation request
// data = accessKey + merchantTxnId + amount + "INR"
func (s *SignatureService) SignOrder(merchantTxnID string, amount float64) string {
	data := s.accessKey + merchantTxnID + formatAmount(amount) + "INR"
	return s.hmacSHA1(data)
}

// SignEnquiry signs enquiry request
// data = merchantTxnId + secretKey
func (s *SignatureService) SignEnquiry(merchantTxnID string) string {
	data := merchantTxnID + s.secretKey
	return s.hmacSHA1(data)
}

// VerifyWebhook compares received signature against computed one
func (s *SignatureService) VerifyWebhook(payload []byte, receivedSig string) bool {
	h := hmac.New(sha1.New, []byte(s.secretKey))
	h.Write(payload)
	expected := hex.EncodeToString(h.Sum(nil))
	// Use hmac.Equal for constant-time comparison
	return hmac.Equal([]byte(expected), []byte(receivedSig))
}

// formatAmount converts float64 to string with 2 decimal places (no trailing zeros)
func formatAmount(amount float64) string {
	// Format to 2 decimal places, remove trailing zeros
	return fmt.Sprintf("%.2f", amount)
}
