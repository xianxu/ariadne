package plan

import "testing"

// composeProse concatenates prose fragments foundation-first (slice order) into
// the AGENTS.md body. Pure. This is the structural fix for the @AGENTS.local.md
// bug: the repo's own fragment is concatenated in directly, not @-imported.

func TestComposeProseFoundationFirst(t *testing.T) {
	// base + layer + local, in that (foundation-first) order. Assert the local
	// fragment is present (the @-import bug, fixed structurally) AND that order
	// is preserved.
	got := composeProse([]string{"BASE", "LAYER", "LOCAL"})
	want := "BASE\n\nLAYER\n\nLOCAL\n"
	if got != want {
		t.Fatalf("composeProse = %q, want %q", got, want)
	}
}

func TestComposeProseEmpty(t *testing.T) {
	if got := composeProse(nil); got != "" {
		t.Fatalf("composeProse(nil) = %q, want empty", got)
	}
}

func TestComposeProseSingle(t *testing.T) {
	if got := composeProse([]string{"ONLY"}); got != "ONLY\n" {
		t.Fatalf("composeProse single = %q, want %q", got, "ONLY\n")
	}
}
