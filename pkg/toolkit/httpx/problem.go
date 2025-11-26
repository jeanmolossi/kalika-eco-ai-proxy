package httpx

import (
	"fmt"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/apperr"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
	"github.com/labstack/echo/v4"
)

type ProblemResponse struct {
	Type     string              `json:"type"`
	Title    string              `json:"title"`
	Status   int                 `json:"status"`
	Detail   string              `json:"detail"`
	Instance string              `json:"instance,omitempty"`
	Code     string              `json:"code,omitempty"`
	Fields   []apperr.FieldError `json:"fields,omitempty"`
	Meta     map[string]any      `json:"meta,omitempty"`
	Cause    string              `json:"cause,omitempty"`
}

// WriteProblem transforms any error into problem+json.
func WriteProblem(c echo.Context, err error) error {
	appErr := apperr.From(err) // ensures falling into *apperr.Error

	// get trace_id/request_id from the context/log middleware
	traceID := c.Response().Header().Get("X-Trace-Id")
	if traceID == "" {
		traceID = c.Response().Header().Get(echo.HeaderXRequestID)
	}

	if traceID == "" {
		traceID = c.Request().Header.Get("X-Trace-Id")
	}

	if traceID == "" {
		traceID = c.Request().Header.Get(echo.HeaderXRequestID)
	}

	var cause string
	if appErr.Unwrap() != nil {
		cause = appErr.Unwrap().Error()
	}

	// You can build "type" dynamically based on Kind:
	problem := ProblemResponse{
		Type:     docsBaseURL(string(appErr.Kind())),
		Title:    appErr.KindTitle(), // we'll see this below
		Status:   appErr.HTTPStatus(),
		Detail:   appErr.Detail(), // getter that you add in apperr
		Instance: traceID,
		Code:     appErr.Code(),
		Fields:   appErr.Fields(),
		Meta:     appErr.Meta(), // getter that you add
		Cause:    cause,
	}

	// correct content-type
	return c.JSON(appErr.HTTPStatus(), problem)
}

func docsBaseURL(code string) string {
	cfg := config.Load() // cached config

	scheme := "http://"
	if cfg.Server.Port == 443 {
		scheme = "https://"
	}

	port := fmt.Sprintf(":%d", cfg.Server.Port)
	baseURL := scheme + cfg.Server.Host + port + cfg.Server.BasePath
	docsURL := baseURL + "/docs/errors/"

	return docsURL + code
}
