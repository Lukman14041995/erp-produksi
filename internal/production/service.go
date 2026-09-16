package production

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/accounting"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/audit"
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/ranji/clothing-erp/internal/inventory"
	"github.com/ranji/clothing-erp/internal/numbering"
	"github.com/ranji/clothing-erp/internal/product"
	"github.com/ranji/clothing-erp/internal/sales"
	"github.com/shopspring/decimal"
)

const (
	acctWIP           = "1-1210" // Persediaan Barang Dalam Proses
	acctRawMaterial   = "1-1200" // Persediaan Bahan Baku
	acctFinishedGoods = "1-1220" // Persediaan Barang Jadi
	acctBank          = "1-1002" // Bank -- labor/overhead are capitalized as paid in cash
)

type Service struct {
	pool      *pgxpool.Pool
	repo      *Repository
	productRepo *product.Repository
	accSvc    *accounting.Service
	invSvc    *inventory.Service
	salesSvc  *sales.Service
}

func NewService(pool *pgxpool.Pool, repo *Repository, productRepo *product.Repository, accSvc *accounting.Service, invSvc *inventory.Service, salesSvc *sales.Service) *Service {
	return &Service{pool: pool, repo: repo, productRepo: productRepo, accSvc: accSvc, invSvc: invSvc, salesSvc: salesSvc}
}

// ---- BOM ----

func (s *Service) CreateBOM(ctx context.Context, in CreateBOMInput) (BOMHeader, error) {
	if in.ProductID == uuid.Nil || in.Name == "" || len(in.Lines) == 0 {
		return BOMHeader{}, apperr.Validation("product_id, name, and at least one line are required")
	}
	var out BOMHeader
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		header, err := s.repo.InsertBOMHeader(ctx, in.ProductID, in.Name)
		if err != nil {
			return apperr.Internal("create bom header", err)
		}
		for _, l := range in.Lines {
			if l.QtyPerUnit.LessThanOrEqual(decimal.Zero) {
				return apperr.Validation("bom line qty_per_unit must be greater than zero")
			}
			line, err := s.repo.InsertBOMLine(ctx, BOMLine{BOMID: header.ID, MaterialID: l.MaterialID, QtyPerUnit: l.QtyPerUnit, UOM: l.UOM})
			if err != nil {
				return apperr.Internal("create bom line", err)
			}
			header.Lines = append(header.Lines, line)
		}
		if err := audit.Log(ctx, tx, "bom_headers", header.ID, audit.Insert, nil, header); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = header
		return nil
	})
	return out, err
}

func (s *Service) GetBOM(ctx context.Context, id uuid.UUID) (BOMHeader, error) {
	h, err := s.repo.GetBOMHeader(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return BOMHeader{}, apperr.NotFound("bom not found")
	}
	if err != nil {
		return BOMHeader{}, err
	}
	lines, err := s.repo.ListBOMLines(ctx, id)
	if err != nil {
		return BOMHeader{}, apperr.Internal("load bom lines", err)
	}
	h.Lines = lines
	return h, nil
}

func (s *Service) ListBOMsByProduct(ctx context.Context, productID uuid.UUID) ([]BOMHeader, error) {
	return s.repo.ListBOMsByProduct(ctx, productID)
}

// ---- Production Orders ----

