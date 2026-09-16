export type PaymentType = 'RECEIPT' | 'DISBURSEMENT'
export type DocStatus = 'DRAFT' | 'POSTED' | 'VOID'

export interface PaymentAllocation {
  id: string
  payment_id: string
  invoice_id: string
  amount_allocated: string
}

export interface SupplierPaymentAllocation {
  id: string
  payment_id: string
  supplier_invoice_id: string
  amount_allocated: string
}

export interface Payment {
  id: string
  payment_number: string
  payment_type: PaymentType
  customer_id?: string
  supplier_id?: string
  cash_bank_account_id: string
  payment_date: string
  amount: string
  method: string
  reference_no: string
  status: DocStatus
  journal_id?: string
  notes: string
  created_at: string
  allocations?: PaymentAllocation[]
  bill_allocations?: SupplierPaymentAllocation[]
}

export interface CreatePaymentInput {
  payment_type: PaymentType
  customer_id?: string
  supplier_id?: string
  cash_bank_account_code: string
  payment_date?: string
  amount: string
  method?: string
  reference_no?: string
  notes?: string
  allocations?: { invoice_id: string; amount: string }[]
  bill_allocations?: { supplier_invoice_id: string; amount: string }[]
}

export interface Expense {
  id: string
  expense_number: string
  expense_date: string
  expense_account_id: string
  paid_from_account_id: string
  supplier_id?: string
  amount: string
  description: string
  status: DocStatus
  journal_id?: string
  created_at: string
}

export interface CreateExpenseInput {
  expense_account_code: string
  paid_from_account_code: string
  supplier_id?: string
  expense_date?: string
  amount: string
  description: string
}
