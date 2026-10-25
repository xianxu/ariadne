package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

func adoptionReadSetCard(id, hours, extra string) tracker.IssueRecord {
	fm := "id: " + id + "\ncreated: 2026-10-14\nupdated: 2026-10-14\nactual_hours: " + hours + "\n" + extra
	raw := []byte("---\n" + fm + "---\n# adoption\n")
	return tracker.IssueRecord{ID: id, Card: &tracker.Record{ID: id, Path: "issues/" + id + ".md", BlobOID: "blob-" + id, Card: issue.Card{Frontmatter: fm}, Raw: raw}}
}

// Count the actual Git boundary rather than timing a scan on a loaded host.
// Doubling candidate records must not introduce per-record history rescans.
func TestAdoptionReadSetValidationHasBoundedGitReads(t *testing.T) {
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	counts := map[int]int{}
	for _, size := range []int{8, 16} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			cards, details := map[string]string{}, map[string]string{}
			for i := 0; i < size; i++ {
				cardPath, card, detailPath, detail := seededIssue(t, fmt.Sprintf("%06d", 640+i), fmt.Sprintf("candidate-%d", i))
				raw, err := issue.SetCardField([]byte(card), "actual_hours", "0.25")
				if err != nil {
					t.Fatal(err)
				}
				cards[cardPath], details[detailPath] = string(raw), detail
			}
			r := newTrackerRepo(t, cards, details)
			detailRoot := filepath.Join(r.root, "workshop/issues")
			records, err := loadIssueRecordsAt(withIssueRecordsScope(context.Background()), r.root, detailRoot, tracker.Fresh)
			if err != nil {
				t.Fatal(err)
			}
			since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
			candidates, err := adoptedExclusionCandidates(records.All(), "636", since, since.Add(time.Hour))
			if err != nil {
				t.Fatal(err)
			}
			if len(candidates) != size {
				t.Fatalf("fixture has %d candidates, want %d", len(candidates), size)
			}
			saved := captureAdoptionReadSet("636", detailRoot, since, since.Add(time.Hour), candidates)
			bin := t.TempDir()
			counter := filepath.Join(bin, "calls")
			body := "#!/bin/sh\nprintf x >> \"$ADOPTION_READSET_COUNTER\"\nexec \"$ADOPTION_READSET_REAL_GIT\" \"$@\"\n"
			if err := os.WriteFile(filepath.Join(bin, "git"), []byte(body), 0755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("ADOPTION_READSET_COUNTER", counter)
			t.Setenv("ADOPTION_READSET_REAL_GIT", realGit)
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			if err := validateObservedAdoptions(context.Background(), r.root, saved); err != nil {
				t.Fatal(err)
			}
			calls, err := os.ReadFile(counter)
			if err != nil {
				t.Fatal(err)
			}
			counts[size] = len(calls)
		})
	}
	if counts[8] == 0 || counts[16] > 2*counts[8]+2 {
		t.Fatalf("Git read growth: N=8: %d; 2N=16: %d", counts[8], counts[16])
	}
	t.Logf("Git boundary calls: N=8: %d; 2N=16: %d", counts[8], counts[16])
}

