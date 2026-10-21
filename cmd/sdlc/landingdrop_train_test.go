package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// This is the operator-approved conformance: real Git merge ancestry and a
// native blocked old suite, with three independent production merge runners.
func TestLandingDropReformsRealTrain(t *testing.T) {
	for _, drop := range []int{1, 2} {
		t.Run(fmt.Sprint(drop), func(t *testing.T) {
			s := newTrainSetup(t, "train")
			roots := map[int]string{1: s.first}
			branches := map[int]string{1: landingTestBranch, 2: "000002-second", 3: "000003-third"}
			for _, i := range []int{2, 3} {
				root := filepath.Join(filepath.Dir(filepath.Dir(s.first)), fmt.Sprintf("sample-slot%d", i), "sample")
				if i == 3 {
					git(t, s.first, "worktree", "add", "-q", "-b", "main-slot3", root, "upstream/main")
				}
				git(t, root, "fetch", "-q", "upstream")
				git(t, root, "reset", "-q", "--hard", "upstream/main")
				trainCar(t, root, branches[i], fmt.Sprintf("car%d", i), i)
				s.addCar(t, branches[i], i)
				roots[i] = root
			}
			gate := filepath.Join(t.TempDir(), "release-old-suite")
			started := gate + ".started"
			droppedHead := s.heads[branches[drop]]
			s.declareSuite(t, fmt.Sprintf("if git merge-base --is-ancestor %s HEAD; then touch '%s'; while test ! -e '%s'; do sleep 0.01; done; fi; echo suite passed", droppedHead, started, gate))
			pending := map[int]<-chan error{}
			defer func() {
				_ = os.WriteFile(gate, nil, 0o600)
				s.blocker.Release()
				for i, done := range pending {
					if err := finished(t, done); err != nil && !strings.Contains(err.Error(), "held") {
						t.Errorf("cleanup runner%d: %v", i, err)
					}
				}
			}()
			for _, i := range []int{1, 2, 3} {
				pending[i] = s.merge(roots[i])
				waitQueued(t, s.line, i)
			}
			waitVouched(t, s.line, 3)
			s.blocker.Release()
			awaitDropCondition(t, "old native suite", func() bool { _, err := os.Stat(started); return err == nil })
			d := defaultLandingDeps()
			d.couchTree = func() (*int, bool) { return nil, false }
			d.dropIdentity = func(dir string) (workspace.Identity, error) {
				if dir == "." {
					dir = roots[drop]
				}
				return workspace.Resolve(execGitRunner{}, dir, "")
			}
			out, err := runLanding(t, d, "drop", "--repo", s.first, "--issue", fmt.Sprint(drop), "--reason", "late defect must stay out")
			if err != nil {
				t.Fatalf("drop refused: %v\n%s", err, out)
			}
			if !strings.Contains(out, "cleanup/recomposition pending") || !strings.Contains(out, branches[drop]) {
				t.Fatalf("drop output: %s", out)
			}
			for _, i := range []int{drop, 1, 2, 3} {
				done, ok := pending[i]
				if !ok {
					continue
				}
				delete(pending, i)
				err := finished(t, done)
				if i == drop {
					if err == nil || !strings.Contains(err.Error(), "held") || !strings.Contains(err.Error(), "late defect") {
						t.Fatalf("dropped runner: %v", err)
					}
				} else if err != nil {
					t.Fatalf("survivor%d: %v", i, err)
				}
			}
			main := git(t, s.gh.remote, "rev-parse", "refs/heads/main")
			for _, i := range []int{1, 2, 3} {
				_, err := trainGitOut(s.gh.remote, "merge-base", "--is-ancestor", s.heads[branches[i]], main)
				want := "MERGED"
				if i == drop {
					want = "OPEN"
					if err == nil {
						t.Fatal("dropped head landed")
					}
				} else if err != nil {
					t.Fatalf("survivor%d absent from main", i)
				}
				if got := s.gh.pr(branches[i]).State; got != want {
					t.Fatalf("car%d: %s want%s", i, got, want)
				}
			}
			events := lineEvents(t, s.line)[landingline.EventTrain]
			if len(s.gh.created) < 2 || s.gh.merges != 1 {
				t.Fatalf("old/reformed train PRs %v merges%d events%v", s.gh.created, s.gh.merges, events)
			}
			old := s.gh.pr(s.gh.created[0])
			if old.State != "CLOSED" {
				t.Fatalf("obsolete train: %s", old.State)
			}
			t.Logf("drop #%d: obsolete PR#%d CLOSED; survivors MERGED on %s; dropped commit excluded", drop, old.Number, main)
			wantNoTrainBranch(t, s.gh.remote)
			wantNoTrainWorktree(t, s.first)
		})
	}
}

func awaitDropCondition(t *testing.T, what string, ready func() bool) {
	t.Helper()
	deadline, ok := t.Deadline()
	now := time.Now()
	end := now.Add(landingCompletionBudget(deadline, ok, now))
	for !ready() {
		if time.Now().After(end) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}
