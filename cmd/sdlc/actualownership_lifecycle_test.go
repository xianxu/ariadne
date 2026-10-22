package main

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// Unlike legacySpanningFixture, this crosses a real pre-cutover issue branch
// and acquires ownership through the public claim command. Reconciliation must
// not manufacture ownership for the intervening work.
func ownershipRecoveryFixture(t *testing.T) (*trackerRepo, string, string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	ownedFixtureAt(t, 0)
	root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	origin := filepath.Join(t.TempDir(), "origin.git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	testfix.Git(t, root, "remote", "add", "origin", origin)
	r := &trackerRepo{t: t, root: root, origin: origin}
	detail := "workshop/issues/000335-spanning.md"
	writeRepoFile(t, root, detail, "---\nid: 000335\nstatus: working\nstarted: 2026-10-14T00:01:00Z\ncreated: 2026-10-14\n---\n\n# Spanning migration\n\n## Problem\n\nPreserve legacy work and exclude unowned work.\n")
	ownedFixtureAt(t, 1)
	r.git("add", detail)
	r.git("commit", "-qm", "#335: legacy started work")
	ownedFixtureAt(t, 5)
	r.git("commit", "--allow-empty", "-qm", "#335: legacy implementation")
	r.git("push", "-q", "-u", "origin", "main")
	worker := filepath.Join(t.TempDir(), "worker")
	testfix.Git(t, root, "worktree", "add", "-q", "-b", "000335-spanning", worker)
	ownedFixtureAt(t, 7)
	testfix.Git(t, worker, "commit", "--allow-empty", "-qm", "#335: pre-cutover branch work")
	ownedFixtureAt(t, 10)
	digest, out, err := migrateDryRun(t)
	if err != nil {
		t.Fatalf("migration dry run: %v %s", err, out)
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--apply", "--expect", digest); err != nil {
		t.Fatalf("migration apply: %v %s", err, stderr)
	}
	r.git("pull", "-q")
	t.Chdir(worker)
	ownedFixtureAt(t, 12)
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--reconcile"); err != nil {
		t.Fatalf("branch reconcile: %v %s", err, stderr)
	}
	if !issue.HasMirror([]byte(testfix.Git(t, worker, "show", "HEAD:"+detail))) {
		t.Fatal("reconcile did not bring the branch's details across")
	}
	ownedFixtureAt(t, 18)
	testfix.Git(t, worker, "commit", "--allow-empty", "-qm", "#335: unowned post-cutover implementation")
	cardPath := tracker.CardPath("000335", "spanning")
	if _, has, err := issue.CardClaimant([]byte(r.card(cardPath))); err != nil || has {
		t.Fatalf("reconcile invented ownership: has=%v err=%v", has, err)
	}
	return r, worker, cardPath
}

