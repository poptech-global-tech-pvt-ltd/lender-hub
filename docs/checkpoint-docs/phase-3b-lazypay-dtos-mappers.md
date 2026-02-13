# Phase 3B: Lazypay DTOs + Mappers

**Date:** February 13, 2025  
**Status:** ✅ Complete

## Overview

This phase implements Lazypay-specific request/response DTOs and bidirectional mappers between canonical domain DTOs and Lazypay API formats. All DTOs match Lazypay's exact JSON shape (camelCase field names), and mappers are pure functions with no side effects.

## Files Created (19 total)

### Common DTOs (3 files)
1. `internal/adapter/lazypay/dto/common/user_details.go` - LPUserDetails
2. `internal/adapter/lazypay/dto/common/amount.go` - LPAmount
3. `internal/adapter/lazypay/dto/common/address.go` - LPAddress

### Request DTOs (4 files)
4. `internal/adapter/lazypay/dto/request/eligibility_request.go` - LPEligibilityRequest
5. `internal/adapter/lazypay/dto/request/onboarding_request.go` - LPOnboardingRequest
6. `internal/adapter/lazypay/dto/request/create_order_request.go` - LPCreateOrderRequest + LPProductLine
7. `internal/adapter/lazypay/dto/request/refund_request.go` - LPRefundRequest

### Response DTOs (4 files)
8. `internal/adapter/lazypay/dto/response/eligibility_response.go` - LPEligibilityResponse + LPEMIPlan
9. `internal/adapter/lazypay/dto/response/onboarding_response.go` - LPOnboardingResponse
10. `internal/adapter/lazypay/dto/response/order_response.go` - LPOrderResponse
11. `internal/adapter/lazypay/dto/response/refund_response.go` - LPRefundResponse

### Webhook DTOs (3 files)
12. `internal/adapter/lazypay/dto/webhook/onboarding_webhook.go` - LPOnboardingWebhook
13. `internal/adapter/lazypay/dto/webhook/order_webhook.go` - LPOrderWebhook
14. `internal/adapter/lazypay/dto/webhook/refund_webhook.go` - LPRefundWebhook

### Mappers (5 files)
15. `internal/adapter/lazypay/mapper/profile_mapper.go` - Profile ↔ Lazypay
16. `internal/adapter/lazypay/mapper/onboarding_mapper.go` - Onboarding ↔ Lazypay
17. `internal/adapter/lazypay/mapper/order_mapper.go` - Order ↔ Lazypay
18. `internal/adapter/lazypay/mapper/refund_mapper.go` - Refund ↔ Lazypay
19. `internal/adapter/lazypay/mapper/error_mapper.go` - LP error codes → canonical errors

## DTO Naming Convention

- **LP Prefix**: All Lazypay-specific DTOs use `LP` prefix (e.g., `LPUserDetails`, `LPEligibilityRequest`)
- **camelCase Fields**: All JSON field names match Lazypay's API exactly (e.g., `merchantId`, `accessKey`, `orderAmount`)
- **Common Types**: Shared structures (user, amount, address) in `dto/common` package

## Mapper Functions

### Direction Convention
- **To***: Canonical domain DTO → Lazypay DTO (e.g., `ToLPEligibilityRequest`)
- **From***: Lazypay DTO → Canonical domain DTO (e.g., `FromLPEligibilityResponse`)

### Profile Mapper
- `ToLPEligibilityRequest(canonical, accessKey, merchantID, signature)` → `*LPEligibilityRequest`
- `FromLPEligibilityResponse(lp, userID)` → `*CustomerStatusResponse`
  - Maps LP status → canonical PayIn3Status
  - Maps LP EMI plans → canonical EmiPlan
  - Sets `onboardingRequired = !creditLineActive`
  - Sets `provider = "LAZYPAY"`

### Onboarding Mapper
- `ToLPOnboardingRequest(canonical, accessKey, merchantID)` → `*LPOnboardingRequest`
  - Maps address, KYC fields, employment details
- `FromLPOnboardingResponse(lp, onboardingTxnID)` → `*OnboardingResponse`
  - Sets `provider = "LAZYPAY"`

### Order Mapper
- `ToLPCreateOrderRequest(canonical, accessKey, merchantID, signature)` → `*LPCreateOrderRequest`
  - Maps product lines with price formatting
  - Maps address if provided
- `FromLPOrderResponse(lp, paymentID)` → `*OrderResponse`
  - Converts string fields to pointers for optional fields

### Refund Mapper
- `ToLPRefundRequest(canonical, paymentID, accessKey, merchantID, signature)` → `*LPRefundRequest`
- `FromLPRefundResponse(lp, refundID, paymentID)` → `*RefundResponse`
  - Sets `provider = "LAZYPAY"`

## Error Mapping Table

