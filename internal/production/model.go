package production

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
	Cancelled  Status = "CANCELLED"
)

type BOMHeader struct {
	ID        uuid.UUID  `json:"id"`
	ProductID uuid.UUID  `json:"product_id"`
	Name      string     `json:"name"`
	Version   int        `json:"version"`
	IsActive  bool       `json:"is_active"`
	Lines     []BOMLine  `json:"lines,omitempty"`
}

type BOMLine struct {
	ID         uuid.UUID       `json:"id"`
	BOMID      uuid.UUID       `json:"bom_id"`
	MaterialID uuid.UUID       `json:"material_id"`
	QtyPerUnit decimal.Decimal `json:"qty_per_unit"`
	UOM        string          `json:"uom"`
}

type CreateBOMLineInput struct {
	MaterialID uuid.UUID       `json:"material_id"`
	QtyPerUnit decimal.Decimal `json:"qty_per_unit"`
	UOM        string          `json:"uom"`
}

type CreateBOMInput struct {
	ProductID uuid.UUID            `json:"product_id"`
	Name      string               `json:"name"`
	Lines     []CreateBOMLineInput `json:"lines"`
}

type Order struct {
	ID             uuid.UUID   `json:"id"`
	ProdNumber     string      `json:"prod_number"`
	SalesOrderID   *uuid.UUID  `json:"sales_order_id,omitempty"`
	ProductID      uuid.UUID   `json:"product_id"`
	BOMID          *uuid.UUID  `json:"bom_id,omitempty"`
	PlannedQty     decimal.Decimal `json:"planned_qty"`
	FinishedQty    decimal.Decimal `json:"finished_qty"`
	Status         Status      `json:"production_status"`
	StartDate      *time.Time  `json:"start_date,omitempty"`
	EndDate        *time.Time  `json:"end_date,omitempty"`
	Notes          string      `json:"notes"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	Items          []OrderItem `json:"items,omitempty"`
	Materials      []Material  `json:"materials,omitempty"`
	Labor          []Labor     `json:"labor,omitempty"`
	Overheads      []Overhead  `json:"overheads,omitempty"`
}

type OrderItem struct {
	ID              uuid.UUID       `json:"id"`
	ProductionOrderID uuid.UUID     `json:"production_order_id"`
	ProductSizeID   uuid.UUID       `json:"product_size_id"`
	PlannedQty      decimal.Decimal `json:"planned_qty"`
	FinishedQty     decimal.Decimal `json:"finished_qty"`
}

type Material struct {
	ID                uuid.UUID       `json:"id"`
	ProductionOrderID uuid.UUID       `json:"production_order_id"`
	MaterialID        uuid.UUID       `json:"material_id"`
	PlannedQty        decimal.Decimal `json:"planned_qty"`
	IssuedQty         decimal.Decimal `json:"issued_qty"`
	UnitCost          decimal.Decimal `json:"unit_cost"`
	TotalCost         decimal.Decimal `json:"total_cost"`
	IssuedAt          *time.Time      `json:"issued_at,omitempty"`
}

type Labor struct {
	ID                uuid.UUID       `json:"id"`
	ProductionOrderID uuid.UUID       `json:"production_order_id"`
	Description       string          `json:"description"`
	Hours             decimal.Decimal `json:"hours"`
	Rate              decimal.Decimal `json:"rate"`
	TotalCost         decimal.Decimal `json:"total_cost"`
}

type Overhead struct {
	ID                uuid.UUID       `json:"id"`
	ProductionOrderID uuid.UUID       `json:"production_order_id"`
	Description       string          `json:"description"`
	AllocationBasis   string          `json:"allocation_basis"`
	Amount            decimal.Decimal `json:"amount"`
}

type Cost struct {
	ID                 uuid.UUID       `json:"id"`
	ProductionOrderID  uuid.UUID       `json:"production_order_id"`
	TotalMaterialCost  decimal.Decimal `json:"total_material_cost"`
	TotalLaborCost     decimal.Decimal `json:"total_labor_cost"`
	TotalOverheadCost  decimal.Decimal `json:"total_overhead_cost"`
	TotalCost          decimal.Decimal `json:"total_cost"`
	FinishedQty        decimal.Decimal `json:"finished_qty"`
	UnitCost           decimal.Decimal `json:"unit_cost"`
	JournalID          *uuid.UUID      `json:"journal_id,omitempty"`
	ComputedAt         time.Time       `json:"computed_at"`
}

type CreateOrderItemInput struct {
	ProductSizeID uuid.UUID       `json:"product_size_id"`
	PlannedQty    decimal.Decimal `json:"planned_qty"`
}

type CreateOrderInput struct {
	SalesOrderID *uuid.UUID              `json:"sales_order_id,omitempty"`
	ProductID    uuid.UUID               `json:"product_id"`
	BOMID        *uuid.UUID              `json:"bom_id,omitempty"`
	Notes        string                  `json:"notes"`
	Items        []CreateOrderItemInput  `json:"items"`
}

type IssueMaterialInput struct {
	Qty decimal.Decimal `json:"qty"`
}

type AddLaborInput struct {
	Description string          `json:"description"`
	Hours       decimal.Decimal `json:"hours"`
	Rate        decimal.Decimal `json:"rate"`
}

type AddOverheadInput struct {
	Description     string          `json:"description"`
	AllocationBasis string          `json:"allocation_basis"`
	Amount          decimal.Decimal `json:"amount"`
}

type CompleteItemInput struct {
	ProductSizeID uuid.UUID       `json:"product_size_id"`
	FinishedQty   decimal.Decimal `json:"finished_qty"`
}

type CompleteOrderInput struct {
	Items []CompleteItemInput `json:"items"`
}
