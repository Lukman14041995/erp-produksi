import { useMemo, useState } from 'react'
import { uid } from '@/lib/uid'
import { useNavigate } from 'react-router-dom'
import { Plus, Trash2 } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Textarea } from '@/components/ui/Input'
import { Card, CardContent } from '@/components/ui/Card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/Select'
import { useCustomers } from '@/hooks/useCustomers'
import { useProduct, useProducts } from '@/hooks/useProducts'
import { useCreateSalesOrder } from '@/hooks/useSales'
import { formatCurrency } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { CreateOrderItemInput } from '@/types/sales'
import type { Product, ProductSize } from '@/types/master'

interface SizeRow {
  qty: string
  unitPrice: string
  discount: string
  taxRate: string
}

interface ProductLine {
  key: string
  productId: string
  rows: Record<string, SizeRow>
}

function emptyRow(unitPrice = ''): SizeRow {
  return { qty: '', unitPrice, discount: '0', taxRate: '0.11' }
}

// The size's `size_multiplier` scales BOM material consumption for
// production planning, but it also doubles as a sensible default sell-price
// scaler here (bigger sizes use more fabric, so they typically sell for
// more) -- still just a starting point, each row's price stays editable.
function defaultUnitPrice(product: Product | undefined, size: ProductSize): string {
  if (!product?.base_price) return ''
  const price = Number(product.base_price) * Number(size.size_multiplier)
  return String(Math.round(price))
}

function ProductLineEditor({
  line,
  onChange,
  onRemove,
}: {
  line: ProductLine
  onChange: (line: ProductLine) => void
  onRemove: () => void
}) {
  const { data: products } = useProducts()
  const { data: product } = useProduct(line.productId || undefined)

  function setProduct(productId: string) {
    onChange({ ...line, productId, rows: {} })
  }

  function setRow(sizeId: string, patch: Partial<SizeRow>) {
    const size = product?.sizes?.find((s) => s.id === sizeId)
    const current = line.rows[sizeId] ?? emptyRow(size ? defaultUnitPrice(product, size) : '')
    onChange({ ...line, rows: { ...line.rows, [sizeId]: { ...current, ...patch } } })
  }

  const lineTotal = useMemo(() => {
    if (!product?.sizes) return 0
    return product.sizes.reduce((sum, size) => {
      const row = line.rows[size.id]
      if (!row?.qty) return sum
      const qty = Number(row.qty) || 0
      const price = Number(row.unitPrice) || 0
      const discount = Number(row.discount) || 0
      const tax = Number(row.taxRate) || 0
      const afterDiscount = qty * price - discount
      return sum + afterDiscount + afterDiscount * tax
    }, 0)
  }, [line, product])

  return (
    <Card>
      <CardContent className="space-y-3 py-4">
        <div className="flex items-end gap-3">
          <div className="flex-1 space-y-1.5">
            <Label>Produk</Label>
            <Select value={line.productId} onValueChange={setProduct}>
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
          <Button type="button" variant="ghost" size="icon" onClick={onRemove}>
            <Trash2 className="h-4 w-4 text-[var(--color-danger)]" />
          </Button>
        </div>

        {product?.sizes && product.sizes.length > 0 && (
          <div className="overflow-x-auto rounded-md border border-slate-200">
            <table className="w-full text-sm">
              <thead className="bg-slate-50 text-left text-xs font-medium uppercase text-slate-500">
                <tr>
                  <th className="px-3 py-2">Ukuran</th>
                  <th className="px-3 py-2">Jml</th>
                  <th className="px-3 py-2">Harga Satuan</th>
                  <th className="px-3 py-2">Diskon</th>
                  <th className="px-3 py-2">Tarif Pajak</th>
                  <th className="px-3 py-2 text-right">Total Baris</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {[...product.sizes]
                  .sort((a, b) => a.sort_order - b.sort_order)
                  .map((size) => {
                    const row = line.rows[size.id] ?? emptyRow(defaultUnitPrice(product, size))
                    const qty = Number(row.qty) || 0
                    const price = Number(row.unitPrice) || 0
                    const discount = Number(row.discount) || 0
                    const tax = Number(row.taxRate) || 0
                    const afterDiscount = qty * price - discount
                    const rowTotal = afterDiscount + afterDiscount * tax
                    return (
                      <tr key={size.id}>
                        <td className="px-3 py-2 font-medium">
                          {size.size_code} <span className="text-xs text-slate-400">&times;{size.size_multiplier}</span>
                        </td>
                        <td className="px-3 py-1.5">
                          <Input
                            type="number"
                            min="0"
                            className="w-20"
                            value={row.qty}
                            onChange={(e) => setRow(size.id, { qty: e.target.value })}
                          />
                        </td>
                        <td className="px-3 py-1.5">
                          <Input
                            type="number"
                            min="0"
                            className="w-28"
                            value={row.unitPrice}
                            onChange={(e) => setRow(size.id, { unitPrice: e.target.value })}
                          />
                        </td>
                        <td className="px-3 py-1.5">
                          <Input
                            type="number"
                            min="0"
                            className="w-24"
                            value={row.discount}
                            onChange={(e) => setRow(size.id, { discount: e.target.value })}
                          />
                        </td>
                        <td className="px-3 py-1.5">
                          <Input
                            type="number"
                            step="0.01"
                            min="0"
                            className="w-20"
                            value={row.taxRate}
                            onChange={(e) => setRow(size.id, { taxRate: e.target.value })}
                          />
                        </td>
                        <td className="px-3 py-2 text-right font-medium">{qty > 0 ? formatCurrency(rowTotal) : '-'}</td>
                      </tr>
                    )
                  })}
              </tbody>
            </table>
          </div>
        )}
        {product && (!product.sizes || product.sizes.length === 0) && (
          <p className="text-sm text-slate-400">Produk ini belum memiliki ukuran. Tambahkan ukuran di Data Master &rarr; Produk.</p>
        )}
        {line.productId && (
          <p className="text-right text-sm font-medium text-slate-700">Subtotal baris: {formatCurrency(lineTotal)}</p>
        )}
      </CardContent>
    </Card>
  )
}

