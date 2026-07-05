package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ollamaExtractor struct {
	baseURL string // eg. "http://localhost:11434"
	model   string // eg. "qwen-3:4b"
	client  *http.Client
}

func NewOllamaExtractor(baseURL, model string) (Extractor, error) {
	return &ollamaExtractor{
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

var _ Extractor = (*ollamaExtractor)(nil)

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Think    bool            `json:"think"`
	Format   json.RawMessage `json:"format,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Model   string        `json:"model"`
	Message ollamaMessage `json:"message"`
	Done    bool          `json:"done"`
}

func (o *ollamaExtractor) Extract(ctx context.Context, input ExtractionInput) (ExtractionResult, error) {
	reqBody := ollamaRequest{
		Model: o.model,
		Messages: []ollamaMessage{
			{Role: "system", Content: ExtractInstructions},
			{Role: "user", Content: "Extract the job posting from this page text: \n\n" + input.PageText},
		},
		Stream: false,
		Think:  false,
		Format: json.RawMessage(`"json"`),
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return ExtractionResult{}, fmt.Errorf("ollama: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/api/chat", bytes.NewReader(payload))
	if err != nil {
		return ExtractionResult{}, fmt.Errorf("ollama: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return ExtractionResult{}, fmt.Errorf("ollama: request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return ExtractionResult{}, fmt.Errorf("ollama: unexpected response: %s", resp.Status)
	}

	var chatResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return ExtractionResult{}, fmt.Errorf("ollama: decode response: %w", err)
	}

	var result ExtractionResult
	if err := json.Unmarshal([]byte(chatResp.Message.Content), &result); err != nil {
		return ExtractionResult{}, fmt.Errorf("ollama: parse extraction JSON %q: %w", chatResp.Message.Content, err)
	}

	return result, nil
}
