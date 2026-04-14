package main

import (
	"context"
	"log"

	"github.com/samnodier/jobradar/internal/convert"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/fetcher"
	"github.com/samnodier/jobradar/internal/stringutils"
)

func (cfg *apiConfig) scrapeRemoteOK(ctx context.Context) {
	url := "https://remoteok.com/api"
	jobs, err := fetcher.FetchRemoteOKJobs(url)
	if err != nil {
		log.Printf("error fetching the jobs form remoteok: %v", err)
		return
	}

	remote := true
	for _, rJob := range jobs {
		cleanTitle := stringutils.Sanitize(rJob.Position)
		cleanDesc := stringutils.Sanitize(rJob.Description)

		_, err := cfg.db.CreateJob(ctx, database.CreateJobParams{

			ExternalID:  rJob.ID,
			Source:      "remoteok",
			Title:       cleanTitle,
			Company:     rJob.Company,
			Description: convert.ToNullString(cleanDesc),
			Url:         rJob.URL,
			SalaryMin:   convert.ToNullInt32(rJob.SalaryMin),
			SalaryMax:   convert.ToNullInt32(rJob.SalaryMax),
			Currency:    convert.ToNullString(""),
			Location:    convert.ToNullString(rJob.Location),
			IsRemote:    convert.ToNullBool(&remote),
			Skills:      rJob.Tags,
			LogoUrl:     convert.ToNullString(rJob.Logo),
		})
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				continue
			}
			log.Printf("could not save job %s: %v", rJob.ID, err)
		}
	}
	log.Printf("RemoteOK scraping finished!")
}
