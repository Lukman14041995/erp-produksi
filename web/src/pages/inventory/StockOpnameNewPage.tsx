import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { PageHeader } from '@/components/PageHeader'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { Card, CardContent } from '@/components/ui/Card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/Select'
import { LoadingState } from '@/components/QueryState'
import { useStockByWarehouse } from '@/hooks/useInventory'
import { useCreateOpname, useWarehouses } from '@/hooks/useWarehouse'
import { useMaterials } from '@/hooks/useMaterials'
import { useProducts } from '@/hooks/useProducts'
import { formatCurrency, formatNumber } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'
import type { InventoryBalance } from '@/types/inventory'

const INVENTORY_ACCOUNTS: Record<'MATERIAL' | 'PRODUCT', { code: string; label: string }> = {
  MATERIAL: { code: '1-1200', label: 'Persediaan Bahan Baku' },
  PRODUCT: { code: '1-1220', label: 'Persediaan Produk Jadi' },
}
const VARIANCE_ACCOUNT = { code: '5-9200', label: 'Selisih Stock Opname' }

interface CountRow {
  balance: InventoryBalance
  included: boolean
  actualQty: string
}

function rowKey(b: InventoryBalance): string {
  return `${b.item_type}-${b.material_id ?? b.product_id}-${b.product_size_id ?? ''}`
}

