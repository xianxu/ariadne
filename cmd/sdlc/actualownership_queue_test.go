package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// Echo #5's real completed attempt queued before its first tracker claim and
// acquired after the last issue commit. Keep the captured duration/provenance;
// only rebase its clock and immutable repository/head identity onto a fixture.
func TestActualOwnershipQueueAcrossFirstClaimUsesOwnedSuffix(t *testing.T) {
	capture, err := os.ReadFile("testdata/actual/echo5-claim-straddling-suite.json")
	if err != nil {
		t.Fatal(err)
	}
	var original SuiteAttempt
	if err := json.Unmarshal(capture, &original); err != nil {
		t.Fatal(err)
	}
	if original.ID != "21fd6681014fb81d65bb1ff403b31d51" || original.QueuedAt == nil || original.AcquiredAt == nil || original.QueueSeconds == nil || *original.QueueSeconds != 5204.94009 {
		t.Fatalf("live capture changed: %+v", original)
	}
	preclaimCapture, err := os.ReadFile("testdata/actual/echo5-preclaim-failed-suite.json")
	if err != nil {
		t.Fatal(err)
	}
	var originalPreclaim SuiteAttempt
	if err := json.Unmarshal(preclaimCapture, &originalPreclaim); err != nil {
		t.Fatal(err)
	}
	if originalPreclaim.ID != "0bbd2bec2bb41073e53e6994aad3d3df" || originalPreclaim.QueuedAt == nil || originalPreclaim.AcquiredAt == nil || originalPreclaim.Exit == 0 {
		t.Fatalf("preclaim failed capture changed: %+v", originalPreclaim)
	}
	for _, harness := range []string{"codex", "claude"} {
		t.Run(harness, func(t *testing.T) {
			r, worker, cardPath := ownershipRecoveryFixture(t)
			owner := hostClaimant(t, worker)
			owner.Workspace = "echo:2"
			withClaimant(t, owner)
			ownedFixtureAt(t, 20)
			if _, stderr, err := executeSDLCTestCommand("claim", "--issue", "335"); err != nil {
				t.Fatalf("real claim: %v %s", err, stderr)
			}
			ownershipNativeSource(t, harness, "legacy-prefix", r.root, 2, 4, 6)
			ownershipNativeSource(t, harness, "worker", worker, 13, 15, 17, 19, 22, 25, 30)
			ownedFixtureAt(t, 30)
			testfix.Git(t, worker, "commit", "--allow-empty", "-qm", "#335: genuine owned implementation")
			head := strings.TrimSpace(testfix.Git(t, worker, "rev-parse", "HEAD"))
			queued := time.Date(2026, 10, 14, 0, 15, 0, 0, time.UTC)
			delta := queued.Sub(*original.QueuedAt)
			attempt := original
			attempt.Repo, attempt.Slot, attempt.SHA = "xianxu/echo", owner.Workspace, head
			attempt.LogPath = filepath.Join("local-checks", head, attempt.Check+".log")
			acquired := original.AcquiredAt.Add(delta)
			attempt.QueuedAt, attempt.AcquiredAt = &queued, &acquired
			attempt.StartedAt = original.StartedAt.Add(delta)
			attempt.FinishedAt = original.FinishedAt.Add(delta)
			if acquired.Sub(queued) != original.AcquiredAt.Sub(*original.QueuedAt) || attempt.StartedAt.Sub(acquired) != original.StartedAt.Sub(*original.AcquiredAt) || attempt.FinishedAt.Sub(attempt.StartedAt) != original.FinishedAt.Sub(original.StartedAt) {
				t.Fatal("rebasing changed captured queue/setup/runtime durations")
			}
			common := strings.TrimSpace(testfix.Git(t, worker, "rev-parse", "--git-common-dir"))
			if !filepath.IsAbs(common) {
				common = filepath.Join(worker, common)
			}
			dir := filepath.Join(common, "sdlc", "test-stats", "attempts")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			writeAttempt := func(a SuiteAttempt) string {
				t.Helper()
				raw, err := json.Marshal(a)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := decodeSuiteAttempt(raw); err != nil {
					t.Fatalf("rebased captured attempt %s is invalid: %v", a.ID, err)
				}
				name := filepath.Join(dir, a.ID+".json")
				if err := os.WriteFile(name, raw, 0600); err != nil {
					t.Fatal(err)
				}
				return name
			}
			name := writeAttempt(attempt)
			// The other captured failed attempt acquired before the claim. Its
			// unknown SHA must be irrelevant before immutable head attribution.
			// An interrupted/cancelled invocation without a record is not invented.
			preclaim := originalPreclaim
			preclaim.Repo, preclaim.Slot, preclaim.SHA = attempt.Repo, attempt.Slot, strings.Repeat("f", 40)
			preEnd := time.Date(2026, 10, 14, 0, 19, 0, 0, time.UTC)
			shift := preEnd.Sub(*originalPreclaim.AcquiredAt)
			preStart := originalPreclaim.QueuedAt.Add(shift)
			preclaim.QueuedAt, preclaim.AcquiredAt = &preStart, &preEnd
			preclaim.StartedAt = originalPreclaim.StartedAt.Add(shift)
			preclaim.FinishedAt = originalPreclaim.FinishedAt.Add(shift)
			writeAttempt(preclaim)
			inputs, err := resolveOwnedActualInputsFor(t.Context(), worker, "", "335", "workshop/plans", actualCloseCandidate)
			if err != nil {
				t.Fatal(err)
			}
			repo, err := tracker.NewRepository(t.Context(), worker, "origin")
			if err != nil {
				t.Fatal(err)
			}
			// Private migration fixtures use a file: publication identity, while
			// persisted suite records require owner/repo names. As in the existing
			// hybrid-queue integration fixture, adapt only that scope at the real
			// attachment seam; retain file-scoped ownership and snapshot authority.
			if err := attachOwnedQueueEvidence(t.Context(), worker, repo, cardPath, "github.com/xianxu/echo", 335, inputs, actualCloseCandidate); err != nil {
				t.Fatalf("claim-straddling captured queue attachment: %v", err)
			}
			measured := computeOwnedActualFor(t.Context(), worker, "335", "", inputs)
			assertOwnershipGap(t, measured)
			if measured.Inputs.Method != gatestate.MeasurementMethodHybridQueueV1 || len(measured.Inputs.QueueWaits) != 1 {
				t.Fatalf("captured queue was not selected exactly once: %+v", measured)
			}
			wait := measured.Inputs.QueueWaits[0]
			if wait.AttemptID != original.ID || wait.Slot != owner.Workspace || wait.Start != "2026-10-14T00:20:00Z" || wait.End != "2026-10-14T00:30:00Z" || wait.QueueSeconds != 600 || math.Abs(wait.RawSeconds-5204.94009) > 1e-9 {
				t.Fatalf("queue did not use exact owned/pinned intersection: %+v", wait)
			}
			// Persist the real output, then make ambient queue/card evidence
			// disagree. Historical replay must retain the adopted clipping and
			// immutable code/tracker pins rather than discovering records again.
			saved := *measured.Inputs
			ledger := gatestate.Ledger{Gate: "boundary-review", IssueNum: 335, IDPrefix: "BR", Rounds: []gatestate.Round{{N: 1, Timestamp: "2026-10-14T00:31:00Z", Reviewed: saved.CodeHead, ActualInputs: &saved}}}
			writeRepoFile(t, worker, "workshop/plans/000335-spanning-close-gate.md", gatestate.Render(ledger, "fixture"))
			testfix.Git(t, worker, "add", "workshop/plans/000335-spanning-close-gate.md")
			ownedFixtureAt(t, 31)
			testfix.Git(t, worker, "commit", "-qm", "#335: retain claim-straddling queue measurement")
			evidence := strings.TrimSpace(testfix.Git(t, worker, "rev-parse", "HEAD"))
			ownedFixtureAt(t, 32)
			ownedFixtureCard(t, r, cardPath, "close-queue-recovered", func(raw []byte) ([]byte, error) {
				return issue.SetCardCompletion(raw, issue.Completion{Token: "close-queue-recovered", Repository: saved.Repository, ReviewedHEAD: saved.CodeHead, EvidenceCommit: evidence})
			})
			if err := os.WriteFile(name, []byte("malformed after immutable adoption\n"), 0600); err != nil {
				t.Fatal(err)
			}
			ownedFixtureAt(t, 35)
			ownedFixtureCard(t, r, cardPath, "queue-rename-after-close", func(raw []byte) ([]byte, error) {
				return issue.SetCardTitle(raw, "Later queue card edit")
			})
			replayed := computeActualForPurpose(t.Context(), worker, "", "335", "", "workshop/plans", actualHistorical)
			assertOwnershipGap(t, replayed)
			if replayed.Hours != measured.Hours || replayed.Inputs.CodeHead != saved.CodeHead || replayed.Inputs.TrackerHead != saved.TrackerHead || replayed.Inputs.Until != saved.Until || len(replayed.Inputs.QueueWaits) != 1 || !reflect.DeepEqual(replayed.Inputs.QueueWaits[0], wait) {
				t.Fatalf("queue replay changed immutable measurement: saved=%+v replay=%+v", measured, replayed)
			}
		})
	}
}
