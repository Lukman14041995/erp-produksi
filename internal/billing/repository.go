package billing

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/shopspring/decimal"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ---- Bank accounts ----

const bankAccountColumns = `id, bank_name, account_number, account_holder, coa_account_code, sort_order, is_active, created_at, updated_at`

func scanBankAccount(row pgx.Row) (BankAccount, error) {
	var b BankAccount
	err := row.Scan(&b.ID, &b.BankName, &b.AccountNumber, &b.AccountHolder, &b.COAAccountCode, &b.SortOrder, &b.IsActive, &b.CreatedAt, &b.UpdatedAt)
	return b, err
}

func (r *Repository) CreateBankAccount(ctx context.Context, in UpsertBankAccountInput) (BankAccount, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO bank_accounts (bank_name, account_number, account_holder, coa_account_code, sort_order)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING `+bankAccountColumns, in.BankName, in.AccountNumber, in.AccountHolder, in.COAAccountCode, in.SortOrder)
	return scanBankAccount(row)
}

func (r *Repository) UpdateBankAccount(ctx context.Context, id uuid.UUID, in UpsertBankAccountInput) (BankAccount, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE bank_accounts SET bank_name=$2, account_number=$3, account_holder=$4, coa_account_code=$5, sort_order=$6, is_active=$7
		WHERE id=$1
		RETURNING `+bankAccountColumns, id, in.BankName, in.AccountNumber, in.AccountHolder, in.COAAccountCode, in.SortOrder, isActive)
	return scanBankAccount(row)
}

func (r *Repository) GetBankAccountByID(ctx context.Context, id uuid.UUID) (BankAccount, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+bankAccountColumns+` FROM bank_accounts WHERE id=$1`, id)
	return scanBankAccount(row)
}

func (r *Repository) ListBankAccounts(ctx context.Context, activeOnly bool) ([]BankAccount, error) {
	q := `SELECT ` + bankAccountColumns + ` FROM bank_accounts`
	if activeOnly {
		q += ` WHERE is_active = true`
	}
	q += ` ORDER BY sort_order`
	rows, err := db.Q(ctx, r.pool).Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []BankAccount{}
	for rows.Next() {
		b, err := scanBankAccount(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// DefaultActiveBankAccount picks the lowest-sort-order active account, used
// to assign a destination account to newly-created installments.
func (r *Repository) DefaultActiveBankAccount(ctx context.Context) (BankAccount, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+bankAccountColumns+` FROM bank_accounts WHERE is_active=true ORDER BY sort_order LIMIT 1`)
	return scanBankAccount(row)
}

// ---- Payment plans ----

const paymentPlanColumns = `id, invoice_id, method, installment_count, created_at`

func scanPaymentPlan(row pgx.Row) (PaymentPlan, error) {
	var p PaymentPlan
	err := row.Scan(&p.ID, &p.InvoiceID, &p.Method, &p.InstallmentCount, &p.CreatedAt)
	return p, err
}

func (r *Repository) InsertPaymentPlan(ctx context.Context, invoiceID uuid.UUID, method Method, count int) (PaymentPlan, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO payment_plans (invoice_id, method, installment_count)
		VALUES ($1,$2,$3)
		RETURNING `+paymentPlanColumns, invoiceID, method, count)
	return scanPaymentPlan(row)
}

func (r *Repository) GetPaymentPlanByInvoice(ctx context.Context, invoiceID uuid.UUID) (PaymentPlan, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+paymentPlanColumns+` FROM payment_plans WHERE invoice_id=$1`, invoiceID)
	return scanPaymentPlan(row)
}

func (r *Repository) GetPaymentPlanByID(ctx context.Context, id uuid.UUID) (PaymentPlan, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+paymentPlanColumns+` FROM payment_plans WHERE id=$1`, id)
	return scanPaymentPlan(row)
}

// ---- Installments ----

const installmentColumns = `id, payment_plan_id, sequence_no, amount, bank_account_id, status, proof_image_url, reject_reason, payment_id, reviewed_by, reviewed_at, submitted_at, created_at`

func scanInstallment(row pgx.Row) (Installment, error) {
	var i Installment
	err := row.Scan(&i.ID, &i.PaymentPlanID, &i.SequenceNo, &i.Amount, &i.BankAccountID, &i.Status, &i.ProofImageURL,
		&i.RejectReason, &i.PaymentID, &i.ReviewedBy, &i.ReviewedAt, &i.SubmittedAt, &i.CreatedAt)
	return i, err
}

func (r *Repository) InsertInstallment(ctx context.Context, planID uuid.UUID, seq int, amount decimal.Decimal, bankAccountID uuid.UUID) (Installment, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO payment_installments (payment_plan_id, sequence_no, amount, bank_account_id)
		VALUES ($1,$2,$3,$4)
		RETURNING `+installmentColumns, planID, seq, amount, bankAccountID)
	return scanInstallment(row)
}

func (r *Repository) ListInstallmentsByPlan(ctx context.Context, planID uuid.UUID) ([]Installment, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+installmentColumns+` FROM payment_installments WHERE payment_plan_id=$1 ORDER BY sequence_no`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Installment{}
	for rows.Next() {
		i, err := scanInstallment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *Repository) GetInstallmentByID(ctx context.Context, id uuid.UUID) (Installment, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+installmentColumns+` FROM payment_installments WHERE id=$1`, id)
	return scanInstallment(row)
}

// GetInstallmentForUpdate row-locks the installment so concurrent
// confirm/reject/upload commands against the same installment serialize.
func (r *Repository) GetInstallmentForUpdate(ctx context.Context, id uuid.UUID) (Installment, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+installmentColumns+` FROM payment_installments WHERE id=$1 FOR UPDATE`, id)
	return scanInstallment(row)
}

func (r *Repository) SetInstallmentProof(ctx context.Context, id uuid.UUID, imageURL string) (Installment, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE payment_installments
		SET proof_image_url=$2, status='PENDING_VERIFICATION', submitted_at=now(), reject_reason=''
		WHERE id=$1
		RETURNING `+installmentColumns, id, imageURL)
	return scanInstallment(row)
}

func (r *Repository) ConfirmInstallment(ctx context.Context, id, staffUserID, paymentID uuid.UUID) (Installment, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE payment_installments
		SET status='CONFIRMED', payment_id=$3, reviewed_by=$2, reviewed_at=now()
		WHERE id=$1
		RETURNING `+installmentColumns, id, staffUserID, paymentID)
	return scanInstallment(row)
}

func (r *Repository) RejectInstallment(ctx context.Context, id, staffUserID uuid.UUID, reason string) (Installment, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE payment_installments
		SET status='REJECTED', reject_reason=$3, reviewed_by=$2, reviewed_at=now()
		WHERE id=$1
		RETURNING `+installmentColumns, id, staffUserID, reason)
	return scanInstallment(row)
}
