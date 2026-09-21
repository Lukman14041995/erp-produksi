-- Denormalize the originating quotation's id/number onto spk_orders, same
-- pattern as so_number/customer_name -- lets production and sales trace an
-- SPK back to the customer's original quotation without a join, and shows
-- it directly on /production/orders/tshirt|jersey.
ALTER TABLE spk_orders
    ADD COLUMN quotation_id     UUID REFERENCES order_quotations(id),
    ADD COLUMN quotation_number VARCHAR(40) NOT NULL DEFAULT '';
