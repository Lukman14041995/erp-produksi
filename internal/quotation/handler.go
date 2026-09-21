package quotation

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/auth"
	"github.com/ranji/clothing-erp/internal/billing"
	"github.com/ranji/clothing-erp/internal/catalog"
	"github.com/ranji/clothing-erp/internal/response"
)

type Handler struct {
	svc       *Service
	uploadDir string
}

func NewHandler(svc *Service, uploadDir string) *Handler {
	return &Handler{svc: svc, uploadDir: uploadDir}
}

// RegisterProtected wires the staff-facing routes: generating/revoking
// order links and reviewing incoming quotations. Callers mount this under
// the authenticated SALES/ADMIN route group.
func (h *Handler) RegisterProtected(g *echo.Group) {
	g.POST("/order-links", h.createLink)
	g.GET("/order-links", h.listLinks)
	g.POST("/order-links/:id/revoke", h.revokeLink)

	g.GET("/quotations", h.listQuotations)
	g.GET("/quotations/:id", h.getQuotation)
	g.POST("/quotations/:id/confirm", h.confirmQuotation)
	g.POST("/quotations/:id/reject", h.rejectQuotation)
	g.GET("/quotations/by-sales-order/:soId", h.getQuotationBySalesOrder)
}

// RegisterPublic wires the customer-facing routes reachable via a link
// token, with no JWT/login involved at all. Callers mount this under the
// unauthenticated public route group.
func (h *Handler) RegisterPublic(g *echo.Group) {
	g.GET("/public/order-links/:token", h.getPublicCatalog)
	g.POST("/public/order-links/:token/estimate", h.estimate)
	g.POST("/public/order-links/:token/quotations", h.submitQuotation)
	g.GET("/public/order-links/:token/quotations", h.listCustomerOrders)
	g.GET("/public/order-links/:token/quotations/:id", h.getPublicOrderDetail)
	g.POST("/public/order-links/:token/invoices/:invoiceId/payment-plan", h.choosePaymentPlan)
	g.GET("/public/order-links/:token/invoices/:invoiceId/pdf", h.invoicePDFForCustomer)
	g.POST("/public/order-links/:token/installments/:installmentId/proof", h.uploadInstallmentProof)
}

func staffUserID(c echo.Context) (uuid.UUID, error) {
	claims, ok := auth.ClaimsFromContext(c.Request().Context())
	if !ok {
		return uuid.Nil, apperr.Forbidden("authentication required")
	}
	return claims.UserID, nil
}

// ---- Protected: order links ----

func (h *Handler) createLink(c echo.Context) error {
	userID, err := staffUserID(c)
	if err != nil {
		return err
	}
	var in CreateLinkInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateLink(c.Request().Context(), userID, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "order link created", out)
}

func (h *Handler) listLinks(c echo.Context) error {
	customerID, err := uuid.Parse(c.QueryParam("customer_id"))
	if err != nil {
		return apperr.Validation("customer_id query param is required")
	}
	out, err := h.svc.ListLinksByCustomer(c.Request().Context(), customerID)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "order links retrieved", out)
}

func (h *Handler) revokeLink(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid order link id")
	}
	if err := h.svc.RevokeLink(c.Request().Context(), id); err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "order link revoked", nil)
}

// ---- Protected: quotation review ----

func (h *Handler) listQuotations(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	out, err := h.svc.ListQuotations(c.Request().Context(), Status(c.QueryParam("status")), limit, offset)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "quotations retrieved", out)
}

func (h *Handler) getQuotation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid quotation id")
	}
	out, err := h.svc.GetQuotation(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "quotation retrieved", out)
}

func (h *Handler) confirmQuotation(c echo.Context) error {
	userID, err := staffUserID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid quotation id")
	}
	var in ConfirmQuotationInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.ConfirmQuotation(c.Request().Context(), id, userID, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "quotation confirmed into sales order", out)
}

func (h *Handler) rejectQuotation(c echo.Context) error {
	userID, err := staffUserID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid quotation id")
	}
	var in RejectQuotationInput
	_ = c.Bind(&in)
	out, err := h.svc.RejectQuotation(c.Request().Context(), id, userID, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "quotation rejected", out)
}

// ---- Public ----

func (h *Handler) getPublicCatalog(c echo.Context) error {
	out, err := h.svc.GetPublicCatalog(c.Request().Context(), c.Param("token"))
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "catalog retrieved", out)
}

func (h *Handler) estimate(c echo.Context) error {
	var in EstimateInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.Estimate(c.Request().Context(), c.Param("token"), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "estimate computed", out)
}

func (h *Handler) submitQuotation(c echo.Context) error {
	var in SubmitQuotationInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.SubmitQuotation(c.Request().Context(), c.Param("token"), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "quotation submitted", out)
}

func (h *Handler) listCustomerOrders(c echo.Context) error {
	out, err := h.svc.ListCustomerOrders(c.Request().Context(), c.Param("token"))
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "orders retrieved", out)
}

func (h *Handler) getPublicOrderDetail(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid quotation id")
	}
	out, err := h.svc.GetPublicOrderDetail(c.Request().Context(), c.Param("token"), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "order detail retrieved", out)
}

func (h *Handler) choosePaymentPlan(c echo.Context) error {
	invoiceID, err := uuid.Parse(c.Param("invoiceId"))
	if err != nil {
		return apperr.Validation("invalid invoice id")
	}
	var in billing.ChoosePaymentPlanInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.ChoosePaymentPlan(c.Request().Context(), c.Param("token"), invoiceID, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "payment plan created", out)
}

func (h *Handler) invoicePDFForCustomer(c echo.Context) error {
	invoiceID, err := uuid.Parse(c.Param("invoiceId"))
	if err != nil {
		return apperr.Validation("invalid invoice id")
	}
	filename, data, err := h.svc.RenderInvoicePDFForCustomer(c.Request().Context(), c.Param("token"), invoiceID)
	if err != nil {
		return err
	}
	return response.PDF(c, "faktur-"+filename+".pdf", data)
}

func (h *Handler) uploadInstallmentProof(c echo.Context) error {
	installmentID, err := uuid.Parse(c.Param("installmentId"))
	if err != nil {
		return apperr.Validation("invalid installment id")
	}
	fh, err := c.FormFile("proof")
	if err != nil {
		return apperr.Validation("proof file is required (field name: proof)")
	}
	url, err := catalog.SaveCatalogImage(fh, "payment-proofs", h.uploadDir)
	if err != nil {
		return err
	}
	out, err := h.svc.UploadInstallmentProof(c.Request().Context(), c.Param("token"), installmentID, url)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "proof of payment uploaded", out)
}

func (h *Handler) getQuotationBySalesOrder(c echo.Context) error {
	soID, err := uuid.Parse(c.Param("soId"))
	if err != nil {
		return apperr.Validation("invalid sales order id")
	}
	out, err := h.svc.GetQuotationBySalesOrder(c.Request().Context(), soID)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "quotation retrieved", out)
}
