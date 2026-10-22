package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

type trainMembershipCounter struct {
	bytes.Buffer
	writes int
}

func (w *trainMembershipCounter) Write(p []byte) (int, error) {
	w.writes++
	return w.Buffer.Write(p)
}

func TestLandingTrainMembershipOutputScalesWithCars(t *testing.T) {
	for _, label := range []string{"candidates", "composed"} {
		for _, n := range []int{8, 16} {
			t.Run(fmt.Sprintf("%s/%d", label, n), func(t *testing.T) {
				cars := make([]landingline.Car, n)
				for i := range cars {
					cars[i] = landingline.Car{Issue: i + 1, PR: i + 101}
				}
				var output trainMembershipCounter
				printTrainMembership(&output, 7, label, cars)
				if output.writes > n+2 {
					t.Fatalf("%d cars took %d writes, want at most %d", n, output.writes, n+2)
				}
				if !strings.HasPrefix(output.String(), fmt.Sprintf("landing train 7 %s: %d car(s):", label, n)) || !strings.HasSuffix(output.String(), "\n") {
					t.Fatalf("missing membership identity: %q", output.String())
				}
				for _, car := range cars {
					member := fmt.Sprintf(" #%d (PR #%d)", car.Issue, car.PR)
					if count := strings.Count(output.String(), member); count != 1 {
						t.Errorf("membership %q occurs %d times", member, count)
					}
				}
			})
		}
	}
}

