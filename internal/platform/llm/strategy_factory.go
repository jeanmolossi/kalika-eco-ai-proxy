package llm

import (
	"context"
	"time"
)

// StrategyFactory creates model strategies backed by langchain-go.
type StrategyFactory struct {
	metrics MetricsRecorder
}

// NewStrategyFactory builds a new StrategyFactory.
func NewStrategyFactory(metrics MetricsRecorder) StrategyFactory {
	if metrics == nil {
		metrics = NoopMetrics{}
	}

	return StrategyFactory{metrics: metrics}
}

// Build constructs a Client using the given provider configuration.
func (f StrategyFactory) Build(cfg ProviderSettings) (Client, error) {
	strategy, err := NewLangChainStrategy(cfg)
	if err != nil {
		return nil, err
	}

	return &StrategyClient{strategy: strategy, metrics: f.metrics}, nil
}

// StrategyClient decorates a ModelStrategy with metrics collection.
type StrategyClient struct {
	strategy ModelStrategy
	metrics  MetricsRecorder
}

// Chat implements the Client interface.
func (s *StrategyClient) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	start := time.Now()
	resp, err := s.strategy.Chat(ctx, req)
	s.metrics.ObserveChat(req.Model, time.Since(start), err)

	return resp, err
}

// Embed implements the Client interface.
func (s *StrategyClient) Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error) {
	start := time.Now()
	resp, err := s.strategy.Embed(ctx, req)
	s.metrics.ObserveEmbed(req.Model, time.Since(start), err)

	return resp, err
}

// Name returns the strategy name for debugging purposes.
func (s *StrategyClient) Name() string {
	return s.strategy.Name()
}
