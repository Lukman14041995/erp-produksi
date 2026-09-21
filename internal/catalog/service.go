package catalog

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

// ---- Product types ----

func (s *Service) CreateProductType(ctx context.Context, in UpsertProductTypeInput) (ProductType, error) {
	if in.Code == "" || in.Name == "" {
		return ProductType{}, apperr.Validation("code and name are required")
	}
	var out ProductType
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		p, err := s.repo.CreateProductType(ctx, in)
		if err != nil {
			return apperr.Internal("create product type", err)
		}
		if err := audit.Log(ctx, tx, "product_types", p.ID, audit.Insert, nil, p); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = p
		return nil
	})
	return out, err
}

func (s *Service) UpdateProductType(ctx context.Context, id uuid.UUID, in UpsertProductTypeInput) (ProductType, error) {
	var out ProductType
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		before, err := s.repo.GetProductTypeByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("product type not found")
			}
			return apperr.Internal("load product type", err)
		}
		after, err := s.repo.UpdateProductType(ctx, id, in)
		if err != nil {
			return apperr.Internal("update product type", err)
		}
		if err := audit.Log(ctx, tx, "product_types", id, audit.Update, before, after); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = after
		return nil
	})
	return out, err
}

func (s *Service) GetProductType(ctx context.Context, id uuid.UUID) (ProductType, error) {
	p, err := s.repo.GetProductTypeByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProductType{}, apperr.NotFound("product type not found")
	}
	return p, err
}

func (s *Service) ListProductTypes(ctx context.Context, activeOnly bool) ([]ProductType, error) {
	return s.repo.ListProductTypes(ctx, activeOnly)
}

// ---- Fabrics ----

func (s *Service) CreateFabric(ctx context.Context, in UpsertFabricInput) (Fabric, error) {
	if in.Code == "" || in.Name == "" {
		return Fabric{}, apperr.Validation("code and name are required")
	}
	if in.ProductTypeID == uuid.Nil {
		return Fabric{}, apperr.Validation("product_type_id is required")
	}
	if in.MaterialID == uuid.Nil {
		return Fabric{}, apperr.Validation("material_id is required")
	}
	if in.SalesPrice.IsNegative() {
		return Fabric{}, apperr.Validation("sales_price cannot be negative")
	}
	var out Fabric
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		f, err := s.repo.CreateFabric(ctx, in)
		if err != nil {
			return apperr.Internal("create fabric", err)
		}
		if err := audit.Log(ctx, tx, "fabrics", f.ID, audit.Insert, nil, f); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = f
		return nil
	})
	return out, err
}

func (s *Service) UpdateFabric(ctx context.Context, id uuid.UUID, in UpsertFabricInput) (Fabric, error) {
	var out Fabric
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		before, err := s.repo.GetFabricByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("fabric not found")
			}
			return apperr.Internal("load fabric", err)
		}
		after, err := s.repo.UpdateFabric(ctx, id, in)
		if err != nil {
			return apperr.Internal("update fabric", err)
		}
		if err := audit.Log(ctx, tx, "fabrics", id, audit.Update, before, after); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = after
		return nil
	})
	return out, err
}

func (s *Service) SetFabricImage(ctx context.Context, id uuid.UUID, imageURL string) (Fabric, error) {
	var out Fabric
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := s.repo.GetFabricByID(ctx, id); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("fabric not found")
			}
			return apperr.Internal("load fabric", err)
		}
		f, err := s.repo.SetFabricImage(ctx, id, imageURL)
		if err != nil {
			return apperr.Internal("set fabric image", err)
		}
		if err := audit.Log(ctx, tx, "fabrics", id, audit.Update, nil, f); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = f
		return nil
	})
	return out, err
}

func (s *Service) GetFabric(ctx context.Context, id uuid.UUID) (Fabric, error) {
	f, err := s.repo.GetFabricByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Fabric{}, apperr.NotFound("fabric not found")
	}
	return f, err
}

func (s *Service) ListFabrics(ctx context.Context, activeOnly bool) ([]Fabric, error) {
	return s.repo.ListFabrics(ctx, activeOnly)
}

// ---- Garment variants (Model potongan) ----

func (s *Service) CreateVariant(ctx context.Context, in UpsertVariantInput) (GarmentVariant, error) {
	if in.Code == "" || in.Name == "" {
		return GarmentVariant{}, apperr.Validation("code and name are required")
	}
	if in.ProductTypeID == uuid.Nil {
		return GarmentVariant{}, apperr.Validation("product_type_id is required")
	}
	if in.PriceAddon.IsNegative() {
		return GarmentVariant{}, apperr.Validation("price_addon cannot be negative")
	}
	var out GarmentVariant
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		v, err := s.repo.CreateVariant(ctx, in)
		if err != nil {
			return apperr.Internal("create garment variant", err)
		}
		if err := audit.Log(ctx, tx, "garment_variants", v.ID, audit.Insert, nil, v); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = v
		return nil
	})
	return out, err
}

func (s *Service) UpdateVariant(ctx context.Context, id uuid.UUID, in UpsertVariantInput) (GarmentVariant, error) {
	var out GarmentVariant
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		before, err := s.repo.GetVariantByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("garment variant not found")
			}
			return apperr.Internal("load garment variant", err)
		}
		after, err := s.repo.UpdateVariant(ctx, id, in)
		if err != nil {
			return apperr.Internal("update garment variant", err)
		}
		if err := audit.Log(ctx, tx, "garment_variants", id, audit.Update, before, after); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = after
		return nil
	})
	return out, err
}

