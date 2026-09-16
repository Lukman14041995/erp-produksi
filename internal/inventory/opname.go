package inventory

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ranji/clothing-erp/internal/accounting"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/audit"
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/ranji/clothing-erp/internal/numbering"
	"github.com/shopspring/decimal"
)

const (
	acctMaterialStock   = "1-1200" // Persediaan Bahan Baku
	acctFinishedGoods   = "1-1220" // Persediaan Barang Jadi
	acctOpnameVariance  = "5-9200" // Selisih Stock Opname
)

const opnameColumns = `id, opname_number, warehouse_id, opname_date, status, notes, journal_id, posted_at, created_at, updated_at`

func scanOpname(row pgx.Row) (StockOpname, error) {
	var o StockOpname
	err := row.Scan(&o.ID, &o.OpnameNumber, &o.WarehouseID, &o.OpnameDate, &o.Status, &o.Notes, &o.JournalID, &o.PostedAt, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}

func (r *Repository) InsertOpname(ctx context.Context, o StockOpname) (StockOpname, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO stock_opnames (opname_number, warehouse_id, opname_date, notes)
		VALUES ($1,$2,$3,$4)
		RETURNING `+opnameColumns, o.OpnameNumber, o.WarehouseID, o.OpnameDate, o.Notes)
	return scanOpname(row)
}

func (r *Repository) InsertOpnameItem(ctx context.Context, i StockOpnameItem) (StockOpnameItem, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO stock_opname_items (stock_opname_id, item_type, material_id, product_id, product_size_id, system_qty, actual_qty, unit_cost, variance_qty, variance_amount)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, stock_opname_id, item_type, material_id, product_id, product_size_id, system_qty, actual_qty, unit_cost, variance_qty, variance_amount
	`, i.StockOpnameID, i.ItemType, i.MaterialID, i.ProductID, i.ProductSizeID, i.SystemQty, i.ActualQty, i.UnitCost, i.VarianceQty, i.VarianceAmount)
	var out StockOpnameItem
	err := row.Scan(&out.ID, &out.StockOpnameID, &out.ItemType, &out.MaterialID, &out.ProductID, &out.ProductSizeID,
		&out.SystemQty, &out.ActualQty, &out.UnitCost, &out.VarianceQty, &out.VarianceAmount)
	return out, err
}

