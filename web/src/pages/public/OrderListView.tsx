import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'
import { Badge, type BadgeTone } from '@/components/ui/Badge'
import { LoadingState } from '@/components/QueryState'
import { useCustomerOrders } from '@/hooks/useQuotations'
import { formatCurrency, formatDateTime } from '@/lib/format'
import type { QuotationStatus } from '@/types/quotation'

const statusLabel: Record<QuotationStatus, { label: string; tone: BadgeTone }> = {
  PENDING_REVIEW: { label: 'Menunggu Review', tone: 'warning' },
  CONFIRMED: { label: 'Dikonfirmasi', tone: 'success' },
  REJECTED: { label: 'Ditolak', tone: 'danger' },
  CANCELLED: { label: 'Dibatalkan', tone: 'neutral' },
}

export function OrderListView({ token, onSelect, onNewOrder }: {
  token: string
  onSelect: (id: string) => void
  onNewOrder: () => void
}) {
  const { data, isLoading } = useCustomerOrders(token)

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-base font-semibold text-slate-900">Pesanan Saya</h2>
        <Button onClick={onNewOrder}>
          <Plus className="h-4 w-4" /> Pesanan Baru
        </Button>
      </div>

      {isLoading && <LoadingState />}
      {data && data.length === 0 && (
        <Card>
          <CardContent className="py-10 text-center text-sm text-slate-400">Belum ada pesanan. Klik "Pesanan Baru" untuk mulai memesan.</CardContent>
        </Card>
      )}
      <div className="space-y-2">
        {data?.map((q) => {
          const status = statusLabel[q.status]
          return (
            <Card key={q.id} className="cursor-pointer hover:border-slate-500" onClick={() => onSelect(q.id)}>
              <CardContent className="flex items-center justify-between py-4">
                <div>
                  <p className="font-medium text-slate-800">{q.quotation_number}</p>
                  <p className="text-xs text-slate-400">{formatDateTime(q.created_at)}</p>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-sm font-medium">{formatCurrency(q.estimated_subtotal)}</span>
                  <Badge tone={status.tone}>{status.label}</Badge>
                </div>
              </CardContent>
            </Card>
          )
        })}
      </div>
    </div>
  )
}
