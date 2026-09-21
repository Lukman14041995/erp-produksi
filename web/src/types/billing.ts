export interface BankAccount {
  id: string
  bank_name: string
  account_number: string
  account_holder: string
  coa_account_code: string
  sort_order: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export type PaymentMethod = 'FULL' | 'INSTALLMENT'
export type InstallmentStatus = 'PENDING_PAYMENT' | 'PENDING_VERIFICATION' | 'CONFIRMED' | 'REJECTED'

export interface Installment {
  id: string
  payment_plan_id: string
  sequence_no: number
  amount: string
  bank_account_id: string
  status: InstallmentStatus
  proof_image_url: string
  reject_reason?: string
  payment_id?: string
  reviewed_by?: string
  reviewed_at?: string
  submitted_at?: string
  created_at: string
  bank_account?: BankAccount
}

export interface PaymentPlan {
  id: string
  invoice_id: string
  method: PaymentMethod
  installment_count: number
  created_at: string
  installments?: Installment[]
}

export interface ChoosePaymentPlanInput {
  method: PaymentMethod
  installment_count: number
}
