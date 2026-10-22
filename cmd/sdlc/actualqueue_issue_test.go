package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// Echo#9: another issue's failed suite spans two owned intervals in the same
// slot. The queue selector must first establish whose tested commit it is.
func TestOwnedActualQueueUsesTestedIssueBeforeSlotOwnership(t *testing.T) {
	for _, kind := range []string{"foreign", "own-crossing", "own-exact", "unknown"} {
		t.Run(kind, func(t *testing.T) {
			r, cardPath, _, owner, _ := ownedResolverFixture(t)
			foreignPath, foreignCard, _, _ := seededIssue(t, "000336", "other")
			foreignRaw, err := issue.SetCardClaimant([]byte(foreignCard), owner)
			if err != nil {
				t.Fatal(err)
			}
			ownedFixtureAt(t, 2)
			if err := r.tracker.Update(foreignPath, "claim-other", func([]byte) ([]byte, error) { return foreignRaw, nil }); err != nil {
				t.Fatal(err)
			}
			ownedFixtureAt(t, 4)
			ownedFixtureCard(t, r, cardPath, "release-first", func(raw []byte) ([]byte, error) {
				return regexp.MustCompile(`(?m)^claimant:\n(?:[ \t].*\n)*`).ReplaceAll(raw, nil), nil
			})
			ownedFixtureAt(t, 6)
			ownedFixtureCard(t, r, cardPath, "claim-second", func(raw []byte) ([]byte, error) { return issue.SetCardClaimant(raw, owner) })
			ownedFixtureAt(t, 11)
			testfix.Git(t, r.root, "commit", "--allow-empty", "-qm", "#336: other issue's failed candidate (see #335)")
			foreignHead := strings.TrimSpace(testfix.Capture(t, r.root, "rev-parse", "HEAD"))
			ownedFixtureAt(t, 12)
			testfix.Git(t, r.root, "commit", "--allow-empty", "-qm", "#335: own candidate")
			out, err := resolveOwnedActualInputsFor(context.Background(), r.root, "", "335", "", actualCloseCandidate)
			if err != nil {
				t.Fatal(err)
			}
			queued := time.Date(2026, 10, 14, 0, 3, 0, 0, time.UTC)
			acquired := queued.Add(4 * time.Minute)
			head := foreignHead
			switch kind {
			case "own-crossing":
				head = out.Snapshot.CodeHead
			case "own-exact":
				head = out.Snapshot.CodeHead
				queued = queued.Add(4 * time.Minute)
				acquired = queued.Add(time.Minute)
			case "unknown":
				head = strings.Repeat("d", 40)
			}
			queue := acquired.Sub(queued).Seconds()
			a := SuiteAttempt{Version: 1, ID: "fcf15650129203dfd1408728d625962b", Repo: "xianxu/ariadne", SHA: head, Check: "suite", Slot: owner.Workspace, QueuedAt: &queued, AcquiredAt: &acquired, QueueSeconds: &queue, StartedAt: acquired, FinishedAt: acquired.Add(time.Minute), Exit: 1, LogSHA256: strings.Repeat("b", 64)}
			common := strings.TrimSpace(testfix.Capture(t, r.root, "rev-parse", "--git-common-dir"))
			if !filepath.IsAbs(common) {
				common = filepath.Join(r.root, common)
			}
			dir := filepath.Join(common, "sdlc", "test-stats", "attempts")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(a)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, a.ID+".json"), raw, 0600); err != nil {
				t.Fatal(err)
			}
			repo, err := tracker.NewRepository(context.Background(), r.root, "origin")
			if err != nil {
				t.Fatal(err)
			}
			got := out
			err = attachOwnedQueueEvidence(context.Background(), r.root, repo, cardPath, "github.com/xianxu/ariadne", 335, got, actualCloseCandidate)
			switch kind {
			case "foreign":
				if err != nil || len(got.Snapshot.QueueWaits) != 0 {
					t.Fatalf("foreign failed suite poisoned own actual: inputs=%+v err=%v", got, err)
				}
			case "own-crossing":
				if err == nil || !strings.Contains(err.Error(), "0 matching pinned slot engagements") {
					t.Fatalf("own ambiguous wait must stay strict: %v", err)
				}
			case "own-exact":
				if err != nil || len(got.Snapshot.QueueWaits) != 1 || got.Snapshot.QueueWaits[0].QueueSeconds != 60 {
					t.Fatalf("own queue lost: inputs=%+v err=%v", got, err)
				}
			case "unknown":
				if err == nil || !strings.Contains(err.Error(), a.ID) || !strings.Contains(err.Error(), head) {
					t.Fatalf("unknown head must name attempt and SHA: %v", err)
				}
			}
		})
	}
}
