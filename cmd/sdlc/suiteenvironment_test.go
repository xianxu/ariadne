package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"iter"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

const workDirFailure = "go: creating work dir: mkdir /blocked/go-build123: permission denied"

func environmentCheck() localCheck {
	return localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", "echo '" + workDirFailure + "'; exit 1"}, TimeoutMinutes: 1}
}

func TestSuiteEnvironmentFailureDoesNotPoisonRepair(t *testing.T) {
	d, sha, gh, stderr := localRunRepo(t, "skipped")
	registryCalls := 0
	d.FlakeRegistry = func(context.Context) (flakes.Registry, error) { registryCalls++; return flakes.Registry{}, nil }
	rec, err := runLocalCheck(context.Background(), d, environmentCheck(), sha)
	if err == nil || !strings.Contains(err.Error(), "environment") || !strings.Contains(err.Error(), "GOTMPDIR") {
		t.Fatalf("expected environment refusal: rec=%+v err=%v\n%s", rec, err, stderr)
	}
	if registryCalls != 0 {
		t.Fatal("startup failure entered flake classification")
	}
	if diagnoseSuiteFailure(context.Background(), err, sha, d) != err {
		t.Fatal("direct environment error triggered baseline diagnosis")
	}
	cTrain := &conductor{ctx: context.Background(), heads: []string{sha}}
	cTrain.run = func(context.Context, localCheck, string) (localRunRecord, error) { return rec, err }
	if event, trainErr := cTrain.runSuite(1, landingline.ExecInitial); event != nil || trainErr == nil || cTrain.baseline != nil {
		t.Fatalf("direct startup blamed train cars: %v %v", event, trainErr)
	}
	if len(gh.statuses[sha]) != 0 || len(gh.comments[9]) != 0 {
		t.Fatalf("environment error published code failure: %v %v", gh.statuses, gh.comments)
	}
	log, readErr := os.ReadFile(rec.LogPath)
	if readErr != nil || !strings.Contains(string(log), workDirFailure) || !strings.HasSuffix(rec.LogPath, ".environment.log") {
		t.Fatalf("environment diagnostic not retained: %s %q %v", rec.LogPath, log, readErr)
	}
	logDir, _ := localCheckLogDir(d)
	rows, warnings, readErr := readSuiteAttempts(suiteStoreRoot(logDir), 100)
	if readErr != nil || len(warnings) != 0 || len(rows) != 0 {
		t.Fatalf("environment became suite evidence: %+v %v %v", rows, warnings, readErr)
	}
	legacy, warnings := readLegacySuiteLogs(filepath.Dir(filepath.Dir(logDir)), d.Repo, nil)
	if len(legacy) != 0 || len(warnings) != 0 {
		t.Fatalf("environment became legacy suite evidence: %+v %v", legacy, warnings)
	}
	c := environmentCheck()
	c.Command = []string{"true"}
	if repaired, err := runLocalCheck(context.Background(), d, c, sha); err != nil || !repaired.passed() || gh.statuses[sha][localStatusContext(suiteCheckName)] != "success" {
		t.Fatalf("same-head repaired environment blocked: %+v %v\n%s", repaired, err, stderr)
	}
}

