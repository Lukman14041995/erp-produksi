import { useNavigate } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Button } from '@/components/ui/Button'
import { StatusBadge } from '@/components/StatusBadge'
import { useStockOpnames, useWarehouses } from '@/hooks/useWarehouse'
import { formatDate } from '@/lib/format'
import type { StockOpname } from '@/types/warehouse'

export function StockOpnamesPage() {
  const { data, isLoading, error } = useStockOpnames()
  const { data: warehouses } = useWarehouses()
  const navigate = useNavigate()

  const warehouseName = (id: string) => warehouses?.find((w) => w.id === id)?.name ?? id.slice(0, 8)

  const columns: Column<StockOpname>[] = [
    { key: 'number', header: 'Nomor Opname', render: (o) => o.opname_number, sortValue: (o) => o.opname_number, csvValue: (o) => o.opname_number },
    { key: 'warehouse', header: 'Gudang', render: (o) => warehouseName(o.warehouse_id), csvValue: (o) => warehouseName(o.warehouse_id) },
    { key: 'date', header: 'Tanggal Opname', render: (o) => formatDate(o.opname_date), sortValue: (o) => o.opname_date, csvValue: (o) => o.opname_date },
    { key: 'status', header: 'Status', render: (o) => <StatusBadge kind="opname" value={o.status} />, csvValue: (o) => o.status },
  ]

  return (
    <div>
      <PageHeader
        title="Stock Opname"
        description="Penghitungan fisik persediaan dan rekonsiliasi selisih otomatis."
        actions={
          <Button onClick={() => navigate('/inventory/opnames/new')}>
            <Plus className="h-4 w-4" /> Opname Baru
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
          exportFilename="stock-opnames"
          searchPlaceholder="Cari nomor opname..."
          searchFn={(o, q) => o.opname_number.toLowerCase().includes(q)}
          onRowClick={(o) => navigate(`/inventory/opnames/${o.id}`)}
        />
      )}
    </div>
  )
}
