package production

import (
	"context"

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

// ---- BOM ----

func (r *Repository) InsertBOMHeader(ctx context.Context, productID uuid.UUID, name string) (BOMHeader, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO bom_headers (product_id, name) VALUES ($1,$2)
		RETURNING id, product_id, name, version, is_active
	`, productID, name)
	var h BOMHeader
	err := row.Scan(&h.ID, &h.ProductID, &h.Name, &h.Version, &h.IsActive)
	return h, err
}

func (r *Repository) InsertBOMLine(ctx context.Context, l BOMLine) (BOMLine, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO bom_lines (bom_id, material_id, qty_per_unit, uom) VALUES ($1,$2,$3,$4)
		RETURNING id, bom_id, material_id, qty_per_unit, uom
	`, l.BOMID, l.MaterialID, l.QtyPerUnit, l.UOM)
	var out BOMLine
	err := row.Scan(&out.ID, &out.BOMID, &out.MaterialID, &out.QtyPerUnit, &out.UOM)
	return out, err
}

func (r *Repository) GetBOMHeader(ctx context.Context, id uuid.UUID) (BOMHeader, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT id, product_id, name, version, is_active FROM bom_headers WHERE id=$1`, id)
	var h BOMHeader
	err := row.Scan(&h.ID, &h.ProductID, &h.Name, &h.Version, &h.IsActive)
	return h, err
}

func (r *Repository) ListBOMLines(ctx context.Context, bomID uuid.UUID) ([]BOMLine, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT id, bom_id, material_id, qty_per_unit, uom FROM bom_lines WHERE bom_id=$1`, bomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []BOMLine{}
	for rows.Next() {
		var l BOMLine
		if err := rows.Scan(&l.ID, &l.BOMID, &l.MaterialID, &l.QtyPerUnit, &l.UOM); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *Repository) ListBOMsByProduct(ctx context.Context, productID uuid.UUID) ([]BOMHeader, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT id, product_id, name, version, is_active FROM bom_headers WHERE product_id=$1 ORDER BY version DESC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []BOMHeader{}
	for rows.Next() {
		var h BOMHeader
		if err := rows.Scan(&h.ID, &h.ProductID, &h.Name, &h.Version, &h.IsActive); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// ---- Production Orders ----

const orderColumns = `id, prod_number, sales_order_id, product_id, bom_id, planned_qty, finished_qty, production_status, start_date, end_date, notes, created_at, updated_at`

func scanOrder(row pgx.Row) (Order, error) {
	var o Order
	err := row.Scan(&o.ID, &o.ProdNumber, &o.SalesOrderID, &o.ProductID, &o.BOMID, &o.PlannedQty, &o.FinishedQty, &o.Status, &o.StartDate, &o.EndDate, &o.Notes, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}

func (r *Repository) InsertOrder(ctx context.Context, o Order) (Order, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO production_orders (prod_number, sales_order_id, product_id, bom_id, planned_qty, notes)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING `+orderColumns, o.ProdNumber, o.SalesOrderID, o.ProductID, o.BOMID, o.PlannedQty, o.Notes)
	return scanOrder(row)
}

func (r *Repository) InsertOrderItem(ctx context.Context, i OrderItem) (OrderItem, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO production_order_items (production_order_id, product_size_id, planned_qty)
		VALUES ($1,$2,$3)
		RETURNING id, production_order_id, product_size_id, planned_qty, finished_qty
	`, i.ProductionOrderID, i.ProductSizeID, i.PlannedQty)
	var out OrderItem
	err := row.Scan(&out.ID, &out.ProductionOrderID, &out.ProductSizeID, &out.PlannedQty, &out.FinishedQty)
	return out, err
}

func (r *Repository) GetOrderForUpdate(ctx context.Context, id uuid.UUID) (Order, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+orderColumns+` FROM production_orders WHERE id=$1 FOR UPDATE`, id)
	return scanOrder(row)
}

func (r *Repository) GetOrderByID(ctx context.Context, id uuid.UUID) (Order, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+orderColumns+` FROM production_orders WHERE id=$1`, id)
	return scanOrder(row)
}

func (r *Repository) ListOrders(ctx context.Context, limit, offset int) ([]Order, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+orderColumns+` FROM production_orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
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

func (r *Repository) ListOrderItems(ctx context.Context, orderID uuid.UUID) ([]OrderItem, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT id, production_order_id, product_size_id, planned_qty, finished_qty
		FROM production_order_items WHERE production_order_id=$1
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []OrderItem{}
	for rows.Next() {
		var i OrderItem
		if err := rows.Scan(&i.ID, &i.ProductionOrderID, &i.ProductSizeID, &i.PlannedQty, &i.FinishedQty); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *Repository) SetItemFinishedQty(ctx context.Context, itemID uuid.UUID, qty decimal.Decimal) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE production_order_items SET finished_qty=$2 WHERE id=$1`, itemID, qty)
	return err
}

func (r *Repository) SetStatus(ctx context.Context, id uuid.UUID, status Status) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE production_orders SET production_status=$2 WHERE id=$1`, id, status)
	return err
}

func (r *Repository) SetStartDate(ctx context.Context, id uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE production_orders SET start_date=CURRENT_DATE WHERE id=$1`, id)
	return err
}

