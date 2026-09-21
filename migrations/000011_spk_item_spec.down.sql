ALTER TABLE spk_orders DROP COLUMN IF EXISTS composition_summary;

ALTER TABLE spk_items
    DROP COLUMN IF EXISTS fabric_name,
    DROP COLUMN IF EXISTS variant_name,
    DROP COLUMN IF EXISTS ink_name,
    DROP COLUMN IF EXISTS size_code;
