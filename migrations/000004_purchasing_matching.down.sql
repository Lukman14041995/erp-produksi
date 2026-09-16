
ALTER TABLE supplier_invoice_items DROP COLUMN IF EXISTS price_variance;
ALTER TABLE supplier_invoice_items DROP COLUMN IF EXISTS goods_receipt_item_id;
ALTER TABLE supplier_invoice_items DROP COLUMN IF EXISTS purchase_order_item_id;
ALTER TABLE supplier_invoices DROP COLUMN IF EXISTS purchase_order_id;

DROP TABLE IF EXISTS goods_receipt_items;
DROP TABLE IF EXISTS goods_receipts;
DROP TABLE IF EXISTS purchase_order_items;
DROP TABLE IF EXISTS purchase_orders;

DELETE FROM accounts WHERE code IN ('2-1100','6-9100');

ALTER TABLE journal_entries DROP CONSTRAINT journal_entries_source_type_check;
ALTER TABLE journal_entries ADD CONSTRAINT journal_entries_source_type_check CHECK (source_type IN
    ('SALES_INVOICE','PAYMENT_RECEIPT','PAYMENT_DISBURSEMENT',
     'MATERIAL_ISSUE','LABOR_COST','OVERHEAD_COST','PRODUCTION_COMPLETION','COGS','EXPENSE',
     'SUPPLIER_BILL','MANUAL','REVERSAL'));

