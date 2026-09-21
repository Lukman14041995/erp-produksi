package spk

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/accounting"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/audit"
	"github.com/ranji/clothing-erp/internal/catalog"
	"github.com/ranji/clothing-erp/internal/customer"
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/ranji/clothing-erp/internal/inventory"
	"github.com/ranji/clothing-erp/internal/material"
	"github.com/ranji/clothing-erp/internal/numbering"
	"github.com/ranji/clothing-erp/internal/sales"
	"github.com/shopspring/decimal"
)

// Same WIP/Persediaan Bahan Baku/Bank GL accounts internal/production uses
// for material issue, labor, and overhead -- these are the same accounting
// events whether they came from a BOM-costed production order or an SPK.
const (
	acctWIP         = "1-1210"
	acctRawMaterial = "1-1200"
	acctBank        = "1-1002" // labor/overhead are capitalized as paid in cash, same simplification internal/production makes
)

type Service struct {
	pool        *pgxpool.Pool
	repo        *Repository
	catalogSvc  *catalog.Service
	customerSvc *customer.Service
	salesSvc    *sales.Service
	materialSvc *material.Service
	invSvc      *inventory.Service
	accSvc      *accounting.Service
}

func NewService(pool *pgxpool.Pool, repo *Repository, catalogSvc *catalog.Service, customerSvc *customer.Service, salesSvc *sales.Service, materialSvc *material.Service, invSvc *inventory.Service, accSvc *accounting.Service) *Service {
	return &Service{pool: pool, repo: repo, catalogSvc: catalogSvc, customerSvc: customerSvc, salesSvc: salesSvc, materialSvc: materialSvc, invSvc: invSvc, accSvc: accSvc}
}

// GenerateForOrder splits a just-confirmed sales order's items into one SPK
// per distinct product type -- a customer can mix Jersey and T-Shirt lines
// in one order through the configurator, but each product type follows its
// own production sequence (see StageFlow), so it gets its own SPK. Called
// once, right after quotation.Service.ConfirmQuotation confirms the order.
// A no-op if items is empty.
func (s *Service) GenerateForOrder(ctx context.Context, order sales.SalesOrder, quotationID uuid.UUID, quotationNumber string, items []GenerateItem) ([]Order, error) {
	if len(items) == 0 {
		return nil, nil
	}

	cust, err := s.customerSvc.Get(ctx, order.CustomerID)
	if err != nil {
		return nil, apperr.Wrapf(err, "load customer for SPK")
	}

	grouped := map[uuid.UUID][]GenerateItem{}
	typeOrder := []uuid.UUID{}
	for _, it := range items {
		if _, ok := grouped[it.ProductTypeID]; !ok {
			typeOrder = append(typeOrder, it.ProductTypeID)
		}
		grouped[it.ProductTypeID] = append(grouped[it.ProductTypeID], it)
	}

	var out []Order
	err = db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		for _, productTypeID := range typeOrder {
			group := grouped[productTypeID]
			pt, err := s.catalogSvc.GetProductType(ctx, productTypeID)
			if err != nil {
				return apperr.Wrapf(err, "load product type for SPK")
			}

			totalQty := decimal.Zero
			for _, it := range group {
				totalQty = totalQty.Add(it.Qty)
			}

			number, err := numbering.Generate(ctx, tx, numbering.SPKOrder, time.Now())
			if err != nil {
				return apperr.Internal("generate SPK number", err)
			}

			spkOrder, err := s.repo.InsertOrder(ctx, Order{
				SPKNumber: number, SalesOrderID: order.ID, SONumber: order.SONumber,
				QuotationID: &quotationID, QuotationNumber: quotationNumber, CustomerName: cust.Name,
				ProductTypeID: productTypeID, ProductTypeCode: pt.Code, ProductTypeName: pt.Name,
				CompositionSummary: compositionSummary(group),
				TotalQty: totalQty, Status: NotStarted, CurrentStageCode: StageDesign, CurrentStageName: StageFlow(pt.Code)[0].Name,
			})
			if err != nil {
				return apperr.Internal("create SPK order", err)
			}

			for _, it := range group {
				item, err := s.repo.InsertItem(ctx, Item{
					SPKOrderID: spkOrder.ID, SalesOrderItemID: it.SalesOrderItemID,
					ProductID: it.ProductID, ProductSizeID: it.ProductSizeID, Qty: it.Qty,
					FabricName: it.FabricName, VariantName: it.VariantName, InkName: it.InkName, SizeCode: it.SizeCode,
					FabricMaterialID: it.FabricMaterialID, InkMaterialID: it.InkMaterialID,
				})
				if err != nil {
					return apperr.Internal("create SPK item", err)
				}
				spkOrder.Items = append(spkOrder.Items, item)
			}

			for i, def := range StageFlow(pt.Code) {
				planned := decimal.Zero
				if def.RequiresQty {
					planned = totalQty
				}
				stage, err := s.repo.InsertStage(ctx, Stage{
					SPKOrderID: spkOrder.ID, StageSeq: i + 1, StageCode: def.Code, StageName: def.Name,
					RequiresQty: def.RequiresQty, PlannedQty: planned, Status: StagePending,
				})
				if err != nil {
					return apperr.Internal("create SPK stage", err)
				}
				spkOrder.Stages = append(spkOrder.Stages, stage)
			}

			if err := audit.Log(ctx, tx, "spk_orders", spkOrder.ID, audit.Insert, nil, spkOrder); err != nil {
				return apperr.Internal("write audit log", err)
			}
			out = append(out, spkOrder)
		}
		return nil
	})
	return out, err
}

