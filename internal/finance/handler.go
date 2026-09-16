package finance

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(g *echo.Group) {
	g.POST("/payments", h.createPayment)
	g.GET("/payments", h.listPayments)
	g.GET("/payments/:id", h.getPayment)
	g.POST("/payments/:id/post", h.postPayment)
	g.POST("/payments/:id/void", h.voidPayment)

	g.POST("/expenses", h.createExpense)
	g.GET("/expenses", h.listExpenses)
	g.GET("/expenses/:id", h.getExpense)
	g.POST("/expenses/:id/post", h.postExpense)
	g.POST("/expenses/:id/void", h.voidExpense)
}

func (h *Handler) createPayment(c echo.Context) error {
	var in CreatePaymentInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreatePayment(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "payment created", out)
}

func (h *Handler) listPayments(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	out, err := h.svc.ListPayments(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "payments retrieved", out)
}

func (h *Handler) getPayment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid payment id")
	}
	out, err := h.svc.GetPayment(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "payment retrieved", out)
}

func (h *Handler) postPayment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid payment id")
	}
	out, err := h.svc.PostPayment(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "payment posted", out)
}

func (h *Handler) voidPayment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid payment id")
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = c.Bind(&body)
	out, err := h.svc.VoidPayment(c.Request().Context(), id, body.Reason)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "payment voided", out)
}

func (h *Handler) createExpense(c echo.Context) error {
	var in CreateExpenseInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateExpense(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "expense created", out)
}

func (h *Handler) listExpenses(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	out, err := h.svc.ListExpenses(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "expenses retrieved", out)
}

func (h *Handler) getExpense(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid expense id")
	}
	out, err := h.svc.GetExpense(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "expense retrieved", out)
}

func (h *Handler) postExpense(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid expense id")
	}
	out, err := h.svc.PostExpense(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "expense posted", out)
}

func (h *Handler) voidExpense(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid expense id")
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = c.Bind(&body)
	out, err := h.svc.VoidExpense(c.Request().Context(), id, body.Reason)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "expense voided", out)
}
