import { useNavigate } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Button } from '@/components/ui/Button'
import { StatusBadge } from '@/components/StatusBadge'
import { useSalesOrders } from '@/hooks/useSales'
import { useCustomers } from '@/hooks/useCustomers'
import { formatCurrency, formatDate } from '@/lib/format'
import type { SalesOrder } from '@/types/sales'

export function SalesOrdersPage() {
  const { data, isLoading, error } = useSalesOrders()
  const { data: customers } = useCustomers()
  const navigate = useNavigate()

  const customerName = (id: string) => customers?.find((c) => c.id === id)?.name ?? id.slice(0, 8)

  const columns: Column<SalesOrder>[] = [
    { key: 'so', header: 'No. SO', render: (o) => o.so_number, sortValue: (o) => o.so_number, csvValue: (o) => o.so_number },
    { key: 'customer', header: 'Pelanggan', render: (o) => customerName(o.customer_id), csvValue: (o) => customerName(o.customer_id) },
    { key: 'date', header: 'Tanggal Pesanan', render: (o) => formatDate(o.order_date), sortValue: (o) => o.order_date, csvValue: (o) => o.order_date },
    { key: 'total', header: 'Total Keseluruhan', render: (o) => formatCurrency(o.grand_total), sortValue: (o) => Number(o.grand_total), csvValue: (o) => o.grand_total },
    { key: 'order_status', header: 'Pesanan', render: (o) => <StatusBadge kind="order" value={o.order_status} />, csvValue: (o) => o.order_status },
    { key: 'payment_status', header: 'Pembayaran', render: (o) => <StatusBadge kind="payment" value={o.payment_status} />, csvValue: (o) => o.payment_status },
    { key: 'production_status', header: 'Produksi', render: (o) => <StatusBadge kind="production" value={o.production_status} />, csvValue: (o) => o.production_status },
    { key: 'delivery_status', header: 'Pengiriman', render: (o) => <StatusBadge kind="delivery" value={o.delivery_status} />, csvValue: (o) => o.delivery_status },
  ]

  return (
    <div>
      <PageHeader
        title="Pesanan Penjualan"
        description="Siklus pesanan dari draf hingga konfirmasi, penagihan, produksi, dan pengiriman."
        actions={
          <Button onClick={() => navigate('/sales/orders/new')}>
            <Plus className="h-4 w-4" /> Pesanan Penjualan Baru
          </Button>
        }
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(o) => o.id}
          exportFilename="sales-orders"
          searchPlaceholder="Cari berdasarkan nomor SO..."
          searchFn={(o, q) => o.so_number.toLowerCase().includes(q)}
          onRowClick={(o) => navigate(`/sales/orders/${o.id}`)}
        />
      )}
    </div>
  )
}
