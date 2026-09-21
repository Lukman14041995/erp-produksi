package billing

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BankAccount struct {
	ID             uuid.UUID `json:"id"`
	BankName       string    `json:"bank_name"`
	AccountNumber  string    `json:"account_number"`
	AccountHolder  string    `json:"account_holder"`
	COAAccountCode string    `json:"coa_account_code"`
	SortOrder      int       `json:"sort_order"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UpsertBankAccountInput struct {
	BankName       string `json:"bank_name"`
	AccountNumber  string `json:"account_number"`
	AccountHolder  string `json:"account_holder"`
	COAAccountCode string `json:"coa_account_code"`
	SortOrder      int    `json:"sort_order"`
	IsActive       *bool  `json:"is_active,omitempty"`
}

type Method string

const (
	MethodFull        Method = "FULL"
	MethodInstallment Method = "INSTALLMENT"
)

type InstallmentStatus string

const (
	InstallmentPendingPayment      InstallmentStatus = "PENDING_PAYMENT"
	InstallmentPendingVerification InstallmentStatus = "PENDING_VERIFICATION"
	InstallmentConfirmed           InstallmentStatus = "CONFIRMED"
	InstallmentRejected            InstallmentStatus = "REJECTED"
)

type PaymentPlan struct {
	ID               uuid.UUID     `json:"id"`
	InvoiceID        uuid.UUID     `json:"invoice_id"`
	Method           Method        `json:"method"`
	InstallmentCount int           `json:"installment_count"`
	CreatedAt        time.Time     `json:"created_at"`
	Installments     []Installment `json:"installments,omitempty"`
}

type Installment struct {
	ID            uuid.UUID         `json:"id"`
	PaymentPlanID uuid.UUID         `json:"payment_plan_id"`
	SequenceNo    int               `json:"sequence_no"`
	Amount        decimal.Decimal   `json:"amount"`
	BankAccountID uuid.UUID         `json:"bank_account_id"`
	Status        InstallmentStatus `json:"status"`
	ProofImageURL string            `json:"proof_image_url"`
	RejectReason  string            `json:"reject_reason,omitempty"`
	PaymentID     *uuid.UUID        `json:"payment_id,omitempty"`
	ReviewedBy    *uuid.UUID        `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time        `json:"reviewed_at,omitempty"`
	SubmittedAt   *time.Time        `json:"submitted_at,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	BankAccount   *BankAccount      `json:"bank_account,omitempty"`
}

type ChoosePaymentPlanInput struct {
	Method           Method `json:"method"`
	InstallmentCount int    `json:"installment_count"`
}

type RejectInstallmentInput struct {
	Reason string `json:"reason"`
}
