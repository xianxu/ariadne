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

// parseSubstrateTargets mirrors lib-deps.sh deps_substrate_targets' parse: keep
// substrate rows' target, strip `#` comments, drop blanks/malformed/data rows.
func TestParseSubstrateTargets(t *testing.T) {
	cases := []struct {
		name, content string
		want          []string
	}{
		{"single", "substrate ../ariadne\n", []string{"../ariadne"}},
		{"data row ignored", "substrate ../ariadne\ndata git@x:y.git data/z\n", []string{"../ariadne"}},
		{"comment stripped mid-line", "substrate ../ariadne # the base\n", []string{"../ariadne"}},
		{"full-line comment + blank", "# header\n\nsubstrate ../up\n", []string{"../up"}},
		{"malformed one-field skipped", "substrate\nsubstrate ../up\n", []string{"../up"}},
		{"absolute target kept", "substrate /abs/ariadne\n", []string{"/abs/ariadne"}},
		{"multiple substrate rows", "substrate ../a\nsubstrate ../b\n", []string{"../a", "../b"}},
		{"empty", "", nil},
	}
	for _, tc := range cases {
		got := parseSubstrateTargets(tc.content)
		if len(got) != len(tc.want) {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("%s: got[%d]=%q, want %q", tc.name, i, got[i], tc.want[i])
			}
		}
	}
}

// substrateChain walks construct/deps transitively, resolving each target
// against its DECLARING root (the exact bug #82 M3 had), deduping, present-skip.
func TestSubstrateChain(t *testing.T) {
	// Build a 3-repo chain C → B → A as sibling dirs under one parent, each
	// declaring its upstream by a repo-root-relative path (`substrate ../X`).
	parent := t.TempDir()
	mk := func(name, deps string) string {
		root := filepath.Join(parent, name)
		if err := os.MkdirAll(filepath.Join(root, "construct"), 0o755); err != nil {
			t.Fatal(err)
		}
		if deps != "" {
			if err := os.WriteFile(filepath.Join(root, "construct", "deps"), []byte(deps), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return root
	}
	rootA := mk("A", "")                    // root: no upstream
	mk("B", "substrate ../A\n")             // B → A
	rootC := mk("C", "substrate ../B\n")    // C → B (→ A transitively)

	// From C: the full transitive path is [B, A], resolved root-relative.
	got := substrateChain(rootC)
	want := []string{filepath.Join(parent, "B"), filepath.Join(parent, "A")}
	if len(got) != len(want) {
		t.Fatalf("chain(C) = %v, want %v", got, want)
	}
	for i := range want {
		// EvalSymlinks may canonicalize /var → /private/var on macOS; compare resolved.
		wr, _ := filepath.EvalSymlinks(want[i])
		gr, _ := filepath.EvalSymlinks(got[i])
		if gr != wr {
			t.Errorf("chain(C)[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	// From the root A: empty chain (no upstream).
	if c := substrateChain(rootA); len(c) != 0 {
		t.Errorf("chain(A) = %v, want empty (root has no upstream)", c)
	}

	// Absent peer is skipped, not fatal.
	lonely := mk("Lonely", "substrate ../DoesNotExist\n")
	if c := substrateChain(lonely); len(c) != 0 {
		t.Errorf("chain(Lonely) = %v, want empty (absent peer skipped)", c)
	}
}
