package http

import (
	"net/http"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant"
	"github.com/labstack/echo/v4"
)

// Handlers exposes tenant endpoints.
type Handlers struct {
	Store tenant.Store
}

// NewHandlers creates a new tenant handler set.
func NewHandlers(store tenant.Store) *Handlers {
	return &Handlers{Store: store}
}

// GetTenantByID returns tenant details for the provided tenant ID.
func (h *Handlers) GetTenantByID(ctx echo.Context) error {
	tenantID := ctx.Param("tenantID")

	tenantData, err := h.Store.FindByID(ctx.Request().Context(), tenantID)
	if err != nil {
		if err == tenant.ErrNotFound {
			return ctx.NoContent(http.StatusNotFound)
		}

		return err
	}

	return ctx.JSON(http.StatusOK, tenantData)
}

// GetTenantByAPIKey returns tenant details based on an API key.
func (h *Handlers) GetTenantByAPIKey(ctx echo.Context) error {
	apiKey := ctx.Param("apiKey")

	tenantData, err := h.Store.FindByAPIKey(ctx.Request().Context(), apiKey)
	if err != nil {
		switch err {
		case tenant.ErrInvalidAPIKey, tenant.ErrNotFound, tenant.ErrInactiveTenant:
			return ctx.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	return ctx.JSON(http.StatusOK, tenantData)
}
