-- Cached running total of material cost issued against an SPK (sum of
-- spk_material_usages.total_cost), maintained incrementally by
-- Service.IssueMaterial so /production/costing's board can show "Biaya
-- Bahan" per SPK without summing spk_material_usages per row. This is
-- material cost only -- SPK doesn't track labor/overhead the way the
-- legacy BOM/HPP module does, so it is NOT a full HPP figure.
ALTER TABLE spk_orders ADD COLUMN material_cost_total NUMERIC(18,2) NOT NULL DEFAULT 0;
