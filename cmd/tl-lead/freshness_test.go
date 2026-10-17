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
		return Landing{Repo: "pair", At: time.Date(2026, 10, 16, 16, 55, 0, 0, time.Local)}, nil
	}
	d := deps{ss: f, now: fixedNow, landing: land}
	out, err := run(t, d, "status", "check", file)
	if err != nil {
		t.Fatal(err)
	}
	want := "stale: ## Threads (updated 10-16 14:00; 3h05m old, budget 2h; pair landed 10-16 16:55)\n" +
		"stale: ## Slots (updated 10-16 16:30; pair landed 10-16 16:55)\n"
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

// The landing source reads origin/main in the SDLC repositories beside the
// status file's (from local refs) and skips a brain and a non-SDLC directory,
// however new their commits.
func TestGitNewestLanding(t *testing.T) {
	parent := t.TempDir()
	repo := func(name string, unix int64, marks ...string) {
		t.Helper()
		dir := filepath.Join(parent, name)
		for _, m := range append(marks, ".") {
			if err := os.MkdirAll(filepath.Join(dir, m), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		date := fmt.Sprintf("@%d +0000", unix)
		for _, args := range [][]string{
			{"init", "-q"},
			{"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "x"},
			{"update-ref", "refs/remotes/origin/main", "HEAD"},
		} {
			cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
			cmd.Env = append(os.Environ(), "GIT_COMMITTER_DATE="+date, "GIT_AUTHOR_DATE="+date)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git %v: %v\n%s", args, err, out)
			}
		}
	}
	repo("ops", 3000, "teams")
	repo("ariadne", 1000, "workshop/issues")
	repo("pair", 2000, "workshop/issues")
	repo("brain", 9000, "workshop/issues", ".brain")
	repo("notes", 9000)
	got, err := gitNewestLanding(filepath.Join(parent, "ops", "teams", "tl-status-op.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Repo != "pair" || got.At.Unix() != 2000 {
		t.Errorf("newest landing = %+v, want pair at 2000", got)
	}
}
