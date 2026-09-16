package production

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/material"
	"github.com/ranji/clothing-erp/internal/pdf"
	"github.com/ranji/clothing-erp/internal/response"
)

func (h *Handler) workOrderPDF(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid production order id")
	}
	ctx := c.Request().Context()

	order, err := h.svc.GetOrder(ctx, id)
	if err != nil {
		return err
	}
	p, err := h.productSvc.Get(ctx, order.ProductID)
	if err != nil {
		return err
	}

	var soNumber string
	if order.SalesOrderID != nil {
		if so, err := h.salesSvc.GetOrder(ctx, *order.SalesOrderID); err == nil {
			soNumber = so.SONumber
		}
	}

	var bomName string
	if order.BOMID != nil {
		if bom, err := h.svc.GetBOM(ctx, *order.BOMID); err == nil {
			bomName = bom.Name
		}
	}

	sizeCode := func(sizeID uuid.UUID) string {
		for _, s := range p.Sizes {
			if s.ID == sizeID {
				return s.SizeCode
			}
		}
		return "-"
	}
	sizes := make([]pdf.ProductionSizeQty, 0, len(order.Items))
	for _, it := range order.Items {
		sizes = append(sizes, pdf.ProductionSizeQty{
			SizeCode: sizeCode(it.ProductSizeID),
			Planned:  it.PlannedQty,
			Finished: it.FinishedQty,
		})
	}

	materials := map[uuid.UUID]material.Material{}
	matLines := make([]pdf.ProductionMaterialLine, 0, len(order.Materials))
	for _, m := range order.Materials {
		mat, ok := materials[m.MaterialID]
		if !ok {
			mat, err = h.materialSvc.Get(ctx, m.MaterialID)
			if err != nil {
				return err
			}
			materials[m.MaterialID] = mat
		}
		matLines = append(matLines, pdf.ProductionMaterialLine{
			MaterialCode: mat.Code,
			MaterialName: mat.Name,
			UOM:          mat.UOM,
			PlannedQty:   m.PlannedQty,
			IssuedQty:    m.IssuedQty,
		})
	}

	out, err := pdf.ProductionWorkOrder(pdf.ProductionWorkOrderData{
		ProdNumber:  order.ProdNumber,
		SONumber:    soNumber,
		ProductName: p.Name,
		ProductCode: p.Code,
		BOMName:     bomName,
		Status:      string(order.Status),
		PlannedQty:  order.PlannedQty,
		FinishedQty: order.FinishedQty,
		StartDate:   order.StartDate,
		EndDate:     order.EndDate,
		Sizes:       sizes,
		Materials:   matLines,
		Notes:       order.Notes,
	})
	if err != nil {
		return apperr.Internal("render production work order pdf", err)
	}
	return response.PDF(c, fmt.Sprintf("work-order-%s.pdf", order.ProdNumber), out)
}
