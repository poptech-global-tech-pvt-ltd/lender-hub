# Phase: Onboarding Boundary Fix + Discarded Errors + Lender Constant

## Completed

### Part 1 — Shared Lender Constant
- Created `pkg/lender/lender.go` with `Lender` type and `Lazypay` constant
- Replaced all `"LAZYPAY"` occurrences across `internal/domain/`, `internal/adapter/lazypay/` with `lender.Lazypay.String()` or equivalent

### Part 2 — Onboarding Port Interfaces
- Created `internal/domain/onboarding/port/profile_updater.go` with `UpdateOnOnboardingSuccess`
- Created `internal/domain/onboarding/port/contact_resolver.go` with `GetContact` and `ContactInfo`

### Part 3 — Onboarding Service Fix
- Onboarding service now uses `port.ProfileUpdater` and `port.ContactResolver` instead of concrete profile types
- Removed `profileService` import from onboarding module entirely

### Part 4 — Wiring
- Created `internal/adapter/contact/onboarding_contact_adapter.go` to adapt `UserContactResolver` to onboarding's `ContactResolver`
- Updated `cmd/server/main.go` to pass `onboardingContactResolver` to onboarding module
- Onboarding module accepts interfaces; profile module's `ProfileUpdater` satisfies `port.ProfileUpdater` structurally

### Part 5 — Discarded Error Logging
- Order service: added logger, warn logs for `RefreshFromSource`, `mappingRepo.Create`, `idempotency.Complete`, `publisher.Publish` failures
- Refund service: warn logs for `repo.Update`, `cache.Set`, `enquiryService.ResolveRefundState` failures
- Refund enquiry service: warn log for `cache.Set` failure
- Onboarding service: added logger, warn log for `profileUpdater.UpdateOnOnboardingSuccess` failure

### Part 6 — Unit Tests
- `pkg/idgen/idgen_test.go`: prefix tests (lps_, ref_, onb_), uniqueness, sortable, concurrent uniqueness
- `internal/domain/order/entity/order_status_test.go`: IsTerminal, OrDefault, NormalizeForDB
- `internal/domain/refund/entity/refund_status_test.go`: IsTerminal, IsResolvable, OrDefault
- `internal/adapter/lazypay/mapper/order_status_test.go`: LP order status → canonical mapping

## Verification

```bash
go build ./...   # ✅
go vet ./...     # ✅
go test ./...    # ✅
```

- `grep -r 'profileService\.' internal/domain/onboarding/` → no results
- `grep -r '"LAZYPAY"' internal/` → no results
- `grep -r '_ = s.cache' internal/` → no results
- `grep -r '_ = s.repo' internal/` → no results

## Remaining (Low Priority)

- `orderservice_test.go`: CreateOrder idempotency and gateway failure tests (require full mock setup)
- `refundservice_test.go`: LPDUPLICATEREFUND path test
- `refundenquiryservice_test.go`: txnRefNo match, lpTxnId fallback, SLA tests
