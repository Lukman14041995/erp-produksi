import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, Trash2 } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Card, CardContent } from '@/components/ui/Card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/Select'
import { useCreateTransfer, useWarehouses } from '@/hooks/useWarehouse'
import { useMaterials } from '@/hooks/useMaterials'
import { useProducts } from '@/hooks/useProducts'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'

interface LineDraft {
  key: string
  itemType: 'MATERIAL' | 'PRODUCT'
  materialId: string
  productId: string
  productSizeId: string
  qty: string
}

function emptyLine(): LineDraft {
  return { key: crypto.randomUUID(), itemType: 'MATERIAL', materialId: '', productId: '', productSizeId: '', qty: '' }
}

export function StockTransferNewPage() {
  const navigate = useNavigate()
  const { data: warehouses } = useWarehouses()
  const { data: materials } = useMaterials()
  const { data: products } = useProducts()
  const createTransfer = useCreateTransfer()

  const [sourceId, setSourceId] = useState('')
  const [destinationId, setDestinationId] = useState('')
  const [transferDate, setTransferDate] = useState('')
  const [notes, setNotes] = useState('')
  const [lines, setLines] = useState<LineDraft[]>([emptyLine()])

  function updateLine(key: string, patch: Partial<LineDraft>) {
    setLines((ls) => ls.map((l) => (l.key === key ? { ...l, ...patch } : l)))
  }

  function submit() {
    if (!sourceId || !destinationId) return toast.error('Pilih gudang asal dan tujuan')
    if (sourceId === destinationId) return toast.error('Gudang asal dan tujuan harus berbeda')

    const items = lines
      .filter((l) => l.qty && (l.itemType === 'MATERIAL' ? l.materialId : l.productId))
      .map((l) => ({
        item_type: l.itemType,
        material_id: l.itemType === 'MATERIAL' ? l.materialId : undefined,
        product_id: l.itemType === 'PRODUCT' ? l.productId : undefined,
        product_size_id: l.itemType === 'PRODUCT' ? l.productSizeId || undefined : undefined,
        qty: l.qty,
      }))
    if (items.length === 0) return toast.error('Tambahkan minimal satu item baris')

    createTransfer
      .mutateAsync({ source_warehouse_id: sourceId, destination_warehouse_id: destinationId, transfer_date: transferDate || undefined, notes, items })
      .then((t) => {
        toast.success('Transfer stok dibuat', t.transfer_number)
        navigate(`/inventory/transfers/${t.id}`)
      })
      .catch((err) => toast.error('Gagal membuat transfer', err instanceof ApiError ? err.message : undefined))
  }

  return (
    <div className="max-w-3xl">
      <PageHeader title="Transfer Stok Baru" description="Pindahkan bahan baku atau barang jadi antar gudang." />

      <div className="space-y-4">
        <Card>
          <CardContent className="grid grid-cols-2 gap-4 py-4">
            <div className="space-y-1.5">
              <Label>Gudang Asal</Label>
              <Select value={sourceId} onValueChange={setSourceId}>
                <SelectTrigger>
                  <SelectValue placeholder="Pilih asal" />
                </SelectTrigger>
                <SelectContent>
                  {warehouses?.map((w) => (
                    <SelectItem key={w.id} value={w.id}>
                      {w.code} &middot; {w.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>Gudang Tujuan</Label>
              <Select value={destinationId} onValueChange={setDestinationId}>
                <SelectTrigger>
                  <SelectValue placeholder="Pilih tujuan" />
                </SelectTrigger>
                <SelectContent>
                  {warehouses?.map((w) => (
                    <SelectItem key={w.id} value={w.id}>
                      {w.code} &middot; {w.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>Tanggal Transfer (opsional)</Label>
              <Input type="date" value={transferDate} onChange={(e) => setTransferDate(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label>Catatan</Label>
              <Input value={notes} onChange={(e) => setNotes(e.target.value)} />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="py-4">
            <p className="mb-3 text-sm font-medium text-slate-700">Item Baris</p>
            <div className="space-y-2">
              {lines.map((line) => {
                const selectedProduct = products?.find((p) => p.id === line.productId)
                return (
                  <div key={line.key} className="flex items-end gap-2">
                    <div className="w-32 space-y-1">
                      <Label className="text-xs">Jenis</Label>
                      <Select
                        value={line.itemType}
                        onValueChange={(v) => updateLine(line.key, { itemType: v as 'MATERIAL' | 'PRODUCT', materialId: '', productId: '', productSizeId: '' })}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="MATERIAL">Bahan Baku</SelectItem>
                          <SelectItem value="PRODUCT">Produk</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    {line.itemType === 'MATERIAL' ? (
                      <div className="flex-1 space-y-1">
                        <Label className="text-xs">Bahan Baku</Label>
                        <Select value={line.materialId} onValueChange={(v) => updateLine(line.key, { materialId: v })}>
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
                    ) : (
                      <>
                        <div className="flex-1 space-y-1">
                          <Label className="text-xs">Produk</Label>
                          <Select value={line.productId} onValueChange={(v) => updateLine(line.key, { productId: v, productSizeId: '' })}>
                            <SelectTrigger>
                              <SelectValue placeholder="Pilih produk" />
                            </SelectTrigger>
                            <SelectContent>
                              {products?.map((p) => (
                                <SelectItem key={p.id} value={p.id}>
                                  {p.name}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </div>
                        <div className="w-28 space-y-1">
                          <Label className="text-xs">Ukuran</Label>
                          <Select value={line.productSizeId} onValueChange={(v) => updateLine(line.key, { productSizeId: v })}>
                            <SelectTrigger>
                              <SelectValue placeholder="Ukuran" />
                            </SelectTrigger>
                            <SelectContent>
                              {selectedProduct?.sizes?.map((s) => (
                                <SelectItem key={s.id} value={s.id}>
                                  {s.size_code}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </div>
                      </>
                    )}
                    <div className="w-28 space-y-1">
                      <Label className="text-xs">Jml</Label>
                      <Input type="number" value={line.qty} onChange={(e) => updateLine(line.key, { qty: e.target.value })} />
                    </div>
                    <Button type="button" variant="ghost" size="icon" onClick={() => setLines((ls) => ls.filter((l) => l.key !== line.key))}>
                      <Trash2 className="h-4 w-4 text-[var(--color-danger)]" />
                    </Button>
                  </div>
                )
              })}
            </div>
            <Button type="button" variant="secondary" size="sm" className="mt-3" onClick={() => setLines((ls) => [...ls, emptyLine()])}>
              <Plus className="h-3.5 w-3.5" /> Tambah Baris
            </Button>
          </CardContent>
        </Card>

        <div className="flex items-center justify-end gap-2 rounded-lg border border-slate-200 bg-[var(--color-surface)] px-5 py-4">
          <Button type="button" variant="secondary" onClick={() => navigate('/inventory/transfers')}>
            Batal
          </Button>
          <Button type="button" onClick={submit} loading={createTransfer.isPending}>
            Buat Transfer
          </Button>
        </div>
      </div>
    </div>
  )
}
