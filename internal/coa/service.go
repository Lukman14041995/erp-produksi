package coa

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

func (s *Service) Create(ctx context.Context, in CreateAccountInput) (Account, error) {
	if in.Code == "" || in.Name == "" {
		return Account{}, apperr.Validation("code and name are required")
	}
	if in.AccountType == "" || in.NormalBalance == "" {
		return Account{}, apperr.Validation("account_type and normal_balance are required")
	}

	var out Account
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		a, err := s.repo.Create(ctx, in)
		if err != nil {
			return apperr.Internal("create account", err)
		}
		if err := audit.Log(ctx, tx, "accounts", a.ID, audit.Insert, nil, a); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = a
		return nil
	})
	return out, err
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in UpdateAccountInput) (Account, error) {
	var out Account
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		before, err := s.repo.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("account not found")
			}
			return apperr.Internal("load account", err)
		}

		after, err := s.repo.Update(ctx, id, in)
		if err != nil {
			return apperr.Internal("update account", err)
		}
		if err := audit.Log(ctx, tx, "accounts", id, audit.Update, before, after); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = after
		return nil
	})
	return out, err
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Account, error) {
	a, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, apperr.NotFound("account not found")
	}
	return a, err
}

func (s *Service) GetByCode(ctx context.Context, code string) (Account, error) {
	a, err := s.repo.GetByCode(ctx, code)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, apperr.NotFound("account with code " + code + " not found")
	}
	return a, err
}

func (s *Service) List(ctx context.Context) ([]Account, error) {
	return s.repo.List(ctx)
}
