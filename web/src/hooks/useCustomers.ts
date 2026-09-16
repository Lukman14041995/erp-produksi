import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost, apiPut } from '@/lib/api'
import type { Customer } from '@/types/master'

export function useCustomers() {
  return useQuery({
    queryKey: ['customers'],
    queryFn: () => apiGet<Customer[]>('/customers', { limit: 200 }),
  })
}

export function useCreateCustomer() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: Partial<Customer>) => apiPost<Customer>('/customers', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['customers'] }),
  })
}

export function useUpdateCustomer() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: Partial<Customer> }) => apiPut<Customer>(`/customers/${id}`, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['customers'] }),
  })
}
