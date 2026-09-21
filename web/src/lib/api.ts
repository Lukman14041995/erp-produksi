import axios, { type InternalAxiosRequestConfig } from 'axios'
import { getAccessToken, getRefreshToken, useAuthStore } from '@/store/authStore'
import type { Envelope } from '@/types/api'
import { ApiError } from '@/types/api'
import type { TokenPair } from '@/types/auth'

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api/v1',
})

let refreshPromise: Promise<string | null> | null = null

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) return null

  try {
    const res = await axios.post<Envelope<{ tokens: TokenPair }>>(
      `${apiClient.defaults.baseURL}/auth/refresh`,
      { refresh_token: refreshToken },
    )
    const tokens = res.data.data?.tokens
    if (!tokens) return null
    useAuthStore.getState().updateTokens(tokens)
    return tokens.access_token
  } catch {
    useAuthStore.getState().logout()
    return null
  }
}

function isExpiringSoon(expiresAt: string, marginSeconds = 30): boolean {
  const expiryMs = new Date(expiresAt).getTime()
  return expiryMs - Date.now() < marginSeconds * 1000
}

apiClient.interceptors.request.use(async (config) => {
  const { tokens } = useAuthStore.getState()
  let token = getAccessToken()

  if (tokens && isExpiringSoon(tokens.access_expires_at)) {
    refreshPromise ??= refreshAccessToken().finally(() => {
      refreshPromise = null
    })
    token = await refreshPromise
  }

  if (token) {
    config.headers.set('Authorization', `Bearer ${token}`)
  }
  return config
})

interface RetryableConfig extends InternalAxiosRequestConfig {
  _retried?: boolean
}

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const config = error.config as RetryableConfig | undefined
    const status = error.response?.status

    // Fallback safety net: if the proactive refresh above missed the window
    // (clock skew, long-lived request) and the server rejected with 403,
    // try one reactive refresh-and-retry before giving up.
    if (status === 403 && config && !config._retried) {
      config._retried = true
      refreshPromise ??= refreshAccessToken().finally(() => {
        refreshPromise = null
      })
      const token = await refreshPromise
      if (token) {
        config.headers = config.headers ?? {}
        config.headers.Authorization = `Bearer ${token}`
        return apiClient.request(config)
      }
    }

    const envelope = error.response?.data as Envelope<unknown> | undefined
    const message = envelope?.message ?? error.message ?? 'Permintaan gagal'
    return Promise.reject(new ApiError(message, status, envelope?.errors))
  },
)

export async function apiGet<T>(url: string, params?: Record<string, unknown>): Promise<T> {
  const res = await apiClient.get<Envelope<T>>(url, { params })
  return res.data.data as T
}

export async function apiPost<T>(url: string, body?: unknown): Promise<T> {
  const res = await apiClient.post<Envelope<T>>(url, body)
  return res.data.data as T
}

export async function apiPut<T>(url: string, body?: unknown): Promise<T> {
  const res = await apiClient.put<Envelope<T>>(url, body)
  return res.data.data as T
}

export async function apiDelete<T>(url: string): Promise<T> {
  const res = await apiClient.delete<Envelope<T>>(url)
  return res.data.data as T
}

// apiUpload sends a multipart/form-data body (a single file under the given
// field name) -- axios sets the boundary header itself from the FormData
// instance, so this must not set Content-Type manually.
export async function apiUpload<T>(url: string, fieldName: string, file: File): Promise<T> {
  const form = new FormData()
  form.append(fieldName, file)
  const res = await apiClient.post<Envelope<T>>(url, form)
  return res.data.data as T
}
