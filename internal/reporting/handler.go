package reporting

import (
	"net/http"
	"time"

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
	g.GET("/reports/trial-balance", h.trialBalance)
	g.GET("/reports/profit-loss", h.profitLoss)
	g.GET("/reports/balance-sheet", h.balanceSheet)
	g.GET("/reports/cash-flow", h.cashFlow)
	g.GET("/reports/ar-aging", h.arAging)
	g.GET("/reports/ap-aging", h.apAging)
	g.GET("/reports/order-profitability", h.orderProfitability)
}

func parseDate(c echo.Context, param string, fallback time.Time) (time.Time, error) {
	v := c.QueryParam(param)
	if v == "" {
		return fallback, nil
	}
	return time.Parse("2006-01-02", v)
}

func (h *Handler) trialBalance(c echo.Context) error {
	asOf, err := parseDate(c, "as_of", time.Now())
	if err != nil {
		return apperr.Validation("as_of must be YYYY-MM-DD")
	}
	out, err := h.svc.TrialBalance(c.Request().Context(), asOf)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "trial balance retrieved", out)
}

func (h *Handler) profitLoss(c echo.Context) error {
	from, err := parseDate(c, "from", time.Now().AddDate(0, -1, 0))
	if err != nil {
		return apperr.Validation("from must be YYYY-MM-DD")
	}
	to, err := parseDate(c, "to", time.Now())
	if err != nil {
		return apperr.Validation("to must be YYYY-MM-DD")
	}
	out, err := h.svc.ProfitLoss(c.Request().Context(), from, to)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "profit and loss retrieved", out)
}

func (h *Handler) balanceSheet(c echo.Context) error {
	asOf, err := parseDate(c, "as_of", time.Now())
	if err != nil {
		return apperr.Validation("as_of must be YYYY-MM-DD")
	}
	out, err := h.svc.BalanceSheet(c.Request().Context(), asOf)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "balance sheet retrieved", out)
}

func (h *Handler) cashFlow(c echo.Context) error {
	from, err := parseDate(c, "from", time.Now().AddDate(0, -1, 0))
	if err != nil {
		return apperr.Validation("from must be YYYY-MM-DD")
	}
	to, err := parseDate(c, "to", time.Now())
	if err != nil {
		return apperr.Validation("to must be YYYY-MM-DD")
	}
	out, err := h.svc.CashFlow(c.Request().Context(), from, to)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "cash flow retrieved", out)
}

func (h *Handler) arAging(c echo.Context) error {
	asOf, err := parseDate(c, "as_of", time.Now())
	if err != nil {
		return apperr.Validation("as_of must be YYYY-MM-DD")
	}
	out, err := h.svc.ARAging(c.Request().Context(), asOf)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "AR aging retrieved", out)
}

func (h *Handler) apAging(c echo.Context) error {
	asOf, err := parseDate(c, "as_of", time.Now())
	if err != nil {
		return apperr.Validation("as_of must be YYYY-MM-DD")
	}
	out, err := h.svc.APAging(c.Request().Context(), asOf)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "AP aging retrieved", out)
}

func (h *Handler) orderProfitability(c echo.Context) error {
	out, err := h.svc.OrderProfitability(c.Request().Context())
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "order profitability retrieved", out)
}
