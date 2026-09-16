package sales

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
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/ranji/clothing-erp/internal/inventory"
	"github.com/ranji/clothing-erp/internal/numbering"
	"github.com/shopspring/decimal"
)

const (
	acctAR            = "1-1100" // Piutang Usaha
	acctRevenue       = "4-1000" // Penjualan
	acctTaxPayable    = "2-2000" // Hutang Pajak
	acctCOGS          = "5-1000" // Harga Pokok Penjualan
	acctFinishedGoods = "1-1220" // Persediaan Barang Jadi
)

type Service struct {
	pool   *pgxpool.Pool
	repo   *Repository
	accSvc *accounting.Service
	invSvc *inventory.Service
}

func NewService(pool *pgxpool.Pool, repo *Repository, accSvc *accounting.Service, invSvc *inventory.Service) *Service {
	return &Service{pool: pool, repo: repo, accSvc: accSvc, invSvc: invSvc}
}

func computeLine(qty, unitPrice, discount, taxRate decimal.Decimal) (lineTotal, taxAmount decimal.Decimal) {
	afterDiscount := qty.Mul(unitPrice).Sub(discount)
	taxAmount = afterDiscount.Mul(taxRate)
	lineTotal = afterDiscount.Add(taxAmount)
	return lineTotal, taxAmount
}

func (s *Service) CreateOrder(ctx context.Context, in CreateOrderInput) (SalesOrder, error) {
	if in.CustomerID == uuid.Nil {
		return SalesOrder{}, apperr.Validation("customer_id is required")
	}
	if len(in.Items) == 0 {
		return SalesOrder{}, apperr.Validation("order must have at least one item")
	}
	orderDate := time.Now()
	if in.OrderDate != "" {
		parsed, err := time.Parse("2006-01-02", in.OrderDate)
		if err != nil {
			return SalesOrder{}, apperr.Validation("order_date must be YYYY-MM-DD")
		}
		orderDate = parsed
	}

	subtotal, discountTotal, taxTotal, grandTotal := decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero
	for _, it := range in.Items {
		if it.Qty.LessThanOrEqual(decimal.Zero) {
			return SalesOrder{}, apperr.Validation("item qty must be greater than zero")
		}
		lineTotal, taxAmount := computeLine(it.Qty, it.UnitPrice, it.Discount, it.TaxRate)
		subtotal = subtotal.Add(it.Qty.Mul(it.UnitPrice))
		discountTotal = discountTotal.Add(it.Discount)
		taxTotal = taxTotal.Add(taxAmount)
		grandTotal = grandTotal.Add(lineTotal)
	}

	var out SalesOrder
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		number, err := numbering.Generate(ctx, tx, numbering.SalesOrder, orderDate)
		if err != nil {
			return apperr.Internal("generate SO number", err)
		}

		order, err := s.repo.InsertOrder(ctx, SalesOrder{
			SONumber: number, CustomerID: in.CustomerID, OrderDate: orderDate, Notes: in.Notes,
			Subtotal: subtotal, DiscountTotal: discountTotal, TaxTotal: taxTotal, GrandTotal: grandTotal,
			CreatedBy: db.ActorFromContext(ctx),
		})
		if err != nil {
			return apperr.Internal("create sales order", err)
		}

		for _, it := range in.Items {
			lineTotal, _ := computeLine(it.Qty, it.UnitPrice, it.Discount, it.TaxRate)
			item, err := s.repo.InsertOrderItem(ctx, SalesOrderItem{
				SalesOrderID: order.ID, ProductID: it.ProductID, ProductSizeID: it.ProductSizeID,
				Qty: it.Qty, UnitPrice: it.UnitPrice, Discount: it.Discount, TaxRate: it.TaxRate, LineTotal: lineTotal,
			})
			if err != nil {
				return apperr.Internal("create sales order item", err)
			}
			order.Items = append(order.Items, item)
		}

		if err := audit.Log(ctx, tx, "sales_orders", order.ID, audit.Insert, nil, order); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = order
		return nil
	})
	return out, err
}

