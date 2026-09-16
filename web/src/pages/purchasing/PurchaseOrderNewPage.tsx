import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, Trash2 } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Card, CardContent } from '@/components/ui/Card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/Select'
import { useSuppliers } from '@/hooks/useSuppliers'
import { useMaterials } from '@/hooks/useMaterials'
import { useCreatePurchaseOrder } from '@/hooks/usePurchasing'
import { formatCurrency } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'

interface LineDraft {
  key: string
  materialId: string
  qty: string
  unitCost: string
}

export function PurchaseOrderNewPage() {
  const navigate = useNavigate()
  const { data: suppliers } = useSuppliers()
  const { data: materials } = useMaterials()
  const createPO = useCreatePurchaseOrder()

  const [supplierId, setSupplierId] = useState('')
  const [expectedDate, setExpectedDate] = useState('')
  const [notes, setNotes] = useState('')
  const [lines, setLines] = useState<LineDraft[]>([{ key: crypto.randomUUID(), materialId: '', qty: '', unitCost: '' }])

  const grandTotal = lines.reduce((sum, l) => sum + (Number(l.qty) || 0) * (Number(l.unitCost) || 0), 0)

  function submit() {
    const items = lines
      .filter((l) => l.materialId && l.qty && l.unitCost)
      .map((l) => ({ material_id: l.materialId, qty: l.qty, unit_cost: l.unitCost }))
    if (!supplierId || items.length === 0) return toast.error('Pilih pemasok dan tambahkan minimal satu item')

    createPO
      .mutateAsync({ supplier_id: supplierId, expected_date: expectedDate || undefined, notes, items })
      .then((po) => {
        toast.success('Pesanan pembelian dibuat', po.po_number)
        navigate(`/purchasing/orders/${po.id}`)
      })
      .catch((err) => toast.error('Gagal membuat pesanan pembelian', err instanceof ApiError ? err.message : undefined))
  }

  return (
    <div className="max-w-3xl">
      <PageHeader title="Pesanan Pembelian Baru" description="Sepakati biaya standar dengan pemasok sebelum barang diterima." />

      <div className="space-y-4">
        <Card>
          <CardContent className="grid grid-cols-2 gap-4 py-4">
            <div className="space-y-1.5">
              <Label>Pemasok</Label>
              <Select value={supplierId} onValueChange={setSupplierId}>
                <SelectTrigger>
                  <SelectValue placeholder="Pilih pemasok" />
                </SelectTrigger>
                <SelectContent>
                  {suppliers?.map((s) => (
                    <SelectItem key={s.id} value={s.id}>
                      {s.code} &middot; {s.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>Tanggal Diharapkan (opsional)</Label>
              <Input type="date" value={expectedDate} onChange={(e) => setExpectedDate(e.target.value)} />
            </div>
            <div className="col-span-2 space-y-1.5">
              <Label>Catatan</Label>
              <Input value={notes} onChange={(e) => setNotes(e.target.value)} />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="py-4">
            <p className="mb-3 text-sm font-medium text-slate-700">Item Baris</p>
            <div className="space-y-2">
              {lines.map((line) => (
                <div key={line.key} className="flex items-end gap-2">
                  <div className="flex-1 space-y-1">
                    <Label className="text-xs">Bahan Baku</Label>
                    <Select value={line.materialId} onValueChange={(v) => setLines((ls) => ls.map((l) => (l.key === line.key ? { ...l, materialId: v } : l)))}>
                      <SelectTrigger>
                        <SelectValue placeholder="Pilih bahan baku" />
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
                  <div className="w-28 space-y-1">
                    <Label className="text-xs">Jml</Label>
                    <Input type="number" value={line.qty} onChange={(e) => setLines((ls) => ls.map((l) => (l.key === line.key ? { ...l, qty: e.target.value } : l)))} />
                  </div>
                  <div className="w-32 space-y-1">
                    <Label className="text-xs">Biaya Satuan Disepakati</Label>
                    <Input
                      type="number"
                      value={line.unitCost}
                      onChange={(e) => setLines((ls) => ls.map((l) => (l.key === line.key ? { ...l, unitCost: e.target.value } : l)))}
                    />
                  </div>
                  <div className="w-32 pb-2 text-right text-sm text-slate-500">
                    {formatCurrency((Number(line.qty) || 0) * (Number(line.unitCost) || 0))}
                  </div>
                  <Button type="button" variant="ghost" size="icon" onClick={() => setLines((ls) => ls.filter((l) => l.key !== line.key))}>
                    <Trash2 className="h-4 w-4 text-[var(--color-danger)]" />
                  </Button>
                </div>
              ))}
            </div>
            <Button
              type="button"
              variant="secondary"
              size="sm"
              className="mt-3"
              onClick={() => setLines((ls) => [...ls, { key: crypto.randomUUID(), materialId: '', qty: '', unitCost: '' }])}
            >
              <Plus className="h-3.5 w-3.5" /> Tambah Baris
            </Button>
          </CardContent>
        </Card>

        <div className="flex items-center justify-between rounded-lg border border-slate-200 bg-white px-5 py-4">
          <div>
            <p className="text-xs uppercase tracking-wide text-slate-400">Total Keseluruhan</p>
            <p className="text-xl font-semibold text-slate-900">{formatCurrency(grandTotal)}</p>
          </div>
          <div className="flex gap-2">
            <Button type="button" variant="secondary" onClick={() => navigate('/purchasing/orders')}>
              Batal
            </Button>
            <Button type="button" onClick={submit} loading={createPO.isPending}>
              Buat Pesanan Pembelian
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
