package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

type legacyProofFixture struct {
	t                           *testing.T
	root, active, archive, base string
}

func newLegacyProofFixture(t *testing.T) legacyProofFixture {
	root := testfix.Repo(t, testfix.InitialCommit())
	f := legacyProofFixture{t, root, "workshop/issues/000001-example.md", "workshop/history/000001-example.md", ""}
	f.write(f.active, "working", "0.80")
	f.commit("start")
	f.base = f.head()
	return f
}
func (f legacyProofFixture) head() string {
	return strings.TrimSpace(testfix.Git(f.t, f.root, "rev-parse", "HEAD"))
}
func (f legacyProofFixture) commit(message string) string {
	testfix.Git(f.t, f.root, "add", ".")
	testfix.Git(f.t, f.root, "commit", "--allow-empty", "-qm", message)
	return f.head()
}
func (f legacyProofFixture) write(path, status, actual string) {
	f.t.Helper()
	writeRepoFile(f.t, f.root, path, fmt.Sprintf("---\nid: 000001\nstatus: %s\ncreated: 2026-01-01\nstarted: 2026-01-01T00:00:00Z\nactual_hours: %s\n---\n\n# Example\n\n## Problem\n\nLegacy fixture.\n", status, actual))
}
func (f legacyProofFixture) close(ledger bool) string {
	head := f.head()
	f.write(f.active, "codecomplete", "0.80")
	file, err := os.OpenFile(filepath.Join(f.root, f.active), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		f.t.Fatal(err)
	}
	_, err = file.WriteString("\nClosed at " + head + "\n")
	_ = file.Close()
	if err != nil {
		f.t.Fatal(err)
	}
	if ledger {
		writeRepoFile(f.t, f.root, "workshop/plans/000001-example-close-gate.md", gatestate.Render(gatestate.Ledger{Gate: "boundary-review", IssueNum: 1, IDPrefix: "BR", Rounds: []gatestate.Round{{N: 1, Reviewed: head}}}, "fixture"))
	}
	return f.commit(fmt.Sprintf("#1: close\n\nReview-Verdict: SHIP\nReview-Window: %s..%s", f.base, head))
}
func (f legacyProofFixture) terminal() string {
	if err := os.MkdirAll(filepath.Dir(filepath.Join(f.root, f.archive)), 0755); err != nil {
		f.t.Fatal(err)
	}
	testfix.Git(f.t, f.root, "mv", f.active, f.archive)
	f.write(f.archive, "done", "0.80")
	return f.commit("archive")
}
func (f legacyProofFixture) resolve() (legacyCompletionProof, error) {
	raw, err := os.ReadFile(filepath.Join(f.root, f.archive))
	if err != nil {
		f.t.Fatal(err)
	}
	return resolveLegacyCompletion(context.Background(), f.root, f.head(), "github.com/test/example", tracker.MigrationFile{Path: f.archive, Raw: raw}, "workshop/plans")
}
func TestLegacyCompletionProofLatestCloseIgnoresMetadata(t *testing.T) {
	f := newLegacyProofFixture(t)
	f.close(true)
	f.commit("code fix")
	latest := f.close(true)
	f.write(f.active, "codecomplete", "0.80")
	writeRepoFile(t, f.root, "note", "metadata")
	f.commit("metadata edit")
	terminal := f.terminal()
	got, err := f.resolve()
	if err != nil {
		t.Fatal(err)
	}
	if !got.HasCompletion || got.Completion.EvidenceCommit != latest || got.Adoption.TerminalCommit != terminal || got.Adoption.WindowClass != issue.LegacyWindowlessClose {
		t.Fatalf("wrong proof: %+v", got)
	}
}
func TestLegacyCompletionProofBornArchivedIsUnbound(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	f := legacyProofFixture{t: t, root: root, archive: "workshop/history/000001-example.md"}
	f.write(f.archive, "done", "0.80")
	terminal := f.commit("import old archive")
	got, err := f.resolve()
	if err != nil {
		t.Fatal(err)
	}
	if got.HasCompletion || got.Adoption.WindowClass != issue.LegacyWindowlessTerminal || got.Adoption.TerminalCommit != terminal {
		t.Fatalf("fabricated close: %+v", got)
	}
}
func TestLegacyCompletionProofRefusesDamagedEvidence(t *testing.T) {
	for _, kind := range []string{"actual", "identity", "ledger", "window", "verdict", "inputs"} {
		t.Run(kind, func(t *testing.T) {
			f := newLegacyProofFixture(t)
			head := f.head()
			f.write(f.active, "codecomplete", "0.80")
			switch kind {
			case "identity":
				writeRepoFile(t, f.root, f.active, "---\nid: 000002\nstatus: codecomplete\nstarted: 2026-01-01T00:00:00Z\nactual_hours: 0.80\n---\n# Other\n")
			case "ledger":
				writeRepoFile(t, f.root, "workshop/plans/000001-example-close-gate.md", "broken ledger")
			case "inputs":
				writeRepoFile(t, f.root, "workshop/plans/000001-example-close-gate.md", gatestate.Render(gatestate.Ledger{Gate: "boundary-review", IssueNum: 1, IDPrefix: "BR", Rounds: []gatestate.Round{{N: 1, Reviewed: head, ActualInputs: &gatestate.MeasurementInputs{}}}}, "fixture"))
			}
			verdict := "SHIP"
			if kind == "verdict" {
				verdict = "REWORK"
			}
			window := f.base + ".." + head
			if kind == "window" {
				window = "HEAD..HEAD"
			}
			f.commit("#1: close\n\nReview-Verdict: " + verdict + "\nReview-Window: " + window)
			f.terminal()
			if kind == "actual" {
				f.write(f.archive, "done", "1.00")
				f.commit("change actual")
			}
			if got, err := f.resolve(); err == nil {
				t.Fatalf("accepted damaged %s: %+v", kind, got)
			}
		})
	}
}

