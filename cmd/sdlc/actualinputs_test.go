package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
)

func TestOwnedActualResolverWithoutTrackerReturnsLegacy(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	got, err := resolveOwnedActualInputs(context.Background(), root, "", "335", "")
	if err != nil || got != nil {
		t.Fatalf("inputs=%+v error=%v", got, err)
	}
}

func TestOwnedActualReplayUsesSavedPrefixAndWindows(t *testing.T) {
	at := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	saved := gatestate.MeasurementInputs{Version: 1, Method: gatestate.MeasurementMethodEngagementV1, Repository: "repo", CodeHead: strings.Repeat("a", 40), TrackerHead: strings.Repeat("b", 40), Since: at.Format(time.RFC3339), Until: at.Add(time.Hour).Format(time.RFC3339), GapMinutes: 15, CommitWeight: 1, Engagements: []gatestate.MeasurementEngagement{{Start: at.Format(time.RFC3339), End: at.Add(time.Hour).Format(time.RFC3339), Began: at.Format(time.RFC3339), Roots: []string{"/repo"}}}, Sources: []gatestate.MeasurementSource{{Harness: "codex", ID: "root", CWD: "/repo", Digest: strings.Repeat("c", 64), Bytes: 42, Started: at.Format(time.RFC3339), Windows: []gatestate.MeasurementInterval{{Start: at.Add(time.Minute).Format(time.RFC3339), End: at.Add(2 * time.Minute).Format(time.RFC3339)}}}}}
	found := []transcripts.SessionEvidence{{Harness: "codex", ID: "root", CWD: "/repo", File: "current-location", Length: 900, Digest: strings.Repeat("d", 64), Started: at}}
	got, err := ownedReplaySources(saved, found)
	if err != nil {
		t.Fatal(err)
	}
	if got.Prefixes["current-location"] != (activetime.SourcePrefix{Bytes: 42, Digest: strings.Repeat("c", 64)}) || len(got.Windows["current-location"]) != 1 || !got.Windows["current-location"][0].Start.Equal(at.Add(time.Minute)) {
		t.Fatalf("replay did not use saved prefix/window: %+v", got)
	}
	if !got.Windows["current-location"][0].Began.Equal(at) {
		t.Fatal("replay lost the dispatch engagement origin")
	}
	broken := saved
	broken.Engagements = nil
	if _, err := ownedReplaySources(broken, found); err == nil {
		t.Fatal("unbound retained dispatch window silently admitted")
	}
	if got.Scope.Head != saved.CodeHead {
		t.Fatal("replay scope does not pin code history")
	}
	if _, err := ownedReplaySources(saved, nil); err == nil {
		t.Fatal("missing retained source silently reconstructed")
	}
}

func ownedFixtureAt(t *testing.T, minute int) {
	t.Helper()
	stamp := fmt.Sprintf("2026-10-14T00:%02d:00Z", minute)
	t.Setenv("GIT_AUTHOR_DATE", stamp)
	t.Setenv("GIT_COMMITTER_DATE", stamp)
}

func ownedResolverFixture(t *testing.T) (*trackerRepo, string, string, issue.Claimant, string) {
	t.Helper()
	ownedFixtureAt(t, 0)
	t.Setenv("HOME", t.TempDir())
	cardPath, card, detailPath, detail := seededIssue(t, "000335", "owned")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	ownerRoot := filepath.Join(t.TempDir(), "owner")
	testfix.Git(t, r.root, "worktree", "add", "-q", "-b", "000335-owned", ownerRoot)
	owner := hostClaimant(t, ownerRoot)
	ownedFixtureAt(t, 1)
	ownedFixtureCard(t, r, cardPath, "claim-first", func(raw []byte) ([]byte, error) {
		raw, err := issue.SetCardClaimant(raw, owner)
		if err != nil {
			return nil, err
		}
		return issue.SetCardField(raw, "started", "2026-10-14T00:01:00Z")
	})
	path := filepath.Join(os.Getenv("HOME"), ".codex", "sessions", "owner.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	data := fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:00:00Z\",\"type\":\"session_meta\",\"payload\":{\"id\":\"owner\",\"cwd\":%q}}\n", owner.Worktree)
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	ownedFixtureAt(t, 10)
	testfix.Git(t, r.root, "commit", "--allow-empty", "-qm", "#335: work")
	return r, cardPath, ownerRoot, owner, path
}

