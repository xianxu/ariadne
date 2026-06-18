package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/layergraph"
)

// TestRenderSkill_DeterministicAndContainsNames: renderSkill injects the joined
// list into the description and is byte-identical on re-run (the byte-stable
// guard depends on this determinism).
func TestRenderSkill_DeterministicAndContainsNames(t *testing.T) {
	names := []string{"alpha", "beta", "gamma"}
	first := renderSkill(names)
	second := renderSkill(names)
	if first != second {
		t.Fatal("renderSkill is non-deterministic (re-run differs)")
	}
	want := "known frontmatter type: alpha, beta, gamma\""
	if !strings.Contains(first, want) {
		t.Fatalf("renderSkill output missing the joined list in the description; want substring %q", want)
	}
	// The placeholder token must be fully consumed.
	if strings.Contains(first, datatypeNamesPlaceholder) {
		t.Fatalf("renderSkill left the placeholder token %q in the output", datatypeNamesPlaceholder)
	}
}

// mergedNames is a test helper: the sorted noun list mergeTypes resolves for a
// single repo root with no project-local datatype/ dir — the merged-render path.
func mergedNames(t *testing.T, root string) []string {
	t.Helper()
	protos, err := mergeTypes(layergraph.OSFS{}, []string{root}, filepath.Join(root, "datatype"))
	if err != nil {
		t.Fatalf("mergeTypes(%s): %v", root, err)
	}
	names := make([]string, len(protos))
	for i, p := range protos {
		names[i] = p.Name
	}
	return names
}

// TestRenderSkill_FaithfulToTemplate is the FAITHFULNESS GATE (#115 M3 re-anchored):
// the committed construct/local/datatype/SKILL.md was retired — its body now lives,
// gitignored, under construct/generated/ (regenerated every compile). So the
// faithfulness reference is the AUTHORED template SKILL.md.tmpl itself (the embedded
// skillTemplate, the single source of truth). Rendering with the REAL ariadne repo
// root (single-layer DAG — ariadne is the base layer) must reproduce the template
// EXACTLY except the one placeholder line: the ONLY allowed change is swapping
// __DATATYPE_NAMES__ for the live noun list. This proves renderSkill is a verbatim
// copy of the prose and only the description differs (the #111 13-noun byte output).
func TestRenderSkill_FaithfulToTemplate(t *testing.T) {
	// Repo root relative to cmd/datatype (the test's cwd).
	const repoRoot = "../.."
	const templatePath = "SKILL.md.tmpl"

	names := mergedNames(t, repoRoot)
	rendered := renderSkill(names)

	tmplBytes, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("read template SKILL.md.tmpl: %v", err)
	}
	// Close the vacuous-pass hole (M3-review #3): the byte-faithful check below
	// computes `expected` with the SAME Replace over the SAME template, so if the
	// placeholder were ever removed from SKILL.md.tmpl, both sides would be the
	// unchanged template and this test would stay GREEN while shipping a skill with
	// NO datatype nouns in its trigger description. Assert the substitution is real.
	if !strings.Contains(string(tmplBytes), datatypeNamesPlaceholder) {
		t.Fatalf("SKILL.md.tmpl no longer contains the %q placeholder — the noun list would never be injected into the eager trigger description", datatypeNamesPlaceholder)
	}
	if strings.Contains(rendered, datatypeNamesPlaceholder) {
		t.Fatalf("rendered output still contains the unsubstituted %q placeholder", datatypeNamesPlaceholder)
	}
	if joined := strings.Join(names, ", "); !strings.Contains(rendered, joined) {
		t.Fatalf("rendered output is missing the injected noun list %q", joined)
	}
	// The ONLY allowed change vs the template is the placeholder→noun-list swap.
	expected := strings.Replace(string(tmplBytes), datatypeNamesPlaceholder, strings.Join(names, ", "), 1)

	if rendered != expected {
		rl := strings.Split(rendered, "\n")
		el := strings.Split(expected, "\n")
		for i := 0; i < len(rl) && i < len(el); i++ {
			if rl[i] != el[i] {
				t.Fatalf("rendered diverges from SKILL.md.tmpl beyond the placeholder line:\n  line %d rendered: %q\n  line %d expected: %q", i+1, rl[i], i+1, el[i])
			}
		}
		t.Fatalf("rendered length %d != expected length %d (trailing divergence)", len(rendered), len(expected))
	}
}

// TestFindRepoRoot_FromNestedSubdir: findRepoRoot walks up from a nested
// subdirectory to the nearest ancestor that contains a construct/ dir — so
// apply-time access (datatype list/show) anchors the repo root, not the agent's
// cwd, and the marker's cwd=construct/local/datatype resolves the repo root.
func TestFindRepoRoot_FromNestedSubdir(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, "construct", "local", "datatype"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "cmd", "datatype"), 0o755); err != nil {
		t.Fatal(err)
	}

	// From the marker's cwd (construct/local/datatype), the nearest ancestor with
	// a construct/ child is the repo root itself.
	got, err := findRepoRoot(filepath.Join(repo, "construct", "local", "datatype"))
	if err != nil {
		t.Fatalf("findRepoRoot from marker cwd: %v", err)
	}
	if canonPath(got) != canonPath(repo) {
		t.Fatalf("findRepoRoot = %q, want repo root %q", got, repo)
	}

	// From a deep cmd/ subdir, same answer.
	got, err = findRepoRoot(filepath.Join(repo, "cmd", "datatype"))
	if err != nil {
		t.Fatalf("findRepoRoot from cmd subdir: %v", err)
	}
	if canonPath(got) != canonPath(repo) {
		t.Fatalf("findRepoRoot(cmd/datatype) = %q, want repo root %q", got, repo)
	}
}

// TestFindRepoRoot_GitFallback: when no ancestor has construct/, findRepoRoot
// falls back to the nearest ancestor containing .git.
func TestFindRepoRoot_GitFallback(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(repo, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := findRepoRoot(sub)
	if err != nil {
		t.Fatalf("findRepoRoot git fallback: %v", err)
	}
	if canonPath(got) != canonPath(repo) {
		t.Fatalf("findRepoRoot = %q, want git root %q", got, repo)
	}
}

// canonPath canonicalizes for comparison (macOS /tmp → /private/tmp).
func canonPath(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return p
}