export function StockOpnameNewPage() {
  const navigate = useNavigate()
  const { data: warehouses } = useWarehouses()
  const { data: materials } = useMaterials()
  const { data: products } = useProducts()
  const [warehouseId, setWarehouseId] = useState('')
  const [opnameDate, setOpnameDate] = useState('')
  const [notes, setNotes] = useState('')
  const { data: balances, isLoading: balancesLoading } = useStockByWarehouse(warehouseId || undefined)
  const createOpname = useCreateOpname()

  const [rows, setRows] = useState<Record<string, CountRow>>({})

  const currentBalances = warehouseId ? balances ?? [] : []

  function ensureRows(bals: InventoryBalance[]) {
    setRows((prev) => {
      const next = { ...prev }
      for (const b of bals) {
        const key = rowKey(b)
        if (!next[key]) next[key] = { balance: b, included: false, actualQty: b.qty_on_hand }
      }
      return next
    })
  }

  if (warehouseId && balances && Object.keys(rows).length === 0 && balances.length > 0) {
    ensureRows(balances)
  }

  function onWarehouseChange(v: string) {
    setWarehouseId(v)
    setRows({})
  }

  const itemName = (b: InventoryBalance): string => {
    if (b.item_type === 'MATERIAL') return materials?.find((m) => m.id === b.material_id)?.name ?? b.material_id?.slice(0, 8) ?? ''
    const p = products?.find((pr) => pr.id === b.product_id)
    const size = p?.sizes?.find((s) => s.id === b.product_size_id)
    return `${p?.name ?? b.product_id}${size ? ' - ' + size.size_code : ''}`
  }

  const countedRows = useMemo(() => Object.values(rows).filter((r) => r.included), [rows])

  const preview = useMemo(() => {
    return countedRows.map((r) => {
      const systemQty = Number(r.balance.qty_on_hand)
      const actualQty = Number(r.actualQty) || 0
      const varianceQty = actualQty - systemQty
      const varianceAmount = varianceQty * Number(r.balance.avg_unit_cost)
      return { row: r, systemQty, actualQty, varianceQty, varianceAmount }
    })
  }, [countedRows])

  const journalTotals = useMemo(() => {
    const totals: Record<string, { debit: number; credit: number }> = {}
    const add = (code: string, debit: number, credit: number) => {
      if (!totals[code]) totals[code] = { debit: 0, credit: 0 }
      totals[code].debit += debit
      totals[code].credit += credit
    }
    for (const p of preview) {
      if (p.varianceAmount === 0) continue
      const invAcct = INVENTORY_ACCOUNTS[p.row.balance.item_type].code
      const amt = Math.abs(p.varianceAmount)
      if (p.varianceAmount > 0) {
        add(invAcct, amt, 0)
        add(VARIANCE_ACCOUNT.code, 0, amt)
      } else {
        add(VARIANCE_ACCOUNT.code, amt, 0)
        add(invAcct, 0, amt)
      }
    }
    return totals
  }, [preview])

  const journalLines = Object.entries(journalTotals)
    .map(([code, t]) => ({ code, net: t.debit - t.credit }))
    .filter((l) => l.net !== 0)

  function submit() {
    if (!warehouseId) return toast.error('Pilih gudang')
    const items = countedRows.map((r) => ({
      item_type: r.balance.item_type,
      material_id: r.balance.item_type === 'MATERIAL' ? r.balance.material_id : undefined,
      product_id: r.balance.item_type === 'PRODUCT' ? r.balance.product_id : undefined,
      product_size_id: r.balance.item_type === 'PRODUCT' ? r.balance.product_size_id : undefined,
      actual_qty: r.actualQty,
    }))
    if (items.length === 0) return toast.error('Pilih minimal satu item untuk dihitung')

    createOpname
      .mutateAsync({ warehouse_id: warehouseId, opname_date: opnameDate || undefined, notes, items })
      .then((o) => {
        toast.success('Stock opname dibuat', o.opname_number)
        navigate(`/inventory/opnames/${o.id}`)
      })
      .catch((err) => toast.error('Gagal membuat opname', err instanceof ApiError ? err.message : undefined))
  }

  return (
    <div className="max-w-4xl">
      <PageHeader title="Meja Kerja Stock Opname" description="Hitung stok fisik dan pratinjau jurnal penyesuaian yang dihasilkan sebelum menyimpan." />

      <div className="space-y-4">
        <Card>
          <CardContent className="grid grid-cols-3 gap-4 py-4">
            <div className="space-y-1.5">
              <Label>Gudang</Label>
              <Select value={warehouseId} onValueChange={onWarehouseChange}>
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
            </div>
            <div className="space-y-1.5">
              <Label>Tanggal Opname (opsional)</Label>
              <Input type="date" value={opnameDate} onChange={(e) => setOpnameDate(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label>Catatan</Label>
              <Input value={notes} onChange={(e) => setNotes(e.target.value)} />
            </div>
          </CardContent>
        </Card>

        {!warehouseId && <p className="text-sm text-slate-400">Pilih gudang untuk memuat saldo stok terkininya.</p>}
        {warehouseId && balancesLoading && <LoadingState />}

        {warehouseId && !balancesLoading && (
          <Card>
            <CardContent className="p-0">
              <table className="w-full text-sm">
                <thead className="bg-slate-50 text-left text-xs font-medium uppercase text-slate-500">
                  <tr>
                    <th className="px-3 py-2.5 w-10"></th>
                    <th className="px-3 py-2.5">Item</th>
                    <th className="px-3 py-2.5">Jml Sistem</th>
                    <th className="px-3 py-2.5">Jml Aktual</th>
                    <th className="px-3 py-2.5">Selisih</th>
                    <th className="px-3 py-2.5 text-right">Selisih Nilai</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {currentBalances.length === 0 && (
                    <tr>
                      <td colSpan={6} className="px-4 py-6 text-center text-slate-400">
                        Tidak ada saldo stok di gudang ini.
                      </td>
                    </tr>
                  )}
                  {currentBalances.map((b) => {
                    const key = rowKey(b)
                    const row = rows[key]
                    if (!row) return null
                    const systemQty = Number(b.qty_on_hand)
                    const actualQty = Number(row.actualQty) || 0
                    const varianceQty = actualQty - systemQty
                    const varianceAmount = varianceQty * Number(b.avg_unit_cost)
                    return (
                      <tr key={key} className={row.included ? '' : 'text-slate-400'}>
                        <td className="px-3 py-2">
                          <input
                            type="checkbox"
                            className="h-4 w-4 rounded border-slate-300 text-[var(--color-primary)] focus:ring-[var(--color-primary)]"
                            checked={row.included}
                            onChange={(e) => setRows((rs) => ({ ...rs, [key]: { ...rs[key], included: e.target.checked } }))}
                          />
                        </td>
                        <td className="px-3 py-2">{itemName(b)}</td>
                        <td className="px-3 py-2">{formatNumber(b.qty_on_hand, 2)}</td>
                        <td className="px-3 py-1.5">
                          <Input
                            type="number"
                            step="0.0001"
                            className="w-28"
                            disabled={!row.included}
                            value={row.actualQty}
                            onChange={(e) => setRows((rs) => ({ ...rs, [key]: { ...rs[key], actualQty: e.target.value } }))}
                          />
                        </td>
                        <td className={`px-3 py-2 ${row.included && varianceQty !== 0 ? (varianceQty > 0 ? 'text-[var(--color-success)]' : 'text-[var(--color-danger)]') : ''}`}>
                          {row.included ? formatNumber(varianceQty, 2) : '-'}
                        </td>
                        <td
                          className={`px-3 py-2 text-right ${row.included && varianceAmount !== 0 ? (varianceAmount > 0 ? 'text-[var(--color-success)]' : 'text-[var(--color-danger)]') : ''}`}
                        >
                          {row.included ? formatCurrency(varianceAmount) : '-'}
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </CardContent>
          </Card>
        )}

        {countedRows.length > 0 && (
          <div className="rounded-md bg-slate-50 p-3 text-sm">
            <p className="mb-1 text-xs font-semibold uppercase text-slate-400">Pratinjau Jurnal</p>
            {journalLines.length === 0 && <p className="text-slate-400">Tidak ada selisih &mdash; tidak ada jurnal yang akan diposting.</p>}
            {journalLines.map((l) => {
              const label = l.code === VARIANCE_ACCOUNT.code ? VARIANCE_ACCOUNT.label : (INVENTORY_ACCOUNTS.MATERIAL.code === l.code ? INVENTORY_ACCOUNTS.MATERIAL.label : INVENTORY_ACCOUNTS.PRODUCT.label)
              return (
                <div key={l.code} className="flex justify-between">
                  <span>
                    {l.net > 0 ? 'DR' : 'CR'} {l.code} &middot; {label}
                  </span>
                  <span>{formatCurrency(Math.abs(l.net))}</span>
                </div>
              )
            })}
          </div>
        )}

        <div className="flex items-center justify-end gap-2 rounded-lg border border-slate-200 bg-[var(--color-surface)] px-5 py-4">
          <Button type="button" variant="secondary" onClick={() => navigate('/inventory/opnames')}>
            Batal
          </Button>
          <Button type="button" onClick={submit} loading={createOpname.isPending}>
            Simpan Opname (Draf)
          </Button>
        </div>
      </div>
    </div>
  )
}