func ownedFixtureCard(t *testing.T, r *trackerRepo, cardPath, token string, edit func([]byte) ([]byte, error)) {
	t.Helper()
	repo, err := tracker.NewRepository(context.Background(), r.root, "origin")
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	card, ok := snapshot.Card("000335")
	if !ok {
		t.Fatal("missing fixture card")
	}
	raw, err := edit(card.Raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateCard(card, raw, token, func(string, string) error { return nil }); err != nil {
		t.Fatal(err)
	}
}

func TestOwnedActualResolverUsesClaimantRatherThanCaller(t *testing.T) {
	r, _, _, owner, _ := ownedResolverFixture(t)
	t.Chdir(t.TempDir())
	got, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", "")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || len(got.Sessions) != 1 || got.Sessions[0].CWD != owner.Worktree || got.Sessions[0].ID != "owner" || got.Snapshot.Engagements[0].Roots[0] != owner.Worktree {
		t.Fatalf("inputs=%+v", got)
	}
	if got.Scope.Head != got.Snapshot.CodeHead {
		t.Fatal("fresh scope does not pin code history")
	}
}

func TestOwnedActualQueueCoverageRefusesUnknownRecordsForCloseAndCorrection(t *testing.T) {
	for _, kind := range []string{"malformed", "unreadable", "missing queue", "missing slot"} {
		t.Run(kind, func(t *testing.T) {
			r, saved, _ := ownedCompletedFixture(t, false)
			common := strings.TrimSpace(testfix.Capture(t, r.root, "rev-parse", "--git-common-dir"))
			if !filepath.IsAbs(common) {
				common = filepath.Join(r.root, common)
			}
			dir := filepath.Join(common, "sdlc", "test-stats", "attempts")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "aabb.json")
			switch kind {
			case "malformed":
				if err := os.WriteFile(path, []byte(`{"broken":`), 0600); err != nil {
					t.Fatal(err)
				}
			case "unreadable":
				if err := os.Symlink(filepath.Join(dir, "missing-target"), path); err != nil {
					t.Fatal(err)
				}
			default:
				started := time.Date(2026, 10, 14, 0, 5, 0, 0, time.UTC)
				q := 0.0
				a := SuiteAttempt{Version: 1, ID: "aabb", Repo: strings.TrimPrefix(saved.Repository, "github.com/"), SHA: saved.CodeHead, Check: "suite", Slot: "ariadne:1", StartedAt: started, FinishedAt: started.Add(time.Minute), QueueSeconds: &q, LogSHA256: strings.Repeat("b", 64)}
				if kind == "missing queue" {
					a.QueueSeconds = nil
				} else {
					a.Slot = ""
				}
				raw, err := json.Marshal(a)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			for _, purpose := range []actualInputPurpose{actualCloseCandidate, actualCorrectionCandidate} {
				_, err := resolveOwnedActualInputsFor(context.Background(), r.root, "", "335", "", purpose)
				if err == nil || !strings.Contains(err.Error(), "queue evidence incomplete") {
					t.Fatalf("%s purpose %d adopted unknown queue coverage: %v", kind, purpose, err)
				}
			}
		})
	}
}

