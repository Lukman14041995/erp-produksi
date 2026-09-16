package material

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

const columns = `id, code, name, uom, unit_cost, is_active, created_at, updated_at`

func scan(row pgx.Row) (Material, error) {
	var m Material
	err := row.Scan(&m.ID, &m.Code, &m.Name, &m.UOM, &m.UnitCost, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

func (r *Repository) Create(ctx context.Context, in UpsertInput) (Material, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO materials (code, name, uom, unit_cost)
		VALUES ($1,$2,$3,$4)
		RETURNING `+columns, in.Code, in.Name, in.UOM, in.UnitCost)
	return scan(row)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, in UpsertInput) (Material, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE materials SET name=$2, uom=$3, unit_cost=$4, is_active=$5
		WHERE id=$1
		RETURNING `+columns, id, in.Name, in.UOM, in.UnitCost, isActive)
	return scan(row)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Material, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+columns+` FROM materials WHERE id=$1`, id)
	return scan(row)
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]Material, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+columns+` FROM materials ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Material{}
	for rows.Next() {
		m, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
