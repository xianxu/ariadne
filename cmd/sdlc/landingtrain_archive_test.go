package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// #494: the conductor's final sweep archives another car's now-done card
// with a generic main commit. Both cars are real merges: a synthetic waiting
// ticket would never exercise the second caller's archive proof and cleanup.
func TestLandingTrainWaiterAcceptsConductorArchive(t *testing.T) {
	testLandingTrainSiblingArchive(t, 2, false)
}

// #512: the same confirmed done/landed/archive state stranded pair:7 (#727)
// and pair:4 (#733). Recreate both issue identities with real two-car merges;
// interrupt cleanup, then start a fresh invocation with no train notification.
func TestLandingTrainArchivedCarFreshResume(t *testing.T) {
	for _, tc := range []struct {
		id   int
		slot string
	}{{727, "pair:7"}, {733, "pair:4"}} {
		t.Run(fmt.Sprintf("%s/#%d", tc.slot, tc.id), func(t *testing.T) {
			testLandingTrainSiblingArchive(t, tc.id, true)
		})
	}
}

func testLandingTrainSiblingArchive(t *testing.T, secondID int, freshResume bool) {
	s := newTrainSetup(t, "train")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tf, err := gitx.NewTrunkFileContext(ctx, s.first, "upstream", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	secondBranch := fmt.Sprintf("%06d-second", secondID)
	secondDetail := "workshop/issues/" + secondBranch + ".md"
	secondHistory := "workshop/history/issues/" + secondBranch + ".md"
	ids := []int{1, secondID}
	paths := []string{"workshop/issue-cards/000001-procedure.md", "workshop/issue-cards/" + secondBranch + ".md"}
	cards := make([][]byte, 2)
	seed := map[string][]byte{tracker.ManifestPath: tracker.ManifestBytes()}
	for i, title := range []string{"Procedure", "second"} {
		cards[i] = []byte(fmt.Sprintf("---\nid: %06d\nstatus: codecomplete\nactual_hours: 1\nupdated: 2026-09-23\n---\n\n# %s\n\n## Problem\nTrain archive regression.\n", ids[i], title))
		seed[paths[i]] = cards[i]
	}
	if _, err := tf.Bootstrap(seed, "bootstrap train cards", func(gitx.BootstrapResult) error { return nil }); err != nil {
		t.Fatal(err)
	}
	// Both issue heads must postdate migration. Their mirrors pin real card
	// blobs in the tracker, allowing the generic sweep to project done status.
	markCutoverOn(t, s.server, "origin", "main")
	git(t, s.first, "fetch", "-q", "upstream")
	git(t, s.first, "branch", "-f", "main-slot1", "upstream/main")
	git(t, s.first, "rebase", "-q", "upstream/main")
	second := filepath.Join(filepath.Dir(filepath.Dir(s.first)), "sample-slot2", "sample")
	git(t, second, "fetch", "-q", "upstream")
	git(t, second, "reset", "-q", "--hard", "upstream/main")
	git(t, second, "switch", "--quiet", "--no-track", "-c", secondBranch)
	procedureWrite(t, second, "second.go", "package second\n")
	git(t, second, "add", "--", "second.go")
	git(t, second, "commit", "-qm", "close second")
	for i, root := range []string{s.first, second} {
		details, err := issue.MirrorDetails(append(append([]byte{}, cards[i]...), []byte("\n## Log\nClosed and reviewed.\n")...), cards[i], "sha1")
		if err != nil {
			t.Fatal(err)
		}
		detail := procedureIssue
		branch := landingTestBranch
		if i == 1 {
			detail, branch = secondDetail, secondBranch
		}
		procedureWrite(t, root, detail, string(details))
		git(t, root, "add", "--", detail)
		git(t, root, "commit", "-q", "--amend", "--no-edit")
		git(t, root, "push", "-q", "--force", "upstream", branch)
		s.addCar(t, branch, ids[i])
	}
	env, err := openTrackerAt(ctx, s.first)
	if err != nil {
		t.Fatal(err)
	}
	for i, branch := range []string{landingTestBranch, secondBranch} {
		cards[i], err = issue.SetCardCompletion(cards[i], issue.Completion{
			Token: fmt.Sprintf("close-train-%d", i+1), Repository: env.target.Repository,
			ReviewedHEAD: s.heads[branch], EvidenceCommit: s.heads[branch],
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := tf.UpdateMany("close both train cars", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
		return gitx.TrunkWrite{Write: map[string][]byte{paths[0]: cards[0], paths[1]: cards[1]}, ExactBytes: true}, nil
	}); err != nil {
		t.Fatal(err)
	}
	rest := git(t, second, "rev-parse", "main-slot2")
	conductorFinished := make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(conductorFinished) }) }
	var workers sync.WaitGroup
	// Unblock and join before fixture cleanups restore shared dependencies,
	// including when an assertion or queue timeout aborts the test early.
	t.Cleanup(func() {
		cancel()
		release()
		s.blocker.Release()
		workers.Wait()
	})
	start := func(root string, stderr io.Writer, waiter bool) chan error {
		done := make(chan error, 1)
		workers.Add(1)
		go func() {
			defer workers.Done()
			f := landingFlags()
			f.Context = withIssueRecordsScope(context.WithValue(ctx, trainRootKey{}, root))
			cmd := markManualLockCommand(&cobra.Command{Use: "merge"})
			cmd.SetContext(f.Context)
			f.Lock = func(run func() error) error {
				// OutcomeLanded confirms integration, before the conductor's
				// cleanup. Hold the waiter outside its transaction lock until
				// that cleanup (including finishLandedLeftovers) has finished.
				if waiter && s.gh.pr(secondBranch).State == "MERGED" {
					select {
					case <-conductorFinished:
						if freshResume {
							return errors.New("simulated runner stopped before cleanup")
						}
					case <-ctx.Done():
						return ctx.Err()
					}
				}
				return withRequiredRepoTransactionLock(cmd, run)
			}
			done <- runMerge(io.Discard, stderr, f)
		}()
		return done
	}
	secondStderr := watchWriter{want: "landed this PR", seen: make(chan struct{})}
	// This regression exercises a positively delivered train outcome. GitHub's
	// merged listing can otherwise win the readiness race and deliberately
	// resume without an outcome, before the final-cleanup barrier below applies.
	// Model a lagging listing until the waiter has consumed its outcome; the
	// underlying real merge remains visible to the archive barrier.
	ghClient = &trainArchiveListingLag{trainFakeGH: s.gh, branch: secondBranch, outcomeSeen: secondStderr.seen}
	firstStderr := watchWriter{want: "landed", seen: make(chan struct{})}
	one := start(s.first, &firstStderr, false)
	waitQueued(t, s.line, 1, []chan error{one})
	two := start(second, &secondStderr, true)
	waitQueued(t, s.line, 2, []chan error{one, two})
	waitVouched(t, s.line, 2, []chan error{one, two})
	s.blocker.Release()
	if err := finished(t, one); err != nil {
		t.Fatalf("conductor merge: %v", err)
	}
	// Establish the exact interleaving before allowing the waiter's cleanup:
	// its active details are gone in a generic archive, without Landing-PR.
	archiveCommit := git(t, s.gh.remote, "log", "-1", "--format=%H", "main", "--", secondHistory)
	if archiveCommit == "" {
		a, _ := tf.Read(paths[1])
		t.Fatalf("no sibling archive; card: %s\nconductor: %s\nevents: %+v", a, firstStderr.String(), lineEvents(t, s.line))
	}
	message := git(t, s.gh.remote, "show", "-s", "--format=%B", archiveCommit)
	if !strings.Contains(message, fmt.Sprintf("#%d: issue: archive after a merge outside sdlc", secondID)) || strings.Contains(message, "Landing-PR:") {
		t.Fatalf("expected the conductor's generic archive, got:\n%s", message)
	}
	if _, err := trainGitOut(s.gh.remote, "cat-file", "-e", "main:"+secondDetail); err == nil {
		t.Fatal("conductor left the second issue active")
	}
	if archived := git(t, s.gh.remote, "show", "main:"+secondHistory); !strings.Contains(archived, "status: done") {
		t.Fatalf("conductor did not archive the done card:\n%s", archived)
	}
	release()
	otherErr := finished(t, two) // join before reading its private stderr
	if freshResume {
		if otherErr == nil || !strings.Contains(otherErr.Error(), "simulated runner stopped") {
			t.Fatalf("interruption: %v", otherErr)
		}
		beforeCard, err := tf.Read(paths[1])
		if err != nil {
			t.Fatal(err)
		}
		merges := s.gh.merges
		ghClient = s.gh
		retry := landingFlags()
		retry.Context = withIssueRecordsScope(context.WithValue(ctx, trainRootKey{}, second))
		// Warm the command-scoped authority just as normal CLI entry does.
		if _, err := loadIssueRecordsAt(retry.Context, second, "workshop/issues", tracker.Fresh); err != nil {
			t.Fatal(err)
		}
		if err := runMerge(io.Discard, &secondStderr, retry); err != nil {
			t.Fatalf("fresh merged retry: %v", err)
		}
		afterCard, err := tf.Read(paths[1])
		if err != nil {
			t.Fatal(err)
		}
		if string(beforeCard) != string(afterCard) || s.gh.merges != merges {
			t.Fatal("retry rewrote completion ownership or repeated server merge")
		}
	} else if otherErr != nil {
		t.Fatalf("waiter merge: %v\n%s", otherErr, secondStderr.String())
	}
	if output := secondStderr.String(); (!freshResume && !strings.Contains(output, "archived by the train")) || strings.Contains(output, "active issue missing without confirmed archive") {
		t.Fatalf("waiter did not report its confirmed train archive:\n%s", output)
	}
	if current := git(t, second, "branch", "--show-current"); current != "main-slot2" {
		t.Fatalf("waiter remains on %s", current)
	}
	if head := procedureHead(t, second); head != rest {
		t.Fatalf("waiter's resting branch moved: %s -> %s", rest, head)
	}
	if branch := git(t, second, "branch", "--list", secondBranch); branch != "" {
		t.Fatalf("waiter's landed branch remains: %s", branch)
	}
}

// trainArchiveListingLag preserves the real train fake's effects while making
// the archive regression's outcome-before-cleanup ordering independent of the
// readiness goroutine's scheduling. The separate no-outcome recovery control
// continues to exercise an immediately visible merged listing.
type trainArchiveListingLag struct {
	*trainFakeGH
	branch      string
	outcomeSeen <-chan struct{}
}

func (g *trainArchiveListingLag) LandingPRs(ctx context.Context, repo, branch string) ([]landingPR, error) {
	prs, err := g.trainFakeGH.LandingPRs(ctx, repo, branch)
	if err != nil || branch != g.branch {
		return prs, err
	}
	select {
	case <-g.outcomeSeen:
		return prs, nil
	default:
	}
	for i := range prs {
		if prs[i].State == "MERGED" {
			prs[i].State, prs[i].MergeOID = "OPEN", ""
		}
	}
	return prs, nil
}
