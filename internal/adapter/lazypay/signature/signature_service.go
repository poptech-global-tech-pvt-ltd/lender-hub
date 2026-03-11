package signature

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

// Lazypay Signature Data Strings (from Postman collection):
//
// | Method              | Data String                                                        |
// |---------------------|--------------------------------------------------------------------|
// | SignOnboarding      | merchantAccessKey={ak}&mobile={mobile}                             |
// | SignEligibility     | {mobile}{amount}INR  OR  {mobile}{email}{amount}INR                |
// | SignCustomerStatus  | merchantAccessKey={ak}&mobile={mobile}                             |
// | SignOrder           | merchantAccessKey={ak}&transactionId={txnId}&amount={amt}          |
// | SignEnquiry         | merchantAccessKey={ak}&merchantTransactionId={txnId}               |
// | SignRefund          | merchantAccessKey={ak}&merchantTxnId={txnId}&amount={amt}          |
//
// IMPORTANT: Note the subtle key name differences:
//   Order uses    "&transactionId="
//   Refund uses   "&merchantTxnId="
//   Enquiry uses  "&merchantTransactionId="

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
// Postman: data = mobile + amount + "INR"  (without email)
// Postman: data = mobile + email + amount + "INR"  (with email)
// Email is conditionally included ONLY if non-empty
func (s *SignatureService) SignEligibility(mobile, email string, amount float64) string {
	var data string
	if email != "" {
		data = mobile + email + formatAmount(amount) + "INR"
	} else {
		data = mobile + formatAmount(amount) + "INR"
	}
	return s.hmacSHA1(data)
}

// SignCustomerStatus signs customer status request
// Postman: data = "merchantAccessKey=" + accessKey + "&mobile=" + mobile
func (s *SignatureService) SignCustomerStatus(mobile string) string {
	data := fmt.Sprintf("merchantAccessKey=%s&mobile=%s", s.accessKey, mobile)
	return s.hmacSHA1(data)
}

// SignOrder signs order creation request
// Postman: data = "merchantAccessKey=" + accessKey + "&transactionId=" + merchantTxnId + "&amount=" + amount
// Note: key is "transactionId" not "merchantTxnId"
// Note: amount is string like "1500.00", NO "INR" suffix
func (s *SignatureService) SignOrder(merchantTxnID string, amount float64) string {
	data := fmt.Sprintf(
		"merchantAccessKey=%s&transactionId=%s&amount=%s",
		s.accessKey, merchantTxnID, formatAmount(amount),
	)
	return s.hmacSHA1(data)
}

// SignEnquiry signs enquiry request
// Postman: data = "merchantAccessKey=" + accessKey + "&merchantTransactionId=" + merchantTxnId
// Note: key is "merchantTransactionId" (not "transactionId" or "merchantTxnId")
// Used for BOTH Order Enquiry and Refund Enquiry
// For Refund Enquiry: pass the ORIGINAL order merchantTxnId, not refundTxnId
func (s *SignatureService) SignEnquiry(merchantTxnID string) string {
	data := fmt.Sprintf(
		"merchantAccessKey=%s&merchantTransactionId=%s",
		s.accessKey, merchantTxnID,
	)
	return s.hmacSHA1(data)
}

// SignOnboarding signs onboarding request
// Postman: data = "merchantAccessKey=" + accessKey + "&mobile=" + mobile
// Same formula as SignCustomerStatus
// Email is NEVER in onboarding signature
func (s *SignatureService) SignOnboarding(mobile string) string {
	data := fmt.Sprintf("merchantAccessKey=%s&mobile=%s", s.accessKey, mobile)
	return s.hmacSHA1(data)
}

// SignRefund signs refund request
// Postman: data = "merchantAccessKey=" + accessKey + "&merchantTxnId=" + merchantTxnId + "&amount=" + amount
// Note: key is "merchantTxnId" (NOT "transactionId" like Order)
// refundTxnId is NOT in signature
// Email is NOT in signature
func (s *SignatureService) SignRefund(merchantTxnID string, amount float64) string {
	data := fmt.Sprintf(
		"merchantAccessKey=%s&merchantTxnId=%s&amount=%s",
		s.accessKey, merchantTxnID, formatAmount(amount),
	)
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
