package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// run executes tl-lead with the given fakes and returns stdout and the error.
func run(t *testing.T, d deps, args ...string) (string, error) {
	t.Helper()
	root := buildRoot(d)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

// statusFile makes a temporary status file (its directory also holds the
// seen-set and .clobbered) and a body file with the given content.
func statusFile(t *testing.T, text string) (file string, body func(string) string) {
	t.Helper()
	dir := t.TempDir()
	file = filepath.Join(dir, "tl-status-op.md")
	if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	n := 0
	return file, func(content string) string {
		n++
		p := filepath.Join(dir, "body"+string(rune('0'+n))+".md")
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
}

func exitCode(err error) int {
	var ee exitError
	if errors.As(err, &ee) {
		return ee.code
	}
	if err != nil {
		return 1
	}
	return 0
}

func TestOperatorLockRefuses(t *testing.T) {
	f := newFake(map[string]string{"Needs you": "1. old"})
	f.holder = "operator"
	file, body := statusFile(t, "")
	_, err := run(t, deps{ss: f}, "status", "section", file, "Needs you", body("1. new"))
	if exitCode(err) != 3 {
		t.Fatalf("exit %d (%v), want 3", exitCode(err), err)
	}
	if !strings.Contains(err.Error(), "the operator holds the lock") || !strings.Contains(err.Error(), "queue") {
		t.Errorf("refusal %q does not say the lock is the operator's and to queue", err)
	}
	if f.writes != 0 || f.writeCalls != 1 {
		t.Errorf("writes=%d calls=%d: a held lock must refuse at once, with no write and no retry", f.writes, f.writeCalls)
	}
	if f.sections["Needs you"] != "1. old" {
		t.Errorf("section changed under the operator's lock: %q", f.sections["Needs you"])
	}
}

// Faults are injected at chosen write calls; each row asserts the exit code,
// the write count and the final section.
func TestWriteSectionRetries(t *testing.T) {
	for _, tc := range []struct {
		name       string
		edits      map[int]map[string]string
		wantExit   int
		wantCalls  int
		wantWrites int
		wantBody   string
	}{
		{name: "no contention", wantCalls: 1, wantWrites: 1, wantBody: "- new\n- old"},
		{name: "stale then success recomputes from the new read",
			edits:     map[int]map[string]string{1: {"Threads": "- theirs\n- old"}},
			wantCalls: 2, wantWrites: 1, wantBody: "- new\n- theirs\n- old"},
		{name: "four stales refuse",
			edits: map[int]map[string]string{
				1: {"Threads": "- e1"}, 2: {"Threads": "- e2"}, 3: {"Threads": "- e3"}, 4: {"Threads": "- e4"}},
			wantExit: 6, wantCalls: 4, wantWrites: 0, wantBody: "- e4"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFake(map[string]string{"Threads": "- old"})
			if tc.edits != nil {
				f.editBefore = tc.edits
			}
			file, body := statusFile(t, "")
			_, err := run(t, deps{ss: f}, "status", "thread", file, body("- new\n"))
			if exitCode(err) != tc.wantExit {
				t.Fatalf("exit %d (%v), want %d", exitCode(err), err, tc.wantExit)
			}
			if f.writeCalls != tc.wantCalls || f.writes != tc.wantWrites {
				t.Errorf("calls=%d writes=%d, want %d/%d", f.writeCalls, f.writes, tc.wantCalls, tc.wantWrites)
			}
			if got := f.sections["Threads"]; got != tc.wantBody {
				t.Errorf("Threads = %q, want %q", got, tc.wantBody)
			}
		})
	}
}

// A submit that lands after the TL's last listing stays owed: --fold-owed
// folds only seen ∩ owed, and the seen-set then drops what was folded.
func TestFoldOwedFoldsOnlySeen(t *testing.T) {
	f := newFake(map[string]string{"Needs you": "1. a"})
	f.owed = []string{"ev-a"}
	file, body := statusFile(t, "")
	if out, err := run(t, deps{ss: f}, "status", "owed", file); err != nil || !strings.Contains(out, "ev-a") {
		t.Fatalf("owed: %q %v", out, err)
	}
	f.owed = append(f.owed, "ev-b") // a submit lands mid-turn
	out, err := run(t, deps{ss: f}, "status", "section", file, "Needs you", body("1. b"), "--fold-owed")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "folded 1") {
		t.Errorf("output %q does not report one fold", out)
	}
	if !slices.Equal(f.owed, []string{"ev-b"}) {
		t.Errorf("owed after the fold = %v, want [ev-b] (unseen stays owed)", f.owed)
	}
	seen, err := loadSeen(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(seen) != 0 {
		t.Errorf("seen-set after folding = %v, want empty (bounded by the owed set)", seen)
	}
}

// Listing owed events replaces the seen-set with the IDs owed now, so it never
// grows past the owed set.
func TestListingPrunesSeenSet(t *testing.T) {
	f := newFake(nil)
	file, _ := statusFile(t, "")
	if err := saveSeen(file, []string{"gone-1", "gone-2"}); err != nil {
		t.Fatal(err)
	}
	f.owed = []string{"ev-c"}
	if _, err := run(t, deps{ss: f}, "status", "owed", file); err != nil {
		t.Fatal(err)
	}
	seen, _ := loadSeen(file)
	if !slices.Equal(seen, []string{"ev-c"}) {
		t.Errorf("seen-set = %v, want [ev-c]", seen)
	}
}

func TestStatusRefusesBadInput(t *testing.T) {
	for _, tc := range []struct {
		name, body, want string
		args          func(file, bodyPath string) []string
	}{
		{"empty section body", "  \n", "is empty",
			func(file, b string) []string { return []string{"status", "section", file, "Needs you", b} }},
		{"thread entry without a bullet", "not a bullet\n", "starts with '- '",
			func(file, b string) []string { return []string{"status", "thread", file, b} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFake(map[string]string{"Needs you": "x", "Threads": "- x"})
			file, body := statusFile(t, "")
			_, err := run(t, deps{ss: f}, tc.args(file, body(tc.body))...)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want one containing %q", err, tc.want)
			}
			if f.writeCalls != 0 {
				t.Errorf("a refused input reached couch (%d write calls)", f.writeCalls)
			}
		})
	}
}

