# Starting the Lending Hub Service

## Prerequisites
- Go 1.23+ installed
- PostgreSQL database running
- (Optional) Redis for caching
- (Optional) Kafka for event streaming

## Configuration

### Option 1: Update `config/config.yaml`
Edit `config/config.yaml` with your database credentials:

```yaml
db:
  host: "your-db-host"
  port: 5432
  user: "your-db-user"
  password: "your-db-password"
  name: "your-db-name"
  sslmode: "require"  # or "disable" for local
```

### Option 2: Use Environment Variables
All config values can be overridden with environment variables prefixed with `LENDING_HUB_`:

```bash
export LENDING_HUB_DB_HOST="your-db-host"
export LENDING_HUB_DB_USER="your-db-user"
export LENDING_HUB_DB_PASSWORD="your-db-password"
export LENDING_HUB_DB_NAME="your-db-name"
export LENDING_HUB_DB_SSLMODE="require"
```

## Database Setup

1. **Run migrations** (if not already done):
   ```bash
   # Connect to your database and run the migration files in order:
   # internal/infrastructure/postgres/migrations/0001_create_enums.sql
   # internal/infrastructure/postgres/migrations/0002_create_lender_user.sql
   # internal/infrastructure/postgres/migrations/0003_create_onboarding_tables.sql
   # internal/infrastructure/postgres/migrations/0004_create_order_tables.sql
   # internal/infrastructure/postgres/migrations/0005_create_lender_refunds.sql
   # internal/infrastructure/postgres/migrations/0006_create_lender_error_code_config.sql
   ```

## Starting the Server

### Build and Run
```bash
# Build
go build -o bin/server ./cmd/server

# Run
./bin/server

# Or run directly
go run ./cmd/server/main.go
```

### With Custom Config
```bash
go run ./cmd/server/main.go -config /path/to/config.yaml
```

## Testing APIs

The server starts on port `8080` by default (configurable via `http.port`).

### Health Check
```bash
curl http://localhost:8080/health
curl http://localhost:8080/health/ready
```

### API Endpoints

All API endpoints require these headers:
- `x-platform`: Platform identifier (e.g., "WEB", "ANDROID", "IOS")
- `x-user-ip`: User IP address

#### 1. Customer Status (Profile)
```bash
curl -X POST http://localhost:8080/v1/payin3/customer-status \
  -H "Content-Type: application/json" \
  -H "x-platform: WEB" \
  -H "x-user-ip: 127.0.0.1" \
  -d '{
    "userId": "user123",
    "lender": "lazypay"
  }'
```

#### 2. Start Onboarding
```bash
curl -X POST http://localhost:8080/v1/payin3/onboarding \
  -H "Content-Type: application/json" \
  -H "x-platform: WEB" \
  -H "x-user-ip: 127.0.0.1" \
  -d '{
    "userId": "user123",
    "lender": "lazypay",
    "phone": "+919876543210",
    "email": "user@example.com"
  }'
```

#### 3. Get Onboarding Status
```bash
curl -X GET "http://localhost:8080/v1/payin3/onboarding/status?onboardingId=ONB_xxx" \
  -H "x-platform: WEB" \
  -H "x-user-ip: 127.0.0.1"
```

#### 4. Create Order
```bash
curl -X POST http://localhost:8080/v1/payin3/order \
  -H "Content-Type: application/json" \
  -H "x-platform: WEB" \
  -H "x-user-ip: 127.0.0.1" \
  -d '{
    "userId": "user123",
    "amount": 1000.00,
    "currency": "INR",
    "merchantTxnId": "merchant-txn-123",
    "emiSelection": {
      "tenure": 3,
      "interestRate": 0.0
    }
  }'
```

#### 5. Get Order Status
```bash
curl -X GET "http://localhost:8080/v1/payin3/order/PAY_xxx" \
  -H "x-platform: WEB" \
  -H "x-user-ip: 127.0.0.1"
```

#### 6. Create Refund
```bash
curl -X POST http://localhost:8080/v1/payin3/refund \
  -H "Content-Type: application/json" \
  -H "x-platform: WEB" \
  -H "x-user-ip: 127.0.0.1" \
  -d '{
    "paymentId": "PAY_xxx",
    "amount": 500.00,
    "reason": "CUSTOMER_REQUEST"
  }'
```

## Configuration Options

### Lazypay Integration
To enable real Lazypay integration (instead of stubs):
```yaml
lazypay:
  enabled: true
  base_url: "https://sandbox.lazypay.in"
  access_key: "your-access-key"
  secret_key: "your-secret-key"
```

### Redis Cache
To enable Redis caching:
```yaml
redis:
  addr: "localhost:6379"
  password: ""
  db: 0
```

### Kafka Events
To enable Kafka event publishing:
```yaml
kafka:
  enabled: true
  brokers:
    - "localhost:9092"
```

## Troubleshooting

1. **Database connection errors**: Check your database credentials and network connectivity
2. **Missing migrations**: Ensure all migration files have been run
3. **Port already in use**: Change `http.port` in config or kill the process using port 8080
4. **Missing headers**: All API requests (except health) require `x-platform` and `x-user-ip` headers
