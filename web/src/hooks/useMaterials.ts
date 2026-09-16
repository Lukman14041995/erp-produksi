import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost, apiPut } from '@/lib/api'
import type { Material } from '@/types/master'

export function useMaterials() {
  return useQuery({
    queryKey: ['materials'],
    queryFn: () => apiGet<Material[]>('/materials', { limit: 200 }),
  })
}

export function useCreateMaterial() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: Partial<Material>) => apiPost<Material>('/materials', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['materials'] }),
  })
}

export function useUpdateMaterial() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: Partial<Material> }) => apiPut<Material>(`/materials/${id}`, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['materials'] }),
  })
}
