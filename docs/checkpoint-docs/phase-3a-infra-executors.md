# Phase 3A: HTTP Executors + Circuit Breaker + Signature Service

**Date:** February 13, 2025  
**Status:** ✅ Complete

## Overview

This phase implements the infrastructure layer for Lazypay HTTP communication: circuit breakers, retry logic, HTTP executors with bulkhead isolation, and HMAC-SHA1 signature service. These components will be used by the Lazypay adapter in Phase 3B/3C.

## Files Created

### Circuit Breaker
1. `internal/infrastructure/http/circuitbreaker/config.go` - Circuit breaker configuration
2. `internal/infrastructure/http/circuitbreaker/breaker.go` - Circuit breaker implementation with state machine

### Retry Logic
3. `internal/infrastructure/http/retry/retry.go` - Retry with exponential backoff and jitter

### HTTP Executors
4. `internal/infrastructure/http/executor/executor.go` - HttpExecutor interface
5. `internal/infrastructure/http/executor/profile_executor.go` - Profile bulkhead executor
6. `internal/infrastructure/http/executor/payment_executor.go` - Payment bulkhead executor

### Lazypay Infrastructure
7. `internal/adapter/lazypay/config/lazypay_config.go` - Lazypay configuration struct
8. `internal/adapter/lazypay/signature/signature_service.go` - HMAC-SHA1 signature service
9. `internal/adapter/lazypay/signature/signature_service_test.go` - Signature service tests
10. `internal/adapter/lazypay/constants.go` - API paths, error codes, headers

## Bulkhead Configuration Summary

| Aspect | Profile Executor | Payment Executor |
|--------|-----------------|------------------|
| **Timeout** | 10s | 5s |
| **MaxIdleConns** | 50 | 100 |
| **MaxIdleConnsPerHost** | 10 | 20 |
| **Circuit Breaker Threshold** | 5 failures | 10 failures |
| **Circuit Breaker Timeout** | 30s | 15s |
| **Retry MaxAttempts** | 3 | 2 |
| **Retry InitialDelay** | 500ms | 200ms |
| **Retry MaxDelay** | 5s | 2s |

## Circuit Breaker State Machine

```
┌─────────┐
│ CLOSED  │ ← Normal operation, requests allowed
└────┬────┘
     │
     │ failureCount >= threshold
     ↓
┌─────────┐
│  OPEN   │ ← Rejecting requests fast
└────┬────┘
     │
     │ timeout elapsed
     ↓
┌──────────┐
│HALF_OPEN │ ← Probing recovery (limited requests)
└────┬─────┘
     │
     ├─ Success → CLOSED
     └─ Failure → OPEN
```

### State Transitions

- **CLOSED → OPEN**: When failureCount >= FailureThreshold
- **OPEN → HALF_OPEN**: After Timeout duration has elapsed
- **HALF_OPEN → CLOSED**: On first success
- **HALF_OPEN → OPEN**: On any failure

### Thread Safety

- Uses `sync.Mutex` for all state changes
- All public methods are thread-safe
- State queries are protected by mutex

## Retry Logic

### Retryable Status Codes
- `0` (network error)
- `502` (Bad Gateway)
- `503` (Service Unavailable)
- `504` (Gateway Timeout)
- `429` (Too Many Requests)

### Exponential Backoff Formula
```
delay = initialDelay * 2^attempt
delay = min(delay, maxDelay)
finalDelay = delay + (delay * jitterFactor * random(0-1))
```

### Retry Flow
1. Execute request
2. If success or non-retryable → return immediately
3. If retryable → wait for calculated delay
4. Retry up to MaxAttempts
5. Return last result/error

## HMAC Signature Formats

### 1. Eligibility Request
```
data = mobile + email + formatAmount(amount) + "INR"
signature = HMAC-SHA1(secretKey, data)
```

### 2. Customer Status Request
```
data = accessKey + mobile
signature = HMAC-SHA1(secretKey, data)
```

