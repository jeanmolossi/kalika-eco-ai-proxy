package remote

import (
"context"
"encoding/json"
"fmt"
"net/http"
"net/url"

pkgtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant"
)

type tenantClient struct {
client  *http.Client
baseURL string
}

func newTenantClient(client *http.Client, baseURL string) tenantStore {
return &tenantClient{client: client, baseURL: baseURL}
}

func (t *tenantClient) FindByAPIKey(ctx context.Context, apiKey string) (*pkgtenant.TenantConfig, error) {
endpoint := fmt.Sprintf("%s/tenants/api-keys/%s", t.baseURL, url.PathEscape(apiKey))

req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
if err != nil {
return nil, err
}

resp, err := t.client.Do(req)
if err != nil {
return nil, err
}
defer resp.Body.Close()

switch resp.StatusCode {
case http.StatusNotFound:
return nil, pkgtenant.ErrNotFound
case http.StatusOK:
var tenant pkgtenant.TenantConfig
if err := json.NewDecoder(resp.Body).Decode(&tenant); err != nil {
return nil, err
}
return &tenant, nil
default:
return nil, fmt.Errorf("tenant service returned %d", resp.StatusCode)
}
}

func (t *tenantClient) FindByID(ctx context.Context, tenantID string) (*pkgtenant.TenantConfig, error) {
endpoint := fmt.Sprintf("%s/tenants/%s", t.baseURL, url.PathEscape(tenantID))
req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
if err != nil {
return nil, err
}

resp, err := t.client.Do(req)
if err != nil {
return nil, err
}
defer resp.Body.Close()

switch resp.StatusCode {
case http.StatusNotFound:
return nil, pkgtenant.ErrNotFound
case http.StatusOK:
var tenant pkgtenant.TenantConfig
if err := json.NewDecoder(resp.Body).Decode(&tenant); err != nil {
return nil, err
}
return &tenant, nil
default:
return nil, fmt.Errorf("tenant service returned %d", resp.StatusCode)
}
}

func (t *tenantClient) RevokeExpired(context.Context) (int64, error) {
return 0, nil
}