// CreateOrder plans required raw material quantities from the BOM, scaling
// each line by the size's cost multiplier (S:0.9, M:1.0, L:1.1, XL:1.2, ...)
// so a size breakdown of the same product consumes proportionally different
// material -- this is the "BOM calculation with size multiplication".
func (s *Service) CreateOrder(ctx context.Context, in CreateOrderInput) (Order, error) {
	if in.ProductID == uuid.Nil {
		return Order{}, apperr.Validation("product_id is required")
	}
	if len(in.Items) == 0 {
		return Order{}, apperr.Validation("production order must have at least one size item")
	}

	plannedQty := decimal.Zero
	for _, it := range in.Items {
		if it.PlannedQty.LessThanOrEqual(decimal.Zero) {
			return Order{}, apperr.Validation("planned_qty must be greater than zero")
		}
		plannedQty = plannedQty.Add(it.PlannedQty)
	}

	var out Order
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		number, err := numbering.Generate(ctx, tx, numbering.ProductionOrder, time.Now())
		if err != nil {
			return apperr.Internal("generate production order number", err)
		}

		order, err := s.repo.InsertOrder(ctx, Order{
			ProdNumber: number, SalesOrderID: in.SalesOrderID, ProductID: in.ProductID, BOMID: in.BOMID,
			PlannedQty: plannedQty, Notes: in.Notes,
		})
		if err != nil {
			return apperr.Internal("create production order", err)
		}

		sizeMultipliers := map[uuid.UUID]decimal.Decimal{}
		for _, it := range in.Items {
			size, err := s.productRepo.GetSizeByID(ctx, it.ProductSizeID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return apperr.Validation("unknown product_size_id")
				}
				return apperr.Internal("load product size", err)
			}
			sizeMultipliers[it.ProductSizeID] = size.SizeMultiplier

			item, err := s.repo.InsertOrderItem(ctx, OrderItem{ProductionOrderID: order.ID, ProductSizeID: it.ProductSizeID, PlannedQty: it.PlannedQty})
			if err != nil {
				return apperr.Internal("create production order item", err)
			}
			order.Items = append(order.Items, item)
		}

		if in.BOMID != nil {
			bomLines, err := s.repo.ListBOMLines(ctx, *in.BOMID)
			if err != nil {
				return apperr.Internal("load bom lines", err)
			}

			materialTotals := map[uuid.UUID]decimal.Decimal{}
			for _, it := range in.Items {
				mult := sizeMultipliers[it.ProductSizeID]
				for _, bl := range bomLines {
					required := bl.QtyPerUnit.Mul(it.PlannedQty).Mul(mult)
					materialTotals[bl.MaterialID] = materialTotals[bl.MaterialID].Add(required)
				}
			}
			for materialID, qty := range materialTotals {
				m, err := s.repo.InsertPlannedMaterial(ctx, Material{ProductionOrderID: order.ID, MaterialID: materialID, PlannedQty: qty})
				if err != nil {
					return apperr.Internal("plan production material", err)
				}
				order.Materials = append(order.Materials, m)
			}
		}

		if err := audit.Log(ctx, tx, "production_orders", order.ID, audit.Insert, nil, order); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = order
		return nil
	})
	return out, err
}

func (s *Service) StartOrder(ctx context.Context, id uuid.UUID) (Order, error) {
	var out Order
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		order, err := s.repo.GetOrderForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("production order not found")
			}
			return apperr.Internal("load production order", err)
		}
		if order.Status != NotStarted {
			return apperr.Conflict("only NOT_STARTED production orders can be started")
		}

		if err := s.repo.SetStatus(ctx, id, InProgress); err != nil {
			return apperr.Internal("start production order", err)
		}
		if err := s.repo.SetStartDate(ctx, id); err != nil {
			return apperr.Internal("stamp start date", err)
		}
		if order.SalesOrderID != nil {
			if err := s.salesSvc.UpdateProductionStatus(ctx, *order.SalesOrderID, sales.ProductionInProgress); err != nil {
				return apperr.Wrapf(err, "sync sales order production status")
			}
		}
		if err := audit.Log(ctx, tx, "production_orders", id, audit.Update, order.Status, InProgress); err != nil {
			return apperr.Internal("write audit log", err)
		}
		order.Status = InProgress
		out = order
		return nil
	})
	return out, err
}

func (s *Service) CancelOrder(ctx context.Context, id uuid.UUID) (Order, error) {
	var out Order
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		order, err := s.repo.GetOrderForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("production order not found")
			}
			return apperr.Internal("load production order", err)
		}
		if order.Status == Completed || order.Status == Cancelled {
			return apperr.Conflict("cannot cancel a production order that is " + string(order.Status))
		}

		if err := s.repo.SetStatus(ctx, id, Cancelled); err != nil {
			return apperr.Internal("cancel production order", err)
		}
		if order.SalesOrderID != nil {
			if err := s.salesSvc.UpdateProductionStatus(ctx, *order.SalesOrderID, sales.ProductionCancelled); err != nil {
				return apperr.Wrapf(err, "sync sales order production status")
			}
		}
		if err := audit.Log(ctx, tx, "production_orders", id, audit.Update, order.Status, Cancelled); err != nil {
			return apperr.Internal("write audit log", err)
		}
		order.Status = Cancelled
		out = order
		return nil
	})
	return out, err
}

