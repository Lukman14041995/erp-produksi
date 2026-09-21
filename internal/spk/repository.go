package spk

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/shopspring/decimal"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const orderColumns = `id, spk_number, sales_order_id, so_number, quotation_id, quotation_number, customer_name, product_type_id, product_type_code, product_type_name, composition_summary, total_qty, status, current_stage_code, current_stage_name, progress_pct, design_done, production_done, shipped, material_cost_total, labor_cost_total, overhead_cost_total, notes, completed_at, created_at, updated_at`

func scanOrder(row pgx.Row) (Order, error) {
	var o Order
	err := row.Scan(&o.ID, &o.SPKNumber, &o.SalesOrderID, &o.SONumber, &o.QuotationID, &o.QuotationNumber, &o.CustomerName, &o.ProductTypeID, &o.ProductTypeCode, &o.ProductTypeName, &o.CompositionSummary,
		&o.TotalQty, &o.Status, &o.CurrentStageCode, &o.CurrentStageName, &o.ProgressPct, &o.DesignDone, &o.ProductionDone, &o.Shipped, &o.MaterialCostTotal, &o.LaborCostTotal, &o.OverheadCostTotal, &o.Notes, &o.CompletedAt, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}

func (r *Repository) InsertOrder(ctx context.Context, o Order) (Order, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO spk_orders (spk_number, sales_order_id, so_number, quotation_id, quotation_number, customer_name, product_type_id, product_type_code, product_type_name, composition_summary, total_qty, status, current_stage_code, current_stage_name)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING `+orderColumns,
		o.SPKNumber, o.SalesOrderID, o.SONumber, o.QuotationID, o.QuotationNumber, o.CustomerName, o.ProductTypeID, o.ProductTypeCode, o.ProductTypeName, o.CompositionSummary, o.TotalQty, o.Status, o.CurrentStageCode, o.CurrentStageName)
	return scanOrder(row)
}

func (r *Repository) GetOrderByID(ctx context.Context, id uuid.UUID) (Order, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+orderColumns+` FROM spk_orders WHERE id=$1`, id)
	return scanOrder(row)
}

func (r *Repository) ListOrders(ctx context.Context, productTypeCode string, limit, offset int) ([]Order, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT `+orderColumns+` FROM spk_orders WHERE product_type_code=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, productTypeCode, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Order{}
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// ListAllOrders backs the cross-jenis production board (/production/costing):
// every SPK regardless of product type, newest first.
func (r *Repository) ListAllOrders(ctx context.Context, limit, offset int) ([]Order, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT `+orderColumns+` FROM spk_orders ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Order{}
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *Repository) ListBySalesOrder(ctx context.Context, salesOrderID uuid.UUID) ([]Order, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT `+orderColumns+` FROM spk_orders WHERE sales_order_id=$1 ORDER BY created_at
	`, salesOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Order{}
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *Repository) UpdateOrderProgress(ctx context.Context, id uuid.UUID, status Status, currentStageCode, currentStageName string, progressPct decimal.Decimal, designDone, productionDone, shipped bool) error {
	var completedAtSQL string
	if status == Completed {
		completedAtSQL = `, completed_at = COALESCE(completed_at, now())`
	}
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		UPDATE spk_orders SET status=$2, current_stage_code=$3, current_stage_name=$4, progress_pct=$5,
		       design_done=$6, production_done=$7, shipped=$8`+completedAtSQL+` WHERE id=$1
	`, id, status, currentStageCode, currentStageName, progressPct, designDone, productionDone, shipped)
	return err
}

func (r *Repository) AllOrdersCompletedForSalesOrder(ctx context.Context, salesOrderID uuid.UUID) (bool, error) {
	var remaining int
	err := db.Q(ctx, r.pool).QueryRow(ctx, `
		SELECT count(*) FROM spk_orders WHERE sales_order_id=$1 AND status <> 'COMPLETED'
	`, salesOrderID).Scan(&remaining)
	return remaining == 0, err
}

// ---- Items ----

const itemColumns = `id, spk_order_id, sales_order_item_id, product_id, product_size_id, qty, fabric_name, variant_name, ink_name, size_code, fabric_material_id, ink_material_id`

func (r *Repository) InsertItem(ctx context.Context, i Item) (Item, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO spk_items (spk_order_id, sales_order_item_id, product_id, product_size_id, qty, fabric_name, variant_name, ink_name, size_code, fabric_material_id, ink_material_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING `+itemColumns,
		i.SPKOrderID, i.SalesOrderItemID, i.ProductID, i.ProductSizeID, i.Qty, i.FabricName, i.VariantName, i.InkName, i.SizeCode, i.FabricMaterialID, i.InkMaterialID)
	var out Item
	err := row.Scan(&out.ID, &out.SPKOrderID, &out.SalesOrderItemID, &out.ProductID, &out.ProductSizeID, &out.Qty, &out.FabricName, &out.VariantName, &out.InkName, &out.SizeCode, &out.FabricMaterialID, &out.InkMaterialID)
	return out, err
}

