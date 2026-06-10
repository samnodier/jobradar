package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samnodier/jobradar/internal/convert"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/matcher"
	"github.com/samnodier/jobradar/internal/queue"
)

/*
handleMatchJob is the handler
1. Parses the job UUID from the payload
2. Fetches the job from the DB
3. Fetches the user's profile (roles, skills, experiences)
4. Runs matcher.MatchJob()
5. Writes the result back with UpsertMatch
*/
func (cfg *apiConfig) handleMatchJob(ctx context.Context, qJob *queue.Job) error {
	var matchJobPayload MatchJobPayload
	err := json.Unmarshal(qJob.Payload, &matchJobPayload)
	if err != nil {
		log.Printf("handleMatchJob: failed to unmarshall payload")
		return nil
	}
	jobIDStr, userIDStr := matchJobPayload.JobID, matchJobPayload.UserID
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		return errors.New("invalid job UUID in payload")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return errors.New("invalid user UUID in payload")
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

	userSkills := make(map[string]float64)
	seenSkills := make(map[string]struct{})
	for _, userSkill := range dbUserSkills {
		lower := strings.ToLower(userSkill.SkillName)
		if _, ok := seenSkills[lower]; !ok {
			seenSkills[lower] = struct{}{}
			userSkills[userSkill.SkillName] = proficiencyToWeight(userSkill.Proficiency)
		}
	}
	for _, expSkill := range dbExpSkills {
		lower := strings.ToLower(expSkill.SkillName)
		if _, ok := seenSkills[lower]; !ok {
			seenSkills[lower] = struct{}{}
			userSkills[expSkill.SkillName] = 0.5
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

	// Run the match algorithm
	jobDesc := ""
	if job.Description != nil {
		jobDesc = *job.Description
	}

	// Jaro-Winkler threshold = 0.55
	result := matcher.MatchJob(job.Title, jobDesc, desiredRoles, userSkills, job.Skills, userExperiences, 0.55)
	var score int32
	titleScore := pgtype.Float8{}
	skillScore := pgtype.Float8{}
	expScore := pgtype.Float8{}
	matchedSkills := []string{}
	missingSkills := []string{}
	notEnriched := false
	if !result.Skipped {
		score = int32(result.Score)
		titleScore = convert.ToFloat8(result.TitleScore)
		skillScore = convert.ToFloat8(result.SkillScore)
		expScore = convert.ToFloat8(result.ExpScore)
		matchedSkills = result.MatchedSkills
		// Compute the missing skills: the job's required skills the user
		// hasn't demonstrated (case-insensitive). These are the gaps the
		// user would need to fill to qualify.
		// NOTE: (design debt) matchedSet is description-derived — it holds the
		// user's skills found in the job's prose, not the user's full skill
		// set. So a skill the user genuinely has can be flagged "missing" if
		// the description never spells it out. Revisit the canonical skill
		// universe before the LLM prompt consumes matched/missing skills.
		matchedSet := make(map[string]bool)
		for _, ms := range matchedSkills {
			matchedSet[strings.ToLower(ms)] = true
		}
		for _, s := range job.Skills {
			jobSkillLower := strings.ToLower(s)
			if !matchedSet[jobSkillLower] {
				missingSkills = append(missingSkills, s)
			}
		}

	} else {
		score = 0
	}
	// TODO: (scale / re-match correctness) we currently upsert a score=0 row
	// even for skipped (non-matching) jobs, so user_job_matches grows to
	// jobs × users including all non-matches. The cleaner design is to NOT
	// write skips. BUT do not ship "don't write skips" alone — if a job that
	// previously scored above threshold drops below it on a profile re-match,
	// the stale high-score row would linger. "Don't write skips" must arrive
	// together with a DeleteMatch query. Fold this into the re-match-scale
	// work (coalescing → batch RematchUser job), see LEARNINGS.md.

	// Update Job in the database with match scores
	err = cfg.db.UpsertMatch(ctx, database.UpsertMatchParams{
		UserID:          userID,
		JobID:           jobID,
		MatchScore:      &score,
		TitleScore:      titleScore,
		SkillScore:      skillScore,
		ExperienceScore: expScore,
		MatchedSkills:   matchedSkills,
		MissingSkills:   missingSkills,
		AiSummary:       nil,
		IsEnriched:      &notEnriched,
	})
	if err != nil {
		return err
	}
	// If this match scored high enough And the user has a Gemini key
	// Push an enrich job
	if score >= cfg.aiMatchThreshold {
		user, err := cfg.db.GetUserByID(ctx, userID)
		if err != nil {
			log.Printf("handlerMatchJob: failed to fetch user %s: %v", userID, err)
			return nil
		}
		if user.HasGeminiKey {
			enrichPayload, err := json.Marshal(EnrichJobPayload{
				JobID:  jobID.String(),
				UserID: userID.String(),
			})
			if err != nil {
				log.Printf("handleMatchJob: failed to marshal enrich payload: %v", err)
				return nil
			}
			err = cfg.queue.Enqueue(ctx, &queue.Job{
				ID:       jobID.String() + "-" + userID.String() + "-enrich",
				Type:     queue.JobEnrichMatch,
				Payload:  enrichPayload,
				MaxRetry: 3,
			})
			if err != nil {
				log.Printf("handleMatchJob: failed to enqueue enrich job: %v", err)
			}
		}

	}

	// Don't log for every match
	// log.Printf("Job %s matched for user %s: score=%d, skipped =%v", job.ID, userID, score, result.Skipped)
	return nil
}

/*
enqueueMatchJobsForUser re-enqueues a match job for the given user against every
job in the DB. Used after a profile change (skills or desired roles) makes the
user's existing match scores stale. It only enqueues — the worker (handleMatchJob)
does the actual fetching and scoring later.
1. Fetches all job IDs from the DB
2. For each job, pushes a {job_id, user_id} payload onto the queue
Best-effort: per-job failures are logged and skipped, not fatal.
*/
func (cfg *apiConfig) enqueueMatchJobsForUser(ctx context.Context, userID uuid.UUID) error {
	// Fetch all job IDs
	jobsIDs, err := cfg.db.GetAllJobIDs(ctx)
	if err != nil {
		return fmt.Errorf("enqueueMatchJobsForUser: error fetching job IDs for rematching: %w", err)
	}
	for _, jobID := range jobsIDs {
		payloadData := MatchJobPayload{
			JobID:  jobID.String(),
			UserID: userID.String(),
		}
		matchJobPayload, err := json.Marshal(payloadData)
		if err != nil {
			log.Printf("enqueueMatchJobsForUser: Failed to marshal match job payload: %v", err)
			continue
		}
		// Push matching task to queue
		err = cfg.queue.Enqueue(ctx, &queue.Job{
			ID:       jobID.String() + "-" + userID.String(),
			Type:     queue.JobMatchJob,
			Payload:  matchJobPayload,
			MaxRetry: 3,
		})
		if err != nil {
			log.Printf("enqueueMatchJobsForUser: failed to enqueue matching job for %s: %v", jobID, err)
		}
	}
	return nil
}

func proficiencyToWeight(p *string) float64 {
	if p == nil {
		return 0.5
	}
	proficiency := strings.ToLower(strings.TrimSpace(*p))
	switch proficiency {
	case "beginner":
		return 0.3
	case "intermediate":
		return 0.5
	case "advanced":
		return 0.8
	case "expert":
		return 1.0
	default:
		return 0.5
	}
}
