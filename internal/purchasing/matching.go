package purchasing

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ranji/clothing-erp/internal/accounting"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/audit"
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/ranji/clothing-erp/internal/inventory"
	"github.com/ranji/clothing-erp/internal/numbering"
	"github.com/shopspring/decimal"
)

const (
	acctUnbilledAP    = "2-1100" // Hutang Belum Ditagih
	acctMaterialStock = "1-1200" // Persediaan Bahan Baku
	acctPriceVariance = "6-9100" // Selisih Harga Pembelian
)

// ---- Purchase Orders ----

func (s *Service) CreatePO(ctx context.Context, in CreatePOInput) (PurchaseOrder, error) {
	if in.SupplierID == uuid.Nil {
		return PurchaseOrder{}, apperr.Validation("supplier_id is required")
	}
	if len(in.Items) == 0 {
		return PurchaseOrder{}, apperr.Validation("purchase order must have at least one item")
	}

	orderDate, err := parseOptionalDate(in.OrderDate, time.Now())
	if err != nil {
		return PurchaseOrder{}, apperr.Validation("order_date must be YYYY-MM-DD")
	}
	var expectedDate *time.Time
	if in.ExpectedDate != "" {
		d, err := time.Parse("2006-01-02", in.ExpectedDate)
		if err != nil {
			return PurchaseOrder{}, apperr.Validation("expected_date must be YYYY-MM-DD")
		}
		expectedDate = &d
	}

	subtotal := decimal.Zero
	for _, it := range in.Items {
		if it.Qty.LessThanOrEqual(decimal.Zero) || it.UnitCost.LessThan(decimal.Zero) {
			return PurchaseOrder{}, apperr.Validation("item qty must be > 0 and unit_cost must be >= 0")
		}
		subtotal = subtotal.Add(it.Qty.Mul(it.UnitCost))
	}

	var out PurchaseOrder
	err = db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		number, err := numbering.Generate(ctx, tx, numbering.PurchaseOrder, orderDate)
		if err != nil {
			return apperr.Internal("generate PO number", err)
		}

		po, err := s.repo.InsertPO(ctx, PurchaseOrder{
			PONumber: number, SupplierID: in.SupplierID, OrderDate: orderDate, ExpectedDate: expectedDate,
			Subtotal: subtotal, GrandTotal: subtotal, Notes: in.Notes,
		})
		if err != nil {
			return apperr.Internal("create purchase order", err)
		}

		for _, it := range in.Items {
			item, err := s.repo.InsertPOItem(ctx, PurchaseOrderItem{
				PurchaseOrderID: po.ID, MaterialID: it.MaterialID, Qty: it.Qty, UnitCost: it.UnitCost,
				LineTotal: it.Qty.Mul(it.UnitCost),
			})
			if err != nil {
				return apperr.Internal("create purchase order item", err)
			}
			po.Items = append(po.Items, item)
		}

		if err := audit.Log(ctx, tx, "purchase_orders", po.ID, audit.Insert, nil, po); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = po
		return nil
	})
	return out, err
}

func (s *Service) ApprovePO(ctx context.Context, id uuid.UUID) (PurchaseOrder, error) {
	var out PurchaseOrder
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		po, err := s.repo.GetPOForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("purchase order not found")
			}
			return apperr.Internal("load purchase order", err)
		}
		if po.Status != POStatusDraft {
			return apperr.Conflict("only DRAFT purchase orders can be approved")
		}
		if err := s.repo.SetPOStatus(ctx, id, POStatusApproved); err != nil {
			return apperr.Internal("approve purchase order", err)
		}
		if err := s.repo.SetPOApprovedAt(ctx, id); err != nil {
			return apperr.Internal("stamp approved_at", err)
		}
		if err := audit.Log(ctx, tx, "purchase_orders", id, audit.Update, po.Status, POStatusApproved); err != nil {
			return apperr.Internal("write audit log", err)
		}
		po.Status = POStatusApproved
		out = po
		return nil
	})
	return out, err
}

