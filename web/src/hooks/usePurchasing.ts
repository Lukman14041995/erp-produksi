import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost } from '@/lib/api'
import type {
  CreateBillFromGRNInput,
  CreateBillInput,
  CreateGRNInput,
  CreatePOInput,
  GoodsReceipt,
  PurchaseOrder,
  SupplierInvoice,
} from '@/types/purchasing'

export function useSupplierInvoices() {
  return useQuery({
    queryKey: ['supplier-invoices'],
    queryFn: () => apiGet<SupplierInvoice[]>('/supplier-invoices', { limit: 200 }),
  })
}

export function useSupplierInvoice(id: string | undefined) {
  return useQuery({
    queryKey: ['supplier-invoices', id],
    queryFn: () => apiGet<SupplierInvoice>(`/supplier-invoices/${id}`),
    enabled: !!id,
  })
}

export function useCreateSupplierInvoice() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateBillInput) => apiPost<SupplierInvoice>('/supplier-invoices', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['supplier-invoices'] }),
  })
}

function useBillCommand(command: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiPost<SupplierInvoice>(`/supplier-invoices/${id}/${command}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['supplier-invoices'] })
      qc.invalidateQueries({ queryKey: ['inventory-balances'] })
      qc.invalidateQueries({ queryKey: ['journals'] })
    },
  })
}

export const usePostBill = () => useBillCommand('post')
export const useVoidBill = () => useBillCommand('void')

// ---- Purchase Orders / 3-way matching ----

export function usePurchaseOrders() {
  return useQuery({
    queryKey: ['purchase-orders'],
    queryFn: () => apiGet<PurchaseOrder[]>('/purchase-orders', { limit: 200 }),
  })
}

export function usePurchaseOrder(id: string | undefined) {
  return useQuery({
    queryKey: ['purchase-orders', id],
    queryFn: () => apiGet<PurchaseOrder>(`/purchase-orders/${id}`),
    enabled: !!id,
  })
}

export function useCreatePurchaseOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreatePOInput) => apiPost<PurchaseOrder>('/purchase-orders', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
  })
}

function usePOCommand(command: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiPost<PurchaseOrder>(`/purchase-orders/${id}/${command}`),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ['purchase-orders'] })
      qc.invalidateQueries({ queryKey: ['purchase-orders', id] })
    },
  })
}

export const useApprovePO = () => usePOCommand('approve')
export const useCancelPO = () => usePOCommand('cancel')

export function useGoodsReceiptsByPO(poId: string | undefined) {
  return useQuery({
    queryKey: ['goods-receipts', 'by-po', poId],
    queryFn: () => apiGet<GoodsReceipt[]>(`/purchase-orders/${poId}/goods-receipts`),
    enabled: !!poId,
  })
}

export function useGoodsReceipts() {
  return useQuery({
    queryKey: ['goods-receipts'],
    queryFn: () => apiGet<GoodsReceipt[]>('/goods-receipts', { limit: 200 }),
  })
}

export function useGoodsReceipt(id: string | undefined) {
  return useQuery({
    queryKey: ['goods-receipts', id],
    queryFn: () => apiGet<GoodsReceipt>(`/goods-receipts/${id}`),
    enabled: !!id,
  })
}

export function useCreateGoodsReceipt() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateGRNInput) => apiPost<GoodsReceipt>('/goods-receipts', input),
    onSuccess: (_, input) => {
      qc.invalidateQueries({ queryKey: ['goods-receipts'] })
      qc.invalidateQueries({ queryKey: ['goods-receipts', 'by-po', input.purchase_order_id] })
      qc.invalidateQueries({ queryKey: ['purchase-orders'] })
      qc.invalidateQueries({ queryKey: ['purchase-orders', input.purchase_order_id] })
      qc.invalidateQueries({ queryKey: ['inventory-balances'] })
      qc.invalidateQueries({ queryKey: ['journals'] })
    },
  })
}

export function useCreateBillFromGRN() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateBillFromGRNInput) => apiPost<SupplierInvoice>('/supplier-invoices/from-grn', input),
    onSuccess: (_, input) => {
      qc.invalidateQueries({ queryKey: ['supplier-invoices'] })
      qc.invalidateQueries({ queryKey: ['purchase-orders'] })
      qc.invalidateQueries({ queryKey: ['purchase-orders', input.purchase_order_id] })
      qc.invalidateQueries({ queryKey: ['goods-receipts', 'by-po', input.purchase_order_id] })
      qc.invalidateQueries({ queryKey: ['journals'] })
    },
  })
}
