package finance

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentType string

const (
	PaymentReceipt       PaymentType = "RECEIPT"
	PaymentDisbursement  PaymentType = "DISBURSEMENT"
)

type DocStatus string

const (
	StatusDraft  DocStatus = "DRAFT"
	StatusPosted DocStatus = "POSTED"
	StatusVoid   DocStatus = "VOID"
)

type Payment struct {
	ID                 uuid.UUID       `json:"id"`
	PaymentNumber      string          `json:"payment_number"`
	PaymentType        PaymentType     `json:"payment_type"`
	CustomerID         *uuid.UUID      `json:"customer_id,omitempty"`
	SupplierID         *uuid.UUID      `json:"supplier_id,omitempty"`
	CashBankAccountID  uuid.UUID       `json:"cash_bank_account_id"`
	PaymentDate        time.Time       `json:"payment_date"`
	Amount             decimal.Decimal `json:"amount"`
	Method             string          `json:"method"`
	ReferenceNo        string          `json:"reference_no"`
	Status             DocStatus       `json:"status"`
	JournalID          *uuid.UUID      `json:"journal_id,omitempty"`
	Notes              string          `json:"notes"`
	CreatedAt          time.Time       `json:"created_at"`
	Allocations        []PaymentAllocation `json:"allocations,omitempty"`
	BillAllocations    []SupplierPaymentAllocation `json:"bill_allocations,omitempty"`
}

type PaymentAllocation struct {
	ID              uuid.UUID       `json:"id"`
	PaymentID       uuid.UUID       `json:"payment_id"`
	InvoiceID       uuid.UUID       `json:"invoice_id"`
	AmountAllocated decimal.Decimal `json:"amount_allocated"`
}

type SupplierPaymentAllocation struct {
	ID                uuid.UUID       `json:"id"`
	PaymentID         uuid.UUID       `json:"payment_id"`
	SupplierInvoiceID uuid.UUID       `json:"supplier_invoice_id"`
	AmountAllocated   decimal.Decimal `json:"amount_allocated"`
}

type Expense struct {
	ID                  uuid.UUID       `json:"id"`
	ExpenseNumber       string          `json:"expense_number"`
	ExpenseDate         time.Time       `json:"expense_date"`
	ExpenseAccountID    uuid.UUID       `json:"expense_account_id"`
	PaidFromAccountID   uuid.UUID       `json:"paid_from_account_id"`
	SupplierID          *uuid.UUID      `json:"supplier_id,omitempty"`
	Amount              decimal.Decimal `json:"amount"`
	Description         string          `json:"description"`
	Status              DocStatus       `json:"status"`
	JournalID           *uuid.UUID      `json:"journal_id,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
}

type AllocationInput struct {
	InvoiceID uuid.UUID       `json:"invoice_id"`
	Amount    decimal.Decimal `json:"amount"`
}

// BillAllocationInput is the DISBURSEMENT-side counterpart of
// AllocationInput, matching a payment against an AP subledger bill instead
// of an AR invoice.
type BillAllocationInput struct {
	SupplierInvoiceID uuid.UUID       `json:"supplier_invoice_id"`
	Amount            decimal.Decimal `json:"amount"`
}

type CreatePaymentInput struct {
	PaymentType       PaymentType       `json:"payment_type"`
	CustomerID        *uuid.UUID        `json:"customer_id,omitempty"`
	SupplierID        *uuid.UUID        `json:"supplier_id,omitempty"`
	CashBankAccountCode string          `json:"cash_bank_account_code"`
	PaymentDate       string            `json:"payment_date,omitempty"` // YYYY-MM-DD, defaults to today
	Amount            decimal.Decimal   `json:"amount"`
	Method            string            `json:"method"`
	ReferenceNo       string            `json:"reference_no"`
	Notes             string            `json:"notes"`
	Allocations       []AllocationInput     `json:"allocations,omitempty"`
	BillAllocations   []BillAllocationInput `json:"bill_allocations,omitempty"`
}

type CreateExpenseInput struct {
	ExpenseAccountCode    string     `json:"expense_account_code"`
	PaidFromAccountCode   string     `json:"paid_from_account_code"`
	SupplierID            *uuid.UUID `json:"supplier_id,omitempty"`
	ExpenseDate           string     `json:"expense_date,omitempty"` // YYYY-MM-DD, defaults to today
	Amount                decimal.Decimal `json:"amount"`
	Description           string     `json:"description"`
}