func (s *Service) CancelPO(ctx context.Context, id uuid.UUID) (PurchaseOrder, error) {
	var out PurchaseOrder
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		po, err := s.repo.GetPOForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("purchase order not found")
			}
			return apperr.Internal("load purchase order", err)
		}
		if po.Status != POStatusDraft && po.Status != POStatusApproved {
			return apperr.Conflict("only DRAFT or APPROVED purchase orders can be cancelled")
		}

		items, err := s.repo.ListPOItems(ctx, id)
		if err != nil {
			return apperr.Internal("load purchase order items", err)
		}
		for _, it := range items {
			if it.QtyReceived.IsPositive() {
				return apperr.Conflict("cannot cancel a purchase order that already has goods received")
			}
		}

		if err := s.repo.SetPOStatus(ctx, id, POStatusCancelled); err != nil {
			return apperr.Internal("cancel purchase order", err)
		}
		if err := s.repo.SetPOCancelledAt(ctx, id); err != nil {
			return apperr.Internal("stamp cancelled_at", err)
		}
		if err := audit.Log(ctx, tx, "purchase_orders", id, audit.Update, po.Status, POStatusCancelled); err != nil {
			return apperr.Internal("write audit log", err)
		}
		po.Status = POStatusCancelled
		out = po
		return nil
	})
	return out, err
}

func (s *Service) GetPO(ctx context.Context, id uuid.UUID) (PurchaseOrder, error) {
	po, err := s.repo.GetPOByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return PurchaseOrder{}, apperr.NotFound("purchase order not found")
	}
	if err != nil {
		return PurchaseOrder{}, err
	}
	items, err := s.repo.ListPOItems(ctx, id)
	if err != nil {
		return PurchaseOrder{}, apperr.Internal("load purchase order items", err)
	}
	po.Items = items
	return po, nil
}

func (s *Service) ListPOs(ctx context.Context, limit, offset int) ([]PurchaseOrder, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListPOs(ctx, limit, offset)
}

func (s *Service) recomputePOReceiptStatus(ctx context.Context, poID uuid.UUID) error {
	items, err := s.repo.ListPOItems(ctx, poID)
	if err != nil {
		return apperr.Internal("load purchase order items", err)
	}
	totalOrdered, totalReceived := decimal.Zero, decimal.Zero
	for _, it := range items {
		totalOrdered = totalOrdered.Add(it.Qty)
		totalReceived = totalReceived.Add(it.QtyReceived)
	}
	status := POStatusApproved
	switch {
	case totalReceived.GreaterThanOrEqual(totalOrdered):
		status = POStatusFullyReceived
	case totalReceived.IsPositive():
		status = POStatusPartiallyReceived
	}
	return s.repo.SetPOStatus(ctx, poID, status)
}

func (s *Service) recomputePOBillingStatus(ctx context.Context, poID uuid.UUID) error {
	items, err := s.repo.ListPOItems(ctx, poID)
	if err != nil {
		return apperr.Internal("load purchase order items", err)
	}
	totalOrdered, totalBilled := decimal.Zero, decimal.Zero
	for _, it := range items {
		totalOrdered = totalOrdered.Add(it.Qty)
		totalBilled = totalBilled.Add(it.QtyBilled)
	}
	status := POBillingUnbilled
	switch {
	case totalBilled.GreaterThanOrEqual(totalOrdered) && totalOrdered.IsPositive():
		status = POBillingFullyBilled
	case totalBilled.IsPositive():
		status = POBillingPartiallyBilled
	}
	return s.repo.SetPOBillingStatus(ctx, poID, status)
}

// ---- Goods Receipts ----