// compositionSummary condenses a group of order lines into one label for
// the list view: the shared bahan/model/tinta combo when every line uses
// the same one, otherwise the first line's combo plus a "+N lainnya" count.
func compositionSummary(items []GenerateItem) string {
	if len(items) == 0 {
		return ""
	}
	label := func(it GenerateItem) string {
		parts := make([]string, 0, 3)
		for _, p := range []string{it.FabricName, it.VariantName, it.InkName} {
			if p != "" {
				parts = append(parts, p)
			}
		}
		return strings.Join(parts, " · ")
	}

	seen := map[string]bool{}
	distinct := 0
	first := label(items[0])
	for _, it := range items {
		l := label(it)
		if !seen[l] {
			seen[l] = true
			distinct++
		}
	}
	if distinct <= 1 {
		return first
	}
	return fmt.Sprintf("%s +%d lainnya", first, distinct-1)
}

// LogStageProgress records today's completed qty for one qty-tracked stage
// (e.g. "20 pcs selesai cutting hari ini"). Stages are independent of each
// other -- any stage can be logged at any time regardless of the others'
// progress, since the production floor works stages in overlapping batches
// rather than strictly one-at-a-time.
func (s *Service) LogStageProgress(ctx context.Context, stageID, userID uuid.UUID, in LogStageInput) (Stage, error) {
	if !in.Qty.IsPositive() {
		return Stage{}, apperr.Validation("qty must be greater than zero")
	}

	var out Stage
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		stage, err := s.repo.GetStageForUpdate(ctx, stageID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("spk stage not found")
			}
			return apperr.Internal("load spk stage", err)
		}
		if !stage.RequiresQty {
			return apperr.Conflict("this stage does not track qty; use the milestone action instead")
		}
		if stage.Status == StageDone {
			return apperr.Conflict("stage is already completed")
		}

		remaining := stage.PlannedQty.Sub(stage.CompletedQty)
		if in.Qty.GreaterThan(remaining) {
			return apperr.Validation("qty exceeds the remaining amount for this stage")
		}

		newCompleted := stage.CompletedQty.Add(in.Qty)
		newStatus := StageInProgress
		var completedAt *time.Time
		if newCompleted.GreaterThanOrEqual(stage.PlannedQty) {
			newStatus = StageDone
			now := time.Now()
			completedAt = &now
		}

		if _, err := s.repo.InsertStageLog(ctx, StageLog{SPKStageID: stageID, Qty: in.Qty, Note: in.Note, LoggedBy: &userID}); err != nil {
			return apperr.Internal("record spk stage log", err)
		}
		if err := s.repo.UpdateStageProgress(ctx, stageID, newCompleted, newStatus, completedAt); err != nil {
			return apperr.Internal("update spk stage", err)
		}
		if stage.Status == StagePending {
			if err := s.repo.SetStageStarted(ctx, stageID); err != nil {
				return apperr.Internal("stamp stage started_at", err)
			}
		}

		if err := s.recomputeOrderProgress(ctx, stage.SPKOrderID); err != nil {
			return err
		}

		stage.CompletedQty = newCompleted
		stage.Status = newStatus
		stage.CompletedAt = completedAt
		if err := audit.Log(ctx, tx, "spk_stages", stageID, audit.Update, nil, stage); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = stage
		return nil
	})
	return out, err
}

