package reporting

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/shopspring/decimal"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// signedBalance returns debit-credit for debit-normal accounts and
// credit-debit for credit-normal accounts, so "balance" is always expressed
// in the direction that account naturally grows.
func signedBalance(normalBalance string, debit, credit decimal.Decimal) decimal.Decimal {
	if normalBalance == "DEBIT" {
		return debit.Sub(credit)
	}
	return credit.Sub(debit)
}

func (s *Service) TrialBalance(ctx context.Context, asOf time.Time) (TrialBalance, error) {
	rows, err := s.repo.TrialBalanceActivity(ctx, asOf)
	if err != nil {
		return TrialBalance{}, apperr.Internal("compute trial balance", err)
	}

	tb := TrialBalance{AsOf: asOf, Rows: []TrialBalanceRow{}, TotalDebit: decimal.Zero, TotalCredit: decimal.Zero}
	for _, r := range rows {
		tb.Rows = append(tb.Rows, TrialBalanceRow{
			AccountCode: r.Code, AccountName: r.Name, AccountType: r.AccountType,
			TotalDebit: r.Debit, TotalCredit: r.Credit, Balance: signedBalance(r.NormalBalance, r.Debit, r.Credit),
		})
		tb.TotalDebit = tb.TotalDebit.Add(r.Debit)
		tb.TotalCredit = tb.TotalCredit.Add(r.Credit)
	}
	return tb, nil
}

func (s *Service) ProfitLoss(ctx context.Context, from, to time.Time) (ProfitLoss, error) {
	pl := ProfitLoss{
		From: from, To: to,
		Revenue: []PLLine{}, COGS: []PLLine{}, Expenses: []PLLine{},
		TotalRevenue: decimal.Zero, TotalCOGS: decimal.Zero, TotalExpenses: decimal.Zero,
	}

	revenueRows, err := s.repo.PeriodActivity(ctx, from, to, []string{"REVENUE"})
	if err != nil {
		return ProfitLoss{}, apperr.Internal("load revenue activity", err)
	}
	for _, r := range revenueRows {
		amount := signedBalance(r.NormalBalance, r.Debit, r.Credit)
		pl.Revenue = append(pl.Revenue, PLLine{AccountCode: r.Code, AccountName: r.Name, Amount: amount})
		pl.TotalRevenue = pl.TotalRevenue.Add(amount)
	}

	cogsRows, err := s.repo.PeriodActivity(ctx, from, to, []string{"COGS"})
	if err != nil {
		return ProfitLoss{}, apperr.Internal("load COGS activity", err)
	}
	for _, r := range cogsRows {
		amount := signedBalance(r.NormalBalance, r.Debit, r.Credit)
		pl.COGS = append(pl.COGS, PLLine{AccountCode: r.Code, AccountName: r.Name, Amount: amount})
		pl.TotalCOGS = pl.TotalCOGS.Add(amount)
	}
	pl.GrossProfit = pl.TotalRevenue.Sub(pl.TotalCOGS)

	expenseRows, err := s.repo.PeriodActivity(ctx, from, to, []string{"EXPENSE"})
	if err != nil {
		return ProfitLoss{}, apperr.Internal("load expense activity", err)
	}
	for _, r := range expenseRows {
		amount := signedBalance(r.NormalBalance, r.Debit, r.Credit)
		pl.Expenses = append(pl.Expenses, PLLine{AccountCode: r.Code, AccountName: r.Name, Amount: amount})
		pl.TotalExpenses = pl.TotalExpenses.Add(amount)
	}
	pl.NetProfit = pl.GrossProfit.Sub(pl.TotalExpenses)

	return pl, nil
}

