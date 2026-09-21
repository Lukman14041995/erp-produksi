import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { StatusBadge } from '@/components/StatusBadge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs'
import { useQuotations } from '@/hooks/useQuotations'
import { useCustomers } from '@/hooks/useCustomers'
import { formatCurrency, formatDateTime } from '@/lib/format'
import type { Quotation, QuotationStatus } from '@/types/quotation'

const tabs: { value: QuotationStatus | 'ALL'; label: string }[] = [
  { value: 'PENDING_REVIEW', label: 'Menunggu Review' },
  { value: 'CONFIRMED', label: 'Dikonfirmasi' },
  { value: 'REJECTED', label: 'Ditolak' },
  { value: 'ALL', label: 'Semua' },
]

export function QuotationsPage() {
  const navigate = useNavigate()
  const [status, setStatus] = useState<QuotationStatus | 'ALL'>('PENDING_REVIEW')
  const { data, isLoading, error } = useQuotations(status === 'ALL' ? undefined : status)
  const { data: customers } = useCustomers()

  const customerName = (id: string) => customers?.find((c) => c.id === id)?.name ?? id.slice(0, 8)

  const columns: Column<Quotation>[] = [
    { key: 'quotation_number', header: 'No. Quotation', render: (q) => q.quotation_number, csvValue: (q) => q.quotation_number },
    { key: 'customer', header: 'Pelanggan', render: (q) => customerName(q.customer_id), csvValue: (q) => customerName(q.customer_id) },
    {
      key: 'estimated_subtotal',
      header: 'Estimasi Subtotal',
      render: (q) => formatCurrency(q.estimated_subtotal),
      sortValue: (q) => Number(q.estimated_subtotal),
      csvValue: (q) => q.estimated_subtotal,
    },
    { key: 'created_at', header: 'Diterima', render: (q) => formatDateTime(q.created_at), sortValue: (q) => q.created_at, csvValue: (q) => q.created_at },
    {
      key: 'status',
      header: 'Status',
      render: (q) => <StatusBadge kind="quotation" value={q.status} />,
      csvValue: (q) => q.status,
    },
  ]

  return (
    <div>
      <PageHeader title="Quotation Masuk" description="Pesanan konfigurasi yang dikirim customer lewat link, menunggu direview dan dikonfirmasi jadi Pesanan Penjualan." />

      <Tabs value={status} onValueChange={(v) => setStatus(v as QuotationStatus | 'ALL')}>
        <TabsList>
          {tabs.map((t) => (
            <TabsTrigger key={t.value} value={t.value}>
              {t.label}
            </TabsTrigger>
          ))}
        </TabsList>
        <TabsContent value={status}>
          {isLoading && <LoadingState />}
          {error && <ErrorState message={(error as Error).message} />}
          {data && (
            <DataTable
              data={data}
              columns={columns}
              rowKey={(q) => q.id}
              exportFilename="quotations"
              searchPlaceholder="Cari nomor quotation..."
              searchFn={(q, query) => q.quotation_number.toLowerCase().includes(query)}
              onRowClick={(q) => navigate(`/sales/quotations/${q.id}`)}
              emptyMessage="Tidak ada quotation."
            />
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}
