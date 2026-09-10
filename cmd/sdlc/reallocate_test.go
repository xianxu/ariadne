package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Filename and frontmatter move TOGETHER. Leaving either behind produces an
// artifact whose name and body disagree — and they are read by different
// consumers (`sdlc claim` matches the filename, `vocabulary validate-instance`
// reads the frontmatter), so a half-rewrite is silent rather than loud.
func TestRewriteIdentity_MovesNameAndFrontmatterTogether(t *testing.T) {
	content := []byte("---\nid: 000207\nstatus: open\ndeps: []\n---\n\n# Title\n\nid: not-frontmatter\n")
	newPath, out, err := rewriteIdentity("workshop/issues/000207-my-slug.md", content, 208)
	if err != nil {
		t.Fatal(err)
	}
	if newPath != "workshop/issues/000208-my-slug.md" {
		t.Errorf("path = %q, want the slug preserved and only the id moved", newPath)
	}
	if !strings.Contains(string(out), "id: 000208\n") {
		t.Errorf("frontmatter not rewritten:\n%s", out)
	}
	if strings.Contains(string(out), "id: 000207") {
		t.Errorf("old id survives:\n%s", out)
	}
	// A later `id:`-looking line in prose must NOT be touched — the frontmatter
	// block is the first thing in the file and only its line is the identity.
	if !strings.Contains(string(out), "id: not-frontmatter") {
		t.Errorf("prose was rewritten:\n%s", out)
	}
	if !strings.Contains(string(out), "status: open") {
		t.Errorf("the rest of the frontmatter was disturbed:\n%s", out)
	}
}

// A file with no `id:` line is the grafted-fragment shape measured 2026-09-09.
// Renaming it alone is a half-rewrite, so refuse rather than produce a
// disagreeing artifact.
func TestRewriteIdentity_RefusesFileWithNoIDLine(t *testing.T) {
	_, _, err := rewriteIdentity("workshop/issues/000207-fragment.md", []byte("\n## Log\n\ntext\n"), 208)
	if err == nil {
		t.Fatal("a file with no id: frontmatter must refuse, not be renamed alone")
	}
	if !strings.Contains(err.Error(), "no `id:` frontmatter") {
		t.Errorf("refusal must say why: %v", err)
	}
}

func TestRewriteIdentity_RefusesNonConventionalName(t *testing.T) {
	if _, _, err := rewriteIdentity("workshop/issues/notes.md", []byte("id: 000207\n"), 208); err == nil {
		t.Error("a name outside the NNNNNN-slug.md convention must refuse")
	}
}

// finish removes the old path AND every orphan an earlier attempt wrote, keeping
// only what was actually published. Retries can write several candidates; all but
// the published one are orphans, and NextID scans local files, so a stray would
// silently reserve an id nothing holds.
func TestReallocation_FinishRemovesOldAndOrphans(t *testing.T) {
	dir := t.TempDir()
	mk := func(name string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	old := mk("000207-slug.md")
	orphan := mk("000208-slug.md") // written by a rejected attempt
	final := mk("000209-slug.md")  // the one that landed

	rc := &reallocation{OldPath: old, NewPath: final, written: []string{orphan, final}}
	if err := rc.finish(); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{old, orphan} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s survived — a stray local file reserves an id nothing published", filepath.Base(p))
		}
	}
	if _, err := os.Stat(final); err != nil {
		t.Errorf("the published file was removed: %v", err)
	}
}

// finish is idempotent: a second call after a partial crash must not fail.
func TestReallocation_FinishIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	final := filepath.Join(dir, "000209-slug.md")
	if err := os.WriteFile(final, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rc := &reallocation{OldPath: filepath.Join(dir, "gone.md"), NewPath: final, written: []string{final}}
	if err := rc.finish(); err != nil {
		t.Fatalf("first: %v", err)
	}
	if err := rc.finish(); err != nil {
		t.Fatalf("second call must be a no-op, got %v", err)
	}
}

// The `id:` match must be ANCHORED to a frontmatter line, not to any occurrence.
//
// The dangerous shape is real and this repo produced one on 2026-09-06: a grafted
// fragment with NO frontmatter that nonetheless mentions an id in prose. Anchored,
// rewriteIdentity refuses (there is no identity line to move). Unanchored, it
// would silently rewrite the PROSE and rename the file — producing an artifact
// whose body now misreports history and whose identity was never actually moved.
//
// Found by mutation: the first version of this test used a non-numeric prose
// value, so an over-broad regex produced the same answer and the test passed
// without pinning the anchoring at all.
func TestRewriteIdentity_AnchorsToFrontmatterNotProse(t *testing.T) {
	fragment := []byte("\n## Log\n\n### 2026-09-06 — grafted evidence\n\nThe collision was on id: 000207 in pair.\n")
	_, _, err := rewriteIdentity("workshop/issues/000207-fragment.md", fragment, 208)
	if err == nil {
		t.Fatal("a file whose only `id:` is in prose has no identity line — must refuse, not rewrite the prose")
	}
	if !strings.Contains(err.Error(), "no `id:` frontmatter") {
		t.Errorf("refusal must name the cause: %v", err)
	}
}
