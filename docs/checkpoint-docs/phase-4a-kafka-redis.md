# Phase 4A: Kafka Producer/Consumer + Redis Cache

**Date:** February 13, 2025  
**Status:** ✅ Complete

## Overview

This phase implements production-ready event publishing with Kafka and Redis caching for profile eligibility. Both components have noop/in-memory fallbacks for local development, ensuring the application works without external dependencies.

## Files Created (12 total)

### Kafka Infrastructure (6 files)
1. `internal/infrastructure/kafka/config.go` - Producer and consumer configuration
2. `internal/infrastructure/kafka/producer.go` - Kafka producer with batching and compression
3. `internal/infrastructure/kafka/consumer.go` - Kafka consumer with message handler
4. `internal/infrastructure/kafka/topics.go` - Topic constants
5. `internal/infrastructure/kafka/event.go` - Domain event envelope and event data types
6. `internal/infrastructure/kafka/noop_producer.go` - Noop producer for local dev

### Event Publisher Adapters (2 files)
7. `internal/infrastructure/kafka/profile_publisher.go` - Profile event publisher adapter
8. `internal/infrastructure/kafka/order_publisher.go` - Order event publisher adapter

### Cache Infrastructure (4 files)
9. `internal/infrastructure/cache/config.go` - Redis configuration
10. `internal/infrastructure/cache/redis_cache.go` - Redis-backed profile cache
11. `internal/infrastructure/cache/memory_cache.go` - In-memory profile cache with TTL
12. `internal/infrastructure/cache/interface.go` - Generic cache interface

### Integration
13. `cmd/server/main.go` - Updated with cache and event publisher initialization

## Kafka Topics

| Topic | Event Types | Key Strategy | Purpose |
|-------|------------|--------------|---------|
| `payin3.profile.events` | ProfileActivated, ProfileBlocked, ProfileUnblocked, LimitUpdated, StatusChanged | `userID:lender` | Profile state changes |
| `payin3.order.events` | OrderCreated, OrderCompleted, OrderFailed, OrderRefunded | `paymentID` | Order lifecycle events |
| `payin3.refund.events` | (Future) | `refundID` | Refund lifecycle events |

## Redis Caching Strategy

### Key Format
```
payin3:profile:{userID}:{lender}
```

### TTL
- **Profile Eligibility**: 60 seconds
- **Rationale**: Short TTL to balance freshness and performance

### Cache-Aside Pattern Flow

```
GetCustomerStatus Request
    ↓
Cache.Get(userID, lender)
    ↓
┌─────────────────────────┐
│ Cache Hit?              │
├─────────────────────────┤
│ YES → Return cached     │
│ NO → Continue          │
└─────────────────────────┘
    ↓
Gateway.CheckEligibility()
    ↓
Cache.Set(userID, lender, response)
    ↓
Return response
```

### Cache Invalidation Triggers

1. **Onboarding Completion** → `Invalidate(userID, lender)`
   - Profile status changes from IN_PROGRESS → ACTIVE
   - Credit limit may have changed

2. **Order Success** → `Invalidate(userID, lender)`
   - Available limit decreased
   - Need fresh eligibility data

3. **User Block** → `Invalidate(userID, lender)`
   - Status changes to BLOCKED
   - Eligibility response changes

4. **Limit Update** → `Invalidate(userID, lender)`
   - Available limit manually updated
   - Credit limit changed

## Fallback Strategy

### Kafka Fallback
- **No Brokers Configured**: Uses `NoopProducer`
- **Connection Failure**: Falls back to `NoopProducer` with warning log
- **Behavior**: Events logged at DEBUG level, no actual publishing

### Redis Fallback
- **No Address Configured**: Uses `MemoryProfileCache`
- **Connection Failure**: Falls back to `MemoryProfileCache` with warning log
- **Behavior**: In-memory cache with 60s TTL, background cleanup every 30s

## Event Publishing Flow

```
Domain Service
  ↓
EventPublisher.Publish(event)
  ↓
Kafka Publisher Adapter
  ↓
DomainEvent Envelope
  ├─ EventID (UUID)
  ├─ EventType
  ├─ Source: "payin3-service"
  ├─ OccurredAt
  ├─ UserID, Lender
  └─ Data (event-specific payload)
  ↓
Kafka Producer
  ├─ Topic: payin3.{module}.events
  ├─ Key: userID or paymentID (for partitioning)
  └─ Value: JSON-encoded DomainEvent
  ↓
Kafka Broker
```

## Key Design Decisions

### 1. Kafka Producer Configuration
- **Async**: `true` - Non-blocking publishes
- **Compression**: `snappy` - Good balance of speed and size
- **BatchSize**: `100` - Messages per batch
- **LingerMs**: `5ms` - Batch wait time
- **RequiredAcks**: `all` - Highest durability
- **Retries**: `3` - Automatic retry on failure

### 2. Kafka Consumer Configuration
- **GroupID**: `payin3-service` - Consumer group
- **AutoCommit**: `true` - Automatic offset commits
- **OffsetReset**: `earliest` - Start from beginning if no offset
- **At-Least-Once**: Errors logged but processing continues

