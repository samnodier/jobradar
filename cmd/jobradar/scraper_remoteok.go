package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/samnodier/jobradar/internal/convert"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/sources"
	"github.com/samnodier/jobradar/internal/stringutils"
)

type MatchJobPayload struct {
	JobID    string `json:"job_id"`
	UserID   string `json:"user_id"`
	Revision int64  `json:"revision,omitempty"`
}
type EnrichJobPayload struct {
	JobID  string `json:"job_id"`
	UserID string `json:"user_id"`
}

func (cfg *apiConfig) scrapeSource(ctx context.Context, source sources.Source) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	jobs, err := source.Fetch(ctx)
	count, skippedOld := 0, 0
	// Don't re-insert postings retention would delete on its next run; otherwise feeds that
	// keep returning old jobs would recreate them every cycle.
	cutoff, retain := cfg.retentionCutoff(time.Now())
	if err == nil {
		for _, j := range jobs {
			if retain && j.PostedAt != nil && j.PostedAt.Before(cutoff) {
				skippedOld++
				continue
			}
			if j.Skills == nil {
				j.Skills = []string{}
			}
			_, saveErr := cfg.db.UpsertPublicJob(ctx, database.UpsertPublicJobParams{
				Category: convert.ToNullString(j.Category), EmploymentType: convert.ToNullString(j.EmploymentType), ExternalID: j.ID, JobSource: j.Source, Title: stringutils.SanitizeStrict(stringutils.FixMojibake(j.Title)),
				CompanyName: stringutils.SanitizeStrict(stringutils.FixMojibake(j.Company)), Description: convert.ToNullString(stringutils.SanitizeDescription(stringutils.FixMojibake(j.Description))),
				SourceUrl: j.URL, SalaryMin: j.SalaryMin, SalaryMax: j.SalaryMax, Currency: convert.ToNullString(j.Currency), JobLocation: convert.ToNullString(j.Location), IsRemote: j.Remote, Skills: j.Skills, LogoUrl: convert.ToNullString(j.Logo), PostedAt: j.PostedAt,
			})
			if saveErr != nil {
				err = fmt.Errorf("save source job: %w", saveErr)
				continue
			}
			count++
		}
	}
	status := "healthy"
	var message *string
	var success *time.Time
	if err != nil {
		status = "degraded"
		m := err.Error()
		message = &m
	} else {
		now := time.Now()
		success = &now
	}
	healthCtx, healthCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer healthCancel()
	_, healthErr := cfg.db.UpdateServiceHealth(healthCtx, database.UpdateServiceHealthParams{ServiceName: source.Name(), LastSuccessAt: success, ServiceStatus: status, LastError: message, JobCountLastRun: convert.ToNullInt32(count)})
	if healthErr != nil {
		log.Printf("source health write failed: %v", healthErr)
	}
	log.Printf("source=%s saved=%d skipped_old=%d status=%s", source.Name(), count, skippedOld, status)
	return err
}
