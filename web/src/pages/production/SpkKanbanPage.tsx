import { useNavigate } from 'react-router-dom'
import { PageHeader } from '@/components/PageHeader'
import { Badge } from '@/components/ui/Badge'
import { Card, CardContent } from '@/components/ui/Card'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { useAllSpkOrders } from '@/hooks/useSpk'
import { formatCurrency, formatDate, formatNumber } from '@/lib/format'
import type { SpkOrder } from '@/types/spk'

type Bucket = 'design' | 'produksi' | 'selesai' | 'pengiriman'

const columns: { bucket: Bucket; label: string }[] = [
  { bucket: 'design', label: 'Design' },
  { bucket: 'produksi', label: 'Produksi' },
  { bucket: 'selesai', label: 'Selesai' },
  { bucket: 'pengiriman', label: 'Pengiriman' },
]

// Stages don't gate each other (an SPK can have several running at once),
// so this buckets on which of the three milestones -- Design, "Selesai",
// "Pengiriman" -- have actually been reached, not on whatever stage was
// most recently touched.
function bucketOf(o: SpkOrder): Bucket {
  if (!o.design_done) return 'design'
  if (!o.production_done) return 'produksi'
  if (!o.shipped) return 'selesai'
  return 'pengiriman'
}

function hppTotal(order: SpkOrder) {
  return Number(order.material_cost_total) + Number(order.labor_cost_total) + Number(order.overhead_cost_total)
}

function OrderCard({ order, onClick }: { order: SpkOrder; onClick: () => void }) {
  const pct = Math.min(100, Math.round(Number(order.progress_pct)))
  return (
    <Card className="cursor-pointer hover:border-slate-300" onClick={onClick}>
      <CardContent className="py-3">
        <div className="flex items-start justify-between gap-2">
          <p className="text-sm font-medium text-slate-900">{order.spk_number}</p>
          <Badge tone={order.product_type_code === 'JERSEY' ? 'info' : 'warning'}>{order.product_type_name}</Badge>
        </div>
        <p className="mt-0.5 text-xs text-slate-500">{order.customer_name}</p>
        {order.composition_summary && <p className="mt-0.5 text-xs text-slate-400">{order.composition_summary}</p>}
        <p className="mt-1 text-xs text-slate-400">
          {order.so_number} &middot; {formatDate(order.created_at)}
        </p>
        <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-slate-100">
          <div className="h-full rounded-full bg-[var(--color-primary)]" style={{ width: `${pct}%` }} />
        </div>
        <div className="mt-1 flex items-center justify-between text-[11px] text-slate-400">
          <span>HPP: {formatCurrency(hppTotal(order))}</span>
          <span>
            {order.current_stage_name} &middot; {formatNumber(order.total_qty)} pcs ({pct}%)
          </span>
        </div>
      </CardContent>
    </Card>
  )
}

export function SpkKanbanPage() {
  const { data, isLoading, error } = useAllSpkOrders()
  const navigate = useNavigate()

  return (
    <div>
      <PageHeader
        title="Papan Produksi"
        description="Ringkasan seluruh SPK (T-Shirt & Jersey) dikelompokkan menurut tahap besarnya. Untuk detail tahap kerja per jenis, buka Pesanan Produksi T-Shirt/Jersey."
      />

      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}

      {data && (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-4">
          {columns.map((col) => {
            const items = data.filter((o) => bucketOf(o) === col.bucket)
            return (
              <div key={col.bucket}>
                <div className="mb-2 flex items-center justify-between px-1">
                  <h3 className="text-xs font-semibold uppercase tracking-wide text-slate-500">{col.label}</h3>
                  <span className="text-xs text-slate-400">{items.length}</span>
                </div>
                <div className="space-y-2">
                  {items.map((o) => (
                    <OrderCard key={o.id} order={o} onClick={() => navigate(`/production/orders/${o.id}`)} />
                  ))}
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
