export type Role = 'ADMIN' | 'SALES' | 'PRODUCTION' | 'FINANCE' | 'ACCOUNTING'

export interface User {
  id: string
  email: string
  name: string
  role: Role
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  access_expires_at: string
  refresh_expires_at: string
}

export interface LoginResponse {
  tokens: TokenPair
  user: User
}
