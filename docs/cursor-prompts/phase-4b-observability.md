# Phase 4B: Structured Logging + Datadog Custom Metrics

**Date:** February 13, 2025  
**Status:** ✅ Complete

## Overview

This phase implements production-ready observability with structured JSON logging (zap) and Datadog statsd custom metrics. Both components have noop fallbacks for local development. **NO OpenTelemetry or distributed tracing** - not required for this phase.

## Files Created (6 total)

### Metrics Infrastructure (4 files)
1. `internal/infrastructure/observability/metrics/config.go` - Datadog configuration
2. `internal/infrastructure/observability/metrics/metrics.go` - MetricsClient interface and metric name constants
3. `internal/infrastructure/observability/metrics/datadog.go` - Datadog statsd client wrapper
4. `internal/infrastructure/observability/metrics/noop.go` - Noop metrics client for local dev

### Logging Infrastructure (2 files)
5. `internal/infrastructure/observability/logging/logger.go` - Structured logger with zap
6. `internal/infrastructure/observability/logging/context.go` - Context helpers and field constructors

## Metrics Table

| Category | Metric Name | Type | Tags | Description |
|----------|-------------|------|------|-------------|
| **Business** | `orders.created` | Count | provider, status | Order creation attempts |
| **Business** | `orders.success` | Count | provider | Successful orders |
| **Business** | `orders.failed` | Count | provider, errorCode | Failed orders |
| **Business** | `eligibility.checks` | Count | provider, source | Eligibility check calls |
| **Business** | `eligibility.eligible` | Count | provider | Users found eligible |
| **Business** | `eligibility.ineligible` | Count | provider, reasonCode | Users found ineligible |
| **Business** | `onboarding.started` | Count | provider, source | Onboarding initiations |
| **Business** | `onboarding.completed` | Count | provider | Successful onboardings |
| **Business** | `onboarding.failed` | Count | provider, reasonCode | Failed onboardings |
| **Business** | `refunds.initiated` | Count | provider | Refund initiation attempts |
| **Business** | `refunds.completed` | Count | provider | Completed refunds |
| **Business** | `idempotency.duplicates` | Count | — | Duplicate requests caught |
| **Provider** | `lazypay.requests` | Count | endpoint, method | Total Lazypay API calls |
| **Provider** | `lazypay.errors` | Count | endpoint, errorCode | Lazypay error responses |
| **Provider** | `lazypay.latency` | Histogram | endpoint | Lazypay response time |
| **Provider** | `lazypay.circuit_breaker.state` | Gauge | executor | 0=closed, 1=open, 2=half-open |
| **Provider** | `lazypay.timeout` | Count | endpoint | Lazypay timeouts |
| **System** | `api.latency` | Timing | method, path, status | API response time |
| **System** | `api.requests` | Count | method, path, status | Total API requests |
| **System** | `api.errors` | Count | method, path, status | API error responses |
| **System** | `db.query.duration` | Histogram | query | DB query latency |
| **System** | `db.connections.active` | Gauge | — | Open DB connections |
| **System** | `cache.hit` | Count | — | Redis cache hits |
| **System** | `cache.miss` | Count | — | Redis cache misses |
| **System** | `panics` | Count | path | Recovered panics |

## Logging JSON Sample

### Production Mode (JSON)
```json
{
  "level": "info",
  "ts": "2026-02-13T02:00:00Z",
  "caller": "handler/customer_status_handler.go:45",
  "msg": "request completed",
  "service": "payin3-service",
  "env": "production",
  "requestId": "abc-123",
  "userId": "user-456",
  "module": "profile",
  "method": "POST",
  "path": "/v1/payin3/customer-status",
  "status": 200,
  "durationMs": 142
}
```

### Development Mode (Colored Console)
```
2026-02-13T02:00:00Z	INFO	handler/customer_status_handler.go:45	request completed	{"service": "payin3-service", "env": "local", "requestId": "abc-123", "userId": "user-456", "module": "profile", "method": "POST", "path": "/v1/payin3/customer-status", "status": 200, "durationMs": 142}
```

