package pdf

import (
	"time"

	"github.com/shopspring/decimal"
)

type ProductionSizeQty struct {
	SizeCode string
	Planned  decimal.Decimal
	Finished decimal.Decimal
}

type ProductionMaterialLine struct {
	MaterialCode string
	MaterialName string
	UOM          string
	PlannedQty   decimal.Decimal
	IssuedQty    decimal.Decimal
}

type ProductionWorkOrderData struct {
	ProdNumber  string
	SONumber    string
	ProductName string
	ProductCode string
	BOMName     string
	Status      string
	PlannedQty  decimal.Decimal
	FinishedQty decimal.Decimal
	StartDate   *time.Time
	EndDate     *time.Time

	Sizes     []ProductionSizeQty
	Materials []ProductionMaterialLine
	Notes     string
}

var processStages = []string{"Cutting", "Printing", "Sewing", "QC"}

// ProductionWorkOrder renders the internal factory work sheet (Slip Kerja
// Produksi): jersey/product spec, size-matrix quantities, material
// requirements, process stage checklist, and a QR/barcode placeholder.
func ProductionWorkOrder(data ProductionWorkOrderData) ([]byte, error) {
	d := newDoc("Production Work Order", data.ProdNumber)

	meta := [][2]string{
		{"Work Order No.", data.ProdNumber},
		{"Status", data.Status},
		{"Start Date", dateStrPtr(data.StartDate)},
		{"End Date", dateStrPtr(data.EndDate)},
	}
	if data.SONumber != "" {
		meta = append(meta, [2]string{"Sales Order", data.SONumber})
	}
	d.Letterhead("SLIP KERJA PRODUKSI", meta)

	d.TwoColumnInfo(
		"Product Specification",
		[]string{data.ProductName, "Code: " + orDash(data.ProductCode), "BOM: " + orDash(data.BOMName)},
		"Quantities",
		[]string{"Planned: " + qty(data.PlannedQty) + " pcs", "Finished: " + qty(data.FinishedQty) + " pcs"},
	)

	d.qrPlaceholder(data.ProdNumber)

	d.SectionTitle("Size Matrix")
	if len(data.Sizes) > 0 {
		headers := make([]string, 0, len(data.Sizes)+1)
		widths := make([]float64, 0, len(data.Sizes)+1)
		aligns := make([]string, 0, len(data.Sizes)+1)
		headers = append(headers, "Metric")
		labelW := 32.0
		widths = append(widths, labelW)
		aligns = append(aligns, "L")
		colW := (usableWidth - labelW) / float64(len(data.Sizes))
		plannedRow := []string{"Planned Qty"}
		finishedRow := []string{"Finished Qty"}
		for _, s := range data.Sizes {
			headers = append(headers, s.SizeCode)
			widths = append(widths, colW)
			aligns = append(aligns, "C")
			plannedRow = append(plannedRow, qty(s.Planned))
			finishedRow = append(finishedRow, qty(s.Finished))
		}
		d.Table(headers, widths, aligns, [][]string{plannedRow, finishedRow})
	}

	d.SectionTitle("Material Requirements")
	headers := []string{"Code", "Material", "UOM", "Planned Qty", "Issued Qty"}
	widths := []float64{24, 66, 20, 34, usableWidth - (24 + 66 + 20 + 34)}
	aligns := []string{"L", "L", "C", "R", "R"}
	rows := make([][]string, 0, len(data.Materials))
	for _, m := range data.Materials {
		rows = append(rows, []string{m.MaterialCode, m.MaterialName, m.UOM, qty(m.PlannedQty), qty(m.IssuedQty)})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"-", "No BOM materials recorded", "", "", ""})
	}
	d.Table(headers, widths, aligns, rows)

	d.SectionTitle("Process Stages")
	d.processChecklist()

	d.NotesBlock(data.Notes)

	d.SignatureBlock([]string{"Cutting Supervisor", "Production Supervisor", "QC Inspector"})

	return d.bytes()
}

// qrPlaceholder draws a small bordered box in the top-right work area
// standing in for a scannable barcode/QR code (not a real generated code).
func (d *Doc) qrPlaceholder(code string) {
	size := 24.0
	x := marginLeft + usableWidth - size
	y := d.GetY()
	d.SetDrawColor(colBorder[0], colBorder[1], colBorder[2])
	d.SetLineWidth(0.3)
	d.Rect(x, y, size, size, "D")
	d.SetFont("Helvetica", "B", 7)
	d.SetTextColor(colMuted[0], colMuted[1], colMuted[2])
	d.SetXY(x, y+size/2-6)
	d.CellFormat(size, 4, "QR CODE", "", 2, "C", false, 0, "")
	d.SetX(x)
	d.SetFont("Helvetica", "", 6)
	d.CellFormat(size, 4, "(placeholder)", "", 2, "C", false, 0, "")
	d.SetX(x)
	d.SetFont("Helvetica", "B", 6.5)
	d.SetTextColor(colPrimary[0], colPrimary[1], colPrimary[2])
	d.CellFormat(size, 4, code, "", 2, "C", false, 0, "")
	d.SetXY(marginLeft, y+size+3)
}

func (d *Doc) processChecklist() {
	n := len(processStages)
	colW := usableWidth / float64(n)
	boxSize := 5.0
	startY := d.GetY()
	d.SetDrawColor(colMuted[0], colMuted[1], colMuted[2])
	d.SetLineWidth(0.3)
	for i, stage := range processStages {
		x := marginLeft + float64(i)*colW
		d.Rect(x, startY, boxSize, boxSize, "D")
		d.SetXY(x+boxSize+2, startY-1)
		d.SetFont("Helvetica", "", 9)
		d.SetTextColor(colPrimary[0], colPrimary[1], colPrimary[2])
		d.CellFormat(colW-boxSize-2, boxSize+2, stage, "", 0, "L", false, 0, "")
	}
	d.SetY(startY + boxSize + 4)
}
