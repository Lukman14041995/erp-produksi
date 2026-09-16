package inventory

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/audit"
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/ranji/clothing-erp/internal/numbering"
	"github.com/shopspring/decimal"
)

func (r *Repository) InsertTransfer(ctx context.Context, t StockTransfer) (StockTransfer, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO stock_transfers (transfer_number, source_warehouse_id, destination_warehouse_id, transfer_date, notes)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, transfer_number, source_warehouse_id, destination_warehouse_id, status, transfer_date, notes, dispatched_at, received_at, created_at, updated_at
	`, t.TransferNumber, t.SourceWarehouseID, t.DestinationWarehouseID, t.TransferDate, t.Notes)
	return scanTransfer(row)
}

func scanTransfer(row pgx.Row) (StockTransfer, error) {
	var t StockTransfer
	err := row.Scan(&t.ID, &t.TransferNumber, &t.SourceWarehouseID, &t.DestinationWarehouseID, &t.Status,
		&t.TransferDate, &t.Notes, &t.DispatchedAt, &t.ReceivedAt, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

const transferColumns = `id, transfer_number, source_warehouse_id, destination_warehouse_id, status, transfer_date, notes, dispatched_at, received_at, created_at, updated_at`

func (r *Repository) InsertTransferItem(ctx context.Context, i StockTransferItem) (StockTransferItem, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO stock_transfer_items (stock_transfer_id, item_type, material_id, product_id, product_size_id, qty, unit_cost)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, stock_transfer_id, item_type, material_id, product_id, product_size_id, qty, unit_cost
	`, i.StockTransferID, i.ItemType, i.MaterialID, i.ProductID, i.ProductSizeID, i.Qty, i.UnitCost)
	var out StockTransferItem
	err := row.Scan(&out.ID, &out.StockTransferID, &out.ItemType, &out.MaterialID, &out.ProductID, &out.ProductSizeID, &out.Qty, &out.UnitCost)
	return out, err
}

func (r *Repository) UpdateTransferItemCost(ctx context.Context, id uuid.UUID, unitCost decimal.Decimal) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE stock_transfer_items SET unit_cost=$2 WHERE id=$1`, id, unitCost)
	return err
}

func (r *Repository) ListTransferItems(ctx context.Context, transferID uuid.UUID) ([]StockTransferItem, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT id, stock_transfer_id, item_type, material_id, product_id, product_size_id, qty, unit_cost
		FROM stock_transfer_items WHERE stock_transfer_id=$1
	`, transferID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []StockTransferItem{}
	for rows.Next() {
		var i StockTransferItem
		if err := rows.Scan(&i.ID, &i.StockTransferID, &i.ItemType, &i.MaterialID, &i.ProductID, &i.ProductSizeID, &i.Qty, &i.UnitCost); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *Repository) GetTransferForUpdate(ctx context.Context, id uuid.UUID) (StockTransfer, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+transferColumns+` FROM stock_transfers WHERE id=$1 FOR UPDATE`, id)
	return scanTransfer(row)
}

func (r *Repository) GetTransferByID(ctx context.Context, id uuid.UUID) (StockTransfer, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+transferColumns+` FROM stock_transfers WHERE id=$1`, id)
	return scanTransfer(row)
}

func (r *Repository) ListTransfers(ctx context.Context, limit, offset int) ([]StockTransfer, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+transferColumns+` FROM stock_transfers ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []StockTransfer{}
	for rows.Next() {
		t, err := scanTransfer(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repository) SetTransferStatus(ctx context.Context, id uuid.UUID, status TransferStatus) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE stock_transfers SET status=$2 WHERE id=$1`, id, status)
	return err
}

func (r *Repository) SetTransferDispatchedAt(ctx context.Context, id uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE stock_transfers SET dispatched_at=now() WHERE id=$1`, id)
	return err
}

func (r *Repository) SetTransferReceivedAt(ctx context.Context, id uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE stock_transfers SET received_at=now() WHERE id=$1`, id)
	return err
}

// ---- Service ----

