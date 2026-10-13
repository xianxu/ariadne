package main

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

const areaDecl = `{"version":1,"enums":{"area":["couch/broadcast/web","sdlc"]}}`

func runSdlc(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := buildRoot()
	var out, errs bytes.Buffer
	root.SetArgs(args)
	root.SetOut(&out)
	root.SetErr(&errs)
	err := root.ExecuteContext(context.Background())
	return out.String() + errs.String(), err
}

func TestRepoEnumsAtRev(t *testing.T) {
	dir := testfix.Repo(t, testfix.InitialCommit())
	git := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		return string(out), err
	}
	if d, err := repoEnumsAtRev(git, "HEAD"); err != nil || d.Allowed("area", "sdlc") {
		t.Fatalf("absent: %+v %v", d, err)
	}
	writeRepoFile(t, dir, ".sdlc/vocabulary.json", areaDecl)
	testfix.Git(t, dir, "add", "-A")
	testfix.Git(t, dir, "commit", "-qm", "declare")
	if d, err := repoEnumsAtRev(git, "HEAD"); err != nil || !d.Allowed("area", "couch") {
		t.Fatalf("declared: %+v %v", d, err)
	}
	writeRepoFile(t, dir, ".sdlc/vocabulary.json", `{"version":1,"enums":{"area":["Bad"]}}`)
	testfix.Git(t, dir, "commit", "-qam", "break")
	if _, err := repoEnumsAtRev(git, "HEAD"); err == nil || !strings.Contains(err.Error(), ".sdlc/vocabulary.json") {
		t.Fatalf("invalid: %v", err)
	}
}

// R1 PQ-4: the declaration is read from main, never from the checked-out
// branch, so a branch cannot declare its own values.
func TestLoadRepoEnums_NeverReadsHead(t *testing.T) {
	dir := testfix.Repo(t, testfix.InitialCommit())
	git := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		return string(out), err
	}
	testfix.Git(t, dir, "branch", "-M", "main")
	testfix.Git(t, dir, "switch", "-qc", "feature")
	writeRepoFile(t, dir, ".sdlc/vocabulary.json", areaDecl)
	testfix.Git(t, dir, "add", "-A")
	testfix.Git(t, dir, "commit", "-qm", "declare on a branch")
	rev, err := repoEnumsMainRev(git)
	if err != nil || rev != "refs/heads/main" {
		t.Fatalf("rev %q %v", rev, err)
	}
	if d, err := repoEnumsAtRev(git, rev); err != nil || d.Allowed("area", "sdlc") {
		t.Fatalf("read the branch's declaration: %+v %v", d, err)
	}
	testfix.Git(t, dir, "branch", "-D", "main")
	if _, err := repoEnumsMainRev(git); err == nil {
		t.Fatal("no main, but a rev was named")
	}
}

func TestIssueSetAreaAndSeverity(t *testing.T) {
	cardPath, card, _, _ := seededIssue(t, "000001", "x")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{".sdlc/vocabulary.json": areaDecl})

	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"issue", "set-area", "--issue", "1", "couch/broadcast/web", "sdlc"}, "area: [couch/broadcast/web, sdlc]"},
		{[]string{"issue", "set-area", "--issue", "1", "couch"}, "area: [couch]"}, // an ancestor of a declared value
		{[]string{"issue", "set-severity", "--issue", "1", "high"}, "severity: high"},
	} {
		if out, err := runSdlc(t, c.args...); err != nil {
			t.Fatalf("%v: %v\n%s", c.args, err, out)
		}
		if got := r.card(cardPath); !strings.Contains(got, c.want) {
			t.Fatalf("%v: card lacks %q:\n%s", c.args, c.want, got)
		}
	}
	before := r.card(cardPath)
	for _, bad := range []struct {
		args []string
		want string
	}{
		{[]string{"issue", "set-area", "--issue", "1", "couch/session"}, ".sdlc/vocabulary.json"},
		{[]string{"issue", "set-area", "--issue", "1", "Couch"}, "lowercase"},
		{[]string{"issue", "set-area", "--issue", "1"}, "--clear"},
		{[]string{"issue", "set-severity", "--issue", "1", "urgent"}, "critical, high, medium, low"},
		{[]string{"issue", "set-severity", "--issue", "1"}, "--clear"},
	} {
		out, err := runSdlc(t, bad.args...)
		if err == nil || !strings.Contains(err.Error()+out, bad.want) {
			t.Fatalf("%v: err %v, want %q\n%s", bad.args, err, bad.want, out)
		}
	}
	if after := r.card(cardPath); after != before {
		t.Fatalf("a refused setter changed the card:\n%s", after)
	}
	for _, args := range [][]string{{"issue", "set-area", "--issue", "1", "--clear"}, {"issue", "set-severity", "--issue", "1", "--clear"}} {
		if out, err := runSdlc(t, args...); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	parsed, err := issue.ParseCard([]byte(r.card(cardPath)))
	if err != nil {
		t.Fatal(err)
	}
	if areas, ok, _ := issue.FieldList(parsed.Frontmatter, "area"); ok || len(areas) != 0 {
		t.Fatalf("area not cleared: %v", areas)
	}
	if sev, _ := issue.FieldString(parsed.Frontmatter, "severity"); sev != "" {
		t.Fatalf("severity not cleared: %q", sev)
	}
}

