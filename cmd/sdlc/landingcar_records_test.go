package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// A merge that started before a correction must see the new close without
// restarting. Use real close/publication logic and a bare tracker remote; only
// GitHub and the reviewer are local fakes.
func TestCarReadinessObservesExternalReclose(t *testing.T) {
	r, _, _ := closeReady(t, 499)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n"+emptyBoundaryFindings)
	closeIssue := func() {
		t.Helper()
		if _, stderr, err := executeSDLCTestCommand("close", "--issue", "499", "--verified", "regression", "--actual", "1", "--no-atlas"); err != nil {
			t.Fatalf("close: %v\n%s", err, stderr)
		}
	}
	closeIssue()
	writeRepoFile(t, r.root, "cmd/a.go", "package a\n\nconst Fixed = true\n")
	r.git("add", "cmd/a.go")
	r.git("commit", "-qm", "#499: correction after old close")
	branch := r.git("branch", "--show-current")
	r.git("push", "-q", "origin", branch)
	pr := landingPR{Number: 499, State: "OPEN", Repo: "test/repo", HeadRef: branch,
		HeadOID: r.git("rev-parse", "HEAD"), BaseRef: "main", BaseOID: r.originMain()}
	gh := &landingFakeGH{t: t, remote: r.origin, pr: pr, followPush: true}
	target := landingTarget{Root: r.root, Remote: "origin", Repo: "test/repo"}
	gates := productionCarGates(execGitRunner{}, gh, target, &mergeFlags{IssuesDir: "workshop/issues", NoCI: true})
	ctx := withIssueRecordsScope(context.Background())
	before, _, _ := carReadiness(ctx, execGitRunner{}, gh, target, branch, gates, time.Now())
	if before.Ready || !strings.Contains(before.Reason, "publish gate:") || !strings.Contains(before.Reason, "cmd/a.go") {
		t.Fatalf("expected uncovered correction: %+v", before)
	}
	// This separate CLI command owns a different context and cannot invalidate
	// the waiting command's records cache.
	closeIssue()
	r.git("push", "-q", "origin", branch)
	after, _, _ := carReadiness(ctx, execGitRunner{}, gh, target, branch, gates, time.Now())
	if !after.Ready || after.Head != r.git("rev-parse", "HEAD") {
		t.Fatalf("corrected close remained unready: %+v", after)
	}
}

