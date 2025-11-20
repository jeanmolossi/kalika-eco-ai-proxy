package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/modules/aiproxy/app"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/apperr"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/httpx"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"
	"github.com/labstack/echo/v4"
)

func (h *Handlers) ChatCompletions(c echo.Context) error {
	r := c.Request()
	ctx := r.Context()

	apiKey := extractAPIKey(r.Header.Get("Authorization"))
	if apiKey == "" {
		return httpx.WriteProblem(c, apperr.Unauthorized(errors.New("missin api key")))
	}

	tcfg, err := h.Tenants.FindByAPIKey(ctx, apiKey)
	if err != nil || tcfg == nil {
		return httpx.WriteProblem(c, apperr.Unauthorized(errors.New("invalid api key")))
	}

	var dto chatRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return httpx.WriteProblem(c, apperr.BadRequest(errors.New("invalid body")))
	}

	req := llm.ChatRequest{
		Model:       dto.Model,
		Messages:    dto.Messages,
		MaxTokens:   dto.MaxTokens,
		Temperature: dto.Temperature,
		TopP:        dto.TopP,
		TenantID:    tcfg.ID,
		Extras:      dto.Metadata,
	}

	out, err := h.ChatUseCase.Chat(ctx, app.ChatInput{
		Request:  req,
		Tenant:   *tcfg,
		Metadata: dto.Metadata,
	})
	if err != nil {
		return httpx.WriteProblem(c, err)
	}

	return c.JSON(http.StatusOK, out)
}
