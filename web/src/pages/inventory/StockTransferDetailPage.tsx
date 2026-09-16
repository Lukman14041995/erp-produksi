import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, PackageCheck, Send, XCircle } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { StatusBadge } from '@/components/StatusBadge'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { useCancelTransfer, useDispatchTransfer, useReceiveTransfer, useStockTransfer, useWarehouses } from '@/hooks/useWarehouse'
import { useMaterials } from '@/hooks/useMaterials'
import { useProducts } from '@/hooks/useProducts'
import { formatCurrency, formatDate, formatDateTime } from '@/lib/format'
import { toast } from '@/store/toastStore'
import type { StockTransferItem } from '@/types/warehouse'

export function StockTransferDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: transfer, isLoading, error } = useStockTransfer(id)
  const { data: warehouses } = useWarehouses()
  const { data: materials } = useMaterials()
  const { data: products } = useProducts()

  const dispatchTransfer = useDispatchTransfer()
  const receiveTransfer = useReceiveTransfer()
  const cancelTransfer = useCancelTransfer()

  const warehouseName = (wid: string) => warehouses?.find((w) => w.id === wid)?.name ?? wid.slice(0, 8)
  const itemName = (it: StockTransferItem): string => {
    if (it.item_type === 'MATERIAL') return materials?.find((m) => m.id === it.material_id)?.name ?? it.material_id?.slice(0, 8) ?? ''
    const p = products?.find((pr) => pr.id === it.product_id)
    const size = p?.sizes?.find((s) => s.id === it.product_size_id)
    return `${p?.name ?? it.product_id}${size ? ' - ' + size.size_code : ''}`
  }

  if (isLoading) return <LoadingState />
  if (error || !transfer) return <ErrorState message={(error as Error)?.message ?? 'Transfer stok tidak ditemukan'} />

  return (
    <div>
      <button onClick={() => navigate('/inventory/transfers')} className="mb-4 flex items-center gap-1 text-sm text-slate-500 hover:text-slate-800">
        <ArrowLeft className="h-4 w-4" /> Kembali ke Transfer Stok
      </button>

      <PageHeader
        title={transfer.transfer_number}
        description={`${warehouseName(transfer.source_warehouse_id)} → ${warehouseName(transfer.destination_warehouse_id)} · ${formatDate(transfer.transfer_date)}`}
        actions={
          <div className="flex gap-2">
            {transfer.status === 'DRAFT' && (
              <Button
                onClick={() => dispatchTransfer.mutate(transfer.id, { onSuccess: () => toast.success('Transfer dikirim') })}
                loading={dispatchTransfer.isPending}
              >
                <Send className="h-4 w-4" /> Kirim
              </Button>
            )}
            {transfer.status === 'DISPATCHED' && (
              <Button
                onClick={() => receiveTransfer.mutate(transfer.id, { onSuccess: () => toast.success('Transfer diterima') })}
                loading={receiveTransfer.isPending}
              >
                <PackageCheck className="h-4 w-4" /> Terima
              </Button>
            )}
            {transfer.status === 'DRAFT' && (
              <Button
                variant="destructive"
                onClick={() => cancelTransfer.mutate(transfer.id, { onSuccess: () => toast.success('Transfer dibatalkan') })}
                loading={cancelTransfer.isPending}
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
              <StatusBadge kind="transfer" value={transfer.status} />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4">
            <p className="text-xs uppercase text-slate-400">Dikirim</p>
            <p className="mt-1.5 text-sm text-slate-700">{transfer.dispatched_at ? formatDateTime(transfer.dispatched_at) : '-'}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4">
            <p className="text-xs uppercase text-slate-400">Diterima</p>
            <p className="mt-1.5 text-sm text-slate-700">{transfer.received_at ? formatDateTime(transfer.received_at) : '-'}</p>
          </CardContent>
        </Card>
      </div>

      {transfer.status === 'DISPATCHED' && (
        <div className="mb-6 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          Stok sedang dalam perjalanan: sudah dikurangi dari {warehouseName(transfer.source_warehouse_id)}, belum ditambahkan ke{' '}
          {warehouseName(transfer.destination_warehouse_id)}.
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Item Baris</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead className="bg-slate-50 text-left text-xs font-medium uppercase text-slate-500">
              <tr>
                <th className="px-4 py-2.5">Item</th>
                <th className="px-4 py-2.5">Jenis</th>
                <th className="px-4 py-2.5">Jml</th>
                <th className="px-4 py-2.5">Biaya Satuan</th>
                <th className="px-4 py-2.5 text-right">Total Baris</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {transfer.items?.map((it) => (
                <tr key={it.id}>
                  <td className="px-4 py-2.5">{itemName(it)}</td>
                  <td className="px-4 py-2.5">{it.item_type === 'MATERIAL' ? 'Bahan Baku' : 'Produk'}</td>
                  <td className="px-4 py-2.5">{it.qty}</td>
                  <td className="px-4 py-2.5">{Number(it.unit_cost) > 0 ? formatCurrency(it.unit_cost) : '-'}</td>
                  <td className="px-4 py-2.5 text-right font-medium">{Number(it.unit_cost) > 0 ? formatCurrency(Number(it.qty) * Number(it.unit_cost)) : '-'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  )
}
