package main

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
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
