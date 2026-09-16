import { useState } from 'react'
import { PageHeader } from '@/components/PageHeader'
import { Card, CardContent } from '@/components/ui/Card'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { DateRangeFilter } from '@/components/DateRangeFilter'
import { useProfitLoss } from '@/hooks/useReporting'
import { formatCurrency, toISODate } from '@/lib/format'

function defaultRange() {
  const now = new Date()
  return { from: toISODate(new Date(now.getFullYear(), now.getMonth(), 1)), to: toISODate(now) }
}

function Row({ label, value, bold, indent }: { label: string; value: string; bold?: boolean; indent?: boolean }) {
  return (
    <div className={`flex justify-between py-1.5 text-sm ${bold ? 'font-semibold' : ''} ${indent ? 'pl-4 text-slate-500' : ''}`}>
      <span>{label}</span>
      <span>{value}</span>
    </div>
  )
}

export function ProfitLossPage() {
  const initial = defaultRange()
  const [from, setFrom] = useState(initial.from)
  const [to, setTo] = useState(initial.to)
  const { data, isLoading, error } = useProfitLoss(from, to)

  return (
    <div>
      <PageHeader
        title="Laba Rugi"
        description="Pendapatan, HPP, dan beban untuk periode yang dipilih."
        actions={<DateRangeFilter from={from} to={to} onFromChange={setFrom} onToChange={setTo} />}
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <Card className="max-w-2xl">
          <CardContent className="divide-y divide-slate-100 py-4">
            <div className="pb-2">
              <p className="mb-1 text-xs font-semibold uppercase text-slate-400">Pendapatan</p>
              {data.revenue.map((r) => (
                <Row key={r.account_code} label={`${r.account_code} ${r.account_name}`} value={formatCurrency(r.amount)} indent />
              ))}
              <Row label="Total Pendapatan" value={formatCurrency(data.total_revenue)} bold />
            </div>
            <div className="py-2">
              <p className="mb-1 text-xs font-semibold uppercase text-slate-400">Harga Pokok Penjualan</p>
              {data.cogs.map((r) => (
                <Row key={r.account_code} label={`${r.account_code} ${r.account_name}`} value={formatCurrency(r.amount)} indent />
              ))}
              <Row label="Total HPP" value={formatCurrency(data.total_cogs)} bold />
            </div>
            <Row label="Laba Kotor" value={formatCurrency(data.gross_profit)} bold />
            <div className="py-2">
              <p className="mb-1 text-xs font-semibold uppercase text-slate-400">Beban Operasional</p>
              {data.expenses.map((r) => (
                <Row key={r.account_code} label={`${r.account_code} ${r.account_name}`} value={formatCurrency(r.amount)} indent />
              ))}
              <Row label="Total Beban" value={formatCurrency(data.total_expenses)} bold />
            </div>
            <div className="pt-2">
              <Row
                label="Laba Bersih"
                value={formatCurrency(data.net_profit)}
                bold
              />
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