// ConfirmOrder is the command endpoint for DRAFT -> CONFIRMED. Generic bulk
// status PUTs are intentionally not supported; every transition is its own
// explicit, validated operation.
func (s *Service) ConfirmOrder(ctx context.Context, id uuid.UUID) (SalesOrder, error) {
	var out SalesOrder
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		order, err := s.repo.GetOrderForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("sales order not found")
			}
			return apperr.Internal("load sales order", err)
		}
		if order.OrderStatus != OrderDraft {
			return apperr.Conflict("only DRAFT orders can be confirmed")
		}

		if err := s.repo.SetOrderStatus(ctx, id, OrderConfirmed); err != nil {
			return apperr.Internal("confirm order", err)
		}
		if err := s.repo.SetConfirmedAt(ctx, id); err != nil {
			return apperr.Internal("stamp confirmed_at", err)
		}
		if err := audit.Log(ctx, tx, "sales_orders", id, audit.Update, order.OrderStatus, OrderConfirmed); err != nil {
			return apperr.Internal("write audit log", err)
		}

		order.OrderStatus = OrderConfirmed
		out = order
		return nil
	})
	return out, err
}

func (s *Service) CancelOrder(ctx context.Context, id uuid.UUID, reason string) (SalesOrder, error) {
	var out SalesOrder
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		order, err := s.repo.GetOrderForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("sales order not found")
			}
			return apperr.Internal("load sales order", err)
		}
		if order.OrderStatus == OrderCancelled || order.OrderStatus == OrderClosed {
			return apperr.Conflict("order is already " + string(order.OrderStatus))
		}
		if order.ProductionStatus != ProductionNotStarted {
			return apperr.Conflict("cannot cancel an order once production has started")
		}

		invoices, err := s.repo.ListInvoicesByOrder(ctx, id)
		if err != nil {
			return apperr.Internal("check existing invoices", err)
		}
		for _, inv := range invoices {
			if inv.Status != InvoiceVoid {
				return apperr.Conflict("cannot cancel an order that has already been invoiced")
			}
		}

		if err := s.repo.SetOrderStatus(ctx, id, OrderCancelled); err != nil {
			return apperr.Internal("cancel order", err)
		}
		if err := s.repo.SetCancelledAt(ctx, id); err != nil {
			return apperr.Internal("stamp cancelled_at", err)
		}
		if err := audit.Log(ctx, tx, "sales_orders", id, audit.Update, order.OrderStatus, map[string]any{"order_status": OrderCancelled, "reason": reason}); err != nil {
			return apperr.Internal("write audit log", err)
		}

		order.OrderStatus = OrderCancelled
		out = order
		return nil
	})
	return out, err
}

// UpdateProductionStatus lets the production domain keep the sales order's
// production_status in sync as its own production order progresses, without
// production ever writing to the sales_orders table directly.
func (s *Service) UpdateProductionStatus(ctx context.Context, orderID uuid.UUID, status ProductionStatus) error {
	return db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		order, err := s.repo.GetOrderForUpdate(ctx, orderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("sales order not found")
			}
			return apperr.Internal("load sales order", err)
		}
		if err := s.repo.SetProductionStatus(ctx, orderID, status); err != nil {
			return apperr.Internal("update production status", err)
		}
		return audit.Log(ctx, tx, "sales_orders", orderID, audit.Update, order.ProductionStatus, status)
	})
}

