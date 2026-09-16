export type ProdStatus = 'NOT_STARTED' | 'IN_PROGRESS' | 'COMPLETED' | 'CANCELLED'

export interface BOMLine {
  id: string
  bom_id: string
  material_id: string
  qty_per_unit: string
  uom: string
}

export interface BOMHeader {
  id: string
  product_id: string
  name: string
  version: number
  is_active: boolean
  lines?: BOMLine[]
}

export interface CreateBOMLineInput {
  material_id: string
  qty_per_unit: string
  uom: string
}

export interface CreateBOMInput {
  product_id: string
  name: string
  lines: CreateBOMLineInput[]
}

export interface ProductionOrderItem {
  id: string
  production_order_id: string
  product_size_id: string
  planned_qty: string
  finished_qty: string
}

export interface ProductionMaterial {
  id: string
  production_order_id: string
  material_id: string
  planned_qty: string
  issued_qty: string
  unit_cost: string
  total_cost: string
  issued_at?: string
}

export interface ProductionLabor {
  id: string
  production_order_id: string
  description: string
  hours: string
  rate: string
  total_cost: string
}

export interface ProductionOverhead {
  id: string
  production_order_id: string
  description: string
  allocation_basis: string
  amount: string
}

export interface ProductionOrder {
  id: string
  prod_number: string
  sales_order_id?: string
  product_id: string
  bom_id?: string
  planned_qty: string
  finished_qty: string
  production_status: ProdStatus
  start_date?: string
  end_date?: string
  notes: string
  created_at: string
  updated_at: string
  items?: ProductionOrderItem[]
  materials?: ProductionMaterial[]
  labor?: ProductionLabor[]
  overheads?: ProductionOverhead[]
}

export interface CreateProductionOrderItemInput {
  product_size_id: string
  planned_qty: string
}

export interface CreateProductionOrderInput {
  sales_order_id?: string
  product_id: string
  bom_id?: string
  notes?: string
  items: CreateProductionOrderItemInput[]
}

export interface ProductionCost {
  id: string
  production_order_id: string
  total_material_cost: string
  total_labor_cost: string
  total_overhead_cost: string
  total_cost: string
  finished_qty: string
  unit_cost: string
  journal_id?: string
  computed_at: string
}
