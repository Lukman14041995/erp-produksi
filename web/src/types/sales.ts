export type OrderStatus = 'DRAFT' | 'CONFIRMED' | 'CANCELLED' | 'CLOSED'
export type PaymentStatus = 'UNPAID' | 'PARTIAL' | 'PAID' | 'OVERPAID'
export type ProductionStatus = 'NOT_STARTED' | 'IN_PROGRESS' | 'COMPLETED' | 'CANCELLED'
export type DeliveryStatus = 'NOT_DELIVERED' | 'PARTIAL' | 'DELIVERED'
export type InvoiceStatus = 'DRAFT' | 'POSTED' | 'PARTIALLY_PAID' | 'PAID' | 'VOID'

export interface SalesOrderItem {
  id: string
  sales_order_id: string
  product_id: string
  product_size_id: string
  qty: string
  unit_price: string
  discount: string
  tax_rate: string
  line_total: string
  quotation_item_id?: string
}

export interface SalesOrder {
  id: string
  so_number: string
  customer_id: string
  order_date: string
  order_status: OrderStatus
  payment_status: PaymentStatus
  production_status: ProductionStatus
  delivery_status: DeliveryStatus
  subtotal: string
  discount_total: string
  tax_total: string
  grand_total: string
  notes: string
  created_by: string
  created_at: string
  updated_at: string
  items?: SalesOrderItem[]
}

export interface CreateOrderItemInput {
  product_id: string
  product_size_id: string
  qty: string
  unit_price: string
  discount: string
  tax_rate: string
}

export interface CreateOrderInput {
  customer_id: string
  order_date?: string
  notes?: string
  items: CreateOrderItemInput[]
}

export interface InvoiceItem {
  id: string
  invoice_id: string
  sales_order_item_id: string
  product_id: string
  product_size_id: string
  qty: string
  unit_price: string
  discount: string
  tax_rate: string
  line_total: string
}

export interface Invoice {
  id: string
  invoice_number: string
  sales_order_id: string
  customer_id: string
  invoice_date: string
  due_date?: string
  status: InvoiceStatus
  subtotal: string
  discount_total: string
  tax_total: string
  grand_total: string
  paid_amount: string
  balance_due: string
  journal_id?: string
  created_at: string
  updated_at: string
  items?: InvoiceItem[]
}
