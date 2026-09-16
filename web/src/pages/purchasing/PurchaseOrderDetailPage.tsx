import { useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, CheckCircle2, PackagePlus, Printer, Receipt, XCircle } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { StatusBadge } from '@/components/StatusBadge'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { PrintPreviewModal } from '@/components/print/PrintPreviewModal'
import { PrintablePurchaseOrder } from '@/components/print/PrintablePurchaseOrder'
import {
  useApprovePO,
  useCancelPO,
  useCreateBillFromGRN,
  useCreateGoodsReceipt,
  useGoodsReceiptsByPO,
  usePurchaseOrder,
} from '@/hooks/usePurchasing'
import { useSuppliers } from '@/hooks/useSuppliers'
import { useMaterials } from '@/hooks/useMaterials'
import { formatCurrency, formatDate } from '@/lib/format'
import { openServerPdf } from '@/lib/pdf'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { GoodsReceiptItem } from '@/types/purchasing'

function ProgressBar({ pct }: { pct: number }) {
  return (
    <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-slate-100">
      <div className="h-full rounded-full bg-[var(--color-primary)]" style={{ width: `${Math.min(100, pct)}%` }} />
    </div>
  )
}

interface MatchLine {
  grnItem: GoodsReceiptItem
  qty: string
  billUnitCost: string
}

