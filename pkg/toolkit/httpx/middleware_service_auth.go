package httpx

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// ServiceAuthMiddleware enforces a shared secret for inter-service HTTP calls.
func ServiceAuthMiddleware(token string) echo.MiddlewareFunc {
	expected := strings.TrimSpace(token)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if expected == "" {
				return next(c)
			}

			if strings.Contains(c.Path(), "healthz") {
				return next(c)
			}

			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "Bearer "+expected {
				return next(c)
			}

			if c.Request().Header.Get("X-Service-Token") == expected {
				return next(c)
			}

			return echo.NewHTTPError(http.StatusUnauthorized, "service authentication failed")
		}
	}
}
