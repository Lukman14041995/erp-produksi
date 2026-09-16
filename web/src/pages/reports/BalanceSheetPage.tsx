import { useState } from 'react'
import { PageHeader } from '@/components/PageHeader'
import { Card, CardContent } from '@/components/ui/Card'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { AsOfFilter } from '@/components/DateRangeFilter'
import { useBalanceSheet } from '@/hooks/useReporting'
import { formatCurrency, toISODate } from '@/lib/format'

export function BalanceSheetPage() {
  const [asOf, setAsOf] = useState(toISODate(new Date()))
  const { data, isLoading, error } = useBalanceSheet(asOf)

  return (
    <div>
      <PageHeader
        title="Neraca"
        description="Aset = Liabilitas + Ekuitas, per tanggal yang dipilih."
        actions={<AsOfFilter value={asOf} onChange={setAsOf} />}
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <Card>
            <CardContent className="py-4">
              <p className="mb-2 text-xs font-semibold uppercase text-slate-400">Aset</p>
              {data.assets.map((a) => (
                <div key={a.account_code} className="flex justify-between py-1 text-sm">
                  <span className="text-slate-500">{a.account_code} {a.account_name}</span>
                  <span>{formatCurrency(a.balance)}</span>
                </div>
              ))}
              <div className="mt-2 flex justify-between border-t border-slate-100 pt-2 text-sm font-semibold">
                <span>Total Aset</span>
                <span>{formatCurrency(data.total_assets)}</span>
              </div>
            </CardContent>
          </Card>

          <div className="space-y-4">
            <Card>
              <CardContent className="py-4">
                <p className="mb-2 text-xs font-semibold uppercase text-slate-400">Liabilitas</p>
                {data.liabilities.map((l) => (
                  <div key={l.account_code} className="flex justify-between py-1 text-sm">
                    <span className="text-slate-500">{l.account_code} {l.account_name}</span>
                    <span>{formatCurrency(l.balance)}</span>
                  </div>
                ))}
                <div className="mt-2 flex justify-between border-t border-slate-100 pt-2 text-sm font-semibold">
                  <span>Total Liabilitas</span>
                  <span>{formatCurrency(data.total_liabilities)}</span>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="py-4">
                <p className="mb-2 text-xs font-semibold uppercase text-slate-400">Ekuitas</p>
                {data.equity.map((e) => (
                  <div key={e.account_code} className="flex justify-between py-1 text-sm">
                    <span className="text-slate-500">{e.account_code} {e.account_name}</span>
                    <span>{formatCurrency(e.balance)}</span>
                  </div>
                ))}
                <div className="flex justify-between py-1 text-sm text-slate-500">
                  <span>Laba Ditahan (berjalan, belum ditutup)</span>
                  <span>{formatCurrency(data.retained_earnings_current)}</span>
                </div>
                <div className="mt-2 flex justify-between border-t border-slate-100 pt-2 text-sm font-semibold">
                  <span>Total Ekuitas</span>
                  <span>{formatCurrency(data.total_equity)}</span>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      )}
    </div>
  )
}
