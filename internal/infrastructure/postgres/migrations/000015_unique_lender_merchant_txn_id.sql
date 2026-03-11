-- Ensure lender_merchant_txn_id (loanId) is unique for GET by loanId
CREATE UNIQUE INDEX IF NOT EXISTS idx_lender_payment_state_merchant_txn
  ON lender_payment_state(lender_merchant_txn_id)
  WHERE lender_merchant_txn_id IS NOT NULL AND lender_merchant_txn_id != '';