// CreateTransfer records the request only -- no stock moves until Dispatch.
func (s *Service) CreateTransfer(ctx context.Context, in CreateTransferInput) (StockTransfer, error) {
	if in.SourceWarehouseID == uuid.Nil || in.DestinationWarehouseID == uuid.Nil {
		return StockTransfer{}, apperr.Validation("source_warehouse_id and destination_warehouse_id are required")
	}
	if in.SourceWarehouseID == in.DestinationWarehouseID {
		return StockTransfer{}, apperr.Validation("source and destination warehouses must differ")
	}
	if len(in.Items) == 0 {
		return StockTransfer{}, apperr.Validation("transfer must have at least one item")
	}

	transferDate := time.Now()
	if in.TransferDate != "" {
		parsed, err := time.Parse("2006-01-02", in.TransferDate)
		if err != nil {
			return StockTransfer{}, apperr.Validation("transfer_date must be YYYY-MM-DD")
		}
		transferDate = parsed
	}

	var out StockTransfer
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		number, err := numbering.Generate(ctx, tx, numbering.StockTransfer, transferDate)
		if err != nil {
			return apperr.Internal("generate transfer number", err)
		}

		transfer, err := s.repo.InsertTransfer(ctx, StockTransfer{
			TransferNumber: number, SourceWarehouseID: in.SourceWarehouseID, DestinationWarehouseID: in.DestinationWarehouseID,
			TransferDate: transferDate, Notes: in.Notes,
		})
		if err != nil {
			return apperr.Internal("create stock transfer", err)
		}

		for _, it := range in.Items {
			if it.Qty.IsZero() || it.Qty.IsNegative() {
				return apperr.Validation("item qty must be greater than zero")
			}
			if it.ItemType == ItemMaterial && it.MaterialID == nil {
				return apperr.Validation("material_id is required for MATERIAL items")
			}
			if it.ItemType == ItemProduct && it.ProductID == nil {
				return apperr.Validation("product_id is required for PRODUCT items")
			}
			item, err := s.repo.InsertTransferItem(ctx, StockTransferItem{
				StockTransferID: transfer.ID, ItemType: it.ItemType, MaterialID: it.MaterialID,
				ProductID: it.ProductID, ProductSizeID: it.ProductSizeID, Qty: it.Qty,
			})
			if err != nil {
				return apperr.Internal("create transfer item", err)
			}
			transfer.Items = append(transfer.Items, item)
		}

		if err := audit.Log(ctx, tx, "stock_transfers", transfer.ID, audit.Insert, nil, transfer); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = transfer
		return nil
	})
	return out, err
}

// DispatchTransfer removes stock from the source warehouse now (TRANSFER_OUT),
// snapshotting the cost each item leaves at. Stock is "in transit" between
// dispatch and receive -- it exists in neither warehouse's balance during
// that window, mirroring the physical reality of a truck on the road.
func (s *Service) DispatchTransfer(ctx context.Context, id uuid.UUID) (StockTransfer, error) {
	var out StockTransfer
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		transfer, err := s.repo.GetTransferForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("stock transfer not found")
			}
			return apperr.Internal("load stock transfer", err)
		}
		if transfer.Status != TransferDraft {
			return apperr.Conflict("only DRAFT transfers can be dispatched")
		}

		items, err := s.repo.ListTransferItems(ctx, id)
		if err != nil {
			return apperr.Internal("load transfer items", err)
		}

		for _, it := range items {
			result, err := s.RecordMovement(ctx, MovementInput{
				TxnType: TxnTransferOut, ItemType: it.ItemType, WarehouseID: transfer.SourceWarehouseID,
				MaterialID: it.MaterialID, ProductID: it.ProductID, ProductSizeID: it.ProductSizeID,
				QtyOut: it.Qty, RefType: "STOCK_TRANSFER", RefID: &transfer.ID, TxnDate: time.Now(),
				Notes: "Dispatch for " + transfer.TransferNumber,
			})
			if err != nil {
				return apperr.Wrapf(err, "dispatch transfer item")
			}
			if err := s.repo.UpdateTransferItemCost(ctx, it.ID, result.Transaction.UnitCost); err != nil {
				return apperr.Internal("snapshot transfer item cost", err)
			}
		}

		if err := s.repo.SetTransferStatus(ctx, id, TransferDispatched); err != nil {
			return apperr.Internal("mark transfer dispatched", err)
		}
		if err := s.repo.SetTransferDispatchedAt(ctx, id); err != nil {
			return apperr.Internal("stamp dispatched_at", err)
		}
		if err := audit.Log(ctx, tx, "stock_transfers", id, audit.Update, transfer.Status, TransferDispatched); err != nil {
			return apperr.Internal("write audit log", err)
		}

		transfer.Status = TransferDispatched
		out = transfer
		return nil
	})
	return out, err
}

