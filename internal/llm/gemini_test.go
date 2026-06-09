package llm

import (
	"errors"
	"testing"

	"google.golang.org/genai"
)

func TestClassifyGeminiError(t *testing.T) {
	tests := []struct {
		name          string
		inputErr      error
		wantPermanent bool
	}{
		// Permanent Errors (Should wrap ErrPermanent)
		{"bad request", genai.APIError{Code: 400}, true},
		{"unauthorized (bad api key)", genai.APIError{Code: 401}, true},
		{"forbidden (region locked)", genai.APIError{Code: 403}, true},
		{"not found (model deprecated)", genai.APIError{Code: 404}, true},

		// Transient Errors (Should NOT wrap ErrPermanent)
		{"rate limited", genai.APIError{Code: 429}, false},
		{"internal server error", genai.APIError{Code: 500}, false},
		{"service unavailable", genai.APIError{Code: 503}, false},

		// Edge Cases
		{"connection reset", errors.New("connection reset by peer"), false}, // Should fall into the default switch block and be retryable
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotError := classifyGeminiError(tt.inputErr)

			isPermanent := errors.Is(gotError, ErrPermanent)

			if isPermanent != tt.wantPermanent {
				t.Errorf("classifyGeminiError(code: %d): wrapped ErrPermanent = %v, want %v", tt.inputErr, isPermanent, tt.wantPermanent)
			}
		})
	}
}
