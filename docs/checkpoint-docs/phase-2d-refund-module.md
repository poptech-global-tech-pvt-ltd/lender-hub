# Phase 2D: Refund Module

**Date:** February 13, 2025  
**Status:** ✅ Complete

## Overview

This phase implements the complete Refund vertical slice with validation logic, limit restoration, and order status management. The module ensures refund amounts don't exceed order amounts and automatically restores user credit limits on successful refunds.

## Files Created/Modified

### Refund Module - Entities
1. `internal/domain/refund/entity/refund_status.go` - Refund status enum
2. `internal/domain/refund/entity/refund_reason.go` - Refund reason enum
3. `internal/domain/refund/entity/refund.go` - Updated with RefundStatus and RefundReason types

### Refund Module - DTOs
4. `internal/domain/refund/dto/request/create_refund.go` - Create refund request
5. `internal/domain/refund/dto/request/refund_callback.go` - Refund callback request
6. `internal/domain/refund/dto/response/refund_response.go` - Refund response with LenderRefID

### Refund Module - Ports
7. `internal/domain/refund/port/gateway.go` - RefundGateway interface
8. `internal/domain/refund/port/repository.go` - Verified from Phase 1

### Refund Module - Services
9. `internal/domain/refund/service/refund_service.go` - RefundService with CreateRefund and ProcessCallback

### Refund Module - Handlers
10. `internal/domain/refund/handler/create_refund_handler.go` - POST /v1/payin3/refund
11. `internal/domain/refund/handler/refund_callback_handler.go` - POST /v1/payin3/callback/refund

### Refund Module - Stubs
12. `internal/domain/refund/stub/stub_gateway.go` - Stub gateway returning fake refund responses

### Refund Module - Wiring
13. `internal/domain/refund/module.go` - Module wiring and route registration
14. `internal/domain/refund/repository/postgres_repository.go` - Updated mapping for RefundStatus and RefundReason

### Profile Module - Enhancement
15. `internal/domain/profile/service/profile_updater.go` - Added AddToLimit method for limit restoration

### Integration
16. `cmd/server/main.go` - Updated to register refund module routes

### Documentation
17. `docs/cursor-prompts/phase-2d-refund-module.md` - This checkpoint document
18. `docs/cursor-prompts/README.md` - Knowledge hub index

## API Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/v1/payin3/refund` | Create a new refund with validation |
| POST | `/v1/payin3/callback/refund` | Process refund callback events |

## Key Design Decisions

### 1. Refund Validation Logic
- **Order Status Check**: Only orders with SUCCESS status can be refunded
- **Amount Validation**: Single refund amount must be > 0
- **Total Validation**: Sum of all SUCCESS refunds + new refund amount must not exceed order amount
- **Locking**: Uses FOR UPDATE on both `lender_payment_state` and `lender_refunds` (via GetForUpdate)

### 2. Refund Validation Flow

```
CreateRefund Request
    ↓
GetForUpdate(order) → Lock order row
    ↓
Validate order.Status == SUCCESS
    ↓
ListByPaymentID(refunds) → Get existing refunds
    ↓
Sum SUCCESS refund amounts
    ↓
Check: totalRefunded + newAmount <= order.Amount
    ├─ YES → Proceed
    └─ NO → Return 422 (REFUND_EXCEEDS_ORDER_AMOUNT)
    ↓
Gateway.ProcessRefund()
    ↓
Create refund entity (status=PENDING)
    ↓
refundRepo.Create()
    ↓
Return RefundResponse
```

### 3. Limit Restoration Flow

```
Refund Callback (SUCCESS)
    ↓
Get refund by refundID
    ↓
Update refund status to SUCCESS
    ↓
profileUpdater.AddToLimit(userID, lender, refund.Amount)
    ├─ GetForUpdate(profile)
    ├─ newAvailable = currentAvailable + refund.Amount
    ├─ Cap at creditLimit if exceeds
    └─ Update profile
    ↓
Check if order fully refunded:
    ├─ ListByPaymentID(all refunds)
    ├─ Sum SUCCESS refund amounts
    ├─ If totalRefunded >= order.Amount:
    │   └─ order.Status = REFUNDED
    └─ Update order
    ↓
Update refund
```

