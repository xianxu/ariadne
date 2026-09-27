package tracker

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

const (
	activeDir  = "workshop/issues/"
	archiveDir = "workshop/history/issues/"
	oidA       = "1111111111111111111111111111111111111111"
	oidB       = "2222222222222222222222222222222222222222"
	oidP       = "3333333333333333333333333333333333333333"
)

func legacy(id, status, extra, body string) []byte {
	return []byte(fmt.Sprintf("---\nid: %s\nstatus: %s\ncreated: 2026-09-01\n%s---\n\n%s", id, status, extra, body))
}

func activeFile(id, slug, status string) MigrationFile {
	extra := ""
	if status == "codecomplete" || status == "done" {
		extra = "actual_hours: 1\n"
	}
	return MigrationFile{Path: activeDir + id + "-" + slug + ".md", Raw: legacy(id, status, extra, "# "+slug+"\n\n## Problem\n\nr\n")}
}

func archivedFile(id, slug, created string) MigrationFile {
	return MigrationFile{Path: archiveDir + id + "-" + slug + ".md", Raw: []byte("---\nid: " + id + "\nstatus: done\ncreated: " + created + "\n---\n\n# " + slug + "\n")}
}

func basicInput() MigrationInput {
	return MigrationInput{Repository: "file:///repo", ObjectFormat: "sha1", Main: oidA}
}

// The planned tracker tree must be exactly what the tracker reader accepts,
// seed max-ID from every used ID, and not depend on inventory order.
func TestPlanTrackerMigrationOverGeneratedPopulations(t *testing.T) {
	rng := rand.New(rand.NewSource(252))
	statuses := []string{"open", "working", "blocked"}
	for round := 0; round < 60; round++ {
		in := basicInput()
		maxID := 0
		for id := 1; id <= 1+rng.Intn(40); id++ {
			pid := fmt.Sprintf("%06d", id)
			switch rng.Intn(3) {
			case 0:
				in.Active = append(in.Active, activeFile(pid, "a"+pid, statuses[rng.Intn(len(statuses))]))
			case 1:
				in.Archived = append(in.Archived, archivedFile(pid, "h"+pid, "2026-05-01"))
				if rng.Intn(4) == 0 { // an archived duplicate
					in.Archived = append(in.Archived, archivedFile(pid, "d"+pid, "2026-06-01"))
				}
			default:
				continue // a gap: never used
			}
			maxID = id
		}
		m := PlanTrackerMigration(in)
		if len(m.Refusals) != 0 {
			t.Fatalf("round %d: refusals %+v", round, m.Refusals)
		}
		var files []gitx.TreeFile
		for p, raw := range m.TrackerFiles() {
			oid, _ := issue.CardBlobOID(raw, "sha1")
			files = append(files, gitx.TreeFile{Path: p, Mode: "100644", OID: oid, Content: raw})
		}
		snap, err := parseSnapshot("planned", files)
		if err != nil {
			t.Fatalf("round %d: planned tracker rejected: %v", round, err)
		}
		if snap.MaxID() != maxID {
			t.Fatalf("round %d: max ID %d, want %d", round, snap.MaxID(), maxID)
		}
		if len(m.Conversions) != len(in.Active) {
			t.Fatalf("round %d: %d conversions for %d active", round, len(m.Conversions), len(in.Active))
		}
		for _, c := range m.Conversions {
			if !issue.HasMirror(c.Raw) {
				t.Fatalf("round %d: %s converted without a mirror", round, c.Path)
			}
		}
		shuffled := in
		shuffled.Active = append([]MigrationFile(nil), in.Active...)
		shuffled.Archived = append([]MigrationFile(nil), in.Archived...)
		rng.Shuffle(len(shuffled.Active), func(i, j int) { shuffled.Active[i], shuffled.Active[j] = shuffled.Active[j], shuffled.Active[i] })
		rng.Shuffle(len(shuffled.Archived), func(i, j int) {
			shuffled.Archived[i], shuffled.Archived[j] = shuffled.Archived[j], shuffled.Archived[i]
		})
		if again := PlanTrackerMigration(shuffled); again.Digest != m.Digest {
			t.Fatalf("round %d: digest depends on inventory order", round)
		}
	}
}

func TestPlanTrackerMigrationDuplicateIDs(t *testing.T) {
	in := basicInput()
	in.Archived = []MigrationFile{archivedFile("000040", "old", "2026-05-27"), archivedFile("000040", "new", "2026-05-28"),
		archivedFile("000096", "archived", "2026-06-14")}
	in.Active = []MigrationFile{activeFile("000096", "active", "open")}
	m := PlanTrackerMigration(in)
	if len(m.Refusals) != 0 {
		t.Fatalf("refusals: %+v", m.Refusals)
	}
	paths := map[string]string{}
	for _, c := range m.Cards {
		paths[c.ID] = c.Path
	}
	if paths["000040"] != CardPath("000040", "new") || paths["000096"] != CardPath("000096", "active") {
		t.Fatalf("seeds: %v", paths)
	}
	if len(m.Duplicates) != 2 {
		t.Fatalf("duplicates not reported: %v", m.Duplicates)
	}
	in.Active = append(in.Active, activeFile("000096", "twin", "open"))
	if m := PlanTrackerMigration(in); len(m.Refusals) != 2 || !strings.Contains(m.Refusals[0].Next, "renumber") {
		t.Fatalf("two active files sharing an ID: %+v", m.Refusals)
	}
}