export function SalesOrderNewPage() {
  const navigate = useNavigate()
  const { data: customers } = useCustomers()
  const createOrder = useCreateSalesOrder()

  const [customerId, setCustomerId] = useState('')
  const [notes, setNotes] = useState('')
  const [lines, setLines] = useState<ProductLine[]>([{ key: uid(), productId: '', rows: {} }])

  function addLine() {
    setLines((ls) => [...ls, { key: uid(), productId: '', rows: {} }])
  }

  function updateLine(key: string, next: ProductLine) {
    setLines((ls) => ls.map((l) => (l.key === key ? next : l)))
  }

  function removeLine(key: string) {
    setLines((ls) => ls.filter((l) => l.key !== key))
  }

  function buildItems(): CreateOrderItemInput[] {
    const items: CreateOrderItemInput[] = []
    for (const line of lines) {
      if (!line.productId) continue
      for (const [sizeId, row] of Object.entries(line.rows)) {
        const qty = Number(row.qty)
        if (!qty || qty <= 0) continue
        items.push({
          product_id: line.productId,
          product_size_id: sizeId,
          qty: row.qty,
          unit_price: row.unitPrice || '0',
          discount: row.discount || '0',
          tax_rate: row.taxRate || '0',
        })
      }
    }
    return items
  }

  const items = buildItems()
  const grandTotal = items.reduce((sum, it) => {
    const after = Number(it.qty) * Number(it.unit_price) - Number(it.discount)
    return sum + after + after * Number(it.tax_rate)
  }, 0)

  function handleSubmit() {
    if (!customerId) return toast.error('Pilih pelanggan terlebih dahulu')
    if (items.length === 0) return toast.error('Tambahkan minimal satu item dengan jumlah')

    createOrder
      .mutateAsync({ customer_id: customerId, notes, items })
      .then((order) => {
        toast.success('Pesanan penjualan dibuat', order.so_number)
        navigate(`/sales/orders/${order.id}`)
      })
      .catch((err) => toast.error('Gagal membuat pesanan', err instanceof ApiError ? err.message : undefined))
  }

  return (
    <div className="max-w-4xl">
      <PageHeader title="Pesanan Penjualan Baru" description="Buat pesanan matriks ukuran untuk satu atau lebih produk jersey/pakaian." />

      <div className="space-y-4">
        <Card>
          <CardContent className="grid grid-cols-2 gap-4 py-4">
            <div className="space-y-1.5">
              <Label>Pelanggan</Label>
              <Select value={customerId} onValueChange={setCustomerId}>
                <SelectTrigger>
                  <SelectValue placeholder="Pilih pelanggan" />
                </SelectTrigger>
                <SelectContent>
                  {customers?.map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.code} &middot; {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>Catatan</Label>
              <Textarea rows={1} value={notes} onChange={(e) => setNotes(e.target.value)} />
            </div>
          </CardContent>
        </Card>

        {lines.map((line) => (
          <ProductLineEditor key={line.key} line={line} onChange={(l) => updateLine(line.key, l)} onRemove={() => removeLine(line.key)} />
        ))}

        <Button type="button" variant="secondary" onClick={addLine}>
          <Plus className="h-4 w-4" /> Tambah Produk
        </Button>

        <div className="flex items-center justify-between rounded-lg border border-slate-200 bg-[var(--color-surface)] px-5 py-4">
          <div>
            <p className="text-xs uppercase tracking-wide text-slate-400">Total Keseluruhan ({items.length} baris)</p>
            <p className="text-xl font-semibold text-slate-900">{formatCurrency(grandTotal)}</p>
          </div>
          <div className="flex gap-2">
            <Button type="button" variant="secondary" onClick={() => navigate('/sales/orders')}>
              Batal
            </Button>
            <Button type="button" onClick={handleSubmit} loading={createOrder.isPending}>
              Buat Pesanan
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