// CreateInvoiceForOrder implements the transaction-driven accounting law:
// posting an invoice automatically posts DR Piutang Usaha / CR Penjualan
// (+ CR Hutang Pajak for the tax portion) in the same transaction.
func (s *Service) CreateInvoiceForOrder(ctx context.Context, orderID uuid.UUID) (Invoice, error) {
	var out Invoice
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		order, err := s.repo.GetOrderForUpdate(ctx, orderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("sales order not found")
			}
			return apperr.Internal("load sales order", err)
		}
		if order.OrderStatus != OrderConfirmed {
			return apperr.Conflict("order must be CONFIRMED before it can be invoiced")
		}

		existing, err := s.repo.ListInvoicesByOrder(ctx, orderID)
		if err != nil {
			return apperr.Internal("check existing invoices", err)
		}
		for _, inv := range existing {
			if inv.Status != InvoiceVoid {
				return apperr.Conflict("order has already been invoiced")
			}
		}

		items, err := s.repo.ListOrderItems(ctx, orderID)
		if err != nil {
			return apperr.Internal("load order items", err)
		}

		invoiceDate := time.Now()
		number, err := numbering.Generate(ctx, tx, numbering.Invoice, invoiceDate)
		if err != nil {
			return apperr.Internal("generate invoice number", err)
		}

		invoice, err := s.repo.InsertInvoice(ctx, Invoice{
			InvoiceNumber: number, SalesOrderID: orderID, CustomerID: order.CustomerID, InvoiceDate: invoiceDate,
			Status: InvoicePosted, Subtotal: order.Subtotal, DiscountTotal: order.DiscountTotal,
			TaxTotal: order.TaxTotal, GrandTotal: order.GrandTotal, PaidAmount: decimal.Zero, BalanceDue: order.GrandTotal,
		})
		if err != nil {
			return apperr.Internal("create invoice", err)
		}

		for _, it := range items {
			if err := s.repo.InsertInvoiceItem(ctx, InvoiceItem{
				InvoiceID: invoice.ID, SalesOrderItemID: it.ID, ProductID: it.ProductID, ProductSizeID: it.ProductSizeID,
				Qty: it.Qty, UnitPrice: it.UnitPrice, Discount: it.Discount, TaxRate: it.TaxRate, LineTotal: it.LineTotal,
			}); err != nil {
				return apperr.Internal("create invoice item", err)
			}
		}

		lines := []accounting.JournalLineInput{
			{AccountCode: acctAR, Debit: invoice.GrandTotal, Description: "AR " + invoice.InvoiceNumber},
			{AccountCode: acctRevenue, Credit: invoice.Subtotal.Sub(invoice.DiscountTotal), Description: "Sales " + invoice.InvoiceNumber},
		}
		if invoice.TaxTotal.IsPositive() {
			lines = append(lines, accounting.JournalLineInput{AccountCode: acctTaxPayable, Credit: invoice.TaxTotal, Description: "Tax on " + invoice.InvoiceNumber})
		}

		journal, err := s.accSvc.PostJournal(ctx, accounting.SourceSalesInvoice, &invoice.ID, invoiceDate, "Sales invoice "+invoice.InvoiceNumber+" for "+order.SONumber, lines)
		if err != nil {
			return apperr.Wrapf(err, "post invoice journal")
		}
		if err := s.repo.SetInvoiceJournal(ctx, invoice.ID, journal.ID); err != nil {
			return apperr.Internal("link invoice journal", err)
		}
		invoice.JournalID = &journal.ID

		if err := audit.Log(ctx, tx, "invoices", invoice.ID, audit.Insert, nil, invoice); err != nil {
			return apperr.Internal("write audit log", err)
		}

		out = invoice
		return nil
	})
	return out, err
}

// DeliverOrder issues finished goods out of inventory at their current
// moving-average cost and posts DR HPP / CR Persediaan Barang Jadi for the
// total COGS, matching the "COGS / Goods Delivery" accounting law.
func (s *Service) DeliverOrder(ctx context.Context, orderID uuid.UUID) (SalesOrder, error) {
	var out SalesOrder
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		order, err := s.repo.GetOrderForUpdate(ctx, orderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("sales order not found")
			}
			return apperr.Internal("load sales order", err)
		}
		if order.OrderStatus == OrderCancelled {
			return apperr.Conflict("cannot deliver a cancelled order")
		}
		if order.DeliveryStatus == DeliveryDelivered {
			return apperr.Conflict("order has already been delivered")
		}

		items, err := s.repo.ListOrderItems(ctx, orderID)
		if err != nil {
			return apperr.Internal("load order items", err)
		}

		fgWarehouse, err := s.invSvc.DefaultFinishedGoodsWarehouse(ctx)
		if err != nil {
			return apperr.Wrapf(err, "resolve finished goods warehouse")
		}

		totalCOGS := decimal.Zero
		deliveryDate := time.Now()
		for _, it := range items {
			productID := it.ProductID
			sizeID := it.ProductSizeID
			result, err := s.invSvc.RecordMovement(ctx, inventory.MovementInput{
				TxnType: inventory.TxnSale, ItemType: inventory.ItemProduct, WarehouseID: fgWarehouse.ID,
				ProductID: &productID, ProductSizeID: &sizeID, QtyOut: it.Qty,
				RefType: "SALES_ORDER", RefID: &orderID, TxnDate: deliveryDate,
				Notes: "Delivery for " + order.SONumber,
			})
			if err != nil {
				return apperr.Wrapf(err, "issue finished goods for delivery")
			}
			totalCOGS = totalCOGS.Add(result.Transaction.TotalCost)
		}

		if totalCOGS.IsPositive() {
			_, err = s.accSvc.PostJournal(ctx, accounting.SourceCOGS, &orderID, deliveryDate,
				"COGS for delivery of "+order.SONumber,
				[]accounting.JournalLineInput{
					{AccountCode: acctCOGS, Debit: totalCOGS, Description: "COGS " + order.SONumber},
					{AccountCode: acctFinishedGoods, Credit: totalCOGS, Description: "FG issued " + order.SONumber},
				})
			if err != nil {
				return apperr.Wrapf(err, "post COGS journal")
			}
		}

		if err := s.repo.SetDeliveryStatus(ctx, orderID, DeliveryDelivered); err != nil {
			return apperr.Internal("update delivery status", err)
		}
		if err := audit.Log(ctx, tx, "sales_orders", orderID, audit.Update, order.DeliveryStatus, DeliveryDelivered); err != nil {
			return apperr.Internal("write audit log", err)
		}

		order.DeliveryStatus = DeliveryDelivered
		out = order
		return nil
	})
	return out, err
}

