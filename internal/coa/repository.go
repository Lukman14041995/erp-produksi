package coa

import (
	"context"
	"errors"
	"fmt"

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

const selectColumns = `id, code, name, account_type, normal_balance, parent_id, is_postable, is_active, created_at, updated_at`

func scanAccount(row pgx.Row) (Account, error) {
	var a Account
	err := row.Scan(&a.ID, &a.Code, &a.Name, &a.AccountType, &a.NormalBalance, &a.ParentID, &a.IsPostable, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

func (r *Repository) Create(ctx context.Context, in CreateAccountInput) (Account, error) {
	isPostable := true
	if in.IsPostable != nil {
		isPostable = *in.IsPostable
	}
	row := db.Q(ctx, r.pool).QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO accounts (code, name, account_type, normal_balance, parent_id, is_postable)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING %s
	`, selectColumns), in.Code, in.Name, in.AccountType, in.NormalBalance, in.ParentID, isPostable)
	return scanAccount(row)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, in UpdateAccountInput) (Account, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, fmt.Sprintf(`
		UPDATE accounts SET name = $2, is_postable = $3, is_active = $4
		WHERE id = $1
		RETURNING %s
	`, selectColumns), id, in.Name, in.IsPostable, in.IsActive)
	return scanAccount(row)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Account, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, fmt.Sprintf(`SELECT %s FROM accounts WHERE id = $1`, selectColumns), id)
	return scanAccount(row)
}

func (r *Repository) GetByCode(ctx context.Context, code string) (Account, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, fmt.Sprintf(`SELECT %s FROM accounts WHERE code = $1`, selectColumns), code)
	return scanAccount(row)
}

var ErrNotFound = errors.New("account not found")

func (r *Repository) List(ctx context.Context) ([]Account, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, fmt.Sprintf(`SELECT %s FROM accounts ORDER BY code`, selectColumns))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Account{}
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