// IssueMaterial pulls qty of a planned material out of raw-material
// inventory at its current moving-average cost, then immediately posts
// DR WIP / CR Persediaan Bahan Baku for that cost -- the "Material Issue"
// accounting law.
func (s *Service) IssueMaterial(ctx context.Context, orderID, materialRowID uuid.UUID, qty decimal.Decimal) (Material, error) {
	if qty.LessThanOrEqual(decimal.Zero) {
		return Material{}, apperr.Validation("qty must be greater than zero")
	}

	var out Material
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		order, err := s.repo.GetOrderForUpdate(ctx, orderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("production order not found")
			}
			return apperr.Internal("load production order", err)
		}
		if order.Status != InProgress {
			return apperr.Conflict("materials can only be issued while production is IN_PROGRESS")
		}

		row, err := s.repo.GetMaterialRowForUpdate(ctx, materialRowID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("production material row not found")
			}
			return apperr.Internal("load production material row", err)
		}
		if row.ProductionOrderID != orderID {
			return apperr.Validation("material row does not belong to this production order")
		}

		matWarehouse, err := s.invSvc.DefaultMaterialWarehouse(ctx)
		if err != nil {
			return apperr.Wrapf(err, "resolve material warehouse")
		}

		materialID := row.MaterialID
		movement, err := s.invSvc.RecordMovement(ctx, inventory.MovementInput{
			TxnType: inventory.TxnProductionIssue, ItemType: inventory.ItemMaterial, WarehouseID: matWarehouse.ID,
			MaterialID: &materialID, QtyOut: qty, RefType: "PRODUCTION_ORDER", RefID: &orderID, TxnDate: time.Now(),
			Notes: "Material issue for " + order.ProdNumber,
		})
		if err != nil {
			return apperr.Wrapf(err, "issue material from inventory")
		}

		newIssuedQty := row.IssuedQty.Add(qty)
		newTotalCost := row.TotalCost.Add(movement.Transaction.TotalCost)
		if err := s.repo.UpdateMaterialIssue(ctx, materialRowID, newIssuedQty, movement.Transaction.UnitCost, newTotalCost); err != nil {
			return apperr.Internal("update production material", err)
		}

		if movement.Transaction.TotalCost.IsPositive() {
			_, err = s.accSvc.PostJournal(ctx, accounting.SourceMaterialIssue, &orderID, time.Now(),
				"Material issue for "+order.ProdNumber,
				[]accounting.JournalLineInput{
					{AccountCode: acctWIP, Debit: movement.Transaction.TotalCost, Description: "WIP " + order.ProdNumber},
					{AccountCode: acctRawMaterial, Credit: movement.Transaction.TotalCost, Description: "Material issued " + order.ProdNumber},
				})
			if err != nil {
				return apperr.Wrapf(err, "post material issue journal")
			}
		}

		if err := audit.Log(ctx, tx, "production_materials", materialRowID, audit.Update, row, map[string]any{"issued_qty": newIssuedQty, "total_cost": newTotalCost}); err != nil {
			return apperr.Internal("write audit log", err)
		}

		row.IssuedQty = newIssuedQty
		row.UnitCost = movement.Transaction.UnitCost
		row.TotalCost = newTotalCost
		out = row
		return nil
	})
	return out, err
}

// AddLabor capitalizes direct labor cost into WIP (DR WIP / CR Beban Gaji),
// rather than letting it hit the P&L directly as a period expense.
func (s *Service) AddLabor(ctx context.Context, orderID uuid.UUID, in AddLaborInput) (Labor, error) {
	if in.Hours.LessThanOrEqual(decimal.Zero) || in.Rate.LessThan(decimal.Zero) {
		return Labor{}, apperr.Validation("hours must be > 0 and rate must be >= 0")
	}
	totalCost := in.Hours.Mul(in.Rate)

	var out Labor
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		order, err := s.repo.GetOrderForUpdate(ctx, orderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("production order not found")
			}
			return apperr.Internal("load production order", err)
		}
		if order.Status != InProgress {
			return apperr.Conflict("labor can only be recorded while production is IN_PROGRESS")
		}

		labor, err := s.repo.InsertLabor(ctx, Labor{ProductionOrderID: orderID, Description: in.Description, Hours: in.Hours, Rate: in.Rate, TotalCost: totalCost})
		if err != nil {
			return apperr.Internal("record labor", err)
		}

		if totalCost.IsPositive() {
			_, err = s.accSvc.PostJournal(ctx, accounting.SourceMaterialIssue, &orderID, time.Now(),
				"Direct labor for "+order.ProdNumber+": "+in.Description,
				[]accounting.JournalLineInput{
					{AccountCode: acctWIP, Debit: totalCost, Description: "WIP labor " + order.ProdNumber},
					{AccountCode: acctBank, Credit: totalCost, Description: "Labor paid (cash) " + order.ProdNumber},
				})
			if err != nil {
				return apperr.Wrapf(err, "post labor journal")
			}
		}

		if err := audit.Log(ctx, tx, "production_labor", labor.ID, audit.Insert, nil, labor); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = labor
		return nil
	})
	return out, err
}