// MarkMilestone flips a non-qty stage (Design, Selesai, Pengiriman) to
// DONE.
func (s *Service) MarkMilestone(ctx context.Context, stageID, userID uuid.UUID, in MarkMilestoneInput) (Stage, error) {
	var out Stage
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		stage, err := s.repo.GetStageForUpdate(ctx, stageID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("spk stage not found")
			}
			return apperr.Internal("load spk stage", err)
		}
		if stage.RequiresQty {
			return apperr.Conflict("this stage tracks qty; use the qty-log action instead")
		}
		if stage.Status == StageDone {
			return apperr.Conflict("stage is already completed")
		}

		now := time.Now()
		if err := s.repo.UpdateStageProgress(ctx, stageID, stage.PlannedQty, StageDone, &now); err != nil {
			return apperr.Internal("update spk stage", err)
		}
		if stage.Status == StagePending {
			if err := s.repo.SetStageStarted(ctx, stageID); err != nil {
				return apperr.Internal("stamp stage started_at", err)
			}
		}
		if _, err := s.repo.InsertStageLog(ctx, StageLog{SPKStageID: stageID, Qty: decimal.Zero, Note: in.Note, LoggedBy: &userID}); err != nil {
			return apperr.Internal("record spk stage log", err)
		}

		if err := s.recomputeOrderProgress(ctx, stage.SPKOrderID); err != nil {
			return err
		}

		// The "Selesai" milestone marks this SPK's production finished. Sync
		// the sales order's production_status the same way the legacy costing
		// production module does, but only once every SPK for that sales
		// order is COMPLETED -- a mixed Jersey+T-Shirt order shouldn't look
		// "production complete" while one half is still on the floor.
		if stage.StageCode == StageFinish {
			spkOrder, err := s.repo.GetOrderByID(ctx, stage.SPKOrderID)
			if err != nil {
				return apperr.Internal("load spk order", err)
			}
			allDone, err := s.repo.AllOrdersCompletedForSalesOrder(ctx, spkOrder.SalesOrderID)
			if err != nil {
				return apperr.Internal("check sibling spk orders", err)
			}
			if allDone {
				if err := s.salesSvc.UpdateProductionStatus(ctx, spkOrder.SalesOrderID, sales.ProductionCompleted); err != nil {
					return apperr.Wrapf(err, "sync sales order production status")
				}
			}
		}

		stage.Status = StageDone
		stage.CompletedAt = &now
		if err := audit.Log(ctx, tx, "spk_stages", stageID, audit.Update, nil, stage); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = stage
		return nil
	})
	return out, err
}

