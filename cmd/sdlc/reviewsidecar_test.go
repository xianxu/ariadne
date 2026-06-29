package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSidecarPath(t *testing.T) {
	cases := []struct{ name, issueFile, milestone, want string }{
		{"close", "000136-review-sidecar.md", "", "workshop/plans/000136-review-sidecar-close-review.md"},
		{"milestone M2", "000136-review-sidecar.md", "M2", "workshop/plans/000136-review-sidecar-m2-review.md"},
		{"milestone lowercased", "000069-x.md", "M4b", "workshop/plans/000069-x-m4b-review.md"},
		{"absolute issue path uses basename", "/tmp/issues/000069-x.md", "", "workshop/plans/000069-x-close-review.md"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := sidecarPath("workshop/plans", c.issueFile, c.milestone); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestRenderReviewEntry(t *testing.T) {
	m := sidecarMeta{
		IssueNum: 136, Title: "sdlc boundary review sidecar", Repo: "ariadne",
		IssueFile: "workshop/issues/000136-review-sidecar.md", Milestone: "",
		Base: "abc1234def", Head: "HEAD", Command: "sdlc close --issue 136",
		Agent: "claude", Timestamp: "2026-06-29T15:40:00-07:00",
		Verdict: "SHIP", Body: "VERDICT: SHIP\n\nLooks good.",
	}

	doc := renderReviewEntry(m, false)
	for _, want := range []string{
		"# Boundary Review — ariadne#136 (whole-issue close)",
		"136 — sdlc boundary review sidecar", "ariadne",
		"workshop/issues/000136-review-sidecar.md", "abc1234def..HEAD",
		"sdlc close --issue 136", "claude", "2026-06-29T15:40:00-07:00",
		"SHIP", "Looks good.", "## Review",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("initial doc missing %q\n---\n%s", want, doc)
		}
	}

	// Milestone boundary renders its tag in the H1, the boundary cell, and the milestone cell.
	mDoc := renderReviewEntry(sidecarMeta{IssueNum: 69, Milestone: "M2", Verdict: "SHIP", Body: "x"}, false)
	if !strings.Contains(mDoc, "milestone M2") {
		t.Errorf("milestone doc should label the boundary as 'milestone M2':\n%s", mDoc)
	}

	rev := renderReviewEntry(m, true)
	if !strings.Contains(rev, "## Re-review — 2026-06-29T15:40:00-07:00 (SHIP)") {
		t.Errorf("revision should head with '## Re-review — <ts> (<verdict>)':\n%s", rev)
	}
	if strings.Contains(rev, "# Boundary Review") {
		t.Errorf("revision section must not repeat the H1:\n%s", rev)
	}
}

func TestWriteReviewSidecar_CreateThenAppend(t *testing.T) {
	dir := t.TempDir()
	issues := filepath.Join(dir, "issues")
	plans := filepath.Join(dir, "plans")
	if err := os.MkdirAll(issues, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(issues, "000136-review-sidecar.md"),
		[]byte("---\nid: 000136\nstatus: working\n---\n# sdlc boundary review sidecar\n## Plan\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	p := boundaryReviewParams{
		IssueNum: 136, Milestone: "", IssuesDir: issues, PlansDir: plans,
		BaseLong: "abc1234", Head: "HEAD", Agent: "claude",
	}

	path1, err := writeReviewSidecar(p, "SHIP", "first review body", "2026-06-29T15:40:00-07:00")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path1) != "000136-review-sidecar-close-review.md" {
		t.Errorf("unexpected sidecar name: %s", filepath.Base(path1))
	}
	// Title is read from the issue file's `# ` line.
	if data, _ := os.ReadFile(path1); !strings.Contains(string(data), "sdlc boundary review sidecar") {
		t.Errorf("first write should embed the issue title")
	}

	path2, err := writeReviewSidecar(p, "FIX-THEN-SHIP", "second review body", "2026-06-29T16:00:00-07:00")
	if err != nil {
		t.Fatal(err)
	}
	if path2 != path1 {
		t.Errorf("re-run should target the same file: %s vs %s", path2, path1)
	}

	data, err := os.ReadFile(path1)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{"first review body", "second review body", "## Re-review", "FIX-THEN-SHIP"} {
		if !strings.Contains(s, want) {
			t.Errorf("re-run output missing %q — prior evidence must be preserved\n---\n%s", want, s)
		}
	}
	// The initial body must survive (no silent overwrite).
	if strings.Count(s, "# Boundary Review") != 1 {
		t.Errorf("expected exactly one H1 after append, got %d", strings.Count(s, "# Boundary Review"))
	}
	// No temp file left behind (atomic write).
	if leaks, _ := filepath.Glob(filepath.Join(plans, "*.tmp")); len(leaks) != 0 {
		t.Errorf("temp file leaked: %v", leaks)
	}
}

func TestWriteReviewSidecar_MissingIssueErrors(t *testing.T) {
	dir := t.TempDir()
	p := boundaryReviewParams{IssueNum: 999, IssuesDir: filepath.Join(dir, "issues"), PlansDir: filepath.Join(dir, "plans")}
	if _, err := writeReviewSidecar(p, "SHIP", "body", "2026-06-29T15:40:00-07:00"); err == nil {
		t.Error("writing a sidecar for a non-existent issue should error")
	}
}
