
ALTER TABLE journal_entries DROP CONSTRAINT journal_entries_source_type_check;
ALTER TABLE journal_entries ADD CONSTRAINT journal_entries_source_type_check CHECK (source_type IN
    ('SALES_INVOICE','PAYMENT_RECEIPT','PAYMENT_DISBURSEMENT',
     'MATERIAL_ISSUE','LABOR_COST','OVERHEAD_COST','PRODUCTION_COMPLETION','COGS','EXPENSE',
     'SUPPLIER_BILL','GOODS_RECEIPT','MANUAL','REVERSAL'));

-- Unbilled AP (accrued liability recognized at goods receipt, before the
-- supplier's actual bill arrives) and the purchase price variance account
-- used when a bill's unit cost differs from the PO's agreed cost.
INSERT INTO accounts (code, name, account_type, normal_balance, is_postable) VALUES
    ('2-1100', 'Hutang Belum Ditagih (Unbilled AP)', 'LIABILITY', 'CREDIT', true),
    ('6-9100', 'Selisih Harga Pembelian (Purchase Price Variance)', 'EXPENSE', 'DEBIT', true)
ON CONFLICT (code) DO NOTHING;

CREATE TABLE purchase_orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    po_number       VARCHAR(40) NOT NULL UNIQUE,
    supplier_id     UUID NOT NULL REFERENCES suppliers(id),
    order_date      DATE NOT NULL DEFAULT CURRENT_DATE,
    expected_date   DATE,
    status          VARCHAR(20) NOT NULL DEFAULT 'DRAFT'
                        CHECK (status IN ('DRAFT','APPROVED','PARTIALLY_RECEIVED','FULLY_RECEIVED','CLOSED','CANCELLED')),
    billing_status  VARCHAR(20) NOT NULL DEFAULT 'UNBILLED'
                        CHECK (billing_status IN ('UNBILLED','PARTIALLY_BILLED','FULLY_BILLED')),
    subtotal        NUMERIC(18,2) NOT NULL DEFAULT 0,
    grand_total     NUMERIC(18,2) NOT NULL DEFAULT 0,
    notes           TEXT,
    approved_at     TIMESTAMPTZ,
    cancelled_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_purchase_orders_supplier_id ON purchase_orders(supplier_id);
CREATE INDEX idx_purchase_orders_created_at ON purchase_orders(created_at);
CREATE INDEX idx_purchase_orders_status ON purchase_orders(status);
CREATE TRIGGER trg_purchase_orders_updated_at BEFORE UPDATE ON purchase_orders
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE purchase_order_items (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id   UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    material_id         UUID NOT NULL REFERENCES materials(id),
    qty                 NUMERIC(18,4) NOT NULL CHECK (qty > 0),
    unit_cost           NUMERIC(18,4) NOT NULL CHECK (unit_cost >= 0),
    qty_received        NUMERIC(18,4) NOT NULL DEFAULT 0,
    qty_billed          NUMERIC(18,4) NOT NULL DEFAULT 0,
    line_total          NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_purchase_order_items_po_id ON purchase_order_items(purchase_order_id);

-- Goods Receipt Note: posts the DR Material Stock / CR Unbilled AP journal
-- and the inventory PURCHASE receipt at the PO's agreed unit cost, ahead of
-- the supplier's actual bill arriving.
CREATE TABLE goods_receipts (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grn_number          VARCHAR(40) NOT NULL UNIQUE,
    purchase_order_id   UUID NOT NULL REFERENCES purchase_orders(id),
    supplier_id         UUID NOT NULL REFERENCES suppliers(id),
    receipt_date        DATE NOT NULL DEFAULT CURRENT_DATE,
    journal_id          UUID REFERENCES journal_entries(id),
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_goods_receipts_po_id ON goods_receipts(purchase_order_id);
CREATE INDEX idx_goods_receipts_created_at ON goods_receipts(created_at);

CREATE TABLE goods_receipt_items (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    goods_receipt_id        UUID NOT NULL REFERENCES goods_receipts(id) ON DELETE CASCADE,
    purchase_order_item_id  UUID NOT NULL REFERENCES purchase_order_items(id),
    material_id             UUID NOT NULL REFERENCES materials(id),
    qty_received            NUMERIC(18,4) NOT NULL CHECK (qty_received > 0),
    unit_cost               NUMERIC(18,4) NOT NULL CHECK (unit_cost >= 0), -- PO's agreed cost, snapshotted
    qty_billed              NUMERIC(18,4) NOT NULL DEFAULT 0,
    line_total              NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_goods_receipt_items_grn_id ON goods_receipt_items(goods_receipt_id);
CREATE INDEX idx_goods_receipt_items_po_item_id ON goods_receipt_items(purchase_order_item_id);

-- 3-way matching linkage: a supplier bill line matched against a specific
-- goods receipt line carries both the PO item (for the standard/agreed
-- cost) and the GRN item (for remaining-unbilled-quantity tracking).
ALTER TABLE supplier_invoices ADD COLUMN purchase_order_id UUID REFERENCES purchase_orders(id);
ALTER TABLE supplier_invoice_items ADD COLUMN purchase_order_item_id UUID REFERENCES purchase_order_items(id);
ALTER TABLE supplier_invoice_items ADD COLUMN goods_receipt_item_id UUID REFERENCES goods_receipt_items(id);
ALTER TABLE supplier_invoice_items ADD COLUMN price_variance NUMERIC(18,2) NOT NULL DEFAULT 0;

CREATE INDEX idx_supplier_invoices_po_id ON supplier_invoices(purchase_order_id);

