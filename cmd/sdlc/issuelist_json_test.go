package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func decodeIssueList(t *testing.T, out string) IssueListJSON {
	t.Helper()
	var doc IssueListJSON
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("decode %q: %v", out, err)
	}
	if doc.SchemaVersion != issueListSchema {
		t.Fatalf("schema_version %d", doc.SchemaVersion)
	}
	return doc
}

// #633: --json lists every card, the terminal card-only ones the active views
// leave out included, with the fields the weekly trend reads; an unreadable
// card stays a row; the text listing is unchanged.
func TestIssueListJSONIncludesTerminalCards(t *testing.T) {
	openPath, openCard, openDetailPath, openDetail := seededIssue(t, "000001", "open")
	workPath, workCard, _, _ := seededIssue(t, "000002", "working")
	workCard = strings.Replace(workCard, "status: open", "status: working", 1)
	donePath, doneCard, _, _ := seededIssue(t, "000003", "done")
	doneCard = strings.Replace(doneCard, "status: open", "status: done\nactual_hours: N/A\nkind: bug", 1)
	doneCard = strings.Replace(doneCard, "updated: 2026-09-01", "updated: 2026-09-03", 1)
	claimed, err := issue.SetCardClaimant([]byte(doneCard), issue.Claimant{Operator: "o", Machine: issue.MachineFingerprint("UUID-1"),
		MachineName: "m", Workspace: "ariadne:3", Worktree: "/w/ariadne-slot3/ariadne", Repository: "github.com/xianxu/ariadne"})
	if err != nil {
		t.Fatal(err)
	}
	badPath, badCard, _, _ := seededIssue(t, "000004", "bad")
	badCard = strings.Replace(badCard, "status: open", "status: working\nclaimant:\n    operator: a\n    machine: RAW-MACHINE-ID\n    machine_name: m\n    worktree: /w\n    repository: r", 1)
	newTrackerRepo(t, map[string]string{openPath: openCard, workPath: workCard, donePath: string(claimed), badPath: badCard},
		map[string]string{openDetailPath: openDetail})

	out, err := runSdlc(t, "issue", "list", "--json")
	if err != nil {
		t.Fatalf("issue list --json: %v\n%s", err, out)
	}
	doc := decodeIssueList(t, out)
	rows := map[string]CardRow{}
	for _, c := range doc.Cards {
		rows[c.ID] = c
	}
	if !doc.Tracker || len(rows) != 4 {
		t.Fatalf("want 4 rows from a tracker repository, got tracker=%v %+v", doc.Tracker, doc.Cards)
	}
	done := rows["000003"]
	if done.Status != "done" || done.Kind != "bug" || done.Created != "2026-09-01" || done.Updated != "2026-09-03" || done.ClaimantWorkspace != "ariadne:3" {
		t.Fatalf("done card row: %+v", done)
	}
	if rows["000001"].Status != "open" || rows["000002"].Status != "working" {
		t.Fatalf("open/working rows: %+v %+v", rows["000001"], rows["000002"])
	}
	if bad := rows["000004"]; bad.Status != "unreadable" || !strings.Contains(bad.Unreadable, "fingerprint") {
		t.Fatalf("unreadable card row: %+v", bad)
	}

	for _, q := range []struct {
		args []string
		want string
	}{
		{[]string{"--kind", "bug"}, "000003,000004"},
		// the triage's closes query: terminal card-only cards, plus the unreadable row
		{[]string{"--status", "done"}, "000003,000004"},
		{[]string{"--status", "open"}, "000001,000004"},
	} {
		out, err := runSdlc(t, append([]string{"issue", "list", "--json"}, q.args...)...)
		if err != nil {
			t.Fatal(err)
		}
		var ids []string
		for _, c := range decodeIssueList(t, out).Cards {
			ids = append(ids, c.ID)
		}
		if strings.Join(ids, ",") != q.want {
			t.Fatalf("%v kept %v; want %s (the unreadable row is never filtered out)", q.args, ids, q.want)
		}
	}

	out, err = runSdlc(t, "issue", "list")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "000003") || !strings.Contains(out, "000001") {
		t.Fatalf("the text listing changed: a terminal card-only issue appeared, or an open one went missing:\n%s", out)
	}
}

// A repository without a tracker lists its details files, and says so.
func TestIssueListJSONWithoutTracker(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	full := issue.Render(issue.ScaffoldSpec{ID: "000001", Title: "Legacy one", Today: "2026-09-01"})
	writeRepoFile(t, root, "workshop/issues/000001-legacy.md", full)
	out, err := runSdlc(t, "issue", "list", "--json")
	if err != nil {
		t.Fatalf("issue list --json: %v\n%s", err, out)
	}
	doc := decodeIssueList(t, out)
	if doc.Tracker || len(doc.Cards) != 1 || doc.Cards[0].ID != "000001" || doc.Cards[0].Created != "2026-09-01" {
		t.Fatalf("legacy listing: %+v", doc)
	}
}

// Schema version 1's keys are pinned: renaming or removing one, or adding one
// to the document without updating this list, fails here, so a format change
// is a decision (ARCH-EVOLVE) and never a side effect of another view's edit.
func TestIssueListJSONFieldSetIsPinned(t *testing.T) {
	keys := func(v any) []string {
		var out []string
		ty := reflect.TypeOf(v)
		for i := 0; i < ty.NumField(); i++ {
			out = append(out, strings.Split(ty.Field(i).Tag.Get("json"), ",")[0])
		}
		return out
	}
	if got, want := keys(IssueListJSON{}), []string{"schema_version", "tracker", "cards"}; !reflect.DeepEqual(got, want) || issueListSchema != 1 {
		t.Fatalf("v1 document keys %v (schema %d), want %v", got, issueListSchema, want)
	}
	want := []string{"id", "title", "status", "created", "updated", "kind", "severity", "area", "claimant_workspace", "unreadable"}
	if got := keys(CardRow{}); !reflect.DeepEqual(got, want) {
		t.Fatalf("v1 row keys %v, want %v", got, want)
	}
}

// A details file that cannot be parsed is an unreadable row with its path.
func TestCardRowOfUnparsedDetails(t *testing.T) {
	row := cardRowOf(IssueState{ID: "000009", Path: "workshop/issues/000009-x.md"})
	if row.Status != "unreadable" || !strings.Contains(row.Unreadable, "000009-x.md") {
		t.Fatalf("%+v", row)
	}
}
