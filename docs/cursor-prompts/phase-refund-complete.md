# Phase: Refund Module — Complete Implementation

## ID Mapping Table

| Field | Source | Description |
|-------|--------|-------------|
| `payment_refund_id` | Caller (POP) | POP's refund reference, e.g. "POP_REF_001" |
| `refund_id` | Server (idgen) | Our generated ID, e.g. "ref_01JN..." — also used as refundTxnId to Lazypay |
| `provider_refund_txnid` | Lazypay | Lazypay's lpTxnId for REFUND transaction, e.g. "TXN1738710" |
| `payment_id` | Order | POP's order paymentId |
| `loan_id` | Order | Our order loanId = merchantTxnId for enquiry |

## State Machine

- **PENDING** → SUCCESS | FAILED | UNKNOWN | PROCESSING
- **PROCESSING** → SUCCESS | FAILED
- **UNKNOWN** → SUCCESS | FAILED | PROCESSING (via enquiry)
- **SUCCESS** | **FAILED** — terminal

## Enquiry Matching Logic

1. Use order's `loanId` as `merchantTxnId` for GET `/lazypay/v3/enquiry`
2. Find transactions with `txnType=REFUND`
3. Primary match: `txnRefNo == refund.RefundID` (our generated ID)
4. Fallback: `txnRefNo` absent and `txn.LpTxnID == refund.ProviderRefundTxnID`

## No-Retry Rule

- Lazypay refund API: **single attempt, no retry**
- Timeout → status UNKNOWN → enquiry resolves
- LPDUPLICATEREFUND → enquiry called automatically

## Config

- `lazypay.refund_enquiry_sla` — after this duration, refund not found in enquiry → FAILED (default: 1h)

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | /v1/payin3/refund | Create refund |
| GET | /v1/payin3/refund/:paymentRefundId | Get by POP's paymentRefundId |
| GET | /v1/payin3/refund/loan/:refundId | Get by our refundId |
| GET | /v1/payin3/refunds?paymentId=... | List refunds for order |
| GET | /v1/payin3/refunds/user?userId=...&page=1&perPage=20 | List refunds for user |

## Curl Examples

### Create refund
```bash
curl -X POST http://localhost:8080/v1/payin3/refund \
  -H "Content-Type: application/json" \
  -d '{
    "paymentId": "POP_PAY_001",
    "paymentRefundId": "POP_REF_001",
    "amount": 1500.00,
    "currency": "INR",
    "reason": "USER_CANCELLED"
  }'
```

### Get by paymentRefundId
```bash
curl http://localhost:8080/v1/payin3/refund/POP_REF_001
```

### Get by refundId
```bash
curl http://localhost:8080/v1/payin3/refund/loan/ref_01JN...
```

### List for order
```bash
curl "http://localhost:8080/v1/payin3/refunds?paymentId=POP_PAY_001"
```

### List by user
```bash
curl "http://localhost:8080/v1/payin3/refunds/user?userId=u_001&page=1&perPage=20"
```
