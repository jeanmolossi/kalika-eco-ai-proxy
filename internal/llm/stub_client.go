package llm

import (
	"context"
	"fmt"
	"time"
)

// StubClient is a fake LLM client that generates deterministic responses
// without calling any external provider. It is used for development and testing.
type StubClient struct{}

// NewStubClient creates a new StubClient instance.
func NewStubClient() *StubClient {
	return &StubClient{}
}

// Chat returns a simple static response based on the last user message.
func (c *StubClient) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	var lastUser string

	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			lastUser = req.Messages[i].Content
			break
		}
	}

	if lastUser == "" && len(req.Messages) > 0 {
		lastUser = req.Messages[len(req.Messages)-1].Content
	}

	now := time.Now().UnixNano()

	respMsg := ChatMessage{
		Role:    "assistant",
		Content: fmt.Sprintf("stub response from model %s for input: %q", req.Model, lastUser),
	}

	return ChatResponse{
		ID:        fmt.Sprintf("stub-%d", now),
		Model:     req.Model,
		Messages:  append(req.Messages, respMsg),
		PromptTok: 42,
		CompTok:   13,
	}, nil
}

// Embed returns a fixed-size embedding vector filled with zeros for each input.
func (c *StubClient) Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error) {
	embs := make([][]float32, len(req.Input))
	for i := range embs {
		embs[i] = make([]float32, 8)
	}

	return EmbedResponse{
		Model:      req.Model,
		Embeddings: embs,
	}, nil
}
