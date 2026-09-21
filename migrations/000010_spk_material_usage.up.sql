-- ============================================================================
-- 1) Ink gets a material_id link, exactly like Fabric already has -- so
-- each tinta color can carry its own harga modal (unit_cost) via the
-- materials master, instead of duplicating a cost field per catalog table.
-- ============================================================================

ALTER TABLE inks ADD COLUMN material_id UUID REFERENCES materials(id);

-- Backfill: give every existing ink row a Material to point at (code
-- derived from the ink's own id so it can never collide, since ink codes
-- are only unique per product_type, not globally like materials.code).
-- unit_cost starts at 0; admin fills in the real harga modal per color
-- afterwards via /master-data/materials or the ink form itself.
INSERT INTO materials (code, name, uom, unit_cost)
SELECT left('INK-' || substr(i.id::text, 1, 8), 30), 'Tinta ' || i.name, 'KG', 0
FROM inks i
WHERE i.material_id IS NULL;

UPDATE inks i
SET material_id = m.id
FROM materials m
WHERE i.material_id IS NULL AND m.code = left('INK-' || substr(i.id::text, 1, 8), 30);

ALTER TABLE inks ALTER COLUMN material_id SET NOT NULL;
CREATE INDEX idx_inks_material_id ON inks(material_id);

-- ============================================================================
-- 2) Material usage recorded against an SPK -- "kain 2 roll", "tinta 500ml",
-- "plastik packing 50 pcs" issued for a specific work order, optionally
-- tied to the stage that consumed it (fabric at Cutting, ink at Sablon,
-- packing plastic at Packing). Unlike spk_stages (pure workflow, no
-- accounting effect), this DOES move real inventory: each row is backed by
-- one inventory_transactions movement (moving-average cost, stock actually
-- decremented) and one journal_entries posting (DR WIP / CR Bahan Baku),
-- the same accounting law internal/production.IssueMaterial already uses
-- for the BOM/HPP costing flow -- see internal/spk/service.go IssueMaterial.
-- ============================================================================

CREATE TABLE spk_material_usages (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spk_order_id        UUID NOT NULL REFERENCES spk_orders(id) ON DELETE CASCADE,
    spk_stage_id        UUID REFERENCES spk_stages(id),
    material_id         UUID NOT NULL REFERENCES materials(id),
    qty                 NUMERIC(18,4) NOT NULL CHECK (qty > 0),
    unit_cost           NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_cost          NUMERIC(18,2) NOT NULL DEFAULT 0,
    warehouse_id        UUID NOT NULL REFERENCES warehouses(id),
    inventory_txn_id    UUID REFERENCES inventory_transactions(id),
    journal_id          UUID REFERENCES journal_entries(id),
    note                TEXT NOT NULL DEFAULT '',
    logged_by           UUID REFERENCES users(id),
    logged_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_spk_material_usages_spk_order_id ON spk_material_usages(spk_order_id);
CREATE INDEX idx_spk_material_usages_spk_stage_id ON spk_material_usages(spk_stage_id);
CREATE INDEX idx_spk_material_usages_material_id ON spk_material_usages(material_id);
