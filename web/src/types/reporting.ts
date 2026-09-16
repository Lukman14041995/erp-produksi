export interface TrialBalanceRow {
  account_code: string
  account_name: string
  account_type: string
  total_debit: string
  total_credit: string
  balance: string
}

export interface TrialBalance {
  as_of: string
  rows: TrialBalanceRow[]
  total_debit: string
  total_credit: string
}

export interface PLLine {
  account_code: string
  account_name: string
  amount: string
}

export interface ProfitLoss {
  from: string
  to: string
  revenue: PLLine[]
  total_revenue: string
  cogs: PLLine[]
  total_cogs: string
  gross_profit: string
  expenses: PLLine[]
  total_expenses: string
  net_profit: string
}

export interface BSLine {
  account_code: string
  account_name: string
  balance: string
}

export interface BalanceSheet {
  as_of: string
  assets: BSLine[]
  total_assets: string
  liabilities: BSLine[]
  total_liabilities: string
  equity: BSLine[]
  retained_earnings_current: string
  total_equity: string
}

export interface CashFlow {
  from: string
  to: string
  beginning_cash_balance: string
  cash_from_customers: string
  cash_for_expenses: string
  cash_for_suppliers: string
  net_change: string
  ending_cash_balance: string
}

export interface AgingBucket {
  current: string
  days_1_30: string
  days_31_60: string
  days_61_90: string
  over_90: string
  total: string
}

export interface ARAgingRow extends AgingBucket {
  customer_id: string
  customer_name: string
}

export interface ARAgingReport extends AgingBucket {
  as_of: string
  rows: ARAgingRow[]
}

export interface APAgingRow extends AgingBucket {
  supplier_id: string
  supplier_name: string
}

export interface APAgingReport extends AgingBucket {
  as_of: string
  rows: APAgingRow[]
}

export interface OrderProfitability {
  sales_order_id: string
  so_number: string
  customer_name: string
  revenue: string
  direct_material: string
  direct_labor: string
  allocated_overhead: string
  total_cost: string
  profit: string
  margin_pct: string
}
