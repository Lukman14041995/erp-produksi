package accounting

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/response"
	"github.com/shopspring/decimal"
)

type Handler struct {
	svc        *Service
	periodRepo *PeriodRepository
}

func NewHandler(svc *Service, periodRepo *PeriodRepository) *Handler {
	return &Handler{svc: svc, periodRepo: periodRepo}
}

// RegisterReadOnly mounts journal-viewing routes that any authenticated
// role may use -- e.g. a sales order's detail page needs to show its own
// posting history without granting SALES general accounting access.
func (h *Handler) RegisterReadOnly(g *echo.Group) {
	g.GET("/journals", h.list)
	g.GET("/journals/audit-trail", h.auditTrail)
	g.GET("/journals/:id", h.get)
	g.GET("/accounting-periods", h.listPeriods)
}

// Register mounts the mutating accounting routes: manual journals,
// reversals, and period close/reopen. Callers gate this group to
// ADMIN/ACCOUNTING.
func (h *Handler) Register(g *echo.Group) {
	g.POST("/journals", h.postManual)
	g.POST("/journals/:id/reverse", h.reverse)

	g.POST("/accounting-periods", h.createPeriod)
	g.POST("/accounting-periods/:period/close", h.closePeriod)
	g.POST("/accounting-periods/:period/reopen", h.reopenPeriod)
}

func (h *Handler) list(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	journals, err := h.svc.ListJournals(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "journals retrieved", journals)
}

func (h *Handler) auditTrail(c echo.Context) error {
	raw := c.QueryParam("source_ids")
	if raw == "" {
		return apperr.Validation("source_ids is required (comma-separated UUIDs)")
	}
	parts := strings.Split(raw, ",")
	ids := make([]uuid.UUID, 0, len(parts))
	for _, p := range parts {
		id, err := uuid.Parse(strings.TrimSpace(p))
		if err != nil {
			return apperr.Validation("invalid uuid in source_ids: " + p)
		}
		ids = append(ids, id)
	}

	journals, err := h.svc.ListForAuditTrail(c.Request().Context(), ids)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "audit trail retrieved", journals)
}

func (h *Handler) get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid journal id")
	}
	j, err := h.svc.GetJournal(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "journal retrieved", j)
}

type manualLineRequest struct {
	AccountCode string          `json:"account_code"`
	Debit       decimal.Decimal `json:"debit"`
	Credit      decimal.Decimal `json:"credit"`
	Description string          `json:"description"`
}

type postManualRequest struct {
	JournalDate string              `json:"journal_date"`
	Description string              `json:"description"`
	Lines       []manualLineRequest `json:"lines"`
}

func (h *Handler) postManual(c echo.Context) error {
	var req postManualRequest
	if err := c.Bind(&req); err != nil {
		return apperr.Validation("invalid request body")
	}

	journalDate := time.Now()
	if req.JournalDate != "" {
		parsed, err := time.Parse("2006-01-02", req.JournalDate)
		if err != nil {
			return apperr.Validation("journal_date must be YYYY-MM-DD")
		}
		journalDate = parsed
	}

	lines := make([]JournalLineInput, len(req.Lines))
	for i, l := range req.Lines {
		lines[i] = JournalLineInput{AccountCode: l.AccountCode, Debit: l.Debit, Credit: l.Credit, Description: l.Description}
	}

	j, err := h.svc.PostJournal(c.Request().Context(), SourceManual, nil, journalDate, req.Description, lines)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "journal posted", j)
}

func (h *Handler) reverse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid journal id")
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = c.Bind(&body)

	j, err := h.svc.ReverseJournal(c.Request().Context(), id, body.Reason)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "journal reversed", j)
}

func (h *Handler) listPeriods(c echo.Context) error {
	periods, err := h.periodRepo.List(c.Request().Context())
	if err != nil {
		return apperr.Internal("list periods", err)
	}
	return response.OK(c, http.StatusOK, "periods retrieved", periods)
}

func (h *Handler) createPeriod(c echo.Context) error {
	var body struct {
		Period    string `json:"period"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	if err := c.Bind(&body); err != nil {
		return apperr.Validation("invalid request body")
	}
	start, err := time.Parse("2006-01-02", body.StartDate)
	if err != nil {
		return apperr.Validation("start_date must be YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", body.EndDate)
	if err != nil {
		return apperr.Validation("end_date must be YYYY-MM-DD")
	}
	p, err := h.periodRepo.Create(c.Request().Context(), body.Period, start, end)
	if err != nil {
		return apperr.Internal("create period", err)
	}
	return response.OK(c, http.StatusCreated, "period created", p)
}

func (h *Handler) closePeriod(c echo.Context) error {
	actor := c.Request().Header.Get("X-User")
	p, err := h.periodRepo.SetStatus(c.Request().Context(), c.Param("period"), PeriodClosed, &actor)
	if err != nil {
		return apperr.Internal("close period", err)
	}
	return response.OK(c, http.StatusOK, "period closed", p)
}

func (h *Handler) reopenPeriod(c echo.Context) error {
	p, err := h.periodRepo.SetStatus(c.Request().Context(), c.Param("period"), PeriodOpen, nil)
	if err != nil {
		return apperr.Internal("reopen period", err)
	}
	return response.OK(c, http.StatusOK, "period reopened", p)
}
