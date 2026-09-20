package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// gitignore.go is weave's generated-runtime ignore mechanism. The pure transform
// (mergeManagedBlock) is unit-tested directly; the IO seam (applyEnsureGitignore,
// via Apply) is tested against a real t.TempDir-rooted OSFS (ARCH: faithful over
// mocked).
//
// The tests below were written against the append-only ensureGitignoreText and
// TRANSLATED to the managed block (#239 M2) rather than deleted: each still
// asserts something true of the new mechanism (creates when absent, idempotent
// when current, preserves the repo's own lines, no trailing-newline glue). Only
// the layout moved — weave's entries now live between markers.

// blockOf is the expected rendering of a managed block carrying entries, with no
// repo-owned lines outside it.
func blockOf(entries ...string) string {
	out := managedBlockOpen + "\n"
	for _, e := range entries {
		out += e + "\n"
	}
	return out + managedBlockClose + "\n"
}

func TestEnsureGitignoreTextAppendsToEmpty(t *testing.T) {
	got, changed, err := mergeManagedBlock("", []string{"/AGENTS.md", "/GEMINI.md"})
	if err != nil || !changed {
		t.Fatalf("err=%v changed=%v, want nil/true on an empty .gitignore", err, changed)
	}
	want := blockOf("/AGENTS.md", "/GEMINI.md")
	if got != want {
		t.Fatalf("mergeManagedBlock = %q, want %q", got, want)
	}
}

func TestEnsureGitignoreTextPreservesExistingAndAppendsAbsent(t *testing.T) {
	// Existing entries + a comment are preserved verbatim; only the truly absent
	// entry is appended (the present one is NOT duplicated — grep -qxF semantics).
	// The repo's comment and its own entry survive verbatim; the loose
	// /AGENTS.md is ABSORBED into the block rather than duplicated.
	current := "# existing comment\n/AGENTS.md\nbin/\n"
	got, changed, err := mergeManagedBlock(current, []string{"/AGENTS.md", "/GEMINI.md"})
	if err != nil || !changed {
		t.Fatalf("err=%v changed=%v, want nil/true", err, changed)
	}
	want := "# existing comment\nbin/\n" + blockOf("/AGENTS.md", "/GEMINI.md")
	if got != want {
		t.Fatalf("mergeManagedBlock = %q, want %q", got, want)
	}
	if strings.Count(got, "/AGENTS.md") != 1 {
		t.Fatalf("/AGENTS.md duplicated:\n%s", got)
	}
}

func TestEnsureGitignoreTextIdempotentWhenAllPresent(t *testing.T) {
	// Every entry already present ⇒ no change, byte-identical (running weave twice
	// never duplicates lines). Built from the canonical list so adding an entry can
	// never silently desync this fixture.
	current := blockOf(GeneratedRuntimeGitignoreEntries...)
	got, changed, err := mergeManagedBlock(current, GeneratedRuntimeGitignoreEntries)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatalf("changed = true when the block is already current, want false; got:\n%s", got)
	}
	if got != current {
		t.Fatalf("content mutated when all present:\n got %q\nwant %q", got, current)
	}
}

// TestGeneratedRuntimeGitignoreCoversConstructGenerated locks the #115 M3
// addition: the per-repo dynamic-skill materialization tree construct/generated/ is
// in the owned ignore set (gitignored EVERYWHERE — it's regenerated every compile
// and must never be tracked; the body lived committed under construct/local before).
func TestGeneratedRuntimeGitignoreCoversConstructGenerated(t *testing.T) {
	found := false
	for _, e := range GeneratedRuntimeGitignoreEntries {
		if e == "/construct/generated/" {
			found = true
		}
	}
	if !found {
		t.Fatalf("GeneratedRuntimeGitignoreEntries missing /construct/generated/: %v", GeneratedRuntimeGitignoreEntries)
	}
}

func TestEnsureGitignoreTextAddsTrailingNewlineBeforeAppend(t *testing.T) {
	// A non-empty file NOT ending in a newline gets one before the appended entry,
	// so the new entry never glues onto the last existing line.
	got, changed, err := mergeManagedBlock("bin/", []string{"/AGENTS.md"})
	if err != nil || !changed {
		t.Fatalf("err=%v changed=%v, want nil/true", err, changed)
	}
	want := "bin/\n" + blockOf("/AGENTS.md")
	if got != want {
		t.Fatalf("mergeManagedBlock = %q, want %q", got, want)
	}
}

