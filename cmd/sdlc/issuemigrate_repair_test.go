package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

func legacyRepairRepo(t *testing.T, close bool) (*trackerRepo, string) {
	t.Helper()
	f := newLegacyProofFixture(t)
	f.archive = "workshop/history/issues/000001-example.md"
	if err := os.MkdirAll(filepath.Join(f.root, filepath.Dir(f.archive)), 0755); err != nil {
		t.Fatal(err)
	}
	closeOID := ""
	if close {
		closeOID = f.close(true)
	}
	f.terminal()
	origin := filepath.Join(t.TempDir(), "remote.git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	testfix.Git(t, f.root, "remote", "add", "origin", origin)
	testfix.Git(t, f.root, "push", "-q", "-u", "origin", "main")
	t.Chdir(f.root)
	return &trackerRepo{t: t, root: f.root, origin: origin}, closeOID
}

type lostReplyLegacyPublisher struct{ *tracker.Repository }

func (p lostReplyLegacyPublisher) UpdateCardWithTrailers(r tracker.Record, raw []byte, token string, trailers []string, before func(string, string) error) error {
	if err := p.Repository.UpdateCardWithTrailers(r, raw, token, trailers, before); err != nil {
		return err
	}
	return fmt.Errorf("%w: fixture lost push response", gitx.ErrPublicationUncertain)
}

func TestLegacyCompletionRepairPublicationFences(t *testing.T) {
	for _, mode := range []string{"stale-card", "lost-reply", "late-cancel"} {
		t.Run(mode, func(t *testing.T) {
			r, _ := oldMigratedLegacy(t)
			env, err := openMigrate(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			plan, err := prepareLegacyCompletionRepair(env, "000001")
			if err != nil {
				t.Fatal(err)
			}
			before := r.git("ls-remote", "origin", "refs/heads/issue-tracker")
			var writer legacyRepairPublisher = env.tracker
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			observer := func(string, string) error { return nil }
			switch mode {
			case "stale-card":
				peer := []byte(strings.Replace(string(plan.Current.Raw), "---\n\n#", "future_peer: kept\n---\n\n#", 1))
				if err := env.tracker.UpdateCard(plan.Current, peer, "peer-change", func(string, string) error { return nil }); err != nil {
					t.Fatal(err)
				}
			case "lost-reply":
				writer = lostReplyLegacyPublisher{env.tracker}
			case "late-cancel":
				observer = func(string, string) error { cancel(); return nil }
			}
			err = applyLegacyCompletionRepair(ctx, writer, plan, observer)
			switch mode {
			case "stale-card":
				if !errors.Is(err, tracker.ErrCardChanged) {
					t.Fatalf("stale write: %v", err)
				}
				if got := r.card(plan.Current.Path); !strings.Contains(got, "future_peer: kept") || strings.Contains(got, "completion:") {
					t.Fatalf("peer overwritten: %s", got)
				}
			case "lost-reply":
				if !errors.Is(err, gitx.ErrPublicationUncertain) {
					t.Fatalf("lost uncertainty: %v", err)
				}
				next, err := prepareLegacyCompletionRepair(env, "000001")
				if err != nil || !next.Complete {
					t.Fatalf("observe lost reply: complete=%v err=%v", next.Complete, err)
				}
				confirmed := r.git("ls-remote", "origin", "refs/heads/issue-tracker")
				if err := applyLegacyCompletionRepair(t.Context(), env.tracker, next, func(string, string) error { return nil }); err != nil {
					t.Fatal(err)
				}
				if r.git("ls-remote", "origin", "refs/heads/issue-tracker") != confirmed {
					t.Fatal("repeat published twice")
				}
			case "late-cancel":
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("ignored observer cancellation: %v", err)
				}
				if r.git("ls-remote", "origin", "refs/heads/issue-tracker") != before {
					t.Fatal("cancelled repair published")
				}
			}
		})
	}
}

func TestIssueMigratePreservesArchivedLegacyClose(t *testing.T) {
	r, closeOID := legacyRepairRepo(t, true)
	digest, out, err := migrateDryRun(t)
	if err != nil {
		t.Fatalf("dry run: %v %s", err, out)
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--apply", "--expect", digest); err != nil {
		t.Fatalf("apply: %v %s", err, stderr)
	}
	raw := []byte(r.card(tracker.CardPath("000001", "example")))
	c, ok, err := issue.CardCompletion(raw)
	if err != nil || !ok || c.EvidenceCommit != closeOID {
		t.Fatalf("archived close lost: %+v %v %s", c, err, raw)
	}
	a, ok, err := issue.CardLegacyAdoption(raw)
	if err != nil || !ok || a.WindowClass != issue.LegacyWindowlessClose {
		t.Fatalf("adoption proof lost: %+v %v", a, err)
	}
	r.git("pull", "-q")
	stdout, stderr, err := executeSDLCTestCommand("issue", "show", "1")
	if err != nil || !strings.Contains(stdout, "windowless-close") {
		t.Fatalf("show lost class: %v %s %s", err, stdout, stderr)
	}
}

func TestLegacyMigrationSourcePinsActualCutover(t *testing.T) {
	r, _ := legacyRepairRepo(t, false)
	if _, _, err := legacyMigrationSource(t.Context(), r.root, "HEAD"); err == nil {
		t.Fatal("accepted missing cutover")
	}
	digest, _, _ := migrateDryRun(t)
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--apply", "--expect", digest); err != nil {
		t.Fatalf("apply: %v %s", err, stderr)
	}
	r.git("pull", "-q")
	source, root, err := legacyMigrationSource(context.Background(), r.root, "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if want := r.git("rev-parse", "HEAD^1"); source != want {
		t.Fatalf("source=%s want=%s", source, want)
	}
	if marker := r.git("show", "HEAD:"+tracker.CutoverMarkerPath); !strings.Contains(marker, root) {
		t.Fatal("wrong root")
	}
	raw, err := os.ReadFile(filepath.Join(r.root, tracker.CutoverMarkerPath))
	if err != nil {
		t.Fatal(err)
	}
	r.git("rm", tracker.CutoverMarkerPath)
	r.git("commit", "-qm", "remove marker")
	writeRepoFile(t, r.root, tracker.CutoverMarkerPath, string(raw))
	r.git("add", tracker.CutoverMarkerPath)
	r.git("commit", "-qm", "readd marker")
	if _, _, err := legacyMigrationSource(t.Context(), r.root, "HEAD"); err == nil {
		t.Fatal("accepted ambiguous cutover")
	}
}

