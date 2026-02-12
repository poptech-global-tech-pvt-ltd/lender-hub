-- lending-hub-service/internal/infrastructure/postgres/migrations/0005_create_lender_refunds.sql

CREATE TABLE IF NOT EXISTS lender_refunds (
  id                BIGSERIAL            PRIMARY KEY,
  refund_id         TEXT                 NOT NULL UNIQUE,
  payment_id        TEXT                 NOT NULL,
  user_id           TEXT                 NOT NULL,
  lender            TEXT                 NOT NULL,
  amount            NUMERIC(14,2)        NOT NULL,
  currency          TEXT                 NOT NULL DEFAULT 'INR',
  status            lender_refund_status NOT NULL DEFAULT 'PENDING',
  reason            refund_reason,
  lender_ref_id     TEXT,
  lender_status     TEXT,
  lender_message    TEXT,
  created_at        TIMESTAMPTZ          NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ          NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refund_payment
  ON lender_refunds(payment_id);

CREATE INDEX IF NOT EXISTS idx_refund_status
  ON lender_refunds(status);

CREATE INDEX IF NOT EXISTS idx_refund_user
  ON lender_refunds(user_id);
