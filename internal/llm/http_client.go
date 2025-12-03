package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	maigo "github.com/jeanmolossi/maigo/pkg/maigo"
	maigocontracts "github.com/jeanmolossi/maigo/pkg/maigo/contracts"
	"github.com/jeanmolossi/maigo/pkg/maigo/header"
	"github.com/jeanmolossi/maigo/pkg/maigo/mime"
)

type ProviderSettings struct {
	Name            string
	BaseURL         string
	APIKey          string
	RequestTimeout  time.Duration
	MaxRetries      int
	EnableStreaming bool
	ChatModels      []string
	EmbedModels     []string
}

// HTTPClient calls an upstream OpenAI-compatible HTTP endpoint.
// It supports basic retries and streaming via server-sent events.
type HTTPClient struct {
	client          maigocontracts.ClientHTTPMethods
	metrics         MetricsRecorder
	maxRetries      int
	enableStreaming bool
}

// NewHTTPClient builds an HTTP-backed LLM client using the provided settings.
func NewHTTPClient(settings ProviderSettings, metrics MetricsRecorder) (*HTTPClient, error) {
	if settings.BaseURL == "" {
		return nil, errors.New("llm: provider base url is required")
	}

	if metrics == nil {
		metrics = NoopMetrics{}
	}

	timeout := settings.RequestTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	retries := settings.MaxRetries
	if retries < 0 {
		retries = 0
	}

	builder := maigo.NewClient(strings.TrimRight(settings.BaseURL, "/"))
	builder.Config().SetTimeout(timeout)

	if settings.APIKey != "" {
		builder.Header().Set(header.Authorization, "Bearer "+settings.APIKey)
	}

	return &HTTPClient{
		client:          builder.Build(),
		metrics:         metrics,
		maxRetries:      retries,
		enableStreaming: settings.EnableStreaming,
	}, nil
}

// Chat sends a chat completion request to the provider.
func (c *HTTPClient) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	start := time.Now()
	resp, err := c.doRequest(ctx, "chat/completions", chatPayload(req))
	latency := time.Since(start)

	defer func() {
		c.metrics.ObserveChat(req.Model, latency, err)
	}()

	if err != nil {
		return ChatResponse{}, err
	}

	if req.Stream {
		return c.consumeStream(resp)
	}

	return parseChatResponse(resp)
}

// Embed sends an embeddings request to the provider.
func (c *HTTPClient) Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error) {
	start := time.Now()
	resp, err := c.doRequest(ctx, "embeddings", embedPayload(req))
	latency := time.Since(start)

	defer func() {
		c.metrics.ObserveEmbed(req.Model, latency, err)
	}()

	if err != nil {
		return EmbedResponse{}, err
	}

	return parseEmbedResponse(resp)
}

func (c *HTTPClient) doRequest(ctx context.Context, path string, body any) (*maigo.Response, error) {
	retries := c.maxRetries + 1

	var (
		resp *maigo.Response
		err  error
	)

	for attempt := 0; attempt < retries; attempt++ {
		resp, err = c.singleRequest(ctx, path, body)
		if err == nil && !resp.Status().Is5xxServerError() {
			return resp, nil
		}

		if resp != nil {
			resp.Body().Close()
		}

		if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			time.Sleep(time.Duration(attempt+1) * 150 * time.Millisecond)
			continue
		}
	}

	return resp, err
}

func (c *HTTPClient) singleRequest(ctx context.Context, path string, body any) (*maigo.Response, error) {
	req := c.client.POST(path).
		Context().Set(ctx).
		Header().AddContentType(mime.JSON).
		Body().AsJSON(body)

	respRaw, err := req.Send()
	if err != nil {
		return nil, err
	}

	resp, ok := respRaw.(*maigo.Response)
	if !ok {
		return nil, fmt.Errorf("llm: unexpected response type %T", respRaw)
	}

	if resp.Status().Is4xxClientError() {
		resp.Body().Close()
		return nil, fmt.Errorf("llm: upstream %d", resp.Status().Code())
	}

	return resp, nil
}

func chatPayload(req ChatRequest) map[string]any {
	return map[string]any{
		"model":       req.Model,
		"messages":    req.Messages,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
		"top_p":       req.TopP,
		"stream":      req.Stream,
		"metadata":    req.Extras,
	}
}

func embedPayload(req EmbedRequest) map[string]any {
	return map[string]any{
		"model": req.Model,
		"input": req.Input,
	}
}

type choice struct {
	Message chatDelta `json:"message"`
}

type chatDelta struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type usageBlock struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type chatEnvelope struct {
	ID      string     `json:"id"`
	Model   string     `json:"model"`
	Choices []choice   `json:"choices"`
	Usage   usageBlock `json:"usage"`
}

type streamDelta struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type streamChoice struct {
	Delta streamDelta `json:"delta"`
}

type streamEnvelope struct {
	ID      string         `json:"id"`
	Model   string         `json:"model"`
	Choices []streamChoice `json:"choices"`
}

func parseChatResponse(resp *maigo.Response) (ChatResponse, error) {
	var envelope chatEnvelope
	if err := resp.Body().AsJSON(&envelope); err != nil {
		return ChatResponse{}, err
	}

	if len(envelope.Choices) == 0 {
		return ChatResponse{}, errors.New("llm: empty choices")
	}

	assistant := envelope.Choices[0].Message
	messages := []ChatMessage{assistantAsMessage(assistant)}

	return ChatResponse{
		ID:        envelope.ID,
		Model:     envelope.Model,
		Messages:  messages,
		PromptTok: envelope.Usage.PromptTokens,
		CompTok:   envelope.Usage.CompletionTokens,
	}, nil
}

func (c *HTTPClient) consumeStream(resp *maigo.Response) (ChatResponse, error) {
	if !c.enableStreaming {
		return ChatResponse{}, errors.New("llm: streaming disabled")
	}

	raw := resp.Raw()
	defer raw.Body.Close()

	scanner := bufio.NewScanner(raw.Body)
	scanner.Buffer(make([]byte, 0, 1024), 1<<20)

	var (
		buffer    strings.Builder
		model, id string
	)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}

		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}

		var chunk streamEnvelope
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return ChatResponse{}, err
		}

		if len(chunk.Choices) > 0 {
			buffer.WriteString(chunk.Choices[0].Delta.Content)
		}

		model = chunk.Model
		id = chunk.ID
	}

	if err := scanner.Err(); err != nil {
		return ChatResponse{}, err
	}

	message := ChatMessage{Role: RoleAssistant, Content: buffer.String()}

	return ChatResponse{ID: id, Model: model, Messages: []ChatMessage{message}}, nil
}

type embedEnvelope struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
}

func parseEmbedResponse(resp *maigo.Response) (EmbedResponse, error) {
	var envelope embedEnvelope
	if err := resp.Body().AsJSON(&envelope); err != nil {
		return EmbedResponse{}, err
	}

	embs := make([][]float32, len(envelope.Data))
	for i, d := range envelope.Data {
		embs[i] = d.Embedding
	}

	return EmbedResponse{Model: envelope.Model, Embeddings: embs}, nil
}

func assistantAsMessage(d chatDelta) ChatMessage {
	role := d.Role
	if role == "" {
		role = RoleAssistant
	}

	return ChatMessage{Role: role, Content: d.Content}
}
