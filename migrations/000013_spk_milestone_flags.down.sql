ALTER TABLE spk_orders
    DROP COLUMN IF EXISTS design_done,
    DROP COLUMN IF EXISTS production_done,
    DROP COLUMN IF EXISTS shipped;
