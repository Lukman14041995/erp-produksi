package catalog

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/response"
)

type Handler struct {
	svc       *Service
	uploadDir string
}

func NewHandler(svc *Service, uploadDir string) *Handler {
	return &Handler{svc: svc, uploadDir: uploadDir}
}

func (h *Handler) Register(g *echo.Group) {
	g.POST("/catalog/product-types", h.createProductType)
	g.GET("/catalog/product-types", h.listProductTypes)
	g.GET("/catalog/product-types/:id", h.getProductType)
	g.PUT("/catalog/product-types/:id", h.updateProductType)

	g.POST("/catalog/fabrics", h.createFabric)
	g.GET("/catalog/fabrics", h.listFabrics)
	g.GET("/catalog/fabrics/:id", h.getFabric)
	g.PUT("/catalog/fabrics/:id", h.updateFabric)
	g.POST("/catalog/fabrics/:id/image", h.uploadFabricImage)

	g.POST("/catalog/variants", h.createVariant)
	g.GET("/catalog/variants", h.listVariants)
	g.GET("/catalog/variants/:id", h.getVariant)
	g.PUT("/catalog/variants/:id", h.updateVariant)
	g.POST("/catalog/variants/:id/icon", h.uploadVariantIcon)

	g.POST("/catalog/sizes", h.createSize)
	g.GET("/catalog/sizes", h.listSizes)
	g.GET("/catalog/sizes/:id", h.getSize)
	g.PUT("/catalog/sizes/:id", h.updateSize)

	g.POST("/catalog/inks", h.createInk)
	g.GET("/catalog/inks", h.listInks)
	g.GET("/catalog/inks/:id", h.getInk)
	g.PUT("/catalog/inks/:id", h.updateInk)
	g.POST("/catalog/inks/:id/image", h.uploadInkImage)
}

// ---- Product types ----

func (h *Handler) createProductType(c echo.Context) error {
	var in UpsertProductTypeInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateProductType(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "product type created", out)
}

func (h *Handler) listProductTypes(c echo.Context) error {
	out, err := h.svc.ListProductTypes(c.Request().Context(), false)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "product types retrieved", out)
}

func (h *Handler) getProductType(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid product type id")
	}
	out, err := h.svc.GetProductType(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "product type retrieved", out)
}

func (h *Handler) updateProductType(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid product type id")
	}
	var in UpsertProductTypeInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.UpdateProductType(c.Request().Context(), id, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "product type updated", out)
}

// ---- Fabrics ----

func (h *Handler) createFabric(c echo.Context) error {
	var in UpsertFabricInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateFabric(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "fabric created", out)
}

func (h *Handler) listFabrics(c echo.Context) error {
	out, err := h.svc.ListFabrics(c.Request().Context(), false)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "fabrics retrieved", out)
}

func (h *Handler) getFabric(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid fabric id")
	}
	out, err := h.svc.GetFabric(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "fabric retrieved", out)
}

func (h *Handler) updateFabric(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid fabric id")
	}
	var in UpsertFabricInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.UpdateFabric(c.Request().Context(), id, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "fabric updated", out)
}

func (h *Handler) uploadFabricImage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid fabric id")
	}
	fh, err := c.FormFile("image")
	if err != nil {
		return apperr.Validation("image file is required (field name: image)")
	}
	url, err := SaveCatalogImage(fh, "fabrics", h.uploadDir)
	if err != nil {
		return err
	}
	out, err := h.svc.SetFabricImage(c.Request().Context(), id, url)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "fabric image uploaded", out)
}

// ---- Garment variants (Model potongan) ----

func (h *Handler) createVariant(c echo.Context) error {
	var in UpsertVariantInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateVariant(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "garment variant created", out)
}

func (h *Handler) listVariants(c echo.Context) error {
	out, err := h.svc.ListVariants(c.Request().Context(), false)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "garment variants retrieved", out)
}

func (h *Handler) getVariant(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid variant id")
	}
	out, err := h.svc.GetVariant(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "garment variant retrieved", out)
}

func (h *Handler) updateVariant(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid variant id")
	}
	var in UpsertVariantInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.UpdateVariant(c.Request().Context(), id, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "garment variant updated", out)
}

func (h *Handler) uploadVariantIcon(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid variant id")
	}
	fh, err := c.FormFile("icon")
	if err != nil {
		return apperr.Validation("icon file is required (field name: icon)")
	}
	url, err := SaveCatalogImage(fh, "variants", h.uploadDir)
	if err != nil {
		return err
	}
	out, err := h.svc.SetVariantIcon(c.Request().Context(), id, url)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "variant icon uploaded", out)
}

// ---- Garment sizes ----

func (h *Handler) createSize(c echo.Context) error {
	var in UpsertSizeInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateGarmentSize(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "garment size created", out)
}

func (h *Handler) listSizes(c echo.Context) error {
	out, err := h.svc.ListGarmentSizes(c.Request().Context(), false)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "garment sizes retrieved", out)
}

func (h *Handler) getSize(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid size id")
	}
	out, err := h.svc.GetGarmentSize(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "garment size retrieved", out)
}

func (h *Handler) updateSize(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid size id")
	}
	var in UpsertSizeInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.UpdateGarmentSize(c.Request().Context(), id, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "garment size updated", out)
}

// ---- Inks (Tinta) ----

func (h *Handler) createInk(c echo.Context) error {
	var in UpsertInkInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.CreateInk(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "ink created", out)
}

func (h *Handler) listInks(c echo.Context) error {
	out, err := h.svc.ListInks(c.Request().Context(), false)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "inks retrieved", out)
}

func (h *Handler) getInk(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid ink id")
	}
	out, err := h.svc.GetInk(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "ink retrieved", out)
}

func (h *Handler) updateInk(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid ink id")
	}
	var in UpsertInkInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	out, err := h.svc.UpdateInk(c.Request().Context(), id, in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "ink updated", out)
}

func (h *Handler) uploadInkImage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.Validation("invalid ink id")
	}
	fh, err := c.FormFile("image")
	if err != nil {
		return apperr.Validation("image file is required (field name: image)")
	}
	url, err := SaveCatalogImage(fh, "inks", h.uploadDir)
	if err != nil {
		return err
	}
	out, err := h.svc.SetInkImage(c.Request().Context(), id, url)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "ink image uploaded", out)
}
