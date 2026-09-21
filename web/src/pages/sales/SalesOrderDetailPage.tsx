import { useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, Check, FileText, ImageIcon, PackageCheck, Printer, Truck, X, XCircle } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { StatusBadge } from '@/components/StatusBadge'
import { Badge, type BadgeTone } from '@/components/ui/Badge'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { PrintPreviewModal } from '@/components/print/PrintPreviewModal'
import { PrintableDeliverySlip } from '@/components/print/PrintableDeliverySlip'
import { PrintableInvoice } from '@/components/print/PrintableInvoice'
import { useCancelOrder, useConfirmOrder, useCreateInvoiceForOrder, useDeliverOrder, useInvoice, useInvoices, useSalesOrder } from '@/hooks/useSales'
import { useCustomers } from '@/hooks/useCustomers'
import { useProducts } from '@/hooks/useProducts'
import { useProductionOrders } from '@/hooks/useProduction'
import { useSpkOrdersBySalesOrder } from '@/hooks/useSpk'
import { usePayments } from '@/hooks/useFinance'
import { useAuditTrail } from '@/hooks/useAccounting'
import { useQuotationBySalesOrder } from '@/hooks/useQuotations'
import { useFabrics, useGarmentSizes, useGarmentVariants, useInks, useProductTypes } from '@/hooks/useCatalog'
import { useConfirmInstallment, usePaymentPlanByInvoice, useRejectInstallment } from '@/hooks/useBilling'
import { formatCurrency, formatDate, formatDateTime, toUploadUrl } from '@/lib/format'
import { openServerPdf } from '@/lib/pdf'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { Invoice } from '@/types/sales'
import type { Installment } from '@/types/billing'
import type { SpkOrder } from '@/types/spk'

type ProductionBucket = 'design' | 'produksi' | 'selesai' | 'pengiriman'
const bucketOrder: ProductionBucket[] = ['design', 'produksi', 'selesai', 'pengiriman']
const bucketLabel: Record<ProductionBucket, string> = { design: 'Design', produksi: 'Produksi', selesai: 'Selesai', pengiriman: 'Pengiriman' }
const bucketTone: Record<ProductionBucket, BadgeTone> = { design: 'neutral', produksi: 'info', selesai: 'warning', pengiriman: 'success' }

// Same milestone-based bucketing as the /production/costing board: stages
// don't gate each other, so this is derived from the DESIGN/DONE/SHIPPING
// milestone flags, not from "furthest stage touched".
function spkBucket(o: SpkOrder): ProductionBucket {
  if (!o.design_done) return 'design'
  if (!o.production_done) return 'produksi'
  if (!o.shipped) return 'selesai'
  return 'pengiriman'
}

