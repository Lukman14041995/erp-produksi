package coa

import (
	"net/http"

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
	g.POST("/accounts", h.create)
	g.GET("/accounts", h.list)
	g.GET("/accounts/:id", h.get)
	g.PUT("/accounts/:id", h.update)
}

func (h *Handler) create(c echo.Context) error {
	var in CreateAccountInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	a, err := h.svc.Create(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "account created", a)
}

func (h *Handler) list(c echo.Context) error {
	accounts, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "accounts retrieved", accounts)
}

func (h *Handler) get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid account id")
	}
	a, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "account retrieved", a)
}

func (h *Handler) update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid account id")
	}
	var in UpdateAccountInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	a, err := h.svc.Update(c.Request().Context(), id, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "account updated", a)
}