func TestSuiteEnvironmentBaselineIsNotRedMain(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "skipped")
	log := filepath.Join(t.TempDir(), "train.log")
	if err := os.WriteFile(log, []byte("--- FAIL: TestActual\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	check := environmentCheck()
	failed := localRunRecord{Exit: 1, LogPath: log}
	err := diagnoseSuiteFailure(context.Background(), &localCheckFailure{Check: check, Run: failed, Head: sha}, sha, d)
	if err == nil || !strings.Contains(err.Error(), "environment") || !strings.Contains(err.Error(), "GOTMPDIR") || strings.Contains(err.Error(), "main is red") {
		t.Fatalf("ordinary baseline misclassified: %v", err)
	}
	c := &conductor{ctx: context.Background(), heads: []string{sha}, base: sha, suite: check, local: d}
	c.run = func(context.Context, localCheck, string) (localRunRecord, error) { return failed, nil }
	if ev, err := c.runSuite(1, landingline.ExecInitial); err == nil || ev != nil || !strings.Contains(err.Error(), "environment") || !strings.Contains(err.Error(), "GOTMPDIR") {
		t.Fatalf("train baseline blamed code: %+v %v", ev, err)
	}
	if gh.statuses[sha] != nil || len(gh.comments[9]) != 0 {
		t.Fatalf("baseline environment posted status/comment: %v %v", gh.statuses, gh.comments)
	}
}

func TestScratchFallbackOverridesInheritedTemp(t *testing.T) {
	d, sha, _, stderr := localRunRepo(t, "skipped")
	root, normal := t.TempDir(), t.TempDir()
	// Checkout setup has a usable private root. Invalidate inherited env only
	// after it exists: this test exercises the child-run fallback boundary.
	t.Setenv("WF_TEST_SCRATCH", "off")
	t.Setenv("TMPDIR", root)
	wt, cleanup, err := detachedWorktree(d.Git, sha)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	helper, err := prepareSuiteHelper(context.Background(), d.Root, "")
	if err != nil {
		t.Fatal(err)
	}
	logDir, err := localCheckLogDir(d)
	if err != nil {
		t.Fatal(err)
	}
	blocked := filepath.Join(root, "not-a-directory")
	if err := os.WriteFile(blocked, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	calls := 0
	d.normalTempDir = func() (string, error) { calls++; return normal, nil }
	t.Setenv("WF_TEST_SCRATCH", blocked)
	t.Setenv("TMPDIR", filepath.Join(blocked, "temp"))
	t.Setenv("GOTMPDIR", filepath.Join(blocked, "go"))
	argv := []string{"sh", "-c", `test -d "$TMPDIR" && test -d "$GOTMPDIR" && printf ok > "$TMPDIR/probe" && printf ok > "$GOTMPDIR/probe" && printf '%s\n%s\n' "$TMPDIR" "$GOTMPDIR"`}
	log := filepath.Join(root, "child.log")
	rec, err := runUnderSlot(context.Background(), d, wt, helper, logDir, "temp-fixture", sha, 1, suitePrioritySubject{}, argv, log, time.Minute)
	if err != nil || rec.Exit != 0 {
		t.Fatalf("fallback inherited invalid temp env: %+v %v\n%s", rec, err, stderr)
	}
	out, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	paths := strings.Split(strings.TrimSpace(string(out)), "\n")
	if calls != 1 || len(paths) != 2 || paths[0] != paths[1] || filepath.Dir(paths[0]) != normal {
		t.Fatalf("child did not use the private fallback: calls=%d paths=%q", calls, paths)
	}
	if _, err := os.Stat(paths[0]); !os.IsNotExist(err) {
		t.Fatalf("run directory survived cleanup: %v", err)
	}
	for _, variable := range []string{"GOTMPDIR", "TMPDIR"} {
		if !strings.Contains(stderr.String(), variable) {
			t.Fatalf("replacement note lacks %s:\n%s", variable, stderr)
		}
	}
}

func TestScratchFallbackUnavailableDoesNotRunOrRecord(t *testing.T) {
	d, sha, gh, stderr := localRunRepo(t, "skipped")
	root := t.TempDir()
	t.Setenv("WF_TEST_SCRATCH", "off")
	t.Setenv("TMPDIR", root)
	calls := 0
	d.normalTempDir = func() (string, error) { calls++; return "", os.ErrPermission }
	d.FlakeRegistry = func(context.Context) (flakes.Registry, error) {
		t.Fatal("temp selection failure entered flake classification")
		return flakes.Registry{}, nil
	}
	marker := filepath.Join(root, "must-not-run")
	c := localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", `printf ran > "$1"`, "fixture", marker}, TimeoutMinutes: 1}
	rec, err := runLocalCheck(context.Background(), d, c, sha)
	if calls != 1 || !errors.Is(err, os.ErrPermission) || !strings.Contains(err.Error(), "environment failure") || rec.Exit != -1 {
		t.Fatalf("unavailable fallback was not an environment refusal: calls=%d rec=%+v err=%v\n%s", calls, rec, err, stderr)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("child ran after selection failed: %v", err)
	}
	if len(gh.statuses[sha]) != 0 || len(gh.comments[9]) != 0 {
		t.Fatalf("environment failure published suite evidence: %v %v", gh.statuses, gh.comments)
	}
	logDir, err := localCheckLogDir(d)
	if err != nil {
		t.Fatal(err)
	}
	rows, warnings, err := readSuiteAttempts(suiteStoreRoot(logDir), 100)
	if err != nil || len(warnings) != 0 || len(rows) != 0 {
		t.Fatalf("environment failure persisted suite evidence: %+v %v %v", rows, warnings, err)
	}
}

func TestSuiteEnvironmentGrammar(t *testing.T) {
	wrapper, err := os.ReadFile("testdata/suite-workdir-startup.log")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, log string
		want      bool
	}{
		{"direct", workDirFailure + "\n", true},
		{"multiple", strings.Repeat(workDirFailure+"\n", 8), true},
		{"wrapper", string(wrapper), true},
		{"make", workDirFailure + "\nmake: *** [test] Error 1\n", true},
		{"empty", "", false},
		{"truncated diagnostic", "go: creating work dir: mkdir /blocked", false},
		{"truncated wrapper", strings.TrimSuffix(string(wrapper), "make: *** [test] Error 1\n")[:100], false},
		{"missing terminal newline", workDirFailure, false},
		{"quoted", "    " + workDirFailure + "\n", false},
		{"incidental", "example: " + workDirFailure + "\n", false},
		{"named failure", workDirFailure + "\n--- FAIL: TestActual (0.00s)\n", false},
		{"pass activity", workDirFailure + "\nok  example/pkg 0.1s\n", false},
		{"running test", "=== RUN   TestActual\n" + workDirFailure + "\n", false},
		{"compiler error", workDirFailure + "\nx.go:3: undefined: broken\n", false},
		{"other go error", "go: downloading broken\n" + workDirFailure + "\n", false},
		{"unknown", workDirFailure + "\nunknown failure\n", false},
		{"unknown traceback", strings.Replace(string(wrapper), "sys.exit(main())", "raise RuntimeError('broken')", 1), false},
		{"long line", workDirFailure + strings.Repeat("x", 1<<20) + "\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := suiteEnvironmentLog(tc.log); got != tc.want {
				t.Fatalf("got %v want %v: %q", got, tc.want, tc.log[:min(len(tc.log), 500)])
			}
		})
	}
}