export function SalesOrderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: order, isLoading, error } = useSalesOrder(id)
  const { data: customers } = useCustomers()
  const { data: products } = useProducts()
  const { data: productionOrders } = useProductionOrders()
  const { data: spkOrders } = useSpkOrdersBySalesOrder(id)
  const { data: invoices } = useInvoices()
  const { data: payments } = usePayments()
  const { data: quotation } = useQuotationBySalesOrder(id)
  const { data: productTypes } = useProductTypes()
  const { data: fabrics } = useFabrics()
  const { data: variants } = useGarmentVariants()
  const { data: inks } = useInks()
  const { data: garmentSizes } = useGarmentSizes()

  const confirmOrder = useConfirmOrder()
  const cancelOrder = useCancelOrder()
  const createInvoice = useCreateInvoiceForOrder()
  const deliverOrder = useDeliverOrder()
  const confirmInstallment = useConfirmInstallment()
  const rejectInstallment = useRejectInstallment()

  const [deliverySlipOpen, setDeliverySlipOpen] = useState(false)
  const [invoicePreviewId, setInvoicePreviewId] = useState<string | null>(null)
  const { data: invoicePreview, isLoading: invoicePreviewLoading } = useInvoice(invoicePreviewId ?? undefined)
  const [downloadingPdf, setDownloadingPdf] = useState(false)
  const [proofPreview, setProofPreview] = useState<string | null>(null)
  const [rejectTarget, setRejectTarget] = useState<Installment | null>(null)
  const [rejectReason, setRejectReason] = useState('')

  const customer = customers?.find((c) => c.id === order?.customer_id)
  const productName = (pid: string) => products?.find((p) => p.id === pid)?.name ?? pid.slice(0, 8)
  const sizeCode = (pid: string, sid: string) => products?.find((p) => p.id === pid)?.sizes?.find((s) => s.id === sid)?.size_code ?? '-'

  const jenisName = (tId: string) => productTypes?.find((t) => t.id === tId)?.name ?? tId.slice(0, 8)
  const fabricName = (fId: string) => fabrics?.find((f) => f.id === fId)?.name ?? fId.slice(0, 8)
  const variantName = (vId: string) => variants?.find((v) => v.id === vId)?.name ?? vId.slice(0, 8)
  const inkName = (iId: string) => inks?.find((i) => i.id === iId)?.name ?? iId.slice(0, 8)
  const garmentSizeCode = (sId: string) => garmentSizes?.find((s) => s.id === sId)?.size_code ?? sId.slice(0, 8)
  const quotationItemFor = (quotationItemId?: string) => quotation?.items?.find((qi) => qi.id === quotationItemId)

  const relatedProduction = useMemo(
    () => productionOrders?.filter((po) => po.sales_order_id === id) ?? [],
    [productionOrders, id],
  )
  const relatedInvoices = useMemo(() => invoices?.filter((inv) => inv.sales_order_id === id) ?? [], [invoices, id])
  const primaryInvoice = relatedInvoices[0]
  const { data: paymentPlan } = usePaymentPlanByInvoice(primaryInvoice?.id)
  const relatedPayments = useMemo(() => {
    const invoiceIds = new Set(relatedInvoices.map((i) => i.id))
    return payments?.filter((p) => p.allocations?.some((a) => invoiceIds.has(a.invoice_id))) ?? []
  }, [payments, relatedInvoices])

  const sourceIds = useMemo(
    () => [id, ...relatedInvoices.map((i) => i.id), ...relatedProduction.map((p) => p.id), ...relatedPayments.map((p) => p.id)].filter(
      (v): v is string => !!v,
    ),
    [id, relatedInvoices, relatedProduction, relatedPayments],
  )
  const auditTrail = useAuditTrail(sourceIds)

  const productionProgress = useMemo(() => {
    if (relatedProduction.length === 0) return null
    const planned = relatedProduction.reduce((s, p) => s + Number(p.planned_qty), 0)
    const finished = relatedProduction.reduce((s, p) => s + Number(p.finished_qty), 0)
    return planned > 0 ? Math.round((finished / planned) * 100) : 0
  }, [relatedProduction])

  // SPK is the real production-floor tracker now (BOM/HPP above is the
  // legacy costing module, rarely used per-order) -- when this SO has SPKs,
  // the summary cards below follow those instead of the coarse
  // production_status/delivery_status enum on the sales order itself.
  const spkProduction = useMemo(() => {
    if (!spkOrders?.length) return null
    const buckets = spkOrders.map(spkBucket)
    const bucket = bucketOrder.find((b) => buckets.includes(b))!
    const pct = Math.round(spkOrders.reduce((s, o) => s + Number(o.progress_pct), 0) / spkOrders.length)
    return { bucket, pct }
  }, [spkOrders])
  const spkShipped = spkOrders?.length ? spkOrders.every((o) => o.shipped) : null

  if (isLoading) return <LoadingState />
  if (error || !order) return <ErrorState message={(error as Error)?.message ?? 'Pesanan tidak ditemukan'} />

  function runCommand(mutation: { mutateAsync: (args: { id: string; body?: unknown }) => Promise<unknown> }, body?: unknown, successMsg?: string) {
    mutation
      .mutateAsync({ id: id!, body })
      .then(() => toast.success(successMsg ?? 'Selesai'))
      .catch((err) => toast.error('Tindakan gagal', err instanceof ApiError ? err.message : undefined))
  }

  async function downloadDeliverySlipPdf() {
    setDownloadingPdf(true)
    try {
      await openServerPdf(`/sales-orders/${id}/delivery-slip/pdf`, `surat-jalan-${order!.so_number}.pdf`)
    } catch (err) {
      toast.error('Gagal membuat PDF', err instanceof ApiError ? err.message : undefined)
    } finally {
      setDownloadingPdf(false)
    }
  }

  async function downloadInvoicePdf(inv: Invoice) {
    setDownloadingPdf(true)
    try {
      await openServerPdf(`/invoices/${inv.id}/pdf`, `faktur-${inv.invoice_number}.pdf`)
    } catch (err) {
      toast.error('Gagal membuat PDF', err instanceof ApiError ? err.message : undefined)
    } finally {
      setDownloadingPdf(false)
    }
  }

  function confirmPayment(installmentId: string) {
    confirmInstallment
      .mutateAsync(installmentId)
      .then(() => toast.success('Pembayaran dikonfirmasi', 'Sudah tercatat di modul finance/akuntansi'))
      .catch((err) => toast.error('Gagal mengonfirmasi', err instanceof ApiError ? err.message : undefined))
  }

  function submitReject() {
    if (!rejectTarget) return
    rejectInstallment
      .mutateAsync({ id: rejectTarget.id, reason: rejectReason })
      .then(() => {
        toast.success('Pembayaran ditolak')
        setRejectTarget(null)
        setRejectReason('')
      })
      .catch((err) => toast.error('Gagal menolak', err instanceof ApiError ? err.message : undefined))
  }

  const installmentStatusLabel: Record<string, string> = {
    PENDING_PAYMENT: 'Menunggu Transfer',
    PENDING_VERIFICATION: 'Menunggu Verifikasi',
    CONFIRMED: 'Terkonfirmasi',
    REJECTED: 'Ditolak',
  }

  return (
    <div>
      <button onClick={() => navigate('/sales/orders')} className="mb-4 flex items-center gap-1 text-sm text-slate-500 hover:text-slate-800">
        <ArrowLeft className="h-4 w-4" /> Kembali ke Pesanan Penjualan
      </button>

      <PageHeader
        title={order.so_number}
        description={`Dibuat oleh ${customer?.name ?? order.customer_id} pada ${formatDate(order.order_date)}`}
        actions={
          <div className="flex gap-2">
            {order.order_status === 'DRAFT' && (
              <Button onClick={() => runCommand(confirmOrder, undefined, 'Pesanan dikonfirmasi')} loading={confirmOrder.isPending}>
                Konfirmasi Pesanan
              </Button>
            )}
            {order.order_status === 'CONFIRMED' && relatedInvoices.length === 0 && (
              <Button
                onClick={() =>
                  createInvoice
                    .mutateAsync(id!)
                    .then(() => toast.success('Faktur dibuat'))
                    .catch((err) => toast.error('Gagal membuat faktur', err instanceof ApiError ? err.message : undefined))
                }
                loading={createInvoice.isPending}
              >
                <FileText className="h-4 w-4" /> Buat Faktur
              </Button>
            )}
            {order.delivery_status !== 'DELIVERED' && order.order_status !== 'CANCELLED' && (
              <Button variant="secondary" onClick={() => runCommand(deliverOrder, undefined, 'Pesanan dikirim')} loading={deliverOrder.isPending}>
                <Truck className="h-4 w-4" /> Kirim
              </Button>
            )}
            <Button variant="secondary" onClick={() => setDeliverySlipOpen(true)}>
              <Printer className="h-4 w-4" /> Surat Jalan
            </Button>
            {order.order_status === 'DRAFT' && (
              <Button
                variant="destructive"
                onClick={() => runCommand(cancelOrder, { reason: 'Cancelled from UI' }, 'Pesanan dibatalkan')}
                loading={cancelOrder.isPending}
              >
                <XCircle className="h-4 w-4" /> Batal
              </Button>
            )}
          </div>
        }
      />

      <div className="mb-6 grid grid-cols-2 gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="py-4">
            <p className="text-xs uppercase text-slate-400">Status Pesanan</p>
            <div className="mt-1.5">
              <StatusBadge kind="order" value={order.order_status} />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4">
            <p className="text-xs uppercase text-slate-400">Pembayaran</p>
            <div className="mt-1.5">
              <StatusBadge kind="payment" value={order.payment_status} />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4">
            <p className="text-xs uppercase text-slate-400">Produksi</p>
            {spkProduction ? (
              <>
                <div className="mt-1.5 flex items-center gap-2">
                  <Badge tone={bucketTone[spkProduction.bucket]}>{bucketLabel[spkProduction.bucket]}</Badge>
                  <span className="text-xs text-slate-400">{spkProduction.pct}%</span>
                </div>
                <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-slate-100">
                  <div className="h-full rounded-full bg-[var(--color-primary)]" style={{ width: `${spkProduction.pct}%` }} />
                </div>
              </>
            ) : (
              <div className="mt-1.5 flex items-center gap-2">
                <StatusBadge kind="production" value={order.production_status} />
                {productionProgress !== null && <span className="text-xs text-slate-400">{productionProgress}%</span>}
              </div>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4">
            <p className="text-xs uppercase text-slate-400">Pengiriman</p>
            <div className="mt-1.5">
              {spkShipped !== null ? (
                <Badge tone={spkShipped ? 'success' : 'neutral'}>{spkShipped ? 'Terkirim' : 'Belum Dikirim'}</Badge>
              ) : (
                <StatusBadge kind="delivery" value={order.delivery_status} />
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">Ringkasan</TabsTrigger>
          <TabsTrigger value="items">Item</TabsTrigger>
          <TabsTrigger value="payment">Pembayaran</TabsTrigger>
          <TabsTrigger value="production">Produksi</TabsTrigger>
          <TabsTrigger value="audit">Jejak Audit Jurnal</TabsTrigger>
        </TabsList>

        <TabsContent value="overview">
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Ringkasan Pesanan</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                <div className="flex justify-between"><span className="text-slate-500">Subtotal</span><span>{formatCurrency(order.subtotal)}</span></div>
                <div className="flex justify-between"><span className="text-slate-500">Diskon</span><span>{formatCurrency(order.discount_total)}</span></div>
                <div className="flex justify-between"><span className="text-slate-500">Pajak</span><span>{formatCurrency(order.tax_total)}</span></div>
                <div className="flex justify-between border-t border-slate-100 pt-2 font-semibold"><span>Total Keseluruhan</span><span>{formatCurrency(order.grand_total)}</span></div>
                {order.notes && <p className="mt-3 rounded-md bg-slate-50 p-2 text-slate-500">{order.notes}</p>}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Faktur &amp; Pembayaran</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3 text-sm">
                {relatedInvoices.length === 0 && <p className="text-slate-400">Belum ada faktur dibuat.</p>}
                {relatedInvoices.map((inv) => (
                  <div key={inv.id} className="flex items-center justify-between rounded-md border border-slate-100 p-2">
                    <div>
                      <p className="font-medium">{inv.invoice_number}</p>
                      <p className="text-xs text-slate-400">Dibayar {formatCurrency(inv.paid_amount)} dari {formatCurrency(inv.grand_total)}</p>
                    </div>
                    <div className="flex items-center gap-2">
                      <StatusBadge kind="invoice" value={inv.status} />
                      <Button variant="ghost" size="icon" onClick={() => setInvoicePreviewId(inv.id)} title="Cetak faktur">
                        <Printer className="h-4 w-4 text-slate-400" />
                      </Button>
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="items">
          <Card>
            <CardContent className="p-0">
              <table className="w-full text-sm">
                <thead className="bg-slate-50 text-left text-xs font-medium uppercase text-slate-500">
                  <tr>
                    {quotation ? (
                      <>
                        <th className="px-4 py-2.5">Jenis</th>
                        <th className="px-4 py-2.5">Model</th>
                        <th className="px-4 py-2.5">Bahan</th>
                        <th className="px-4 py-2.5">Tinta</th>
                        <th className="px-4 py-2.5">Ukuran</th>
                      </>
                    ) : (
                      <>
                        <th className="px-4 py-2.5">Produk</th>
                        <th className="px-4 py-2.5">Ukuran</th>
                      </>
                    )}
                    <th className="px-4 py-2.5">Jml</th>
                    <th className="px-4 py-2.5">Harga Satuan</th>
                    <th className="px-4 py-2.5">Diskon</th>
                    <th className="px-4 py-2.5">Pajak</th>
                    <th className="px-4 py-2.5 text-right">Total Baris</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {order.items?.map((it) => {
                    const qi = quotationItemFor(it.quotation_item_id)
                    return (
                      <tr key={it.id}>
                        {qi ? (
                          <>
                            <td className="px-4 py-2.5">{jenisName(qi.product_type_id)}</td>
                            <td className="px-4 py-2.5">{variantName(qi.variant_id)}</td>
                            <td className="px-4 py-2.5">{fabricName(qi.fabric_id)}</td>
                            <td className="px-4 py-2.5">{inkName(qi.ink_id)}</td>
                            <td className="px-4 py-2.5">{garmentSizeCode(qi.garment_size_id)}</td>
                          </>
                        ) : (
                          <>
                            <td className="px-4 py-2.5">{productName(it.product_id)}</td>
                            <td className="px-4 py-2.5">{sizeCode(it.product_id, it.product_size_id)}</td>
                          </>
                        )}
                        <td className="px-4 py-2.5">{it.qty}</td>
                        <td className="px-4 py-2.5">{formatCurrency(it.unit_price)}</td>
                        <td className="px-4 py-2.5">{formatCurrency(it.discount)}</td>
                        <td className="px-4 py-2.5">{Number(it.tax_rate) * 100}%</td>
                        <td className="px-4 py-2.5 text-right font-medium">{formatCurrency(it.line_total)}</td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="payment">
          {!primaryInvoice && (
            <Card>
              <CardContent className="py-8 text-center text-slate-400">Pesanan ini belum memiliki faktur.</CardContent>
            </Card>
          )}
          {primaryInvoice && !paymentPlan && (
            <Card>
              <CardContent className="py-8 text-center text-slate-400">Customer belum memilih metode pembayaran.</CardContent>
            </Card>
          )}
          {primaryInvoice && paymentPlan && (
            <Card>
              <CardHeader>
                <CardTitle>
                  {paymentPlan.method === 'FULL' ? 'Lunas' : `Termin ${paymentPlan.installment_count}x`} &middot; {primaryInvoice.invoice_number}
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {paymentPlan.installments?.map((inst) => (
                  <div key={inst.id} className="flex items-center justify-between rounded-md border border-slate-100 p-3 text-sm">
                    <div>
                      <p className="font-medium text-slate-800">
                        Cicilan {inst.sequence_no} &middot; {formatCurrency(inst.amount)}
                      </p>
                      <p className="text-xs text-slate-400">
                        {inst.bank_account ? `${inst.bank_account.bank_name} - ${inst.bank_account.account_number} a.n. ${inst.bank_account.account_holder}` : ''}
                      </p>
                      {inst.status === 'REJECTED' && inst.reject_reason && (
                        <p className="mt-1 text-xs text-[var(--color-danger)]">Alasan ditolak: {inst.reject_reason}</p>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      {inst.proof_image_url && (
                        <Button variant="ghost" size="sm" onClick={() => setProofPreview(inst.proof_image_url)}>
                          <ImageIcon className="h-4 w-4" /> Bukti
                        </Button>
                      )}
                      <span className="text-xs text-slate-500">{installmentStatusLabel[inst.status] ?? inst.status}</span>
                      {inst.status === 'PENDING_VERIFICATION' && (
                        <>
                          <Button size="sm" onClick={() => confirmPayment(inst.id)} loading={confirmInstallment.isPending}>
                            <Check className="h-4 w-4" /> Konfirmasi
                          </Button>
                          <Button size="sm" variant="destructive" onClick={() => setRejectTarget(inst)}>
                            <X className="h-4 w-4" /> Tolak
                          </Button>
                        </>
                      )}
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="production">
          <div className="mb-6">
            <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-500">SPK &middot; Tahapan Produksi</h3>
            {!spkOrders?.length ? (
              <Card>
                <CardContent className="py-6 text-center text-sm text-slate-400">
                  Belum ada SPK -- dibuat otomatis begitu pesanan ini dikonfirmasi.
                </CardContent>
              </Card>
            ) : (
              <div className="space-y-3">
                {spkOrders.map((spkOrder) => {
                  const pct = Math.min(100, Math.round(Number(spkOrder.progress_pct)))
                  return (
                    <Card key={spkOrder.id} className="cursor-pointer hover:border-slate-300" onClick={() => navigate(`/production/orders/${spkOrder.id}`)}>
                      <CardContent className="py-4">
                        <div className="flex items-center justify-between">
                          <div>
                            <p className="font-medium text-slate-900">{spkOrder.spk_number} &middot; {spkOrder.product_type_name}</p>
                            <p className="text-xs text-slate-400">{spkOrder.current_stage_name} &middot; {spkOrder.total_qty} pcs</p>
                          </div>
                          <span className="text-xs font-medium text-slate-500">{pct}%</span>
                        </div>
                        <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-slate-100">
                          <div className="h-full rounded-full bg-[var(--color-primary)]" style={{ width: `${pct}%` }} />
                        </div>
                      </CardContent>
                    </Card>
                  )
                })}
              </div>
            )}
          </div>

          <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-500">HPP Produksi (BOM)</h3>
          {relatedProduction.length === 0 ? (
            <Card>
              <CardContent className="py-8 text-center text-slate-400">
                Belum ada pesanan produksi (HPP) yang terkait dengan pesanan penjualan ini.
                <div className="mt-3">
                  <Button variant="secondary" size="sm" onClick={() => navigate('/production/hpp/new')}>
                    <PackageCheck className="h-4 w-4" /> Buat Pesanan Produksi
                  </Button>
                </div>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-3">
              {relatedProduction.map((po) => (
                <Card key={po.id} className="cursor-pointer hover:border-slate-300" onClick={() => navigate(`/production/hpp/${po.id}`)}>
                  <CardContent className="flex items-center justify-between py-4">
                    <div>
                      <p className="font-medium text-slate-900">{po.prod_number}</p>
                      <p className="text-xs text-slate-400">{po.finished_qty} / {po.planned_qty} unit selesai</p>
                    </div>
                    <StatusBadge kind="production" value={po.production_status} />
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="audit">
          {auditTrail.isLoading && <LoadingState />}
          {auditTrail.data && auditTrail.data.length === 0 && <p className="text-sm text-slate-400">Belum ada jurnal yang diposting untuk pesanan ini.</p>}
          <div className="space-y-3">
            {auditTrail.data?.map((j) => (
              <Card key={j.id}>
                <CardHeader>
                  <div>
                    <Link to={`/accounting/journals/${j.id}`} className="font-medium text-[var(--color-primary)] hover:underline">
                      {j.journal_number}
                    </Link>
                    <p className="text-xs text-slate-400">{j.source_type} &middot; {formatDateTime(j.created_at)} &middot; oleh {j.created_by}</p>
                  </div>
                  <StatusBadge kind="journal" value={j.status} />
                </CardHeader>
                <CardContent className="p-0">
                  <table className="w-full text-sm">
                    <tbody className="divide-y divide-slate-100">
                      {j.lines?.map((l) => (
                        <tr key={l.id}>
                          <td className="px-4 py-2">{l.account_code} &middot; {l.account_name}</td>
                          <td className="px-4 py-2 text-right">{Number(l.debit) > 0 ? formatCurrency(l.debit) : ''}</td>
                          <td className="px-4 py-2 text-right">{Number(l.credit) > 0 ? formatCurrency(l.credit) : ''}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>
      </Tabs>

      <PrintPreviewModal
        open={deliverySlipOpen}
        onOpenChange={setDeliverySlipOpen}
        title={`Surat Jalan – ${order.so_number}`}
        onDownloadPdf={downloadDeliverySlipPdf}
        downloadPending={downloadingPdf}
      >
        <PrintableDeliverySlip order={order} customer={customer} products={products} />
      </PrintPreviewModal>

      <PrintPreviewModal
        open={!!invoicePreviewId}
        onOpenChange={(o) => !o && setInvoicePreviewId(null)}
        title={`Faktur – ${invoicePreview?.invoice_number ?? ''}`}
        onDownloadPdf={invoicePreview ? () => downloadInvoicePdf(invoicePreview) : undefined}
        downloadPending={downloadingPdf}
      >
        {invoicePreviewLoading && <LoadingState />}
        {invoicePreview && <PrintableInvoice invoice={invoicePreview} soNumber={order.so_number} customer={customer} products={products} />}
      </PrintPreviewModal>

      <Dialog open={!!proofPreview} onOpenChange={(o) => !o && setProofPreview(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Bukti Transfer</DialogTitle>
          </DialogHeader>
          {proofPreview && <img src={toUploadUrl(proofPreview)} alt="Bukti transfer" className="w-full rounded-md" />}
        </DialogContent>
      </Dialog>

      <Dialog open={!!rejectTarget} onOpenChange={(o) => !o && setRejectTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Tolak Pembayaran</DialogTitle>
          </DialogHeader>
          <Input value={rejectReason} onChange={(e) => setRejectReason(e.target.value)} placeholder="Alasan penolakan" />
          <DialogFooter>
            <Button variant="secondary" onClick={() => setRejectTarget(null)}>
              Batal
            </Button>
            <Button variant="destructive" onClick={submitReject} loading={rejectInstallment.isPending}>
              Tolak
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