## Log Field Helpers

| Helper Function | Field Name | Type | Example |
|----------------|------------|------|---------|
| `RequestID(id)` | `requestId` | string | `"abc-123"` |
| `UserID(id)` | `userId` | string | `"user-456"` |
| `Module(name)` | `module` | string | `"profile"` |
| `PaymentID(id)` | `paymentId` | string | `"pay-789"` |
| `RefundID(id)` | `refundId` | string | `"refund-012"` |
| `Provider(name)` | `provider` | string | `"LAZYPAY"` |
| `Endpoint(ep)` | `endpoint` | string | `"/v7/payment/eligibility"` |
| `DurationMs(ms)` | `durationMs` | int64 | `142` |
| `Status(s)` | `status` | string | `"SUCCESS"` |
| `ErrorCode(code)` | `errorCode` | string | `"USER_INELIGIBLE"` |
| `HTTPStatus(code)` | `httpStatus` | int | `200` |

## Key Design Decisions

### 1. Metrics Client Interface
- **Unified Interface**: `MetricsClient` interface for both Datadog and Noop
- **Method Types**: Count, Gauge, Histogram, Timing
- **Tag Support**: All methods accept tags for filtering/grouping
- **Graceful Shutdown**: `Close()` method for cleanup

### 2. Datadog Integration
- **Namespace**: All metrics prefixed with `payin3.` automatically
- **Global Tags**: Service, env, version tags applied to all metrics
- **Statsd Protocol**: Uses UDP statsd protocol (port 8125)
- **Error Handling**: Errors are ignored (fire-and-forget pattern)

### 3. Structured Logging
- **Production**: JSON output for log aggregation systems
- **Development**: Colored console output for readability
- **Context Fields**: Service name and env in all logs
- **Caller Info**: File and line number included

### 4. Context-Based Logging
- **Logger in Context**: Store logger in request context
- **Automatic Propagation**: Logger passed through service layers
- **Noop Fallback**: Returns noop logger if not in context
- **Field Helpers**: Convenient constructors for common fields

### 5. Noop Pattern
- **Metrics**: `NoopClient` implements `MetricsClient` with no-ops
- **Logging**: `zap.NewNop()` for noop logger
- **Fallback**: Automatic when Datadog/logging not configured
- **Zero Overhead**: No performance impact when disabled

## Verification Steps

### 1. Build Verification
```bash
go mod tidy
go build ./...
```

### 2. Verify Interface Compliance
```bash
go build ./internal/infrastructure/observability/metrics/...
```

**Verification:**
- `DatadogClient` satisfies `MetricsClient`
- `NoopClient` satisfies `MetricsClient`

### 3. Test Logger JSON Output
```go
// Test in production mode
logger, _ := logging.NewLogger("payin3-service", "production")
logger.Info("test message",
    logging.RequestID("req-123"),
    logging.UserID("user-456"),
    logging.Module("profile"),
)
logger.Sync()
```

**Expected Output (JSON):**
```json
{"level":"info","ts":"...","caller":"...","msg":"test message","service":"payin3-service","env":"production","requestId":"req-123","userId":"user-456","module":"profile"}
```

### 4. Test Logger Colored Output
```go
// Test in development mode
logger, _ := logging.NewLogger("payin3-service", "local")
logger.Info("test message",
    logging.RequestID("req-123"),
    logging.UserID("user-456"),
)
logger.Sync()
```

**Expected Output (Colored Console):**
```
2026-02-13T...	INFO	...	test message	{"service": "payin3-service", "env": "local", "requestId": "req-123", "userId": "user-456"}
```

### 5. Test Metrics Client
```go
// Test Datadog client (if configured)
metrics, _ := metrics.NewDatadogClient(metrics.DefaultConfig())
metrics.Count(metrics.MetricOrdersCreated, 1, []string{"provider:LAZYPAY"})
metrics.Close()
```

