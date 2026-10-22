package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// Exercise both production IO boundaries using the same persisted v4 records
// and retained raw logs. Reader-only facts cannot be smuggled in via fixtures:
// each report must independently hydrate its evidence and verify the digest.
func TestWeeklyTypedConfirmationMatchesInteractive(t *testing.T) {
	for _, name := range []string{"script:tests/check.sh", "lua:tests/native_spec.lua#native retries admission"} {
		t.Run(name, func(t *testing.T) {
			root := testfix.Repo(t, testfix.InitialCommit())
			origin := filepath.Join(t.TempDir(), "origin.git")
			testfix.Git(t, root, "init", "--bare", "-q", origin)
			testfix.Git(t, root, "remote", "add", "origin", origin)
			testfix.Git(t, root, "push", "-q", "origin", "main")
			bootstrapTracker(t, root, nil)
			common := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "--path-format=absolute", "--git-common-dir"))
			sha := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
			repo := filepath.Base(root)
			store := filepath.Join(common, "sdlc", "test-stats")
			for _, dir := range []string{"attempts", "failed"} {
				if err := os.MkdirAll(filepath.Join(store, dir), 0700); err != nil {
					t.Fatal(err)
				}
			}
			now := time.Date(2026, 10, 22, 20, 0, 0, 0, time.UTC)
			window := WeeklyWindow{Start: now.Add(-7 * 24 * time.Hour), End: now}
			id, err := flakes.ParseIdentity(name)
			if err != nil {
				t.Fatal(err)
			}
			logs := [][]byte{typedConfirmationLog(t, id.Unit(), id.Case(), true), typedConfirmationLog(t, id.Unit(), id.Case(), false)}
			rows := []SuiteAttempt{
				{Version: 4, ID: "a1", Repo: "fixture/" + repo, SHA: sha, Check: suiteCheckName, StartedAt: now.Add(-2 * time.Hour), FinishedAt: now.Add(-2*time.Hour + time.Second), Exit: 1, Failures: []string{name}, Outcome: string(flakes.Unknown), RunnerContractSHA256: strings.Repeat("b", 64)},
				{Version: 4, ID: "b2", Repo: "fixture/" + repo, SHA: sha, Check: suiteCheckName, StartedAt: now.Add(-time.Hour), FinishedAt: now.Add(-time.Hour + time.Second), Exit: 0, Outcome: string(flakes.Passed), RunnerContractSHA256: strings.Repeat("b", 64)},
			}
			for i := range rows {
				rows[i].LogPath = filepath.Join("failed", rows[i].ID+".log")
				rows[i].LogSHA256 = resumeRecordDigest(logs[i])
				raw, err := json.Marshal(rows[i])
				if err != nil {
					t.Fatal(err)
				}
				if _, err := decodeSuiteAttempt(raw); err != nil {
					t.Fatalf("invalid persisted fixture: %v", err)
				}
				if err := os.WriteFile(filepath.Join(store, "attempts", rows[i].ID+".json"), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			deps := testStatsDeps{
				now: func() time.Time { return now },
				identity: func() (workspace.Identity, error) {
					return workspace.Identity{PrimaryRoot: root, FleetRoot: filepath.Dir(root)}, nil
				},
				fleetDirs: func(string) ([]string, error) { return []string{root}, nil },
				commonDir: func(string) (string, error) { return common, nil },
			}
			for _, variant := range []string{"exact", "missing-failure-log", "missing-pass-log", "tampered-failure-log", "tampered-pass-log"} {
				t.Run(variant, func(t *testing.T) {
					for i := range rows {
						path := suiteAttemptLogFile(store, rows[i])
						missing := variant == "missing-failure-log" && i == 0 || variant == "missing-pass-log" && i == 1
						if missing {
							if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
								t.Fatal(err)
							}
							continue
						}
						raw := logs[i]
						if variant == "tampered-failure-log" && i == 0 || variant == "tampered-pass-log" && i == 1 {
							raw = append(append([]byte(nil), raw...), []byte("altered after publication\n")...)
						}
						if err := os.WriteFile(path, raw, 0600); err != nil {
							t.Fatal(err)
						}
					}
					interactive, err := buildTestStatsReport(testStatsFlags{Since: "7d", Repo: repo}, deps)
					if err != nil {
						t.Fatal(err)
					}
					if len(interactive.Repos) != 1 || len(interactive.Warnings) != 0 {
						t.Fatalf("interactive source: %+v", interactive)
					}
					evidence := readWeeklyEvidence(context.Background(), repo, root, window)
					if len(evidence.CoverageUnknown) != 0 || len(evidence.Attempts) != 2 {
						t.Fatalf("weekly source: %+v", evidence)
					}
					weekly := aggregateWeeklyRepo(repo, evidence.Attempts, window)
					want := 0
					kind := "unknown"
					if variant == "exact" {
						want = 1
						kind = "confirmed"
					}
					if interactive.Repos[0].ConfirmedFlakeAttempts != want {
						t.Errorf("interactive confirmed=%d want %d", interactive.Repos[0].ConfirmedFlakeAttempts, want)
					}
					if weekly.Suite.ConfirmedFlakeAttempts != want {
						t.Errorf("weekly confirmed=%d want %d", weekly.Suite.ConfirmedFlakeAttempts, want)
					}
					if !reflect.DeepEqual(weekly.Suite, interactive.Repos[0]) {
						t.Errorf("reporting parity:\nweekly=%+v\ninteractive=%+v", weekly.Suite, interactive.Repos[0])
					}
					if len(weekly.FailedTests) != 1 || weekly.FailedTests[0].Name != name || weekly.FailedTests[0].Kind != kind {
						t.Errorf("weekly failed_tests=%+v want %s", weekly.FailedTests, kind)
					}
				})
			}
		})
	}
}
