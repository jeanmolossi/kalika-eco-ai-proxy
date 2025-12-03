package observability

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/observability/audit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/observability/usage"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
	"github.com/labstack/echo/v4"
	"github.com/segmentio/kafka-go"
)

const ModuleName = "observability"

type module struct {
	closers []io.Closer
}

func NewModule() core.Module { return &module{} }

func (m *module) Name() string                                  { return ModuleName }
func (m *module) Weight() int                                   { return 7 }
func (m *module) Routes(g *echo.Group, c *core.Container) error { return registerRoutes(g, c) }

func (m *module) Provide(_ context.Context, c *core.Container) error {
	conf := c.Config()

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

	return nil
}

func (m *module) Start(_ context.Context, c *core.Container) (func(context.Context) error, error) {
	return func(ctx context.Context) error {
		log := c.Logger()

		for _, closer := range m.closers {
			if err := closer.Close(); err != nil {
				log.ErrorContext(ctx, "close observability resource", "err", err)
			}
		}

		return nil
	}, nil
}

func (m *module) Migrations(ctx context.Context, c *core.Container) ([]core.MigrationFile, error) {
	return make([]core.MigrationFile, 0), nil
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
		publisher, err := usage.NewFilePublisher(resolveSinkPath(conf.UsageSink.FilePath, filepath.Join("logs", "usage-events.log")))
		if err != nil {
			return nil, fmt.Errorf("creating usage file publisher: %w", err)
		}

		m.closers = append(m.closers, publisher)

		return publisher, nil
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
		publisher, err := audit.NewFilePublisher(resolveSinkPath(conf.AuditSink.FilePath, filepath.Join("logs", "audit-events.log")))
		if err != nil {
			return nil, fmt.Errorf("creating audit file publisher: %w", err)
		}

		m.closers = append(m.closers, publisher)

		return publisher, nil
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
