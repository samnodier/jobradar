package main

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/matcher"
	"github.com/samnodier/jobradar/internal/queue"
)

func (cfg *apiConfig) handleMatchJob(ctx context.Context, qJob *queue.Job) error {
	jobIDStr := string(qJob.Payload)
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		return errors.New("invalid job UUID in payload")
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

	// Fetch the primary user the first user in our system
	user, err := cfg.db.GetFirstUser(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("No users registered in the system yet. Skipping match")
			return nil
		}
		return err
	}

	// Compile the user's profile
	// Desired Roles
	desiredRolesRow, err := cfg.db.GetUserDesiredRoles(ctx, user.ID)
	if err != nil {
		return err
	}
	var desiredRoles []string
	for _, r := range desiredRolesRow {
		desiredRoles = append(desiredRoles, r.RoleTitle)
	}

	// Skills
	skillsRow, err := cfg.db.GetUserSkillsByUserID(ctx, user.ID)
	if err != nil {
		return err
	}
	var userSkills []string
	for _, s := range skillsRow {
		userSkills = append(userSkills, s.SkillName)
	}

	// Experiences
	expsRows, err := cfg.db.GetExperiencesByUserID(ctx, user.ID)
	if err != nil {
		return err
	}
	var userExps []string
	for _, exp := range expsRows {
		text := exp.RoleTitle + " at " + exp.CompanyName
		if exp.Description != nil {
			text += " " + *exp.Description
		}
		for _, ach := range exp.Achievements {
			text += " " + ach
		}
		userExps = append(userExps, text)
	}

	// Run the match algorithm
	jobDesc := ""
	if job.Description != nil {
		jobDesc = *job.Description
	}

	// Jaro-Winkler threshold = 0.55
	result := matcher.MatchJob(job.Title, jobDesc, desiredRoles, userSkills, userExps, 0.55)
	var scoreVal int32
	var scorePtr *int32
	var summaryVal string
	var summaryPtr *string
	var matchedSkills []string
	var missingSkills []string
	if !result.Skipped {
		scoreVal = int32(result.Score)
		scorePtr = &scoreVal
		matchedSkills = result.MatchedSkills

		summaryVal = "Algorithmic match completed"
		summaryPtr = &summaryVal
		matchedSkills = result.MatchedSkills
		// Compute the missing skills (user skills that weren't matched)
		matchedSet := make(map[string]bool)
		for _, ms := range result.MatchedSkills {
			matchedSet[ms] = true
		}
		for _, s := range userSkills {
			if !matchedSet[s] {
				missingSkills = append(missingSkills, s)
			}
		}

	} else {
		scoreVal = 0
		scorePtr = &scoreVal
		summaryVal = "Skipped due to low title match"
		summaryPtr = &summaryVal
	}

	// Update Job in the database with match scores
	err = cfg.db.UpdateJobMatchingResult(ctx, database.UpdateJobMatchingResultParams{
		ID:            job.ID,
		MatchScore:    scorePtr,
		AiSummary:     summaryPtr,
		MatchedSkills: matchedSkills,
		MissingSkills: missingSkills,
	})
	if err != nil {
		return err
	}
	log.Printf("Job %s matched for user %s: score=%d, skipped =%v", job.ID, user.Email, scoreVal, result.Skipped)
	return nil
}
