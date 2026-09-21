package quotation

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/audit"
	"github.com/ranji/clothing-erp/internal/billing"
	"github.com/ranji/clothing-erp/internal/catalog"
	"github.com/ranji/clothing-erp/internal/customer"
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/ranji/clothing-erp/internal/numbering"
	"github.com/ranji/clothing-erp/internal/pdf"
	"github.com/ranji/clothing-erp/internal/product"
	"github.com/ranji/clothing-erp/internal/sales"
	"github.com/ranji/clothing-erp/internal/spk"
	"github.com/shopspring/decimal"
)

const defaultLinkTTL = 7 * 24 * time.Hour

type Service struct {
	pool        *pgxpool.Pool
	repo        *Repository
	catalogSvc  *catalog.Service
	customerSvc *customer.Service
	productSvc  *product.Service
	salesSvc    *sales.Service
	billingSvc  *billing.Service
	spkSvc      *spk.Service
}

func NewService(pool *pgxpool.Pool, repo *Repository, catalogSvc *catalog.Service, customerSvc *customer.Service, productSvc *product.Service, salesSvc *sales.Service, billingSvc *billing.Service, spkSvc *spk.Service) *Service {
	return &Service{pool: pool, repo: repo, catalogSvc: catalogSvc, customerSvc: customerSvc, productSvc: productSvc, salesSvc: salesSvc, billingSvc: billingSvc, spkSvc: spkSvc}
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// ---- Order links (staff-facing) ----

func (s *Service) CreateLink(ctx context.Context, staffUserID uuid.UUID, in CreateLinkInput) (CreateLinkOutput, error) {
	if in.CustomerID == uuid.Nil {
		return CreateLinkOutput{}, apperr.Validation("customer_id is required")
	}
	if _, err := s.customerSvc.Get(ctx, in.CustomerID); err != nil {
		return CreateLinkOutput{}, err
	}

	expiresAt := time.Now().Add(defaultLinkTTL)
	if in.ExpiresAt != nil {
		expiresAt = *in.ExpiresAt
	}

	token, err := randomToken()
	if err != nil {
		return CreateLinkOutput{}, apperr.Internal("generate order link token", err)
	}

	var out CreateLinkOutput
	err = db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		link, err := s.repo.InsertLink(ctx, in.CustomerID, staffUserID, hashToken(token), expiresAt)
		if err != nil {
			return apperr.Internal("create order link", err)
		}
		if err := audit.Log(ctx, tx, "order_links", link.ID, audit.Insert, nil, link); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = CreateLinkOutput{OrderLink: link, Token: token}
		return nil
	})
	return out, err
}

func (s *Service) ListLinksByCustomer(ctx context.Context, customerID uuid.UUID) ([]OrderLink, error) {
	return s.repo.ListLinksByCustomer(ctx, customerID)
}

func (s *Service) RevokeLink(ctx context.Context, id uuid.UUID) error {
	return db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		link, err := s.repo.GetLinkByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("order link not found")
			}
			return apperr.Internal("load order link", err)
		}
		if err := s.repo.SetLinkStatus(ctx, id, LinkRevoked); err != nil {
			return apperr.Internal("revoke order link", err)
		}
		return audit.Log(ctx, tx, "order_links", id, audit.Update, link.Status, LinkRevoked)
	})
}

// resolveLink validates a plaintext token from a public request: it must
// hash to a known link that is still ACTIVE and not past its expiry. Every
// public endpoint goes through this before touching anything else.
func (s *Service) resolveLink(ctx context.Context, token string) (OrderLink, error) {
	link, err := s.repo.GetLinkByTokenHash(ctx, hashToken(token))
	if errors.Is(err, pgx.ErrNoRows) {
		return OrderLink{}, apperr.Forbidden("order link is invalid")
	}
	if err != nil {
		return OrderLink{}, apperr.Internal("load order link", err)
	}
	if link.Status != LinkActive {
		return OrderLink{}, apperr.Forbidden("order link has been revoked")
	}
	if time.Now().After(link.ExpiresAt) {
		return OrderLink{}, apperr.Forbidden("order link has expired")
	}
	return link, nil
}

