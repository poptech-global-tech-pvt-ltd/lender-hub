# Cursor Prompt Knowledge Hub

This folder contains checkpoint documents for each implementation phase.
Each .md file records what was built, design decisions, and verification steps.

## Phase Index

| Phase | File | Description |
|-------|------|-------------|
| 2A | [phase-2a-profile-module.md](./phase-2a-profile-module.md) | Shared kernel + Profile module |
| 2B | [phase-2b-onboarding-module.md](./phase-2b-onboarding-module.md) | Onboarding module + event sourcing |
| 2C | [phase-2c-order-module.md](./phase-2c-order-module.md) | Order module + DB idempotency |
| 2D | [phase-2d-refund-module.md](./phase-2d-refund-module.md) | Refund module + limit restoration |
| 3A | [phase-3a-infra-executors.md](./phase-3a-infra-executors.md) | HTTP executors, circuit breaker, signature |
| 3B | [phase-3b-lazypay-dtos-mappers.md](./phase-3b-lazypay-dtos-mappers.md) | Lazypay DTOs + mappers |
| 3C | [phase-3c-lazypay-clients.md](./phase-3c-lazypay-clients.md) | Lazypay clients + adapter wiring |

## How to Use

1. **Run phases in order** (2A → 2B → 2C → 2D)
2. **After each phase**: Run `go build ./...` to verify compilation
3. **Check the checkpoint .md** for curl test commands
4. **Each phase lists its dependencies** clearly

## Phase Dependencies

```
Phase 0 (Infrastructure)
  ├─ Config, Postgres, Middleware
  └─ Migrations

Phase 1 (Models & Repositories)
  ├─ GORM Models
  ├─ Repository Interfaces
  └─ Repository Implementations

Phase 2A (Shared Kernel + Profile)
  ├─ Response Envelope
  ├─ Error Handling
  └─ Profile Module

Phase 2B (Onboarding)
  └─ Depends on: 2A (ProfileUpdater)

Phase 2C (Order)
  └─ Depends on: 2A (ProfileUpdater)

Phase 2D (Refund)
  └─ Depends on: 2A (ProfileUpdater), 2C (OrderRepository)

Phase 3A (Infrastructure)
  └─ No dependencies (foundation layer)

Phase 3B (DTOs + Mappers)
  └─ Depends on: 3A (signature service), 2A-2D (domain DTOs)

Phase 3C (Clients + Wiring)
  └─ Depends on: 3A (executors, signature), 3B (DTOs, mappers), 2A-2D (gateway interfaces)
```

## Quick Start

```bash
# 1. Verify all phases compile
go build ./...

# 2. Run server
go run cmd/server/main.go

# 3. Test endpoints (see individual phase docs for curl commands)
```

## Next Steps

After Phase 2D, proceed to **Phase 3: Lazypay Adapter** to replace stub implementations with real HTTP clients.
