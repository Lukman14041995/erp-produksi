import { create } from 'zustand'
import type { TokenPair, User } from '@/types/auth'

const STORAGE_KEY = 'clothing-erp-auth'

interface StoredAuth {
  user: User
  tokens: TokenPair
}

function loadStored(): StoredAuth | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? (JSON.parse(raw) as StoredAuth) : null
  } catch {
    return null
  }
}

function persist(auth: StoredAuth | null) {
  try {
    if (auth) localStorage.setItem(STORAGE_KEY, JSON.stringify(auth))
    else localStorage.removeItem(STORAGE_KEY)
  } catch {
    // localStorage unavailable (private mode, etc.) -- session simply won't survive a reload
  }
}

interface AuthState {
  user: User | null
  tokens: TokenPair | null
  setSession: (user: User, tokens: TokenPair) => void
  updateTokens: (tokens: TokenPair) => void
  logout: () => void
}

const stored = loadStored()

export const useAuthStore = create<AuthState>((set, get) => ({
  user: stored?.user ?? null,
  tokens: stored?.tokens ?? null,
  setSession: (user, tokens) => {
    persist({ user, tokens })
    set({ user, tokens })
  },
  updateTokens: (tokens) => {
    const user = get().user
    if (user) persist({ user, tokens })
    set({ tokens })
  },
  logout: () => {
    persist(null)
    set({ user: null, tokens: null })
  },
}))

export function getAccessToken(): string | null {
  return useAuthStore.getState().tokens?.access_token ?? null
}

export function getRefreshToken(): string | null {
  return useAuthStore.getState().tokens?.refresh_token ?? null
}
