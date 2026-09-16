import { useState } from 'react'
import { PageHeader } from '@/components/PageHeader'
import { Card, CardContent } from '@/components/ui/Card'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { DateRangeFilter } from '@/components/DateRangeFilter'
import { useCashFlow } from '@/hooks/useReporting'
import { formatCurrency, toISODate } from '@/lib/format'

export function CashFlowPage() {
  const now = new Date()
  const [from, setFrom] = useState(toISODate(new Date(now.getFullYear(), now.getMonth(), 1)))
  const [to, setTo] = useState(toISODate(now))
  const { data, isLoading, error } = useCashFlow(from, to)

  return (
    <div>
      <PageHeader
        title="Arus Kas"
        description="Pergerakan kas metode langsung pada akun Kas/Bank."
        actions={<DateRangeFilter from={from} to={to} onFromChange={setFrom} onToChange={setTo} />}
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <Card className="max-w-xl">
          <CardContent className="space-y-1.5 divide-y divide-slate-100 py-4 text-sm">
            <div className="flex justify-between pb-1.5"><span className="text-slate-500">Saldo Kas Awal</span><span>{formatCurrency(data.beginning_cash_balance)}</span></div>
            <div className="flex justify-between py-1.5"><span className="text-slate-500">Kas dari Pelanggan</span><span className="text-[var(--color-success)]">+{formatCurrency(data.cash_from_customers)}</span></div>
            <div className="flex justify-between py-1.5"><span className="text-slate-500">Kas untuk Pemasok</span><span className="text-[var(--color-danger)]">-{formatCurrency(data.cash_for_suppliers)}</span></div>
            <div className="flex justify-between py-1.5"><span className="text-slate-500">Kas untuk Beban</span><span className="text-[var(--color-danger)]">-{formatCurrency(data.cash_for_expenses)}</span></div>
            <div className="flex justify-between py-1.5 font-semibold"><span>Perubahan Bersih</span><span>{formatCurrency(data.net_change)}</span></div>
            <div className="flex justify-between pt-1.5 text-base font-semibold"><span>Saldo Kas Akhir</span><span>{formatCurrency(data.ending_cash_balance)}</span></div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