func legacyProofCard(t *testing.T, f legacyProofFixture, proof legacyCompletionProof) []byte {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(f.root, f.archive))
	if err != nil {
		t.Fatal(err)
	}
	card, _, err := issue.SplitCardWithFormat(source, "sha1")
	if err != nil {
		t.Fatal(err)
	}
	if proof.HasCompletion {
		card, err = issue.SetCardCompletion(card, proof.Completion)
		if err != nil {
			t.Fatal(err)
		}
	}
	card, err = issue.SetCardLegacyAdoption(card, proof.Adoption)
	if err != nil {
		t.Fatal(err)
	}
	return card
}
func legacyProofRecord(t *testing.T, raw []byte) tracker.IssueRecord {
	t.Helper()
	card, err := issue.ParseCard(raw)
	if err != nil {
		t.Fatal(err)
	}
	return tracker.IssueRecord{ID: card.ID, Card: &tracker.Record{ID: card.ID, Path: tracker.CardPath(card.ID, "example"), Card: card, Raw: raw}}
}
func TestVerifyLegacyAdoptionRequiresCutoverAuthority(t *testing.T) {
	f := newLegacyProofFixture(t)
	f.close(true)
	f.terminal()
	proof, err := f.resolve()
	if err != nil {
		t.Fatal(err)
	}
	card := legacyProofCard(t, f, proof)
	rec := legacyProofRecord(t, card)
	if end, err := verifyLegacyAdoption(context.Background(), f.root, f.trackerPin(), rec, proof.Adoption.Repository, "workshop/plans"); err == nil {
		t.Fatalf("accepted uncut-over provenance: %v", end)
	}
}
func TestVerifyLegacyAdoptionRevalidatesCurrentGeneration(t *testing.T) {
	for _, closeBacked := range []bool{false, true} {
		t.Run(fmt.Sprint(closeBacked), func(t *testing.T) {
			f := newLegacyProofFixture(t)
			if closeBacked {
				f.close(true)
				f.terminal()
			} else {
				testfix.Git(t, f.root, "rm", f.active)
				f.write(f.archive, "done", "0.80")
				f.commit("born archived")
			}
			sourcePath := filepath.Join(f.root, f.archive)
			rawSource, e := os.ReadFile(sourcePath)
			if e != nil {
				t.Fatal(e)
			}
			quoted := strings.ReplaceAll(string(rawSource), "id: 000001", "id: '000001'")
			quoted = strings.ReplaceAll(quoted, "started: 2026-01-01T00:00:00Z", "started: '2026-01-01T00:00:00Z'")
			if e := os.WriteFile(sourcePath, []byte(quoted), 0600); e != nil {
				t.Fatal(e)
			}
			f.commit("quote-equivalent terminal metadata")
			proof, err := f.resolve()
			if err != nil {
				t.Fatal(err)
			}
			card := legacyProofCard(t, f, proof)
			origin := filepath.Join(t.TempDir(), "origin.git")
			testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
			testfix.Git(t, f.root, "remote", "add", "origin", origin)
			testfix.Git(t, f.root, "push", "-q", "-u", "origin", "main")
			bootstrapTracker(t, f.root, map[string]string{tracker.CardPath("000001", "example"): string(card)})
			rec := legacyProofRecord(t, card)
			end, err := verifyLegacyAdoption(context.Background(), f.root, f.trackerPin(), rec, proof.Adoption.Repository, "workshop/plans")
			if err != nil || end.IsZero() {
				t.Fatalf("valid provenance: %v, %v", end, err)
			}
			for _, mutation := range []string{"actual_hours", "started", "status"} {
				t.Run(mutation, func(t *testing.T) {
					value := "1.00"
					if mutation == "started" {
						value = "2026-01-02T00:00:00Z"
					}
					if mutation == "status" {
						value = "working"
					}
					changed, e := issue.SetCardField(card, mutation, value)
					if e != nil {
						t.Fatal(e)
					}
					_, e = verifyLegacyAdoption(context.Background(), f.root, f.trackerPin(), legacyProofRecord(t, changed), proof.Adoption.Repository, "workshop/plans")
					if e == nil {
						t.Fatalf("accepted changed %s", mutation)
					}
				})
			}
		})
	}
}
func TestLegacyCompletionProofRetainsOriginalInputs(t *testing.T) {
	f := newLegacyProofFixture(t)
	head := f.head()
	f.write(f.active, "codecomplete", "0.80")
	saved := &gatestate.MeasurementInputs{Version: 1, Method: gatestate.MeasurementMethodLegacyV1, Repository: "github.com/test/example", CodeHead: head, Since: "2026-01-01T00:00:00Z", Until: "2026-01-01T01:00:00Z", GapMinutes: 10}
	writeRepoFile(t, f.root, "workshop/plans/000001-example-close-gate.md", gatestate.Render(gatestate.Ledger{Gate: "boundary-review", IssueNum: 1, IDPrefix: "BR", Rounds: []gatestate.Round{{N: 1, Reviewed: head, ActualInputs: saved}}}, "fixture"))
	close := f.commit(fmt.Sprintf("#1: close\n\nReview-Verdict: SHIP\nReview-Window: %s..%s", f.base, head))
	f.terminal()
	proof, err := f.resolve()
	if err != nil {
		t.Fatal(err)
	}
	if proof.Adoption.WindowClass != issue.LegacyRetained || proof.Completion.EvidenceCommit != close {
		t.Fatalf("lost original inputs reference: %+v", proof)
	}
	end, err := verifyLegacyAdoption(context.Background(), f.root, f.trackerPin(), legacyProofRecord(t, legacyProofCard(t, f, proof)), proof.Adoption.Repository, "workshop/plans")
	if err != nil || !end.IsZero() {
		t.Fatalf("retained proof granted temporal shortcut: %v %v", end, err)
	}
}

