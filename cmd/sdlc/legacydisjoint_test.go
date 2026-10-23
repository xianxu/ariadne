package main

import (
	"context"
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

// The real #151 ledger says protocol_error in its whole-issue round. An
// archive can prove temporal disjointness without repairing that review.
func legacyCapturedClose(t *testing.T, f legacyProofFixture) {
	t.Helper()
	head := f.head()
	f.write(f.active, "codecomplete", "0.80")
	ledger, err := os.ReadFile("testdata/legacy-adoption/pair151-close-gate.txt")
	if err != nil {
		t.Fatal(err)
	}
	writeRepoFile(t, f.root, "workshop/plans/000001-example-close-gate.md", strings.Replace(string(ledger), "issue: 151\n", "issue: 1\n", 1))
	f.commit("#1: close\n\nReview-Verdict: SHIP\nReview-Window: " + f.base + ".." + head)
}

func legacyDamagedCloseImport(t *testing.T) (legacyProofFixture, time.Time) {
	t.Helper()
	ownedFixtureAt(t, 0)
	f := newLegacyProofFixture(t)
	f.archive = "workshop/history/issues/000001-example.md"
	legacyCapturedClose(t, f)
	terminal := f.terminal()
	raw, err := os.ReadFile(filepath.Join(f.root, f.archive))
	if err != nil {
		t.Fatal(err)
	}
	manifest := tracker.PlanTrackerMigration(tracker.MigrationInput{Repository: "github.com/test/example", ObjectFormat: "sha1", Main: f.head(), Archived: []tracker.MigrationFile{{Path: f.archive, Raw: raw}}})
	if len(manifest.Refusals) != 0 || len(manifest.Cards) != 1 {
		t.Fatalf("old import: %+v", manifest)
	}
	card := manifest.Cards[0].Raw
	origin := filepath.Join(t.TempDir(), "origin.git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	testfix.Git(t, f.root, "remote", "add", "origin", origin)
	testfix.Git(t, f.root, "push", "-q", "-u", "origin", "main")
	bootstrapTracker(t, f.root, map[string]string{tracker.CardPath("000001", "example"): string(card)})
	stamp := strings.TrimSpace(testfix.Git(t, f.root, "show", "-s", "--format=%cI", terminal))
	end, err := time.Parse(time.RFC3339Nano, stamp)
	if err != nil {
		t.Fatal(err)
	}
	return f, end
}

func TestLegacyDisjointArchiveGenerationFences(t *testing.T) {
	for _, kind := range []string{"actual mismatch", "started mismatch", "modern completion", "reopened then restored", "provenance then stripped"} {
		t.Run(kind, func(t *testing.T) {
			f, end := legacyDamagedCloseImport(t)
			t.Chdir(f.root)
			repo, err := tracker.NewRepository(t.Context(), f.root, "origin")
			if err != nil {
				t.Fatal(err)
			}
			snap, err := repo.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			original, ok := snap.Card("000001")
			if !ok {
				t.Fatal("missing imported issue")
			}
			var changed []byte
			switch kind {
			case "actual mismatch":
				changed, err = issue.SetCardField(original.Raw, "actual_hours", "1.00")
			case "started mismatch":
				changed, err = issue.SetCardField(original.Raw, "started", "2026-01-02T00:00:00Z")
			case "modern completion":
				head := f.head()
				changed, err = issue.SetCardCompletion(original.Raw, issue.Completion{Token: "modern", Repository: "github.com/test/example", ReviewedHEAD: head, EvidenceCommit: head})
			case "reopened then restored":
				changed, err = issue.SetCardField(original.Raw, "status", "working")
			case "provenance then stripped":
				changed, err = issue.SetCardLegacyAdoption(original.Raw, issue.LegacyAdoption{Repository: "github.com/test/example", SourceMain: f.head(), SourcePath: f.archive, SourceBlob: strings.Repeat("a", 40), TerminalCommit: f.head(), WindowClass: issue.LegacyWindowlessTerminal})
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := repo.UpdateCard(original, changed, "changed-generation", func(string, string) error { return nil }); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(kind, "restored") || strings.Contains(kind, "stripped") {
				snap, err = repo.Snapshot()
				if err != nil {
					t.Fatal(err)
				}
				current, _ := snap.Card("000001")
				if err := repo.UpdateCard(current, original.Raw, "restore-old-bytes", func(string, string) error { return nil }); err != nil {
					t.Fatal(err)
				}
			}
			current := &ownedActualInputs{Snapshot: gatestate.MeasurementInputs{Repository: "github.com/test/example", Since: end.Add(time.Minute).Format(time.RFC3339Nano), Until: end.Add(time.Hour).Format(time.RFC3339Nano)}}
			if _, err := loadAdoptedExclusions(t.Context(), f.root, "162", "workshop/plans", current); err == nil || len(current.Warnings) != 0 {
				t.Fatalf("%s received legacy disjointness authority: err=%v warnings=%v", kind, err, current.Warnings)
			}
		})
	}
}

func TestLegacyDisjointArchiveDoesNotRepairCapturedClose(t *testing.T) {
	f, end := legacyDamagedCloseImport(t)
	t.Chdir(f.root)
	refs := testfix.Git(t, f.root, "for-each-ref", "--format=%(refname) %(objectname)")
	status := testfix.Git(t, f.root, "status", "--porcelain=v1")
	for _, tc := range []struct {
		name  string
		since time.Time
		fail  bool
	}{
		{"later legacy prefix", end.Add(time.Minute), false},
		{"equal earliest bound", end, true},
		{"legacy prefix overlaps despite later claim", end.Add(-time.Minute), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			current := &ownedActualInputs{Snapshot: gatestate.MeasurementInputs{Repository: "github.com/test/example", Since: tc.since.Format(time.RFC3339Nano), Until: end.Add(time.Hour).Format(time.RFC3339Nano)}}
			got, err := loadAdoptedExclusions(context.Background(), f.root, "162", "workshop/plans", current)
			if (err != nil) != tc.fail || len(got) != 0 {
				t.Fatalf("exclusions=%v err=%v want failure=%v", got, err, tc.fail)
			}
			if !tc.fail {
				label := strings.Join(current.Warnings, "\n")
				for _, required := range []string{"legacy adopted issue #1", end.Format(time.RFC3339Nano), "strictly before", "not replay-verified"} {
					if !strings.Contains(label, required) {
						t.Errorf("missing %q in disjointness label %q", required, label)
					}
				}
			}
		})
	}
	if _, err := f.resolve(); err == nil || !strings.Contains(err.Error(), "no finalized whole-issue close") {
		t.Fatalf("archive relevance repaired the invalid close: %v", err)
	}
	if refs != testfix.Git(t, f.root, "for-each-ref", "--format=%(refname) %(objectname)") || status != testfix.Git(t, f.root, "status", "--porcelain=v1") {
		t.Fatal("read-only disjointness changed refs or checkout")
	}
}
