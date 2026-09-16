package inventory

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/response"
	"github.com/shopspring/decimal"
)

type Handler struct {
	svc    *Service
	repo   *Repository
	accSvc opnameAccountingIface
}

func NewHandler(svc *Service, repo *Repository, accSvc opnameAccountingIface) *Handler {
	return &Handler{svc: svc, repo: repo, accSvc: accSvc}
}

func (h *Handler) Register(g *echo.Group) {
	g.GET("/inventory/balances", h.listBalances)
	g.GET("/inventory/stock-by-warehouse", h.stockByWarehouse)
	g.GET("/inventory/transactions", h.listTransactions)
	g.POST("/inventory/adjustments", h.adjust)

	g.POST("/warehouses", h.createWarehouse)
	g.GET("/warehouses", h.listWarehouses)
	g.PUT("/warehouses/:id", h.updateWarehouse)

	g.POST("/inventory/transfers", h.createTransfer)
	g.GET("/inventory/transfers", h.listTransfers)
	g.GET("/inventory/transfers/:id", h.getTransfer)
	g.POST("/inventory/transfers/:id/dispatch", h.dispatchTransfer)
	g.POST("/inventory/transfers/:id/receive", h.receiveTransfer)
	g.POST("/inventory/transfers/:id/cancel", h.cancelTransfer)

	g.POST("/inventory/opnames", h.createOpname)
	g.GET("/inventory/opnames", h.listOpnames)
	g.GET("/inventory/opnames/:id", h.getOpname)
	g.POST("/inventory/opnames/:id/post", h.postOpname)
	g.POST("/inventory/opnames/:id/cancel", h.cancelOpname)
}

func (h *Handler) listBalances(c echo.Context) error {
	out, err := h.repo.ListBalances(c.Request().Context())
	if err != nil {
		return apperr.Internal("list balances", err)
	}
	return response.OK(c, http.StatusOK, "inventory balances retrieved", out)
}

func (h *Handler) stockByWarehouse(c echo.Context) error {
	var warehouseID *uuid.UUID
	if v := c.QueryParam("warehouse_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return apperr.Validation("invalid warehouse_id")
		}
		warehouseID = &id
	}
	out, err := h.svc.ListBalancesByWarehouse(c.Request().Context(), warehouseID)
	if err != nil {
		return apperr.Internal("list stock by warehouse", err)
	}
	return response.OK(c, http.StatusOK, "stock by warehouse retrieved", out)
}

func (h *Handler) listTransactions(c echo.Context) error {
	var materialID, productID, warehouseID *uuid.UUID
	if v := c.QueryParam("material_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return apperr.Validation("invalid material_id")
		}
		materialID = &id
	}
	if v := c.QueryParam("product_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return apperr.Validation("invalid product_id")
		}
		productID = &id
	}
	if v := c.QueryParam("warehouse_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return apperr.Validation("invalid warehouse_id")
		}
		warehouseID = &id
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	out, err := h.repo.ListTransactions(c.Request().Context(), materialID, productID, warehouseID, limit, offset)
	if err != nil {
		return apperr.Internal("list transactions", err)
	}
	return response.OK(c, http.StatusOK, "inventory transactions retrieved", out)
}

type adjustRequest struct {
	ItemType      ItemType        `json:"item_type"`
	WarehouseID   uuid.UUID       `json:"warehouse_id"`
	MaterialID    *uuid.UUID      `json:"material_id,omitempty"`
	ProductID     *uuid.UUID      `json:"product_id,omitempty"`
	ProductSizeID *uuid.UUID      `json:"product_size_id,omitempty"`
	Qty           decimal.Decimal `json:"qty"` // positive = increase stock, negative = decrease stock
	UnitCost      decimal.Decimal `json:"unit_cost"`
	Notes         string          `json:"notes"`
}

func (h *Handler) adjust(c echo.Context) error {
	var req adjustRequest
	if err := c.Bind(&req); err != nil {
		return apperr.Validation("invalid request body")
	}

	in := MovementInput{
		TxnType: TxnAdjustment, ItemType: req.ItemType, WarehouseID: req.WarehouseID,
		MaterialID: req.MaterialID, ProductID: req.ProductID, ProductSizeID: req.ProductSizeID,
		UnitCost: req.UnitCost, RefType: "MANUAL", TxnDate: time.Now(), Notes: req.Notes,
	}
	if req.Qty.IsPositive() {
		in.QtyIn = req.Qty
	} else {
		in.QtyOut = req.Qty.Neg()
	}

	out, err := h.svc.RecordMovement(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "inventory adjusted", out)
}

// ---- Warehouses ----

func (h *Handler) createWarehouse(c echo.Context) error {
	var in UpsertWarehouseInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateWarehouse(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "warehouse created", out)
}

func (h *Handler) listWarehouses(c echo.Context) error {
	out, err := h.svc.ListWarehouses(c.Request().Context())
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "warehouses retrieved", out)
}

func (h *Handler) updateWarehouse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid warehouse id")
	}
	var in UpsertWarehouseInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.UpdateWarehouse(c.Request().Context(), id, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "warehouse updated", out)
}

// ---- Stock transfers ----

func (h *Handler) createTransfer(c echo.Context) error {
	var in CreateTransferInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateTransfer(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "stock transfer created", out)
}

func (h *Handler) listTransfers(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	out, err := h.svc.ListTransfers(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "stock transfers retrieved", out)
}

func (h *Handler) getTransfer(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid transfer id")
	}
	out, err := h.svc.GetTransfer(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "stock transfer retrieved", out)
}

func (h *Handler) dispatchTransfer(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid transfer id")
	}
	out, err := h.svc.DispatchTransfer(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "stock transfer dispatched", out)
}

func (h *Handler) receiveTransfer(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid transfer id")
	}
	out, err := h.svc.ReceiveTransfer(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "stock transfer received", out)
}

func (h *Handler) cancelTransfer(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid transfer id")
	}
	out, err := h.svc.CancelTransfer(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "stock transfer cancelled", out)
}

// ---- Stock opname ----

func (h *Handler) createOpname(c echo.Context) error {
	var in CreateOpnameInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateOpname(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "stock opname created", out)
}

func (h *Handler) listOpnames(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	out, err := h.svc.ListOpnames(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "stock opnames retrieved", out)
}

func (h *Handler) getOpname(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid opname id")
	}
	out, err := h.svc.GetOpname(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "stock opname retrieved", out)
}

func (h *Handler) postOpname(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid opname id")
	}
	out, err := h.svc.PostOpname(c.Request().Context(), id, h.accSvc)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "stock opname posted", out)
}

func (h *Handler) cancelOpname(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid opname id")
	}
	out, err := h.svc.CancelOpname(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "stock opname cancelled", out)
}
