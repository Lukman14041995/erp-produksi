package quotation

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/db"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ---- Order links ----

const linkColumns = `id, customer_id, created_by, status, expires_at, created_at`

func scanLink(row pgx.Row) (OrderLink, error) {
	var l OrderLink
	err := row.Scan(&l.ID, &l.CustomerID, &l.CreatedBy, &l.Status, &l.ExpiresAt, &l.CreatedAt)
	return l, err
}

func (r *Repository) InsertLink(ctx context.Context, customerID, createdBy uuid.UUID, tokenHash string, expiresAt time.Time) (OrderLink, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO order_links (token_hash, customer_id, created_by, expires_at)
		VALUES ($1,$2,$3,$4)
		RETURNING `+linkColumns, tokenHash, customerID, createdBy, expiresAt)
	return scanLink(row)
}

func (r *Repository) GetLinkByTokenHash(ctx context.Context, tokenHash string) (OrderLink, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+linkColumns+` FROM order_links WHERE token_hash=$1`, tokenHash)
	return scanLink(row)
}

func (r *Repository) GetLinkByID(ctx context.Context, id uuid.UUID) (OrderLink, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+linkColumns+` FROM order_links WHERE id=$1`, id)
	return scanLink(row)
}

func (r *Repository) ListLinksByCustomer(ctx context.Context, customerID uuid.UUID) ([]OrderLink, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+linkColumns+` FROM order_links WHERE customer_id=$1 ORDER BY created_at DESC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []OrderLink{}
	for rows.Next() {
		l, err := scanLink(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *Repository) SetLinkStatus(ctx context.Context, id uuid.UUID, status LinkStatus) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE order_links SET status=$2 WHERE id=$1`, id, status)
	return err
}

// ---- Quotations ----

const quotationColumns = `id, quotation_number, order_link_id, customer_id, status, estimated_subtotal, notes,
	reviewed_by, reviewed_at, sales_order_id, created_at, updated_at`

func scanQuotation(row pgx.Row) (Quotation, error) {
	var q Quotation
	err := row.Scan(&q.ID, &q.QuotationNumber, &q.OrderLinkID, &q.CustomerID, &q.Status, &q.EstimatedSubtotal, &q.Notes,
		&q.ReviewedBy, &q.ReviewedAt, &q.SalesOrderID, &q.CreatedAt, &q.UpdatedAt)
	return q, err
}

func (r *Repository) InsertQuotation(ctx context.Context, q Quotation) (Quotation, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO order_quotations (quotation_number, order_link_id, customer_id, notes, estimated_subtotal)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING `+quotationColumns, q.QuotationNumber, q.OrderLinkID, q.CustomerID, q.Notes, q.EstimatedSubtotal)
	return scanQuotation(row)
}

func (r *Repository) GetQuotationByID(ctx context.Context, id uuid.UUID) (Quotation, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+quotationColumns+` FROM order_quotations WHERE id=$1`, id)
	return scanQuotation(row)
}

// GetQuotationForUpdate row-locks the quotation so concurrent confirm/reject
// commands against the same quotation serialize, mirroring
// sales.Repository.GetOrderForUpdate.
func (r *Repository) GetQuotationForUpdate(ctx context.Context, id uuid.UUID) (Quotation, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+quotationColumns+` FROM order_quotations WHERE id=$1 FOR UPDATE`, id)
	return scanQuotation(row)
}

func (r *Repository) ListQuotations(ctx context.Context, status Status, limit, offset int) ([]Quotation, error) {
	var rows pgx.Rows
	var err error
	if status == "" {
		rows, err = db.Q(ctx, r.pool).Query(ctx, `
			SELECT `+quotationColumns+` FROM order_quotations ORDER BY created_at DESC LIMIT $1 OFFSET $2
		`, limit, offset)
	} else {
		rows, err = db.Q(ctx, r.pool).Query(ctx, `
			SELECT `+quotationColumns+` FROM order_quotations WHERE status=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3
		`, status, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Quotation{}
	for rows.Next() {
		item, err := scanQuotation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ListQuotationsByCustomer powers the customer's own order-list page: every
// quotation they've ever submitted, across any order link, newest first.
func (r *Repository) ListQuotationsByCustomer(ctx context.Context, customerID uuid.UUID) ([]Quotation, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT `+quotationColumns+` FROM order_quotations WHERE customer_id=$1 ORDER BY created_at DESC
	`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Quotation{}
	for rows.Next() {
		q, err := scanQuotation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

func (r *Repository) GetQuotationBySalesOrderID(ctx context.Context, salesOrderID uuid.UUID) (Quotation, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+quotationColumns+` FROM order_quotations WHERE sales_order_id=$1`, salesOrderID)
	return scanQuotation(row)
}

func (r *Repository) UpdateQuotationReview(ctx context.Context, id uuid.UUID, status Status, reviewedBy uuid.UUID, salesOrderID *uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		UPDATE order_quotations SET status=$2, reviewed_by=$3, reviewed_at=now(), sales_order_id=$4
		WHERE id=$1
	`, id, status, reviewedBy, salesOrderID)
	return err
}

// ---- Quotation items ----

const quotationItemColumns = `id, quotation_id, product_type_id, fabric_id, garment_size_id, variant_id, ink_id, qty, unit_price, line_total, created_at`

func scanQuotationItem(row pgx.Row) (QuotationItem, error) {
	var i QuotationItem
	err := row.Scan(&i.ID, &i.QuotationID, &i.ProductTypeID, &i.FabricID, &i.GarmentSizeID, &i.VariantID, &i.InkID, &i.Qty, &i.UnitPrice, &i.LineTotal, &i.CreatedAt)
	return i, err
}

func (r *Repository) InsertQuotationItem(ctx context.Context, i QuotationItem) (QuotationItem, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO order_quotation_items (quotation_id, product_type_id, fabric_id, garment_size_id, variant_id, ink_id, qty, unit_price, line_total)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING `+quotationItemColumns, i.QuotationID, i.ProductTypeID, i.FabricID, i.GarmentSizeID, i.VariantID, i.InkID, i.Qty, i.UnitPrice, i.LineTotal)
	return scanQuotationItem(row)
}

func (r *Repository) ListQuotationItems(ctx context.Context, quotationID uuid.UUID) ([]QuotationItem, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+quotationItemColumns+` FROM order_quotation_items WHERE quotation_id=$1`, quotationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []QuotationItem{}
	for rows.Next() {
		i, err := scanQuotationItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}
