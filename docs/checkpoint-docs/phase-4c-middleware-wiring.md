# Phase 4C: Middleware + Request Context + Health + Main.go Wiring

**Date:** February 13, 2025  
**Status:** ✅ Complete

## Overview

This phase implements HTTP middleware, request context propagation, health check endpoints, and wires everything together in `main.go`. **NO metrics middleware** - Datadog APM auto-instruments Gin routes. **NO OpenTelemetry or distributed tracing**.

## Files Created (7 total)

### Request Context (1 file)
1. `internal/shared/context/request_context.go` - Canonical request context struct

### Middleware (4 files)
2. `internal/infrastructure/middleware/request_id.go` - Request ID generation/extraction
3. `internal/infrastructure/middleware/logging.go` - Structured request/response logging
4. `internal/infrastructure/middleware/recovery.go` - Panic recovery with stack trace logging
5. `internal/infrastructure/middleware/context_headers.go` - Platform context header extraction

### Health (1 file)
6. `internal/infrastructure/health/health.go` - Health check endpoints (liveness + readiness)

### Main Application (1 file)
7. `cmd/server/main.go` - Complete application entry point with all infrastructure wiring

## Middleware Execution Order

| Order | Middleware | Purpose | File |
|-------|------------|---------|------|
| 1 | RequestID | Generate/extract X-Request-ID | `request_id.go` |
| 2 | Recovery | Panic catch, log stack trace | `recovery.go` |
| 3 | RequestLogging | Structured JSON request log | `logging.go` |
| 4 | ContextHeaders | Extract x-platform, x-device-id, x-user-ip | `context_headers.go` |

**NOTE:** No RequestMetrics middleware — Datadog APM auto-instruments Gin routes for request count, latency, and error rate.

## Recovery Middleware

- **Does NOT emit custom statsd metric** for panics
- **Logs panic** with full stack trace via structured logger
- **Returns 500** with canonical error envelope
- **Datadog log-based metric** can be created from "panic recovered" log entries
- **DD APM records** the 500 status automatically

## RequestContext Fields

| Field | Source | Required | Purpose |
|-------|--------|----------|---------|
| RequestID | X-Request-ID header | Auto | Correlation across services |
| UserID | Set by handler | No | User identification |
| Platform | x-platform header | Yes | Analytics, risk evaluation |
| DeviceID | x-device-id header | No | Fraud detection |
| UserIP | x-user-ip header | No | Geo checks, audit trail |
| Source | Set by handler | No | PDP, CART, CHECKOUT, CX |

## Health Endpoints

| Endpoint | Type | Checks | K8s Probe |
|----------|------|--------|-----------|
| `GET /health` | Liveness | Process running | `livenessProbe` |
| `GET /health/ready` | Readiness | DB + Redis | `readinessProbe` |

### Health Response Examples

**Liveness (`/health`):**
```json
{
  "status": "ok",
  "service": "payin3-service"
}
```

**Readiness (`/health/ready`) - Healthy:**
```json
{
  "status": "ok",
  "checks": {
    "database": "ok",
    "redis": "ok"
  }
}
```

**Readiness (`/health/ready`) - Degraded:**
```json
{
  "status": "degraded",
  "checks": {
    "database": "ok",
    "redis": "unhealthy: connection refused"
  }
}
```

## Main.go Wiring Order

1. **config.Load()** - Load configuration from file/env
2. **logging.NewLogger()** - Initialize structured logger
3. **metrics.NewDatadogClient()** or **NewNoopClient()** - Business metrics only
4. **pg.NewDB()** + pool settings - GORM database connection
5. **cache.NewRedisProfileCache()** or **NewMemoryProfileCache()** - Profile cache
6. **kafka.NewProducer()** or **NewNoopProducer()** - Event publisher
7. **[placeholder]** Lazypay adapter init — Phase 5
8. **[placeholder]** Domain module init — Phase 6+
9. **gin.New()** + 4 middleware (no metrics middleware)
10. **http.Server ListenAndServe** with graceful shutdown

## What Datadog Agent Auto-Handles

| Category | What | DD Feature |
|----------|------|------------|
| HTTP/API | Request count, latency, errors | dd-trace-go + Gin |
| Database | Query duration, connections | Postgres integration |
| Redis | Commands, latency, hit/miss | Redis integration |
| Provider HTTP | Lazypay call count, latency, errors | dd-trace-go net/http |
| Go Runtime | Goroutines, GC, memory | runtime_metrics |
| Infra | Container CPU, memory, network | DD Agent |

## Request Logging Format

