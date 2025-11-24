package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/modules/aiproxy/app"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/apperr"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/httpx"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/router"
	"github.com/labstack/echo/v4"
)

func (h *Handlers) Embeddings(c echo.Context) error {
	r := c.Request()

	ctx := r.Context()

	apiKey := extractAPIKey(r.Header.Get("Authorization"))
	if apiKey == "" {
		return httpx.WriteProblem(c, apperr.Unauthorized(errors.New("missing api key")))
	}

	tcfg, err := h.Tenants.FindByAPIKey(ctx, apiKey)
	if err != nil {
		return httpx.WriteProblem(c, apperr.Unauthorized(err))
	}

	if tcfg == nil {
		return httpx.WriteProblem(c, apperr.Unauthorized(errors.New("missing config")))
	}

	var dto embedRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return httpx.WriteProblem(c, apperr.BadRequest(err))
	}

	model, err := router.ResolveEmbedModel(*tcfg, dto.Model)
	if err != nil {
		return httpx.WriteProblem(c, apperr.BadRequest(err))
	}

	req := llm.EmbedRequest{
		Model: model,
		Input: dto.Input,
	}

	tokenCount, err := h.Tokenizr.CountEmbeddingTokens(model, dto.Input)
	if err != nil {
		return httpx.WriteProblem(c, err)
	}

	limited, err := h.Limiter.Allow(ctx, tcfg.ID, "embeddings", tokenCount)
	if err != nil {
		return httpx.WriteProblem(c, err)
	}

	httpx.SetRateLimitHeaders(c.Response().Header(), limited, time.Now())

	if !limited.Allowed {
		return httpx.WriteProblem(c, apperr.TooManyRequests(limited.AsError()))
	}

	out, err := h.EmbeddingsUseCase.Embeddings(ctx, app.EmbeddingsInput{
		Request:  req,
		Tenant:   *tcfg,
		Metadata: dto.Metadata,
	})
	if err != nil {
		return httpx.WriteProblem(c, err)
	}

	return c.JSON(http.StatusOK, out)
}
