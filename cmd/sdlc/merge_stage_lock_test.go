package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

type mergeStageFixture struct {
	root, server, target, admission, admit, running, release, built string
	gh                                                              *landingFakeGH
}

func newMergeStageFixture(t *testing.T) *mergeStageFixture {
	t.Helper()
	root, server, gh := integrationFixture(t)
	dir := t.TempDir()
	f := &mergeStageFixture{root: root, server: server, admission: filepath.Join(dir, "admission"), admit: filepath.Join(dir, "admit"), running: filepath.Join(dir, "running"), release: filepath.Join(dir, "release"), built: filepath.Join(dir, "built"), gh: gh}
	id := procedureIdentity(t, root, ":0")
	f.target = id.WorktreeRoot
	for name, value := range map[string]string{
		"MERGE_STAGE_ADMISSION": f.admission, "MERGE_STAGE_ADMIT": f.admit,
		"MERGE_STAGE_RUNNING": f.running, "MERGE_STAGE_RELEASE": f.release, "MERGE_STAGE_BUILD": f.built,
	} {
		t.Setenv(name, value)
	}
	// The helper is a real subprocess with private admission. Capability probes
	// must return before the acquisition barrier is armed.
	helper := fakeSuiteLockHelper("import os, time\nopen(os.environ['MERGE_STAGE_ADMISSION'], 'w').close()\nwhile not os.path.exists(os.environ['MERGE_STAGE_ADMIT']):\n    time.sleep(0.01)\nprint('acquired')\n")
	for path, content := range map[string]string{
		"scripts/test-lock.py": helper,
		"Makefile":             "build:\n\t@git rev-parse HEAD > \"$$MERGE_STAGE_BUILD\"\n",
		localChecksPath:        `{"version":1,"checks":[{"name":"suite","required":true,"command":["sh","-c","touch \"$MERGE_STAGE_RUNNING\"; while [ ! -e \"$MERGE_STAGE_RELEASE\" ]; do sleep 0.01; done"]}]}`,
	} {
		procedureWrite(t, server, path, content)
	}
	git(t, server, "add", ".")
	git(t, server, "commit", "-qm", "command lock fixture")
	git(t, server, "push", "-q", "origin", "main")
	git(t, root, "remote", "add", "origin", gh.remote)
	git(t, root, "fetch", "upstream", "main")
	git(t, root, "merge", "--no-edit", "upstream/main")
	git(t, root, "push", "upstream", landingTestBranch)
	gh.pr.HeadOID = procedureHead(t, root)
	gh.ci.setHead(gh.pr.Number, gh.pr.HeadOID)
	if _, _, err := executeSDLCTestCommand("stage", ":0"); err != nil {
		t.Fatal(err)
	}
	return f
}

func mergeStageAwait(t *testing.T, name string, ready func() bool) {
	t.Helper()
	deadline, hasDeadline := t.Deadline()
	budget := landingCompletionBudget(deadline, hasDeadline, time.Now())
	timer := time.NewTimer(budget)
	defer timer.Stop()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for !ready() {
		select {
		case <-tick.C:
		case <-timer.C:
			var stacks bytes.Buffer
			_ = pprof.Lookup("goroutine").WriteTo(&stacks, 2)
			t.Fatalf("waiting for %s exhausted %s; goroutine stacks:\n%s", name, budget, stacks.String())
		}
	}
}

func mergeStageTouch(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
}

func mergeStageFile(t *testing.T, path string) {
	t.Helper()
	mergeStageAwait(t, filepath.Base(path), func() bool { _, err := os.Stat(path); return err == nil })
}

func mergeStageJoined(t *testing.T, name string, w *blockedMergeWorker) error {
	t.Helper()
	mergeStageAwait(t, name, func() bool {
		select {
		case <-w.done:
			return true
		default:
			return false
		}
	})
	return w.err
}

func (f *mergeStageFixture) merge(t *testing.T) *blockedMergeWorker {
	t.Helper()
	w := startBlockedMergeWorker(t, f.release, func(ctx context.Context) error {
		_, err := executeSDLCTestCommandWith(ctx, io.Discard, "merge", "--yes", "--no-judge", "--no-validate")
		return err
	})
	// Registered after the worker: admission is freed before its cancellation
	// and join, which precede the fixture's cwd/global adapter restoration.
	t.Cleanup(func() { _ = os.WriteFile(f.admit, nil, 0600) })
	return w
}

func (f *mergeStageFixture) stage(t *testing.T, action string, stderr io.Writer) *blockedMergeWorker {
	t.Helper()
	args := []string{"stage", ":0", "--build"}
	if action == "clear" {
		args = []string{"stage", "--clear", ":0"}
	}
	return startBlockedMergeWorker(t, f.release, func(ctx context.Context) error {
		_, err := executeSDLCTestCommandWith(ctx, stderr, args...)
		return err
	})
}