func TestPlanTrackerMigrationRefusesLegacyDivergence(t *testing.T) {
	in := basicInput()
	in.Active = []MigrationFile{activeFile("000001", "one", "open"), activeFile("000002", "two", "open")}
	in.Archived = []MigrationFile{archivedFile("000003", "three", "2026-05-01")}
	edited := activeFile("000001", "one", "open")
	edited.Raw = append(edited.Raw, []byte("\n## Log\n- branch-only design\n")...)
	in.Branches = []MigrationBranchFile{
		{Branch: "feat-a", Path: edited.Path, Raw: edited.Raw},                                                  // detail-owned edits: fine
		{Branch: "feat-b", Path: activeDir + "000002-two.md", Raw: activeFile("000002", "two", "working").Raw},  // unpublished claim
		{Branch: "feat-c", Path: activeDir + "000009-new.md", Raw: activeFile("000009", "new", "open").Raw},     // branch-only issue
		{Branch: "feat-d", Path: activeDir + "000003-three.md", Raw: activeFile("000003", "three", "open").Raw}, // edits an archived issue
		{Branch: "feat-e", Path: activeDir + "000001-one.md"},                                                   // a deletion: the branch's own archive move
	}
	in.DirtyIssuePaths = []string{"/slot2: workshop/issues/000001-one.md"}
	m := PlanTrackerMigration(in)
	got := map[string]string{}
	for _, r := range m.Refusals {
		got[r.Subject] = r.Reason
	}
	want := map[string]string{
		"feat-b: " + activeDir + "000002-two.md":   "status",
		"feat-c: " + activeDir + "000009-new.md":   "only on this branch",
		"feat-d: " + activeDir + "000003-three.md": "archived",
		"/slot2: workshop/issues/000001-one.md":    "uncommitted",
	}
	if len(got) != len(want) {
		t.Fatalf("refusals %v, want %v", got, want)
	}
	for subject, reason := range want {
		if !strings.Contains(got[subject], reason) {
			t.Errorf("%s: %q, want %q", subject, got[subject], reason)
		}
	}
}

func TestPlanTrackerMigrationBindsOnlyAProvableLegacyClose(t *testing.T) {
	cases := []struct {
		name    string
		anchors []MigrationAnchor
		refuse  string
	}{
		{"one close on a branch and its remote copy", []MigrationAnchor{{Ref: "feat", Anchor: oidB, Parent: oidP}, {Ref: "origin/feat", Anchor: oidB, Parent: oidP}}, ""},
		{"no close", nil, "0 legacy close"},
		{"two closes", []MigrationAnchor{{Ref: "a", Anchor: oidA, Parent: oidP}, {Ref: "b", Anchor: oidB, Parent: oidP}}, "2 legacy close"},
		{"code after the close", []MigrationAnchor{{Ref: "feat", Anchor: oidB, Parent: oidP, CodeAfter: true}}, "code after"},
		{"the branch close beside main's legacy publication of it", []MigrationAnchor{{Ref: "main", Anchor: oidA, Parent: oidP, OnMain: true}, {Ref: "feat", Anchor: oidB, Parent: oidP}}, ""},
		{"a close made directly on main", []MigrationAnchor{{Ref: "main", Anchor: oidB, Parent: oidP, OnMain: true}}, ""},
	}
	for _, c := range cases {
		in := basicInput()
		in.Active = []MigrationFile{activeFile("000005", "five", "codecomplete")}
		in.Anchors = map[string][]MigrationAnchor{"000005": c.anchors}
		m := PlanTrackerMigration(in)
		if c.refuse != "" {
			if len(m.Refusals) != 1 || !strings.Contains(m.Refusals[0].Reason, c.refuse) || len(m.Cards) != 0 {
				t.Errorf("%s: %+v", c.name, m.Refusals)
			}
			continue
		}
		if len(m.Refusals) != 0 || len(m.Cards) != 1 {
			t.Fatalf("%s: %+v", c.name, m.Refusals)
		}
		b, ok, err := issue.CardCompletion(m.Cards[0].Raw)
		if err != nil || !ok || b.Token != MigrationToken("000005") || b.EvidenceCommit != oidB || b.ReviewedHEAD != oidP || b.Repository != "file:///repo" {
			t.Fatalf("%s: binding %+v %v", c.name, b, err)
		}
	}
}
