package finance

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
	"github.com/ranji/clothing-erp/internal/numbering"
	"github.com/ranji/clothing-erp/internal/purchasing"
	"github.com/ranji/clothing-erp/internal/sales"
	"github.com/shopspring/decimal"
)

const acctAR = "1-1100" // Piutang Usaha
const acctAP = "2-1000" // Hutang Usaha

type Service struct {
	pool          *pgxpool.Pool
	repo          *Repository
	accSvc        *accounting.Service
	accountRepo   *coa.Repository
	salesSvc      *sales.Service
	purchasingSvc *purchasing.Service
}

func NewService(pool *pgxpool.Pool, repo *Repository, accSvc *accounting.Service, accountRepo *coa.Repository, salesSvc *sales.Service, purchasingSvc *purchasing.Service) *Service {
	return &Service{pool: pool, repo: repo, accSvc: accSvc, accountRepo: accountRepo, salesSvc: salesSvc, purchasingSvc: purchasingSvc}
}

// ---- Payments ----

func (s *Service) CreatePayment(ctx context.Context, in CreatePaymentInput) (Payment, error) {
	if in.Amount.LessThanOrEqual(decimal.Zero) {
		return Payment{}, apperr.Validation("amount must be greater than zero")
	}
	switch in.PaymentType {
	case PaymentReceipt:
		if in.CustomerID == nil {
			return Payment{}, apperr.Validation("customer_id is required for a RECEIPT payment")
		}
		if len(in.BillAllocations) > 0 {
			return Payment{}, apperr.Validation("bill_allocations are only valid for DISBURSEMENT payments")
		}
	case PaymentDisbursement:
		if in.SupplierID == nil {
			return Payment{}, apperr.Validation("supplier_id is required for a DISBURSEMENT payment")
		}
		if len(in.Allocations) > 0 {
			return Payment{}, apperr.Validation("allocations are only valid for RECEIPT payments")
		}
	default:
		return Payment{}, apperr.Validation("payment_type must be RECEIPT or DISBURSEMENT")
	}

	allocatedTotal := decimal.Zero
	for _, a := range in.Allocations {
		if a.Amount.LessThanOrEqual(decimal.Zero) {
			return Payment{}, apperr.Validation("allocation amount must be greater than zero")
		}
		allocatedTotal = allocatedTotal.Add(a.Amount)
	}
	for _, a := range in.BillAllocations {
		if a.Amount.LessThanOrEqual(decimal.Zero) {
			return Payment{}, apperr.Validation("bill allocation amount must be greater than zero")
		}
		allocatedTotal = allocatedTotal.Add(a.Amount)
	}
	if allocatedTotal.GreaterThan(in.Amount) {
		return Payment{}, apperr.Validation("total allocations cannot exceed payment amount")
	}

	paymentDate := time.Now()
	if in.PaymentDate != "" {
		parsed, err := time.Parse("2006-01-02", in.PaymentDate)
		if err != nil {
			return Payment{}, apperr.Validation("payment_date must be YYYY-MM-DD")
		}
		paymentDate = parsed
	}

	var out Payment
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		account, err := s.accountRepo.GetByCode(ctx, in.CashBankAccountCode)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.Validation("unknown cash/bank account code: " + in.CashBankAccountCode)
			}
			return apperr.Internal("resolve cash/bank account", err)
		}

		number, err := numbering.Generate(ctx, tx, numbering.PaymentReceipt, paymentDate)
		if err != nil {
			return apperr.Internal("generate payment number", err)
		}

		payment, err := s.repo.InsertPayment(ctx, Payment{
			PaymentNumber: number, PaymentType: in.PaymentType, CustomerID: in.CustomerID, SupplierID: in.SupplierID,
			CashBankAccountID: account.ID, PaymentDate: paymentDate, Amount: in.Amount, Method: in.Method,
			ReferenceNo: in.ReferenceNo, Notes: in.Notes,
		})
		if err != nil {
			return apperr.Internal("create payment", err)
		}

		for _, a := range in.Allocations {
			alloc, err := s.repo.InsertAllocation(ctx, PaymentAllocation{PaymentID: payment.ID, InvoiceID: a.InvoiceID, AmountAllocated: a.Amount})
			if err != nil {
				return apperr.Internal("create payment allocation", err)
			}
			payment.Allocations = append(payment.Allocations, alloc)
		}
		for _, a := range in.BillAllocations {
			alloc, err := s.repo.InsertSupplierAllocation(ctx, SupplierPaymentAllocation{PaymentID: payment.ID, SupplierInvoiceID: a.SupplierInvoiceID, AmountAllocated: a.Amount})
			if err != nil {
				return apperr.Internal("create bill allocation", err)
			}
			payment.BillAllocations = append(payment.BillAllocations, alloc)
		}

		if err := audit.Log(ctx, tx, "payments", payment.ID, audit.Insert, nil, payment); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = payment
		return nil
	})
	return out, err
}

