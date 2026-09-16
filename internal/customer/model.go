package customer

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Customer struct {
	ID             uuid.UUID       `json:"id"`
	Code           string          `json:"code"`
	Name           string          `json:"name"`
	ContactPerson  string          `json:"contact_person"`
	Phone          string          `json:"phone"`
	Email          string          `json:"email"`
	Address        string          `json:"address"`
	TaxID          string          `json:"tax_id"`
	CreditLimit    decimal.Decimal `json:"credit_limit"`
	IsActive       bool            `json:"is_active"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type UpsertInput struct {
	Code          string          `json:"code"`
	Name          string          `json:"name"`
	ContactPerson string          `json:"contact_person"`
	Phone         string          `json:"phone"`
	Email         string          `json:"email"`
	Address       string          `json:"address"`
	TaxID         string          `json:"tax_id"`
	CreditLimit   decimal.Decimal `json:"credit_limit"`
	IsActive      *bool           `json:"is_active,omitempty"`
}
