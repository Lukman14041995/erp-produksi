package sales

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/shopspring/decimal"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const orderColumns = `id, so_number, customer_id, order_date, order_status, payment_status, production_status, delivery_status,
	subtotal, discount_total, tax_total, grand_total, notes, created_by, created_at, updated_at`

func scanOrder(row pgx.Row) (SalesOrder, error) {
	var o SalesOrder
	err := row.Scan(&o.ID, &o.SONumber, &o.CustomerID, &o.OrderDate, &o.OrderStatus, &o.PaymentStatus, &o.ProductionStatus, &o.DeliveryStatus,
		&o.Subtotal, &o.DiscountTotal, &o.TaxTotal, &o.GrandTotal, &o.Notes, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}

func (r *Repository) InsertOrder(ctx context.Context, o SalesOrder) (SalesOrder, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO sales_orders (so_number, customer_id, order_date, notes, subtotal, discount_total, tax_total, grand_total, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING `+orderColumns, o.SONumber, o.CustomerID, o.OrderDate, o.Notes, o.Subtotal, o.DiscountTotal, o.TaxTotal, o.GrandTotal, o.CreatedBy)
	return scanOrder(row)
}

func (r *Repository) InsertOrderItem(ctx context.Context, i SalesOrderItem) (SalesOrderItem, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO sales_order_items (sales_order_id, product_id, product_size_id, qty, unit_price, discount, tax_rate, line_total)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, sales_order_id, product_id, product_size_id, qty, unit_price, discount, tax_rate, line_total
	`, i.SalesOrderID, i.ProductID, i.ProductSizeID, i.Qty, i.UnitPrice, i.Discount, i.TaxRate, i.LineTotal)
	var out SalesOrderItem
	err := row.Scan(&out.ID, &out.SalesOrderID, &out.ProductID, &out.ProductSizeID, &out.Qty, &out.UnitPrice, &out.Discount, &out.TaxRate, &out.LineTotal)
	return out, err
}

func (r *Repository) GetOrderByID(ctx context.Context, id uuid.UUID) (SalesOrder, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+orderColumns+` FROM sales_orders WHERE id=$1`, id)
	return scanOrder(row)
}

// GetOrderForUpdate row-locks the sales order so concurrent commands
// (confirm, cancel, create-invoice, deliver) against the same order
// serialize instead of racing on its status columns.
func (r *Repository) GetOrderForUpdate(ctx context.Context, id uuid.UUID) (SalesOrder, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+orderColumns+` FROM sales_orders WHERE id=$1 FOR UPDATE`, id)
	return scanOrder(row)
}

func (r *Repository) ListOrderItems(ctx context.Context, orderID uuid.UUID) ([]SalesOrderItem, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT id, sales_order_id, product_id, product_size_id, qty, unit_price, discount, tax_rate, line_total
		FROM sales_order_items WHERE sales_order_id=$1
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SalesOrderItem{}
	for rows.Next() {
		var i SalesOrderItem
		if err := rows.Scan(&i.ID, &i.SalesOrderID, &i.ProductID, &i.ProductSizeID, &i.Qty, &i.UnitPrice, &i.Discount, &i.TaxRate, &i.LineTotal); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *Repository) ListOrders(ctx context.Context, limit, offset int) ([]SalesOrder, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+orderColumns+` FROM sales_orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SalesOrder{}
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *Repository) SetOrderStatus(ctx context.Context, id uuid.UUID, status OrderStatus) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE sales_orders SET order_status=$2 WHERE id=$1`, id, status)
	return err
}

func (r *Repository) SetConfirmedAt(ctx context.Context, id uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE sales_orders SET confirmed_at = now() WHERE id=$1`, id)
	return err
}

func (r *Repository) SetCancelledAt(ctx context.Context, id uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE sales_orders SET cancelled_at = now() WHERE id=$1`, id)
	return err
}

func (r *Repository) SetProductionStatus(ctx context.Context, id uuid.UUID, status ProductionStatus) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE sales_orders SET production_status=$2 WHERE id=$1`, id, status)
	return err
}

func (r *Repository) SetDeliveryStatus(ctx context.Context, id uuid.UUID, status DeliveryStatus) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE sales_orders SET delivery_status=$2 WHERE id=$1`, id, status)
	return err
}

func (r *Repository) SetPaymentStatus(ctx context.Context, id uuid.UUID, status PaymentStatus) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE sales_orders SET payment_status=$2 WHERE id=$1`, id, status)
	return err
}

