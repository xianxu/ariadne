package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

func TestPublishCompletionStatusNamesRecovery(t *testing.T) {
	for _, status := range vocab.Issue().AllStatuses() {
		err := publishCompletionStatus("000502", status)
		if status == "codecomplete" || status == "done" {
			if err != nil {
				t.Fatalf("%s: %v", status, err)
			}
			continue
		}
		if err == nil {
			t.Fatalf("accepted %s", status)
		}
		for _, phrase := range []string{"#502", status, "sdlc close --issue 502 --verified '<evidence>'", "push"} {
			if !strings.Contains(err.Error(), phrase) {
				t.Fatalf("%s: missing %q in %v", status, phrase, err)
			}
		}
	}
}

// Portable tracker fixture; cards and bindings are real persisted records.
func publishEligibilityFixture(t *testing.T, n int, mode string) (*trackerRepo, *trackerEnv, tracker.Records, string, string) {
	t.Helper()
	cards, details := map[string]string{}, map[string]string{}
	for i := 1; i <= n; i++ {
		cp, c, dp, d := seededIssue(t, fmt.Sprintf("%06d", i), "eligibility")
		c = strings.Replace(c, "status: open", "status: working", 1)
		cards[cp], details[dp] = c, d
	}
	r := newTrackerRepo(t, cards, details)
	base := r.git("rev-parse", "HEAD")
	r.git("switch", "-qc", "candidate")
	r.git("commit", "--allow-empty", "-qm", "candidate work")
	head := r.git("rev-parse", "HEAD")
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	writer, err := gitx.NewTrunkFileContext(context.Background(), r.root, "origin", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	bound := map[string][]byte{}
	for cp, c := range cards {
		evidence, repository := head, env.target.Repository
		if mode == "on-base" {
			evidence = base
		}
		if mode == "foreign" {
			repository = "another/repository"
		}
		if mode == "not-carried" {
			evidence = strings.Repeat("f", 40)
		}
		if mode == "no-binding" {
			bound[cp] = []byte(c)
			continue
		}
		if mode == "two-issues" && strings.Contains(cp, "000002-") {
			c = strings.Replace(c, "status: working", "status: codecomplete\nactual_hours: 1", 1)
		}
		b, err := issue.SetCardCompletion([]byte(c), issue.Completion{Token: "close-" + cp, Repository: repository, ReviewedHEAD: evidence, EvidenceCommit: evidence})
		if err != nil {
			t.Fatal(err)
		}
		bound[cp] = b
	}
	if err := writer.UpdateMany("bind fixture", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
		return gitx.TrunkWrite{Write: bound, ExactBytes: true}, nil
	}); err != nil {
		t.Fatal(err)
	}
	rs, err := loadIssueRecords(withIssueRecordsScope(context.Background()), filepath.Join(r.root, "workshop/issues"), tracker.Fresh)
	if err != nil {
		t.Fatal(err)
	}
	return r, env, rs, head, base
}

func TestPublishEligibilityPreservesIssueFreeBranches(t *testing.T) {
	for _, mode := range []string{"no-binding", "on-base", "foreign", "not-carried"} {
		t.Run(mode, func(t *testing.T) {
			_, env, rs, head, base := publishEligibilityFixture(t, 2, mode)
			if err := refuseReopenedCompletions(env, rs, head, base); err != nil {
				t.Fatal(err)
			}
			pr := landingPR{HeadOID: head, BaseOID: base}
			if err := runLandingPublishGate(withIssueRecordsScope(context.Background()), pr, "workshop/issues", io.Discard); err != nil {
				t.Fatal(err)
			}
			if err := runPublishGate(withIssueRecordsScope(context.Background()), base, "workshop/issues", io.Discard); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPublishEligibilityChecksEveryCarriedIssue(t *testing.T) {
	_, env, rs, head, base := publishEligibilityFixture(t, 2, "two-issues")
	// The eligible issue really supplies an anchor; it cannot hide #1's status.
	owned, err := ownedCompletions(env, rs, head, base, false)
	if err != nil || len(owned) != 1 || owned[0].ID != "000002" {
		t.Fatalf("eligible anchor: %+v %v", owned, err)
	}
	for _, gate := range []func() error{
		func() error {
			return runPublishGate(withIssueRecordsScope(context.Background()), base, "workshop/issues", io.Discard)
		},
		func() error {
			return runLandingPublishGate(withIssueRecordsScope(context.Background()), landingPR{HeadOID: head, BaseOID: base}, "workshop/issues", io.Discard)
		},
	} {
		if err := gate(); err == nil || !strings.Contains(err.Error(), "#1 is working") {
			t.Fatalf("eligible issue hid reopen: %v", err)
		}
	}
}

func TestPublishEligibilityOwnershipProbesScaleLinearly(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			_, env, rs, head, base := publishEligibilityFixture(t, n, "on-base")
			trace := filepath.Join(t.TempDir(), "git.trace")
			t.Setenv("GIT_TRACE", trace)
			if err := refuseReopenedCompletions(env, rs, head, base); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(trace)
			if err != nil {
				t.Fatal(err)
			}
			probes := 0
			for _, line := range strings.Split(string(raw), "\n") {
				if strings.Contains(line, "built-in: git merge-base --is-ancestor ") {
					probes++
				}
			}
			if probes != 2*n {
				t.Fatalf("%d cards: %d ownership ancestry probes, want %d", n, probes, 2*n)
			}
		})
	}
}