func TestAdoptionReadSetCandidateChanges(t *testing.T) {
	since := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	original := adoptionReadSetCard("000649", "0.50", "")
	initial := []tracker.IssueRecord{original, adoptionReadSetCard("000636", "1.0", ""), adoptionReadSetCard("000650", "N/A", "claimant: old\n")}
	candidates, err := adoptedExclusionCandidates(initial, "636", since, since.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	saved := captureAdoptionReadSet("636", "workshop/issues", since, since.Add(time.Hour), candidates)
	cases := []struct {
		name    string
		records []tracker.IssueRecord
		stale   bool
	}{
		{"unchanged", initial, false},
		{"own adoption changes", []tracker.IssueRecord{original, adoptionReadSetCard("000636", "2.0", "")}, false},
		{"unrelated claim changes", []tracker.IssueRecord{original, adoptionReadSetCard("000650", "N/A", "claimant: new\n")}, false},
		{"new adoption", []tracker.IssueRecord{original, adoptionReadSetCard("000650", "0.25", "")}, true},
		{"corrected adoption", []tracker.IssueRecord{adoptionReadSetCard("000649", "0.75", "")}, true},
		{"completion changes", []tracker.IssueRecord{adoptionReadSetCard("000649", "0.50", "completion: revised\n")}, true},
		{"removed adoption", nil, true},
		{"unreadable card", []tracker.IssueRecord{original, {ID: "000651", CardErr: errors.New("bad YAML")}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := compareAdoptionReadSet(saved, tc.records)
			if (err != nil) != tc.stale {
				t.Fatalf("stale=%v, err=%v", tc.stale, err)
			}
		})
	}
	// The capture must own its bytes, not retain a mutable tracker view.
	original.Card.Raw[0] = '!'
	if err := compareAdoptionReadSet(saved, []tracker.IssueRecord{adoptionReadSetCard("000649", "0.50", "")}); err != nil {
		t.Fatalf("capture aliased source: %v", err)
	}
}

func TestAdoptionReadSetCandidateFilterRemainsConservative(t *testing.T) {
	since := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	recent := adoptionReadSetCard("000649", "0.5", "")
	old := adoptionReadSetCard("000650", "0.5", "")
	old.Card.Card.Frontmatter = strings.ReplaceAll(old.Card.Card.Frontmatter, "2026-10-14", "2026-09-01")
	malformed := tracker.IssueRecord{ID: "000651", CardErr: errors.New("bad YAML")}
	got, err := adoptedExclusionCandidates([]tracker.IssueRecord{recent, old}, "636", since, since.Add(time.Hour))
	if err != nil || len(got) != 1 || got[0].ID != recent.ID {
		t.Fatalf("candidates=%v err=%v", got, err)
	}
	if _, err := adoptedExclusionCandidates([]tracker.IssueRecord{malformed}, "636", since, since.Add(time.Hour)); err == nil || !strings.Contains(err.Error(), "651") {
		t.Fatalf("unreadable authority silently skipped: %v", err)
	}
}

// A real private tracker changes between preparation and admission. Keeping a
// populated preparation scope proves validation cannot reuse its cached view.
func TestAdoptionReadSetFreshValidation(t *testing.T) {
	ownPath, ownCard, ownDetail, ownBody := seededIssue(t, "000636", "candidate")
	otherPath, otherCard, otherDetail, otherBody := seededIssue(t, "000649", "other")
	r := newTrackerRepo(t, map[string]string{ownPath: ownCard, otherPath: otherCard}, map[string]string{ownDetail: ownBody, otherDetail: otherBody})
	ctx := withIssueRecordsScope(context.Background())
	details := filepath.Join(r.root, "workshop/issues")
	if _, err := loadIssueRecordsAt(ctx, r.root, details, tracker.Fresh); err != nil {
		t.Fatal(err)
	}
	current := &ownedActualInputs{Snapshot: gatestate.MeasurementInputs{Since: "2026-09-01T00:00:00Z", Until: "2026-09-01T01:00:00Z"}}
	var saved adoptionReadSet
	if _, err := loadAdoptedExclusionsWithReadSet(ctx, r.root, "636", "", current, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.cards == nil || len(saved.cards) != 0 {
		t.Fatalf("original read not captured: %+v", saved)
	}
	if err := validateObservedAdoptions(ctx, r.root, &saved); err != nil {
		t.Fatal(err)
	}
	repo, err := tracker.NewRepository(context.Background(), r.root, "origin")
	if err != nil {
		t.Fatal(err)
	}
	snap, err := repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	card, ok := snap.Card("000649")
	if !ok {
		t.Fatal("missing other card")
	}
	raw, err := issue.SetCardField(card.Raw, "actual_hours", "0.25")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateCard(card, raw, "adoption-after-preparation", func(string, string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	// This error is about authority movement, not replay: an incomplete numeric
	// card still must invalidate preparation before a writer can use the result.
	if err := validateObservedAdoptions(ctx, r.root, &saved); err == nil || !strings.Contains(err.Error(), "649") {
		t.Fatalf("new adoption escaped fresh validation: %v", err)
	}
}

// Exercise the same fresh observation and pure decision used by the close
// guard, without duplicating either policy in this test helper.
func validateObservedAdoptions(ctx context.Context, root string, saved *adoptionReadSet) error {
	records, err := observeClosePreparationCards(ctx, root, saved.details)
	if err != nil {
		return err
	}
	if !records.Tracker {
		return errors.New("tracker evidence unavailable")
	}
	return compareAdoptionReadSet(saved, records.All())
}
