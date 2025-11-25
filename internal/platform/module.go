package platform

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/audit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/cache"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/config"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/database"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/database/pg"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/guardrails"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/ratelimit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/router"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/tenant"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/tokenizer"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/usage"
	"github.com/labstack/echo/v4"
	"github.com/segmentio/kafka-go"
)

// ModuleName is the identifier of this module.
const ModuleName = "platform"

// module implements core.Module and is responsible for the AI proxy data-plane.
type module struct {
	closers []io.Closer
}

// NewModule creates a new AI proxy module.
func NewModule() core.Module {
	return &module{}
}

// Name implements core.Module.
func (m *module) Name() string { return ModuleName }

// Weight implements core.Module.
func (m *module) Weight() int { return 9 }

// Provide implements core.Module.
func (m *module) Provide(ctx context.Context, c *core.Container) error {
	conf := c.Config()

	conn := core.MustGet[*pg.DB](c, database.PgConn)

	rl, err := ratelimit.NewLimiter(conf.RateLimit)
	if err != nil {
		return err
	}

	if conf.LLM.BaseURL == "" {
		return fmt.Errorf("llm base url is required; set LLM_BASE_URL")
	}

	aliases := make(map[string]string)

	for _, m := range append(append([]string{}, conf.LLM.ChatModels...), conf.LLM.EmbedModels...) {
		if m != "" {
			aliases[m] = m
		}
	}

	llmDefaults := llm.ProviderSettings{
		Name:            conf.LLM.ProviderName,
		BaseURL:         conf.LLM.BaseURL,
		APIKey:          conf.LLM.APIKey,
		RequestTimeout:  conf.LLM.RequestTimeout,
		MaxRetries:      conf.LLM.MaxRetries,
		EnableStreaming: conf.LLM.EnableStreaming,
		ChatModels:      conf.LLM.ChatModels,
		EmbedModels:     conf.LLM.EmbedModels,
	}

	pool := llm.NewProviderPool(llmDefaults, llm.NoopMetrics{})

	tokenzr := tokenizer.NewOpenAITikTokenCounter(aliases)

	c.Set(core.TenantStoreModule, tenant.NewPostgresStore(conn.Pool()))
	c.Set(core.RateLimiterModule, rl)
	c.Set(core.SemanticCacheModule, cache.NewNoopSemanticCache())
	c.Set(core.GuardrailsModule, guardrails.ProvideGuardrails(c))

	usagePub, err := m.provideUsagePublisher(conf)
	if err != nil {
		return err
	}

	auditPub, err := m.provideAuditPublisher(conf)
	if err != nil {
		return err
	}

	c.Set(core.UsagePublisherModule, usagePub)
	c.Set(core.AuditPublisherModule, auditPub)
	c.Set(core.RouterModule, router.NewSimpleRouter(pool))
	c.Set(core.TokenizerModule, tokenzr)

	return nil
}

func (m *module) provideUsagePublisher(conf *config.Config) (usage.Publisher, error) {
	switch strings.ToLower(conf.UsageSink.Mode) {
	case "kafka":
		writer, err := m.buildKafkaWriter(
			conf.UsageSink.Topic,
			conf.Kafka.Brokers,
			"usage kafka topic is required when USAGE_MODE=kafka",
			"kafka brokers are required when USAGE_MODE=kafka",
		)
		if err != nil {
			return nil, err
		}

		return usage.NewKafkaPublisher(writer), nil
	default:
		return usage.NewFilePublisher(resolveSinkPath(conf.UsageSink.FilePath, filepath.Join("logs", "usage-events.log")))
	}
}

func (m *module) provideAuditPublisher(conf *config.Config) (audit.Publisher, error) {
	switch strings.ToLower(conf.AuditSink.Mode) {
	case "kafka":
		writer, err := m.buildKafkaWriter(
			conf.AuditSink.Topic,
			conf.Kafka.Brokers,
			"audit kafka topic is required when AUDIT_MODE=kafka",
			"kafka brokers are required when AUDIT_MODE=kafka",
		)
		if err != nil {
			return nil, err
		}

		return audit.NewKafkaPublisher(writer), nil
	default:
		return audit.NewFilePublisher(resolveSinkPath(conf.AuditSink.FilePath, filepath.Join("logs", "audit-events.log")))
	}
}

func (m *module) buildKafkaWriter(topic string, brokers []string, topicErr, brokersErr string) (*kafka.Writer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("%s", brokersErr)
	}

	if topic == "" {
		return nil, fmt.Errorf("%s", topicErr)
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.Hash{},
		AllowAutoTopicCreation: false,
		RequiredAcks:           kafka.RequireAll,
	}

	m.closers = append(m.closers, writer)

	return writer, nil
}

func resolveSinkPath(path, fallback string) string {
	if path != "" {
		return path
	}

	return fallback
}

// Routes implements core.Module.
func (m *module) Routes(g *echo.Group, c *core.Container) error {
	return nil
}

// Start implements core.Module.
func (m *module) Start(ctx context.Context, c *core.Container) (stop func(context.Context) error, err error) {
	store := core.MustGet[tenant.Store](c, core.TenantStoreModule)
	log := c.Logger()
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(1 * time.Hour)

		defer func() {
			ticker.Stop()
			close(done)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if revoked, err := store.RevokeExpired(ctx); err != nil {
					log.ErrorContext(ctx, "api key revocation", "err", err)
				} else if revoked > 0 {
					log.InfoContext(ctx, "revoked expired api keys", "count", revoked)
				}
			}
		}
	}()

	return func(ctx context.Context) error {
		cancel()

		select {
		case <-done:
		case <-ctx.Done():
		}

		for _, closer := range m.closers {
			if err := closer.Close(); err != nil {
				log.ErrorContext(ctx, "close module resource", "err", err)
			}
		}

		return nil
	}, nil
}

// Migrations implements core.Module.
func (m *module) Migrations(ctx context.Context, c *core.Container) ([]core.MigrationFile, error) {
	return make([]core.MigrationFile, 0), nil
}