// IssueMaterial records real material consumption against an SPK -- "kain 2
// roll", "tinta 500ml", "plastik packing 50 pcs" -- optionally tied to the
// stage that consumed it. Unlike stage progress, this moves real inventory:
// it deducts stock from the default material warehouse at that warehouse's
// current moving-average cost and posts DR WIP / CR Bahan Baku, the same
// accounting law internal/production.IssueMaterial uses for the BOM/HPP
// costing flow -- both are "material issued into work-in-process".
func (s *Service) IssueMaterial(ctx context.Context, spkOrderID, userID uuid.UUID, in IssueMaterialInput) (MaterialUsage, error) {
	if in.MaterialID == uuid.Nil {
		return MaterialUsage{}, apperr.Validation("material_id is required")
	}
	if !in.Qty.IsPositive() {
		return MaterialUsage{}, apperr.Validation("qty must be greater than zero")
	}

	var out MaterialUsage
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		spkOrder, err := s.repo.GetOrderByID(ctx, spkOrderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("spk order not found")
			}
			return apperr.Internal("load spk order", err)
		}

		if in.SPKStageID != nil {
			stage, err := s.repo.GetStageForUpdate(ctx, *in.SPKStageID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return apperr.NotFound("spk stage not found")
				}
				return apperr.Internal("load spk stage", err)
			}
			if stage.SPKOrderID != spkOrderID {
				return apperr.Validation("stage does not belong to this spk order")
			}
		}

		mat, err := s.materialSvc.Get(ctx, in.MaterialID)
		if err != nil {
			if ae, ok := apperr.As(err); ok && ae.Kind == apperr.KindNotFound {
				return apperr.Validation("unknown material_id")
			}
			return apperr.Internal("load material", err)
		}

		warehouse, err := s.invSvc.DefaultMaterialWarehouse(ctx)
		if err != nil {
			return apperr.Wrapf(err, "resolve material warehouse")
		}

		materialID := in.MaterialID
		movement, err := s.invSvc.RecordMovement(ctx, inventory.MovementInput{
			TxnType: inventory.TxnProductionIssue, ItemType: inventory.ItemMaterial, WarehouseID: warehouse.ID,
			MaterialID: &materialID, QtyOut: in.Qty, RefType: "SPK_ORDER", RefID: &spkOrderID, TxnDate: time.Now(),
			Notes: "Material issue for " + spkOrder.SPKNumber,
		})
		if err != nil {
			return apperr.Wrapf(err, "issue material from inventory")
		}

		var journalID *uuid.UUID
		if movement.Transaction.TotalCost.IsPositive() {
			journal, err := s.accSvc.PostJournal(ctx, accounting.SourceMaterialIssue, &spkOrderID, time.Now(),
				"Material issue for "+spkOrder.SPKNumber+": "+mat.Name,
				[]accounting.JournalLineInput{
					{AccountCode: acctWIP, Debit: movement.Transaction.TotalCost, Description: "WIP " + spkOrder.SPKNumber},
					{AccountCode: acctRawMaterial, Credit: movement.Transaction.TotalCost, Description: "Material issued " + spkOrder.SPKNumber},
				})
			if err != nil {
				return apperr.Wrapf(err, "post material issue journal")
			}
			journalID = &journal.ID
		}

		usage, err := s.repo.InsertMaterialUsage(ctx, MaterialUsage{
			SPKOrderID: spkOrderID, SPKStageID: in.SPKStageID, MaterialID: in.MaterialID,
			Qty: in.Qty, UnitCost: movement.Transaction.UnitCost, TotalCost: movement.Transaction.TotalCost,
			WarehouseID: warehouse.ID, InventoryTxnID: &movement.Transaction.ID, JournalID: journalID,
			Note: in.Note, LoggedBy: &userID,
		})
		if err != nil {
			return apperr.Internal("record material usage", err)
		}
		usage.MaterialName = mat.Name
		usage.MaterialUOM = mat.UOM

		if err := s.repo.IncrementMaterialCost(ctx, spkOrderID, usage.TotalCost); err != nil {
			return apperr.Internal("update spk material cost total", err)
		}

		if err := audit.Log(ctx, tx, "spk_material_usages", usage.ID, audit.Insert, nil, usage); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = usage
		return nil
	})
	return out, err
}

