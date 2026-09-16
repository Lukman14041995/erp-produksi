
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS journal_entry_lines;
DROP TABLE IF EXISTS journal_entries CASCADE;
DROP TABLE IF EXISTS inventory_balances;
DROP TABLE IF EXISTS inventory_transactions;
DROP TABLE IF EXISTS production_costs;
DROP TABLE IF EXISTS production_overheads;
DROP TABLE IF EXISTS production_labor;
DROP TABLE IF EXISTS production_materials;
DROP TABLE IF EXISTS production_order_items;
DROP TABLE IF EXISTS production_orders;
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS payment_allocations;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS invoice_items;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS sales_order_items;
DROP TABLE IF EXISTS sales_orders;
DROP TABLE IF EXISTS bom_lines;
DROP TABLE IF EXISTS bom_headers;
DROP TABLE IF EXISTS accounting_periods;
DROP TABLE IF EXISTS document_sequences;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS materials;
DROP TABLE IF EXISTS product_sizes;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS suppliers;
DROP TABLE IF EXISTS customers;

DROP FUNCTION IF EXISTS set_updated_at();

