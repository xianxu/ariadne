//go:build unix

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/scratchdisk"
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
	mk("sdlc-run-suite-100-1", now)   // dead pid
	mk("sdlc-run-suite-200-2", now)   // live pid
	mk("sdlc-run-suite-200-3", old)   // live pid, but older than a day: a reused pid
	mk("TestKeep123", old)            // not ours
	mk("sdlc-run-go-lint-100-4", now) // a hyphenated check name, dead pid
	mk("sdlc-run-go-lint-200-5", now) // a hyphenated check name, live pid
	mk("other-tool-100-6", old)       // another tool's <name>-<n>-<n> folder: never ours to take
	alive := func(pid int) bool { return pid == 200 }

	sweepStaleTemp(dir, false, now, alive, io.Discard)
	if _, err := os.Stat(filepath.Join(dir, "sdlc-run-suite-100-1")); err != nil {
		t.Fatal("swept a run folder in a folder that holds none")
	}
	sweepStaleTemp(dir, true, now, alive, io.Discard)
	want := []string{"TestKeep123", "go-build222", "other-tool-100-6", "sdlc-run-go-lint-200-5", "sdlc-run-suite-200-2"}
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
			scratchRoot, _ := scratchdisk.Use()
			usingScratch := scratchRoot != ""
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
			if usingScratch {
				checkouts = filepath.Join(root, "local-checks")
			}
			if real, err := filepath.EvalSymlinks(checkouts); err == nil {
				checkouts = real
			}
			if len(lines) != 3 || filepath.Dir(lines[2]) != checkouts || !strings.HasPrefix(filepath.Base(lines[2]), localCheckTmpPrefix) {
				t.Fatalf("checkout at %q, want under %s", lines, checkouts)
			}
			if !usingScratch {
				if lines[0] != lines[1] || filepath.Dir(lines[0]) != os.TempDir() || !strings.HasPrefix(filepath.Base(lines[0]), scratchRunPrefix+"suite-") {
					t.Fatalf("off: want owned normal temp env, got %q", lines)
				}
				if _, err := os.Stat(lines[0]); !os.IsNotExist(err) {
					t.Fatalf("normal run folder survived: %v", err)
				}
				return
			}
			base := filepath.Join(root, "tmp")
			if lines[0] != lines[1] || filepath.Dir(lines[0]) != base || !strings.HasPrefix(filepath.Base(lines[0]), scratchRunPrefix+"suite-") {
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

func TestPrepareScratchUnsupportedFileSystem(t *testing.T) {
	root := t.TempDir()
	t.Setenv("WF_TEST_SCRATCH", root)
	p := scratchdisk.Read()
	if p.Err != nil {
		t.Fatal(p.Err)
	}
	if p.FileSystem != "hfs" {
		t.Skip("requires an HFS fixture root")
	}
	var note bytes.Buffer
	prepared, err := prepareScratch(&note, "suite", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer prepared.Cleanup()
	if len(prepared.Env) != 2 || prepared.OnScratch || !strings.Contains(note.String(), "is hfs; APFS required") || strings.Count(note.String(), "using the normal folder") != 1 {
		t.Fatalf("prepared=%+v note=%q", prepared, note.String())
	}
	if _, err := os.Stat(filepath.Join(root, "tmp")); !os.IsNotExist(err) {
		t.Fatalf("rejected scratch was written: %v", err)
	}
}

func TestPrepareScratchNormalRunCleanup(t *testing.T) {
	root := t.TempDir()
	t.Setenv("WF_TEST_SCRATCH", "off")
	t.Setenv("TMPDIR", root)
	stale := filepath.Join(root, "sdlc-run-suite-1-123")
	keep := filepath.Join(root, "other-tool-1-123")
	for _, dir := range []string{stale, keep} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		old := time.Now().Add(-25 * time.Hour)
		if err := os.Chtimes(dir, old, old); err != nil {
			t.Fatal(err)
		}
	}
	var note bytes.Buffer
	p, err := prepareScratch(&note, "suite", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Cleanup()
	if p.OnScratch || filepath.Dir(p.Dir) != root {
		t.Fatalf("normal preparation: %+v", p)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale normal run survived: %v", err)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("swept another tool: %v", err)
	}
	p.Cleanup()
	if _, err := os.Stat(p.Dir); !os.IsNotExist(err) {
		t.Fatalf("normal run survived cleanup: %v", err)
	}
}

// darwinNormalFallback owns the Darwin per-user fallback in the fixture.
func darwinNormalFallback(t *testing.T) (root, normal, blocked string) {
	t.Helper()
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin per-user fallback fixture; Unix candidates covered in scratchdisk")
	}
	root = t.TempDir()
	normal = filepath.Join(root, "normal")
	bin := filepath.Join(root, "bin")
	blocked = filepath.Join(root, "blocked")
	for _, dir := range []string{normal, bin} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(blocked, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "getconf"), []byte("#!/bin/sh\nprintf '%s\n' \"$SDLC_USER_TEMP_FIXTURE\"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SDLC_USER_TEMP_FIXTURE", normal)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("WF_TEST_SCRATCH", "off")
	return root, normal, blocked
}

func TestSweepStaleLocalWorktreesRepairedRoot(t *testing.T) {
	d, _, _, _ := localRunRepo(t, "skipped")
	root, normal, blocked := darwinNormalFallback(t)
	stale := filepath.Join(normal, localCheckTmpPrefix+"old")
	fresh := filepath.Join(normal, localCheckTmpPrefix+"new")
	if err := os.Mkdir(fresh, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, inherited := range []string{filepath.Join(blocked, "tmp"), root} {
		t.Setenv("TMPDIR", inherited)
		if err := os.Mkdir(stale, 0o700); err != nil {
			t.Fatal(err)
		}
		old := time.Now().Add(-localCheckStaleAge - time.Hour)
		if err := os.Chtimes(stale, old, old); err != nil {
			t.Fatal(err)
		}
		sweepStaleLocalWorktrees(d)
		if _, err := os.Stat(stale); !os.IsNotExist(err) {
			t.Fatalf("repaired-root stale checkout survived with TMPDIR=%s: %v", inherited, err)
		}
		if _, err := os.Stat(fresh); err != nil {
			t.Fatalf("fresh checkout removed: %v", err)
		}
	}
}

func TestPrepareScratchRepairedRunRootAfterRecovery(t *testing.T) {
	root, normal, blocked := darwinNormalFallback(t)
	stale := filepath.Join(normal, "sdlc-run-suite-1-123")
	live := filepath.Join(normal, fmt.Sprintf("sdlc-run-suite-%d-456", os.Getpid()))
	other := filepath.Join(normal, "other-tool-1-789")
	for _, path := range []string{live, other} {
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, inherited := range []string{filepath.Join(blocked, "tmp"), root} {
		t.Setenv("TMPDIR", inherited)
		if err := os.Mkdir(stale, 0o700); err != nil {
			t.Fatal(err)
		}
		old := time.Now().Add(-staleTempAge - time.Hour)
		if err := os.Chtimes(stale, old, old); err != nil {
			t.Fatal(err)
		}
		var note bytes.Buffer
		p, err := prepareScratch(&note, "suite", nil)
		if err != nil {
			t.Fatal(err)
		}
		p.Cleanup()
		if _, err := os.Stat(stale); !os.IsNotExist(err) {
			t.Fatalf("former fallback run survives with TMPDIR=%s: %v", inherited, err)
		}
		for _, path := range []string{live, other} {
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("fresh/live or unrelated folder removed: %s: %v", path, err)
			}
		}
	}
}

// readOnlyModCache builds what a Go test leaves when it sets GOMODCACHE under
// its TMPDIR (#631): a module tree whose directories are read-only, which a
// plain os.RemoveAll cannot take.
func readOnlyModCache(t *testing.T, dir string) {
	t.Helper()
	mod := filepath.Join(dir, "launcher-1", "go", "pkg", "mod", "example.com", "m@v1.0.0", "sub")
	if err := os.MkdirAll(mod, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mod, "x.go"), []byte("package sub\n"), 0o444); err != nil {
		t.Fatal(err)
	}
	for p := mod; p != filepath.Join(dir, "launcher-1", "go", "pkg", "mod"); p = filepath.Dir(p) {
		if err := os.Chmod(p, 0o555); err != nil {
			t.Fatal(err)
		}
	}
	// Directories without owner read too: a fix that must open them fails here.
	for name, perm := range map[string]os.FileMode{"search-only": 0o311, "none": 0o000} {
		p := filepath.Join(dir, "launcher-1", name)
		if err := os.MkdirAll(filepath.Join(p, "inner"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(p, perm); err != nil {
			t.Fatal(err)
		}
	}
	// Leave nothing behind whatever the code under test does: repair with a
	// walk of its own, never makeTreeWritable, so a broken fix (or a review's
	// revert check) cannot strand the tree.
	t.Cleanup(func() { removeFixtureTree(t, dir) })
}

func removeFixtureTree(t *testing.T, dir string) {
	t.Helper()
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			_ = os.Chmod(p, 0o755)
		}
		return nil
	})
	if err := os.RemoveAll(dir); err != nil {
		t.Errorf("fixture tree left behind: %v", err)
	}
	if _, err := os.Lstat(dir); !os.IsNotExist(err) {
		t.Errorf("fixture tree %s still exists: %v", dir, err)
	}
}

