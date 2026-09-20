package digest

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/mailer"
)

type Service struct {
	Pool   *pgxpool.Pool
	Config Config
	Sender mailer.Sender
}

func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		callCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
		if err := s.Tick(callCtx, time.Now(), false); err != nil && ctx.Err() == nil {
			log.Printf("digest: %v", err)
		}
		cancel()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (s *Service) Preview(ctx context.Context) (string, error) {
	q := database.New(s.Pool)
	id, err := q.GetDigestUser(ctx, s.Config.Recipient)
	if err != nil {
		return "", fmt.Errorf("digest recipient must be a registered user with notifications enabled")
	}
	rows, err := q.GetDigestCandidates(ctx, database.GetDigestCandidatesParams{UserID: id, MatchScore: &[]int32{int32(s.Config.MinScore)}[0], Limit: int32(s.Config.MaxJobs)})
	if err != nil {
		return "", err
	}
	_, html, err := Render(rows)
	return html, err
}
func (s *Service) Tick(ctx context.Context, now time.Time, force bool) error {
	if !s.Config.Enabled {
		return nil
	}
	// A dedicated session advisory lock prevents two schedulers from sending together.
	conn, err := s.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	var locked bool
	if err = conn.QueryRow(ctx, "SELECT pg_try_advisory_lock(7261646172)").Scan(&locked); err != nil {
		return err
	}
	if !locked {
		return nil
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if _, e := conn.Exec(unlockCtx, "SELECT pg_advisory_unlock(7261646172)"); e != nil {
			_ = conn.Conn().Close(unlockCtx)
		}
	}()
	q := database.New(s.Pool)
	userID, err := q.GetDigestUser(ctx, s.Config.Recipient)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	local := now.In(s.Config.Location)
	var pending int
	if err := s.Pool.QueryRow(ctx, "SELECT count(*) FROM pending_matches WHERE user_id=$1", userID).Scan(&pending); err != nil {
		return err
	}
	if pending == 0 && (force || local.Hour() >= s.Config.Hour) {
		date := pgtype.Date{Time: time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC), Valid: true}
		_, err = q.GetDailyDigest(ctx, database.GetDailyDigestParams{UserID: userID, DigestDate: date})
		if errors.Is(err, pgx.ErrNoRows) {
			tx, e := s.Pool.Begin(ctx)
			if e != nil {
				return e
			}
			e = func() error {
				defer tx.Rollback(context.Background())
				tq := q.WithTx(tx)
				score := int32(s.Config.MinScore)
				jobs, e := tq.GetDigestCandidates(ctx, database.GetDigestCandidatesParams{UserID: userID, MatchScore: &score, Limit: int32(s.Config.MaxJobs)})
				if e != nil {
					return e
				}
				if len(jobs) == 0 {
					return nil
				}
				subject, html, e := Render(jobs)
				if e != nil {
					return e
				}
				d, e := tq.CreateDigest(ctx, database.CreateDigestParams{UserID: userID, DigestDate: date, Recipient: s.Config.Recipient, Sender: s.Config.From, Provider: s.Config.Provider, Subject: subject, Html: html})
				if e != nil {
					return e
				}
				for _, j := range jobs {
					if e = tq.ReserveDigestJob(ctx, database.ReserveDigestJobParams{DigestID: d.ID, UserID: userID, JobID: j.ID}); e != nil {
						return e
					}
				}
				return tx.Commit(ctx)
			}()
			if e != nil {
				return e
			}
		} else if err != nil {
			return err
		}
	}
	d, err := q.GetPendingDigest(ctx, s.Config.Recipient)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	fail := func(status, message string) error {
		return q.FailDigest(ctx, database.FailDigestParams{ID: d.ID, Status: status, LastError: &message})
	}
	if d.Provider != s.Config.Provider || d.Sender != s.Config.From {
		return fail("failed", "Email configuration changed; review the pending digest before resending")
	}
	if d.Status == "sending" && !s.Sender.RetrySafe() {
		return fail("uncertain", "Previous send was interrupted; check the provider before retrying")
	}
	if d.FirstAttemptAt != nil && now.Sub(*d.FirstAttemptAt) >= 23*time.Hour {
		return fail("uncertain", "Idempotency retry window expired; check the provider before retrying")
	}
	if err = q.StartDigestAttempt(ctx, d.ID); err != nil {
		return err
	}
	id, sendErr := s.Sender.Send(ctx, "jobradar/digest/"+d.ID.String(), mailer.Message{From: d.Sender, To: []string{d.Recipient}, Subject: d.Subject, HTML: d.Html})
	if sendErr != nil {
		status := "pending"
		var apiErr *mailer.SendError
		if errors.As(sendErr, &apiErr) {
			if apiErr.Ambiguous && !s.Sender.RetrySafe() {
				status = "uncertain"
			} else if apiErr.Code >= 400 && apiErr.Code < 500 && apiErr.Code != 429 {
				status = "failed"
			}
		}
		if d.Attempts >= 7 && status == "pending" {
			status = "failed"
		}
		if err := fail(status, sendErr.Error()); err != nil {
			return err
		}
		return sendErr
	}
	// If this write fails, the durable 'sending' state causes a safe replay (Resend)
	// or manual reconciliation (GHL). Never pretend a provider success was recorded.
	return q.FinishDigest(ctx, database.FinishDigestParams{ID: d.ID, ProviderMessageID: &id})
}
