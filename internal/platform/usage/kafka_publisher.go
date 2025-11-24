package usage

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/segmentio/kafka-go"
)

const usageQueueSize = 256

var ErrUsagePublisherQueueFull = errors.New("usage publisher queue is full")

// KafkaPublisher streams usage events into a Kafka topic.
type KafkaPublisher struct {
	writer *kafka.Writer
	queue  chan kafka.Message
}

// NewKafkaPublisher builds a publisher using the provided writer.
func NewKafkaPublisher(writer *kafka.Writer) *KafkaPublisher {
	p := &KafkaPublisher{
		writer: writer,
		queue:  make(chan kafka.Message, usageQueueSize),
	}

	go p.run()

	return p
}

// Publish sends the serialized event to Kafka using the request ID as the key when available.
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
		return ErrUsagePublisherQueueFull
	}
}

// Close shuts down the underlying writer.
func (p *KafkaPublisher) Close() error {
	close(p.queue)

	return p.writer.Close()
}

func (p *KafkaPublisher) run() {
	for msg := range p.queue {
		_ = p.writer.WriteMessages(context.Background(), msg)
	}
}
