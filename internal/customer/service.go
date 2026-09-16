package customer

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

func (s *Service) Create(ctx context.Context, in UpsertInput) (Customer, error) {
	if in.Code == "" || in.Name == "" {
		return Customer{}, apperr.Validation("code and name are required")
	}
	var out Customer
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		c, err := s.repo.Create(ctx, in)
		if err != nil {
			return apperr.Internal("create customer", err)
		}
		if err := audit.Log(ctx, tx, "customers", c.ID, audit.Insert, nil, c); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = c
		return nil
	})
	return out, err
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in UpsertInput) (Customer, error) {
	var out Customer
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		before, err := s.repo.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("customer not found")
			}
			return apperr.Internal("load customer", err)
		}
		after, err := s.repo.Update(ctx, id, in)
		if err != nil {
			return apperr.Internal("update customer", err)
		}
		if err := audit.Log(ctx, tx, "customers", id, audit.Update, before, after); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = after
		return nil
	})
	return out, err
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Customer, error) {
	c, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Customer{}, apperr.NotFound("customer not found")
	}
	return c, err
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]Customer, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.List(ctx, limit, offset)
}
