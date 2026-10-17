package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// unstamped is a section body without its freshness stamp.
func unstamped(body string) string {
	_, rest := splitStamp(body)
	return rest
}

var clock = time.Date(2026, 10, 16, 17, 5, 0, 0, time.Local)

func fixedNow() time.Time { return clock }

func noLanding(string) (Landing, error) { return Landing{}, nil }

// Every write stamps its section's first line; a stamp the operator deleted
// comes back, and one already in the current section or the body is replaced,
// never doubled. The operator's own section is never stamped.
func TestWriteStamps(t *testing.T) {
	f := newFake(map[string]string{
		"Needs you":         "\n1. old",
		"Threads":           "\n_updated 2026-10-15 09:00_\n\n### Old\nx",
		instructionsHeading: "\n1. ask",
	})
	file, body := statusFile(t, "")
	d := deps{ss: f, now: fixedNow, landing: noLanding}
	const stamp = "_updated 2026-10-16 17:05_\n\n"

	if _, err := run(t, d, "status", "section", file, "Needs you", body("1. new")); err != nil {
		t.Fatal(err)
	}
	if got := f.sections["Needs you"]; got != stamp+"1. new" {
		t.Errorf("unstamped section after a write: %q", got)
	}
	if _, err := run(t, d, "status", "section", file, "Needs you", body("_updated 2020-01-01 00:00_\n\n1. again")); err != nil {
		t.Fatal(err)
	}
	if got := f.sections["Needs you"]; got != stamp+"1. again" {
		t.Errorf("a stamp in the body file: %q", got)
	}
	if _, err := run(t, d, "status", "thread", file, body("### New\ny\n")); err != nil {
		t.Fatal(err)
	}
	if got := f.sections["Threads"]; got != stamp+"### Old\nx\n\n### New\ny" {
		t.Errorf("thread upsert over a stamped section: %q", got)
	}
	if _, err := run(t, d, "status", "section", file, instructionsHeading, body("1. ask → [answer](qa/a.md)")); err != nil {
		t.Fatal(err)
	}
	if got := f.sections[instructionsHeading]; strings.Contains(got, "_updated") {
		t.Errorf("the operator's section got a stamp: %q", got)
	}
}

// A stamp the operator put on their own section is not a bullet.
func TestStampIsNotABullet(t *testing.T) {
	bs := instructionBullets("_updated 2026-10-16 17:05_\n\n1. ask\n2. another")
	if len(bs) != 2 || bs[0].Lines[0] != "1. ask" {
		t.Errorf("bullets = %+v", bs)
	}
}

func TestSplitStamp(t *testing.T) {
	for _, tc := range []struct{ in, rest string }{
		{"\n_updated 2026-10-16 17:05_\n\nbody", "body"},
		{"_updated 2026-10-16 17:05_\nbody", "body"},
		{"\nbody\n_updated 2026-10-16 17:05_", "\nbody\n_updated 2026-10-16 17:05_"}, // only the first line
		{"_updated yesterday_\nbody", "_updated yesterday_\nbody"},
	} {
		at, rest := splitStamp(tc.in)
		if rest != tc.rest || at.IsZero() != (rest == tc.in) {
			t.Errorf("splitStamp(%q) = %v, %q", tc.in, at, rest)
		}
	}
}

// check flags a section past its budget, one older than the newest landing,
// and an unstamped one; Needs you only when unstamped; never the operator's
// section; a missing section is no finding. Owed ends with the same lines.
func TestStatusCheck(t *testing.T) {
	st := func(at string) string { return "\n_updated 2026-10-16 " + at + "_\n\nbody" }
	f := newFake(map[string]string{
		instructionsHeading: "\n1. ask",
		"Needs you":         st("01:00"), // 16 h, no budget
		"Threads":           st("14:00"), // 3 h > 2 h
		"Slots":             st("16:30"), // fresh by age; pair landed 16:55
	})
	file, _ := statusFile(t, "")
	land := func(string) (Landing, error) {
		return Landing{Repo: "pair", Commit: "0123456789abcdef", At: time.Date(2026, 10, 16, 16, 55, 0, 0, time.Local)}, nil
	}
	d := deps{ss: f, now: fixedNow, landing: land}
	out, err := run(t, d, "status", "check", file)
	if err != nil {
		t.Fatal(err)
	}
	want := "stale: ## Threads (updated 10-16 14:00; 3h05m old, budget 2h; pair merge 0123456789ab landed 10-16 16:55)\n" +
		"stale: ## Slots (updated 10-16 16:30; pair merge 0123456789ab landed 10-16 16:55)\n"
	if out != want {
		t.Errorf("check:\n%s\nwant:\n%s", out, want)
	}

	f.sections["Needs you"] = "\n- an ask"
	f.sections["Threads"] = st("17:00")
	f.sections["Slots"] = st("17:00")
	out, _ = run(t, d, "status", "owed", file)
	if !strings.HasSuffix(out, "stale: ## Needs you (unstamped)\n") || strings.Contains(out, "Threads") {
		t.Errorf("owed after the TL refreshed Threads and Slots:\n%s", out)
	}

	f.sections["Needs you"] = st("17:00")
	if out, _ = run(t, d, "status", "check", file); out != "fresh: every budgeted section\n" {
		t.Errorf("all fresh: %q", out)
	}
}