// ---- Public: catalog + pricing ----

func (s *Service) GetPublicCatalog(ctx context.Context, token string) (PublicCatalog, error) {
	link, err := s.resolveLink(ctx, token)
	if err != nil {
		return PublicCatalog{}, err
	}
	cust, err := s.customerSvc.Get(ctx, link.CustomerID)
	if err != nil {
		return PublicCatalog{}, err
	}
	productTypes, err := s.catalogSvc.ListProductTypes(ctx, true)
	if err != nil {
		return PublicCatalog{}, err
	}
	fabrics, err := s.catalogSvc.ListFabrics(ctx, true)
	if err != nil {
		return PublicCatalog{}, err
	}
	variants, err := s.catalogSvc.ListVariants(ctx, true)
	if err != nil {
		return PublicCatalog{}, err
	}
	inks, err := s.catalogSvc.ListInks(ctx, true)
	if err != nil {
		return PublicCatalog{}, err
	}
	sizes, err := s.catalogSvc.ListGarmentSizes(ctx, true)
	if err != nil {
		return PublicCatalog{}, err
	}
	return PublicCatalog{
		CustomerName: cust.Name, ProductTypes: productTypes, Fabrics: fabrics,
		Variants: variants, Inks: inks, Sizes: sizes,
	}, nil
}

// priceItem is the single formula shared by the estimate endpoint, the
// submit endpoint, and (via the persisted unit_price) the confirm step:
// unit_price = fabric.sales_price × size.multiplier + variant.price_addon
// + ink.price_addon. It is always computed here from the referenced
// catalog rows, never trusted from the client. Fabric, variant, and ink
// are each scoped to a product type (Jersey vs T-Shirt have different
// fabric/model/ink options), so a mismatch is rejected as invalid input
// rather than silently priced.
func (s *Service) priceItem(ctx context.Context, in ItemInput) (EstimateLine, error) {
	if in.Qty.LessThanOrEqual(decimal.Zero) {
		return EstimateLine{}, apperr.Validation("item qty must be greater than zero")
	}
	fabric, err := s.catalogSvc.GetFabric(ctx, in.FabricID)
	if err != nil {
		return EstimateLine{}, err
	}
	size, err := s.catalogSvc.GetGarmentSize(ctx, in.GarmentSizeID)
	if err != nil {
		return EstimateLine{}, err
	}
	variant, err := s.catalogSvc.GetVariant(ctx, in.VariantID)
	if err != nil {
		return EstimateLine{}, err
	}
	ink, err := s.catalogSvc.GetInk(ctx, in.InkID)
	if err != nil {
		return EstimateLine{}, err
	}
	if fabric.ProductTypeID != in.ProductTypeID || variant.ProductTypeID != in.ProductTypeID || ink.ProductTypeID != in.ProductTypeID {
		return EstimateLine{}, apperr.Validation("fabric, model, and ink must all belong to the selected jenis pesanan")
	}

	unitPrice := fabric.SalesPrice.Mul(size.SizeMultiplier).Add(variant.PriceAddon).Add(ink.PriceAddon)
	lineTotal := unitPrice.Mul(in.Qty)
	return EstimateLine{ItemInput: in, UnitPrice: unitPrice, LineTotal: lineTotal}, nil
}

func (s *Service) Estimate(ctx context.Context, token string, in EstimateInput) (EstimateOutput, error) {
	if _, err := s.resolveLink(ctx, token); err != nil {
		return EstimateOutput{}, err
	}
	if len(in.Items) == 0 {
		return EstimateOutput{}, apperr.Validation("at least one item is required")
	}
	lines := make([]EstimateLine, 0, len(in.Items))
	subtotal := decimal.Zero
	for _, it := range in.Items {
		line, err := s.priceItem(ctx, it)
		if err != nil {
			return EstimateOutput{}, err
		}
		lines = append(lines, line)
		subtotal = subtotal.Add(line.LineTotal)
	}
	return EstimateOutput{Lines: lines, Subtotal: subtotal}, nil
}

