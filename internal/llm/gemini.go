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

type geminiExtractor struct {
	client *genai.Client
}

var (
	_            Enricher  = (*geminiEnricher)(nil)
	_            Extractor = (*geminiExtractor)(nil)
	ErrPermanent           = errors.New("permanent enrichment failure")
)

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

// NewGeminiExtractor creates a new instance of the GeminiExtractor
func NewGeminiExtractor(ctx context.Context, apiKey string) (Extractor, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to create client: %w", err)
	}
	return &geminiExtractor{client: client}, nil
}

const EnrichInstructions = `You are a career assistant speaking directly to a job-seeker about a specific role.
Write a concise 2-3 sentence summary addressing the user as "you"
	— for example: "You have strong experience in X" or "You're missing Y which this role requires."
  - Be specific
	— mention relevant skills and experience by name.
	Do not invent skills or experience the user does not have.
	Respond only with valid JSON matching the schema provided. No markdown, no preamble.`

const ExtractInstructions = `You are a precise job-posting data extractor.
	You are given the raw text content of a single job posting page.
	Your only task is to extract structured fields from that text and return them as JSON matching the provided schema.
	Rules:
	- Extract only what is explicitly stated in the text. Never infer, guess, or fabricate a value.
	- For any field not clearly present in the text, return null. Do not substitute a placeholder, an empty string, or a zero.
	- "title" is the job title only
	— not the company name or a marketing tagline.
	- "company_name" is the hiring organization's name.
	- "description" should capture the role's responsibilities and requirements as written, in plain text. Do not summarize or editorialize.
	- "skills" is an array of concrete technologies, tools, or competencies named as requirements (e.g. "Go", "PostgreSQL", "Kubernetes").
	Return an empty array if none are stated. Do not invent skills the posting does not mention.
	- "salary_min" and "salary_max" are integers with no currency symbols, separators, or text. Put the currency in "currency" (e.g. "USD").
	If only one salary figure is given, set both to that figure. If no salary is stated, return null for all three.
	- "is_remote" is true only if the posting explicitly describes the role as remote.
	If it is on-site, hybrid, or unspecified, return false or null accordingly — do not assume.
	- "job_location" is the location as written (e.g. "Berlin, Germany" or "Remote — US").
Respond only with JSON matching the schema. Nomarkdown, no code fences, no preamble.`

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
			SystemInstruction: genai.NewContentFromText(EnrichInstructions, ""),
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

func (g *geminiExtractor) Extract(ctx context.Context, input ExtractionInput) (ExtractionResult, error) {
	userContent := input.PageText
	result, err := g.client.Models.GenerateContent(
		ctx,
		model,
		genai.Text(userContent),
		&genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(ExtractInstructions, ""),
			ResponseMIMEType:  "application/json",
			ResponseSchema: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"title":        {Type: genai.TypeString},
					"company_name": {Type: genai.TypeString},
					"description":  {Type: genai.TypeString},
					"salary_min":   {Type: genai.TypeInteger},
					"salary_max":   {Type: genai.TypeInteger},
					"currency":     {Type: genai.TypeString},
					"job_location": {Type: genai.TypeString},
					"is_remote":    {Type: genai.TypeBoolean},
					"skills": {
						Type:  genai.TypeArray,
						Items: &genai.Schema{Type: genai.TypeString},
					},
				},
				Required: []string{
					"title", "company_name", "description",
				},
				Nullable: genai.Ptr(true),
			},
		},
	)
	if err != nil {
		return ExtractionResult{}, classifyGeminiError(err)
	}

	raw := result.Text()
	if raw == "" {
		return ExtractionResult{}, fmt.Errorf("gemini: empty response returned")
	}
	var parsed ExtractionResult
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return ExtractionResult{}, fmt.Errorf("gemini: failed to parse response: %w", err)
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
