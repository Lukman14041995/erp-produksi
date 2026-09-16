import { useState } from 'react'
import { PageHeader } from '@/components/PageHeader'
import { LoadingState, ErrorState } from '@/components/QueryState'
import { AsOfFilter } from '@/components/DateRangeFilter'
import { AgingTable } from '@/components/AgingTable'
import { useAPAging } from '@/hooks/useReporting'
import { toISODate } from '@/lib/format'

export function APAgingPage() {
  const [asOf, setAsOf] = useState(toISODate(new Date()))
  const { data, isLoading, error } = useAPAging(asOf)

  return (
    <div>
      <PageHeader
        title="Umur Utang"
        description="Tagihan pemasok yang belum lunas, dikelompokkan berdasarkan hari keterlambatan."
        actions={<AsOfFilter value={asOf} onChange={setAsOf} />}
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={(error as Error).message} />}
      {data && <AgingTable rows={data.rows} nameKey={(r) => r.supplier_name} total={data} />}
    </div>
  )
}
