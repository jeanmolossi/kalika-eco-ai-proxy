package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/guardrails"
)

type guardrailsClient struct {
	client  *http.Client
	baseURL string
}

func newGuardrailsClient(client *http.Client, baseURL string) guardrailsEngine {
	return &guardrailsClient{client: client, baseURL: baseURL}
}

func (g *guardrailsClient) EvaluateInput(ctx context.Context, gx guardrails.Context) (guardrails.Decision, error) {
	return g.call(ctx, "/guardrails/evaluate/input", gx)
}

func (g *guardrailsClient) EvaluateOutput(ctx context.Context, gx guardrails.Context) (guardrails.Decision, error) {
	return g.call(ctx, "/guardrails/evaluate/output", gx)
}

func (g *guardrailsClient) call(ctx context.Context, path string, gx guardrails.Context) (guardrails.Decision, error) {
	payload, err := json.Marshal(gx)
	if err != nil {
		return guardrails.Decision{}, err
	}

	endpoint := g.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return guardrails.Decision{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return guardrails.Decision{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return guardrails.Decision{}, fmt.Errorf("guardrails service returned %d", resp.StatusCode)
	}

	var decision guardrails.Decision
	if err := json.NewDecoder(resp.Body).Decode(&decision); err != nil {
		return guardrails.Decision{}, err
	}

	return decision, nil
}
