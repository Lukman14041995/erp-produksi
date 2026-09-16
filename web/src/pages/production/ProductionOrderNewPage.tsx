import { useState } from 'react'
import { uid } from '@/lib/uid'
import { useNavigate } from 'react-router-dom'
import { Plus, Trash2 } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Card, CardContent } from '@/components/ui/Card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/Select'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { useProduct, useProducts } from '@/hooks/useProducts'
import { useMaterials } from '@/hooks/useMaterials'
import { useBOMsByProduct, useCreateBOM, useCreateProductionOrder } from '@/hooks/useProduction'
import { useSalesOrders } from '@/hooks/useSales'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'

interface BomLineDraft {
  key: string
  materialId: string
  qtyPerUnit: string
}

export function ProductionOrderNewPage() {
  const navigate = useNavigate()
  const { data: products } = useProducts()
  const { data: materials } = useMaterials()
  const { data: salesOrders } = useSalesOrders()

  const [productId, setProductId] = useState('')
  const [salesOrderId, setSalesOrderId] = useState<string>('')
  const [bomId, setBomId] = useState('')
  const [plannedBySize, setPlannedBySize] = useState<Record<string, string>>({})
  const [notes, setNotes] = useState('')

  const { data: product } = useProduct(productId || undefined)
  const { data: boms } = useBOMsByProduct(productId || undefined)
  const createOrder = useCreateProductionOrder()

  const [bomDialogOpen, setBomDialogOpen] = useState(false)
  const [bomName, setBomName] = useState('Standard BOM')
  const [bomLines, setBomLines] = useState<BomLineDraft[]>([{ key: uid(), materialId: '', qtyPerUnit: '' }])
  const createBom = useCreateBOM()

  function selectProduct(id: string) {
    setProductId(id)
    setBomId('')
    setPlannedBySize({})
  }

  function submitBom() {
    const lines = bomLines.filter((l) => l.materialId && l.qtyPerUnit)
    if (!productId || lines.length === 0) return toast.error('Tambahkan minimal satu baris bahan')
    createBom
      .mutateAsync({
        product_id: productId,
        name: bomName,
        lines: lines.map((l) => ({ material_id: l.materialId, qty_per_unit: l.qtyPerUnit, uom: 'PCS' })),
      })
      .then((bom) => {
        toast.success('BOM berhasil dibuat')
        setBomId(bom.id)
        setBomDialogOpen(false)
      })
      .catch((err) => toast.error('Gagal membuat BOM', err instanceof ApiError ? err.message : undefined))
  }

  function submitOrder() {
    if (!productId) return toast.error('Pilih produk')
    const items = Object.entries(plannedBySize)
      .filter(([, qty]) => Number(qty) > 0)
      .map(([sizeId, qty]) => ({ product_size_id: sizeId, planned_qty: qty }))
    if (items.length === 0) return toast.error('Masukkan jumlah rencana untuk minimal satu ukuran')

    createOrder
      .mutateAsync({
        product_id: productId,
        sales_order_id: salesOrderId || undefined,
        bom_id: bomId || undefined,
        notes,
        items,
      })
      .then((order) => {
        toast.success('Pesanan produksi berhasil dibuat', order.prod_number)
        navigate(`/production/orders/${order.id}`)
      })
      .catch((err) => toast.error('Gagal membuat pesanan produksi', err instanceof ApiError ? err.message : undefined))
  }

  return (
    <div className="max-w-3xl">
      <PageHeader title="Pesanan Produksi Baru" description="Rencanakan proses produksi dan kebutuhan bahan yang berasal dari BOM." />

      <div className="space-y-4">
        <Card>
          <CardContent className="grid grid-cols-2 gap-4 py-4">
            <div className="space-y-1.5">
              <Label>Produk</Label>
              <Select value={productId} onValueChange={selectProduct}>
                <SelectTrigger>
                  <SelectValue placeholder="Pilih produk" />
                </SelectTrigger>
                <SelectContent>
                  {products?.map((p) => (
                    <SelectItem key={p.id} value={p.id}>
                      {p.code} &middot; {p.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>Pesanan Penjualan Terkait (opsional)</Label>
              <Select value={salesOrderId || '__none'} onValueChange={(v) => setSalesOrderId(v === '__none' ? '' : v)}>
                <SelectTrigger>
                  <SelectValue placeholder="Produksi untuk stok" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__none">Tidak ada (produksi untuk stok)</SelectItem>
                  {salesOrders?.filter((o) => o.order_status === 'CONFIRMED').map((o) => (
                    <SelectItem key={o.id} value={o.id}>
                      {o.so_number}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="col-span-2 space-y-1.5">
              <div className="flex items-center justify-between">
                <Label>Bill of Materials (opsional)</Label>
                {productId && (
                  <button type="button" className="text-xs text-[var(--color-primary)] hover:underline" onClick={() => setBomDialogOpen(true)}>
                    + BOM Baru
                  </button>
                )}
              </div>
              <Select value={bomId || '__none'} onValueChange={(v) => setBomId(v === '__none' ? '' : v)} disabled={!productId}>
                <SelectTrigger>
                  <SelectValue placeholder="Tanpa BOM (bahan dicatat manual)" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__none">Tanpa BOM</SelectItem>
                  {boms?.map((b) => (
                    <SelectItem key={b.id} value={b.id}>
                      {b.name} (v{b.version})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="col-span-2 space-y-1.5">
              <Label>Catatan</Label>
              <Input value={notes} onChange={(e) => setNotes(e.target.value)} />
            </div>
          </CardContent>
        </Card>

        {product?.sizes && product.sizes.length > 0 && (
          <Card>
            <CardContent className="py-4">
              <p className="mb-3 text-sm font-medium text-slate-700">Jumlah Rencana per Ukuran</p>
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                {[...product.sizes].sort((a, b) => a.sort_order - b.sort_order).map((size) => (
                  <div key={size.id} className="space-y-1.5">
                    <Label>{size.size_code} <span className="text-slate-400">(&times;{size.size_multiplier})</span></Label>
                    <Input
                      type="number"
                      min="0"
                      value={plannedBySize[size.id] ?? ''}
                      onChange={(e) => setPlannedBySize((s) => ({ ...s, [size.id]: e.target.value }))}
                    />
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        )}

        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={() => navigate('/production/orders')}>
            Batal
          </Button>
          <Button onClick={submitOrder} loading={createOrder.isPending}>
            Buat Pesanan Produksi
          </Button>
        </div>
      </div>

      <Dialog open={bomDialogOpen} onOpenChange={setBomDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Bill of Materials Baru</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div className="space-y-1.5">
              <Label>Nama</Label>
              <Input value={bomName} onChange={(e) => setBomName(e.target.value)} />
            </div>
            {bomLines.map((line, idx) => (
              <div key={line.key} className="flex items-end gap-2">
                <div className="flex-1 space-y-1.5">
                  <Label>Bahan</Label>
                  <Select
                    value={line.materialId}
                    onValueChange={(v) => setBomLines((ls) => ls.map((l) => (l.key === line.key ? { ...l, materialId: v } : l)))}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Pilih bahan" />
                    </SelectTrigger>
                    <SelectContent>
                      {materials?.map((m) => (
                        <SelectItem key={m.id} value={m.id}>
                          {m.code} &middot; {m.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="w-32 space-y-1.5">
                  <Label>Jml / unit</Label>
                  <Input
                    type="number"
                    step="0.0001"
                    value={line.qtyPerUnit}
                    onChange={(e) => setBomLines((ls) => ls.map((l) => (l.key === line.key ? { ...l, qtyPerUnit: e.target.value } : l)))}
                  />
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() => setBomLines((ls) => ls.filter((_, i) => i !== idx))}
                >
                  <Trash2 className="h-4 w-4 text-[var(--color-danger)]" />
                </Button>
              </div>
            ))}
            <Button type="button" variant="secondary" size="sm" onClick={() => setBomLines((ls) => [...ls, { key: uid(), materialId: '', qtyPerUnit: '' }])}>
              <Plus className="h-3.5 w-3.5" /> Tambah Baris
            </Button>
          </div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setBomDialogOpen(false)}>
              Batal
            </Button>
            <Button onClick={submitBom} loading={createBom.isPending}>
              Simpan BOM
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