### 6. Test Noop Fallback
```go
// Test noop client
noop := metrics.NewNoopClient()
noop.Count(metrics.MetricOrdersCreated, 1, []string{})
noop.Close() // Should not error
```

## Dependencies on Previous Phases

- **Phase 0**: Config structure (for metrics/logging config)
- **Phase 2A-2D**: Domain modules (will use metrics/logging in Phase 4C)

## What Comes Next

### Phase 4C: Middleware + Main.go Wiring
- Update middleware to use structured logger
- Add metrics middleware for API requests
- Wire metrics and logger into main.go
- Add metrics emission in services
- Add logging in handlers and services

### Future Enhancements
- Custom dashboards in Datadog
- Alert rules based on metrics
- Log aggregation with ELK/CloudWatch
- Performance profiling
- Error tracking integration

## Notes

- **NO OpenTelemetry**: Distributed tracing not implemented
- **NO Tracing Package**: Completely skipped as per requirements
- **Metrics Namespace**: All metrics prefixed with `payin3.` automatically
- **Log Context**: Logger stored in request context for propagation
- **Field Helpers**: 12 helper functions for common log fields
- **Noop Fallback**: Both metrics and logging have noop implementations
- **Production JSON**: Structured JSON output for log aggregation
- **Development Colored**: Human-readable colored output for local dev
- **Graceful Shutdown**: Both metrics and logging support `Close()`/`Sync()`

---

## 4B-FIX: Namespace Rename + Metric Cleanup

**Date:** 2026-02-13  
**Reason:** Datadog Agent auto-tracks system and provider metrics

**Changes:**
1. Namespace: "payin3" → "lsp" (all metrics now prefixed `lsp.*`)
2. Dropped system metrics — handled by DD Agent APM + integrations:
   - `api.requests`, `api.latency`, `api.errors`
   - `db.query.duration`, `db.connections.active`
   - `cache.hit`, `cache.miss`
   - `panics`
3. Dropped provider metrics — handled by dd-trace-go HTTP client wrapping:
   - `lazypay.requests`, `lazypay.errors`, `lazypay.latency`
   - `lazypay.circuit_breaker.state`, `lazypay.timeout`
4. Retained 12 business-only metrics (see table below)

**Final metric inventory:**

| Metric Name | Type | Tags | Purpose |
|-------------|------|------|---------|
| `lsp.eligibility.checked` | Count | provider, source | Eligibility check calls |
| `lsp.eligibility.eligible` | Count | provider | Users found eligible |
| `lsp.eligibility.ineligible` | Count | provider, reasonCode | Users found ineligible |
| `lsp.onboarding.started` | Count | provider, source | Onboarding initiations |
| `lsp.onboarding.completed` | Count | provider | Successful onboardings |
| `lsp.onboarding.failed` | Count | provider, reasonCode | Failed onboardings |
| `lsp.order.created` | Count | provider | Order creation attempts |
| `lsp.order.success` | Count | provider | Successful payments |
| `lsp.order.failed` | Count | provider, errorCode | Failed payments |
| `lsp.refund.initiated` | Count | provider | Refund requests |
| `lsp.refund.completed` | Count | provider | Completed refunds |
| `lsp.idempotency.duplicate` | Count | — | Duplicate order attempts caught |

**Files modified:** `config.go`, `metrics.go` (2 files only)  
**Files unchanged:** `datadog.go`, `noop.go`, `logging/logger.go`, `logging/context.go`

**What Datadog Agent Handles Automatically:**
- HTTP request count, latency, errors per endpoint → DD APM + dd-trace-go/contrib/gin
- Database query duration, connection pool stats → DD Postgres integration
- Redis commands, latency, hit/miss → DD Redis integration
- Outbound HTTP calls to Lazypay (count, latency, errors) → dd-trace-go/contrib/net/http
- Go runtime (goroutines, GC, memory) → DD runtime metrics
- Container/infra (CPU, memory, network) → DD Agent
