-- ============================================================================
-- Clothing / Jersey ERP - Initial Schema
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============================================================================
-- SECTION 0: SHARED TRIGGER FUNCTION (updated_at maintenance)
-- ============================================================================
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- SECTION 1: MASTER DATA
-- ============================================================================

CREATE TABLE customers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(30) NOT NULL UNIQUE,
    name            VARCHAR(150) NOT NULL,
    contact_person  VARCHAR(150),
    phone           VARCHAR(30),
    email           VARCHAR(150),
    address         TEXT,
    tax_id          VARCHAR(40),
    credit_limit    NUMERIC(18,2) NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_customers_created_at ON customers(created_at);
CREATE INDEX idx_customers_is_active ON customers(is_active);
CREATE TRIGGER trg_customers_updated_at BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE suppliers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(30) NOT NULL UNIQUE,
    name            VARCHAR(150) NOT NULL,
    contact_person  VARCHAR(150),
    phone           VARCHAR(30),
    email           VARCHAR(150),
    address         TEXT,
    tax_id          VARCHAR(40),
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_suppliers_created_at ON suppliers(created_at);
CREATE TRIGGER trg_suppliers_updated_at BEFORE UPDATE ON suppliers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE products (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(30) NOT NULL UNIQUE,
    name            VARCHAR(150) NOT NULL,
    category        VARCHAR(80),
    uom             VARCHAR(20) NOT NULL DEFAULT 'PCS',
    base_price      NUMERIC(18,2) NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_products_created_at ON products(created_at);
CREATE TRIGGER trg_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Size breakdown with cost multiplication factors (S:0.9, M:1.0, L:1.1, XL:1.2, ...)
CREATE TABLE product_sizes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    size_code       VARCHAR(10) NOT NULL,
    size_multiplier NUMERIC(6,4) NOT NULL DEFAULT 1.0000,
    sort_order      INTEGER NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (product_id, size_code)
);
CREATE INDEX idx_product_sizes_product_id ON product_sizes(product_id);
CREATE TRIGGER trg_product_sizes_updated_at BEFORE UPDATE ON product_sizes
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE materials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(30) NOT NULL UNIQUE,
    name            VARCHAR(150) NOT NULL,
    uom             VARCHAR(20) NOT NULL DEFAULT 'PCS',
    unit_cost       NUMERIC(18,4) NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_materials_created_at ON materials(created_at);
CREATE TRIGGER trg_materials_updated_at BEFORE UPDATE ON materials
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Chart of Accounts (standard 1-6: Asset/Liability/Equity/Revenue/COGS/Expense)
CREATE TABLE accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(20) NOT NULL UNIQUE,
    name            VARCHAR(150) NOT NULL,
    account_type    VARCHAR(20) NOT NULL CHECK (account_type IN
                        ('ASSET','LIABILITY','EQUITY','REVENUE','COGS','EXPENSE')),
    normal_balance  VARCHAR(6) NOT NULL CHECK (normal_balance IN ('DEBIT','CREDIT')),
    parent_id       UUID REFERENCES accounts(id),
    is_postable     BOOLEAN NOT NULL DEFAULT true,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_accounts_type ON accounts(account_type);
CREATE TRIGGER trg_accounts_updated_at BEFORE UPDATE ON accounts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Auto-numbering generator backing store (pkg/numbering)
CREATE TABLE document_sequences (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doc_type        VARCHAR(20) NOT NULL,      -- SO, INV, PAY, PRD, JV, EXP, ADJ ...
    period          VARCHAR(6)  NOT NULL,      -- YYYYMM
    prefix          VARCHAR(20) NOT NULL,
    last_number     BIGINT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (doc_type, period)
);

