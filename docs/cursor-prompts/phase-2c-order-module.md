# Phase 2C: Order Module

**Date:** February 13, 2025  
**Status:** ✅ Complete

## Overview

This phase implements the complete Order vertical slice with DB-enforced idempotency, payment state management, and integration with the Profile module for limit management. The module uses SHA256 request hashing for idempotency validation and FOR UPDATE locking for safe state transitions.

## Files Created/Modified

### Order Module - Entities
1. `internal/domain/order/entity/order_status.go` - Order status enum with terminal state check
2. `internal/domain/order/entity/emi_selection.go` - EMI selection value object
3. `internal/domain/order/entity/order.go` - Updated with OrderStatus type
4. `internal/domain/order/entity/payment_mapping.go` - Payment mapping entity (already existed)
5. `internal/domain/order/entity/idempotency_key.go` - Updated with Key and IdempotencyStatus types

### Order Module - DTOs
6. `internal/domain/order/dto/request/create_order.go` - Create order request with EMI selection
7. `internal/domain/order/dto/request/order_callback.go` - Order callback request
8. `internal/domain/order/dto/response/order_response.go` - Order creation response
9. `internal/domain/order/dto/response/order_status.go` - Order status response

### Order Module - Ports
10. `internal/domain/order/port/gateway.go` - OrderGateway interface
11. `internal/domain/order/port/event_publisher.go` - OrderEventPublisher interface
12. `internal/domain/order/port/repository.go` - Verified from Phase 1
13. `internal/domain/order/port/idempotency_repository.go` - Verified from Phase 1

### Order Module - Services
14. `internal/domain/order/service/idempotency.go` - IdempotencyService with hash computation
15. `internal/domain/order/service/order_service.go` - OrderService with CreateOrder, GetStatus, ProcessCallback

### Order Module - Handlers
16. `internal/domain/order/handler/create_order_handler.go` - POST /v1/payin3/order
17. `internal/domain/order/handler/get_order_handler.go` - GET /v1/payin3/order/:paymentId
18. `internal/domain/order/handler/order_callback_handler.go` - POST /v1/payin3/callback/order

### Order Module - Stubs
19. `internal/domain/order/stub/stub_gateway.go` - Stub gateway returning fake order responses
20. `internal/domain/order/stub/stub_event_publisher.go` - No-op event publisher

### Order Module - Wiring
21. `internal/domain/order/module.go` - Module wiring and route registration
22. `internal/domain/order/repository/postgres_repository.go` - Updated mapping for OrderStatus
23. `internal/domain/order/repository/idempotency_repository.go` - Updated mapping for IdempotencyStatus and Key field

### Integration
24. `cmd/server/main.go` - Updated to register order module routes

### Documentation
25. `docs/cursor-prompts/phase-2c-order-module.md` - This checkpoint document

## API Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/v1/payin3/order` | Create a new order with idempotency |
| GET | `/v1/payin3/order/:paymentId` | Get order status |
| POST | `/v1/payin3/callback/order` | Process order callback events |

## Key Design Decisions

### 1. DB-Enforced Idempotency
- `paymentId` is the idempotency key
- Uses `INSERT ON CONFLICT` pattern via `TryAcquire` method
- Unique constraint on `idempotency_key` column ensures atomicity
- Request hash (SHA256) validates request body matches

### 2. Idempotency Flow

```
1. ComputeHash(request) → SHA256 hash
2. TryAcquire(paymentId, hash)
   ├─ NEW: Create PROCESSING record → proceed
   ├─ DUPLICATE: Hash matches + COMPLETED → return cached response
   ├─ CONFLICT: Hash matches + PROCESSING → return 409
   └─ MISMATCH: Hash differs → return 422
3. Create order + mapping
4. MarkCompleted(paymentId, response, lenderOrderID)
   └─ On failure: MarkFailed(paymentId)
```

### 3. SHA256 Request Hashing
- Entire request body is JSON-marshaled and hashed
- Prevents same `paymentId` with different request body
- Hash mismatch returns 422 (IDEMPOTENCY_HASH_MISMATCH)
- Cached responses stored in `response_payload` JSONB field

### 4. FOR UPDATE Locking
- `GetForUpdate` uses PostgreSQL row-level locking
- Prevents race conditions during callback processing
- Used in `ProcessCallback` for safe state updates
- Terminal states are idempotent (ignore duplicate callbacks)

### 5. Profile Limit Management
- On SUCCESS callback, deducts amount from available limit
- Calls `profileUpdater.UpdateLimit` to update credit line
- Currently simplified (TODO: calculate from current available - amount)
- Ensures credit limit is enforced at order creation time

### 6. Payment Mapping
- Creates mapping between `paymentId` and `lender_merchant_txn_id`
- Enables reverse lookup by lender transaction ID
- Stored in `lender_payment_mapping` table
- Used for reconciliation and webhook processing

### 7. Event Publishing
- Order events published for: Created, Completed, Failed, Refunded
- Stub publisher is no-op (production would use Kafka)
- Events include full context (payment ID, status, lender IDs, errors)

