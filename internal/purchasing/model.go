package purchasing

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BillStatus string

const (
	BillDraft         BillStatus = "DRAFT"
	BillPosted        BillStatus = "POSTED"
	BillPartiallyPaid BillStatus = "PARTIALLY_PAID"
	BillPaid          BillStatus = "PAID"
	BillVoid          BillStatus = "VOID"
)

type SupplierInvoice struct {
	ID               uuid.UUID             `json:"id"`
	BillNumber       string                `json:"bill_number"`
	SupplierID       uuid.UUID             `json:"supplier_id"`
	PurchaseOrderID  *uuid.UUID            `json:"purchase_order_id,omitempty"`
	BillDate         time.Time             `json:"bill_date"`
	DueDate          *time.Time            `json:"due_date,omitempty"`
	DebitAccountCode string                `json:"debit_account_code"`
	Status           BillStatus            `json:"status"`
	Subtotal         decimal.Decimal       `json:"subtotal"`
	TaxTotal         decimal.Decimal       `json:"tax_total"`
	GrandTotal       decimal.Decimal       `json:"grand_total"`
	PaidAmount       decimal.Decimal       `json:"paid_amount"`
	BalanceDue       decimal.Decimal       `json:"balance_due"`
	JournalID        *uuid.UUID            `json:"journal_id,omitempty"`
	Notes            string                `json:"notes"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
	Items            []SupplierInvoiceItem `json:"items,omitempty"`
}

type SupplierInvoiceItem struct {
	ID                 uuid.UUID       `json:"id"`
	SupplierInvoiceID  uuid.UUID       `json:"supplier_invoice_id"`
	MaterialID         uuid.UUID       `json:"material_id"`
	Qty                decimal.Decimal `json:"qty"`
	UnitCost           decimal.Decimal `json:"unit_cost"`
	LineTotal          decimal.Decimal `json:"line_total"`
	PurchaseOrderItemID *uuid.UUID     `json:"purchase_order_item_id,omitempty"`
	GoodsReceiptItemID  *uuid.UUID     `json:"goods_receipt_item_id,omitempty"`
	PriceVariance       decimal.Decimal `json:"price_variance"`
}

type CreateBillItemInput struct {
	MaterialID uuid.UUID       `json:"material_id"`
	Qty        decimal.Decimal `json:"qty"`
	UnitCost   decimal.Decimal `json:"unit_cost"`
}

type CreateBillInput struct {
	SupplierID       uuid.UUID             `json:"supplier_id"`
	BillDate         string                `json:"bill_date,omitempty"` // YYYY-MM-DD, defaults to today
	DueDate          string                `json:"due_date,omitempty"`
	DebitAccountCode string                `json:"debit_account_code"` // e.g. 1-1200 Persediaan Bahan Baku
	TaxRate          decimal.Decimal       `json:"tax_rate"`
	Notes            string                `json:"notes"`
	Items            []CreateBillItemInput `json:"items"`
}

// ---- Purchase Orders / 3-way matching ----

type POStatus string

const (
	POStatusDraft             POStatus = "DRAFT"
	POStatusApproved          POStatus = "APPROVED"
	POStatusPartiallyReceived POStatus = "PARTIALLY_RECEIVED"
	POStatusFullyReceived     POStatus = "FULLY_RECEIVED"
	POStatusClosed            POStatus = "CLOSED"
	POStatusCancelled         POStatus = "CANCELLED"
)

type POBillingStatus string

const (
	POBillingUnbilled        POBillingStatus = "UNBILLED"
	POBillingPartiallyBilled POBillingStatus = "PARTIALLY_BILLED"
	POBillingFullyBilled     POBillingStatus = "FULLY_BILLED"
)

type PurchaseOrder struct {
	ID             uuid.UUID           `json:"id"`
	PONumber       string              `json:"po_number"`
	SupplierID     uuid.UUID           `json:"supplier_id"`
	OrderDate      time.Time           `json:"order_date"`
	ExpectedDate   *time.Time          `json:"expected_date,omitempty"`
	Status         POStatus            `json:"status"`
	BillingStatus  POBillingStatus     `json:"billing_status"`
	Subtotal       decimal.Decimal     `json:"subtotal"`
	GrandTotal     decimal.Decimal     `json:"grand_total"`
	Notes          string              `json:"notes"`
	ApprovedAt     *time.Time          `json:"approved_at,omitempty"`
	CancelledAt    *time.Time          `json:"cancelled_at,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	Items          []PurchaseOrderItem `json:"items,omitempty"`
}