| Lazypay Error Code | Canonical Code | HTTP Status | Retryable | Description |
|-------------------|----------------|-------------|-----------|-------------|
| `LP_USER_BLOCKED` | `PAYIN3_USER_BLOCKED` | 422 | No | User is blocked |
| `COF_INSUFFICIENT_BALANCE` | `PAYIN3_INSUFFICIENT_LIMIT` | 422 | No | Insufficient credit limit |
| `INVALID_MOBILE_FORMAT` | `INVALID_REQUEST` | 400 | No | Invalid mobile format |
| `INVALID_PAN_FORMAT` | `INVALID_REQUEST` | 400 | No | Invalid PAN format |
| `PAN_ALREADY_REGISTERED` | `IDENTITY_DOCUMENT_ALREADY_REGISTERED` | 422 | No | PAN already registered |
| `USER_INELIGIBLE` | `USER_INELIGIBLE` | 422 | No | User ineligible |
| `BUREAU_TIMEOUT` | `CREDIT_BUREAU_TIMEOUT` | 500 | Yes | Bureau timeout |
| `SERVICE_UNAVAILABLE` | `SERVICE_TEMPORARILY_UNAVAILABLE` | 503 | Yes | Service unavailable |
| `INTERNAL_ERROR` | `INTERNAL_PROVIDER_ERROR` | 500 | Yes | Provider internal error |
| `RATE_LIMIT_EXCEEDED` | `RATE_LIMIT_EXCEEDED` | 429 | Yes | Rate limit exceeded |
| `KYC_FAILED` | `KYC_VERIFICATION_FAILED` | 422 | Yes | KYC verification failed |
| `PAN_VERIFICATION_LIMIT_EXHAUSTED` | `VERIFICATION_ATTEMPT_LIMIT_EXHAUSTED` | 422 | Yes | Verification limit exhausted |

### Error Mapper Function
- `MapLPError(lpErrorCode)` → `*DomainError`
  - Returns mapped error if found in `LPErrorMapping`
  - Returns generic `INTERNAL_ERROR` for unknown codes

## Webhook DTO Shapes

### Onboarding Webhook
```json
{
  "onboardingId": "string",
  "mobile": "string",
  "eventType": "string",
  "status": "string",
  "step": "string?",
  "errorCode": "string?",
  "message": "string?",
  "eventTime": "string"
}
```

### Order Webhook
```json
{
  "merchantTxnId": "string",
  "orderId": "string",
  "status": "string",
  "txnId": "string?",
  "errorCode": "string?",
  "errorMessage": "string?",
  "eventTime": "string"
}
```

### Refund Webhook
```json
{
  "refundTxnId": "string",
  "merchantTxnId": "string",
  "status": "string",
  "lenderRefId": "string?",
  "message": "string?",
  "eventTime": "string"
}
```

## Key Design Decisions

### 1. Pure Functions
- All mapper functions are pure (no side effects)
- No database or HTTP calls
- Deterministic transformations only

### 2. Bidirectional Mapping
- Separate functions for each direction
- Clear naming: `To*` (canonical→LP), `From*` (LP→canonical)
- Handles optional fields and type conversions

### 3. Amount Formatting
- Uses `fmt.Sprintf("%.2f", amount)` for consistent 2-decimal formatting
- Matches Lazypay's expected string format

### 4. Status Mapping
- Profile: LP status → canonical PayIn3Status enum
- Handles blocked state separately
- Maps `APPROVED` + `creditLineActive` → `ACTIVE` or `IN_PROGRESS`

### 5. Pointer Fields
- Optional fields use pointers (`*string`, `*common.LPAddress`)
- Properly handles nil values in mappings

### 6. Error Mapping
- Centralized error code mapping
- Retryable errors marked with `NewRetryable`
- Unknown errors fall back to generic `INTERNAL_ERROR`

## Verification Steps

### 1. Build Verification
```bash
go mod tidy
go build ./...
```

### 2. Import Verification
- All imports resolve correctly
- No circular dependencies
- Canonical DTOs ↔ Adapter DTOs properly separated

### 3. Type Safety
- All mapper functions compile
- Type conversions are explicit
- Pointer handling is correct

## Dependencies on Previous Phases

- **Phase 3A**: Signature service (used for signing requests in Phase 3C)
- **Phase 2A-2D**: Domain DTOs (canonical request/response types)

## What Comes Next

### Phase 3C: Lazypay HTTP Clients
- Profile client (eligibility, customer status)
- Onboarding client (create, status)
- Order client (create, enquiry)
- Refund client (process)
- Webhook handlers (onboarding, order, refund)
- Replace all stub gateways with real implementations
- Wire into domain modules

## Notes

- All DTO field names match Lazypay API exactly (camelCase)
- Mappers handle optional fields gracefully
- Error mapper provides fallback for unknown codes
- Webhook DTOs support optional error fields
- Amount values formatted as strings with 2 decimals
- Provider field set to "LAZYPAY" in all responses
