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
	Title       string
	CompanyName string
	Description string
	SalaryMin   *int32
	SalaryMax   *int32
	Currency    *string
	JobLocation *string
	IsRemote    *bool
	Skills      []string
	LogoURL     *string
}

type Enricher interface {
	Enrich(ctx context.Context, input EnrichmentInput) (EnrichmentResult, error)
}

type Extractor interface {
	Extract(ctx context.Context, input ExtractionInput) (ExtractionResult, error)
}
