package main

import (
	"context"
	"encoding/json"
	"errors"
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
		return errors.New("failed to unmarshall payload")
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

	// Skills
	skillsRow, err := cfg.db.GetUserSkillsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	var userSkills []string
	for _, s := range skillsRow {
		userSkills = append(userSkills, s.SkillName)
	}

	// Experiences
	expsRows, err := cfg.db.GetExperiencesByUserID(ctx, userID)
	if err != nil {
		return err
	}
	var userExps []string
	for _, exp := range expsRows {
		var text strings.Builder
		text.WriteString(exp.RoleTitle + " at " + exp.CompanyName)
		if exp.Description != nil {
			text.WriteString(" " + *exp.Description)
		}
		for _, ach := range exp.Achievements {
			text.WriteString(" " + ach)
		}
		userExps = append(userExps, text.String())
	}

	// Run the match algorithm
	jobDesc := ""
	if job.Description != nil {
		jobDesc = *job.Description
	}

	// Jaro-Winkler threshold = 0.55
	result := matcher.MatchJob(job.Title, jobDesc, desiredRoles, userSkills, userExps, 0.55)
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
		// Compute the missing skills (user skills that weren't matched)
		matchedSet := make(map[string]bool)
		for _, ms := range matchedSkills {
			matchedSet[ms] = true
		}
		for _, s := range userSkills {
			if !matchedSet[s] {
				missingSkills = append(missingSkills, s)
			}
		}

	} else {
		score = 0
	}

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
	log.Printf("Job %s matched for user %s: score=%d, skipped =%v", job.ID, userID, score, result.Skipped)
	return nil
}
