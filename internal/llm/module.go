package llm

import (
	"context"
	"fmt"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/llm/router"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/llm/tokenizer"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
	"github.com/labstack/echo/v4"
)

const ModuleName = "llm"

type module struct{}

func NewModule() core.Module { return &module{} }

func (m *module) Name() string                                  { return ModuleName }
func (m *module) Weight() int                                   { return 6 }
func (m *module) Routes(_ *echo.Group, _ *core.Container) error { return nil }

func (m *module) Provide(_ context.Context, c *core.Container) error {
	conf := c.Config()
	if err := ensureLLMConfig(conf); err != nil {
		return err
	}

	aliases := buildAliases(conf)

	llmDefaults := ProviderSettings{
		Name:            conf.LLM.ProviderName,
		BaseURL:         conf.LLM.BaseURL,
		APIKey:          conf.LLM.APIKey,
		RequestTimeout:  conf.LLM.RequestTimeout,
		MaxRetries:      conf.LLM.MaxRetries,
		EnableStreaming: conf.LLM.EnableStreaming,
		ChatModels:      conf.LLM.ChatModels,
		EmbedModels:     conf.LLM.EmbedModels,
	}

	pool := NewProviderPool(llmDefaults, NoopMetrics{})
	tokenzr := tokenizer.NewOpenAITikTokenCounter(aliases)

	c.Set(core.RouterModule, router.NewSimpleRouter(pool))
	c.Set(core.TokenizerModule, tokenzr)

	return nil
}

func ensureLLMConfig(conf *config.Config) error {
	if conf.LLM.BaseURL == "" {
		return fmt.Errorf("llm base url is required; set LLM_BASE_URL")
	}

	return nil
}

func buildAliases(conf *config.Config) map[string]string {
	aliases := make(map[string]string)

	for _, model := range append(append([]string{}, conf.LLM.ChatModels...), conf.LLM.EmbedModels...) {
		if model != "" {
			aliases[model] = model
		}
	}

	return aliases
}

func (m *module) Start(_ context.Context, _ *core.Container) (func(context.Context) error, error) {
	return nil, nil
}

func (m *module) Migrations(ctx context.Context, c *core.Container) ([]core.MigrationFile, error) {
	return make([]core.MigrationFile, 0), nil
}
