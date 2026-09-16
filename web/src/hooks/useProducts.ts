import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost, apiPut } from '@/lib/api'
import type { Product, ProductSize } from '@/types/master'

export function useProducts() {
  return useQuery({
    queryKey: ['products'],
    queryFn: () => apiGet<Product[]>('/products', { limit: 200 }),
  })
}

export function useProduct(id: string | undefined) {
  return useQuery({
    queryKey: ['products', id],
    queryFn: () => apiGet<Product>(`/products/${id}`),
    enabled: !!id,
  })
}

export function useCreateProduct() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: Partial<Product>) => apiPost<Product>('/products', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['products'] }),
  })
}

export function useUpdateProduct() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: Partial<Product> }) => apiPut<Product>(`/products/${id}`, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['products'] }),
  })
}

export function useAddProductSize() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ productId, input }: { productId: string; input: Partial<ProductSize> }) =>
      apiPost<ProductSize>(`/products/${productId}/sizes`, input),
    onSuccess: (_, { productId }) => {
      qc.invalidateQueries({ queryKey: ['products', productId] })
      qc.invalidateQueries({ queryKey: ['products'] })
    },
  })
}