### 3. Order Creation Request
```
data = accessKey + merchantTxnId + formatAmount(amount) + "INR"
signature = HMAC-SHA1(secretKey, data)
```

### 4. Webhook Verification
```
signature = HMAC-SHA1(webhookSecret, payload)
verify = hmac.Equal(receivedSignature, computedSignature)
```

### Amount Formatting
- Uses `fmt.Sprintf("%.2f", amount)`
- Always 2 decimal places (e.g., "1000.00", "500.50")

## Verification Steps

### 1. Build Verification
```bash
go mod tidy
go build ./...
```

### 2. Run Signature Tests
```bash
go test ./internal/adapter/lazypay/signature/... -v
```

**Expected Output:**
```
=== RUN   TestSignEligibility
--- PASS: TestSignEligibility (0.00s)
=== RUN   TestSignCustomerStatus
--- PASS: TestSignCustomerStatus (0.00s)
=== RUN   TestSignOrder
--- PASS: TestSignOrder (0.00s)
=== RUN   TestVerifyWebhook
--- PASS: TestVerifyWebhook (0.00s)
=== RUN   TestVerifyWebhook_TimingSafe
--- PASS: TestVerifyWebhook_TimingSafe (0.00s)
=== RUN   TestFormatAmount
--- PASS: TestFormatAmount (0.00s)
PASS
```

### 3. Verify Circuit Breaker Thread Safety
```bash
go test -race ./internal/infrastructure/http/circuitbreaker/...
```

## Key Design Decisions

### 1. Bulkhead Isolation
- **Profile Executor**: Higher timeout (10s), more retries (3), lower failure threshold (5)
  - Used for eligibility checks, customer status
  - More tolerant of slow responses
- **Payment Executor**: Lower timeout (5s), fewer retries (2), higher failure threshold (10)
  - Used for order creation, refunds
  - Faster failure detection, less retry overhead

### 2. Circuit Breaker Implementation
- **Thread-Safe**: All operations protected by `sync.Mutex`
- **State Machine**: Clear transitions with timeout-based recovery
- **Half-Open Probing**: Limited requests to test recovery
- **Metrics-Ready**: `GetState()` method for observability

### 3. Retry Strategy
- **Exponential Backoff**: 2^attempt multiplier
- **Jitter**: ±10% random variation to prevent thundering herd
- **Context-Aware**: Respects context cancellation
- **Selective**: Only retries transient errors (5xx, timeouts)

### 4. Signature Service
- **HMAC-SHA1**: Lazypay's required algorithm
- **Constant-Time Comparison**: Uses `hmac.Equal` for webhook verification
- **Per-API Formats**: Different signature formats for different endpoints
- **Amount Formatting**: Consistent 2-decimal format

### 5. HTTP Client Configuration
- **Connection Pooling**: MaxIdleConns and MaxIdleConnsPerHost
- **Timeouts**: Per-executor timeout configuration
- **Transport Reuse**: Single transport per executor instance

## Dependencies on Previous Phases

- **Phase 0**: Config structure (for Lazypay config)
- **Phase 1-2D**: Domain modules (no direct dependency, but will be used by adapters)

## What Comes Next

### Phase 3B: Lazypay DTOs + Mappers
- Request/response DTOs matching Lazypay API contracts
- Mappers between domain DTOs and Lazypay DTOs
- Error code mapping from Lazypay to domain errors

### Phase 3C: Lazypay HTTP Clients
- Profile client (eligibility, customer status)
- Onboarding client (create, status)
- Order client (create, enquiry)
- Refund client (process)
- Replace all stub gateways with real implementations

## Notes

- All public functions take `context.Context` as first argument
- Circuit breaker is goroutine-safe using `sync.Mutex`
- Retry logic respects context cancellation
- Signature service uses constant-time comparison for security
- HTTP executors are isolated by bulkhead (Profile vs Payment)
- Executors integrate circuit breaker and retry automatically
- Response includes timing metadata (Duration field)