func TestOwnedActualLateLegacyAttemptRefusesCloseAndCorrection(t *testing.T) {
	r, saved, _ := ownedCompletedFixture(t, false)
	common := strings.TrimSpace(testfix.Capture(t, r.root, "rev-parse", "--git-common-dir"))
	if !filepath.IsAbs(common) {
		common = filepath.Join(r.root, common)
	}
	dir := filepath.Join(common, "sdlc", "test-stats", "attempts")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	until, err := time.Parse(time.RFC3339, saved.Until)
	if err != nil {
		t.Fatal(err)
	}
	queue := 1.0
	started := until.Add(5 * time.Second)
	a := SuiteAttempt{Version: 1, ID: "late-legacy", Repo: strings.TrimPrefix(saved.Repository, "github.com/"), SHA: saved.CodeHead, Check: "suite", Slot: "ariadne:1", StartedAt: started, FinishedAt: started.Add(time.Minute), QueueSeconds: &queue, LogSHA256: strings.Repeat("b", 64)}
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "late-legacy.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	for _, purpose := range []actualInputPurpose{actualCloseCandidate, actualCorrectionCandidate} {
		_, err := resolveOwnedActualInputsFor(context.Background(), r.root, "", "335", "", purpose)
		if err == nil || !strings.Contains(err.Error(), "queue evidence incomplete") {
			t.Fatalf("purpose %d adopted a late legacy attempt: %v", purpose, err)
		}
	}
}

func TestOwnedActualResolverReleaseReclaimSplitsEngagements(t *testing.T) {
	r, cardPath, _, owner, _ := ownedResolverFixture(t)
	ownedFixtureAt(t, 4)
	ownedFixtureCard(t, r, cardPath, "release-first", func(raw []byte) ([]byte, error) {
		return regexp.MustCompile(`(?m)^claimant:\n(?:[ \t].*\n)*`).ReplaceAll(raw, nil), nil
	})
	ownedFixtureAt(t, 6)
	ownedFixtureCard(t, r, cardPath, "claim-second", func(raw []byte) ([]byte, error) { return issue.SetCardClaimant(raw, owner) })
	got, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Snapshot.Engagements) != 2 || got.Snapshot.Engagements[0].End != "2026-10-14T00:04:00Z" || got.Snapshot.Engagements[1].Began != "2026-10-14T00:06:00Z" || len(got.Windows[got.Files[0]]) != 2 {
		t.Fatalf("engagements=%+v windows=%+v", got.Snapshot.Engagements, got.Windows)
	}
}

func ownedCompletedFixture(t *testing.T, mismatch bool) (*trackerRepo, gatestate.MeasurementInputs, string) {
	t.Helper()
	r, cardPath, _, _, source := ownedResolverFixture(t)
	inputs, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", "")
	if err != nil {
		t.Fatal(err)
	}
	saved := inputs.Snapshot
	head := saved.CodeHead
	if mismatch {
		saved.CodeHead = r.git("rev-parse", "HEAD^")
	}
	saved.Cuts = []gatestate.MeasurementCut{{Finalized: saved.Until, Dispatched: "2026-10-14T00:09:00Z", Finalizing: true}}
	ledger := gatestate.Ledger{Gate: "boundary-review", IssueNum: 335, IDPrefix: "BR", Rounds: []gatestate.Round{{N: 1, Timestamp: saved.Until, Dispatched: "2026-10-14T00:09:00Z", Reviewed: head, ActualInputs: &saved}}}
	writeRepoFile(t, r.root, "workshop/plans/000335-owned-close-gate.md", gatestate.Render(ledger, "fixture"))
	testfix.Git(t, r.root, "add", "workshop/plans/000335-owned-close-gate.md")
	ownedFixtureAt(t, 11)
	testfix.Git(t, r.root, "commit", "-qm", "#335: close evidence")
	evidence := r.git("rev-parse", "HEAD")
	ownedFixtureAt(t, 12)
	ownedFixtureCard(t, r, cardPath, "close-final", func(raw []byte) ([]byte, error) {
		raw, err := issue.SetCardCompletion(raw, issue.Completion{Token: "close-final", Repository: saved.Repository, ReviewedHEAD: head, EvidenceCommit: evidence})
		if err != nil {
			return nil, err
		}
		raw, err = issue.SetCardField(raw, "actual_hours", "0.5")
		if err != nil {
			return nil, err
		}
		return issue.SetCardField(raw, "status", "codecomplete")
	})
	return r, saved, source
}