// AddLabor records direct labor time against an SPK -- "5 jam jahit" --
// and capitalizes it into WIP (DR WIP / CR Bank), completing the HPP
// alongside material and overhead. See package doc on MaterialCostTotal.
func (s *Service) AddLabor(ctx context.Context, spkOrderID, userID uuid.UUID, in AddLaborInput) (Labor, error) {
	if in.Hours.LessThanOrEqual(decimal.Zero) {
		return Labor{}, apperr.Validation("hours must be greater than zero")
	}
	if in.Rate.IsNegative() {
		return Labor{}, apperr.Validation("rate cannot be negative")
	}
	totalCost := in.Hours.Mul(in.Rate)

	var out Labor
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		spkOrder, err := s.repo.GetOrderByID(ctx, spkOrderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("spk order not found")
			}
			return apperr.Internal("load spk order", err)
		}

		if in.SPKStageID != nil {
			stage, err := s.repo.GetStageForUpdate(ctx, *in.SPKStageID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return apperr.NotFound("spk stage not found")
				}
				return apperr.Internal("load spk stage", err)
			}
			if stage.SPKOrderID != spkOrderID {
				return apperr.Validation("stage does not belong to this spk order")
			}
		}

		var journalID *uuid.UUID
		if totalCost.IsPositive() {
			journal, err := s.accSvc.PostJournal(ctx, accounting.SourceLaborCost, &spkOrderID, time.Now(),
				"Direct labor for "+spkOrder.SPKNumber+": "+in.Description,
				[]accounting.JournalLineInput{
					{AccountCode: acctWIP, Debit: totalCost, Description: "WIP labor " + spkOrder.SPKNumber},
					{AccountCode: acctBank, Credit: totalCost, Description: "Labor paid (cash) " + spkOrder.SPKNumber},
				})
			if err != nil {
				return apperr.Wrapf(err, "post labor journal")
			}
			journalID = &journal.ID
		}

		labor, err := s.repo.InsertLabor(ctx, Labor{
			SPKOrderID: spkOrderID, SPKStageID: in.SPKStageID, Description: in.Description,
			Hours: in.Hours, Rate: in.Rate, TotalCost: totalCost, JournalID: journalID, LoggedBy: &userID,
		})
		if err != nil {
			return apperr.Internal("record spk labor", err)
		}

		if err := s.repo.IncrementLaborCost(ctx, spkOrderID, totalCost); err != nil {
			return apperr.Internal("update spk labor cost total", err)
		}
		if err := audit.Log(ctx, tx, "spk_labor", labor.ID, audit.Insert, nil, labor); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = labor
		return nil
	})
	return out, err
}

// AddOverhead records manufacturing overhead against an SPK, capitalized
// into WIP the same way AddLabor does.
func (s *Service) AddOverhead(ctx context.Context, spkOrderID, userID uuid.UUID, in AddOverheadInput) (Overhead, error) {
	if in.Amount.LessThanOrEqual(decimal.Zero) {
		return Overhead{}, apperr.Validation("amount must be greater than zero")
	}

	var out Overhead
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		spkOrder, err := s.repo.GetOrderByID(ctx, spkOrderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("spk order not found")
			}
			return apperr.Internal("load spk order", err)
		}

		journal, err := s.accSvc.PostJournal(ctx, accounting.SourceOverheadCost, &spkOrderID, time.Now(),
			"Manufacturing overhead for "+spkOrder.SPKNumber+": "+in.Description,
			[]accounting.JournalLineInput{
				{AccountCode: acctWIP, Debit: in.Amount, Description: "WIP overhead " + spkOrder.SPKNumber},
				{AccountCode: acctBank, Credit: in.Amount, Description: "Overhead paid (cash) " + spkOrder.SPKNumber},
			})
		if err != nil {
			return apperr.Wrapf(err, "post overhead journal")
		}

		overhead, err := s.repo.InsertOverhead(ctx, Overhead{
			SPKOrderID: spkOrderID, Description: in.Description, AllocationBasis: in.AllocationBasis,
			Amount: in.Amount, JournalID: &journal.ID, LoggedBy: &userID,
		})
		if err != nil {
			return apperr.Internal("record spk overhead", err)
		}

		if err := s.repo.IncrementOverheadCost(ctx, spkOrderID, in.Amount); err != nil {
			return apperr.Internal("update spk overhead cost total", err)
		}
		if err := audit.Log(ctx, tx, "spk_overheads", overhead.ID, audit.Insert, nil, overhead); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = overhead
		return nil
	})
	return out, err
}

