package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTickMilestoneTaskRow_Match(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
		n    int
	}{
		{
			"unchecked",
			"- [ ] do work [ariadne#31 M1]\n",
			"- [x] do work [ariadne#31 M1]\n",
			1,
		},
		{
			"in_progress_dot",
			"- [.] do work [ariadne#31 M1]\n",
			"- [x] do work [ariadne#31 M1]\n",
			1,
		},
		{
			"blocked_dash",
			"- [-] do work [ariadne#31 M1]\n",
			"- [x] do work [ariadne#31 M1]\n",
			1,
		},
		{
			"cancelled_tilde",
			"- [~] do work [ariadne#31 M1]\n",
			"- [x] do work [ariadne#31 M1]\n",
			1,
		},
		{
			"already_x_no_change",
			"- [x] do work [ariadne#31 M1]\n",
			"- [x] do work [ariadne#31 M1]\n",
			0,
		},
		{
			"different_milestone_no_match",
			"- [ ] do work [ariadne#31 M2]\n",
			"- [ ] do work [ariadne#31 M2]\n",
			0,
		},
		{
			"different_repo_no_match",
			"- [ ] do work [nous#31 M1]\n",
			"- [ ] do work [nous#31 M1]\n",
			0,
		},
		{
			"superstring_milestone_no_match",
			// M1-extra must NOT match M1
			"- [ ] do work [ariadne#31 M1-extra]\n",
			"- [ ] do work [ariadne#31 M1-extra]\n",
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, n := TickMilestoneTaskRow(tt.in, "ariadne", "31", "M1")
			if got != tt.want {
				t.Errorf("got %q want %q", got, tt.want)
			}
			if n != tt.n {
				t.Errorf("n = %d want %d", n, tt.n)
			}
		})
	}
}

func TestTickAllTaskRowsForIssue(t *testing.T) {
	in := "- [ ] M1 task [ariadne#31 M1]\n" +
		"- [.] M2 task [ariadne#31 M2]\n" +
		"- [x] done task [ariadne#31 M3]\n" +
		"- [-] cancelled [ariadne#31 M4]\n" +
		"- [ ] other issue [ariadne#99 M1]\n" +
		"- [ ] bare ref [ariadne#31]\n"
	want := "- [x] M1 task [ariadne#31 M1]\n" +
		"- [x] M2 task [ariadne#31 M2]\n" +
		"- [x] done task [ariadne#31 M3]\n" +
		// cancelled stays cancelled — character class is [ .] not [ .\-~]
		"- [-] cancelled [ariadne#31 M4]\n" +
		"- [ ] other issue [ariadne#99 M1]\n" +
		"- [x] bare ref [ariadne#31]\n"
	got, n := TickAllTaskRowsForIssue(in, "ariadne", "31")
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
	// 3 ticks: M1 + M2 + bare ref (M3 already x, M4 cancelled not in class, #99 different issue)
	if n != 3 {
		t.Errorf("n = %d want 3", n)
	}
}

func TestUpsertDetailBlockFields_FieldPresent_Replaces(t *testing.T) {
	doc := `## details

<a id="ariadne-31-m1"></a>
### ariadne#31 M1 — port close-issue

**est:** 4h
**actual:** 0h
**closed:** TBD

shipped: port done.

<a id="ariadne-31-m2"></a>
### ariadne#31 M2 — state

**est:** 3h
`
	out, found := UpsertDetailBlockFields(doc, "ariadne-31-m1", map[string]string{
		"actual": "6.5h",
		"closed": "2026-05-25",
	})
	if !found {
		t.Fatal("expected found=true")
	}
	if !strings.Contains(out, "**actual:** 6.5h") {
		t.Errorf("expected '**actual:** 6.5h' in output:\n%s", out)
	}
	if !strings.Contains(out, "**closed:** 2026-05-25") {
		t.Errorf("expected '**closed:** 2026-05-25' in output:\n%s", out)
	}
	if strings.Contains(out, "**actual:** 0h") {
		t.Errorf("old actual still present:\n%s", out)
	}
	if strings.Contains(out, "**closed:** TBD") {
		t.Errorf("old closed still present:\n%s", out)
	}
	// M2 block must remain untouched.
	if !strings.Contains(out, "<a id=\"ariadne-31-m2\"></a>\n### ariadne#31 M2 — state\n\n**est:** 3h\n") {
		t.Errorf("M2 block was perturbed:\n%s", out)
	}
}

func TestUpsertDetailBlockFields_FieldAbsent_InsertsAfterEst(t *testing.T) {
	doc := `<a id="ariadne-31-m1"></a>
### ariadne#31 M1 — port close-issue

**est:** 4h

shipped: port done.
`
	out, found := UpsertDetailBlockFields(doc, "ariadne-31-m1", map[string]string{
		"actual": "6.5h",
	})
	if !found {
		t.Fatal("expected found=true")
	}
	// actual should appear immediately after est, no other field in between.
	idx := strings.Index(out, "**est:** 4h\n**actual:** 6.5h")
	if idx == -1 {
		t.Errorf("expected **actual:** to follow **est:** with no blank line:\n%s", out)
	}
}

