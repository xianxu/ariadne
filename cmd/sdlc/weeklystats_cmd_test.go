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
		read: func(context.Context, string, string, WeeklyWindow) weeklyRepoEvidence {
			return weeklyRepoEvidence{Repo: "ariadne", Root: "/fake/ariadne", CoverageUnknown: []string{"tracker unavailable"}}
		},
		previous: func(string, WeeklyWindow) previousResourceResult {
			return previousResourceResult{State: "baseline", Reason: "previous week missing"}
		},
		resources: func(context.Context, workspace.Identity, []string, weeklySnapshotFlags) (ResourceCollection, ResourceMachine) {
			return ResourceCollection{}, ResourceMachine{}
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
	if h := got.Repos[0].Headline; h.AttemptsPerLandedIssue != nil || h.ObservedLandedIssues != 0 || !strings.Contains(h.Text, "0 of 0 observed landed issues") || len(h.TopUnknownReasons) != 1 || h.TopUnknownReasons[0].Reason != "tracker unavailable" {
		t.Fatalf("headline must expose source coverage: %+v", h)
	}
	out.Reset()
	if err := runWeeklySnapshot(context.Background(), &out, weeklySnapshotFlags{End: end.Format(time.RFC3339), BrainRoot: brain}, deps); err != nil || strings.TrimSpace(out.String()) != path {
		t.Fatalf("stable rerun: %v, %s", err, out.String())
	}
}

func TestRunWeeklySnapshotAttachesResourceSectionOnce(t *testing.T) {
	end := time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
	brain := t.TempDir()
	if err := os.Mkdir(filepath.Join(brain, ".brain"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(brain, ".brain", "config.md"), []byte("local-only"), 0o644); err != nil {
		t.Fatal(err)
	}
	collects := 0
	deps := weeklySnapshotDeps{
		now:      func() time.Time { return end.Add(time.Hour) },
		identity: func() (workspace.Identity, error) { return workspace.Identity{FleetRoot: "/fleet"}, nil },
		discover: func(workspace.Identity) ([]string, error) { return []string{"/fleet/ariadne"}, nil },
		read: func(context.Context, string, string, WeeklyWindow) weeklyRepoEvidence {
			return weeklyRepoEvidence{Repo: "ariadne", Root: "/fleet/ariadne", CoverageUnknown: []string{"tracker unavailable"}}
		},
		previous: func(string, WeeklyWindow) previousResourceResult {
			return previousResourceResult{State: "baseline", Reason: "previous week missing"}
		},
		resources: func(_ context.Context, id workspace.Identity, paths []string, _ weeklySnapshotFlags) (ResourceCollection, ResourceMachine) {
			collects++
			if id.FleetRoot != "/fleet" || len(paths) != 1 || paths[0] != "/fleet/ariadne" {
				t.Fatalf("collector inputs id=%+v paths=%v", id, paths)
			}
			return ResourceCollection{Families: []ResourceObservation{{Kind: "couch-ledgers", Definition: "v1", Bytes: 130, Complete: true}}}, ResourceMachine{}
		},
		publish: publishBrainSnapshot,
	}
	var out bytes.Buffer
	if err := runWeeklySnapshot(context.Background(), &out, weeklySnapshotFlags{End: end.Format(time.RFC3339), BrainRoot: brain}, deps); err != nil {
		t.Fatal(err)
	}
	var got WeeklySnapshot
	data, err := os.ReadFile(strings.TrimSpace(out.String()))
	if err != nil || json.Unmarshal(data, &got) != nil {
		t.Fatalf("snapshot read/decode: %v", err)
	}
	if collects != 1 || got.SchemaVersion != 2 || got.Resources == nil || !got.Resources.ObservedAt.Equal(end.Add(time.Hour)) || len(got.Resources.Families) != 1 || got.Resources.Families[0].State != "baseline" || got.Resources.Families[0].Bytes != 130 {
		t.Fatalf("attached resources = %+v, collects=%d", got, collects)
	}
}
