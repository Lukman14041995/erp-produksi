package pdf

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type InvoiceLine struct {
	ProductName string
	SizeCode    string
	Qty         decimal.Decimal
	UnitPrice   decimal.Decimal
	Discount    decimal.Decimal
	TaxRate     decimal.Decimal
	LineTotal   decimal.Decimal
}

type InvoiceData struct {
	InvoiceNumber string
	SONumber      string
	InvoiceDate   time.Time
	DueDate       *time.Time
	Status        string

	CustomerName    string
	CustomerAddress string
	CustomerPhone   string
	CustomerTaxID   string

	Lines []InvoiceLine

	Subtotal      decimal.Decimal
	DiscountTotal decimal.Decimal
	TaxTotal      decimal.Decimal
	GrandTotal    decimal.Decimal
	PaidAmount    decimal.Decimal
	BalanceDue    decimal.Decimal

	Notes string
}

// Invoice renders a Sales Invoice (Faktur Penjualan): letterhead, bill-to
// block, itemized size-matrix lines, tax/discount totals, bank transfer
// info, and a signature block.
func Invoice(data InvoiceData) ([]byte, error) {
	d := newDoc("Faktur", data.InvoiceNumber)

	meta := [][2]string{
		{"No. Faktur", data.InvoiceNumber},
		{"Tanggal Faktur", dateStr(data.InvoiceDate)},
		{"Jatuh Tempo", dateStrPtr(data.DueDate)},
		{"Status", data.Status},
	}
	if data.SONumber != "" {
		meta = append(meta, [2]string{"Pesanan Penjualan", data.SONumber})
	}
	d.Letterhead("FAKTUR PENJUALAN", meta)

	d.TwoColumnInfo(
		"Ditagihkan Kepada",
		[]string{data.CustomerName, orDash(data.CustomerAddress), "Telepon: " + orDash(data.CustomerPhone), "NPWP: " + orDash(data.CustomerTaxID)},
		"Pembayaran Ke",
		[]string{Company.BankName, Company.BankAccountName, "No. Rek. " + Company.BankAccountNumber},
	)

	d.SectionTitle("Rincian Item")
	headers := []string{"Produk", "Ukuran", "Jml", "Harga Satuan", "Diskon", "Pajak", "Total Baris"}
	widths := []float64{50, 16, 16, 28, 20, 16, usableWidth - (50 + 16 + 16 + 28 + 20 + 16)}
	aligns := []string{"L", "C", "R", "R", "R", "R", "R"}
	rows := make([][]string, 0, len(data.Lines))
	for _, l := range data.Lines {
		rows = append(rows, []string{
			l.ProductName,
			l.SizeCode,
			qty(l.Qty),
			money(l.UnitPrice),
			money(l.Discount),
			l.TaxRate.Mul(decimal.NewFromInt(100)).StringFixed(1) + "%",
			money(l.LineTotal),
		})
	}
	d.Table(headers, widths, aligns, rows)

	totalRows := [][2]string{
		{"Subtotal", money(data.Subtotal)},
		{"Diskon", "-" + money(data.DiscountTotal)},
		{"Pajak", money(data.TaxTotal)},
	}
	d.TotalsBlock(totalRows, "Total Keseluruhan", money(data.GrandTotal))

	paidRows := [][2]string{
		{"Jumlah Dibayar", money(data.PaidAmount)},
	}
	d.TotalsBlock(paidRows, "Sisa Tagihan", money(data.BalanceDue))

	d.NotesBlock(data.Notes)

	d.SetFont("Helvetica", "I", 8)
	d.SetTextColor(colMuted[0], colMuted[1], colMuted[2])
	d.MultiCell(usableWidth, 4.2, fmt.Sprintf("Silakan transfer pembayaran ke %s (%s) - No. Rek. %s", Company.BankName, Company.BankAccountName, Company.BankAccountNumber), "", "L", false)

	d.SignatureBlock([]string{"Dibuat Oleh", "Disetujui Oleh", "Diterima Oleh (Pelanggan)"})

	return d.bytes()
}
