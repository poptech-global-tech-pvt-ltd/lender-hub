# Phase: Kafka Complete — Producers + Consumers + DLQ + Graceful Shutdown

## Summary

- **Topics**: All from `config.Kafka.Topics` — no hardcoded topic strings
- **Event schema**: `BaseEvent` with `version: "v1"` in all payloads
- **Partition key**: `loanId` for orders, `refundId` for refunds
- **DLQ**: Both consumers send failed messages to DLQ topics
- **Retry**: Exponential backoff (100ms, 400ms, 900ms), max 3 retries
- **Shutdown**: signal.NotifyContext → HTTP shutdown → close readers → wait consumers → close writers

## Topics (Produced)

| Topic | Event | Key |
|-------|-------|-----|
| lsp.order.created | OrderCreatedEvent | loanId |
| lsp.order.status_updated | OrderStatusUpdatedEvent | loanId |
| lsp.order.support_updated | OrderSupportUpdatedEvent | loanId |
| lsp.refund.created | RefundCreatedEvent | refundId |
| lsp.refund.status_updated | RefundStatusUpdatedEvent | refundId |

## Topics (Consumed)

| Topic | Event | Consumer Group |
|-------|-------|----------------|
| lsp.lazypay.order.callback | OrderCallbackEvent | lending-hub.order-callback |
| lsp.lazypay.refund.callback | RefundCallbackEvent | lending-hub.refund-callback |

## DLQ Topics

| Original Topic | DLQ Topic |
|----------------|-----------|
| lsp.lazypay.order.callback | lsp.lazypay.order.callback.dlq |
| lsp.lazypay.refund.callback | lsp.lazypay.refund.callback.dlq |

## Files Created/Modified

### Config
- `config/config.go` — KafkaTopics, KafkaConsumerGroups, KafkaProducerConfig, KafkaConsumerConfig
- `config/config.yaml` — full Kafka section

### Infrastructure
- `internal/infrastructure/kafka/writer.go` — TopicWriter
- `internal/infrastructure/kafka/reader.go` — TopicReader
- `internal/infrastructure/kafka/events/base.go` — BaseEvent
- `internal/infrastructure/kafka/events/orderevents.go`
- `internal/infrastructure/kafka/events/refundevents.go`
- `internal/infrastructure/kafka/events/callbackevents.go`
- `internal/infrastructure/kafka/orderevent_publisher.go`
- `internal/infrastructure/kafka/refundevent_publisher.go`
- `internal/infrastructure/kafka/consumer/order_callback.go`
- `internal/infrastructure/kafka/consumer/refund_callback.go`

### Domain
- `internal/domain/order/port/event_publisher.go` — PublishOrderCreated, PublishOrderStatusUpdated, PublishOrderSupportUpdated
- `internal/domain/refund/port/event_publisher.go` — PublishRefundCreated, PublishRefundStatusUpdated
- `internal/domain/refund/dto/request/refund_callback.go`
- `internal/domain/refund/service/refund_service.go` — ProcessRefundCallback

### Main
- `cmd/server/main.go` — order/refund publishers, consumers, graceful shutdown

## Verification

```bash
go build ./...   # passes
go vet ./...     # passes
```
