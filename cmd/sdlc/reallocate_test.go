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
	if !strings.Contains(err.Error(), "no `id:` line in a leading frontmatter block") {
		t.Errorf("refusal must say why: %v", err)
	}
}

func TestRewriteIdentity_RefusesNonConventionalName(t *testing.T) {
	if _, _, err := rewriteIdentity("workshop/issues/notes.md", []byte("id: 000207\n"), 208); err == nil {
		t.Error("a name outside the NNNNNN-slug.md convention must refuse")
	}
}

// finish removes the old path AND every orphan an earlier attempt wrote,
// keeping only what was published. Retries can write several candidates; NextID
// scans local files, so a stray would silently reserve an id nothing holds.
func TestPublishResult_FinishRemovesOldAndOrphans(t *testing.T) {
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

	res := &publishResult{
		reallocs:   []*reallocation{{OldPath: old, NewPath: final}},
		candidates: []string{orphan, final},
	}
	if err := res.finish(); err != nil {
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
func TestPublishResult_FinishIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	final := filepath.Join(dir, "000209-slug.md")
	if err := os.WriteFile(final, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := &publishResult{
		reallocs:   []*reallocation{{OldPath: filepath.Join(dir, "gone.md"), NewPath: final}},
		candidates: []string{final},
	}
	if err := res.finish(); err != nil {
		t.Fatalf("first: %v", err)
	}
	if err := res.finish(); err != nil {
		t.Fatalf("second call must be a no-op, got %v", err)
	}
}

// BR-14: only the FRONTMATTER id moves. A body that also mentions `id: NNNNNN`
// in prose keeps it — that line is content, not identity.
func TestRewriteIdentity_LeavesAProseIDLineAlone(t *testing.T) {
	content := []byte("---\nid: 000207\nstatus: open\n---\n\n# Title\n\nThe collision was on\nid: 000999\nin pair.\n")
	_, out, err := rewriteIdentity("workshop/issues/000207-x.md", content, 208)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "id: 000208\n") {
		t.Errorf("frontmatter not rewritten:\n%s", out)
	}
	if !strings.Contains(string(out), "id: 000999\n") {
		t.Errorf("a prose id line was rewritten:\n%s", out)
	}
}

// Frontmatter without an `id:` line, body WITH one: the identity is absent, so
// re-allocation must refuse rather than rewrite a body line.
//
// The prose test above cannot show this — the frontmatter id is the first match
// either way, so it passes with the span check removed. Only a file whose ONLY
// `id:` line is in the body distinguishes "search the frontmatter" from "search
// the document" (#207 BR-14).
func TestRewriteIdentity_RefusesWhenOnlyTheBodyHasAnIDLine(t *testing.T) {
	content := []byte("---\nstatus: open\n---\n\n# Title\n\nid: 000999\n")
	_, out, err := rewriteIdentity("workshop/issues/000207-x.md", content, 208)
	if err == nil {
		t.Fatalf("rewrote a body id line as if it were the identity:\n%s", out)
	}
	if !strings.Contains(err.Error(), "no `id:` line") {
		t.Errorf("refusal should name the missing frontmatter id, got: %v", err)
	}
}
