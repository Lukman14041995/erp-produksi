ALTER TABLE spk_orders
    DROP COLUMN IF EXISTS labor_cost_total,
    DROP COLUMN IF EXISTS overhead_cost_total;

DROP TABLE IF EXISTS spk_overheads;
DROP TABLE IF EXISTS spk_labor;