func TestApplyEnsureGitignoreCreatesAndAppends(t *testing.T) {
	// Apply on a repo with no .gitignore creates it carrying the fixed entries.
	root := t.TempDir()
	if err := Apply(weavefs.OSFS{}, root, []Action{
		EnsureGitignore{Entries: GeneratedRuntimeGitignoreEntries},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	for _, entry := range GeneratedRuntimeGitignoreEntries {
		if !strings.Contains(string(got), entry+"\n") {
			t.Fatalf(".gitignore missing %q:\n%s", entry, got)
		}
	}
}

func TestApplyEnsureGitignoreIdempotent(t *testing.T) {
	// Two applies in a row leave the .gitignore byte-identical (no churn / no dup).
	root := t.TempDir()
	gi := filepath.Join(root, ".gitignore")
	act := []Action{EnsureGitignore{Entries: GeneratedRuntimeGitignoreEntries}}

	if err := Apply(weavefs.OSFS{}, root, act); err != nil {
		t.Fatalf("Apply (1st): %v", err)
	}
	first, err := os.ReadFile(gi)
	if err != nil {
		t.Fatalf("read after 1st: %v", err)
	}
	if err := Apply(weavefs.OSFS{}, root, act); err != nil {
		t.Fatalf("Apply (2nd): %v", err)
	}
	second, err := os.ReadFile(gi)
	if err != nil {
		t.Fatalf("read after 2nd: %v", err)
	}
	if string(first) != string(second) {
		t.Fatalf("re-weave changed .gitignore:\n1st: %q\n2nd: %q", first, second)
	}
}

func TestApplyEnsureGitignorePreservesExisting(t *testing.T) {
	// Apply preserves pre-existing entries/comments and appends only the absent.
	root := t.TempDir()
	gi := filepath.Join(root, ".gitignore")
	if err := os.WriteFile(gi, []byte("# hand notes\nbin/\n/AGENTS.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Apply(weavefs.OSFS{}, root, []Action{
		EnsureGitignore{Entries: GeneratedRuntimeGitignoreEntries},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got, err := os.ReadFile(gi)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	for _, keep := range []string{"# hand notes", "bin/"} {
		if !strings.Contains(string(got), keep) {
			t.Fatalf(".gitignore dropped pre-existing %q:\n%s", keep, got)
		}
	}
	if strings.Count(string(got), "/AGENTS.md") != 1 {
		t.Fatalf("/AGENTS.md duplicated:\n%s", got)
	}
}

// --- the managed block (#239 M2) -------------------------------------------
//
// Append-only could never RETIRE an entry: a manifest row removed upstream left
// its ignore line in every derivative forever. Harmless at 9 hardcoded entries,
// actively dangerous once the list is manifest-derived (M3), because a stale
// line can silently untrack a repo-owned file that later takes that path.

func TestManagedBlockAppendsWhenAbsent(t *testing.T) {
	got, changed, err := mergeManagedBlock("mine/\n", []string{"/CLAUDE.md"})
	if err != nil || !changed {
		t.Fatalf("err=%v changed=%v", err, changed)
	}
	want := "mine/\n" + managedBlockOpen + "\n/CLAUDE.md\n" + managedBlockClose + "\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestManagedBlockReplacesWholesaleSoRetiredEntriesDisappear(t *testing.T) {
	current := "mine/\n" + managedBlockOpen + "\n/CLAUDE.md\n/RETIRED.md\n" + managedBlockClose + "\ntail/\n"
	got, changed, err := mergeManagedBlock(current, []string{"/CLAUDE.md"})
	if err != nil || !changed {
		t.Fatalf("err=%v changed=%v", err, changed)
	}
	if strings.Contains(got, "/RETIRED.md") {
		t.Fatalf("retired entry survived: %q", got)
	}
	if !strings.Contains(got, "mine/") || !strings.Contains(got, "tail/") {
		t.Fatalf("repo-owned entries lost: %q", got)
	}
}

// pair's bin/* + !bin/*.sh must round-trip verbatim — the #64 regression where a
// blanket ignore made tracked shell scripts look disposable and a sweep rm'd them.
func TestManagedBlockRoundTripsRepoOwnedNegations(t *testing.T) {
	repo := "bin/*\n!bin/*.sh\n!bin/pair-dev\ncache/\n"
	first, _, err := mergeManagedBlock(repo, []string{"/CLAUDE.md"})
	if err != nil {
		t.Fatal(err)
	}
	second, changed, err := mergeManagedBlock(first, []string{"/CLAUDE.md"})
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("second weave rewrote an already-current .gitignore")
	}
	if second != first {
		t.Fatalf("not idempotent:\n%q\n%q", first, second)
	}
	if !strings.HasPrefix(second, repo) {
		t.Fatalf("repo entries moved or changed: %q", second)
	}
}

func TestManagedBlockAbsorbsLooseDuplicates(t *testing.T) {
	got, _, err := mergeManagedBlock("/CLAUDE.md\nmine/\n", []string{"/CLAUDE.md"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(got, "/CLAUDE.md") != 1 {
		t.Fatalf("duplicate entry: %q", got)
	}
	if !strings.Contains(got, "mine/") {
		t.Fatalf("repo entry lost: %q", got)
	}
}

// A derivative never runs the M2 binary — it jumps pre-M2 to post-M3, where the
// retired BLANKET entries match no per-path derived entry. Without an explicit
// absorb they would sit outside the block FOREVER as permanent directory
// ignores: the pair#64 hazard this issue exists to remove, left standing.
func TestManagedBlockAbsorbsRetiredBlanketEntries(t *testing.T) {
	current := "mine/\n/.claude/skills/\n/.agents/skills/\n/.colima/\n"
	got, _, err := mergeManagedBlock(current, []string{"/.claude/skills/xx-fix", "/.colima/Makefile"})
	if err != nil {
		t.Fatal(err)
	}
	for _, legacy := range []string{"/.claude/skills/\n", "/.agents/skills/\n", "/.colima/\n"} {
		if strings.Contains(got, legacy) {
			t.Fatalf("legacy blanket entry %q survived: %q", legacy, got)
		}
	}
	if !strings.Contains(got, "mine/") {
		t.Fatalf("repo entry lost: %q", got)
	}
}

// ARCH-SECURE: .gitignore is input weave did not produce — hand-edited, written
// by older versions, merged by git. A block it cannot parse is an error naming
// the remedy, never a guessed splice: a wrong guess deletes the repo's own
// rules. The "doubled" case is what a git MERGE CONFLICT produces.
func TestManagedBlockRefusesMalformedMarkers(t *testing.T) {
	for name, current := range map[string]string{
		"unterminated": managedBlockOpen + "\n/CLAUDE.md\n",
		"doubled":      managedBlockOpen + "\n" + managedBlockClose + "\n" + managedBlockOpen + "\n" + managedBlockClose + "\n",
		"close_first":  managedBlockClose + "\n" + managedBlockOpen + "\n",
	} {
		t.Run(name, func(t *testing.T) {
			_, _, err := mergeManagedBlock(current, []string{"/CLAUDE.md"})
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if !strings.Contains(err.Error(), "make weave") {
				t.Fatalf("error must name the remedy, got: %v", err)
			}
		})
	}
}

// An UNREADABLE .gitignore must abort, never be treated as empty: the block is
// written WHOLESALE, so "empty" would replace every repo-owned entry with
// weave's block alone. This test exists because the guard shipped without one —
// nothing in the package could fault a read (#239 M2 BR-19).
func TestApplyEnsureGitignoreFailsClosedOnReadError(t *testing.T) {
	root := t.TempDir()
	gitignore := filepath.Join(root, ".gitignore")
	mustWrite(t, gitignore, "bin/\n!bin/keep.sh\n")

	err := Apply(materializationFaultFS{operation: "read", destination: gitignore}, root,
		[]Action{EnsureGitignore{Entries: []string{"/CLAUDE.md"}}})
	if err == nil {
		t.Fatal("unreadable .gitignore was treated as empty — the repo's own entries would be destroyed")
	}
	if got := mustRead(t, gitignore); got != "bin/\n!bin/keep.sh\n" {
		t.Fatalf("file mutated despite the read failure: %q", got)
	}
}

// The entry list is de-duplicated where the block is emitted. The retired
// ensureGitignoreText guarded this; dropping the guard regressed it silently
// because the test named for the property had been rewritten to pass a single
// entry (#239 M2 BR-19). This one passes the duplicate it is named for.
func TestManagedBlockDedupsRepeatedInputEntry(t *testing.T) {
	got, _, err := mergeManagedBlock("", []string{"/AGENTS.md", "/GEMINI.md", "/AGENTS.md"})
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(got, "/AGENTS.md"); n != 1 {
		t.Fatalf("repeated input entry emitted %d times:\n%s", n, got)
	}
	// Order-stable: first occurrence wins, so the block stays deterministic.
	want := blockOf("/AGENTS.md", "/GEMINI.md")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// A .gitignore that quotes the open marker verbatim on its own line IS
// ambiguous — whole-line matching cannot tell a documenting comment from the
// real thing. The docstring used to claim it could; this pins the truth instead,
// and the truth is acceptable because the failure is loud and the message names
// the repair (#239 M2 BR-19).
func TestManagedBlockTreatsAQuotedMarkerAsAMarker(t *testing.T) {
	current := "# our convention is:\n" + managedBlockOpen + "\nmine/\n" +
		managedBlockOpen + "\n/CLAUDE.md\n" + managedBlockClose + "\n"
	_, _, err := mergeManagedBlock(current, []string{"/CLAUDE.md"})
	if err == nil {
		t.Fatal("a verbatim marker line must be read as a marker, not as prose")
	}
	if !strings.Contains(err.Error(), "make weave") {
		t.Fatalf("error must name the remedy, got: %v", err)
	}
}
