package purchasing

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/accounting"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/audit"
	"github.com/ranji/clothing-erp/internal/coa"
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/ranji/clothing-erp/internal/inventory"
	"github.com/ranji/clothing-erp/internal/numbering"
	"github.com/shopspring/decimal"
)

const acctAP = "2-1000" // Hutang Usaha

type Service struct {
	pool        *pgxpool.Pool
	repo        *Repository
	accountRepo *coa.Repository
	accSvc      *accounting.Service
	invSvc      *inventory.Service
}

func NewService(pool *pgxpool.Pool, repo *Repository, accountRepo *coa.Repository, accSvc *accounting.Service, invSvc *inventory.Service) *Service {
	return &Service{pool: pool, repo: repo, accountRepo: accountRepo, accSvc: accSvc, invSvc: invSvc}
}

func parseOptionalDate(value string, fallback time.Time) (time.Time, error) {
	if value == "" {
		return fallback, nil
	}
	return time.Parse("2006-01-02", value)
}

func (s *Service) CreateBill(ctx context.Context, in CreateBillInput) (SupplierInvoice, error) {
	if in.SupplierID == uuid.Nil {
		return SupplierInvoice{}, apperr.Validation("supplier_id is required")
	}
	if in.DebitAccountCode == "" {
		return SupplierInvoice{}, apperr.Validation("debit_account_code is required")
	}
	if len(in.Items) == 0 {
		return SupplierInvoice{}, apperr.Validation("bill must have at least one item")
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

	subtotal := decimal.Zero
	for _, it := range in.Items {
		if it.Qty.LessThanOrEqual(decimal.Zero) || it.UnitCost.LessThan(decimal.Zero) {
			return SupplierInvoice{}, apperr.Validation("item qty must be > 0 and unit_cost must be >= 0")
		}
		subtotal = subtotal.Add(it.Qty.Mul(it.UnitCost))
	}
	taxTotal := subtotal.Mul(in.TaxRate)
	grandTotal := subtotal.Add(taxTotal)

	var out SupplierInvoice
	err = db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		number, err := numbering.Generate(ctx, tx, numbering.SupplierBill, billDate)
		if err != nil {
			return apperr.Internal("generate bill number", err)
		}

		bill, err := s.repo.InsertBill(ctx, SupplierInvoice{
			BillNumber: number, SupplierID: in.SupplierID, BillDate: billDate, DueDate: dueDate,
			DebitAccountCode: in.DebitAccountCode, Subtotal: subtotal, TaxTotal: taxTotal, GrandTotal: grandTotal, Notes: in.Notes,
		})
		if err != nil {
			return apperr.Internal("create supplier bill", err)
		}

		for _, it := range in.Items {
			item, err := s.repo.InsertBillItem(ctx, SupplierInvoiceItem{
				SupplierInvoiceID: bill.ID, MaterialID: it.MaterialID, Qty: it.Qty, UnitCost: it.UnitCost,
				LineTotal: it.Qty.Mul(it.UnitCost),
			})
			if err != nil {
				return apperr.Internal("create bill item", err)
			}
			bill.Items = append(bill.Items, item)
		}

		if err := audit.Log(ctx, tx, "supplier_invoices", bill.ID, audit.Insert, nil, bill); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = bill
		return nil
	})
	return out, err
}

