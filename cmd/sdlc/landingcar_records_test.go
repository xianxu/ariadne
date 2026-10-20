package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
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
	before, _ := carReadiness(ctx, execGitRunner{}, gh, target, branch, gates, time.Now())
	if before.Ready || !strings.Contains(before.Reason, "publish gate:") || !strings.Contains(before.Reason, "cmd/a.go") {
		t.Fatalf("expected uncovered correction: %+v", before)
	}
	// This separate CLI command owns a different context and cannot invalidate
	// the waiting command's records cache.
	closeIssue()
	r.git("push", "-q", "origin", branch)
	after, _ := carReadiness(ctx, execGitRunner{}, gh, target, branch, gates, time.Now())
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
			ready, _ := carReadiness(ctx, execGitRunner{}, gh, target, branch, gates, time.Now())
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
			ready, _ = carReadiness(ctx, execGitRunner{}, gh, target, branch, gates, time.Now())
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
