package supplier

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/audit"
	"github.com/ranji/clothing-erp/internal/db"
)

type Service struct {
	pool *pgxpool.Pool
	repo *Repository
}

func NewService(pool *pgxpool.Pool, repo *Repository) *Service {
	return &Service{pool: pool, repo: repo}
}

func (s *Service) Create(ctx context.Context, in UpsertInput) (Supplier, error) {
	if in.Code == "" || in.Name == "" {
		return Supplier{}, apperr.Validation("code and name are required")
	}
	var out Supplier
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		sup, err := s.repo.Create(ctx, in)
		if err != nil {
			return apperr.Internal("create supplier", err)
		}
		if err := audit.Log(ctx, tx, "suppliers", sup.ID, audit.Insert, nil, sup); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = sup
		return nil
	})
	return out, err
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in UpsertInput) (Supplier, error) {
	var out Supplier
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		before, err := s.repo.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("supplier not found")
			}
			return apperr.Internal("load supplier", err)
		}
		after, err := s.repo.Update(ctx, id, in)
		if err != nil {
			return apperr.Internal("update supplier", err)
		}
		if err := audit.Log(ctx, tx, "suppliers", id, audit.Update, before, after); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = after
		return nil
	})
	return out, err
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Supplier, error) {
	sup, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Supplier{}, apperr.NotFound("supplier not found")
	}
	return sup, err
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]Supplier, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.List(ctx, limit, offset)
}
