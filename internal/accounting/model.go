package accounting

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type SourceType string

const (
	SourceSalesInvoice          SourceType = "SALES_INVOICE"
	SourcePaymentReceipt        SourceType = "PAYMENT_RECEIPT"
	SourcePaymentDisbursement   SourceType = "PAYMENT_DISBURSEMENT"
	SourceMaterialIssue         SourceType = "MATERIAL_ISSUE"
	SourceLaborCost             SourceType = "LABOR_COST"
	SourceOverheadCost          SourceType = "OVERHEAD_COST"
	SourceProductionCompletion  SourceType = "PRODUCTION_COMPLETION"
	SourceCOGS                  SourceType = "COGS"
	SourceExpense               SourceType = "EXPENSE"
	SourceSupplierBill          SourceType = "SUPPLIER_BILL"
	SourceGoodsReceipt          SourceType = "GOODS_RECEIPT"
	SourceStockOpname           SourceType = "STOCK_OPNAME"
	SourceManual                SourceType = "MANUAL"
	SourceReversal              SourceType = "REVERSAL"
)

type Status string

const (
	StatusPosted   Status = "POSTED"
	StatusReversed Status = "REVERSED"
)

// JournalLineInput is what callers (sales, finance, production services)
// build to describe one leg of a double-entry posting. AccountCode is
// resolved to an account_id inside PostJournal.
type JournalLineInput struct {
	AccountCode string
	Debit       decimal.Decimal
	Credit      decimal.Decimal
	Description string
}

type JournalEntryLine struct {
	ID          uuid.UUID       `json:"id"`
	LineNo      int             `json:"line_no"`
	AccountID   uuid.UUID       `json:"account_id"`
	AccountCode string          `json:"account_code"`
	AccountName string          `json:"account_name"`
	Debit       decimal.Decimal `json:"debit"`
	Credit      decimal.Decimal `json:"credit"`
	Description string          `json:"description"`
}

type JournalEntry struct {
	ID                uuid.UUID          `json:"id"`
	JournalNumber     string             `json:"journal_number"`
	JournalDate       time.Time          `json:"journal_date"`
	SourceType        SourceType         `json:"source_type"`
	SourceID          *uuid.UUID         `json:"source_id,omitempty"`
	Description       string             `json:"description"`
	Status            Status             `json:"status"`
	ReversedJournalID *uuid.UUID         `json:"reversed_journal_id,omitempty"`
	TotalDebit        decimal.Decimal    `json:"total_debit"`
	TotalCredit       decimal.Decimal    `json:"total_credit"`
	CreatedBy         string             `json:"created_by"`
	CreatedAt         time.Time          `json:"created_at"`
	Lines             []JournalEntryLine `json:"lines,omitempty"`
}
