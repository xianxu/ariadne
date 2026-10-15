package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Real bare origin, real merges and real suite subprocesses; the GH fake
// observes pushes and records statuses on their exact commits.
func integrationFixture(t *testing.T) (root, server string, gh *landingFakeGH) {
	t.Helper()
	oldPoll := integrationPollInterval
	integrationPollInterval = time.Millisecond
	t.Cleanup(func() { integrationPollInterval = oldPoll })
	roots, remote, gh := landingFixture(t, 1)
	root = roots[1]
	server = filepath.Join(t.TempDir(), "server")
	git(t, "", "clone", "-q", remote, server)
	git(t, server, "config", "user.name", "fixture")
	git(t, server, "config", "user.email", "fixture@example.com")
	gh.followPush = true
	gh.ci = newCIFakeGH()
	gh.ci.setHead(gh.pr.Number, gh.pr.HeadOID)
	for path, content := range map[string]string{
		localChecksPath:     `{"version":1,"checks":[{"name":"suite","required":true,"command":["sh","-c","echo suite passed"]}]}`,
		"Makefile.workflow": "# fixture\n", "scripts/test-lock.py": fakeSuiteLockHelper("print('skipped')\n"),
	} {
		procedureWrite(t, server, path, content)
	}
	git(t, server, "add", ".")
	git(t, server, "commit", "-qm", "declare suite on main")
	git(t, server, "push", "-q", "origin", "main")
	return root, server, gh
}

func integrationAdvance(t *testing.T, server, path, content string) string {
	t.Helper()
	procedureWrite(t, server, path, content)
	git(t, server, "add", ".")
	git(t, server, "commit", "-qm", "main moves")
	git(t, server, "push", "-q", "origin", "main")
	return procedureHead(t, server)
}

func TestLandingIntegratesAMovedMain(t *testing.T) {
	root, server, gh := integrationFixture(t)
	before, main := procedureHead(t, root), procedureHead(t, server)
	gh.headLag = 1
	f := landingFlags()
	f.NoJudge = false
	oldGuard := guardTransferredDetailsFn
	guardTransferredDetailsFn = func(context.Context) error { return nil }
	t.Cleanup(func() { guardTransferredDetailsFn = oldGuard })
	if err := runMerge(io.Discard, io.Discard, f); err != nil {
		t.Fatal(err)
	}
	head := gh.pr.HeadOID
	if gh.merges != 1 || gh.ci.statuses[head]["sdlc-local/suite"] != "success" {
		t.Fatalf("merges=%d statuses=%v", gh.merges, gh.ci.statuses)
	}
	if got := git(t, root, "show", "-s", "--format=%P", head); got != before+" "+main {
		t.Fatalf("parents=%s", got)
	}
	if got := git(t, root, "show", "-s", "--format=%B", head); !strings.Contains(got, "Integrated-Main: "+main) || !strings.Contains(got, "\nAgent: "+testfix.FixedAgent) {
		t.Fatalf("missing trailer: %s", got)
	}
	if got := git(t, root, "branch", "--show-current"); got != "main-slot1" {
		t.Fatal(got)
	}
}

func TestLandingRegistrationPollingDoesNotHoldRepoLock(t *testing.T) {
	_, _, gh := integrationFixture(t)
	gh.headLag = 2
	cmd := markManualLockCommand(&cobra.Command{Use: "merge"})
	cmd.SetContext(context.Background())
	f := landingFlags()
	f.Lock = func(run func() error) error { return withRequiredRepoTransactionLock(cmd, run) }
	queries := 0
	gh.afterLaggedQuery = func() {
		queries++
		other := markManualLockCommand(&cobra.Command{Use: "claim"})
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		other.SetContext(ctx)
		if err := withRequiredRepoTransactionLock(other, func() error { return nil }); err != nil {
			t.Errorf("repo lock held during GitHub registration query: %v", err)
		}
	}
	if err := runMerge(io.Discard, io.Discard, f); err != nil {
		t.Fatal(err)
	}
	if queries < 2 || gh.merges != 1 {
		t.Fatalf("queries=%d merges=%d", queries, gh.merges)
	}
}

