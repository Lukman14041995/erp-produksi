-- ============================================================================
-- Customer-facing payment flow: after a quotation is confirmed into a Sales
-- Order (+ invoice), the customer picks Lunas or Termin (1-3x) on their
-- order-link page, transfers to a bank account we display, uploads proof,
-- and staff verifies it -- which posts a real finance.Payment (journal +
-- AR allocation) exactly like a manually-entered payment would.
-- ============================================================================

-- Bank accounts customers are shown as the transfer destination. Linked to
-- a COA account so a confirmed installment posts to the correct ledger
-- account (which real bank the money lands in).
CREATE TABLE bank_accounts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_name         VARCHAR(100) NOT NULL,
    account_number    VARCHAR(50) NOT NULL,
    account_holder    VARCHAR(150) NOT NULL,
    coa_account_code  VARCHAR(20) NOT NULL REFERENCES accounts(code),
    sort_order        INTEGER NOT NULL DEFAULT 0,
    is_active         BOOLEAN NOT NULL DEFAULT true,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_bank_accounts_updated_at BEFORE UPDATE ON bank_accounts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- One plan per invoice: the customer's one-time choice of Lunas (1x) vs
-- Termin (2-3x). installment_count also covers "Lunas" as a 1-count plan,
-- so there's a single mental model for "how many installments" downstream.
CREATE TABLE payment_plans (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id        UUID NOT NULL UNIQUE REFERENCES invoices(id),
    method            VARCHAR(20) NOT NULL CHECK (method IN ('FULL','INSTALLMENT')),
    installment_count INTEGER NOT NULL CHECK (installment_count BETWEEN 1 AND 3),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Each installment is its own transfer-and-verify cycle. amount is a fixed
-- equal split of the invoice grand_total (remainder folded into the last
-- installment) computed once when the plan is created -- never recomputed,
-- so it stays stable even if catalog prices change later.
CREATE TABLE payment_installments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_plan_id   UUID NOT NULL REFERENCES payment_plans(id) ON DELETE CASCADE,
    sequence_no       INTEGER NOT NULL,
    amount            NUMERIC(18,2) NOT NULL CHECK (amount > 0),
    bank_account_id   UUID NOT NULL REFERENCES bank_accounts(id),
    status            VARCHAR(20) NOT NULL DEFAULT 'PENDING_PAYMENT'
                          CHECK (status IN ('PENDING_PAYMENT','PENDING_VERIFICATION','CONFIRMED','REJECTED')),
    proof_image_url   TEXT NOT NULL DEFAULT '',
    reject_reason     TEXT NOT NULL DEFAULT '',
    payment_id        UUID REFERENCES payments(id),
    reviewed_by       UUID REFERENCES users(id),
    reviewed_at       TIMESTAMPTZ,
    submitted_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (payment_plan_id, sequence_no)
);
CREATE INDEX idx_payment_installments_plan_id ON payment_installments(payment_plan_id);