func TestActualQueueCorrectionCandidateRequiresRecordedWait(t *testing.T) {
	r, saved, _ := ownedCompletedFixture(t, false)
	historical, err := resolveOwnedActualInputsFor(context.Background(), r.root, "", "335", "", actualHistorical)
	if err != nil || !historical.Replay || historical.Snapshot.CodeHead != saved.CodeHead {
		t.Fatalf("historical replay changed: %+v %v", historical, err)
	}
	_, err = resolveOwnedActualInputsFor(context.Background(), r.root, "", "335", "", actualCorrectionCandidate)
	if err == nil || !strings.Contains(err.Error(), "no complete, owned suite queue records") {
		t.Fatalf("candidate without records should refuse: %v", err)
	}
}

func TestOwnedActualResolverReplaysCompletionDespiteAppends(t *testing.T) {
	r, saved, source := ownedCompletedFixture(t, false)
	ownedFixtureAt(t, 20)
	testfix.Git(t, r.root, "commit", "--allow-empty", "-qm", "#335: later unrelated edit")
	f, err := os.OpenFile(source, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString("{\"timestamp\":\"2026-10-14T00:30:00Z\",\"type\":\"event_msg\",\"payload\":{\"text\":\"later work\"}}\n")
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	got, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", "")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Replay || got.Snapshot.CodeHead != saved.CodeHead || got.Snapshot.TrackerHead != saved.TrackerHead || got.Snapshot.Until != saved.Until || got.Prefixes[source].Bytes != saved.Sources[0].Bytes || got.Scope.Head != saved.CodeHead || len(got.Ledger.Rounds) != 1 {
		t.Fatalf("replay=%+v", got)
	}
}

func TestOwnedActualResolverRejectsCompletionSnapshotMismatch(t *testing.T) {
	r, _, _ := ownedCompletedFixture(t, true)
	if _, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", ""); err == nil {
		t.Fatal("snapshot for a different code generation accepted")
	}
}

func TestOwnedActualResolverReplaysFinalizedMilestoneOnActiveIssue(t *testing.T) {
	r, _, _, _, _ := ownedResolverFixture(t)
	inputs, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", "")
	if err != nil {
		t.Fatal(err)
	}
	saved := inputs.Snapshot
	saved.Cuts = []gatestate.MeasurementCut{{Boundary: "M1", Finalized: saved.Until, Dispatched: "2026-10-14T00:09:00Z", Finalizing: true}}
	ledger := gatestate.Ledger{Gate: "boundary-review", IssueNum: 335, IDPrefix: "BR", Rounds: []gatestate.Round{{N: 1, Boundary: "M1", Timestamp: saved.Until, Dispatched: "2026-10-14T00:09:00Z", Reviewed: saved.CodeHead, ActualInputs: &saved}}}
	writeRepoFile(t, r.root, "workshop/plans/000335-owned-close-gate.md", gatestate.Render(ledger, "fixture"))
	testfix.Git(t, r.root, "add", "workshop/plans/000335-owned-close-gate.md")
	ownedFixtureAt(t, 11)
	testfix.Git(t, r.root, "commit", "-qm", "#335 M1: close evidence")
	ownedFixtureAt(t, 20)
	testfix.Git(t, r.root, "commit", "--allow-empty", "-qm", "#335: next milestone work")
	got, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", "", "M1")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Replay || got.Snapshot.CodeHead != saved.CodeHead || got.Snapshot.Until != saved.Until || got.Scope.Head != saved.CodeHead {
		t.Fatalf("milestone replay=%+v", got)
	}
	aggregate, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", "")
	if err != nil {
		t.Fatal(err)
	}
	if aggregate.Replay || aggregate.Snapshot.CodeHead != r.git("rev-parse", "HEAD") {
		t.Fatalf("whole issue reused old milestone: %+v", aggregate)
	}
}

