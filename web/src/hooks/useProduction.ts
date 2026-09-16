import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost } from '@/lib/api'
import type {
  BOMHeader,
  CreateBOMInput,
  CreateProductionOrderInput,
  ProductionCost,
  ProductionOrder,
} from '@/types/production'

export function useBOMsByProduct(productId: string | undefined) {
  return useQuery({
    queryKey: ['boms', productId],
    queryFn: () => apiGet<BOMHeader[]>(`/products/${productId}/boms`),
    enabled: !!productId,
  })
}

export function useCreateBOM() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateBOMInput) => apiPost<BOMHeader>('/boms', input),
    onSuccess: (_, input) => qc.invalidateQueries({ queryKey: ['boms', input.product_id] }),
  })
}

export function useProductionOrders() {
  return useQuery({
    queryKey: ['production-orders'],
    queryFn: () => apiGet<ProductionOrder[]>('/production-orders', { limit: 200 }),
  })
}

export function useProductionOrder(id: string | undefined) {
  return useQuery({
    queryKey: ['production-orders', id],
    queryFn: () => apiGet<ProductionOrder>(`/production-orders/${id}`),
    enabled: !!id,
  })
}

export function useProductionCost(id: string | undefined) {
  return useQuery({
    queryKey: ['production-orders', id, 'cost'],
    queryFn: () => apiGet<ProductionCost>(`/production-orders/${id}/cost`),
    enabled: !!id,
    retry: false,
  })
}

export function useCreateProductionOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateProductionOrderInput) => apiPost<ProductionOrder>('/production-orders', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['production-orders'] }),
  })
}

function invalidateOrder(qc: ReturnType<typeof useQueryClient>, id: string) {
  qc.invalidateQueries({ queryKey: ['production-orders'] })
  qc.invalidateQueries({ queryKey: ['production-orders', id] })
  qc.invalidateQueries({ queryKey: ['sales-orders'] })
  qc.invalidateQueries({ queryKey: ['inventory-balances'] })
}

export function useStartProductionOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiPost<ProductionOrder>(`/production-orders/${id}/start`),
    onSuccess: (_, id) => invalidateOrder(qc, id),
  })
}

export function useCancelProductionOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiPost<ProductionOrder>(`/production-orders/${id}/cancel`),
    onSuccess: (_, id) => invalidateOrder(qc, id),
  })
}

export function useIssueMaterial() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ orderId, materialRowId, qty }: { orderId: string; materialRowId: string; qty: string }) =>
      apiPost(`/production-orders/${orderId}/materials/${materialRowId}/issue`, { qty }),
    onSuccess: (_, { orderId }) => invalidateOrder(qc, orderId),
  })
}

export function useAddLabor() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ orderId, description, hours, rate }: { orderId: string; description: string; hours: string; rate: string }) =>
      apiPost(`/production-orders/${orderId}/labor`, { description, hours, rate }),
    onSuccess: (_, { orderId }) => invalidateOrder(qc, orderId),
  })
}

export function useAddOverhead() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      orderId,
      description,
      allocation_basis,
      amount,
    }: {
      orderId: string
      description: string
      allocation_basis: string
      amount: string
    }) => apiPost(`/production-orders/${orderId}/overheads`, { description, allocation_basis, amount }),
    onSuccess: (_, { orderId }) => invalidateOrder(qc, orderId),
  })
}

export function useCompleteProductionOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ orderId, items }: { orderId: string; items: { product_size_id: string; finished_qty: string }[] }) =>
      apiPost<ProductionCost>(`/production-orders/${orderId}/complete`, { items }),
    onSuccess: (_, { orderId }) => {
      invalidateOrder(qc, orderId)
      qc.invalidateQueries({ queryKey: ['production-orders', orderId, 'cost'] })
    },
  })
}