func TestSweepStaleTempRemovesReadOnlyTrees(t *testing.T) {
	dir := t.TempDir()
	dead := filepath.Join(dir, "sdlc-run-suite-100-1")
	readOnlyModCache(t, dead)
	outside := t.TempDir()
	if err := os.Chmod(outside, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(outside, 0o755) })
	if err := os.Symlink(outside, filepath.Join(dead, "link")); err != nil {
		t.Fatal(err)
	}
	var log bytes.Buffer
	sweepStaleTemp(dir, true, time.Now(), func(int) bool { return false }, &log)
	if _, err := os.Stat(dead); !os.IsNotExist(err) {
		t.Fatalf("a dead run folder with a read-only tree survived: %v\n%s", err, log.String())
	}
	if info, err := os.Stat(outside); err != nil || info.Mode().Perm() != 0o555 {
		t.Fatalf("the sweep followed a symlink out of the tree: %v %v", info.Mode(), err)
	}
	for _, want := range []string{"removing " + dead + " (owner pid 100 exited)", "removed " + dead + " in "} {
		if !strings.Contains(log.String(), want) {
			t.Fatalf("log lacks %q:\n%s", want, log.String())
		}
	}
}

func TestPrepareScratchCleanupRemovesReadOnlyTrees(t *testing.T) {
	t.Setenv("WF_TEST_SCRATCH", "off")
	t.Setenv("TMPDIR", t.TempDir())
	var note bytes.Buffer
	p, err := prepareScratch(&note, "suite", nil)
	if err != nil {
		t.Fatal(err)
	}
	readOnlyModCache(t, p.Dir)
	p.Cleanup()
	if _, err := os.Stat(p.Dir); !os.IsNotExist(err) {
		t.Fatalf("a run's own cleanup left its read-only tree: %v", err)
	}
}

