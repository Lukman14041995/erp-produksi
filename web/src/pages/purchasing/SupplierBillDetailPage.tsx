import { Link, useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, Send, XCircle } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Button } from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { StatusBadge } from '@/components/StatusBadge'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { usePostBill, usePurchaseOrder, useSupplierInvoice, useVoidBill } from '@/hooks/usePurchasing'
import { useSuppliers } from '@/hooks/useSuppliers'
import { useMaterials } from '@/hooks/useMaterials'
import { formatCurrency, formatDate } from '@/lib/format'
import { toast } from '@/store/toastStore'
import { ApiError } from '@/types/api'

export function SupplierBillDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: bill, isLoading, error } = useSupplierInvoice(id)
  const { data: suppliers } = useSuppliers()
  const { data: materials } = useMaterials()
  const { data: po } = usePurchaseOrder(bill?.purchase_order_id)
  const postBill = usePostBill()
  const voidBill = useVoidBill()

  if (isLoading) return <LoadingState />
  if (error || !bill) return <ErrorState message={(error as Error)?.message ?? 'Tagihan tidak ditemukan'} />

  const supplier = suppliers?.find((s) => s.id === bill.supplier_id)
  const materialName = (matId: string) => materials?.find((m) => m.id === matId)?.name ?? matId.slice(0, 8)
  const hasVariance = bill.items?.some((it) => Number(it.price_variance) !== 0)

  return (
    <div>
      <button onClick={() => navigate('/purchasing/bills')} className="mb-4 flex items-center gap-1 text-sm text-slate-500 hover:text-slate-800">
        <ArrowLeft className="h-4 w-4" /> Kembali ke Tagihan Pemasok
      </button>

      <PageHeader
        title={bill.bill_number}
        description={`${supplier?.name ?? bill.supplier_id} · ${formatDate(bill.bill_date)}`}
        actions={
          <div className="flex gap-2">
            {bill.status === 'DRAFT' && (
              <>
                <Button
                  onClick={() =>
                    postBill
                      .mutateAsync(id!)
                      .then(() => toast.success('Tagihan diposting, bahan baku diterima'))
                      .catch((err) => toast.error('Gagal posting tagihan', err instanceof ApiError ? err.message : undefined))
                  }
                  loading={postBill.isPending}
                >
                  <Send className="h-4 w-4" /> Posting Tagihan
                </Button>
                <Button
                  variant="destructive"
                  onClick={() =>
                    voidBill
                      .mutateAsync(id!)
                      .then(() => toast.success('Tagihan dibatalkan'))
                      .catch((err) => toast.error('Gagal membatalkan tagihan', err instanceof ApiError ? err.message : undefined))
                  }
                  loading={voidBill.isPending}
                >
                  <XCircle className="h-4 w-4" /> Batalkan
                </Button>
              </>
            )}
          </div>
        }
      />

      <div className="mb-6 flex items-center gap-2">
        <StatusBadge kind="bill" value={bill.status} />
        {po && (
          <Badge tone="info">
            Tercocok 3 arah &middot;{' '}
            <Link to={`/purchasing/orders/${po.id}`} className="underline">
              {po.po_number}
            </Link>
          </Badge>
        )}
      </div>

      <Card>
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead className="bg-slate-50 text-left text-xs font-medium uppercase text-slate-500">
              <tr>
                <th className="px-4 py-2.5">Bahan Baku</th>
                <th className="px-4 py-2.5">Jml</th>
                <th className="px-4 py-2.5">Biaya Satuan</th>
                {hasVariance && <th className="px-4 py-2.5">Selisih Harga</th>}
                <th className="px-4 py-2.5 text-right">Total Baris</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {bill.items?.map((it) => (
                <tr key={it.id}>
                  <td className="px-4 py-2.5">{materialName(it.material_id)}</td>
                  <td className="px-4 py-2.5">{it.qty}</td>
                  <td className="px-4 py-2.5">{formatCurrency(it.unit_cost)}</td>
                  {hasVariance && (
                    <td className={Number(it.price_variance) > 0 ? 'px-4 py-2.5 text-[var(--color-danger)]' : Number(it.price_variance) < 0 ? 'px-4 py-2.5 text-[var(--color-success)]' : 'px-4 py-2.5 text-slate-400'}>
                      {formatCurrency(it.price_variance)}
                    </td>
                  )}
                  <td className="px-4 py-2.5 text-right font-medium">{formatCurrency(it.line_total)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>

      <div className="mt-4 flex justify-end">
        <div className="w-64 space-y-1 text-sm">
          <div className="flex justify-between"><span className="text-slate-500">Subtotal</span><span>{formatCurrency(bill.subtotal)}</span></div>
          <div className="flex justify-between"><span className="text-slate-500">Pajak</span><span>{formatCurrency(bill.tax_total)}</span></div>
          <div className="flex justify-between border-t border-slate-100 pt-1 font-semibold"><span>Total Keseluruhan</span><span>{formatCurrency(bill.grand_total)}</span></div>
          <div className="flex justify-between text-[var(--color-success)]"><span>Dibayar</span><span>{formatCurrency(bill.paid_amount)}</span></div>
          <div className="flex justify-between font-semibold text-[var(--color-danger)]"><span>Sisa Tagihan</span><span>{formatCurrency(bill.balance_due)}</span></div>
        </div>
      </div>
    </div>
  )
}