func TestOwnedReviewWindowRequiresExactConcreteEvidence(t *testing.T) {
	for _, tc := range []struct {
		name, message, base, head string
		wantErr                   bool
	}{
		{"valid", "Closed.\n\nReview-Verdict: SHIP\nReview-Window: abc12345..def67890\n", "abc12345", "def67890", false},
		{"missing", "Closed.", "", "", true},
		{"ambient head", "Review-Window: abc12345..HEAD", "", "", true},
		{"duplicate", "Review-Window: abc12345..def67890\nReview-Window: abc12345..def67890", "", "", true},
		{"malformed", "Review-Window: definitely..invalid", "", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base, head, err := ownedReviewWindow(tc.message)
			if (err != nil) != tc.wantErr || base != tc.base || head != tc.head {
				t.Fatalf("window=%s..%s err=%v", base, head, err)
			}
		})
	}
}

func ownedLegacyCompletionFixture(t *testing.T, multiple bool) (*trackerRepo, string, string) {
	t.Helper()
	r, cardPath, _, owner, _ := ownedResolverFixture(t)
	head := r.git("rev-parse", "HEAD")
	base := r.git("rev-parse", "HEAD^")
	ledger := gatestate.Ledger{Gate: "boundary-review", IssueNum: 335, IDPrefix: "BR", Rounds: []gatestate.Round{{N: 1, Timestamp: "2026-10-14T00:10:00Z", Reviewed: head}}}
	if multiple {
		ledger.Rounds = append([]gatestate.Round{{N: 1, Boundary: "M1", Timestamp: "2026-10-14T00:08:00Z", Reviewed: base}}, ledger.Rounds...)
	}
	writeRepoFile(t, r.root, "workshop/plans/000335-owned-close-gate.md", gatestate.Render(ledger, "fixture"))
	testfix.Git(t, r.root, "add", "workshop/plans/000335-owned-close-gate.md")
	ownedFixtureAt(t, 11)
	testfix.Git(t, r.root, "commit", "-qm", fmt.Sprintf("#335: close evidence\n\nReview-Verdict: SHIP\nReview-Window: %s..%s", base[:8], head[:8]))
	evidence := r.git("rev-parse", "HEAD")
	ownedFixtureAt(t, 12)
	ownedFixtureCard(t, r, cardPath, "close-final", func(raw []byte) ([]byte, error) {
		raw, err := issue.SetCardCompletion(raw, issue.Completion{Token: "close-final", Repository: owner.Repository, ReviewedHEAD: head, EvidenceCommit: evidence})
		if err != nil {
			return nil, err
		}
		raw, err = issue.SetCardField(raw, "actual_hours", "0.5")
		if err != nil {
			return nil, err
		}
		return issue.SetCardField(raw, "status", "codecomplete")
	})
	ownedFixtureAt(t, 20)
	testfix.Git(t, r.root, "commit", "--allow-empty", "-qm", "#999: later main work")
	return r, base, head
}

func TestOwnedActualResolverReconstructsHistoricalScopeFromEvidence(t *testing.T) {
	r, base, head := ownedLegacyCompletionFixture(t, false)
	if currentBase := r.git("merge-base", "main", head); currentBase != head {
		t.Fatal("fixture did not advance main beyond old reviewed head")
	}
	got, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Scope.BranchPoint != base || got.Scope.Head != head || got.Snapshot.BranchPoint != base {
		t.Fatalf("historical scope widened to current main: %+v", got.Scope)
	}
}

func TestOwnedActualResolverRefusesUnpinnedMultiBoundaryHistory(t *testing.T) {
	r, _, _ := ownedLegacyCompletionFixture(t, true)
	if _, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", ""); err == nil {
		t.Fatal("incomplete multi-boundary historical scope guessed")
	}
}

