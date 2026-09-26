package main

import (
	"bytes"
	"context"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"path/filepath"
	"strings"
	"testing"
)

func TestListIssuesComposesCardsWithDetails(t *testing.T) {
	cp9, c9, dp9, d9 := seededIssue(t, "000009", "nine")
	cp10, c10, _, _ := seededIssue(t, "000010", "card-only")
	cp11, c11, _, _ := seededIssue(t, "000011", "archived")
	c11 = strings.Replace(c11, "status: open", "status: done\nactual_hours: 1", 1)
	r := newTrackerRepo(t, map[string]string{cp9: c9, cp10: c10, cp11: c11}, map[string]string{dp9: d9})
	retitleElsewhere(t, r, "000009", "Card Title Wins")
	got, stale, err := listIssueStates(context.Background(), filepath.Join(r.root, "workshop/issues"))
	if err != nil || stale {
		t.Fatalf("list: %v stale=%v", err, stale)
	}
	if len(got) != 2 {
		t.Fatalf("want #9 and card-only #10, got %+v", got)
	}
	if got[0].ID != "000009" || got[0].Title != "Card Title Wins" || got[0].CardOnly {
		t.Errorf("#9 did not take its title from the card: %+v", got[0])
	}
	if got[1].ID != "000010" || !got[1].CardOnly || got[1].Status != "open" || got[1].Path != "" {
		t.Errorf("card-only #10: %+v", got[1])
	}

	git(t, r.root, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))
	got, stale, err = listIssueStates(context.Background(), filepath.Join(r.root, "workshop/issues"))
	if err != nil || !stale || len(got) != 2 || got[0].Title != "Card Title Wins" {
		t.Fatalf("offline read: %+v stale=%v err=%v", got, stale, err)
	}
	var out bytes.Buffer
	if err := renderProse(&out, State{Issues: got, TrackerStale: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "tracker unreachable") || !strings.Contains(out.String(), "card only") {
		t.Errorf("state did not label stale cards or card-only rows:\n%s", out.String())
	}
}

func TestProjectIssueMetaReadsCardFieldsAndDetailDeps(t *testing.T) {
	cp9, c9, dp9, d9 := seededIssue(t, "000009", "nine")
	c9 = strings.Replace(c9, "estimate_hours:", "estimate_hours: 4", 1)
	d9 = strings.Replace(d9, "deps: []", "deps: [r#10]", 1)
	d9 = strings.Replace(d9, "estimate_hours:", "estimate_hours: 99", 1) // stale mirror, must not count
	cp10, c10, _, _ := seededIssue(t, "000010", "card-only")
	r := newTrackerRepo(t, map[string]string{cp9: c9, cp10: c10}, map[string]string{dp9: d9})
	meta, err := lookupIssueMeta(context.Background(), filepath.Base(r.root)+"#9", r.root)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Status != "open" || meta.EstimateHours != 4 || len(meta.Deps) != 1 || meta.Deps[0] != "r#10" || meta.DepsUnknown {
		t.Fatalf("#9 meta: %+v", meta)
	}
	only, err := lookupIssueMeta(context.Background(), filepath.Base(r.root)+"#10", r.root)
	if err != nil || only.Status != "open" || !only.DepsUnknown {
		t.Fatalf("card-only meta: %+v %v", only, err)
	}
}

func TestActualTrackerInputsUseTheCardStamp(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err != nil {
		t.Fatal(err)
	}
	refs, started, carded := actualTrackerInputs(context.Background(), r.root, "9")
	if !carded || started == "" || len(refs) != 1 || refs[0] != "refs/remotes/origin/issue-tracker" {
		t.Fatalf("refs %v started %q carded %v", refs, started, carded)
	}
	if !strings.Contains(r.card(cardPath), "started: "+started) {
		t.Fatalf("started %q is not the card's stamp", started)
	}
}

// BR-24: one tracker fetch per command. Within a command's records scope a
// second read is served without network IO; a card write invalidates it; a
// stale (offline) read never satisfies a fresh request.
func TestRecordsScopeFetchesOncePerCommand(t *testing.T) {
	cardPath, card, _, _ := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, nil)
	dir := filepath.Join(r.root, "workshop/issues")
	ctx := withIssueRecordsScope(context.Background())
	if _, err := loadIssueRecords(ctx, dir, tracker.Fresh); err != nil {
		t.Fatal(err)
	}
	git(t, r.root, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))
	if rs, err := loadIssueRecords(ctx, dir, tracker.Fresh); err != nil || !rs.Tracker {
		t.Fatalf("second read in the command fetched again: %v", err)
	}
	if _, err := loadIssueRecords(context.Background(), dir, tracker.Fresh); err == nil {
		t.Fatal("fixture: an unscoped fresh read should need the (gone) remote")
	}
	invalidateIssueRecords(ctx)
	if _, err := loadIssueRecords(ctx, dir, tracker.Fresh); err == nil {
		t.Fatal("a read after a card write reused the invalidated view")
	}
	stale, err := loadIssueRecords(ctx, dir, tracker.PreferFresh)
	if err != nil || !stale.Stale {
		t.Fatalf("offline preferred read: %+v %v", stale, err)
	}
	if _, err := loadIssueRecords(ctx, dir, tracker.Fresh); err == nil {
		t.Fatal("a cached stale view satisfied a fresh request")
	}
}
