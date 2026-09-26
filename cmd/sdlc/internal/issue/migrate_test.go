package issue

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func legacyActive(fm, body string) []byte { return []byte("---\n" + fm + "\n---\n\n" + body) }

// Every generated legacy active file converts into a card and mirrored details
// the rest of the system accepts, and reconciling an unchanged branch copy of
// the same file yields exactly main's conversion — so the later merge applies
// identical changes on both sides.
func TestMigrateActiveDetailsRoundTripsOverGeneratedLegacyFiles(t *testing.T) {
	fields := []string{"estimate_hours: 3", "started: 2026-09-01T10:00:00Z", "github_issue: 12", "deps: [000001]", "references: [x]"}
	bodies := map[string]string{
		"problem":     "# Title\n\n## Problem\n\nA report.\n\n## Spec\n\nS.\n",
		"preamble":    "# Title\n\nThe report as a preamble.\n\nMore of it.\n\n## Done when\n\n- x\n",
		"no-preamble": "# Title\n\n## Done when\n\n- x\n",
		"fenced":      "# Title\n\nPreamble.\n\n```\n## Problem\n```\n\n## Plan\n",
	}
	statuses := []string{"open", "working", "blocked", "codecomplete"}
	for mask := 0; mask < 1<<len(fields); mask++ {
		for name, body := range bodies {
			for _, status := range statuses {
				fm := []string{"id: 000042", "status: " + status, "created: 2026-09-01", "updated: 2026-09-02"}
				if status == "codecomplete" {
					fm = append(fm, "actual_hours: 1.5")
				}
				for i, f := range fields {
					if mask&(1<<i) != 0 {
						fm = append(fm, f)
					}
				}
				label := fmt.Sprintf("%s/%s/%b", name, status, mask)
				legacy := legacyActive(strings.Join(fm, "\n"), body)
				card, mirrored, inferences, err := MigrateActiveDetails(legacy, "sha1")
				if err != nil {
					t.Fatalf("%s: %v", label, err)
				}
				if _, err := ParseCard(card); err != nil {
					t.Fatalf("%s: card invalid: %v", label, err)
				}
				oid, _ := CardBlobOID(card, "sha1")
				if got, err := MirrorBaselineOID(mirrored); err != nil || got != oid {
					t.Fatalf("%s: mirror %q (%v), want %s", label, got, err, oid)
				}
				if _, err := RefreshMirror(mirrored, card, card); err != nil {
					t.Fatalf("%s: mirrored details refuse their own card: %v", label, err)
				}
				if (name == "problem") != (len(inferences) == 0) {
					t.Fatalf("%s: inferences %v", label, inferences)
				}
				reconciled, err := ReconcileLegacyDetails(legacy, card, "sha1")
				if err != nil || string(reconciled) != string(mirrored) {
					t.Fatalf("%s: branch reconcile differs from main's conversion (%v):\n%s\n---\n%s", label, err, reconciled, mirrored)
				}
			}
		}
	}
}

func TestMigrateActiveDetailsNormalizesOnlyTheProblemHeading(t *testing.T) {
	legacy := legacyActive("id: 000042\nstatus: open", "# Title\n\nThe report.\n\n## Done when\n\n- x\n")
	card, mirrored, _, err := MigrateActiveDetails(legacy, "sha1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(card), "## Problem\n\nThe report.") {
		t.Fatalf("card Problem is not the preamble:\n%s", card)
	}
	if !strings.Contains(string(mirrored), "# Title\n\n## Problem\n\nThe report.\n\n## Done when") {
		t.Fatalf("details heading not inserted above the preamble:\n%s", mirrored)
	}
	bare := legacyActive("id: 000042\nstatus: open", "# Title\n\n## Done when\n")
	card, _, inferences, err := MigrateActiveDetails(bare, "sha1")
	if err != nil || !strings.Contains(string(card), NoProblemPlaceholder) || len(inferences) != 1 || !strings.Contains(inferences[0], "placeholder") {
		t.Fatalf("placeholder: %v %v\n%s", err, inferences, card)
	}
	for name, bad := range map[string][]byte{
		"status":       legacyActive("id: 000042\nstatus: shipped", "# T\n\n## Problem\nx\n"),
		"closed hours": legacyActive("id: 000042\nstatus: done", "# T\n\n## Problem\nx\n"),
		"no title":     legacyActive("id: 000042\nstatus: open", "Just text.\n"),
		"mirrored":     legacyActive("id: 000042\nstatus: open\ncard_mirror: 'abc'", "# T\n\n## Problem\nx\n"),
	} {
		if _, _, _, err := MigrateActiveDetails(bad, "sha1"); err == nil {
			t.Errorf("%s: migrated invalid active details", name)
		}
	}
}

