
ALTER TABLE journal_entries DROP CONSTRAINT journal_entries_source_type_check;
ALTER TABLE journal_entries ADD CONSTRAINT journal_entries_source_type_check CHECK (source_type IN
    ('SALES_INVOICE','PAYMENT_RECEIPT','PAYMENT_DISBURSEMENT',
     'MATERIAL_ISSUE','LABOR_COST','OVERHEAD_COST','PRODUCTION_COMPLETION','COGS','EXPENSE',
     'SUPPLIER_BILL','MANUAL','REVERSAL'));

-- Supplier bills: the AP subledger. Mirrors invoices/invoice_items on the AR
-- side so AP aging can be computed the same way -- from actual unpaid
-- bills, not approximated from payment history.
CREATE TABLE supplier_invoices (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bill_number         VARCHAR(40) NOT NULL UNIQUE,
    supplier_id         UUID NOT NULL REFERENCES suppliers(id),
    bill_date           DATE NOT NULL DEFAULT CURRENT_DATE,
    due_date            DATE,
    debit_account_code  VARCHAR(20) NOT NULL, -- e.g. 1-1200 Persediaan Bahan Baku, or an expense account
    status              VARCHAR(20) NOT NULL DEFAULT 'DRAFT'
                            CHECK (status IN ('DRAFT','POSTED','PARTIALLY_PAID','PAID','VOID')),
    subtotal            NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax_total           NUMERIC(18,2) NOT NULL DEFAULT 0,
    grand_total         NUMERIC(18,2) NOT NULL DEFAULT 0,
    paid_amount         NUMERIC(18,2) NOT NULL DEFAULT 0,
    balance_due         NUMERIC(18,2) NOT NULL DEFAULT 0,
    journal_id          UUID REFERENCES journal_entries(id),
    notes               TEXT,
    voided_at           TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_supplier_invoices_supplier_id ON supplier_invoices(supplier_id);
CREATE INDEX idx_supplier_invoices_created_at ON supplier_invoices(created_at);
CREATE INDEX idx_supplier_invoices_status ON supplier_invoices(status);
CREATE TRIGGER trg_supplier_invoices_updated_at BEFORE UPDATE ON supplier_invoices
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE supplier_invoice_items (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_invoice_id UUID NOT NULL REFERENCES supplier_invoices(id) ON DELETE CASCADE,
    material_id         UUID NOT NULL REFERENCES materials(id),
    qty                 NUMERIC(18,4) NOT NULL CHECK (qty > 0),
    unit_cost           NUMERIC(18,4) NOT NULL CHECK (unit_cost >= 0),
    line_total          NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_supplier_invoice_items_bill_id ON supplier_invoice_items(supplier_invoice_id);

-- Parallel to payment_allocations, but against the AP subledger instead of
-- AR invoices, used when payments.payment_type = 'DISBURSEMENT'.
CREATE TABLE supplier_payment_allocations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id          UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    supplier_invoice_id UUID NOT NULL REFERENCES supplier_invoices(id),
    amount_allocated    NUMERIC(18,2) NOT NULL CHECK (amount_allocated > 0),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_supplier_payment_allocations_payment_id ON supplier_payment_allocations(payment_id);
CREATE INDEX idx_supplier_payment_allocations_bill_id ON supplier_payment_allocations(supplier_invoice_id);