func (r *Repository) ListOpnameItems(ctx context.Context, opnameID uuid.UUID) ([]StockOpnameItem, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT id, stock_opname_id, item_type, material_id, product_id, product_size_id, system_qty, actual_qty, unit_cost, variance_qty, variance_amount
		FROM stock_opname_items WHERE stock_opname_id=$1
	`, opnameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []StockOpnameItem{}
	for rows.Next() {
		var i StockOpnameItem
		if err := rows.Scan(&i.ID, &i.StockOpnameID, &i.ItemType, &i.MaterialID, &i.ProductID, &i.ProductSizeID,
			&i.SystemQty, &i.ActualQty, &i.UnitCost, &i.VarianceQty, &i.VarianceAmount); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *Repository) GetOpnameForUpdate(ctx context.Context, id uuid.UUID) (StockOpname, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+opnameColumns+` FROM stock_opnames WHERE id=$1 FOR UPDATE`, id)
	return scanOpname(row)
}

func (r *Repository) GetOpnameByID(ctx context.Context, id uuid.UUID) (StockOpname, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+opnameColumns+` FROM stock_opnames WHERE id=$1`, id)
	return scanOpname(row)
}

func (r *Repository) ListOpnames(ctx context.Context, limit, offset int) ([]StockOpname, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+opnameColumns+` FROM stock_opnames ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []StockOpname{}
	for rows.Next() {
		o, err := scanOpname(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// journalID is nil when the opname had no variance (nothing to post to the
// GL) -- passed through as SQL NULL, never the zero UUID, since journal_id
// carries a foreign key to journal_entries.
func (r *Repository) SetOpnamePosted(ctx context.Context, id uuid.UUID, journalID *uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		UPDATE stock_opnames SET status='POSTED', journal_id=$2, posted_at=now() WHERE id=$1
	`, id, journalID)
	return err
}

func (r *Repository) SetOpnameCancelled(ctx context.Context, id uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE stock_opnames SET status='CANCELLED' WHERE id=$1`, id)
	return err
}

// ---- Service ----

// opnameAccountingIface avoids importing internal/accounting's concrete
// *Service type as a hard dependency here (accounting doesn't import
// inventory, but several other domains import both, so this keeps the
// dependency direction explicit) -- main.go passes the real
// *accounting.Service, which satisfies this interface structurally.
type opnameAccountingIface interface {
	PostJournal(ctx context.Context, sourceType accounting.SourceType, sourceID *uuid.UUID, journalDate time.Time, description string, lines []accounting.JournalLineInput) (accounting.JournalEntry, error)
}

// CreateOpname freezes system_qty and the current moving-average cost for
// each counted item from the warehouse's live balance, and records the
// physically counted actual_qty alongside it. Nothing touches the stock
// ledger or GL yet -- that only happens on Post, so a miscount can simply
// be discarded (never posted) instead of needing a correction entry.
func (s *Service) CreateOpname(ctx context.Context, in CreateOpnameInput) (StockOpname, error) {
	if in.WarehouseID == uuid.Nil {
		return StockOpname{}, apperr.Validation("warehouse_id is required")
	}
	if len(in.Items) == 0 {
		return StockOpname{}, apperr.Validation("opname must count at least one item")
	}

	opnameDate := time.Now()
	if in.OpnameDate != "" {
		parsed, err := time.Parse("2006-01-02", in.OpnameDate)
		if err != nil {
			return StockOpname{}, apperr.Validation("opname_date must be YYYY-MM-DD")
		}
		opnameDate = parsed
	}

	var out StockOpname
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := s.repo.GetWarehouseByID(ctx, in.WarehouseID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.Validation("unknown warehouse_id")
			}
			return apperr.Internal("load warehouse", err)
		}

		number, err := numbering.Generate(ctx, tx, numbering.StockOpname, opnameDate)
		if err != nil {
			return apperr.Internal("generate opname number", err)
		}

		opname, err := s.repo.InsertOpname(ctx, StockOpname{OpnameNumber: number, WarehouseID: in.WarehouseID, OpnameDate: opnameDate, Notes: in.Notes})
		if err != nil {
			return apperr.Internal("create stock opname", err)
		}

		for _, it := range in.Items {
			if it.ItemType == ItemMaterial && it.MaterialID == nil {
				return apperr.Validation("material_id is required for MATERIAL items")
			}
			if it.ItemType == ItemProduct && it.ProductID == nil {
				return apperr.Validation("product_id is required for PRODUCT items")
			}
			if it.ActualQty.IsNegative() {
				return apperr.Validation("actual_qty cannot be negative")
			}

			balance, err := s.repo.GetBalance(ctx, it.ItemType, in.WarehouseID, it.MaterialID, it.ProductID, it.ProductSizeID)
			if errors.Is(err, pgx.ErrNoRows) {
				balance = Balance{QtyOnHand: decimal.Zero, AvgUnitCost: decimal.Zero}
			} else if err != nil {
				return apperr.Internal("load current balance", err)
			}

			varianceQty := it.ActualQty.Sub(balance.QtyOnHand)
			varianceAmount := varianceQty.Mul(balance.AvgUnitCost)

			item, err := s.repo.InsertOpnameItem(ctx, StockOpnameItem{
				StockOpnameID: opname.ID, ItemType: it.ItemType, MaterialID: it.MaterialID, ProductID: it.ProductID, ProductSizeID: it.ProductSizeID,
				SystemQty: balance.QtyOnHand, ActualQty: it.ActualQty, UnitCost: balance.AvgUnitCost,
				VarianceQty: varianceQty, VarianceAmount: varianceAmount,
			})
			if err != nil {
				return apperr.Internal("create opname item", err)
			}
			opname.Items = append(opname.Items, item)
		}

		if err := audit.Log(ctx, tx, "stock_opnames", opname.ID, audit.Insert, nil, opname); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = opname
		return nil
	})
	return out, err
}

