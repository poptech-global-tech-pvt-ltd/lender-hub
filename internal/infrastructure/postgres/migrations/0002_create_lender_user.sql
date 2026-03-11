-- lending-hub-service/internal/infrastructure/postgres/migrations/0002_create_lender_user.sql

CREATE TABLE IF NOT EXISTS lender_user (
  id                    BIGSERIAL                PRIMARY KEY,
  user_id               TEXT                     NOT NULL,
  lender                TEXT                     NOT NULL,
  current_status        lender_profile_status    NOT NULL DEFAULT 'NOT_STARTED',
  onboarding_done       BOOLEAN                  DEFAULT false,
  ntb_status            BOOLEAN,
  credit_limit          NUMERIC(14,2),
  available_limit       NUMERIC(14,2),
  credit_line_active    BOOLEAN                  NOT NULL DEFAULT false,
  credit_line_summary   JSONB,
  is_blocked            BOOLEAN                  DEFAULT false,
  block_reason          TEXT,
  block_source          TEXT,
  next_eligible_at      TIMESTAMPTZ,
  last_onboarding_id    BIGINT,
  last_limit_refresh_at TIMESTAMPTZ,
  created_at            TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, lender)
);

CREATE INDEX IF NOT EXISTS idx_lender_user_status
  ON lender_user(current_status);

CREATE INDEX IF NOT EXISTS idx_lender_user_blocked
  ON lender_user(is_blocked);
