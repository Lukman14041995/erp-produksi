package supplier

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

const columns = `id, code, name, contact_person, phone, email, address, tax_id, is_active, created_at, updated_at`

func scan(row pgx.Row) (Supplier, error) {
	var s Supplier
	err := row.Scan(&s.ID, &s.Code, &s.Name, &s.ContactPerson, &s.Phone, &s.Email, &s.Address, &s.TaxID, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *Repository) Create(ctx context.Context, in UpsertInput) (Supplier, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO suppliers (code, name, contact_person, phone, email, address, tax_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING `+columns, in.Code, in.Name, in.ContactPerson, in.Phone, in.Email, in.Address, in.TaxID)
	return scan(row)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, in UpsertInput) (Supplier, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE suppliers SET name=$2, contact_person=$3, phone=$4, email=$5, address=$6, tax_id=$7, is_active=$8
		WHERE id=$1
		RETURNING `+columns, id, in.Name, in.ContactPerson, in.Phone, in.Email, in.Address, in.TaxID, isActive)
	return scan(row)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Supplier, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+columns+` FROM suppliers WHERE id=$1`, id)
	return scan(row)
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]Supplier, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+columns+` FROM suppliers ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Supplier{}
	for rows.Next() {
		s, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
