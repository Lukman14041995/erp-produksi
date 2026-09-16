export type TxnType =
  | 'PURCHASE'
  | 'PRODUCTION_ISSUE'
  | 'PRODUCTION_RECEIPT'
  | 'SALE'
  | 'ADJUSTMENT'
  | 'TRANSFER_OUT'
  | 'TRANSFER_IN'
export type ItemType = 'MATERIAL' | 'PRODUCT'

export interface InventoryTransaction {
  id: string
  txn_type: TxnType
  item_type: ItemType
  warehouse_id: string
  material_id?: string
  product_id?: string
  product_size_id?: string
  qty_in: string
  qty_out: string
  unit_cost: string
  total_cost: string
  ref_type: string
  ref_id?: string
  txn_date: string
  notes: string
  created_at: string
}

export interface InventoryBalance {
  item_type: ItemType
  warehouse_id: string
  material_id?: string
  product_id?: string
  product_size_id?: string
  qty_on_hand: string
  avg_unit_cost: string
}
