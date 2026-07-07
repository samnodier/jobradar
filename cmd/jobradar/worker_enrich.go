package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/llm"
	"github.com/samnodier/jobradar/internal/queue"
)

func (cfg *apiConfig) handleEnrichJob(ctx context.Context, qJob *queue.Job) error {
	var payload EnrichJobPayload
	if err := json.Unmarshal(qJob.Payload, &payload); err != nil {
		log.Printf("handleEnrichJob: failed to unmarshal payload: %v", err)
		return nil
	}

	jobID, err := uuid.Parse(payload.JobID)
	if err != nil {
		log.Printf("handleEnrichJob: failed to parse job UUID: %v", err)
		return nil
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		log.Printf("handleEnrichJob: failed to parse user UUID: %v", err)
		return nil
	}

	// Re-check the key defensively — the producer's gate can go stale
	keys, err := cfg.selectProviderKeys(ctx, userID)
	if err != nil {
		if errors.Is(err, errNoAPIKey) {
			log.Printf("handleEnrichJob: no api key found for user %s", userID)
			return nil // nothing to retry — the user removed their key
		}
		return err // DB/decrypt trouble — let the backoff retry it
	}

	// Fetch the job
	job, err := cfg.db.GetJobByID(ctx, database.GetJobByIDParams{
		// Pass the real user ID: GetJobByID only returns public-or-owned rows,
		// so uuid.Nil would hide this user's imported (private) jobs and the
		// enrichment would silently drop with "not found".
		UserID: userID,
		ID:     jobID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("Job %s not found in DB skipping matching", jobID)
			return nil
		}
		return err
	}

	// Compile the user's profile
	// Desired Roles
	desiredRolesRow, err := cfg.db.GetUserDesiredRoles(ctx, userID)
	if err != nil {
		return err
	}
	var desiredRoles []string
	for _, r := range desiredRolesRow {
		desiredRoles = append(desiredRoles, r.RoleTitle)
	}

	// User skills
	dbUserSkills, err := cfg.db.GetUserSkillsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	// Experiences skills
	dbExpSkills, err := cfg.db.GetExperiencesSkillsByUserID(ctx, userID)
	if err != nil {
		return err
	}

	var userSkills []string
	seenSkills := make(map[string]struct{})
	for _, userSkill := range dbUserSkills {
		lower := strings.ToLower(userSkill.SkillName)
		if _, ok := seenSkills[lower]; !ok {
			seenSkills[lower] = struct{}{}
			userSkills = append(userSkills, userSkill.SkillName)
		}
	}
	for _, expSkill := range dbExpSkills {
		lower := strings.ToLower(expSkill.SkillName)
		if _, ok := seenSkills[lower]; !ok {
			seenSkills[lower] = struct{}{}
			userSkills = append(userSkills, expSkill.SkillName)
		}
	}

	// Experiences
	expsRows, err := cfg.db.GetExperiencesByUserID(ctx, userID)
	if err != nil {
		return err
	}
	var userExperiences []string
	for _, exp := range expsRows {
		var text strings.Builder
		text.WriteString(exp.RoleTitle)
		text.WriteString(" at ")
		text.WriteString(exp.CompanyName)
		if exp.Description != nil {
			text.WriteString(" ")
			text.WriteString(*exp.Description)
		}
		for _, ach := range exp.Achievements {
			text.WriteString(" ")
			text.WriteString(ach)
		}
		userExperiences = append(userExperiences, text.String())
	}

	// Run the enrinch
	jobDescription := ""
	if job.Description != nil {
		jobDescription = *job.Description
	}
	jobLocation := ""
	if job.JobLocation != nil {
		jobLocation = *job.JobLocation
	}

	input := llm.EnrichmentInput{
		JobTitle:       job.Title,
		JobDescription: jobDescription,
		JobSkills:      job.Skills,
		JobLocation:    jobLocation,

		DesiredRoles:    desiredRoles,
		UserSkills:      userSkills,
		UserExperiences: userExperiences,
	}

	// Try each provider in priority order; fall back to the next on failure
	var result llm.EnrichmentResult
	var enrichErr error
	for _, pk := range keys {
		var llmEnricher llm.Enricher
		llmEnricher, enrichErr = cfg.newEnricher(ctx, pk.provider, pk.apiKey)
		if enrichErr != nil {
			log.Printf("handleEnrichJob: failed to create %s enricher for user %s: %v", pk.provider, userID, enrichErr)
			continue
		}
		result, enrichErr = llmEnricher.Enrich(ctx, input)
		if enrichErr == nil {
			break
		}
		log.Printf("handleEnrichJob: %s enrich failed for job %s user %s (trying next provider if any): %v", pk.provider, jobID, userID, enrichErr)
	}
	if enrichErr != nil {
		if errors.Is(enrichErr, llm.ErrPermanent) {
			log.Printf("handleEnrichJob: permanent failure for job %s user %s, dropping: %v", jobID, userID, enrichErr)
			return nil // drop it, don't retry
		}
		return enrichErr // transient — the queue's backoff retries the job
	}

	err = cfg.db.UpdateMatchEnrichment(ctx, database.UpdateMatchEnrichmentParams{
		UserID:    userID,
		JobID:     jobID,
		AiSummary: &result.Summary,
	})
	if err != nil {
		log.Printf("handleEnrichJob: failed to update match enrichment for job %s for user %s: %v", jobID, userID, err)
		return nil
	}

	return nil
}
