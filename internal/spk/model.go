// Package spk tracks the production-floor workflow for a confirmed sales
// order: which stage it's at, how many pieces are done per stage, and the
// resulting percentage a customer or sales can see. It is intentionally
// separate from internal/production (BOM/material issue/labor/HPP costing)
// -- SPK has no accounting side effects, it is pure workflow visibility.
package spk

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Status string

const (
	NotStarted Status = "NOT_STARTED"
	InProgress Status = "IN_PROGRESS"
	Completed  Status = "COMPLETED"
)

type StageStatus string

const (
	StagePending    StageStatus = "PENDING"
	StageInProgress StageStatus = "IN_PROGRESS"
	StageDone       StageStatus = "DONE"
)

// Stage codes shared between both product-type flows; only the order and
// membership differ between T-Shirt and Jersey -- see StageFlow.
const (
	StageDesign   = "DESIGN"
	StageCutting  = "CUTTING"
	StageSewing   = "SEWING"
	StagePrinting = "PRINTING"
	StageQC       = "QC"
	StagePacking  = "PACKING"
	StageFinish   = "DONE"
	StageShipping = "SHIPPING"
)

type stageDef struct {
	Code        string
	Name        string
	RequiresQty bool
}

// tshirtFlow: Design -> Cutting -> Jahit -> Sablon -> QC -> Packing ->
// Selesai -> Pengiriman.
var tshirtFlow = []stageDef{
	{StageDesign, "Design", false},
	{StageCutting, "Cutting Bahan", true},
	{StageSewing, "Jahit (Sewing)", true},
	{StagePrinting, "Cetak Design (Sablon)", true},
	{StageQC, "QC", true},
	{StagePacking, "Packing", true},
	{StageFinish, "Selesai", false},
	{StageShipping, "Pengiriman", false},
}

// jerseyFlow: Design -> Sablon -> Cutting -> Jahit -> QC -> Packing ->
// Selesai -> Pengiriman (print happens on flat fabric before it's cut/sewn).
var jerseyFlow = []stageDef{
	{StageDesign, "Design", false},
	{StagePrinting, "Cetak Design (Sablon)", true},
	{StageCutting, "Cutting Bahan", true},
	{StageSewing, "Jahit (Sewing)", true},
	{StageQC, "QC", true},
	{StagePacking, "Packing", true},
	{StageFinish, "Selesai", false},
	{StageShipping, "Pengiriman", false},
}

// StageFlow returns the ordered stage template for a product type code.
// Only "TSHIRT" and "JERSEY" are defined by the business today; any other
// code falls back to the T-Shirt flow rather than failing, since a new
// product type is meant to be addable from master data without a code
// change (see catalog.ProductType).
func StageFlow(productTypeCode string) []stageDef {
	if productTypeCode == "JERSEY" {
		return jerseyFlow
	}
	return tshirtFlow
}

