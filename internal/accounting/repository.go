package accounting

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

func (r *Repository) InsertHeader(ctx context.Context, e JournalEntry) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO journal_entries
			(journal_number, journal_date, source_type, source_id, description, status, total_debit, total_credit, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, e.JournalNumber, e.JournalDate, e.SourceType, e.SourceID, e.Description, e.Status, e.TotalDebit, e.TotalCredit, e.CreatedBy).Scan(&id)
	return id, err
}

func (r *Repository) InsertLine(ctx context.Context, journalID uuid.UUID, lineNo int, accountID uuid.UUID, debit, credit decimal.Decimal, description string) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		INSERT INTO journal_entry_lines (journal_id, line_no, account_id, debit, credit, description)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, journalID, lineNo, accountID, debit, credit, description)
	return err
}

func (r *Repository) MarkReversed(ctx context.Context, journalID, reversalJournalID uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		UPDATE journal_entries SET status = 'REVERSED', reversed_journal_id = $2
		WHERE id = $1
	`, journalID, reversalJournalID)
	return err
}

const headerColumns = `id, journal_number, journal_date, source_type, source_id, description, status, reversed_journal_id, total_debit, total_credit, created_by, created_at`

func scanHeader(row pgx.Row) (JournalEntry, error) {
	var e JournalEntry
	err := row.Scan(&e.ID, &e.JournalNumber, &e.JournalDate, &e.SourceType, &e.SourceID, &e.Description, &e.Status, &e.ReversedJournalID, &e.TotalDebit, &e.TotalCredit, &e.CreatedBy, &e.CreatedAt)
	return e, err
}

func (r *Repository) GetHeader(ctx context.Context, id uuid.UUID) (JournalEntry, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+headerColumns+` FROM journal_entries WHERE id = $1`, id)
	return scanHeader(row)
}

func (r *Repository) GetLines(ctx context.Context, journalID uuid.UUID) ([]JournalEntryLine, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT l.id, l.line_no, l.account_id, a.code, a.name, l.debit, l.credit, l.description
		FROM journal_entry_lines l
		JOIN accounts a ON a.id = l.account_id
		WHERE l.journal_id = $1
		ORDER BY l.line_no
	`, journalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []JournalEntryLine{}
	for rows.Next() {
		var l JournalEntryLine
		if err := rows.Scan(&l.ID, &l.LineNo, &l.AccountID, &l.AccountCode, &l.AccountName, &l.Debit, &l.Credit, &l.Description); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]JournalEntry, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT `+headerColumns+` FROM journal_entries
		ORDER BY journal_date DESC, created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []JournalEntry{}
	for rows.Next() {
		e, err := scanHeader(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListBySourceIDs finds every journal (across whatever source_type posted
// it) tied to any of the given source ids -- used to build an audit trail
// for one business document by also following its downstream effects (e.g.
// a sales order's own COGS journal plus its invoice's and payments').
func (r *Repository) ListBySourceIDs(ctx context.Context, sourceIDs []uuid.UUID) ([]JournalEntry, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT `+headerColumns+` FROM journal_entries
		WHERE source_id = ANY($1)
		ORDER BY journal_date, created_at
	`, sourceIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []JournalEntry{}
	for rows.Next() {
		e, err := scanHeader(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
