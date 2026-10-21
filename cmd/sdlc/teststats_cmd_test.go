package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/pkg/workspace"
)

func TestSuiteStatsCommandDiscoversFleetFromNumberedSlot(t *testing.T) {
	fleet := t.TempDir()
	ariadne := filepath.Join(fleet, "ariadne")
	pair := filepath.Join(fleet, "pair")
	slot := filepath.Join(fleet, "worktree", "ariadne-slot1", "ariadne")
	for _, dir := range []string{ariadne, pair, slot} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	sha := strings.Repeat("a", 40)
	writeAttemptFixture := func(repoDir, repo string) {
		t.Helper()
		root := filepath.Join(repoDir, ".git", "sdlc", "test-stats")
		if err := os.MkdirAll(filepath.Join(root, "attempts"), 0o755); err != nil {
			t.Fatal(err)
		}
		q, w := 3.0, 90.0
		a := SuiteAttempt{Version: 1, ID: "aabbccdd", Repo: "xianxu/" + repo, SHA: sha, Check: "suite", StartedAt: now.Add(-time.Hour), FinishedAt: now.Add(-time.Hour + 90*time.Second), QueueSeconds: &q, WallSeconds: &w, LogSHA256: strings.Repeat("b", 64), LogPath: filepath.Join("local-checks", sha, "suite.log")}
		b, err := json.Marshal(a)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "attempts", a.ID+".json"), b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	writeAttemptFixture(ariadne, "ariadne")
	legacyDir := filepath.Join(pair, ".git", "sdlc", "local-checks", strings.Repeat("c", 40))
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, "suite.log"), []byte("ok  \tgithub.com/xianxu/pair/cmd/tool\t5.1s\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	badPath := filepath.Join(ariadne, ".git", "sdlc", "test-stats", "attempts", "bad.json")
	if err := os.WriteFile(badPath, []byte("not JSON"), 0o600); err != nil {
		t.Fatal(err)
	}
	deps := testStatsDeps{
		now: func() time.Time { return now },
		identity: func() (workspace.Identity, error) {
			return workspace.Identity{Repo: "ariadne", WorktreeRoot: slot, PrimaryRoot: ariadne, FleetRoot: fleet}, nil
		},
		fleetDirs: func(_ string) ([]string, error) { return []string{ariadne, pair}, nil },
		commonDir: func(root string) (string, error) { return filepath.Join(root, ".git"), nil },
	}
	var out bytes.Buffer
	if err := runTestStats(&out, testStatsFlags{Since: "7d", JSON: true}, deps); err != nil {
		t.Fatal(err)
	}
	var report testStatsReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("JSON: %v: %s", err, out.String())
	}
	if len(report.Repos) != 2 || report.Repos[0].Repo != "ariadne" || report.Repos[1].Repo != "pair" {
		t.Fatalf("repos = %#v", report.Repos)
	}
	if report.Repos[0].Runs != 1 || report.Repos[0].Wall == nil || report.Repos[0].Wall.P50 != 90 {
		t.Fatalf("ariadne = %#v", report.Repos[0])
	}
	if report.Repos[1].Runs != 1 || report.Repos[1].FullyRecorded != 0 || report.Repos[1].Wall != nil || len(report.Repos[1].Slow) != 1 {
		t.Fatalf("legacy pair = %#v", report.Repos[1])
	}
	if len(report.Warnings) == 0 || !strings.Contains(strings.Join(report.Warnings, " "), "invalid attempt") {
		t.Fatalf("warnings = %q", report.Warnings)
	}
	out.Reset()
	if err := runTestStats(&out, testStatsFlags{Since: "7d", Repo: "pair"}, deps); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "ariadne") || !strings.Contains(out.String(), "pair") || !strings.Contains(out.String(), "n/a") || !strings.Contains(out.String(), "partial") {
		t.Fatalf("text = %q", out.String())
	}
	out.Reset()
	if err := runTestStats(&out, testStatsFlags{Since: "30m", JSON: true}, deps); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Repos) != 1 || report.Repos[0].Repo != "pair" {
		t.Fatalf("since cutoff retained old ariadne attempt: %#v", report.Repos)
	}
	if err := runTestStats(&out, testStatsFlags{Since: "nonsense"}, deps); err == nil {
		t.Fatal("invalid --since accepted")
	}
}

