ALTER TABLE journal_entries DROP CONSTRAINT journal_entries_source_type_check;
ALTER TABLE journal_entries ADD CONSTRAINT journal_entries_source_type_check CHECK (source_type IN
    ('SALES_INVOICE','PAYMENT_RECEIPT','PAYMENT_DISBURSEMENT',
     'MATERIAL_ISSUE','LABOR_COST','OVERHEAD_COST','PRODUCTION_COMPLETION','COGS','EXPENSE',
     'SUPPLIER_BILL','GOODS_RECEIPT','STOCK_OPNAME','MANUAL','REVERSAL'));

INSERT INTO accounts (code, name, account_type, normal_balance, is_postable) VALUES
    ('5-9200', 'Selisih Stock Opname (Inventory Adjustment)', 'COGS', 'DEBIT', true)
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- Warehouses
-- ============================================================================
CREATE TABLE warehouses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(20) NOT NULL UNIQUE,
    name            VARCHAR(150) NOT NULL,
    address         TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_warehouses_updated_at BEFORE UPDATE ON warehouses
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- address is '' rather than NULL: the Go model's Address field is a plain
-- string (matching every other master-data address column in this schema),
-- so a NULL here would fail to scan.
INSERT INTO warehouses (code, name, address) VALUES
    ('WH-MAT', 'Gudang Utama Bahan', ''),
    ('WH-FG',  'Gudang Produk Jadi', ''),
    ('WH-WIP', 'Gudang Transit/Produksi', '');

-- ============================================================================
-- inventory_transactions / inventory_balances become warehouse-scoped.
-- Existing rows predate multi-warehouse support, so backfill them into the
-- matching default warehouse by item_type before enforcing NOT NULL --
-- there is no ambiguity to resolve since every prior transaction/balance
-- implicitly lived in "the" warehouse for its item type.
-- ============================================================================
ALTER TABLE inventory_transactions ADD COLUMN warehouse_id UUID REFERENCES warehouses(id);

UPDATE inventory_transactions SET warehouse_id = (SELECT id FROM warehouses WHERE code = 'WH-MAT') WHERE item_type = 'MATERIAL';
UPDATE inventory_transactions SET warehouse_id = (SELECT id FROM warehouses WHERE code = 'WH-FG')  WHERE item_type = 'PRODUCT';

ALTER TABLE inventory_transactions ALTER COLUMN warehouse_id SET NOT NULL;
CREATE INDEX idx_inventory_txn_warehouse_id ON inventory_transactions(warehouse_id);

ALTER TABLE inventory_balances ADD COLUMN warehouse_id UUID REFERENCES warehouses(id);

UPDATE inventory_balances SET warehouse_id = (SELECT id FROM warehouses WHERE code = 'WH-MAT') WHERE item_type = 'MATERIAL';
UPDATE inventory_balances SET warehouse_id = (SELECT id FROM warehouses WHERE code = 'WH-FG')  WHERE item_type = 'PRODUCT';

ALTER TABLE inventory_balances ALTER COLUMN warehouse_id SET NOT NULL;
ALTER TABLE inventory_balances DROP CONSTRAINT inventory_balances_item_type_material_id_product_id_product_key;
ALTER TABLE inventory_balances ADD CONSTRAINT inventory_balances_unique_per_warehouse
    UNIQUE (item_type, material_id, product_id, product_size_id, warehouse_id);
CREATE INDEX idx_inventory_balances_warehouse_id ON inventory_balances(warehouse_id);

-- New movement types for inter-warehouse transfers.
ALTER TABLE inventory_transactions DROP CONSTRAINT inventory_transactions_txn_type_check;
ALTER TABLE inventory_transactions ADD CONSTRAINT inventory_transactions_txn_type_check CHECK (txn_type IN
    ('PURCHASE','PRODUCTION_ISSUE','PRODUCTION_RECEIPT','SALE','ADJUSTMENT','TRANSFER_OUT','TRANSFER_IN'));

