package accounting

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/db"
)

type PeriodStatus string

const (
	PeriodOpen   PeriodStatus = "OPEN"
	PeriodClosed PeriodStatus = "CLOSED"
)

type Period struct {
	ID        uuid.UUID    `json:"id"`
	Period    string       `json:"period"`
	StartDate time.Time    `json:"start_date"`
	EndDate   time.Time    `json:"end_date"`
	Status    PeriodStatus `json:"status"`
	ClosedAt  *time.Time   `json:"closed_at,omitempty"`
	ClosedBy  *string      `json:"closed_by,omitempty"`
}

type PeriodRepository struct {
	pool *pgxpool.Pool
}

func NewPeriodRepository(pool *pgxpool.Pool) *PeriodRepository {
	return &PeriodRepository{pool: pool}
}

const periodColumns = `id, period, start_date, end_date, status, closed_at, closed_by`

func scanPeriod(row pgx.Row) (Period, error) {
	var p Period
	err := row.Scan(&p.ID, &p.Period, &p.StartDate, &p.EndDate, &p.Status, &p.ClosedAt, &p.ClosedBy)
	return p, err
}

func (r *PeriodRepository) GetByPeriod(ctx context.Context, period string) (Period, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+periodColumns+` FROM accounting_periods WHERE period = $1`, period)
	return scanPeriod(row)
}

func (r *PeriodRepository) Create(ctx context.Context, period string, start, end time.Time) (Period, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO accounting_periods (period, start_date, end_date) VALUES ($1, $2, $3)
		RETURNING `+periodColumns, period, start, end)
	return scanPeriod(row)
}

func (r *PeriodRepository) List(ctx context.Context) ([]Period, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+periodColumns+` FROM accounting_periods ORDER BY period DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Period{}
	for rows.Next() {
		p, err := scanPeriod(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PeriodRepository) SetStatus(ctx context.Context, period string, status PeriodStatus, closedBy *string) (Period, error) {
	var row pgx.Row
	if status == PeriodClosed {
		row = db.Q(ctx, r.pool).QueryRow(ctx, `
			UPDATE accounting_periods SET status = $2, closed_at = now(), closed_by = $3
			WHERE period = $1 RETURNING `+periodColumns, period, status, closedBy)
	} else {
		row = db.Q(ctx, r.pool).QueryRow(ctx, `
			UPDATE accounting_periods SET status = $2, closed_at = NULL, closed_by = NULL
			WHERE period = $1 RETURNING `+periodColumns, period, status)
	}
	return scanPeriod(row)
}

// EnsurePeriodOpen enforces rule 6 (lock postings once a period is closed).
// A month with no accounting_periods row is treated as open by default, so
// operators are not forced to pre-create a period before they can post.
func EnsurePeriodOpen(ctx context.Context, repo *PeriodRepository, on time.Time) error {
	p, err := repo.GetByPeriod(ctx, on.Format("200601"))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return apperr.Internal("check accounting period", err)
	}
	if p.Status == PeriodClosed {
		return apperr.Conflict("accounting period " + p.Period + " is closed")
	}
	return nil
}
