import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, CheckCircle2, PackageMinus, Play, Printer, XCircle } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { StatusBadge } from '@/components/StatusBadge'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { PrintPreviewModal } from '@/components/print/PrintPreviewModal'
import { PrintableWorkOrder } from '@/components/print/PrintableWorkOrder'
import {
  useAddLabor,
  useAddOverhead,
  useBOMsByProduct,
  useCancelProductionOrder,
  useCompleteProductionOrder,
  useIssueMaterial,
  useProductionCost,
  useProductionOrder,
  useStartProductionOrder,
} from '@/hooks/useProduction'
import { useMaterials } from '@/hooks/useMaterials'
import { useProduct } from '@/hooks/useProducts'
import { useSalesOrder } from '@/hooks/useSales'
import { formatCurrency, formatDateTime } from '@/lib/format'
import { openServerPdf } from '@/lib/pdf'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'

export function ProductionOrderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: order, isLoading, error } = useProductionOrder(id)
  const { data: materials } = useMaterials()
  const { data: product } = useProduct(order?.product_id)
  const { data: soForOrder } = useSalesOrder(order?.sales_order_id)
  const { data: boms } = useBOMsByProduct(order?.product_id)
  const cost = useProductionCost(order?.production_status === 'COMPLETED' ? id : undefined)
  const bom = boms?.find((b) => b.id === order?.bom_id)

  const startOrder = useStartProductionOrder()
  const cancelOrder = useCancelProductionOrder()
  const issueMaterial = useIssueMaterial()
  const addLabor = useAddLabor()
  const addOverhead = useAddOverhead()
  const completeOrder = useCompleteProductionOrder()

  const [issueDialog, setIssueDialog] = useState<{ rowId: string; materialName: string } | null>(null)
  const [issueQty, setIssueQty] = useState('')
  const [laborDesc, setLaborDesc] = useState('')
  const [laborHours, setLaborHours] = useState('')
  const [laborRate, setLaborRate] = useState('')
  const [overheadDesc, setOverheadDesc] = useState('')
  const [overheadAmount, setOverheadAmount] = useState('')
  const [completeOpen, setCompleteOpen] = useState(false)
  const [finishedBySize, setFinishedBySize] = useState<Record<string, string>>({})
  const [printOpen, setPrintOpen] = useState(false)
  const [downloadingPdf, setDownloadingPdf] = useState(false)

  const materialName = (matId: string) => materials?.find((m) => m.id === matId)?.name ?? matId.slice(0, 8)
  const sizeCode = (sizeId: string) => product?.sizes?.find((s) => s.id === sizeId)?.size_code ?? sizeId.slice(0, 6)

  if (isLoading) return <LoadingState />
  if (error || !order) return <ErrorState message={(error as Error)?.message ?? 'Pesanan produksi tidak ditemukan'} />

  function submitIssue() {
    if (!issueDialog || !issueQty) return
    issueMaterial
      .mutateAsync({ orderId: id!, materialRowId: issueDialog.rowId, qty: issueQty })
      .then(() => {
        toast.success('Bahan berhasil dikeluarkan')
        setIssueDialog(null)
        setIssueQty('')
      })
      .catch((err) => toast.error('Gagal mengeluarkan bahan', err instanceof ApiError ? err.message : undefined))
  }

  function submitLabor() {
    if (!laborDesc || !laborHours || !laborRate) return toast.error('Isi semua kolom tenaga kerja')
    addLabor
      .mutateAsync({ orderId: id!, description: laborDesc, hours: laborHours, rate: laborRate })
      .then(() => {
        toast.success('Tenaga kerja berhasil dicatat')
        setLaborDesc('')
        setLaborHours('')
        setLaborRate('')
      })
      .catch((err) => toast.error('Gagal mencatat tenaga kerja', err instanceof ApiError ? err.message : undefined))
  }

  function submitOverhead() {
    if (!overheadDesc || !overheadAmount) return toast.error('Isi semua kolom overhead')
    addOverhead
      .mutateAsync({ orderId: id!, description: overheadDesc, allocation_basis: 'FIXED', amount: overheadAmount })
      .then(() => {
        toast.success('Overhead berhasil dicatat')
        setOverheadDesc('')
        setOverheadAmount('')
      })
      .catch((err) => toast.error('Gagal mencatat overhead', err instanceof ApiError ? err.message : undefined))
  }

  function submitComplete() {
    const items = Object.entries(finishedBySize)
      .filter(([, qty]) => Number(qty) > 0)
      .map(([sizeId, qty]) => ({ product_size_id: sizeId, finished_qty: qty }))
    if (items.length === 0) return toast.error('Masukkan jumlah selesai untuk minimal satu ukuran')
    completeOrder
      .mutateAsync({ orderId: id!, items })
      .then(() => {
        toast.success('Produksi selesai, HPP telah dihitung')
        setCompleteOpen(false)
      })
      .catch((err) => toast.error('Gagal menyelesaikan produksi', err instanceof ApiError ? err.message : undefined))
  }

  async function downloadWorkOrderPdf() {
    setDownloadingPdf(true)
    try {
      await openServerPdf(`/production-orders/${id}/pdf`, `work-order-${order!.prod_number}.pdf`)
    } catch (err) {
      toast.error('Gagal membuat PDF', err instanceof ApiError ? err.message : undefined)
    } finally {
      setDownloadingPdf(false)
    }
  }

  return (
    <div>
      <button onClick={() => navigate('/production/orders')} className="mb-4 flex items-center gap-1 text-sm text-slate-500 hover:text-slate-800">
        <ArrowLeft className="h-4 w-4" /> Kembali ke Pesanan Produksi
      </button>

      <PageHeader
        title={order.prod_number}
        description={`${order.finished_qty} / ${order.planned_qty} unit selesai`}
        actions={
          <div className="flex gap-2">
            <Button variant="secondary" onClick={() => setPrintOpen(true)}>
              <Printer className="h-4 w-4" /> Cetak Slip Kerja
            </Button>
            {order.production_status === 'NOT_STARTED' && (
              <Button
                onClick={() =>
                  startOrder
                    .mutateAsync(id!)
                    .then(() => toast.success('Produksi dimulai'))
                    .catch((err) => toast.error('Gagal memulai', err instanceof ApiError ? err.message : undefined))
                }
                loading={startOrder.isPending}
              >
                <Play className="h-4 w-4" /> Mulai Produksi
              </Button>
            )}
            {order.production_status === 'IN_PROGRESS' && (
              <Button onClick={() => setCompleteOpen(true)}>
                <CheckCircle2 className="h-4 w-4" /> Selesaikan &amp; Hitung HPP
              </Button>
            )}
            {(order.production_status === 'NOT_STARTED' || order.production_status === 'IN_PROGRESS') && (
              <Button
                variant="destructive"
                onClick={() =>
                  cancelOrder
                    .mutateAsync(id!)
                    .then(() => toast.success('Pesanan produksi dibatalkan'))
                    .catch((err) => toast.error('Gagal membatalkan', err instanceof ApiError ? err.message : undefined))
                }
                loading={cancelOrder.isPending}
              >
                <XCircle className="h-4 w-4" /> Batal
              </Button>
            )}
          </div>
        }
      />

      <div className="mb-6">
        <StatusBadge kind="production" value={order.production_status} />
      </div>

      <Tabs defaultValue="materials">
        <TabsList>
          <TabsTrigger value="materials">Bahan Baku</TabsTrigger>
          <TabsTrigger value="labor">Tenaga Kerja &amp; Overhead</TabsTrigger>
          <TabsTrigger value="hpp">Snapshot HPP</TabsTrigger>
        </TabsList>

        <TabsContent value="materials">
          <Card>
            <CardContent className="p-0">
              <table className="w-full text-sm">
                <thead className="bg-slate-50 text-left text-xs font-medium uppercase text-slate-500">
                  <tr>
                    <th className="px-4 py-2.5">Bahan</th>
                    <th className="px-4 py-2.5">Jml Rencana</th>
                    <th className="px-4 py-2.5">Jml Dikeluarkan</th>
                    <th className="px-4 py-2.5">Biaya Satuan</th>
                    <th className="px-4 py-2.5">Total Biaya</th>
                    <th className="px-4 py-2.5" />
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {order.materials?.length ? (
                    order.materials.map((m) => (
                      <tr key={m.id}>
                        <td className="px-4 py-2.5">{materialName(m.material_id)}</td>
                        <td className="px-4 py-2.5">{m.planned_qty}</td>
                        <td className="px-4 py-2.5">{m.issued_qty}</td>
                        <td className="px-4 py-2.5">{formatCurrency(m.unit_cost)}</td>
                        <td className="px-4 py-2.5">{formatCurrency(m.total_cost)}</td>
                        <td className="px-4 py-2.5 text-right">
                          {order.production_status === 'IN_PROGRESS' && (
                            <Button
                              size="sm"
                              variant="secondary"
                              onClick={() => setIssueDialog({ rowId: m.id, materialName: materialName(m.material_id) })}
                            >
                              <PackageMinus className="h-3.5 w-3.5" /> Keluarkan
                            </Button>
                          )}
                        </td>
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td colSpan={6} className="px-4 py-6 text-center text-slate-400">
                        Tidak ada bahan dari BOM untuk pesanan ini.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="labor">
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Tenaga Kerja Langsung</CardTitle>
              </CardHeader>
              <CardContent>
                <table className="mb-4 w-full text-sm">
                  <tbody className="divide-y divide-slate-100">
                    {order.labor?.map((l) => (
                      <tr key={l.id}>
                        <td className="py-1.5">{l.description}</td>
                        <td className="py-1.5 text-right text-slate-500">{l.hours} jam &times; {formatCurrency(l.rate)}</td>
                        <td className="py-1.5 text-right font-medium">{formatCurrency(l.total_cost)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                {order.production_status === 'IN_PROGRESS' && (
                  <div className="space-y-2">
                    <Input placeholder="Deskripsi" value={laborDesc} onChange={(e) => setLaborDesc(e.target.value)} />
                    <div className="flex gap-2">
                      <Input type="number" placeholder="Jam" value={laborHours} onChange={(e) => setLaborHours(e.target.value)} />
                      <Input type="number" placeholder="Tarif/jam" value={laborRate} onChange={(e) => setLaborRate(e.target.value)} />
                    </div>
                    <Button size="sm" onClick={submitLabor} loading={addLabor.isPending}>
                      Tambah Tenaga Kerja
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Overhead Pabrik</CardTitle>
              </CardHeader>
              <CardContent>
                <table className="mb-4 w-full text-sm">
                  <tbody className="divide-y divide-slate-100">
                    {order.overheads?.map((o) => (
                      <tr key={o.id}>
                        <td className="py-1.5">{o.description}</td>
                        <td className="py-1.5 text-right font-medium">{formatCurrency(o.amount)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                {order.production_status === 'IN_PROGRESS' && (
                  <div className="space-y-2">
                    <Input placeholder="Deskripsi" value={overheadDesc} onChange={(e) => setOverheadDesc(e.target.value)} />
                    <Input type="number" placeholder="Jumlah" value={overheadAmount} onChange={(e) => setOverheadAmount(e.target.value)} />
                    <Button size="sm" onClick={submitOverhead} loading={addOverhead.isPending}>
                      Tambah Overhead
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="hpp">
          {order.production_status !== 'COMPLETED' && (
            <p className="text-sm text-slate-400">HPP dihitung saat produksi ditandai selesai.</p>
          )}
          {cost.isLoading && <LoadingState />}
          {cost.data && (
            <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
              <Card>
                <CardContent className="py-4">
                  <p className="text-xs uppercase text-slate-400">Biaya Bahan</p>
                  <p className="mt-1 text-lg font-semibold">{formatCurrency(cost.data.total_material_cost)}</p>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="py-4">
                  <p className="text-xs uppercase text-slate-400">Biaya Tenaga Kerja</p>
                  <p className="mt-1 text-lg font-semibold">{formatCurrency(cost.data.total_labor_cost)}</p>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="py-4">
                  <p className="text-xs uppercase text-slate-400">Biaya Overhead</p>
                  <p className="mt-1 text-lg font-semibold">{formatCurrency(cost.data.total_overhead_cost)}</p>
                </CardContent>
              </Card>
              <Card className="border-[var(--color-primary)]">
                <CardContent className="py-4">
                  <p className="text-xs uppercase text-slate-400">Biaya Satuan (HPP)</p>
                  <p className="mt-1 text-lg font-semibold text-[var(--color-primary)]">{formatCurrency(cost.data.unit_cost)}</p>
                </CardContent>
              </Card>
              <Card className="col-span-2 lg:col-span-4">
                <CardContent className="flex items-center justify-between py-4 text-sm">
                  <span className="text-slate-500">
                    Total Biaya {formatCurrency(cost.data.total_cost)} &divide; {cost.data.finished_qty} unit selesai
                  </span>
                  <span className="text-slate-400">Dihitung {formatDateTime(cost.data.computed_at)}</span>
                </CardContent>
              </Card>
            </div>
          )}
        </TabsContent>
      </Tabs>

      <Dialog open={!!issueDialog} onOpenChange={(o) => !o && setIssueDialog(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Keluarkan {issueDialog?.materialName}</DialogTitle>
          </DialogHeader>
          <div className="space-y-1.5">
            <Label>Jumlah yang dikeluarkan</Label>
            <Input type="number" step="0.0001" value={issueQty} onChange={(e) => setIssueQty(e.target.value)} />
          </div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setIssueDialog(null)}>
              Batal
            </Button>
            <Button onClick={submitIssue} loading={issueMaterial.isPending}>
              Keluarkan Bahan
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={completeOpen} onOpenChange={setCompleteOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Selesaikan Produksi</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            {order.items?.map((it) => (
              <div key={it.id} className="flex items-center justify-between gap-3">
                <Label className="flex-1">
                  {sizeCode(it.product_size_id)} <span className="text-slate-400">(rencana {it.planned_qty})</span>
                </Label>
                <Input
                  type="number"
                  min="0"
                  max={it.planned_qty}
                  className="w-32"
                  value={finishedBySize[it.product_size_id] ?? ''}
                  onChange={(e) => setFinishedBySize((s) => ({ ...s, [it.product_size_id]: e.target.value }))}
                />
              </div>
            ))}
          </div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setCompleteOpen(false)}>
              Batal
            </Button>
            <Button onClick={submitComplete} loading={completeOrder.isPending}>
              Selesaikan &amp; Hitung HPP
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <PrintPreviewModal
        open={printOpen}
        onOpenChange={setPrintOpen}
        title={`Work Order – ${order.prod_number}`}
        onDownloadPdf={downloadWorkOrderPdf}
        downloadPending={downloadingPdf}
      >
        <PrintableWorkOrder order={order} product={product} materials={materials} soNumber={soForOrder?.so_number} bomName={bom?.name} />
      </PrintPreviewModal>
    </div>
  )
}
