# Phase 3C: Lazypay Clients + Adapter Wiring

**Date:** February 13, 2025  
**Status:** ✅ Complete

## Overview

This phase implements real Lazypay HTTP clients that replace the stub gateways from Phase 2. All clients implement their respective domain gateway interfaces and use the infrastructure components from Phase 3A (executors, circuit breaker, retry, signature service) and mappers from Phase 3B.

## Files Created/Modified (7 total)

### Lazypay Clients (5 files)
1. `internal/adapter/lazypay/client/lazypay_client.go` - Main client aggregator
2. `internal/adapter/lazypay/client/profile_client.go` - Profile gateway implementation
3. `internal/adapter/lazypay/client/onboarding_client.go` - Onboarding gateway implementation
4. `internal/adapter/lazypay/client/payment_client.go` - Order gateway implementation
5. `internal/adapter/lazypay/client/refund_client.go` - Refund gateway implementation

### Adapter Wiring (1 file)
6. `internal/adapter/lazypay/adapter.go` - Adapter factory function

### Main Application (1 file)
7. `cmd/server/main.go` - Updated with adapter selection logic

### Constants Package (1 file)
8. `internal/adapter/lazypay/constants/constants.go` - Moved from lazypay package to break import cycle

## Client → Gateway Mapping

| Client | Gateway Interface | Domain Module |
|--------|------------------|---------------|
| `ProfileClient` | `profile.ProfileGateway` | Profile |
| `OnboardingClient` | `onboarding.OnboardingGateway` | Onboarding |
| `PaymentClient` | `order.OrderGateway` | Order |
| `RefundClient` | `refund.RefundGateway` | Refund |

## Client → Executor Mapping

| Client | Executor | Timeout | MaxIdleConns | MaxIdleConnsPerHost |
|--------|----------|---------|--------------|---------------------|
| `ProfileClient` | `profileExec` | 10s | 50 | 10 |
| `OnboardingClient` | `profileExec` | 10s | 50 | 10 |
| `PaymentClient` | `paymentExec` | 5s | 100 | 20 |
| `RefundClient` | `paymentExec` | 5s | 100 | 20 |

## Request Flow

```
Handler
  ↓
Service
  ↓
Gateway (interface)
  ↓
Client (implementation)
  ↓
Mapper (ToLP*)
  ↓
Signature Service
  ↓
Executor (with circuit breaker + retry)
  ↓
Lazypay API
  ↓
Executor Response
  ↓
Mapper (FromLP*)
  ↓
Error Mapper (if error)
  ↓
DomainError
  ↓
Service
  ↓
Handler
```

## Adapter Selection Logic

The application automatically selects between real and stub adapters based on configuration:

```go
if cfg.Lazypay.BaseURL != "" && cfg.Lazypay.AccessKey != "" {
    // Real Lazypay adapter
    lpClient := lazypay.NewAdapter(lpCfg)
    profileGW = lpClient.ProfileGateway()
    // ... other gateways
} else {
    // Stub gateways for local dev
    profileGW = profileStub.NewStubProfileGateway()
    // ... other stubs
}
```

### Configuration Requirements

**Real Adapter:**
- `lazypay.base_url` (required)
- `lazypay.access_key` (required)
- `lazypay.secret_key` (required)

**Stub Adapter:**
- No Lazypay config needed
- All gateways use in-memory stubs

## Client Implementation Details

### ProfileClient

**Methods:**
- `CheckEligibility(ctx, req)` → POST `/v7/payment/eligibility`
  - Signs with `SignEligibility(mobile, email, 0)`
  - Maps to `LPEligibilityRequest`
  - Returns `CustomerStatusResponse`

- `GetCustomerStatus(ctx, mobile)` → GET `/v7/payment/customerStatus`
  - Signs with `SignCustomerStatus(mobile)`
  - Returns `CustomerStatusResponse`

**Executor:** `profileExec` (10s timeout, profile bulkhead)

### OnboardingClient

**Methods:**
- `StartOnboarding(ctx, req)` → POST `/v7/createStandaloneOnboarding`
  - Maps to `LPOnboardingRequest`
  - Returns `OnboardingResponse`

- `GetOnboardingStatus(ctx, mobile)` → GET `/v7/onboarding/status`
  - Returns `OnboardingStatusResponse`

**Executor:** `profileExec` (10s timeout, profile bulkhead)

### PaymentClient

