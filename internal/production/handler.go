package production

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/material"
	"github.com/ranji/clothing-erp/internal/product"
	"github.com/ranji/clothing-erp/internal/response"
	"github.com/ranji/clothing-erp/internal/sales"
)

type Handler struct {
	svc         *Service
	productSvc  *product.Service
	materialSvc *material.Service
	salesSvc    *sales.Service
}

func NewHandler(svc *Service, productSvc *product.Service, materialSvc *material.Service, salesSvc *sales.Service) *Handler {
	return &Handler{svc: svc, productSvc: productSvc, materialSvc: materialSvc, salesSvc: salesSvc}
}

func (h *Handler) Register(g *echo.Group) {
	g.POST("/boms", h.createBOM)
	g.GET("/boms/:id", h.getBOM)
	g.GET("/products/:productId/boms", h.listBOMsByProduct)

	g.POST("/production-orders", h.create)
	g.GET("/production-orders", h.list)
	g.GET("/production-orders/:id", h.get)
	g.GET("/production-orders/:id/cost", h.getCost)
	g.GET("/production-orders/:id/pdf", h.workOrderPDF)
	g.POST("/production-orders/:id/start", h.start)
	g.POST("/production-orders/:id/cancel", h.cancel)
	g.POST("/production-orders/:id/materials/:materialRowId/issue", h.issueMaterial)
	g.POST("/production-orders/:id/labor", h.addLabor)
	g.POST("/production-orders/:id/overheads", h.addOverhead)
	g.POST("/production-orders/:id/complete", h.complete)
}

func (h *Handler) createBOM(c echo.Context) error {
	var in CreateBOMInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateBOM(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "bom created", out)
}

func (h *Handler) getBOM(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid bom id")
	}
	out, err := h.svc.GetBOM(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "bom retrieved", out)
}

func (h *Handler) listBOMsByProduct(c echo.Context) error {
	id, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		return apperr.Validation("invalid product id")
	}
	out, err := h.svc.ListBOMsByProduct(c.Request().Context(), id)
	if err != nil {
		return apperr.Internal("list boms", err)
	}
	return response.OK(c, http.StatusOK, "boms retrieved", out)
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
	return response.OK(c, http.StatusCreated, "production order created", out)
}

func (h *Handler) list(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	out, err := h.svc.ListOrders(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "production orders retrieved", out)
}

func (h *Handler) get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid production order id")
	}
	out, err := h.svc.GetOrder(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "production order retrieved", out)
}

func (h *Handler) getCost(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid production order id")
	}
	out, err := h.svc.GetCost(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "HPP snapshot retrieved", out)
}

func (h *Handler) start(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid production order id")
	}
	out, err := h.svc.StartOrder(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "production order started", out)
}

func (h *Handler) cancel(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid production order id")
	}
	out, err := h.svc.CancelOrder(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "production order cancelled", out)
}

func (h *Handler) issueMaterial(c echo.Context) error {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid production order id")
	}
	materialRowID, err := uuid.Parse(c.Param("materialRowId"))
	if err != nil {
		return apperr.Validation("invalid material row id")
	}
	var in IssueMaterialInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.IssueMaterial(c.Request().Context(), orderID, materialRowID, in.Qty)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "material issued", out)
}

func (h *Handler) addLabor(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid production order id")
	}
	var in AddLaborInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.AddLabor(c.Request().Context(), id, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "labor recorded", out)
}

func (h *Handler) addOverhead(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid production order id")
	}
	var in AddOverheadInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.AddOverhead(c.Request().Context(), id, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "overhead recorded", out)
}

func (h *Handler) complete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid production order id")
	}
	var in CompleteOrderInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CompleteOrder(c.Request().Context(), id, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "production order completed", out)
}
