package material

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Material struct {
	ID        uuid.UUID       `json:"id"`
	Code      string          `json:"code"`
	Name      string          `json:"name"`
	UOM       string          `json:"uom"`
	UnitCost  decimal.Decimal `json:"unit_cost"`
	IsActive  bool            `json:"is_active"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type UpsertInput struct {
	Code     string          `json:"code"`
	Name     string          `json:"name"`
	UOM      string          `json:"uom"`
	UnitCost decimal.Decimal `json:"unit_cost"`
	IsActive *bool           `json:"is_active,omitempty"`
}
