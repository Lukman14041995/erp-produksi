package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
)

const (
	marginLeft  = 15.0
	marginTop   = 15.0
	marginRight = 15.0
	pageWidth   = 210.0 // A4, mm
	usableWidth = pageWidth - marginLeft - marginRight
)

var (
	colPrimary  = [3]int{30, 41, 59}    // slate-800
	colMuted    = [3]int{100, 116, 139} // slate-500
	colBorder   = [3]int{203, 213, 225} // slate-300
	colHeaderBg = [3]int{241, 245, 249} // slate-100
	colStripeBg = [3]int{248, 250, 252} // slate-50
	colAccent   = [3]int{37, 99, 235}   // blue-600
)

// Doc wraps an fpdf document with the layout helpers shared by every
// generated business document (letterhead, tables, totals, signatures).
type Doc struct {
	*fpdf.Fpdf
}

// newDoc starts a fresh A4 portrait document with a standard footer
// (page number + generation timestamp) wired up on every page.
func newDoc(docType, docNumber string) *Doc {
	f := fpdf.New("P", "mm", "A4", "")
	f.SetMargins(marginLeft, marginTop, marginRight)
	f.SetAutoPageBreak(true, 18)
	f.SetTitle(fmt.Sprintf("%s %s", docType, docNumber), false)
	f.AliasNbPages("")
	f.SetFooterFunc(func() {
		f.SetY(-15)
		f.SetDrawColor(colBorder[0], colBorder[1], colBorder[2])
		f.SetLineWidth(0.2)
		f.Line(marginLeft, f.GetY(), marginLeft+usableWidth, f.GetY())
		f.Ln(2)
		f.SetFont("Helvetica", "I", 7)
		f.SetTextColor(colMuted[0], colMuted[1], colMuted[2])
		f.CellFormat(usableWidth/2, 5, "Dibuat oleh Clothing ERP - "+time.Now().Format("02 Jan 2006 15:04"), "", 0, "L", false, 0, "")
		f.CellFormat(usableWidth/2, 5, fmt.Sprintf("Halaman %d / {nb}", f.PageNo()), "", 0, "R", false, 0, "")
	})
	f.AddPage()
	return &Doc{Fpdf: f}
}

func (d *Doc) bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := d.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Letterhead draws the company block on the left and a document title +
// metadata block (number/date/status...) right-aligned on the right, then a
// separating rule.
func (d *Doc) Letterhead(docTitle string, meta [][2]string) {
	leftW := usableWidth * 0.58
	rightW := usableWidth - leftW
	rightX := marginLeft + leftW

	startY := d.GetY()

	d.SetFont("Helvetica", "B", 15)
	d.SetTextColor(colPrimary[0], colPrimary[1], colPrimary[2])
	d.CellFormat(leftW, 6.5, Company.Name, "", 2, "L", false, 0, "")
	d.SetX(marginLeft)
	d.SetFont("Helvetica", "", 8.5)
	d.SetTextColor(colMuted[0], colMuted[1], colMuted[2])
	d.MultiCell(leftW, 4.2, Company.Tagline+"\n"+Company.Address+"\n"+Company.Phone+" - "+Company.Email+"\n"+Company.TaxID, "", "L", false)

	d.SetXY(rightX, startY)
	d.SetFont("Helvetica", "B", 15)
	d.SetTextColor(colAccent[0], colAccent[1], colAccent[2])
	d.CellFormat(rightW, 6.5, docTitle, "", 2, "R", false, 0, "")

	d.SetFont("Helvetica", "", 9)
	for _, kv := range meta {
		d.SetX(rightX)
		d.SetTextColor(colMuted[0], colMuted[1], colMuted[2])
		d.CellFormat(rightW*0.45, 5.2, kv[0], "", 0, "R", false, 0, "")
		d.SetFont("Helvetica", "B", 9)
		d.SetTextColor(colPrimary[0], colPrimary[1], colPrimary[2])
		d.CellFormat(rightW*0.55, 5.2, kv[1], "", 2, "R", false, 0, "")
		d.SetFont("Helvetica", "", 9)
	}

	bottomY := d.GetY()
	d.SetY(bottomY + 3)
	y := d.GetY()
	d.SetDrawColor(colAccent[0], colAccent[1], colAccent[2])
	d.SetLineWidth(0.6)
	d.Line(marginLeft, y, marginLeft+usableWidth, y)
	d.Ln(5)
}

// TwoColumnInfo renders two titled text blocks side by side (e.g. "Bill To"
// / "Ship To", or "Supplier" / "Delivery Info").
func (d *Doc) TwoColumnInfo(leftTitle string, leftLines []string, rightTitle string, rightLines []string) {
	colW := usableWidth/2 - 3
	startY := d.GetY()

	d.SetXY(marginLeft, startY)
	d.SetFont("Helvetica", "B", 9)
	d.SetTextColor(colMuted[0], colMuted[1], colMuted[2])
	d.CellFormat(colW, 5, leftTitle, "", 2, "L", false, 0, "")
	d.SetX(marginLeft)
	d.SetFont("Helvetica", "", 9.5)
	d.SetTextColor(colPrimary[0], colPrimary[1], colPrimary[2])
	for _, line := range leftLines {
		d.SetX(marginLeft)
		d.MultiCell(colW, 4.6, line, "", "L", false)
	}
	leftEndY := d.GetY()

	rightX := marginLeft + usableWidth/2 + 3
	d.SetXY(rightX, startY)
	d.SetFont("Helvetica", "B", 9)
	d.SetTextColor(colMuted[0], colMuted[1], colMuted[2])
	d.CellFormat(colW, 5, rightTitle, "", 2, "L", false, 0, "")
	d.SetX(rightX)
	d.SetFont("Helvetica", "", 9.5)
	d.SetTextColor(colPrimary[0], colPrimary[1], colPrimary[2])
	for _, line := range rightLines {
		d.SetX(rightX)
		d.MultiCell(colW, 4.6, line, "", "L", false)
	}
	rightEndY := d.GetY()

	endY := leftEndY
	if rightEndY > endY {
		endY = rightEndY
	}
	d.SetY(endY + 4)
}