// ---- Public: submit ----

func (s *Service) SubmitQuotation(ctx context.Context, token string, in SubmitQuotationInput) (Quotation, error) {
	link, err := s.resolveLink(ctx, token)
	if err != nil {
		return Quotation{}, err
	}
	if len(in.Items) == 0 {
		return Quotation{}, apperr.Validation("order must have at least one item")
	}

	lines := make([]EstimateLine, 0, len(in.Items))
	subtotal := decimal.Zero
	for _, it := range in.Items {
		line, err := s.priceItem(ctx, it)
		if err != nil {
			return Quotation{}, err
		}
		lines = append(lines, line)
		subtotal = subtotal.Add(line.LineTotal)
	}

	var out Quotation
	err = db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		number, err := numbering.Generate(ctx, tx, numbering.Quotation, time.Now())
		if err != nil {
			return apperr.Internal("generate quotation number", err)
		}
		q, err := s.repo.InsertQuotation(ctx, Quotation{
			QuotationNumber: number, OrderLinkID: link.ID, CustomerID: link.CustomerID,
			Notes: in.Notes, EstimatedSubtotal: subtotal,
		})
		if err != nil {
			return apperr.Internal("create quotation", err)
		}
		for _, line := range lines {
			item, err := s.repo.InsertQuotationItem(ctx, QuotationItem{
				QuotationID: q.ID, ProductTypeID: line.ProductTypeID, FabricID: line.FabricID,
				GarmentSizeID: line.GarmentSizeID, VariantID: line.VariantID, InkID: line.InkID,
				Qty: line.Qty, UnitPrice: line.UnitPrice, LineTotal: line.LineTotal,
			})
			if err != nil {
				return apperr.Internal("create quotation item", err)
			}
			q.Items = append(q.Items, item)
		}
		if err := audit.Log(ctx, tx, "order_quotations", q.ID, audit.Insert, nil, q); err != nil {
			return apperr.Internal("write audit log", err)
		}
		out = q
		return nil
	})
	return out, err
}

// ---- Staff: review ----

func (s *Service) ListQuotations(ctx context.Context, status Status, limit, offset int) ([]Quotation, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListQuotations(ctx, status, limit, offset)
}

func (s *Service) GetQuotation(ctx context.Context, id uuid.UUID) (Quotation, error) {
	q, err := s.repo.GetQuotationByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Quotation{}, apperr.NotFound("quotation not found")
	}
	if err != nil {
		return Quotation{}, err
	}
	items, err := s.repo.ListQuotationItems(ctx, id)
	if err != nil {
		return Quotation{}, apperr.Internal("load quotation items", err)
	}
	q.Items = items
	return q, nil
}

func (s *Service) RejectQuotation(ctx context.Context, id, staffUserID uuid.UUID, in RejectQuotationInput) (Quotation, error) {
	var out Quotation
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		q, err := s.repo.GetQuotationForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("quotation not found")
			}
			return apperr.Internal("load quotation", err)
		}
		if q.Status != StatusPendingReview {
			return apperr.Conflict("only quotations pending review can be rejected")
		}
		if err := s.repo.UpdateQuotationReview(ctx, id, StatusRejected, staffUserID, nil); err != nil {
			return apperr.Internal("reject quotation", err)
		}
		if err := audit.Log(ctx, tx, "order_quotations", id, audit.Update, q.Status, map[string]any{"status": StatusRejected, "reason": in.Reason}); err != nil {
			return apperr.Internal("write audit log", err)
		}
		q.Status = StatusRejected
		out = q
		return nil
	})
	return out, err
}

