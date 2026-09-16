package inventory

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TxnType string

const (
	TxnPurchase          TxnType = "PURCHASE"
	TxnProductionIssue   TxnType = "PRODUCTION_ISSUE"
	TxnProductionReceipt TxnType = "PRODUCTION_RECEIPT"
	TxnSale              TxnType = "SALE"
	TxnAdjustment        TxnType = "ADJUSTMENT"
	TxnTransferOut       TxnType = "TRANSFER_OUT"
	TxnTransferIn        TxnType = "TRANSFER_IN"
)

type ItemType string

const (
	ItemMaterial ItemType = "MATERIAL"
	ItemProduct  ItemType = "PRODUCT"
)

type Warehouse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpsertWarehouseInput struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// Well-known warehouse codes used as defaults by domains that don't yet
// expose an explicit warehouse picker in their own UI (e.g. a sales
// delivery always ships finished goods from WH-FG). Stock transfers and
// stock opname always require an explicit warehouse from the caller.
const (
	DefaultMaterialWarehouseCode = "WH-MAT"
	DefaultFinishedGoodsWarehouseCode = "WH-FG"
)

type Transaction struct {
	ID            uuid.UUID       `json:"id"`
	TxnType       TxnType         `json:"txn_type"`
	ItemType      ItemType        `json:"item_type"`
	WarehouseID   uuid.UUID       `json:"warehouse_id"`
	MaterialID    *uuid.UUID      `json:"material_id,omitempty"`
	ProductID     *uuid.UUID      `json:"product_id,omitempty"`
	ProductSizeID *uuid.UUID      `json:"product_size_id,omitempty"`
	QtyIn         decimal.Decimal `json:"qty_in"`
	QtyOut        decimal.Decimal `json:"qty_out"`
	UnitCost      decimal.Decimal `json:"unit_cost"`
	TotalCost     decimal.Decimal `json:"total_cost"`
	RefType       string          `json:"ref_type"`
	RefID         *uuid.UUID      `json:"ref_id,omitempty"`
	TxnDate       time.Time       `json:"txn_date"`
	Notes         string          `json:"notes"`
	CreatedAt     time.Time       `json:"created_at"`
}

type Balance struct {
	ItemType      ItemType        `json:"item_type"`
	WarehouseID   uuid.UUID       `json:"warehouse_id"`
	MaterialID    *uuid.UUID      `json:"material_id,omitempty"`
	ProductID     *uuid.UUID      `json:"product_id,omitempty"`
	ProductSizeID *uuid.UUID      `json:"product_size_id,omitempty"`
	QtyOnHand     decimal.Decimal `json:"qty_on_hand"`
	AvgUnitCost   decimal.Decimal `json:"avg_unit_cost"`
}

// MovementInput describes one stock movement. Exactly one of MaterialID or
// (ProductID [+ ProductSizeID]) must be set, matching ItemType. WarehouseID
// is always required -- every movement is bound to a specific location.
type MovementInput struct {
	TxnType       TxnType
	ItemType      ItemType
	WarehouseID   uuid.UUID
	MaterialID    *uuid.UUID
	ProductID     *uuid.UUID
	ProductSizeID *uuid.UUID
	QtyIn         decimal.Decimal
	QtyOut        decimal.Decimal
	UnitCost      decimal.Decimal // required for QtyIn movements; ignored (looked up) for QtyOut movements
	RefType       string
	RefID         *uuid.UUID
	TxnDate       time.Time
	Notes         string
}

// MovementResult reports the transaction as recorded, including the
// effective unit cost actually applied (== moving average for issues).
type MovementResult struct {
	Transaction Transaction
	Balance     Balance
}

// ---- Stock transfers ----

type TransferStatus string

const (
	TransferDraft      TransferStatus = "DRAFT"
	TransferDispatched TransferStatus = "DISPATCHED"
	TransferReceived   TransferStatus = "RECEIVED"
	TransferCancelled  TransferStatus = "CANCELLED"
)

