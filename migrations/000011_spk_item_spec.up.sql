-- ============================================================================
-- Production needs to see WHAT they're making, not just an SPK number and
-- qty. Denormalize the customer's configurator choices (bahan/model/tinta/
-- ukuran) onto spk_items at SPK-generation time, the same way spk_orders
-- already denormalizes so_number/customer_name -- avoids spk depending on
-- the quotation package (which already depends on spk) just to join catalog
-- names at read time. See internal/spk/service.go GenerateForOrder and
-- internal/quotation/service.go ConfirmQuotation.
-- ============================================================================

ALTER TABLE spk_items
    ADD COLUMN fabric_name  TEXT NOT NULL DEFAULT '',
    ADD COLUMN variant_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN ink_name     TEXT NOT NULL DEFAULT '',
    ADD COLUMN size_code    TEXT NOT NULL DEFAULT '';

-- Cached one-line summary of the SPK's composition for the list view (e.g.
-- "Cotton Combed · Normal · Rubber", or "... +2 lainnya" when the
-- order mixes more than one fabric/model/ink combination), so
-- /production/orders/tshirt|jersey doesn't need to join items per row.
ALTER TABLE spk_orders ADD COLUMN composition_summary TEXT NOT NULL DEFAULT '';
