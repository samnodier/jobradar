// Package llm provides a simple interface for LLMs to enrich job postings with summaries
package llm

import "context"

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
	LogoURL     *string  `json:"logo_url"`
}

type Enricher interface {
	Enrich(ctx context.Context, input EnrichmentInput) (EnrichmentResult, error)
}

type Extractor interface {
	Extract(ctx context.Context, input ExtractionInput) (ExtractionResult, error)
}
