export interface Customer {
  id: string
  code: string
  name: string
  contact_person: string
  phone: string
  email: string
  address: string
  tax_id: string
  credit_limit: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface Supplier {
  id: string
  code: string
  name: string
  contact_person: string
  phone: string
  email: string
  address: string
  tax_id: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface ProductSize {
  id: string
  product_id: string
  size_code: string
  size_multiplier: string
  sort_order: number
  is_active: boolean
}

export interface Product {
  id: string
  code: string
  name: string
  category: string
  uom: string
  base_price: string
  is_active: boolean
  created_at: string
  updated_at: string
  sizes?: ProductSize[]
}

export interface Material {
  id: string
  code: string
  name: string
  uom: string
  unit_cost: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export type AccountType = 'ASSET' | 'LIABILITY' | 'EQUITY' | 'REVENUE' | 'COGS' | 'EXPENSE'
export type NormalBalance = 'DEBIT' | 'CREDIT'

export interface Account {
  id: string
  code: string
  name: string
  account_type: AccountType
  normal_balance: NormalBalance
  parent_id?: string
  is_postable: boolean
  is_active: boolean
  created_at: string
  updated_at: string
}
