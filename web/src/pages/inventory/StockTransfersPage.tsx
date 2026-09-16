import { useNavigate } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Button } from '@/components/ui/Button'
import { StatusBadge } from '@/components/StatusBadge'
import { useStockTransfers, useWarehouses } from '@/hooks/useWarehouse'
import { formatDate } from '@/lib/format'
import type { StockTransfer } from '@/types/warehouse'

export function StockTransfersPage() {
  const { data, isLoading, error } = useStockTransfers()
  const { data: warehouses } = useWarehouses()
  const navigate = useNavigate()

  const warehouseName = (id: string) => warehouses?.find((w) => w.id === id)?.name ?? id.slice(0, 8)

  const columns: Column<StockTransfer>[] = [
    { key: 'number', header: 'Nomor Transfer', render: (t) => t.transfer_number, sortValue: (t) => t.transfer_number, csvValue: (t) => t.transfer_number },
    { key: 'source', header: 'Asal', render: (t) => warehouseName(t.source_warehouse_id), csvValue: (t) => warehouseName(t.source_warehouse_id) },
    { key: 'dest', header: 'Tujuan', render: (t) => warehouseName(t.destination_warehouse_id), csvValue: (t) => warehouseName(t.destination_warehouse_id) },
    { key: 'date', header: 'Tanggal Transfer', render: (t) => formatDate(t.transfer_date), sortValue: (t) => t.transfer_date, csvValue: (t) => t.transfer_date },
    { key: 'status', header: 'Status', render: (t) => <StatusBadge kind="transfer" value={t.status} />, csvValue: (t) => t.status },
  ]

  return (
    <div>
      <PageHeader
        title="Transfer Stok"
        description="Mutasi antar gudang: Draf -> Dikirim -> Diterima."
        actions={
          <Button onClick={() => navigate('/inventory/transfers/new')}>
            <Plus className="h-4 w-4" /> Transfer Baru
          </Button>
        }
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(t) => t.id}
          exportFilename="stock-transfers"
          searchPlaceholder="Cari nomor transfer..."
          searchFn={(t, q) => t.transfer_number.toLowerCase().includes(q)}
          onRowClick={(t) => navigate(`/inventory/transfers/${t.id}`)}
        />
      )}
    </div>
  )
}
