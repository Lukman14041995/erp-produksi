import { useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, FileText, PackageCheck, Printer, Truck, XCircle } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { StatusBadge } from '@/components/StatusBadge'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs'
import { PrintPreviewModal } from '@/components/print/PrintPreviewModal'
import { PrintableDeliverySlip } from '@/components/print/PrintableDeliverySlip'
import { PrintableInvoice } from '@/components/print/PrintableInvoice'
import { useCancelOrder, useConfirmOrder, useCreateInvoiceForOrder, useDeliverOrder, useInvoice, useInvoices, useSalesOrder } from '@/hooks/useSales'
import { useCustomers } from '@/hooks/useCustomers'
import { useProducts } from '@/hooks/useProducts'
import { useProductionOrders } from '@/hooks/useProduction'
import { usePayments } from '@/hooks/useFinance'
import { useAuditTrail } from '@/hooks/useAccounting'
import { formatCurrency, formatDate, formatDateTime } from '@/lib/format'
import { openServerPdf } from '@/lib/pdf'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { Invoice } from '@/types/sales'

export function SalesOrderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: order, isLoading, error } = useSalesOrder(id)
  const { data: customers } = useCustomers()
  const { data: products } = useProducts()
  const { data: productionOrders } = useProductionOrders()
  const { data: invoices } = useInvoices()
  const { data: payments } = usePayments()

  const confirmOrder = useConfirmOrder()
  const cancelOrder = useCancelOrder()
  const createInvoice = useCreateInvoiceForOrder()
  const deliverOrder = useDeliverOrder()

  const [deliverySlipOpen, setDeliverySlipOpen] = useState(false)
  const [invoicePreviewId, setInvoicePreviewId] = useState<string | null>(null)
  const { data: invoicePreview, isLoading: invoicePreviewLoading } = useInvoice(invoicePreviewId ?? undefined)
  const [downloadingPdf, setDownloadingPdf] = useState(false)

  const customer = customers?.find((c) => c.id === order?.customer_id)
  const productName = (pid: string) => products?.find((p) => p.id === pid)?.name ?? pid.slice(0, 8)
  const sizeCode = (pid: string, sid: string) => products?.find((p) => p.id === pid)?.sizes?.find((s) => s.id === sid)?.size_code ?? '-'

  const relatedProduction = useMemo(
    () => productionOrders?.filter((po) => po.sales_order_id === id) ?? [],
    [productionOrders, id],
  )
  const relatedInvoices = useMemo(() => invoices?.filter((inv) => inv.sales_order_id === id) ?? [], [invoices, id])
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
            <div className="mt-1.5 flex items-center gap-2">
              <StatusBadge kind="production" value={order.production_status} />
              {productionProgress !== null && <span className="text-xs text-slate-400">{productionProgress}%</span>}
            </div>
            {productionProgress !== null && (
              <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-slate-100">
                <div className="h-full rounded-full bg-[var(--color-primary)]" style={{ width: `${productionProgress}%` }} />
              </div>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4">
            <p className="text-xs uppercase text-slate-400">Pengiriman</p>
            <div className="mt-1.5">
              <StatusBadge kind="delivery" value={order.delivery_status} />
            </div>
          </CardContent>
        </Card>
      </div>

      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">Ringkasan</TabsTrigger>
          <TabsTrigger value="items">Item</TabsTrigger>
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
                    <th className="px-4 py-2.5">Produk</th>
                    <th className="px-4 py-2.5">Ukuran</th>
                    <th className="px-4 py-2.5">Jml</th>
                    <th className="px-4 py-2.5">Harga Satuan</th>
                    <th className="px-4 py-2.5">Diskon</th>
                    <th className="px-4 py-2.5">Pajak</th>
                    <th className="px-4 py-2.5 text-right">Total Baris</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {order.items?.map((it) => (
                    <tr key={it.id}>
                      <td className="px-4 py-2.5">{productName(it.product_id)}</td>
                      <td className="px-4 py-2.5">{sizeCode(it.product_id, it.product_size_id)}</td>
                      <td className="px-4 py-2.5">{it.qty}</td>
                      <td className="px-4 py-2.5">{formatCurrency(it.unit_price)}</td>
                      <td className="px-4 py-2.5">{formatCurrency(it.discount)}</td>
                      <td className="px-4 py-2.5">{Number(it.tax_rate) * 100}%</td>
                      <td className="px-4 py-2.5 text-right font-medium">{formatCurrency(it.line_total)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="production">
          {relatedProduction.length === 0 ? (
            <Card>
              <CardContent className="py-8 text-center text-slate-400">
                Belum ada pesanan produksi yang terkait dengan pesanan penjualan ini.
                <div className="mt-3">
                  <Button variant="secondary" size="sm" onClick={() => navigate('/production/orders/new')}>
                    <PackageCheck className="h-4 w-4" /> Buat Pesanan Produksi
                  </Button>
                </div>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-3">
              {relatedProduction.map((po) => (
                <Card key={po.id} className="cursor-pointer hover:border-slate-300" onClick={() => navigate(`/production/orders/${po.id}`)}>
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
    </div>
  )
}
