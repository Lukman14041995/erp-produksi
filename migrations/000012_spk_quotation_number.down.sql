ALTER TABLE spk_orders
    DROP COLUMN IF EXISTS quotation_id,
    DROP COLUMN IF EXISTS quotation_number;
