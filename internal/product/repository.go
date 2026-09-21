package product

import (
	"context"

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

const productColumns = `id, code, name, category, uom, base_price, is_active, created_at, updated_at`

func scanProduct(row pgx.Row) (Product, error) {
	var p Product
	err := row.Scan(&p.ID, &p.Code, &p.Name, &p.Category, &p.UOM, &p.BasePrice, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *Repository) Create(ctx context.Context, in UpsertProductInput) (Product, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO products (code, name, category, uom, base_price)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING `+productColumns, in.Code, in.Name, in.Category, in.UOM, in.BasePrice)
	return scanProduct(row)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, in UpsertProductInput) (Product, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE products SET name=$2, category=$3, uom=$4, base_price=$5, is_active=$6
		WHERE id=$1
		RETURNING `+productColumns, id, in.Name, in.Category, in.UOM, in.BasePrice, isActive)
	return scanProduct(row)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Product, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+productColumns+` FROM products WHERE id=$1`, id)
	return scanProduct(row)
}

// FindByCode looks up a product by its unique code. Used by the quotation
// domain to find (or decide whether to create) the materialized product
// behind a confirmed design+fabric combination.
func (r *Repository) FindByCode(ctx context.Context, code string) (Product, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+productColumns+` FROM products WHERE code=$1`, code)
	return scanProduct(row)
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]Product, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+productColumns+` FROM products ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Product{}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

const sizeColumns = `id, product_id, size_code, size_multiplier, sort_order, is_active`

func scanSize(row pgx.Row) (ProductSize, error) {
	var s ProductSize
	err := row.Scan(&s.ID, &s.ProductID, &s.SizeCode, &s.SizeMultiplier, &s.SortOrder, &s.IsActive)
	return s, err
}

func (r *Repository) CreateSize(ctx context.Context, productID uuid.UUID, in UpsertSizeInput) (ProductSize, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO product_sizes (product_id, size_code, size_multiplier, sort_order)
		VALUES ($1,$2,$3,$4)
		RETURNING `+sizeColumns, productID, in.SizeCode, in.SizeMultiplier, in.SortOrder)
	return scanSize(row)
}

func (r *Repository) UpdateSize(ctx context.Context, id uuid.UUID, in UpsertSizeInput) (ProductSize, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE product_sizes SET size_code=$2, size_multiplier=$3, sort_order=$4, is_active=$5
		WHERE id=$1
		RETURNING `+sizeColumns, id, in.SizeCode, in.SizeMultiplier, in.SortOrder, isActive)
	return scanSize(row)
}

func (r *Repository) GetSizeByID(ctx context.Context, id uuid.UUID) (ProductSize, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+sizeColumns+` FROM product_sizes WHERE id=$1`, id)
	return scanSize(row)
}

// FindSizeByCode looks up a product size by its code within a product. Used
// by the quotation domain alongside FindByCode when materializing products.
func (r *Repository) FindSizeByCode(ctx context.Context, productID uuid.UUID, sizeCode string) (ProductSize, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+sizeColumns+` FROM product_sizes WHERE product_id=$1 AND size_code=$2`, productID, sizeCode)
	return scanSize(row)
}

// ListAllSizes loads every product's sizes in one query, so List (which
// returns many products at once) can attach each product's sizes without
// an N+1 round trip per product.
func (r *Repository) ListAllSizes(ctx context.Context) ([]ProductSize, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+sizeColumns+` FROM product_sizes ORDER BY product_id, sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ProductSize{}
	for rows.Next() {
		s, err := scanSize(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) ListSizesByProduct(ctx context.Context, productID uuid.UUID) ([]ProductSize, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+sizeColumns+` FROM product_sizes WHERE product_id=$1 ORDER BY sort_order`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ProductSize{}
	for rows.Next() {
		s, err := scanSize(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