### 3. Redis Cache Configuration
- **PoolSize**: `25` - Connection pool size
- **MinIdleConns**: `5` - Minimum idle connections
- **TTL**: `60s` - Profile eligibility cache TTL
- **Key Prefix**: `payin3:profile:` - Namespace isolation

### 4. Memory Cache Implementation
- **sync.Map**: Thread-safe map for concurrent access
- **TTL-Based Expiry**: Background cleanup every 30s
- **No Persistence**: Lost on restart (acceptable for local dev)

### 5. Event Envelope Design
- **Standard Format**: All events use `DomainEvent` envelope
- **EventID**: UUID for deduplication
- **Source**: Service identifier
- **OccurredAt**: Timestamp for ordering
- **Data**: Type-specific payload (ProfileEventData, OrderEventData, etc.)

### 6. Partitioning Strategy
- **Profile Events**: Key = `userID:lender` - Ensures user events in same partition
- **Order Events**: Key = `paymentID` - Ensures order events in same partition
- **Refund Events**: Key = `refundID` - Ensures refund events in same partition

## Verification Steps

### 1. Build Verification
```bash
go mod tidy
go build ./...
```

### 2. Start Server WITHOUT Kafka/Redis Config
```bash
# No kafka/redis config in config.yaml
go run cmd/server/main.go
```

**Expected Output:**
```
Using memory cache (no Redis config)
Using noop event publishers (no Kafka config)
Using stub gateways (no Lazypay config)
Server listening on :8080
```

**Verification:**
- All endpoints work
- Cache operations succeed (in-memory)
- Events logged but not published
- No external dependencies required

### 3. Start Server WITH Redis (No Kafka)
```yaml
# config.yaml
redis:
  addr: "localhost:6379"
```

```bash
go run cmd/server/main.go
```

**Expected Output:**
```
Using Redis cache
Using noop event publishers (no Kafka config)
Server listening on :8080
```

**Verification:**
- Redis connection established
- Cache SET/GET/DEL operations work
- Test with: `redis-cli GET "payin3:profile:u1:LAZYPAY"`

### 4. Start Server WITH Kafka (No Redis)
```yaml
# config.yaml
kafka:
  enabled: true
  brokers:
    - "localhost:9092"
```

```bash
go run cmd/server/main.go
```

**Expected Output:**
```
Using memory cache (no Redis config)
Using Kafka event publishers
Server listening on :8080
```

**Verification:**
- Kafka producer initialized
- Events published to topics
- Check with: `kafka-console-consumer --bootstrap-server localhost:9092 --topic payin3.profile.events`

### 5. Verify Interface Compliance
```bash
go build ./internal/infrastructure/kafka/...
go build ./internal/infrastructure/cache/...
```

**Verification:**
- `ProfileEventPublisher` satisfies `profilePort.ProfileEventPublisher`
- `OrderEventPublisher` satisfies `orderPort.OrderEventPublisher`
- `RedisProfileCache` satisfies `profilePort.ProfileCache`
- `MemoryProfileCache` satisfies `profilePort.ProfileCache`

## Cache Invalidation Events

### Onboarding Completion
```go
// In OnboardingService.ProcessCallback
if newStatus == OnboardingSuccess {
    profileUpdater.UpdateOnOnboardingSuccess(...)
    // ProfileUpdater should invalidate cache
    cache.Invalidate(ctx, userID, lender)
}
```

### Order Success
```go
// In OrderService.ProcessCallback
if status == OrderSuccess {
    profileUpdater.UpdateLimit(...)
    // ProfileUpdater should invalidate cache
    cache.Invalidate(ctx, userID, lender)
}
```

### User Block/Unblock
```go
// In ProfileUpdater.BlockUser/UnblockUser
cache.Invalidate(ctx, userID, lender)
```

## Dependencies on Previous Phases

- **Phase 2A**: Profile module with `ProfileEventPublisher` and `ProfileCache` interfaces
- **Phase 2C**: Order module with `OrderEventPublisher` interface
- **Phase 3A-3C**: Lazypay adapter (no direct dependency, but used together)

## What Comes Next

### Phase 4B: Observability
- Metrics (Prometheus)
- Distributed tracing (OpenTelemetry)
- Structured logging
- Health check enhancements
- Performance monitoring

### Future Enhancements
- Kafka consumer for cross-service events
- Redis pub/sub for cache invalidation
- Event replay capabilities
- Dead letter queue handling
- Cache warming strategies
- Event schema versioning

## Notes

- All components support graceful shutdown via `Close()` method
- Kafka producer uses async publishing for non-blocking operation
- Redis cache uses connection pooling for efficiency
- Memory cache uses background cleanup to prevent memory leaks
- Event publishers wrap domain events in standard `DomainEvent` envelope
- Cache invalidation is explicit (not automatic) - services must call `Invalidate`
- Fallback to noop/memory ensures local development works without external services
- Kafka compression reduces network bandwidth and storage
- Event keys ensure proper partitioning for ordering guarantees
