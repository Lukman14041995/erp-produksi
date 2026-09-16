package purchasing

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/material"
	"github.com/ranji/clothing-erp/internal/pdf"
	"github.com/ranji/clothing-erp/internal/response"
)

func (h *Handler) purchaseOrderPDF(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid purchase order id")
	}
	ctx := c.Request().Context()

	po, err := h.svc.GetPO(ctx, id)
	if err != nil {
		return err
	}
	sup, err := h.supplierSvc.Get(ctx, po.SupplierID)
	if err != nil {
		return err
	}

	materials := map[uuid.UUID]material.Material{}
	lines := make([]pdf.PurchaseOrderLine, 0, len(po.Items))
	for _, it := range po.Items {
		m, ok := materials[it.MaterialID]
		if !ok {
			m, err = h.materialSvc.Get(ctx, it.MaterialID)
			if err != nil {
				return err
			}
			materials[it.MaterialID] = m
		}
		lines = append(lines, pdf.PurchaseOrderLine{
			MaterialCode: m.Code,
			MaterialName: m.Name,
			UOM:          m.UOM,
			Qty:          it.Qty,
			UnitCost:     it.UnitCost,
			LineTotal:    it.LineTotal,
		})
	}

	out, err := pdf.PurchaseOrder(pdf.PurchaseOrderData{
		PONumber:        po.PONumber,
		OrderDate:       po.OrderDate,
		ExpectedDate:    po.ExpectedDate,
		Status:          string(po.Status),
		SupplierName:    sup.Name,
		SupplierAddress: sup.Address,
		SupplierPhone:   sup.Phone,
		SupplierContact: sup.ContactPerson,
		SupplierTaxID:   sup.TaxID,
		Lines:           lines,
		Subtotal:        po.Subtotal,
		GrandTotal:      po.GrandTotal,
		Notes:           po.Notes,
	})
	if err != nil {
		return apperr.Internal("render purchase order pdf", err)
	}
	return response.PDF(c, fmt.Sprintf("po-%s.pdf", po.PONumber), out)
}
