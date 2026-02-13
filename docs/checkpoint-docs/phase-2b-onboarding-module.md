# Phase 2B: Onboarding Module

**Date:** February 13, 2025  
**Status:** ✅ Complete

## Overview

This phase implements the complete Onboarding vertical slice with event sourcing, callback processing, error classification, and integration with the Profile module. The module supports user-driven retry (no automatic system retries) and uses stub implementations for the Lazypay gateway.

## Files Created/Modified

### Onboarding Module - Entities
1. `internal/domain/onboarding/entity/onboarding_status.go` - Onboarding status enum with terminal state check
2. `internal/domain/onboarding/entity/onboarding_step.go` - Onboarding step enum and step status
3. `internal/domain/onboarding/entity/onboarding.go` - Updated with new entity types
4. `internal/domain/onboarding/entity/onboarding_event.go` - Updated with OnboardingStep type

### Onboarding Module - DTOs
5. `internal/domain/onboarding/dto/request/start_onboarding.go` - Start onboarding request with full KYC snapshot
6. `internal/domain/onboarding/dto/request/onboarding_callback.go` - Callback request DTO
7. `internal/domain/onboarding/dto/response/onboarding_response.go` - Start onboarding response
8. `internal/domain/onboarding/dto/response/onboarding_status.go` - Status response with step details

### Onboarding Module - Ports
9. `internal/domain/onboarding/port/gateway.go` - OnboardingGateway interface
10. `internal/domain/onboarding/port/repository.go` - Verified from Phase 1
11. `internal/domain/onboarding/port/event_store.go` - Verified from Phase 1

### Onboarding Module - Services
12. `internal/domain/onboarding/service/onboarding_service.go` - Main service with StartOnboarding, GetStatus, ProcessCallback
13. `internal/domain/onboarding/service/event_processor.go` - Event processing with status determination and retry calculation
14. `internal/domain/onboarding/service/error_config.go` - Error classification map with 10 error codes

### Onboarding Module - Handlers
15. `internal/domain/onboarding/handler/start_onboarding_handler.go` - POST /v1/payin3/onboarding
16. `internal/domain/onboarding/handler/onboarding_status_handler.go` - GET /v1/payin3/onboarding/status
17. `internal/domain/onboarding/handler/onboarding_callback_handler.go` - POST /v1/payin3/callback/onboarding

### Onboarding Module - Stubs
18. `internal/domain/onboarding/stub/stub_gateway.go` - Stub gateway returning fake redirect URLs

### Onboarding Module - Wiring
19. `internal/domain/onboarding/module.go` - Module wiring and route registration
20. `internal/domain/onboarding/repository/postgres_repository.go` - Updated mapping functions for new entity types
21. `internal/domain/onboarding/repository/postgres_event_store.go` - Updated mapping functions for OnboardingStep type

### Integration
22. `cmd/server/main.go` - Updated to register onboarding module routes

### Documentation
23. `docs/cursor-prompts/phase-2b-onboarding-module.md` - This checkpoint document

## API Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/v1/payin3/onboarding` | Start a new onboarding flow |
| GET | `/v1/payin3/onboarding/status` | Get onboarding status with step details |
| POST | `/v1/payin3/callback/onboarding` | Process onboarding callback events |

## Key Design Decisions

### 1. Event Sourcing
- All callback events are appended to `lender_onboarding_events` table
- Events are immutable and append-only
- Unique constraint on `(provider, mobile, onboarding_id, event_time, step)` ensures idempotency
- Event store is queried to build step-by-step status

### 2. Callback Idempotency
- Callbacks are idempotent via unique constraint in database
- Duplicate events are silently ignored (no error on conflict)
- Event is appended first, then projection is updated
- This ensures eventual consistency even if callback is retried

### 3. Error Classification
- 10 error codes classified with:
  - Retryability (true/false)
  - Target status (FAILED or INELIGIBLE)
  - HTTP status code
  - User-friendly message
  - Retry parameters (initial delay, max delay, max retries)
- Unknown errors default to non-retryable FAILED status

### 4. User-Driven Retry Philosophy
- No automatic system retries
- User must initiate retry by calling StartOnboarding again
- System calculates `nextRetryAt` for informational purposes only
- `isRetryable` flag indicates if retry is possible
- Retry count tracks number of retry attempts

### 5. Exponential Backoff with Jitter
- Retry delays calculated as: `initialDelay * 2^retryCount`
- Capped at `maxDelay` from error classification
- ±20% jitter added to prevent thundering herd
- Only calculated for retryable errors

### 6. Profile Integration
- On SUCCESS status, automatically calls `profileUpdater.UpdateOnOnboardingSuccess`
- Updates profile to ACTIVE status with credit limit
- Credit limit currently hardcoded to 50000 (TODO: get from gateway response)