func TestLegacyCompletionProofRefusesDeletedRecreatedPath(t *testing.T) {
	f := newLegacyProofFixture(t)
	f.close(true)
	f.terminal()
	testfix.Git(t, f.root, "rm", f.archive)
	f.commit("delete old archive")
	f.write(f.archive, "done", "0.80")
	f.commit("reuse same identity")
	if got, err := f.resolve(); err == nil {
		t.Fatalf("accepted recreated generation: %+v", got)
	}
}
func TestLegacyCompletionProofPreservesTrailingSourceBytes(t *testing.T) {
	f := newLegacyProofFixture(t)
	f.close(true)
	f.terminal()
	file, err := os.OpenFile(filepath.Join(f.root, f.archive), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.WriteString("\n\n  \n")
	_ = file.Close()
	if err != nil {
		t.Fatal(err)
	}
	f.commit("preserve original whitespace")
	if _, err = f.resolve(); err != nil {
		t.Fatal(err)
	}
}

func TestLegacyCompletionProofHistoryScansScaleLinearly(t *testing.T) {
	first := newLegacyProofFixture(t)
	first.close(true)
	first.terminal()
	second := newLegacyProofFixture(t)
	second.close(true)
	second.terminal()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	counter := filepath.Join(t.TempDir(), "scans")
	wrapper := "#!/bin/sh\nfor arg in \"$@\"; do\n if [ \"$arg\" = --follow ]; then echo scan >> \"$LEGACY_SCAN_COUNTER\"; fi\ndone\nexec \"$LEGACY_REAL_GIT\" \"$@\"\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(wrapper), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LEGACY_SCAN_COUNTER", counter)
	t.Setenv("LEGACY_REAL_GIT", realGit)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	for round := 1; round <= 2; round++ {
		for _, fixture := range []legacyProofFixture{first, second} {
			if _, err := fixture.resolve(); err != nil {
				t.Fatal(err)
			}
		}
		raw, err := os.ReadFile(counter)
		if err != nil {
			t.Fatal(err)
		}
		if got := len(strings.Fields(string(raw))); got != 2*round {
			t.Fatalf("%d resolutions made %d history walks", 2*round, got)
		}
	}
}

