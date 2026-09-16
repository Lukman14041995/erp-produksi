export type SourceType =
  | 'SALES_INVOICE'
  | 'PAYMENT_RECEIPT'
  | 'PAYMENT_DISBURSEMENT'
  | 'MATERIAL_ISSUE'
  | 'LABOR_COST'
  | 'OVERHEAD_COST'
  | 'PRODUCTION_COMPLETION'
  | 'COGS'
  | 'EXPENSE'
  | 'SUPPLIER_BILL'
  | 'MANUAL'
  | 'REVERSAL'

export type JournalStatus = 'POSTED' | 'REVERSED'

export interface JournalEntryLine {
  id: string
  line_no: number
  account_id: string
  account_code: string
  account_name: string
  debit: string
  credit: string
  description: string
}

export interface JournalEntry {
  id: string
  journal_number: string
  journal_date: string
  source_type: SourceType
  source_id?: string
  description: string
  status: JournalStatus
  reversed_journal_id?: string
  total_debit: string
  total_credit: string
  created_by: string
  created_at: string
  lines?: JournalEntryLine[]
}

export type PeriodStatus = 'OPEN' | 'CLOSED'

export interface AccountingPeriod {
  id: string
  period: string
  start_date: string
  end_date: string
  status: PeriodStatus
  closed_at?: string
  closed_by?: string
}