func (s *Service) BalanceSheet(ctx context.Context, asOf time.Time) (BalanceSheet, error) {
	bs := BalanceSheet{
		AsOf: asOf,
		Assets: []BSLine{}, Liabilities: []BSLine{}, Equity: []BSLine{},
		TotalAssets: decimal.Zero, TotalLiabilities: decimal.Zero, TotalEquity: decimal.Zero,
	}

	assetRows, err := s.repo.CumulativeActivity(ctx, asOf, []string{"ASSET"})
	if err != nil {
		return BalanceSheet{}, apperr.Internal("load asset balances", err)
	}
	for _, r := range assetRows {
		bal := signedBalance(r.NormalBalance, r.Debit, r.Credit)
		bs.Assets = append(bs.Assets, BSLine{AccountCode: r.Code, AccountName: r.Name, Balance: bal})
		bs.TotalAssets = bs.TotalAssets.Add(bal)
	}

	liabilityRows, err := s.repo.CumulativeActivity(ctx, asOf, []string{"LIABILITY"})
	if err != nil {
		return BalanceSheet{}, apperr.Internal("load liability balances", err)
	}
	for _, r := range liabilityRows {
		bal := signedBalance(r.NormalBalance, r.Debit, r.Credit)
		bs.Liabilities = append(bs.Liabilities, BSLine{AccountCode: r.Code, AccountName: r.Name, Balance: bal})
		bs.TotalLiabilities = bs.TotalLiabilities.Add(bal)
	}

	equityRows, err := s.repo.CumulativeActivity(ctx, asOf, []string{"EQUITY"})
	if err != nil {
		return BalanceSheet{}, apperr.Internal("load equity balances", err)
	}
	for _, r := range equityRows {
		bal := signedBalance(r.NormalBalance, r.Debit, r.Credit)
		bs.Equity = append(bs.Equity, BSLine{AccountCode: r.Code, AccountName: r.Name, Balance: bal})
		bs.TotalEquity = bs.TotalEquity.Add(bal)
	}

	// Revenue - COGS - Expense accounts have not been closed to equity via a
	// formal closing journal, so their cumulative net is surfaced as current
	// retained earnings to keep Assets == Liabilities + Equity.
	pl, err := s.ProfitLoss(ctx, time.Time{}, asOf)
	if err != nil {
		return BalanceSheet{}, err
	}
	bs.RetainedEarnings = pl.NetProfit
	bs.TotalEquity = bs.TotalEquity.Add(pl.NetProfit)

	return bs, nil
}

func (s *Service) CashFlow(ctx context.Context, from, to time.Time) (CashFlow, error) {
	beginning, err := s.repo.CashBalanceAsOf(ctx, from)
	if err != nil {
		return CashFlow{}, apperr.Internal("compute beginning cash balance", err)
	}
	ending, err := s.repo.CashBalanceAsOf(ctx, to)
	if err != nil {
		return CashFlow{}, apperr.Internal("compute ending cash balance", err)
	}

	activity, err := s.repo.CashActivity(ctx, from, to)
	if err != nil {
		return CashFlow{}, apperr.Internal("compute cash activity", err)
	}

	cf := CashFlow{From: from, To: to, BeginningCashBalance: beginning, EndingCashBalance: ending, NetChange: ending.Sub(beginning)}
	for _, a := range activity {
		switch a.SourceType {
		case "PAYMENT_RECEIPT":
			cf.CashFromCustomers = cf.CashFromCustomers.Add(a.Debit).Sub(a.Credit)
		case "EXPENSE":
			cf.CashForExpenses = cf.CashForExpenses.Add(a.Credit).Sub(a.Debit)
		case "PAYMENT_DISBURSEMENT":
			cf.CashForSuppliers = cf.CashForSuppliers.Add(a.Credit).Sub(a.Debit)
		}
	}
	return cf, nil
}

func bucketFor(dueDate time.Time, asOf time.Time) string {
	daysOverdue := int(asOf.Sub(dueDate).Hours() / 24)
	switch {
	case daysOverdue <= 0:
		return "current"
	case daysOverdue <= 30:
		return "1-30"
	case daysOverdue <= 60:
		return "31-60"
	case daysOverdue <= 90:
		return "61-90"
	default:
		return "90+"
	}
}

