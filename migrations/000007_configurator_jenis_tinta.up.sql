-- ============================================================================
-- Redesign of the order configurator flow: customer now picks
-- Jenis (Jersey/T-Shirt) -> Model potongan (icon) -> Bahan (photo) ->
-- Tinta (photo) -> Ukuran+Qty, instead of browsing a pre-made "Desain"
-- catalog. Bahan, Model, and Tinta each have their own option list per
-- Jenis (jersey fabric/print differs from t-shirt fabric/print).
-- ============================================================================

-- Jenis pesanan (Jersey / T-Shirt). A proper admin-manageable master table
-- rather than a hardcoded enum, so a third type can be added later without
-- a code change.
CREATE TABLE product_types (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code          VARCHAR(30) NOT NULL UNIQUE,
    name          VARCHAR(100) NOT NULL,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_product_types_updated_at BEFORE UPDATE ON product_types
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

INSERT INTO product_types (code, name, sort_order) VALUES
    ('JERSEY', 'Jersey', 0),
    ('TSHIRT', 'T-Shirt', 1);

-- Model potongan (Normal/Long Sleeve/Raglan): now scoped per jenis and
-- shown as a small icon in the configurator, not a full sample photo --
-- hence icon_url (single image) rather than a gallery.
ALTER TABLE garment_variants
    ADD COLUMN product_type_id UUID REFERENCES product_types(id),
    ADD COLUMN icon_url TEXT NOT NULL DEFAULT '';

-- Backfill the NORMAL row seeded by 000006 onto Jersey so it isn't
-- orphaned; admin adds the T-Shirt equivalent (and Long Sleeve/Raglan)
-- through the catalog UI.
UPDATE garment_variants SET product_type_id = (SELECT id FROM product_types WHERE code = 'JERSEY') WHERE product_type_id IS NULL;
ALTER TABLE garment_variants ALTER COLUMN product_type_id SET NOT NULL;

-- code was globally unique; now it only needs to be unique within a jenis
-- (e.g. both Jersey and T-Shirt can each have their own "NORMAL").
ALTER TABLE garment_variants DROP CONSTRAINT garment_variants_code_key;
ALTER TABLE garment_variants ADD CONSTRAINT garment_variants_type_code_key UNIQUE (product_type_id, code);
CREATE INDEX idx_garment_variants_product_type_id ON garment_variants(product_type_id);

-- Bahan: scoped per jenis, with a sample photo (single image, not a
-- gallery -- one representative photo per fabric option).
ALTER TABLE fabrics
    ADD COLUMN product_type_id UUID REFERENCES product_types(id),
    ADD COLUMN image_url TEXT NOT NULL DEFAULT '';

UPDATE fabrics SET product_type_id = (SELECT id FROM product_types WHERE code = 'JERSEY') WHERE product_type_id IS NULL;
ALTER TABLE fabrics ALTER COLUMN product_type_id SET NOT NULL;

ALTER TABLE fabrics DROP CONSTRAINT fabrics_code_key;
ALTER TABLE fabrics ADD CONSTRAINT fabrics_type_code_key UNIQUE (product_type_id, code);
CREATE INDEX idx_fabrics_product_type_id ON fabrics(product_type_id);

-- Tinta (Rubber, Plastisol, ...): scoped per jenis, with a sample photo of
-- the print finish. price_addon works exactly like a variant's -- a fixed
-- extra cost per pcs, since different ink processes cost differently.
CREATE TABLE inks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_type_id UUID NOT NULL REFERENCES product_types(id),
    code            VARCHAR(30) NOT NULL,
    name            VARCHAR(100) NOT NULL,
    price_addon     NUMERIC(18,2) NOT NULL DEFAULT 0,
    image_url       TEXT NOT NULL DEFAULT '',
    sort_order      INTEGER NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (product_type_id, code)
);
CREATE INDEX idx_inks_product_type_id ON inks(product_type_id);
CREATE TRIGGER trg_inks_updated_at BEFORE UPDATE ON inks
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Quotation items: drop design_id (and the FK to designs it carries), add
-- product_type_id (so a line's jenis is explicit and validated against its
-- fabric/variant/ink, all of which are now jenis-scoped) and ink_id. No
-- existing rows reference design_id yet -- this configurator flow hasn't
-- shipped to a real customer.
ALTER TABLE order_quotation_items DROP CONSTRAINT order_quotation_items_quotation_id_design_id_fabric_id_garm_key;
ALTER TABLE order_quotation_items
    DROP COLUMN design_id,
    ADD COLUMN product_type_id UUID NOT NULL REFERENCES product_types(id),
    ADD COLUMN ink_id UUID NOT NULL REFERENCES inks(id);
ALTER TABLE order_quotation_items ADD CONSTRAINT order_quotation_items_unique_combo
    UNIQUE (quotation_id, product_type_id, fabric_id, garment_size_id, variant_id, ink_id);

-- The "browse a ready-made design" catalog from 000006 is fully replaced by
-- this flow -- now safe to drop, nothing references it any more.
DROP TABLE IF EXISTS design_images;
DROP TABLE IF EXISTS designs;
