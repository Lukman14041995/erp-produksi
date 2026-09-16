import { useNavigate } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { Button } from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { useProductionOrders } from '@/hooks/useProduction'
import { useProducts } from '@/hooks/useProducts'
import { formatDate } from '@/lib/format'
import type { ProdStatus, ProductionOrder } from '@/types/production'

const columns: { status: ProdStatus; label: string }[] = [
  { status: 'NOT_STARTED', label: 'Belum Dimulai' },
  { status: 'IN_PROGRESS', label: 'Sedang Berjalan' },
  { status: 'COMPLETED', label: 'Selesai' },
  { status: 'CANCELLED', label: 'Dibatalkan' },
]

function OrderCard({ order, productName, onClick }: { order: ProductionOrder; productName: string; onClick: () => void }) {
  const pct = Number(order.planned_qty) > 0 ? Math.round((Number(order.finished_qty) / Number(order.planned_qty)) * 100) : 0
  return (
    <Card className="cursor-pointer hover:border-slate-300" onClick={onClick}>
      <CardContent className="py-3">
        <p className="text-sm font-medium text-slate-900">{order.prod_number}</p>
        <p className="mt-0.5 text-xs text-slate-500">{productName}</p>
        <p className="mt-1 text-xs text-slate-400">{formatDate(order.created_at)}</p>
        <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-slate-100">
          <div className="h-full rounded-full bg-[var(--color-primary)]" style={{ width: `${pct}%` }} />
        </div>
        <p className="mt-1 text-right text-[11px] text-slate-400">
          {order.finished_qty} / {order.planned_qty} ({pct}%)
        </p>
      </CardContent>
    </Card>
  )
}

export function ProductionOrdersPage() {
  const { data, isLoading, error } = useProductionOrders()
  const { data: products } = useProducts()
  const navigate = useNavigate()

  const productName = (id: string) => products?.find((p) => p.id === id)?.name ?? id.slice(0, 8)

  return (
    <div>
      <PageHeader
        title="Pesanan Produksi"
        description="Lacak produksi dari perencanaan hingga pengeluaran bahan, tenaga kerja/overhead, dan penyelesaian."
        actions={
          <Button onClick={() => navigate('/production/orders/new')}>
            <Plus className="h-4 w-4" /> Pesanan Produksi Baru
          </Button>
        }
      />

      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}

      {data && (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-4">
          {columns.map((col) => (
            <div key={col.status}>
              <div className="mb-2 flex items-center justify-between px-1">
                <h3 className="text-xs font-semibold uppercase tracking-wide text-slate-500">{col.label}</h3>
                <span className="text-xs text-slate-400">{data.filter((o) => o.production_status === col.status).length}</span>
              </div>
              <div className="space-y-2">
                {data
                  .filter((o) => o.production_status === col.status)
                  .map((o) => (
                    <OrderCard key={o.id} order={o} productName={productName(o.product_id)} onClick={() => navigate(`/production/orders/${o.id}`)} />
                  ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
