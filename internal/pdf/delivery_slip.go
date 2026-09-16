package pdf

import (
	"time"

	"github.com/shopspring/decimal"
)

type DeliverySlipLine struct {
	ProductName string
	SizeCode    string
	Qty         decimal.Decimal
}

type DeliverySlipData struct {
	SONumber       string
	OrderDate      time.Time
	DeliveryDate   time.Time
	DeliveryStatus string

	CustomerName    string
	CustomerAddress string
	CustomerPhone   string

	Lines []DeliverySlipLine
	Notes string
}

// DeliverySlip renders a shipping document (Surat Jalan): shipping address,
// order reference, size-grid quantity breakdown, and signature/stamp
// placeholders for the recipient, driver, and warehouse.
func DeliverySlip(data DeliverySlipData) ([]byte, error) {
	d := newDoc("Surat Jalan", data.SONumber)

	meta := [][2]string{
		{"No. SO", data.SONumber},
		{"Tanggal Pesanan", dateStr(data.OrderDate)},
		{"Tanggal Kirim", dateStr(data.DeliveryDate)},
		{"Status", data.DeliveryStatus},
	}
	d.Letterhead("SURAT JALAN", meta)

	d.TwoColumnInfo(
		"Dikirim Kepada",
		[]string{data.CustomerName, orDash(data.CustomerAddress), "Telepon: " + orDash(data.CustomerPhone)},
		"Dikirim Dari",
		[]string{Company.Name, Company.Address, Company.Phone},
	)

	d.SectionTitle("Rincian Ukuran")
	headers := []string{"Produk", "Ukuran", "Jml Dikirim"}
	widths := []float64{usableWidth - 40 - 40, 40, 40}
	aligns := []string{"L", "C", "R"}
	rows := make([][]string, 0, len(data.Lines))
	var totalQty decimal.Decimal
	for _, l := range data.Lines {
		rows = append(rows, []string{l.ProductName, l.SizeCode, qty(l.Qty)})
		totalQty = totalQty.Add(l.Qty)
	}
	d.Table(headers, widths, aligns, rows)

	d.TotalsBlock(nil, "Total Jml", qty(totalQty)+" pcs")

	d.NotesBlock(data.Notes)

	d.SetFont("Helvetica", "I", 8)
	d.SetTextColor(colMuted[0], colMuted[1], colMuted[2])
	d.MultiCell(usableWidth, 4.2, "Mohon periksa jumlah dan kondisi barang saat diterima. Klaim setelah tanda tangan tidak dapat diproses.", "", "L", false)

	d.SignatureBlock([]string{"Gudang", "Pengemudi", "Diterima Oleh"})

	return d.bytes()
}