func ownedFixtureCorrection(t *testing.T, r *trackerRepo, saved gatestate.MeasurementInputs, exclude, badBinding bool) {
	t.Helper()
	saved.GapMinutes = 7
	raw, err := json.Marshal(saved)
	if err != nil {
		t.Fatal(err)
	}
	ownedFixtureAt(t, 15)
	ownedFixtureCard(t, r, tracker.CardPath("000335", "owned"), "correction-final", func(card []byte) ([]byte, error) {
		completion, ok, err := issue.CardCompletion(card)
		if err != nil || !ok {
			t.Fatal("missing completion fixture")
		}
		newActual := "0.7"
		if exclude {
			newActual = "N/A"
		}
		record := issue.ActualCorrection{Version: 1, Token: "correction-final", CompletionToken: completion.Token, Repository: completion.Repository, EvidenceCommit: completion.EvidenceCommit, ReviewedHEAD: completion.ReviewedHEAD, PreviousActual: "0.5", NewActual: newActual, Reason: "fixture correction", MeasurementDigest: correctionDigest(raw), Inputs: string(raw), Exclude: exclude}
		card, err = issue.SetCardActualCorrection(card, record)
		if err != nil {
			return nil, err
		}
		card, err = issue.SetCardField(card, "actual_hours", newActual)
		if err != nil {
			return nil, err
		}
		if badBinding {
			card = bytesReplaceCorrectionBinding(card)
		}
		return card, nil
	})
}

func bytesReplaceCorrectionBinding(card []byte) []byte {
	return []byte(strings.Replace(string(card), "completion_token: close-final", "completion_token: unrelated-generation", 1))
}

func TestOwnedActualResolverPrefersCardCorrectionInputs(t *testing.T) {
	r, saved, _ := ownedCompletedFixture(t, false)
	ownedFixtureCorrection(t, r, saved, false, false)
	got, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", "")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Replay || got.Snapshot.GapMinutes != 7 {
		t.Fatalf("card correction ignored: %+v", got.Snapshot)
	}
}

func TestOwnedActualResolverRefusesWrongCorrectionBinding(t *testing.T) {
	r, saved, _ := ownedCompletedFixture(t, false)
	ownedFixtureCorrection(t, r, saved, false, true)
	if _, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", ""); err == nil {
		t.Fatal("correction for unrelated completion accepted")
	}
}

func TestOwnedActualResolverDoesNotAdoptExcludedCorrection(t *testing.T) {
	r, saved, _ := ownedCompletedFixture(t, false)
	ownedFixtureCorrection(t, r, saved, true, false)
	if _, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", ""); err == nil || !strings.Contains(err.Error(), "excluded") {
		t.Fatalf("excluded sample diagnostic missing: %v", err)
	}
}

func TestOwnedActualResolverReconstructsProvenMainIntegrationScope(t *testing.T) {
	r, cardPath, _, owner, _ := ownedResolverFixture(t)
	branch := "000335-integrated"
	testfix.Git(t, r.root, "switch", "-q", "-c", branch)
	ownedFixtureAt(t, 12)
	testfix.Git(t, r.root, "commit", "--allow-empty", "-qm", "#335 M1: milestone work")
	testfix.Git(t, r.root, "switch", "-q", "main")
	ownedFixtureAt(t, 13)
	testfix.Git(t, r.root, "commit", "--allow-empty", "-qm", "#999: concurrent main work")
	base := r.git("rev-parse", "HEAD")
	testfix.Git(t, r.root, "push", "-q", "origin", "main")
	testfix.Git(t, r.root, "switch", "-q", branch)
	ownedFixtureAt(t, 14)
	testfix.Git(t, r.root, "merge", "--no-ff", "-qm", "Merge remote-tracking branch 'origin/main' into "+branch, "origin/main")
	head := r.git("rev-parse", "HEAD")
	ledger := gatestate.Ledger{Gate: "boundary-review", IssueNum: 335, IDPrefix: "BR", Rounds: []gatestate.Round{{N: 1, Boundary: "M1", Timestamp: "2026-10-14T00:12:00Z", Reviewed: r.git("rev-parse", "HEAD^1")}, {N: 2, Timestamp: "2026-10-14T00:14:00Z", Reviewed: head}}}
	writeRepoFile(t, r.root, "workshop/plans/000335-owned-close-gate.md", gatestate.Render(ledger, "fixture"))
	testfix.Git(t, r.root, "add", "workshop/plans/000335-owned-close-gate.md")
	ownedFixtureAt(t, 15)
	testfix.Git(t, r.root, "commit", "-qm", fmt.Sprintf("#335: close evidence\n\nReview-Verdict: SHIP\nReview-Window: %s..%s", base[:8], head[:8]))
	evidence := r.git("rev-parse", "HEAD")
	ownedFixtureAt(t, 16)
	ownedFixtureCard(t, r, cardPath, "close-final", func(raw []byte) ([]byte, error) {
		raw, err := issue.SetCardCompletion(raw, issue.Completion{Token: "close-final", Repository: owner.Repository, ReviewedHEAD: head, EvidenceCommit: evidence})
		if err != nil {
			return nil, err
		}
		raw, err = issue.SetCardField(raw, "actual_hours", "0.5")
		if err != nil {
			return nil, err
		}
		return issue.SetCardField(raw, "status", "codecomplete")
	})
	got, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Scope.BranchPoint != base || got.Scope.Head != head {
		t.Fatalf("recorded main integration scope=%+v", got.Scope)
	}
}

