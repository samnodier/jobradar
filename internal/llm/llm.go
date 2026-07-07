// Package llm provides a simple interface for LLMs to enrich job postings with summaries
package llm

import (
	"context"
	"fmt"
	"strings"
)

type EnrichmentInput struct {
	JobTitle       string
	JobDescription string
	JobSkills      []string
	JobLocation    string

	DesiredRoles    []string
	UserSkills      []string
	UserExperiences []string
}

type EnrichmentResult struct {
	Summary string `json:"summary"`
}

type ExtractionInput struct {
	PageText string
}

type ExtractionResult struct {
	Title       string   `json:"title"`
	CompanyName string   `json:"company_name"`
	Description string   `json:"description"`
	SalaryMin   *int32   `json:"salary_min"`
	SalaryMax   *int32   `json:"salary_max"`
	Currency    *string  `json:"currency"`
	JobLocation *string  `json:"job_location"`
	IsRemote    *bool    `json:"is_remote"`
	Skills      []string `json:"skills"`
}

type Enricher interface {
	Enrich(ctx context.Context, input EnrichmentInput) (EnrichmentResult, error)
}

type Extractor interface {
	Extract(ctx context.Context, input ExtractionInput) (ExtractionResult, error)
}

// buildEnrichPrompt renders the user-content half of an enrichment request.
// Shared by all providers so they reason over identical inputs.
func buildEnrichPrompt(input EnrichmentInput) string {
	return fmt.Sprintf(
		`
		Job Title: %s
		Job Location: %s
		Job Description: %s

		Candidate Desired Roles: %s
		Candidate Skills: %s
		Candidate Experience: %s
		`,
		input.JobTitle,
		input.JobLocation,
		input.JobDescription,
		strings.Join(input.DesiredRoles, ", "),
		strings.Join(input.UserSkills, ", "),
		formatExperience(input.UserExperiences),
	)
}
