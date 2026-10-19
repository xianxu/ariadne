//go:build unix

package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSweepStaleTemp(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	old := now.Add(-25 * time.Hour)
	mk := func(name string, mtime time.Time) {
		p := filepath.Join(dir, name)
		if err := os.Mkdir(p, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, mtime, mtime); err != nil {
			t.Fatal(err)
		}
	}
	mk("go-build111", old)
	mk("go-build222", now)
	mk("suite-100-1", now)   // dead pid
	mk("suite-200-2", now)   // live pid
	mk("suite-200-3", old)   // live pid, but older than a day: a reused pid
	mk("TestKeep123", old)   // not ours
	mk("go-lint-100-4", now) // a hyphenated check name, dead pid
	mk("go-lint-200-5", now) // a hyphenated check name, live pid
	alive := func(pid int) bool { return pid == 200 }

	sweepStaleTemp(dir, false, now, alive)
	if _, err := os.Stat(filepath.Join(dir, "suite-100-1")); err != nil {
		t.Fatal("swept a run folder in a folder that holds none")
	}
	sweepStaleTemp(dir, true, now, alive)
	want := []string{"TestKeep123", "go-build222", "go-lint-200-5", "suite-200-2"}
	entries, _ := os.ReadDir(dir)
	var got []string
	for _, e := range entries {
		got = append(got, e.Name())
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("left %v, want %v", got, want)
	}
}

func TestScratchLeaksAttributesOnlyThisRun(t *testing.T) {
	tests := map[string]string{"TestFoo": "cmd/a", "TestFooBar": "cmd/b"}
	before := map[string]bool{"/T/TestFoo1": true}
	after := map[string]bool{
		"/T/TestFoo1":            true, // there before: not this run's
		"/T/TestFoo22":           true,
		"/T/TestFooBarsubcase33": true, // a subtest: the slash is dropped
		"/T/TestOther44":         true, // a test this worktree doesn't declare
		"/T/go-build55":          true,
		"/T/tmp.abc":             true,
	}
	byPkg, unattributed := scratchLeaks(before, after, tests)
	if unattributed != 3 || len(byPkg["cmd/a"]) != 1 || len(byPkg["cmd/b"]) != 1 {
		t.Fatalf("byPkg=%v unattributed=%d", byPkg, unattributed)
	}
	var out bytes.Buffer
	reportScratchLeaks(&out, "suite", byPkg, map[string]int64{"/T/TestFoo22": 2048, "/T/TestFooBarsubcase33": 10}, unattributed)
	for _, want := range []string{"unattributed: 3", "2 t.TempDir folder(s), 2.0 KiB", "cmd/b: 1, 10 B", "/T/TestFooBarsubcase33 (10 B)", "cmd/a: 1, 2.0 KiB"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("report lacks %q:\n%s", want, out.String())
		}
	}
	// The names list is capped; the count and size stay whole.
	many := map[string][]string{"cmd/c": nil}
	for i := 0; i < leakListCap+3; i++ {
		many["cmd/c"] = append(many["cmd/c"], fmt.Sprintf("/T/TestC%d", i))
	}
	out.Reset()
	reportScratchLeaks(&out, "suite", many, map[string]int64{}, 0)
	if !strings.Contains(out.String(), "cmd/c: 23,") || !strings.Contains(out.String(), "… 3 more") {
		t.Fatalf("capped report:\n%s", out.String())
	}
	out.Reset()
	reportScratchLeaks(&out, "suite", map[string][]string{}, nil, 0)
	if out.Len() != 0 {
		t.Fatalf("a clean run reported: %q", out.String())
	}
}

func TestDeclaredTests(t *testing.T) {
	wt := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(wt, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("pkg/x/x_test.go", "package x\nfunc TestX(t *testing.T) {}\nfunc helper() {}\n")
	write("pkg/x/testdata/y_test.go", "func TestIgnored(t *testing.T) {}\n")
	write("main.go", "func TestNotATest(t *testing.T) {}\n")
	got := declaredTests(wt)
	if len(got) != 1 || got["TestX"] != filepath.Join("pkg", "x") {
		t.Fatalf("declaredTests = %v", got)
	}
}

// A suite's child gets TMPDIR and GOTMPDIR in its own run folder on the
// scratch root, removed after the run; with the feature off it inherits.
func TestRunUnderSlotUsesScratch(t *testing.T) {
	for _, mode := range []string{"scratch", "off"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			if mode == "off" {
				t.Setenv("WF_TEST_SCRATCH", "off")
			} else {
				t.Setenv("WF_TEST_SCRATCH", root)
			}
			d, sha, _, _ := localRunRepo(t, "skipped")
			seen := filepath.Join(t.TempDir(), "env")
			c := localCheck{Name: "suite", Command: []string{"sh", "-c", `printf '%s\n%s\n%s\n' "$TMPDIR" "$GOTMPDIR" "$(pwd -P)" > "$1"; touch "$TMPDIR/probe"`, "fixture", seen}, TimeoutMinutes: 1}
			if _, err := runLocalCheck(context.Background(), d, c, sha); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(seen)
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
			// The run's checkout follows the same rule (#468).
			checkouts := os.TempDir()
			if mode == "scratch" {
				checkouts = filepath.Join(root, "local-checks")
			}
			if real, err := filepath.EvalSymlinks(checkouts); err == nil {
				checkouts = real
			}
			if len(lines) != 3 || filepath.Dir(lines[2]) != checkouts || !strings.HasPrefix(filepath.Base(lines[2]), localCheckTmpPrefix) {
				t.Fatalf("checkout at %q, want under %s", lines, checkouts)
			}
			if mode == "off" {
				if lines[0] != os.Getenv("TMPDIR") {
					t.Fatalf("off: TMPDIR = %q, want the inherited %q", lines[0], os.Getenv("TMPDIR"))
				}
				return
			}
			base := filepath.Join(root, "tmp")
			if lines[0] != lines[1] || filepath.Dir(lines[0]) != base || !strings.HasPrefix(filepath.Base(lines[0]), "suite-") {
				t.Fatalf("child env = %q, want one run folder under %s", lines, base)
			}
			if _, err := os.Stat(lines[0]); !os.IsNotExist(err) {
				t.Fatalf("run folder %s not removed after the run: %v", lines[0], err)
			}
		})
	}
}

// The stale-checkout sweep looks on the RAM disk as well as the system temp
// folder: a killed run's checkout is reclaimed wherever it started.
func TestSweepStaleLocalWorktreesBothRoots(t *testing.T) {
	root := t.TempDir()
	t.Setenv("WF_TEST_SCRATCH", root)
	d, _, _, _ := localRunRepo(t, "skipped")
	ram := filepath.Join(root, "local-checks")
	stale := filepath.Join(ram, localCheckTmpPrefix+"old")
	fresh := filepath.Join(ram, localCheckTmpPrefix+"new")
	for _, p := range []string{stale, fresh} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	past := time.Now().Add(-localCheckStaleAge - time.Hour)
	if err := os.Chtimes(stale, past, past); err != nil {
		t.Fatal(err)
	}
	sweepStaleLocalWorktrees(d)
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale checkout on the RAM disk survived: %v", err)
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Fatalf("fresh checkout swept: %v", err)
	}
}
