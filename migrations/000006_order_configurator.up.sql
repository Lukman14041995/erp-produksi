-- ============================================================================
-- Order configurator: customer self-service pesanan (Desain + Bahan + Ukuran
-- + Varian potongan) via link token, masuk sebagai draft/quotation sebelum
-- staf sales mengonfirmasinya jadi Sales Order resmi.
-- ============================================================================

-- Varian potongan (mis. NORMAL, LONG_SLEEVE). Harga varian adalah biaya
-- tambahan TETAP per pcs, bukan multiplier -- mencerminkan biaya jahit/bahan
-- ekstra riil, bukan skala proporsional seperti size_multiplier.
CREATE TABLE garment_variants (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code          VARCHAR(30) NOT NULL UNIQUE,
    name          VARCHAR(100) NOT NULL,
    price_addon   NUMERIC(18,2) NOT NULL DEFAULT 0,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_garment_variants_updated_at BEFORE UPDATE ON garment_variants
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

INSERT INTO garment_variants (code, name, price_addon, sort_order)
VALUES ('NORMAL', 'Normal', 0, 0);

-- Ukuran generik untuk configurator, sengaja dilepas dari product_sizes
-- (yang terikat wajib ke satu product_id) karena tidak ada lagi Product
-- tetap per desain di alur ini.
CREATE TABLE garment_sizes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    size_code       VARCHAR(10) NOT NULL UNIQUE,
    size_multiplier NUMERIC(6,4) NOT NULL DEFAULT 1.0000,
    sort_order      INTEGER NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_garment_sizes_updated_at BEFORE UPDATE ON garment_sizes
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Katalog desain yang bisa dipilih customer, dengan sample gambar.
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

-- Bahan: harga jual ke customer, ditautkan ke materials (inventory/BOM cost
-- basis) supaya konsisten dengan modul produksi nanti.
CREATE TABLE fabrics (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code          VARCHAR(30) NOT NULL UNIQUE,
    material_id   UUID NOT NULL REFERENCES materials(id),
    name          VARCHAR(150) NOT NULL,
    sales_price   NUMERIC(18,2) NOT NULL DEFAULT 0,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_fabrics_material_id ON fabrics(material_id);
CREATE TRIGGER trg_fabrics_updated_at BEFORE UPDATE ON fabrics
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Link token per customer (tanpa akun/login). Hanya token_hash yang
-- disimpan -- plaintext token cuma dikirim sekali ke staf saat dibuat,
-- persis pola refresh token di internal/auth.
CREATE TABLE order_links (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash    VARCHAR(64) NOT NULL UNIQUE,
    customer_id   UUID NOT NULL REFERENCES customers(id),
    created_by    UUID NOT NULL REFERENCES users(id),
    status        VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                      CHECK (status IN ('ACTIVE','REVOKED')),
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_order_links_customer_id ON order_links(customer_id);

-- Draft pesanan hasil konfigurasi customer -- belum jadi Sales Order resmi
-- sampai staf sales me-review dan mengonfirmasinya.
CREATE TABLE order_quotations (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quotation_number     VARCHAR(40) NOT NULL UNIQUE,
    order_link_id        UUID NOT NULL REFERENCES order_links(id),
    customer_id          UUID NOT NULL REFERENCES customers(id),
    status               VARCHAR(20) NOT NULL DEFAULT 'PENDING_REVIEW'
                             CHECK (status IN ('PENDING_REVIEW','CONFIRMED','REJECTED','CANCELLED')),
    estimated_subtotal   NUMERIC(18,2) NOT NULL DEFAULT 0,
    notes                TEXT,
    reviewed_by          UUID REFERENCES users(id),
    reviewed_at          TIMESTAMPTZ,
    sales_order_id       UUID REFERENCES sales_orders(id),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_order_quotations_status ON order_quotations(status);
CREATE INDEX idx_order_quotations_customer_id ON order_quotations(customer_id);
CREATE TRIGGER trg_order_quotations_updated_at BEFORE UPDATE ON order_quotations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE order_quotation_items (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quotation_id      UUID NOT NULL REFERENCES order_quotations(id) ON DELETE CASCADE,
    design_id         UUID NOT NULL REFERENCES designs(id),
    fabric_id         UUID NOT NULL REFERENCES fabrics(id),
    garment_size_id   UUID NOT NULL REFERENCES garment_sizes(id),
    variant_id        UUID NOT NULL REFERENCES garment_variants(id),
    qty               NUMERIC(18,2) NOT NULL CHECK (qty > 0),
    unit_price        NUMERIC(18,2) NOT NULL CHECK (unit_price >= 0),
    line_total        NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (quotation_id, design_id, fabric_id, garment_size_id, variant_id)
);
CREATE INDEX idx_order_quotation_items_quotation_id ON order_quotation_items(quotation_id);

-- Tautan balik dari baris SO yang dihasilkan lewat konfirmasi quotation ke
-- baris konfigurasi aslinya, supaya detail desain/bahan/varian pilihan
-- customer tetap tertelusur walau sales_order_items sendiri tidak diubah
-- strukturnya (tetap product_id/product_size_id seperti sebelumnya --
-- lihat catatan materialisasi produk di internal/quotation).
ALTER TABLE sales_order_items
    ADD COLUMN quotation_item_id UUID REFERENCES order_quotation_items(id);