// The landing source selects actual mainline merges, ignoring direct main
// publications and merges made inside an issue branch before it was landed.
func TestGitNewestLanding(t *testing.T) {
	parent := t.TempDir()
	makeRepo := func(name string, marks ...string) (string, func(int64, string, ...string) string) {
		t.Helper()
		dir := filepath.Join(parent, name)
		for _, m := range append(marks, ".") {
			if err := os.MkdirAll(filepath.Join(dir, m), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		runGit := func(unix int64, args ...string) string {
			t.Helper()
			cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
			date := fmt.Sprintf("@%d +0000", unix)
			cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull,
				"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t", "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
				"GIT_COMMITTER_DATE="+date, "GIT_AUTHOR_DATE="+date)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("git %v: %v\n%s", args, err, out)
			}
			return strings.TrimSpace(string(out))
		}
		runGit(1, "init", "-q")
		tree := runGit(1, "write-tree")
		return dir, func(unix int64, message string, parents ...string) string {
			t.Helper()
			args := []string{"commit-tree", tree, "-m", message}
			for _, p := range parents {
				args = append(args, "-p", p)
			}
			head := runGit(unix, args...)
			runGit(unix, "update-ref", "refs/remotes/origin/main", head)
			return head
		}
	}
	_, _ = makeRepo("ops", "teams")
	file := filepath.Join(parent, "ops", "teams", "tl-status-op.md")
	_, direct := makeRepo("direct", "workshop/issues")
	base := direct(500, "initial")
	for i, msg := range []string{"#1: issue: new", "#1: issue: publish", "archive completed issues to history", "project: tick", "Merge pull request #99"} {
		base = direct(int64(9000+i), msg, base)
	}
	if got, err := gitNewestLanding(file); err != nil || !got.At.IsZero() {
		t.Fatalf("without a merged PR: %+v, %v; want no landing", got, err)
	}
	for _, name := range []string{"ariadne", "pair", "brain", "notes", "missing-ref"} {
		marks := []string{"workshop/issues"}
		if name == "brain" {
			marks = append(marks, ".brain")
		}
		if name == "notes" {
			marks = nil
		}
		_, commit := makeRepo(name, marks...)
		if name == "missing-ref" {
			continue
		}
		root := commit(500, "initial")
		main := commit(600, "main", root)
		branch := commit(700, "issue", root)
		// Deliberately newer nested merge: it must not become the landing.
		nested := commit(10000, "merge main into issue", branch, main)
		at := int64(1000)
		if name == "pair" {
			at = 2000
		}
		if name == "brain" || name == "notes" {
			at = 11000
		}
		merge := commit(at, "Merge pull request #42", main, nested)
		commit(12000, "archive completed issues to history", merge)
	}
	got, err := gitNewestLanding(file)
	if err != nil {
		t.Fatal(err)
	}
	if got.Repo != "pair" || got.At.Unix() != 2000 {
		t.Fatalf("newest landing = %+v, want pair at 2000", got)
	}
	merge, err := git(filepath.Join(parent, "pair"), "rev-parse", "refs/remotes/origin/main^")
	if err != nil {
		t.Fatal(err)
	}
	f := newFake(map[string]string{"Slots": "_updated " + time.Unix(1000, 0).Format("2006-01-02 15:04") + "_\n\nbody"})
	out, err := run(t, deps{ss: f, now: func() time.Time { return time.Unix(2001, 0) }, landing: gitNewestLanding}, "status", "check", file)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"pair", merge[:12], got.At.Format("01-02 15:04")} {
		if !strings.Contains(out, want) {
			t.Errorf("stale diagnostic lacks %q: %s", want, out)
		}
	}
}

// A landing in the same minute as the stamp is not newer (the stamp has
// minutes only); an unreadable landing source or section is reported, not
// silently fresh.
func TestStatusCheckEdges(t *testing.T) {
	f := newFake(map[string]string{"Slots": "\n_updated 2026-10-16 17:05_\n\nbody"})
	file, _ := statusFile(t, "")
	sameMinute := func(string) (Landing, error) {
		return Landing{Repo: "pair", Commit: "0123456789abcdef", At: time.Date(2026, 10, 16, 17, 5, 40, 0, time.Local)}, nil
	}
	if got := staleLines(deps{ss: f, now: fixedNow, landing: sameMinute}, file); len(got) != 0 {
		t.Errorf("a landing in the stamp's minute: %q", got)
	}
	broken := func(string) (Landing, error) { return Landing{}, fmt.Errorf("git: not a repository") }
	got := staleLines(deps{ss: failingRead{f, "Threads"}, now: fixedNow, landing: broken}, file)
	want := []string{"check: newest landing unknown: git: not a repository", "check: ## Threads unread: couch: timed out"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("check lines = %q, want %q", got, want)
	}
}

// failingRead fails reads of one section with something other than a missing heading.
type failingRead struct {
	*fakeSessionSync
	heading string
}

func (r failingRead) ReadSection(file, heading string) (string, string, error) {
	if heading == r.heading {
		return "", "", fmt.Errorf("couch: timed out")
	}
	return r.fakeSessionSync.ReadSection(file, heading)
}

func TestLandingHelp(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"status", "check", "--help"}} {
		out := helpOutput(t, args...)
		for _, want := range []string{"first-parent merge", "filing", "merge hash"} {
			if !strings.Contains(out, want) {
				t.Errorf("help %v lacks %q", args, want)
			}
		}
	}
}
