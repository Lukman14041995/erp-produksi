package quotation

import (
	"time"

	"github.com/google/uuid"
	"github.com/ranji/clothing-erp/internal/billing"
	"github.com/ranji/clothing-erp/internal/catalog"
	"github.com/ranji/clothing-erp/internal/sales"
	"github.com/ranji/clothing-erp/internal/spk"
	"github.com/shopspring/decimal"
)

type LinkStatus string

const (
	LinkActive  LinkStatus = "ACTIVE"
	LinkRevoked LinkStatus = "REVOKED"
)

type Status string

const (
	StatusPendingReview Status = "PENDING_REVIEW"
	StatusConfirmed     Status = "CONFIRMED"
	StatusRejected      Status = "REJECTED"
	StatusCancelled     Status = "CANCELLED"
)

type OrderLink struct {
	ID         uuid.UUID  `json:"id"`
	CustomerID uuid.UUID  `json:"customer_id"`
	CreatedBy  uuid.UUID  `json:"created_by"`
	Status     LinkStatus `json:"status"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// CreateLinkInput.ExpiresAt is optional; the service defaults to 7 days out
// when omitted (see Service.CreateLink).
type CreateLinkInput struct {
	CustomerID uuid.UUID  `json:"customer_id"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// CreateLinkOutput carries the plaintext token exactly once -- it is never
// persisted or retrievable again after this response, matching how
// internal/auth hands back a refresh token.
type CreateLinkOutput struct {
	OrderLink
	Token string `json:"token"`
}

type Quotation struct {
	ID                uuid.UUID       `json:"id"`
	QuotationNumber   string          `json:"quotation_number"`
	OrderLinkID       uuid.UUID       `json:"order_link_id"`
	CustomerID        uuid.UUID       `json:"customer_id"`
	Status            Status          `json:"status"`
	EstimatedSubtotal decimal.Decimal `json:"estimated_subtotal"`
	Notes             string          `json:"notes"`
	ReviewedBy        *uuid.UUID      `json:"reviewed_by,omitempty"`
	ReviewedAt        *time.Time      `json:"reviewed_at,omitempty"`
	SalesOrderID      *uuid.UUID      `json:"sales_order_id,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	Items             []QuotationItem `json:"items,omitempty"`
}

type QuotationItem struct {
	ID            uuid.UUID       `json:"id"`
	QuotationID   uuid.UUID       `json:"quotation_id"`
	ProductTypeID uuid.UUID       `json:"product_type_id"`
	FabricID      uuid.UUID       `json:"fabric_id"`
	GarmentSizeID uuid.UUID       `json:"garment_size_id"`
	VariantID     uuid.UUID       `json:"variant_id"`
	InkID         uuid.UUID       `json:"ink_id"`
	Qty           decimal.Decimal `json:"qty"`
	UnitPrice     decimal.Decimal `json:"unit_price"`
	LineTotal     decimal.Decimal `json:"line_total"`
	CreatedAt     time.Time       `json:"created_at"`
}

// ItemInput is what the customer's browser sends for one configured line:
// Jenis (product type) + Model (variant) + Bahan (fabric) + Tinta (ink) +
// Ukuran, with Qty the only customer-controlled number that reaches
// pricing -- unit price is always computed server-side from the
// referenced catalog rows.
type ItemInput struct {
	ProductTypeID uuid.UUID       `json:"product_type_id"`
	FabricID      uuid.UUID       `json:"fabric_id"`
	GarmentSizeID uuid.UUID       `json:"garment_size_id"`
	VariantID     uuid.UUID       `json:"variant_id"`
	InkID         uuid.UUID       `json:"ink_id"`
	Qty           decimal.Decimal `json:"qty"`
}

type EstimateInput struct {
	Items []ItemInput `json:"items"`
}

type EstimateLine struct {
	ItemInput
	UnitPrice decimal.Decimal `json:"unit_price"`
	LineTotal decimal.Decimal `json:"line_total"`
}

type EstimateOutput struct {
	Lines    []EstimateLine  `json:"lines"`
	Subtotal decimal.Decimal `json:"subtotal"`
}

type SubmitQuotationInput struct {
	Notes string      `json:"notes"`
	Items []ItemInput `json:"items"`
}

// ConfirmQuotationInput.TaxRate is a fraction (e.g. 0.11), not a percentage
// -- same convention as sales.CreateOrderItemInput.TaxRate.
type ConfirmQuotationInput struct {
	TaxRate decimal.Decimal `json:"tax_rate"`
}

type RejectQuotationInput struct {
	Reason string `json:"reason"`
}

// PublicOrderDetail is everything the customer's order-detail view needs in
// one call: the quotation itself, the Sales Order and Invoice it became
// once confirmed (nil until then), and the payment plan/installments the
// customer has chosen (nil until they pick one).
type PublicOrderDetail struct {
	Quotation   Quotation            `json:"quotation"`
	SalesOrder  *sales.SalesOrder    `json:"sales_order,omitempty"`
	Invoice     *sales.Invoice       `json:"invoice,omitempty"`
	PaymentPlan *billing.PaymentPlan `json:"payment_plan,omitempty"`
	// SPKOrders is the production workflow progress for this order -- one
	// entry per product type in the order (see spk.Service.GenerateForOrder),
	// so the customer can see which stage their T-Shirt/Jersey pieces are at.
	SPKOrders []spk.Order `json:"spk_orders,omitempty"`
}

// PublicCatalog is what a customer sees behind a valid order link: every
// active product type/fabric/variant/ink/size, tagged with which product
// type each fabric/variant/ink belongs to. The frontend filters
// fabrics/variants/inks down to the selected Jenis (product type) as the
// customer steps through the wizard.
type PublicCatalog struct {
	CustomerName string                   `json:"customer_name"`
	ProductTypes []catalog.ProductType    `json:"product_types"`
	Fabrics      []catalog.Fabric         `json:"fabrics"`
	Variants     []catalog.GarmentVariant `json:"variants"`
	Inks         []catalog.Ink            `json:"inks"`
	Sizes        []catalog.GarmentSize    `json:"sizes"`
}