func ownershipNativeSource(t *testing.T, harness, id, root string, minutes ...int) string {
	t.Helper()
	var name, data string
	switch harness {
	case "codex":
		name = filepath.Join(os.Getenv("HOME"), ".codex", "sessions", id+".jsonl")
		data = fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:00:00Z\",\"type\":\"session_meta\",\"payload\":{\"id\":%q,\"cwd\":%q}}\n", id, root)
		for _, minute := range minutes {
			data += fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:%02d:00Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":\"#335 development\"}}\n", minute)
		}
	case "claude":
		project := strings.NewReplacer("/", "-", ".", "-").Replace(root)
		name = filepath.Join(os.Getenv("HOME"), ".claude", "projects", project, id+".jsonl")
		for _, minute := range minutes {
			data += fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:%02d:00Z\",\"type\":\"user\",\"sessionId\":%q,\"cwd\":%q,\"message\":{\"content\":\"#335 development\"}}\n", minute, id, root)
		}
	default:
		t.Fatalf("unsupported native fixture harness %q", harness)
	}
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestActualOwnershipRealMigrationClaimAndReplay(t *testing.T) {
	for _, harness := range []string{"codex", "claude"} {
		t.Run(harness, func(t *testing.T) {
			r, worker, cardPath := ownershipRecoveryFixture(t)
			ownershipNativeSource(t, harness, "legacy-prefix", r.root, 2, 4, 6)
			ownershipNativeSource(t, harness, "worker", worker, 13, 15, 17)
			for _, purpose := range []actualInputPurpose{actualHistorical, actualCloseCandidate} {
				refused := computeActualForPurpose(t.Context(), worker, "", "335", "", "workshop/plans", purpose)
				var rendered bytes.Buffer
				printActual(&rendered, refused, "")
				want := "no owner since cutover: run sdlc claim --issue 335 from the slot doing the work; work between cutover and claim is unowned and not counted"
				if refused.Status != actualError || !strings.Contains(rendered.String(), want) {
					t.Fatalf("purpose %d lacks exact owner recovery: %+v\n%s", purpose, refused, rendered.String())
				}
				for _, wrong := range []string{"restore the named", "migrate --reconcile"} {
					if strings.Contains(rendered.String(), wrong) {
						t.Fatalf("purpose %d repeats unsafe recovery %q: %s", purpose, wrong, rendered.String())
					}
				}
			}
			ownedFixtureAt(t, 20)
			if _, stderr, err := executeSDLCTestCommand("claim", "--issue", "335"); err != nil {
				t.Fatalf("real claim after reconcile: %v %s", err, stderr)
			}
			owner, has, err := issue.CardClaimant([]byte(r.card(cardPath)))
			if err != nil || !has || owner.Worktree != canonRoot(worker) {
				t.Fatalf("real claim did not establish working-slot owner: %+v %v %v", owner, has, err)
			}
			ownershipNativeSource(t, harness, "worker", worker, 22, 25, 30)
			ownedFixtureAt(t, 30)
			testfix.Git(t, worker, "commit", "--allow-empty", "-qm", "#335: genuine owned implementation")
			base := computeActualForPurpose(t.Context(), worker, "", "335", "", "workshop/plans", actualCloseCandidate)
			assertOwnershipGap(t, base)
			// Pre-claim events can supply continuity into the owned 20→22 interval;
			// that existing boundary behavior can change totals. The independent
			// contract is that no credited activity enters the unowned 10→20 gap.
			ownershipNativeSource(t, harness, "legacy-prefix", r.root, 2, 4, 6, 11, 13, 15, 17, 19)
			workerSource := ownershipNativeSource(t, harness, "worker", worker, 11, 13, 15, 17, 19, 22, 25, 30)
			measured := computeActualForPurpose(t.Context(), worker, "", "335", "", "workshop/plans", actualCloseCandidate)
			assertOwnershipGap(t, measured)
			saved := *measured.Inputs
			if saved.LegacyPrefix == nil || saved.LegacyPrefix.Until != "2026-10-14T00:10:00Z" || len(saved.Engagements) != 1 || saved.Engagements[0].Start != "2026-10-14T00:20:00Z" {
				t.Fatalf("wrong real migration/claim bounds: %+v", saved)
			}
			ledger := gatestate.Ledger{Gate: "boundary-review", IssueNum: 335, IDPrefix: "BR", Rounds: []gatestate.Round{{N: 1, Timestamp: "2026-10-14T00:31:00Z", Reviewed: saved.CodeHead, ActualInputs: &saved}}}
			writeRepoFile(t, worker, "workshop/plans/000335-spanning-close-gate.md", gatestate.Render(ledger, "fixture"))
			testfix.Git(t, worker, "add", "workshop/plans/000335-spanning-close-gate.md")
			ownedFixtureAt(t, 31)
			testfix.Git(t, worker, "commit", "-qm", "#335: close owned recovery measurement")
			evidence := strings.TrimSpace(testfix.Git(t, worker, "rev-parse", "HEAD"))
			ownedFixtureAt(t, 32)
			ownedFixtureCard(t, r, cardPath, "close-recovered", func(raw []byte) ([]byte, error) {
				return issue.SetCardCompletion(raw, issue.Completion{Token: "close-recovered", Repository: saved.Repository, ReviewedHEAD: saved.CodeHead, EvidenceCommit: evidence})
			})
			ownedFixtureAt(t, 35)
			ownedFixtureCard(t, r, cardPath, "rename-after-close", func(raw []byte) ([]byte, error) {
				return issue.SetCardTitle(raw, "Later card edit")
			})
			ownedFixtureAt(t, 38)
			testfix.Git(t, worker, "commit", "--allow-empty", "-qm", "#335: later work outside completed generation")
			// Append without rewriting the captured prefix: replay must use its
			// saved byte boundary even for a newly appended in-window event.
			file, err := os.OpenFile(workerSource, os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				t.Fatal(err)
			}
			row := "{\"timestamp\":\"2026-10-14T00:29:00Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":\"#335 appended after completion\"}}\n"
			if harness == "claude" {
				row = fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:29:00Z\",\"type\":\"user\",\"sessionId\":\"worker\",\"cwd\":%q,\"message\":{\"content\":\"#335 appended after completion\"}}\n", worker)
			}
			if _, err := file.WriteString(row); err != nil {
				_ = file.Close()
				t.Fatal(err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
			ownershipNativeSource(t, harness, "later-source", worker, 37, 38)
			replayed := computeActualForPurpose(t.Context(), worker, "", "335", "", "workshop/plans", actualHistorical)
			assertOwnershipGap(t, replayed)
			if replayed.Hours != measured.Hours || replayed.Inputs.CodeHead != saved.CodeHead || replayed.Inputs.TrackerHead != saved.TrackerHead || replayed.Inputs.Until != saved.Until {
				t.Fatalf("finalized measurement authority changed: saved=%+v replay=%+v", measured, replayed)
			}
			// Corrections still require fresh complete queue evidence; a claim
			// diagnostic must not replace this historical-generation refusal.
			correction := computeActualForPurpose(t.Context(), worker, "", "335", "", "workshop/plans", actualCorrectionCandidate)
			if correction.Status != actualError || !strings.Contains(correction.Detail, "no complete, owned suite queue records") || strings.Contains(correction.Detail, "sdlc claim") {
				t.Fatalf("correction candidate lost its pinned evidence policy: %+v", correction)
			}
		})
	}
}

func assertOwnershipGap(t *testing.T, result actualResult) {
	t.Helper()
	if result.Status != actualMeasured || result.Inputs == nil {
		t.Fatalf("expected measured recovered ownership: %+v", result)
	}
	gapStart := time.Date(2026, 10, 14, 0, 10, 0, 0, time.UTC)
	gapEnd := gapStart.Add(10 * time.Minute)
	var creditedMinutes float64
	// Phases are broad timeline labels, not proof of activity. Intersect each
	// positively allocated segment with its counted phases to inspect exactly
	// the intervals that contribute to Hours, excluding any verification tail.
	for _, segment := range result.Segments {
		allocated := segment.Alloc["335"]
		if allocated <= 0 {
			continue
		}
		for _, phase := range result.CountedPhases {
			start, end := segment.Start, segment.End
			if phase.Start.After(start) {
				start = phase.Start
			}
			if phase.End.Before(end) {
				end = phase.End
			}
			if !end.After(start) {
				continue
			}
			if start.Before(gapEnd) && end.After(gapStart) {
				t.Fatalf("credited interval %s→%s overlaps unowned gap %s→%s", start, end, gapStart, gapEnd)
			}
			creditedMinutes += allocated * end.Sub(start).Seconds() / segment.End.Sub(segment.Start).Seconds()
		}
	}
	if creditedMinutes <= 0 || math.Abs(creditedMinutes-result.Hours*60) > 1e-9 {
		t.Fatalf("credit oracle omitted adopted activity: intervals=%v minutes actual=%v hours", creditedMinutes, result.Hours)
	}
	var rendered bytes.Buffer
	printActual(&rendered, result, "")
	for _, want := range []string{
		"work between cutover 2026-10-14T00:10:00Z and claim 2026-10-14T00:20:00Z is unowned and not counted",
		"tracker ownership: 2026-10-14T00:20:00Z",
	} {
		if !strings.Contains(rendered.String(), want) {
			t.Fatalf("missing visible attribution limit %q:\n%s", want, rendered.String())
		}
	}
}

func TestActualOwnershipLaterClaimAndReleaseAreNotNeverOwned(t *testing.T) {
	r, worker, cardPath := ownershipRecoveryFixture(t)
	ownedFixtureAt(t, 20)
	if _, stderr, err := executeSDLCTestCommand("claim", "--issue", "335"); err != nil {
		t.Fatalf("claim: %v %s", err, stderr)
	}
	// Claim refreshes the branch's details mirror. Preserve that real edit
	// before asking unclaim to publish a clean handoff of this working branch.
	testfix.Git(t, worker, "add", "workshop/issues/000335-spanning.md")
	testfix.Git(t, worker, "commit", "-qm", "#335: retain refreshed claim mirror")
	for _, released := range []bool{false, true} {
		if released {
			ownedFixtureAt(t, 21)
			if _, stderr, err := executeSDLCTestCommand("unclaim", "--issue", "335"); err != nil {
				t.Fatalf("release: %v %s", err, stderr)
			}
		}
		repo, err := tracker.NewRepository(t.Context(), worker, "origin")
		if err != nil {
			t.Fatal(err)
		}
		snapshot, err := repo.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		// Even when the selected cutoff precedes all ownership, the complete
		// pinned history proves this is not an issue that has never had an owner.
		until := time.Date(2026, 10, 14, 0, 18, 0, 0, time.UTC)
		owner := hostClaimant(t, worker)
		_, err = ownershipSourceWindows(t.Context(), repo, snapshot.Ref(), cardPath, owner.Repository, until, "335")
		var missing *unownedActualError
		if err == nil || errors.As(err, &missing) || !strings.Contains(err.Error(), "before its pinned right edge") {
			t.Fatalf("released=%v later ownership misclassified as never-owned: %v", released, err)
		}
		if _, has, err := issue.CardClaimant([]byte(r.card(cardPath))); err != nil || has == released {
			t.Fatalf("released=%v wrong fixture owner state: has=%v err=%v", released, has, err)
		}
	}
}
