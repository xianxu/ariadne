package main

import (
	"bytes"
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
	got, stale, err := listIssueStates(filepath.Join(r.root, "workshop/issues"))
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
	got, stale, err = listIssueStates(filepath.Join(r.root, "workshop/issues"))
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
	meta, err := lookupIssueMeta(filepath.Base(r.root)+"#9", r.root)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Status != "open" || meta.EstimateHours != 4 || len(meta.Deps) != 1 || meta.Deps[0] != "r#10" || meta.DepsUnknown {
		t.Fatalf("#9 meta: %+v", meta)
	}
	only, err := lookupIssueMeta(filepath.Base(r.root)+"#10", r.root)
	if err != nil || only.Status != "open" || !only.DepsUnknown {
		t.Fatalf("card-only meta: %+v %v", only, err)
	}
}
