package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// ModelStrategy defines how we interact with a concrete LLM/embedding backend.
type ModelStrategy interface {
	Name() string
	Supports(model string) bool
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
	Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error)
}

// LangChainStrategy relies on langchain-go to speak with OpenAI-compatible models.
type LangChainStrategy struct {
	name             string
	llm              llms.Model
	embedder         embeddings.Embedder
	streamingEnabled bool
	chatModels       map[string]struct{}
	embedModels      map[string]struct{}
}

// NewLangChainStrategy builds a LangChain-backed strategy using provider settings.
func NewLangChainStrategy(cfg ProviderSettings) (ModelStrategy, error) {
	return newOpenAIStrategy(cfg)
}

func newOpenAIStrategy(cfg ProviderSettings) (ModelStrategy, error) {
	opts := []openai.Option{
		openai.WithToken(cfg.APIKey),
		openai.WithBaseURL(trimURL(cfg.BaseURL)),
	}

	if len(cfg.ChatModels) > 0 {
		opts = append(opts, openai.WithModel(cfg.ChatModels[0]))
	}

	if len(cfg.EmbedModels) > 0 {
		opts = append(opts, openai.WithEmbeddingModel(cfg.EmbedModels[0]))
	}

	client, err := openai.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("llm: unable to create openai strategy: %w", err)
	}

	embedder, err := embeddings.NewEmbedder(embeddings.EmbedderClientFunc(client.CreateEmbedding))
	if err != nil {
		return nil, fmt.Errorf("llm: unable to create openai embedder: %w", err)
	}

	return &LangChainStrategy{
		name:             cfg.Name,
		llm:              client,
		embedder:         embedder,
		streamingEnabled: cfg.EnableStreaming,
		chatModels:       toSet(cfg.ChatModels),
		embedModels:      toSet(cfg.EmbedModels),
	}, nil
}

func newAnthropicStrategy(cfg ProviderSettings) (ModelStrategy, error) {
	opts := []anthropic.Option{anthropic.WithToken(cfg.APIKey)}

	if trimmed := trimURL(cfg.BaseURL); trimmed != "" {
		opts = append(opts, anthropic.WithBaseURL(trimmed))
	}

	if len(cfg.ChatModels) > 0 {
		opts = append(opts, anthropic.WithModel(cfg.ChatModels[0]))
	}

	client, err := anthropic.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("llm: unable to create anthropic strategy: %w", err)
	}

	return &LangChainStrategy{
		name:             cfg.Name,
		llm:              client,
		streamingEnabled: cfg.EnableStreaming,
		chatModels:       toSet(cfg.ChatModels),
		embedModels:      toSet(cfg.EmbedModels),
	}, nil
}

func newOllamaStrategy(cfg ProviderSettings) (ModelStrategy, error) {
	opts := []ollama.Option{ollama.WithServerURL(trimURL(cfg.BaseURL))}

	if len(cfg.ChatModels) > 0 {
		opts = append(opts, ollama.WithModel(cfg.ChatModels[0]))
	}

	if len(cfg.EmbedModels) > 0 {
		opts = append(opts, ollama.WithModel(cfg.EmbedModels[0]))
	}

	client, err := ollama.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("llm: unable to create ollama strategy: %w", err)
	}

	embedder, err := embeddings.NewEmbedder(embeddings.EmbedderClientFunc(client.CreateEmbedding))
	if err != nil {
		return nil, fmt.Errorf("llm: unable to create ollama embedder: %w", err)
	}

	return &LangChainStrategy{
		name:             cfg.Name,
		llm:              client,
		embedder:         embedder,
		streamingEnabled: cfg.EnableStreaming,
		chatModels:       toSet(cfg.ChatModels),
		embedModels:      toSet(cfg.EmbedModels),
	}, nil
}

// Name returns the configured provider name.
func (s *LangChainStrategy) Name() string {
	return s.name
}

