export interface ProductType {
  id: string
  code: string
  name: string
  sort_order: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface Fabric {
  id: string
  product_type_id: string
  code: string
  material_id: string
  name: string
  sales_price: string
  image_url: string
  sort_order: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface GarmentVariant {
  id: string
  product_type_id: string
  code: string
  name: string
  price_addon: string
  icon_url: string
  sort_order: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface GarmentSize {
  id: string
  size_code: string
  size_multiplier: string
  sort_order: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface Ink {
  id: string
  product_type_id: string
  material_id: string
  code: string
  name: string
  price_addon: string
  image_url: string
  sort_order: number
  is_active: boolean
  created_at: string
  updated_at: string
}
