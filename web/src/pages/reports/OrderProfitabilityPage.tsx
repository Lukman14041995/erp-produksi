import { Link } from 'react-router-dom'
import { PageHeader } from '@/components/PageHeader'
import { DataTable, type Column } from '@/components/DataTable'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { useOrderProfitability } from '@/hooks/useReporting'
import { formatCurrency, formatNumber } from '@/lib/format'
import type { OrderProfitability } from '@/types/reporting'

export function OrderProfitabilityPage() {
  const { data, isLoading, error } = useOrderProfitability()

  const columns: Column<OrderProfitability>[] = [
    {
      key: 'so',
      header: 'No. SO',
      render: (r) => (
        <Link to={`/sales/orders/${r.sales_order_id}`} className="font-medium text-[var(--color-primary)] hover:underline">
          {r.so_number}
        </Link>
      ),
      csvValue: (r) => r.so_number,
    },
    { key: 'customer', header: 'Pelanggan', render: (r) => r.customer_name, sortValue: (r) => r.customer_name, csvValue: (r) => r.customer_name },
    { key: 'revenue', header: 'Pendapatan', render: (r) => formatCurrency(r.revenue), sortValue: (r) => Number(r.revenue), csvValue: (r) => r.revenue },
    { key: 'material', header: 'Bahan Langsung', render: (r) => formatCurrency(r.direct_material), csvValue: (r) => r.direct_material },
    { key: 'labor', header: 'Tenaga Kerja Langsung', render: (r) => formatCurrency(r.direct_labor), csvValue: (r) => r.direct_labor },
    { key: 'overhead', header: 'Overhead', render: (r) => formatCurrency(r.allocated_overhead), csvValue: (r) => r.allocated_overhead },
    { key: 'cost', header: 'Total HPP', render: (r) => formatCurrency(r.total_cost), sortValue: (r) => Number(r.total_cost), csvValue: (r) => r.total_cost },
    {
      key: 'profit',
      header: 'Laba',
      render: (r) => (
        <span className={Number(r.profit) >= 0 ? 'text-[var(--color-success)]' : 'text-[var(--color-danger)]'}>{formatCurrency(r.profit)}</span>
      ),
      sortValue: (r) => Number(r.profit),
      csvValue: (r) => r.profit,
    },
    { key: 'margin', header: 'Margin %', render: (r) => `${formatNumber(r.margin_pct, 1)}%`, sortValue: (r) => Number(r.margin_pct), csvValue: (r) => r.margin_pct },
  ]

  return (
    <div>
      <PageHeader title="Profitabilitas Pesanan" description="Pendapatan dikurangi HPP langsung (bahan + tenaga kerja + overhead) per pesanan penjualan." />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <DataTable
          data={data}
          columns={columns}
          rowKey={(r) => r.sales_order_id}
          exportFilename="order-profitability"
          searchPlaceholder="Cari pesanan..."
          searchFn={(r, q) => r.so_number.toLowerCase().includes(q) || r.customer_name.toLowerCase().includes(q)}
          pageSize={20}
        />
      )}
    </div>
  )
}
