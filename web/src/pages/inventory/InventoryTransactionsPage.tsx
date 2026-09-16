import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { Badge, type BadgeTone } from '@/components/ui/Badge'
import { useInventoryTransactions } from '@/hooks/useInventory'
import { useMaterials } from '@/hooks/useMaterials'
import { useProducts } from '@/hooks/useProducts'
import { useWarehouses } from '@/hooks/useWarehouse'
import { formatCurrency, formatDateTime, formatNumber } from '@/lib/format'
import type { InventoryTransaction, TxnType } from '@/types/inventory'

const txnTone: Record<TxnType, BadgeTone> = {
  PURCHASE: 'info',
  PRODUCTION_ISSUE: 'warning',
  PRODUCTION_RECEIPT: 'success',
  SALE: 'danger',
  ADJUSTMENT: 'neutral',
  TRANSFER_OUT: 'warning',
  TRANSFER_IN: 'success',
}

const txnLabel: Record<TxnType, string> = {
  PURCHASE: 'Pembelian',
  PRODUCTION_ISSUE: 'Keluar Produksi',
  PRODUCTION_RECEIPT: 'Masuk Produksi',
  SALE: 'Penjualan',
  ADJUSTMENT: 'Penyesuaian',
  TRANSFER_OUT: 'Transfer Keluar',
  TRANSFER_IN: 'Transfer Masuk',
}

export function InventoryTransactionsPage() {
  const { data, isLoading, error } = useInventoryTransactions()
  const { data: materials } = useMaterials()
  const { data: products } = useProducts()
  const { data: warehouses } = useWarehouses()

  const itemName = (t: InventoryTransaction): string => {
    if (t.item_type === 'MATERIAL') return materials?.find((m) => m.id === t.material_id)?.name ?? t.material_id?.slice(0, 8) ?? ''
    return products?.find((p) => p.id === t.product_id)?.name ?? t.product_id?.slice(0, 8) ?? ''
  }
  const warehouseName = (t: InventoryTransaction): string => warehouses?.find((w) => w.id === t.warehouse_id)?.name ?? t.warehouse_id.slice(0, 8)

  const columns: Column<InventoryTransaction>[] = [
    { key: 'date', header: 'Tanggal', render: (t) => formatDateTime(t.created_at), sortValue: (t) => t.created_at, csvValue: (t) => t.created_at },
    { key: 'type', header: 'Jenis', render: (t) => <Badge tone={txnTone[t.txn_type]}>{txnLabel[t.txn_type]}</Badge>, csvValue: (t) => txnLabel[t.txn_type] },
    { key: 'warehouse', header: 'Gudang', render: warehouseName, csvValue: warehouseName },
    { key: 'item', header: 'Item', render: itemName, csvValue: itemName },
    { key: 'in', header: 'Jml Masuk', render: (t) => (Number(t.qty_in) > 0 ? formatNumber(t.qty_in, 2) : '-'), csvValue: (t) => t.qty_in },
    { key: 'out', header: 'Jml Keluar', render: (t) => (Number(t.qty_out) > 0 ? formatNumber(t.qty_out, 2) : '-'), csvValue: (t) => t.qty_out },
    { key: 'cost', header: 'Biaya Satuan', render: (t) => formatCurrency(t.unit_cost), csvValue: (t) => t.unit_cost },
    { key: 'total', header: 'Total Biaya', render: (t) => formatCurrency(t.total_cost), sortValue: (t) => Number(t.total_cost), csvValue: (t) => t.total_cost },
    { key: 'ref', header: 'Referensi', render: (t) => t.ref_type, csvValue: (t) => t.ref_type },
  ]

  return (
    <div>
      <PageHeader title="Transaksi Persediaan" description="Buku besar pergerakan lengkap untuk pembelian, produksi, dan penjualan." />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && <DataTable data={data} columns={columns} rowKey={(t) => t.id} exportFilename="inventory-transactions" pageSize={20} />}
    </div>
  )
}
