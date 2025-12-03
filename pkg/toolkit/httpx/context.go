package httpx

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type requestID struct{}

func RequestIDFromCtx(ctx context.Context) string {
	val, ok := ctx.Value(requestID{}).(string)
	if !ok {
		return ""
	}

	return val
}

func SetRequestID(ctx context.Context, rID string) context.Context {
	return context.WithValue(ctx, requestID{}, rID)
}

func RequestID() echo.MiddlewareFunc {
	return middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Skipper:      middleware.DefaultSkipper,
		Generator:    middleware.DefaultRequestIDConfig.Generator,
		TargetHeader: middleware.DefaultRequestIDConfig.TargetHeader,
		RequestIDHandler: func(c echo.Context, rid string) {
			req := c.Request()
			ctx := SetRequestID(req.Context(), rid)
			c.SetRequest(req.WithContext(ctx))
		},
	})
}
