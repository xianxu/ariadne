package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/pkg/workspace"
)

func TestRunWeeklySnapshotPublishesVersionedPartialCoverage(t *testing.T) {
	if newWeeklySnapshotCmd().Flags().Lookup("repo") != nil {
		t.Fatal("a repo-filtered run could claim the immutable fleet-wide date path")
	}
	end := time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
	brain := t.TempDir()
	if err := os.Mkdir(filepath.Join(brain, ".brain"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(brain, ".brain", "config.md"), []byte("local-only"), 0o644); err != nil {
		t.Fatal(err)
	}
	deps := weeklySnapshotDeps{
		now:      func() time.Time { return end.Add(time.Hour) },
		identity: func() (workspace.Identity, error) { return workspace.Identity{}, nil },
		discover: func(workspace.Identity) ([]string, error) { return []string{"/fake/ariadne"}, nil },
		read: func(context.Context, string, string) weeklyRepoEvidence {
			return weeklyRepoEvidence{Repo: "ariadne", Root: "/fake/ariadne", CoverageUnknown: []string{"tracker unavailable"}}
		},
		publish: publishBrainSnapshot,
	}
	var out bytes.Buffer
	if err := runWeeklySnapshot(context.Background(), &out, weeklySnapshotFlags{End: end.Format(time.RFC3339), BrainRoot: brain}, deps); err != nil {
		t.Fatal(err)
	}
	path := strings.TrimSpace(out.String())
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got WeeklySnapshot
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != weeklySnapshotSchemaVersion || len(got.Repos) != 1 || got.Repos[0].AttemptsPerLandedIssue != nil || len(got.Repos[0].CoverageUnknown) == 0 {
		t.Fatalf("partial snapshot = %+v", got)
	}
	out.Reset()
	if err := runWeeklySnapshot(context.Background(), &out, weeklySnapshotFlags{End: end.Format(time.RFC3339), BrainRoot: brain}, deps); err != nil || strings.TrimSpace(out.String()) != path {
		t.Fatalf("stable rerun: %v, %s", err, out.String())
	}
}