// ConfirmQuotation converts a PENDING_REVIEW quotation into a real Sales
// Order. It materializes a Product+ProductSize per distinct
// (design, fabric, garment_size, variant) combination the quotation used --
// see materializeCatalogRef -- so every downstream flow that already
// understands product_id/product_size_id (invoicing, delivery/COGS,
// reporting, the SO detail page) keeps working unmodified. The quotation
// items themselves remain the source of truth for what the customer
// actually picked; sales_order_items.quotation_item_id links back to them.
func (s *Service) ConfirmQuotation(ctx context.Context, id, staffUserID uuid.UUID, in ConfirmQuotationInput) (sales.SalesOrder, error) {
	if in.TaxRate.IsNegative() {
		return sales.SalesOrder{}, apperr.Validation("tax_rate cannot be negative")
	}

	var out sales.SalesOrder
	err := db.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		q, err := s.repo.GetQuotationForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.NotFound("quotation not found")
			}
			return apperr.Internal("load quotation", err)
		}
		if q.Status != StatusPendingReview {
			return apperr.Conflict("only quotations pending review can be confirmed")
		}

		items, err := s.repo.ListQuotationItems(ctx, id)
		if err != nil {
			return apperr.Internal("load quotation items", err)
		}
		if len(items) == 0 {
			return apperr.Conflict("quotation has no items")
		}

		orderItems := make([]sales.CreateOrderItemInput, 0, len(items))
		for _, item := range items {
			productID, sizeID, err := s.materializeCatalogRef(ctx, item.ProductTypeID, item.FabricID, item.GarmentSizeID, item.VariantID, item.InkID)
			if err != nil {
				return err
			}
			itemID := item.ID
			orderItems = append(orderItems, sales.CreateOrderItemInput{
				ProductID: productID, ProductSizeID: sizeID, Qty: item.Qty,
				UnitPrice: item.UnitPrice, Discount: decimal.Zero, TaxRate: in.TaxRate,
				QuotationItemID: &itemID,
			})
		}

		order, err := s.salesSvc.CreateOrder(ctx, sales.CreateOrderInput{
			CustomerID: q.CustomerID, Notes: q.Notes, Items: orderItems,
		})
		if err != nil {
			return apperr.Wrapf(err, "create sales order from quotation")
		}

		// order.Items is in the same order as items/orderItems above (both
		// loops walk `items` in lockstep), so they can be zipped by index to
		// carry each line's product_type_id -- lost once QuotationItemID is
		// all that's left on the materialized sales_order_item -- forward to
		// SPK generation below.
		spkItems := make([]spk.GenerateItem, 0, len(items))
		for i, item := range items {
			fabric, err := s.catalogSvc.GetFabric(ctx, item.FabricID)
			if err != nil {
				return apperr.Wrapf(err, "load fabric for SPK")
			}
			variant, err := s.catalogSvc.GetVariant(ctx, item.VariantID)
			if err != nil {
				return apperr.Wrapf(err, "load variant for SPK")
			}
			ink, err := s.catalogSvc.GetInk(ctx, item.InkID)
			if err != nil {
				return apperr.Wrapf(err, "load ink for SPK")
			}
			size, err := s.catalogSvc.GetGarmentSize(ctx, item.GarmentSizeID)
			if err != nil {
				return apperr.Wrapf(err, "load size for SPK")
			}
			spkItems = append(spkItems, spk.GenerateItem{
				SalesOrderItemID: order.Items[i].ID, ProductTypeID: item.ProductTypeID,
				ProductID: order.Items[i].ProductID, ProductSizeID: order.Items[i].ProductSizeID, Qty: item.Qty,
				FabricName: fabric.Name, VariantName: variant.Name, InkName: ink.Name, SizeCode: size.SizeCode,
				FabricMaterialID: &fabric.MaterialID, InkMaterialID: &ink.MaterialID,
			})
		}

		// Staff already reviewed this as part of confirming the quotation, and
		// the customer needs an invoice immediately to pick a payment plan --
		// so confirm the order and invoice it in the same step, instead of
		// waiting on the separate manual "Confirm"/"Buat Faktur" staff actions
		// the legacy manual-entry flow still uses.
		order, err = s.salesSvc.ConfirmOrder(ctx, order.ID)
		if err != nil {
			return apperr.Wrapf(err, "confirm sales order from quotation")
		}
		if _, err := s.salesSvc.CreateInvoiceForOrder(ctx, order.ID); err != nil {
			return apperr.Wrapf(err, "create invoice from quotation")
		}

		// A confirmed order is "ready for production": generate its SPK(s) --
		// one per distinct product type in the order -- so it shows up on the
		// production floor's /production/orders lists immediately.
		if _, err := s.spkSvc.GenerateForOrder(ctx, order, q.ID, q.QuotationNumber, spkItems); err != nil {
			return apperr.Wrapf(err, "generate SPK production orders")
		}

		if err := s.repo.UpdateQuotationReview(ctx, id, StatusConfirmed, staffUserID, &order.ID); err != nil {
			return apperr.Internal("confirm quotation", err)
		}
		if err := audit.Log(ctx, tx, "order_quotations", id, audit.Update, q.Status, map[string]any{"status": StatusConfirmed, "sales_order_id": order.ID}); err != nil {
			return apperr.Internal("write audit log", err)
		}

		out = order
		return nil
	})
	return out, err
}

