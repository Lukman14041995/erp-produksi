import { useRef, useState } from 'react'
import { ArrowLeft, Download, ImageIcon, Upload } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Badge, type BadgeTone } from '@/components/ui/Badge'
import { LoadingState } from '@/components/QueryState'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { useChoosePaymentPlan, usePublicCatalog, usePublicOrderDetail, useUploadInstallmentProof } from '@/hooks/useQuotations'
import { formatCurrency, toUploadUrl } from '@/lib/format'
import { openServerPdf } from '@/lib/pdf'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { PaymentMethod } from '@/types/billing'
import type { QuotationStatus } from '@/types/quotation'
import type { SpkOrder } from '@/types/spk'

const statusLabel: Record<QuotationStatus, { label: string; tone: BadgeTone }> = {
  PENDING_REVIEW: { label: 'Menunggu Review', tone: 'warning' },
  CONFIRMED: { label: 'Dikonfirmasi', tone: 'success' },
  REJECTED: { label: 'Ditolak', tone: 'danger' },
  CANCELLED: { label: 'Dibatalkan', tone: 'neutral' },
}

const installmentStatusLabel: Record<string, { label: string; tone: BadgeTone }> = {
  PENDING_PAYMENT: { label: 'Menunggu Transfer', tone: 'neutral' },
  PENDING_VERIFICATION: { label: 'Menunggu Verifikasi', tone: 'warning' },
  CONFIRMED: { label: 'Lunas', tone: 'success' },
  REJECTED: { label: 'Ditolak', tone: 'danger' },
}

