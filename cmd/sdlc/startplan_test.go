package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartPlanCmd_Registered(t *testing.T) {
	cmd := NewStartPlanCmd()
	if cmd.Flags().Lookup("issue") == nil {
		t.Error("start-plan command missing --issue flag")
	}
}

// #75 M2: start-plan delivers the at-plan architecture lens (the forward
// injection) to the main thread, labeled with the issue.
func TestRunStartPlan_RendersAtPlanLens(t *testing.T) {
	var b strings.Builder
	runStartPlan(&b, 75)
	out := b.String()
	for _, want := range []string{"#75", "ARCH-DRY", "at-plan", "change-code"} {
		if !strings.Contains(out, want) {
			t.Errorf("start-plan output missing %q:\n%s", want, out)
		}
	}
	// No --issue → generic label, still renders the principles.
	var b2 strings.Builder
	runStartPlan(&b2, 0)
	if !strings.Contains(b2.String(), "ARCH-PURE") {
		t.Error("start-plan with no issue should still render the principles")
	}
}

// ── #82 M3: base-contention summary (pure), isBaseRepo, issueRef ─────────────

func TestBaseContentionSummary(t *testing.T) {
	cases := []struct {
		name string
		in   baseContention
		want []string // substrings the line must contain
	}{
		{
			name: "clean main no others → clear to plan",
			in:   baseContention{Repo: "ariadne", Branch: "main"},
			want: []string{"base (ariadne)", "clean main", "clear to plan"},
		},
		{
			name: "branched base",
			in:   baseContention{Repo: "ariadne", Branch: "000082-foo"},
			want: []string{"on branch `000082-foo`", "moving base"},
		},
		{
			name: "dirty code on main",
			in:   baseContention{Repo: "ariadne", Branch: "main", DirtyCode: 3},
			want: []string{"3 uncommitted code file(s)", "moving base"},
		},
		{
			name: "concurrent issues excluded-self already done by gather",
			in:   baseContention{Repo: "ariadne", Branch: "main", Others: []inFlightIssue{{ID: "000081"}, {ID: "000076"}}},
			want: []string{"2 other issue(s) in-flight", "#81", "#76", "moving base"},
		},
		{
			name: "branched + dirty + others compose",
			in:   baseContention{Repo: "ariadne", Branch: "wip", DirtyCode: 1, Others: []inFlightIssue{{ID: "000081"}}},
			want: []string{"on branch `wip`", "1 uncommitted code file(s)", "1 other issue(s) in-flight (#81)"},
		},
		{
			name: "detached head",
			in:   baseContention{Repo: "ariadne", Branch: ""},
			want: []string{"detached HEAD", "moving base"},
		},
	}
	for _, tc := range cases {
		got := baseContentionSummary(tc.in)
		for _, w := range tc.want {
			if !strings.Contains(got, w) {
				t.Errorf("%s: summary %q missing %q", tc.name, got, w)
			}
		}
		// Clean() must agree with the wording.
		if tc.in.Clean() != strings.Contains(got, "clear to plan") {
			t.Errorf("%s: Clean()=%v but wording %q disagrees", tc.name, tc.in.Clean(), got)
		}
	}
}

func TestIssueRef(t *testing.T) {
	for _, tc := range []struct{ id, want string }{
		{"000081", "#81"},
		{"000076", "#76"},
		{"42", "#42"},
		{"weird", "#weird"},
	} {
		if got := issueRef(tc.id); got != tc.want {
			t.Errorf("issueRef(%q)=%q, want %q", tc.id, got, tc.want)
		}
	}
}

// isBaseRepo: a real construct/ dir → base; a symlinked construct/ → derivative.
func TestIsBaseRepo(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "construct"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !isBaseRepo(base) {
		t.Error("real construct/ dir should read as the base repo")
	}

	deriv := t.TempDir()
	realConstruct := filepath.Join(base, "construct")
	if err := os.Symlink(realConstruct, filepath.Join(deriv, "construct")); err != nil {
		t.Fatal(err)
	}
	if isBaseRepo(deriv) {
		t.Error("symlinked construct/ should NOT read as the base repo (it's a derivative)")
	}

	none := t.TempDir()
	if isBaseRepo(none) {
		t.Error("no construct/ at all should not read as the base repo")
	}
}
