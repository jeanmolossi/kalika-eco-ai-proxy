package tokenizer

import internaltokenizer "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/tokenizer"

// Tokenizer contracts exposed for other modules.
type (
	TokenCounter         = internaltokenizer.TokenCounter
	TokenCounterResolver = internaltokenizer.TokenCounterResolver
)
