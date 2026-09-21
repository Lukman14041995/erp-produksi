package billing

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/auth"
	"github.com/ranji/clothing-erp/internal/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(g *echo.Group) {
	g.POST("/billing/bank-accounts", h.createBankAccount)
	g.GET("/billing/bank-accounts", h.listBankAccounts)
	g.PUT("/billing/bank-accounts/:id", h.updateBankAccount)

	g.GET("/billing/payment-plans/by-invoice/:invoiceId", h.getPlanByInvoice)
	g.POST("/billing/installments/:id/confirm", h.confirmInstallment)
	g.POST("/billing/installments/:id/reject", h.rejectInstallment)
}

func (h *Handler) createBankAccount(c echo.Context) error {
	var in UpsertBankAccountInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateBankAccount(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "bank account created", out)
}

func (h *Handler) listBankAccounts(c echo.Context) error {
	out, err := h.svc.ListBankAccounts(c.Request().Context(), false)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "bank accounts retrieved", out)
}

func (h *Handler) updateBankAccount(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid bank account id")
	}
	var in UpsertBankAccountInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.UpdateBankAccount(c.Request().Context(), id, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "bank account updated", out)
}

func (h *Handler) getPlanByInvoice(c echo.Context) error {
	invoiceID, err := uuid.Parse(c.Param("invoiceId"))
	if err != nil {
		return apperr.Validation("invalid invoice id")
	}
	out, err := h.svc.GetPaymentPlanByInvoice(c.Request().Context(), invoiceID)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "payment plan retrieved", out)
}

func (h *Handler) confirmInstallment(c echo.Context) error {
	claims, ok := auth.ClaimsFromContext(c.Request().Context())
	if !ok {
		return apperr.Forbidden("authentication required")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid installment id")
	}
	out, err := h.svc.ConfirmInstallment(c.Request().Context(), id, claims.UserID)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "installment confirmed", out)
}

func (h *Handler) rejectInstallment(c echo.Context) error {
	claims, ok := auth.ClaimsFromContext(c.Request().Context())
	if !ok {
		return apperr.Forbidden("authentication required")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid installment id")
	}
	var in RejectInstallmentInput
	_ = c.Bind(&in)
	out, err := h.svc.RejectInstallment(c.Request().Context(), id, claims.UserID, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "installment rejected", out)
}