type PurchaseOrderItem struct {
	ID              uuid.UUID       `json:"id"`
	PurchaseOrderID uuid.UUID       `json:"purchase_order_id"`
	MaterialID      uuid.UUID       `json:"material_id"`
	Qty             decimal.Decimal `json:"qty"`
	UnitCost        decimal.Decimal `json:"unit_cost"`
	QtyReceived     decimal.Decimal `json:"qty_received"`
	QtyBilled       decimal.Decimal `json:"qty_billed"`
	LineTotal       decimal.Decimal `json:"line_total"`
}

type CreatePOItemInput struct {
	MaterialID uuid.UUID       `json:"material_id"`
	Qty        decimal.Decimal `json:"qty"`
	UnitCost   decimal.Decimal `json:"unit_cost"`
}

type CreatePOInput struct {
	SupplierID   uuid.UUID           `json:"supplier_id"`
	OrderDate    string              `json:"order_date,omitempty"`
	ExpectedDate string              `json:"expected_date,omitempty"`
	Notes        string              `json:"notes"`
	Items        []CreatePOItemInput `json:"items"`
}

type GoodsReceipt struct {
	ID              uuid.UUID          `json:"id"`
	GRNNumber       string             `json:"grn_number"`
	PurchaseOrderID uuid.UUID          `json:"purchase_order_id"`
	SupplierID      uuid.UUID          `json:"supplier_id"`
	ReceiptDate     time.Time          `json:"receipt_date"`
	JournalID       *uuid.UUID         `json:"journal_id,omitempty"`
	Notes           string             `json:"notes"`
	CreatedAt       time.Time          `json:"created_at"`
	Items           []GoodsReceiptItem `json:"items,omitempty"`
}

type GoodsReceiptItem struct {
	ID                   uuid.UUID       `json:"id"`
	GoodsReceiptID       uuid.UUID       `json:"goods_receipt_id"`
	PurchaseOrderItemID  uuid.UUID       `json:"purchase_order_item_id"`
	MaterialID           uuid.UUID       `json:"material_id"`
	QtyReceived          decimal.Decimal `json:"qty_received"`
	UnitCost             decimal.Decimal `json:"unit_cost"`
	QtyBilled            decimal.Decimal `json:"qty_billed"`
	LineTotal            decimal.Decimal `json:"line_total"`
}

type CreateGRNItemInput struct {
	PurchaseOrderItemID uuid.UUID       `json:"purchase_order_item_id"`
	QtyReceived         decimal.Decimal `json:"qty_received"`
}

type CreateGRNInput struct {
	PurchaseOrderID uuid.UUID            `json:"purchase_order_id"`
	ReceiptDate     string               `json:"receipt_date,omitempty"`
	Notes           string               `json:"notes"`
	Items           []CreateGRNItemInput `json:"items"`
}

// BillFromGRNItemInput is one matched line: qty and the supplier's actual
// billed unit cost, matched against a specific (already-received) GRN line.
type BillFromGRNItemInput struct {
	GoodsReceiptItemID uuid.UUID       `json:"goods_receipt_item_id"`
	Qty                decimal.Decimal `json:"qty"`
	BillUnitCost       decimal.Decimal `json:"bill_unit_cost"`
}

type CreateBillFromGRNInput struct {
	PurchaseOrderID uuid.UUID               `json:"purchase_order_id"`
	BillDate        string                  `json:"bill_date,omitempty"`
	DueDate         string                  `json:"due_date,omitempty"`
	Notes           string                  `json:"notes"`
	Items           []BillFromGRNItemInput  `json:"items"`
}
