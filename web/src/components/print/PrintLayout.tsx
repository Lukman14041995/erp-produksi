import type { ReactNode } from 'react'
import { COMPANY } from './company'

export function PrintLetterhead({ docTitle, meta }: { docTitle: string; meta: [string, string][] }) {
  return (
    <div className="mb-4 border-b-2 border-[var(--color-primary)] pb-3">
      <div className="flex items-start justify-between gap-6">
        <div>
          <h1 className="text-lg font-bold text-slate-900">{COMPANY.name}</h1>
          <p className="mt-0.5 text-xs text-slate-500">{COMPANY.tagline}</p>
          <p className="text-xs text-slate-500">{COMPANY.address}</p>
          <p className="text-xs text-slate-500">
            {COMPANY.phone} &middot; {COMPANY.email}
          </p>
          <p className="text-xs text-slate-500">{COMPANY.taxId}</p>
        </div>
        <div className="shrink-0 text-right">
          <h2 className="text-lg font-bold text-[var(--color-primary)]">{docTitle}</h2>
          <table className="mt-1 text-xs">
            <tbody>
              {meta.map(([k, v]) => (
                <tr key={k}>
                  <td className="pr-2 text-right text-slate-500">{k}</td>
                  <td className="text-right font-semibold text-slate-900">{v}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

export function PrintTwoColumn({
  leftTitle,
  leftLines,
  rightTitle,
  rightLines,
}: {
  leftTitle: string
  leftLines: string[]
  rightTitle: string
  rightLines: string[]
}) {
  return (
    <div className="mb-4 grid grid-cols-2 gap-6 text-xs">
      <div>
        <p className="mb-1 font-semibold uppercase text-slate-400">{leftTitle}</p>
        {leftLines.map((l, i) => (
          <p key={i} className="text-slate-800">
            {l || '-'}
          </p>
        ))}
      </div>
      <div>
        <p className="mb-1 font-semibold uppercase text-slate-400">{rightTitle}</p>
        {rightLines.map((l, i) => (
          <p key={i} className="text-slate-800">
            {l || '-'}
          </p>
        ))}
      </div>
    </div>
  )
}

export function PrintSectionTitle({ children }: { children: ReactNode }) {
  return <p className="mb-1.5 text-sm font-semibold text-slate-900">{children}</p>
}

export function PrintTable({
  headers,
  aligns,
  rows,
}: {
  headers: string[]
  aligns: ('left' | 'center' | 'right')[]
  rows: ReactNode[][]
}) {
  return (
    <table className="mb-4 w-full border-collapse text-xs">
      <thead>
        <tr className="bg-slate-100">
          {headers.map((h, i) => (
            <th key={i} className="border border-slate-300 px-2 py-1.5 font-semibold text-slate-700" style={{ textAlign: aligns[i] }}>
              {h}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map((row, ri) => (
          <tr key={ri} className={ri % 2 === 1 ? 'bg-slate-50' : ''}>
            {row.map((cell, ci) => (
              <td key={ci} className="border border-slate-300 px-2 py-1.5" style={{ textAlign: aligns[ci] }}>
                {cell}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  )
}

export function PrintTotals({ rows, grandLabel, grandValue }: { rows: [string, string][]; grandLabel: string; grandValue: string }) {
  return (
    <div className="mb-4 flex justify-end">
      <table className="w-64 text-xs">
        <tbody>
          {rows.map(([label, value]) => (
            <tr key={label}>
              <td className="py-0.5 text-slate-500">{label}</td>
              <td className="py-0.5 text-right text-slate-900">{value}</td>
            </tr>
          ))}
          <tr className="border-t border-slate-300">
            <td className="pt-1.5 text-sm font-bold text-[var(--color-primary)]">{grandLabel}</td>
            <td className="pt-1.5 text-right text-sm font-bold text-[var(--color-primary)]">{grandValue}</td>
          </tr>
        </tbody>
      </table>
    </div>
  )
}

export function PrintNotes({ notes }: { notes: string }) {
  if (!notes) return null
  return (
    <div className="mb-4 text-xs">
      <p className="mb-1 font-semibold text-slate-500">Catatan</p>
      <p className="text-slate-800">{notes}</p>
    </div>
  )
}

export function PrintSignatures({ labels }: { labels: string[] }) {
  return (
    <div className="mt-10 grid gap-6 text-center text-xs" style={{ gridTemplateColumns: `repeat(${labels.length}, minmax(0, 1fr))` }}>
      {labels.map((label) => (
        <div key={label}>
          <p className="mb-10 font-semibold text-slate-800">{label}</p>
          <div className="border-t border-slate-400 pt-1 text-slate-400">Nama / Tanggal</div>
        </div>
      ))}
    </div>
  )
}
