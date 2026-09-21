package billing

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/audit"
	"github.com/ranji/clothing-erp/internal/coa"
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/ranji/clothing-erp/internal/finance"
	"github.com/ranji/clothing-erp/internal/sales"
	"github.com/shopspring/decimal"
)

type Service struct {
	pool       *pgxpool.Pool
	repo       *Repository
	coaSvc     *coa.Service
	financeSvc *finance.Service
	salesSvc   *sales.Service
}

func NewService(pool *pgxpool.Pool, repo *Repository, coaSvc *coa.Service, financeSvc *finance.Service, salesSvc *sales.Service) *Service {
	return &Service{pool: pool, repo: repo, coaSvc: coaSvc, financeSvc: financeSvc, salesSvc: salesSvc}
}

// ---- Bank accounts (staff) ----

func (s *Service) CreateBankAccount(ctx context.Context, in UpsertBankAccountInput) (BankAccount, error) {
	if in.BankName == "" || in.AccountNumber == "" || in.AccountHolder == "" {
		return BankAccount{}, apperr.Validation("bank_name, account_number, and account_holder are required")
	}
	if _, err := s.coaSvc.GetByCode(ctx, in.COAAccountCode); err != nil {
		return BankAccount{}, apperr.Validation("unknown coa_account_code: " + in.COAAccountCode)
	}
	var out BankAccount
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		b, err := s.repo.CreateBankAccount(ctx, in)
		if err != nil {
			return apperr.Internal("create bank account", err)
		}
		if err := audit.Log(ctx, tx, "bank_accounts", b.ID, audit.Insert, nil, b); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = b
		return nil
	})
	return out, err
}

func (s *Service) UpdateBankAccount(ctx context.Context, id uuid.UUID, in UpsertBankAccountInput) (BankAccount, error) {
	if _, err := s.coaSvc.GetByCode(ctx, in.COAAccountCode); err != nil {
		return BankAccount{}, apperr.Validation("unknown coa_account_code: " + in.COAAccountCode)
	}
	var out BankAccount
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		before, err := s.repo.GetBankAccountByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("bank account not found")
			}
			return apperr.Internal("load bank account", err)
		}
		after, err := s.repo.UpdateBankAccount(ctx, id, in)
		if err != nil {
			return apperr.Internal("update bank account", err)
		}
		if err := audit.Log(ctx, tx, "bank_accounts", id, audit.Update, before, after); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = after
		return nil
	})
	return out, err
}

func (s *Service) ListBankAccounts(ctx context.Context, activeOnly bool) ([]BankAccount, error) {
	return s.repo.ListBankAccounts(ctx, activeOnly)
}

// ---- Payment plan (customer chooses Lunas/Termin) ----

// ChoosePaymentPlan splits the invoice's grand_total into N equal
// installments (remainder folded into the last one so the sum always
// matches exactly), each assigned the same default active bank account.
// Called once per invoice -- a plan can't be re-chosen after creation.
func (s *Service) ChoosePaymentPlan(ctx context.Context, invoiceID uuid.UUID, in ChoosePaymentPlanInput) (PaymentPlan, error) {
	count := in.InstallmentCount
	if in.Method == MethodFull {
		count = 1
	}
	if count < 1 || count > 3 {
		return PaymentPlan{}, apperr.Validation("installment_count must be between 1 and 3")
	}
	if in.Method != MethodFull && in.Method != MethodInstallment {
		return PaymentPlan{}, apperr.Validation("method must be FULL or INSTALLMENT")
	}

	var out PaymentPlan
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := s.repo.GetPaymentPlanByInvoice(ctx, invoiceID); err == nil {
			return apperr.Conflict("this invoice already has a payment plan")
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return apperr.Internal("check existing payment plan", err)
		}

		invoice, err := s.salesSvc.GetInvoice(ctx, invoiceID)
		if err != nil {
			return err
		}

		bank, err := s.repo.DefaultActiveBankAccount(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.Conflict("no active bank account is configured yet")
			}
			return apperr.Internal("load default bank account", err)
		}

		plan, err := s.repo.InsertPaymentPlan(ctx, invoiceID, in.Method, count)
		if err != nil {
			return apperr.Internal("create payment plan", err)
		}

		countDec := decimal.NewFromInt(int64(count))
		base := invoice.GrandTotal.Div(countDec).Round(0)
		allocated := decimal.Zero
		for seq := 1; seq <= count; seq++ {
			amount := base
			if seq == count {
				amount = invoice.GrandTotal.Sub(allocated)
			}
			allocated = allocated.Add(amount)
			inst, err := s.repo.InsertInstallment(ctx, plan.ID, seq, amount, bank.ID)
			if err != nil {
				return apperr.Internal("create installment", err)
			}
			plan.Installments = append(plan.Installments, inst)
		}

		if err := audit.Log(ctx, tx, "payment_plans", plan.ID, audit.Insert, nil, plan); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = plan
		return nil
	})
	return out, err
}

func (s *Service) GetPlanByID(ctx context.Context, id uuid.UUID) (PaymentPlan, error) {
	plan, err := s.repo.GetPaymentPlanByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return PaymentPlan{}, apperr.NotFound("payment plan not found")
	}
	return plan, err
}

