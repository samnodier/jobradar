package main

import (
	"context"
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

func (cfg *apiConfig) scrapeRemoteOK(ctx context.Context) error {
	log.Println("Starting RemoteOK scrape...")

	url := "https://remoteok.com/api"
	jobs, err := fetcher.FetchRemoteOKJobs(url)
	if err != nil {
		cfg.db.UpdateServiceHealth(ctx, database.UpdateServiceHealthParams{
			ServiceName:     "remoteok",
			LastSuccessAt:   convert.ToNullTime(time.Time{}),
			ServiceStatus:   "failing",
			LastError:       convert.ToNullString(err.Error()),
			JobCountLastRun: convert.ToNullInt32(0),
		})
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

			// Push matching task to queue
			jobPayload := []byte(createdJob.ID.String())
			err = cfg.queue.Enqueue(ctx, &queue.Job{
				ID:       createdJob.ID.String(),
				Type:     queue.JobMatchJob,
				Payload:  jobPayload,
				MaxRetry: 3,
			})
			if err != nil {
				log.Printf("failed to enqueue matching job for %s: %v", createdJob.ID, err)
			}
		}
	}
	log.Printf("RemoteOK scraping finished!")

	cfg.db.UpdateServiceHealth(ctx, database.UpdateServiceHealthParams{
		ServiceName:     "remoteok",
		LastSuccessAt:   convert.ToNullTime(time.Now()),
		ServiceStatus:   "healthy",
		LastError:       convert.ToNullString(""),
		JobCountLastRun: convert.ToNullInt32(count),
	})
	return nil
}
