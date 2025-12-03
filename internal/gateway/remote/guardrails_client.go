package remote

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/guardrails"
	maigocontracts "github.com/jeanmolossi/maigo/pkg/maigo/contracts"
	"github.com/jeanmolossi/maigo/pkg/maigo/mime"
)

type guardrailsClient struct {
	client maigocontracts.ClientHTTPMethods
}

func newGuardrailsClient(client maigocontracts.ClientHTTPMethods) guardrailsEngine {
	return &guardrailsClient{client: client}
}

func (g *guardrailsClient) EvaluateInput(ctx context.Context, gx guardrails.Context) (guardrails.Decision, error) {
	return g.call(ctx, "guardrails/evaluate/input", gx)
}

func (g *guardrailsClient) EvaluateOutput(ctx context.Context, gx guardrails.Context) (guardrails.Decision, error) {
	return g.call(ctx, "guardrails/evaluate/output", gx)
}

func (g *guardrailsClient) call(ctx context.Context, path string, gx guardrails.Context) (guardrails.Decision, error) {
	resp, err := g.client.POST(path).
		Context().Set(ctx).
		Header().AddContentType(mime.JSON).
		Body().AsJSON(gx).
		Send()
	if err != nil {
		return guardrails.Decision{}, err
	}

	if resp.Status().Code() != http.StatusOK {
		return guardrails.Decision{}, fmt.Errorf("guardrails service returned %d", resp.Status().Code())
	}

	var decision guardrails.Decision
	if err := resp.Body().AsJSON(&decision); err != nil {
		return guardrails.Decision{}, err
	}

	return decision, nil
}
