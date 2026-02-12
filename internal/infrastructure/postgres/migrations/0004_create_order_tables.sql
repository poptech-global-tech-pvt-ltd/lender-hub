-- lending-hub-service/internal/infrastructure/postgres/migrations/0004_create_order_tables.sql

-- lender_payment_state
CREATE TABLE IF NOT EXISTS lender_payment_state (
  id                        BIGSERIAL              PRIMARY KEY,
  payment_id                TEXT                   NOT NULL UNIQUE,
  user_id                   TEXT                   NOT NULL,
  merchant_id               TEXT                   NOT NULL,
  lender                    TEXT                   NOT NULL,
  amount                    NUMERIC(14,2)          NOT NULL,
  currency                  TEXT                   NOT NULL DEFAULT 'INR',
  status                    lender_payment_status  NOT NULL DEFAULT 'PENDING',
  source                    request_source,
  return_url                TEXT,
  emi_plan                  JSONB,
  lender_order_id           TEXT,
  lender_merchant_txn_id    TEXT,
  lender_last_status        TEXT,
  lender_last_txn_id        TEXT,
  lender_last_txn_status    TEXT,
  lender_last_txn_message   TEXT,
  lender_last_txn_time      TIMESTAMPTZ,
  last_error_code           TEXT,
  last_error_message        TEXT,
  created_at                TIMESTAMPTZ            NOT NULL DEFAULT NOW(),
  updated_at                TIMESTAMPTZ            NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_user
  ON lender_payment_state(user_id);

CREATE INDEX IF NOT EXISTS idx_payment_status
  ON lender_payment_state(status);

CREATE INDEX IF NOT EXISTS idx_payment_created
  ON lender_payment_state(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_lender_order
  ON lender_payment_state(lender_order_id);

-- lender_payment_mapping
CREATE TABLE IF NOT EXISTS lender_payment_mapping (
  id                        BIGSERIAL    PRIMARY KEY,
  payment_id                TEXT         NOT NULL,
  user_id                   TEXT         NOT NULL,
  lender                    TEXT         NOT NULL,
  lender_merchant_txn_id    TEXT         NOT NULL,
  lender_order_id           TEXT,
  eligibility_response_id   TEXT,
  created_at                TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at                TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  UNIQUE (payment_id, lender)
);

CREATE INDEX IF NOT EXISTS idx_mapping_merchant_txn
  ON lender_payment_mapping(lender_merchant_txn_id);

-- lender_idempotency_keys
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
