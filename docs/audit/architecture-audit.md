# Architecture & Go Practices Audit

**Date:** 2025-02-19  
**Scope:** `internal/`, `pkg/`, `cmd/`  
**Service:** lending-hub-service (Go, Gin, GORM, Postgres, Redis, Kafka)

---

## Executive Summary

- **go build ./...**: ✅ Passes
- **go vet ./...**: ✅ Passes
- **Import cycles**: None detected
- **Unused imports**: None (goimports -l clean)

---

## Violations Found & Fixes Applied

### Category 1 — Module Boundary Violations

| Severity | File | Issue | Fix Applied |
|----------|------|-------|-------------|
| HIGH | `order/service` | Concrete `*profileService.ProfileUpdater` | ✅ `port.ProfileUpdater` interface in `order/port/profile_updater.go` |
| HIGH | `order/service` | Concrete `*profileService.UserContactResolver` | ✅ `port.ContactResolver` interface + `OrderContactAdapter` in `adapter/contact/` |
| HIGH | `refund/service` | Concrete `*profileService.ProfileUpdater` | ✅ `port.ProfileUpdater` with `AddToLimit` in `refund/port/profile_updater.go` |
| MEDIUM | `onboarding/service` | Concrete `*profileService.ProfileUpdater`, `*profileService.UserContactResolver` | ❌ TODO: Add onboarding port interfaces and adapter |

### Category 2 — Interface Over-Exposure

- **ProfileUpdater** split per consumer: order uses `UpdateLimit`, refund uses `AddToLimit` — narrow interfaces defined in each consuming module.
- No unnecessary interface embedding identified.

### Category 3 — Concrete Types as Dependencies

| Module | Constructor | Status |
|--------|-------------|--------|
| order | `NewOrderService(profileUpdater port.ProfileUpdater, contactResolver port.ContactResolver, ...)` | ✅ Interfaces |
| refund | `NewRefundService(profileUpdater port.ProfileUpdater, ...)` | ✅ Interface |
| order module | `NewModule(profileUpdater port.ProfileUpdater, contactResolver port.ContactResolver, ...)` | ✅ Interfaces |
| refund module | `NewModule(profileUpdater refundPort.ProfileUpdater, ...)` | ✅ Interface |

### Category 4 — Error Handling

| File | Issue | Fix |
|------|-------|-----|
| `refund/service` | `refund, _ = s.repo.GetByPaymentRefundID` — nil overwrite risk | ✅ Use `refreshed, err :=` and only overwrite on success |
| `refund/service` | `refund, _ = s.repo.GetByRefundID` | ✅ Same fix |
| Multiple | `_ = s.repo.Update`, `_ = s.cache.Set` | TODO: Log at warn level if non-fatal |
| Multiple | Discarded errors in non-critical paths | Documented; consider structured logging |

### Category 4e — Goroutine Panic Recovery

| File | Issue | Fix |
|------|-------|-----|
| `refund/service` | `go s.profileUpdater.AddToLimit(...)` | ✅ Wrapped in `defer recover()` |
| `profile/service` | `go s.syncUpstream(...)` | ✅ Added `defer recover()` in `syncUpstream` |

### Category 5 — Context Propagation

- All I/O functions accept `ctx context.Context` as first parameter.
- No context stored in structs.
- `context.Background()` in async paths (AddToLimit, syncUpstream) — intentional (request context cancelled when handler returns). ✅

### Category 6 — GORM Patterns

- Repositories use `db.WithContext(ctx)` on queries.
- `FOR UPDATE` uses `clause.Locking` where required.
- Transaction wrapping present for multi-step operations.

### Category 7 — Concurrency Safety

- **idgen**: Uses `sync.Mutex` to protect entropy. ✅
- No unsafe package-level mutable state identified.

### Category 8 — HTTP Handler Patterns

- Handlers are thin: bind → service call → response.
- Error responses use shared error mapper.
- Request ID propagation via middleware.

### Category 9 — Naming & Go Idioms

- Acronyms: `UserID`, `PaymentID`, `LoanID` — consistent.
- Constructor return types: concrete `*OrderService` — acceptable (callers rarely need interface for ordering).

### Category 10 — Logging & Observability

- Structured zap-based logging throughout.
- Request/correlation IDs in handlers.
- Panic recovery logs with context.

### Category 11 — Configuration

- Hardcoded `"LAZYPAY"` in profile service — consider config/constant.
- Timeouts and keys from config where applicable.

### Category 12 — Test Coverage Gaps

- idgen uniqueness/prefix
- OrderStatus.IsTerminal(), OrDefault()
- RefundStatus transitions
- MapLPOrderStatusToCanonical edge cases
- Idempotency: duplicate paymentId on CreateOrder
- LPDUPLICATEREFUND path in CreateRefund
- Enquiry txnRefNo fallback logic

---

## Files Created

- `internal/domain/order/port/profile_updater.go`
- `internal/domain/order/port/contact_resolver.go`
- `internal/domain/refund/port/profile_updater.go`
- `internal/adapter/contact/order_contact_adapter.go`

---

## Remaining TODOs (Medium/Low)

1. **Onboarding module**: Add `port.ProfileUpdater` and `port.ContactResolver` in onboarding, adapter in main.
2. **Discarded errors**: Add warn-level logging for non-fatal `_ = repo.Update`, `_ = cache.Set` etc.
3. **Lender constant**: Move `"LAZYPAY"` to config or shared constant.
4. **Tests**: Add tests per Category 12.

---

## Verification

```bash
go build ./...   # ✅
go vet ./...     # ✅
go list ./...    # No import cycles
goimports -l ./internal ./cmd ./pkg  # No unused imports
```
