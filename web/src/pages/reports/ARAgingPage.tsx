import { useState } from 'react'
import { PageHeader } from '@/components/PageHeader'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { AsOfFilter } from '@/components/DateRangeFilter'
import { AgingTable } from '@/components/AgingTable'
import { useARAging } from '@/hooks/useReporting'
import { toISODate } from '@/lib/format'

export function ARAgingPage() {
  const [asOf, setAsOf] = useState(toISODate(new Date()))
  const { data, isLoading, error } = useARAging(asOf)

  return (
    <div>
      <PageHeader
        title="Umur Piutang"
        description="Faktur pelanggan yang belum lunas, dikelompokkan berdasarkan hari keterlambatan."
        actions={<AsOfFilter value={asOf} onChange={setAsOf} />}
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && <AgingTable rows={data.rows} nameKey={(r) => r.customer_name} total={data} />}
    </div>
  )
}
