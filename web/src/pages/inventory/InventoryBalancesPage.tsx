import { useState } from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm } from 'react-hook-form'
import { z } from 'zod'
import { Plus } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Badge } from '@/components/ui/Badge'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/Dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/Select'
import { useAdjustInventory, useInventoryBalances } from '@/hooks/useInventory'
import { useWarehouses } from '@/hooks/useWarehouse'
import { useMaterials } from '@/hooks/useMaterials'
import { useProducts } from '@/hooks/useProducts'
import { formatCurrency, formatNumber } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { InventoryBalance, ItemType } from '@/types/inventory'

const schema = z.object({
  item_type: z.enum(['MATERIAL', 'PRODUCT']),
  warehouse_id: z.string().min(1, 'Wajib diisi'),
  material_id: z.string().optional(),
  product_id: z.string().optional(),
  product_size_id: z.string().optional(),
  qty: z.string().min(1, 'Wajib diisi'),
  unit_cost: z.string().min(1, 'Wajib diisi'),
  notes: z.string().optional(),
})
type FormValues = z.infer<typeof schema>

export function InventoryBalancesPage() {
  const { data, isLoading, error } = useInventoryBalances()
  const { data: warehouses } = useWarehouses()
  const { data: materials } = useMaterials()
  const { data: products } = useProducts()
  const adjust = useAdjustInventory()
  const [open, setOpen] = useState(false)

  const { register, handleSubmit, control, watch, reset } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { item_type: 'MATERIAL' },
  })
  const itemType = watch('item_type')
  const productId = watch('product_id')
  const selectedProduct = products?.find((p) => p.id === productId)

  const warehouseName = (id: string) => warehouses?.find((w) => w.id === id)?.name ?? id.slice(0, 8)

  const itemName = (row: InventoryBalance): string => {
    if (row.item_type === 'MATERIAL') return materials?.find((m) => m.id === row.material_id)?.name ?? row.material_id ?? ''
    const p = products?.find((pr) => pr.id === row.product_id)
    const size = p?.sizes?.find((s) => s.id === row.product_size_id)
    return `${p?.name ?? row.product_id}${size ? ' - ' + size.size_code : ''}`
  }

  function onSubmit(values: FormValues) {
    adjust
      .mutateAsync({
        item_type: values.item_type as ItemType,
        warehouse_id: values.warehouse_id,
        material_id: values.item_type === 'MATERIAL' ? values.material_id : undefined,
        product_id: values.item_type === 'PRODUCT' ? values.product_id : undefined,
        product_size_id: values.item_type === 'PRODUCT' ? values.product_size_id : undefined,
        qty: values.qty,
        unit_cost: values.unit_cost,
        notes: values.notes,
      })
      .then(() => {
        toast.success('Persediaan disesuaikan')
        setOpen(false)
        reset({ item_type: 'MATERIAL' })
      })
      .catch((err) => toast.error('Penyesuaian gagal', err instanceof ApiError ? err.message : undefined))
  }

  const itemTypeLabel = (t: ItemType) => (t === 'MATERIAL' ? 'Bahan Baku' : 'Produk')

  const columns: Column<InventoryBalance>[] = [
    { key: 'warehouse', header: 'Gudang', render: (r) => warehouseName(r.warehouse_id), csvValue: (r) => warehouseName(r.warehouse_id) },
    { key: 'item', header: 'Item', render: itemName, csvValue: itemName },
    { key: 'type', header: 'Jenis', render: (r) => <Badge tone={r.item_type === 'MATERIAL' ? 'info' : 'success'}>{itemTypeLabel(r.item_type)}</Badge>, csvValue: (r) => itemTypeLabel(r.item_type) },
    { key: 'qty', header: 'Jml Tersedia', render: (r) => formatNumber(r.qty_on_hand, 2), sortValue: (r) => Number(r.qty_on_hand), csvValue: (r) => r.qty_on_hand },
    { key: 'cost', header: 'Biaya Satuan Rata-rata', render: (r) => formatCurrency(r.avg_unit_cost), sortValue: (r) => Number(r.avg_unit_cost), csvValue: (r) => r.avg_unit_cost },
    {
      key: 'value',
      header: 'Total Nilai',
      render: (r) => formatCurrency(Number(r.qty_on_hand) * Number(r.avg_unit_cost)),
      sortValue: (r) => Number(r.qty_on_hand) * Number(r.avg_unit_cost),
      csvValue: (r) => String(Number(r.qty_on_hand) * Number(r.avg_unit_cost)),
    },
  ]

  return (
    <div>
      <PageHeader
        title="Saldo Persediaan"
        description="Saldo biaya rata-rata bergerak untuk bahan baku dan barang jadi, per gudang."
        actions={
          <Button onClick={() => setOpen(true)}>
            <Plus className="h-4 w-4" /> Sesuaikan Stok
          </Button>
        }
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(r) => `${r.warehouse_id}-${r.item_type}-${r.material_id ?? r.product_id}-${r.product_size_id ?? ''}`}
          exportFilename="inventory-balances"
          pageSize={20}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Penyesuaian Stok Manual</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Jenis Item</Label>
                <Controller
                  control={control}
                  name="item_type"
                  render={({ field }) => (
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="MATERIAL">Bahan Baku</SelectItem>
                        <SelectItem value="PRODUCT">Produk Jadi</SelectItem>
                      </SelectContent>
                    </Select>
                  )}
                />
              </div>
              <div className="space-y-1.5">
                <Label>Gudang</Label>
                <Controller
                  control={control}
                  name="warehouse_id"
                  render={({ field }) => (
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger>
                        <SelectValue placeholder="Pilih gudang" />
                      </SelectTrigger>
                      <SelectContent>
                        {warehouses?.map((w) => (
                          <SelectItem key={w.id} value={w.id}>
                            {w.code} &middot; {w.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                />
              </div>
            </div>

            {itemType === 'MATERIAL' ? (
              <div className="space-y-1.5">
                <Label>Bahan Baku</Label>
                <Controller
                  control={control}
                  name="material_id"
                  render={({ field }) => (
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger>
                        <SelectValue placeholder="Pilih bahan" />
                      </SelectTrigger>
                      <SelectContent>
                        {materials?.map((m) => (
                          <SelectItem key={m.id} value={m.id}>
                            {m.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                />
              </div>
            ) : (
              <>
                <div className="space-y-1.5">
                  <Label>Produk</Label>
                  <Controller
                    control={control}
                    name="product_id"
                    render={({ field }) => (
                      <Select value={field.value} onValueChange={field.onChange}>
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
                    )}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label>Ukuran</Label>
                  <Controller
                    control={control}
                    name="product_size_id"
                    render={({ field }) => (
                      <Select value={field.value} onValueChange={field.onChange}>
                        <SelectTrigger>
                          <SelectValue placeholder="Pilih ukuran" />
                        </SelectTrigger>
                        <SelectContent>
                          {selectedProduct?.sizes?.map((s) => (
                            <SelectItem key={s.id} value={s.id}>
                              {s.size_code}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    )}
                  />
                </div>
              </>
            )}

            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Jml (+ / -)</Label>
                <Input type="number" step="0.0001" {...register('qty')} placeholder="cth. 50 atau -10" />
              </div>
              <div className="space-y-1.5">
                <Label>Biaya Satuan</Label>
                <Input type="number" step="0.0001" {...register('unit_cost')} />
              </div>
            </div>
            <div className="space-y-1.5">
              <Label>Catatan</Label>
              <Input {...register('notes')} />
            </div>
            <DialogFooter>
              <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
                Batal
              </Button>
              <Button type="submit" loading={adjust.isPending}>
                Terapkan Penyesuaian
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
