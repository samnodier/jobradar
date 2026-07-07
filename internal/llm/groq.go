package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	groqBaseURL = "https://api.groq.com/openai/v1"
	groqModel   = "llama-3.3-70b-versatile"
)

// Groq's json_object mode guarantees syntactically valid JSON but has no
// schema channel (unlike Gemini's ResponseSchema), so the expected shape
// must be spelled out in the prompt itself.
const groqExtractSchema = `The JSON object must have exactly these keys:
{"title": string, "company_name": string, "description": string,
 "salary_min": integer or null, "salary_max": integer or null,
 "currency": string or null, "job_location": string or null,
 "is_remote": boolean or null, "skills": array of strings}`

const groqEnrichSchema = `The JSON object must have exactly this key:
{"summary": string}`

// groqClient talks to Groq's OpenAI-compatible chat completions API.
// It implements both Extractor and Enricher.
type groqClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

var (
	_ Extractor = (*groqClient)(nil)
	_ Enricher  = (*groqClient)(nil)
)

func newGroqClient(apiKey string) *groqClient {
	return &groqClient{
		apiKey:  apiKey,
		baseURL: groqBaseURL,
		model:   groqModel,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponseFormat struct {
	Type string `json:"type"`
}

type groqChatRequest struct {
	Model          string              `json:"model"`
	Messages       []groqMessage       `json:"messages"`
	Temperature    float64             `json:"temperature"`
	ResponseFormat *groqResponseFormat `json:"response_format,omitempty"`
}

type groqChatResponse struct {
	Choices []struct {
		Message groqMessage `json:"message"`
	} `json:"choices"`
}

type groqErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (g *groqClient) Extract(ctx context.Context, input ExtractionInput) (ExtractionResult, error) {
	content, err := g.chat(ctx,
		ExtractInstructions+"\n"+groqExtractSchema,
		"Extract the job posting from this page text:\n\n"+input.PageText,
	)
	if err != nil {
		return ExtractionResult{}, err
	}

	var parsed ExtractionResult
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return ExtractionResult{}, fmt.Errorf("groq: failed to parse extraction JSON: %w", err)
	}
	return parsed, nil
}

func (g *groqClient) Enrich(ctx context.Context, input EnrichmentInput) (EnrichmentResult, error) {
	content, err := g.chat(ctx,
		EnrichInstructions+"\n"+groqEnrichSchema,
		buildEnrichPrompt(input),
	)
	if err != nil {
		return EnrichmentResult{}, err
	}

	var parsed EnrichmentResult
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return EnrichmentResult{}, fmt.Errorf("groq: failed to parse enrichment JSON: %w", err)
	}
	return parsed, nil
}

// chat sends one system+user exchange in JSON mode and returns the raw
// message content.
func (g *groqClient) chat(ctx context.Context, system, user string) (string, error) {
	reqBody := groqChatRequest{
		Model: g.model,
		Messages: []groqMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature:    0,
		ResponseFormat: &groqResponseFormat{Type: "json_object"},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("groq: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("groq: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("groq: request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", classifyGroqError(resp)
	}

	var chatResp groqChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("groq: decode response: %w", err)
	}
	if len(chatResp.Choices) == 0 || chatResp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("groq: empty response returned")
	}
	return chatResp.Choices[0].Message.Content, nil
}

// classifyGroqError maps a non-200 response onto the retry taxonomy:
// wrap ErrPermanent when a retry cannot succeed (bad request, bad key),
// return a plain error when it might (rate limit, server trouble).
func classifyGroqError(resp *http.Response) error {
	apiMsg := resp.Status
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err == nil {
		var parsed groqErrorResponse
		if json.Unmarshal(body, &parsed) == nil && parsed.Error.Message != "" {
			apiMsg = fmt.Sprintf("%s: %s", resp.Status, parsed.Error.Message)
		}
	}

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return fmt.Errorf("groq: authentication failed: %w: %s", ErrPermanent, apiMsg)
	case resp.StatusCode == http.StatusTooManyRequests:
		return fmt.Errorf("groq: rate limit exceeded: %s", apiMsg)
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		return fmt.Errorf("groq: invalid request: %w: %s", ErrPermanent, apiMsg)
	default:
		return fmt.Errorf("groq: service unavailable: %s", apiMsg)
	}
}
