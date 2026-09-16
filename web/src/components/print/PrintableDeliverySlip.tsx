import { PrintLetterhead, PrintNotes, PrintSectionTitle, PrintSignatures, PrintTable, PrintTotals, PrintTwoColumn } from './PrintLayout'
import { COMPANY } from './company'
import { formatDate, formatNumber } from '@/lib/format'
import type { SalesOrder } from '@/types/sales'
import type { Customer, Product } from '@/types/master'

export function PrintableDeliverySlip({ order, customer, products }: { order: SalesOrder; customer?: Customer; products?: Product[] }) {
  const productName = (id: string) => products?.find((p) => p.id === id)?.name ?? id.slice(0, 8)
  const sizeCode = (pid: string, sid: string) => products?.find((p) => p.id === pid)?.sizes?.find((s) => s.id === sid)?.size_code ?? '-'

  const items = order.items ?? []
  const totalQty = items.reduce((s, it) => s + Number(it.qty), 0)
  const deliveryDate = order.delivery_status === 'NOT_DELIVERED' ? order.order_date : order.updated_at

  return (
    <div>
      <PrintLetterhead
        docTitle="SURAT JALAN"
        meta={[
          ['No. SO', order.so_number],
          ['Tanggal Pesanan', formatDate(order.order_date)],
          ['Tanggal Kirim', formatDate(deliveryDate)],
          ['Status', order.delivery_status],
        ]}
      />
      <PrintTwoColumn
        leftTitle="Dikirim Kepada"
        leftLines={[customer?.name ?? '-', customer?.address ?? '', 'Telepon: ' + (customer?.phone || '-')]}
        rightTitle="Dikirim Dari"
        rightLines={[COMPANY.name, COMPANY.address, COMPANY.phone]}
      />
      <PrintSectionTitle>Rincian Ukuran</PrintSectionTitle>
      <PrintTable
        headers={['Produk', 'Ukuran', 'Jml Dikirim']}
        aligns={['left', 'center', 'right']}
        rows={items.map((it) => [productName(it.product_id), sizeCode(it.product_id, it.product_size_id), formatNumber(it.qty, 2)])}
      />
      <PrintTotals rows={[]} grandLabel="Total Jml" grandValue={`${formatNumber(totalQty, 2)} pcs`} />
      <PrintNotes notes={order.notes} />
      <p className="mb-4 text-xs italic text-slate-500">
        Mohon periksa jumlah dan kondisi barang saat diterima. Klaim setelah tanda tangan tidak dapat diproses.
      </p>
      <PrintSignatures labels={['Gudang', 'Pengemudi', 'Diterima Oleh']} />
    </div>
  )
}
