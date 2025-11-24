package audit

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/segmentio/kafka-go"
)

const auditQueueSize = 256

var ErrAuditPublisherQueueFull = errors.New("audit publisher queue is full")

// KafkaPublisher writes audit events into a Kafka topic.
type KafkaPublisher struct {
	writer *kafka.Writer
	queue  chan kafka.Message
}

// NewKafkaPublisher builds a publisher backed by the provided writer.
func NewKafkaPublisher(writer *kafka.Writer) *KafkaPublisher {
	p := &KafkaPublisher{
		writer: writer,
		queue:  make(chan kafka.Message, auditQueueSize),
	}

	go p.run()

	return p
}

// Publish emits the audit event to Kafka.
func (p *KafkaPublisher) Publish(ctx context.Context, ev Event) error {
	payload, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	key := []byte(ev.RequestID)
	if len(key) == 0 {
		key = []byte(ev.TenantID)
	}

	msg := kafka.Message{
		Key:   key,
		Value: payload,
		Time:  time.Now(),
	}

	select {
	case p.queue <- msg:
		return nil
	default:
		return ErrAuditPublisherQueueFull
	}
}

// Close stops the writer and flushes pending messages.
func (p *KafkaPublisher) Close() error {
	close(p.queue)

	return p.writer.Close()
}

func (p *KafkaPublisher) run() {
	for msg := range p.queue {
		_ = p.writer.WriteMessages(context.Background(), msg)
	}
}
