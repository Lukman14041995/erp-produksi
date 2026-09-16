import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost } from '@/lib/api'
import type { Role, User } from '@/types/auth'

export function useUsers() {
  return useQuery({
    queryKey: ['users'],
    queryFn: () => apiGet<User[]>('/users'),
  })
}

export interface RegisterUserInput {
  email: string
  password: string
  name: string
  role: Role
}

export function useRegisterUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: RegisterUserInput) => apiPost<User>('/auth/register', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] }),
  })
}
