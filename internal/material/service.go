package material

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

func (s *Service) Create(ctx context.Context, in UpsertInput) (Material, error) {
	if in.Code == "" || in.Name == "" {
		return Material{}, apperr.Validation("code and name are required")
	}
	var out Material
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		m, err := s.repo.Create(ctx, in)
		if err != nil {
			return apperr.Internal("create material", err)
		}
		if err := audit.Log(ctx, tx, "materials", m.ID, audit.Insert, nil, m); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = m
		return nil
	})
	return out, err
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in UpsertInput) (Material, error) {
	var out Material
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		before, err := s.repo.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("material not found")
			}
			return apperr.Internal("load material", err)
		}
		after, err := s.repo.Update(ctx, id, in)
		if err != nil {
			return apperr.Internal("update material", err)
		}
		if err := audit.Log(ctx, tx, "materials", id, audit.Update, before, after); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = after
		return nil
	})
	return out, err
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Material, error) {
	m, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Material{}, apperr.NotFound("material not found")
	}
	return m, err
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]Material, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.List(ctx, limit, offset)
}
