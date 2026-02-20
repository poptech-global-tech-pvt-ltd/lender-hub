# Phase: Order Module — Complete

## Summary

This phase implements the full order flow per spec:

- **ID generation**: `lps_`, `ref_`, `onb_` prefixes via `pkg/idgen`
- **Create order**: paymentId always server-generated; idempotency by request hash
- **Get order**: `GET /v1/payin3/order/:paymentId` — returns status, emiPlan, etc.
- **List orders**: `GET /v1/payin3/orders?userId=...&merchantId=...&status=...&page=1&perPage=20`
- **Support override**: `PATCH /v1/payin3/order/:paymentId/status` — requires `X-Internal-Token`
- **Callbacks**: via Kafka (no POST /callback/order)

## ID Generation Rules

| Prefix | Use |
|--------|-----|
| lps | lenderpaymentstate — primary order identifier |
| ref | refund id |
| onb | onboarding id |

**Rule**: paymentId is NEVER accepted from request body. Always server-generated.

## Status Mapping

| Canonical | Description |
|-----------|-------------|
| PENDING | Initial / in progress |
| SUCCESS | Completed |
| FAILED | Failed |
| REFUNDED | Fully refunded |
| EXPIRED | Expired |
| CANCELLED | Cancelled |

**Rule**: status is NEVER empty in responses. Default: PENDING.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | /v1/payin3/order | Create order (returns lps_xxx) |
| GET | /v1/payin3/order/:paymentId | Get order status |
| GET | /v1/payin3/orders | List orders by userId |
| PATCH | /v1/payin3/order/:paymentId/status | Support status override (FAILED/CANCELLED) |

## Support Transitions

- Allowed: PENDING → FAILED, PENDING → CANCELLED
- Requires: `X-Internal-Token` header matching `config.internal_api_token`
- Body: `{ "status": "FAILED"|"CANCELLED", "reason": "...", "actor": "..." }`

## Create Order Request (new format)

```json
{
  "userId": "019c6ff1-0898-7137-b7f6-2dbbfe156336",
  "merchantId": "270",
  "amount": 1000,
  "currency": "INR",
  "source": "CHECKOUT",
  "returnUrl": "https://yourapp.com/callback",
  "emiPlan": { "tenure": 3 }
}
```

Legacy: `emiSelection: { "tenure": 3, "type": "PAY_IN_PARTS" }` also supported. `paymentId` ignored if sent.

## Curl Examples

```bash
# Create order (new format)
curl -X POST http://localhost:8080/v1/payin3/order \
  -H "Content-Type: application/json" \
  -H "x-platform: website" \
  -H "x-user-ip: 127.0.0.1" \
  -d '{
    "userId": "019c6ff1-0898-7137-b7f6-2dbbfe156336",
    "merchantId": "270",
    "amount": 1000,
    "currency": "INR",
    "source": "CHECKOUT",
    "returnUrl": "https://yourapp.com/callback",
    "emiPlan": { "tenure": 3 }
  }'

# Get order
curl -X GET "http://localhost:8080/v1/payin3/order/lps_01JN..." \
  -H "x-platform: website" \
  -H "x-user-ip: 127.0.0.1"

# List orders
curl -X GET "http://localhost:8080/v1/payin3/orders?userId=019c6ff1-0898-7137-b7f6-2dbbfe156336&page=1&perPage=20" \
  -H "x-platform: website" \
  -H "x-user-ip: 127.0.0.1"

# Support override (requires X-Internal-Token)
curl -X PATCH "http://localhost:8080/v1/payin3/order/lps_01JN.../status" \
  -H "Content-Type: application/json" \
  -H "X-Internal-Token: your-internal-token" \
  -H "x-platform: website" \
  -H "x-user-ip: 127.0.0.1" \
  -d '{"status":"FAILED","reason":"test","actor":"support@pop.in"}'
```

## Config

Add to `config/config.yaml`:

```yaml
internal_api_token: "your-secret-token"  # For PATCH order status
```
