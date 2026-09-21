-- Cached milestone flags for the cross-jenis production board
-- (/production/costing): since spk_stages are independent of each other
-- (no sequential gating -- see migration 000009), placing one SPK into a
-- coarse Design/Produksi/Selesai/Pengiriman bucket needs to know whether
-- three specific milestone stages (DESIGN, DONE="Selesai",
-- SHIPPING="Pengiriman") are done, not just "furthest stage touched"
-- (current_stage_code). Recomputed alongside progress_pct/status in
-- internal/spk/service.go recomputeOrderProgress.
ALTER TABLE spk_orders
    ADD COLUMN design_done     BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN production_done BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN shipped         BOOLEAN NOT NULL DEFAULT false;
