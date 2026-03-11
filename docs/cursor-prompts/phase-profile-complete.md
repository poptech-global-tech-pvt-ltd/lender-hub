# Phase Profile Module — Complete

## Summary

Profile module refactor: split response DTOs, gateway port with port types, repository upserts from gateway results, upstream sync (MockClient), and combined GET API.

## Completed

### Part 1 — Split Response DTOs
- **Adapter**: `LPEligibilityResponse` (COF section), `LPCustomerStatusResponse` — already aligned
- **Domain**:
  - `EligibilityResponse` — userId, lender, txnEligible, eligibilityCode, eligibilityReason, availableLimit, creditLimit, emiPlans, existingUser, checkedAt
  - `CustomerStatusResponse` — userId, lender, preApproved, onboardingRequired, onboardingDone, ntbEligible, availableLimit, checkedAt
  - `UserProfileResponse` — combined; status NEVER empty (ACTIVE|INELIGIBLE|NOT_STARTED|BLOCKED)

### Part 2 — Gateway Port
- `ProfileGateway.CheckEligibility` → `*port.EligibilityResult`
- `ProfileGateway.GetCustomerStatus` → `*port.CustomerStatusResult`
- `EligibilityResult`, `CustomerStatusResult`, `EmiPlanResult` defined in port

### Part 3 — Lazypay Mapper
- `MapEligibilityResponse(lp)` → `*port.EligibilityResult` (COF only)
- `MapCustomerStatusResponse(lp)` → `*port.CustomerStatusResult`

### Part 4 — Repository Upserts
- `UpsertFromEligibility(ctx, userID, lender, result *port.EligibilityResult)`
- `UpsertFromCustomerStatus(ctx, userID, lender, result *port.CustomerStatusResult)`
- `UpsertFromCombined(ctx, userID, lender, cs *port.CustomerStatusResult, el *port.EligibilityResult)`
- Status derived: ACTIVE, INELIGIBLE, NOT_STARTED

### Part 5 — Upstream Sync
- `internal/infrastructure/userprofile/mock_client.go` — `MockClient` logs TODO
- `port.ProfileSyncer` — `UpdateLenderProfile(ctx, req) error`

### Part 6 — Profile Service
- `CheckEligibility` — resolve contact → gateway → persist → async sync → return response
- `GetCustomerStatus` — same flow
- `GetCombinedProfile` — parallel calls when amount > 0, eligibility non-fatal

### Part 7 — Combined GET Handler
- `GET /v1/payin3/profile/:userId?amount=&currency=INR&source=`
- amount optional; 0 = CustomerStatus only

### Part 8 — Module Wiring
- `POST /v1/payin3/eligibility`
- `POST /v1/payin3/customer-status`
- `GET /v1/payin3/profile/:userId`
- `profileSyncer := userprofile.NewMockClient(logger)` in main.go

## Verification

- [x] `go build ./...` — zero errors
- [x] `go vet ./...` — zero warnings
