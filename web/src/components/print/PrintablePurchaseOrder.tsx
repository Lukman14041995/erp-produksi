import { PrintLetterhead, PrintNotes, PrintSectionTitle, PrintSignatures, PrintTable, PrintTotals, PrintTwoColumn } from './PrintLayout'
import { COMPANY } from './company'
import { formatCurrency, formatDate, formatNumber } from '@/lib/format'
import type { PurchaseOrder } from '@/types/purchasing'
import type { Material, Supplier } from '@/types/master'

export function PrintablePurchaseOrder({ po, supplier, materials }: { po: PurchaseOrder; supplier?: Supplier; materials?: Material[] }) {
  const material = (id: string) => materials?.find((m) => m.id === id)

  return (
    <div>
      <PrintLetterhead
        docTitle="PESANAN PEMBELIAN"
        meta={[
          ['No. PO', po.po_number],
          ['Tanggal Pesanan', formatDate(po.order_date)],
          ['Tanggal Diharapkan', po.expected_date ? formatDate(po.expected_date) : '-'],
          ['Status', po.status],
        ]}
      />
      <PrintTwoColumn
        leftTitle="Pemasok"
        leftLines={[
          supplier?.name ?? '-',
          supplier?.address ?? '',
          'Kontak: ' + (supplier?.contact_person || '-'),
          'Telepon: ' + (supplier?.phone || '-'),
          'NPWP: ' + (supplier?.tax_id || '-'),
        ]}
        rightTitle="Dikirim Ke"
        rightLines={[COMPANY.name, COMPANY.address, COMPANY.phone]}
      />
      <PrintSectionTitle>Spesifikasi Item</PrintSectionTitle>
      <PrintTable
        headers={['Kode', 'Bahan Baku', 'Satuan', 'Jml', 'Biaya Satuan', 'Total Baris']}
        aligns={['left', 'left', 'center', 'right', 'right', 'right']}
        rows={(po.items ?? []).map((it) => {
          const m = material(it.material_id)
          return [m?.code ?? '-', m?.name ?? it.material_id.slice(0, 8), m?.uom ?? '-', formatNumber(it.qty, 2), formatCurrency(it.unit_cost), formatCurrency(it.line_total)]
        })}
      />
      <PrintTotals rows={[['Subtotal', formatCurrency(po.subtotal)]]} grandLabel="Total Keseluruhan" grandValue={formatCurrency(po.grand_total)} />
      <PrintNotes notes={po.notes} />
      <p className="mb-4 text-xs italic text-slate-500">
        Biaya satuan di atas telah disepakati dengan pemasok sebelum pengiriman dan dicocokkan dengan penerimaan barang serta tagihan pemasok.
      </p>
      <PrintSignatures labels={['Diminta Oleh', 'Disetujui Oleh', 'Persetujuan Pemasok']} />
    </div>
  )
}