// A conductor excluded during composition must return before publishing or
// running a surviving car's suite. All conflict outcomes still get delivered.
func TestLandingTrainHeldConductorReleasesSurvivors(t *testing.T) {
	s := newTrainSetup(t, "train")
	third := s.pushCar(t, "000003-third", "third", 3, 3)
	procedureWrite(t, third, "shipped.txt", "third conflicts with main\n")
	git(t, third, "add", "shipped.txt")
	git(t, third, "commit", "-qm", "third conflicts")
	git(t, third, "push", "-q", "upstream", "000003-third")
	s.addCar(t, "000003-third", 3)
	integrationAdvance(t, s.server, "shipped.txt", "main conflicts with first and third\n")
	gate := filepath.Join(t.TempDir(), "release-suite")
	started := gate + ".started"
	if err := os.WriteFile(gate, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	// The initial run is unblocked, so old code fails an assertion instead of
	// hanging. The reformed train blocks until its ownership is checked.
	s.declareSuite(t, fmt.Sprintf("touch '%s'; while test ! -e '%s'; do sleep 0.01; done; echo suite passed", started, gate))
	second := trainCarWithIssue(t, s, "---\nid: 000002\nstatus: codecomplete\n---\n# second\n")
	s.pushCar(t, "000004-fourth", "fourth", 4, 4)
	branchSuites := s.gh.suiteHeads()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	head, err := s.line.Enqueue(landingline.Ticket{Issue: 1, Branch: landingTestBranch})
	if err != nil {
		t.Fatal(err)
	}
	defer head.Release()
	secondTicket := s.queuedCar(t, "000002-second", 2, 2)
	thirdTicket := s.queuedCar(t, "000003-third", 3, 3)
	fourthTicket := s.queuedCar(t, "000004-fourth", 4, 4)
	s.blocker.Release()
	if err := s.line.Wait(ctx, head); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	turn := &landingTurn{line: s.line, handle: head, log: &lineLog{line: s.line, stderr: &stderr}, stderr: &stderr}
	pr := s.gh.pr(landingTestBranch)
	self := func(context.Context) (landingline.Readiness, bool, error) {
		return landingline.Readiness{Branch: landingTestBranch, Head: pr.HeadOID, PR: pr.Number, Ready: true}, false, nil
	}
	target := landingTarget{Root: s.first, Remote: "upstream", Repo: "test/repo"}
	_, _, err = maybeConductTrain(ctx, execGitRunner{}, s.gh, s.gh, target, landingFlags(), turn, landingTestBranch, pr.HeadOID, pr, landingline.Schedule{Interval: time.Hour}, self, &stderr)
	if err == nil || !strings.Contains(err.Error(), "held") || !strings.Contains(err.Error(), "conflicts") {
		t.Fatalf("conductor result: %v\n%s", err, &stderr)
	}
	if s.gh.merges != 0 || len(s.gh.created) != 0 {
		t.Fatalf("held conductor published or merged surviving cars: merges %d, created %v\n%s", s.gh.merges, s.gh.created, &stderr)
	}
	if got := s.gh.suiteHeads(); !slices.Equal(got, branchSuites) {
		t.Fatalf("held conductor ran another car's suite: %v, want only existing branch suites %v", got, branchSuites)
	}
	output := stderr.String()
	for _, want := range []string{"candidates:", "composed: 2 car(s): #2 (PR #2) #4 (PR #4)", "#1 held:", "fix the reported problem, integrate main, re-close", "sdlc merge --branch " + landingTestBranch + " --yes"} {
		if !strings.Contains(output, want) {
			t.Errorf("missing %q:\n%s", want, output)
		}
	}
	if strings.Contains(output, "suite passed") || strings.Contains(output, "landed; 4 car(s)") {
		t.Fatalf("held caller followed another car's suite:\n%s", output)
	}
	results := map[int]string{}
	for _, e := range lineEvents(t, s.line)[landingline.EventCar] {
		results[e.Issue] = e.Result
	}
	for id, want := range map[int]string{1: "held", 2: "released", 3: "held", 4: "released"} {
		if results[id] != want {
			t.Fatalf("car #%d outcome %q, want %q; all: %v", id, results[id], want, results)
		}
	}
	t.Logf("held conductor returned without publication or suite; car outcomes: %v", results)
	wantNoTrainBranch(t, s.gh.remote)
	wantNoTrainWorktree(t, s.first)

	// Requeue the survivors in a real prepared checkout, retaining the held
	// head's line lock until both are ready so the new conductor boards both.
	secondTicket.Release()
	thirdTicket.Release()
	fourthTicket.Release()
	if err := os.Remove(gate); err != nil {
		t.Fatal(err)
	}
	done := s.merge(second)
	// Join before fixture globals restore, even if an ownership assertion fails.
	defer func() {
		_ = os.WriteFile(gate, nil, 0o600)
		head.Release()
		if err := finished(t, done); err != nil {
			t.Errorf("reformed train: %v", err)
		}
		for _, branch := range []string{"000002-second", "000004-fourth"} {
			if car := s.gh.pr(branch); car.State != "MERGED" {
				t.Errorf("survivor %s remains %s", branch, car.State)
			}
		}
		for _, branch := range []string{landingTestBranch, "000003-third"} {
			if car := s.gh.pr(branch); car.State != "OPEN" {
				t.Errorf("held car %s became %s", branch, car.State)
			}
		}
	}()
	waitQueued(t, s.line, 1)
	s.queuedCar(t, "000004-fourth", 4, 4)
	waitVouched(t, s.line, 2)
	head.Release()
	logDir, err := localCheckLogDir(localRunDeps{Root: second, Git: func(args ...string) (string, error) { return trainGitOut(second, args...) }})
	if err != nil {
		t.Fatal(err)
	}
	runs := localCheckRunsDir(logDir)
	var runID string
	outerDeadline, hasDeadline := t.Deadline()
	now := time.Now()
	deadline := now.Add(landingCompletionBudget(outerDeadline, hasDeadline, now))
	for {
		files, err := filepath.Glob(filepath.Join(runs, "*.json"))
		if err != nil {
			t.Fatal(err)
		}
		_, startedErr := os.Stat(started)
		if len(files) == 1 && startedErr == nil {
			raw, err := os.ReadFile(files[0])
			if err != nil {
				t.Fatal(err)
			}
			var lease runLease
			if err := json.Unmarshal(raw, &lease); err != nil {
				t.Fatal(err)
			}
			if lease.Root != canonicalRoot(second) {
				t.Fatalf("reformed suite belongs to %s, want new conductor %s", lease.Root, second)
			}
			runID = strings.TrimSuffix(filepath.Base(files[0]), ".json")
			t.Logf("live reformed suite belongs to new conductor %s", lease.Root)
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("reformed train did not start a leased suite")
		}
		time.Sleep(time.Millisecond)
	}
	var cancelled bytes.Buffer
	if n, err := cancelLocalChecks(runs, canonicalRoot(s.first), "suite", "", &cancelled, time.Second); err != nil || n != 0 || !strings.Contains(cancelled.String(), "other checkouts left running") {
		t.Fatalf("held caller cancelled reformed train: %d, %v\n%s", n, err, &cancelled)
	}
	t.Logf("foreign cancel from held checkout: %s", strings.TrimSpace(cancelled.String()))
	if n, err := cancelLocalChecks(runs, canonicalRoot(s.first), "", runID, &cancelled, time.Second); n != 0 || err == nil || !strings.Contains(err.Error(), "belongs to another checkout") {
		t.Fatalf("held caller cancelled exact reformed run %s: %d, %v", runID, n, err)
	}
	select {
	case err := <-done:
		// The deferred join still needs this result.
		returned := make(chan error, 1)
		returned <- err
		done = returned
		t.Fatalf("reformed suite stopped before release: %v", err)
	default:
	}
}

func TestLandingTrainHeldWaiterReportsRecovery(t *testing.T) {
	s := newTrainSetup(t, "train")
	second := trainCarWithIssue(t, s, "---\nid: 000002\nstatus: codecomplete\n---\n# second\n")
	procedureWrite(t, second, "shipped.txt", "second conflicts with first\n")
	git(t, second, "add", "shipped.txt")
	git(t, second, "commit", "-qm", "second conflicts")
	git(t, second, "push", "-q", "upstream", "000002-second")
	s.addCar(t, "000002-second", 2)
	stderr := &watchWriter{want: "#2 held:", seen: make(chan struct{})}
	oldStderr := trainStderr
	trainStderr = stderr
	t.Cleanup(func() { trainStderr = oldStderr })
	firstErr, otherErr := runTwoCars(t, s, second)
	if firstErr != nil || otherErr == nil || !strings.Contains(otherErr.Error(), "held") {
		t.Fatalf("cars: %v / %v\n%s", firstErr, otherErr, stderr.String())
	}
	output := stderr.String()
	for _, want := range []string{"composed: 1 car(s): #1 (PR #1)", "#2 held:", "fix the reported problem, integrate main, re-close", "sdlc merge --branch 000002-second --yes", "landed; 1 car(s)"} {
		if !strings.Contains(output, want) {
			t.Errorf("missing %q:\n%s", want, output)
		}
	}
	if strings.Contains(output, "landed; 2 car(s)") {
		t.Fatalf("summary counted excluded car:\n%s", output)
	}
}
