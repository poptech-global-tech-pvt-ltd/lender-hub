-- Phase 11: Rollback refund enquiry states migration

-- Drop indexes
DROP INDEX IF EXISTS idx_refund_provider_txn;
DROP INDEX IF EXISTS uq_lender_refund_refid;

-- Remove columns (Note: Cannot remove enum values, but columns can be dropped)
ALTER TABLE lender_refunds
    DROP COLUMN IF EXISTS provider_merchant_txn_id,
    DROP COLUMN IF EXISTS provider_parent_txn_id,
    DROP COLUMN IF EXISTS provider_refund_txn_id,
    DROP COLUMN IF EXISTS provider_refund_ref_id,
    DROP COLUMN IF EXISTS last_enquired_at;

-- Note: PostgreSQL does not support removing enum values, so PROCESSING and UNKNOWN
-- will remain in the enum type. This is acceptable as they won't be used after rollback.