// A registered root is never swept, and TestMain registers the machine's
// shared temp roots, so tests only ever sweep roots they made (#631).
func TestExcludedScratchRootIsNeverSwept(t *testing.T) {
	dir := t.TempDir()
	dead := filepath.Join(dir, "sdlc-run-suite-100-1")
	if err := os.Mkdir(dead, 0o755); err != nil {
		t.Fatal(err)
	}
	excludeScratchSweep(dir)
	t.Cleanup(func() { scratchSweepExcluded.Delete(sweepRootKey(dir)) })
	var log bytes.Buffer
	sweepStaleTemp(dir, true, time.Now(), func(int) bool { return false }, &log)
	if _, err := os.Stat(dead); err != nil || log.Len() != 0 {
		t.Fatalf("an excluded root was swept: %v\n%s", err, log.String())
	}
	_, shared, _ := scratchdisk.Roots("tmp", os.TempDir())
	for _, root := range append(shared, "/tmp") {
		if _, ok := scratchSweepExcluded.Load(sweepRootKey(root)); !ok {
			t.Errorf("shared root %s is sweepable from tests", root)
		}
	}
}

func TestStaleTempReason(t *testing.T) {
	now := time.Now()
	old := now.Add(-staleTempAge - time.Hour)
	alive := func(pid int) bool { return pid == 200 }
	for _, c := range []struct {
		name    string
		mtime   time.Time
		runDirs bool
		want    string
	}{
		{"go-build1", old, false, "go-build older than 24h0m0s"},
		{"go-build2", now, true, ""},
		{"sdlc-run-suite-100-1", now, true, "owner pid 100 exited"},
		{"sdlc-run-suite-100-1", now, false, ""},
		{"sdlc-run-suite-200-2", now, true, ""},
		{"sdlc-run-suite-200-3", old, true, "older than 24h0m0s, owner pid 200 reused"},
		{"sdlc-run-go-lint-100-4", now, true, "owner pid 100 exited"},
		{"other-tool-100-6", old, true, ""},
	} {
		if got := staleTempReason(c.name, c.mtime, c.runDirs, now, alive); got != c.want {
			t.Errorf("%s (runDirs %v): %q, want %q", c.name, c.runDirs, got, c.want)
		}
	}
}