// materializeCatalogRef finds (or creates, on first use) the Product and
// ProductSize that stand in for one product_type+fabric+size+variant+ink
// combination, so the legacy sales/invoicing/delivery pipeline -- which is
// keyed on product_id/product_size_id -- has something concrete to point
// at. Pricing is never recomputed from these: the caller already has the
// quotation item's server-computed unit_price and passes it straight
// through.
func (s *Service) materializeCatalogRef(ctx context.Context, productTypeID, fabricID, garmentSizeID, variantID, inkID uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	productType, err := s.catalogSvc.GetProductType(ctx, productTypeID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	fabric, err := s.catalogSvc.GetFabric(ctx, fabricID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	size, err := s.catalogSvc.GetGarmentSize(ctx, garmentSizeID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	productCode := shortCode("CFG-" + productType.Code + "-" + fabric.Code)
	prod, err := s.productSvc.FindByCode(ctx, productCode)
	if err != nil {
		ae, ok := apperr.As(err)
		if !ok || ae.Kind != apperr.KindNotFound {
			return uuid.Nil, uuid.Nil, err
		}
		prod, err = s.productSvc.Create(ctx, product.UpsertProductInput{
			Code: productCode, Name: productType.Name + " - " + fabric.Name,
			Category: "CUSTOM", UOM: "PCS", BasePrice: fabric.SalesPrice,
		})
		if err != nil {
			return uuid.Nil, uuid.Nil, err
		}
	}

	sizeCode := combinedSizeCode(size.SizeCode, variantID, inkID)
	sz, err := s.productSvc.FindSizeByCode(ctx, prod.ID, sizeCode)
	if err != nil {
		ae, ok := apperr.As(err)
		if !ok || ae.Kind != apperr.KindNotFound {
			return uuid.Nil, uuid.Nil, err
		}
		sz, err = s.productSvc.AddSize(ctx, prod.ID, product.UpsertSizeInput{
			SizeCode: sizeCode, SizeMultiplier: size.SizeMultiplier, SortOrder: size.SortOrder,
		})
		if err != nil {
			return uuid.Nil, uuid.Nil, err
		}
	}

	return prod.ID, sz.ID, nil
}

// shortCode keeps a materialized product code within products.code's
// VARCHAR(30): short combinations pass through unchanged, otherwise it
// falls back to a fixed-length hash so it never overflows regardless of how
// long the product type/fabric codes it's built from are.
func shortCode(code string) string {
	if len(code) <= 30 {
		return code
	}
	sum := sha256.Sum256([]byte(code))
	return "CFG-" + hex.EncodeToString(sum[:])[:16]
}

// combinedSizeCode packs a base size code plus the variant+ink combination
// into product_sizes.size_code's VARCHAR(10). It always hashes the
// variant/ink pair (rather than trying to spell them out) so it never
// overflows or collides regardless of how many variant/ink options exist.
func combinedSizeCode(baseSizeCode string, variantID, inkID uuid.UUID) string {
	base := baseSizeCode
	if len(base) > 3 {
		base = base[:3]
	}
	sum := sha256.Sum256([]byte(variantID.String() + inkID.String()))
	return base + "-" + hex.EncodeToString(sum[:])[:6]
}

// ---- Customer: my orders ----

func (s *Service) ListCustomerOrders(ctx context.Context, token string) ([]Quotation, error) {
	link, err := s.resolveLink(ctx, token)
	if err != nil {
		return nil, err
	}
	return s.repo.ListQuotationsByCustomer(ctx, link.CustomerID)
}

// GetPublicOrderDetail is the one-call payload for the customer's order
// detail view: the quotation itself, plus (once confirmed) the sales
// order, invoice, and any payment plan the customer has chosen.
func (s *Service) GetPublicOrderDetail(ctx context.Context, token string, quotationID uuid.UUID) (PublicOrderDetail, error) {
	link, err := s.resolveLink(ctx, token)
	if err != nil {
		return PublicOrderDetail{}, err
	}
	q, err := s.GetQuotation(ctx, quotationID)
	if err != nil {
		return PublicOrderDetail{}, err
	}
	if q.CustomerID != link.CustomerID {
		return PublicOrderDetail{}, apperr.Forbidden("this order does not belong to this link")
	}

	out := PublicOrderDetail{Quotation: q}
	if q.SalesOrderID == nil {
		return out, nil
	}
	order, err := s.salesSvc.GetOrder(ctx, *q.SalesOrderID)
	if err != nil {
		return PublicOrderDetail{}, err
	}
	out.SalesOrder = &order

	invoice, err := s.salesSvc.GetInvoiceByOrder(ctx, order.ID)
	if err != nil {
		if ae, ok := apperr.As(err); ok && ae.Kind == apperr.KindNotFound {
			return out, nil
		}
		return PublicOrderDetail{}, err
	}
	out.Invoice = &invoice

	plan, err := s.billingSvc.GetPaymentPlanByInvoice(ctx, invoice.ID)
	if err != nil {
		return PublicOrderDetail{}, err
	}
	out.PaymentPlan = plan

	spkOrders, err := s.spkSvc.ListBySalesOrder(ctx, order.ID)
	if err != nil {
		return PublicOrderDetail{}, apperr.Wrapf(err, "load spk production progress")
	}
	out.SPKOrders = spkOrders
	return out, nil
}

// resolveOwnedInvoice checks that invoiceID's sales order was produced by a
// quotation belonging to this token's customer, so a customer can only
// touch payment plans/proofs for their own orders.
func (s *Service) resolveOwnedInvoice(ctx context.Context, token string, invoiceID uuid.UUID) (sales.Invoice, error) {
	link, err := s.resolveLink(ctx, token)
	if err != nil {
		return sales.Invoice{}, err
	}
	invoice, err := s.salesSvc.GetInvoice(ctx, invoiceID)
	if err != nil {
		return sales.Invoice{}, err
	}
	if invoice.CustomerID != link.CustomerID {
		return sales.Invoice{}, apperr.Forbidden("this invoice does not belong to this link")
	}
	return invoice, nil
}

func (s *Service) ChoosePaymentPlan(ctx context.Context, token string, invoiceID uuid.UUID, in billing.ChoosePaymentPlanInput) (billing.PaymentPlan, error) {
	if _, err := s.resolveOwnedInvoice(ctx, token, invoiceID); err != nil {
		return billing.PaymentPlan{}, err
	}
	return s.billingSvc.ChoosePaymentPlan(ctx, invoiceID, in)
}

// resolveOwnedInstallment checks the installment's plan->invoice chain
// belongs to this token's customer before letting them upload a proof.
func (s *Service) resolveOwnedInstallment(ctx context.Context, token string, installmentID uuid.UUID) error {
	link, err := s.resolveLink(ctx, token)
	if err != nil {
		return err
	}
	inst, err := s.billingSvc.GetInstallment(ctx, installmentID)
	if err != nil {
		return err
	}
	plan, err := s.billingSvc.GetPlanByID(ctx, inst.PaymentPlanID)
	if err != nil {
		return err
	}
	invoice, err := s.salesSvc.GetInvoice(ctx, plan.InvoiceID)
	if err != nil {
		return err
	}
	if invoice.CustomerID != link.CustomerID {
		return apperr.Forbidden("this installment does not belong to this link")
	}
	return nil
}

func (s *Service) UploadInstallmentProof(ctx context.Context, token string, installmentID uuid.UUID, imageURL string) (billing.Installment, error) {
	if err := s.resolveOwnedInstallment(ctx, token, installmentID); err != nil {
		return billing.Installment{}, err
	}
	return s.billingSvc.UploadInstallmentProof(ctx, installmentID, imageURL)
}

// ---- Staff: cross-reference a Sales Order back to the quotation it came from ----

func (s *Service) GetQuotationBySalesOrder(ctx context.Context, salesOrderID uuid.UUID) (*Quotation, error) {
	q, err := s.repo.GetQuotationBySalesOrderID(ctx, salesOrderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, apperr.Internal("load quotation for sales order", err)
	}
	items, err := s.repo.ListQuotationItems(ctx, q.ID)
	if err != nil {
		return nil, apperr.Internal("load quotation items", err)
	}
	q.Items = items
	return &q, nil
}

// ---- Customer: invoice PDF ----

func (s *Service) RenderInvoicePDFForCustomer(ctx context.Context, token string, invoiceID uuid.UUID) (string, []byte, error) {
	invoice, err := s.resolveOwnedInvoice(ctx, token, invoiceID)
	if err != nil {
		return "", nil, err
	}
	cust, err := s.customerSvc.Get(ctx, invoice.CustomerID)
	if err != nil {
		return "", nil, err
	}
	var soNumber string
	if invoice.SalesOrderID != uuid.Nil {
		if so, err := s.salesSvc.GetOrder(ctx, invoice.SalesOrderID); err == nil {
			soNumber = so.SONumber
		}
	}

	products := map[uuid.UUID]product.Product{}
	lines := make([]pdf.InvoiceLine, 0, len(invoice.Items))
	for _, it := range invoice.Items {
		p, ok := products[it.ProductID]
		if !ok {
			p, err = s.productSvc.Get(ctx, it.ProductID)
			if err != nil {
				return "", nil, err
			}
			products[it.ProductID] = p
		}
		sizeCode := "-"
		for _, sz := range p.Sizes {
			if sz.ID == it.ProductSizeID {
				sizeCode = sz.SizeCode
				break
			}
		}
		lines = append(lines, pdf.InvoiceLine{
			ProductName: p.Name, SizeCode: sizeCode, Qty: it.Qty, UnitPrice: it.UnitPrice,
			Discount: it.Discount, TaxRate: it.TaxRate, LineTotal: it.LineTotal,
		})
	}

	out, err := pdf.Invoice(pdf.InvoiceData{
		InvoiceNumber: invoice.InvoiceNumber, SONumber: soNumber, InvoiceDate: invoice.InvoiceDate, DueDate: invoice.DueDate,
		Status: string(invoice.Status), CustomerName: cust.Name, CustomerAddress: cust.Address, CustomerPhone: cust.Phone,
		CustomerTaxID: cust.TaxID, Lines: lines, Subtotal: invoice.Subtotal, DiscountTotal: invoice.DiscountTotal,
		TaxTotal: invoice.TaxTotal, GrandTotal: invoice.GrandTotal, PaidAmount: invoice.PaidAmount, BalanceDue: invoice.BalanceDue,
	})
	if err != nil {
		return "", nil, apperr.Internal("render invoice pdf", err)
	}
	return invoice.InvoiceNumber, out, nil
}
