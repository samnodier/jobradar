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

	geminiKeyPointer, err := cfg.db.GetGeminiKeyByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if geminiKeyPointer.EncryptedGeminiApiKey == nil {
		log.Printf("handleEnrichJob: no gemini key found for user %s", userID)
		return nil
	}

	geminiKey, err := cfg.crypto.Decrypt(*geminiKeyPointer.EncryptedGeminiApiKey)
	if err != nil {
		log.Printf("handleEnrichJob: failed to decrypt gemini key for user %s: %v", userID, err)
		return nil
	}

	llmEnricher, err := cfg.newEnricher(ctx, geminiKey)
	if err != nil {
		log.Printf("handleEnrichJob: failed to create enricher for user %s: %v", userID, err)
		return nil
	}

	// Fetch the job
	job, err := cfg.db.GetJobByID(ctx, database.GetJobByIDParams{
		// GetJobByID needs a user ID because of the left join with saved_jobs/applications.
		// We can pass a blank UUID since we only need the raw job description/title details
		UserID: uuid.Nil,
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
		text.WriteString(exp.RoleTitle + " at " + exp.CompanyName)
		if exp.Description != nil {
			text.WriteString(" " + *exp.Description)
		}
		for _, ach := range exp.Achievements {
			text.WriteString(" " + ach)
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

	result, err := llmEnricher.Enrich(ctx, input)
	if err != nil {
		if errors.Is(err, llm.ErrPermanent) {
			log.Printf("handleEnrichJob: permanent failure for job %s user %s, dropping: %v", jobID, userID, err)
			return nil // drop it, don't retry
		}
		return err // transient — Asynq backoff retries it
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
