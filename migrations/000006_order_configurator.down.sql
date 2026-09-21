ALTER TABLE sales_order_items DROP COLUMN IF EXISTS quotation_item_id;

DROP TABLE IF EXISTS order_quotation_items;
DROP TABLE IF EXISTS order_quotations;
DROP TABLE IF EXISTS order_links;
DROP TABLE IF EXISTS fabrics;
DROP TABLE IF EXISTS design_images;
DROP TABLE IF EXISTS designs;
DROP TABLE IF EXISTS garment_sizes;
DROP TABLE IF EXISTS garment_variants;
