import { Link, useNavigate } from 'react-router-dom'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { Badge, type BadgeTone } from '@/components/ui/Badge'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { useSpkOrders } from '@/hooks/useSpk'
import { useAuthStore } from '@/store/authStore'
import { formatDate, formatNumber } from '@/lib/format'
import type { SpkOrder, SpkStatus } from '@/types/spk'

const statusTone: Record<SpkStatus, BadgeTone> = {
  NOT_STARTED: 'neutral',
  IN_PROGRESS: 'info',
  COMPLETED: 'success',
}

const statusLabel: Record<SpkStatus, string> = {
  NOT_STARTED: 'Belum Dimulai',
  IN_PROGRESS: 'Sedang Berjalan',
  COMPLETED: 'Selesai',
}

function ProgressCell({ order }: { order: SpkOrder }) {
  const pct = Math.min(100, Math.round(Number(order.progress_pct)))
  return (
    <div className="w-40">
      <div className="flex items-center justify-between text-xs text-slate-500">
        <span>{order.current_stage_name}</span>
        <span>{pct}%</span>
      </div>
      <div className="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-slate-100">
        <div className="h-full rounded-full bg-[var(--color-primary)]" style={{ width: `${pct}%` }} />
      </div>
    </div>
  )
}

export function SpkListPage({ productTypeCode, title }: { productTypeCode: string; title: string }) {
  const { data, isLoading, error } = useSpkOrders(productTypeCode)
  const navigate = useNavigate()
  const canOpenQuotation = useAuthStore((s) => s.user?.role === 'ADMIN' || s.user?.role === 'SALES')

  const columns: Column<SpkOrder>[] = [
    { key: 'spk_number', header: 'No. SPK', render: (o) => <span className="font-medium text-slate-900">{o.spk_number}</span>, sortValue: (o) => o.spk_number, csvValue: (o) => o.spk_number },
    {
      key: 'quotation_number',
      header: 'No. Quotation',
      render: (o) =>
        o.quotation_number && canOpenQuotation ? (
          <Link
            to={`/sales/quotations/${o.quotation_id}`}
            onClick={(e) => e.stopPropagation()}
            className="text-[var(--color-primary)] hover:underline"
          >
            {o.quotation_number}
          </Link>
        ) : (
          (o.quotation_number ?? '-')
        ),
      sortValue: (o) => o.quotation_number,
      csvValue: (o) => o.quotation_number,
    },
    { key: 'so_number', header: 'Pesanan', render: (o) => o.so_number, sortValue: (o) => o.so_number, csvValue: (o) => o.so_number },
    { key: 'customer_name', header: 'Customer', render: (o) => o.customer_name, sortValue: (o) => o.customer_name, csvValue: (o) => o.customer_name },
    {
      key: 'composition_summary',
      header: 'Bahan / Model / Tinta',
      render: (o) => <span className="text-slate-600">{o.composition_summary || '-'}</span>,
      csvValue: (o) => o.composition_summary,
    },
    { key: 'total_qty', header: 'Total Qty', render: (o) => formatNumber(o.total_qty), sortValue: (o) => Number(o.total_qty), csvValue: (o) => o.total_qty },
    { key: 'progress', header: 'Tahap Saat Ini', render: (o) => <ProgressCell order={o} /> },
    {
      key: 'status',
      header: 'Status',
      render: (o) => <Badge tone={statusTone[o.status]}>{statusLabel[o.status]}</Badge>,
      csvValue: (o) => statusLabel[o.status],
    },
    { key: 'created_at', header: 'Dibuat', render: (o) => formatDate(o.created_at), sortValue: (o) => o.created_at, csvValue: (o) => o.created_at },
  ]

  return (
    <div>
      <PageHeader
        title={title}
        description="Daftar SPK yang siap/sedang diproduksi, dibuat otomatis begitu pesanan customer dikonfirmasi."
      />

      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}

      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(o) => o.id}
          searchPlaceholder="Cari no. SPK, quotation, pesanan, atau customer..."
          searchFn={(o, q) =>
            o.spk_number.toLowerCase().includes(q) ||
            o.quotation_number.toLowerCase().includes(q) ||
            o.so_number.toLowerCase().includes(q) ||
            o.customer_name.toLowerCase().includes(q) ||
            o.composition_summary.toLowerCase().includes(q)
          }
          exportFilename={`spk-${productTypeCode.toLowerCase()}`}
          onRowClick={(o) => navigate(`/production/orders/${o.id}`)}
          emptyMessage="Belum ada SPK untuk jenis ini."
        />
      )}
    </div>
  )
}