func TestLegacyCompletionProofRefusesHistoryOverflowBeforeObjectReads(t *testing.T) {
	f := newLegacyProofFixture(t)
	f.close(true)
	f.terminal()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	history := filepath.Join(t.TempDir(), "history")
	synthetic := strings.Repeat("@@"+strings.Repeat("0", 40)+"\n\n"+f.archive+"\n", legacyCompletionHistoryLimit+1)
	if err := os.WriteFile(history, []byte(synthetic), 0600); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\nfor arg in \"$@\"; do\n if [ \"$arg\" = --follow ]; then cat \"$LEGACY_HISTORY\"; exit 0; fi\ndone\nexec \"$LEGACY_REAL_GIT\" \"$@\"\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LEGACY_REAL_GIT", realGit)
	t.Setenv("LEGACY_HISTORY", history)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	_, err = f.resolve()
	if err == nil || !strings.Contains(err.Error(), "exceeds complete-history bound") {
		t.Fatalf("history sentinel was not rejected before object reads: %v", err)
	}
}
func TestLegacyCompletionProofHonorsCancellation(t *testing.T) {
	f := newLegacyProofFixture(t)
	f.terminal()
	raw, err := os.ReadFile(filepath.Join(f.root, f.archive))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = resolveLegacyCompletion(ctx, f.root, f.head(), "github.com/test/example", tracker.MigrationFile{Path: f.archive, Raw: raw}, "workshop/plans")
	if err == nil {
		t.Fatal("ignored cancellation")
	}
}

func TestLegacyCompletionProofRefusesIncomparableCloses(t *testing.T) {
	f := newLegacyProofFixture(t)
	testfix.Git(t, f.root, "checkout", "-qb", "alternate-close")
	f.commit("alternate feature")
	f.close(true)
	testfix.Git(t, f.root, "checkout", "-q", "main")
	f.commit("main feature")
	f.close(true)
	testfix.Git(t, f.root, "merge", "-q", "-s", "ours", "--no-edit", "alternate-close")
	f.terminal()
	if got, err := f.resolve(); err == nil || !strings.Contains(err.Error(), "incomparable") {
		t.Fatalf("incomparable proof result: %+v, %v", got, err)
	}
}

func TestLegacyCompletionProofAcceptsMergedClose(t *testing.T) {
	f := newLegacyProofFixture(t)
	testfix.Git(t, f.root, "checkout", "-qb", "issue-close")
	f.commit("feature")
	close := f.close(true)
	testfix.Git(t, f.root, "checkout", "-q", "main")
	writeRepoFile(t, f.root, "unrelated", "other work")
	f.commit("unrelated main")
	testfix.Git(t, f.root, "merge", "-q", "--no-edit", "issue-close")
	f.terminal()
	got, err := f.resolve()
	if err != nil {
		t.Fatal(err)
	}
	if !got.HasCompletion || got.Completion.EvidenceCommit != close || got.Completion.LandedCommit != "" {
		t.Fatalf("wrong merged binding: %+v", got)
	}
}

func TestLegacyCompletionProofArchiveTransitionBoundsAlreadyDone(t *testing.T) {
	f := newLegacyProofFixture(t)
	f.write(f.active, "done", "0.80")
	f.commit("old terminal record")
	archive := f.terminal()
	got, err := f.resolve()
	if err != nil {
		t.Fatal(err)
	}
	if got.Adoption.TerminalCommit != archive {
		t.Fatalf("archive bound selected earlier active done: %+v", got)
	}
}

func TestLegacyCompletionProofQuotedActualEquivalent(t *testing.T) {
	f := newLegacyProofFixture(t)
	f.close(true)
	f.terminal()
	file := filepath.Join(f.root, f.archive)
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(strings.ReplaceAll(string(raw), "actual_hours: 0.80", "actual_hours: '0.80'")), 0600); err != nil {
		t.Fatal(err)
	}
	f.commit("quote-equivalent legacy actual")
	if _, err = f.resolve(); err != nil {
		t.Fatal(err)
	}
}