## Idempotency Flow Diagram

```
Client Request (paymentId: "pay-123", amount: 1000)
    ↓
ComputeHash(request) → "abc123..."
    ↓
TryAcquire("pay-123", "abc123...")
    ↓
┌─────────────────────────────────────┐
│ Database Check                      │
├─────────────────────────────────────┤
│ IF NOT EXISTS:                      │
│   INSERT (PROCESSING) → NEW         │
│ ELSE IF hash matches + COMPLETED:    │
│   RETURN cached → DUPLICATE         │
│ ELSE IF hash matches + PROCESSING:   │
│   RETURN → CONFLICT (409)           │
│ ELSE IF hash differs:                │
│   RETURN → MISMATCH (422)           │
└─────────────────────────────────────┘
    ↓ (if NEW)
Create Order + Mapping
    ↓
Gateway.CreateOrder()
    ↓
MarkCompleted("pay-123", response, orderID)
    ↓
Return OrderResponse
```

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

### 3. Test Create Order
```bash
curl -X POST http://localhost:8080/v1/payin3/order \
  -H "Content-Type: application/json" \
  -d '{
    "paymentId": "pay-123",
    "userId": "u1",
    "mobile": "9876543210",
    "amount": 1000,
    "currency": "INR",
    "merchantId": "m1",
    "returnUrl": "https://merchant.com/return",
    "emiSelection": {
      "tenure": 3,
      "emi": 333.33,
      "totalPayableAmount": 1000
    }
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "paymentId": "pay-123",
    "status": "PENDING",
    "lenderOrderId": "LP-ORDER-pay-123",
    "redirectUrl": "https://stub.lazypay.in/payment/..."
  }
}
```

### 4. Test Idempotency (Duplicate Request)
```bash
# Send same request again
curl -X POST http://localhost:8080/v1/payin3/order \
  -H "Content-Type: application/json" \
  -d '{
    "paymentId": "pay-123",
    "userId": "u1",
    "mobile": "9876543210",
    "amount": 1000,
    "currency": "INR",
    "merchantId": "m1",
    "returnUrl": "https://merchant.com/return",
    "emiSelection": {
      "tenure": 3,
      "emi": 333.33,
      "totalPayableAmount": 1000
    }
  }'
```

**Expected Response:** Same as first request (cached response)

### 5. Test Hash Mismatch
```bash
# Same paymentId, different amount
curl -X POST http://localhost:8080/v1/payin3/order \
  -H "Content-Type: application/json" \
  -d '{
    "paymentId": "pay-123",
    "userId": "u1",
    "mobile": "9876543210",
    "amount": 2000,
    "currency": "INR",
    "merchantId": "m1",
    "returnUrl": "https://merchant.com/return",
    "emiSelection": {
      "tenure": 3,
      "emi": 666.67,
      "totalPayableAmount": 2000
    }
  }'
```

**Expected Response:**
```json
{
  "success": false,
  "error": {
    "code": "IDEMPOTENCY_HASH_MISMATCH",
    "message": "request hash mismatch: same paymentId with different request body",
    "statusCode": 422
  }
}
```

### 6. Test Get Order Status
```bash
curl http://localhost:8080/v1/payin3/order/pay-123
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "paymentId": "pay-123",
    "userId": "u1",
    "merchantId": "m1",
    "amount": 1000,
    "currency": "INR",
    "status": "PENDING",
    "lenderOrderId": "LP-ORDER-pay-123",
    "createdAt": "...",
    "updatedAt": "..."
  }
}
```

### 7. Test Order Callback
```bash
curl -X POST http://localhost:8080/v1/payin3/callback/order \
  -H "Content-Type: application/json" \
  -d '{
    "paymentId": "pay-123",
    "provider": "LAZYPAY",
    "status": "SUCCESS",
    "lenderOrderId": "LP-ORDER-pay-123",
    "lenderTxnId": "LP-TXN-123",
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
- **Phase 2B**: Onboarding module (for reference on event sourcing pattern)

## What Comes Next

### Phase 2D: Refund Module
- Refund initiation endpoint
- Refund status tracking
- Integration with order module
- Refund callback processing
- Profile limit restoration on refund

### Future Enhancements
- Replace stub gateway with real Lazypay adapter
- Replace stub event publisher with Kafka producer
- Implement proper limit calculation (current available - amount)
- Add webhook signature verification
- Add comprehensive unit and integration tests
- Add idempotency key expiration cleanup job
- Add order expiration handling

## Notes

- All service methods use `context.Context` for cancellation/timeout support
- Idempotency keys expire after 24 hours
- Terminal order states (SUCCESS, FAILED, REFUNDED, EXPIRED, CANCELLED) are idempotent
- Payment mapping enables reverse lookup by lender transaction ID
- Profile limit is updated on order SUCCESS (deducted) and refund (restored)
- Repository mapping functions handle type conversions between entity types and DB strings
- Module follows dependency injection pattern
