CREATE TABLE designs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code          VARCHAR(30) NOT NULL UNIQUE,
    name          VARCHAR(150) NOT NULL,
    description   TEXT,
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_designs_updated_at BEFORE UPDATE ON designs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE design_images (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    design_id     UUID NOT NULL REFERENCES designs(id) ON DELETE CASCADE,
    image_url     TEXT NOT NULL,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_design_images_design_id ON design_images(design_id);

ALTER TABLE order_quotation_items DROP CONSTRAINT IF EXISTS order_quotation_items_unique_combo;
ALTER TABLE order_quotation_items
    DROP COLUMN IF EXISTS ink_id,
    DROP COLUMN IF EXISTS product_type_id,
    ADD COLUMN design_id UUID REFERENCES designs(id);
ALTER TABLE order_quotation_items ADD CONSTRAINT order_quotation_items_quotation_id_design_id_fabric_id_garm_key
    UNIQUE (quotation_id, design_id, fabric_id, garment_size_id, variant_id);

DROP TABLE IF EXISTS inks;

ALTER TABLE fabrics DROP CONSTRAINT IF EXISTS fabrics_type_code_key;
ALTER TABLE fabrics ADD CONSTRAINT fabrics_code_key UNIQUE (code);
ALTER TABLE fabrics
    DROP COLUMN IF EXISTS product_type_id,
    DROP COLUMN IF EXISTS image_url;

ALTER TABLE garment_variants DROP CONSTRAINT IF EXISTS garment_variants_type_code_key;
ALTER TABLE garment_variants ADD CONSTRAINT garment_variants_code_key UNIQUE (code);
ALTER TABLE garment_variants
    DROP COLUMN IF EXISTS product_type_id,
    DROP COLUMN IF EXISTS icon_url;

DROP TABLE IF EXISTS product_types;