func legacyUnmarkedTerminalFixture(t *testing.T, closeBacked bool, sourceMismatch bool) (legacyProofFixture, tracker.IssueRecord) {
	t.Helper()
	var f legacyProofFixture
	if closeBacked {
		f = newLegacyProofFixture(t)
		f.archive = "workshop/history/issues/000001-example.md"
		f.close(true)
		f.terminal()
	} else {
		f = legacyProofFixture{t: t, root: testfix.Repo(t, testfix.InitialCommit()), archive: "workshop/history/issues/000001-example.md"}
		f.write(f.archive, "done", "0.80")
		f.commit("born archived")
	}
	raw, err := os.ReadFile(filepath.Join(f.root, f.archive))
	if err != nil {
		t.Fatal(err)
	}
	manifest := tracker.PlanTrackerMigration(tracker.MigrationInput{Repository: "github.com/test/example", ObjectFormat: "sha1", Main: f.head(), Archived: []tracker.MigrationFile{{Path: f.archive, Raw: raw}}})
	if len(manifest.Refusals) > 0 || len(manifest.Cards) != 1 {
		t.Fatalf("legacy migration fixture: %+v", manifest)
	}
	card := manifest.Cards[0].Raw
	if sourceMismatch {
		card, err = issue.SetCardField(card, "actual_hours", "1.00")
		if err != nil {
			t.Fatal(err)
		}
	}
	origin := filepath.Join(t.TempDir(), "origin.git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	testfix.Git(t, f.root, "remote", "add", "origin", origin)
	testfix.Git(t, f.root, "push", "-q", "-u", "origin", "main")
	bootstrapTracker(t, f.root, map[string]string{tracker.CardPath("000001", "example"): string(card)})
	return f, legacyProofRecord(t, card)
}
func TestVerifyLegacyAdoptionRecoversOldTerminalWithoutWrites(t *testing.T) {
	f, rec := legacyUnmarkedTerminalFixture(t, false, false)
	refs := testfix.Git(t, f.root, "for-each-ref", "--format=%(refname) %(objectname)")
	status := testfix.Git(t, f.root, "status", "--porcelain=v1")
	head := f.head()
	raw := append([]byte(nil), rec.Card.Raw...)
	for i := 0; i < 2; i++ {
		end, err := verifyLegacyAdoption(context.Background(), f.root, f.trackerPin(), rec, "github.com/test/example", "workshop/plans")
		if err != nil || end.IsZero() {
			t.Fatalf("old terminal not recovered: %v %v", end, err)
		}
	}
	if refs != testfix.Git(t, f.root, "for-each-ref", "--format=%(refname) %(objectname)") || status != testfix.Git(t, f.root, "status", "--porcelain=v1") || head != f.head() || string(raw) != string(rec.Card.Raw) {
		t.Fatal("read-only recovery changed repository/card")
	}
}
func TestVerifyLegacyAdoptionUnmarkedRecoveryHasNoModernFallback(t *testing.T) {
	t.Run("modern absent from root", func(t *testing.T) {
		f, rec := legacyUnmarkedTerminalFixture(t, false, false)
		raw := []byte(strings.ReplaceAll(string(rec.Card.Raw), "000001", "000002"))
		modern := legacyProofRecord(t, raw)
		modern.Card.Path = tracker.CardPath("000002", "modern")
		end, err := verifyLegacyAdoption(context.Background(), f.root, f.trackerPin(), modern, "github.com/test/example", "workshop/plans")
		if err != nil || !end.IsZero() {
			t.Fatalf("modern card inherited root proof: %v %v", end, err)
		}
	})
	t.Run("changed numeric generation", func(t *testing.T) {
		f, rec := legacyUnmarkedTerminalFixture(t, false, false)
		raw, err := issue.SetCardField(rec.Card.Raw, "actual_hours", "1.00")
		if err != nil {
			t.Fatal(err)
		}
		end, err := verifyLegacyAdoption(context.Background(), f.root, f.trackerPin(), legacyProofRecord(t, raw), "github.com/test/example", "workshop/plans")
		if !end.IsZero() {
			t.Fatalf("changed card inherited root proof: %v %v", end, err)
		}
	})
	t.Run("mismatching archived source", func(t *testing.T) {
		f, rec := legacyUnmarkedTerminalFixture(t, false, true)
		end, err := verifyLegacyAdoption(context.Background(), f.root, f.trackerPin(), rec, "github.com/test/example", "workshop/plans")
		if err == nil || !end.IsZero() {
			t.Fatalf("changed source accepted: %v %v", end, err)
		}
	})
	t.Run("close-backed needs explicit repair", func(t *testing.T) {
		f, rec := legacyUnmarkedTerminalFixture(t, true, false)
		end, err := verifyLegacyAdoption(context.Background(), f.root, f.trackerPin(), rec, "github.com/test/example", "workshop/plans")
		if err != nil || !end.IsZero() {
			t.Fatalf("close-backed import silently repaired: %v %v", end, err)
		}
	})
}

func TestVerifyLegacyAdoptionUnmarkedCorrectionCannotRecover(t *testing.T) {
	f, rec := legacyUnmarkedTerminalFixture(t, false, false)
	head := f.head()
	c := issue.Completion{Token: "new-generation", Repository: "github.com/test/example", ReviewedHEAD: head, EvidenceCommit: head}
	raw, err := issue.SetCardCompletion(rec.Card.Raw, c)
	if err != nil {
		t.Fatal(err)
	}
	inputs := fmt.Sprintf(`{"repository":%q,"code_head":%q}`, c.Repository, c.ReviewedHEAD)
	correction := issue.ActualCorrection{Version: 1, Token: "corrected", CompletionToken: c.Token, Repository: c.Repository, EvidenceCommit: c.EvidenceCommit, ReviewedHEAD: c.ReviewedHEAD, PreviousActual: "0.80", NewActual: "0.80", Reason: "corrected generation", MeasurementDigest: fmt.Sprintf("%x", sha256.Sum256([]byte(inputs))), Inputs: inputs}
	raw, err = issue.SetCardActualCorrection(raw, correction)
	if err != nil {
		t.Fatal(err)
	}
	end, err := verifyLegacyAdoption(context.Background(), f.root, f.trackerPin(), legacyProofRecord(t, raw), c.Repository, "workshop/plans")
	if err != nil || !end.IsZero() {
		t.Fatalf("corrected modern generation inherited old bound: %v %v", end, err)
	}
}

func TestVerifyLegacyAdoptionUnmarkedActiveDoneHasNoArchiveFallback(t *testing.T) {
	f := newLegacyProofFixture(t)
	f.write(f.active, "done", "0.80")
	f.commit("active done without archive")
	raw, err := os.ReadFile(filepath.Join(f.root, f.active))
	if err != nil {
		t.Fatal(err)
	}
	manifest := tracker.PlanTrackerMigration(tracker.MigrationInput{Repository: "github.com/test/example", ObjectFormat: "sha1", Main: f.head(), Active: []tracker.MigrationFile{{Path: f.active, Raw: raw}}})
	if len(manifest.Refusals) > 0 || len(manifest.Cards) != 1 {
		t.Fatalf("active done fixture: %+v", manifest)
	}
	card := manifest.Cards[0].Raw
	origin := filepath.Join(t.TempDir(), "origin.git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	testfix.Git(t, f.root, "remote", "add", "origin", origin)
	testfix.Git(t, f.root, "push", "-q", "-u", "origin", "main")
	bootstrapTracker(t, f.root, map[string]string{tracker.CardPath("000001", "example"): string(card)})
	end, err := verifyLegacyAdoption(context.Background(), f.root, f.trackerPin(), legacyProofRecord(t, card), "github.com/test/example", "workshop/plans")
	if err != nil || !end.IsZero() {
		t.Fatalf("active record inherited archive-only fallback: %v %v", end, err)
	}
}

// Echo's native --follow capture followed a 1%-similar project template past
// the issue's creation. Keep the real Git heuristic in this regression.
func TestLegacyCompletionProofDoesNotFollowUnrelatedTemplate(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	f := legacyProofFixture{t: t, root: root, active: "workshop/issues/000001-example.md", archive: "workshop/history/issues/000001-example.md"}
	common := "---\nid: 000001\nstatus: working\ncreated: 2026-01-01\nstarted: 2026-01-01T00:00:00Z\nactual_hours: 0.80\n---\n\n# Example\n\n## Problem\n\nLegacy fixture.\n"
	writeRepoFile(t, root, "workshop/projects/control-board.md", common+strings.Repeat("Project-only context.\n", 30))
	f.commit("project template")
	writeRepoFile(t, root, f.active, common+strings.Repeat("Issue-only context.\n", 30))
	f.commit("create issue from template")
	f.base = f.head()
	close := f.close(true)
	f.terminal()
	capture := testfix.Git(t, root, "log", "--topo-order", "--full-history", "--follow", "-M1%", "--format=@@%H", "--name-only", "HEAD", "--", f.archive)
	if !strings.Contains(capture, "workshop/projects/control-board.md") {
		t.Fatalf("fixture did not reproduce Echo's false template ancestry: %s", capture)
	}
	proof, err := f.resolve()
	if err != nil || !proof.HasCompletion || proof.Completion.EvidenceCommit != close {
		t.Fatalf("real archive rename lost its close through unrelated template history: %+v, %v", proof, err)
	}
}

func TestLegacyHistoryIssueBirthChecksEveryParent(t *testing.T) {
	f := newLegacyProofFixture(t)
	birth, err := legacyHistoryIssueBirth(context.Background(), f.root, legacySourceState{commit: f.head(), path: f.active})
	if err != nil || !birth {
		t.Fatalf("new issue path not proved: %v %v", birth, err)
	}
	withIssue := f.head()
	testfix.Git(t, f.root, "checkout", "-qb", "without-issue", withIssue+"^")
	writeRepoFile(t, f.root, "unrelated", "main work")
	f.commit("other parent")
	testfix.Git(t, f.root, "merge", "-q", "--no-edit", "main")
	birth, err = legacyHistoryIssueBirth(context.Background(), f.root, legacySourceState{commit: f.head(), path: f.active})
	if err != nil || birth {
		t.Fatalf("path present in second parent mistaken for creation: %v %v", birth, err)
	}
}

func legacyFixtureTrackerWrite(t *testing.T, f legacyProofFixture, raw []byte, token string) tracker.IssueRecord {
	t.Helper()
	repo, err := tracker.NewRepository(context.Background(), f.root, "origin")
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	current, err := snapshot.Require("000001")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateCard(current, raw, token, func(string, string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	snapshot, err = repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	record, err := snapshot.Require("000001")
	if err != nil {
		t.Fatal(err)
	}
	return tracker.IssueRecord{ID: record.ID, Card: &record}
}
func legacyFixtureExplicitTerminal(t *testing.T, f legacyProofFixture, rec tracker.IssueRecord) tracker.IssueRecord {
	t.Helper()
	sourceMain, _, err := legacyMigrationSource(context.Background(), f.root, "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(f.root, f.archive))
	if err != nil {
		t.Fatal(err)
	}
	proof, err := resolveLegacyCompletion(context.Background(), f.root, sourceMain, "github.com/test/example", tracker.MigrationFile{Path: f.archive, Raw: raw}, "workshop/plans")
	if err != nil {
		t.Fatal(err)
	}
	card, err := issue.SetCardLegacyAdoption(rec.Card.Raw, proof.Adoption)
	if err != nil {
		t.Fatal(err)
	}
	return legacyFixtureTrackerWrite(t, f, card, "record-terminal")
}
func TestVerifyLegacyAdoptionTrackerHistoryRejectsRestoredReopen(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(fmt.Sprint(explicit), func(t *testing.T) {
			f, rec := legacyUnmarkedTerminalFixture(t, false, false)
			if explicit {
				rec = legacyFixtureExplicitTerminal(t, f, rec)
			}
			original := append([]byte(nil), rec.Card.Raw...)
			reopened, err := issue.SetCardField(original, "status", "working")
			if err != nil {
				t.Fatal(err)
			}
			legacyFixtureTrackerWrite(t, f, reopened, "reopened")
			rec = legacyFixtureTrackerWrite(t, f, original, "restore-done")
			end, err := verifyLegacyAdoption(context.Background(), f.root, f.trackerPin(), rec, "github.com/test/example", "workshop/plans")
			if err == nil || !end.IsZero() {
				t.Fatalf("restored reopen inherited old generation: %v %v", end, err)
			}
		})
	}
}

func (f legacyProofFixture) trackerPin() string {
	pin, err := ownedActualGit(context.Background(), f.root, "rev-parse", "--verify", "refs/remotes/origin/issue-tracker")
	if err != nil {
		return ""
	}
	return pin
}

func TestVerifyLegacyAdoptionTrackerHistoryAllowsMetadataAndRepair(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(fmt.Sprint(explicit), func(t *testing.T) {
			f, rec := legacyUnmarkedTerminalFixture(t, false, false)
			if explicit {
				rec = legacyFixtureExplicitTerminal(t, f, rec)
			}
			raw, err := issue.SetCardField(rec.Card.Raw, "updated", "2026-10-23")
			if err != nil {
				t.Fatal(err)
			}
			rec = legacyFixtureTrackerWrite(t, f, raw, "metadata")
			end, err := verifyLegacyAdoption(context.Background(), f.root, f.trackerPin(), rec, "github.com/test/example", "workshop/plans")
			if err != nil || end.IsZero() {
				t.Fatalf("metadata/monotone adoption rejected: %v %v", end, err)
			}
		})
	}
	t.Run("matching repaired close", func(t *testing.T) {
		f, rec := legacyUnmarkedTerminalFixture(t, true, false)
		sourceMain, _, err := legacyMigrationSource(context.Background(), f.root, "HEAD")
		if err != nil {
			t.Fatal(err)
		}
		source, err := os.ReadFile(filepath.Join(f.root, f.archive))
		if err != nil {
			t.Fatal(err)
		}
		proof, err := resolveLegacyCompletion(context.Background(), f.root, sourceMain, "github.com/test/example", tracker.MigrationFile{Path: f.archive, Raw: source}, "workshop/plans")
		if err != nil {
			t.Fatal(err)
		}
		raw, err := issue.SetCardCompletion(rec.Card.Raw, proof.Completion)
		if err != nil {
			t.Fatal(err)
		}
		raw, err = issue.SetCardLegacyAdoption(raw, proof.Adoption)
		if err != nil {
			t.Fatal(err)
		}
		rec = legacyFixtureTrackerWrite(t, f, raw, "matching-close-repair")
		end, err := verifyLegacyAdoption(context.Background(), f.root, f.trackerPin(), rec, "github.com/test/example", "workshop/plans")
		if err != nil || end.IsZero() {
			t.Fatalf("matching repair rejected: %v %v", end, err)
		}
	})
}
func TestVerifyLegacyAdoptionTrackerSnapshotMustMatchCard(t *testing.T) {
	for _, kind := range []string{"raw", "blob", "pin"} {
		t.Run(kind, func(t *testing.T) {
			f, rec := legacyUnmarkedTerminalFixture(t, false, false)
			pin := f.trackerPin()
			switch kind {
			case "raw":
				raw, err := issue.SetCardField(rec.Card.Raw, "updated", "2026-10-23")
				if err != nil {
					t.Fatal(err)
				}
				rec = legacyProofRecord(t, raw)
			case "blob":
				rec.Card.BlobOID = strings.Repeat("1", 40)
			case "pin":
				pin = f.head()
			}
			end, err := verifyLegacyAdoption(context.Background(), f.root, pin, rec, "github.com/test/example", "workshop/plans")
			if err == nil || !end.IsZero() {
				t.Fatalf("unmatched %s snapshot accepted: %v %v", kind, end, err)
			}
		})
	}
}
func TestVerifyLegacyAdoptionTrackerHistoryReadFailuresRefuse(t *testing.T) {
	for _, kind := range []string{"failed", "truncated", "overflow"} {
		t.Run(kind, func(t *testing.T) {
			f, rec := legacyUnmarkedTerminalFixture(t, false, false)
			raw, err := issue.SetCardField(rec.Card.Raw, "updated", "2026-10-23")
			if err != nil {
				t.Fatal(err)
			}
			rec = legacyFixtureTrackerWrite(t, f, raw, "metadata")
			pin := f.trackerPin()
			realGit, err := exec.LookPath("git")
			if err != nil {
				t.Fatal(err)
			}
			bin := t.TempDir()
			history := filepath.Join(t.TempDir(), "history")
			output := pin + "\n"
			if kind == "overflow" {
				output = strings.Repeat(output, legacyCompletionHistoryLimit+1)
			}
			if err := os.WriteFile(history, []byte(output), 0600); err != nil {
				t.Fatal(err)
			}
			script := "#!/bin/sh\nscan=0\npin=0\nfor arg in \"$@\"; do\n [ \"$arg\" = --format=%H ] && scan=1\n [ \"$arg\" = \"$LEGACY_TRACKER_PIN\" ] && pin=1\ndone\nif [ \"$scan$pin\" = 11 ]; then\n [ \"$LEGACY_HISTORY_MODE\" = failed ] && exit 91\n cat \"$LEGACY_HISTORY\"\n exit 0\nfi\nexec \"$LEGACY_REAL_GIT\" \"$@\"\n"
			if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("LEGACY_REAL_GIT", realGit)
			t.Setenv("LEGACY_HISTORY", history)
			t.Setenv("LEGACY_HISTORY_MODE", kind)
			t.Setenv("LEGACY_TRACKER_PIN", pin)
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			end, err := verifyLegacyAdoption(context.Background(), f.root, pin, rec, "github.com/test/example", "workshop/plans")
			if err == nil || !end.IsZero() || !strings.Contains(err.Error(), "tracker history") {
				t.Fatalf("incomplete %s history accepted: %v %v", kind, end, err)
			}
		})
	}
}

func TestVerifyLegacyAdoptionTrackerHistoryRejectsRestoredAuthority(t *testing.T) {
	for _, kind := range []string{"actual_hours", "started", "completion", "provenance", "correction"} {
		t.Run(kind, func(t *testing.T) {
			f, rec := legacyUnmarkedTerminalFixture(t, false, false)
			rec = legacyFixtureExplicitTerminal(t, f, rec)
			original := append([]byte(nil), rec.Card.Raw...)
			var changed []byte
			var err error
			switch kind {
			case "actual_hours":
				changed, err = issue.SetCardField(original, kind, "1.00")
			case "started":
				changed, err = issue.SetCardField(original, kind, "2026-01-02T00:00:00Z")
			case "completion":
				head := f.head()
				changed, err = issue.SetCardCompletion(original, issue.Completion{Token: "another-close", Repository: "github.com/test/example", ReviewedHEAD: head, EvidenceCommit: head})
			case "correction":
				head := f.head()
				completion := issue.Completion{Token: "another-close", Repository: "github.com/test/example", ReviewedHEAD: head, EvidenceCommit: head}
				changed, err = issue.SetCardCompletion(original, completion)
				if err == nil {
					inputs := fmt.Sprintf(`{"repository":%q,"code_head":%q}`, completion.Repository, head)
					changed, err = issue.SetCardActualCorrection(changed, issue.ActualCorrection{Version: 1, Token: "intervening-correction", CompletionToken: completion.Token, Repository: completion.Repository, EvidenceCommit: head, ReviewedHEAD: head, PreviousActual: "0.80", NewActual: "0.80", Reason: "same-value correction", MeasurementDigest: fmt.Sprintf("%x", sha256.Sum256([]byte(inputs))), Inputs: inputs})
				}
			case "provenance":
				adoption, _, e := issue.CardLegacyAdoption(original)
				if e != nil {
					t.Fatal(e)
				}
				adoption.SourceBlob = strings.Repeat("1", 40)
				changed, err = issue.SetCardLegacyAdoption(original, adoption)
			}
			if err != nil {
				t.Fatal(err)
			}
			legacyFixtureTrackerWrite(t, f, changed, "changed-authority")
			rec = legacyFixtureTrackerWrite(t, f, original, "restore-authority")
			end, err := verifyLegacyAdoption(context.Background(), f.root, f.trackerPin(), rec, "github.com/test/example", "workshop/plans")
			if err == nil || !end.IsZero() {
				t.Fatalf("restored %s authority inherited old bound: %v %v", kind, end, err)
			}
			if kind == "correction" && !strings.Contains(err.Error(), "actual correction") {
				t.Fatalf("correction guard was not exercised: %v", err)
			}
		})
	}
}
