package purchasing

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

const billColumns = `id, bill_number, supplier_id, purchase_order_id, bill_date, due_date, debit_account_code, status,
	subtotal, tax_total, grand_total, paid_amount, balance_due, journal_id, notes, created_at, updated_at`

func scanBill(row pgx.Row) (SupplierInvoice, error) {
	var b SupplierInvoice
	err := row.Scan(&b.ID, &b.BillNumber, &b.SupplierID, &b.PurchaseOrderID, &b.BillDate, &b.DueDate, &b.DebitAccountCode, &b.Status,
		&b.Subtotal, &b.TaxTotal, &b.GrandTotal, &b.PaidAmount, &b.BalanceDue, &b.JournalID, &b.Notes, &b.CreatedAt, &b.UpdatedAt)
	return b, err
}

func (r *Repository) InsertBill(ctx context.Context, b SupplierInvoice) (SupplierInvoice, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO supplier_invoices (bill_number, supplier_id, purchase_order_id, bill_date, due_date, debit_account_code, subtotal, tax_total, grand_total, balance_due, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING `+billColumns, b.BillNumber, b.SupplierID, b.PurchaseOrderID, b.BillDate, b.DueDate, b.DebitAccountCode, b.Subtotal, b.TaxTotal, b.GrandTotal, b.GrandTotal, b.Notes)
	return scanBill(row)
}

func (r *Repository) InsertBillItem(ctx context.Context, i SupplierInvoiceItem) (SupplierInvoiceItem, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO supplier_invoice_items (supplier_invoice_id, material_id, qty, unit_cost, line_total, purchase_order_item_id, goods_receipt_item_id, price_variance)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, supplier_invoice_id, material_id, qty, unit_cost, line_total, purchase_order_item_id, goods_receipt_item_id, price_variance
	`, i.SupplierInvoiceID, i.MaterialID, i.Qty, i.UnitCost, i.LineTotal, i.PurchaseOrderItemID, i.GoodsReceiptItemID, i.PriceVariance)
	var out SupplierInvoiceItem
	err := row.Scan(&out.ID, &out.SupplierInvoiceID, &out.MaterialID, &out.Qty, &out.UnitCost, &out.LineTotal, &out.PurchaseOrderItemID, &out.GoodsReceiptItemID, &out.PriceVariance)
	return out, err
}

func (r *Repository) ListBillItems(ctx context.Context, billID uuid.UUID) ([]SupplierInvoiceItem, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT id, supplier_invoice_id, material_id, qty, unit_cost, line_total, purchase_order_item_id, goods_receipt_item_id, price_variance
		FROM supplier_invoice_items WHERE supplier_invoice_id=$1
	`, billID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SupplierInvoiceItem{}
	for rows.Next() {
		var i SupplierInvoiceItem
		if err := rows.Scan(&i.ID, &i.SupplierInvoiceID, &i.MaterialID, &i.Qty, &i.UnitCost, &i.LineTotal, &i.PurchaseOrderItemID, &i.GoodsReceiptItemID, &i.PriceVariance); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *Repository) GetBillForUpdate(ctx context.Context, id uuid.UUID) (SupplierInvoice, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+billColumns+` FROM supplier_invoices WHERE id=$1 FOR UPDATE`, id)
	return scanBill(row)
}

func (r *Repository) GetBillByID(ctx context.Context, id uuid.UUID) (SupplierInvoice, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+billColumns+` FROM supplier_invoices WHERE id=$1`, id)
	return scanBill(row)
}

func (r *Repository) ListBills(ctx context.Context, limit, offset int) ([]SupplierInvoice, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+billColumns+` FROM supplier_invoices ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SupplierInvoice{}
	for rows.Next() {
		b, err := scanBill(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *Repository) SetBillPosted(ctx context.Context, id, journalID uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE supplier_invoices SET status='POSTED', journal_id=$2 WHERE id=$1`, id, journalID)
	return err
}

// SetBillTotalsAndPosted persists the totals computed during 3-way matching
// (unknown until all matched lines are processed) and marks the bill posted
// in a single update.
func (r *Repository) SetBillTotalsAndPosted(ctx context.Context, id uuid.UUID, subtotal, grandTotal decimal.Decimal, journalID uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		UPDATE supplier_invoices SET subtotal=$2, grand_total=$3, balance_due=$3, status='POSTED', journal_id=$4 WHERE id=$1
	`, id, subtotal, grandTotal, journalID)
	return err
}

func (r *Repository) SetBillVoid(ctx context.Context, id uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE supplier_invoices SET status='VOID', voided_at=now() WHERE id=$1`, id)
	return err
}

func (r *Repository) UpdateBillPayment(ctx context.Context, id uuid.UUID, paidAmount, balanceDue decimal.Decimal, status BillStatus) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		UPDATE supplier_invoices SET paid_amount=$2, balance_due=$3, status=$4 WHERE id=$1
	`, id, paidAmount, balanceDue, status)
	return err
}