### Production (JSON)
```json
{
  "level": "info",
  "ts": "2026-02-13T02:00:00Z",
  "caller": "middleware/logging.go:45",
  "msg": "request completed",
  "service": "payin3-service",
  "env": "production",
  "requestId": "abc-123",
  "userId": "user-456",
  "method": "POST",
  "path": "/v1/payin3/customer-status",
  "httpStatus": 200,
  "durationMs": 142,
  "bodySize": 1024
}
```

### Error Log (4xx/5xx)
```json
{
  "level": "error",
  "ts": "2026-02-13T02:00:00Z",
  "caller": "middleware/logging.go:45",
  "msg": "request completed",
  "service": "payin3-service",
  "env": "production",
  "requestId": "abc-123",
  "method": "POST",
  "path": "/v1/payin3/customer-status",
  "httpStatus": 500,
  "durationMs": 2500,
  "bodySize": 0
}
```

## Context Header Validation

- **x-platform** is **required** for all non-health endpoints
- Missing x-platform → 400 Bad Request with error envelope
- Health endpoints (`/health`, `/health/ready`) skip context header validation
- RequestContext is stored in both Gin context and Go context for downstream use

## Panic Recovery

When a panic occurs:
1. Stack trace captured via `debug.Stack()`
2. Structured error log with full stack
3. 500 response with canonical error envelope:
   ```json
   {
     "success": false,
     "data": null,
     "error": {
       "code": "PAYIN3_INTERNAL_ERROR",
       "message": "An internal error occurred",
       "statusCode": 500,
       "retryable": true
     }
   }
   ```
4. DD APM records the 500 status automatically
5. Log-based metric can be created from "panic recovered" entries

## Graceful Shutdown

- **Signal handling**: SIGINT, SIGTERM
- **Shutdown timeout**: From config (default 10s)
- **Resource cleanup**: Logger.Sync(), MetricsClient.Close(), Producer.Close(), DB.Close()
- **Order**: Stop accepting requests → Wait for in-flight → Close resources

## Verification Steps

### 1. Build Verification
```bash
go mod tidy
go build ./...
```

### 2. Verify NO Metrics Import in Middleware
```bash
grep -r "metrics\." internal/infrastructure/middleware/
# Should return 0 results
```

### 3. Test Health Endpoints
```bash
# Liveness (always 200)
curl http://localhost:8080/health

# Readiness (checks DB + Redis)
curl http://localhost:8080/health/ready
```

### 4. Test Context Header Validation
```bash
# Missing x-platform → 400
curl -X POST http://localhost:8080/v1/payin3/customer-status \
  -H "Content-Type: application/json" \
  -d '{"userId": "u1", "mobile": "1234567890"}'

# With x-platform → passes (if handler exists)
curl -X POST http://localhost:8080/v1/payin3/customer-status \
  -H "Content-Type: application/json" \
  -H "x-platform: WEB" \
  -d '{"userId": "u1", "mobile": "1234567890"}'
```

### 5. Test Panic Recovery
```go
// In a test handler
router.GET("/test-panic", func(c *gin.Context) {
    panic("test panic")
})
```

**Expected:**
- 500 response with error envelope
- Structured log with stack trace
- No process crash

### 6. Verify Request Logging
```bash
# Make a request and check logs
curl http://localhost:8080/health

# Should see structured log with:
# - requestId
# - method: GET
# - path: /health
# - httpStatus: 200
# - durationMs: <number>
```

## Dependencies on Previous Phases

- **Phase 4A**: Kafka producer, Redis cache (used in main.go)
- **Phase 4B + 4B-FIX**: Logger, business metrics (used in middleware and main.go)
- **Phase 0**: Config structure, Postgres connection

## What Comes Next

### Phase 5: Lazypay Adapter
- HTTP executors (already done in Phase 3A)
- Signature service (already done in Phase 3A)
- Lazypay clients (already done in Phase 3C)
- Wire into main.go

### Phase 6+: Domain Modules
- Wire domain modules into main.go
- Register module routes
- Complete end-to-end flow

## Notes

- **NO metrics middleware**: Datadog APM handles request metrics automatically
- **NO OpenTelemetry**: Distributed tracing not implemented
- **Request context propagation**: Via both Gin context and Go context
- **Health endpoints**: Skip context header validation
- **Panic recovery**: Logs stack trace, returns 500, no custom metric
- **Graceful shutdown**: All resources cleaned up properly
- **Middleware order matters**: RequestID → Recovery → Logging → ContextHeaders
- **Structured logging**: All requests logged with requestId, method, path, status, duration
- **Context headers**: x-platform required for non-health endpoints
- **Redis health check**: Only if Redis is configured (nil if disabled)
