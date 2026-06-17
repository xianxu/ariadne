package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTypeNames_SortedFilenamesNoMD: typeNames returns each prototype's filename
// without `.md`, sorted ascending — ignoring subdirs and the `type:` frontmatter.
func TestTypeNames_SortedFilenamesNoMD(t *testing.T) {
	dir := t.TempDir()
	// Out-of-order on disk; a subdir + a non-.md file must be ignored.
	for _, f := range []string{"c.md", "a.md", "b.md", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("---\ntype: type\n---\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := typeNames(dir)
	if err != nil {
		t.Fatalf("typeNames: %v", err)
	}
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("typeNames = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("typeNames = %v, want %v (sorted, filename sans .md, no subdir/non-md)", got, want)
		}
	}
}

// TestRenderSkill_DeterministicAndContainsNames: renderSkill injects the joined
// list into the description and is byte-identical on re-run (the drift guard
// depends on this determinism).
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

// TestRenderSkill_FaithfulToCommittedSKILL is the FAITHFULNESS GATE: rendering
// with the REAL construct/datatype must reproduce the current committed
// construct/local/datatype/SKILL.md EXACTLY except the one description line. We
// read the committed file, swap ITS description tail to the live noun list, and
// assert byte-equality with renderSkill's output — proving the template is a
// verbatim copy of the prose and only the description differs.
func TestRenderSkill_FaithfulToCommittedSKILL(t *testing.T) {
	// Paths relative to cmd/datatype (the test's cwd) → the repo root.
	const datatypeDir = "../../construct/datatype"
	const committedSkill = "../../construct/local/datatype/SKILL.md"

	names, err := typeNames(datatypeDir)
	if err != nil {
		t.Fatalf("typeNames(real construct/datatype): %v", err)
	}
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

// TestWriteSkill_WritesOutputDir: writeSkill writes <output>/SKILL.md byte-equal
// to renderSkill over the same dir (the thin IO shell adds nothing).
func TestWriteSkill_WritesOutputDir(t *testing.T) {
	const datatypeDir = "../../construct/datatype"
	out := t.TempDir()
	if err := writeSkill(datatypeDir, out); err != nil {
		t.Fatalf("writeSkill: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(out, "SKILL.md"))
	if err != nil {
		t.Fatalf("read written SKILL.md: %v", err)
	}
	names, err := typeNames(datatypeDir)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != renderSkill(names) {
		t.Fatal("writeSkill output differs from renderSkill (IO shell altered content)")
	}
}
