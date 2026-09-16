package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/ranji/clothing-erp/internal/auth"
	dbpkg "github.com/ranji/clothing-erp/internal/db"
)

// Actor stamps the request context with the identity to use for audit and
// accounting created_by/changed_by fields. When RequireAuth has already run
// (the normal case), it uses the authenticated user's email; otherwise it
// falls back to "system" for the few public routes (login, bootstrap
// register, health).
func Actor() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			actor := "system"
			if claims, ok := auth.ClaimsFromContext(c.Request().Context()); ok {
				actor = claims.Email
			}
			ctx := dbpkg.WithActor(c.Request().Context(), actor)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