func TestSuiteEnvironmentLinearScan(t *testing.T) {
	for _, n := range []int{8, 16} {
		visits := 0
		lines := func(yield func(string) bool) {
			for range n {
				visits++
				if !yield(workDirFailure + "\n") {
					return
				}
			}
		}
		if !suiteEnvironmentLines(iter.Seq[string](lines)) || visits != n {
			t.Fatalf("N=%d visits=%d", n, visits)
		}
	}
}

func TestSuiteEnvironmentPreservesPriorCodeFailure(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "skipped")
	if rec, err := runLocalCheck(context.Background(), d, suiteFailing(gateUnknown), sha); err != nil || rec.passed() {
		t.Fatalf("first failure: %+v %v", rec, err)
	}
	if _, err := runLocalCheck(context.Background(), d, environmentCheck(), sha); err == nil {
		t.Fatal("environment failure accepted")
	}
	pass := environmentCheck()
	pass.Command = []string{"true"}
	_, err := runLocalCheck(context.Background(), d, pass, sha)
	var unexplained *unexplainedFailureError
	if !errors.As(err, &unexplained) || gh.statuses[sha]["sdlc-local/suite"] != "failure" {
		t.Fatalf("real failure erased: %v %v", err, gh.statuses[sha])
	}
}

func TestSuiteEnvironmentKnownFlakeRerun(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "skipped")
	d.FlakeRegistry = gateRegistry(t, gateKnown)
	scriptRerun(t, "echo '"+workDirFailure+"'; exit 1")
	_, err := runLocalCheck(context.Background(), d, suiteFailing(gateKnown), sha)
	var environment *suiteEnvironmentError
	if !errors.As(err, &environment) {
		t.Fatalf("rerun did not report environment failure: %v", err)
	}
	if !strings.Contains(err.Error(), "sdlc local-check run --name suite --rev "+sha+";") || strings.Contains(err.Error(), "--name suite-rerun") {
		t.Fatalf("rerun recovery names an internal record: %v", err)
	}
	rows := gateAttempts(t, d)
	if len(rows) != 1 || rows[0].Check != suiteCheckName || len(gh.statuses[sha]) != 0 {
		t.Fatalf("environment became rerun failure/status: %+v %v", rows, gh.statuses[sha])
	}
	if log, err := os.ReadFile(environment.LogPath); err != nil || !strings.Contains(string(log), workDirFailure) {
		t.Fatalf("rerun diagnostic: %q %v", log, err)
	}
}