func (s *Service) GetInstallment(ctx context.Context, id uuid.UUID) (Installment, error) {
	inst, err := s.repo.GetInstallmentByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Installment{}, apperr.NotFound("installment not found")
	}
	return inst, err
}

func (s *Service) GetPaymentPlanByInvoice(ctx context.Context, invoiceID uuid.UUID) (*PaymentPlan, error) {
	plan, err := s.repo.GetPaymentPlanByInvoice(ctx, invoiceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, apperr.Internal("load payment plan", err)
	}
	installments, err := s.repo.ListInstallmentsByPlan(ctx, plan.ID)
	if err != nil {
		return nil, apperr.Internal("load installments", err)
	}
	banks, err := s.repo.ListBankAccounts(ctx, false)
	if err != nil {
		return nil, apperr.Internal("load bank accounts", err)
	}
	banksByID := map[uuid.UUID]BankAccount{}
	for _, b := range banks {
		banksByID[b.ID] = b
	}
	for i := range installments {
		if b, ok := banksByID[installments[i].BankAccountID]; ok {
			installments[i].BankAccount = &b
		}
	}
	plan.Installments = installments
	return &plan, nil
}

// ---- Installment proof + verification ----

func (s *Service) UploadInstallmentProof(ctx context.Context, id uuid.UUID, imageURL string) (Installment, error) {
	var out Installment
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		inst, err := s.repo.GetInstallmentForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("installment not found")
			}
			return apperr.Internal("load installment", err)
		}
		if inst.Status != InstallmentPendingPayment && inst.Status != InstallmentRejected {
			return apperr.Conflict("this installment does not need a new proof of payment")
		}
		updated, err := s.repo.SetInstallmentProof(ctx, id, imageURL)
		if err != nil {
			return apperr.Internal("save proof of payment", err)
		}
		if err := audit.Log(ctx, tx, "payment_installments", id, audit.Update, inst.Status, updated.Status); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = updated
		return nil
	})
	return out, err
}

// ConfirmInstallment is the staff verification step: it books a real
// finance.Payment (RECEIPT, allocated to the invoice) so the confirmed
// transfer flows straight into accounting -- same journal-posting and
// AR-allocation path a manually-entered payment would take.
func (s *Service) ConfirmInstallment(ctx context.Context, id, staffUserID uuid.UUID) (Installment, error) {
	var out Installment
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		inst, err := s.repo.GetInstallmentForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("installment not found")
			}
			return apperr.Internal("load installment", err)
		}
		if inst.Status != InstallmentPendingVerification {
			return apperr.Conflict("only installments pending verification can be confirmed")
		}
		bank, err := s.repo.GetBankAccountByID(ctx, inst.BankAccountID)
		if err != nil {
			return apperr.Internal("load bank account", err)
		}

		// Walk back from the installment to its invoice/customer via the plan.
		planRow, err := s.repo.GetPaymentPlanByID(ctx, inst.PaymentPlanID)
		if err != nil {
			return apperr.Internal("load payment plan", err)
		}
		invoice, err := s.salesSvc.GetInvoice(ctx, planRow.InvoiceID)
		if err != nil {
			return err
		}

		payment, err := s.financeSvc.CreatePayment(ctx, finance.CreatePaymentInput{
			PaymentType: finance.PaymentReceipt, CustomerID: &invoice.CustomerID,
			CashBankAccountCode: bank.COAAccountCode, Amount: inst.Amount, Method: "TRANSFER",
			ReferenceNo: "Installment " + invoice.InvoiceNumber,
			Allocations: []finance.AllocationInput{{InvoiceID: invoice.ID, Amount: inst.Amount}},
		})
		if err != nil {
			return apperr.Wrapf(err, "record payment for installment")
		}
		if _, err := s.financeSvc.PostPayment(ctx, payment.ID); err != nil {
			return apperr.Wrapf(err, "post payment for installment")
		}

		updated, err := s.repo.ConfirmInstallment(ctx, id, staffUserID, payment.ID)
		if err != nil {
			return apperr.Internal("confirm installment", err)
		}
		if err := audit.Log(ctx, tx, "payment_installments", id, audit.Update, inst.Status, updated.Status); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = updated
		return nil
	})
	return out, err
}

func (s *Service) RejectInstallment(ctx context.Context, id, staffUserID uuid.UUID, in RejectInstallmentInput) (Installment, error) {
	var out Installment
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		inst, err := s.repo.GetInstallmentForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("installment not found")
			}
			return apperr.Internal("load installment", err)
		}
		if inst.Status != InstallmentPendingVerification {
			return apperr.Conflict("only installments pending verification can be rejected")
		}
		updated, err := s.repo.RejectInstallment(ctx, id, staffUserID, in.Reason)
		if err != nil {
			return apperr.Internal("reject installment", err)
		}
		if err := audit.Log(ctx, tx, "payment_installments", id, audit.Update, inst.Status, map[string]any{"status": InstallmentRejected, "reason": in.Reason}); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = updated
		return nil
	})
	return out, err
}
