package product

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Product struct {
	ID        uuid.UUID       `json:"id"`
	Code      string          `json:"code"`
	Name      string          `json:"name"`
	Category  string          `json:"category"`
	UOM       string          `json:"uom"`
	BasePrice decimal.Decimal `json:"base_price"`
	IsActive  bool            `json:"is_active"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Sizes     []ProductSize   `json:"sizes,omitempty"`
}

type ProductSize struct {
	ID             uuid.UUID       `json:"id"`
	ProductID      uuid.UUID       `json:"product_id"`
	SizeCode       string          `json:"size_code"`
	SizeMultiplier decimal.Decimal `json:"size_multiplier"`
	SortOrder      int             `json:"sort_order"`
	IsActive       bool            `json:"is_active"`
}

type UpsertProductInput struct {
	Code      string          `json:"code"`
	Name      string          `json:"name"`
	Category  string          `json:"category"`
	UOM       string          `json:"uom"`
	BasePrice decimal.Decimal `json:"base_price"`
	IsActive  *bool           `json:"is_active,omitempty"`
}

type UpsertSizeInput struct {
	SizeCode       string          `json:"size_code"`
	SizeMultiplier decimal.Decimal `json:"size_multiplier"`
	SortOrder      int             `json:"sort_order"`
	IsActive       *bool           `json:"is_active,omitempty"`
}