export function PurchaseOrderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: po, isLoading, error } = usePurchaseOrder(id)
  const { data: grns } = useGoodsReceiptsByPO(id)
  const { data: suppliers } = useSuppliers()
  const { data: materials } = useMaterials()

  const approvePO = useApprovePO()
  const cancelPO = useCancelPO()
  const createGRN = useCreateGoodsReceipt()
  const createBillFromGRN = useCreateBillFromGRN()

  const [receiveOpen, setReceiveOpen] = useState(false)
  const [receiveQty, setReceiveQty] = useState<Record<string, string>>({})
  const [matchOpen, setMatchOpen] = useState(false)
  const [matchLines, setMatchLines] = useState<Record<string, MatchLine>>({})
  const [printOpen, setPrintOpen] = useState(false)
  const [downloadingPdf, setDownloadingPdf] = useState(false)

  const supplier = suppliers?.find((s) => s.id === po?.supplier_id)
  const supplierName = supplier?.name
  const materialName = (matId: string) => materials?.find((m) => m.id === matId)?.name ?? matId.slice(0, 8)

  const totals = useMemo(() => {
    if (!po?.items) return { ordered: 0, received: 0, billed: 0 }
    return po.items.reduce(
      (acc, it) => ({
        ordered: acc.ordered + Number(it.qty),
        received: acc.received + Number(it.qty_received),
        billed: acc.billed + Number(it.qty_billed),
      }),
      { ordered: 0, received: 0, billed: 0 },
    )
  }, [po])

  const unbilledGRNItems = useMemo(() => {
    const items: GoodsReceiptItem[] = []
    for (const grn of grns ?? []) {
      for (const it of grn.items ?? []) {
        if (Number(it.qty_received) - Number(it.qty_billed) > 0) items.push(it)
      }
    }
    return items
  }, [grns])

  if (isLoading) return <LoadingState />
  if (error || !po) return <ErrorState message={(error as Error)?.message ?? 'Pesanan pembelian tidak ditemukan'} />

  const receivedPct = totals.ordered > 0 ? Math.round((totals.received / totals.ordered) * 100) : 0
  const billedPct = totals.ordered > 0 ? Math.round((totals.billed / totals.ordered) * 100) : 0

  function openReceiveDialog() {
    const initial: Record<string, string> = {}
    po!.items?.forEach((it) => {
      const remaining = Number(it.qty) - Number(it.qty_received)
      if (remaining > 0) initial[it.id] = ''
    })
    setReceiveQty(initial)
    setReceiveOpen(true)
  }

  function submitReceive() {
    const items = Object.entries(receiveQty)
      .filter(([, qty]) => Number(qty) > 0)
      .map(([purchase_order_item_id, qty]) => ({ purchase_order_item_id, qty_received: qty }))
    if (items.length === 0) return toast.error('Masukkan jumlah diterima untuk minimal satu baris')

    createGRN
      .mutateAsync({ purchase_order_id: id!, items })
      .then((grn) => {
        toast.success('Penerimaan barang diposting', grn.grn_number)
        setReceiveOpen(false)
      })
      .catch((err) => toast.error('Gagal posting penerimaan barang', err instanceof ApiError ? err.message : undefined))
  }

  function openMatchDialog() {
    const initial: Record<string, MatchLine> = {}
    for (const it of unbilledGRNItems) {
      const remaining = Number(it.qty_received) - Number(it.qty_billed)
      initial[it.id] = { grnItem: it, qty: String(remaining), billUnitCost: it.unit_cost }
    }
    setMatchLines(initial)
    setMatchOpen(true)
  }

  function updateMatchLine(id: string, patch: Partial<MatchLine>) {
    setMatchLines((ls) => ({ ...ls, [id]: { ...ls[id], ...patch } }))
  }

  const matchPreview = Object.values(matchLines)
    .filter((l) => Number(l.qty) > 0)
    .map((l) => {
      const qty = Number(l.qty)
      const standardCost = Number(l.grnItem.unit_cost)
      const billCost = Number(l.billUnitCost) || 0
      const standard = qty * standardCost
      const actual = qty * billCost
      return { ...l, qty, standard, actual, variance: actual - standard }
    })
  const previewStandardTotal = matchPreview.reduce((s, l) => s + l.standard, 0)
  const previewActualTotal = matchPreview.reduce((s, l) => s + l.actual, 0)
  const previewVarianceTotal = matchPreview.reduce((s, l) => s + l.variance, 0)

  function submitMatch() {
    const items = matchPreview.map((l) => ({
      goods_receipt_item_id: l.grnItem.id,
      qty: String(l.qty),
      bill_unit_cost: l.billUnitCost,
    }))
    if (items.length === 0) return toast.error('Masukkan jumlah untuk ditagih pada minimal satu baris')

    createBillFromGRN
      .mutateAsync({ purchase_order_id: id!, items })
      .then((bill) => {
        toast.success('Tagihan dicocokkan dan diposting', bill.bill_number)
        setMatchOpen(false)
      })
      .catch((err) => toast.error('Gagal mencocokkan tagihan', err instanceof ApiError ? err.message : undefined))
  }

  async function downloadPoPdf() {
    setDownloadingPdf(true)
    try {
      await openServerPdf(`/purchase-orders/${id}/pdf`, `po-${po!.po_number}.pdf`)
    } catch (err) {
      toast.error('Gagal membuat PDF', err instanceof ApiError ? err.message : undefined)
    } finally {
      setDownloadingPdf(false)
    }
  }

  return (
    <div>
      <button onClick={() => navigate('/purchasing/orders')} className="mb-4 flex items-center gap-1 text-sm text-slate-500 hover:text-slate-800">
        <ArrowLeft className="h-4 w-4" /> Kembali ke Pesanan Pembelian
      </button>

      <PageHeader
        title={po.po_number}
        description={`${supplierName ?? po.supplier_id} · ${formatDate(po.order_date)}`}
        actions={
          <div className="flex gap-2">
            <Button variant="secondary" onClick={() => setPrintOpen(true)}>
              <Printer className="h-4 w-4" /> Cetak PO
            </Button>
            {po.status === 'DRAFT' && (
              <Button onClick={() => approvePO.mutate(po.id, { onSuccess: () => toast.success('Pesanan pembelian disetujui') })} loading={approvePO.isPending}>
                <CheckCircle2 className="h-4 w-4" /> Setujui
              </Button>
            )}
            {(po.status === 'APPROVED' || po.status === 'PARTIALLY_RECEIVED') && (
              <Button onClick={openReceiveDialog}>
                <PackagePlus className="h-4 w-4" /> Terima Barang
              </Button>
            )}
            {unbilledGRNItems.length > 0 && (
              <Button variant="secondary" onClick={openMatchDialog}>
                <Receipt className="h-4 w-4" /> Cocokkan Tagihan
              </Button>
            )}
            {(po.status === 'DRAFT' || po.status === 'APPROVED') && (
              <Button
                variant="destructive"
                onClick={() => cancelPO.mutate(po.id, { onSuccess: () => toast.success('Pesanan pembelian dibatalkan') })}
                loading={cancelPO.isPending}
              >
                <XCircle className="h-4 w-4" /> Batal
              </Button>
            )}
          </div>
        }
      />

      <div className="mb-6 grid grid-cols-1 gap-4 md:grid-cols-3">
        <Card>
          <CardContent className="py-4">
            <p className="text-xs uppercase text-slate-400">Status PO</p>
            <div className="mt-1.5">
              <StatusBadge kind="po" value={po.status} />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4">
            <p className="text-xs uppercase text-slate-400">Diterima</p>
            <div className="mt-1.5 text-sm text-slate-700">
              {totals.received} / {totals.ordered} unit <span className="text-slate-400">({receivedPct}%)</span>
            </div>
            <ProgressBar pct={receivedPct} />
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4">
            <p className="text-xs uppercase text-slate-400">Penagihan</p>
            <div className="mt-1.5 flex items-center gap-2 text-sm">
              <StatusBadge kind="poBilling" value={po.billing_status} />
              <span className="text-slate-400">{billedPct}%</span>
            </div>
            <ProgressBar pct={billedPct} />
          </CardContent>
        </Card>
      </div>

      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Item Baris</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead className="bg-slate-50 text-left text-xs font-medium uppercase text-slate-500">
              <tr>
                <th className="px-4 py-2.5">Bahan Baku</th>
                <th className="px-4 py-2.5">Dipesan</th>
                <th className="px-4 py-2.5">Diterima</th>
                <th className="px-4 py-2.5">Ditagih</th>
                <th className="px-4 py-2.5">Biaya Satuan</th>
                <th className="px-4 py-2.5 text-right">Total Baris</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {po.items?.map((it) => (
                <tr key={it.id}>
                  <td className="px-4 py-2.5">{materialName(it.material_id)}</td>
                  <td className="px-4 py-2.5">{it.qty}</td>
                  <td className="px-4 py-2.5">{it.qty_received}</td>
                  <td className="px-4 py-2.5">{it.qty_billed}</td>
                  <td className="px-4 py-2.5">{formatCurrency(it.unit_cost)}</td>
                  <td className="px-4 py-2.5 text-right font-medium">{formatCurrency(it.line_total)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Penerimaan Barang</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead className="bg-slate-50 text-left text-xs font-medium uppercase text-slate-500">
              <tr>
                <th className="px-4 py-2.5">Nomor GRN</th>
                <th className="px-4 py-2.5">Tanggal Terima</th>
                <th className="px-4 py-2.5">Baris</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {(!grns || grns.length === 0) && (
                <tr>
                  <td colSpan={3} className="px-4 py-6 text-center text-slate-400">
                    Belum ada barang diterima.
                  </td>
                </tr>
              )}
              {grns?.map((g) => (
                <tr key={g.id}>
                  <td className="px-4 py-2.5 font-medium">{g.grn_number}</td>
                  <td className="px-4 py-2.5">{formatDate(g.receipt_date)}</td>
                  <td className="px-4 py-2.5">{g.items?.length ?? 0}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>

      <Dialog open={receiveOpen} onOpenChange={setReceiveOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Terima Barang</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            {po.items
              ?.filter((it) => Number(it.qty) - Number(it.qty_received) > 0)
              .map((it) => {
                const remaining = Number(it.qty) - Number(it.qty_received)
                return (
                  <div key={it.id} className="flex items-center justify-between gap-3">
                    <Label className="flex-1">
                      {materialName(it.material_id)} <span className="text-slate-400">(sisa {remaining})</span>
                    </Label>
                    <Input
                      type="number"
                      min="0"
                      max={remaining}
                      className="w-32"
                      value={receiveQty[it.id] ?? ''}
                      onChange={(e) => setReceiveQty((q) => ({ ...q, [it.id]: e.target.value }))}
                    />
                  </div>
                )
              })}
          </div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setReceiveOpen(false)}>
              Batal
            </Button>
            <Button onClick={submitReceive} loading={createGRN.isPending}>
              Posting Penerimaan Barang
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={matchOpen} onOpenChange={setMatchOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Meja Pencocokan Tagihan Pemasok</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div className="overflow-x-auto rounded-md border border-slate-200">
              <table className="w-full text-sm">
                <thead className="bg-slate-50 text-left text-xs font-medium uppercase text-slate-500">
                  <tr>
                    <th className="px-3 py-2">Bahan Baku</th>
                    <th className="px-3 py-2">Biaya PO</th>
                    <th className="px-3 py-2">Jml Tagihan</th>
                    <th className="px-3 py-2">Biaya Satuan Tagihan</th>
                    <th className="px-3 py-2 text-right">Selisih</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {unbilledGRNItems.map((it) => {
                    const line = matchLines[it.id]
                    if (!line) return null
                    const remaining = Number(it.qty_received) - Number(it.qty_billed)
                    const qty = Number(line.qty) || 0
                    const variance = (Number(line.billUnitCost) - Number(it.unit_cost)) * qty
                    return (
                      <tr key={it.id}>
                        <td className="px-3 py-2">
                          {materialName(it.material_id)}
                          <div className="text-xs text-slate-400">sisa {remaining}</div>
                        </td>
                        <td className="px-3 py-2">{formatCurrency(it.unit_cost)}</td>
                        <td className="px-3 py-1.5">
                          <Input
                            type="number"
                            min="0"
                            max={remaining}
                            className="w-20"
                            value={line.qty}
                            onChange={(e) => updateMatchLine(it.id, { qty: e.target.value })}
                          />
                        </td>
                        <td className="px-3 py-1.5">
                          <Input
                            type="number"
                            className="w-28"
                            value={line.billUnitCost}
                            onChange={(e) => updateMatchLine(it.id, { billUnitCost: e.target.value })}
                          />
                        </td>
                        <td className={`px-3 py-2 text-right ${variance > 0 ? 'text-[var(--color-danger)]' : variance < 0 ? 'text-[var(--color-success)]' : 'text-slate-400'}`}>
                          {formatCurrency(variance)}
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>

            <div className="rounded-md bg-slate-50 p-3 text-sm">
              <p className="mb-1 text-xs font-semibold uppercase text-slate-400">Pratinjau Jurnal</p>
              <div className="flex justify-between"><span>DR Hutang Belum Ditagih</span><span>{formatCurrency(previewStandardTotal)}</span></div>
              {previewVarianceTotal !== 0 && (
                <div className="flex justify-between">
                  <span>{previewVarianceTotal > 0 ? 'DR' : 'CR'} Selisih Harga Pembelian</span>
                  <span>{formatCurrency(Math.abs(previewVarianceTotal))}</span>
                </div>
              )}
              <div className="flex justify-between font-medium"><span>CR Hutang Usaha</span><span>{formatCurrency(previewActualTotal)}</span></div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setMatchOpen(false)}>
              Batal
            </Button>
            <Button onClick={submitMatch} loading={createBillFromGRN.isPending}>
              Posting Tagihan Tercocok
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <PrintPreviewModal
        open={printOpen}
        onOpenChange={setPrintOpen}
        title={`Pesanan Pembelian – ${po.po_number}`}
        onDownloadPdf={downloadPoPdf}
        downloadPending={downloadingPdf}
      >
        <PrintablePurchaseOrder po={po} supplier={supplier} materials={materials} />
      </PrintPreviewModal>
    </div>
  )
}