// CreateGoodsReceipt posts DR Persediaan Bahan Baku / CR Hutang Belum
// Ditagih (Unbilled AP) at the PO's agreed unit cost, and receives the
// material into inventory -- stock and the accrued liability move together,
// before the supplier's actual bill ever arrives.
func (s *Service) CreateGoodsReceipt(ctx context.Context, in CreateGRNInput) (GoodsReceipt, error) {
	if in.PurchaseOrderID == uuid.Nil {
		return GoodsReceipt{}, apperr.Validation("purchase_order_id is required")
	}
	if len(in.Items) == 0 {
		return GoodsReceipt{}, apperr.Validation("goods receipt must have at least one item")
	}
	receiptDate, err := parseOptionalDate(in.ReceiptDate, time.Now())
	if err != nil {
		return GoodsReceipt{}, apperr.Validation("receipt_date must be YYYY-MM-DD")
	}

	var out GoodsReceipt
	err = db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		po, err := s.repo.GetPOForUpdate(ctx, in.PurchaseOrderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("purchase order not found")
			}
			return apperr.Internal("load purchase order", err)
		}
		if po.Status != POStatusApproved && po.Status != POStatusPartiallyReceived {
			return apperr.Conflict("purchase order must be APPROVED or PARTIALLY_RECEIVED to receive goods")
		}

		number, err := numbering.Generate(ctx, tx, numbering.GoodsReceipt, receiptDate)
		if err != nil {
			return apperr.Internal("generate GRN number", err)
		}

		grn, err := s.repo.InsertGRN(ctx, GoodsReceipt{
			GRNNumber: number, PurchaseOrderID: po.ID, SupplierID: po.SupplierID, ReceiptDate: receiptDate, Notes: in.Notes,
		})
		if err != nil {
			return apperr.Internal("create goods receipt", err)
		}

		matWarehouse, err := s.invSvc.DefaultMaterialWarehouse(ctx)
		if err != nil {
			return apperr.Wrapf(err, "resolve material warehouse")
		}

		totalValue := decimal.Zero
		for _, it := range in.Items {
			if it.QtyReceived.LessThanOrEqual(decimal.Zero) {
				return apperr.Validation("qty_received must be greater than zero")
			}
			poItem, err := s.repo.GetPOItemForUpdate(ctx, it.PurchaseOrderItemID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return apperr.Validation("unknown purchase_order_item_id")
				}
				return apperr.Internal("load purchase order item", err)
			}
			if poItem.PurchaseOrderID != po.ID {
				return apperr.Validation("purchase order item does not belong to this purchase order")
			}
			remaining := poItem.Qty.Sub(poItem.QtyReceived)
			if it.QtyReceived.GreaterThan(remaining) {
				return apperr.Validation("qty_received exceeds remaining ordered quantity for this line")
			}

			grnItem, err := s.repo.InsertGRNItem(ctx, GoodsReceiptItem{
				GoodsReceiptID: grn.ID, PurchaseOrderItemID: poItem.ID, MaterialID: poItem.MaterialID,
				QtyReceived: it.QtyReceived, UnitCost: poItem.UnitCost, LineTotal: it.QtyReceived.Mul(poItem.UnitCost),
			})
			if err != nil {
				return apperr.Internal("create goods receipt item", err)
			}
			grn.Items = append(grn.Items, grnItem)

			if err := s.repo.UpdatePOItemReceived(ctx, poItem.ID, poItem.QtyReceived.Add(it.QtyReceived)); err != nil {
				return apperr.Internal("update received quantity", err)
			}

			materialID := poItem.MaterialID
			if _, err := s.invSvc.RecordMovement(ctx, inventory.MovementInput{
				TxnType: inventory.TxnPurchase, ItemType: inventory.ItemMaterial, WarehouseID: matWarehouse.ID,
				MaterialID: &materialID, QtyIn: it.QtyReceived, UnitCost: poItem.UnitCost,
				RefType: "GOODS_RECEIPT", RefID: &grn.ID, TxnDate: receiptDate,
				Notes: "Goods receipt " + number + " for " + po.PONumber,
			}); err != nil {
				return apperr.Wrapf(err, "receive material into inventory")
			}

			totalValue = totalValue.Add(it.QtyReceived.Mul(poItem.UnitCost))
		}

		journal, err := s.accSvc.PostJournal(ctx, accounting.SourceGoodsReceipt, &grn.ID, receiptDate,
			"Goods receipt "+number+" for "+po.PONumber,
			[]accounting.JournalLineInput{
				{AccountCode: acctMaterialStock, Debit: totalValue, Description: "Material received " + number},
				{AccountCode: acctUnbilledAP, Credit: totalValue, Description: "Unbilled AP accrued " + number},
			})
		if err != nil {
			return apperr.Wrapf(err, "post goods receipt journal")
		}
		if err := s.repo.SetGRNJournal(ctx, grn.ID, journal.ID); err != nil {
			return apperr.Internal("link goods receipt journal", err)
		}
		grn.JournalID = &journal.ID

		if err := s.recomputePOReceiptStatus(ctx, po.ID); err != nil {
			return err
		}

		if err := audit.Log(ctx, tx, "goods_receipts", grn.ID, audit.Insert, nil, grn); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = grn
		return nil
	})
	return out, err
}