func (r *Repository) SetEndDateAndFinishedQty(ctx context.Context, id uuid.UUID, finishedQty decimal.Decimal) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE production_orders SET end_date=CURRENT_DATE, finished_qty=$2 WHERE id=$1`, id, finishedQty)
	return err
}

// ---- Materials / Labor / Overhead ----

func (r *Repository) InsertPlannedMaterial(ctx context.Context, m Material) (Material, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO production_materials (production_order_id, material_id, planned_qty)
		VALUES ($1,$2,$3)
		RETURNING id, production_order_id, material_id, planned_qty, issued_qty, unit_cost, total_cost, issued_at
	`, m.ProductionOrderID, m.MaterialID, m.PlannedQty)
	var out Material
	err := row.Scan(&out.ID, &out.ProductionOrderID, &out.MaterialID, &out.PlannedQty, &out.IssuedQty, &out.UnitCost, &out.TotalCost, &out.IssuedAt)
	return out, err
}

func (r *Repository) GetMaterialRowForUpdate(ctx context.Context, id uuid.UUID) (Material, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		SELECT id, production_order_id, material_id, planned_qty, issued_qty, unit_cost, total_cost, issued_at
		FROM production_materials WHERE id=$1 FOR UPDATE
	`, id)
	var out Material
	err := row.Scan(&out.ID, &out.ProductionOrderID, &out.MaterialID, &out.PlannedQty, &out.IssuedQty, &out.UnitCost, &out.TotalCost, &out.IssuedAt)
	return out, err
}

func (r *Repository) UpdateMaterialIssue(ctx context.Context, id uuid.UUID, issuedQty, unitCost, totalCost decimal.Decimal) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		UPDATE production_materials SET issued_qty=$2, unit_cost=$3, total_cost=$4, issued_at=now() WHERE id=$1
	`, id, issuedQty, unitCost, totalCost)
	return err
}

func (r *Repository) ListMaterials(ctx context.Context, orderID uuid.UUID) ([]Material, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT id, production_order_id, material_id, planned_qty, issued_qty, unit_cost, total_cost, issued_at
		FROM production_materials WHERE production_order_id=$1
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Material{}
	for rows.Next() {
		var m Material
		if err := rows.Scan(&m.ID, &m.ProductionOrderID, &m.MaterialID, &m.PlannedQty, &m.IssuedQty, &m.UnitCost, &m.TotalCost, &m.IssuedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) InsertLabor(ctx context.Context, l Labor) (Labor, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO production_labor (production_order_id, description, hours, rate, total_cost)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, production_order_id, description, hours, rate, total_cost
	`, l.ProductionOrderID, l.Description, l.Hours, l.Rate, l.TotalCost)
	var out Labor
	err := row.Scan(&out.ID, &out.ProductionOrderID, &out.Description, &out.Hours, &out.Rate, &out.TotalCost)
	return out, err
}

func (r *Repository) ListLabor(ctx context.Context, orderID uuid.UUID) ([]Labor, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT id, production_order_id, description, hours, rate, total_cost FROM production_labor WHERE production_order_id=$1
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Labor{}
	for rows.Next() {
		var l Labor
		if err := rows.Scan(&l.ID, &l.ProductionOrderID, &l.Description, &l.Hours, &l.Rate, &l.TotalCost); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *Repository) InsertOverhead(ctx context.Context, o Overhead) (Overhead, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO production_overheads (production_order_id, description, allocation_basis, amount)
		VALUES ($1,$2,$3,$4)
		RETURNING id, production_order_id, description, allocation_basis, amount
	`, o.ProductionOrderID, o.Description, o.AllocationBasis, o.Amount)
	var out Overhead
	err := row.Scan(&out.ID, &out.ProductionOrderID, &out.Description, &out.AllocationBasis, &out.Amount)
	return out, err
}

func (r *Repository) ListOverheads(ctx context.Context, orderID uuid.UUID) ([]Overhead, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT id, production_order_id, description, allocation_basis, amount FROM production_overheads WHERE production_order_id=$1
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Overhead{}
	for rows.Next() {
		var o Overhead
		if err := rows.Scan(&o.ID, &o.ProductionOrderID, &o.Description, &o.AllocationBasis, &o.Amount); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// ---- HPP snapshot ----

func (r *Repository) InsertCost(ctx context.Context, c Cost) (Cost, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO production_costs (production_order_id, total_material_cost, total_labor_cost, total_overhead_cost, total_cost, finished_qty, unit_cost, journal_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, production_order_id, total_material_cost, total_labor_cost, total_overhead_cost, total_cost, finished_qty, unit_cost, journal_id, computed_at
	`, c.ProductionOrderID, c.TotalMaterialCost, c.TotalLaborCost, c.TotalOverheadCost, c.TotalCost, c.FinishedQty, c.UnitCost, c.JournalID)
	var out Cost
	err := row.Scan(&out.ID, &out.ProductionOrderID, &out.TotalMaterialCost, &out.TotalLaborCost, &out.TotalOverheadCost, &out.TotalCost, &out.FinishedQty, &out.UnitCost, &out.JournalID, &out.ComputedAt)
	return out, err
}

func (r *Repository) GetCostByOrder(ctx context.Context, orderID uuid.UUID) (Cost, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		SELECT id, production_order_id, total_material_cost, total_labor_cost, total_overhead_cost, total_cost, finished_qty, unit_cost, journal_id, computed_at
		FROM production_costs WHERE production_order_id=$1
	`, orderID)
	var out Cost
	err := row.Scan(&out.ID, &out.ProductionOrderID, &out.TotalMaterialCost, &out.TotalLaborCost, &out.TotalOverheadCost, &out.TotalCost, &out.FinishedQty, &out.UnitCost, &out.JournalID, &out.ComputedAt)
	return out, err
}
