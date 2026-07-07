package llm

import (
	"context"
	"fmt"
)

const (
	ProviderGemini = "gemini"
	ProviderGroq   = "groq"
)

// ProviderPriority is the order providers are tried when a user has keys for
// more than one. Groq first: much faster inference and a friendlier free tier.
var ProviderPriority = []string{ProviderGroq, ProviderGemini}

// IsKnownProvider reports whether name is a provider this package can build
// clients for. Handlers use it to validate user input at the edge; the DB
// CHECK constraint on user_api_keys.provider is the backstop.
func IsKnownProvider(name string) bool {
	switch name {
	case ProviderGemini, ProviderGroq:
		return true
	}
	return false
}

// NewExtractor builds the Extractor for the given provider.
func NewExtractor(ctx context.Context, provider, apiKey string) (Extractor, error) {
	switch provider {
	case ProviderGemini:
		return NewGeminiExtractor(ctx, apiKey)
	case ProviderGroq:
		return newGroqClient(apiKey), nil
	default:
		return nil, fmt.Errorf("llm: unknown extractor provider %q", provider)
	}
}

// NewEnricher builds the Enricher for the given provider.
func NewEnricher(ctx context.Context, provider, apiKey string) (Enricher, error) {
	switch provider {
	case ProviderGemini:
		return NewGeminiEnricher(ctx, apiKey)
	case ProviderGroq:
		return newGroqClient(apiKey), nil
	default:
		return nil, fmt.Errorf("llm: unknown enricher provider %q", provider)
	}
}
