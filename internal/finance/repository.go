package finance

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/db"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const paymentColumns = `id, payment_number, payment_type, customer_id, supplier_id, cash_bank_account_id, payment_date, amount, method, reference_no, status, journal_id, notes, created_at`

func scanPayment(row pgx.Row) (Payment, error) {
	var p Payment
	err := row.Scan(&p.ID, &p.PaymentNumber, &p.PaymentType, &p.CustomerID, &p.SupplierID, &p.CashBankAccountID, &p.PaymentDate,
		&p.Amount, &p.Method, &p.ReferenceNo, &p.Status, &p.JournalID, &p.Notes, &p.CreatedAt)
	return p, err
}

func (r *Repository) InsertPayment(ctx context.Context, p Payment) (Payment, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO payments (payment_number, payment_type, customer_id, supplier_id, cash_bank_account_id, payment_date, amount, method, reference_no, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING `+paymentColumns, p.PaymentNumber, p.PaymentType, p.CustomerID, p.SupplierID, p.CashBankAccountID, p.PaymentDate, p.Amount, p.Method, p.ReferenceNo, p.Notes)
	return scanPayment(row)
}

func (r *Repository) GetPaymentForUpdate(ctx context.Context, id uuid.UUID) (Payment, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+paymentColumns+` FROM payments WHERE id=$1 FOR UPDATE`, id)
	return scanPayment(row)
}

func (r *Repository) GetPaymentByID(ctx context.Context, id uuid.UUID) (Payment, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+paymentColumns+` FROM payments WHERE id=$1`, id)
	return scanPayment(row)
}

func (r *Repository) ListPayments(ctx context.Context, limit, offset int) ([]Payment, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+paymentColumns+` FROM payments ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Payment{}
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) SetPaymentPosted(ctx context.Context, id, journalID uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		UPDATE payments SET status='POSTED', journal_id=$2, posted_at=now() WHERE id=$1
	`, id, journalID)
	return err
}

func (r *Repository) SetPaymentVoid(ctx context.Context, id uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE payments SET status='VOID', voided_at=now() WHERE id=$1`, id)
	return err
}

func (r *Repository) InsertAllocation(ctx context.Context, a PaymentAllocation) (PaymentAllocation, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO payment_allocations (payment_id, invoice_id, amount_allocated)
		VALUES ($1,$2,$3)
		RETURNING id, payment_id, invoice_id, amount_allocated
	`, a.PaymentID, a.InvoiceID, a.AmountAllocated)
	var out PaymentAllocation
	err := row.Scan(&out.ID, &out.PaymentID, &out.InvoiceID, &out.AmountAllocated)
	return out, err
}

func (r *Repository) ListAllocations(ctx context.Context, paymentID uuid.UUID) ([]PaymentAllocation, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT id, payment_id, invoice_id, amount_allocated FROM payment_allocations WHERE payment_id=$1
	`, paymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PaymentAllocation{}
	for rows.Next() {
		var a PaymentAllocation
		if err := rows.Scan(&a.ID, &a.PaymentID, &a.InvoiceID, &a.AmountAllocated); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *Repository) InsertSupplierAllocation(ctx context.Context, a SupplierPaymentAllocation) (SupplierPaymentAllocation, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO supplier_payment_allocations (payment_id, supplier_invoice_id, amount_allocated)
		VALUES ($1,$2,$3)
		RETURNING id, payment_id, supplier_invoice_id, amount_allocated
	`, a.PaymentID, a.SupplierInvoiceID, a.AmountAllocated)
	var out SupplierPaymentAllocation
	err := row.Scan(&out.ID, &out.PaymentID, &out.SupplierInvoiceID, &out.AmountAllocated)
	return out, err
}

func (r *Repository) ListSupplierAllocations(ctx context.Context, paymentID uuid.UUID) ([]SupplierPaymentAllocation, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `
		SELECT id, payment_id, supplier_invoice_id, amount_allocated FROM supplier_payment_allocations WHERE payment_id=$1
	`, paymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SupplierPaymentAllocation{}
	for rows.Next() {
		var a SupplierPaymentAllocation
		if err := rows.Scan(&a.ID, &a.PaymentID, &a.SupplierInvoiceID, &a.AmountAllocated); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ---- Expenses ----

const expenseColumns = `id, expense_number, expense_date, expense_account_id, paid_from_account_id, supplier_id, amount, description, status, journal_id, created_at`

func scanExpense(row pgx.Row) (Expense, error) {
	var e Expense
	err := row.Scan(&e.ID, &e.ExpenseNumber, &e.ExpenseDate, &e.ExpenseAccountID, &e.PaidFromAccountID, &e.SupplierID, &e.Amount, &e.Description, &e.Status, &e.JournalID, &e.CreatedAt)
	return e, err
}

func (r *Repository) InsertExpense(ctx context.Context, e Expense) (Expense, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO expenses (expense_number, expense_date, expense_account_id, paid_from_account_id, supplier_id, amount, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING `+expenseColumns, e.ExpenseNumber, e.ExpenseDate, e.ExpenseAccountID, e.PaidFromAccountID, e.SupplierID, e.Amount, e.Description)
	return scanExpense(row)
}

func (r *Repository) GetExpenseForUpdate(ctx context.Context, id uuid.UUID) (Expense, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+expenseColumns+` FROM expenses WHERE id=$1 FOR UPDATE`, id)
	return scanExpense(row)
}

func (r *Repository) GetExpenseByID(ctx context.Context, id uuid.UUID) (Expense, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+expenseColumns+` FROM expenses WHERE id=$1`, id)
	return scanExpense(row)
}

func (r *Repository) ListExpenses(ctx context.Context, limit, offset int) ([]Expense, error) {
	rows, err := db.Q(ctx, r.pool).Query(ctx, `SELECT `+expenseColumns+` FROM expenses ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Expense{}
	for rows.Next() {
		e, err := scanExpense(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *Repository) SetExpensePosted(ctx context.Context, id, journalID uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `
		UPDATE expenses SET status='POSTED', journal_id=$2, posted_at=now() WHERE id=$1
	`, id, journalID)
	return err
}

func (r *Repository) SetExpenseVoid(ctx context.Context, id uuid.UUID) error {
	_, err := db.Q(ctx, r.pool).Exec(ctx, `UPDATE expenses SET status='VOID', voided_at=now() WHERE id=$1`, id)
	return err
}