// The readiness boundary supplies every gate one coherent snapshot, isolated
// from its command's snapshot. A remote edit during a pass appears next pass.
func TestCarReadinessRecordsGeneration(t *testing.T) {
	for _, reads := range []int{8, 16} {
		t.Run(fmt.Sprint(reads), func(t *testing.T) {
			cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
			r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
			r.git("switch", "-qc", "000009-nine")
			branch := r.git("branch", "--show-current")
			r.git("push", "-q", "origin", branch)
			head := r.git("rev-parse", "HEAD")
			gh := &landingFakeGH{pr: landingPR{Number: 9, State: "OPEN", HeadOID: head}}
			target := landingTarget{Root: r.root, Remote: "origin", Repo: "test/repo"}
			dir := filepath.Join(r.root, "workshop/issues")
			ctx := withIssueRecordsScope(context.Background())
			load := func(ctx context.Context) tracker.Records {
				t.Helper()
				rs, err := loadIssueRecords(ctx, dir, tracker.Fresh)
				if err != nil {
					t.Fatal(err)
				}
				return rs
			}
			load(ctx) // prime parent before observing external changes
			external := filepath.Join(t.TempDir(), "external")
			testfix.Git(t, "", "clone", "-q", r.origin, external)
			writer, err := gitx.NewTrunkFileContext(context.Background(), external, "origin", "issue-tracker")
			if err != nil {
				t.Fatal(err)
			}
			trace := filepath.Join(t.TempDir(), "git.trace")
			t.Setenv("GIT_TRACE", trace)
			var generation string
			gates := carGates{publish: func(pass context.Context, _ landingPR) error {
				rs := load(pass)
				generation = rs.Ref
				// Make the remote unavailable for all remaining reads this pass:
				// a per-issue refresh would fail, rather than reuse one generation.
				r.git("remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))
				return nil
			}, ci: func(pass context.Context, _ landingPR) error {
				for i := 0; i < reads; i++ {
					if rs := load(pass); rs.Ref != generation {
						t.Fatal("mixed generations")
					}
				}
				return nil
			}}
			ready, _, _ := carReadiness(ctx, execGitRunner{}, gh, target, branch, gates, time.Now())
			if !ready.Ready {
				t.Fatalf("first pass: %+v", ready)
			}
			r.git("remote", "set-url", "origin", r.origin)
			// Restore readiness's transport, then change status in another checkout.
			if err := writer.UpdateMany("external status change", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
				return gitx.TrunkWrite{Write: map[string][]byte{cardPath: []byte(strings.Replace(card, "status: open", "status: working", 1))}, ExactBytes: true}, nil
			}); err != nil {
				t.Fatal(err)
			}
			// Count only readiness's own fetches, excluding the external writer.
			if err := os.WriteFile(trace, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			gates.publish = func(pass context.Context, _ landingPR) error {
				for i := 0; i < reads; i++ {
					rec, ok, err := load(pass).Require("000009")
					if err != nil || !ok || rec.Status() != "working" {
						t.Fatal("external status change remained cached")
					}
				}
				return fmt.Errorf("external status change")
			}
			ready, _, _ = carReadiness(ctx, execGitRunner{}, gh, target, branch, gates, time.Now())
			if ready.Ready || !strings.Contains(ready.Reason, "external status change") {
				t.Fatalf("second pass: %+v", ready)
			}
			log, err := os.ReadFile(trace)
			if err != nil {
				t.Fatal(err)
			}
			fetches := 0
			for _, line := range strings.Split(string(log), "\n") {
				if strings.Contains(line, "built-in: git fetch ") && strings.Contains(line, "issue-tracker") {
					fetches++
				}
			}
			if fetches != 1 {
				t.Fatalf("%d reads: got %d tracker fetches\n%s", reads, fetches, log)
			}
			rec, _, err := load(ctx).Require("000009")
			if err != nil || rec.Status() != "open" {
				t.Fatal("readiness changed parent scope")
			}
		})
	}
}

// BR-1: preserve BOTH production publish checks. A raw snapshot read in the
// transferred-details guard followed by a scoped selection can mix generations.
func TestCarReadinessPublishChecksShareGeneration(t *testing.T) {
	r, cardPath, _ := closeReady(t, 498)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n"+emptyBoundaryFindings)
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "498", "--verified", "coherence regression", "--actual", "1", "--no-atlas"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	branch, head := r.git("branch", "--show-current"), r.git("rev-parse", "HEAD")
	r.git("push", "-q", "origin", branch)
	external := filepath.Join(t.TempDir(), "external")
	testfix.Git(t, "", "clone", "-q", r.origin, external)
	writer, err := gitx.NewTrunkFileContext(context.Background(), external, "origin", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	before := strings.Fields(r.git("ls-remote", "origin", "refs/heads/issue-tracker"))[0]
	trace := filepath.Join(t.TempDir(), "production-git.trace")
	t.Setenv("GIT_TRACE", trace)
	oldGuard := guardTransferredDetailsFn
	t.Cleanup(func() { guardTransferredDetailsFn = oldGuard })
	changed := false
	guardTransferredDetailsFn = func(ctx context.Context) error {
		if err := oldGuard(ctx); err != nil {
			return err
		}
		if !changed {
			changed = true
			// Exclude the independently owned writer's network traffic.
			t.Setenv("GIT_TRACE", "")
			defer t.Setenv("GIT_TRACE", trace)
			// A newer malformed card must become visible next pass, never
			// halfway through the old pass's production authority checks.
			return writer.UpdateMany("external malformed edit between checks", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
				return gitx.TrunkWrite{Write: map[string][]byte{cardPath: []byte("---\nid: 000498\nstatus: [broken\n---\n")}, ExactBytes: true}, nil
			})
		}
		return nil
	}
	pr := landingPR{Number: 498, State: "OPEN", Repo: "test/repo", HeadRef: branch,
		HeadOID: head, BaseRef: "main", BaseOID: r.originMain()}
	gh := &landingFakeGH{pr: pr}
	target := landingTarget{Root: r.root, Remote: "origin", Repo: "test/repo"}
	gates := productionCarGates(execGitRunner{}, gh, target, &mergeFlags{IssuesDir: "workshop/issues", NoCI: true})
	gates.ci = func(ctx context.Context, _ landingPR) error {
		for _, reads := range []int{8, 16} {
			for i := 0; i < reads; i++ {
				for _, dir := range []string{"workshop/issues", "custom/issues"} {
					rs, err := loadIssueRecords(ctx, dir, tracker.Fresh)
					if err != nil {
						return err
					}
					if rs.Ref != before {
						return fmt.Errorf("mixed tracker generations in %s: before %s, after %s", dir, before, rs.Ref)
					}
				}
			}
			log, err := os.ReadFile(trace)
			if err != nil {
				return err
			}
			probes, fetches := trackerTraceCounts(string(log))
			if probes != 1 || fetches > 1 {
				return fmt.Errorf("%d repeated production reads: %d tracker probes, %d fetches; want one probe and at most one fetch", reads, probes, fetches)
			}
		}
		return nil
	}
	ctx := withIssueRecordsScope(context.Background())
	ready, _, _ := carReadiness(ctx, execGitRunner{}, gh, target, branch, gates, time.Now())
	if !changed || !ready.Ready {
		t.Fatalf("production checks did not retain their initial generation: %+v (external edit=%v)", ready, changed)
	}
	ready, _, _ = carReadiness(ctx, execGitRunner{}, gh, target, branch, gates, time.Now())
	if ready.Ready || !strings.Contains(ready.Reason, "publish gate:") {
		t.Fatalf("next pass accepted malformed external authority: %+v", ready)
	}
}

func trackerTraceCounts(trace string) (probes, fetches int) {
	for _, line := range strings.Split(trace, "\n") {
		if !strings.Contains(line, "issue-tracker") {
			continue
		}
		if strings.Contains(line, "built-in: git ls-remote ") {
			probes++
		}
		if strings.Contains(line, "built-in: git fetch ") {
			fetches++
		}
	}
	return
}

// #502: a close carried by a waiting car remains an obligation when an
// independent tracker writer reopens it. Readiness refresh cannot erase it.
func TestCarReadinessRefusesExternallyReopenedClose(t *testing.T) {
	for _, mode := range []string{"original", "rebased", "published-details"} {
		t.Run(mode, func(t *testing.T) { testCarReadinessReopen(t, mode) })
	}
}

func testCarReadinessReopen(t *testing.T, mode string) {
	r, cardPath, detailPath := closeReady(t, 502)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n"+emptyBoundaryFindings)
	closeIssue := func() {
		t.Helper()
		if _, stderr, err := executeSDLCTestCommand("close", "--issue", "502", "--verified", "reopen regression", "--actual", "1", "--no-atlas"); err != nil {
			t.Fatalf("close: %v\n%s", err, stderr)
		}
	}
	closeIssue()
	switch mode {
	case "rebased":
		peerCommit(t, r, "peer.go")
		r.git("fetch", "-q", "origin")
		r.git("rebase", "-q", "origin/main")
	case "published-details":
		main, err := gitx.NewTrunkFileContext(context.Background(), r.root, "origin", "main")
		if err != nil {
			t.Fatal(err)
		}
		body := []byte(r.git("show", "HEAD:"+detailPath) + "\n")
		if err := main.UpdateMany("independent details publish", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
			return gitx.TrunkWrite{Write: map[string][]byte{detailPath: body}, ExactBytes: true}, nil
		}); err != nil {
			t.Fatal(err)
		}
		r.git("fetch", "-q", "origin")
	}
	branch := r.git("branch", "--show-current")
	r.git("push", "-q", "--force", "origin", branch)
	pr := landingPR{Number: 502, State: "OPEN", Repo: "test/repo", HeadRef: branch, HeadOID: r.git("rev-parse", "HEAD"), BaseRef: "main", BaseOID: r.originMain()}
	gh := &landingFakeGH{t: t, remote: r.origin, pr: pr, followPush: true}
	target := landingTarget{Root: r.root, Remote: "origin", Repo: "test/repo"}
	gates := productionCarGates(execGitRunner{}, gh, target, &mergeFlags{IssuesDir: "workshop/issues", NoCI: true})
	ctx := withIssueRecordsScope(context.Background())
	check := func(wantReady bool) {
		t.Helper()
		before := r.git("rev-parse", "HEAD")
		ready, _, _ := carReadiness(ctx, execGitRunner{}, gh, target, branch, gates, time.Now())
		if ready.Ready != wantReady {
			t.Fatalf("ready=%t, want %t: %+v", ready.Ready, wantReady, ready)
		}
		if !wantReady {
			for _, phrase := range []string{"#502", "working", "sdlc close --issue 502 --verified"} {
				if !strings.Contains(ready.Reason, phrase) {
					t.Fatalf("missing recovery %q: %+v", phrase, ready)
				}
			}
		}
		if after := r.git("rev-parse", "HEAD"); after != before {
			t.Fatal("readiness mutated branch")
		}
	}
	check(true)
	external := filepath.Join(t.TempDir(), "external")
	testfix.Git(t, "", "clone", "-q", r.origin, external)
	writer, err := gitx.NewTrunkFileContext(context.Background(), external, "origin", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	if err := writer.UpdateMany("external reopen", func(v *gitx.TrunkView) (gitx.TrunkWrite, error) {
		raw, err := v.Read(cardPath)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		next, err := issue.SetCardField(raw, "status", "working")
		return gitx.TrunkWrite{Write: map[string][]byte{cardPath: next}, ExactBytes: true}, err
	}); err != nil {
		t.Fatal(err)
	}
	reopened := r.card(cardPath)
	check(false)
	env, err := openTracker(ctx)
	if err != nil {
		t.Fatal(err)
	}
	rs, err := loadIssueRecords(withIssueRecordsScope(ctx), filepath.Join(r.root, "workshop/issues"), tracker.Fresh)
	if err != nil {
		t.Fatal(err)
	}
	if owned, err := ownedCompletions(env, rs, r.git("rev-parse", "HEAD"), "", false); err != nil || len(owned) != 0 {
		t.Fatalf("settlement selector acquired reopened work: %+v %v", owned, err)
	}

	if r.card(cardPath) != reopened {
		t.Fatal("readiness mutated reopened card")
	}
	err = runPublishGate(withIssueRecordsScope(context.Background()), "origin/main", "workshop/issues", io.Discard)
	if err == nil || !strings.Contains(err.Error(), "sdlc close --issue 502 --verified") {
		t.Fatalf("ordinary publish accepted reopen or lacks action: %v", err)
	}
	if mode == "published-details" {
		// Re-close updates the published details. Base that edit on their latest
		// main version, as the independent transfer guard requires.
		r.git("merge", "-q", "--no-edit", "origin/main")
	}
	closeIssue()
	r.git("push", "-q", "origin", branch)
	check(true)
}