func (s *Service) SetVariantIcon(ctx context.Context, id uuid.UUID, iconURL string) (GarmentVariant, error) {
	var out GarmentVariant
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := s.repo.GetVariantByID(ctx, id); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("garment variant not found")
			}
			return apperr.Internal("load garment variant", err)
		}
		v, err := s.repo.SetVariantIcon(ctx, id, iconURL)
		if err != nil {
			return apperr.Internal("set variant icon", err)
		}
		if err := audit.Log(ctx, tx, "garment_variants", id, audit.Update, nil, v); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = v
		return nil
	})
	return out, err
}

func (s *Service) GetVariant(ctx context.Context, id uuid.UUID) (GarmentVariant, error) {
	v, err := s.repo.GetVariantByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return GarmentVariant{}, apperr.NotFound("garment variant not found")
	}
	return v, err
}

func (s *Service) ListVariants(ctx context.Context, activeOnly bool) ([]GarmentVariant, error) {
	return s.repo.ListVariants(ctx, activeOnly)
}

// ---- Garment sizes ----

func (s *Service) CreateGarmentSize(ctx context.Context, in UpsertSizeInput) (GarmentSize, error) {
	if in.SizeCode == "" {
		return GarmentSize{}, apperr.Validation("size_code is required")
	}
	if in.SizeMultiplier.IsZero() || in.SizeMultiplier.IsNegative() {
		return GarmentSize{}, apperr.Validation("size_multiplier must be greater than zero")
	}
	var out GarmentSize
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		sz, err := s.repo.CreateGarmentSize(ctx, in)
		if err != nil {
			return apperr.Internal("create garment size", err)
		}
		if err := audit.Log(ctx, tx, "garment_sizes", sz.ID, audit.Insert, nil, sz); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = sz
		return nil
	})
	return out, err
}

func (s *Service) UpdateGarmentSize(ctx context.Context, id uuid.UUID, in UpsertSizeInput) (GarmentSize, error) {
	var out GarmentSize
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		before, err := s.repo.GetGarmentSizeByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("garment size not found")
			}
			return apperr.Internal("load garment size", err)
		}
		after, err := s.repo.UpdateGarmentSize(ctx, id, in)
		if err != nil {
			return apperr.Internal("update garment size", err)
		}
		if err := audit.Log(ctx, tx, "garment_sizes", id, audit.Update, before, after); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = after
		return nil
	})
	return out, err
}

func (s *Service) GetGarmentSize(ctx context.Context, id uuid.UUID) (GarmentSize, error) {
	sz, err := s.repo.GetGarmentSizeByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return GarmentSize{}, apperr.NotFound("garment size not found")
	}
	return sz, err
}

func (s *Service) ListGarmentSizes(ctx context.Context, activeOnly bool) ([]GarmentSize, error) {
	return s.repo.ListGarmentSizes(ctx, activeOnly)
}

// ---- Inks (Tinta) ----

func (s *Service) CreateInk(ctx context.Context, in UpsertInkInput) (Ink, error) {
	if in.Code == "" || in.Name == "" {
		return Ink{}, apperr.Validation("code and name are required")
	}
	if in.ProductTypeID == uuid.Nil {
		return Ink{}, apperr.Validation("product_type_id is required")
	}
	if in.MaterialID == uuid.Nil {
		return Ink{}, apperr.Validation("material_id is required")
	}
	if in.PriceAddon.IsNegative() {
		return Ink{}, apperr.Validation("price_addon cannot be negative")
	}
	var out Ink
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		i, err := s.repo.CreateInk(ctx, in)
		if err != nil {
			return apperr.Internal("create ink", err)
		}
		if err := audit.Log(ctx, tx, "inks", i.ID, audit.Insert, nil, i); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = i
		return nil
	})
	return out, err
}

func (s *Service) UpdateInk(ctx context.Context, id uuid.UUID, in UpsertInkInput) (Ink, error) {
	var out Ink
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		before, err := s.repo.GetInkByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("ink not found")
			}
			return apperr.Internal("load ink", err)
		}
		after, err := s.repo.UpdateInk(ctx, id, in)
		if err != nil {
			return apperr.Internal("update ink", err)
		}
		if err := audit.Log(ctx, tx, "inks", id, audit.Update, before, after); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = after
		return nil
	})
	return out, err
}

func (s *Service) SetInkImage(ctx context.Context, id uuid.UUID, imageURL string) (Ink, error) {
	var out Ink
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := s.repo.GetInkByID(ctx, id); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("ink not found")
			}
			return apperr.Internal("load ink", err)
		}
		i, err := s.repo.SetInkImage(ctx, id, imageURL)
		if err != nil {
			return apperr.Internal("set ink image", err)
		}
		if err := audit.Log(ctx, tx, "inks", id, audit.Update, nil, i); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = i
		return nil
	})
	return out, err
}

func (s *Service) GetInk(ctx context.Context, id uuid.UUID) (Ink, error) {
	i, err := s.repo.GetInkByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Ink{}, apperr.NotFound("ink not found")
	}
	return i, err
}

func (s *Service) ListInks(ctx context.Context, activeOnly bool) ([]Ink, error) {
	return s.repo.ListInks(ctx, activeOnly)
}
