import { Link, useNavigate } from 'react-router-dom'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { StatusBadge } from '@/components/StatusBadge'
import { useInvoices, useSalesOrders } from '@/hooks/useSales'
import { formatCurrency, formatDate } from '@/lib/format'
import type { Invoice } from '@/types/sales'

export function InvoicesPage() {
  const { data, isLoading, error } = useInvoices()
  const { data: orders } = useSalesOrders()
  const navigate = useNavigate()

  const soNumber = (id: string) => orders?.find((o) => o.id === id)?.so_number ?? id.slice(0, 8)

  const columns: Column<Invoice>[] = [
    { key: 'inv', header: 'No. Faktur', render: (i) => i.invoice_number, sortValue: (i) => i.invoice_number, csvValue: (i) => i.invoice_number },
    {
      key: 'so',
      header: 'Pesanan Penjualan',
      render: (i) => (
        <Link to={`/sales/orders/${i.sales_order_id}`} className="text-[var(--color-primary)] hover:underline" onClick={(e) => e.stopPropagation()}>
          {soNumber(i.sales_order_id)}
        </Link>
      ),
      csvValue: (i) => soNumber(i.sales_order_id),
    },
    { key: 'date', header: 'Tanggal Faktur', render: (i) => formatDate(i.invoice_date), sortValue: (i) => i.invoice_date, csvValue: (i) => i.invoice_date },
    { key: 'total', header: 'Total Keseluruhan', render: (i) => formatCurrency(i.grand_total), sortValue: (i) => Number(i.grand_total), csvValue: (i) => i.grand_total },
    { key: 'paid', header: 'Dibayar', render: (i) => formatCurrency(i.paid_amount), sortValue: (i) => Number(i.paid_amount), csvValue: (i) => i.paid_amount },
    { key: 'balance', header: 'Sisa Tagihan', render: (i) => formatCurrency(i.balance_due), sortValue: (i) => Number(i.balance_due), csvValue: (i) => i.balance_due },
    { key: 'status', header: 'Status', render: (i) => <StatusBadge kind="invoice" value={i.status} />, csvValue: (i) => i.status },
  ]

  return (
    <div>
      <PageHeader title="Faktur" description="Faktur piutang yang dibuat dari pesanan penjualan yang telah dikonfirmasi." />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(i) => i.id}
          exportFilename="invoices"
          searchPlaceholder="Cari nomor faktur..."
          searchFn={(i, q) => i.invoice_number.toLowerCase().includes(q)}
          onRowClick={(i) => navigate(`/sales/orders/${i.sales_order_id}`)}
        />
      )}
    </div>
  )
}
