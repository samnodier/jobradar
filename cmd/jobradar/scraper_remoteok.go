package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/samnodier/jobradar/internal/convert"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/fetcher"
	"github.com/samnodier/jobradar/internal/stringutils"
)

func (cfg *apiConfig) scrapeRemoteOK(ctx context.Context) {
	log.Println("Starting RemoteOK scrape...")

	url := "https://remoteok.com/api"
	jobs, err := fetcher.FetchRemoteOKJobs(url)
	if err != nil {
		cfg.db.UpdateServiceHealth(ctx, database.UpdateServiceHealthParams{
			ServiceName:     "remoteok",
			LastSuccessAt:   convert.ToNullTime(time.Time{}),
			Status:          "failing",
			LastError:       convert.ToNullString(err.Error()),
			JobCountLastRun: convert.ToNullInt32(0),
		})
		log.Printf("error fetching the jobs form remoteok: %v", err)
		return
	}

	count := 0

	remote := true
	for _, rJob := range jobs {
		cleanTitle := stringutils.SanitizeStrict(rJob.Position)

		_, err := cfg.db.CreateJob(ctx, database.CreateJobParams{

			ExternalID:  rJob.ID,
			Source:      "remoteok",
			Title:       cleanTitle,
			Company:     rJob.Company,
			Description: convert.ToNullString(stringutils.SanitizeDescription(rJob.Description)),
			Url:         rJob.URL,
			SalaryMin:   convert.ToNullInt32(rJob.SalaryMin),
			SalaryMax:   convert.ToNullInt32(rJob.SalaryMax),
			Currency:    convert.ToNullString(""),
			Location:    convert.ToNullString(rJob.Location),
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
		}
	}
	log.Printf("RemoteOK scraping finished!")

	cfg.db.UpdateServiceHealth(ctx, database.UpdateServiceHealthParams{
		ServiceName:     "remoteok",
		LastSuccessAt:   convert.ToNullTime(time.Now()),
		Status:          "healthy",
		LastError:       convert.ToNullString(""),
		JobCountLastRun: convert.ToNullInt32(count),
	})
}
