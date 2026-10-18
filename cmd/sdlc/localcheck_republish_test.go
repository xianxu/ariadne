package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSavedAttemptRecord(t *testing.T) {
	wall := 61.4
	base := SuiteAttempt{ID: "a1", Repo: "o/r", SHA: "abc123", Check: "suite", Exit: 0, WallSeconds: &wall, LogSHA256: "feed"}
	with := func(f func(*SuiteAttempt)) SuiteAttempt { a := base; f(&a); return a }
	for _, tc := range []struct {
		name   string
		a      SuiteAttempt
		logSHA string
		want   string // "" publishes
	}{
		{"matching record publishes", base, "feed", ""},
		{"a failure publishes as a failure", with(func(a *SuiteAttempt) { a.Exit = 2 }), "feed", ""},
		{"tampered log", base, "beef", "no longer hashes"},
		{"other commit", with(func(a *SuiteAttempt) { a.SHA = "def456" }), "feed", "not abc123"},
		{"other check", with(func(a *SuiteAttempt) { a.Check = "e2e" }), "feed", "not suite"},
		{"other repository", with(func(a *SuiteAttempt) { a.Repo = "x/y" }), "feed", "not o/r"},
		{"main baseline", with(func(a *SuiteAttempt) { a.MainBaseline = true }), "feed", "baseline"},
		{"incomplete: no duration", with(func(a *SuiteAttempt) { a.WallSeconds = nil }), "feed", "incomplete"},
		{"incomplete: no log hash", with(func(a *SuiteAttempt) { a.LogSHA256 = "" }), "", "incomplete"},
	} {
		rec, err := savedAttemptRecord(tc.a, tc.logSHA, []string{"tail"}, "o/r", "abc123", "suite", "host1")
		if tc.want == "" {
			if err != nil {
				t.Errorf("%s: %v", tc.name, err)
				continue
			}
			if rec.Exit != tc.a.Exit || rec.SHA != "abc123" || rec.Name != "suite" || rec.Host != "host1" || rec.Duration != 61400*time.Millisecond || rec.LogSHA256 != "feed" {
				t.Errorf("%s: record %+v", tc.name, rec)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: got %v, want an error containing %q", tc.name, err, tc.want)
		}
	}
}

// flakyPostGH fails the first n status posts, as an unreachable GitHub does.
type flakyPostGH struct {
	*ciFakeGH
	mu    sync.Mutex
	fails int
	posts int
}

func (g *flakyPostGH) PostStatus(ctx context.Context, repo, sha, statusContext, state, desc string) error {
	g.mu.Lock()
	g.posts++
	failing := g.fails > 0
	if failing {
		g.fails--
	}
	g.mu.Unlock()
	if failing {
		return errors.New("error connecting to api.github.com")
	}
	return g.ciFakeGH.PostStatus(ctx, repo, sha, statusContext, state, desc)
}

func suiteCheck(command ...string) localCheck {
	return localCheck{Name: "suite", Command: command, TimeoutMinutes: 1, Weight: 1}
}

// A brief outage is ridden out by the bounded retry; a longer one leaves the
// result saved, and the failure names the command that publishes it.
func TestRunLocalCheckPostRetryAndRecovery(t *testing.T) {
	ctx := context.Background()
	t.Run("retry rides out a brief outage", func(t *testing.T) {
		d, sha, gh, stderr := localRunRepo(t, "skipped")
		flaky := &flakyPostGH{ciFakeGH: gh, fails: 2}
		d.GH = flaky
		if _, err := runLocalCheck(ctx, d, suiteCheck("true"), sha); err != nil {
			t.Fatalf("run: %v\n%s", err, stderr.String())
		}
		if flaky.posts != 3 || gh.statuses[sha][localStatusContext("suite")] != "success" {
			t.Errorf("posts %d, statuses %v", flaky.posts, gh.statuses[sha])
		}
	})
	t.Run("a lasting outage names the republish command", func(t *testing.T) {
		d, sha, gh, stderr := localRunRepo(t, "skipped")
		d.GH = &flakyPostGH{ciFakeGH: gh, fails: 1 + len(localCheckPostBackoff)}
		_, err := runLocalCheck(ctx, d, suiteCheck("true"), sha)
		if err == nil || !strings.Contains(err.Error(), republishCommand("suite", sha)) || !strings.Contains(err.Error(), "no rerun needed") {
			t.Fatalf("want the recovery command in: %v\n%s", err, stderr.String())
		}
		if gh.statuses[sha] != nil {
			t.Fatalf("nothing should be recorded yet: %v", gh.statuses[sha])
		}
		// The recovery, as the republish verb runs it: load and post the saved result.
		d.GH = gh
		logDir, _ := localCheckLogDir(d)
		rec, _, err := loadSavedAttempt(suiteStoreRoot(logDir), d.Repo, sha, "suite", "", "host1")
		if err != nil {
			t.Fatal(err)
		}
		if err := postLocalCheckRecord(ctx, d, rec); err != nil {
			t.Fatal(err)
		}
		if gh.statuses[sha][localStatusContext("suite")] != "success" {
			t.Errorf("republished status: %v", gh.statuses[sha])
		}
	})
}

func TestLoadSavedAttemptRefuses(t *testing.T) {
	ctx := context.Background()
	d, sha, _, stderr := localRunRepo(t, "skipped")
	if _, err := runLocalCheck(ctx, d, suiteCheck("sh", "-c", "echo suite output"), sha); err != nil {
		t.Fatalf("run: %v\n%s", err, stderr.String())
	}
	logDir, _ := localCheckLogDir(d)
	store := suiteStoreRoot(logDir)
	if _, _, err := loadSavedAttempt(store, d.Repo, sha, "suite", "", "h"); err != nil {
		t.Fatalf("the saved attempt should verify: %v", err)
	}
	other := strings.Repeat("0", len(sha))
	if _, _, err := loadSavedAttempt(store, d.Repo, other, "suite", "", "h"); !errors.Is(err, errNoSavedAttempt) {
		t.Errorf("other commit: %v", err)
	}
	if _, _, err := loadSavedAttempt(store, d.Repo, sha, "e2e", "", "h"); !errors.Is(err, errNoSavedAttempt) {
		t.Errorf("other check: %v", err)
	}
	if _, _, err := loadSavedAttempt(store, d.Repo, sha, "suite", "0123456789abcdef0123456789abcdef", "h"); !errors.Is(err, errNoSavedAttempt) {
		t.Errorf("missing attempt id: %v", err)
	}
	if _, _, err := loadSavedAttempt(store, "x/y", sha, "suite", "", "h"); !errors.Is(err, errSavedAttemptRefused) {
		t.Errorf("other repository: %v", err)
	}
	log := filepath.Join(logDir, sha, "suite.log")
	if err := os.WriteFile(log, []byte("tampered\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadSavedAttempt(store, d.Repo, sha, "suite", "", "h"); !errors.Is(err, errSavedAttemptRefused) || !strings.Contains(err.Error(), "no longer hashes") {
		t.Errorf("tampered log: %v", err)
	}
	if err := os.Remove(log); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadSavedAttempt(store, d.Repo, sha, "suite", "", "h"); !errors.Is(err, errSavedAttemptRefused) {
		t.Errorf("missing log: %v", err)
	}
}

// Merge's runner posts a saved, verified pass for the exact head instead of
// rerunning; a saved failure is never reused.
func TestLocalCheckRunnerReusesSavedPass(t *testing.T) {
	ctx := context.Background()
	t.Run("saved pass is posted, not rerun", func(t *testing.T) {
		d, sha, gh, stderr := localRunRepo(t, "skipped")
		d.GH = &flakyPostGH{ciFakeGH: gh, fails: 1 + len(localCheckPostBackoff)}
		if _, err := runLocalCheck(ctx, d, suiteCheck("true"), sha); err == nil {
			t.Fatal("the first post should have failed")
		}
		d.GH = gh
		marker := filepath.Join(t.TempDir(), "ran")
		rec, err := newLocalCheckRunner(d)(ctx, suiteCheck("./marker-check.sh", marker), sha)
		if err != nil || !rec.passed() {
			t.Fatalf("runner: %+v %v\n%s", rec, err, stderr.String())
		}
		if _, err := os.Stat(marker); err == nil {
			t.Error("the check was rerun despite a saved pass")
		}
		if gh.statuses[sha][localStatusContext("suite")] != "success" {
			t.Errorf("saved pass not posted: %v", gh.statuses[sha])
		}
	})
	t.Run("saved failure is rerun", func(t *testing.T) {
		d, sha, gh, stderr := localRunRepo(t, "skipped")
		d.GH = &flakyPostGH{ciFakeGH: gh, fails: 1 + len(localCheckPostBackoff)}
		_, _ = runLocalCheck(ctx, d, suiteCheck("false"), sha)
		d.GH = gh
		marker := filepath.Join(t.TempDir(), "ran")
		if _, err := newLocalCheckRunner(d)(ctx, suiteCheck("./marker-check.sh", marker), sha); err != nil {
			t.Fatalf("runner: %v\n%s", err, stderr.String())
		}
		if _, err := os.Stat(marker); err != nil {
			t.Error("a saved failure must not stand in for a run")
		}
	})
}

// prFakeGH answers the branch's PR, so a republished failure also comments.
type prFakeGH struct {
	*ciFakeGH
	pr string
}

func (g prFakeGH) PRListForBranch(string, string) (string, error) { return g.pr, nil }

// The verb, end to end from the checkout: --rev resolution, the store path,
// the PR lookup, and its refusals.
func TestLocalCheckRepublishCmd(t *testing.T) {
	ctx := context.Background()
	d, sha, gh, stderr := localRunRepo(t, "skipped")
	d.GH = &flakyPostGH{ciFakeGH: gh, fails: 1 + len(localCheckPostBackoff)}
	if _, err := runLocalCheck(ctx, d, suiteCheck("sh", "-c", "echo boom; exit 3"), sha); err == nil {
		t.Fatalf("the post should have failed\n%s", stderr.String())
	}
	previousGH, previousRepo := ghClient, detectRepo
	ghClient = prFakeGH{ciFakeGH: gh, pr: "9"}
	detectRepo = func() (string, error) { return d.Repo, nil }
	t.Cleanup(func() { ghClient, detectRepo = previousGH, previousRepo })
	t.Chdir(d.Root)
	var out strings.Builder
	if err := runLocalCheckRepublish(ctx, &out, republishFlags{}); err == nil || !strings.Contains(err.Error(), "give --name") {
		t.Errorf("missing --name: %v", err)
	}
	if err := runLocalCheckRepublish(ctx, &out, republishFlags{Name: "suite", Rev: "HEAD"}); err != nil {
		t.Fatalf("republish: %v\n%s", err, out.String())
	}
	if gh.statuses[sha][localStatusContext("suite")] != "failure" {
		t.Errorf("a saved failure posts as a failure: %v", gh.statuses[sha])
	}
	if len(gh.comments[9]) != 1 || !strings.Contains(gh.comments[9][0], "boom") {
		t.Errorf("failure comment on the branch's PR: %v", gh.comments[9])
	}
	if out, err := d.Git("commit", "-q", "--allow-empty", "-m", "later"); err != nil {
		t.Fatal(out, err)
	}
	if err := runLocalCheckRepublish(ctx, &out, republishFlags{Name: "suite", Rev: "HEAD"}); !errors.Is(err, errNoSavedAttempt) {
		t.Errorf("another commit has no saved attempt: %v", err)
	}
	if err := runLocalCheckRepublish(ctx, &out, republishFlags{Name: "suite", Rev: "no-such-rev"}); err == nil || !strings.Contains(err.Error(), "not a commit") {
		t.Errorf("bad --rev: %v", err)
	}
}

func TestLocalCheckRunnerSavedPassEdges(t *testing.T) {
	ctx := context.Background()
	t.Run("a failed republish stops with the recovery line, never reruns", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "skipped")
		d.GH = &flakyPostGH{ciFakeGH: gh, fails: 1 + len(localCheckPostBackoff)}
		_, _ = runLocalCheck(ctx, d, suiteCheck("true"), sha)
		d.GH = &flakyPostGH{ciFakeGH: gh, fails: 1 + len(localCheckPostBackoff)}
		marker := filepath.Join(t.TempDir(), "ran")
		_, err := newLocalCheckRunner(d)(ctx, suiteCheck("./marker-check.sh", marker), sha)
		if err == nil || !strings.Contains(err.Error(), republishCommand("suite", sha)) {
			t.Errorf("want the recovery line: %v", err)
		}
		if _, statErr := os.Stat(marker); statErr == nil {
			t.Error("a failed republish fell back to rerunning")
		}
	})
	t.Run("a pass with a later unverifiable attempt is rerun", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "skipped")
		d.GH = &flakyPostGH{ciFakeGH: gh, fails: 2 * (1 + len(localCheckPostBackoff))}
		logDir, _ := localCheckLogDir(d)
		store, passLog := suiteStoreRoot(logDir), filepath.Join(logDir, sha, "suite.log")
		_, _ = runLocalCheck(ctx, d, suiteCheck("sh", "-c", "echo pass output"), sha)
		passBytes, err := os.ReadFile(passLog)
		if err != nil {
			t.Fatal(err)
		}
		// The later failure overwrites the shared per-head log and keeps its own copy.
		_, _ = runLocalCheck(ctx, d, suiteCheck("sh", "-c", "echo fail output; exit 1"), sha)
		// Reach the guard on purpose: the pass verifies again (its bytes restored),
		// the later failure cannot (its own log pruned).
		if err := os.WriteFile(passLog, passBytes, 0o644); err != nil {
			t.Fatal(err)
		}
		failed, _ := filepath.Glob(filepath.Join(store, "failed", "*.log"))
		if len(failed) != 1 {
			t.Fatalf("want the failure's own log, got %v", failed)
		}
		if err := os.Remove(failed[0]); err != nil {
			t.Fatal(err)
		}
		rec, newest, err := loadSavedAttempt(store, d.Repo, sha, "suite", "", "h")
		if err != nil || !rec.passed() || newest {
			t.Fatalf("want a verifying pass that is not the newest: %+v newest=%v %v", rec, newest, err)
		}
		d.GH = gh
		var stderr strings.Builder
		d.Stderr = &stderr
		marker := filepath.Join(t.TempDir(), "ran")
		if _, err := newLocalCheckRunner(d)(ctx, suiteCheck("./marker-check.sh", marker), sha); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(marker); err != nil {
			t.Error("an older pass stood in although a later attempt exists")
		}
		if !strings.Contains(stderr.String(), "a later attempt exists and does not verify") {
			t.Errorf("the not-newest branch did not report itself:\n%s", stderr.String())
		}
	})
	t.Run("an incomplete store read leaves newest unknown", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "skipped")
		d.GH = &flakyPostGH{ciFakeGH: gh, fails: 1 + len(localCheckPostBackoff)}
		_, _ = runLocalCheck(ctx, d, suiteCheck("true"), sha)
		logDir, _ := localCheckLogDir(d)
		store := suiteStoreRoot(logDir)
		// An unreadable record could be a later attempt of this very check.
		if err := os.WriteFile(filepath.Join(store, "attempts", "ffffffffffffffffffffffffffffffff.json"), []byte("{"), 0o600); err != nil {
			t.Fatal(err)
		}
		if rec, newest, err := loadSavedAttempt(store, d.Repo, sha, "suite", "", "h"); err != nil || !rec.passed() || newest {
			t.Fatalf("want a verifying pass whose newness is unknown: newest=%v %v", newest, err)
		}
		d.GH = gh
		marker := filepath.Join(t.TempDir(), "ran")
		if _, err := newLocalCheckRunner(d)(ctx, suiteCheck("./marker-check.sh", marker), sha); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(marker); err != nil {
			t.Error("a pass stood in although the store read was incomplete")
		}
	})
}