**Methods:**
- `CreateOrder(ctx, req)` → POST `/cof/v0/payment/order`
  - Signs with `SignOrder(paymentId, amount)`
  - Maps to `LPCreateOrderRequest`
  - Returns `OrderResponse`

- `GetOrderStatus(ctx, paymentID)` → GET `/cof/v0/payment/enquiry`
  - Returns `OrderStatusResponse`

**Executor:** `paymentExec` (5s timeout, payment bulkhead)

### RefundClient

**Methods:**
- `ProcessRefund(ctx, req)` → POST `/v7/refund`
  - Signs with `SignOrder(paymentId, amount)` (reuses order signature format)
  - Maps to `LPRefundRequest`
  - Returns `RefundResponse`

**Executor:** `paymentExec` (5s timeout, payment bulkhead)

## Error Handling

All clients follow the same error handling pattern:

1. **HTTP Status >= 400**: Parse error response body
2. **Extract Error Code**: Look for `errorCode` field
3. **Map Error**: Use `mapper.MapLPError(lpErrorCode)`
4. **Return DomainError**: Returns canonical error with retryability

```go
if resp.StatusCode >= 400 {
    var lpError struct {
        ErrorCode    string `json:"errorCode"`
        ErrorMessage string `json:"errorMessage"`
    }
    if err := json.Unmarshal(resp.Body, &lpError); err == nil && lpError.ErrorCode != "" {
        return nil, mapper.MapLPError(lpError.ErrorCode)
    }
    return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "provider error")
}
```

## Verification Steps

### 1. Build Verification
```bash
go mod tidy
go build ./...
```

### 2. Start Server WITHOUT Lazypay Config
```bash
# No lazypay config in config.yaml
go run cmd/server/main.go
```

**Expected Output:**
```
Using stub gateways (no Lazypay config)
Server listening on :8080
```

**Verification:**
- All endpoints work with stub responses
- No HTTP calls to Lazypay
- Circuit breaker not initialized

### 3. Start Server WITH Lazypay Config
```yaml
# config.yaml
lazypay:
  base_url: "https://sandbox.lazypay.in"
  access_key: "your-access-key"
  secret_key: "your-secret-key"
```

```bash
go run cmd/server/main.go
```

**Expected Output:**
```
Using Lazypay adapter
Server listening on :8080
```

**Verification:**
- Real adapter initialized
- Circuit breakers active
- HTTP executors configured
- Actual calls to Lazypay (may fail if credentials invalid, but no compile errors)

### 4. Verify No Circular Imports
```bash
go build ./...
# Should compile without import cycle errors
```

## Key Design Decisions

### 1. Import Cycle Resolution
- **Problem**: `lazypay` package imports `lazypay/client`, but `lazypay/client` needs constants from `lazypay`
- **Solution**: Moved constants to separate `lazypay/constants` package
- **Result**: No circular dependencies

### 2. Bulkhead Isolation
- **Profile Operations**: Use `profileExec` (slower, more tolerant)
- **Payment Operations**: Use `paymentExec` (faster, stricter)
- **Rationale**: Different SLA requirements

### 3. Adapter Selection
- **Environment-Based**: Automatic based on config presence
- **No Code Changes**: Same code path, different implementations
- **Fallback**: Stubs always available for local dev

### 4. Error Mapping
- **Centralized**: All clients use `mapper.MapLPError`
- **Consistent**: Same LP error code → same canonical error
- **Retryable**: Errors marked appropriately

### 5. Signature Service
- **Per-API Formats**: Different signature formats for different endpoints
- **Reused**: Same signature service instance across all clients
- **Secure**: HMAC-SHA1 with constant-time comparison

## Dependencies on Previous Phases

- **Phase 3A**: HTTP executors, circuit breaker, retry, signature service
- **Phase 3B**: DTOs, mappers, error mapper
- **Phase 2A-2D**: Domain modules with gateway interfaces

## What Comes Next

### Phase 4: Production Infrastructure
- Kafka producer for event publishing
- Redis cache for profile eligibility
- Observability (metrics, tracing, logging)
- Webhook signature verification
- Rate limiting
- API documentation (OpenAPI/Swagger)

## Notes

- All clients implement domain gateway interfaces (dependency inversion)
- Same executor instances reused across clients (bulkhead isolation)
- Circuit breakers track state per executor
- Retry logic handles transient errors automatically
- Signature service shared across all clients
- Error responses parsed and mapped to canonical errors
- Adapter selection is transparent to domain modules
