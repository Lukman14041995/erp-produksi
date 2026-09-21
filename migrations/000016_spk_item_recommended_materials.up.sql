-- Carries the Material each line's Fabric/Ink is linked to (fabrics.material_id
-- / inks.material_id) onto spk_items, so the "Bahan Baku Terpakai" picker on
-- the SPK detail page can surface "bahan yang sudah dipilih customer" (e.g.
-- Cotton fabric, Rubber ink) as quick picks instead of the whole Materials
-- master -- see internal/spk/service.go GenerateForOrder and
-- internal/quotation/service.go ConfirmQuotation.
ALTER TABLE spk_items
    ADD COLUMN fabric_material_id UUID REFERENCES materials(id),
    ADD COLUMN ink_material_id    UUID REFERENCES materials(id);
