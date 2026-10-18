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
			return ResourceCollection{CoverageUnknown: []string{"inventory unavailable"}}, fakeWeeklyResourceMachine()
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
	now := end.Add(time.Hour)
	deps := weeklySnapshotDeps{
		now:      func() time.Time { return now },
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
			return ResourceCollection{Families: []ResourceObservation{{Kind: "couch-ledgers", Definition: "v1", Bytes: 130, Files: 1, Complete: true}}}, fakeWeeklyResourceMachine()
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
	now = now.Add(time.Second)
	out.Reset()
	if err := runWeeklySnapshot(context.Background(), &out, weeklySnapshotFlags{End: end.Format(time.RFC3339), BrainRoot: brain}, deps); err != nil || strings.TrimSpace(out.String()) != filepath.Join(brain, "memory/work/devinfra/test-economy/weekly/2026-10-19.json") || collects != 1 {
		t.Fatalf("advancing clock must preserve existing snapshot without recollection: err=%v path=%q collects=%d", err, out.String(), collects)
	}
}

func TestRunWeeklySnapshotRecoversConcurrentPublication(t *testing.T) {
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
			return weeklyRepoEvidence{Repo: "ariadne"}
		},
		previous: func(string, WeeklyWindow) previousResourceResult { return previousResourceResult{State: "baseline"} },
		resources: func(context.Context, workspace.Identity, []string, weeklySnapshotFlags) (ResourceCollection, ResourceMachine) {
			return ResourceCollection{CoverageUnknown: []string{"inventory unavailable"}}, fakeWeeklyResourceMachine()
		},
		publish: func(root string, end time.Time, data []byte) (string, error) {
			var snap WeeklySnapshot
			if err := json.Unmarshal(data, &snap); err != nil {
				t.Fatal(err)
			}
			snap.Resources.ObservedAt = snap.Resources.ObservedAt.Add(-time.Second)
			winner, err := json.Marshal(snap)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := publishBrainSnapshot(root, end, winner); err != nil {
				t.Fatal(err)
			}
			return "", os.ErrExist
		},
	}
	var out bytes.Buffer
	if err := runWeeklySnapshot(context.Background(), &out, weeklySnapshotFlags{End: end.Format(time.RFC3339), BrainRoot: brain}, deps); err != nil || !strings.Contains(out.String(), "2026-10-19.json") {
		t.Fatalf("concurrent valid winner must be returned: %v %q", err, out.String())
	}
}

func fakeWeeklyResourceMachine() ResourceMachine {
	obs := func(kind string) ResourceObservation {
		return ResourceObservation{Kind: kind, Definition: "/source", Complete: true}
	}
	return ResourceMachine{
		GoCache: obs("go-cache"), AgentScratchFree: obs("agentscratch-free"), AgentScratchTotal: obs("agentscratch-total"),
		BytesWritten: DeviceGauge{Device: "disk0", Epoch: "123", Observed: true, State: "observed"},
		SSDWear:      DeviceGauge{Device: "disk0", Observed: true, State: "observed"},
	}
}

func TestReadExistingWeeklySnapshotRejectsIncompatibleOrCorrupt(t *testing.T) {
	root := t.TempDir()
	end := time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
	window, err := newWeeklyWindow(end, end)
	if err != nil {
		t.Fatal(err)
	}
	path := weeklySnapshotPath(root, end)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{
		`{"schema_version":1,"window":{"start":"2026-10-12T00:00:00Z","end":"2026-10-19T00:00:00Z"},"repos":[{"repo":"ariadne"}]}`,
		`{not-json}`,
		`{"schema_version":2,"window":{"start":"2026-10-12T00:00:00Z","end":"2026-10-19T00:00:00Z"},"repos":[{"repo":"ariadne"}],"resources":{"observed_at":"2026-10-19T01:00:00Z","families":[],"suites":[]}}`,
	} {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, exists, err := readExistingWeeklySnapshot(root, window); err == nil || exists {
			t.Fatalf("invalid existing artifact accepted: %s (%v)", body, err)
		}
	}
}
