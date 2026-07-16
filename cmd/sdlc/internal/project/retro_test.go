package project

import "testing"

func TestLatestRetroDate(t *testing.T) {
	d, err := ParseDoc("---\ntype: project\n---\n## Log\n### 2026-07-01 — retro\nold\n### 2026-07-12 — retro\nnew\n")
	if err != nil {
		t.Fatal(err)
	}
	if got := LatestRetroDate(d); got != "2026-07-12" {
		t.Fatalf("LatestRetroDate = %q", got)
	}
}
