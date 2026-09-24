package main

import (
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
	"strings"
	"testing"
)

func TestArchiveDestination(t *testing.T) {
	for _, tc := range []struct {
		kind vocab.ArchiveKind
		want string
	}{{vocab.ArchiveIssues, "history/issues/000246-x.md"}, {vocab.ArchivePlans, "history/plans/000246-x.md"}} {
		if got := archiveDestination("history", tc.kind, "000246-x.md"); got != tc.want {
			t.Fatalf("%q want %q", got, tc.want)
		}
	}
}
func TestPlanArtifactBelongsToIssue(t *testing.T) {
	for _, tc := range []struct {
		name string
		want bool
	}{{"000246-x-plan.md", true}, {"000246-x-close-review.md", true}, {"000246-other-plan-gate.md", true}, {"000245-x-plan.md", false}, {"0002460-x.md", false}, {"dir/000246-x-plan.md", false}} {
		if got := planArtifactBelongsToIssue("000246-x.md", tc.name); got != tc.want {
			t.Fatalf("%s=%v want %v", tc.name, got, tc.want)
		}
	}
	if planArtifactBelongsToIssue("bad", "bad-plan.md") {
		t.Fatal("invalid issue matched")
	}
}
func TestPublishedIssueContent(t *testing.T) {
	original := "---\nid: 000246\nstatus: codecomplete\nactual_hours: 1\nupdated: 2026-01-01\ncustom: retained\n---\n# Keep body\n"
	fm, body, err := issue.Parse(original)
	if err != nil {
		t.Fatal(err)
	}
	got, err := publishedIssueContent(fm, body, "2026-09-23")
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(strings.Replace(original, "status: codecomplete", "status: done", 1), "updated: 2026-01-01", "updated: 2026-09-23", 1)
	if string(got) != want {
		t.Fatalf("got %q want %q", got, want)
	}
	for _, status := range []string{"working", "open", "done"} {
		if _, err := publishedIssueContent(strings.Replace(fm, "codecomplete", status, 1), body, "2026-09-23"); err == nil {
			t.Fatalf("published status %s", status)
		}
	}
}
