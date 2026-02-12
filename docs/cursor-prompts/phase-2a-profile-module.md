# Phase 2A: Shared Kernel + Profile Module

**Date:** February 13, 2025  
**Status:** ✅ Complete

## Overview

This phase implements the shared kernel components (error handling, response envelope) and completes the Profile vertical slice with full DDD architecture including entities, services, handlers, and stub implementations.

## Files Created/Modified

### Shared Kernel
1. `internal/shared/response/envelope.go` - Standard API response wrapper
2. `internal/shared/errors/codes.go` - Canonical error codes
3. `internal/shared/errors/error.go` - DomainError type with HTTP status mapping
4. `internal/shared/middleware/request_id.go` - Updated with GetRequestID helper

### Profile Module - Entities
5. `internal/domain/profile/entity/profile_status.go` - Profile status enum with state transition validation
6. `internal/domain/profile/entity/credit_line.go` - Credit line value object
7. `internal/domain/profile/entity/block_info.go` - Block information value object
8. `internal/domain/profile/entity/user_profile.go` - Updated aggregate root with new structure

### Profile Module - DTOs
9. `internal/domain/profile/dto/request/customer_status.go` - Request DTO
10. `internal/domain/profile/dto/response/customer_status.go` - Response DTO with EMI plans

### Profile Module - Ports
11. `internal/domain/profile/port/gateway.go` - ProfileGateway interface
12. `internal/domain/profile/port/cache.go` - ProfileCache interface
13. `internal/domain/profile/port/event_publisher.go` - ProfileEventPublisher interface

### Profile Module - Services
14. `internal/domain/profile/service/profile_service.go` - ProfileService with cache-first strategy
15. `internal/domain/profile/service/profile_updater.go` - ProfileUpdater for state transitions

### Profile Module - Handler
16. `internal/domain/profile/handler/customer_status_handler.go` - HTTP handler with validation

### Profile Module - Stubs
17. `internal/domain/profile/stub/stub_gateway.go` - Stub gateway returning hardcoded ACTIVE response
18. `internal/domain/profile/stub/stub_cache.go` - In-memory cache using sync.Map
19. `internal/domain/profile/stub/stub_event_publisher.go` - No-op event publisher

### Profile Module - Wiring
20. `internal/domain/profile/module.go` - Module wiring and route registration
21. `internal/domain/profile/repository/postgres_repository.go` - Updated mapping functions for new entity structure
22. `cmd/server/main.go` - Updated to register profile module routes

### Documentation
23. `docs/cursor-prompts/phase-2a-profile-module.md` - This checkpoint document

## API Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/v1/payin3/customer-status` | Check customer eligibility and status |

## Key Design Decisions

### 1. State Machine
- Profile status transitions are validated using `CanTransitionTo()` method
- Allowed transitions:
  - NOT_STARTED → IN_PROGRESS, ACTIVE, INELIGIBLE
  - IN_PROGRESS → ACTIVE, INELIGIBLE, BLOCKED
  - ACTIVE → BLOCKED, INELIGIBLE
  - BLOCKED → ACTIVE, INELIGIBLE
  - INELIGIBLE → IN_PROGRESS (retry)

### 2. Cache-First Strategy
- ProfileService checks cache before calling gateway
- Cache key format: `userID:lender`
- Stub cache uses in-memory sync.Map (production would use Redis)

### 3. Stub Gateway
- Returns hardcoded ACTIVE status with 50,000 INR limit
- Includes 2 sample EMI plans (3 and 6 months)
- No external API calls in development mode

### 4. FOR UPDATE Locking
- `GetForUpdate` uses PostgreSQL row-level locking
- Prevents race conditions during state transitions
- Used in ProfileUpdater for all update operations

### 5. Event Publishing
- Profile events published for: Activation, Block, Unblock, Limit Update, Status Change
- Stub publisher is no-op (production would use Kafka)
- Events include full context (previous/new status, limits, block info)

### 6. Error Handling
- DomainError type with HTTP status mapping
- Standardized error codes (e.g., PAYIN3_USER_BLOCKED)
- Response envelope wraps all API responses

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

### 3. Test Endpoint
```bash
curl -X POST http://localhost:8080/v1/payin3/customer-status \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "u1",
    "mobile": "9876543210",
    "merchantId": "m1",
    "source": "CHECKOUT"
  }'
```

### Expected Response
```json
{
  "success": true,
  "data": {
    "userId": "u1",
    "provider": "LAZYPAY",
    "preApproved": true,
    "availableLimit": 50000,
    "creditLineActive": true,
    "onboardingRequired": false,
    "status": "ACTIVE",
    "reasonCode": "",
    "reasonMessage": "",
    "emiPlans": [
      {
        "tenureMonths": 3,
        "emiAmount": 16666.67,
        "totalAmount": 50000
      },
      {
        "tenureMonths": 6,
        "emiAmount": 8333.33,
        "totalAmount": 50000
      }
    ]
  }
}
```

## Dependencies on Previous Phases

- **Phase 0**: Config, Postgres connection, middleware
- **Phase 1**: GORM models, repository interfaces, Postgres implementations
- **Phase 1b**: Database migrations (enums, tables, indexes)

## What Comes Next

### Phase 2B: Onboarding Module
- Onboarding initiation endpoint
- Step-by-step onboarding flow
- Event sourcing for onboarding events
- Integration with Lazypay gateway (stub)
- Onboarding status tracking

### Future Enhancements
- Replace stub gateway with real Lazypay adapter
- Replace stub cache with Redis implementation
- Replace stub event publisher with Kafka producer
- Add rate limiting middleware
- Add authentication/authorization
- Add comprehensive unit and integration tests

## Notes

- All service methods use `context.Context` for cancellation/timeout support
- Handler validates input using Gin's binding
- Domain errors are properly mapped to HTTP status codes
- Repository uses proper entity-to-model mapping
- Module follows dependency injection pattern
