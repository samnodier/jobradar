package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/llm"
)

// errNoAPIKey means the user has no key for any provider we support.
// Callers translate it: handlers to a 422, workers to a silent drop.
var errNoAPIKey = errors.New("no AI provider API key configured")

// providerKey pairs a provider name with the user's decrypted API key for it.
type providerKey struct {
	provider string
	apiKey   string
}

// selectProviderKeys returns every provider the user has a key for, decrypted,
// in llm.ProviderPriority order. Callers try them in order, so when the first
// provider fails (down, rate-limited, revoked key) they fall back to the next.
func (cfg *apiConfig) selectProviderKeys(ctx context.Context, userID uuid.UUID) ([]providerKey, error) {
	configured, err := cfg.db.ListUserAPIKeyProviders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list api key providers: %w", err)
	}

	have := make(map[string]struct{}, len(configured))
	for _, p := range configured {
		have[p] = struct{}{}
	}

	var keys []providerKey
	for _, p := range llm.ProviderPriority {
		if _, ok := have[p]; !ok {
			continue
		}
		ciphertext, err := cfg.db.GetUserAPIKey(ctx, database.GetUserAPIKeyParams{
			UserID:   userID,
			Provider: p,
		})
		if err != nil {
			return nil, fmt.Errorf("fetch %s key: %w", p, err)
		}
		plain, err := cfg.crypto.Decrypt(ciphertext)
		if err != nil {
			// A decrypt failure is ours (corrupt ciphertext or wrong
			// ENCRYPTION_KEY) — fail loud, don't quietly skip the provider.
			return nil, fmt.Errorf("decrypt %s key: %w", p, err)
		}
		keys = append(keys, providerKey{provider: p, apiKey: plain})
	}

	if len(keys) == 0 {
		return nil, errNoAPIKey
	}
	return keys, nil
}