func (r *Repository) ListItems(ctx context.Context, spkOrderID uuid.UUID) ([]Item, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT `+itemColumns+` FROM spk_items WHERE spk_order_id=$1
	`, spkOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Item{}
	for rows.Next() {
		var i Item
		if err := rows.Scan(&i.ID, &i.SPKOrderID, &i.SalesOrderItemID, &i.ProductID, &i.ProductSizeID, &i.Qty, &i.FabricName, &i.VariantName, &i.InkName, &i.SizeCode, &i.FabricMaterialID, &i.InkMaterialID); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

// ---- Stages ----

const stageColumns = `id, spk_order_id, stage_seq, stage_code, stage_name, requires_qty, planned_qty, completed_qty, status, started_at, completed_at`

func scanStage(row pgx.Row) (Stage, error) {
	var s Stage
	err := row.Scan(&s.ID, &s.SPKOrderID, &s.StageSeq, &s.StageCode, &s.StageName, &s.RequiresQty, &s.PlannedQty, &s.CompletedQty, &s.Status, &s.StartedAt, &s.CompletedAt)
	return s, err
}

func (r *Repository) InsertStage(ctx context.Context, s Stage) (Stage, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO spk_stages (spk_order_id, stage_seq, stage_code, stage_name, requires_qty, planned_qty, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING `+stageColumns,
		s.SPKOrderID, s.StageSeq, s.StageCode, s.StageName, s.RequiresQty, s.PlannedQty, s.Status)
	return scanStage(row)
}

func (r *Repository) GetStageForUpdate(ctx context.Context, id uuid.UUID) (Stage, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+stageColumns+` FROM spk_stages WHERE id=$1 FOR UPDATE`, id)
	return scanStage(row)
}

func (r *Repository) ListStages(ctx context.Context, spkOrderID uuid.UUID) ([]Stage, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+stageColumns+` FROM spk_stages WHERE spk_order_id=$1 ORDER BY stage_seq`, spkOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Stage{}
	for rows.Next() {
		s, err := scanStage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) UpdateStageProgress(ctx context.Context, id uuid.UUID, completedQty decimal.Decimal, status StageStatus, completedAt *time.Time) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		UPDATE spk_stages SET completed_qty=$2, status=$3, completed_at=$4 WHERE id=$1
	`, id, completedQty, status, completedAt)
	return err
}

func (r *Repository) SetStageStarted(ctx context.Context, id uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE spk_stages SET started_at=COALESCE(started_at, now()) WHERE id=$1`, id)
	return err
}

// ---- Stage logs ----

func (r *Repository) InsertStageLog(ctx context.Context, l StageLog) (StageLog, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO spk_stage_logs (spk_stage_id, qty, note, logged_by)
		VALUES ($1,$2,$3,$4)
		RETURNING id, spk_stage_id, qty, note, logged_by, logged_at
	`, l.SPKStageID, l.Qty, l.Note, l.LoggedBy)
	var out StageLog
	err := row.Scan(&out.ID, &out.SPKStageID, &out.Qty, &out.Note, &out.LoggedBy, &out.LoggedAt)
	return out, err
}

// ---- Material usage ----

func (r *Repository) IncrementMaterialCost(ctx context.Context, spkOrderID uuid.UUID, amount decimal.Decimal) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		UPDATE spk_orders SET material_cost_total = material_cost_total + $2 WHERE id=$1
	`, spkOrderID, amount)
	return err
}

func (r *Repository) IncrementLaborCost(ctx context.Context, spkOrderID uuid.UUID, amount decimal.Decimal) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		UPDATE spk_orders SET labor_cost_total = labor_cost_total + $2 WHERE id=$1
	`, spkOrderID, amount)
	return err
}

func (r *Repository) IncrementOverheadCost(ctx context.Context, spkOrderID uuid.UUID, amount decimal.Decimal) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		UPDATE spk_orders SET overhead_cost_total = overhead_cost_total + $2 WHERE id=$1
	`, spkOrderID, amount)
	return err
}

// ---- Labor ----

