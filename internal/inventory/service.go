package inventory

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/audit"
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/shopspring/decimal"
)

type Service struct {
	pool *pgxpool.Pool
	repo *Repository
}

func NewService(pool *pgxpool.Pool, repo *Repository) *Service {
	return &Service{pool: pool, repo: repo}
}

// RecordMovement is the single entry point for every stock change in the
// system (rule: double-entry movement tracking). It maintains a moving
// weighted-average cost per item *per warehouse*: receipts blend their unit
// cost into that warehouse's average, issues consume stock at the
// warehouse's current average (that consumed cost is what
// production/COGS/transfer journals should post with).
func (s *Service) RecordMovement(ctx context.Context, in MovementInput) (MovementResult, error) {
	if in.ItemType == ItemMaterial && in.MaterialID == nil {
		return MovementResult{}, apperr.Validation("material_id is required for MATERIAL movements")
	}
	if in.ItemType == ItemProduct && in.ProductID == nil {
		return MovementResult{}, apperr.Validation("product_id is required for PRODUCT movements")
	}
	if in.WarehouseID == uuid.Nil {
		return MovementResult{}, apperr.Validation("warehouse_id is required")
	}
	if in.QtyIn.IsNegative() || in.QtyOut.IsNegative() {
		return MovementResult{}, apperr.Validation("quantities cannot be negative")
	}
	if in.QtyIn.IsZero() && in.QtyOut.IsZero() {
		return MovementResult{}, apperr.Validation("movement must have a non-zero qty_in or qty_out")
	}

	var result MovementResult
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		balance, err := s.repo.GetBalanceForUpdate(ctx, in.ItemType, in.WarehouseID, in.MaterialID, in.ProductID, in.ProductSizeID)
		if err != nil {
			return apperr.Internal("lock inventory balance", err)
		}

		effectiveUnitCost := in.UnitCost
		totalCost := decimal.Zero

		if in.QtyIn.IsPositive() {
			totalCost = in.QtyIn.Mul(in.UnitCost)
			newQty := balance.QtyOnHand.Add(in.QtyIn)
			if newQty.IsPositive() {
				existingValue := balance.QtyOnHand.Mul(balance.AvgUnitCost)
				balance.AvgUnitCost = existingValue.Add(totalCost).Div(newQty)
			}
			balance.QtyOnHand = newQty
		} else {
			effectiveUnitCost = balance.AvgUnitCost
			totalCost = in.QtyOut.Mul(effectiveUnitCost)
			newQty := balance.QtyOnHand.Sub(in.QtyOut)
			if newQty.IsNegative() {
				return apperr.Conflict("insufficient stock on hand in this warehouse for this movement")
			}
			balance.QtyOnHand = newQty
		}

		if err := s.repo.SaveBalance(ctx, balance); err != nil {
			return apperr.Internal("save inventory balance", err)
		}

		txn, err := s.repo.InsertTransaction(ctx, Transaction{
			TxnType: in.TxnType, ItemType: in.ItemType, WarehouseID: in.WarehouseID,
			MaterialID: in.MaterialID, ProductID: in.ProductID, ProductSizeID: in.ProductSizeID,
			QtyIn: in.QtyIn, QtyOut: in.QtyOut, UnitCost: effectiveUnitCost, TotalCost: totalCost,
			RefType: in.RefType, RefID: in.RefID, TxnDate: in.TxnDate, Notes: in.Notes,
		})
		if err != nil {
			return apperr.Internal("insert inventory transaction", err)
		}

		result = MovementResult{Transaction: txn, Balance: balance}
		return nil
	})

	return result, err
}

func (s *Service) GetMaterialBalance(ctx context.Context, warehouseID, materialID uuid.UUID) (Balance, error) {
	return s.repo.GetBalance(ctx, ItemMaterial, warehouseID, &materialID, nil, nil)
}

func (s *Service) GetProductBalance(ctx context.Context, warehouseID, productID uuid.UUID, productSizeID *uuid.UUID) (Balance, error) {
	return s.repo.GetBalance(ctx, ItemProduct, warehouseID, nil, &productID, productSizeID)
}

// ---- Warehouses ----

func (s *Service) CreateWarehouse(ctx context.Context, in UpsertWarehouseInput) (Warehouse, error) {
	if in.Code == "" || in.Name == "" {
		return Warehouse{}, apperr.Validation("code and name are required")
	}
	var out Warehouse
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		w, err := s.repo.CreateWarehouse(ctx, in)
		if err != nil {
			return apperr.Internal("create warehouse", err)
		}
		if err := audit.Log(ctx, tx, "warehouses", w.ID, audit.Insert, nil, w); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = w
		return nil
	})
	return out, err
}

func (s *Service) UpdateWarehouse(ctx context.Context, id uuid.UUID, in UpsertWarehouseInput) (Warehouse, error) {
	var out Warehouse
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		before, err := s.repo.GetWarehouseByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("warehouse not found")
			}
			return apperr.Internal("load warehouse", err)
		}
		after, err := s.repo.UpdateWarehouse(ctx, id, in)
		if err != nil {
			return apperr.Internal("update warehouse", err)
		}
		if err := audit.Log(ctx, tx, "warehouses", id, audit.Update, before, after); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = after
		return nil
	})
	return out, err
}

func (s *Service) ListWarehouses(ctx context.Context) ([]Warehouse, error) {
	return s.repo.ListWarehouses(ctx)
}

func (s *Service) GetWarehouseByCode(ctx context.Context, code string) (Warehouse, error) {
	w, err := s.repo.GetWarehouseByCode(ctx, code)
	if errors.Is(err, pgx.ErrNoRows) {
		return Warehouse{}, apperr.Internal("resolve default warehouse "+code, err)
	}
	return w, err
}

// DefaultMaterialWarehouse / DefaultFinishedGoodsWarehouse are used by
// domains (production, purchasing, sales) that don't yet expose an explicit
// warehouse picker of their own.
func (s *Service) DefaultMaterialWarehouse(ctx context.Context) (Warehouse, error) {
	return s.GetWarehouseByCode(ctx, DefaultMaterialWarehouseCode)
}

func (s *Service) DefaultFinishedGoodsWarehouse(ctx context.Context) (Warehouse, error) {
	return s.GetWarehouseByCode(ctx, DefaultFinishedGoodsWarehouseCode)
}

func (s *Service) ListBalancesByWarehouse(ctx context.Context, warehouseID *uuid.UUID) ([]Balance, error) {
	return s.repo.ListBalancesByWarehouse(ctx, warehouseID)
}