func TestArchivedCardReportsEveryInference(t *testing.T) {
	cases := []struct {
		name, file, raw string
		want            []string // substrings of the card
		infer           []string // substrings of inferences, in order
	}{
		{"clean", "000007-seven.md", "---\nid: 000007\nstatus: done\nactual_hours: 2\ncreated: 2026-05-01\n---\n\n# Seven\n\n## Problem\n\nBroke.\n",
			[]string{"status: done", "actual_hours: 2", "created: 2026-05-01", "# Seven", "Broke."}, nil},
		{"no hours", "000007-seven.md", "---\nid: 000007\nstatus: done\n---\n\n# Seven\n\nPreamble report.\n\n## Log\n",
			[]string{"actual_hours: N/A", "Preamble report."}, []string{"N/A", "Problem from the preamble"}},
		{"bad hours and date", "000007-seven.md", "---\nid: 000007\nstatus: done\nactual_hours: ~2h\ncreated: May 1\n---\n\n# Seven\n",
			[]string{"actual_hours: N/A"}, []string{"created", "unreadable: N/A", "Problem points at"}},
		{"no frontmatter", "000060-q.md", "# Q\n\nquestion\n",
			[]string{"id: '000060'", "status: done", "# Q"}, []string{"no frontmatter", "status inferred done", "N/A", "preamble"}},
		{"slug title", "000109-cross-file-markers.md", "---\nid: 000109\nstatus: done\nactual_hours: 1\n---\n\nno heading\n",
			[]string{"# cross file markers"}, []string{"title from the slug", "Problem points at"}},
		{"two titles", "000090-refactor.md", "---\nid: 000090\nstatus: done\nactual_hours: 1\n---\n\n# First\n\nwhy\n\n# Second\n",
			[]string{"# First", "why"}, []string{"first of", "preamble"}},
		{"id mismatch", "000040-two.md", "---\nid: 000041\nstatus: done\nactual_hours: 1\n---\n\n# Two\n\n## Problem\nx\n",
			[]string{"id: '000040'"}, []string{"replaced by the filename"}},
		{"open in archive", "000008-eight.md", "---\nid: 000008\nstatus: open\n---\n\n# Eight\n\n## Problem\nx\n",
			[]string{"status: open"}, nil},
	}
	for _, c := range cases {
		card, inferences, err := ArchivedCard([]byte(c.raw), c.file, "sha1")
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if _, err := ParseCard(card); err != nil {
			t.Fatalf("%s: invalid card: %v\n%s", c.name, err, card)
		}
		for _, w := range c.want {
			if !strings.Contains(string(card), w) {
				t.Errorf("%s: card lacks %q:\n%s", c.name, w, card)
			}
		}
		if len(inferences) != len(c.infer) {
			t.Fatalf("%s: inferences %q, want %d like %q", c.name, inferences, len(c.infer), c.infer)
		}
		for i, w := range c.infer {
			if !strings.Contains(inferences[i], w) {
				t.Errorf("%s: inference %d = %q, want %q", c.name, i, inferences[i], w)
			}
		}
	}
	if _, _, err := ArchivedCard([]byte("---\nid: 1\nid: 2\n---\n# x\n"), "000001-x.md", "sha1"); err == nil {
		t.Error("duplicate keys accepted")
	}
}

func TestReconcileLegacyDetailsRefusesDivergedCardFields(t *testing.T) {
	main := legacyActive("id: 000042\nstatus: working\nupdated: 2026-09-02", "# Title\n\n## Problem\n\nA report.\n")
	card, _, _, err := MigrateActiveDetails(main, "sha1")
	if err != nil {
		t.Fatal(err)
	}
	branch := legacyActive("id: 000042\nstatus: codecomplete\nactual_hours: 1\nupdated: 2026-09-03", "# Title\n\n## Problem\n\nA report.\n")
	if _, err := ReconcileLegacyDetails(branch, card, "sha1"); err == nil || !strings.Contains(err.Error(), "actual_hours, status, updated") {
		t.Fatalf("diverged branch reconciled: %v", err)
	}
	edited := legacyActive("id: 000042\nstatus: working\nupdated: 2026-09-02\ndeps: [000001]", "# Title\n\n## Problem\n\nA revised report.\n\n## Log\n- more\n")
	out, err := ReconcileLegacyDetails(edited, card, "sha1")
	if err != nil {
		t.Fatalf("detail-owned edits must reconcile: %v", err)
	}
	if !strings.Contains(string(out), "A revised report.") || !strings.Contains(string(out), "deps: [000001]") {
		t.Fatalf("reconcile lost branch edits:\n%s", out)
	}
	if _, err := ReconcileLegacyDetails(out, card, "sha1"); !errors.Is(err, ErrNotLegacy) {
		t.Fatalf("mirrored details reconciled again: %v", err)
	}
}