func (f *mergeStageFixture) assertStage(t *testing.T, action, source string) {
	t.Helper()
	want, branch := source, ""
	if action == "clear" {
		want, branch = git(t, f.root, "rev-parse", "origin/main"), "main"
	}
	if got := procedureHead(t, f.target); got != want {
		t.Fatalf("target HEAD=%s want %s", got, want)
	}
	if got := git(t, f.target, "branch", "--show-current"); got != branch {
		t.Fatalf("target branch=%q want %q", got, branch)
	}
	lease, err := fleet.ReadStageLease(execGitRunner{}, f.target)
	if err != nil {
		t.Fatal(err)
	}
	if action == "clear" && lease != (fleet.StageLease{}) {
		t.Fatalf("clear retained lease %+v", lease)
	}
	if action == "build" && (lease.Commit != want || lease.Holder != "#1") {
		t.Fatalf("build lease %+v", lease)
	}
	stamp, err := os.ReadFile(f.built)
	if err != nil || strings.TrimSpace(string(stamp)) != want {
		t.Fatalf("build stamp=%q want %s err=%v", stamp, want, err)
	}
	if got := procedureHead(t, f.root); got != source {
		t.Fatalf("stage changed source HEAD %s -> %s", source, got)
	}
	if got := git(t, f.root, "branch", "--show-current"); got != landingTestBranch {
		t.Fatalf("stage changed source branch to %s", got)
	}
}

func TestMergeCommandStageDuringSuiteWaits(t *testing.T) {
	for _, phase := range []string{"admission", "execution"} {
		for _, action := range []string{"clear", "build"} {
			t.Run(phase+"/"+action, func(t *testing.T) {
				f := newMergeStageFixture(t)
				if phase == "execution" {
					mergeStageTouch(t, f.admit)
				}
				merge := f.merge(t)
				barrier := f.admission
				if phase == "execution" {
					barrier = f.running
				}
				mergeStageFile(t, barrier)
				source := procedureHead(t, f.root)
				stage := f.stage(t, action, io.Discard)
				if err := mergeStageJoined(t, "stage "+action+" during "+phase, stage); err != nil {
					t.Fatal(err)
				}
				f.assertStage(t, action, source)
				select {
				case <-merge.done:
					t.Fatal("merge ended before its barrier was released")
				default:
				}
				mergeStageTouch(t, f.admit)
				mergeStageTouch(t, f.release)
				if err := mergeStageJoined(t, "merge after suite release", merge); err != nil {
					t.Fatal(err)
				}
				if f.gh.merges != 1 {
					t.Fatalf("GitHub merges=%d", f.gh.merges)
				}
			})
		}
	}
}

func TestMergeCommandStageWaitsForTransaction(t *testing.T) {
	f := newMergeStageFixture(t)
	integrationAdvance(t, f.server, "post-fixture.txt", "main advanced before the merge command\n")
	before := slotState(t, f.target)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	held, release := make(chan struct{}), make(chan struct{})
	previousRunner := mergeRunner
	mergeRunner = landingHookRunner{gitRunner: previousRunner, after: func(root string, args []string) error {
		if root != f.root || len(args) < 2 || args[0] != "push" || !strings.HasPrefix(args[1], "--force-with-lease=refs/heads/"+landingTestBranch+":") {
			return nil
		}
		// integrateLandingMain holds its real repository transaction across
		// this lease push. Admission stays held after the push is released.
		close(held)
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}}
	t.Cleanup(func() { mergeRunner = previousRunner })
	merge := f.merge(t)
	t.Cleanup(cancel)
	mergeStageAwait(t, "integration push transaction held", func() bool {
		select {
		case <-held:
			return true
		default:
			return false
		}
	})
	source := procedureHead(t, f.root)
	watch := testfix.NewWatch("waiting for sdlc repo lock held by")
	stage := f.stage(t, "clear", watch)
	mergeStageAwait(t, "stage transaction wait diagnostic", func() bool {
		select {
		case <-watch.Seen():
			return true
		default:
			return false
		}
	})
	if got := slotState(t, f.target); got != before {
		t.Fatal("stage mutated target under held transaction")
	}
	if !strings.Contains(watch.String(), "sdlc merge") {
		t.Fatalf("wait diagnostic omitted actual merge holder: %s", watch.String())
	}
	close(release)
	if err := mergeStageJoined(t, "stage after transaction", stage); err != nil {
		t.Fatal(err)
	}
	f.assertStage(t, "clear", source)
	mergeStageFile(t, f.admission)
	mergeStageTouch(t, f.admit)
	mergeStageTouch(t, f.release)
	if err := mergeStageJoined(t, "merge after transaction and suite", merge); err != nil {
		t.Fatal(err)
	}
	if f.gh.merges != 1 {
		t.Fatalf("GitHub merges=%d", f.gh.merges)
	}
}

func TestMergeCommandRefusesHeadDriftAfterSuite(t *testing.T) {
	f := newMergeStageFixture(t)
	mergeStageTouch(t, f.admit)
	merge := f.merge(t)
	mergeStageFile(t, f.running)
	git(t, f.root, "commit", "--allow-empty", "-qm", "concurrent source work")
	want := procedureHead(t, f.root)
	mergeStageTouch(t, f.release)
	err := mergeStageJoined(t, "merge refusal after source drift", merge)
	if err == nil || !strings.Contains(err.Error(), "issue branch changed") {
		t.Fatalf("head drift refusal=%v", err)
	}
	if f.gh.merges != 0 || procedureHead(t, f.root) != want {
		t.Fatalf("source drift lost or merged: merges=%d HEAD=%s want=%s", f.gh.merges, procedureHead(t, f.root), want)
	}
}
