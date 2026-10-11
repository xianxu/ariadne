package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// catchUpFixture is a claimed issue #9 in a slot whose construct/deps names one
// substrate checkout; both publish through bare origins.
type catchUpFixture struct {
	slot, dep *trackerRepo
	cardPath  string
}

func newCatchUpFixture(t *testing.T) catchUpFixture {
	t.Helper()
	depRoot := testfix.Repo(t, testfix.InitialCommit())
	if resolved, err := filepath.EvalSymlinks(depRoot); err == nil {
		depRoot = resolved
	}
	depOrigin := filepath.Join(t.TempDir(), "dep.git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", depOrigin)
	testfix.Git(t, depRoot, "remote", "add", "origin", depOrigin)
	testfix.Git(t, depRoot, "push", "-q", "-u", "origin", "main")
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	slot := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{
		detailPath:       detail,
		"construct/deps": "substrate " + depRoot + "\n",
	})
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	return catchUpFixture{slot: slot, dep: &trackerRepo{t: t, root: depRoot, origin: depOrigin}, cardPath: cardPath}
}

// fakeWeave installs a weaveRefresh that records each call's branch state and
// then runs do.
func fakeWeave(t *testing.T, f catchUpFixture, do func(out io.Writer) error) *[]string {
	t.Helper()
	var calls []string
	prev := weaveRefresh
	weaveRefresh = func(_ context.Context, root string, out io.Writer) error {
		calls = append(calls, root+" on "+f.slot.git("branch", "--show-current"))
		return do(out)
	}
	t.Cleanup(func() { weaveRefresh = prev })
	return &calls
}

func pinCutoffClock(t *testing.T, offset time.Duration) {
	t.Helper()
	prev := integrateNow
	integrateNow = func() time.Time { return time.Now().Add(offset) }
	t.Cleanup(func() { integrateNow = prev })
}

// TestStartPlanCatchesUpSlotAndDependency is #303's fixture: a slot and its
// dependency both trail origin/main. start-plan runs weave refresh on the
// resting branch, before any issue branch exists, then stops for a re-run with
// the card untouched; the re-run finds everything current and plans.
func TestStartPlanCatchesUpSlotAndDependency(t *testing.T) {
	f := newCatchUpFixture(t)
	peerAdd(t, f.slot, "slot.go", "package slot\n")
	peerAdd(t, f.dep, "dep.go", "package dep\n")
	claimed := f.slot.card(f.cardPath)
	calls := fakeWeave(t, f, func(out io.Writer) error {
		for _, r := range []*trackerRepo{f.slot, f.dep} {
			r.git("merge", "-q", "--ff-only", "origin/main")
		}
		fmt.Fprintln(out, "fake weave: refreshed")
		return nil
	})
	var out bytes.Buffer
	err := startPlanBranch(context.Background(), &out, 9)
	if err == nil || !strings.Contains(err.Error(), "Re-run `sdlc start-plan --issue 9`") {
		t.Fatalf("a catch-up must stop for a re-run: %v\n%s", err, out.String())
	}
	if len(*calls) != 1 || (*calls)[0] != f.slot.root+" on main" {
		t.Fatalf("weave refresh must run once in the slot on its resting branch: %q", *calls)
	}
	for _, r := range []*trackerRepo{f.slot, f.dep} {
		if r.git("rev-parse", "HEAD") != r.originMain() {
			t.Fatalf("%s was not caught up", r.root)
		}
	}
	if f.slot.git("branch", "--list", "000009-nine") != "" || f.slot.card(f.cardPath) != claimed {
		t.Fatal("a catch-up stop must not branch or start the card")
	}
	if err := startPlanBranch(context.Background(), &out, 9); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 {
		t.Fatalf("a current slot and dependency must not run weave again: %q", *calls)
	}
	if f.slot.git("branch", "--show-current") != "000009-nine" || !strings.Contains(f.slot.card(f.cardPath), "status: working") {
		t.Fatal("the re-run did not plan the issue")
	}
}

// A refresh that fails refuses when a lagging checkout lacks code main had by the
// cutoff, carrying refresh's account of what is dirty and naming the missing code.
func TestStartPlanRefusesAStaleCheckoutRefreshCannotCatchUp(t *testing.T) {
	f := newCatchUpFixture(t)
	peerAdd(t, f.dep, "dep.go", "package dep\n")
	pinCutoffClock(t, 48*time.Hour) // the dependency's landing is before the cutoff
	claimed := f.slot.card(f.cardPath)
	fakeWeave(t, f, func(out io.Writer) error {
		fmt.Fprintln(out, "fake weave: dirty: dep/notes.go")
		return errors.New("exit status 1")
	})
	var out bytes.Buffer
	err := startPlanBranch(context.Background(), &out, 9)
	if err == nil || !strings.Contains(err.Error(), f.dep.root+" (dep.go)") || !strings.Contains(err.Error(), "Nothing was branched or started") {
		t.Fatalf("must refuse naming the stale dependency and its missing code: %v", err)
	}
	if strings.Contains(err.Error(), f.slot.root+" (") {
		t.Fatalf("the current slot must not be named stale: %v", err)
	}
	if !strings.Contains(out.String(), "dirty: dep/notes.go") {
		t.Fatalf("refresh's output must reach the operator:\n%s", out.String())
	}
	if f.slot.git("branch", "--show-current") != "main" || f.slot.card(f.cardPath) != claimed {
		t.Fatal("a refused start-plan changed the branch or the card")
	}
}

// A refresh that fails over a lag the freshness rule tolerates (landed after the
// cutoff, or docs only) warns and plans, as close's integrate gate would accept it.
func TestStartPlanToleratesALagFreshnessAccepts(t *testing.T) {
	for _, tc := range []struct {
		name, file string
		offset     time.Duration
	}{
		{"code after the cutoff", "dep.go", -48 * time.Hour},
		{"docs before the cutoff", "NOTES.md", 48 * time.Hour},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newCatchUpFixture(t)
			peerAdd(t, f.dep, tc.file, "x\n")
			pinCutoffClock(t, tc.offset)
			fakeWeave(t, f, func(io.Writer) error { return errors.New("exit status 1") })
			var out bytes.Buffer
			if err := startPlanBranch(context.Background(), &out, 9); err != nil {
				t.Fatalf("a tolerated lag must not refuse: %v", err)
			}
			if !strings.Contains(out.String(), "planning on the current checkouts") {
				t.Fatalf("the failed refresh must be warned:\n%s", out.String())
			}
			if f.slot.git("branch", "--show-current") != "000009-nine" {
				t.Fatal("start-plan did not plan the issue")
			}
		})
	}
}
