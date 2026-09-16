import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost } from '@/lib/api'
import type { CreateExpenseInput, CreatePaymentInput, Expense, Payment } from '@/types/finance'

export function usePayments() {
  return useQuery({
    queryKey: ['payments'],
    queryFn: () => apiGet<Payment[]>('/payments', { limit: 200 }),
  })
}

export function usePayment(id: string | undefined) {
  return useQuery({
    queryKey: ['payments', id],
    queryFn: () => apiGet<Payment>(`/payments/${id}`),
    enabled: !!id,
  })
}

export function useCreatePayment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreatePaymentInput) => apiPost<Payment>('/payments', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['payments'] }),
  })
}

function usePaymentCommand(command: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body?: unknown }) => apiPost<Payment>(`/payments/${id}/${command}`, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['payments'] })
      qc.invalidateQueries({ queryKey: ['invoices'] })
      qc.invalidateQueries({ queryKey: ['supplier-invoices'] })
      qc.invalidateQueries({ queryKey: ['sales-orders'] })
      qc.invalidateQueries({ queryKey: ['journals'] })
    },
  })
}

export const usePostPayment = () => usePaymentCommand('post')
export const useVoidPayment = () => usePaymentCommand('void')

export function useExpenses() {
  return useQuery({
    queryKey: ['expenses'],
    queryFn: () => apiGet<Expense[]>('/expenses', { limit: 200 }),
  })
}

export function useCreateExpense() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateExpenseInput) => apiPost<Expense>('/expenses', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expenses'] }),
  })
}

function useExpenseCommand(command: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body?: unknown }) => apiPost<Expense>(`/expenses/${id}/${command}`, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['expenses'] })
      qc.invalidateQueries({ queryKey: ['journals'] })
    },
  })
}

export const usePostExpense = () => useExpenseCommand('post')
export const useVoidExpense = () => useExpenseCommand('void')
