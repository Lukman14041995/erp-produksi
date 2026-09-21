export type SpkStatus = 'NOT_STARTED' | 'IN_PROGRESS' | 'COMPLETED'
export type SpkStageStatus = 'PENDING' | 'IN_PROGRESS' | 'DONE'

export interface SpkItem {
  id: string
  spk_order_id: string
  sales_order_item_id: string
  product_id: string
  product_size_id: string
  qty: string
  fabric_name: string
  variant_name: string
  ink_name: string
  size_code: string
  fabric_material_id?: string
  ink_material_id?: string
}

export interface SpkStageLog {
  id: string
  spk_stage_id: string
  qty: string
  note: string
  logged_by?: string
  logged_by_name?: string
  logged_at: string
}

export interface SpkMaterialUsage {
  id: string
  spk_order_id: string
  spk_stage_id?: string
  material_id: string
  material_name: string
  material_uom: string
  qty: string
  unit_cost: string
  total_cost: string
  warehouse_id: string
  inventory_txn_id?: string
  journal_id?: string
  note: string
  logged_by?: string
  logged_by_name?: string
  logged_at: string
}

export interface IssueMaterialInput {
  spk_stage_id?: string
  material_id: string
  qty: string
  note?: string
}

export interface SpkLabor {
  id: string
  spk_order_id: string
  spk_stage_id?: string
  description: string
  hours: string
  rate: string
  total_cost: string
  journal_id?: string
  logged_by?: string
  logged_by_name?: string
  logged_at: string
}

export interface AddLaborInput {
  spk_stage_id?: string
  description: string
  hours: string
  rate: string
}

export interface SpkOverhead {
  id: string
  spk_order_id: string
  description: string
  allocation_basis: string
  amount: string
  journal_id?: string
  logged_by?: string
  logged_by_name?: string
  logged_at: string
}

export interface AddOverheadInput {
  description: string
  allocation_basis?: string
  amount: string
}

export interface SpkStage {
  id: string
  spk_order_id: string
  stage_seq: number
  stage_code: string
  stage_name: string
  requires_qty: boolean
  planned_qty: string
  completed_qty: string
  status: SpkStageStatus
  started_at?: string
  completed_at?: string
  logs?: SpkStageLog[]
}

export interface SpkOrder {
  id: string
  spk_number: string
  sales_order_id: string
  so_number: string
  quotation_id?: string
  quotation_number: string
  customer_name: string
  product_type_id: string
  product_type_code: string
  product_type_name: string
  composition_summary: string
  total_qty: string
  status: SpkStatus
  current_stage_code: string
  current_stage_name: string
  progress_pct: string
  design_done: boolean
  production_done: boolean
  shipped: boolean
  material_cost_total: string
  labor_cost_total: string
  overhead_cost_total: string
  notes: string
  completed_at?: string
  created_at: string
  updated_at: string
  items?: SpkItem[]
  stages?: SpkStage[]
  material_usages?: SpkMaterialUsage[]
  labor?: SpkLabor[]
  overheads?: SpkOverhead[]
}

export interface LogStageInput {
  qty: string
  note?: string
}

export interface MarkMilestoneInput {
  note?: string
}
