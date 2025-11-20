package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/llm"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/router"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant"
)

type EmbeddingsHandler struct {
	Tenants tenant.Store
	Router  router.Router
}

type embedRequestDTO struct {
	Model    string            `json:"model"`
	Input    []string          `json:"input"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type embedVectorDTO struct {
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

type embedResponseDTO struct {
	Model string           `json:"model"`
	Data  []embedVectorDTO `json:"data"`
}

// ServeHTTP handles the embeddings request and forwards it through the router.
func (h *EmbeddingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	apiKey := extractAPIKey(r.Header.Get("Authorization"))
	if apiKey == "" {
		http.Error(w, "missing api key", http.StatusUnauthorized)
		return
	}

	tcfg, err := h.Tenants.FindByAPIKey(apiKey)
	if err != nil || tcfg == nil {
		http.Error(w, "invalid api key", http.StatusUnauthorized)
		return
	}

	var dto embedRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	req := llm.EmbedRequest{
		Model: dto.Model,
		Input: dto.Input,
	}

	resp, err := h.Router.RouteEmbed(ctx, *tcfg, req)
	if err != nil {
		http.Error(w, "model error", http.StatusBadGateway)
		return
	}

	out := embedResponseDTO{
		Model: resp.Model,
		Data:  make([]embedVectorDTO, len(resp.Embeddings)),
	}

	for i, emb := range resp.Embeddings {
		out.Data[i] = embedVectorDTO{
			Index:     i,
			Embedding: emb,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