func TestSuiteStatsCommandSinceFilter(t *testing.T) {
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	// Parsing durations is a CLI contract; a cutoff outside retention is still
	// accepted, but it cannot invent missing history.
	if got, err := parseTestStatsSince("7d", now); err != nil || !got.Equal(now.Add(-7*24*time.Hour)) {
		t.Fatalf("7d = %v, %v", got, err)
	}
	if got, err := parseTestStatsSince("12h", now); err != nil || !got.Equal(now.Add(-12*time.Hour)) {
		t.Fatalf("12h = %v, %v", got, err)
	}
}

func TestSuiteStatsCommandSmokeRecordDoesNotHideLegacySuite(t *testing.T) {
	fleet := t.TempDir()
	repo := filepath.Join(fleet, "ariadne")
	common := filepath.Join(repo, ".git")
	sha := strings.Repeat("a", 40)
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	root := filepath.Join(common, "sdlc", "test-stats", "attempts")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	a := SuiteAttempt{Version: 1, ID: "aabbccdd", Repo: "xianxu/ariadne", SHA: sha, Check: "smoke", StartedAt: now.Add(-time.Minute), FinishedAt: now, LogSHA256: strings.Repeat("b", 64), LogPath: filepath.Join("local-checks", sha, "smoke.log")}
	b, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, a.ID+".json"), b, 0o600); err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(common, "sdlc", "local-checks", sha)
	if err := os.MkdirAll(legacy, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "suite.log"), []byte("ok github.com/xianxu/ariadne/pkg/stats 2.5s\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	deps := testStatsDeps{now: func() time.Time { return now }, identity: func() (workspace.Identity, error) {
		return workspace.Identity{PrimaryRoot: repo, FleetRoot: fleet}, nil
	}, fleetDirs: func(string) ([]string, error) { return []string{repo}, nil }, commonDir: func(string) (string, error) { return common, nil }}
	var out bytes.Buffer
	if err := runTestStats(&out, testStatsFlags{Since: "7d", JSON: true}, deps); err != nil {
		t.Fatal(err)
	}
	var report testStatsReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Repos) != 1 || report.Repos[0].Runs != 1 || report.Repos[0].FullyRecorded != 0 || len(report.Repos[0].Slow) != 1 {
		t.Fatalf("legacy suite hidden by smoke record: %+v", report)
	}
}

func TestSuiteStatsUnsupportedCoverageStaysScoped(t *testing.T) {
	root := t.TempDir()
	common := filepath.Join(root, ".git")
	attempts := filepath.Join(common, "sdlc", "test-stats", "attempts")
	if err := os.MkdirAll(attempts, 0o700); err != nil {
		t.Fatal(err)
	}
	a := storedAttemptFixture("a1")
	a.Exit = 0
	a.Failures = nil
	wall := 1.0
	a.WallSeconds = &wall
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(attempts, "a1.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	future := strings.Replace(string(futureProofFixture("d3")), strings.Repeat("a", 40), strings.Repeat("b", 40), 1)
	if err := os.WriteFile(filepath.Join(attempts, "d3.json"), []byte(future), 0o600); err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(common, "sdlc", "local-checks", strings.Repeat("b", 40))
	if err := os.MkdirAll(legacy, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "suite.log"), []byte("ok x/pkg 9s\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	deps := testStatsDeps{now: func() time.Time { return a.FinishedAt }, identity: func() (workspace.Identity, error) {
		return workspace.Identity{PrimaryRoot: root, FleetRoot: filepath.Dir(root)}, nil
	}, fleetDirs: func(string) ([]string, error) { return []string{root}, nil }, commonDir: func(string) (string, error) { return common, nil }}
	var out bytes.Buffer
	if err := runTestStats(&out, testStatsFlags{Since: "7d", JSON: true}, deps); err != nil {
		t.Fatal(err)
	}
	var report testStatsReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Repos) != 1 || report.Repos[0].Runs != 1 || report.Repos[0].FullyRecorded != 1 || len(report.Warnings) != 1 || !strings.Contains(report.Warnings[0], "needs a newer sdlc") {
		t.Fatalf("scoped statistics: %+v", report)
	}
}
