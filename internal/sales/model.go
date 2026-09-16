package sales

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	OrderDraft     OrderStatus = "DRAFT"
	OrderConfirmed OrderStatus = "CONFIRMED"
	OrderCancelled OrderStatus = "CANCELLED"
	OrderClosed    OrderStatus = "CLOSED"
)

type PaymentStatus string

const (
	PaymentUnpaid   PaymentStatus = "UNPAID"
	PaymentPartial  PaymentStatus = "PARTIAL"
	PaymentPaid     PaymentStatus = "PAID"
	PaymentOverpaid PaymentStatus = "OVERPAID"
)

type ProductionStatus string

const (
	ProductionNotStarted ProductionStatus = "NOT_STARTED"
	ProductionInProgress ProductionStatus = "IN_PROGRESS"
	ProductionCompleted  ProductionStatus = "COMPLETED"
	ProductionCancelled  ProductionStatus = "CANCELLED"
)

type DeliveryStatus string

const (
	DeliveryNotDelivered DeliveryStatus = "NOT_DELIVERED"
	DeliveryPartial      DeliveryStatus = "PARTIAL"
	DeliveryDelivered    DeliveryStatus = "DELIVERED"
)

type InvoiceStatus string

const (
	InvoiceDraft          InvoiceStatus = "DRAFT"
	InvoicePosted         InvoiceStatus = "POSTED"
	InvoicePartiallyPaid  InvoiceStatus = "PARTIALLY_PAID"
	InvoicePaid           InvoiceStatus = "PAID"
	InvoiceVoid           InvoiceStatus = "VOID"
)

type SalesOrder struct {
	ID                uuid.UUID        `json:"id"`
	SONumber          string           `json:"so_number"`
	CustomerID        uuid.UUID        `json:"customer_id"`
	OrderDate         time.Time        `json:"order_date"`
	OrderStatus       OrderStatus      `json:"order_status"`
	PaymentStatus     PaymentStatus    `json:"payment_status"`
	ProductionStatus  ProductionStatus `json:"production_status"`
	DeliveryStatus    DeliveryStatus   `json:"delivery_status"`
	Subtotal          decimal.Decimal  `json:"subtotal"`
	DiscountTotal     decimal.Decimal  `json:"discount_total"`
	TaxTotal          decimal.Decimal  `json:"tax_total"`
	GrandTotal        decimal.Decimal  `json:"grand_total"`
	Notes             string           `json:"notes"`
	CreatedBy         string           `json:"created_by"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	Items             []SalesOrderItem `json:"items,omitempty"`
}

type SalesOrderItem struct {
	ID              uuid.UUID       `json:"id"`
	SalesOrderID    uuid.UUID       `json:"sales_order_id"`
	ProductID       uuid.UUID       `json:"product_id"`
	ProductSizeID   uuid.UUID       `json:"product_size_id"`
	Qty             decimal.Decimal `json:"qty"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	Discount        decimal.Decimal `json:"discount"`
	TaxRate         decimal.Decimal `json:"tax_rate"`
	LineTotal       decimal.Decimal `json:"line_total"`
}

type Invoice struct {
	ID             uuid.UUID       `json:"id"`
	InvoiceNumber  string          `json:"invoice_number"`
	SalesOrderID   uuid.UUID       `json:"sales_order_id"`
	CustomerID     uuid.UUID       `json:"customer_id"`
	InvoiceDate    time.Time       `json:"invoice_date"`
	DueDate        *time.Time      `json:"due_date,omitempty"`
	Status         InvoiceStatus   `json:"status"`
	Subtotal       decimal.Decimal `json:"subtotal"`
	DiscountTotal  decimal.Decimal `json:"discount_total"`
	TaxTotal       decimal.Decimal `json:"tax_total"`
	GrandTotal     decimal.Decimal `json:"grand_total"`
	PaidAmount     decimal.Decimal `json:"paid_amount"`
	BalanceDue     decimal.Decimal `json:"balance_due"`
	JournalID      *uuid.UUID      `json:"journal_id,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	Items          []InvoiceItem   `json:"items,omitempty"`
}

type InvoiceItem struct {
	ID               uuid.UUID       `json:"id"`
	InvoiceID        uuid.UUID       `json:"invoice_id"`
	SalesOrderItemID uuid.UUID       `json:"sales_order_item_id"`
	ProductID        uuid.UUID       `json:"product_id"`
	ProductSizeID    uuid.UUID       `json:"product_size_id"`
	Qty              decimal.Decimal `json:"qty"`
	UnitPrice        decimal.Decimal `json:"unit_price"`
	Discount         decimal.Decimal `json:"discount"`
	TaxRate          decimal.Decimal `json:"tax_rate"`
	LineTotal        decimal.Decimal `json:"line_total"`
}

type CreateOrderItemInput struct {
	ProductID     uuid.UUID       `json:"product_id"`
	ProductSizeID uuid.UUID       `json:"product_size_id"`
	Qty           decimal.Decimal `json:"qty"`
	UnitPrice     decimal.Decimal `json:"unit_price"`
	Discount      decimal.Decimal `json:"discount"`
	TaxRate       decimal.Decimal `json:"tax_rate"`
}

type CreateOrderInput struct {
	CustomerID uuid.UUID              `json:"customer_id"`
	OrderDate  string                 `json:"order_date,omitempty"` // YYYY-MM-DD, defaults to today
	Notes      string                 `json:"notes"`
	Items      []CreateOrderItemInput `json:"items"`
}
