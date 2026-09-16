package accounting

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/audit"
	"github.com/ranji/clothing-erp/internal/coa"
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/ranji/clothing-erp/internal/numbering"
	"github.com/shopspring/decimal"
)

// Service is the single gateway through which every other domain posts
// accounting entries. No other package writes to journal_entries /
// journal_entry_lines directly, which is what makes accounting
// transaction-driven rather than a manually re-keyed module.
type Service struct {
	pool        *pgxpool.Pool
	repo        *Repository
	accountRepo *coa.Repository
	periodRepo  *PeriodRepository
}

func NewService(pool *pgxpool.Pool, repo *Repository, accountRepo *coa.Repository, periodRepo *PeriodRepository) *Service {
	return &Service{pool: pool, repo: repo, accountRepo: accountRepo, periodRepo: periodRepo}
}

// PostJournal validates that debits equal credits, resolves each line's
// account code, and writes an immutable journal. It participates in the
// caller's ambient transaction (via db.WithTx's reuse behavior) so that, for
// example, an invoice row and its journal commit or roll back together.
func (s *Service) PostJournal(ctx context.Context, sourceType SourceType, sourceID *uuid.UUID, journalDate time.Time, description string, lines []JournalLineInput) (JournalEntry, error) {
	if len(lines) < 2 {
		return JournalEntry{}, apperr.Validation("a journal requires at least two lines")
	}

	totalDebit := decimal.Zero
	totalCredit := decimal.Zero
	for _, l := range lines {
		if l.Debit.IsNegative() || l.Credit.IsNegative() {
			return JournalEntry{}, apperr.Validation("journal line amounts cannot be negative")
		}
		if l.Debit.IsPositive() && l.Credit.IsPositive() {
			return JournalEntry{}, apperr.Validation("a journal line cannot carry both a debit and a credit")
		}
		if l.Debit.IsZero() && l.Credit.IsZero() {
			return JournalEntry{}, apperr.Validation("a journal line must carry either a debit or a credit")
		}
		totalDebit = totalDebit.Add(l.Debit)
		totalCredit = totalCredit.Add(l.Credit)
	}

	if !totalDebit.Equal(totalCredit) {
		return JournalEntry{}, apperr.Validation("journal is not balanced: total debit must equal total credit")
	}

	var result JournalEntry
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		if err := EnsurePeriodOpen(ctx, s.periodRepo, journalDate); err != nil {
			return err
		}

		number, err := numbering.Generate(ctx, tx, numbering.Journal, journalDate)
		if err != nil {
			return apperr.Internal("generate journal number", err)
		}

		header := JournalEntry{
			JournalNumber: number,
			JournalDate:   journalDate,
			SourceType:    sourceType,
			SourceID:      sourceID,
			Description:   description,
			Status:        StatusPosted,
			TotalDebit:    totalDebit,
			TotalCredit:   totalCredit,
			CreatedBy:     db.ActorFromContext(ctx),
		}

		id, err := s.repo.InsertHeader(ctx, header)
		if err != nil {
			return apperr.Internal("insert journal header", err)
		}
		header.ID = id

		for i, l := range lines {
			acc, err := s.accountRepo.GetByCode(ctx, l.AccountCode)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return apperr.Validation("unknown account code: " + l.AccountCode)
				}
				return apperr.Internal("resolve account code", err)
			}
			if !acc.IsPostable {
				return apperr.Validation("account is not postable: " + l.AccountCode)
			}

			if err := s.repo.InsertLine(ctx, id, i+1, acc.ID, l.Debit, l.Credit, l.Description); err != nil {
				return apperr.Internal("insert journal line", err)
			}
			header.Lines = append(header.Lines, JournalEntryLine{
				LineNo: i + 1, AccountID: acc.ID, AccountCode: acc.Code, AccountName: acc.Name,
				Debit: l.Debit, Credit: l.Credit, Description: l.Description,
			})
		}

		if err := audit.Log(ctx, tx, "journal_entries", id, audit.Insert, nil, header); err != nil {
			return apperr.Internal("write audit log", err)
		}

		result = header
		return nil
	})

	return result, err
}

// ReverseJournal posts an equal-and-opposite journal and links it back to
// the original, which is then flagged REVERSED. The original's number,
// date, and lines are never altered or deleted -- this is the only
// sanctioned way to undo a posted entry.
func (s *Service) ReverseJournal(ctx context.Context, journalID uuid.UUID, reason string) (JournalEntry, error) {
	var result JournalEntry
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		original, err := s.repo.GetHeader(ctx, journalID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("journal not found")
			}
			return apperr.Internal("load journal", err)
		}
		if original.Status == StatusReversed {
			return apperr.Conflict("journal already reversed")
		}

		lines, err := s.repo.GetLines(ctx, journalID)
		if err != nil {
			return apperr.Internal("load journal lines", err)
		}

		reversedLines := make([]JournalLineInput, len(lines))
		for i, l := range lines {
			reversedLines[i] = JournalLineInput{
				AccountCode: l.AccountCode,
				Debit:       l.Credit,
				Credit:      l.Debit,
				Description: "Reversal: " + l.Description,
			}
		}

		reversal, err := s.PostJournal(ctx, SourceReversal, &journalID, time.Now(), "Reversal of "+original.JournalNumber+": "+reason, reversedLines)
		if err != nil {
			return err
		}

		if err := s.repo.MarkReversed(ctx, journalID, reversal.ID); err != nil {
			return apperr.Internal("mark journal reversed", err)
		}
		if err := audit.Log(ctx, tx, "journal_entries", journalID, audit.Update, original, map[string]any{"status": StatusReversed, "reversed_journal_id": reversal.ID}); err != nil {
			return apperr.Internal("write audit log", err)
		}

		result = reversal
		return nil
	})
	return result, err
}

func (s *Service) GetJournal(ctx context.Context, id uuid.UUID) (JournalEntry, error) {
	header, err := s.repo.GetHeader(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return JournalEntry{}, apperr.NotFound("journal not found")
		}
		return JournalEntry{}, err
	}
	lines, err := s.repo.GetLines(ctx, id)
	if err != nil {
		return JournalEntry{}, err
	}
	header.Lines = lines
	return header, nil
}

func (s *Service) ListJournals(ctx context.Context, limit, offset int) ([]JournalEntry, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.List(ctx, limit, offset)
}

// ListForAuditTrail returns every journal (with lines) tied to any of the
// given source ids, so a caller can pass a document's own id plus its
// downstream documents' ids (invoice, payments, production order) to
// reconstruct the full posting history for that document.
func (s *Service) ListForAuditTrail(ctx context.Context, sourceIDs []uuid.UUID) ([]JournalEntry, error) {
	headers, err := s.repo.ListBySourceIDs(ctx, sourceIDs)
	if err != nil {
		return nil, apperr.Internal("list journals by source", err)
	}
	for i := range headers {
		lines, err := s.repo.GetLines(ctx, headers[i].ID)
		if err != nil {
			return nil, apperr.Internal("load journal lines", err)
		}
		headers[i].Lines = lines
	}
	return headers, nil
}
