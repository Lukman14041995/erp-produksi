import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost } from '@/lib/api'
import type {
  AddLaborInput,
  AddOverheadInput,
  IssueMaterialInput,
  LogStageInput,
  MarkMilestoneInput,
  SpkLabor,
  SpkMaterialUsage,
  SpkOrder,
  SpkOverhead,
  SpkStage,
} from '@/types/spk'

export function useSpkOrders(productTypeCode: string) {
  return useQuery({
    queryKey: ['spk-orders', productTypeCode],
    queryFn: () => apiGet<SpkOrder[]>('/spk-orders', { product_type_code: productTypeCode, limit: 200 }),
  })
}

// useAllSpkOrders backs the cross-jenis production board (/production/costing).
export function useAllSpkOrders() {
  return useQuery({
    queryKey: ['spk-orders', 'all'],
    queryFn: () => apiGet<SpkOrder[]>('/spk-orders', { limit: 200 }),
  })
}

export function useSpkOrder(id: string | undefined) {
  return useQuery({
    queryKey: ['spk-orders', 'detail', id],
    queryFn: () => apiGet<SpkOrder>(`/spk-orders/${id}`),
    enabled: !!id,
  })
}

export function useSpkOrdersBySalesOrder(salesOrderId: string | undefined) {
  return useQuery({
    queryKey: ['spk-orders', 'by-sales-order', salesOrderId],
    queryFn: () => apiGet<SpkOrder[]>('/spk-orders', { sales_order_id: salesOrderId }),
    enabled: !!salesOrderId,
  })
}

function invalidateSpkOrder(qc: ReturnType<typeof useQueryClient>, id: string) {
  qc.invalidateQueries({ queryKey: ['spk-orders'] })
  qc.invalidateQueries({ queryKey: ['spk-orders', 'detail', id] })
}

export function useLogSpkStage() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ stageId, ...input }: { stageId: string; spkOrderId: string } & LogStageInput) =>
      apiPost<SpkStage>(`/spk-stages/${stageId}/log`, input),
    onSuccess: (_, { spkOrderId }) => invalidateSpkOrder(qc, spkOrderId),
  })
}

export function useMarkSpkMilestone() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ stageId, ...input }: { stageId: string; spkOrderId: string } & MarkMilestoneInput) =>
      apiPost<SpkStage>(`/spk-stages/${stageId}/milestone`, input),
    onSuccess: (_, { spkOrderId }) => invalidateSpkOrder(qc, spkOrderId),
  })
}

export function useIssueSpkMaterial() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ spkOrderId, ...input }: { spkOrderId: string } & IssueMaterialInput) =>
      apiPost<SpkMaterialUsage>(`/spk-orders/${spkOrderId}/materials`, input),
    onSuccess: (_, { spkOrderId }) => {
      invalidateSpkOrder(qc, spkOrderId)
      qc.invalidateQueries({ queryKey: ['inventory-balances'] })
    },
  })
}

export function useAddSpkLabor() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ spkOrderId, ...input }: { spkOrderId: string } & AddLaborInput) =>
      apiPost<SpkLabor>(`/spk-orders/${spkOrderId}/labor`, input),
    onSuccess: (_, { spkOrderId }) => invalidateSpkOrder(qc, spkOrderId),
  })
}

export function useAddSpkOverhead() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ spkOrderId, ...input }: { spkOrderId: string } & AddOverheadInput) =>
      apiPost<SpkOverhead>(`/spk-orders/${spkOrderId}/overheads`, input),
    onSuccess: (_, { spkOrderId }) => invalidateSpkOrder(qc, spkOrderId),
  })
}
