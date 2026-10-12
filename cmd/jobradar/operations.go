package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/mailer"
)

type operationState struct {
	Status         string    `json:"status"`
	TargetRevision string    `json:"target_revision"`
	CheckedAt      time.Time `json:"checked_at"`
	CompletedAt    time.Time `json:"completed_at"`
}
type operationAlert struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

func stateDirectory() string {
	if dir := os.Getenv("OPERATIONS_STATE_DIR"); dir != "" {
		return dir
	}
	return "deploy-state"
}
func readOperationState(name string) operationState {
	var state operationState
	// Files are generated locally, but still bound reads and expose an allowlisted shape.
	file, err := os.Open(filepath.Join(stateDirectory(), name))
	if err != nil {
		return state
	}
	defer file.Close()
	_ = json.NewDecoder(io.LimitReader(file, 8192)).Decode(&state)
	return state
}

func collectOperationAlerts(ctx context.Context, pool *pgxpool.Pool, rdb *redis.Client, now time.Time) []operationAlert {
	alerts := []operationAlert{}
	update := readOperationState("update.json")
	if update.Status == "failed" {
		alerts = append(alerts, operationAlert{"deployment_failed", "The latest deployment failed. The previous app may be stopped. Inspect the update service log."})
	}
	if !update.CheckedAt.IsZero() && now.Sub(update.CheckedAt) > 20*time.Minute {
		alerts = append(alerts, operationAlert{"updater_stale", "No deployment check was recorded in the last 20 minutes. Check the update timer."})
	}
	if update.Status == "waiting_for_images" && now.Sub(update.CheckedAt) < 20*time.Minute {
		alerts = append(alerts, operationAlert{"publication_pending", "The selected release is waiting for both image artifacts. The previous application has been kept running."})
	}
	backup := readOperationState("backup.json")
	if backup.CompletedAt.IsZero() || now.Sub(backup.CompletedAt) > 36*time.Hour {
		alerts = append(alerts, operationAlert{"backup_stale", "No successful server backup has been recorded within 36 hours. Check the backup timer and restore procedure."})
	}
	if rdb.Ping(ctx).Err() != nil {
		alerts = append(alerts, operationAlert{"redis_unavailable", "Redis is unavailable. Sessions and background delivery may be interrupted."})
	}
	if pool.Ping(ctx) != nil {
		return append(alerts, operationAlert{"database_unavailable", "The database is unavailable. Collection, matching and digest delivery cannot proceed."})
	}
	q := database.New(pool)
	sources, err := q.ListServiceHealth(ctx)
	if err != nil {
		alerts = append(alerts, operationAlert{"sources_unknown", "Could not read source collection status."})
	} else {
		for _, source := range sources {
			if source.ServiceStatus == "unhealthy" || source.ServiceStatus == "failed" || source.ServiceStatus == "degraded" || source.ServiceStatus == "error" {
				alerts = append(alerts, operationAlert{"source_failed_" + source.ServiceName, source.ServiceName + ": the latest collection failed. Inspect collection logs; provider errors are not exported."})
			} else if source.LastSuccessAt == nil || now.Sub(*source.LastSuccessAt) > 48*time.Hour {
				alerts = append(alerts, operationAlert{"source_stale_" + source.ServiceName, source.ServiceName + ": no successful collection in 48 hours."})
			}
		}
	}
	if os.Getenv("DIGEST_ENABLED") != "true" {
		return alerts
	}
	userID, err := q.GetDigestUser(ctx, os.Getenv("DIGEST_RECIPIENT"))
	if err != nil {
		return alerts
	}
	settings, err := q.GetMatchSettings(ctx, userID)
	if err != nil || !settings.EmailEnabled {
		return alerts
	}
	location, err := time.LoadLocation(settings.DigestTimezone)
	if err != nil {
		return alerts
	}
	local := now.In(location)
	if start := os.Getenv("DIGEST_START_DATE"); start != "" && local.Format("2006-01-02") < start {
		return alerts
	}
	slot := int32(-1)
	for _, hour := range settings.DigestHours {
		if int(hour) <= local.Hour() && hour > slot {
			slot = hour
		}
	}
	if slot < 0 {
		return alerts
	}
	due := time.Date(local.Year(), local.Month(), local.Day(), int(slot), 0, 0, 0, location)
	if now.Sub(due) < 2*time.Hour {
		return alerts
	}
	var status string
	err = pool.QueryRow(ctx, "SELECT status FROM email_digests WHERE user_id=$1 AND digest_date=$2::date AND digest_slot=$3", userID, local.Format("2006-01-02"), slot).Scan(&status)
	if err == nil && status == "sent" {
		return alerts
	}
	if err == nil {
		return append(alerts, operationAlert{"digest_overdue", "The latest scheduled digest has not been accepted by the email provider after two hours. Its recorded state is " + status + "."})
	}
	// No matches means no email by design. Count without the pending-work exclusion
	// so an indefinitely stuck match backlog is not mistaken for a healthy empty day.
	var candidates int
	err = pool.QueryRow(ctx, `SELECT count(*) FROM jobs j JOIN user_job_matches m ON m.job_id=j.id AND m.user_id=$1
 WHERE m.match_score >= $2 AND COALESCE(j.posted_at,j.created_at)>=now()-interval '30 days'
 AND (j.expires_at IS NULL OR j.expires_at>now())
 AND NOT EXISTS(SELECT 1 FROM email_digest_jobs d WHERE d.user_id=$1 AND d.job_id=j.id)
 AND NOT EXISTS(SELECT 1 FROM job_feedback f WHERE f.user_id=$1 AND f.job_id=j.id AND f.sentiment=-1)`, userID, settings.MinScore).Scan(&candidates)
	if err == nil && candidates > 0 {
		alerts = append(alerts, operationAlert{"digest_missing", "The latest scheduled digest is over two hours late despite undelivered qualifying jobs. Check the matching backlog and scheduler."})
	}
	return alerts
}

