package reporting

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TrialBalanceRow struct {
	AccountCode string          `json:"account_code"`
	AccountName string          `json:"account_name"`
	AccountType string          `json:"account_type"`
	TotalDebit  decimal.Decimal `json:"total_debit"`
	TotalCredit decimal.Decimal `json:"total_credit"`
	Balance     decimal.Decimal `json:"balance"` // signed per the account's normal balance
}

type TrialBalance struct {
	AsOf  time.Time          `json:"as_of"`
	Rows  []TrialBalanceRow  `json:"rows"`
	TotalDebit  decimal.Decimal `json:"total_debit"`
	TotalCredit decimal.Decimal `json:"total_credit"`
}

type PLLine struct {
	AccountCode string          `json:"account_code"`
	AccountName string          `json:"account_name"`
	Amount      decimal.Decimal `json:"amount"`
}

type ProfitLoss struct {
	From          time.Time       `json:"from"`
	To            time.Time       `json:"to"`
	Revenue       []PLLine        `json:"revenue"`
	TotalRevenue  decimal.Decimal `json:"total_revenue"`
	COGS          []PLLine        `json:"cogs"`
	TotalCOGS     decimal.Decimal `json:"total_cogs"`
	GrossProfit   decimal.Decimal `json:"gross_profit"`
	Expenses      []PLLine        `json:"expenses"`
	TotalExpenses decimal.Decimal `json:"total_expenses"`
	NetProfit     decimal.Decimal `json:"net_profit"`
}

type BSLine struct {
	AccountCode string          `json:"account_code"`
	AccountName string          `json:"account_name"`
	Balance     decimal.Decimal `json:"balance"`
}

type BalanceSheet struct {
	AsOf             time.Time       `json:"as_of"`
	Assets           []BSLine        `json:"assets"`
	TotalAssets      decimal.Decimal `json:"total_assets"`
	Liabilities      []BSLine        `json:"liabilities"`
	TotalLiabilities decimal.Decimal `json:"total_liabilities"`
	Equity           []BSLine        `json:"equity"`
	RetainedEarnings decimal.Decimal `json:"retained_earnings_current"` // cumulative net income not yet closed to equity
	TotalEquity      decimal.Decimal `json:"total_equity"`
}

type CashFlow struct {
	From                  time.Time       `json:"from"`
	To                    time.Time       `json:"to"`
	BeginningCashBalance  decimal.Decimal `json:"beginning_cash_balance"`
	CashFromCustomers     decimal.Decimal `json:"cash_from_customers"`
	CashForExpenses       decimal.Decimal `json:"cash_for_expenses"`
	CashForSuppliers      decimal.Decimal `json:"cash_for_suppliers"`
	NetChange             decimal.Decimal `json:"net_change"`
	EndingCashBalance     decimal.Decimal `json:"ending_cash_balance"`
}

type AgingBucket struct {
	Current   decimal.Decimal `json:"current"` // not yet due
	Days1to30 decimal.Decimal `json:"days_1_30"`
	Days31to60 decimal.Decimal `json:"days_31_60"`
	Days61to90 decimal.Decimal `json:"days_61_90"`
	Over90    decimal.Decimal `json:"over_90"`
	Total     decimal.Decimal `json:"total"`
}

type ARAgingRow struct {
	CustomerID   uuid.UUID `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	AgingBucket
}

type ARAgingReport struct {
	AsOf time.Time    `json:"as_of"`
	Rows []ARAgingRow `json:"rows"`
	AgingBucket
}

type APAgingRow struct {
	SupplierID   uuid.UUID `json:"supplier_id"`
	SupplierName string    `json:"supplier_name"`
	AgingBucket
}

// APAgingReport buckets unpaid supplier bills (the AP subledger) by due date,
// mirroring ARAgingReport on the receivables side.
type APAgingReport struct {
	AsOf time.Time    `json:"as_of"`
	Rows []APAgingRow `json:"rows"`
	AgingBucket
}

type OrderProfitability struct {
	SalesOrderID     uuid.UUID       `json:"sales_order_id"`
	SONumber         string          `json:"so_number"`
	CustomerName     string          `json:"customer_name"`
	Revenue          decimal.Decimal `json:"revenue"`
	DirectMaterial   decimal.Decimal `json:"direct_material"`
	DirectLabor      decimal.Decimal `json:"direct_labor"`
	AllocatedOverhead decimal.Decimal `json:"allocated_overhead"`
	TotalCost        decimal.Decimal `json:"total_cost"`
	Profit           decimal.Decimal `json:"profit"`
	MarginPct        decimal.Decimal `json:"margin_pct"`
}
