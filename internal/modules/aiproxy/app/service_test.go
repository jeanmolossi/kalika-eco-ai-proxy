package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/audit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/httpx"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/ratelimit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/usage"
	pkgguardrails "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/guardrails"
	pkgllm "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/llm"
	pkgtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant"
)

func TestChatRateLimitAndGuardrails(t *testing.T) {
	ctx := httpx.SetRequestID(context.Background(), "rid-test")
	guard := &fakeGuardrails{}
	limiter := &fakeLimiter{result: ratelimit.Result{Allowed: false}}
	rt := &fakeRouter{chatResp: pkgllm.ChatResponse{
		Model:     "gpt-4o",
		Messages:  []pkgllm.ChatMessage{{Role: pkgllm.RoleAssistant, Content: "ok"}},
		PromptTok: 2,
		CompTok:   1,
	}}
	up := &captureUsage{}
	ap := &captureAudit{}
	tok := &fixedTokenizer{}

	svc := NewService(limiter, nil, guard, rt, up, ap, tok)

	_, err := svc.Chat(ctx, ChatInput{
		Tenant:  pkgtenant.TenantConfig{ID: "t1", ParsedPolicyConfig: &pkgtenant.PolicyConfig{ModelsAllowed: []string{"gpt-4o"}}},
		Request: pkgllm.ChatRequest{Model: "gpt-4o", Messages: []pkgllm.ChatMessage{{Role: pkgllm.RoleUser, Content: "hi"}}},
	})
	if err == nil {
		t.Fatalf("expected rate limit error")
	}

	limiter.result = ratelimit.Result{Allowed: true}
	guard.inputDecision = pkgguardrails.Decision{Action: pkgguardrails.ActionAllow}
	guard.outputDecision = pkgguardrails.Decision{Action: pkgguardrails.ActionAllow}

	resp, err := svc.Chat(ctx, ChatInput{
		Tenant:  pkgtenant.TenantConfig{ID: "t1", ParsedPolicyConfig: &pkgtenant.PolicyConfig{ModelsAllowed: []string{"gpt-4o"}}},
		Request: pkgllm.ChatRequest{Model: "gpt-4o", Messages: []pkgllm.ChatMessage{{Role: pkgllm.RoleUser, Content: "hi"}}},
	})
	if err != nil {
		t.Fatalf("chat should succeed: %v", err)
	}

	if resp.Model != "gpt-4o" {
		t.Fatalf("unexpected model %s", resp.Model)
	}

	if len(up.events) != 1 {
		t.Fatalf("usage not published")
	}

	if up.events[0].RequestID != "rid-test" {
		t.Fatalf("request id not propagated")
	}

	if len(ap.events) != 1 {
		t.Fatalf("audit not published")
	}
}

func TestEmbeddingsRateLimit(t *testing.T) {
	ctx := context.Background()
	limiter := &fakeLimiter{result: ratelimit.Result{Allowed: true}}
	rt := &fakeRouter{embedResp: pkgllm.EmbedResponse{Model: "emb-1", Embeddings: [][]float32{{1, 2, 3}}}}
	up := &captureUsage{}
	ap := &captureAudit{}
	tok := &fixedTokenizer{embTokens: 4}
	svc := NewService(limiter, nil, &fakeGuardrails{}, rt, up, ap, tok)

	_, err := svc.Embeddings(ctx, EmbeddingsInput{
		Tenant:  pkgtenant.TenantConfig{ID: "t1"},
		Request: pkgllm.EmbedRequest{Model: "emb-1", Input: []string{"txt"}},
	})
	if err != nil {
		t.Fatalf("embeddings failed: %v", err)
	}

	if len(up.events) != 1 || up.events[0].PromptTokens != 4 {
		t.Fatalf("usage not recorded with tokens")
	}

	limiter.result = ratelimit.Result{Allowed: false}

	_, err = svc.Embeddings(ctx, EmbeddingsInput{
		Tenant:  pkgtenant.TenantConfig{ID: "t1"},
		Request: pkgllm.EmbedRequest{Model: "emb-1", Input: []string{"txt"}},
	})
	if err == nil {
		t.Fatalf("expected rate limit on embeddings")
	}
}

type fakeLimiter struct {
	result ratelimit.Result
	err    error
}

func (f *fakeLimiter) Allow(context.Context, string, string, int) (ratelimit.Result, error) {
	return f.result, f.err
}

type fakeGuardrails struct {
	inputDecision  pkgguardrails.Decision
	outputDecision pkgguardrails.Decision
}

func (f *fakeGuardrails) EvaluateInput(context.Context, pkgguardrails.Context) (pkgguardrails.Decision, error) {
	return f.inputDecision, nil
}

func (f *fakeGuardrails) EvaluateOutput(context.Context, pkgguardrails.Context) (pkgguardrails.Decision, error) {
	return f.outputDecision, nil
}

type fakeRouter struct {
	chatResp  pkgllm.ChatResponse
	embedResp pkgllm.EmbedResponse
}

func (f *fakeRouter) RouteChat(context.Context, pkgtenant.TenantConfig, pkgllm.ChatRequest) (pkgllm.ChatResponse, error) {
	if f.chatResp.ID == "" {
		f.chatResp.ID = uuid.NewString()
	}

	return f.chatResp, nil
}

func (f *fakeRouter) RouteEmbed(context.Context, pkgtenant.TenantConfig, pkgllm.EmbedRequest) (pkgllm.EmbedResponse, error) {
	return f.embedResp, nil
}

type captureUsage struct {
	events []usage.Event
}

func (c *captureUsage) Publish(_ context.Context, ev usage.Event) error {
	c.events = append(c.events, ev)
	return nil
}

type captureAudit struct {
	events []audit.Event
}

func (c *captureAudit) Publish(_ context.Context, ev audit.Event) error {
	c.events = append(c.events, ev)
	return nil
}

type fixedTokenizer struct {
	embTokens int
}

func (fixedTokenizer) CountChatTokens(string, []pkgllm.ChatMessage) (int, error) { return 2, nil }

func (f *fixedTokenizer) CountEmbeddingTokens(string, []string) (int, error) { return f.embTokens, nil }