// ApplyPaymentToInvoice is called by the finance domain, inside its own
// transaction, after it has posted the payment's journal and allocation
// record. It never posts accounting entries itself -- that stays finance's
// responsibility -- it only updates the invoice/order status columns.
func (s *Service) ApplyPaymentToInvoice(ctx context.Context, invoiceID uuid.UUID, amount decimal.Decimal) (Invoice, error) {
	invoice, err := s.repo.GetInvoiceForUpdate(ctx, invoiceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Invoice{}, apperr.NotFound("invoice not found")
		}
		return Invoice{}, apperr.Internal("load invoice", err)
	}
	if invoice.Status == InvoiceVoid {
		return Invoice{}, apperr.Conflict("cannot pay a voided invoice")
	}

	newPaid := invoice.PaidAmount.Add(amount)
	newBalance := invoice.GrandTotal.Sub(newPaid)

	status := InvoicePosted
	switch {
	case newBalance.LessThanOrEqual(decimal.Zero):
		status = InvoicePaid
	case newPaid.IsPositive():
		status = InvoicePartiallyPaid
	}

	if err := s.repo.UpdateInvoicePayment(ctx, invoiceID, newPaid, newBalance, status); err != nil {
		return Invoice{}, apperr.Internal("update invoice payment", err)
	}

	if err := s.recomputeOrderPaymentStatus(ctx, invoice.SalesOrderID); err != nil {
		return Invoice{}, err
	}

	invoice.PaidAmount = newPaid
	invoice.BalanceDue = newBalance
	invoice.Status = status
	return invoice, nil
}

func (s *Service) recomputeOrderPaymentStatus(ctx context.Context, orderID uuid.UUID) error {
	invoices, err := s.repo.ListInvoicesByOrder(ctx, orderID)
	if err != nil {
		return apperr.Internal("load order invoices", err)
	}

	totalGrand, totalPaid := decimal.Zero, decimal.Zero
	for _, inv := range invoices {
		if inv.Status == InvoiceVoid {
			continue
		}
		totalGrand = totalGrand.Add(inv.GrandTotal)
		totalPaid = totalPaid.Add(inv.PaidAmount)
	}

	status := PaymentUnpaid
	switch {
	case totalPaid.IsZero():
		status = PaymentUnpaid
	case totalPaid.GreaterThan(totalGrand):
		status = PaymentOverpaid
	case totalPaid.Equal(totalGrand):
		status = PaymentPaid
	default:
		status = PaymentPartial
	}

	if err := s.repo.SetPaymentStatus(ctx, orderID, status); err != nil {
		return apperr.Internal("update order payment status", err)
	}
	return nil
}

func (s *Service) GetOrder(ctx context.Context, id uuid.UUID) (SalesOrder, error) {
	order, err := s.repo.GetOrderByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return SalesOrder{}, apperr.NotFound("sales order not found")
	}
	if err != nil {
		return SalesOrder{}, err
	}
	items, err := s.repo.ListOrderItems(ctx, id)
	if err != nil {
		return SalesOrder{}, apperr.Internal("load order items", err)
	}
	order.Items = items
	return order, nil
}

func (s *Service) ListOrders(ctx context.Context, limit, offset int) ([]SalesOrder, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListOrders(ctx, limit, offset)
}

func (s *Service) GetInvoice(ctx context.Context, id uuid.UUID) (Invoice, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Invoice{}, apperr.NotFound("invoice not found")
	}
	if err != nil {
		return Invoice{}, err
	}
	items, err := s.repo.ListInvoiceItems(ctx, id)
	if err != nil {
		return Invoice{}, apperr.Internal("load invoice items", err)
	}
	inv.Items = items
	return inv, nil
}

func (s *Service) ListInvoices(ctx context.Context, limit, offset int) ([]Invoice, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListInvoices(ctx, limit, offset)
}
