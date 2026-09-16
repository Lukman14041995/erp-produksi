import { Link } from 'react-router-dom'
import { Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { PageHeader } from '@/components/PageHeader'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { LoadingState } from '@/components/QueryState'
import { useProfitLoss } from '@/hooks/useReporting'
import { useCashFlow } from '@/hooks/useReporting'
import { useOrderProfitability } from '@/hooks/useReporting'
import { formatCurrency, formatNumber, toISODate } from '@/lib/format'
import { DataTable, type Column } from '@/components/DataTable'
import type { OrderProfitability } from '@/types/reporting'

function monthRange() {
  const now = new Date()
  const from = new Date(now.getFullYear(), now.getMonth(), 1)
  return { from: toISODate(from), to: toISODate(now) }
}

function MetricCard({ label, value, tone }: { label: string; value: string; tone?: 'success' | 'danger' }) {
  return (
    <Card>
      <CardContent className="py-5">
        <p className="text-xs font-medium uppercase tracking-wide text-slate-400">{label}</p>
        <p
          className={
            'mt-1.5 text-2xl font-semibold ' +
            (tone === 'success' ? 'text-[var(--color-success)]' : tone === 'danger' ? 'text-[var(--color-danger)]' : 'text-slate-900')
          }
        >
          {value}
        </p>
      </CardContent>
    </Card>
  )
}

export function DashboardPage() {
  const { from, to } = monthRange()
  const pl = useProfitLoss(from, to)
  const cf = useCashFlow(from, to)
  const profitability = useOrderProfitability()

  const chartData = pl.data
    ? [
        { name: 'Pendapatan', value: Number(pl.data.total_revenue) },
        { name: 'HPP', value: Number(pl.data.total_cogs) },
        { name: 'Laba Kotor', value: Number(pl.data.gross_profit) },
        { name: 'Beban', value: Number(pl.data.total_expenses) },
        { name: 'Laba Bersih', value: Number(pl.data.net_profit) },
      ]
    : []

  const columns: Column<OrderProfitability>[] = [
    { key: 'so', header: 'No. SO', render: (r) => (
        <Link to={`/sales/orders/${r.sales_order_id}`} className="font-medium text-[var(--color-primary)] hover:underline">
          {r.so_number}
        </Link>
      ), csvValue: (r) => r.so_number },
    { key: 'customer', header: 'Pelanggan', render: (r) => r.customer_name, csvValue: (r) => r.customer_name },
    { key: 'revenue', header: 'Pendapatan', render: (r) => formatCurrency(r.revenue), sortValue: (r) => Number(r.revenue), csvValue: (r) => r.revenue },
    { key: 'cost', header: 'Total HPP', render: (r) => formatCurrency(r.total_cost), sortValue: (r) => Number(r.total_cost), csvValue: (r) => r.total_cost },
    {
      key: 'profit',
      header: 'Laba',
      render: (r) => (
        <span className={Number(r.profit) >= 0 ? 'text-[var(--color-success)]' : 'text-[var(--color-danger)]'}>
          {formatCurrency(r.profit)}
        </span>
      ),
      sortValue: (r) => Number(r.profit),
      csvValue: (r) => r.profit,
    },
    { key: 'margin', header: 'Margin %', render: (r) => `${formatNumber(r.margin_pct, 1)}%`, sortValue: (r) => Number(r.margin_pct), csvValue: (r) => r.margin_pct },
  ]

  return (
    <div>
      <PageHeader title="Dasbor" description={`Metrik utama untuk ${new Date(from).toLocaleString('id-ID', { month: 'long', year: 'numeric' })} hingga hari ini`} />

      {pl.isLoading || cf.isLoading ? (
        <LoadingState />
      ) : (
        <>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
            <MetricCard label="Pendapatan" value={formatCurrency(pl.data?.total_revenue ?? '0')} />
            <MetricCard label="HPP" value={formatCurrency(pl.data?.total_cogs ?? '0')} />
            <MetricCard label="Margin Kotor" value={formatCurrency(pl.data?.gross_profit ?? '0')} />
            <MetricCard
              label="Laba Bersih"
              value={formatCurrency(pl.data?.net_profit ?? '0')}
              tone={Number(pl.data?.net_profit ?? 0) >= 0 ? 'success' : 'danger'}
            />
            <MetricCard label="Saldo Kas" value={formatCurrency(cf.data?.ending_cash_balance ?? '0')} />
          </div>

          <Card className="mt-6">
            <CardHeader>
              <CardTitle>Ringkasan Laba Rugi (Bulan Berjalan)</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={chartData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#2a2a2c" />
                    <XAxis dataKey="name" tick={{ fontSize: 12, fill: '#a1a1a6' }} />
                    <YAxis tickFormatter={(v) => formatCurrency(v)} width={90} tick={{ fontSize: 11, fill: '#a1a1a6' }} />
                    <Tooltip
                      formatter={(v) => formatCurrency(Number(v))}
                      contentStyle={{ backgroundColor: '#161617', border: '1px solid #2a2a2c', borderRadius: 8, color: '#f5f5f5' }}
                      labelStyle={{ color: '#f5f5f5' }}
                      itemStyle={{ color: '#f5f5f5' }}
                    />
                    <Bar dataKey="value" fill="var(--color-primary)" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>
        </>
      )}

      <Card className="mt-6">
        <CardHeader>
          <CardTitle>Profitabilitas Pesanan (Pendapatan - HPP Langsung = Laba)</CardTitle>
        </CardHeader>
        <CardContent>
          {profitability.isLoading ? (
            <LoadingState />
          ) : (
            <DataTable
              data={profitability.data ?? []}
              columns={columns}
              rowKey={(r) => r.sales_order_id}
              exportFilename="order-profitability"
              searchPlaceholder="Cari pesanan..."
              searchFn={(r, q) => r.so_number.toLowerCase().includes(q) || r.customer_name.toLowerCase().includes(q)}
              pageSize={8}
            />
          )}
        </CardContent>
      </Card>
    </div>
  )
}