// PostBill posts DR <debit_account> / CR Hutang Usaha for the bill total and
// receives each line's material into inventory at its billed unit cost --
// this is the "PURCHASE" inventory movement that had no proper source
// before this subledger existed.
func (s *Service) PostBill(ctx context.Context, id uuid.UUID) (SupplierInvoice, error) {
	var out SupplierInvoice
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		bill, err := s.repo.GetBillForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("supplier bill not found")
			}
			return apperr.Internal("load supplier bill", err)
		}
		if bill.Status != BillDraft {
			return apperr.Conflict("only DRAFT bills can be posted")
		}

		if _, err := s.accountRepo.GetByCode(ctx, bill.DebitAccountCode); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.Validation("unknown debit_account_code: " + bill.DebitAccountCode)
			}
			return apperr.Internal("resolve debit account", err)
		}

		journal, err := s.accSvc.PostJournal(ctx, accounting.SourceSupplierBill, &bill.ID, bill.BillDate,
			"Supplier bill "+bill.BillNumber,
			[]accounting.JournalLineInput{
				{AccountCode: bill.DebitAccountCode, Debit: bill.GrandTotal, Description: "Bill " + bill.BillNumber},
				{AccountCode: acctAP, Credit: bill.GrandTotal, Description: "AP " + bill.BillNumber},
			})
		if err != nil {
			return apperr.Wrapf(err, "post supplier bill journal")
		}

		items, err := s.repo.ListBillItems(ctx, id)
		if err != nil {
			return apperr.Internal("load bill items", err)
		}
		matWarehouse, err := s.invSvc.DefaultMaterialWarehouse(ctx)
		if err != nil {
			return apperr.Wrapf(err, "resolve material warehouse")
		}
		for _, it := range items {
			materialID := it.MaterialID
			if _, err := s.invSvc.RecordMovement(ctx, inventory.MovementInput{
				TxnType: inventory.TxnPurchase, ItemType: inventory.ItemMaterial, WarehouseID: matWarehouse.ID,
				MaterialID: &materialID, QtyIn: it.Qty, UnitCost: it.UnitCost,
				RefType: "SUPPLIER_INVOICE", RefID: &bill.ID, TxnDate: bill.BillDate,
				Notes: "Purchase receipt for " + bill.BillNumber,
			}); err != nil {
				return apperr.Wrapf(err, "receive purchased material into inventory")
			}
		}

		if err := s.repo.SetBillPosted(ctx, id, journal.ID); err != nil {
			return apperr.Internal("mark bill posted", err)
		}
		if err := audit.Log(ctx, tx, "supplier_invoices", id, audit.Update, bill.Status, BillPosted); err != nil {
			return apperr.Internal("write audit log", err)
		}

		bill.Status = BillPosted
		bill.JournalID = &journal.ID
		bill.Items = items
		out = bill
		return nil
	})
	return out, err
}

// VoidBill only supports DRAFT bills. Once a bill is POSTED it has already
// received material into inventory (and possibly been consumed downstream
// by production), so undoing it safely requires a purchase-return workflow,
// which is out of scope here rather than pretending a silent reversal is safe.
func (s *Service) VoidBill(ctx context.Context, id uuid.UUID) (SupplierInvoice, error) {
	var out SupplierInvoice
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		bill, err := s.repo.GetBillForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("supplier bill not found")
			}
			return apperr.Internal("load supplier bill", err)
		}
		if bill.Status != BillDraft {
			return apperr.Conflict("only DRAFT bills can be voided; a posted bill requires a purchase return")
		}
		if err := s.repo.SetBillVoid(ctx, id); err != nil {
			return apperr.Internal("void bill", err)
		}
		if err := audit.Log(ctx, tx, "supplier_invoices", id, audit.Update, bill.Status, BillVoid); err != nil {
			return apperr.Internal("write audit log", err)
		}
		bill.Status = BillVoid
		out = bill
		return nil
	})
	return out, err
}

// ApplyPaymentToBill is called by the finance domain after it posts a
// DISBURSEMENT payment's journal and allocation record, mirroring
// sales.Service.ApplyPaymentToInvoice on the AR side.
func (s *Service) ApplyPaymentToBill(ctx context.Context, billID uuid.UUID, amount decimal.Decimal) (SupplierInvoice, error) {
	bill, err := s.repo.GetBillForUpdate(ctx, billID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SupplierInvoice{}, apperr.NotFound("supplier bill not found")
		}
		return SupplierInvoice{}, apperr.Internal("load supplier bill", err)
	}
	if bill.Status == BillVoid {
		return SupplierInvoice{}, apperr.Conflict("cannot pay a voided bill")
	}

	newPaid := bill.PaidAmount.Add(amount)
	newBalance := bill.GrandTotal.Sub(newPaid)
	status := BillPosted
	switch {
	case newBalance.LessThanOrEqual(decimal.Zero):
		status = BillPaid
	case newPaid.IsPositive():
		status = BillPartiallyPaid
	}

	if err := s.repo.UpdateBillPayment(ctx, billID, newPaid, newBalance, status); err != nil {
		return SupplierInvoice{}, apperr.Internal("update bill payment", err)
	}

	bill.PaidAmount = newPaid
	bill.BalanceDue = newBalance
	bill.Status = status
	return bill, nil
}

func (s *Service) GetBill(ctx context.Context, id uuid.UUID) (SupplierInvoice, error) {
	bill, err := s.repo.GetBillByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return SupplierInvoice{}, apperr.NotFound("supplier bill not found")
	}
	if err != nil {
		return SupplierInvoice{}, err
	}
	items, err := s.repo.ListBillItems(ctx, id)
	if err != nil {
		return SupplierInvoice{}, apperr.Internal("load bill items", err)
	}
	bill.Items = items
	return bill, nil
}

func (s *Service) ListBills(ctx context.Context, limit, offset int) ([]SupplierInvoice, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListBills(ctx, limit, offset)
}
