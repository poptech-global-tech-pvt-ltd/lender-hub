-- Phase 11: Add PROCESSING and UNKNOWN states to lender_refund_status enum
-- Also add provider fields and last_enquired_at to lender_refunds table

-- Update enum to add PROCESSING and UNKNOWN
DO $$
BEGIN
    -- Check if PROCESSING exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'PROCESSING' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'lender_refund_status')
    ) THEN
        ALTER TYPE lender_refund_status ADD VALUE 'PROCESSING';
    END IF;

    -- Check if UNKNOWN exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'UNKNOWN' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'lender_refund_status')
    ) THEN
        ALTER TYPE lender_refund_status ADD VALUE 'UNKNOWN';
    END IF;
END$$;

-- Add new columns to lender_refunds table
ALTER TABLE lender_refunds
    ADD COLUMN IF NOT EXISTS provider_merchant_txn_id TEXT,
    ADD COLUMN IF NOT EXISTS provider_parent_txn_id TEXT,
    ADD COLUMN IF NOT EXISTS provider_refund_txn_id TEXT,
    ADD COLUMN IF NOT EXISTS provider_refund_ref_id TEXT,
    ADD COLUMN IF NOT EXISTS last_enquired_at TIMESTAMPTZ;

-- Create unique constraint on lender + provider_refund_ref_id
CREATE UNIQUE INDEX IF NOT EXISTS uq_lender_refund_refid
    ON lender_refunds(lender, provider_refund_ref_id)
    WHERE provider_refund_ref_id IS NOT NULL;

-- Create index on provider_refund_txn_id for enquiry lookups
CREATE INDEX IF NOT EXISTS idx_refund_provider_txn
    ON lender_refunds(provider_refund_txn_id)
    WHERE provider_refund_txn_id IS NOT NULL;
