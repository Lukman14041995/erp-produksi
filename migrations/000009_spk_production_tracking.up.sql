-- ============================================================================
-- SPK (Surat Perintah Kerja) production-floor workflow tracking.
--
-- This is deliberately separate from production_orders/BOM/HPP (see
-- internal/production): that module is a costing engine (material issue,
-- labor/overhead capitalization, WIP -> Finished Goods journals). SPK is
-- pure workflow visibility for the production floor, sales, and the
-- customer -- "which stage is this order at, how many pieces are done" --
-- with no accounting side effects.
--
-- One SPK is generated per (sales_order_id, product_type_id) the moment a
-- sales order is confirmed, since a single customer order from the
-- configurator can mix Jersey and T-Shirt lines, and each product type
-- follows its own stage sequence (see internal/spk/model.go: tshirtFlow,
-- jerseyFlow). Stages are seeded up front from that hardcoded template --
-- there are only two known flows, so this is not a configurable
-- workflow-builder.
-- ============================================================================

CREATE TABLE spk_orders (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spk_number          VARCHAR(40) NOT NULL UNIQUE,
    sales_order_id      UUID NOT NULL REFERENCES sales_orders(id),
    so_number           VARCHAR(40) NOT NULL,
    customer_name       VARCHAR(150) NOT NULL,
    product_type_id     UUID NOT NULL REFERENCES product_types(id),
    product_type_code   VARCHAR(30) NOT NULL,
    product_type_name   VARCHAR(100) NOT NULL,
    total_qty           NUMERIC(18,2) NOT NULL DEFAULT 0,
    status              VARCHAR(20) NOT NULL DEFAULT 'NOT_STARTED'
                            CHECK (status IN ('NOT_STARTED','IN_PROGRESS','COMPLETED')),
    current_stage_code  VARCHAR(30) NOT NULL DEFAULT 'DESIGN',
    current_stage_name  VARCHAR(60) NOT NULL DEFAULT 'Design',
    progress_pct        NUMERIC(5,2) NOT NULL DEFAULT 0,
    notes               TEXT NOT NULL DEFAULT '',
    completed_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (sales_order_id, product_type_id)
);
CREATE INDEX idx_spk_orders_sales_order_id ON spk_orders(sales_order_id);
CREATE INDEX idx_spk_orders_product_type_code ON spk_orders(product_type_code);
CREATE INDEX idx_spk_orders_status ON spk_orders(status);
CREATE TRIGGER trg_spk_orders_updated_at BEFORE UPDATE ON spk_orders
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Which sales_order_items (and therefore which product/size) feed this
-- SPK's total_qty -- only the items matching this SPK's product type.
CREATE TABLE spk_items (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spk_order_id        UUID NOT NULL REFERENCES spk_orders(id) ON DELETE CASCADE,
    sales_order_item_id UUID NOT NULL REFERENCES sales_order_items(id),
    product_id          UUID NOT NULL REFERENCES products(id),
    product_size_id     UUID NOT NULL REFERENCES product_sizes(id),
    qty                 NUMERIC(18,2) NOT NULL CHECK (qty > 0)
);
CREATE INDEX idx_spk_items_spk_order_id ON spk_items(spk_order_id);

-- One row per stage per SPK, seeded from the product type's stage template
-- at creation time. requires_qty=false stages (Design/Selesai/Pengiriman)
-- are milestones toggled done with no quantity; the rest track
-- completed_qty against planned_qty (= the SPK's total_qty). Stages are
-- intentionally independent of each other (no sequential lock) -- the
-- production floor works them in whatever order/overlap makes sense, and
-- current_stage_code on spk_orders is just "furthest stage touched so far"
-- for the list view, not a gate.
CREATE TABLE spk_stages (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spk_order_id        UUID NOT NULL REFERENCES spk_orders(id) ON DELETE CASCADE,
    stage_seq           INTEGER NOT NULL,
    stage_code          VARCHAR(30) NOT NULL,
    stage_name          VARCHAR(60) NOT NULL,
    requires_qty        BOOLEAN NOT NULL DEFAULT true,
    planned_qty         NUMERIC(18,2) NOT NULL DEFAULT 0,
    completed_qty       NUMERIC(18,2) NOT NULL DEFAULT 0,
    status              VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                            CHECK (status IN ('PENDING','IN_PROGRESS','DONE')),
    started_at          TIMESTAMPTZ,
    completed_at        TIMESTAMPTZ,
    UNIQUE (spk_order_id, stage_seq)
);
CREATE INDEX idx_spk_stages_spk_order_id ON spk_stages(spk_order_id);

-- Append-only ledger of what the production team logged, when, and by
-- whom -- "hari ini tim produksi input berapa yang selesai di tahap X".
-- spk_stages.completed_qty is the running sum of this table's qty column
-- for that stage.
CREATE TABLE spk_stage_logs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spk_stage_id        UUID NOT NULL REFERENCES spk_stages(id) ON DELETE CASCADE,
    qty                 NUMERIC(18,2) NOT NULL DEFAULT 0,
    note                TEXT NOT NULL DEFAULT '',
    logged_by           UUID REFERENCES users(id),
    logged_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_spk_stage_logs_spk_stage_id ON spk_stage_logs(spk_stage_id);
