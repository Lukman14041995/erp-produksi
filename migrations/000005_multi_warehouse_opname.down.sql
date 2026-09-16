DROP TABLE IF EXISTS stock_opname_items;
DROP TABLE IF EXISTS stock_opnames;
DROP TABLE IF EXISTS stock_transfer_items;
DROP TABLE IF EXISTS stock_transfers;

ALTER TABLE inventory_transactions DROP CONSTRAINT IF EXISTS inventory_transactions_txn_type_check;
ALTER TABLE inventory_transactions ADD CONSTRAINT inventory_transactions_txn_type_check CHECK (txn_type IN
    ('PURCHASE','PRODUCTION_ISSUE','PRODUCTION_RECEIPT','SALE','ADJUSTMENT'));

ALTER TABLE inventory_balances DROP CONSTRAINT IF EXISTS inventory_balances_unique_per_warehouse;
ALTER TABLE inventory_balances ADD CONSTRAINT inventory_balances_item_type_material_id_product_id_product_key
    UNIQUE (item_type, material_id, product_id, product_size_id);
ALTER TABLE inventory_balances DROP COLUMN IF EXISTS warehouse_id;

ALTER TABLE inventory_transactions DROP COLUMN IF EXISTS warehouse_id;

DROP TABLE IF EXISTS warehouses;

DELETE FROM accounts WHERE code = '5-9200';

ALTER TABLE journal_entries DROP CONSTRAINT journal_entries_source_type_check;
ALTER TABLE journal_entries ADD CONSTRAINT journal_entries_source_type_check CHECK (source_type IN
    ('SALES_INVOICE','PAYMENT_RECEIPT','PAYMENT_DISBURSEMENT',
     'MATERIAL_ISSUE','LABOR_COST','OVERHEAD_COST','PRODUCTION_COMPLETION','COGS','EXPENSE',
     'SUPPLIER_BILL','GOODS_RECEIPT','MANUAL','REVERSAL'));
