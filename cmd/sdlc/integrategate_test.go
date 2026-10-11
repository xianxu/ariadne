package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// integrateFixture is a working repository (cwd) with a bare origin, and a second
// clone that lands commits on origin's main. The gate's clock is pinned two days
// ahead, so everything landed in the test is before the cutoff.
type integrateFixture struct {
	t                *testing.T
	work, peer, bare string
}

func newIntegrateFixture(t *testing.T) *integrateFixture {
	t.Helper()
	fx := &integrateFixture{t: t, bare: t.TempDir()}
	testfix.Git(t, fx.bare, "init", "-q", "--bare", "-b", "main")
	fx.work = testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	testfix.Git(t, fx.work, "remote", "add", "origin", fx.bare)
	testfix.Git(t, fx.work, "push", "-q", "origin", "main")
	testfix.Git(t, fx.work, "fetch", "-q", "origin")
	fx.peer = filepath.Join(t.TempDir(), "peer")
	testfix.Git(t, "", "clone", "-q", fx.bare, fx.peer)
	for _, kv := range [][2]string{{"user.email", "t@t"}, {"user.name", "t"}, {"commit.gpgsign", "false"}} {
		testfix.Git(t, fx.peer, "config", kv[0], kv[1])
	}
	testfix.Git(t, fx.work, "checkout", "-q", "-b", "feature")
	fx.commit(fx.work, "feature.go")

	prev := integrateNow
	integrateNow = func() time.Time { return time.Now().Add(48 * time.Hour) }
	t.Cleanup(func() { integrateNow = prev })
	return fx
}

func (fx *integrateFixture) commit(dir, path string) {
	fx.t.Helper()
	full := filepath.Join(dir, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		fx.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(path+"\n"), 0o644); err != nil {
		fx.t.Fatal(err)
	}
	testfix.Git(fx.t, dir, "add", path)
	testfix.Git(fx.t, dir, "commit", "-q", "-m", "add "+path)
}

// land commits path on origin's main through the peer clone.
func (fx *integrateFixture) land(path string) {
	fx.t.Helper()
	fx.commit(fx.peer, path)
	testfix.Git(fx.t, fx.peer, "push", "-q", "origin", "main")
}

func runIntegrateGate(t *testing.T, f *closeFlags) (out, refusal string, refused bool) {
	t.Helper()
	var stderr bytes.Buffer
	refusal, refused = expectDie(t, func() { integrateGate(context.Background(), &stderr, f, "269") })
	return stderr.String(), refusal, refused
}

func TestIntegrateGate_FreshPassesSilently(t *testing.T) {
	newIntegrateFixture(t)
	if out, msg, refused := runIntegrateGate(t, &closeFlags{}); refused || out != "" {
		t.Fatalf("a branch with nothing new on main should pass silently: refused=%v %q %q", refused, msg, out)
	}
}