func TestIssueNewAreaSeverity(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, map[string]string{".sdlc/vocabulary.json": areaDecl})
	slotClaimant(t, "ariadne:0")
	f := newIssueFlags()
	f.Areas, f.Severity = []string{"couch/broadcast"}, "critical"
	var out, errs bytes.Buffer
	if err := runIssueNew(context.Background(), &out, &errs, f, []string{"Classified"}); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	card := r.card("workshop/issue-cards/000008-classified.md")
	if !strings.Contains(card, "area: [couch/broadcast]") || !strings.Contains(card, "severity: critical") {
		t.Fatalf("card:\n%s", card)
	}
	if _, err := issue.ParseCard([]byte(card)); err != nil {
		t.Fatal(err)
	}
	f.Areas, f.Severity = []string{"undeclared"}, ""
	err := runIssueNew(context.Background(), &out, &errs, f, []string{"Refused"})
	if err == nil || !strings.Contains(err.Error(), ".sdlc/vocabulary.json") {
		t.Fatalf("undeclared: %v", err)
	}
	if r.card("workshop/issue-cards/000009-refused.md") != "" {
		t.Fatal("a refused issue was reserved")
	}
}

// R1 PQ-3: publishing issue details checks their classification against
// main's declaration (legacy --commit publication).
func TestIssuePublish_RefusesUndeclaredArea(t *testing.T) {
	repo, origin := syncRepo(t)
	writeRepoFile(t, repo, ".sdlc/vocabulary.json", areaDecl)
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-qm", "declare")
	git(t, repo, "push", "-q", "origin", "main")
	publish := func(fm string) error {
		writeSyncIssue(t, repo, "000206-issue-sync-verb.md", "---\nid: 000206\nstatus: open\n"+fm+"---\n\n# x\n\n## Spec\n\nedit\n")
		git(t, repo, "commit", "-qam", "edit")
		var out, errs bytes.Buffer
		err := publishIssueCommit(&out, &errs, repo, strings.TrimSpace(git(t, repo, "rev-parse", "HEAD")), syncIssuesDir, "workshop/history")
		if err != nil {
			git(t, repo, "reset", "-q", "--hard", "HEAD~1") // a refused edit is not the next edit's parent
		}
		return err
	}
	if err := publish("area: [couch/session]\n"); err == nil || !strings.Contains(err.Error(), ".sdlc/vocabulary.json") {
		t.Fatalf("undeclared area published: %v", err)
	}
	if on := git(t, origin, "show", "main:"+issuePath206); strings.Contains(on, "couch/session") {
		t.Fatal("the refused edit reached main")
	}
	if err := publish("severity: urgent\n"); err == nil || !strings.Contains(err.Error(), "severity") {
		t.Fatalf("invalid severity published: %v", err)
	}
	if err := publish("area:\n  - couch/broadcast\nseverity: \"low\"\n"); err != nil {
		t.Fatalf("declared values refused: %v", err)
	}
}

// #348: on-demand validation checks each file's classification and, with --all
// in a tracker repository, every card's (the cards are the authority).
func TestIssueValidateClassification(t *testing.T) {
	prev := validateFrontmatterFn
	validateFrontmatterFn = func(_, _ string) (string, bool, error) { return "", true, nil }
	t.Cleanup(func() { validateFrontmatterFn = prev })
	cardPath, card, detailPath, detail := seededIssue(t, "000001", "x")
	card = strings.Replace(card, "status: open", "status: open\narea: [couch/session]", 1)
	newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{".sdlc/vocabulary.json": areaDecl, detailPath: detail})

	var out, errs bytes.Buffer
	err := runIssueValidate(context.Background(), &out, &errs, &issueValidateFlags{All: true, IssuesDir: "workshop/issues"}, nil)
	if err == nil || !strings.Contains(out.String(), "card #000001") || !strings.Contains(out.String(), "couch/session") {
		t.Fatalf("err %v\n%s\n%s", err, out.String(), errs.String())
	}

	// A file's own classification is checked too.
	bad := strings.Replace(detail, "status: open", "status: open\nseverity: urgent", 1)
	writeRepoFile(t, ".", detailPath, bad)
	out.Reset()
	err = runIssueValidate(context.Background(), &out, &errs, &issueValidateFlags{IssuesDir: "workshop/issues"}, []string{detailPath})
	if err == nil || !strings.Contains(out.String(), "urgent") {
		t.Fatalf("err %v\n%s", err, out.String())
	}
}
