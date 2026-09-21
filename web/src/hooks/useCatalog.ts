import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost, apiPut, apiUpload } from '@/lib/api'
import type { Fabric, GarmentSize, GarmentVariant, Ink, ProductType } from '@/types/catalog'

// ---- Product types (Jenis Pesanan) ----

export function useProductTypes() {
  return useQuery({
    queryKey: ['product-types'],
    queryFn: () => apiGet<ProductType[]>('/catalog/product-types'),
  })
}

export function useCreateProductType() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: Partial<ProductType>) => apiPost<ProductType>('/catalog/product-types', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['product-types'] }),
  })
}

export function useUpdateProductType() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: Partial<ProductType> }) =>
      apiPut<ProductType>(`/catalog/product-types/${id}`, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['product-types'] }),
  })
}

// ---- Fabrics ----

export function useFabrics() {
  return useQuery({
    queryKey: ['fabrics'],
    queryFn: () => apiGet<Fabric[]>('/catalog/fabrics'),
  })
}

export function useCreateFabric() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: Partial<Fabric>) => apiPost<Fabric>('/catalog/fabrics', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['fabrics'] }),
  })
}

export function useUpdateFabric() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: Partial<Fabric> }) => apiPut<Fabric>(`/catalog/fabrics/${id}`, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['fabrics'] }),
  })
}

export function useUploadFabricImage() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, file }: { id: string; file: File }) => apiUpload<Fabric>(`/catalog/fabrics/${id}/image`, 'image', file),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['fabrics'] }),
  })
}

// ---- Garment variants (Model potongan) ----

export function useGarmentVariants() {
  return useQuery({
    queryKey: ['garment-variants'],
    queryFn: () => apiGet<GarmentVariant[]>('/catalog/variants'),
  })
}

export function useCreateGarmentVariant() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: Partial<GarmentVariant>) => apiPost<GarmentVariant>('/catalog/variants', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['garment-variants'] }),
  })
}

export function useUpdateGarmentVariant() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: Partial<GarmentVariant> }) =>
      apiPut<GarmentVariant>(`/catalog/variants/${id}`, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['garment-variants'] }),
  })
}

export function useUploadVariantIcon() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, file }: { id: string; file: File }) => apiUpload<GarmentVariant>(`/catalog/variants/${id}/icon`, 'icon', file),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['garment-variants'] }),
  })
}

// ---- Garment sizes ----

export function useGarmentSizes() {
  return useQuery({
    queryKey: ['garment-sizes'],
    queryFn: () => apiGet<GarmentSize[]>('/catalog/sizes'),
  })
}

export function useCreateGarmentSize() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: Partial<GarmentSize>) => apiPost<GarmentSize>('/catalog/sizes', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['garment-sizes'] }),
  })
}

export function useUpdateGarmentSize() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: Partial<GarmentSize> }) =>
      apiPut<GarmentSize>(`/catalog/sizes/${id}`, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['garment-sizes'] }),
  })
}

// ---- Inks (Tinta) ----

export function useInks() {
  return useQuery({
    queryKey: ['inks'],
    queryFn: () => apiGet<Ink[]>('/catalog/inks'),
  })
}

export function useCreateInk() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: Partial<Ink>) => apiPost<Ink>('/catalog/inks', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['inks'] }),
  })
}

export function useUpdateInk() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: Partial<Ink> }) => apiPut<Ink>(`/catalog/inks/${id}`, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['inks'] }),
  })
}

export function useUploadInkImage() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, file }: { id: string; file: File }) => apiUpload<Ink>(`/catalog/inks/${id}/image`, 'image', file),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['inks'] }),
  })
}