// The unpushed branch is told to rebase, and the named remedy, run from the refused
// state, makes the gate pass.
func TestIntegrateGate_UnpushedStaleRefusesWithRebase(t *testing.T) {
	fx := newIntegrateFixture(t)
	fx.land("pkg/new.go")
	fx.land("notes.md")
	_, msg, refused := runIntegrateGate(t, &closeFlags{})
	if !refused {
		t.Fatal("main gained code before the cutoff: want a refusal")
	}
	for _, want := range []string{"pkg/new.go", "git rebase origin/main", "pass --no-integrate (or --force) with the reason in --verified", "sdlc close --issue 269"} {
		if !strings.Contains(msg, want) {
			t.Errorf("refusal lacks %q:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, "notes.md") || strings.Contains(msg, "git merge") {
		t.Errorf("refusal lists docs or the pushed remedy:\n%s", msg)
	}
	testfix.Git(t, fx.work, "rebase", "-q", "origin/main")
	if out, msg, refused := runIntegrateGate(t, &closeFlags{}); refused || out != "" {
		t.Fatalf("after the rebase: refused=%v %q %q", refused, msg, out)
	}
}

func TestIntegrateGate_PushedStaleRefusesWithMerge(t *testing.T) {
	fx := newIntegrateFixture(t)
	testfix.Git(t, fx.work, "push", "-q", "origin", "feature")
	fx.land("pkg/new.go")
	_, msg, refused := runIntegrateGate(t, &closeFlags{Milestone: "M2"})
	if !refused || !strings.Contains(msg, "git merge origin/main") || strings.Contains(msg, "git rebase") {
		t.Fatalf("pushed branch: want the merge remedy; refused=%v\n%s", refused, msg)
	}
	if !strings.Contains(msg, "sdlc milestone-close --issue 269 --milestone M2") {
		t.Errorf("a milestone refusal re-runs milestone-close:\n%s", msg)
	}
	testfix.Git(t, fx.work, "merge", "-q", "--no-edit", "origin/main")
	if _, msg, refused := runIntegrateGate(t, &closeFlags{}); refused {
		t.Fatalf("after the merge: %s", msg)
	}
}

func TestIntegrateGate_DocsOnlyPassesWithInfo(t *testing.T) {
	fx := newIntegrateFixture(t)
	fx.land("atlas/map.md")
	out, msg, refused := runIntegrateGate(t, &closeFlags{})
	if refused || !strings.Contains(out, "1 doc-only file(s)") {
		t.Fatalf("docs-only: refused=%v %q\n%s", refused, msg, out)
	}
}

// Embedded markdown ships in the binary, so it is code (the publish gate's predicate).
func TestIntegrateGate_EmbeddedMarkdownIsCode(t *testing.T) {
	fx := newIntegrateFixture(t)
	fx.land("cmd/sdlc/helptext/close.md")
	if _, msg, refused := runIntegrateGate(t, &closeFlags{}); !refused || !strings.Contains(msg, "cmd/sdlc/helptext/close.md") {
		t.Fatalf("embedded helptext must refuse: refused=%v\n%s", refused, msg)
	}
}

// A main that cannot be fetched is never reported as up to date.
func TestIntegrateGate_FetchFailureIsInconclusive(t *testing.T) {
	fx := newIntegrateFixture(t)
	testfix.Git(t, fx.work, "remote", "set-url", "origin", filepath.Join(fx.bare, "missing"))
	out, msg, refused := runIntegrateGate(t, &closeFlags{})
	if refused || !strings.Contains(out, "could not check") || !strings.Contains(out, "NOT verified as up to date") {
		t.Fatalf("fetch failure: refused=%v %q\n%s", refused, msg, out)
	}
}

func TestIntegrateGate_Bypass(t *testing.T) {
	fx := newIntegrateFixture(t)
	fx.land("pkg/new.go")
	for _, f := range []*closeFlags{{NoIntegrate: true}, {Force: true}} {
		out, msg, refused := runIntegrateGate(t, f)
		if refused || !strings.Contains(out, "--no-integrate (or --force): skipping the integrate gate") {
			t.Fatalf("%+v: refused=%v %q\n%s", f, refused, msg, out)
		}
	}
}

func TestIntegrateGate_BadTimezoneConfigRefuses(t *testing.T) {
	fx := newIntegrateFixture(t)
	if err := os.MkdirAll(filepath.Join(fx.work, ".sdlc"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fx.work, ".sdlc/freshness.json"), []byte(`{"version":1,"timezone":"Mars/Olympus"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, msg, refused := runIntegrateGate(t, &closeFlags{}); !refused || !strings.Contains(msg, ".sdlc/freshness.json") {
		t.Fatalf("an unknown zone must refuse naming the file: refused=%v\n%s", refused, msg)
	}
}

// computeClose runs the gate before planning the window: a stale branch refuses the
// close itself, not just the helper.
func TestCloseRefusesAStaleBranchBeforeTheReview(t *testing.T) {
	issuesDir := quickCloseRepo(t, 269, "", map[string]string{"x.go": goLines(3)})
	work, _ := os.Getwd()
	bare := t.TempDir()
	testfix.Git(t, bare, "init", "-q", "--bare", "-b", "main")
	testfix.Git(t, work, "remote", "add", "origin", bare)
	testfix.Git(t, work, "push", "-q", "origin", "main")
	testfix.Git(t, work, "checkout", "-q", "-b", "000269-x")
	peer := filepath.Join(t.TempDir(), "peer")
	testfix.Git(t, "", "clone", "-q", bare, peer)
	if err := os.WriteFile(filepath.Join(peer, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, peer, "add", "main.go")
	testfix.Git(t, peer, "-c", "user.email=t@t", "-c", "user.name=t", "-c", "commit.gpgsign=false", "commit", "-q", "-m", "land main.go")
	testfix.Git(t, peer, "push", "-q", "origin", "main")
	prev := integrateNow
	integrateNow = func() time.Time { return time.Now().Add(48 * time.Hour) }
	t.Cleanup(func() { integrateNow = prev })

	f := quickFlags(issuesDir, 269)
	f.DryRun = true
	msg, refused := expectDie(t, func() { computeClose(&bytes.Buffer{}, f) })
	if !refused || !strings.Contains(msg, "integrate gate") || !strings.Contains(msg, "main.go") {
		t.Fatalf("close on a stale branch: refused=%v\n%s", refused, msg)
	}
}

// In a tracker repository the gate reads the review window's fetched main (a pinned
// commit) but names the tracking ref in the remedy, and an unreachable origin is
// inconclusive.
func TestIntegrateGate_TrackerRepository(t *testing.T) {
	r := newTrackerRepo(t, nil, nil)
	testfix.Git(t, r.root, "checkout", "-q", "-b", "feature")
	peer := filepath.Join(t.TempDir(), "peer")
	testfix.Git(t, "", "clone", "-q", r.origin, peer)
	if err := os.WriteFile(filepath.Join(peer, "new.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, peer, "add", "new.go")
	testfix.Git(t, peer, "-c", "user.email=t@t", "-c", "user.name=t", "-c", "commit.gpgsign=false", "commit", "-q", "-m", "land new.go")
	testfix.Git(t, peer, "push", "-q", "origin", "main")
	prev := integrateNow
	integrateNow = func() time.Time { return time.Now().Add(48 * time.Hour) }
	t.Cleanup(func() { integrateNow = prev })

	_, msg, refused := runIntegrateGate(t, &closeFlags{})
	if !refused || !strings.Contains(msg, "new.go") || !strings.Contains(msg, "git rebase origin/main") {
		t.Fatalf("tracker repo, stale branch: refused=%v\n%s", refused, msg)
	}

	if err := os.Rename(r.origin, r.origin+".gone"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Rename(r.origin+".gone", r.origin) })
	// A new process: main is fetched once per process and checkout.
	reviewMainMu.Lock()
	reviewMainCache = map[string]reviewMainEntry{}
	reviewMainMu.Unlock()
	out, msg, refused := runIntegrateGate(t, &closeFlags{})
	if refused || !strings.Contains(out, "could not check") {
		t.Fatalf("unreachable tracker origin: refused=%v %q\n%s", refused, msg, out)
	}
}

// A branch whose upstream is main itself (`-b x origin/main`) was never pushed, so
// it is told to rebase.
func TestIntegrateGate_UpstreamOfMainIsNotPushed(t *testing.T) {
	fx := newIntegrateFixture(t)
	testfix.Git(t, fx.work, "branch", "-q", "--set-upstream-to=origin/main")
	fx.land("pkg/new.go")
	if _, msg, refused := runIntegrateGate(t, &closeFlags{}); !refused || !strings.Contains(msg, "git rebase origin/main") {
		t.Fatalf("upstream = main: want the rebase remedy; refused=%v\n%s", refused, msg)
	}
}

// A fetch that outlives its deadline is inconclusive, not a hang.
func TestIntegrateGate_FetchTimeoutIsInconclusive(t *testing.T) {
	fx := newIntegrateFixture(t)
	hang := filepath.Join(t.TempDir(), "hang-upload-pack")
	if err := os.WriteFile(hang, []byte("#!/bin/sh\nsleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, fx.work, "config", "remote.origin.uploadpack", hang)
	prev := integrateFetchTimeout
	integrateFetchTimeout = 500 * time.Millisecond
	t.Cleanup(func() { integrateFetchTimeout = prev })
	start := time.Now()
	out, msg, refused := runIntegrateGate(t, &closeFlags{})
	if refused || !strings.Contains(out, "could not check") {
		t.Fatalf("hung fetch: refused=%v %q\n%s", refused, msg, out)
	}
	if took := time.Since(start); took > 20*time.Second {
		t.Fatalf("the fetch deadline did not bound the gate: %s", took)
	}
}