// ---- Purchase Orders ----

const poColumns = `id, po_number, supplier_id, order_date, expected_date, status, billing_status, subtotal, grand_total, notes, approved_at, cancelled_at, created_at, updated_at`

func scanPO(row pgx.Row) (PurchaseOrder, error) {
	var p PurchaseOrder
	err := row.Scan(&p.ID, &p.PONumber, &p.SupplierID, &p.OrderDate, &p.ExpectedDate, &p.Status, &p.BillingStatus,
		&p.Subtotal, &p.GrandTotal, &p.Notes, &p.ApprovedAt, &p.CancelledAt, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *Repository) InsertPO(ctx context.Context, p PurchaseOrder) (PurchaseOrder, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO purchase_orders (po_number, supplier_id, order_date, expected_date, subtotal, grand_total, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING `+poColumns, p.PONumber, p.SupplierID, p.OrderDate, p.ExpectedDate, p.Subtotal, p.GrandTotal, p.Notes)
	return scanPO(row)
}

func (r *Repository) InsertPOItem(ctx context.Context, i PurchaseOrderItem) (PurchaseOrderItem, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO purchase_order_items (purchase_order_id, material_id, qty, unit_cost, line_total)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, purchase_order_id, material_id, qty, unit_cost, qty_received, qty_billed, line_total
	`, i.PurchaseOrderID, i.MaterialID, i.Qty, i.UnitCost, i.LineTotal)
	var out PurchaseOrderItem
	err := row.Scan(&out.ID, &out.PurchaseOrderID, &out.MaterialID, &out.Qty, &out.UnitCost, &out.QtyReceived, &out.QtyBilled, &out.LineTotal)
	return out, err
}

func (r *Repository) ListPOItems(ctx context.Context, poID uuid.UUID) ([]PurchaseOrderItem, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT id, purchase_order_id, material_id, qty, unit_cost, qty_received, qty_billed, line_total
		FROM purchase_order_items WHERE purchase_order_id=$1
	`, poID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PurchaseOrderItem{}
	for rows.Next() {
		var i PurchaseOrderItem
		if err := rows.Scan(&i.ID, &i.PurchaseOrderID, &i.MaterialID, &i.Qty, &i.UnitCost, &i.QtyReceived, &i.QtyBilled, &i.LineTotal); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *Repository) GetPOItemForUpdate(ctx context.Context, id uuid.UUID) (PurchaseOrderItem, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		SELECT id, purchase_order_id, material_id, qty, unit_cost, qty_received, qty_billed, line_total
		FROM purchase_order_items WHERE id=$1 FOR UPDATE
	`, id)
	var i PurchaseOrderItem
	err := row.Scan(&i.ID, &i.PurchaseOrderID, &i.MaterialID, &i.Qty, &i.UnitCost, &i.QtyReceived, &i.QtyBilled, &i.LineTotal)
	return i, err
}

func (r *Repository) UpdatePOItemReceived(ctx context.Context, id uuid.UUID, qtyReceived decimal.Decimal) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE purchase_order_items SET qty_received=$2 WHERE id=$1`, id, qtyReceived)
	return err
}

func (r *Repository) UpdatePOItemBilled(ctx context.Context, id uuid.UUID, qtyBilled decimal.Decimal) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE purchase_order_items SET qty_billed=$2 WHERE id=$1`, id, qtyBilled)
	return err
}

func (r *Repository) GetPOForUpdate(ctx context.Context, id uuid.UUID) (PurchaseOrder, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+poColumns+` FROM purchase_orders WHERE id=$1 FOR UPDATE`, id)
	return scanPO(row)
}

func (r *Repository) GetPOByID(ctx context.Context, id uuid.UUID) (PurchaseOrder, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+poColumns+` FROM purchase_orders WHERE id=$1`, id)
	return scanPO(row)
}

func (r *Repository) ListPOs(ctx context.Context, limit, offset int) ([]PurchaseOrder, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+poColumns+` FROM purchase_orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PurchaseOrder{}
	for rows.Next() {
		p, err := scanPO(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) SetPOStatus(ctx context.Context, id uuid.UUID, status POStatus) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE purchase_orders SET status=$2 WHERE id=$1`, id, status)
	return err
}

