package remote

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	pkgtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant"
	maigocontracts "github.com/jeanmolossi/maigo/pkg/maigo/contracts"
)

type tenantClient struct {
	client maigocontracts.ClientHTTPMethods
}

func newTenantClient(client maigocontracts.ClientHTTPMethods) pkgtenant.Store {
	return &tenantClient{client: client}
}

func (t *tenantClient) FindByAPIKey(ctx context.Context, apiKey string) (*pkgtenant.TenantConfig, error) {
	endpoint := fmt.Sprintf("tenants/api-keys/%s", url.PathEscape(apiKey))

	return t.fetchTenant(ctx, endpoint)
}

func (t *tenantClient) FindByID(ctx context.Context, tenantID string) (*pkgtenant.TenantConfig, error) {
	endpoint := fmt.Sprintf("tenants/%s", url.PathEscape(tenantID))

	return t.fetchTenant(ctx, endpoint)
}

func (t *tenantClient) RevokeExpired(context.Context) (int64, error) {
	return 0, nil
}

func (t *tenantClient) fetchTenant(ctx context.Context, endpoint string) (*pkgtenant.TenantConfig, error) {
	resp, err := t.client.GET(endpoint).Context().Set(ctx).Send()
	if err != nil {
		return nil, err
	}

	switch resp.Status().Code() {
	case http.StatusNotFound:
		return nil, pkgtenant.ErrNotFound
	case http.StatusOK:
		var tenant pkgtenant.TenantConfig
		if err := resp.Body().AsJSON(&tenant); err != nil {
			return nil, err
		}

		return &tenant, nil
	default:
		return nil, fmt.Errorf("tenant service returned %d", resp.Status().Code())
	}
}