func TestUpsertDetailBlockFields_AnchorMissing_FoundFalse(t *testing.T) {
	doc := `<a id="ariadne-31-m2"></a>
### ariadne#31 M2 — state

**est:** 3h
`
	out, found := UpsertDetailBlockFields(doc, "ariadne-31-m1", map[string]string{"actual": "1h"})
	if found {
		t.Errorf("expected found=false, got true")
	}
	if out != doc {
		t.Errorf("text was modified when not found")
	}
}

func TestAnchorFor(t *testing.T) {
	tests := []struct {
		repo, id, ms, want string
	}{
		{"ariadne", "31", "M1", "ariadne-31-m1"},
		{"ariadne", "31", "M4b", "ariadne-31-m4b"},
		{"charon", "13", "M2 review", "charon-13-m2-review"},
	}
	for _, tt := range tests {
		if got := AnchorFor(tt.repo, tt.id, tt.ms); got != tt.want {
			t.Errorf("AnchorFor(%q,%q,%q) = %q want %q", tt.repo, tt.id, tt.ms, got, tt.want)
		}
	}
}

func TestFindTaskTitle(t *testing.T) {
	doc := "- [x] port close-issue to Go [ariadne#31 M1]\n"
	if got := FindTaskTitle(doc, "ariadne", "31", "M1"); got != "port close-issue to Go" {
		t.Errorf("got %q want %q", got, "port close-issue to Go")
	}
	// em-dash trimming (Python's .strip(' —')):
	doc2 := "- [x] — port close-issue — [ariadne#31 M1]\n"
	if got := FindTaskTitle(doc2, "ariadne", "31", "M1"); got != "port close-issue" {
		t.Errorf("got %q want %q", got, "port close-issue")
	}
}

func TestFindByIssueRef(t *testing.T) {
	dir := t.TempDir()
	projectsDir := filepath.Join(dir, "data", "project")
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite := func(name, body string) {
		if err := os.WriteFile(filepath.Join(projectsDir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("alpha.md", "irrelevant content\n")
	mustWrite("beta.md", "## tasks\n\n- [ ] thing [ariadne#31 M1]\n")
	mustWrite("gamma.md", "## tasks\n\n- [ ] thing [nous#31]\n")

	t.Run("one match", func(t *testing.T) {
		got, err := FindByIssueRef(dir, "ariadne", "31")
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Base(got) != "beta.md" {
			t.Errorf("got %q want .../beta.md", got)
		}
	})
	t.Run("no match", func(t *testing.T) {
		got, err := FindByIssueRef(dir, "ariadne", "99")
		if err != nil {
			t.Fatal(err)
		}
		if got != "" {
			t.Errorf("got %q want \"\"", got)
		}
	})
	t.Run("multiple matches", func(t *testing.T) {
		mustWrite("delta.md", "## tasks\n\n- [ ] thing [ariadne#31 M2]\n")
		_, err := FindByIssueRef(dir, "ariadne", "31")
		if err == nil {
			t.Errorf("expected error for multiple matches")
		}
	})
}

func TestSkeleton_Render(t *testing.T) {
	s := Skeleton{
		Anchor:    "ariadne-31-m1",
		RefLabel:  "ariadne#31 M1",
		Title:     "port close-issue",
		Est:       "4h",
		Actual:    "6.5h",
		ClosedISO: "2026-05-25",
	}
	skel, ref := s.Render()
	if !strings.Contains(skel, "<a id=\"ariadne-31-m1\"></a>") {
		t.Errorf("skeleton missing anchor: %q", skel)
	}
	if !strings.Contains(skel, "### ariadne#31 M1 — port close-issue") {
		t.Errorf("skeleton missing heading: %q", skel)
	}
	if !strings.Contains(skel, "**est:** 4h\n**actual:** 6.5h\n**closed:** 2026-05-25\n") {
		t.Errorf("skeleton missing fields block: %q", skel)
	}
	if ref != "[ariadne#31 M1]: #ariadne-31-m1" {
		t.Errorf("refDef = %q", ref)
	}
}

func TestSkeleton_Render_Fallbacks(t *testing.T) {
	s := Skeleton{
		Anchor:    "ariadne-31-m1",
		RefLabel:  "ariadne#31 M1",
		Actual:    "6.5h",
		ClosedISO: "2026-05-25",
	}
	skel, _ := s.Render()
	if !strings.Contains(skel, "<title for milestone>") {
		t.Errorf("expected title fallback, got: %q", skel)
	}
	if !strings.Contains(skel, "<copy from issue estimate_hours, or omit>") {
		t.Errorf("expected est fallback, got: %q", skel)
	}
}
