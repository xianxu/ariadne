package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// heldLine opens the fixture repository's landing line and admits a head that
// the test holds, standing in for another slot's merge mid-landing.
func heldLine(t *testing.T, root string) (*landingline.Line, *landingline.Handle) {
	t.Helper()
	old := landingLinePoll
	landingLinePoll = 5 * time.Millisecond
	t.Cleanup(func() { landingLinePoll = old })
	dir, err := landingLineDir(mergeRunner, root)
	if err != nil {
		t.Fatal(err)
	}
	line, err := landingline.Open(dir, landingline.Options{Poll: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	h, err := line.Enqueue(landingline.Ticket{Issue: 999, Branch: "another-slot"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(h.Release)
	if err := line.Wait(context.Background(), h); err != nil {
		t.Fatal(err)
	}
	return line, h
}

// mergeAsync runs one merge in the background. The fake GitHub is shared with
// the test goroutine, which only touches it while the merge waits in the line.
func mergeAsync(f *mergeFlags) <-chan error {
	done := make(chan error, 1)
	go func() { done <- runMerge(io.Discard, io.Discard, f) }()
	return done
}

func waitQueued(t *testing.T, line *landingline.Line, n int) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		if _, waiting, err := line.Status(); err == nil && len(waiting) == n {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("no %d waiting merges", n)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func finished(t *testing.T, done <-chan error) error {
	t.Helper()
	return awaitLandingCompletion(t, done, "merge did not finish")
}

func lineEvents(t *testing.T, line *landingline.Line) map[string][]landingline.Event {
	t.Helper()
	events, _, err := line.Events()
	if err != nil {
		t.Fatal(err)
	}
	byKind := map[string][]landingline.Event{}
	for _, e := range events {
		byKind[e.Kind] = append(byKind[e.Kind], e)
	}
	return byKind
}

// The point of the line: while another merge holds it, this one doesn't
// integrate, test or land. Once that merge has landed (main moved) and left,
// this one integrates the new main once, runs one suite, and lands.
func TestLandingLineWaitsThenIntegratesOnce(t *testing.T) {
	root, server, gh := integrationFixture(t)
	line, head := heldLine(t, root)
	var tested []string
	gh.afterStatus = func(sha string) { tested = append(tested, sha) }
	before := git(t, server, "ls-remote", "origin", "refs/heads/"+landingTestBranch)
	done := mergeAsync(landingFlags())
	waitQueued(t, line, 1)
	time.Sleep(50 * time.Millisecond) // several polls while the head holds the line
	if gh.merges != 0 || len(tested) != 0 || git(t, server, "ls-remote", "origin", "refs/heads/"+landingTestBranch) != before {
		t.Fatalf("worked while waiting: merges=%d suites=%v", gh.merges, tested)
	}
	moved := integrationAdvance(t, server, "head.go", "the other slot landed\n")
	head.Release()
	if err := finished(t, done); err != nil {
		t.Fatal(err)
	}
	if gh.merges != 1 || len(tested) != 1 {
		t.Fatalf("merges=%d suites=%v", gh.merges, tested)
	}
	if included, err := integrationAncestor(mergeRunner, root, moved, tested[0]); err != nil || !included {
		t.Fatalf("the suite's head %s doesn't contain the main that moved while waiting (%s)", tested[0], moved)
	}
	events := lineEvents(t, line)
	// The test's own head logs nothing; the merge logs its enqueue, admit and landing.
	if len(events[landingline.EventEnqueue]) != 1 || len(events[landingline.EventAdmit]) != 1 || len(events[landingline.EventLanded]) != 1 {
		t.Fatalf("events %v", events)
	}
	enq, adm, landed := events[landingline.EventEnqueue][0], events[landingline.EventAdmit][0], events[landingline.EventLanded][0]
	if landed.PR != gh.pr.Number || landed.Branch != landingTestBranch || landed.Issue != 1 || enq.Issue != 1 ||
		adm.Seq != enq.Seq || landed.Seq != enq.Seq || len(adm.Displaced) != 0 {
		t.Fatalf("events enqueue %+v admit %+v landed %+v", enq, adm, landed)
	}
	if _, waiting, _ := line.Status(); len(waiting) != 0 {
		t.Fatalf("left in line: %v", waiting)
	}
}

// PQ-1: a duplicate invocation waits while the original lands the PR. Admitted
// afterwards, it observes the PR anew and resumes as merged: no second merge.
func TestLandingLineDuplicateResumesAsMerged(t *testing.T) {
	root, _, gh := integrationFixture(t)
	line, head := heldLine(t, root)
	done := mergeAsync(landingFlags())
	waitQueued(t, line, 1)
	if err := gh.LandingMerge(context.Background(), gh.pr.Repo, gh.pr.Number, gh.pr.HeadOID); err != nil {
		t.Fatal(err)
	}
	head.Release()
	if err := finished(t, done); err != nil {
		t.Fatal(err)
	}
	if gh.merges != 1 {
		t.Fatalf("merges=%d", gh.merges)
	}
	// A resume isn't a landing: it mustn't count twice in the pilot's delays.
	if events := lineEvents(t, line); len(events[landingline.EventResumed]) != 1 || len(events[landingline.EventLanded]) != 0 {
		t.Fatalf("events %v", events)
	}
}

// A reused branch name still queues the current OPEN PR. Historical completed
// heads cannot bypass admission or block integration and cleanup afterwards.
func TestLandingLineReusedBranchStillQueues(t *testing.T) {
	root, _, gh := integrationFixture(t)
	gh.older = []landingPR{{Number: 380, State: "MERGED", Repo: gh.pr.Repo, HeadRef: landingTestBranch, HeadOID: gh.pr.BaseOID, BaseRef: "main", BaseOID: gh.pr.BaseOID, MergeOID: gh.pr.BaseOID}}
	line, head := heldLine(t, root)
	done := mergeAsync(landingFlags())
	waitQueued(t, line, 1)
	time.Sleep(50 * time.Millisecond)
	if gh.merges != 0 {
		t.Fatal("landed outside the line")
	}
	head.Release()
	if err := finished(t, done); err != nil {
		t.Fatal(err)
	}
	if events := lineEvents(t, line); len(events[landingline.EventAdmit]) != 1 {
		t.Fatalf("the open PR's merge was never admitted: %v", events)
	}
}

// An already-merged PR only resumes cleanup: it takes no turn, so it finishes
// while another merge holds the line.
func TestLandingLineAlreadyMergedTakesNoTurn(t *testing.T) {
	root, _, gh := integrationFixture(t)
	line, _ := heldLine(t, root)
	if err := gh.LandingMerge(context.Background(), gh.pr.Repo, gh.pr.Number, gh.pr.HeadOID); err != nil {
		t.Fatal(err)
	}
	if err := finished(t, mergeAsync(landingFlags())); err != nil {
		t.Fatal(err)
	}
	if events := lineEvents(t, line); len(events[landingline.EventEnqueue]) != 0 {
		t.Fatalf("an already-merged PR queued: %v", events)
	}
}

// A refusal after admission frees the line and is logged as refused.
func TestLandingLineRefusalLeavesTheLine(t *testing.T) {
	root, server, gh := integrationFixture(t)
	procedureWrite(t, server, localChecksPath, `{"version":1,"checks":[{"name":"suite","required":true,"command":["sh","-c","exit 1"]}]}`)
	git(t, server, "add", localChecksPath)
	git(t, server, "commit", "-qm", "a failing suite")
	git(t, server, "push", "-q", "origin", "main")
	line, head := heldLine(t, root)
	head.Release()
	if err := finished(t, mergeAsync(landingFlags())); err == nil {
		t.Fatal("landed with a failing suite")
	}
	if gh.merges != 0 {
		t.Fatalf("merges=%d", gh.merges)
	}
	if h, waiting, _ := line.Status(); h != nil || len(waiting) != 0 {
		t.Fatalf("line not freed: head %v waiting %v", h, waiting)
	}
	if events := lineEvents(t, line); len(events[landingline.EventRefused]) != 1 {
		t.Fatalf("events %v", events)
	}
}

// PQ-1: the branch moves while the merge waits (its author pushes a fix from
// the slot). Admitted afterwards, the landing selects the new head, integrates
// it once and lands it.
func TestLandingLinePushedWhileWaiting(t *testing.T) {
	root, _, gh := integrationFixture(t)
	line, head := heldLine(t, root)
	var tested []string
	gh.afterStatus = func(sha string) { tested = append(tested, sha) }
	done := mergeAsync(landingFlags())
	waitQueued(t, line, 1)
	procedureWrite(t, root, "fix.txt", "pushed while waiting\n")
	git(t, root, "add", "fix.txt")
	git(t, root, "commit", "-qm", "a fix pushed while waiting")
	git(t, root, "push", "-q", "upstream", landingTestBranch)
	fix := procedureHead(t, root)
	head.Release()
	if err := finished(t, done); err != nil {
		t.Fatal(err)
	}
	if gh.merges != 1 || len(tested) != 1 {
		t.Fatalf("merges=%d suites=%v", gh.merges, tested)
	}
	if included, err := integrationAncestor(mergeRunner, root, fix, tested[0]); err != nil || !included {
		t.Fatalf("the landed head %s doesn't contain the fix %s pushed while waiting", tested[0], fix)
	}
}

// The rule (M2 review I-A): a merge without a turn acts only on an already
// merged PR. A lagging first listing that shows only an old merged PR skips
// the line, so when the next read shows the open PR, the merge refuses having
// changed nothing, instead of landing outside the line.
func TestLandingLineLaggingListingRefusesOutsideTheLine(t *testing.T) {
	root, server, gh := integrationFixture(t)
	old := landingPR{Number: 0, State: "MERGED", Repo: gh.pr.Repo, HeadRef: landingTestBranch, HeadOID: gh.pr.BaseOID, BaseRef: "main", BaseOID: gh.pr.BaseOID, MergeOID: gh.pr.BaseOID}
	gh.lagFirst, gh.firstListing = true, []landingPR{old}
	line, _ := heldLine(t, root)
	before := git(t, server, "ls-remote", "origin", "refs/heads/"+landingTestBranch)
	err := finished(t, mergeAsync(landingFlags()))
	if err == nil || !strings.Contains(err.Error(), "appeared open after this merge chose not to queue") {
		t.Fatalf("err %v", err)
	}
	if gh.merges != 0 || git(t, server, "ls-remote", "origin", "refs/heads/"+landingTestBranch) != before {
		t.Fatal("acted outside the line")
	}
	if events := lineEvents(t, line); len(events[landingline.EventEnqueue]) != 0 {
		t.Fatalf("events %v", events)
	}
}

// An empty first listing (GitHub not yet showing a new PR) is not a resume:
// the merge queues.
func TestLandingLineEmptyListingQueues(t *testing.T) {
	root, _, gh := integrationFixture(t)
	gh.lagFirst, gh.firstListing = true, nil
	line, head := heldLine(t, root)
	done := mergeAsync(landingFlags())
	waitQueued(t, line, 1)
	head.Release()
	if err := finished(t, done); err != nil {
		t.Fatal(err)
	}
	if gh.merges != 1 {
		t.Fatalf("merges=%d", gh.merges)
	}
}

// Lost audit data is visible: a log write failure warns once, and a landed
// issue whose grants can't be read says its priority wasn't ended.
func TestLandingLineLogFailuresWarn(t *testing.T) {
	dir := t.TempDir()
	var stderr bytes.Buffer
	line, log, err := openLandingLine(dir, nil, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{landingline.EventsFile, landingline.GrantsFile} {
		if err := os.Mkdir(filepath.Join(dir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	h, err := line.Enqueue(landingline.Ticket{Issue: 5, Branch: "000005-x"})
	if err != nil {
		t.Fatal(err)
	}
	turn := &landingTurn{line: line, handle: h, log: log, stderr: &stderr}
	log.record(landingline.Event{Kind: landingline.EventEnqueue, Seq: 1})
	turn.leave(nil)
	out := stderr.String()
	if strings.Count(out, "could not log") != 1 || !strings.Contains(out, "could not read grants to end #5's priority") {
		t.Fatalf("stderr %q", out)
	}
}

func TestLandingGrantHoldOutputSerialAndTrain(t *testing.T) {
	for _, trains := range []bool{false, true} {
		t.Run(fmt.Sprint(trains), func(t *testing.T) {
			var out bytes.Buffer
			var sched func() landingline.Schedule
			if trains {
				sched = func() landingline.Schedule { return landingline.Schedule{Interval: time.Hour, MaxCars: 8} }
			}
			line, _, err := openLandingLine(t.TempDir(), sched, &out)
			if err != nil {
				t.Fatal(err)
			}
			expires := time.Now().Add(time.Hour)
			if err := line.AddGrant(landingline.Grant{Issue: 470, Grantor: landingline.Grantor{Address: "operator shell"}, Reason: "main is red", Expires: expires}); err != nil {
				t.Fatal(err)
			}
			h, err := line.Enqueue(landingline.Ticket{Issue: 2})
			if err != nil {
				t.Fatal(err)
			}
			defer h.Release()
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if err := line.Wait(ctx, h); err != context.Canceled {
				t.Fatalf("wait %v", err)
			}
			prefix := "landing line: held:"
			if trains {
				prefix = "landing train: held:"
			}
			for _, want := range []string{prefix, "#470", "main is red", expires.Format(time.RFC3339), "sdlc landing revoke --issue 470"} {
				if !strings.Contains(out.String(), want) {
					t.Fatalf("missing %q: %s", want, out.String())
				}
			}
		})
	}
}
