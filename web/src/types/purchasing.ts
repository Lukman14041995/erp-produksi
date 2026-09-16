export type BillStatus = 'DRAFT' | 'POSTED' | 'PARTIALLY_PAID' | 'PAID' | 'VOID'

export interface SupplierInvoiceItem {
  id: string
  supplier_invoice_id: string
  material_id: string
  qty: string
  unit_cost: string
  line_total: string
  purchase_order_item_id?: string
  goods_receipt_item_id?: string
  price_variance: string
}

export interface SupplierInvoice {
  id: string
  bill_number: string
  supplier_id: string
  purchase_order_id?: string
  bill_date: string
  due_date?: string
  debit_account_code: string
  status: BillStatus
  subtotal: string
  tax_total: string
  grand_total: string
  paid_amount: string
  balance_due: string
  journal_id?: string
  notes: string
  created_at: string
  updated_at: string
  items?: SupplierInvoiceItem[]
}

export interface CreateBillItemInput {
  material_id: string
  qty: string
  unit_cost: string
}

export interface CreateBillInput {
  supplier_id: string
  bill_date?: string
  due_date?: string
  debit_account_code: string
  tax_rate: string
  notes?: string
  items: CreateBillItemInput[]
}

// ---- Purchase Orders / 3-way matching ----

export type POStatus = 'DRAFT' | 'APPROVED' | 'PARTIALLY_RECEIVED' | 'FULLY_RECEIVED' | 'CLOSED' | 'CANCELLED'
export type POBillingStatus = 'UNBILLED' | 'PARTIALLY_BILLED' | 'FULLY_BILLED'

export interface PurchaseOrderItem {
  id: string
  purchase_order_id: string
  material_id: string
  qty: string
  unit_cost: string
  qty_received: string
  qty_billed: string
  line_total: string
}

export interface PurchaseOrder {
  id: string
  po_number: string
  supplier_id: string
  order_date: string
  expected_date?: string
  status: POStatus
  billing_status: POBillingStatus
  subtotal: string
  grand_total: string
  notes: string
  approved_at?: string
  cancelled_at?: string
  created_at: string
  updated_at: string
  items?: PurchaseOrderItem[]
}

export interface CreatePOItemInput {
  material_id: string
  qty: string
  unit_cost: string
}

export interface CreatePOInput {
  supplier_id: string
  order_date?: string
  expected_date?: string
  notes?: string
  items: CreatePOItemInput[]
}

export interface GoodsReceiptItem {
  id: string
  goods_receipt_id: string
  purchase_order_item_id: string
  material_id: string
  qty_received: string
  unit_cost: string
  qty_billed: string
  line_total: string
}

export interface GoodsReceipt {
  id: string
  grn_number: string
  purchase_order_id: string
  supplier_id: string
  receipt_date: string
  journal_id?: string
  notes: string
  created_at: string
  items?: GoodsReceiptItem[]
}

export interface CreateGRNItemInput {
  purchase_order_item_id: string
  qty_received: string
}

export interface CreateGRNInput {
  purchase_order_id: string
  receipt_date?: string
  notes?: string
  items: CreateGRNItemInput[]
}

export interface BillFromGRNItemInput {
  goods_receipt_item_id: string
  qty: string
  bill_unit_cost: string
}

export interface CreateBillFromGRNInput {
  purchase_order_id: string
  bill_date?: string
  due_date?: string
  notes?: string
  items: BillFromGRNItemInput[]
}