func (s *Service) GetGoodsReceipt(ctx context.Context, id uuid.UUID) (GoodsReceipt, error) {
	grn, err := s.repo.GetGRNByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return GoodsReceipt{}, apperr.NotFound("goods receipt not found")
	}
	if err != nil {
		return GoodsReceipt{}, err
	}
	items, err := s.repo.ListGRNItems(ctx, id)
	if err != nil {
		return GoodsReceipt{}, apperr.Internal("load goods receipt items", err)
	}
	grn.Items = items
	return grn, nil
}

// ListGoodsReceiptsByPO includes each GRN's items (unlike the paginated
// ListGoodsReceipts), since callers use this to build the bill-matching
// desk, which needs per-line remaining-unbilled quantities.
func (s *Service) ListGoodsReceiptsByPO(ctx context.Context, poID uuid.UUID) ([]GoodsReceipt, error) {
	grns, err := s.repo.ListGRNsByPO(ctx, poID)
	if err != nil {
		return nil, err
	}
	for i := range grns {
		items, err := s.repo.ListGRNItems(ctx, grns[i].ID)
		if err != nil {
			return nil, apperr.Internal("load goods receipt items", err)
		}
		grns[i].Items = items
	}
	return grns, nil
}

func (s *Service) ListGoodsReceipts(ctx context.Context, limit, offset int) ([]GoodsReceipt, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListGRNs(ctx, limit, offset)
}

// ---- 3-way matched Supplier Bill ----

