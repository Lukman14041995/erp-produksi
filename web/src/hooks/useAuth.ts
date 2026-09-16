import { useMutation } from '@tanstack/react-query'
import { apiPost } from '@/lib/api'
import { useAuthStore } from '@/store/authStore'
import type { LoginResponse } from '@/types/auth'
import { ApiError } from '@/types/api'

export function useCurrentUser() {
  return useAuthStore((s) => s.user)
}

export function useLogin() {
  const setSession = useAuthStore((s) => s.setSession)
  return useMutation({
    mutationFn: (input: { email: string; password: string }) => apiPost<LoginResponse>('/auth/login', input),
    onSuccess: (data) => setSession(data.user, data.tokens),
  })
}

export function useLogout() {
  const logout = useAuthStore((s) => s.logout)
  return () => logout()
}

export function isApiError(err: unknown): err is ApiError {
  return err instanceof ApiError
}
