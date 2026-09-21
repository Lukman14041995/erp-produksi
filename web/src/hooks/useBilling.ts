import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost, apiPut } from '@/lib/api'
import type { BankAccount, PaymentPlan } from '@/types/billing'

export function useBankAccounts() {
  return useQuery({
    queryKey: ['bank-accounts'],
    queryFn: () => apiGet<BankAccount[]>('/billing/bank-accounts'),
  })
}

export function useCreateBankAccount() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: Partial<BankAccount>) => apiPost<BankAccount>('/billing/bank-accounts', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['bank-accounts'] }),
  })
}

export function useUpdateBankAccount() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: Partial<BankAccount> }) =>
      apiPut<BankAccount>(`/billing/bank-accounts/${id}`, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['bank-accounts'] }),
  })
}

export function usePaymentPlanByInvoice(invoiceId: string | undefined) {
  return useQuery({
    queryKey: ['payment-plans', 'by-invoice', invoiceId],
    queryFn: () => apiGet<PaymentPlan | null>(`/billing/payment-plans/by-invoice/${invoiceId}`),
    enabled: !!invoiceId,
  })
}

export function useConfirmInstallment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiPost(`/billing/installments/${id}/confirm`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['payment-plans'] })
      qc.invalidateQueries({ queryKey: ['invoices'] })
      qc.invalidateQueries({ queryKey: ['sales-orders'] })
    },
  })
}

export function useRejectInstallment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) => apiPost(`/billing/installments/${id}/reject`, { reason }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['payment-plans'] }),
  })
}