func TestLandingMainMovesDuringRegistrationBeforeFirstSuite(t *testing.T) {
	root, server, gh := integrationFixture(t)
	gh.headLag = 1
	var moved string
	gh.afterLaggedQuery = func() {
		if moved == "" {
			moved = integrationAdvance(t, server, "registration.go", "main moved during registration\n")
		}
	}
	var tested []string
	gh.afterStatus = func(sha string) { tested = append(tested, sha) }
	if err := runMerge(io.Discard, io.Discard, landingFlags()); err != nil {
		t.Fatal(err)
	}
	if moved == "" || len(tested) != 1 || gh.merges != 1 {
		t.Fatalf("moved=%q suite heads=%v merges=%d", moved, tested, gh.merges)
	}
	included, err := integrationAncestor(mergeRunner, root, moved, tested[0])
	if err != nil || !included {
		t.Fatalf("suite head %s did not contain moved main %s", tested[0], moved)
	}
}

func TestLandingReleasesRepoLockDuringSuite(t *testing.T) {
	_, server, gh := integrationFixture(t)
	signalDir := t.TempDir()
	started, release := filepath.Join(signalDir, "started"), filepath.Join(signalDir, "release")
	command := fmt.Sprintf("touch %q; i=0; while [ ! -e %q ] && [ \"$i\" -lt 600 ]; do sleep 0.05; i=$((i+1)); done; test -e %q", started, release, release)
	procedureWrite(t, server, localChecksPath, `{"version":1,"checks":[{"name":"suite","required":true,"command":["sh","-c",`+strconv.Quote(command)+`]}]}`)
	git(t, server, "add", localChecksPath)
	git(t, server, "commit", "-qm", "block the suite")
	git(t, server, "push", "-q", "origin", "main")

	cmd := markManualLockCommand(&cobra.Command{Use: "merge"})
	cmd.SetContext(context.Background())
	var locks atomic.Int32
	f := landingFlags()
	f.Lock = func(run func() error) error {
		locks.Add(1)
		return withRequiredRepoTransactionLock(cmd, run)
	}
	worker := startBlockedMergeWorker(t, release, func(ctx context.Context) error {
		f.Context = ctx
		cmd.SetContext(ctx)
		return runMerge(io.Discard, io.Discard, f)
	})
	deadline := time.Now().Add(30 * time.Second)
	for {
		if _, err := os.Stat(started); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("suite did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if locks.Load() == 0 {
		t.Fatal("merge never acquired the repo transaction lock")
	}

	other := markManualLockCommand(&cobra.Command{Use: "claim"})
	lockCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	other.SetContext(lockCtx)
	if err := withRequiredRepoTransactionLock(other, func() error { return nil }); err != nil {
		t.Fatalf("another command could not acquire the repo lock while suite ran: %v", err)
	}
	if err := os.WriteFile(release, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	select {
	case <-worker.done:
		if worker.err != nil || gh.merges != 1 {
			t.Fatalf("merge err=%v merges=%d", worker.err, gh.merges)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("merge did not finish after releasing the suite")
	}
}

func TestLandingMainMovesDuringReplayResetsUnpublishedMerge(t *testing.T) {
	root, server, gh := integrationFixture(t)
	before := procedureHead(t, root)
	moved := ""
	oldGuard := guardTransferredDetailsFn
	guardTransferredDetailsFn = func(context.Context) error {
		if moved == "" {
			moved = integrationAdvance(t, server, "during-replay.txt", "new main\n")
		}
		return nil
	}
	t.Cleanup(func() { guardTransferredDetailsFn = oldGuard })
	f := landingFlags()
	f.NoJudge = false
	if err := runMerge(io.Discard, io.Discard, f); err != nil {
		t.Fatal(err)
	}
	if moved == "" || gh.merges != 1 {
		t.Fatalf("replay did not advance main or land once: moved=%s merges=%d", moved, gh.merges)
	}
	log := git(t, gh.remote, "log", "--format=%B", before+".."+gh.pr.HeadOID)
	if got := strings.Count(log, "Integrated-Main:"); got != 1 {
		t.Fatalf("published %d integration commits after pre-push drift, want one", got)
	}
	if !strings.Contains(log, "Integrated-Main: "+moved) {
		t.Fatalf("published integration omitted latest main %s: %s", moved, log)
	}
}

func TestLandingReacquiresLockAndRefusesHeadDriftAfterSuite(t *testing.T) {
	root, _, gh := integrationFixture(t)
	cmd := markManualLockCommand(&cobra.Command{Use: "merge"})
	cmd.SetContext(context.Background())
	var locks atomic.Int32
	f := landingFlags()
	f.Lock = func(run func() error) error {
		locks.Add(1)
		return withRequiredRepoTransactionLock(cmd, run)
	}
	locksBeforeDrift := int32(-1)
	newHead := ""
	gh.afterStatus = func(string) {
		if newHead != "" {
			return
		}
		locksBeforeDrift = locks.Load()
		git(t, root, "commit", "--allow-empty", "-qm", "concurrent local work")
		newHead = procedureHead(t, root)
	}
	err := runMerge(io.Discard, io.Discard, f)
	if err == nil || !strings.Contains(err.Error(), "issue branch changed") {
		t.Fatalf("head drift was not refused: %v", err)
	}
	if locksBeforeDrift < 0 || locks.Load() <= locksBeforeDrift || gh.merges != 0 || procedureHead(t, root) != newHead {
		t.Fatalf("no locked revalidation after suite: before=%d after=%d merges=%d head=%s want=%s", locksBeforeDrift, locks.Load(), gh.merges, procedureHead(t, root), newHead)
	}
}

func TestLandingConflictRefusesAndLeavesTheBranch(t *testing.T) {
	root, server, gh := integrationFixture(t)
	integrationAdvance(t, server, "shipped.txt", "main's conflicting feature\n")
	before := procedureHead(t, root)
	var out bytes.Buffer
	err := runMerge(io.Discard, &out, landingFlags())
	if err == nil || !strings.Contains(err.Error(), "shipped.txt") || !strings.Contains(err.Error(), "conflict") {
		t.Fatalf("wrong refusal: %v", err)
	}
	if procedureHead(t, root) != before || git(t, gh.remote, "rev-parse", "refs/heads/"+landingTestBranch) != before || gh.merges != 0 {
		t.Fatal("conflict changed heads or landed")
	}
	if err := landingClean(mergeRunner, root); err != nil {
		t.Fatalf("merge not aborted: %v", err)
	}
}

func TestLandingMainMovesDuringSuite(t *testing.T) {
	for _, moves := range []int{1, 3} {
		t.Run(fmt.Sprint(moves), func(t *testing.T) {
			_, server, gh := integrationFixture(t)
			cmd := markManualLockCommand(&cobra.Command{Use: "merge"})
			cmd.SetContext(context.Background())
			f := landingFlags()
			f.Lock = func(run func() error) error { return withRequiredRepoTransactionLock(cmd, run) }
			var tested []string
			gh.afterStatus = func(sha string) {
				tested = append(tested, sha)
				other := markManualLockCommand(&cobra.Command{Use: "claim"})
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				other.SetContext(ctx)
				if err := withRequiredRepoTransactionLock(other, func() error { return nil }); err != nil {
					t.Errorf("repo lock unavailable after suite run %d: %v", len(tested), err)
				}
				if len(tested) <= moves {
					integrationAdvance(t, server, "moving.go", fmt.Sprintf("change %d\n", len(tested)))
				}
			}
			err := runMerge(io.Discard, io.Discard, f)
			if moves == 1 {
				if err != nil || gh.merges != 1 || len(tested) != 2 {
					t.Fatalf("err=%v merges=%d tested=%v", err, gh.merges, tested)
				}
			} else if err == nil || !strings.Contains(err.Error(), "moved too often") || gh.merges != 0 || len(tested) != 3 {
				t.Fatalf("err=%v merges=%d tested=%v", err, gh.merges, tested)
			}
			for _, sha := range tested {
				if gh.ci.statuses[sha]["sdlc-local/suite"] != "success" {
					t.Fatal("suite missing for " + sha)
				}
			}
		})
	}
}

func TestLandingSuiteFailureLabels(t *testing.T) {
	for _, tc := range []struct {
		name, command string
		want          []string
	}{
		{
			name:    "introduced-and-preexisting",
			command: `echo '--- FAIL: TestB'; if test -f shipped.txt; then echo '--- FAIL: TestA'; fi; exit 1`,
			want:    []string{"introduced: TestA", "pre-existing: TestB"},
		},
		{
			name:    "unparsed-main-failure",
			command: `echo 'suite infrastructure failed'; exit 1`,
			want:    []string{"main is red (unparsed)"},
		},
		{
			name:    "introduced-only",
			command: `if test -f shipped.txt; then echo '--- FAIL: TestA'; exit 1; fi`,
			want:    []string{"introduced: TestA", "main baseline passed"},
		},
		{
			name:    "merged-unparsed-failure",
			command: `if test -f shipped.txt; then echo 'suite infrastructure failed'; exit 1; fi`,
			want:    []string{"merged suite failed (unparsed)", "main baseline passed"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, server, gh := integrationFixture(t)
			procedureWrite(t, server, localChecksPath, `{"version":1,"checks":[{"name":"suite","required":true,"command":["sh","-c",`+strconv.Quote(tc.command)+`]}]}`)
			git(t, server, "add", localChecksPath)
			git(t, server, "commit", "-qm", "failing suite on main")
			git(t, server, "push", "-q", "origin", "main")
			main := procedureHead(t, server)
			err := runMerge(io.Discard, io.Discard, landingFlags())
			if err == nil {
				t.Fatal("failing suite allowed landing")
			}
			for _, want := range tc.want {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("missing %q in refusal: %v", want, err)
				}
			}
			if gh.merges != 0 || gh.ci.statuses[main] != nil {
				t.Fatalf("baseline landed or posted a status: merges=%d main statuses=%v", gh.merges, gh.ci.statuses[main])
			}
		})
	}
}

func TestLandingResumesAnInterruptedIntegration(t *testing.T) {
	for _, mode := range []string{"valid", "pushed-but-unregistered", "ordinary", "wrong-trailer", "wrong-first-parent", "modified-tree"} {
		t.Run(mode, func(t *testing.T) {
			root, server, gh := integrationFixture(t)
			main := procedureHead(t, server)
			git(t, root, "fetch", "-q", "upstream")
			before := procedureHead(t, root)
			if mode == "ordinary" {
				git(t, root, "commit", "--allow-empty", "-qm", "unpublished work")
			} else {
				trailer := main
				if mode == "wrong-trailer" {
					trailer = before
				}
				git(t, root, "merge", "--no-ff", "-m", "integration\n\nIntegrated-Main: "+trailer, main)
				if mode == "wrong-first-parent" {
					git(t, root, "commit", "--allow-empty", "-qm", "forged\n\nIntegrated-Main: "+main)
				}
				if mode == "modified-tree" {
					procedureWrite(t, root, "forged.go", "unreviewed\n")
					git(t, root, "add", ".")
					git(t, root, "commit", "--amend", "--no-edit")
				}
			}
			local := procedureHead(t, root)
			if mode == "pushed-but-unregistered" {
				git(t, root, "push", "-q", "upstream", landingTestBranch)
				gh.headLag = 1
			}
			err := runMerge(io.Discard, io.Discard, landingFlags())
			if mode == "valid" || mode == "pushed-but-unregistered" {
				if err != nil || gh.merges != 1 || gh.pr.HeadOID != local {
					t.Fatalf("err=%v merges=%d head=%s", err, gh.merges, gh.pr.HeadOID)
				}
			} else if err == nil || gh.merges != 0 || procedureHead(t, root) != local || git(t, gh.remote, "rev-parse", "refs/heads/"+landingTestBranch) != before {
				t.Fatalf("invalid resume changed state: %v", err)
			}
		})
	}
}

func TestLandingLeaseRefusesConcurrentPush(t *testing.T) {
	root, _, gh := integrationFixture(t)
	before := procedureHead(t, root)
	concurrent := ""
	mergeRunner = landingHookRunner{gitRunner: mergeRunner, before: func(_ string, args []string) error {
		if len(args) > 0 && args[0] == "push" && concurrent == "" {
			tree := git(t, root, "rev-parse", before+"^{tree}")
			concurrent = git(t, root, "commit-tree", tree, "-p", before, "-m", "concurrent push")
			git(t, root, "push", "-q", "upstream", concurrent+":refs/heads/"+landingTestBranch)
		}
		return nil
	}}
	err := runMerge(io.Discard, io.Discard, landingFlags())
	if err == nil || !strings.Contains(err.Error(), "lease") || concurrent == "" {
		t.Fatalf("wrong refusal: %v", err)
	}
	if gh.merges != 0 || git(t, gh.remote, "rev-parse", "refs/heads/"+landingTestBranch) != concurrent {
		t.Fatal("concurrent work overwritten")
	}
}

func TestLandingIntegrationPublishCollisionResets(t *testing.T) {
	for _, resume := range []bool{false, true} {
		t.Run(fmt.Sprint(resume), func(t *testing.T) {
			root, server, gh := integrationFixture(t)
			// Branch attributes permit a clean actual merge. Main's attributes govern
			// replay of the reviewed patch, so its same-line edit must still refuse.
			integrationAdvance(t, server, "shared.go", "base\n")
			git(t, root, "fetch", "-q", "upstream")
			git(t, root, "merge", "--no-edit", "upstream/main")
			procedureWrite(t, root, "shared.go", "reviewed\n")
			procedureWrite(t, root, ".gitattributes", "shared.go merge=keep\n")
			procedureWrite(t, root, procedureIssue, "---\nid: 000001\nstatus: codecomplete\n---\n# Procedure\n\n## Log\n\nReviewed shared code.\n")
			git(t, root, "add", ".")
			git(t, root, "commit", "-qm", "#1: close")
			git(t, root, "push", "-q", "upstream", landingTestBranch)
			git(t, root, "config", "merge.keep.driver", "true")
			before := procedureHead(t, root)
			gh.pr.HeadOID = before
			gh.ci.setHead(1, before)
			main := integrationAdvance(t, server, "shared.go", "main rewrites reviewed line\n")
			if resume {
				git(t, root, "fetch", "-q", "upstream")
				git(t, root, "merge", "--no-ff", "-m", "integration\n\nIntegrated-Main: "+main, main)
			}
			f := landingFlags()
			f.NoJudge = false
			oldGuard := guardTransferredDetailsFn
			guardTransferredDetailsFn = func(context.Context) error { return nil }
			t.Cleanup(func() { guardTransferredDetailsFn = oldGuard })
			err := runMerge(io.Discard, io.Discard, f)
			if err == nil || !strings.Contains(err.Error(), "fixed-then-ship") || !strings.Contains(err.Error(), "shared.go") {
				t.Fatalf("did not reach replay collision: %v", err)
			}
			if procedureHead(t, root) != before || gh.merges != 0 || git(t, gh.remote, "rev-parse", "refs/heads/"+landingTestBranch) != before {
				t.Fatal("replay refusal failed to reset before publishing")
			}
		})
	}
}

func TestIntegrationPRSync(t *testing.T) {
	oldTimeout, oldPoll := integrationSyncTimeout, integrationPollInterval
	integrationSyncTimeout, integrationPollInterval = 20*time.Millisecond, time.Millisecond
	t.Cleanup(func() { integrationSyncTimeout, integrationPollInterval = oldTimeout, oldPoll })
	oldHead, newHead := strings.Repeat("a", 40), strings.Repeat("b", 40)
	target := landingTarget{Repo: "test/repo"}
	previous := landingPR{Number: 1, Repo: target.Repo, HeadRef: landingTestBranch, HeadOID: oldHead, BaseRef: "main", State: "OPEN"}
	for _, mode := range []string{"timeout", "cancelled", "identity", "other-head", "merged"} {
		t.Run(mode, func(t *testing.T) {
			gh := &landingFakeGH{pr: previous}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			want := ""
			switch mode {
			case "timeout":
				want = "not registered"
			case "cancelled":
				cancel()
				want = "not registered"
			case "identity":
				gh.pr.Number++
				want = "identity"
			case "other-head":
				gh.pr.HeadOID = strings.Repeat("c", 40)
				want = "concurrently"
			case "merged":
				gh.pr.State = "MERGED"
				want = "state changed"
			}
			var out bytes.Buffer
			_, err := syncIntegrationPR(ctx, gh, target, landingTestBranch, newHead, previous, &out)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("err=%v want=%s", err, want)
			}
			if mode == "cancelled" && !errors.Is(err, context.Canceled) {
				t.Fatalf("cancellation lost: %v", err)
			}
			if mode == "timeout" && (!errors.Is(err, context.DeadlineExceeded) || !strings.Contains(out.String(), "PR #1")) {
				t.Fatalf("timeout err=%v output=%s", err, out.String())
			}
			if gh.merges != 0 {
				t.Fatal("sync merged a PR")
			}
		})
	}
}