// Supports checks whether the given model is allowed for the strategy.
func (s *LangChainStrategy) Supports(model string) bool {
	if model == "" {
		return true
	}

	if len(s.chatModels) == 0 && len(s.embedModels) == 0 {
		return true
	}

	_, okChat := s.chatModels[model]
	_, okEmbed := s.embedModels[model]

	return okChat || okEmbed
}

// Chat sends the request using langchain-go and maps it back to the internal format.
func (s *LangChainStrategy) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if !s.Supports(req.Model) {
		return ChatResponse{}, fmt.Errorf("llm: model %s not supported by strategy %s", req.Model, s.name)
	}

	messages := make([]llms.MessageContent, 0, len(req.Messages))
	for _, msg := range req.Messages {
		messages = append(messages, llms.MessageContent{
			Role:  roleToLangChain(msg.Role),
			Parts: []llms.ContentPart{llms.TextPart(msg.Content)},
		})
	}

	opts := s.buildCallOptions(req)

	var streamed strings.Builder

	if req.Stream {
		if !s.streamingEnabled {
			return ChatResponse{}, errors.New("llm: streaming disabled for provider")
		}

		opts = append(opts, llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			streamed.Write(chunk)
			return nil
		}))
	}

	resp, err := s.llm.GenerateContent(ctx, messages, opts...)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("llm: chat failed: %w", err)
	}

	content := streamed.String()
	if content == "" {
		if resp == nil || len(resp.Choices) == 0 || resp.Choices[0] == nil {
			return ChatResponse{}, errors.New("llm: empty chat response")
		}

		content = resp.Choices[0].Content
	}

	assistant := ChatMessage{Role: RoleAssistant, Content: content}

	return ChatResponse{
		Model:     req.Model,
		Messages:  append(req.Messages, assistant),
		PromptTok: extractUsage(resp, "prompt_tokens"),
		CompTok:   extractUsage(resp, "completion_tokens"),
	}, nil
}

// Embed creates embeddings using the langchain-go embedder.
func (s *LangChainStrategy) Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error) {
	if !s.Supports(req.Model) {
		return EmbedResponse{}, fmt.Errorf("llm: model %s not supported by strategy %s", req.Model, s.name)
	}

	if s.embedder == nil {
		return EmbedResponse{}, errors.New("llm: embeddings not supported by provider")
	}

	vectors, err := s.embedder.EmbedDocuments(ctx, req.Input)
	if err != nil {
		return EmbedResponse{}, fmt.Errorf("llm: embed failed: %w", err)
	}

	return EmbedResponse{Model: req.Model, Embeddings: vectors}, nil
}

func (s *LangChainStrategy) buildCallOptions(req ChatRequest) []llms.CallOption {
	opts := []llms.CallOption{}

	if req.Model != "" {
		opts = append(opts, llms.WithModel(req.Model))
	}

	if req.MaxTokens > 0 {
		opts = append(opts, llms.WithMaxTokens(req.MaxTokens))
	}

	if req.Temperature > 0 {
		opts = append(opts, llms.WithTemperature(float64(req.Temperature)))
	}

	if req.TopP > 0 {
		opts = append(opts, llms.WithTopP(float64(req.TopP)))
	}

	return opts
}

func roleToLangChain(role string) llms.ChatMessageType {
	switch role {
	case RoleAssistant:
		return llms.ChatMessageTypeAI
	case RoleUser:
		return llms.ChatMessageTypeHuman
	default:
		return llms.ChatMessageTypeSystem
	}
}

func toSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, v := range values {
		if v != "" {
			set[v] = struct{}{}
		}
	}

	return set
}

func extractUsage(resp *llms.ContentResponse, key string) int {
	if resp == nil || len(resp.Choices) == 0 || resp.Choices[0] == nil {
		return 0
	}

	v, ok := resp.Choices[0].GenerationInfo[key]
	if !ok {
		return 0
	}

	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	default:
		return 0
	}
}

func trimURL(url string) string {
	return strings.TrimRight(url, "/")
}