// ReceiveTransfer adds stock into the destination warehouse (TRANSFER_IN) at
// the cost snapshotted when it left the source.
func (s *Service) ReceiveTransfer(ctx context.Context, id uuid.UUID) (StockTransfer, error) {
	var out StockTransfer
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		transfer, err := s.repo.GetTransferForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("stock transfer not found")
			}
			return apperr.Internal("load stock transfer", err)
		}
		if transfer.Status != TransferDispatched {
			return apperr.Conflict("only DISPATCHED transfers can be received")
		}

		items, err := s.repo.ListTransferItems(ctx, id)
		if err != nil {
			return apperr.Internal("load transfer items", err)
		}

		for _, it := range items {
			if _, err := s.RecordMovement(ctx, MovementInput{
				TxnType: TxnTransferIn, ItemType: it.ItemType, WarehouseID: transfer.DestinationWarehouseID,
				MaterialID: it.MaterialID, ProductID: it.ProductID, ProductSizeID: it.ProductSizeID,
				QtyIn: it.Qty, UnitCost: it.UnitCost, RefType: "STOCK_TRANSFER", RefID: &transfer.ID, TxnDate: time.Now(),
				Notes: "Receipt for " + transfer.TransferNumber,
			}); err != nil {
				return apperr.Wrapf(err, "receive transfer item")
			}
		}

		if err := s.repo.SetTransferStatus(ctx, id, TransferReceived); err != nil {
			return apperr.Internal("mark transfer received", err)
		}
		if err := s.repo.SetTransferReceivedAt(ctx, id); err != nil {
			return apperr.Internal("stamp received_at", err)
		}
		if err := audit.Log(ctx, tx, "stock_transfers", id, audit.Update, transfer.Status, TransferReceived); err != nil {
			return apperr.Internal("write audit log", err)
		}

		transfer.Status = TransferReceived
		out = transfer
		return nil
	})
	return out, err
}

func (s *Service) CancelTransfer(ctx context.Context, id uuid.UUID) (StockTransfer, error) {
	var out StockTransfer
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		transfer, err := s.repo.GetTransferForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("stock transfer not found")
			}
			return apperr.Internal("load stock transfer", err)
		}
		if transfer.Status != TransferDraft {
			return apperr.Conflict("only DRAFT transfers can be cancelled; a dispatched transfer must be received then corrected with a reverse transfer")
		}
		if err := s.repo.SetTransferStatus(ctx, id, TransferCancelled); err != nil {
			return apperr.Internal("cancel transfer", err)
		}
		if err := audit.Log(ctx, tx, "stock_transfers", id, audit.Update, transfer.Status, TransferCancelled); err != nil {
			return apperr.Internal("write audit log", err)
		}
		transfer.Status = TransferCancelled
		out = transfer
		return nil
	})
	return out, err
}

func (s *Service) GetTransfer(ctx context.Context, id uuid.UUID) (StockTransfer, error) {
	transfer, err := s.repo.GetTransferByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return StockTransfer{}, apperr.NotFound("stock transfer not found")
	}
	if err != nil {
		return StockTransfer{}, err
	}
	items, err := s.repo.ListTransferItems(ctx, id)
	if err != nil {
		return StockTransfer{}, apperr.Internal("load transfer items", err)
	}
	transfer.Items = items
	return transfer, nil
}

func (s *Service) ListTransfers(ctx context.Context, limit, offset int) ([]StockTransfer, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListTransfers(ctx, limit, offset)
}
