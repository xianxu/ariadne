package gitx

import (
	"path/filepath"
	"testing"
)

// TestEscapes_IsComponentWise is BR-24: a prefix test on ".." rejects a
// directory whose name merely starts with two dots, which is a legal name and
// not an escape.
func TestEscapes_IsComponentWise(t *testing.T) {
	for _, tc := range []struct {
		rel  string
		want bool
	}{
		{"..", true},
		{filepath.Join("..", "elsewhere"), true},
		{filepath.Join("..", "..", "etc"), true},
		{"", true},
		{"workshop/issues", false},
		{"..config", false},               // a legal directory name, not an escape
		{"..git-keep/issues", false},      // ditto, nested
		{filepath.Join("a", ".."), false}, // filepath.Rel never returns this, but it resolves inside
	} {
		if got := Escapes(tc.rel); got != tc.want {
			t.Errorf("Escapes(%q) = %v, want %v", tc.rel, got, tc.want)
		}
	}
}

func TestInsideRoot_RelativeJoinsOnRootNotCwd(t *testing.T) {
	root := t.TempDir()
	got, err := InsideRoot(root, "workshop/issues")
	if err != nil {
		t.Fatal(err)
	}
	if got != "workshop/issues" {
		t.Errorf("InsideRoot = %q, want workshop/issues", got)
	}
	if _, err := InsideRoot(root, filepath.Join(root, "..", "elsewhere")); err == nil {
		t.Error("a path outside the root must be refused")
	}
}
