package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/modules/aiproxy/app"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/router"
	pkgllm "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/llm"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/apperr"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/httpx"
	"github.com/labstack/echo/v4"
)

func (h *Handlers) ChatCompletions(c echo.Context) error {
	r := c.Request()
	ctx := r.Context()

	apiKey := extractAPIKey(r.Header.Get("Authorization"))
	if apiKey == "" {
		return httpx.WriteProblem(c, apperr.Unauthorized(errors.New("missing api key")))
	}

	tcfg, err := h.Tenants.FindByAPIKey(ctx, apiKey)
	if err != nil || tcfg == nil {
		return httpx.WriteProblem(c, apperr.Unauthorized(errors.New("invalid api key")))
	}

	var dto chatRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return httpx.WriteProblem(c, apperr.BadRequest(errors.New("invalid body")))
	}

	model, err := router.ResolveChatModel(*tcfg, dto.Model)
	if err != nil {
		return httpx.WriteProblem(c, apperr.BadRequest(err))
	}

	req := pkgllm.ChatRequest{
		Model:       model,
		Messages:    dto.Messages,
		MaxTokens:   dto.MaxTokens,
		Temperature: dto.Temperature,
		TopP:        dto.TopP,
		TenantID:    tcfg.ID,
		Extras:      dto.Metadata,
	}

	tokenCount, err := h.Tokenizr.CountChatTokens(model, dto.Messages)
	if err != nil {
		return httpx.WriteProblem(c, err)
	}

	res, err := h.Limiter.Allow(ctx, tcfg.ID, "chat", tokenCount)
	if err != nil {
		return httpx.WriteProblem(c, err)
	}

	httpx.SetRateLimitHeaders(c.Response().Header(), res, time.Now())

	if !res.Allowed {
		return httpx.WriteProblem(c, apperr.TooManyRequests(res.AsError()))
	}

	out, err := h.ChatUseCase.Chat(ctx, app.ChatInput{
		Request:  req,
		Tenant:   *tcfg,
		APIKey:   apiKey,
		Metadata: dto.Metadata,
	})
	if err != nil {
		return httpx.WriteProblem(c, err)
	}

	return c.JSON(http.StatusOK, out)
}
