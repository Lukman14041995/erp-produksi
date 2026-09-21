import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost, apiUpload } from '@/lib/api'
import type { SalesOrder } from '@/types/sales'
import type { ChoosePaymentPlanInput, Installment } from '@/types/billing'
import type { CreateLinkOutput, ItemInput, OrderLink, PublicCatalog, PublicOrderDetail, Quotation, QuotationStatus } from '@/types/quotation'

// ---- Staff: order links ----

export function useOrderLinks(customerId: string | undefined) {
  return useQuery({
    queryKey: ['order-links', customerId],
    queryFn: () => apiGet<OrderLink[]>('/order-links', { customer_id: customerId }),
    enabled: !!customerId,
  })
}

export function useCreateOrderLink() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { customer_id: string; expires_at?: string }) => apiPost<CreateLinkOutput>('/order-links', input),
    onSuccess: (_, { customer_id }) => qc.invalidateQueries({ queryKey: ['order-links', customer_id] }),
  })
}

export function useRevokeOrderLink() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiPost(`/order-links/${id}/revoke`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['order-links'] }),
  })
}

// ---- Staff: quotation review ----

export function useQuotations(status?: QuotationStatus) {
  return useQuery({
    queryKey: ['quotations', status ?? 'ALL'],
    queryFn: () => apiGet<Quotation[]>('/quotations', { status, limit: 200 }),
  })
}

export function useQuotation(id: string | undefined) {
  return useQuery({
    queryKey: ['quotations', 'detail', id],
    queryFn: () => apiGet<Quotation>(`/quotations/${id}`),
    enabled: !!id,
  })
}

export function useConfirmQuotation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, taxRate }: { id: string; taxRate: string }) =>
      apiPost<SalesOrder>(`/quotations/${id}/confirm`, { tax_rate: taxRate }),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['quotations'] })
      qc.invalidateQueries({ queryKey: ['quotations', 'detail', id] })
      qc.invalidateQueries({ queryKey: ['sales-orders'] })
    },
  })
}

export function useRejectQuotation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) => apiPost<Quotation>(`/quotations/${id}/reject`, { reason }),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['quotations'] })
      qc.invalidateQueries({ queryKey: ['quotations', 'detail', id] })
    },
  })
}

// useQuotationBySalesOrder powers the admin SO detail page's enriched item
// table (Jenis/Model/Bahan/Tinta/Ukuran instead of generic Produk/Ukuran)
// for orders that came from the customer configurator. Returns null for
// orders created the old manual way.
export function useQuotationBySalesOrder(salesOrderId: string | undefined) {
  return useQuery({
    queryKey: ['quotations', 'by-sales-order', salesOrderId],
    queryFn: () => apiGet<Quotation | null>(`/quotations/by-sales-order/${salesOrderId}`),
    enabled: !!salesOrderId,
  })
}

// ---- Public: customer configurator (no auth) ----

export function usePublicCatalog(token: string | undefined) {
  return useQuery({
    queryKey: ['public-order-link', token],
    queryFn: () => apiGet<PublicCatalog>(`/public/order-links/${token}`),
    enabled: !!token,
    retry: false,
  })
}

export function useSubmitQuotation(token: string | undefined) {
  return useMutation({
    mutationFn: (input: { notes: string; items: ItemInput[] }) =>
      apiPost<Quotation>(`/public/order-links/${token}/quotations`, input),
  })
}

export function useCustomerOrders(token: string | undefined) {
  return useQuery({
    queryKey: ['public-orders', token],
    queryFn: () => apiGet<Quotation[]>(`/public/order-links/${token}/quotations`),
    enabled: !!token,
  })
}

export function usePublicOrderDetail(token: string | undefined, quotationId: string | undefined) {
  return useQuery({
    queryKey: ['public-orders', token, quotationId],
    queryFn: () => apiGet<PublicOrderDetail>(`/public/order-links/${token}/quotations/${quotationId}`),
    enabled: !!token && !!quotationId,
  })
}

export function useChoosePaymentPlan(token: string | undefined) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ invoiceId, input }: { invoiceId: string; input: ChoosePaymentPlanInput }) =>
      apiPost(`/public/order-links/${token}/invoices/${invoiceId}/payment-plan`, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['public-orders', token] }),
  })
}

export function useUploadInstallmentProof(token: string | undefined) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ installmentId, file }: { installmentId: string; file: File }) =>
      apiUpload<Installment>(`/public/order-links/${token}/installments/${installmentId}/proof`, 'proof', file),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['public-orders', token] }),
  })
}
