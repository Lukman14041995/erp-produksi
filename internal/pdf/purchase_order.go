package pdf

import (
	"time"

	"github.com/shopspring/decimal"
)

type PurchaseOrderLine struct {
	MaterialCode string
	MaterialName string
	UOM          string
	Qty          decimal.Decimal
	UnitCost     decimal.Decimal
	LineTotal    decimal.Decimal
}

type PurchaseOrderData struct {
	PONumber     string
	OrderDate    time.Time
	ExpectedDate *time.Time
	Status       string

	SupplierName    string
	SupplierAddress string
	SupplierPhone   string
	SupplierContact string
	SupplierTaxID   string

	Lines      []PurchaseOrderLine
	Subtotal   decimal.Decimal
	GrandTotal decimal.Decimal
	Notes      string
}

// PurchaseOrder renders an official PO document (Pesanan Pembelian) sent to
// a supplier: letterhead, supplier + delivery info, standard unit costs,
// and an approval signature block.
func PurchaseOrder(data PurchaseOrderData) ([]byte, error) {
	d := newDoc("Pesanan Pembelian", data.PONumber)

	meta := [][2]string{
		{"No. PO", data.PONumber},
		{"Tanggal Pesanan", dateStr(data.OrderDate)},
		{"Tanggal Diharapkan", dateStrPtr(data.ExpectedDate)},
		{"Status", data.Status},
	}
	d.Letterhead("PESANAN PEMBELIAN", meta)

	d.TwoColumnInfo(
		"Pemasok",
		[]string{data.SupplierName, orDash(data.SupplierAddress), "Kontak: " + orDash(data.SupplierContact), "Telepon: " + orDash(data.SupplierPhone), "NPWP: " + orDash(data.SupplierTaxID)},
		"Dikirim Ke",
		[]string{Company.Name, Company.Address, Company.Phone},
	)

	d.SectionTitle("Spesifikasi Item")
	headers := []string{"Kode", "Bahan Baku", "Satuan", "Jml", "Biaya Satuan", "Total Baris"}
	widths := []float64{22, 62, 18, 24, 30, usableWidth - (22 + 62 + 18 + 24 + 30)}
	aligns := []string{"L", "L", "C", "R", "R", "R"}
	rows := make([][]string, 0, len(data.Lines))
	for _, l := range data.Lines {
		rows = append(rows, []string{l.MaterialCode, l.MaterialName, l.UOM, qty(l.Qty), money(l.UnitCost), money(l.LineTotal)})
	}
	d.Table(headers, widths, aligns, rows)

	d.TotalsBlock([][2]string{{"Subtotal", money(data.Subtotal)}}, "Total Keseluruhan", money(data.GrandTotal))

	d.NotesBlock(data.Notes)

	d.SetFont("Helvetica", "I", 8)
	d.SetTextColor(colMuted[0], colMuted[1], colMuted[2])
	d.MultiCell(usableWidth, 4.2, "Biaya satuan di atas telah disepakati dengan pemasok sebelum pengiriman dan dicocokkan dengan penerimaan barang serta tagihan pemasok.", "", "L", false)

	d.SignatureBlock([]string{"Diminta Oleh", "Disetujui Oleh", "Persetujuan Pemasok"})

	return d.bytes()
}
