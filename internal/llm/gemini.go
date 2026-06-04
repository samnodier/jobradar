package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type geminiEnricher struct {
	client *genai.Client
}

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

const systemInstructions = `You are a career assistant job-match analyst. Given a job posting and a candidate's profile,
write a concise 2-3 sentence summary of how well the candidate fits the role.
Be specific — mention relevant skills and experience by name.
Do not invent skills or experience the candidate does not have.
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
		`, input.JobTitle, input.JobLocation, input.JobDescription, strings.Join(input.UserRoles, ", "), strings.Join(input.UserSkills, ", "), formatExperience(input.UserExperiences),
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
		return EnrichmentResult{}, fmt.Errorf("gemini: generate failed: %w", err)
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
