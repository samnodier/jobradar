package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
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
	if os.Getenv("APP_ENV") == "production" {
		monitor := readOperationsSnapshot()
		if monitor.CheckedAt.IsZero() {
			alerts = append(alerts, operationAlert{"monitor_unknown", "No operational monitor check is recorded. Install the monitor timer."})
		} else if now.Sub(monitor.CheckedAt) > 45*time.Minute {
			alerts = append(alerts, operationAlert{"monitor_stale", "The operational monitor has not recorded a check in 45 minutes. Inspect its timer and service logs."})
		}
	}

	update := readOperationState("update.json")
	if update.Status == "failed" {
		alerts = append(alerts, operationAlert{"deployment_failed", "The latest deployment failed. The previous app may be stopped. Inspect the update service log."})
	}
	if update.CheckedAt.IsZero() && os.Getenv("APP_ENV") == "production" {
		alerts = append(alerts, operationAlert{"updater_unknown", "No deployment check has been recorded. Install and inspect the update timer."})
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
				alerts = append(alerts, operationAlert{"source_failed_" + source.ServiceName, source.ServiceName + ": " + sourceFailureSummary(source.ServiceStatus, source.LastError)})
			} else if source.LastSuccessAt == nil || now.Sub(*source.LastSuccessAt) > 48*time.Hour {
				alerts = append(alerts, operationAlert{"source_stale_" + source.ServiceName, source.ServiceName + ": no successful collection in 48 hours."})
			}
		}
	}
	if os.Getenv("DIGEST_ENABLED") != "true" {
		return alerts
	}
	userID, err := q.GetDigestUser(ctx, os.Getenv("DIGEST_RECIPIENT"))
	if errors.Is(err, pgx.ErrNoRows) {
		var enabled bool
		lookupErr := pool.QueryRow(ctx, "SELECT COALESCE(notify_jobs,false) FROM users WHERE email=$1", os.Getenv("DIGEST_RECIPIENT")).Scan(&enabled)
		if lookupErr == nil && !enabled {
			return alerts
		} // intentional account opt-out
		return append(alerts, operationAlert{"digest_account_unknown", "The configured digest recipient could not be loaded. Check account and delivery settings."})
	}
	if err != nil {
		return append(alerts, operationAlert{"digest_account_unknown", "Could not read the configured digest account."})
	}
	settings, err := q.GetMatchSettings(ctx, userID)
	if err != nil {
		return append(alerts, operationAlert{"digest_settings_unknown", "Could not load digest delivery settings."})
	}
	if !settings.EmailEnabled {
		return alerts
	}
	location, err := time.LoadLocation(settings.DigestTimezone)
	if err != nil {
		return append(alerts, operationAlert{"digest_timezone_invalid", "The configured digest timezone is invalid."})
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
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return append(alerts, operationAlert{"digest_status_unknown", "Could not read the latest scheduled digest state."})
	}
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
 WHERE (j.created_by_user_id IS NULL OR j.created_by_user_id=$1) AND m.match_score >= $2 AND COALESCE(j.posted_at,j.created_at)>=now()-interval '30 days'
 AND (j.expires_at IS NULL OR j.expires_at>now())
 AND NOT EXISTS(SELECT 1 FROM job_availability a WHERE a.job_id=j.id AND a.status='unavailable' AND a.checked_at>now()-interval '48 hours')
 AND NOT EXISTS(SELECT 1 FROM email_digest_jobs d WHERE d.user_id=$1 AND d.job_id=j.id)
 AND NOT EXISTS(SELECT 1 FROM job_feedback f WHERE f.user_id=$1 AND f.job_id=j.id AND f.sentiment=-1)`, userID, settings.MinScore).Scan(&candidates)
	if err != nil {
		return append(alerts, operationAlert{"digest_candidates_unknown", "Could not determine whether undelivered digest jobs are waiting."})
	}
	if candidates > 0 {
		alerts = append(alerts, operationAlert{"digest_missing", "The latest scheduled digest is over two hours late despite undelivered qualifying jobs. Check the matching backlog and scheduler."})
	}
	return alerts
}

type operationsSnapshot struct {
	CheckedAt   time.Time        `json:"checked_at"`
	Alerts      []operationAlert `json:"alerts"`
	EmailStatus string           `json:"email_status"`
	EmailDetail string           `json:"email_detail"`
}

func readOperationsSnapshot() operationsSnapshot {
	var snapshot operationsSnapshot
	file, err := os.Open(filepath.Join(stateDirectory(), "monitor.json"))
	if err != nil {
		return snapshot
	}
	defer file.Close()
	_ = json.NewDecoder(io.LimitReader(file, 65536)).Decode(&snapshot)
	return snapshot
}
func writeOperationJSON(file string, value any, mode os.FileMode) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if err = os.WriteFile(file+".tmp", data, mode); err != nil {
		return err
	}
	if err = os.Chmod(file+".tmp", mode); err != nil {
		return err
	}
	return os.Rename(file+".tmp", file)
}

type alertDelivery struct {
	Status   string         `json:"status"`
	Attempts int            `json:"attempts"`
	Message  mailer.Message `json:"message"`
	Provider string         `json:"provider"`
}

// Persist the immutable payload before a send. Retry-safe providers reuse both
// the exact body and key; ambiguous non-idempotent sends stop for owner review.
func deliverOperationAlert(ctx context.Context, sender mailer.Sender, dir string, now time.Time, from, recipient string, alert operationAlert) (string, error) {
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(alert.Code+":"+now.UTC().Format("2006-01-02"))))
	file := filepath.Join(dir, "alert-"+key)
	record := alertDelivery{Status: "pending", Provider: fmt.Sprintf("%T", sender), Message: mailer.Message{From: from, To: []string{recipient}, Subject: "Jobradar needs attention: " + alert.Code, HTML: "<h1>Jobradar operational alert</h1><p>" + html.EscapeString(alert.Detail) + "</p><p>Check System health and the server service logs. Provider acceptance does not confirm inbox delivery.</p>"}}
	if data, err := os.ReadFile(file); err == nil {
		if json.Unmarshal(data, &record) != nil {
			return "uncertain", nil
		} // legacy/crashed claim: do not guess
	} else if !os.IsNotExist(err) {
		return "unconfirmed", err
	}
	if record.Provider != fmt.Sprintf("%T", sender) || len(record.Message.To) != 1 || !strings.EqualFold(record.Message.To[0], recipient) || record.Message.From != from {
		return "uncertain", nil
	}
	if record.Status == "sent" {
		return "accepted", nil
	}
	if record.Status == "uncertain" || (record.Status == "sending" && !sender.RetrySafe()) {
		return "uncertain", nil
	}
	if record.Attempts >= 3 {
		return "retry_limit", nil
	}
	record.Status = "sending"
	record.Attempts++
	if err := writeOperationJSON(file, record, 0600); err != nil {
		return "unconfirmed", err
	}
	callCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	_, err := sender.Send(callCtx, "jobradar-ops-"+key, record.Message)
	if err == nil {
		record.Status = "sent"
	} else {
		var sendErr *mailer.SendError
		definitive := errors.As(err, &sendErr) && !sendErr.Ambiguous
		if sender.RetrySafe() || definitive {
			record.Status = "failed"
		} else {
			record.Status = "uncertain"
		}
	}
	if saveErr := writeOperationJSON(file, record, 0600); saveErr != nil {
		return "unconfirmed", saveErr
	}
	if record.Status == "sent" {
		return "accepted", nil
	}
	return record.Status, nil
}

// The host timer serializes this command. The monitor has a separate lifetime
// from the app, but a whole-host outage still requires an external monitor.
func runOperationsCheck(ctx context.Context) (resultErr error) {
	pool, err := pgxpool.New(ctx, os.Getenv("DB_URL"))
	if err != nil {
		return fmt.Errorf("invalid operations database configuration")
	}
	defer pool.Close()
	rdb := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR"), ContextTimeoutEnabled: true})
	defer rdb.Close()
	now := time.Now().UTC()
	// Dependency failure cannot consume the deadline reserved for incident email.
	checkCtx, cancelCheck := context.WithTimeout(ctx, 15*time.Second)
	alerts := collectOperationAlerts(checkCtx, pool, rdb, now)
	healthURL := os.Getenv("OPS_HEALTH_URL")
	if healthURL == "" {
		healthURL = "http://app:8080/api/health"
	}
	cancelCheck()
	healthCtx, cancelHealth := context.WithTimeout(ctx, 5*time.Second)
	defer cancelHealth()
	req, _ := http.NewRequestWithContext(healthCtx, http.MethodGet, healthURL, nil)
	if req == nil {
		cancelCheck()
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
	cancelCheck()
	dir := stateDirectory()
	if err = os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	snapshot := operationsSnapshot{CheckedAt: now, Alerts: alerts, EmailStatus: "disabled"}
	defer func() {
		if resultErr != nil {
			snapshot.EmailStatus = "unconfirmed"
			snapshot.EmailDetail = "Operational email could not be confirmed. Check the monitor service log and provider history."
		}
		if err := writeOperationJSON(filepath.Join(dir, "monitor.json"), snapshot, 0644); err != nil && resultErr == nil {
			resultErr = err
		}
	}()
	if os.Getenv("OPS_EMAIL_ENABLED") != "true" {
		return nil
	}
	snapshot.EmailStatus = "no_incidents"
	if len(alerts) == 0 {
		return nil
	}
	recipient, from := os.Getenv("DIGEST_RECIPIENT"), os.Getenv("EMAIL_FROM")
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
	emailCtx, cancelEmail := context.WithTimeout(ctx, 2*time.Minute)
	defer cancelEmail()
	snapshot.EmailStatus = "no_incidents"
	for _, alert := range alerts {
		if alert.Code == "publication_pending" || alert.Code == "monitor_unknown" || alert.Code == "monitor_stale" {
			continue
		}
		if snapshot.EmailStatus == "no_incidents" {
			snapshot.EmailStatus = "accepted"
		}
		state, err := deliverOperationAlert(emailCtx, sender, dir, now, from, recipient, alert)
		if err != nil {
			return fmt.Errorf("could not persist operational email delivery state")
		}
		if state != "accepted" {
			snapshot.EmailStatus = "unconfirmed"
			snapshot.EmailDetail = "Some operational alerts are unconfirmed. Retry-safe failures get up to three attempts per UTC day; ambiguous non-idempotent delivery requires provider review."
		}
	}
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
