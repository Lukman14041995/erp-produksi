package customer

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

const columns = `id, code, name, contact_person, phone, email, address, tax_id, credit_limit, is_active, created_at, updated_at`

func scan(row pgx.Row) (Customer, error) {
	var c Customer
	err := row.Scan(&c.ID, &c.Code, &c.Name, &c.ContactPerson, &c.Phone, &c.Email, &c.Address, &c.TaxID, &c.CreditLimit, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *Repository) Create(ctx context.Context, in UpsertInput) (Customer, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO customers (code, name, contact_person, phone, email, address, tax_id, credit_limit)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING `+columns, in.Code, in.Name, in.ContactPerson, in.Phone, in.Email, in.Address, in.TaxID, in.CreditLimit)
	return scan(row)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, in UpsertInput) (Customer, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE customers SET name=$2, contact_person=$3, phone=$4, email=$5, address=$6, tax_id=$7, credit_limit=$8, is_active=$9
		WHERE id=$1
		RETURNING `+columns, id, in.Name, in.ContactPerson, in.Phone, in.Email, in.Address, in.TaxID, in.CreditLimit, isActive)
	return scan(row)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Customer, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+columns+` FROM customers WHERE id=$1`, id)
	return scan(row)
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]Customer, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+columns+` FROM customers ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Customer{}
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