-- ============================================================================
-- Inter-warehouse transfers: request -> dispatch (TRANSFER_OUT) -> receive
-- (TRANSFER_IN). No GL impact -- moving stock between company warehouses
-- doesn't change what the company owns, only where it sits.
-- ============================================================================
CREATE TABLE stock_transfers (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_number         VARCHAR(40) NOT NULL UNIQUE,
    source_warehouse_id     UUID NOT NULL REFERENCES warehouses(id),
    destination_warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    status                  VARCHAR(20) NOT NULL DEFAULT 'DRAFT'
                                CHECK (status IN ('DRAFT','DISPATCHED','RECEIVED','CANCELLED')),
    transfer_date           DATE NOT NULL DEFAULT CURRENT_DATE,
    notes                   TEXT,
    dispatched_at           TIMESTAMPTZ,
    received_at             TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (source_warehouse_id <> destination_warehouse_id)
);
CREATE INDEX idx_stock_transfers_created_at ON stock_transfers(created_at);
CREATE INDEX idx_stock_transfers_status ON stock_transfers(status);
CREATE TRIGGER trg_stock_transfers_updated_at BEFORE UPDATE ON stock_transfers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE stock_transfer_items (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stock_transfer_id   UUID NOT NULL REFERENCES stock_transfers(id) ON DELETE CASCADE,
    item_type           VARCHAR(10) NOT NULL CHECK (item_type IN ('MATERIAL','PRODUCT')),
    material_id         UUID REFERENCES materials(id),
    product_id          UUID REFERENCES products(id),
    product_size_id     UUID REFERENCES product_sizes(id),
    qty                 NUMERIC(18,4) NOT NULL CHECK (qty > 0),
    unit_cost           NUMERIC(18,4) NOT NULL DEFAULT 0, -- snapshotted from source warehouse at dispatch time
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (
        (item_type = 'MATERIAL' AND material_id IS NOT NULL AND product_id IS NULL) OR
        (item_type = 'PRODUCT' AND product_id IS NOT NULL AND material_id IS NULL)
    )
);
CREATE INDEX idx_stock_transfer_items_transfer_id ON stock_transfer_items(stock_transfer_id);

-- ============================================================================
-- Stock opname (physical inventory count) and its GL reconciliation.
-- ============================================================================
CREATE TABLE stock_opnames (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    opname_number   VARCHAR(40) NOT NULL UNIQUE,
    warehouse_id    UUID NOT NULL REFERENCES warehouses(id),
    opname_date     DATE NOT NULL DEFAULT CURRENT_DATE,
    status          VARCHAR(20) NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT','POSTED','CANCELLED')),
    notes           TEXT,
    journal_id      UUID REFERENCES journal_entries(id),
    posted_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_stock_opnames_warehouse_id ON stock_opnames(warehouse_id);
CREATE INDEX idx_stock_opnames_created_at ON stock_opnames(created_at);
CREATE TRIGGER trg_stock_opnames_updated_at BEFORE UPDATE ON stock_opnames
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE stock_opname_items (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stock_opname_id         UUID NOT NULL REFERENCES stock_opnames(id) ON DELETE CASCADE,
    item_type               VARCHAR(10) NOT NULL CHECK (item_type IN ('MATERIAL','PRODUCT')),
    material_id             UUID REFERENCES materials(id),
    product_id              UUID REFERENCES products(id),
    product_size_id         UUID REFERENCES product_sizes(id),
    system_qty              NUMERIC(18,4) NOT NULL, -- frozen at draft-creation time
    actual_qty              NUMERIC(18,4) NOT NULL,
    unit_cost               NUMERIC(18,4) NOT NULL, -- frozen moving-average cost at draft-creation time
    variance_qty            NUMERIC(18,4) NOT NULL DEFAULT 0,
    variance_amount         NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (
        (item_type = 'MATERIAL' AND material_id IS NOT NULL AND product_id IS NULL) OR
        (item_type = 'PRODUCT' AND product_id IS NOT NULL AND material_id IS NULL)
    )
);
CREATE INDEX idx_stock_opname_items_opname_id ON stock_opname_items(stock_opname_id);