// A new production close retires the prior generation's repair while a landing
// of that generation preserves it. Both replacement and exclusion are covered.
func TestOwnedActualResolverRecloseRetiresPriorCorrection(t *testing.T) {
	for _, excluded := range []bool{false, true} {
		t.Run(fmt.Sprintf("exclude=%v", excluded), func(t *testing.T) {
			r, saved, _ := ownedCompletedFixture(t, false)
			ownedFixtureCorrection(t, r, saved, excluded, false)
			cardPath := tracker.CardPath("000335", "owned")
			ownedFixtureAt(t, 16)
			ownedFixtureCard(t, r, cardPath, "land-same-generation", func(raw []byte) ([]byte, error) {
				prior, ok, err := issue.CardActualCorrection(raw)
				if err != nil || !ok {
					t.Fatal("missing correction")
				}
				c, _, err := issue.CardCompletion(raw)
				if err != nil {
					return nil, err
				}
				c.LandedCommit = c.EvidenceCommit
				next, err := issue.SetCardCompletion(raw, c)
				if err != nil {
					return nil, err
				}
				retained, ok, err := issue.CardActualCorrection(next)
				if err != nil || !ok || retained.Token != prior.Token || retained.MeasurementDigest != prior.MeasurementDigest {
					t.Fatal("landing retired same-generation correction")
				}
				return next, nil
			})
			got, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", "")
			if excluded {
				if err == nil || !strings.Contains(err.Error(), "explicitly excluded") {
					t.Fatalf("landing lost exclusion: %v", err)
				}
			} else if err != nil || got.Snapshot.GapMinutes != 7 {
				t.Fatalf("landing lost corrected inputs: %+v %v", got, err)
			}
			ownedFixtureAt(t, 17)
			ownedFixtureCard(t, r, cardPath, "close-next-generation", func(raw []byte) ([]byte, error) {
				c, _, err := issue.CardCompletion(raw)
				if err != nil {
					return nil, err
				}
				return tracker.CodecompleteCard(raw, tracker.ReceiptSpec{Token: "close-next", IssueID: "000335", Repository: saved.Repository, ReviewedHEAD: saved.CodeHead, EvidenceMessage: "#335: new close\n\nClose-Actual: 0.5"}, c.EvidenceCommit, "2026-10-14")
			})
			got, err = resolveOwnedActualInputs(context.Background(), r.root, "", "335", "")
			if err != nil {
				t.Fatalf("new completion retained stale correction: %v", err)
			}
			if !got.Replay || got.Snapshot.GapMinutes != saved.GapMinutes {
				t.Fatalf("new close did not restore its own snapshot: %+v", got.Snapshot)
			}
		})
	}
}
