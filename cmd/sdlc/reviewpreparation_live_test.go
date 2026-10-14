package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewbox"
)

// Opt-in downstream conformance uses pair's real generator and runtimebundle
// tests. It never edits the source; its declaration exists only in a disposable
// clone. A shell stands in for the reviewer while production Dispatch owns
// confinement, preparation, execution and response validation.
func TestReviewPreparationLivePair(t *testing.T) {
	source := os.Getenv("WF_LIVE_PAIR_REPO")
	if source == "" {
		t.Skip("set WF_LIVE_PAIR_REPO to a local pair checkout")
	}
	t.Setenv("WF_REVIEW_TIMEOUT", "15m")
	git := func(dir string, args ...string) string {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
		return string(out)
	}
	snapshot := func() string {
		return git(source, "rev-parse", "HEAD") + git(source, "show-ref") + git(source, "status", "--porcelain=v1", "--untracked-files=all") + git(source, "diff", "HEAD", "--binary", "--no-ext-diff")
	}
	before := snapshot()
	t.Cleanup(func() {
		if after := snapshot(); after != before {
			t.Error("pair source HEAD, refs, status or tracked contents changed")
		}
	})
	pinned := strings.TrimSpace(git(source, "rev-parse", "origin/main^{commit}"))
	fence := liveReviewFence(t)
	co, err := reviewbox.Create(t.TempDir(), source, pinned, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := co.Close(); err != nil {
			t.Error(err)
		}
	}()
	asset := filepath.Join(co.Dir, "cmd/internal/runtimebundle/assets/runtime/manifest.json")
	if _, err := os.Stat(asset); !os.IsNotExist(err) {
		t.Fatalf("expected generated manifest absent before preparation: %v", err)
	}
	// Establish the downstream failure the preparation must resolve.
	baseline := exec.Command("go", "test", "./cmd/internal/runtimebundle", "-count=1")
	baseline.Dir = co.Dir
	if out, err := baseline.CombinedOutput(); err == nil || !strings.Contains(string(out), "no matching files found") {
		t.Fatalf("expected missing embed assets before preparation, got %v: %s", err, out)
	} else {
		t.Logf("before preparation: %s", out)
	}
	if err := os.MkdirAll(filepath.Join(co.Dir, ".sdlc"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(co.Dir, reviewPreparationPath), []byte(`{"version":1,"command":["make","runtimebundle-generate"],"timeout_minutes":10}`), 0600); err != nil {
		t.Fatal(err)
	}
	git(co.Dir, "add", "--", reviewPreparationPath)
	git(co.Dir, "-c", "user.name=Review preparation conformance", "-c", "user.email=conformance@example.invalid", "-c", "core.hooksPath=/dev/null", "commit", "--no-gpg-sign", "-qm", "test: declare disposable runtime preparation")
	head := strings.TrimSpace(git(co.Dir, "rev-parse", "HEAD"))
	cfg, err := readReviewPreparation(co.Dir, head)
	if err != nil || cfg == nil {
		t.Fatalf("read pinned declaration: %v, %v", cfg, err)
	}
	common := strings.TrimSpace(git(source, "rev-parse", "--path-format=absolute", "--git-common-dir"))
	protect, err := reviewbox.Canonical([]string{source, common})
	if err != nil {
		t.Fatal(err)
	}
	iso := reviewbox.Isolation{Fence: fence, Dir: co.Dir, Protect: protect}
	evidence := &judge.DispatchEvidence{}
	reviewerRuns := 0
	var diagnostics bytes.Buffer
	started := time.Now()
	output, err := judge.Dispatch(context.Background(), judge.DispatchOptions{
		Agent: judge.AgentCodex, Checkout: co.Dir, Containment: judge.ContainFenced,
		Preparation: cfg, PreparationHead: head, Evidence: evidence, Stderr: &diagnostics,
		Wrap: func(name string, args []string) (string, []string) {
			if name == "codex" {
				reviewerRuns++
				name = "/bin/sh"
				args = []string{"-c", `test -s cmd/internal/runtimebundle/assets/runtime/manifest.json && go test ./cmd/internal/runtimebundle -count=1 >&2 && printf 'VERDICT: SHIP\n'`}
			}
			return iso.Wrap(name, args)
		},
		ValidateResponse: func(output string) error {
			if strings.TrimSpace(output) != "VERDICT: SHIP" {
				return fmt.Errorf("runtimebundle conformance did not pass: %q", output)
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("dispatch: %v; evidence=%+v", err, evidence)
	}
	if reviewerRuns != 1 || len(evidence.Preparation) != 1 || evidence.Preparation[0].Error != "" {
		t.Fatalf("unexpected attempts/evidence: %d %+v", reviewerRuns, evidence)
	}
	if _, err := os.Stat(asset); err != nil {
		t.Fatal(err)
	}
	if dirty := git(co.Dir, "diff", "HEAD", "--name-only"); dirty != "" {
		t.Fatalf("tracked clone mutations: %s", dirty)
	}
	t.Logf("source origin/main=%s; declaration head=%s; fence=%s; prep=%s; total=%s; result=%s; source unchanged", pinned, head, fence, evidence.Preparation[0].Duration, time.Since(started), strings.TrimSpace(output))
	t.Logf("preparation output: %s", evidence.Preparation[0].Output)
	t.Logf("reviewer test diagnostics: %s", diagnostics.String())
}

// TestMain deliberately replaces Probe; prove the selected platform fence with
// real writes instead of accepting that hermetic checkout-only default.
func liveReviewFence(t *testing.T) reviewbox.Fence {
	t.Helper()
	var fence reviewbox.Fence
	switch runtime.GOOS {
	case "darwin":
		fence = reviewbox.FenceSeatbelt
	case "linux":
		fence = reviewbox.FenceBwrap
	default:
		t.Fatalf("no live fence supported on %s", runtime.GOOS)
	}
	paths, err := reviewbox.Canonical([]string{t.TempDir(), t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	iso := reviewbox.Isolation{Fence: fence, Dir: paths[0], Protect: paths[1:]}
	name, args := iso.Wrap("/bin/sh", []string{"-c", `: > "$1/allowed" && if ( : > "$2/denied" ) 2>/dev/null; then exit 3; fi`, "sh", paths[0], paths[1]})
	if out, err := exec.Command(name, args...).CombinedOutput(); err != nil {
		t.Fatalf("live fence conformance: %v: %s", err, out)
	}
	if _, err := os.Stat(filepath.Join(paths[0], "allowed")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(paths[1], "denied")); !os.IsNotExist(err) {
		t.Fatalf("fence allowed protected write: %v", err)
	}
	return fence
}

// Requires installed Codex and its existing credentials. The real agent, runner
// and Wrap execute ps; no injected Run or substituted reviewer command is used.
func TestReviewPreparationLiveCodexPS(t *testing.T) {
	if os.Getenv("WF_LIVE_REVIEW_AGENT") != "codex" {
		t.Skip("set WF_LIVE_REVIEW_AGENT=codex for installed-agent conformance")
	}
	t.Setenv("WF_REVIEW_TIMEOUT", "5m")
	fence := liveReviewFence(t)
	source := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"-c", "user.name=Conformance", "-c", "user.email=conformance@example.invalid", "-c", "core.hooksPath=/dev/null", "commit", "--no-gpg-sign", "--allow-empty", "-qm", "test: live reviewer fixture"}} {
		if out, err := exec.Command("git", append([]string{"-C", source}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git: %v: %s", err, out)
		}
	}
	co, err := reviewbox.Create(t.TempDir(), source, "HEAD", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := co.Close(); err != nil {
			t.Error(err)
		}
	}()
	protect, err := reviewbox.Canonical([]string{source, filepath.Join(source, ".git")})
	if err != nil {
		t.Fatal(err)
	}
	iso := reviewbox.Isolation{Fence: fence, Dir: co.Dir, Protect: protect}
	var diagnostics bytes.Buffer
	started := time.Now()
	output, err := judge.Dispatch(context.Background(), judge.DispatchOptions{
		Agent: judge.AgentCodex, Checkout: co.Dir, Containment: judge.ContainFenced, Wrap: iso.Wrap, Stderr: &diagnostics,
		Prompt:           `This is a bounded live OS fence conformance check. Execute exactly /bin/ps -p 1 -o pid=,comm= in the foreground using your shell tool. Do not inspect or edit files or run any other commands. Only if that actual command exits zero and returns PID 1, lead your final response with a fenced verdict block containing verdict: SHIP, then PS_READ_OK and the actual ps output, then a fenced findings block containing findings: [] and dispose: []. Otherwise emit verdict: REWORK and a blocking kind: unverifiable finding naming the failed command, observed failure and evidence needed. Do not claim success from the prompt or from prior knowledge.` + "\nUse these literal fences (not yaml fences):\n```verdict\nverdict: SHIP\n```\nPS_READ_OK and actual output\n```findings\nfindings: []\ndispose: []\n```",
		ValidateResponse: func(output string) error { return validateBoundaryResponse(output, gatestate.Ledger{}) },
		BeforeRetry:      co.Reset,
	})
	t.Logf("installed Codex transcript (fence=%s, mode=danger-full-access under outer fence, duration=%s):\n%s", fence, time.Since(started), diagnostics.String())
	t.Logf("semantic response: %s", output)
	if err != nil {
		t.Fatal(err)
	}
	if judge.ParseVerdict(output) != judge.VerdictShip || !strings.Contains(output, "PS_READ_OK") {
		t.Fatalf("ps inspection did not succeed: %s", output)
	}
	if !strings.Contains(diagnostics.String(), "/bin/ps -p 1 -o pid=,comm=") || !strings.Contains(diagnostics.String(), "succeeded in") {
		t.Fatal("transcript lacks successful real tool execution")
	}
}