// AddOverhead capitalizes manufacturing overhead into WIP the same way AddLabor does.
func (s *Service) AddOverhead(ctx context.Context, orderID uuid.UUID, in AddOverheadInput) (Overhead, error) {
	if in.Amount.LessThanOrEqual(decimal.Zero) {
		return Overhead{}, apperr.Validation("amount must be greater than zero")
	}

	var out Overhead
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		order, err := s.repo.GetOrderForUpdate(ctx, orderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("production order not found")
			}
			return apperr.Internal("load production order", err)
		}
		if order.Status != InProgress {
			return apperr.Conflict("overhead can only be recorded while production is IN_PROGRESS")
		}

		overhead, err := s.repo.InsertOverhead(ctx, Overhead{ProductionOrderID: orderID, Description: in.Description, AllocationBasis: in.AllocationBasis, Amount: in.Amount})
		if err != nil {
			return apperr.Internal("record overhead", err)
		}

		_, err = s.accSvc.PostJournal(ctx, accounting.SourceOverheadCost, &orderID, time.Now(),
			"Manufacturing overhead for "+order.ProdNumber+": "+in.Description,
			[]accounting.JournalLineInput{
				{AccountCode: acctWIP, Debit: in.Amount, Description: "WIP overhead " + order.ProdNumber},
				{AccountCode: acctBank, Credit: in.Amount, Description: "Overhead paid (cash) " + order.ProdNumber},
			})
		if err != nil {
			return apperr.Wrapf(err, "post overhead journal")
		}

		if err := audit.Log(ctx, tx, "production_overheads", overhead.ID, audit.Insert, nil, overhead); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = overhead
		return nil
	})
	return out, err
}

