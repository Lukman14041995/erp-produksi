package inventory

import (
	"context"
	"errors"

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

// ---- Warehouses ----

const warehouseColumns = `id, code, name, address, is_active, created_at, updated_at`

func scanWarehouse(row pgx.Row) (Warehouse, error) {
	var w Warehouse
	err := row.Scan(&w.ID, &w.Code, &w.Name, &w.Address, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	return w, err
}

func (r *Repository) CreateWarehouse(ctx context.Context, in UpsertWarehouseInput) (Warehouse, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO warehouses (code, name, address) VALUES ($1,$2,$3)
		RETURNING `+warehouseColumns, in.Code, in.Name, in.Address)
	return scanWarehouse(row)
}

func (r *Repository) UpdateWarehouse(ctx context.Context, id uuid.UUID, in UpsertWarehouseInput) (Warehouse, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE warehouses SET name=$2, address=$3, is_active=$4 WHERE id=$1
		RETURNING `+warehouseColumns, id, in.Name, in.Address, isActive)
	return scanWarehouse(row)
}

func (r *Repository) GetWarehouseByID(ctx context.Context, id uuid.UUID) (Warehouse, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+warehouseColumns+` FROM warehouses WHERE id=$1`, id)
	return scanWarehouse(row)
}

func (r *Repository) GetWarehouseByCode(ctx context.Context, code string) (Warehouse, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+warehouseColumns+` FROM warehouses WHERE code=$1`, code)
	return scanWarehouse(row)
}

func (r *Repository) ListWarehouses(ctx context.Context) ([]Warehouse, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+warehouseColumns+` FROM warehouses ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Warehouse{}
	for rows.Next() {
		w, err := scanWarehouse(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// ---- Transactions & balances ----

const txnColumns = `id, txn_type, item_type, warehouse_id, material_id, product_id, product_size_id, qty_in, qty_out, unit_cost, total_cost, ref_type, ref_id, txn_date, notes, created_at`

func scanTxn(row pgx.Row) (Transaction, error) {
	var t Transaction
	err := row.Scan(&t.ID, &t.TxnType, &t.ItemType, &t.WarehouseID, &t.MaterialID, &t.ProductID, &t.ProductSizeID,
		&t.QtyIn, &t.QtyOut, &t.UnitCost, &t.TotalCost, &t.RefType, &t.RefID, &t.TxnDate, &t.Notes, &t.CreatedAt)
	return t, err
}

func (r *Repository) InsertTransaction(ctx context.Context, t Transaction) (Transaction, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO inventory_transactions
			(txn_type, item_type, warehouse_id, material_id, product_id, product_size_id, qty_in, qty_out, unit_cost, total_cost, ref_type, ref_id, txn_date, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING `+txnColumns,
		t.TxnType, t.ItemType, t.WarehouseID, t.MaterialID, t.ProductID, t.ProductSizeID, t.QtyIn, t.QtyOut, t.UnitCost, t.TotalCost, t.RefType, t.RefID, t.TxnDate, t.Notes)
	return scanTxn(row)
}

// GetBalanceForUpdate locks (or creates, if absent) the balance row for the
// given item+warehouse so concurrent movements against the same slot
// serialize instead of racing on the read-modify-write of qty_on_hand /
// avg_unit_cost.
func (r *Repository) GetBalanceForUpdate(ctx context.Context, itemType ItemType, warehouseID uuid.UUID, materialID, productID, productSizeID *uuid.UUID) (Balance, error) {
	q := db.Q(ctx, r.pool)

	row := q.QueryRow(ctx, `
		SELECT item_type, warehouse_id, material_id, product_id, product_size_id, qty_on_hand, avg_unit_cost
		FROM inventory_balances
		WHERE item_type = $1 AND warehouse_id = $2
		  AND material_id IS NOT DISTINCT FROM $3
		  AND product_id IS NOT DISTINCT FROM $4
		  AND product_size_id IS NOT DISTINCT FROM $5
		FOR UPDATE
	`, itemType, warehouseID, materialID, productID, productSizeID)

	var b Balance
	err := row.Scan(&b.ItemType, &b.WarehouseID, &b.MaterialID, &b.ProductID, &b.ProductSizeID, &b.QtyOnHand, &b.AvgUnitCost)
	if errors.Is(err, pgx.ErrNoRows) {
		_, err = q.Exec(ctx, `
			INSERT INTO inventory_balances (item_type, warehouse_id, material_id, product_id, product_size_id, qty_on_hand, avg_unit_cost)
			VALUES ($1,$2,$3,$4,$5,0,0)
		`, itemType, warehouseID, materialID, productID, productSizeID)
		if err != nil {
			return Balance{}, err
		}
		return Balance{ItemType: itemType, WarehouseID: warehouseID, MaterialID: materialID, ProductID: productID, ProductSizeID: productSizeID, QtyOnHand: decimal.Zero, AvgUnitCost: decimal.Zero}, nil
	}
	return b, err
}

