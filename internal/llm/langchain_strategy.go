package llm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	maigo "github.com/jeanmolossi/maigo/pkg/maigo"
	maigocontracts "github.com/jeanmolossi/maigo/pkg/maigo/contracts"
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
	embedderFactory  func(model string) (embeddings.Embedder, error)
	streamingEnabled bool
	chatModels       map[string]struct{}
	embedModels      map[string]struct{}
	requestTimeout   time.Duration
	maxRetries       int
}

// NewLangChainStrategy builds a LangChain-backed strategy using provider settings.
func NewLangChainStrategy(cfg ProviderSettings) (ModelStrategy, error) {
	return newOpenAIStrategy(cfg)
}

func newOpenAIStrategy(cfg ProviderSettings) (ModelStrategy, error) {
	timeout := normalizeTimeout(cfg.RequestTimeout)
	retries := normalizeRetries(cfg.MaxRetries)

	httpClient := newMaiGoHTTPClient(cfg.BaseURL, timeout)

	opts := []openai.Option{
		openai.WithToken(cfg.APIKey),
		openai.WithBaseURL(trimURL(cfg.BaseURL)),
		openai.WithHTTPClient(httpClient),
	}

	if len(cfg.ChatModels) > 0 {
		opts = append(opts, openai.WithModel(cfg.ChatModels[0]))
	}

	client, err := openai.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("llm: unable to create openai strategy: %w", err)
	}

	return &LangChainStrategy{
		name: cfg.Name,
		llm:  client,
		embedderFactory: func(model string) (embeddings.Embedder, error) {
			selectedModel := model
			if selectedModel == "" && len(cfg.EmbedModels) > 0 {
				selectedModel = cfg.EmbedModels[0]
			}

			embedOpts := []openai.Option{
				openai.WithToken(cfg.APIKey),
				openai.WithBaseURL(trimURL(cfg.BaseURL)),
				openai.WithHTTPClient(httpClient),
			}

			if selectedModel != "" {
				embedOpts = append(embedOpts, openai.WithEmbeddingModel(selectedModel))
			}

			embedClient, err := openai.New(embedOpts...)
			if err != nil {
				return nil, fmt.Errorf("llm: unable to create openai embedder: %w", err)
			}

			return embeddings.NewEmbedder(embeddings.EmbedderClientFunc(embedClient.CreateEmbedding))
		},
		streamingEnabled: cfg.EnableStreaming,
		chatModels:       toSet(cfg.ChatModels),
		embedModels:      toSet(cfg.EmbedModels),
		requestTimeout:   timeout,
		maxRetries:       retries,
	}, nil
}

func newAnthropicStrategy(cfg ProviderSettings) (ModelStrategy, error) {
	timeout := normalizeTimeout(cfg.RequestTimeout)
	retries := normalizeRetries(cfg.MaxRetries)

	httpClient := newMaiGoHTTPClient(cfg.BaseURL, timeout)

	opts := []anthropic.Option{anthropic.WithToken(cfg.APIKey), anthropic.WithHTTPClient(httpClient)}

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
		requestTimeout:   timeout,
		maxRetries:       retries,
	}, nil
}

func newOllamaStrategy(cfg ProviderSettings) (ModelStrategy, error) {
	timeout := normalizeTimeout(cfg.RequestTimeout)
	retries := normalizeRetries(cfg.MaxRetries)

	httpClient := newMaiGoHTTPClient(cfg.BaseURL, timeout)

	opts := []ollama.Option{ollama.WithServerURL(trimURL(cfg.BaseURL)), ollama.WithHTTPClient(httpClient)}

	if len(cfg.ChatModels) > 0 {
		opts = append(opts, ollama.WithModel(cfg.ChatModels[0]))
	}

	client, err := ollama.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("llm: unable to create ollama strategy: %w", err)
	}

	return &LangChainStrategy{
		name: cfg.Name,
		llm:  client,
		embedderFactory: func(model string) (embeddings.Embedder, error) {
			selectedModel := model
			if selectedModel == "" && len(cfg.EmbedModels) > 0 {
				selectedModel = cfg.EmbedModels[0]
			}

			embedOpts := []ollama.Option{ollama.WithServerURL(trimURL(cfg.BaseURL)), ollama.WithHTTPClient(httpClient)}

			if selectedModel != "" {
				embedOpts = append(embedOpts, ollama.WithModel(selectedModel))
			}

			embedClient, err := ollama.New(embedOpts...)
			if err != nil {
				return nil, fmt.Errorf("llm: unable to create ollama embedder: %w", err)
			}

			return embeddings.NewEmbedder(embeddings.EmbedderClientFunc(embedClient.CreateEmbedding))
		},
		streamingEnabled: cfg.EnableStreaming,
		chatModels:       toSet(cfg.ChatModels),
		embedModels:      toSet(cfg.EmbedModels),
		requestTimeout:   timeout,
		maxRetries:       retries,
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

	var resp *llms.ContentResponse

	err := s.callWithRetry(ctx, func(callCtx context.Context) error {
		r, err := s.llm.GenerateContent(callCtx, messages, opts...)
		if err != nil {
			return err
		}

		resp = r

		return nil
	})
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

	if s.embedderFactory == nil {
		return EmbedResponse{}, errors.New("llm: embeddings not supported by provider")
	}

	embedder, err := s.embedderFactory(req.Model)
	if err != nil {
		return EmbedResponse{}, err
	}

	var vectors [][]float32

	err = s.callWithRetry(ctx, func(callCtx context.Context) error {
		vectors, err = embedder.EmbedDocuments(callCtx, req.Input)
		return err
	})
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

func (s *LangChainStrategy) callWithRetry(ctx context.Context, fn func(ctx context.Context) error) error {
	retries := s.maxRetries + 1

	var err error

	for attempt := 0; attempt < retries; attempt++ {
		callCtx, cancel := context.WithTimeout(ctx, s.requestTimeout)

		err = fn(callCtx)

		cancel()

		if err == nil {
			return nil
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}

		if attempt < retries-1 {
			time.Sleep(time.Duration(attempt+1) * 150 * time.Millisecond)
		}
	}

	return err
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

func newMaiGoHTTPClient(baseURL string, timeout time.Duration) *http.Client {
	trimmed := trimURL(baseURL)

	if trimmed == "" {
		return &http.Client{Timeout: timeout}
	}

	client := maigo.NewClient(trimmed).Config().SetTimeout(timeout).Build()

	config, ok := client.(maigocontracts.ClientConfig)
	if !ok {
		return &http.Client{Timeout: timeout}
	}

	httpCfg := config.HttpClient()

	return &http.Client{
		Timeout:   httpCfg.Timeout(),
		Transport: httpCfg.Transport(),
	}
}

func trimURL(url string) string {
	return strings.TrimRight(url, "/")
}

func normalizeTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return 30 * time.Second
	}

	return timeout
}

func normalizeRetries(retries int) int {
	if retries < 0 {
		return 0
	}

	return retries
}