// oldMigratedLegacy reproduces the pre-fix cutover using the same planner,
// deliberately omitting the new inventory evidence before bootstrap.
func oldMigratedLegacy(t *testing.T) (*trackerRepo, string) {
	return oldMigratedLegacyWithClose(t, true)
}

func oldMigratedLegacyWithClose(t *testing.T, close bool) (*trackerRepo, string) {
	t.Helper()
	r, closeOID := legacyRepairRepo(t, close)
	env, err := openMigrate(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	view, err := env.main.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	in, err := migrationInventory(env, view)
	if err != nil {
		t.Fatal(err)
	}
	in.Legacy = nil
	m := tracker.PlanTrackerMigration(in)
	root, err := applyTrackerBootstrap(env, m)
	if err != nil {
		t.Fatal(err)
	}
	if err := applyMainCutover(env, m, root); err != nil {
		t.Fatal(err)
	}
	r.git("pull", "-q")
	return r, closeOID
}

func TestIssueShowNamesReadOnlyRecoveredLegacyClass(t *testing.T) {
	r, _ := oldMigratedLegacyWithClose(t, false)
	cardPath := tracker.CardPath("000001", "example")
	before := r.card(cardPath)
	ref := r.git("ls-remote", "origin", "refs/heads/issue-tracker")
	stdout, stderr, err := executeSDLCTestCommand("issue", "show", "1")
	if err != nil || !strings.Contains(stdout, "windowless-terminal") || !strings.Contains(stdout, "read-only") {
		t.Fatalf("missing recovered class: %v %s %s", err, stdout, stderr)
	}
	if r.card(cardPath) != before || r.git("ls-remote", "origin", "refs/heads/issue-tracker") != ref {
		t.Fatal("observation changed imported card")
	}
}

func TestIssueMigrateRepairsMissingCompletionWithoutCheckoutChanges(t *testing.T) {
	r, closeOID := oldMigratedLegacy(t)
	main, head, status := r.originMain(), r.git("rev-parse", "HEAD"), r.git("status", "--porcelain")
	stdout, stderr, err := executeSDLCTestCommand("issue", "migrate", "--repair-completions", "--issue", "1")
	if err != nil {
		t.Fatalf("repair dry: %v %s %s", err, stdout, stderr)
	}
	m := digestLine.FindStringSubmatch(stdout)
	if len(m) != 2 {
		t.Fatalf("no repair digest: %s", stdout)
	}
	cardPath := tracker.CardPath("000001", "example")
	before := r.card(cardPath)
	if _, ok, _ := issue.CardCompletion([]byte(before)); ok {
		t.Fatal("dry run wrote completion")
	}
	if _, _, err := executeSDLCTestCommand("issue", "migrate", "--repair-completions", "--issue", "1", "--apply", "--expect", "0000000000000000"); err == nil {
		t.Fatal("wrong digest applied")
	}
	_, stderr, err = executeSDLCTestCommand("issue", "migrate", "--repair-completions", "--issue", "1", "--apply", "--expect", m[1])
	if err != nil {
		t.Fatalf("repair apply: %v %s", err, stderr)
	}
	after := r.card(cardPath)
	c, ok, err := issue.CardCompletion([]byte(after))
	if err != nil || !ok || c.EvidenceCommit != closeOID {
		t.Fatalf("repair failed: %+v %v", c, err)
	}
	trackerBefore := r.git("ls-remote", "origin", "refs/heads/issue-tracker")
	stdout, stderr, err = executeSDLCTestCommand("issue", "migrate", "--repair-completions", "--issue", "1", "--apply", "--expect", m[1])
	if err != nil || !strings.Contains(stdout+stderr, "already") {
		t.Fatalf("repeat: %v %s %s", err, stdout, stderr)
	}
	if got := r.git("ls-remote", "origin", "refs/heads/issue-tracker"); got != trackerBefore {
		t.Fatal("repeat created a commit")
	}
	if r.originMain() != main || r.git("rev-parse", "HEAD") != head || r.git("status", "--porcelain") != status {
		t.Fatal("repair changed main/checkout")
	}
}

// Repair must use the same monotone authority fence as actual exclusion. An
// endpoint restored to identical bytes does not undo intervening evidence loss.
func TestLegacyCompletionRepairRejectsRestoredAuthority(t *testing.T) {
	for _, mode := range []string{"remove-completion-and-provenance", "remove-provenance", "replace-completion", "replace-provenance"} {
		t.Run(mode, func(t *testing.T) {
			r, _ := oldMigratedLegacy(t)
			env, err := openMigrate(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			plan, err := prepareLegacyCompletionRepair(env, "000001")
			if err != nil {
				t.Fatal(err)
			}
			original := append([]byte(nil), plan.Current.Raw...)
			if err := applyLegacyCompletionRepair(t.Context(), env.tracker, plan, func(string, string) error { return nil }); err != nil {
				t.Fatal(err)
			}
			snapshot, err := env.tracker.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			repaired, err := snapshot.Require("000001")
			if err != nil {
				t.Fatal(err)
			}
			end, err := verifyLegacyAdoption(t.Context(), r.root, snapshot.Ref(), tracker.IssueRecord{ID: repaired.ID, Card: &repaired}, env.target.Repository, plansDir())
			if err != nil || end.IsZero() {
				t.Fatalf("monotone repair unusable by actual: %v %v", end, err)
			}
			var changed []byte
			switch mode {
			case "remove-completion-and-provenance":
				changed = original
			case "remove-provenance":
				changed, err = issue.SetCardCompletion(original, plan.Proof.Completion)
			case "replace-completion":
				completion := plan.Proof.Completion
				completion.Token = "different-close"
				changed, err = issue.SetCardCompletion(repaired.Raw, completion)
			case "replace-provenance":
				adoption := plan.Proof.Adoption
				adoption.SourceBlob = strings.Repeat("1", 40)
				changed, err = issue.SetCardLegacyAdoption(repaired.Raw, adoption)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := env.tracker.UpdateCard(repaired, changed, "change-authority", func(string, string) error { return nil }); err != nil {
				t.Fatal(err)
			}
			// The exact originally-unbound restoration reproduces the review finding;
			// the other cases restore repaired bytes to exercise hidden history changes.
			if mode != "remove-completion-and-provenance" {
				snapshot, err = env.tracker.Snapshot()
				if err != nil {
					t.Fatal(err)
				}
				current, err := snapshot.Require("000001")
				if err != nil {
					t.Fatal(err)
				}
				if err := env.tracker.UpdateCard(current, repaired.Raw, "restore-authority", func(string, string) error { return nil }); err != nil {
					t.Fatal(err)
				}
			}
			snapshot, err = env.tracker.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			current, err := snapshot.Require("000001")
			if err != nil {
				t.Fatal(err)
			}
			end, err = verifyLegacyAdoption(t.Context(), r.root, snapshot.Ref(), tracker.IssueRecord{ID: current.ID, Card: &current}, env.target.Repository, plansDir())
			if !end.IsZero() {
				t.Fatalf("actual granted bound for non-monotone %s history: %v %v", mode, end, err)
			}
			if mode != "remove-completion-and-provenance" && (err == nil || !strings.Contains(err.Error(), "tracker history")) {
				t.Fatalf("actual did not reject restored %s authority: %v", mode, err)
			}
			before := r.git("ls-remote", "origin", "refs/heads/issue-tracker")
			if next, err := prepareLegacyCompletionRepair(env, "000001"); err == nil || !strings.Contains(err.Error(), "tracker history") {
				t.Fatalf("repair accepted non-monotone %s history (already repaired=%v): %v", mode, next.Complete, err)
			}
			if after := r.git("ls-remote", "origin", "refs/heads/issue-tracker"); after != before {
				t.Fatal("refused repair changed tracker")
			}
		})
	}
}