func (r *Repository) SaveBalance(ctx context.Context, b Balance) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		UPDATE inventory_balances SET qty_on_hand = $6, avg_unit_cost = $7, updated_at = now()
		WHERE item_type = $1 AND warehouse_id = $2
		  AND material_id IS NOT DISTINCT FROM $3
		  AND product_id IS NOT DISTINCT FROM $4
		  AND product_size_id IS NOT DISTINCT FROM $5
	`, b.ItemType, b.WarehouseID, b.MaterialID, b.ProductID, b.ProductSizeID, b.QtyOnHand, b.AvgUnitCost)
	return err
}

func (r *Repository) ListTransactions(ctx context.Context, materialID, productID, warehouseID *uuid.UUID, limit, offset int) ([]Transaction, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT `+txnColumns+`
		FROM inventory_transactions
		WHERE ($1::uuid IS NULL OR material_id = $1)
		  AND ($2::uuid IS NULL OR product_id = $2)
		  AND ($3::uuid IS NULL OR warehouse_id = $3)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`, materialID, productID, warehouseID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Transaction{}
	for rows.Next() {
		t, err := scanTxn(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repository) ListBalances(ctx context.Context) ([]Balance, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT item_type, warehouse_id, material_id, product_id, product_size_id, qty_on_hand, avg_unit_cost
		FROM inventory_balances
		ORDER BY item_type
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Balance{}
	for rows.Next() {
		var b Balance
		if err := rows.Scan(&b.ItemType, &b.WarehouseID, &b.MaterialID, &b.ProductID, &b.ProductSizeID, &b.QtyOnHand, &b.AvgUnitCost); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// ListBalancesByWarehouse powers the stock-by-warehouse breakdown: every
// balance row, optionally filtered to one warehouse.
func (r *Repository) ListBalancesByWarehouse(ctx context.Context, warehouseID *uuid.UUID) ([]Balance, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT item_type, warehouse_id, material_id, product_id, product_size_id, qty_on_hand, avg_unit_cost
		FROM inventory_balances
		WHERE ($1::uuid IS NULL OR warehouse_id = $1)
		ORDER BY warehouse_id, item_type
	`, warehouseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Balance{}
	for rows.Next() {
		var b Balance
		if err := rows.Scan(&b.ItemType, &b.WarehouseID, &b.MaterialID, &b.ProductID, &b.ProductSizeID, &b.QtyOnHand, &b.AvgUnitCost); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *Repository) GetBalance(ctx context.Context, itemType ItemType, warehouseID uuid.UUID, materialID, productID, productSizeID *uuid.UUID) (Balance, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		SELECT item_type, warehouse_id, material_id, product_id, product_size_id, qty_on_hand, avg_unit_cost
		FROM inventory_balances
		WHERE item_type = $1 AND warehouse_id = $2
		  AND material_id IS NOT DISTINCT FROM $3
		  AND product_id IS NOT DISTINCT FROM $4
		  AND product_size_id IS NOT DISTINCT FROM $5
	`, itemType, warehouseID, materialID, productID, productSizeID)
	var b Balance
	err := row.Scan(&b.ItemType, &b.WarehouseID, &b.MaterialID, &b.ProductID, &b.ProductSizeID, &b.QtyOnHand, &b.AvgUnitCost)
	return b, err
}
