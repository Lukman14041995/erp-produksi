import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost, apiPut } from '@/lib/api'
import type {
  CreateOpnameInput,
  CreateTransferInput,
  StockOpname,
  StockTransfer,
  UpsertWarehouseInput,
  Warehouse,
} from '@/types/warehouse'

// ---- Warehouses ----

export function useWarehouses() {
  return useQuery({
    queryKey: ['warehouses'],
    queryFn: () => apiGet<Warehouse[]>('/warehouses'),
  })
}

export function useCreateWarehouse() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: UpsertWarehouseInput) => apiPost<Warehouse>('/warehouses', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['warehouses'] }),
  })
}

export function useUpdateWarehouse() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpsertWarehouseInput }) => apiPut<Warehouse>(`/warehouses/${id}`, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['warehouses'] }),
  })
}

// ---- Stock transfers ----

export function useStockTransfers() {
  return useQuery({
    queryKey: ['stock-transfers'],
    queryFn: () => apiGet<StockTransfer[]>('/inventory/transfers', { limit: 200 }),
  })
}

export function useStockTransfer(id: string | undefined) {
  return useQuery({
    queryKey: ['stock-transfers', id],
    queryFn: () => apiGet<StockTransfer>(`/inventory/transfers/${id}`),
    enabled: !!id,
  })
}

export function useCreateTransfer() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateTransferInput) => apiPost<StockTransfer>('/inventory/transfers', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['stock-transfers'] }),
  })
}

function useTransferCommand(command: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiPost<StockTransfer>(`/inventory/transfers/${id}/${command}`),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ['stock-transfers'] })
      qc.invalidateQueries({ queryKey: ['stock-transfers', id] })
      qc.invalidateQueries({ queryKey: ['stock-by-warehouse'] })
      qc.invalidateQueries({ queryKey: ['inventory-balances'] })
    },
  })
}

export const useDispatchTransfer = () => useTransferCommand('dispatch')
export const useReceiveTransfer = () => useTransferCommand('receive')
export const useCancelTransfer = () => useTransferCommand('cancel')

// ---- Stock opname ----

export function useStockOpnames() {
  return useQuery({
    queryKey: ['stock-opnames'],
    queryFn: () => apiGet<StockOpname[]>('/inventory/opnames', { limit: 200 }),
  })
}

export function useStockOpname(id: string | undefined) {
  return useQuery({
    queryKey: ['stock-opnames', id],
    queryFn: () => apiGet<StockOpname>(`/inventory/opnames/${id}`),
    enabled: !!id,
  })
}

export function useCreateOpname() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateOpnameInput) => apiPost<StockOpname>('/inventory/opnames', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['stock-opnames'] }),
  })
}

export function usePostOpname() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiPost<StockOpname>(`/inventory/opnames/${id}/post`),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ['stock-opnames'] })
      qc.invalidateQueries({ queryKey: ['stock-opnames', id] })
      qc.invalidateQueries({ queryKey: ['stock-by-warehouse'] })
      qc.invalidateQueries({ queryKey: ['inventory-balances'] })
      qc.invalidateQueries({ queryKey: ['journals'] })
    },
  })
}

export function useCancelOpname() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiPost<StockOpname>(`/inventory/opnames/${id}/cancel`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['stock-opnames'] }),
  })
}
