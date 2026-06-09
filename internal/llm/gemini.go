package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type geminiEnricher struct {
	client *genai.Client
}

var ErrPermanent = errors.New("permanent enrichment failure")

// NewGeminiEnricher creates a new instance of the GeminiEnricher
func NewGeminiEnricher(ctx context.Context, apiKey string) (Enricher, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to create client: %w", err)
	}

	return &geminiEnricher{client: client}, nil
}

var _ Enricher = (*geminiEnricher)(nil)

const systemInstructions = `You are a career assistant speaking directly to a job-seeker about a specific role.
Write a concise 2-3 sentence summary addressing the user as "you" — for example: "You have strong experience in X" or "You're missing Y which this role requires."
Be specific — mention relevant skills and experience by name.
Do not invent skills or experience the user does not have.
Respond only with valid JSON matching the schema provided. No markdown, no preamble.`

const model = "gemini-3.5-flash"

func (g *geminiEnricher) Enrich(ctx context.Context, input EnrichmentInput) (EnrichmentResult, error) {
	userContent := fmt.Sprintf(
		`
		Job Title: %s
		Job Location: %s
    Job Description: %s

		Candidate Desired Roles: %s
		Candidate Skills: %s
		Candidate Experience: %s
		`, input.JobTitle, input.JobLocation, input.JobDescription, strings.Join(input.DesiredRoles, ", "), strings.Join(input.UserSkills, ", "), formatExperience(input.UserExperiences),
	)

	result, err := g.client.Models.GenerateContent(
		ctx,
		model,
		genai.Text(userContent),
		&genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(systemInstructions, ""),
			ResponseMIMEType:  "application/json",
			ResponseSchema: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"summary": {Type: genai.TypeString},
				},
				Required: []string{"summary"},
			},
		},
	)
	if err != nil {
		return EnrichmentResult{}, classifyGeminiError(err)
	}

	raw := result.Text()
	if raw == "" {
		return EnrichmentResult{}, fmt.Errorf("gemini: empty response returned")
	}
	var parsed EnrichmentResult
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return EnrichmentResult{}, fmt.Errorf("gemini: failed to parse response: %w", err)
	}
	return parsed, nil
}

func formatExperience(experiences []string) string {
	var out strings.Builder
	for i, exp := range experiences {
		fmt.Fprintf(&out, "%d. %s\n", i+1, exp)
	}
	return out.String()
}

func classifyGeminiError(err error) error {
	var apiErr genai.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Code {
		case 400, 404:
			return fmt.Errorf("invalid request: %w: %w", ErrPermanent, apiErr)
		case 401, 403:
			return fmt.Errorf("authentication failed: %w: %w", ErrPermanent, apiErr)
		case 429:
			return fmt.Errorf("gemini rate limit exceeded: %w", apiErr)
		case 500, 503:
			return fmt.Errorf("gemini service unavailable: %w", apiErr)
		default:
			return fmt.Errorf("unexpected gemini API error (code %d): %w", apiErr.Code, apiErr)
		}
	}
	// 2. Fallback for network timeouts, context cancellations, etc.
	return fmt.Errorf("gemini: network or generic generation failure: %w", err)
}
