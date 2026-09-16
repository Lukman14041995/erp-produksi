package material

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
	g.POST("/materials", h.create)
	g.GET("/materials", h.list)
	g.GET("/materials/:id", h.get)
	g.PUT("/materials/:id", h.update)
}

func (h *Handler) create(c echo.Context) error {
	var in UpsertInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.Create(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "material created", out)
}

func (h *Handler) list(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	out, err := h.svc.List(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "materials retrieved", out)
}

func (h *Handler) get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid material id")
	}
	out, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "material retrieved", out)
}

func (h *Handler) update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid material id")
	}
	var in UpsertInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.Update(c.Request().Context(), id, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "material updated", out)
}
