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

// TestRenderSkill_FaithfulToCommittedSKILL is the FAITHFULNESS GATE, now over
// the merged-render path: rendering with the REAL ariadne repo root (single-layer
// DAG — ariadne is the base layer) must reproduce the current committed
// construct/local/datatype/SKILL.md EXACTLY except the one description line. The
// single-layer merge == #111's exact 13-noun byte output. We read the committed
// file, swap ITS description tail to the live noun list, and assert byte-equality
// with renderSkill's output — proving the template is a verbatim copy of the prose
// and only the description differs.
func TestRenderSkill_FaithfulToCommittedSKILL(t *testing.T) {
	// Repo root relative to cmd/datatype (the test's cwd).
	const repoRoot = "../.."
	const committedSkill = "../../construct/local/datatype/SKILL.md"

	names := mergedNames(t, repoRoot)
	rendered := renderSkill(names)

	committedBytes, err := os.ReadFile(committedSkill)
	if err != nil {
		t.Fatalf("read committed SKILL.md: %v", err)
	}
	committed := string(committedBytes)

	// Independently compute what the committed file's description SHOULD become:
	// the committed file currently ends `…known frontmatter type:"`; appending the
	// live nouns before the closing quote is the ONLY allowed change.
	const tail = "known frontmatter type:"
	expected := strings.Replace(committed, tail+"\"", tail+" "+strings.Join(names, ", ")+"\"", 1)

	if rendered != expected {
		// Surface the first differing line for a useful failure.
		rl := strings.Split(rendered, "\n")
		el := strings.Split(expected, "\n")
		for i := 0; i < len(rl) && i < len(el); i++ {
			if rl[i] != el[i] {
				t.Fatalf("rendered diverges from committed SKILL.md beyond the description line:\n  line %d rendered: %q\n  line %d expected: %q", i+1, rl[i], i+1, el[i])
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
