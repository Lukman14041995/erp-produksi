package auth

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

const userColumns = `id, email, password_hash, name, role, is_active, created_at, updated_at`

func scanUser(row pgx.Row) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, name string, role Role) (User, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO users (email, password_hash, name, role) VALUES ($1,$2,$3,$4)
		RETURNING `+userColumns, email, passwordHash, name, role)
	return scanUser(row)
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (User, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email=$1`, email)
	return scanUser(row)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id=$1`, id)
	return scanUser(row)
}

func (r *Repository) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+userColumns+` FROM users ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *Repository) CountUsers(ctx context.Context) (int, error) {
	var count int
	err := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func (r *Repository) InsertRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1,$2,$3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (r *Repository) GetActiveRefreshToken(ctx context.Context, tokenHash string) (RefreshToken, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at FROM refresh_tokens
		WHERE token_hash=$1 AND revoked_at IS NULL AND expires_at > now()
	`, tokenHash)
	var t RefreshToken
	err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt)
	return t, err
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE refresh_tokens SET revoked_at=now() WHERE id=$1`, id)
	return err
}