// CreateBillFromGRN matches a supplier's actual bill against already-received
// GRN lines. It clears the Unbilled AP accrued at receipt time, books any
// price variance between the PO's agreed cost and the bill's actual cost to
// a dedicated variance account, and recognizes the real Hutang Usaha payable
// -- all in one atomic, immediately-posted transaction (no separate DRAFT
// step, since every figure is already validated against the GRN).
func (s *Service) CreateBillFromGRN(ctx context.Context, in CreateBillFromGRNInput) (SupplierInvoice, error) {
	if in.PurchaseOrderID == uuid.Nil {
		return SupplierInvoice{}, apperr.Validation("purchase_order_id is required")
	}
	if len(in.Items) == 0 {
		return SupplierInvoice{}, apperr.Validation("bill must match at least one goods receipt line")
	}
	billDate, err := parseOptionalDate(in.BillDate, time.Now())
	if err != nil {
		return SupplierInvoice{}, apperr.Validation("bill_date must be YYYY-MM-DD")
	}
	var dueDate *time.Time
	if in.DueDate != "" {
		d, err := time.Parse("2006-01-02", in.DueDate)
		if err != nil {
			return SupplierInvoice{}, apperr.Validation("due_date must be YYYY-MM-DD")
		}
		dueDate = &d
	}

	var out SupplierInvoice
	err = db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		po, err := s.repo.GetPOForUpdate(ctx, in.PurchaseOrderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("purchase order not found")
			}
			return apperr.Internal("load purchase order", err)
		}
		if po.Status != POStatusPartiallyReceived && po.Status != POStatusFullyReceived {
			return apperr.Conflict("purchase order has no received goods to bill against")
		}

		number, err := numbering.Generate(ctx, tx, numbering.SupplierBill, billDate)
		if err != nil {
			return apperr.Internal("generate bill number", err)
		}

		bill, err := s.repo.InsertBill(ctx, SupplierInvoice{
			BillNumber: number, SupplierID: po.SupplierID, PurchaseOrderID: &po.ID, BillDate: billDate, DueDate: dueDate,
			DebitAccountCode: acctUnbilledAP, Notes: in.Notes,
		})
		if err != nil {
			return apperr.Internal("create supplier bill", err)
		}

		standardTotal, actualTotal, varianceTotal := decimal.Zero, decimal.Zero, decimal.Zero

		for _, it := range in.Items {
			if it.Qty.LessThanOrEqual(decimal.Zero) {
				return apperr.Validation("matched qty must be greater than zero")
			}
			if it.BillUnitCost.LessThan(decimal.Zero) {
				return apperr.Validation("bill_unit_cost must be >= 0")
			}

			grnItem, err := s.repo.GetGRNItemForUpdate(ctx, it.GoodsReceiptItemID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return apperr.Validation("unknown goods_receipt_item_id")
				}
				return apperr.Internal("load goods receipt item", err)
			}

			poItem, err := s.repo.GetPOItemForUpdate(ctx, grnItem.PurchaseOrderItemID)
			if err != nil {
				return apperr.Internal("load purchase order item", err)
			}
			if poItem.PurchaseOrderID != po.ID {
				return apperr.Validation("goods receipt item does not belong to this purchase order")
			}

			// Rule: a bill can never exceed what was actually received and not yet billed.
			remainingUnbilled := grnItem.QtyReceived.Sub(grnItem.QtyBilled)
			if it.Qty.GreaterThan(remainingUnbilled) {
				return apperr.Validation("bill quantity exceeds unbilled received quantity for this line")
			}

			standardCost := grnItem.UnitCost // the PO's agreed cost, snapshotted at receipt time
			lineStandard := it.Qty.Mul(standardCost)
			lineActual := it.Qty.Mul(it.BillUnitCost)
			lineVariance := lineActual.Sub(lineStandard)

			standardTotal = standardTotal.Add(lineStandard)
			actualTotal = actualTotal.Add(lineActual)
			varianceTotal = varianceTotal.Add(lineVariance)

			item, err := s.repo.InsertBillItem(ctx, SupplierInvoiceItem{
				SupplierInvoiceID: bill.ID, MaterialID: grnItem.MaterialID, Qty: it.Qty, UnitCost: it.BillUnitCost,
				LineTotal: lineActual, PurchaseOrderItemID: &poItem.ID, GoodsReceiptItemID: &grnItem.ID, PriceVariance: lineVariance,
			})
			if err != nil {
				return apperr.Internal("create bill item", err)
			}
			bill.Items = append(bill.Items, item)

			if err := s.repo.UpdateGRNItemBilled(ctx, grnItem.ID, grnItem.QtyBilled.Add(it.Qty)); err != nil {
				return apperr.Internal("update goods receipt item billed qty", err)
			}
			if err := s.repo.UpdatePOItemBilled(ctx, poItem.ID, poItem.QtyBilled.Add(it.Qty)); err != nil {
				return apperr.Internal("update purchase order item billed qty", err)
			}
		}

		lines := []accounting.JournalLineInput{
			{AccountCode: acctUnbilledAP, Debit: standardTotal, Description: "Clear unbilled AP " + number},
		}
		if varianceTotal.IsPositive() {
			lines = append(lines, accounting.JournalLineInput{AccountCode: acctPriceVariance, Debit: varianceTotal, Description: "Unfavorable price variance " + number})
		} else if varianceTotal.IsNegative() {
			lines = append(lines, accounting.JournalLineInput{AccountCode: acctPriceVariance, Credit: varianceTotal.Neg(), Description: "Favorable price variance " + number})
		}
		lines = append(lines, accounting.JournalLineInput{AccountCode: acctAP, Credit: actualTotal, Description: "AP recognized " + number})

		journal, err := s.accSvc.PostJournal(ctx, accounting.SourceSupplierBill, &bill.ID, billDate,
			"3-way matched bill "+number+" for "+po.PONumber, lines)
		if err != nil {
			return apperr.Wrapf(err, "post matched bill journal")
		}

		if err := s.repo.SetBillTotalsAndPosted(ctx, bill.ID, actualTotal, actualTotal, journal.ID); err != nil {
			return apperr.Internal("finalize bill totals", err)
		}
		bill.Status = BillPosted
		bill.Subtotal = actualTotal
		bill.GrandTotal = actualTotal
		bill.BalanceDue = actualTotal
		bill.JournalID = &journal.ID

		if err := s.recomputePOBillingStatus(ctx, po.ID); err != nil {
			return err
		}

		if err := audit.Log(ctx, tx, "supplier_invoices", bill.ID, audit.Insert, nil, bill); err != nil {
			return apperr.Internal("write audit log", err)
		}

		out = bill
		return nil
	})
	return out, err
}