// PostPayment posts DR Bank/CR AR (receipt) or DR AP/CR Bank (disbursement)
// and, for receipts, applies each allocation to its invoice via the sales
// domain -- this is the payment allocation engine.
func (s *Service) PostPayment(ctx context.Context, id uuid.UUID) (Payment, error) {
	var out Payment
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		payment, err := s.repo.GetPaymentForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("payment not found")
			}
			return apperr.Internal("load payment", err)
		}
		if payment.Status != StatusDraft {
			return apperr.Conflict("only DRAFT payments can be posted")
		}

		account, err := s.accountRepo.GetByID(ctx, payment.CashBankAccountID)
		if err != nil {
			return apperr.Internal("resolve cash/bank account", err)
		}

		var lines []accounting.JournalLineInput
		var sourceType accounting.SourceType
		if payment.PaymentType == PaymentReceipt {
			sourceType = accounting.SourcePaymentReceipt
			lines = []accounting.JournalLineInput{
				{AccountCode: account.Code, Debit: payment.Amount, Description: "Receipt " + payment.PaymentNumber},
				{AccountCode: acctAR, Credit: payment.Amount, Description: "AR settled " + payment.PaymentNumber},
			}
		} else {
			sourceType = accounting.SourcePaymentDisbursement
			lines = []accounting.JournalLineInput{
				{AccountCode: acctAP, Debit: payment.Amount, Description: "AP settled " + payment.PaymentNumber},
				{AccountCode: account.Code, Credit: payment.Amount, Description: "Disbursement " + payment.PaymentNumber},
			}
		}

		journal, err := s.accSvc.PostJournal(ctx, sourceType, &payment.ID, payment.PaymentDate, "Payment "+payment.PaymentNumber, lines)
		if err != nil {
			return apperr.Wrapf(err, "post payment journal")
		}

		if err := s.repo.SetPaymentPosted(ctx, id, journal.ID); err != nil {
			return apperr.Internal("mark payment posted", err)
		}

		var allocations []PaymentAllocation
		var billAllocations []SupplierPaymentAllocation
		if payment.PaymentType == PaymentReceipt {
			allocations, err = s.repo.ListAllocations(ctx, id)
			if err != nil {
				return apperr.Internal("load payment allocations", err)
			}
			for _, a := range allocations {
				if _, err := s.salesSvc.ApplyPaymentToInvoice(ctx, a.InvoiceID, a.AmountAllocated); err != nil {
					return apperr.Wrapf(err, "apply payment to invoice")
				}
			}
		} else {
			billAllocations, err = s.repo.ListSupplierAllocations(ctx, id)
			if err != nil {
				return apperr.Internal("load bill allocations", err)
			}
			for _, a := range billAllocations {
				if _, err := s.purchasingSvc.ApplyPaymentToBill(ctx, a.SupplierInvoiceID, a.AmountAllocated); err != nil {
					return apperr.Wrapf(err, "apply payment to supplier bill")
				}
			}
		}

		if err := audit.Log(ctx, tx, "payments", id, audit.Update, payment.Status, StatusPosted); err != nil {
			return apperr.Internal("write audit log", err)
		}

		payment.Status = StatusPosted
		payment.BillAllocations = billAllocations
		payment.JournalID = &journal.ID
		payment.Allocations = allocations
		out = payment
		return nil
	})
	return out, err
}

