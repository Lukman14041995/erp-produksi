import { PrintLetterhead, PrintSectionTitle, PrintSignatures, PrintTable, PrintTotals, PrintTwoColumn } from './PrintLayout'
import { COMPANY } from './company'
import { formatCurrency, formatDate, formatNumber } from '@/lib/format'
import type { Invoice } from '@/types/sales'
import type { Customer, Product } from '@/types/master'

export function PrintableInvoice({ invoice, soNumber, customer, products }: { invoice: Invoice; soNumber?: string; customer?: Customer; products?: Product[] }) {
  const productName = (id: string) => products?.find((p) => p.id === id)?.name ?? id.slice(0, 8)
  const sizeCode = (pid: string, sid: string) => products?.find((p) => p.id === pid)?.sizes?.find((s) => s.id === sid)?.size_code ?? '-'

  const meta: [string, string][] = [
    ['No. Faktur', invoice.invoice_number],
    ['Tanggal Faktur', formatDate(invoice.invoice_date)],
    ['Jatuh Tempo', invoice.due_date ? formatDate(invoice.due_date) : '-'],
    ['Status', invoice.status],
  ]
  if (soNumber) meta.push(['Pesanan Penjualan', soNumber])

  return (
    <div>
      <PrintLetterhead docTitle="FAKTUR PENJUALAN" meta={meta} />
      <PrintTwoColumn
        leftTitle="Ditagihkan Kepada"
        leftLines={[customer?.name ?? '-', customer?.address ?? '', 'Telepon: ' + (customer?.phone || '-'), 'NPWP: ' + (customer?.tax_id || '-')]}
        rightTitle="Pembayaran Ke"
        rightLines={[COMPANY.bankName, COMPANY.bankAccountName, 'No. Rek. ' + COMPANY.bankAccountNumber]}
      />
      <PrintSectionTitle>Rincian Item</PrintSectionTitle>
      <PrintTable
        headers={['Produk', 'Ukuran', 'Jml', 'Harga Satuan', 'Diskon', 'Pajak', 'Total Baris']}
        aligns={['left', 'center', 'right', 'right', 'right', 'right', 'right']}
        rows={(invoice.items ?? []).map((it) => [
          productName(it.product_id),
          sizeCode(it.product_id, it.product_size_id),
          formatNumber(it.qty, 2),
          formatCurrency(it.unit_price),
          formatCurrency(it.discount),
          `${(Number(it.tax_rate) * 100).toFixed(1)}%`,
          formatCurrency(it.line_total),
        ])}
      />
      <PrintTotals
        rows={[
          ['Subtotal', formatCurrency(invoice.subtotal)],
          ['Diskon', '-' + formatCurrency(invoice.discount_total)],
          ['Pajak', formatCurrency(invoice.tax_total)],
        ]}
        grandLabel="Total Keseluruhan"
        grandValue={formatCurrency(invoice.grand_total)}
      />
      <PrintTotals rows={[['Jumlah Dibayar', formatCurrency(invoice.paid_amount)]]} grandLabel="Sisa Tagihan" grandValue={formatCurrency(invoice.balance_due)} />
      <p className="mb-4 text-xs italic text-slate-500">
        Silakan transfer pembayaran ke {COMPANY.bankName} ({COMPANY.bankAccountName}) - No. Rek. {COMPANY.bankAccountNumber}
      </p>
      <PrintSignatures labels={['Dibuat Oleh', 'Disetujui Oleh', 'Diterima Oleh (Pelanggan)']} />
    </div>
  )
}