function SpkProgressCard({ spkOrder }: { spkOrder: SpkOrder }) {
  const pct = Math.min(100, Math.round(Number(spkOrder.progress_pct)))
  return (
    <Card className="mb-4">
      <CardHeader>
        <CardTitle>
          Progress Produksi &middot; {spkOrder.product_type_name} ({spkOrder.total_qty} pcs)
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="mb-4 flex items-center justify-between text-sm">
          <span className="font-medium text-slate-700">{spkOrder.current_stage_name}</span>
          <span className="text-slate-500">{pct}%</span>
        </div>
        <div className="mb-5 h-2 w-full overflow-hidden rounded-full bg-slate-100">
          <div className="h-full rounded-full bg-[var(--color-primary)]" style={{ width: `${pct}%` }} />
        </div>
        <div className="flex flex-wrap gap-x-1 gap-y-2">
          {spkOrder.stages?.map((stage, i) => (
            <div key={stage.id} className="flex items-center">
              <div className="flex flex-col items-center gap-1">
                <div
                  className={`flex h-6 w-6 items-center justify-center rounded-full text-[10px] font-semibold ${
                    stage.status === 'DONE'
                      ? 'bg-[var(--color-primary)] text-white'
                      : stage.status === 'IN_PROGRESS'
                        ? 'border-2 border-[var(--color-primary)] text-[var(--color-primary)]'
                        : 'border border-slate-200 text-slate-300'
                  }`}
                >
                  {stage.status === 'DONE' ? '✓' : stage.stage_seq}
                </div>
                <span className="w-16 text-center text-[10px] leading-tight text-slate-500">{stage.stage_name}</span>
              </div>
              {i < (spkOrder.stages?.length ?? 0) - 1 && <div className="mx-1 h-px w-4 bg-slate-200" />}
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

export function OrderDetailView({ token, quotationId, onBack }: { token: string; quotationId: string; onBack: () => void }) {
  const { data, isLoading } = usePublicOrderDetail(token, quotationId)
  const { data: catalog } = usePublicCatalog(token)
  const choosePlan = useChoosePaymentPlan(token)
  const uploadProof = useUploadInstallmentProof(token)

  const [method, setMethod] = useState<PaymentMethod>('FULL')
  const [installmentCount, setInstallmentCount] = useState(2)
  const [uploadingFor, setUploadingFor] = useState<string | null>(null)
  const [proofPreview, setProofPreview] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const jenisName = (id: string) => catalog?.product_types.find((t) => t.id === id)?.name ?? '-'
  const fabricName = (id: string) => catalog?.fabrics.find((f) => f.id === id)?.name ?? '-'
  const variantName = (id: string) => catalog?.variants.find((v) => v.id === id)?.name ?? '-'
  const inkName = (id: string) => catalog?.inks.find((i) => i.id === id)?.name ?? '-'
  const sizeCode = (id: string) => catalog?.sizes.find((s) => s.id === id)?.size_code ?? '-'

  if (isLoading || !data) return <LoadingState />

  const { quotation, sales_order: salesOrder, invoice, payment_plan: plan, spk_orders: spkOrders } = data
  const status = statusLabel[quotation.status]

  function submitPlan() {
    if (!invoice) return
    choosePlan
      .mutateAsync({ invoiceId: invoice.id, input: { method, installment_count: method === 'FULL' ? 1 : installmentCount } })
      .then(() => toast.success('Metode pembayaran dipilih'))
      .catch((err) => toast.error('Gagal memilih metode pembayaran', err instanceof ApiError ? err.message : undefined))
  }

  function onPickProof(file: File | undefined) {
    if (!file || !uploadingFor) return
    uploadProof
      .mutateAsync({ installmentId: uploadingFor, file })
      .then(() => toast.success('Bukti transfer diunggah', 'Menunggu verifikasi dari kami'))
      .catch((err) => toast.error('Gagal mengunggah', err instanceof ApiError ? err.message : undefined))
      .finally(() => {
        setUploadingFor(null)
        if (fileInputRef.current) fileInputRef.current.value = ''
      })
  }

  async function downloadInvoice() {
    if (!invoice) return
    try {
      await openServerPdf(`/public/order-links/${token}/invoices/${invoice.id}/pdf`, `faktur-${invoice.invoice_number}.pdf`)
    } catch (err) {
      toast.error('Gagal mengunduh faktur', err instanceof ApiError ? err.message : undefined)
    }
  }

  return (
    <div>
      <button onClick={onBack} className="mb-4 flex items-center gap-1 text-sm text-slate-400 hover:text-slate-200">
        <ArrowLeft className="h-4 w-4" /> Kembali ke Pesanan Saya
      </button>

      <div className="mb-4 flex items-center justify-between">
        <div>
          <h2 className="text-base font-semibold text-slate-900">{quotation.quotation_number}</h2>
          <p className="text-xs text-slate-400">Estimasi {formatCurrency(quotation.estimated_subtotal)}</p>
        </div>
        <Badge tone={status.tone}>{status.label}</Badge>
      </div>

      <Card className="mb-4">
        <CardHeader>
          <CardTitle>Item Pesanan</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {quotation.items?.map((it) => (
            <div key={it.id} className="flex items-center justify-between rounded-md border border-slate-100 px-3 py-2 text-sm">
              <div>
                <p className="font-medium text-slate-800">
                  {jenisName(it.product_type_id)} &middot; {variantName(it.variant_id)} &middot; {fabricName(it.fabric_id)} &middot; {inkName(it.ink_id)}
                </p>
                <p className="text-xs text-slate-400">
                  {sizeCode(it.garment_size_id)} &middot; {it.qty} pcs &times; {formatCurrency(it.unit_price)}
                </p>
              </div>
              <span className="font-medium">{formatCurrency(it.line_total)}</span>
            </div>
          ))}
        </CardContent>
      </Card>

      {quotation.status === 'PENDING_REVIEW' && (
        <Card className="mb-4">
          <CardContent className="py-6 text-center text-sm text-slate-400">
            Pesanan Anda sedang direview oleh tim kami. Kami akan menghubungi Anda untuk konfirmasi.
          </CardContent>
        </Card>
      )}

      {quotation.status === 'REJECTED' && (
        <Card className="mb-4">
          <CardContent className="py-6 text-center text-sm text-[var(--color-danger)]">Maaf, pesanan ini ditolak oleh tim kami.</CardContent>
        </Card>
      )}

      {spkOrders?.map((spkOrder) => (
        <SpkProgressCard key={spkOrder.id} spkOrder={spkOrder} />
      ))}

      {salesOrder && invoice && (
        <Card className="mb-4">
          <CardHeader>
            <CardTitle>Pembayaran &middot; {salesOrder.so_number}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between rounded-md bg-slate-50 px-3 py-2 text-sm">
              <span className="text-slate-500">Total Tagihan</span>
              <span className="font-semibold">{formatCurrency(invoice.grand_total)}</span>
            </div>

            <Button variant="secondary" size="sm" onClick={downloadInvoice}>
              <Download className="h-4 w-4" /> Unduh Faktur
            </Button>

            {!plan && (
              <div className="space-y-3 rounded-md border border-slate-100 p-3">
                <p className="text-sm font-medium text-slate-700">Pilih Metode Pembayaran</p>
                <div className="flex gap-2">
                  <button
                    onClick={() => setMethod('FULL')}
                    className={`rounded-md border-2 px-4 py-2 text-sm ${method === 'FULL' ? 'border-[var(--color-primary)]' : 'border-slate-200'}`}
                  >
                    Lunas
                  </button>
                  <button
                    onClick={() => setMethod('INSTALLMENT')}
                    className={`rounded-md border-2 px-4 py-2 text-sm ${method === 'INSTALLMENT' ? 'border-[var(--color-primary)]' : 'border-slate-200'}`}
                  >
                    Termin
                  </button>
                </div>
                {method === 'INSTALLMENT' && (
                  <div className="flex gap-2">
                    {[1, 2, 3].map((n) => (
                      <button
                        key={n}
                        onClick={() => setInstallmentCount(n)}
                        className={`h-9 w-9 rounded-md border-2 text-sm ${installmentCount === n ? 'border-[var(--color-primary)]' : 'border-slate-200'}`}
                      >
                        {n}x
                      </button>
                    ))}
                  </div>
                )}
                <Button onClick={submitPlan} loading={choosePlan.isPending}>
                  Konfirmasi Metode Pembayaran
                </Button>
              </div>
            )}

            {plan && (
              <div className="space-y-2">
                <p className="text-sm font-medium text-slate-700">
                  {plan.method === 'FULL' ? 'Lunas' : `Termin ${plan.installment_count}x`}
                </p>
                {plan.installments?.map((inst) => {
                  const st = installmentStatusLabel[inst.status] ?? { label: inst.status, tone: 'neutral' as BadgeTone }
                  const canUpload = inst.status === 'PENDING_PAYMENT' || inst.status === 'REJECTED'
                  return (
                    <div key={inst.id} className="rounded-md border border-slate-100 p-3 text-sm">
                      <div className="flex items-center justify-between">
                        <p className="font-medium text-slate-800">
                          Cicilan {inst.sequence_no} &middot; {formatCurrency(inst.amount)}
                        </p>
                        <Badge tone={st.tone}>{st.label}</Badge>
                      </div>
                      {inst.bank_account && (
                        <p className="mt-1 text-xs text-slate-400">
                          Transfer ke {inst.bank_account.bank_name} {inst.bank_account.account_number} a.n. {inst.bank_account.account_holder}
                        </p>
                      )}
                      {inst.status === 'REJECTED' && inst.reject_reason && (
                        <p className="mt-1 text-xs text-[var(--color-danger)]">Alasan: {inst.reject_reason}</p>
                      )}
                      <div className="mt-2 flex items-center gap-2">
                        {inst.proof_image_url && (
                          <Button variant="ghost" size="sm" onClick={() => setProofPreview(inst.proof_image_url)}>
                            <ImageIcon className="h-4 w-4" /> Lihat Bukti
                          </Button>
                        )}
                        {canUpload && (
                          <Button
                            variant="secondary"
                            size="sm"
                            onClick={() => {
                              setUploadingFor(inst.id)
                              fileInputRef.current?.click()
                            }}
                            loading={uploadProof.isPending && uploadingFor === inst.id}
                          >
                            <Upload className="h-4 w-4" /> Upload Bukti Transfer
                          </Button>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      <input ref={fileInputRef} type="file" accept="image/jpeg,image/png,image/webp" className="hidden" onChange={(e) => onPickProof(e.target.files?.[0])} />

      <Dialog open={!!proofPreview} onOpenChange={(o) => !o && setProofPreview(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Bukti Transfer</DialogTitle>
          </DialogHeader>
          {proofPreview && <img src={toUploadUrl(proofPreview)} alt="Bukti transfer" className="w-full rounded-md" />}
        </DialogContent>
      </Dialog>
    </div>
  )
}
