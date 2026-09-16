export interface Warehouse {
  id: string
  code: string
  name: string
  address: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface UpsertWarehouseInput {
  code?: string
  name: string
  address?: string
  is_active?: boolean
}

// ---- Stock transfers ----

export type TransferStatus = 'DRAFT' | 'DISPATCHED' | 'RECEIVED' | 'CANCELLED'

export interface StockTransferItem {
  id: string
  stock_transfer_id: string
  item_type: 'MATERIAL' | 'PRODUCT'
  material_id?: string
  product_id?: string
  product_size_id?: string
  qty: string
  unit_cost: string
}

export interface StockTransfer {
  id: string
  transfer_number: string
  source_warehouse_id: string
  destination_warehouse_id: string
  status: TransferStatus
  transfer_date: string
  notes: string
  dispatched_at?: string
  received_at?: string
  created_at: string
  updated_at: string
  items?: StockTransferItem[]
}

export interface CreateTransferItemInput {
  item_type: 'MATERIAL' | 'PRODUCT'
  material_id?: string
  product_id?: string
  product_size_id?: string
  qty: string
}

export interface CreateTransferInput {
  source_warehouse_id: string
  destination_warehouse_id: string
  transfer_date?: string
  notes?: string
  items: CreateTransferItemInput[]
}

// ---- Stock opname ----

export type OpnameStatus = 'DRAFT' | 'POSTED' | 'CANCELLED'

export interface StockOpnameItem {
  id: string
  stock_opname_id: string
  item_type: 'MATERIAL' | 'PRODUCT'
  material_id?: string
  product_id?: string
  product_size_id?: string
  system_qty: string
  actual_qty: string
  unit_cost: string
  variance_qty: string
  variance_amount: string
}

export interface StockOpname {
  id: string
  opname_number: string
  warehouse_id: string
  opname_date: string
  status: OpnameStatus
  notes: string
  journal_id?: string
  posted_at?: string
  created_at: string
  updated_at: string
  items?: StockOpnameItem[]
}

export interface CreateOpnameItemInput {
  item_type: 'MATERIAL' | 'PRODUCT'
  material_id?: string
  product_id?: string
  product_size_id?: string
  actual_qty: string
}

export interface CreateOpnameInput {
  warehouse_id: string
  opname_date?: string
  notes?: string
  items: CreateOpnameItemInput[]
}