func (r *Repository) InsertLabor(ctx context.Context, l Labor) (Labor, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO spk_labor (spk_order_id, spk_stage_id, description, hours, rate, total_cost, journal_id, logged_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, spk_order_id, spk_stage_id, description, hours, rate, total_cost, journal_id, logged_by, logged_at
	`, l.SPKOrderID, l.SPKStageID, l.Description, l.Hours, l.Rate, l.TotalCost, l.JournalID, l.LoggedBy)
	var out Labor
	err := row.Scan(&out.ID, &out.SPKOrderID, &out.SPKStageID, &out.Description, &out.Hours, &out.Rate, &out.TotalCost, &out.JournalID, &out.LoggedBy, &out.LoggedAt)
	return out, err
}

func (r *Repository) ListLabor(ctx context.Context, spkOrderID uuid.UUID) ([]Labor, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT l.id, l.spk_order_id, l.spk_stage_id, l.description, l.hours, l.rate, l.total_cost, l.journal_id, l.logged_by, COALESCE(u.name, ''), l.logged_at
		FROM spk_labor l
		LEFT JOIN users u ON u.id = l.logged_by
		WHERE l.spk_order_id=$1 ORDER BY l.logged_at DESC
	`, spkOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Labor{}
	for rows.Next() {
		var l Labor
		if err := rows.Scan(&l.ID, &l.SPKOrderID, &l.SPKStageID, &l.Description, &l.Hours, &l.Rate, &l.TotalCost, &l.JournalID, &l.LoggedBy, &l.LoggedByName, &l.LoggedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// ---- Overhead ----

func (r *Repository) InsertOverhead(ctx context.Context, o Overhead) (Overhead, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO spk_overheads (spk_order_id, description, allocation_basis, amount, journal_id, logged_by)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, spk_order_id, description, allocation_basis, amount, journal_id, logged_by, logged_at
	`, o.SPKOrderID, o.Description, o.AllocationBasis, o.Amount, o.JournalID, o.LoggedBy)
	var out Overhead
	err := row.Scan(&out.ID, &out.SPKOrderID, &out.Description, &out.AllocationBasis, &out.Amount, &out.JournalID, &out.LoggedBy, &out.LoggedAt)
	return out, err
}

func (r *Repository) ListOverheads(ctx context.Context, spkOrderID uuid.UUID) ([]Overhead, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT o.id, o.spk_order_id, o.description, o.allocation_basis, o.amount, o.journal_id, o.logged_by, COALESCE(u.name, ''), o.logged_at
		FROM spk_overheads o
		LEFT JOIN users u ON u.id = o.logged_by
		WHERE o.spk_order_id=$1 ORDER BY o.logged_at DESC
	`, spkOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Overhead{}
	for rows.Next() {
		var o Overhead
		if err := rows.Scan(&o.ID, &o.SPKOrderID, &o.Description, &o.AllocationBasis, &o.Amount, &o.JournalID, &o.LoggedBy, &o.LoggedByName, &o.LoggedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *Repository) InsertMaterialUsage(ctx context.Context, u MaterialUsage) (MaterialUsage, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO spk_material_usages (spk_order_id, spk_stage_id, material_id, qty, unit_cost, total_cost, warehouse_id, inventory_txn_id, journal_id, note, logged_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, spk_order_id, spk_stage_id, material_id, qty, unit_cost, total_cost, warehouse_id, inventory_txn_id, journal_id, note, logged_by, logged_at
	`, u.SPKOrderID, u.SPKStageID, u.MaterialID, u.Qty, u.UnitCost, u.TotalCost, u.WarehouseID, u.InventoryTxnID, u.JournalID, u.Note, u.LoggedBy)
	var out MaterialUsage
	err := row.Scan(&out.ID, &out.SPKOrderID, &out.SPKStageID, &out.MaterialID, &out.Qty, &out.UnitCost, &out.TotalCost, &out.WarehouseID, &out.InventoryTxnID, &out.JournalID, &out.Note, &out.LoggedBy, &out.LoggedAt)
	return out, err
}

func (r *Repository) ListMaterialUsages(ctx context.Context, spkOrderID uuid.UUID) ([]MaterialUsage, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT u.id, u.spk_order_id, u.spk_stage_id, u.material_id, m.name, m.uom, u.qty, u.unit_cost, u.total_cost,
		       u.warehouse_id, u.inventory_txn_id, u.journal_id, u.note, u.logged_by, COALESCE(usr.name, ''), u.logged_at
		FROM spk_material_usages u
		JOIN materials m ON m.id = u.material_id
		LEFT JOIN users usr ON usr.id = u.logged_by
		WHERE u.spk_order_id=$1
		ORDER BY u.logged_at DESC
	`, spkOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MaterialUsage{}
	for rows.Next() {
		var u MaterialUsage
		if err := rows.Scan(&u.ID, &u.SPKOrderID, &u.SPKStageID, &u.MaterialID, &u.MaterialName, &u.MaterialUOM, &u.Qty, &u.UnitCost, &u.TotalCost,
			&u.WarehouseID, &u.InventoryTxnID, &u.JournalID, &u.Note, &u.LoggedBy, &u.LoggedByName, &u.LoggedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *Repository) ListStageLogs(ctx context.Context, stageID uuid.UUID) ([]StageLog, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT l.id, l.spk_stage_id, l.qty, l.note, l.logged_by, l.logged_at, COALESCE(u.name, '')
		FROM spk_stage_logs l
		LEFT JOIN users u ON u.id = l.logged_by
		WHERE l.spk_stage_id=$1 ORDER BY l.logged_at DESC
	`, stageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []StageLog{}
	for rows.Next() {
		var l StageLog
		if err := rows.Scan(&l.ID, &l.SPKStageID, &l.Qty, &l.Note, &l.LoggedBy, &l.LoggedAt, &l.LoggedByName); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}
