package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/audit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/cache"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/guardrails"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/llm"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/ratelimit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/router"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/usage"
)

type ChatHandler struct {
	Tenants    tenant.Store
	Router     router.Router
	Cache      cache.SemanticCache
	Guardrails guardrails.Guardrails
	UsagePub   usage.Publisher
	AuditPub   audit.Publisher
	Limiter    ratelimit.Limiter
}

type chatRequestDTO struct {
	Model       string            `json:"model"`
	Messages    []llm.ChatMessage `json:"messages"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float32           `json:"temperature,omitempty"`
	TopP        float32           `json:"top_p,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	if ok, _ := h.Limiter.Allow(ctx, tcfg.ID, "chat", 1); !ok {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	var dto chatRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
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

	// Guardrails pré
	req, err = h.Guardrails.PreProcessChat(ctx, *tcfg, req)
	if err != nil {
		http.Error(w, "blocked by guardrails", http.StatusForbidden)
		return
	}

	// Cache
	if tcfg.EnableSemanticCache {
		if cached, ok, _ := h.Cache.LookupChat(ctx, tcfg.ID, req); ok {
			writeChatResponse(w, *cached)
			_ = h.UsagePub.Publish(ctx, usage.Event{
				TenantID: tcfg.ID,
				Model:    cached.Model,
				// tokens podem ser estimados
			})

			return
		}
	}

	// Routing + chamada ao modelo
	resp, err := h.Router.RouteChat(ctx, *tcfg, req)
	if err != nil {
		http.Error(w, "model error", http.StatusBadGateway)
		return
	}

	// Guardrails pós
	resp, _ = h.Guardrails.PostProcessChat(ctx, *tcfg, req, resp)

	// Async events
	_ = h.UsagePub.Publish(ctx, usage.Event{
		TenantID:         tcfg.ID,
		Model:            resp.Model,
		PromptTokens:     resp.PromptTok,
		CompletionTokens: resp.CompTok,
	})

	_ = h.AuditPub.Publish(ctx, audit.Event{
		TenantID:  tcfg.ID,
		Model:     resp.Model,
		RequestID: "", // pode vir do header ou gerar
		Prompt:    req.Messages,
		Response:  resp.Messages,
	})

	if tcfg.EnableSemanticCache {
		_ = h.Cache.StoreChat(ctx, tcfg.ID, req, resp)
	}

	writeChatResponse(w, resp)
}

func writeChatResponse(w http.ResponseWriter, resp llm.ChatResponse) {
	type choice struct {
		Index   int             `json:"index"`
		Message llm.ChatMessage `json:"message"`
	}

	out := struct {
		ID    string `json:"id"`
		Model string `json:"model"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
		Choices []choice `json:"choices"`
	}{
		ID:    resp.ID,
		Model: resp.Model,
		Choices: []choice{
			{
				Index:   0,
				Message: resp.Messages[len(resp.Messages)-1],
			},
		},
	}
	out.Usage.PromptTokens = resp.PromptTok
	out.Usage.CompletionTokens = resp.CompTok
	out.Usage.TotalTokens = resp.PromptTok + resp.CompTok

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
