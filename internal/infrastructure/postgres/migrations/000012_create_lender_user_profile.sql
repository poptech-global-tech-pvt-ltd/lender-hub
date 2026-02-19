-- Create lender_user_profile table for caching user contact info
CREATE TABLE IF NOT EXISTS lender_user_profile (
    id              BIGSERIAL       PRIMARY KEY,
    user_id         TEXT            NOT NULL UNIQUE,  -- global_user_id from caller
    mobile          TEXT            NOT NULL,          -- 10 digit, no prefix
    email           TEXT,                              -- nullable
    raw_phone       TEXT,                              -- original "+919686019629" for audit
    source          TEXT            NOT NULL DEFAULT 'PROFILE_SERVICE', -- PROFILE_SERVICE | ONBOARDING | ORDER
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lender_user_profile_mobile ON lender_user_profile(mobile);
CREATE INDEX IF NOT EXISTS idx_lender_user_profile_user_id ON lender_user_profile(user_id);
