package purchasing

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/material"
	"github.com/ranji/clothing-erp/internal/response"
	"github.com/ranji/clothing-erp/internal/supplier"
)

type Handler struct {
	svc         *Service
	supplierSvc *supplier.Service
	materialSvc *material.Service
}

func NewHandler(svc *Service, supplierSvc *supplier.Service, materialSvc *material.Service) *Handler {
	return &Handler{svc: svc, supplierSvc: supplierSvc, materialSvc: materialSvc}
}

func (h *Handler) Register(g *echo.Group) {
	g.POST("/supplier-invoices", h.create)
	g.GET("/supplier-invoices", h.list)
	g.GET("/supplier-invoices/:id", h.get)
	g.POST("/supplier-invoices/:id/post", h.post)
	g.POST("/supplier-invoices/:id/void", h.void)
	g.POST("/supplier-invoices/from-grn", h.createFromGRN)

	g.POST("/purchase-orders", h.createPO)
	g.GET("/purchase-orders", h.listPOs)
	g.GET("/purchase-orders/:id", h.getPO)
	g.POST("/purchase-orders/:id/approve", h.approvePO)
	g.POST("/purchase-orders/:id/cancel", h.cancelPO)
	g.GET("/purchase-orders/:id/goods-receipts", h.listGRNsByPO)
	g.GET("/purchase-orders/:id/pdf", h.purchaseOrderPDF)

	g.POST("/goods-receipts", h.createGRN)
	g.GET("/goods-receipts", h.listGRNs)
	g.GET("/goods-receipts/:id", h.getGRN)
}

func (h *Handler) create(c echo.Context) error {
	var in CreateBillInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateBill(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "supplier bill created", out)
}

func (h *Handler) list(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	out, err := h.svc.ListBills(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "supplier bills retrieved", out)
}

func (h *Handler) get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid bill id")
	}
	out, err := h.svc.GetBill(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "supplier bill retrieved", out)
}

func (h *Handler) post(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid bill id")
	}
	out, err := h.svc.PostBill(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "supplier bill posted", out)
}

func (h *Handler) void(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid bill id")
	}
	out, err := h.svc.VoidBill(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "supplier bill voided", out)
}

func (h *Handler) createFromGRN(c echo.Context) error {
	var in CreateBillFromGRNInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateBillFromGRN(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "supplier bill matched and posted", out)
}

// ---- Purchase Orders ----

func (h *Handler) createPO(c echo.Context) error {
	var in CreatePOInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreatePO(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "purchase order created", out)
}

func (h *Handler) listPOs(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	out, err := h.svc.ListPOs(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "purchase orders retrieved", out)
}

func (h *Handler) getPO(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid purchase order id")
	}
	out, err := h.svc.GetPO(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "purchase order retrieved", out)
}

func (h *Handler) approvePO(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid purchase order id")
	}
	out, err := h.svc.ApprovePO(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "purchase order approved", out)
}

func (h *Handler) cancelPO(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid purchase order id")
	}
	out, err := h.svc.CancelPO(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "purchase order cancelled", out)
}

func (h *Handler) listGRNsByPO(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid purchase order id")
	}
	out, err := h.svc.ListGoodsReceiptsByPO(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "goods receipts retrieved", out)
}

// ---- Goods Receipts ----

func (h *Handler) createGRN(c echo.Context) error {
	var in CreateGRNInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateGoodsReceipt(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "goods receipt posted", out)
}

func (h *Handler) listGRNs(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	out, err := h.svc.ListGoodsReceipts(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "goods receipts retrieved", out)
}

func (h *Handler) getGRN(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid goods receipt id")
	}
	out, err := h.svc.GetGoodsReceipt(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "goods receipt retrieved", out)
}
