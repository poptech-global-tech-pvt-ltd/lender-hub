-- Phase: Refund module complete — add payment_refund_id (POP) and loan_id

-- Add payment_refund_id (POP's refund reference)
ALTER TABLE lender_refunds
    ADD COLUMN IF NOT EXISTS payment_refund_id TEXT;

-- Add loan_id (order's loanId = merchantTxnId for enquiry)
ALTER TABLE lender_refunds
    ADD COLUMN IF NOT EXISTS loan_id TEXT;

-- Unique: one POP refund reference per lender
CREATE UNIQUE INDEX IF NOT EXISTS uq_lender_refunds_payment_refund_id
    ON lender_refunds(lender, payment_refund_id)
    WHERE payment_refund_id IS NOT NULL AND payment_refund_id != '';

-- Index for GET by paymentRefundId
CREATE INDEX IF NOT EXISTS idx_refund_payment_refund_id
    ON lender_refunds(payment_refund_id)
    WHERE payment_refund_id IS NOT NULL;

-- Index for list by loan_id
CREATE INDEX IF NOT EXISTS idx_refund_loan_id
    ON lender_refunds(loan_id)
    WHERE loan_id IS NOT NULL;
