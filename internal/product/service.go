package product

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

func (s *Service) Create(ctx context.Context, in UpsertProductInput) (Product, error) {
	if in.Code == "" || in.Name == "" {
		return Product{}, apperr.Validation("code and name are required")
	}
	var out Product
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		p, err := s.repo.Create(ctx, in)
		if err != nil {
			return apperr.Internal("create product", err)
		}
		if err := audit.Log(ctx, tx, "products", p.ID, audit.Insert, nil, p); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = p
		return nil
	})
	return out, err
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in UpsertProductInput) (Product, error) {
	var out Product
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		before, err := s.repo.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("product not found")
			}
			return apperr.Internal("load product", err)
		}
		after, err := s.repo.Update(ctx, id, in)
		if err != nil {
			return apperr.Internal("update product", err)
		}
		if err := audit.Log(ctx, tx, "products", id, audit.Update, before, after); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = after
		return nil
	})
	return out, err
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, apperr.NotFound("product not found")
	}
	if err != nil {
		return Product{}, err
	}
	sizes, err := s.repo.ListSizesByProduct(ctx, id)
	if err != nil {
		return Product{}, apperr.Internal("load product sizes", err)
	}
	p.Sizes = sizes
	return p, nil
}

// FindByCode returns apperr.NotFound if no product has this code, so callers
// (e.g. the quotation domain's find-or-create materialization) can branch on
// errors.As instead of a raw pgx.ErrNoRows check.
func (s *Service) FindByCode(ctx context.Context, code string) (Product, error) {
	p, err := s.repo.FindByCode(ctx, code)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, apperr.NotFound("product not found")
	}
	return p, err
}

func (s *Service) FindSizeByCode(ctx context.Context, productID uuid.UUID, sizeCode string) (ProductSize, error) {
	sz, err := s.repo.FindSizeByCode(ctx, productID, sizeCode)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProductSize{}, apperr.NotFound("product size not found")
	}
	return sz, err
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]Product, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	products, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	allSizes, err := s.repo.ListAllSizes(ctx)
	if err != nil {
		return nil, apperr.Internal("load product sizes", err)
	}
	sizesByProduct := map[uuid.UUID][]ProductSize{}
	for _, sz := range allSizes {
		sizesByProduct[sz.ProductID] = append(sizesByProduct[sz.ProductID], sz)
	}
	for i := range products {
		products[i].Sizes = sizesByProduct[products[i].ID]
	}
	return products, nil
}

func (s *Service) AddSize(ctx context.Context, productID uuid.UUID, in UpsertSizeInput) (ProductSize, error) {
	if in.SizeCode == "" {
		return ProductSize{}, apperr.Validation("size_code is required")
	}
	if in.SizeMultiplier.IsZero() {
		return ProductSize{}, apperr.Validation("size_multiplier must be greater than zero")
	}
	var out ProductSize
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := s.repo.GetByID(ctx, productID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("product not found")
			}
			return apperr.Internal("load product", err)
		}
		size, err := s.repo.CreateSize(ctx, productID, in)
		if err != nil {
			return apperr.Internal("create product size", err)
		}
		if err := audit.Log(ctx, tx, "product_sizes", size.ID, audit.Insert, nil, size); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = size
		return nil
	})
	return out, err
}

func (s *Service) UpdateSize(ctx context.Context, sizeID uuid.UUID, in UpsertSizeInput) (ProductSize, error) {
	var out ProductSize
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		before, err := s.repo.GetSizeByID(ctx, sizeID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("product size not found")
			}
			return apperr.Internal("load product size", err)
		}
		after, err := s.repo.UpdateSize(ctx, sizeID, in)
		if err != nil {
			return apperr.Internal("update product size", err)
		}
		if err := audit.Log(ctx, tx, "product_sizes", sizeID, audit.Update, before, after); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = after
		return nil
	})
	return out, err
}

func (s *Service) ListSizes(ctx context.Context, productID uuid.UUID) ([]ProductSize, error) {
	return s.repo.ListSizesByProduct(ctx, productID)
}
