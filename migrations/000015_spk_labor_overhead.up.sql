-- Labor and overhead recorded against an SPK, completing the material cost
-- already tracked (spk_material_usages) into a real HPP figure: material +
-- labor + overhead. Same accounting law as internal/production.AddLabor/
-- AddOverhead -- capitalized into WIP (DR WIP / CR Bank, paid-in-cash
-- simplification), not a period expense -- but scoped to spk_orders
-- instead of the legacy BOM production_orders. See
-- internal/spk/service.go AddLabor/AddOverhead.

CREATE TABLE spk_labor (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spk_order_id    UUID NOT NULL REFERENCES spk_orders(id) ON DELETE CASCADE,
    spk_stage_id    UUID REFERENCES spk_stages(id),
    description     TEXT NOT NULL,
    hours           NUMERIC(18,2) NOT NULL CHECK (hours > 0),
    rate            NUMERIC(18,2) NOT NULL CHECK (rate >= 0),
    total_cost      NUMERIC(18,2) NOT NULL DEFAULT 0,
    journal_id      UUID REFERENCES journal_entries(id),
    logged_by       UUID REFERENCES users(id),
    logged_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_spk_labor_spk_order_id ON spk_labor(spk_order_id);
CREATE INDEX idx_spk_labor_spk_stage_id ON spk_labor(spk_stage_id);

CREATE TABLE spk_overheads (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spk_order_id        UUID NOT NULL REFERENCES spk_orders(id) ON DELETE CASCADE,
    description         TEXT NOT NULL,
    allocation_basis    TEXT NOT NULL DEFAULT '',
    amount              NUMERIC(18,2) NOT NULL CHECK (amount > 0),
    journal_id          UUID REFERENCES journal_entries(id),
    logged_by           UUID REFERENCES users(id),
    logged_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_spk_overheads_spk_order_id ON spk_overheads(spk_order_id);

-- Cached running totals, same pattern as material_cost_total, so HPP =
-- material_cost_total + labor_cost_total + overhead_cost_total can be read
-- straight off spk_orders without joining three child tables per row.
ALTER TABLE spk_orders
    ADD COLUMN labor_cost_total    NUMERIC(18,2) NOT NULL DEFAULT 0,
    ADD COLUMN overhead_cost_total NUMERIC(18,2) NOT NULL DEFAULT 0;