type StockTransfer struct {
	ID                     uuid.UUID          `json:"id"`
	TransferNumber         string             `json:"transfer_number"`
	SourceWarehouseID      uuid.UUID          `json:"source_warehouse_id"`
	DestinationWarehouseID uuid.UUID          `json:"destination_warehouse_id"`
	Status                 TransferStatus     `json:"status"`
	TransferDate           time.Time          `json:"transfer_date"`
	Notes                  string             `json:"notes"`
	DispatchedAt           *time.Time         `json:"dispatched_at,omitempty"`
	ReceivedAt             *time.Time         `json:"received_at,omitempty"`
	CreatedAt              time.Time          `json:"created_at"`
	UpdatedAt              time.Time          `json:"updated_at"`
	Items                  []StockTransferItem `json:"items,omitempty"`
}

type StockTransferItem struct {
	ID              uuid.UUID       `json:"id"`
	StockTransferID uuid.UUID       `json:"stock_transfer_id"`
	ItemType        ItemType        `json:"item_type"`
	MaterialID      *uuid.UUID      `json:"material_id,omitempty"`
	ProductID       *uuid.UUID      `json:"product_id,omitempty"`
	ProductSizeID   *uuid.UUID      `json:"product_size_id,omitempty"`
	Qty             decimal.Decimal `json:"qty"`
	UnitCost        decimal.Decimal `json:"unit_cost"`
}

type CreateTransferItemInput struct {
	ItemType      ItemType   `json:"item_type"`
	MaterialID    *uuid.UUID `json:"material_id,omitempty"`
	ProductID     *uuid.UUID `json:"product_id,omitempty"`
	ProductSizeID *uuid.UUID `json:"product_size_id,omitempty"`
	Qty           decimal.Decimal `json:"qty"`
}

type CreateTransferInput struct {
	SourceWarehouseID      uuid.UUID                  `json:"source_warehouse_id"`
	DestinationWarehouseID uuid.UUID                  `json:"destination_warehouse_id"`
	TransferDate           string                     `json:"transfer_date,omitempty"`
	Notes                  string                     `json:"notes"`
	Items                  []CreateTransferItemInput  `json:"items"`
}

// ---- Stock opname ----

type OpnameStatus string

const (
	OpnameDraft     OpnameStatus = "DRAFT"
	OpnamePosted    OpnameStatus = "POSTED"
	OpnameCancelled OpnameStatus = "CANCELLED"
)

type StockOpname struct {
	ID            uuid.UUID         `json:"id"`
	OpnameNumber  string            `json:"opname_number"`
	WarehouseID   uuid.UUID         `json:"warehouse_id"`
	OpnameDate    time.Time         `json:"opname_date"`
	Status        OpnameStatus      `json:"status"`
	Notes         string            `json:"notes"`
	JournalID     *uuid.UUID        `json:"journal_id,omitempty"`
	PostedAt      *time.Time        `json:"posted_at,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Items         []StockOpnameItem `json:"items,omitempty"`
}

type StockOpnameItem struct {
	ID              uuid.UUID       `json:"id"`
	StockOpnameID   uuid.UUID       `json:"stock_opname_id"`
	ItemType        ItemType        `json:"item_type"`
	MaterialID      *uuid.UUID      `json:"material_id,omitempty"`
	ProductID       *uuid.UUID      `json:"product_id,omitempty"`
	ProductSizeID   *uuid.UUID      `json:"product_size_id,omitempty"`
	SystemQty       decimal.Decimal `json:"system_qty"`
	ActualQty       decimal.Decimal `json:"actual_qty"`
	UnitCost        decimal.Decimal `json:"unit_cost"`
	VarianceQty     decimal.Decimal `json:"variance_qty"`
	VarianceAmount  decimal.Decimal `json:"variance_amount"`
}

// CreateOpnameItemInput identifies one counted item and its physically
// counted quantity; system_qty/unit_cost are frozen server-side from the
// current warehouse balance, never trusted from the client.
type CreateOpnameItemInput struct {
	ItemType      ItemType        `json:"item_type"`
	MaterialID    *uuid.UUID      `json:"material_id,omitempty"`
	ProductID     *uuid.UUID      `json:"product_id,omitempty"`
	ProductSizeID *uuid.UUID      `json:"product_size_id,omitempty"`
	ActualQty     decimal.Decimal `json:"actual_qty"`
}

type CreateOpnameInput struct {
	WarehouseID uuid.UUID               `json:"warehouse_id"`
	OpnameDate  string                  `json:"opname_date,omitempty"`
	Notes       string                  `json:"notes"`
	Items       []CreateOpnameItemInput `json:"items"`
}
