-- lending-hub-service/internal/infrastructure/postgres/migrations/0003_create_onboarding_tables.sql

-- lender_customer_link
CREATE TABLE IF NOT EXISTS lender_customer_link (
  id                    BIGSERIAL    PRIMARY KEY,
  user_id               TEXT         NOT NULL,
  merchant_id           TEXT         NOT NULL,
  provider              TEXT         NOT NULL,
  mobile                TEXT         NOT NULL,
  latest_onboarding_id  BIGINT,
  created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  UNIQUE (provider, merchant_id, user_id),
  UNIQUE (provider, mobile, merchant_id)
);

-- lender_onboarding
CREATE TABLE IF NOT EXISTS lender_onboarding (
  id                        BIGSERIAL                PRIMARY KEY,
  onboarding_id             TEXT                     NOT NULL,
  provider_onboarding_id    TEXT,
  user_id                   TEXT                     NOT NULL,
  merchant_id               TEXT                     NOT NULL,
  provider                  TEXT                     NOT NULL,
  mobile                    TEXT                     NOT NULL,
  source                    request_source           NOT NULL,
  channel                   channel_type,
  status                    lender_onboarding_status NOT NULL DEFAULT 'PENDING',
  last_step                 onboarding_step,
  rejection_reason_code     TEXT,
  rejection_reason_message  TEXT,
  cof_eligible              BOOLEAN                  DEFAULT false,
  redirect_url              TEXT,
  is_retryable              BOOLEAN                  NOT NULL DEFAULT false,
  retry_count               INT                      NOT NULL DEFAULT 0,
  next_retry_at             TIMESTAMPTZ,
  last_retry_at             TIMESTAMPTZ,
  raw_request               JSONB,
  raw_response              JSONB,
  created_at                TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
  updated_at                TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
  UNIQUE (provider, merchant_id, user_id, onboarding_id)
);

CREATE INDEX IF NOT EXISTS idx_onboarding_user
  ON lender_onboarding(user_id);

CREATE INDEX IF NOT EXISTS idx_onboarding_status
  ON lender_onboarding(status);

CREATE INDEX IF NOT EXISTS idx_onboarding_retry
  ON lender_onboarding(status, is_retryable, next_retry_at)
  WHERE status = 'FAILED' AND is_retryable = true;

-- lender_onboarding_events
CREATE TABLE IF NOT EXISTS lender_onboarding_events (
  id                BIGSERIAL    PRIMARY KEY,
  provider          TEXT         NOT NULL,
  merchant_id       TEXT         NOT NULL,
  user_id           TEXT         NOT NULL,
  mobile            TEXT         NOT NULL,
  onboarding_id     TEXT         NOT NULL,
  event_type        TEXT         NOT NULL,
  status            TEXT         NOT NULL,
  step              onboarding_step,
  error_code        TEXT,
  message           TEXT,
  event_time        TIMESTAMPTZ  NOT NULL,
  raw_payload       JSONB,
  created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  UNIQUE (provider, mobile, onboarding_id, event_time, step)
);

CREATE INDEX IF NOT EXISTS idx_onboarding_events_lookup
  ON lender_onboarding_events(provider, onboarding_id);
