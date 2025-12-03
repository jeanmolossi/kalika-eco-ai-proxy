package observability

import "math"

var perTokenUSD = map[string]struct {
	Prompt     float64
	Completion float64
}{
	"gpt-4o-mini":    {Prompt: 0.00000015, Completion: 0.0000006},
	"gpt-4o":         {Prompt: 0.000005, Completion: 0.000015},
	"text-embedding": {Prompt: 0.00000002, Completion: 0},
}

// CalculateUSD computes an estimated cost in USD for the given model and token counts.
// If the model is unknown, it returns zero to avoid blocking the flow while still publishing usage.
func CalculateUSD(model string, promptTokens, completionTokens int) float64 {
	price, ok := perTokenUSD[model]
	if !ok {
		return 0
	}

	prompt := float64(promptTokens) * price.Prompt
	completion := float64(completionTokens) * price.Completion

	return math.Round((prompt+completion)*1000000) / 1000000
}