func (s *Service) VoidPayment(ctx context.Context, id uuid.UUID, reason string) (Payment, error) {
	var out Payment
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		payment, err := s.repo.GetPaymentForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("payment not found")
			}
			return apperr.Internal("load payment", err)
		}
		if payment.Status == StatusVoid {
			return apperr.Conflict("payment already void")
		}

		if payment.Status == StatusPosted {
			if payment.JournalID == nil {
				return apperr.Internal("posted payment missing journal reference", nil)
			}
			if _, err := s.accSvc.ReverseJournal(ctx, *payment.JournalID, reason); err != nil {
				return apperr.Wrapf(err, "reverse payment journal")
			}
			if payment.PaymentType == PaymentReceipt {
				allocations, err := s.repo.ListAllocations(ctx, id)
				if err != nil {
					return apperr.Internal("load payment allocations", err)
				}
				for _, a := range allocations {
					if _, err := s.salesSvc.ApplyPaymentToInvoice(ctx, a.InvoiceID, a.AmountAllocated.Neg()); err != nil {
						return apperr.Wrapf(err, "reverse invoice allocation")
					}
				}
			} else {
				billAllocations, err := s.repo.ListSupplierAllocations(ctx, id)
				if err != nil {
					return apperr.Internal("load bill allocations", err)
				}
				for _, a := range billAllocations {
					if _, err := s.purchasingSvc.ApplyPaymentToBill(ctx, a.SupplierInvoiceID, a.AmountAllocated.Neg()); err != nil {
						return apperr.Wrapf(err, "reverse bill allocation")
					}
				}
			}
		}

		if err := s.repo.SetPaymentVoid(ctx, id); err != nil {
			return apperr.Internal("void payment", err)
		}
		if err := audit.Log(ctx, tx, "payments", id, audit.Update, payment.Status, map[string]any{"status": StatusVoid, "reason": reason}); err != nil {
			return apperr.Internal("write audit log", err)
		}

		payment.Status = StatusVoid
		out = payment
		return nil
	})
	return out, err
}

func (s *Service) GetPayment(ctx context.Context, id uuid.UUID) (Payment, error) {
	p, err := s.repo.GetPaymentByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Payment{}, apperr.NotFound("payment not found")
	}
	if err != nil {
		return Payment{}, err
	}
	allocations, err := s.repo.ListAllocations(ctx, id)
	if err != nil {
		return Payment{}, apperr.Internal("load allocations", err)
	}
	p.Allocations = allocations
	billAllocations, err := s.repo.ListSupplierAllocations(ctx, id)
	if err != nil {
		return Payment{}, apperr.Internal("load bill allocations", err)
	}
	p.BillAllocations = billAllocations
	return p, nil
}

func (s *Service) ListPayments(ctx context.Context, limit, offset int) ([]Payment, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListPayments(ctx, limit, offset)
}

// ---- Expenses ----

func (s *Service) CreateExpense(ctx context.Context, in CreateExpenseInput) (Expense, error) {
	if in.Amount.LessThanOrEqual(decimal.Zero) {
		return Expense{}, apperr.Validation("amount must be greater than zero")
	}
	expenseDate := time.Now()
	if in.ExpenseDate != "" {
		parsed, err := time.Parse("2006-01-02", in.ExpenseDate)
		if err != nil {
			return Expense{}, apperr.Validation("expense_date must be YYYY-MM-DD")
		}
		expenseDate = parsed
	}

	var out Expense
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		expenseAcct, err := s.accountRepo.GetByCode(ctx, in.ExpenseAccountCode)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.Validation("unknown expense account code: " + in.ExpenseAccountCode)
			}
			return apperr.Internal("resolve expense account", err)
		}
		paidFromAcct, err := s.accountRepo.GetByCode(ctx, in.PaidFromAccountCode)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.Validation("unknown paid_from account code: " + in.PaidFromAccountCode)
			}
			return apperr.Internal("resolve paid_from account", err)
		}

		number, err := numbering.Generate(ctx, tx, numbering.Expense, expenseDate)
		if err != nil {
			return apperr.Internal("generate expense number", err)
		}

		expense, err := s.repo.InsertExpense(ctx, Expense{
			ExpenseNumber: number, ExpenseDate: expenseDate, ExpenseAccountID: expenseAcct.ID,
			PaidFromAccountID: paidFromAcct.ID, SupplierID: in.SupplierID, Amount: in.Amount, Description: in.Description,
		})
		if err != nil {
			return apperr.Internal("create expense", err)
		}

		if err := audit.Log(ctx, tx, "expenses", expense.ID, audit.Insert, nil, expense); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = expense
		return nil
	})
	return out, err
}

