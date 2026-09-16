import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost, apiPut } from '@/lib/api'
import type { Supplier } from '@/types/master'

export function useSuppliers() {
  return useQuery({
    queryKey: ['suppliers'],
    queryFn: () => apiGet<Supplier[]>('/suppliers', { limit: 200 }),
  })
}

export function useCreateSupplier() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: Partial<Supplier>) => apiPost<Supplier>('/suppliers', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suppliers'] }),
  })
}

export function useUpdateSupplier() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: Partial<Supplier> }) => apiPut<Supplier>(`/suppliers/${id}`, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suppliers'] }),
  })
}
