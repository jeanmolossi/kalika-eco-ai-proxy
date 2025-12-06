package observability

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/observability/audit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/observability/usage"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
	"github.com/segmentio/kafka-go"
)

// DepsKey is the container key used to store observability dependencies.
const DepsKey = "observability:deps"

// Deps groups dependencies produced by the observability module.
type Deps struct {
	UsagePublisher usage.Publisher
	AuditPublisher audit.Publisher
	Closers        []io.Closer
}

// MustDepsFromContainer retrieves observability dependencies or panics if missing.
func MustDepsFromContainer(c *core.Container) Deps {
	v := c.MustGet(DepsKey)

	deps, ok := v.(Deps)
	if !ok {
		panic("observability: invalid deps type stored in container")
	}

	return deps
}

func buildDependencies(c *core.Container) (Deps, error) {
	conf := c.Config()

	usagePub, usageClosers, err := buildUsagePublisher(conf)
	if err != nil {
		return Deps{}, err
	}

	auditPub, auditClosers, err := buildAuditPublisher(conf)
	if err != nil {
		return Deps{}, err
	}

	deps := Deps{
		UsagePublisher: usagePub,
		AuditPublisher: auditPub,
		Closers:        append(usageClosers, auditClosers...),
	}

	return deps, nil
}

func buildUsagePublisher(conf *config.Config) (usage.Publisher, []io.Closer, error) {
	switch strings.ToLower(conf.UsageSink.Mode) {
	case "kafka":
		writer, err := buildKafkaWriter(
			conf.UsageSink.Topic,
			conf.Kafka.Brokers,
			"usage kafka topic is required when USAGE_MODE=kafka",
			"kafka brokers are required when USAGE_MODE=kafka",
		)
		if err != nil {
			return nil, nil, err
		}

		pub := usage.NewKafkaPublisher(writer)

		return pub, []io.Closer{pub}, nil
	default:
		publisher, err := usage.NewFilePublisher(resolveSinkPath(conf.UsageSink.FilePath, filepath.Join("logs", "usage-events.log")))
		if err != nil {
			return nil, nil, fmt.Errorf("creating usage file publisher: %w", err)
		}

		return publisher, []io.Closer{publisher}, nil
	}
}

func buildAuditPublisher(conf *config.Config) (audit.Publisher, []io.Closer, error) {
	switch strings.ToLower(conf.AuditSink.Mode) {
	case "kafka":
		writer, err := buildKafkaWriter(
			conf.AuditSink.Topic,
			conf.Kafka.Brokers,
			"audit kafka topic is required when AUDIT_MODE=kafka",
			"kafka brokers are required when AUDIT_MODE=kafka",
		)
		if err != nil {
			return nil, nil, err
		}

		pub := audit.NewKafkaPublisher(writer)

		return pub, []io.Closer{pub}, nil
	default:
		publisher, err := audit.NewFilePublisher(resolveSinkPath(conf.AuditSink.FilePath, filepath.Join("logs", "audit-events.log")))
		if err != nil {
			return nil, nil, fmt.Errorf("creating audit file publisher: %w", err)
		}

		return publisher, []io.Closer{publisher}, nil
	}
}

func buildKafkaWriter(topic string, brokers []string, topicErr, brokersErr string) (*kafka.Writer, error) {
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

	return writer, nil
}

func resolveSinkPath(path, fallback string) string {
	if path != "" {
		return path
	}

	return fallback
}