// CompleteOrder is the HPP snapshot engine: unit_cost = (material + labor +
// overhead) / finished_qty. It posts DR Finished Goods / CR WIP for the
// accumulated cost and receives the finished units into inventory at that
// unit cost, all inside one transaction.
func (s *Service) CompleteOrder(ctx context.Context, orderID uuid.UUID, in CompleteOrderInput) (Cost, error) {
	if len(in.Items) == 0 {
		return Cost{}, apperr.Validation("completion must report finished qty per size")
	}

	var out Cost
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		order, err := s.repo.GetOrderForUpdate(ctx, orderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("production order not found")
			}
			return apperr.Internal("load production order", err)
		}
		if order.Status != InProgress {
			return apperr.Conflict("only IN_PROGRESS production orders can be completed")
		}

		orderItems, err := s.repo.ListOrderItems(ctx, orderID)
		if err != nil {
			return apperr.Internal("load production order items", err)
		}
		itemBySize := map[uuid.UUID]OrderItem{}
		for _, it := range orderItems {
			itemBySize[it.ProductSizeID] = it
		}

		finishedQty := decimal.Zero
		for _, ci := range in.Items {
			item, ok := itemBySize[ci.ProductSizeID]
			if !ok {
				return apperr.Validation("size is not part of this production order")
			}
			if ci.FinishedQty.LessThan(decimal.Zero) || ci.FinishedQty.GreaterThan(item.PlannedQty) {
				return apperr.Validation("finished_qty must be between 0 and the planned qty for the size")
			}
			finishedQty = finishedQty.Add(ci.FinishedQty)
		}
		if !finishedQty.IsPositive() {
			return apperr.Validation("total finished qty must be greater than zero")
		}

		materials, err := s.repo.ListMaterials(ctx, orderID)
		if err != nil {
			return apperr.Internal("load production materials", err)
		}
		totalMaterialCost := decimal.Zero
		for _, m := range materials {
			totalMaterialCost = totalMaterialCost.Add(m.TotalCost)
		}

		labor, err := s.repo.ListLabor(ctx, orderID)
		if err != nil {
			return apperr.Internal("load production labor", err)
		}
		totalLaborCost := decimal.Zero
		for _, l := range labor {
			totalLaborCost = totalLaborCost.Add(l.TotalCost)
		}

		overheads, err := s.repo.ListOverheads(ctx, orderID)
		if err != nil {
			return apperr.Internal("load production overheads", err)
		}
		totalOverheadCost := decimal.Zero
		for _, o := range overheads {
			totalOverheadCost = totalOverheadCost.Add(o.Amount)
		}

		totalCost := totalMaterialCost.Add(totalLaborCost).Add(totalOverheadCost)
		unitCost := decimal.Zero
		if finishedQty.IsPositive() {
			unitCost = totalCost.DivRound(finishedQty, 4)
		}

		var journalID *uuid.UUID
		if totalCost.IsPositive() {
			journal, err := s.accSvc.PostJournal(ctx, accounting.SourceProductionCompletion, &orderID, time.Now(),
				"Production completion for "+order.ProdNumber,
				[]accounting.JournalLineInput{
					{AccountCode: acctFinishedGoods, Debit: totalCost, Description: "FG received " + order.ProdNumber},
					{AccountCode: acctWIP, Credit: totalCost, Description: "WIP relieved " + order.ProdNumber},
				})
			if err != nil {
				return apperr.Wrapf(err, "post production completion journal")
			}
			journalID = &journal.ID
		}

		fgWarehouse, err := s.invSvc.DefaultFinishedGoodsWarehouse(ctx)
		if err != nil {
			return apperr.Wrapf(err, "resolve finished goods warehouse")
		}

		productID := order.ProductID
		for _, ci := range in.Items {
			if !ci.FinishedQty.IsPositive() {
				continue
			}
			sizeID := ci.ProductSizeID
			if _, err := s.invSvc.RecordMovement(ctx, inventory.MovementInput{
				TxnType: inventory.TxnProductionReceipt, ItemType: inventory.ItemProduct, WarehouseID: fgWarehouse.ID,
				ProductID: &productID, ProductSizeID: &sizeID, QtyIn: ci.FinishedQty, UnitCost: unitCost,
				RefType: "PRODUCTION_ORDER", RefID: &orderID, TxnDate: time.Now(),
				Notes: "Finished goods receipt for " + order.ProdNumber,
			}); err != nil {
				return apperr.Wrapf(err, "receive finished goods into inventory")
			}
			if err := s.repo.SetItemFinishedQty(ctx, itemBySize[ci.ProductSizeID].ID, ci.FinishedQty); err != nil {
				return apperr.Internal("update finished qty for size", err)
			}
		}

		if err := s.repo.SetEndDateAndFinishedQty(ctx, orderID, finishedQty); err != nil {
			return apperr.Internal("update production order finished qty", err)
		}
		if err := s.repo.SetStatus(ctx, orderID, Completed); err != nil {
			return apperr.Internal("mark production order completed", err)
		}

		cost, err := s.repo.InsertCost(ctx, Cost{
			ProductionOrderID: orderID, TotalMaterialCost: totalMaterialCost, TotalLaborCost: totalLaborCost,
			TotalOverheadCost: totalOverheadCost, TotalCost: totalCost, FinishedQty: finishedQty, UnitCost: unitCost, JournalID: journalID,
		})
		if err != nil {
			return apperr.Internal("save HPP snapshot", err)
		}

		if order.SalesOrderID != nil {
			if err := s.salesSvc.UpdateProductionStatus(ctx, *order.SalesOrderID, sales.ProductionCompleted); err != nil {
				return apperr.Wrapf(err, "sync sales order production status")
			}
		}

		if err := audit.Log(ctx, tx, "production_costs", cost.ID, audit.Insert, nil, cost); err != nil {
			return apperr.Internal("write audit log", err)
		}

		out = cost
		return nil
	})
	return out, err
}

func (s *Service) GetOrder(ctx context.Context, id uuid.UUID) (Order, error) {
	order, err := s.repo.GetOrderByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, apperr.NotFound("production order not found")
	}
	if err != nil {
		return Order{}, err
	}
	if order.Items, err = s.repo.ListOrderItems(ctx, id); err != nil {
		return Order{}, apperr.Internal("load order items", err)
	}
	if order.Materials, err = s.repo.ListMaterials(ctx, id); err != nil {
		return Order{}, apperr.Internal("load order materials", err)
	}
	if order.Labor, err = s.repo.ListLabor(ctx, id); err != nil {
		return Order{}, apperr.Internal("load order labor", err)
	}
	if order.Overheads, err = s.repo.ListOverheads(ctx, id); err != nil {
		return Order{}, apperr.Internal("load order overheads", err)
	}
	return order, nil
}

func (s *Service) ListOrders(ctx context.Context, limit, offset int) ([]Order, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListOrders(ctx, limit, offset)
}

func (s *Service) GetCost(ctx context.Context, orderID uuid.UUID) (Cost, error) {
	c, err := s.repo.GetCostByOrder(ctx, orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cost{}, apperr.NotFound("HPP snapshot not found for this production order")
	}
	return c, err
}