// ---- Invoices ----

const invoiceColumns = `id, invoice_number, sales_order_id, customer_id, invoice_date, due_date, status,
	subtotal, discount_total, tax_total, grand_total, paid_amount, balance_due, journal_id, created_at, updated_at`

func scanInvoice(row pgx.Row) (Invoice, error) {
	var i Invoice
	err := row.Scan(&i.ID, &i.InvoiceNumber, &i.SalesOrderID, &i.CustomerID, &i.InvoiceDate, &i.DueDate, &i.Status,
		&i.Subtotal, &i.DiscountTotal, &i.TaxTotal, &i.GrandTotal, &i.PaidAmount, &i.BalanceDue, &i.JournalID, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

func (r *Repository) InsertInvoice(ctx context.Context, i Invoice) (Invoice, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO invoices (invoice_number, sales_order_id, customer_id, invoice_date, due_date, status, subtotal, discount_total, tax_total, grand_total, paid_amount, balance_due, journal_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING `+invoiceColumns, i.InvoiceNumber, i.SalesOrderID, i.CustomerID, i.InvoiceDate, i.DueDate, i.Status,
		i.Subtotal, i.DiscountTotal, i.TaxTotal, i.GrandTotal, i.PaidAmount, i.BalanceDue, i.JournalID)
	return scanInvoice(row)
}

func (r *Repository) InsertInvoiceItem(ctx context.Context, i InvoiceItem) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		INSERT INTO invoice_items (invoice_id, sales_order_item_id, product_id, product_size_id, qty, unit_price, discount, tax_rate, line_total)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, i.InvoiceID, i.SalesOrderItemID, i.ProductID, i.ProductSizeID, i.Qty, i.UnitPrice, i.Discount, i.TaxRate, i.LineTotal)
	return err
}

func (r *Repository) GetInvoiceByID(ctx context.Context, id uuid.UUID) (Invoice, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+invoiceColumns+` FROM invoices WHERE id=$1`, id)
	return scanInvoice(row)
}

func (r *Repository) GetInvoiceForUpdate(ctx context.Context, id uuid.UUID) (Invoice, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+invoiceColumns+` FROM invoices WHERE id=$1 FOR UPDATE`, id)
	return scanInvoice(row)
}

func (r *Repository) ListInvoiceItems(ctx context.Context, invoiceID uuid.UUID) ([]InvoiceItem, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT id, invoice_id, sales_order_item_id, product_id, product_size_id, qty, unit_price, discount, tax_rate, line_total
		FROM invoice_items WHERE invoice_id=$1
	`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []InvoiceItem{}
	for rows.Next() {
		var i InvoiceItem
		if err := rows.Scan(&i.ID, &i.InvoiceID, &i.SalesOrderItemID, &i.ProductID, &i.ProductSizeID, &i.Qty, &i.UnitPrice, &i.Discount, &i.TaxRate, &i.LineTotal); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *Repository) ListInvoicesByOrder(ctx context.Context, orderID uuid.UUID) ([]Invoice, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+invoiceColumns+` FROM invoices WHERE sales_order_id=$1 ORDER BY created_at`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Invoice{}
	for rows.Next() {
		i, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *Repository) ListInvoices(ctx context.Context, limit, offset int) ([]Invoice, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+invoiceColumns+` FROM invoices ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Invoice{}
	for rows.Next() {
		i, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *Repository) SetInvoiceJournal(ctx context.Context, id, journalID uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE invoices SET journal_id=$2 WHERE id=$1`, id, journalID)
	return err
}

func (r *Repository) UpdateInvoicePayment(ctx context.Context, id uuid.UUID, paidAmount, balanceDue decimal.Decimal, status InvoiceStatus) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		UPDATE invoices SET paid_amount=$2, balance_due=$3, status=$4 WHERE id=$1
	`, id, paidAmount, balanceDue, status)
	return err
}

// AROutstandingByCustomer sums balance_due across POSTED/PARTIALLY_PAID invoices, used by AR aging.
func (r *Repository) ListOpenInvoicesByCustomer(ctx context.Context, customerID uuid.UUID) ([]Invoice, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT `+invoiceColumns+` FROM invoices
		WHERE customer_id=$1 AND status IN ('POSTED','PARTIALLY_PAID')
		ORDER BY invoice_date
	`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Invoice{}
	for rows.Next() {
		i, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}
