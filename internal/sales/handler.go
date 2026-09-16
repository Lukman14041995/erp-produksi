package sales

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/customer"
	"github.com/ranji/clothing-erp/internal/product"
	"github.com/ranji/clothing-erp/internal/response"
)

type Handler struct {
	svc         *Service
	customerSvc *customer.Service
	productSvc  *product.Service
}

func NewHandler(svc *Service, customerSvc *customer.Service, productSvc *product.Service) *Handler {
	return &Handler{svc: svc, customerSvc: customerSvc, productSvc: productSvc}
}

func (h *Handler) Register(g *echo.Group) {
	g.POST("/sales-orders", h.create)
	g.GET("/sales-orders", h.list)
	g.GET("/sales-orders/:id", h.get)
	g.POST("/sales-orders/:id/confirm", h.confirm)
	g.POST("/sales-orders/:id/cancel", h.cancel)
	g.POST("/sales-orders/:id/create-invoice", h.createInvoice)
	g.POST("/sales-orders/:id/deliver", h.deliver)
	g.GET("/sales-orders/:id/delivery-slip/pdf", h.deliverySlipPDF)

	g.GET("/invoices", h.listInvoices)
	g.GET("/invoices/:id", h.getInvoice)
	g.GET("/invoices/:id/pdf", h.invoicePDF)
}

func (h *Handler) create(c echo.Context) error {
	var in CreateOrderInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateOrder(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "sales order created", out)
}

func (h *Handler) list(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	out, err := h.svc.ListOrders(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "sales orders retrieved", out)
}

func (h *Handler) get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid sales order id")
	}
	out, err := h.svc.GetOrder(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "sales order retrieved", out)
}

func (h *Handler) confirm(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid sales order id")
	}
	out, err := h.svc.ConfirmOrder(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "sales order confirmed", out)
}

func (h *Handler) cancel(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid sales order id")
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = c.Bind(&body)
	out, err := h.svc.CancelOrder(c.Request().Context(), id, body.Reason)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "sales order cancelled", out)
}

func (h *Handler) createInvoice(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid sales order id")
	}
	out, err := h.svc.CreateInvoiceForOrder(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "invoice created", out)
}

func (h *Handler) deliver(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid sales order id")
	}
	out, err := h.svc.DeliverOrder(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "sales order delivered", out)
}

func (h *Handler) listInvoices(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	out, err := h.svc.ListInvoices(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "invoices retrieved", out)
}

func (h *Handler) getInvoice(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid invoice id")
	}
	out, err := h.svc.GetInvoice(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "invoice retrieved", out)
}
