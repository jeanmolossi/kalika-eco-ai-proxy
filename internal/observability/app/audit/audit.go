package audit

import (
	"context"
	"io"

	observability "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/observability"
)

type Event = observability.AuditEvent

type Publisher interface {
	Publish(ctx context.Context, ev Event) error
	io.Closer
}