// recomputeOrderProgress recalculates an SPK's overall progress % and
// status from its stages' current state and caches both on spk_orders so
// list views don't need to aggregate stages per row. Every stage counts as
// an equal 1/N share of the order (a milestone contributes 0% or 100% of
// its share; a qty-tracked stage contributes completed_qty/planned_qty of
// its share) since stages can overlap rather than gating each other.
func (s *Service) recomputeOrderProgress(ctx context.Context, orderID uuid.UUID) error {
	stages, err := s.repo.ListStages(ctx, orderID)
	if err != nil {
		return apperr.Internal("load spk stages", err)
	}
	if len(stages) == 0 {
		return nil
	}

	share := decimal.NewFromInt(1).DivRound(decimal.NewFromInt(int64(len(stages))), 8)
	total := decimal.Zero
	allDone := true
	anyStarted := false
	currentStageCode := stages[0].StageCode
	currentStageName := stages[0].StageName
	var designDone, productionDone, shipped bool
	for _, st := range stages {
		switch {
		case st.Status == StageDone:
			total = total.Add(share)
		case st.RequiresQty && st.PlannedQty.IsPositive():
			total = total.Add(share.Mul(st.CompletedQty.Div(st.PlannedQty)))
		}
		if st.Status != StageDone {
			allDone = false
		}
		if st.Status != StagePending {
			currentStageCode = st.StageCode
			currentStageName = st.StageName
			anyStarted = true
		}

		// Cache the three milestones the production board buckets on --
		// stages don't gate each other, so "furthest stage touched" above
		// isn't enough to tell whether these specific ones are DONE.
		switch st.StageCode {
		case StageDesign:
			designDone = st.Status == StageDone
		case StageFinish:
			productionDone = st.Status == StageDone
		case StageShipping:
			shipped = st.Status == StageDone
		}
	}

	status := NotStarted
	switch {
	case allDone:
		status = Completed
	case anyStarted:
		status = InProgress
	}

	pct := total.Mul(decimal.NewFromInt(100)).Round(2)
	if pct.GreaterThan(decimal.NewFromInt(100)) {
		pct = decimal.NewFromInt(100)
	}
	return s.repo.UpdateOrderProgress(ctx, orderID, status, currentStageCode, currentStageName, pct, designDone, productionDone, shipped)
}

func (s *Service) GetOrder(ctx context.Context, id uuid.UUID) (Order, error) {
	order, err := s.repo.GetOrderByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, apperr.NotFound("spk order not found")
	}
	if err != nil {
		return Order{}, err
	}
	if order.Items, err = s.repo.ListItems(ctx, id); err != nil {
		return Order{}, apperr.Internal("load spk items", err)
	}
	stages, err := s.repo.ListStages(ctx, id)
	if err != nil {
		return Order{}, apperr.Internal("load spk stages", err)
	}
	for i := range stages {
		logs, err := s.repo.ListStageLogs(ctx, stages[i].ID)
		if err != nil {
			return Order{}, apperr.Internal("load spk stage logs", err)
		}
		stages[i].Logs = logs
	}
	order.Stages = stages
	usages, err := s.repo.ListMaterialUsages(ctx, id)
	if err != nil {
		return Order{}, apperr.Internal("load spk material usages", err)
	}
	order.MaterialUsages = usages
	labor, err := s.repo.ListLabor(ctx, id)
	if err != nil {
		return Order{}, apperr.Internal("load spk labor", err)
	}
	order.Labor = labor
	overheads, err := s.repo.ListOverheads(ctx, id)
	if err != nil {
		return Order{}, apperr.Internal("load spk overheads", err)
	}
	order.Overheads = overheads
	return order, nil
}

func (s *Service) ListOrders(ctx context.Context, productTypeCode string, limit, offset int) ([]Order, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	return s.repo.ListOrders(ctx, productTypeCode, limit, offset)
}

// ListAllOrders backs the cross-jenis production board (/production/costing).
func (s *Service) ListAllOrders(ctx context.Context, limit, offset int) ([]Order, error) {
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	return s.repo.ListAllOrders(ctx, limit, offset)
}

// ListBySalesOrder is used by the Sales Order detail page (staff) and the
// customer's public order-detail view, both of which want the full stage
// breakdown (not just the cached current_stage/progress_pct) to render a
// stepper -- unlike ListOrders, which backs the /production/orders lists
// and stays light on purpose.
func (s *Service) ListBySalesOrder(ctx context.Context, salesOrderID uuid.UUID) ([]Order, error) {
	orders, err := s.repo.ListBySalesOrder(ctx, salesOrderID)
	if err != nil {
		return nil, err
	}
	for i := range orders {
		stages, err := s.repo.ListStages(ctx, orders[i].ID)
		if err != nil {
			return nil, apperr.Internal("load spk stages", err)
		}
		orders[i].Stages = stages
	}
	return orders, nil
}