func (s *Service) PostExpense(ctx context.Context, id uuid.UUID) (Expense, error) {
	var out Expense
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		expense, err := s.repo.GetExpenseForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("expense not found")
			}
			return apperr.Internal("load expense", err)
		}
		if expense.Status != StatusDraft {
			return apperr.Conflict("only DRAFT expenses can be posted")
		}

		expenseAcct, err := s.accountRepo.GetByID(ctx, expense.ExpenseAccountID)
		if err != nil {
			return apperr.Internal("resolve expense account", err)
		}
		paidFromAcct, err := s.accountRepo.GetByID(ctx, expense.PaidFromAccountID)
		if err != nil {
			return apperr.Internal("resolve paid_from account", err)
		}

		journal, err := s.accSvc.PostJournal(ctx, accounting.SourceExpense, &expense.ID, expense.ExpenseDate,
			"Expense "+expense.ExpenseNumber+": "+expense.Description,
			[]accounting.JournalLineInput{
				{AccountCode: expenseAcct.Code, Debit: expense.Amount, Description: expense.Description},
				{AccountCode: paidFromAcct.Code, Credit: expense.Amount, Description: "Payment for " + expense.ExpenseNumber},
			})
		if err != nil {
			return apperr.Wrapf(err, "post expense journal")
		}

		if err := s.repo.SetExpensePosted(ctx, id, journal.ID); err != nil {
			return apperr.Internal("mark expense posted", err)
		}
		if err := audit.Log(ctx, tx, "expenses", id, audit.Update, expense.Status, StatusPosted); err != nil {
			return apperr.Internal("write audit log", err)
		}

		expense.Status = StatusPosted
		expense.JournalID = &journal.ID
		out = expense
		return nil
	})
	return out, err
}

func (s *Service) VoidExpense(ctx context.Context, id uuid.UUID, reason string) (Expense, error) {
	var out Expense
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		expense, err := s.repo.GetExpenseForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("expense not found")
			}
			return apperr.Internal("load expense", err)
		}
		if expense.Status == StatusVoid {
			return apperr.Conflict("expense already void")
		}
		if expense.Status == StatusPosted {
			if expense.JournalID == nil {
				return apperr.Internal("posted expense missing journal reference", nil)
			}
			if _, err := s.accSvc.ReverseJournal(ctx, *expense.JournalID, reason); err != nil {
				return apperr.Wrapf(err, "reverse expense journal")
			}
		}
		if err := s.repo.SetExpenseVoid(ctx, id); err != nil {
			return apperr.Internal("void expense", err)
		}
		if err := audit.Log(ctx, tx, "expenses", id, audit.Update, expense.Status, map[string]any{"status": StatusVoid, "reason": reason}); err != nil {
			return apperr.Internal("write audit log", err)
		}
		expense.Status = StatusVoid
		out = expense
		return nil
	})
	return out, err
}

func (s *Service) GetExpense(ctx context.Context, id uuid.UUID) (Expense, error) {
	e, err := s.repo.GetExpenseByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Expense{}, apperr.NotFound("expense not found")
	}
	return e, err
}

func (s *Service) ListExpenses(ctx context.Context, limit, offset int) ([]Expense, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListExpenses(ctx, limit, offset)
}