// PostOpname applies every item's variance to the stock ledger (via an
// ADJUSTMENT movement) and posts one journal covering the whole opname:
// each inventory account's net surplus/shortage nets against
// acctOpnameVariance, so surplus and shortage lines on different materials
// don't need separate journals.
func (s *Service) PostOpname(ctx context.Context, id uuid.UUID, accSvc opnameAccountingIface) (StockOpname, error) {
	var out StockOpname
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		opname, err := s.repo.GetOpnameForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("stock opname not found")
			}
			return apperr.Internal("load stock opname", err)
		}
		if opname.Status != OpnameDraft {
			return apperr.Conflict("only DRAFT opnames can be posted")
		}

		items, err := s.repo.ListOpnameItems(ctx, id)
		if err != nil {
			return apperr.Internal("load opname items", err)
		}

		type accountTotal struct{ debit, credit decimal.Decimal }
		totals := map[string]*accountTotal{}
		addDebit := func(code string, amt decimal.Decimal) {
			if totals[code] == nil {
				totals[code] = &accountTotal{}
			}
			totals[code].debit = totals[code].debit.Add(amt)
		}
		addCredit := func(code string, amt decimal.Decimal) {
			if totals[code] == nil {
				totals[code] = &accountTotal{}
			}
			totals[code].credit = totals[code].credit.Add(amt)
		}

		for _, it := range items {
			if it.VarianceQty.IsZero() {
				continue
			}

			invAcct := acctMaterialStock
			if it.ItemType == ItemProduct {
				invAcct = acctFinishedGoods
			}

			if it.VarianceAmount.IsPositive() {
				// Surplus: DR inventory / CR variance (gain).
				addDebit(invAcct, it.VarianceAmount)
				addCredit(acctOpnameVariance, it.VarianceAmount)
			} else {
				// Shortage: DR variance (loss) / CR inventory.
				amt := it.VarianceAmount.Neg()
				addDebit(acctOpnameVariance, amt)
				addCredit(invAcct, amt)
			}

			// Adjust the stock ledger to the physically counted quantity.
			movement := MovementInput{
				TxnType: TxnAdjustment, ItemType: it.ItemType, WarehouseID: opname.WarehouseID,
				MaterialID: it.MaterialID, ProductID: it.ProductID, ProductSizeID: it.ProductSizeID,
				UnitCost: it.UnitCost, RefType: "STOCK_OPNAME", RefID: &opname.ID, TxnDate: opname.OpnameDate,
				Notes: "Stock opname " + opname.OpnameNumber,
			}
			if it.VarianceQty.IsPositive() {
				movement.QtyIn = it.VarianceQty
			} else {
				movement.QtyOut = it.VarianceQty.Neg()
			}
			if _, err := s.RecordMovement(ctx, movement); err != nil {
				return apperr.Wrapf(err, "adjust stock for opname item")
			}
		}

		if len(totals) == 0 {
			// No discrepancies found: nothing to post, just close the sheet.
			if err := s.repo.SetOpnamePosted(ctx, id, nil); err != nil {
				return apperr.Internal("mark opname posted", err)
			}
			opname.Status = OpnamePosted
			out = opname
			return nil
		}

		var lines []accounting.JournalLineInput
		for code, t := range totals {
			net := t.debit.Sub(t.credit)
			switch {
			case net.IsPositive():
				lines = append(lines, accounting.JournalLineInput{AccountCode: code, Debit: net, Description: "Stock opname " + opname.OpnameNumber})
			case net.IsNegative():
				lines = append(lines, accounting.JournalLineInput{AccountCode: code, Credit: net.Neg(), Description: "Stock opname " + opname.OpnameNumber})
			}
		}

		journal, err := accSvc.PostJournal(ctx, accounting.SourceStockOpname, &opname.ID, opname.OpnameDate,
			"Stock opname "+opname.OpnameNumber, lines)
		if err != nil {
			return apperr.Wrapf(err, "post stock opname journal")
		}

		if err := s.repo.SetOpnamePosted(ctx, id, &journal.ID); err != nil {
			return apperr.Internal("mark opname posted", err)
		}
		if err := audit.Log(ctx, tx, "stock_opnames", id, audit.Update, opname.Status, OpnamePosted); err != nil {
			return apperr.Internal("write audit log", err)
		}

		opname.Status = OpnamePosted
		opname.JournalID = &journal.ID
		opname.Items = items
		out = opname
		return nil
	})
	return out, err
}

func (s *Service) CancelOpname(ctx context.Context, id uuid.UUID) (StockOpname, error) {
	var out StockOpname
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		opname, err := s.repo.GetOpnameForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("stock opname not found")
			}
			return apperr.Internal("load stock opname", err)
		}
		if opname.Status != OpnameDraft {
			return apperr.Conflict("only DRAFT opnames can be cancelled")
		}
		if err := s.repo.SetOpnameCancelled(ctx, id); err != nil {
			return apperr.Internal("cancel opname", err)
		}
		if err := audit.Log(ctx, tx, "stock_opnames", id, audit.Update, opname.Status, OpnameCancelled); err != nil {
			return apperr.Internal("write audit log", err)
		}
		opname.Status = OpnameCancelled
		out = opname
		return nil
	})
	return out, err
}

func (s *Service) GetOpname(ctx context.Context, id uuid.UUID) (StockOpname, error) {
	opname, err := s.repo.GetOpnameByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return StockOpname{}, apperr.NotFound("stock opname not found")
	}
	if err != nil {
		return StockOpname{}, err
	}
	items, err := s.repo.ListOpnameItems(ctx, id)
	if err != nil {
		return StockOpname{}, apperr.Internal("load opname items", err)
	}
	opname.Items = items
	return opname, nil
}

func (s *Service) ListOpnames(ctx context.Context, limit, offset int) ([]StockOpname, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListOpnames(ctx, limit, offset)
}