// Runs in a separate timer process so an app crash does not silence the monitor.
// A complete host outage still needs monitoring from another machine.
func runOperationsCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, os.Getenv("DB_URL"))
	if err != nil {
		return fmt.Errorf("invalid operations database configuration")
	}
	defer pool.Close()
	rdb := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR")})
	defer rdb.Close()
	now := time.Now().UTC()
	alerts := collectOperationAlerts(ctx, pool, rdb, now)
	healthURL := os.Getenv("OPS_HEALTH_URL")
	if healthURL == "" {
		healthURL = "http://app:8080/api/health"
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if req == nil {
		return fmt.Errorf("invalid operations health URL")
	}
	response, healthErr := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if healthErr != nil {
		alerts = append(alerts, operationAlert{"application_unavailable", "The application did not answer its internal health check. Inspect the app container and deployment log."})
	} else {
		response.Body.Close()
		if response.StatusCode != 200 {
			alerts = append(alerts, operationAlert{"application_unavailable", "The application health check returned a failure status."})
		}
	}

	data, _ := json.Marshal(map[string]any{"checked_at": now, "alerts": alerts})
	dir := stateDirectory()
	if err = os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(dir, "monitor.json.tmp"), data, 0600); err != nil {
		return err
	}
	if err = os.Rename(filepath.Join(dir, "monitor.json.tmp"), filepath.Join(dir, "monitor.json")); err != nil {
		return err
	}
	if os.Getenv("OPS_EMAIL_ENABLED") != "true" || len(alerts) == 0 {
		return nil
	}
	recipient := os.Getenv("DIGEST_RECIPIENT")
	from := os.Getenv("EMAIL_FROM")
	if _, err = mail.ParseAddress(recipient); err != nil {
		return fmt.Errorf("operations recipient is invalid")
	}
	if _, err = mail.ParseAddress(from); err != nil {
		return fmt.Errorf("operations sender is invalid")
	}
	if owner := os.Getenv("PERSONAL_EMAIL"); owner != "" && !strings.EqualFold(owner, recipient) {
		return fmt.Errorf("operations recipient must match the owner")
	}
	var sender mailer.Sender
	switch os.Getenv("EMAIL_PROVIDER") {
	case "resend":
		if os.Getenv("RESEND_API_KEY") == "" {
			return fmt.Errorf("operations email key missing")
		}
		sender = &mailer.Resend{APIKey: os.Getenv("RESEND_API_KEY")}
	case "gohighlevel":
		if os.Getenv("GHL_API_KEY") == "" || os.Getenv("GHL_CONTACT_ID") == "" {
			return fmt.Errorf("operations email configuration missing")
		}
		sender = &mailer.GoHighLevel{APIKey: os.Getenv("GHL_API_KEY"), ContactID: os.Getenv("GHL_CONTACT_ID")}
	default:
		return fmt.Errorf("operations email provider is not configured")
	}
	for _, alert := range alerts {
		// Publication waiting is expected during CI, not an incident email.
		if alert.Code == "publication_pending" {
			continue
		}
		key := fmt.Sprintf("%x", sha256.Sum256([]byte(alert.Code+":"+now.Format("2006-01-02"))))
		marker := filepath.Join(dir, "alert-"+key)
		if _, err := os.Stat(marker); err == nil {
			continue
		}
		// Claim before sending. A lost reply must not cause repeated non-idempotent
		// sends; once/day is deliberately conservative for an operations notification.
		claim, err := os.OpenFile(marker, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		claim.Close()
		_, err = sender.Send(ctx, "jobradar-ops-"+key, mailer.Message{From: from, To: []string{recipient}, Subject: "Jobradar needs attention: " + alert.Code, HTML: "<h1>Jobradar operational alert</h1><p>" + html.EscapeString(alert.Detail) + "</p><p>Check System health and the server service logs. This reports provider acceptance, not inbox delivery.</p>"})
		if err != nil {
			return fmt.Errorf("operations email could not be confirmed; inspect provider delivery history")
		}
	}
	// Remove deduplication markers older than a week; never touch database backups.
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "alert-") {
			if info, e := entry.Info(); e == nil && now.Sub(info.ModTime()) > 7*24*time.Hour {
				_ = os.Remove(filepath.Join(dir, entry.Name()))
			}
		}
	}
	return nil
}
