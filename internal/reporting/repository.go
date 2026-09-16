package reporting

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/shopspring/decimal"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type accountActivityRow struct {
	Code        string
	Name        string
	AccountType string
	NormalBalance string
	Debit       decimal.Decimal
	Credit      decimal.Decimal
}

// activityByAccount sums posted journal line debits/credits per account for
// journal_date in (from, to]. Pass a zero from to mean "from the beginning
// of time" (used for balance-sheet style cumulative balances).
func (r *Repository) activityByAccount(ctx context.Context, from, to time.Time, accountTypes []string) ([]accountActivityRow, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT a.code, a.name, a.account_type, a.normal_balance,
		       COALESCE(SUM(l.debit), 0), COALESCE(SUM(l.credit), 0)
		FROM accounts a
		LEFT JOIN journal_entry_lines l ON l.account_id = a.id
		LEFT JOIN journal_entries j ON j.id = l.journal_id
		    AND j.status = 'POSTED'
		    AND j.journal_date > $1
		    AND j.journal_date <= $2
		WHERE a.account_type = ANY($3)
		GROUP BY a.code, a.name, a.account_type, a.normal_balance
		ORDER BY a.code
	`, from, to, accountTypes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []accountActivityRow{}
	for rows.Next() {
		var a accountActivityRow
		if err := rows.Scan(&a.Code, &a.Name, &a.AccountType, &a.NormalBalance, &a.Debit, &a.Credit); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *Repository) TrialBalanceActivity(ctx context.Context, asOf time.Time) ([]accountActivityRow, error) {
	return r.activityByAccount(ctx, time.Time{}, asOf, []string{"ASSET", "LIABILITY", "EQUITY", "REVENUE", "COGS", "EXPENSE"})
}

func (r *Repository) PeriodActivity(ctx context.Context, from, to time.Time, accountTypes []string) ([]accountActivityRow, error) {
	return r.activityByAccount(ctx, from, to, accountTypes)
}

func (r *Repository) CumulativeActivity(ctx context.Context, asOf time.Time, accountTypes []string) ([]accountActivityRow, error) {
	return r.activityByAccount(ctx, time.Time{}, asOf, accountTypes)
}

type invoiceAgingRow struct {
	CustomerID   string
	CustomerName string
	DueDate      *time.Time
	InvoiceDate  time.Time
	BalanceDue   decimal.Decimal
}

func (r *Repository) OpenInvoicesForAging(ctx context.Context, asOf time.Time) ([]invoiceAgingRow, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT c.id::text, c.name, i.due_date, i.invoice_date, i.balance_due
		FROM invoices i
		JOIN customers c ON c.id = i.customer_id
		WHERE i.status IN ('POSTED','PARTIALLY_PAID') AND i.balance_due > 0 AND i.invoice_date <= $1
	`, asOf)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []invoiceAgingRow{}
	for rows.Next() {
		var a invoiceAgingRow
		if err := rows.Scan(&a.CustomerID, &a.CustomerName, &a.DueDate, &a.InvoiceDate, &a.BalanceDue); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

type billAgingRow struct {
	SupplierID   string
	SupplierName string
	DueDate      *time.Time
	BillDate     time.Time
	BalanceDue   decimal.Decimal
}

func (r *Repository) OpenBillsForAging(ctx context.Context, asOf time.Time) ([]billAgingRow, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT s.id::text, s.name, b.due_date, b.bill_date, b.balance_due
		FROM supplier_invoices b
		JOIN suppliers s ON s.id = b.supplier_id
		WHERE b.status IN ('POSTED','PARTIALLY_PAID') AND b.balance_due > 0 AND b.bill_date <= $1
	`, asOf)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []billAgingRow{}
	for rows.Next() {
		var a billAgingRow
		if err := rows.Scan(&a.SupplierID, &a.SupplierName, &a.DueDate, &a.BillDate, &a.BalanceDue); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

type cashActivityRow struct {
	SourceType string
	Debit      decimal.Decimal
	Credit     decimal.Decimal
}

// CashActivity sums debits/credits to the Kas/Bank accounts (1-1001, 1-1002)
// grouped by the posting's source_type, for the direct-method cash flow
// statement.
func (r *Repository) CashActivity(ctx context.Context, from, to time.Time) ([]cashActivityRow, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT j.source_type, COALESCE(SUM(l.debit),0), COALESCE(SUM(l.credit),0)
		FROM journal_entry_lines l
		JOIN journal_entries j ON j.id = l.journal_id
		JOIN accounts a ON a.id = l.account_id
		WHERE a.code IN ('1-1001','1-1002') AND j.status = 'POSTED'
		  AND j.journal_date > $1 AND j.journal_date <= $2
		GROUP BY j.source_type
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []cashActivityRow{}
	for rows.Next() {
		var a cashActivityRow
		if err := rows.Scan(&a.SourceType, &a.Debit, &a.Credit); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *Repository) CashBalanceAsOf(ctx context.Context, asOf time.Time) (decimal.Decimal, error) {
	var balance decimal.Decimal
	err := db.Q(ctx, r.pool).QueryRow(ctx, `
		SELECT COALESCE(SUM(l.debit) - SUM(l.credit), 0)
		FROM journal_entry_lines l
		JOIN journal_entries j ON j.id = l.journal_id
		JOIN accounts a ON a.id = l.account_id
		WHERE a.code IN ('1-1001','1-1002') AND j.status = 'POSTED' AND j.journal_date <= $1
	`, asOf).Scan(&balance)
	return balance, err
}

type orderProfitabilityRow struct {
	SalesOrderID   string
	SONumber       string
	CustomerName   string
	Revenue        decimal.Decimal
	DirectMaterial decimal.Decimal
	DirectLabor    decimal.Decimal
	Overhead       decimal.Decimal
}

func (r *Repository) OrderProfitability(ctx context.Context) ([]orderProfitabilityRow, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT so.id::text, so.so_number, c.name, so.grand_total,
		       COALESCE(SUM(pc.total_material_cost), 0),
		       COALESCE(SUM(pc.total_labor_cost), 0),
		       COALESCE(SUM(pc.total_overhead_cost), 0)
		FROM sales_orders so
		JOIN customers c ON c.id = so.customer_id
		LEFT JOIN production_orders po ON po.sales_order_id = so.id
		LEFT JOIN production_costs pc ON pc.production_order_id = po.id
		WHERE so.order_status != 'CANCELLED'
		GROUP BY so.id, so.so_number, c.name, so.grand_total
		ORDER BY so.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []orderProfitabilityRow{}
	for rows.Next() {
		var o orderProfitabilityRow
		if err := rows.Scan(&o.SalesOrderID, &o.SONumber, &o.CustomerName, &o.Revenue, &o.DirectMaterial, &o.DirectLabor, &o.Overhead); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}
