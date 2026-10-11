package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// #320: the union-merge attributes land in a delimited block, tested against a real
// t.TempDir-rooted OSFS like the .gitignore block.

func applyAttrs(t *testing.T, root string) string {
	t.Helper()
	if err := Apply(weavefs.OSFS{}, root, []Action{EnsureGitattributes{Entries: UnionMergeAttributes}}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, ".gitattributes"))
	if err != nil {
		t.Fatalf("read .gitattributes: %v", err)
	}
	return string(got)
}

func TestApplyEnsureGitattributesCreatesBlock(t *testing.T) {
	got := applyAttrs(t, t.TempDir())
	want := ignoreBegin + "\nworkshop/lessons.md merge=union\n" + ignoreEnd + "\n"
	if got != want {
		t.Fatalf(".gitattributes = %q, want %q", got, want)
	}
}

func TestApplyEnsureGitattributesIdempotentAndKeepsAuthored(t *testing.T) {
	root := t.TempDir()
	authored := "*.png binary\nworkshop/lessons.md -merge\n"
	if err := os.WriteFile(filepath.Join(root, ".gitattributes"), []byte(authored), 0o644); err != nil {
		t.Fatal(err)
	}
	first := applyAttrs(t, root)
	if second := applyAttrs(t, root); first != second {
		t.Fatalf("re-weave changed .gitattributes:\n1st: %q\n2nd: %q", first, second)
	}
	// The block comes first, so an authored override after it still wins.
	if !strings.HasPrefix(first, ignoreBegin+"\n") || !strings.HasSuffix(first, ignoreEnd+"\n"+authored) {
		t.Fatalf("authored lines must follow the block untouched:\n%s", first)
	}
}

func TestApplyEnsureGitattributesRefusesNonRegular(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".gitattributes"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := Apply(weavefs.OSFS{}, root, []Action{EnsureGitattributes{Entries: UnionMergeAttributes}})
	if err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("Apply = %v, want a not-a-regular-file refusal", err)
	}
}

// Every entry is a union merge of a workshop list: the file set is deliberate.
func TestUnionMergeAttributesAreWorkshopUnionEntries(t *testing.T) {
	for _, e := range UnionMergeAttributes {
		if !strings.HasPrefix(e, "workshop/") || !strings.HasSuffix(e, " merge=union") {
			t.Errorf("entry %q: want a workshop/ path with merge=union", e)
		}
	}
}
