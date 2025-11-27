package remote

import (
"bytes"
"context"
"encoding/json"
"fmt"
"net/http"

"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/observability"
)

type usageClient struct {
client  *http.Client
baseURL string
}

func newUsageClient(client *http.Client, baseURL string) usagePublisher {
return &usageClient{client: client, baseURL: baseURL}
}

func (u *usageClient) Publish(ctx context.Context, ev observability.UsageEvent) error {
return u.post(ctx, "/observability/usage", ev)
}

type auditClient struct {
client  *http.Client
baseURL string
}

func newAuditClient(client *http.Client, baseURL string) auditPublisher {
return &auditClient{client: client, baseURL: baseURL}
}

func (a *auditClient) Publish(ctx context.Context, ev observability.AuditEvent) error {
return a.post(ctx, "/observability/audit", ev)
}

func (u *usageClient) post(ctx context.Context, path string, ev any) error {
return postEvent(ctx, u.client, u.baseURL, path, ev)
}

func (a *auditClient) post(ctx context.Context, path string, ev any) error {
return postEvent(ctx, a.client, a.baseURL, path, ev)
}

func postEvent(ctx context.Context, client *http.Client, baseURL, path string, ev any) error {
payload, err := json.Marshal(ev)
if err != nil {
return err
}

endpoint := baseURL + path
req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
if err != nil {
return err
}
req.Header.Set("Content-Type", "application/json")

resp, err := client.Do(req)
if err != nil {
return err
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusAccepted {
return fmt.Errorf("observability service returned %d", resp.StatusCode)
}

return nil
}

