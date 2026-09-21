package catalog

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductType is the top-level "Jenis Pesanan" choice (Jersey / T-Shirt).
// Fabrics, garment variants (model potongan), and inks are each scoped to
// one product type, since jersey fabric/print differs from t-shirt.
type ProductType struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Fabric struct {
	ID            uuid.UUID       `json:"id"`
	ProductTypeID uuid.UUID       `json:"product_type_id"`
	Code          string          `json:"code"`
	MaterialID    uuid.UUID       `json:"material_id"`
	Name          string          `json:"name"`
	SalesPrice    decimal.Decimal `json:"sales_price"`
	ImageURL      string          `json:"image_url"`
	SortOrder     int             `json:"sort_order"`
	IsActive      bool            `json:"is_active"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// GarmentVariant is the "Model" potongan (Normal/Long Sleeve/Raglan),
// picked by icon rather than a sample photo.
type GarmentVariant struct {
	ID            uuid.UUID       `json:"id"`
	ProductTypeID uuid.UUID       `json:"product_type_id"`
	Code          string          `json:"code"`
	Name          string          `json:"name"`
	PriceAddon    decimal.Decimal `json:"price_addon"`
	IconURL       string          `json:"icon_url"`
	SortOrder     int             `json:"sort_order"`
	IsActive      bool            `json:"is_active"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type GarmentSize struct {
	ID             uuid.UUID       `json:"id"`
	SizeCode       string          `json:"size_code"`
	SizeMultiplier decimal.Decimal `json:"size_multiplier"`
	SortOrder      int             `json:"sort_order"`
	IsActive       bool            `json:"is_active"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// Ink is the "Tinta" choice (Rubber, Plastisol, ...), scoped per product
// type since jersey and t-shirt printing use different processes.
type Ink struct {
	ID            uuid.UUID       `json:"id"`
	ProductTypeID uuid.UUID       `json:"product_type_id"`
	MaterialID    uuid.UUID       `json:"material_id"`
	Code          string          `json:"code"`
	Name          string          `json:"name"`
	PriceAddon    decimal.Decimal `json:"price_addon"`
	ImageURL      string          `json:"image_url"`
	SortOrder     int             `json:"sort_order"`
	IsActive      bool            `json:"is_active"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type UpsertProductTypeInput struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
	IsActive  *bool  `json:"is_active,omitempty"`
}

type UpsertFabricInput struct {
	ProductTypeID uuid.UUID       `json:"product_type_id"`
	Code          string          `json:"code"`
	MaterialID    uuid.UUID       `json:"material_id"`
	Name          string          `json:"name"`
	SalesPrice    decimal.Decimal `json:"sales_price"`
	SortOrder     int             `json:"sort_order"`
	IsActive      *bool           `json:"is_active,omitempty"`
}

type UpsertVariantInput struct {
	ProductTypeID uuid.UUID       `json:"product_type_id"`
	Code          string          `json:"code"`
	Name          string          `json:"name"`
	PriceAddon    decimal.Decimal `json:"price_addon"`
	SortOrder     int             `json:"sort_order"`
	IsActive      *bool           `json:"is_active,omitempty"`
}

type UpsertSizeInput struct {
	SizeCode       string          `json:"size_code"`
	SizeMultiplier decimal.Decimal `json:"size_multiplier"`
	SortOrder      int             `json:"sort_order"`
	IsActive       *bool           `json:"is_active,omitempty"`
}

type UpsertInkInput struct {
	ProductTypeID uuid.UUID       `json:"product_type_id"`
	MaterialID    uuid.UUID       `json:"material_id"`
	Code          string          `json:"code"`
	Name          string          `json:"name"`
	PriceAddon    decimal.Decimal `json:"price_addon"`
	SortOrder     int             `json:"sort_order"`
	IsActive      *bool           `json:"is_active,omitempty"`
}