func (r *Repository) SetPOApprovedAt(ctx context.Context, id uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE purchase_orders SET approved_at=now() WHERE id=$1`, id)
	return err
}

func (r *Repository) SetPOCancelledAt(ctx context.Context, id uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE purchase_orders SET cancelled_at=now() WHERE id=$1`, id)
	return err
}

func (r *Repository) SetPOBillingStatus(ctx context.Context, id uuid.UUID, status POBillingStatus) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE purchase_orders SET billing_status=$2 WHERE id=$1`, id, status)
	return err
}

// ---- Goods Receipts ----

const grnColumns = `id, grn_number, purchase_order_id, supplier_id, receipt_date, journal_id, notes, created_at`

func scanGRN(row pgx.Row) (GoodsReceipt, error) {
	var g GoodsReceipt
	err := row.Scan(&g.ID, &g.GRNNumber, &g.PurchaseOrderID, &g.SupplierID, &g.ReceiptDate, &g.JournalID, &g.Notes, &g.CreatedAt)
	return g, err
}

func (r *Repository) InsertGRN(ctx context.Context, g GoodsReceipt) (GoodsReceipt, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO goods_receipts (grn_number, purchase_order_id, supplier_id, receipt_date, notes)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING `+grnColumns, g.GRNNumber, g.PurchaseOrderID, g.SupplierID, g.ReceiptDate, g.Notes)
	return scanGRN(row)
}

func (r *Repository) SetGRNJournal(ctx context.Context, id, journalID uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE goods_receipts SET journal_id=$2 WHERE id=$1`, id, journalID)
	return err
}

func (r *Repository) InsertGRNItem(ctx context.Context, i GoodsReceiptItem) (GoodsReceiptItem, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO goods_receipt_items (goods_receipt_id, purchase_order_item_id, material_id, qty_received, unit_cost, line_total)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, goods_receipt_id, purchase_order_item_id, material_id, qty_received, unit_cost, qty_billed, line_total
	`, i.GoodsReceiptID, i.PurchaseOrderItemID, i.MaterialID, i.QtyReceived, i.UnitCost, i.LineTotal)
	var out GoodsReceiptItem
	err := row.Scan(&out.ID, &out.GoodsReceiptID, &out.PurchaseOrderItemID, &out.MaterialID, &out.QtyReceived, &out.UnitCost, &out.QtyBilled, &out.LineTotal)
	return out, err
}

func (r *Repository) ListGRNItems(ctx context.Context, grnID uuid.UUID) ([]GoodsReceiptItem, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT id, goods_receipt_id, purchase_order_item_id, material_id, qty_received, unit_cost, qty_billed, line_total
		FROM goods_receipt_items WHERE goods_receipt_id=$1
	`, grnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []GoodsReceiptItem{}
	for rows.Next() {
		var i GoodsReceiptItem
		if err := rows.Scan(&i.ID, &i.GoodsReceiptID, &i.PurchaseOrderItemID, &i.MaterialID, &i.QtyReceived, &i.UnitCost, &i.QtyBilled, &i.LineTotal); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *Repository) GetGRNItemForUpdate(ctx context.Context, id uuid.UUID) (GoodsReceiptItem, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		SELECT id, goods_receipt_id, purchase_order_item_id, material_id, qty_received, unit_cost, qty_billed, line_total
		FROM goods_receipt_items WHERE id=$1 FOR UPDATE
	`, id)
	var i GoodsReceiptItem
	err := row.Scan(&i.ID, &i.GoodsReceiptID, &i.PurchaseOrderItemID, &i.MaterialID, &i.QtyReceived, &i.UnitCost, &i.QtyBilled, &i.LineTotal)
	return i, err
}

func (r *Repository) UpdateGRNItemBilled(ctx context.Context, id uuid.UUID, qtyBilled decimal.Decimal) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE goods_receipt_items SET qty_billed=$2 WHERE id=$1`, id, qtyBilled)
	return err
}

func (r *Repository) GetGRNByID(ctx context.Context, id uuid.UUID) (GoodsReceipt, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+grnColumns+` FROM goods_receipts WHERE id=$1`, id)
	return scanGRN(row)
}

func (r *Repository) ListGRNsByPO(ctx context.Context, poID uuid.UUID) ([]GoodsReceipt, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+grnColumns+` FROM goods_receipts WHERE purchase_order_id=$1 ORDER BY created_at`, poID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []GoodsReceipt{}
	for rows.Next() {
		g, err := scanGRN(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *Repository) ListGRNs(ctx context.Context, limit, offset int) ([]GoodsReceipt, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+grnColumns+` FROM goods_receipts ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []GoodsReceipt{}
	for rows.Next() {
		g, err := scanGRN(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}