func TestSuiteEnvironmentLegacyFallback(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "skipped")
	logDir, _ := localCheckLogDir(d)
	store := suiteStoreRoot(logDir)
	if err := os.MkdirAll(filepath.Dir(store), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store, []byte("blocked store"), 0o600); err != nil {
		t.Fatal(err)
	}
	rec, err := runLocalCheck(context.Background(), d, environmentCheck(), sha)
	var environment *suiteEnvironmentError
	if !errors.As(err, &environment) || len(gh.statuses[sha]) != 0 {
		t.Fatalf("legacy startup: %+v %v", rec, err)
	}
	if _, err := os.Stat(filepath.Join(logDir, sha, "suite.log")); !os.IsNotExist(err) {
		t.Fatalf("legacy code evidence remains: %v", err)
	}
	if _, err := os.ReadFile(rec.LogPath); err != nil {
		t.Fatal(err)
	}
}

func TestSuiteEnvironmentRequiresFailedCompletedSuite(t *testing.T) {
	for _, mode := range []string{"pass", "timeout", "other-check"} {
		t.Run(mode, func(t *testing.T) {
			d, sha, gh, _ := localRunRepo(t, "skipped")
			c := environmentCheck()
			switch mode {
			case "pass":
				c.Command[2] = "echo '" + workDirFailure + "'; exit 0"
			case "timeout":
				c.Command[2] = "echo '" + workDirFailure + "'; sleep 10"
			case "other-check":
				c.Name = "other"
			}
			rec, err := runLocalCheck(context.Background(), d, c, sha)
			if err != nil || len(gh.statuses[sha]) == 0 || strings.HasSuffix(rec.LogPath, ".environment.log") {
				t.Fatalf("ordinary outcome changed: %+v %v", rec, err)
			}
		})
	}
}

func TestSuiteEnvironmentRetentionFailureKeepsRawEvidence(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(fmt.Sprint(legacy), func(t *testing.T) {
			d, sha, gh, _ := localRunRepo(t, "skipped")
			logDir, _ := localCheckLogDir(d)
			if err := os.MkdirAll(filepath.Join(logDir, sha, "suite.environment.log"), 0o700); err != nil {
				t.Fatal(err)
			}
			if legacy {
				store := suiteStoreRoot(logDir)
				if err := os.MkdirAll(store, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(store, "inflight"), nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			rec, err := runLocalCheck(context.Background(), d, environmentCheck(), sha)
			var environment *suiteEnvironmentError
			if !errors.As(err, &environment) || !strings.Contains(err.Error(), "retain environment diagnostic") || len(gh.statuses[sha]) != 0 {
				t.Fatalf("retention error hidden: %+v %v", rec, err)
			}
			if log, err := os.ReadFile(rec.LogPath); err != nil || !strings.Contains(string(log), workDirFailure) {
				t.Fatalf("raw diagnostic deleted: %q %v", log, err)
			}
			if _, err := os.Stat(filepath.Join(logDir, sha, "suite.log")); !os.IsNotExist(err) {
				t.Fatalf("raw environment became legacy result: %v", err)
			}
		})
	}
}
