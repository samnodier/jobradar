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

type Enricher interface {
	Enrich(ctx context.Context, input EnrichmentInput) (EnrichmentResult, error)
}
