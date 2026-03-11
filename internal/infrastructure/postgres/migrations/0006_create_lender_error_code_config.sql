-- lending-hub-service/internal/infrastructure/postgres/migrations/0006_create_lender_error_code_config.sql

CREATE TABLE IF NOT EXISTS lender_error_code_config (
  id                      BIGSERIAL    PRIMARY KEY,
  provider                TEXT         NOT NULL,
  canonical_error_code    TEXT         NOT NULL,
  is_retryable            BOOLEAN      NOT NULL,
  canonical_status        TEXT         NOT NULL,
  initial_delay_seconds   INT,
  max_delay_seconds       INT,
  max_retries             INT,
  http_status_code        INT,
  user_message            TEXT,
  description             TEXT,
  created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  UNIQUE (provider, canonical_error_code)
);
