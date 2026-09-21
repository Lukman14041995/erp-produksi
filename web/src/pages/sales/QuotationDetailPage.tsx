import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, Check, X } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { StatusBadge } from '@/components/StatusBadge'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { useConfirmQuotation, useQuotation, useRejectQuotation } from '@/hooks/useQuotations'
import { useCustomers } from '@/hooks/useCustomers'
import { useFabrics, useGarmentSizes, useGarmentVariants, useInks, useProductTypes } from '@/hooks/useCatalog'
import { formatCurrency, formatDateTime } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'

export function QuotationDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: quotation, isLoading, error } = useQuotation(id)
  const { data: customers } = useCustomers()
  const { data: productTypes } = useProductTypes()
  const { data: fabrics } = useFabrics()
  const { data: variants } = useGarmentVariants()
  const { data: inks } = useInks()
  const { data: sizes } = useGarmentSizes()

  const confirmQuotation = useConfirmQuotation()
  const rejectQuotation = useRejectQuotation()
  const [taxRate, setTaxRate] = useState('0.11')
  const [rejectOpen, setRejectOpen] = useState(false)
  const [reason, setReason] = useState('')

  if (isLoading) return <LoadingState />
  if (error || !quotation) return <ErrorState message={(error as Error)?.message ?? 'Quotation tidak ditemukan'} />

  const customer = customers?.find((c) => c.id === quotation.customer_id)
  const jenisName = (tId: string) => productTypes?.find((t) => t.id === tId)?.name ?? tId.slice(0, 8)
  const fabricName = (fId: string) => fabrics?.find((f) => f.id === fId)?.name ?? fId.slice(0, 8)
  const sizeCode = (sId: string) => sizes?.find((s) => s.id === sId)?.size_code ?? sId.slice(0, 8)
  const variantName = (vId: string) => variants?.find((v) => v.id === vId)?.name ?? vId.slice(0, 8)
  const inkName = (iId: string) => inks?.find((i) => i.id === iId)?.name ?? iId.slice(0, 8)

  function confirm() {
    confirmQuotation
      .mutateAsync({ id: id!, taxRate })
      .then((so) => {
        toast.success('Quotation dikonfirmasi', `Pesanan Penjualan ${so.so_number} dibuat`)
        navigate(`/sales/orders/${so.id}`)
      })
      .catch((err) => toast.error('Gagal mengonfirmasi', err instanceof ApiError ? err.message : undefined))
  }

  function reject() {
    rejectQuotation
      .mutateAsync({ id: id!, reason })
      .then(() => {
        toast.success('Quotation ditolak')
        setRejectOpen(false)
      })
      .catch((err) => toast.error('Gagal menolak', err instanceof ApiError ? err.message : undefined))
  }

  return (
    <div>
      <button onClick={() => navigate('/sales/quotations')} className="mb-4 flex items-center gap-1 text-sm text-slate-500 hover:text-slate-800">
        <ArrowLeft className="h-4 w-4" /> Kembali ke Quotation
      </button>

      <PageHeader
        title={quotation.quotation_number}
        description={`Dari ${customer?.name ?? quotation.customer_id} · diterima ${formatDateTime(quotation.created_at)}`}
        actions={
          quotation.status === 'PENDING_REVIEW' && (
            <div className="flex items-center gap-2">
              <Input
                type="number"
                step="0.0001"
                value={taxRate}
                onChange={(e) => setTaxRate(e.target.value)}
                className="w-24"
                title="Tarif pajak (contoh: 0.11 = 11%)"
              />
              <Button onClick={confirm} loading={confirmQuotation.isPending}>
                <Check className="h-4 w-4" /> Konfirmasi jadi SO
              </Button>
              <Button variant="destructive" onClick={() => setRejectOpen(true)}>
                <X className="h-4 w-4" /> Tolak
              </Button>
            </div>
          )
        }
      />

      <div className="mb-6 grid grid-cols-2 gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="py-4">
            <p className="text-xs uppercase text-slate-400">Status</p>
            <div className="mt-1.5">
              <StatusBadge kind="quotation" value={quotation.status} />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4">
            <p className="text-xs uppercase text-slate-400">Estimasi Subtotal</p>
            <p className="mt-1.5 text-sm font-semibold">{formatCurrency(quotation.estimated_subtotal)}</p>
          </CardContent>
        </Card>
        {quotation.sales_order_id && (
          <Card>
            <CardContent className="py-4">
              <p className="text-xs uppercase text-slate-400">Pesanan Penjualan</p>
              <button
                onClick={() => navigate(`/sales/orders/${quotation.sales_order_id}`)}
                className="mt-1.5 text-sm font-medium text-[var(--color-primary)] hover:underline"
              >
                Lihat SO &rarr;
              </button>
            </CardContent>
          </Card>
        )}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Item Pesanan</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead className="bg-slate-50 text-left text-xs font-medium uppercase text-slate-500">
              <tr>
                <th className="px-4 py-2.5">Jenis</th>
                <th className="px-4 py-2.5">Model</th>
                <th className="px-4 py-2.5">Bahan</th>
                <th className="px-4 py-2.5">Tinta</th>
                <th className="px-4 py-2.5">Ukuran</th>
                <th className="px-4 py-2.5">Jml</th>
                <th className="px-4 py-2.5">Harga Satuan</th>
                <th className="px-4 py-2.5 text-right">Total Baris</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {quotation.items?.map((it) => (
                <tr key={it.id}>
                  <td className="px-4 py-2.5">{jenisName(it.product_type_id)}</td>
                  <td className="px-4 py-2.5">{variantName(it.variant_id)}</td>
                  <td className="px-4 py-2.5">{fabricName(it.fabric_id)}</td>
                  <td className="px-4 py-2.5">{inkName(it.ink_id)}</td>
                  <td className="px-4 py-2.5">{sizeCode(it.garment_size_id)}</td>
                  <td className="px-4 py-2.5">{it.qty}</td>
                  <td className="px-4 py-2.5">{formatCurrency(it.unit_price)}</td>
                  <td className="px-4 py-2.5 text-right font-medium">{formatCurrency(it.line_total)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>

      {quotation.notes && (
        <Card className="mt-4">
          <CardHeader>
            <CardTitle>Catatan Customer</CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-slate-600">{quotation.notes}</CardContent>
        </Card>
      )}

      <Dialog open={rejectOpen} onOpenChange={setRejectOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Tolak Quotation</DialogTitle>
          </DialogHeader>
          <div className="space-y-1.5">
            <Label>Alasan</Label>
            <Input value={reason} onChange={(e) => setReason(e.target.value)} placeholder="Contoh: duplikat, salah pilih ukuran" />
          </div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setRejectOpen(false)}>
              Batal
            </Button>
            <Button variant="destructive" onClick={reject} loading={rejectQuotation.isPending}>
              Tolak Quotation
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