### 4. Order Status Management
- When all refunds for an order are SUCCESS and total equals order amount
- Order status is automatically set to REFUNDED
- Uses FOR UPDATE locking to prevent race conditions
- Only counts SUCCESS refunds in the total

### 5. Profile Limit Restoration
- `AddToLimit` method added to ProfileUpdater
- Adds refund amount to current available limit
- Caps at credit limit if addition would exceed it
- Publishes LimitUpdated event
- Used in refund callback on SUCCESS status

### 6. FOR UPDATE Locking
- `GetForUpdate` used on order during refund creation
- Prevents concurrent refunds from exceeding order amount
- Ensures atomic validation and creation
- Used in callback to safely update order status

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

### 3. Test Create Refund (Valid)
```bash
# First, create an order (use paymentId from order creation)
# Then create refund:
curl -X POST http://localhost:8080/v1/payin3/refund \
  -H "Content-Type: application/json" \
  -d '{
    "refundId": "refund-123",
    "paymentId": "pay-123",
    "amount": 500,
    "currency": "INR",
    "reason": "USER_CANCELLED"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "refundId": "refund-123",
    "paymentId": "pay-123",
    "provider": "LAZYPAY",
    "status": "PENDING",
    "amount": 500,
    "currency": "INR"
  }
}
```

### 4. Test Refund Exceeding Order Amount
```bash
# Try to refund more than order amount
curl -X POST http://localhost:8080/v1/payin3/refund \
  -H "Content-Type: application/json" \
  -d '{
    "refundId": "refund-124",
    "paymentId": "pay-123",
    "amount": 2000,
    "currency": "INR",
    "reason": "USER_CANCELLED"
  }'
```

**Expected Response:**
```json
{
  "success": false,
  "error": {
    "code": "REFUND_EXCEEDS_ORDER_AMOUNT",
    "message": "refund amount exceeds order amount",
    "statusCode": 422
  }
}
```

### 5. Test Refund Callback (SUCCESS)
```bash
curl -X POST http://localhost:8080/v1/payin3/callback/refund \
  -H "Content-Type: application/json" \
  -d '{
    "refundId": "refund-123",
    "paymentId": "pay-123",
    "provider": "LAZYPAY",
    "status": "SUCCESS",
    "lenderRefId": "LP-REFUND-refund-123",
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

**Verification:**
- Refund status updated to SUCCESS
- User available limit increased by refund amount
- If all refunds complete and total == order amount, order status set to REFUNDED

## Dependencies on Previous Phases

- **Phase 0**: Config, Postgres connection, middleware
- **Phase 1**: GORM models, repository interfaces, Postgres implementations, migrations
- **Phase 2A**: Shared kernel (response envelope, error codes), Profile module (ProfileUpdater with AddToLimit)
- **Phase 2B**: Onboarding module (for reference on callback pattern)
- **Phase 2C**: Order module (OrderRepository for validation and status updates)

## What Comes Next

### Phase 3: Lazypay Adapter
- Real HTTP client implementation
- HMAC signature generation and verification
- Request/response mappers
- Error handling and retries
- Replace all stub gateways with real implementations

### Future Enhancements
- Add comprehensive unit and integration tests
- Add refund status query endpoint
- Add refund history endpoint
- Add webhook signature verification
- Add refund expiration handling
- Add partial refund support with better validation
- Add refund reversal support

## Notes

- All service methods use `context.Context` for cancellation/timeout support
- Refund validation uses FOR UPDATE locking to prevent race conditions
- Limit restoration uses `AddToLimit` which caps at credit limit
- Order status automatically set to REFUNDED when fully refunded
- Repository mapping functions handle type conversions between entity types and DB strings
- Module follows dependency injection pattern
- Refund reasons: USER_CANCELLED, PRODUCT_RETURN, ORDER_CANCELLED
