import type { Fabric, GarmentSize, GarmentVariant, Ink, ProductType } from '@/types/catalog'
import type { PaymentPlan } from '@/types/billing'
import type { Invoice, SalesOrder } from '@/types/sales'
import type { SpkOrder } from '@/types/spk'

export type LinkStatus = 'ACTIVE' | 'REVOKED'
export type QuotationStatus = 'PENDING_REVIEW' | 'CONFIRMED' | 'REJECTED' | 'CANCELLED'

export interface OrderLink {
  id: string
  customer_id: string
  created_by: string
  status: LinkStatus
  expires_at: string
  created_at: string
}

export interface CreateLinkOutput extends OrderLink {
  token: string
}

export interface ItemInput {
  product_type_id: string
  fabric_id: string
  garment_size_id: string
  variant_id: string
  ink_id: string
  qty: string
}

export interface QuotationItem extends ItemInput {
  id: string
  quotation_id: string
  unit_price: string
  line_total: string
  created_at: string
}

export interface Quotation {
  id: string
  quotation_number: string
  order_link_id: string
  customer_id: string
  status: QuotationStatus
  estimated_subtotal: string
  notes: string
  reviewed_by?: string
  reviewed_at?: string
  sales_order_id?: string
  created_at: string
  updated_at: string
  items?: QuotationItem[]
}

export interface EstimateLine extends ItemInput {
  unit_price: string
  line_total: string
}

export interface EstimateOutput {
  lines: EstimateLine[]
  subtotal: string
}

export interface PublicOrderDetail {
  quotation: Quotation
  sales_order?: SalesOrder
  invoice?: Invoice
  payment_plan?: PaymentPlan
  spk_orders?: SpkOrder[]
}

export interface PublicCatalog {
  customer_name: string
  product_types: ProductType[]
  fabrics: Fabric[]
  variants: GarmentVariant[]
  inks: Ink[]
  sizes: GarmentSize[]
}