func addToBucket(b *AgingBucket, bucket string, amount decimal.Decimal) {
	switch bucket {
	case "current":
		b.Current = b.Current.Add(amount)
	case "1-30":
		b.Days1to30 = b.Days1to30.Add(amount)
	case "31-60":
		b.Days31to60 = b.Days31to60.Add(amount)
	case "61-90":
		b.Days61to90 = b.Days61to90.Add(amount)
	default:
		b.Over90 = b.Over90.Add(amount)
	}
	b.Total = b.Total.Add(amount)
}

func (s *Service) ARAging(ctx context.Context, asOf time.Time) (ARAgingReport, error) {
	invoices, err := s.repo.OpenInvoicesForAging(ctx, asOf)
	if err != nil {
		return ARAgingReport{}, apperr.Internal("load open invoices", err)
	}

	byCustomer := map[string]*ARAgingRow{}
	report := ARAgingReport{AsOf: asOf, Rows: []ARAgingRow{}}

	for _, inv := range invoices {
		dueDate := inv.InvoiceDate.AddDate(0, 0, 30)
		if inv.DueDate != nil {
			dueDate = *inv.DueDate
		}
		bucket := bucketFor(dueDate, asOf)

		row, ok := byCustomer[inv.CustomerID]
		if !ok {
			id, _ := uuid.Parse(inv.CustomerID)
			row = &ARAgingRow{CustomerID: id, CustomerName: inv.CustomerName}
			byCustomer[inv.CustomerID] = row
		}
		addToBucket(&row.AgingBucket, bucket, inv.BalanceDue)
		addToBucket(&report.AgingBucket, bucket, inv.BalanceDue)
	}

	for _, row := range byCustomer {
		report.Rows = append(report.Rows, *row)
	}
	return report, nil
}

func (s *Service) APAging(ctx context.Context, asOf time.Time) (APAgingReport, error) {
	bills, err := s.repo.OpenBillsForAging(ctx, asOf)
	if err != nil {
		return APAgingReport{}, apperr.Internal("load open supplier bills", err)
	}

	bySupplier := map[string]*APAgingRow{}
	report := APAgingReport{AsOf: asOf, Rows: []APAgingRow{}}

	for _, b := range bills {
		dueDate := b.BillDate.AddDate(0, 0, 30)
		if b.DueDate != nil {
			dueDate = *b.DueDate
		}
		bucket := bucketFor(dueDate, asOf)

		row, ok := bySupplier[b.SupplierID]
		if !ok {
			id, _ := uuid.Parse(b.SupplierID)
			row = &APAgingRow{SupplierID: id, SupplierName: b.SupplierName}
			bySupplier[b.SupplierID] = row
		}
		addToBucket(&row.AgingBucket, bucket, b.BalanceDue)
		addToBucket(&report.AgingBucket, bucket, b.BalanceDue)
	}

	for _, row := range bySupplier {
		report.Rows = append(report.Rows, *row)
	}
	return report, nil
}

func (s *Service) OrderProfitability(ctx context.Context) ([]OrderProfitability, error) {
	rows, err := s.repo.OrderProfitability(ctx)
	if err != nil {
		return nil, apperr.Internal("compute order profitability", err)
	}

	out := make([]OrderProfitability, 0, len(rows))
	for _, r := range rows {
		id, _ := uuid.Parse(r.SalesOrderID)
		totalCost := r.DirectMaterial.Add(r.DirectLabor).Add(r.Overhead)
		profit := r.Revenue.Sub(totalCost)
		marginPct := decimal.Zero
		if r.Revenue.IsPositive() {
			marginPct = profit.Div(r.Revenue).Mul(decimal.NewFromInt(100)).Round(2)
		}
		out = append(out, OrderProfitability{
			SalesOrderID: id, SONumber: r.SONumber, CustomerName: r.CustomerName, Revenue: r.Revenue,
			DirectMaterial: r.DirectMaterial, DirectLabor: r.DirectLabor, AllocatedOverhead: r.Overhead,
			TotalCost: totalCost, Profit: profit, MarginPct: marginPct,
		})
	}
	return out, nil
}
