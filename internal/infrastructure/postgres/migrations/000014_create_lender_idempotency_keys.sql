-- Ensure idempotency enum exists (required by lender_idempotency_keys)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'idempotency_status') THEN
        CREATE TYPE idempotency_status AS ENUM (
          'PROCESSING',
          'COMPLETED',
          'FAILED'
        );
    END IF;
END$$;

-- Create idempotency keys table if missing (e.g. DB had 0004 run before this table was added)
CREATE TABLE IF NOT EXISTS lender_idempotency_keys (
  id                BIGSERIAL          PRIMARY KEY,
  idempotency_key   TEXT               NOT NULL UNIQUE,
  request_hash      TEXT               NOT NULL,
  status            idempotency_status NOT NULL DEFAULT 'PROCESSING',
  response_payload  JSONB,
  lender_order_id   TEXT,
  created_at        TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
  expires_at        TIMESTAMPTZ        NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_idempotency_expires
  ON lender_idempotency_keys(expires_at);
