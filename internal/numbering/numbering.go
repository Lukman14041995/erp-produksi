// Package numbering generates gapless, human-readable document numbers such
// as SO-202609-000001 backed by the document_sequences table. Callers must
// invoke Generate from within the same database transaction as the document
// insert it numbers, so a rolled-back operation does not burn a number.
package numbering

import (
	"context"
	"fmt"
	"time"

	"github.com/ranji/clothing-erp/internal/db"
)

type DocType string

const (
	SalesOrder      DocType = "SO"
	Invoice         DocType = "INV"
	PaymentReceipt  DocType = "PAY"
	ProductionOrder DocType = "PRD"
	Journal         DocType = "JV"
	Expense         DocType = "EXP"
	Adjustment      DocType = "ADJ"
	SupplierBill    DocType = "BILL"
	PurchaseOrder   DocType = "PO"
	GoodsReceipt    DocType = "GRN"
	StockTransfer   DocType = "MUT"
	StockOpname     DocType = "OPN"
)

// Generate atomically reserves and returns the next number for docType in
// the calendar month of `on`, formatted as "<PREFIX>-<YYYYMM>-<000001>".
// The underlying UPSERT takes a row lock on document_sequences, so
// concurrent callers in different transactions serialize instead of
// colliding on the same number.
func Generate(ctx context.Context, q db.Querier, docType DocType, on time.Time) (string, error) {
	period := on.Format("200601")

	var lastNumber int64
	err := q.QueryRow(ctx, `
		INSERT INTO document_sequences (doc_type, period, prefix, last_number)
		VALUES ($1, $2, $1, 1)
		ON CONFLICT (doc_type, period)
		DO UPDATE SET last_number = document_sequences.last_number + 1,
		              updated_at = now()
		RETURNING last_number
	`, string(docType), period).Scan(&lastNumber)
	if err != nil {
		return "", fmt.Errorf("reserve document number for %s: %w", docType, err)
	}

	return fmt.Sprintf("%s-%s-%06d", docType, period, lastNumber), nil
}
