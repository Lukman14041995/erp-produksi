import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost } from '@/lib/api'
import type { CreateOrderInput, Invoice, SalesOrder } from '@/types/sales'

export function useSalesOrders() {
  return useQuery({
    queryKey: ['sales-orders'],
    queryFn: () => apiGet<SalesOrder[]>('/sales-orders', { limit: 200 }),
  })
}

export function useSalesOrder(id: string | undefined) {
  return useQuery({
    queryKey: ['sales-orders', id],
    queryFn: () => apiGet<SalesOrder>(`/sales-orders/${id}`),
    enabled: !!id,
  })
}

export function useCreateSalesOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateOrderInput) => apiPost<SalesOrder>('/sales-orders', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['sales-orders'] }),
  })
}

function useOrderCommand(command: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body?: unknown }) => apiPost<SalesOrder>(`/sales-orders/${id}/${command}`, body),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['sales-orders'] })
      qc.invalidateQueries({ queryKey: ['sales-orders', id] })
      qc.invalidateQueries({ queryKey: ['invoices'] })
      qc.invalidateQueries({ queryKey: ['inventory-balances'] })
    },
  })
}

export const useConfirmOrder = () => useOrderCommand('confirm')
export const useCancelOrder = () => useOrderCommand('cancel')
export const useDeliverOrder = () => useOrderCommand('deliver')

export function useCreateInvoiceForOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (orderId: string) => apiPost<Invoice>(`/sales-orders/${orderId}/create-invoice`),
    onSuccess: (_, orderId) => {
      qc.invalidateQueries({ queryKey: ['sales-orders'] })
      qc.invalidateQueries({ queryKey: ['sales-orders', orderId] })
      qc.invalidateQueries({ queryKey: ['invoices'] })
    },
  })
}

export function useInvoices() {
  return useQuery({
    queryKey: ['invoices'],
    queryFn: () => apiGet<Invoice[]>('/invoices', { limit: 200 }),
  })
}

export function useInvoice(id: string | undefined) {
  return useQuery({
    queryKey: ['invoices', id],
    queryFn: () => apiGet<Invoice>(`/invoices/${id}`),
    enabled: !!id,
  })
}