// SectionTitle draws a small bold section heading.
func (d *Doc) SectionTitle(title string) {
	d.SetFont("Helvetica", "B", 10)
	d.SetTextColor(colPrimary[0], colPrimary[1], colPrimary[2])
	d.CellFormat(usableWidth, 6, title, "", 2, "L", false, 0, "")
}

// Table renders a bordered, header-shaded, zebra-striped table.
// len(widths) == len(aligns) == len(headers) == len(row) for every row.
func (d *Doc) Table(headers []string, widths []float64, aligns []string, rows [][]string) {
	const rowH = 6.5

	d.SetFont("Helvetica", "B", 8)
	d.SetFillColor(colHeaderBg[0], colHeaderBg[1], colHeaderBg[2])
	d.SetTextColor(colPrimary[0], colPrimary[1], colPrimary[2])
	d.SetDrawColor(colBorder[0], colBorder[1], colBorder[2])
	d.SetLineWidth(0.2)
	for i, h := range headers {
		d.CellFormat(widths[i], rowH, h, "1", 0, aligns[i], true, 0, "")
	}
	d.Ln(rowH)

	d.SetFont("Helvetica", "", 8.5)
	for rIdx, row := range rows {
		if rIdx%2 == 1 {
			d.SetFillColor(colStripeBg[0], colStripeBg[1], colStripeBg[2])
		} else {
			d.SetFillColor(255, 255, 255)
		}
		d.SetTextColor(colPrimary[0], colPrimary[1], colPrimary[2])
		for i, cell := range row {
			d.CellFormat(widths[i], rowH, cell, "1", 0, aligns[i], true, 0, "")
		}
		d.Ln(rowH)
	}
	d.Ln(2)
}

// TotalsBlock renders right-aligned label/value rows followed by an
// emphasized grand-total row, anchored to the right edge of the page.
func (d *Doc) TotalsBlock(rows [][2]string, grandLabel, grandValue string) {
	width := 78.0
	startX := marginLeft + usableWidth - width

	d.SetFont("Helvetica", "", 9)
	for _, r := range rows {
		d.SetX(startX)
		d.SetTextColor(colMuted[0], colMuted[1], colMuted[2])
		d.CellFormat(width*0.5, 6, r[0], "", 0, "L", false, 0, "")
		d.SetTextColor(colPrimary[0], colPrimary[1], colPrimary[2])
		d.CellFormat(width*0.5, 6, r[1], "", 2, "R", false, 0, "")
	}

	d.SetX(startX)
	y := d.GetY()
	d.SetDrawColor(colBorder[0], colBorder[1], colBorder[2])
	d.Line(startX, y, startX+width, y)
	d.Ln(1.5)

	d.SetFont("Helvetica", "B", 11)
	d.SetX(startX)
	d.SetTextColor(colAccent[0], colAccent[1], colAccent[2])
	d.CellFormat(width*0.5, 7.5, grandLabel, "", 0, "L", false, 0, "")
	d.CellFormat(width*0.5, 7.5, grandValue, "", 2, "R", false, 0, "")
	d.Ln(2)
}

// SignatureBlock lays out N equal-width signature columns with a caption,
// a blank line to sign above, and a "Name / Date" hint underneath.
func (d *Doc) SignatureBlock(labels []string) {
	if d.GetY() > 240 {
		d.AddPage()
	}
	d.Ln(8)
	n := len(labels)
	colW := usableWidth / float64(n)
	startY := d.GetY()

	d.SetFont("Helvetica", "B", 9)
	d.SetTextColor(colPrimary[0], colPrimary[1], colPrimary[2])
	for i, label := range labels {
		d.SetXY(marginLeft+float64(i)*colW, startY)
		d.CellFormat(colW, 5, label, "", 0, "C", false, 0, "")
	}

	lineY := startY + 22
	d.SetDrawColor(colMuted[0], colMuted[1], colMuted[2])
	d.SetLineWidth(0.2)
	for i := range labels {
		x := marginLeft + float64(i)*colW + colW*0.15
		d.Line(x, lineY, x+colW*0.7, lineY)
	}

	d.SetFont("Helvetica", "", 7.5)
	d.SetTextColor(colMuted[0], colMuted[1], colMuted[2])
	for i := range labels {
		d.SetXY(marginLeft+float64(i)*colW, lineY+1.5)
		d.CellFormat(colW, 4, "Nama / Tanggal", "", 0, "C", false, 0, "")
	}
	d.SetY(lineY + 8)
}

// NotesBlock renders a small "Notes" paragraph if non-empty.
func (d *Doc) NotesBlock(notes string) {
	if notes == "" {
		return
	}
	d.SetFont("Helvetica", "B", 8.5)
	d.SetTextColor(colMuted[0], colMuted[1], colMuted[2])
	d.CellFormat(usableWidth, 5, "Catatan", "", 2, "L", false, 0, "")
	d.SetFont("Helvetica", "", 8.5)
	d.SetTextColor(colPrimary[0], colPrimary[1], colPrimary[2])
	d.MultiCell(usableWidth, 4.4, notes, "", "L", false)
	d.Ln(2)
}
