import { formatCurrency } from '@/lib/format'
import type { AgingBucket } from '@/types/reporting'

export function AgingTable<T extends AgingBucket>({
  rows,
  nameKey,
  total,
}: {
  rows: T[] | null | undefined
  nameKey: (row: T) => string
  total: AgingBucket
}) {
  const safeRows = rows ?? []
  return (
    <div className="overflow-x-auto rounded-lg border border-slate-200">
      <table className="w-full text-sm">
        <thead className="bg-slate-50 text-left text-xs font-medium uppercase text-slate-500">
          <tr>
            <th className="px-4 py-2.5">Nama</th>
            <th className="px-4 py-2.5 text-right">Saat Ini</th>
            <th className="px-4 py-2.5 text-right">1-30 Hari</th>
            <th className="px-4 py-2.5 text-right">31-60 Hari</th>
            <th className="px-4 py-2.5 text-right">61-90 Hari</th>
            <th className="px-4 py-2.5 text-right">&gt;90 Hari</th>
            <th className="px-4 py-2.5 text-right">Total</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-100">
          {safeRows.length === 0 && (
            <tr>
              <td colSpan={7} className="px-4 py-8 text-center text-slate-400">
                Tidak ada saldo terutang.
              </td>
            </tr>
          )}
          {safeRows.map((row, i) => (
            <tr key={i}>
              <td className="px-4 py-2.5 font-medium">{nameKey(row)}</td>
              <td className="px-4 py-2.5 text-right">{formatCurrency(row.current)}</td>
              <td className="px-4 py-2.5 text-right">{formatCurrency(row.days_1_30)}</td>
              <td className="px-4 py-2.5 text-right">{formatCurrency(row.days_31_60)}</td>
              <td className="px-4 py-2.5 text-right">{formatCurrency(row.days_61_90)}</td>
              <td className="px-4 py-2.5 text-right text-[var(--color-danger)]">{formatCurrency(row.over_90)}</td>
              <td className="px-4 py-2.5 text-right font-semibold">{formatCurrency(row.total)}</td>
            </tr>
          ))}
        </tbody>
        {safeRows.length > 0 && (
          <tfoot className="border-t-2 border-slate-200 font-semibold">
            <tr>
              <td className="px-4 py-2.5">Jumlah Total</td>
              <td className="px-4 py-2.5 text-right">{formatCurrency(total.current)}</td>
              <td className="px-4 py-2.5 text-right">{formatCurrency(total.days_1_30)}</td>
              <td className="px-4 py-2.5 text-right">{formatCurrency(total.days_31_60)}</td>
              <td className="px-4 py-2.5 text-right">{formatCurrency(total.days_61_90)}</td>
              <td className="px-4 py-2.5 text-right">{formatCurrency(total.over_90)}</td>
              <td className="px-4 py-2.5 text-right">{formatCurrency(total.total)}</td>
            </tr>
          </tfoot>
        )}
      </table>
    </div>
  )
}