type Order struct {
	ID               uuid.UUID       `json:"id"`
	SPKNumber        string          `json:"spk_number"`
	SalesOrderID     uuid.UUID       `json:"sales_order_id"`
	SONumber         string          `json:"so_number"`
	QuotationID      *uuid.UUID      `json:"quotation_id,omitempty"`
	QuotationNumber  string          `json:"quotation_number"`
	CustomerName     string          `json:"customer_name"`
	ProductTypeID    uuid.UUID       `json:"product_type_id"`
	ProductTypeCode  string          `json:"product_type_code"`
	ProductTypeName  string          `json:"product_type_name"`
	// CompositionSummary is a cached one-line summary of the order's
	// bahan/model/tinta combination (e.g. "Cotton Combed · Normal ·
	// Rubber"), for the list view -- see Item's per-line fields for the full
	// breakdown.
	CompositionSummary string          `json:"composition_summary"`
	TotalQty           decimal.Decimal `json:"total_qty"`
	Status           Status          `json:"status"`
	CurrentStageCode string          `json:"current_stage_code"`
	CurrentStageName string          `json:"current_stage_name"`
	ProgressPct      decimal.Decimal `json:"progress_pct"`
	// DesignDone/ProductionDone/Shipped mirror the DESIGN/DONE/SHIPPING
	// stages' status, cached for the cross-jenis production board's coarse
	// Design/Produksi/Selesai/Pengiriman bucketing (see
	// recomputeOrderProgress -- stages don't gate each other, so this can't
	// be derived from CurrentStageCode alone).
	DesignDone       bool            `json:"design_done"`
	ProductionDone   bool            `json:"production_done"`
	Shipped          bool            `json:"shipped"`
	// MaterialCostTotal/LaborCostTotal/OverheadCostTotal are running sums
	// maintained by IssueMaterial/AddLabor/AddOverhead; their sum is this
	// SPK's HPP (Harga Pokok Produksi) -- material + labor + overhead, the
	// same three buckets the legacy BOM/HPP module tracks, just scoped to
	// an SPK instead of a BOM production order.
	MaterialCostTotal decimal.Decimal `json:"material_cost_total"`
	LaborCostTotal    decimal.Decimal `json:"labor_cost_total"`
	OverheadCostTotal decimal.Decimal `json:"overhead_cost_total"`
	Notes            string          `json:"notes"`
	CompletedAt      *time.Time      `json:"completed_at,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	Items            []Item          `json:"items,omitempty"`
	Stages           []Stage         `json:"stages,omitempty"`
	MaterialUsages   []MaterialUsage `json:"material_usages,omitempty"`
	Labor            []Labor         `json:"labor,omitempty"`
	Overheads        []Overhead      `json:"overheads,omitempty"`
}

type Item struct {
	ID               uuid.UUID       `json:"id"`
	SPKOrderID       uuid.UUID       `json:"spk_order_id"`
	SalesOrderItemID uuid.UUID       `json:"sales_order_item_id"`
	ProductID        uuid.UUID       `json:"product_id"`
	ProductSizeID    uuid.UUID       `json:"product_size_id"`
	Qty              decimal.Decimal `json:"qty"`
	// FabricName/VariantName/InkName/SizeCode are the customer's
	// configurator choices for this line (Bahan/Model/Tinta/Ukuran),
	// denormalized at SPK-generation time so production sees exactly what
	// to cut/sew/print without following product_id back through the
	// catalog. See Service.GenerateForOrder.
	FabricName  string `json:"fabric_name"`
	VariantName string `json:"variant_name"`
	InkName     string `json:"ink_name"`
	SizeCode    string `json:"size_code"`
	// FabricMaterialID/InkMaterialID are the Materials the customer's
	// chosen fabric/ink are linked to (catalog.Fabric.MaterialID /
	// catalog.Ink.MaterialID) -- lets the "Bahan Baku Terpakai" picker
	// surface "bahan yang sudah dipilih customer" as quick picks.
	FabricMaterialID *uuid.UUID `json:"fabric_material_id,omitempty"`
	InkMaterialID    *uuid.UUID `json:"ink_material_id,omitempty"`
}

type Stage struct {
	ID           uuid.UUID       `json:"id"`
	SPKOrderID   uuid.UUID       `json:"spk_order_id"`
	StageSeq     int             `json:"stage_seq"`
	StageCode    string          `json:"stage_code"`
	StageName    string          `json:"stage_name"`
	RequiresQty  bool            `json:"requires_qty"`
	PlannedQty   decimal.Decimal `json:"planned_qty"`
	CompletedQty decimal.Decimal `json:"completed_qty"`
	Status       StageStatus     `json:"status"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	Logs         []StageLog      `json:"logs,omitempty"`
}

// MaterialUsage is real material issued against an SPK -- unlike Stage
// progress (pure workflow, no accounting effect), each row here is backed
// by one real inventory movement (stock actually decremented at moving-
// average cost) and one journal posting (DR WIP / CR Bahan Baku), mirroring
// internal/production.IssueMaterial's accounting law. See
// Service.IssueMaterial.
type MaterialUsage struct {
	ID              uuid.UUID       `json:"id"`
	SPKOrderID      uuid.UUID       `json:"spk_order_id"`
	SPKStageID      *uuid.UUID      `json:"spk_stage_id,omitempty"`
	MaterialID      uuid.UUID       `json:"material_id"`
	MaterialName    string          `json:"material_name"`
	MaterialUOM     string          `json:"material_uom"`
	Qty             decimal.Decimal `json:"qty"`
	UnitCost        decimal.Decimal `json:"unit_cost"`
	TotalCost       decimal.Decimal `json:"total_cost"`
	WarehouseID     uuid.UUID       `json:"warehouse_id"`
	InventoryTxnID  *uuid.UUID      `json:"inventory_txn_id,omitempty"`
	JournalID       *uuid.UUID      `json:"journal_id,omitempty"`
	Note            string          `json:"note"`
	LoggedBy        *uuid.UUID      `json:"logged_by,omitempty"`
	LoggedByName    string          `json:"logged_by_name,omitempty"`
	LoggedAt        time.Time       `json:"logged_at"`
}

