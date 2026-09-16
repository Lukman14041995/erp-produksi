import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, CheckCircle2, XCircle } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { StatusBadge } from '@/components/StatusBadge'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { useCancelOpname, usePostOpname, useStockOpname, useWarehouses } from '@/hooks/useWarehouse'
import { useMaterials } from '@/hooks/useMaterials'
import { useProducts } from '@/hooks/useProducts'
import { formatCurrency, formatDate, formatDateTime, formatNumber } from '@/lib/format'
import { toast } from '@/store/toastStore'
import type { StockOpnameItem } from '@/types/warehouse'

export function StockOpnameDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: opname, isLoading, error } = useStockOpname(id)
  const { data: warehouses } = useWarehouses()
  const { data: materials } = useMaterials()
  const { data: products } = useProducts()

  const postOpname = usePostOpname()
  const cancelOpname = useCancelOpname()

  const warehouseName = (wid: string) => warehouses?.find((w) => w.id === wid)?.name ?? wid.slice(0, 8)
  const itemName = (it: StockOpnameItem): string => {
    if (it.item_type === 'MATERIAL') return materials?.find((m) => m.id === it.material_id)?.name ?? it.material_id?.slice(0, 8) ?? ''
    const p = products?.find((pr) => pr.id === it.product_id)
    const size = p?.sizes?.find((s) => s.id === it.product_size_id)
    return `${p?.name ?? it.product_id}${size ? ' - ' + size.size_code : ''}`
  }

  if (isLoading) return <LoadingState />
  if (error || !opname) return <ErrorState message={(error as Error)?.message ?? 'Stock opname tidak ditemukan'} />

  const totalVariance = (opname.items ?? []).reduce((sum, it) => sum + Number(it.variance_amount), 0)

  return (
    <div>
      <button onClick={() => navigate('/inventory/opnames')} className="mb-4 flex items-center gap-1 text-sm text-slate-500 hover:text-slate-800">
        <ArrowLeft className="h-4 w-4" /> Kembali ke Stock Opname
      </button>

      <PageHeader
        title={opname.opname_number}
        description={`${warehouseName(opname.warehouse_id)} · ${formatDate(opname.opname_date)}`}
        actions={
          <div className="flex gap-2">
            {opname.status === 'DRAFT' && (
              <Button onClick={() => postOpname.mutate(opname.id, { onSuccess: () => toast.success('Opname diposting') })} loading={postOpname.isPending}>
                <CheckCircle2 className="h-4 w-4" /> Posting
              </Button>
            )}
            {opname.status === 'DRAFT' && (
              <Button
                variant="destructive"
                onClick={() => cancelOpname.mutate(opname.id, { onSuccess: () => toast.success('Opname dibatalkan') })}
                loading={cancelOpname.isPending}
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
            <p className="text-xs uppercase text-slate-400">Status</p>
            <div className="mt-1.5">
              <StatusBadge kind="opname" value={opname.status} />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4">
            <p className="text-xs uppercase text-slate-400">Selisih Bersih</p>
            <p className={`mt-1.5 text-sm font-medium ${totalVariance > 0 ? 'text-[var(--color-success)]' : totalVariance < 0 ? 'text-[var(--color-danger)]' : 'text-slate-700'}`}>
              {formatCurrency(totalVariance)}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4">
            <p className="text-xs uppercase text-slate-400">Diposting Pada</p>
            <p className="mt-1.5 text-sm text-slate-700">{opname.posted_at ? formatDateTime(opname.posted_at) : '-'}</p>
          </CardContent>
        </Card>
      </div>

      {opname.status === 'POSTED' && (
        <div className="mb-6 rounded-md border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800">
          Opname ini telah diposting dan entri jurnalnya telah dibuat. Data sekarang tidak dapat diubah.
          {opname.journal_id && (
            <>
              {' '}
              <button className="font-medium underline" onClick={() => navigate(`/accounting/journals/${opname.journal_id}`)}>
                Lihat entri jurnal
              </button>
            </>
          )}
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Item Dihitung</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead className="bg-slate-50 text-left text-xs font-medium uppercase text-slate-500">
              <tr>
                <th className="px-4 py-2.5">Item</th>
                <th className="px-4 py-2.5">Jml Sistem</th>
                <th className="px-4 py-2.5">Jml Aktual</th>
                <th className="px-4 py-2.5">Selisih Jml</th>
                <th className="px-4 py-2.5">Biaya Satuan</th>
                <th className="px-4 py-2.5 text-right">Selisih Nilai</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {opname.items?.map((it) => {
                const variance = Number(it.variance_amount)
                return (
                  <tr key={it.id}>
                    <td className="px-4 py-2.5">{itemName(it)}</td>
                    <td className="px-4 py-2.5">{formatNumber(it.system_qty, 2)}</td>
                    <td className="px-4 py-2.5">{formatNumber(it.actual_qty, 2)}</td>
                    <td className={`px-4 py-2.5 ${Number(it.variance_qty) > 0 ? 'text-[var(--color-success)]' : Number(it.variance_qty) < 0 ? 'text-[var(--color-danger)]' : ''}`}>
                      {formatNumber(it.variance_qty, 2)}
                    </td>
                    <td className="px-4 py-2.5">{formatCurrency(it.unit_cost)}</td>
                    <td className={`px-4 py-2.5 text-right font-medium ${variance > 0 ? 'text-[var(--color-success)]' : variance < 0 ? 'text-[var(--color-danger)]' : ''}`}>
                      {formatCurrency(it.variance_amount)}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  )
}