-- Accounting periods (locking for closed periods)
CREATE TABLE accounting_periods (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period          VARCHAR(6) NOT NULL UNIQUE, -- YYYYMM
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    status          VARCHAR(10) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN','CLOSED')),
    closed_at       TIMESTAMPTZ,
    closed_by       VARCHAR(150),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Bill of Materials
CREATE TABLE bom_headers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      UUID NOT NULL REFERENCES products(id),
    name            VARCHAR(150) NOT NULL,
    version         INTEGER NOT NULL DEFAULT 1,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_bom_headers_product_id ON bom_headers(product_id);
CREATE TRIGGER trg_bom_headers_updated_at BEFORE UPDATE ON bom_headers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE bom_lines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bom_id          UUID NOT NULL REFERENCES bom_headers(id) ON DELETE CASCADE,
    material_id     UUID NOT NULL REFERENCES materials(id),
    qty_per_unit    NUMERIC(18,6) NOT NULL,  -- quantity of material for 1 base (size-1.0) unit
    uom             VARCHAR(20) NOT NULL DEFAULT 'PCS',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_bom_lines_bom_id ON bom_lines(bom_id);

-- ============================================================================
-- SECTION 2: SALES DOMAIN
-- ============================================================================

CREATE TABLE sales_orders (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    so_number           VARCHAR(40) NOT NULL UNIQUE,
    customer_id         UUID NOT NULL REFERENCES customers(id),
    order_date          DATE NOT NULL DEFAULT CURRENT_DATE,
    -- Decoupled statuses (never collapse into one field)
    order_status        VARCHAR(20) NOT NULL DEFAULT 'DRAFT'
                            CHECK (order_status IN ('DRAFT','CONFIRMED','CANCELLED','CLOSED')),
    payment_status      VARCHAR(20) NOT NULL DEFAULT 'UNPAID'
                            CHECK (payment_status IN ('UNPAID','PARTIAL','PAID','OVERPAID')),
    production_status   VARCHAR(20) NOT NULL DEFAULT 'NOT_STARTED'
                            CHECK (production_status IN ('NOT_STARTED','IN_PROGRESS','COMPLETED','CANCELLED')),
    delivery_status     VARCHAR(20) NOT NULL DEFAULT 'NOT_DELIVERED'
                            CHECK (delivery_status IN ('NOT_DELIVERED','PARTIAL','DELIVERED')),
    subtotal            NUMERIC(18,2) NOT NULL DEFAULT 0,
    discount_total      NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax_total           NUMERIC(18,2) NOT NULL DEFAULT 0,
    grand_total         NUMERIC(18,2) NOT NULL DEFAULT 0,
    notes               TEXT,
    created_by          VARCHAR(150),
    confirmed_at        TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_sales_orders_customer_id ON sales_orders(customer_id);
CREATE INDEX idx_sales_orders_created_at ON sales_orders(created_at);
CREATE INDEX idx_sales_orders_order_status ON sales_orders(order_status);
CREATE TRIGGER trg_sales_orders_updated_at BEFORE UPDATE ON sales_orders
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE sales_order_items (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sales_order_id      UUID NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    product_id          UUID NOT NULL REFERENCES products(id),
    product_size_id     UUID NOT NULL REFERENCES product_sizes(id),
    qty                 NUMERIC(18,2) NOT NULL CHECK (qty > 0),
    unit_price          NUMERIC(18,2) NOT NULL CHECK (unit_price >= 0),
    discount            NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax_rate            NUMERIC(6,4) NOT NULL DEFAULT 0,
    line_total          NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_sales_order_items_so_id ON sales_order_items(sales_order_id);

CREATE TABLE invoices (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number      VARCHAR(40) NOT NULL UNIQUE,
    sales_order_id      UUID NOT NULL REFERENCES sales_orders(id),
    customer_id         UUID NOT NULL REFERENCES customers(id),
    invoice_date        DATE NOT NULL DEFAULT CURRENT_DATE,
    due_date            DATE,
    status              VARCHAR(20) NOT NULL DEFAULT 'POSTED'
                            CHECK (status IN ('DRAFT','POSTED','PARTIALLY_PAID','PAID','VOID')),
    subtotal            NUMERIC(18,2) NOT NULL DEFAULT 0,
    discount_total      NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax_total           NUMERIC(18,2) NOT NULL DEFAULT 0,
    grand_total         NUMERIC(18,2) NOT NULL DEFAULT 0,
    paid_amount         NUMERIC(18,2) NOT NULL DEFAULT 0,
    balance_due         NUMERIC(18,2) NOT NULL DEFAULT 0,
    journal_id          UUID,
    voided_at           TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_invoices_customer_id ON invoices(customer_id);
CREATE INDEX idx_invoices_sales_order_id ON invoices(sales_order_id);
CREATE INDEX idx_invoices_created_at ON invoices(created_at);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE TRIGGER trg_invoices_updated_at BEFORE UPDATE ON invoices
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE invoice_items (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id          UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    sales_order_item_id UUID NOT NULL REFERENCES sales_order_items(id),
    product_id          UUID NOT NULL REFERENCES products(id),
    product_size_id     UUID NOT NULL REFERENCES product_sizes(id),
    qty                 NUMERIC(18,2) NOT NULL CHECK (qty > 0),
    unit_price          NUMERIC(18,2) NOT NULL CHECK (unit_price >= 0),
    discount            NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax_rate            NUMERIC(6,4) NOT NULL DEFAULT 0,
    line_total          NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_invoice_items_invoice_id ON invoice_items(invoice_id);

-- ============================================================================
-- SECTION 3: FINANCE / PAYMENT DOMAIN
-- ============================================================================

CREATE TABLE payments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_number      VARCHAR(40) NOT NULL UNIQUE,
    payment_type        VARCHAR(15) NOT NULL CHECK (payment_type IN ('RECEIPT','DISBURSEMENT')),
    customer_id         UUID REFERENCES customers(id),
    supplier_id         UUID REFERENCES suppliers(id),
    cash_bank_account_id UUID NOT NULL REFERENCES accounts(id),
    payment_date        DATE NOT NULL DEFAULT CURRENT_DATE,
    amount              NUMERIC(18,2) NOT NULL CHECK (amount > 0),
    method              VARCHAR(30) NOT NULL DEFAULT 'BANK_TRANSFER',
    reference_no        VARCHAR(80),
    status              VARCHAR(15) NOT NULL DEFAULT 'DRAFT'
                            CHECK (status IN ('DRAFT','POSTED','VOID')),
    journal_id          UUID,
    notes               TEXT,
    posted_at           TIMESTAMPTZ,
    voided_at           TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (
        (payment_type = 'RECEIPT' AND customer_id IS NOT NULL AND supplier_id IS NULL) OR
        (payment_type = 'DISBURSEMENT' AND supplier_id IS NOT NULL AND customer_id IS NULL)
    )
);
CREATE INDEX idx_payments_customer_id ON payments(customer_id);
CREATE INDEX idx_payments_supplier_id ON payments(supplier_id);
CREATE INDEX idx_payments_created_at ON payments(created_at);
CREATE INDEX idx_payments_status ON payments(status);
CREATE TRIGGER trg_payments_updated_at BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE payment_allocations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id          UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    invoice_id          UUID NOT NULL REFERENCES invoices(id),
    amount_allocated    NUMERIC(18,2) NOT NULL CHECK (amount_allocated > 0),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_payment_allocations_payment_id ON payment_allocations(payment_id);
CREATE INDEX idx_payment_allocations_invoice_id ON payment_allocations(invoice_id);

CREATE TABLE expenses (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_number      VARCHAR(40) NOT NULL UNIQUE,
    expense_date        DATE NOT NULL DEFAULT CURRENT_DATE,
    expense_account_id  UUID NOT NULL REFERENCES accounts(id),
    paid_from_account_id UUID NOT NULL REFERENCES accounts(id),
    supplier_id         UUID REFERENCES suppliers(id),
    amount              NUMERIC(18,2) NOT NULL CHECK (amount > 0),
    description         TEXT,
    status              VARCHAR(15) NOT NULL DEFAULT 'DRAFT'
                            CHECK (status IN ('DRAFT','POSTED','VOID')),
    journal_id          UUID,
    posted_at           TIMESTAMPTZ,
    voided_at           TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_expenses_created_at ON expenses(created_at);
CREATE INDEX idx_expenses_status ON expenses(status);
CREATE TRIGGER trg_expenses_updated_at BEFORE UPDATE ON expenses
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================================
-- SECTION 4: PRODUCTION / HPP DOMAIN
-- ============================================================================

CREATE TABLE production_orders (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prod_number         VARCHAR(40) NOT NULL UNIQUE,
    sales_order_id      UUID REFERENCES sales_orders(id),
    product_id          UUID NOT NULL REFERENCES products(id),
    bom_id              UUID REFERENCES bom_headers(id),
    planned_qty         NUMERIC(18,2) NOT NULL CHECK (planned_qty > 0),
    finished_qty        NUMERIC(18,2) NOT NULL DEFAULT 0,
    production_status   VARCHAR(20) NOT NULL DEFAULT 'NOT_STARTED'
                            CHECK (production_status IN ('NOT_STARTED','IN_PROGRESS','COMPLETED','CANCELLED')),
    start_date          DATE,
    end_date            DATE,
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_production_orders_sales_order_id ON production_orders(sales_order_id);
CREATE INDEX idx_production_orders_created_at ON production_orders(created_at);
CREATE INDEX idx_production_orders_status ON production_orders(production_status);
CREATE TRIGGER trg_production_orders_updated_at BEFORE UPDATE ON production_orders
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE production_order_items (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    production_order_id UUID NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    product_size_id     UUID NOT NULL REFERENCES product_sizes(id),
    planned_qty         NUMERIC(18,2) NOT NULL CHECK (planned_qty > 0),
    finished_qty        NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_production_order_items_po_id ON production_order_items(production_order_id);

CREATE TABLE production_materials (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    production_order_id UUID NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    material_id         UUID NOT NULL REFERENCES materials(id),
    planned_qty         NUMERIC(18,4) NOT NULL DEFAULT 0,
    issued_qty          NUMERIC(18,4) NOT NULL DEFAULT 0,
    unit_cost           NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_cost          NUMERIC(18,2) NOT NULL DEFAULT 0,
    issued_at           TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_production_materials_po_id ON production_materials(production_order_id);

CREATE TABLE production_labor (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    production_order_id UUID NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    description         VARCHAR(150) NOT NULL,
    hours               NUMERIC(10,2) NOT NULL DEFAULT 0,
    rate                NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_cost          NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_production_labor_po_id ON production_labor(production_order_id);

CREATE TABLE production_overheads (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    production_order_id UUID NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    description         VARCHAR(150) NOT NULL,
    allocation_basis    VARCHAR(30) NOT NULL DEFAULT 'FIXED',
    amount              NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_production_overheads_po_id ON production_overheads(production_order_id);

-- HPP Snapshot: unit_cost = (material + labor + overhead) / finished_qty
CREATE TABLE production_costs (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    production_order_id    UUID NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    total_material_cost    NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_labor_cost       NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_overhead_cost    NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_cost             NUMERIC(18,2) NOT NULL DEFAULT 0,
    finished_qty           NUMERIC(18,2) NOT NULL DEFAULT 0,
    unit_cost              NUMERIC(18,4) NOT NULL DEFAULT 0,
    journal_id              UUID,
    computed_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (production_order_id)
);

-- ============================================================================
-- SECTION 5: INVENTORY LEDGER DOMAIN
-- ============================================================================

CREATE TABLE inventory_transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    txn_type        VARCHAR(20) NOT NULL CHECK (txn_type IN
                        ('PURCHASE','PRODUCTION_ISSUE','PRODUCTION_RECEIPT','SALE','ADJUSTMENT')),
    item_type       VARCHAR(10) NOT NULL CHECK (item_type IN ('MATERIAL','PRODUCT')),
    material_id     UUID REFERENCES materials(id),
    product_id      UUID REFERENCES products(id),
    product_size_id UUID REFERENCES product_sizes(id),
    qty_in          NUMERIC(18,4) NOT NULL DEFAULT 0,
    qty_out         NUMERIC(18,4) NOT NULL DEFAULT 0,
    unit_cost       NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_cost      NUMERIC(18,2) NOT NULL DEFAULT 0,
    ref_type        VARCHAR(30) NOT NULL,   -- SALES_ORDER, PRODUCTION_ORDER, PURCHASE, MANUAL
    ref_id          UUID,
    txn_date        DATE NOT NULL DEFAULT CURRENT_DATE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (
        (item_type = 'MATERIAL' AND material_id IS NOT NULL AND product_id IS NULL) OR
        (item_type = 'PRODUCT' AND product_id IS NOT NULL AND material_id IS NULL)
    )
);
CREATE INDEX idx_inventory_txn_material_id ON inventory_transactions(material_id);
CREATE INDEX idx_inventory_txn_product_id ON inventory_transactions(product_id);
CREATE INDEX idx_inventory_txn_created_at ON inventory_transactions(created_at);
CREATE INDEX idx_inventory_txn_ref ON inventory_transactions(ref_type, ref_id);

-- Running balance per item (maintained transactionally alongside inventory_transactions)
CREATE TABLE inventory_balances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_type       VARCHAR(10) NOT NULL CHECK (item_type IN ('MATERIAL','PRODUCT')),
    material_id     UUID REFERENCES materials(id),
    product_id      UUID REFERENCES products(id),
    product_size_id UUID REFERENCES product_sizes(id),
    qty_on_hand     NUMERIC(18,4) NOT NULL DEFAULT 0,
    avg_unit_cost   NUMERIC(18,4) NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (item_type, material_id, product_id, product_size_id)
);

-- ============================================================================
-- SECTION 6: ACCOUNTING DOMAIN (immutable double-entry journals)
-- ============================================================================

CREATE TABLE journal_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_number  VARCHAR(40) NOT NULL UNIQUE,
    journal_date    DATE NOT NULL DEFAULT CURRENT_DATE,
    source_type     VARCHAR(30) NOT NULL CHECK (source_type IN
                        ('SALES_INVOICE','PAYMENT_RECEIPT','PAYMENT_DISBURSEMENT',
                         'MATERIAL_ISSUE','LABOR_COST','OVERHEAD_COST','PRODUCTION_COMPLETION','COGS','EXPENSE',
                         'MANUAL','REVERSAL')),
    source_id       UUID,
    description     TEXT,
    status          VARCHAR(10) NOT NULL DEFAULT 'POSTED' CHECK (status IN ('POSTED','REVERSED')),
    reversed_journal_id UUID REFERENCES journal_entries(id),
    total_debit     NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_credit    NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_by      VARCHAR(150),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (total_debit = total_credit)
);
CREATE INDEX idx_journal_entries_journal_date ON journal_entries(journal_date);
CREATE INDEX idx_journal_entries_source ON journal_entries(source_type, source_id);
CREATE INDEX idx_journal_entries_created_at ON journal_entries(created_at);

CREATE TABLE journal_entry_lines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_id      UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    line_no         INTEGER NOT NULL,
    account_id      UUID NOT NULL REFERENCES accounts(id),
    debit           NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (debit >= 0),
    credit          NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (credit >= 0),
    description     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (NOT (debit > 0 AND credit > 0)),
    CHECK (debit > 0 OR credit > 0)
);
CREATE INDEX idx_journal_entry_lines_journal_id ON journal_entry_lines(journal_id);
CREATE INDEX idx_journal_entry_lines_account_id ON journal_entry_lines(account_id);

-- Deferred FKs from transactional tables to journal_entries (added after table exists)
ALTER TABLE invoices           ADD CONSTRAINT fk_invoices_journal           FOREIGN KEY (journal_id) REFERENCES journal_entries(id);
ALTER TABLE payments           ADD CONSTRAINT fk_payments_journal           FOREIGN KEY (journal_id) REFERENCES journal_entries(id);
ALTER TABLE expenses           ADD CONSTRAINT fk_expenses_journal           FOREIGN KEY (journal_id) REFERENCES journal_entries(id);
ALTER TABLE production_costs   ADD CONSTRAINT fk_production_costs_journal  FOREIGN KEY (journal_id) REFERENCES journal_entries(id);

-- ============================================================================
-- SECTION 7: AUDIT LOG
-- ============================================================================

CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name      VARCHAR(80) NOT NULL,
    record_id       UUID NOT NULL,
    action          VARCHAR(10) NOT NULL CHECK (action IN ('INSERT','UPDATE','DELETE')),
    old_data        JSONB,
    new_data        JSONB,
    changed_by      VARCHAR(150),
    changed_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_logs_table_record ON audit_logs(table_name, record_id);
CREATE INDEX idx_audit_logs_changed_at ON audit_logs(changed_at);

-- ============================================================================
-- SECTION 8: SEED DATA - Chart of Accounts (standard 1-6)
-- ============================================================================

INSERT INTO accounts (code, name, account_type, normal_balance, is_postable) VALUES
    ('1-1001', 'Kas',                              'ASSET',     'DEBIT', true),
    ('1-1002', 'Bank',                              'ASSET',     'DEBIT', true),
    ('1-1100', 'Piutang Usaha',                     'ASSET',     'DEBIT', true),
    ('1-1200', 'Persediaan Bahan Baku',             'ASSET',     'DEBIT', true),
    ('1-1210', 'Persediaan Barang Dalam Proses',    'ASSET',     'DEBIT', true),
    ('1-1220', 'Persediaan Barang Jadi',            'ASSET',     'DEBIT', true),
    ('1-1500', 'Aset Tetap',                        'ASSET',     'DEBIT', true),
    ('2-1000', 'Hutang Usaha',                      'LIABILITY', 'CREDIT', true),
    ('2-2000', 'Hutang Pajak',                      'LIABILITY', 'CREDIT', true),
    ('3-1000', 'Modal Disetor',                     'EQUITY',    'CREDIT', true),
    ('3-2000', 'Laba Ditahan',                      'EQUITY',    'CREDIT', true),
    ('4-1000', 'Penjualan',                         'REVENUE',   'CREDIT', true),
    ('4-2000', 'Retur Penjualan',                   'REVENUE',   'DEBIT', true),
    ('5-1000', 'Harga Pokok Penjualan (HPP)',       'COGS',      'DEBIT', true),
    ('6-1000', 'Beban Gaji',                        'EXPENSE',   'DEBIT', true),
    ('6-1100', 'Beban Operasional',                 'EXPENSE',   'DEBIT', true),
    ('6-1200', 'Beban Listrik & Air',                'EXPENSE',   'DEBIT', true),
    ('6-1300', 'Beban Sewa',                         'EXPENSE',   'DEBIT', true),
    ('6-9000', 'Beban Lain-lain',                    'EXPENSE',   'DEBIT', true);