func TestThreadDrop(t *testing.T) {
	f := newFake(map[string]string{"Threads": "- rollout: pair\n  detail line\n- incident: x\n- keep me"})
	file, body := statusFile(t, "")
	if _, err := run(t, deps{ss: f}, "status", "thread", file, body("- rollout: done\n"), "--drop", "- rollout", "--drop", "- incident"); err != nil {
		t.Fatal(err)
	}
	if got, want := f.sections["Threads"], "- rollout: done\n- keep me"; got != want {
		t.Errorf("Threads = %q, want %q", got, want)
	}
}

func TestSubmit(t *testing.T) {
	text := "---\ntype: session-sync\n---\nTo comment: write 🤖[...] beside a line\n## Needs you\n1. ship? 🤖[yes, ship it]\n"
	f := newFake(nil)
	f.diff = "@@ operator edit @@\n"
	f.owed = []string{"ev-1"}
	file, _ := statusFile(t, text)
	if err := os.WriteFile(file+".clobbered", []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, deps{ss: f}, "status", "submit", file)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"@@ operator edit @@", "markers:\n  6: 1. ship? 🤖[yes, ship it]", "removed .clobbered (identical", "owed: ev-1"} {
		if !strings.Contains(out, want) {
			t.Errorf("submit output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "To comment:") {
		t.Error("submit listed the legend line as a marker")
	}
	if _, err := os.Stat(file + ".clobbered"); !os.IsNotExist(err) {
		t.Error("an identical .clobbered was not removed")
	}
	if seen, _ := loadSeen(file); !slices.Equal(seen, []string{"ev-1"}) {
		t.Errorf("submit did not record the owed events as seen: %v", seen)
	}
}

func TestClobberedDiffersThenDone(t *testing.T) {
	file, _ := statusFile(t, "a\nb\n")
	if err := os.WriteFile(file+".clobbered", []byte("a\nmine\nb\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, deps{ss: newFake(nil)}, "status", "clobbered", file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "differs") || !strings.Contains(out, "+mine") {
		t.Errorf("output does not show the clobbered copy's extra line:\n%s", out)
	}
	if _, err := os.Stat(file + ".clobbered"); err != nil {
		t.Fatal("a differing .clobbered was removed before the TL merged it")
	}
	if _, err := run(t, deps{ss: newFake(nil)}, "status", "clobbered", file, "--done"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(file + ".clobbered"); !os.IsNotExist(err) {
		t.Error("--done did not remove .clobbered")
	}
}

func TestSeenMarksWholeFile(t *testing.T) {
	f := newFake(nil)
	file, _ := statusFile(t, "")
	if _, err := run(t, deps{ss: f}, "status", "seen", file); err != nil {
		t.Fatal(err)
	}
	if !f.seenAll {
		t.Error("seen did not read the whole file")
	}
}

func TestLineDiff(t *testing.T) {
	got := lineDiff("a\nb\nc", "a\nx\nc")
	if want := " a\n-b\n+x\n c\n"; got != want {
		t.Errorf("lineDiff = %q, want %q", got, want)
	}
}

func TestSeenSetMalformedIsAnError(t *testing.T) {
	file, _ := statusFile(t, "")
	if err := os.WriteFile(seenPath(file), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	f := newFake(map[string]string{"Needs you": "x"})
	f.owed = []string{"ev"}
	_, body := statusFile(t, "")
	_, err := run(t, deps{ss: f}, "status", "section", file, "Needs you", body("y"), "--fold-owed")
	if err == nil || !strings.Contains(err.Error(), "malformed") {
		t.Errorf("err = %v; a malformed seen-set must refuse a fold, never read as empty", err)
	}
	if f.writes != 0 {
		t.Error("wrote despite an unreadable seen-set")
	}
}
