package http

import (
	"encoding/json"
	"net/http"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/modules/aiproxy/app"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/httpx"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"
	"github.com/labstack/echo/v4"
)

func (h *Handlers) Embeddings(c echo.Context) error {
	r := c.Request()
	w := c.Response()

	ctx := r.Context()

	apiKey := extractAPIKey(r.Header.Get("Authorization"))
	if apiKey == "" {
		http.Error(w, "missing api key", http.StatusUnauthorized)
		return nil
	}

	tcfg, err := h.Tenants.FindByAPIKey(apiKey)
	if err != nil || tcfg == nil {
		http.Error(w, "invalid api key", http.StatusUnauthorized)
		return nil
	}

	var dto embedRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return nil
	}

	req := llm.EmbedRequest{
		Model: dto.Model,
		Input: dto.Input,
	}

	out, err := h.EmbeddingsUseCase.Embeddings(ctx, app.EmbeddingsInput{
		UserID:  "uiser-id",
		Request: req,
		Tenant:  *tcfg,
	})
	if err != nil {
		return httpx.WriteProblem(c, err)
	}

	return c.JSON(http.StatusOK, out)
}
