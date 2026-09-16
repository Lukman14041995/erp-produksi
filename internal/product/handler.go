package product

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
	g.POST("/products", h.create)
	g.GET("/products", h.list)
	g.GET("/products/:id", h.get)
	g.PUT("/products/:id", h.update)
	g.POST("/products/:id/sizes", h.addSize)
	g.GET("/products/:id/sizes", h.listSizes)
	g.PUT("/products/:id/sizes/:sizeId", h.updateSize)
}

func (h *Handler) create(c echo.Context) error {
	var in UpsertProductInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.Create(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "product created", out)
}

func (h *Handler) list(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	out, err := h.svc.List(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "products retrieved", out)
}

func (h *Handler) get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid product id")
	}
	out, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "product retrieved", out)
}

func (h *Handler) update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid product id")
	}
	var in UpsertProductInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.Update(c.Request().Context(), id, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "product updated", out)
}

func (h *Handler) addSize(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid product id")
	}
	var in UpsertSizeInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.AddSize(c.Request().Context(), id, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "product size created", out)
}

func (h *Handler) listSizes(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid product id")
	}
	out, err := h.svc.ListSizes(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "product sizes retrieved", out)
}

func (h *Handler) updateSize(c echo.Context) error {
	sizeID, err := uuid.Parse(c.Param("sizeId"))
	if err != nil {
		return apperr.Validation("invalid size id")
	}
	var in UpsertSizeInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.UpdateSize(c.Request().Context(), sizeID, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "product size updated", out)
}
