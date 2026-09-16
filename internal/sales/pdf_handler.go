package sales

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/pdf"
	"github.com/ranji/clothing-erp/internal/product"
	"github.com/ranji/clothing-erp/internal/response"
)

func sizeCodeOf(p product.Product, sizeID uuid.UUID) string {
	for _, s := range p.Sizes {
		if s.ID == sizeID {
			return s.SizeCode
		}
	}
	return "-"
}

func (h *Handler) invoicePDF(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid invoice id")
	}
	ctx := c.Request().Context()

	inv, err := h.svc.GetInvoice(ctx, id)
	if err != nil {
		return err
	}
	cust, err := h.customerSvc.Get(ctx, inv.CustomerID)
	if err != nil {
		return err
	}

	var soNumber string
	if inv.SalesOrderID != uuid.Nil {
		if so, err := h.svc.GetOrder(ctx, inv.SalesOrderID); err == nil {
			soNumber = so.SONumber
		}
	}

	products := map[uuid.UUID]product.Product{}
	lines := make([]pdf.InvoiceLine, 0, len(inv.Items))
	for _, it := range inv.Items {
		p, ok := products[it.ProductID]
		if !ok {
			p, err = h.productSvc.Get(ctx, it.ProductID)
			if err != nil {
				return err
			}
			products[it.ProductID] = p
		}
		lines = append(lines, pdf.InvoiceLine{
			ProductName: p.Name,
			SizeCode:    sizeCodeOf(p, it.ProductSizeID),
			Qty:         it.Qty,
			UnitPrice:   it.UnitPrice,
			Discount:    it.Discount,
			TaxRate:     it.TaxRate,
			LineTotal:   it.LineTotal,
		})
	}

	out, err := pdf.Invoice(pdf.InvoiceData{
		InvoiceNumber:   inv.InvoiceNumber,
		SONumber:        soNumber,
		InvoiceDate:     inv.InvoiceDate,
		DueDate:         inv.DueDate,
		Status:          string(inv.Status),
		CustomerName:    cust.Name,
		CustomerAddress: cust.Address,
		CustomerPhone:   cust.Phone,
		CustomerTaxID:   cust.TaxID,
		Lines:           lines,
		Subtotal:        inv.Subtotal,
		DiscountTotal:   inv.DiscountTotal,
		TaxTotal:        inv.TaxTotal,
		GrandTotal:      inv.GrandTotal,
		PaidAmount:      inv.PaidAmount,
		BalanceDue:      inv.BalanceDue,
	})
	if err != nil {
		return apperr.Internal("render invoice pdf", err)
	}
	return response.PDF(c, fmt.Sprintf("faktur-%s.pdf", inv.InvoiceNumber), out)
}

func (h *Handler) deliverySlipPDF(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid sales order id")
	}
	ctx := c.Request().Context()

	order, err := h.svc.GetOrder(ctx, id)
	if err != nil {
		return err
	}
	cust, err := h.customerSvc.Get(ctx, order.CustomerID)
	if err != nil {
		return err
	}

	products := map[uuid.UUID]product.Product{}
	lines := make([]pdf.DeliverySlipLine, 0, len(order.Items))
	for _, it := range order.Items {
		p, ok := products[it.ProductID]
		if !ok {
			p, err = h.productSvc.Get(ctx, it.ProductID)
			if err != nil {
				return err
			}
			products[it.ProductID] = p
		}
		lines = append(lines, pdf.DeliverySlipLine{
			ProductName: p.Name,
			SizeCode:    sizeCodeOf(p, it.ProductSizeID),
			Qty:         it.Qty,
		})
	}

	var deliveryDate = order.UpdatedAt
	if order.DeliveryStatus == DeliveryNotDelivered {
		deliveryDate = order.OrderDate
	}

	out, err := pdf.DeliverySlip(pdf.DeliverySlipData{
		SONumber:        order.SONumber,
		OrderDate:       order.OrderDate,
		DeliveryDate:    deliveryDate,
		DeliveryStatus:  string(order.DeliveryStatus),
		CustomerName:    cust.Name,
		CustomerAddress: cust.Address,
		CustomerPhone:   cust.Phone,
		Lines:           lines,
		Notes:           order.Notes,
	})
	if err != nil {
		return apperr.Internal("render delivery slip pdf", err)
	}
	return response.PDF(c, fmt.Sprintf("surat-jalan-%s.pdf", order.SONumber), out)
}
