package main

import (
	"encoding/json"
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
	rows := map[string]IssueState{}
	for _, is := range doc.Issues {
		rows[is.ID] = is
	}
	if !doc.Tracker || len(rows) != 4 {
		t.Fatalf("want 4 rows from a tracker repository, got tracker=%v %+v", doc.Tracker, doc.Issues)
	}
	done := rows["000003"]
	if done.Status != "done" || done.Kind != "bug" || done.Created != "2026-09-01" || done.Updated != "2026-09-03" || done.Owner == nil || done.Owner.Workspace != "ariadne:3" {
		t.Fatalf("done card row: %+v (owner %+v)", done, done.Owner)
	}
	if rows["000001"].Status != "open" || rows["000002"].Status != "working" {
		t.Fatalf("open/working rows: %+v %+v", rows["000001"], rows["000002"])
	}
	if bad := rows["000004"]; bad.Status != "unreadable" || !strings.Contains(bad.Unreadable, "fingerprint") {
		t.Fatalf("unreadable card row: %+v", bad)
	}

	out, err = runSdlc(t, "issue", "list", "--json", "--kind", "bug")
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, is := range decodeIssueList(t, out).Issues {
		ids = append(ids, is.ID)
	}
	if strings.Join(ids, ",") != "000003,000004" {
		t.Fatalf("--kind bug kept %v; want the bug and the unreadable row", ids)
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
	if doc.Tracker || len(doc.Issues) != 1 || doc.Issues[0].ID != "000001" || doc.Issues[0].Created != "2026-09-01" {
		t.Fatalf("legacy listing: %+v", doc)
	}
}
