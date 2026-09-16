import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost } from '@/lib/api'
import type { InventoryBalance, InventoryTransaction, ItemType } from '@/types/inventory'

export function useInventoryBalances() {
  return useQuery({
    queryKey: ['inventory-balances'],
    queryFn: () => apiGet<InventoryBalance[]>('/inventory/balances'),
  })
}

export function useStockByWarehouse(warehouseId?: string) {
  return useQuery({
    queryKey: ['stock-by-warehouse', warehouseId],
    queryFn: () => apiGet<InventoryBalance[]>('/inventory/stock-by-warehouse', warehouseId ? { warehouse_id: warehouseId } : undefined),
  })
}

export function useInventoryTransactions() {
  return useQuery({
    queryKey: ['inventory-transactions'],
    queryFn: () => apiGet<InventoryTransaction[]>('/inventory/transactions', { limit: 200 }),
  })
}

export interface AdjustInput {
  item_type: ItemType
  warehouse_id: string
  material_id?: string
  product_id?: string
  product_size_id?: string
  qty: string
  unit_cost: string
  notes?: string
}

export function useAdjustInventory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: AdjustInput) => apiPost('/inventory/adjustments', input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['inventory-balances'] })
      qc.invalidateQueries({ queryKey: ['inventory-transactions'] })
      qc.invalidateQueries({ queryKey: ['stock-by-warehouse'] })
    },
  })
}
