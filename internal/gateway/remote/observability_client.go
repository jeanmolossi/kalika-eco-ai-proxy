package remote

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/observability"
	maigocontracts "github.com/jeanmolossi/maigo/pkg/maigo/contracts"
	"github.com/jeanmolossi/maigo/pkg/maigo/mime"
)

type usageClient struct {
	client maigocontracts.ClientHTTPMethods
}

func newUsageClient(client maigocontracts.ClientHTTPMethods) usagePublisher {
	return &usageClient{client: client}
}

func (u *usageClient) Publish(ctx context.Context, ev observability.UsageEvent) error {
	return u.post(ctx, "observability/usage", ev)
}

type auditClient struct {
	client maigocontracts.ClientHTTPMethods
}

func newAuditClient(client maigocontracts.ClientHTTPMethods) auditPublisher {
	return &auditClient{client: client}
}

func (a *auditClient) Publish(ctx context.Context, ev observability.AuditEvent) error {
	return a.post(ctx, "observability/audit", ev)
}

func (u *usageClient) post(ctx context.Context, path string, ev any) error {
	return postEvent(ctx, u.client, path, ev)
}

func (a *auditClient) post(ctx context.Context, path string, ev any) error {
	return postEvent(ctx, a.client, path, ev)
}

func postEvent(ctx context.Context, client maigocontracts.ClientHTTPMethods, path string, ev any) error {
	resp, err := client.POST(path).
		Context().Set(ctx).
		Header().AddContentType(mime.JSON).
		Body().AsJSON(ev).
		Send()
	if err != nil {
		return err
	}

	if resp.Status().Code() != http.StatusAccepted {
		return fmt.Errorf("observability service returned %d", resp.Status().Code())
	}

	return nil
}
