package main

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/mailer"
	"github.com/samnodier/jobradar/internal/testutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOperationsExposeSanitizedFailureAndBackupState(t *testing.T) {
	pool := testutil.Pool(t)
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR required")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	dir := t.TempDir()
	t.Setenv("OPERATIONS_STATE_DIR", dir)
	t.Setenv("DIGEST_ENABLED", "false")
	now := time.Now().UTC()
	write := func(name string, state operationState) {
		data, _ := json.Marshal(state)
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("update.json", operationState{Status: "failed", CheckedAt: now, TargetRevision: "release"})
	write("backup.json", operationState{CompletedAt: now})
	alerts := collectOperationAlerts(context.Background(), pool, rdb, now)
	found := false
	for _, alert := range alerts {
		if alert.Code == "deployment_failed" {
			found = true
		}
		if alert.Code == "backup_stale" {
			t.Fatal("recent backup flagged")
		}
	}
	if !found {
		t.Fatal("failed deployment was hidden")
	}
	write("backup.json", operationState{CompletedAt: now.Add(-48 * time.Hour)})
	found = false
	for _, alert := range collectOperationAlerts(context.Background(), pool, rdb, now) {
		if alert.Code == "backup_stale" {
			found = true
		}
	}
	if !found {
		t.Fatal("stale backup was hidden")
	}
}

type operationsTestSender struct {
	safe     bool
	calls    int
	fail     error
	messages []mailer.Message
	keys     []string
}

func (s *operationsTestSender) RetrySafe() bool { return s.safe }
func (s *operationsTestSender) Send(_ context.Context, key string, message mailer.Message) (string, error) {
	s.calls++
	s.messages = append(s.messages, message)
	s.keys = append(s.keys, key)
	if s.calls == 1 {
		return "", s.fail
	}
	return "accepted", nil
}
func TestOperationEmailRetriesImmutablePayloadAndStopsAfterAcceptance(t *testing.T) {
	sender := &operationsTestSender{safe: true, fail: &mailer.SendError{Ambiguous: true}}
	dir := t.TempDir()
	now := time.Now().UTC()
	alert := operationAlert{Code: "example", Detail: "First incident detail"}
	state, err := deliverOperationAlert(context.Background(), sender, dir, now, "jobs@example.test", "owner@example.test", alert)
	if err != nil || state != "failed" {
		t.Fatal(state, err)
	}
	alert.Detail = "Changed incident detail"
	state, err = deliverOperationAlert(context.Background(), sender, dir, now, "jobs@example.test", "owner@example.test", alert)
	if err != nil || state != "accepted" || sender.calls != 2 {
		t.Fatal(state, err, sender.calls)
	}
	if sender.keys[0] != sender.keys[1] || sender.messages[0].HTML != sender.messages[1].HTML {
		t.Fatal("retry changed key or payload")
	}
	state, err = deliverOperationAlert(context.Background(), sender, dir, now, "jobs@example.test", "owner@example.test", alert)
	if err != nil || state != "accepted" || sender.calls != 2 {
		t.Fatal("sent again", state, err)
	}
}
func TestOperationEmailDoesNotRepeatAmbiguousNonIdempotentSend(t *testing.T) {
	sender := &operationsTestSender{fail: &mailer.SendError{Ambiguous: true}}
	dir := t.TempDir()
	now := time.Now().UTC()
	alert := operationAlert{Code: "example", Detail: "incident"}
	for i := 0; i < 2; i++ {
		state, err := deliverOperationAlert(context.Background(), sender, dir, now, "jobs@example.test", "owner@example.test", alert)
		if err != nil || state != "uncertain" {
			t.Fatal(state, err)
		}
	}
	if sender.calls != 1 {
		t.Fatal("ambiguous delivery repeated")
	}
}
func TestOperationEmailRetriesDefinitiveRejection(t *testing.T) {
	sender := &operationsTestSender{fail: &mailer.SendError{Code: 429}}
	dir := t.TempDir()
	now := time.Now().UTC()
	alert := operationAlert{Code: "example", Detail: "incident"}
	if state, err := deliverOperationAlert(context.Background(), sender, dir, now, "jobs@example.test", "owner@example.test", alert); err != nil || state != "failed" {
		t.Fatal(state, err)
	}
	if state, err := deliverOperationAlert(context.Background(), sender, dir, now, "jobs@example.test", "owner@example.test", alert); err != nil || state != "accepted" {
		t.Fatal(state, err)
	}
	if sender.calls != 2 {
		t.Fatal("definitive refusal not retried")
	}
}
