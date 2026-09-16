import { PrintLetterhead, PrintNotes, PrintSectionTitle, PrintSignatures, PrintTable, PrintTwoColumn } from './PrintLayout'
import { formatDate, formatNumber } from '@/lib/format'
import type { ProductionOrder } from '@/types/production'
import type { Material, Product } from '@/types/master'

const PROCESS_STAGES = ['Cutting', 'Printing', 'Sewing', 'QC']

export function PrintableWorkOrder({
  order,
  product,
  materials,
  soNumber,
  bomName,
}: {
  order: ProductionOrder
  product?: Product
  materials?: Material[]
  soNumber?: string
  bomName?: string
}) {
  const sizeCode = (id: string) => product?.sizes?.find((s) => s.id === id)?.size_code ?? '-'
  const items = order.items ?? []
  const material = (id: string) => materials?.find((m) => m.id === id)

  const meta: [string, string][] = [
    ['Work Order No.', order.prod_number],
    ['Status', order.production_status],
    ['Start Date', order.start_date ? formatDate(order.start_date) : '-'],
    ['End Date', order.end_date ? formatDate(order.end_date) : '-'],
  ]
  if (soNumber) meta.push(['Sales Order', soNumber])

  return (
    <div>
      <PrintLetterhead docTitle="SLIP KERJA PRODUKSI" meta={meta} />
      <div className="mb-4 flex items-start justify-between gap-6">
        <PrintTwoColumn
          leftTitle="Product Specification"
          leftLines={[product?.name ?? '-', 'Code: ' + (product?.code || '-'), 'BOM: ' + (bomName || '-')]}
          rightTitle="Quantities"
          rightLines={[`Planned: ${formatNumber(order.planned_qty, 2)} pcs`, `Finished: ${formatNumber(order.finished_qty, 2)} pcs`]}
        />
        <div className="w-24 shrink-0 rounded border border-slate-300 p-2 text-center">
          <p className="text-[10px] font-semibold text-slate-500">QR CODE</p>
          <p className="text-[9px] text-slate-400">(placeholder)</p>
          <p className="mt-1 break-all text-[9px] font-semibold text-slate-800">{order.prod_number}</p>
        </div>
      </div>

      {items.length > 0 && (
        <>
          <PrintSectionTitle>Size Matrix</PrintSectionTitle>
          <PrintTable
            headers={['Metric', ...items.map((it) => sizeCode(it.product_size_id))]}
            aligns={['left', ...items.map(() => 'center' as const)]}
            rows={[
              ['Planned Qty', ...items.map((it) => formatNumber(it.planned_qty, 2))],
              ['Finished Qty', ...items.map((it) => formatNumber(it.finished_qty, 2))],
            ]}
          />
        </>
      )}

      <PrintSectionTitle>Material Requirements</PrintSectionTitle>
      <PrintTable
        headers={['Code', 'Material', 'UOM', 'Planned Qty', 'Issued Qty']}
        aligns={['left', 'left', 'center', 'right', 'right']}
        rows={
          (order.materials ?? []).length > 0
            ? (order.materials ?? []).map((m) => {
                const mat = material(m.material_id)
                return [mat?.code ?? '-', mat?.name ?? m.material_id.slice(0, 8), mat?.uom ?? '-', formatNumber(m.planned_qty, 2), formatNumber(m.issued_qty, 2)]
              })
            : [['-', 'No BOM materials recorded', '', '', '']]
        }
      />

      <PrintSectionTitle>Process Stages</PrintSectionTitle>
      <div className="mb-4 flex gap-8">
        {PROCESS_STAGES.map((stage) => (
          <label key={stage} className="flex items-center gap-2 text-xs text-slate-800">
            <span className="inline-block h-3.5 w-3.5 border border-slate-400" />
            {stage}
          </label>
        ))}
      </div>

      <PrintNotes notes={order.notes} />
      <PrintSignatures labels={['Cutting Supervisor', 'Production Supervisor', 'QC Inspector']} />
    </div>
  )
}
