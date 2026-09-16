import { useNavigate } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Button } from '@/components/ui/Button'
import { StatusBadge } from '@/components/StatusBadge'
import { usePurchaseOrders } from '@/hooks/usePurchasing'
import { useSuppliers } from '@/hooks/useSuppliers'
import { formatCurrency, formatDate } from '@/lib/format'
import type { PurchaseOrder } from '@/types/purchasing'

export function PurchaseOrdersPage() {
  const { data, isLoading, error } = usePurchaseOrders()
  const { data: suppliers } = useSuppliers()
  const navigate = useNavigate()

  const supplierName = (id: string) => suppliers?.find((s) => s.id === id)?.name ?? id.slice(0, 8)

  const columns: Column<PurchaseOrder>[] = [
    { key: 'po', header: 'Nomor PO', render: (p) => p.po_number, sortValue: (p) => p.po_number, csvValue: (p) => p.po_number },
    { key: 'supplier', header: 'Pemasok', render: (p) => supplierName(p.supplier_id), csvValue: (p) => supplierName(p.supplier_id) },
    { key: 'date', header: 'Tanggal Pesan', render: (p) => formatDate(p.order_date), sortValue: (p) => p.order_date, csvValue: (p) => p.order_date },
    { key: 'total', header: 'Total Keseluruhan', render: (p) => formatCurrency(p.grand_total), sortValue: (p) => Number(p.grand_total), csvValue: (p) => p.grand_total },
    { key: 'status', header: 'Status', render: (p) => <StatusBadge kind="po" value={p.status} />, csvValue: (p) => p.status },
    { key: 'billing', header: 'Penagihan', render: (p) => <StatusBadge kind="poBilling" value={p.billing_status} />, csvValue: (p) => p.billing_status },
  ]

  return (
    <div>
      <PageHeader
        title="Pesanan Pembelian"
        description="Pencocokan 3 arah: Pesanan Pembelian -> Penerimaan Barang -> Tagihan Pemasok."
        actions={
          <Button onClick={() => navigate('/purchasing/orders/new')}>
            <Plus className="h-4 w-4" /> Pesanan Pembelian Baru
          </Button>
        }
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(p) => p.id}
          exportFilename="purchase-orders"
          searchPlaceholder="Cari nomor PO..."
          searchFn={(p, q) => p.po_number.toLowerCase().includes(q)}
          onRowClick={(p) => navigate(`/purchasing/orders/${p.id}`)}
        />
      )}
    </div>
  )
}