### 7. Step Building from Events
- Steps are built by processing events chronologically
- All 6 steps initialized as PENDING
- Events update step status to SUCCESS or FAILED
- CompletedAt timestamp set from event time

## Verification Steps

### 1. Build Verification
```bash
go mod tidy
go build ./...
```

### 2. Run Server
```bash
go run cmd/server/main.go
```

### 3. Test Start Onboarding
```bash
curl -X POST http://localhost:8080/v1/payin3/onboarding \
  -H "Content-Type: application/json" \
  -d '{
    "onboardingTxnId": "txn-123",
    "userId": "u1",
    "merchantId": "m1",
    "channelId": "WEB",
    "source": "CHECKOUT",
    "returnUrl": "https://merchant.com/return",
    "userContact": {
      "mobile": "9876543210",
      "email": "user@example.com"
    },
    "kycSnapshot": {
      "pan": "ABCDE1234F",
      "fullLegalName": "John Doe",
      "bureauPullConsent": true
    }
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "onboardingId": "<uuid>",
    "onboardingTxnId": "txn-123",
    "provider": "LAZYPAY",
    "redirectUrl": "https://stub.lazypay.in/onboarding/...",
    "status": "PENDING"
  }
}
```

### 4. Test Get Status
```bash
curl "http://localhost:8080/v1/payin3/onboarding/status?userId=u1&onboardingId=<onboardingId>&merchantId=m1"
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "onboardingId": "<onboardingId>",
    "userId": "u1",
    "provider": "LAZYPAY",
    "status": "PENDING",
    "cofEligible": false,
    "steps": [
      {"step": "USER_DATA", "status": "PENDING"},
      {"step": "EMI_SELECTION", "status": "PENDING"},
      {"step": "KYC", "status": "PENDING"},
      {"step": "KFS", "status": "PENDING"},
      {"step": "MITC", "status": "PENDING"},
      {"step": "AUTOPAY", "status": "PENDING"}
    ],
    "retrying": false,
    "retryCount": 0,
    "updatedAt": "..."
  }
}
```

### 5. Test Callback
```bash
curl -X POST http://localhost:8080/v1/payin3/callback/onboarding \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "u1",
    "merchantId": "m1",
    "provider": "LAZYPAY",
    "onboardingId": "<onboardingId>",
    "mobile": "9876543210",
    "eventType": "STEP_COMPLETED",
    "status": "SUCCESS",
    "step": "KYC",
    "eventTime": "2025-02-13T10:00:00Z"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "accepted": true
  }
}
```

## Dependencies on Previous Phases

- **Phase 0**: Config, Postgres connection, middleware
- **Phase 1**: GORM models, repository interfaces, Postgres implementations, migrations
- **Phase 2A**: Shared kernel (response envelope, error codes), Profile module (ProfileUpdater)

## Error Codes Classified

1. `INVALID_MOBILE_FORMAT` - Non-retryable, FAILED, 400
2. `PAN_ALREADY_REGISTERED` - Non-retryable, INELIGIBLE, 422
3. `USER_INELIGIBLE` - Non-retryable, INELIGIBLE, 422
4. `BUREAU_TIMEOUT` - Retryable, FAILED, 500 (5s initial, 1h max, 5 retries)
5. `SERVICE_UNAVAILABLE` - Retryable, FAILED, 503 (5s initial, 1h max, 5 retries)
6. `KYC_FAILED` - Retryable, FAILED, 422 (30m initial, 48h max, 2 retries)
7. `INVALID_PAN_FORMAT` - Non-retryable, FAILED, 400
8. `KFS_FAILED` - Retryable, FAILED, 422 (30m initial, 48h max, 2 retries)
9. `MITC_FAILED` - Retryable, FAILED, 422 (30m initial, 48h max, 2 retries)
10. `PROVIDER_ERROR` - Retryable, FAILED, 500 (5s initial, 1h max, 5 retries)

## What Comes Next

### Phase 2C: Order Module
- Order creation with idempotency
- Payment state management
- Payment mapping
- Order status tracking
- Integration with onboarding and profile

### Future Enhancements
- Replace stub gateway with real Lazypay adapter
- Add GetLatestByUserAndMerchant to repository
- Get credit limit from gateway response
- Add comprehensive unit and integration tests
- Add event replay capability
- Add webhook signature verification

## Notes

- All service methods use `context.Context` for cancellation/timeout support
- Callback handler appends event first (idempotent), then updates projection
- Event processor determines status transitions based on event status and error codes
- Profile is automatically updated to ACTIVE on onboarding SUCCESS
- Repository mapping functions handle type conversions between entity types and DB strings
- Module follows dependency injection pattern