type IssueMaterialInput struct {
	SPKStageID *uuid.UUID      `json:"spk_stage_id,omitempty"`
	MaterialID uuid.UUID       `json:"material_id"`
	Qty        decimal.Decimal `json:"qty"`
	Note       string          `json:"note"`
}

// Labor is direct labor time capitalized into an SPK's WIP, optionally
// tied to the stage it was spent on (e.g. "5 jam jahit" against SEWING).
type Labor struct {
	ID           uuid.UUID       `json:"id"`
	SPKOrderID   uuid.UUID       `json:"spk_order_id"`
	SPKStageID   *uuid.UUID      `json:"spk_stage_id,omitempty"`
	Description  string          `json:"description"`
	Hours        decimal.Decimal `json:"hours"`
	Rate         decimal.Decimal `json:"rate"`
	TotalCost    decimal.Decimal `json:"total_cost"`
	JournalID    *uuid.UUID      `json:"journal_id,omitempty"`
	LoggedBy     *uuid.UUID      `json:"logged_by,omitempty"`
	LoggedByName string          `json:"logged_by_name,omitempty"`
	LoggedAt     time.Time       `json:"logged_at"`
}

type AddLaborInput struct {
	SPKStageID  *uuid.UUID      `json:"spk_stage_id,omitempty"`
	Description string          `json:"description"`
	Hours       decimal.Decimal `json:"hours"`
	Rate        decimal.Decimal `json:"rate"`
}

// Overhead is manufacturing overhead capitalized into an SPK's WIP the
// same way Labor is.
type Overhead struct {
	ID              uuid.UUID       `json:"id"`
	SPKOrderID      uuid.UUID       `json:"spk_order_id"`
	Description     string          `json:"description"`
	AllocationBasis string          `json:"allocation_basis"`
	Amount          decimal.Decimal `json:"amount"`
	JournalID       *uuid.UUID      `json:"journal_id,omitempty"`
	LoggedBy        *uuid.UUID      `json:"logged_by,omitempty"`
	LoggedByName    string          `json:"logged_by_name,omitempty"`
	LoggedAt        time.Time       `json:"logged_at"`
}

type AddOverheadInput struct {
	Description     string          `json:"description"`
	AllocationBasis string          `json:"allocation_basis"`
	Amount          decimal.Decimal `json:"amount"`
}

type StageLog struct {
	ID           uuid.UUID       `json:"id"`
	SPKStageID   uuid.UUID       `json:"spk_stage_id"`
	Qty          decimal.Decimal `json:"qty"`
	Note         string          `json:"note"`
	LoggedBy     *uuid.UUID      `json:"logged_by,omitempty"`
	LoggedByName string          `json:"logged_by_name,omitempty"`
	LoggedAt     time.Time       `json:"logged_at"`
}

// GenerateItem is what quotation.Service passes in per sales-order-item
// when a just-confirmed order is split into per-product-type SPKs.
type GenerateItem struct {
	SalesOrderItemID uuid.UUID
	ProductTypeID    uuid.UUID
	ProductID        uuid.UUID
	ProductSizeID    uuid.UUID
	Qty              decimal.Decimal
	FabricName       string
	VariantName      string
	InkName          string
	SizeCode         string
	FabricMaterialID *uuid.UUID
	InkMaterialID    *uuid.UUID
}

type LogStageInput struct {
	Qty  decimal.Decimal `json:"qty"`
	Note string          `json:"note"`
}

type MarkMilestoneInput struct {
	Note string `json:"note"`
}
