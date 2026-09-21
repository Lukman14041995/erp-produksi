package spk

import (
	"net/http"
	"strconv"

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

// RegisterReadOnly wires the list/detail routes under the broadly-shared
// route group: progress visibility is the point of this module ("tim
// produksi, customer, dan tim lain bisa melihat progress"), so any
// authenticated role can read it.
func (h *Handler) RegisterReadOnly(g *echo.Group) {
	g.GET("/spk-orders", h.list)
	g.GET("/spk-orders/:id", h.get)
}

// Register wires the mutating routes -- logging stage progress and
// milestones is restricted to PRODUCTION/ADMIN by the caller's route group.
func (h *Handler) Register(g *echo.Group) {
	g.POST("/spk-stages/:id/log", h.logStage)
	g.POST("/spk-stages/:id/milestone", h.markMilestone)
	g.POST("/spk-orders/:id/materials", h.issueMaterial)
	g.POST("/spk-orders/:id/labor", h.addLabor)
	g.POST("/spk-orders/:id/overheads", h.addOverhead)
}

func actorUserID(c echo.Context) (uuid.UUID, error) {
	claims, ok := auth.ClaimsFromContext(c.Request().Context())
	if !ok {
		return uuid.Nil, apperr.Forbidden("authentication required")
	}
	return claims.UserID, nil
}

func (h *Handler) list(c echo.Context) error {
	if soID := c.QueryParam("sales_order_id"); soID != "" {
		id, err := uuid.Parse(soID)
		if err != nil {
			return apperr.Validation("invalid sales_order_id")
		}
		out, err := h.svc.ListBySalesOrder(c.Request().Context(), id)
		if err != nil {
			return err
		}
		return response.OK(c, http.StatusOK, "spk orders retrieved", out)
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	productTypeCode := c.QueryParam("product_type_code")
	if productTypeCode == "" {
		out, err := h.svc.ListAllOrders(c.Request().Context(), limit, offset)
		if err != nil {
			return err
		}
		return response.OK(c, http.StatusOK, "spk orders retrieved", out)
	}

	out, err := h.svc.ListOrders(c.Request().Context(), productTypeCode, limit, offset)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "spk orders retrieved", out)
}

func (h *Handler) get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid spk order id")
	}
	out, err := h.svc.GetOrder(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "spk order retrieved", out)
}

func (h *Handler) logStage(c echo.Context) error {
	stageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid spk stage id")
	}
	userID, err := actorUserID(c)
	if err != nil {
		return err
	}
	var in LogStageInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.LogStageProgress(c.Request().Context(), stageID, userID, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "progress recorded", out)
}

func (h *Handler) issueMaterial(c echo.Context) error {
	spkOrderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid spk order id")
	}
	userID, err := actorUserID(c)
	if err != nil {
		return err
	}
	var in IssueMaterialInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.IssueMaterial(c.Request().Context(), spkOrderID, userID, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "material issued", out)
}

func (h *Handler) addLabor(c echo.Context) error {
	spkOrderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid spk order id")
	}
	userID, err := actorUserID(c)
	if err != nil {
		return err
	}
	var in AddLaborInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.AddLabor(c.Request().Context(), spkOrderID, userID, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "labor recorded", out)
}

func (h *Handler) addOverhead(c echo.Context) error {
	spkOrderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid spk order id")
	}
	userID, err := actorUserID(c)
	if err != nil {
		return err
	}
	var in AddOverheadInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.AddOverhead(c.Request().Context(), spkOrderID, userID, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "overhead recorded", out)
}

func (h *Handler) markMilestone(c echo.Context) error {
	stageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid spk stage id")
	}
	userID, err := actorUserID(c)
	if err != nil {
		return err
	}
	var in MarkMilestoneInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.MarkMilestone(c.Request().Context(), stageID, userID, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "stage marked done", out)
}
