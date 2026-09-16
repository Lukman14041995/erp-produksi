package auth

import (
	"net/http"

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

// RegisterPublic mounts the unauthenticated auth routes (login, refresh,
// and register -- register self-gates: open only for the very first user).
func (h *Handler) RegisterPublic(g *echo.Group) {
	g.POST("/auth/register", h.register)
	g.POST("/auth/login", h.login)
	g.POST("/auth/refresh", h.refresh)
}

// RegisterProtected mounts routes that require an already-valid session.
func (h *Handler) RegisterProtected(g *echo.Group) {
	g.GET("/auth/me", h.me)
}

// RegisterAdmin mounts user-management routes. Callers gate this group to ADMIN.
func (h *Handler) RegisterAdmin(g *echo.Group) {
	g.GET("/users", h.listUsers)
}

func (h *Handler) listUsers(c echo.Context) error {
	users, err := h.svc.ListUsers(c.Request().Context())
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "users retrieved", users)
}

func (h *Handler) register(c echo.Context) error {
	ctx := c.Request().Context()

	bootstrap, err := h.svc.NeedsBootstrap(ctx)
	if err != nil {
		return err
	}

	if !bootstrap {
		claims, err := extractClaims(c, h.svc)
		if err != nil {
			return err
		}
		if claims.Role != RoleAdmin {
			return apperr.Forbidden("only an admin can register new users")
		}
	}

	var in RegisterInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}

	user, err := h.svc.Register(ctx, in, bootstrap)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusCreated, "user registered", user)
}

func (h *Handler) login(c echo.Context) error {
	var in LoginInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	pair, user, err := h.svc.Login(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "login successful", map[string]any{"tokens": pair, "user": user})
}

func (h *Handler) refresh(c echo.Context) error {
	var in RefreshInput
	if err := c.Bind(&in); err != nil {
		return apperr.Validation("invalid request body")
	}
	if in.RefreshToken == "" {
		return apperr.Validation("refresh_token is required")
	}
	pair, user, err := h.svc.Refresh(c.Request().Context(), in.RefreshToken)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "token refreshed", map[string]any{"tokens": pair, "user": user})
}

func (h *Handler) me(c echo.Context) error {
	claims, ok := ClaimsFromContext(c.Request().Context())
	if !ok {
		return apperr.Forbidden("authentication required")
	}
	user, err := h.svc.Me(c.Request().Context(), claims.UserID)
	if err != nil {
		return err
	}
	return response.OK(c, http.StatusOK, "current user", user)
}
