package auth

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/ranji/clothing-erp/internal/apperr"
)

type claimsCtxKey struct{}

// RequireAuth verifies the Bearer access token and attaches its claims to
// the request context. It must run before any RequireRole middleware.
func RequireAuth(svc *Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, err := extractClaims(c, svc)
			if err != nil {
				return err
			}
			ctx := context.WithValue(c.Request().Context(), claimsCtxKey{}, claims)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func extractClaims(c echo.Context, svc *Service) (*Claims, error) {
	header := c.Request().Header.Get("Authorization")
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		return nil, apperr.Forbidden("missing bearer token")
	}
	tokenString := strings.TrimPrefix(header, "Bearer ")
	return svc.ValidateAccessToken(tokenString)
}

// RequireRole must run after RequireAuth. It rejects the request unless the
// authenticated user's role is one of the allowed roles.
func RequireRole(roles ...Role) echo.MiddlewareFunc {
	allowed := make(map[Role]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := ClaimsFromContext(c.Request().Context())
			if !ok {
				return apperr.Forbidden("authentication required")
			}
			if !allowed[claims.Role] {
				return apperr.Forbidden("your role does not have access to this action")
			}
			return next(c)
		}
	}
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsCtxKey{}).(*Claims)
	return claims, ok
}
