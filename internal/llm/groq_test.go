package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestGroqClient points a groqClient at a fake server.
func newTestGroqClient(serverURL string) *groqClient {
	c := newGroqClient("test-key")
	c.baseURL = serverURL
	return c
}

func TestGroqExtract_Success(t *testing.T) {
	// The content field is a JSON *string* holding the extraction object,
	// exactly as Groq's OpenAI-compatible API returns it.
	extraction := `{"title":"Backend Engineer","company_name":"Acme","description":"Build APIs","salary_min":90000,"salary_max":120000,"currency":"USD","job_location":"Berlin, Germany","is_remote":true,"skills":["Go","PostgreSQL"]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("unexpected auth header: %q", got)
		}
		var req groqChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.ResponseFormat == nil || req.ResponseFormat.Type != "json_object" {
			t.Error("expected json_object response format")
		}

		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"role": "assistant", "content": extraction}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestGroqClient(server.URL)
	result, err := client.Extract(context.Background(), ExtractionInput{PageText: "some page text"})
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if result.Title != "Backend Engineer" {
		t.Errorf("title = %q, want %q", result.Title, "Backend Engineer")
	}
	if result.IsRemote == nil || !*result.IsRemote {
		t.Error("is_remote should be true")
	}
	if len(result.Skills) != 2 {
		t.Errorf("skills = %v, want 2 entries", result.Skills)
	}
}

func TestGroqErrors_RetryClassification(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		wantPermanent bool
	}{
		{"bad key is permanent", http.StatusUnauthorized, true},
		{"forbidden is permanent", http.StatusForbidden, true},
		{"bad request is permanent", http.StatusBadRequest, true},
		{"rate limit is transient", http.StatusTooManyRequests, false},
		{"server error is transient", http.StatusInternalServerError, false},
		{"unavailable is transient", http.StatusServiceUnavailable, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				// Groq returns OpenAI-style error bodies
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]any{"message": "boom", "type": "invalid_request_error"},
				})
			}))
			defer server.Close()

			client := newTestGroqClient(server.URL)
			_, err := client.Extract(context.Background(), ExtractionInput{PageText: "text"})
			if err == nil {
				t.Fatal("expected an error")
			}
			if got := errors.Is(err, ErrPermanent); got != tt.wantPermanent {
				t.Errorf("errors.Is(err, ErrPermanent) = %v, want %v (err: %v)", got, tt.wantPermanent, err)
			}
		})
	}
}

func TestGroqExtract_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{}})
	}))
	defer server.Close()

	client := newTestGroqClient(server.URL)
	_, err := client.Extract(context.Background(), ExtractionInput{PageText: "text"})
	if err == nil {
		t.Fatal("expected an error for empty choices")
	}
}
