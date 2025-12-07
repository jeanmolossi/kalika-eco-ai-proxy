package http

import (
	"net/http"

	tenantapp "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant/app"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	Store tenantapp.Store
}

func NewHandlers(store tenantapp.Store) *Handlers {
	return &Handlers{Store: store}
}

func (h *Handlers) GetByAPIKey(ctx echo.Context) error {
	apiKey := ctx.Param("apiKey")

	tenant, err := h.Store.FindByAPIKey(ctx.Request().Context(), apiKey)
	if err != nil {
		switch err {
		case tenantapp.ErrNotFound:
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case tenantapp.ErrInvalidAPIKey:
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		case tenantapp.ErrInactiveTenant:
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		default:
			return err
		}
	}

	return ctx.JSON(http.StatusOK, tenant)
}

func (h *Handlers) GetByTenantID(ctx echo.Context) error {
	tenantID := ctx.Param("tenantID")

	tenant, err := h.Store.FindByID(ctx.Request().Context(), tenantID)
	if err != nil {
		switch err {
		case tenantapp.ErrNotFound:
			return echo.NewHTTPError(http.StatusNotFound)
		case tenantapp.ErrInactiveTenant:
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		default:
			return err
		}
	}

	return ctx.JSON(http.StatusOK, tenant)
}
