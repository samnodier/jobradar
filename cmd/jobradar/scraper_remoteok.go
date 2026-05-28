package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/samnodier/jobradar/internal/convert"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/fetcher"
	"github.com/samnodier/jobradar/internal/queue"
	"github.com/samnodier/jobradar/internal/stringutils"
)

type MatchJobPayload struct {
	JobID  string `json:"job_id"`
	UserID string `json:"user_id"`
}

func (cfg *apiConfig) scrapeRemoteOK(ctx context.Context) error {
	log.Println("Starting RemoteOK scrape...")

	url := "https://remoteok.com/api"
	jobs, err := fetcher.FetchRemoteOKJobs(url)
	if err != nil {
		// Log the health update error, but still return the primary fetch error below
		_, healthErr := cfg.db.UpdateServiceHealth(ctx, database.UpdateServiceHealthParams{
			ServiceName:     "remoteok",
			LastSuccessAt:   convert.ToNullTime(time.Time{}),
			ServiceStatus:   "failing",
			LastError:       convert.ToNullString(err.Error()),
			JobCountLastRun: convert.ToNullInt32(0),
		})
		if healthErr != nil {
			log.Printf("warning: failed to update service health to failing: %v", healthErr)
		}

		log.Printf("error fetching the jobs form remoteok: %v", err)
		return fmt.Errorf("failed to fetch RemoteOK jobs: %w", err)
	}

	count := 0
	remote := true

	for _, rJob := range jobs {
		cleanTitle := stringutils.SanitizeStrict(rJob.Position)

		createdJob, err := cfg.db.CreateJob(ctx, database.CreateJobParams{
			ExternalID:  rJob.ID,
			JobSource:   "remoteok",
			Title:       cleanTitle,
			CompanyName: rJob.Company,
			Description: convert.ToNullString(stringutils.SanitizeDescription(rJob.Description)),
			SourceUrl:   rJob.URL,
			SalaryMin:   convert.ToNullInt32(rJob.SalaryMin),
			SalaryMax:   convert.ToNullInt32(rJob.SalaryMax),
			Currency:    convert.ToNullString(""),
			JobLocation: convert.ToNullString(rJob.Location),
			IsRemote:    &remote,
			Skills:      rJob.Tags,
			LogoUrl:     convert.ToNullString(rJob.Logo),
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			log.Printf("could not save job %s: %v", rJob.ID, err)
		} else {
			count++

			// For each new job saved to db, get all users and run the matcher
			userIDs, err := cfg.db.GetAllUsers(ctx)
			if err != nil {
				log.Printf("could not get all users for matching job %s: %v", rJob.ID, err)
				continue
			}
			for _, userID := range userIDs {

				// Push matching task to queue
				payloadData := MatchJobPayload{
					JobID:  createdJob.ID.String(),
					UserID: userID.String(),
				}
				matchJobPayload, err := json.Marshal(payloadData)
				if err != nil {
					log.Printf("Failed to marshal match job payload: %v", err)
					continue
				}
				err = cfg.queue.Enqueue(ctx, &queue.Job{
					ID:       createdJob.ID.String() + "-" + userID.String(),
					Type:     queue.JobMatchJob,
					Payload:  matchJobPayload,
					MaxRetry: 3,
				})
				if err != nil {
					log.Printf("failed to enqueue matching job for %s: %v", createdJob.ID, err)
				}
			}
		}
	}
	log.Printf("RemoteOK scraping finished!")

	// If we got jobs from the API, but successfully saved 0 of them, something is wrong internally.
	if len(jobs) > 0 && count == 0 {
		_, healthErr := cfg.db.UpdateServiceHealth(ctx, database.UpdateServiceHealthParams{
			ServiceName:     "remoteok",
			LastSuccessAt:   convert.ToNullTime(time.Time{}), // Don't update success time
			ServiceStatus:   "degraded",                      // Or "failing"
			LastError:       convert.ToNullString("all jobs failed to insert during this run"),
			JobCountLastRun: convert.ToNullInt32(0),
		})
		if healthErr != nil {
			log.Printf("warning: failed to update service health to degraded: %v", healthErr)
		}
		return fmt.Errorf("scraped %d jobs but failed to save any to the database", len(jobs))
	}

	// Success case
	_, healthErr := cfg.db.UpdateServiceHealth(ctx, database.UpdateServiceHealthParams{
		ServiceName:     "remoteok",
		LastSuccessAt:   convert.ToNullTime(time.Now()),
		ServiceStatus:   "healthy",
		LastError:       convert.ToNullString(""),
		JobCountLastRun: convert.ToNullInt32(count),
	})
	if healthErr != nil {
		log.Printf("warning: failed to update service health to healthy: %v", healthErr)
	}

	return nil
}
