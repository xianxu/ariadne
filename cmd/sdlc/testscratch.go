// testscratch.go — suites write their temp files to the TestScratch RAM disk
// when it is mounted (#468), sparing the shared SSD the fsync storm a suite's
// temp churn causes. runUnderSlot is the one seam: every recorded suite and
// known-flake rerun starts there.
package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/scratchdisk"
)

// staleTempAge is how old a go-build or run folder must be before a sweep
// calls it abandoned.
const staleTempAge = 24 * time.Hour

// scratchPreparation owns a run's temp directory until the child has stopped.
// OnScratch is explicit: a normal fallback also overrides the child's env.
type scratchPreparation struct {
	Env       []string
	Dir       string
	OnScratch bool
	Cleanup   func()
}

func prepareScratch(stderr io.Writer, check string, normalTempDir func() (string, error)) (scratchPreparation, error) {
	// Creation and stale cleanup share every possible root, independently
	// of today's placement. Former fallback runs stay there after recovery.
	_, cleanupRoots, _ := scratchdisk.Roots("tmp", os.TempDir())
	for _, root := range cleanupRoots {
		sweepStaleTemp(root, true, time.Now(), processAlive, stderr)
	}
	makeRun := func(base string, onScratch bool) (scratchPreparation, error) {
		if err := os.MkdirAll(base, 0o755); err != nil {
			return scratchPreparation{}, err
		}
		dir, err := os.MkdirTemp(base, scratchRunPrefix+check+"-"+strconv.Itoa(os.Getpid())+"-")
		if err != nil {
			return scratchPreparation{}, err
		}
		p := scratchPreparation{Env: []string{"TMPDIR=" + dir, "GOTMPDIR=" + dir}, Dir: dir, OnScratch: onScratch, Cleanup: func() { _ = removeTree(dir) }}
		for _, name := range []string{"TMPDIR", "GOTMPDIR"} {
			if inherited := os.Getenv(name); inherited != "" {
				if err := scratchdisk.CheckTempDir(inherited); err != nil {
					fmt.Fprintf(stderr, "[!] inherited %s=%q is unusable (%v); using %s\n", name, inherited, err, dir)
				}
			}
		}
		return p, nil
	}
	root, note := scratchdisk.Use()
	if note != "" {
		fmt.Fprintf(stderr, "[!] %s\n", note)
	}
	if root != "" {
		base := filepath.Join(root, "tmp")
		if p, err := makeRun(base, true); err == nil {
			return p, nil
		}
		fmt.Fprintf(stderr, "[!] test scratch %s not writable; using the normal temp folder\n", base)
	}
	if normalTempDir == nil {
		normalTempDir = scratchdisk.NormalTempDir
	}
	normal, err := normalTempDir()
	if err != nil {
		return scratchPreparation{}, fmt.Errorf("environment failure: cannot prepare TMPDIR/GOTMPDIR: %w", err)
	}
	p, err := makeRun(normal, false)
	if err != nil {
		return scratchPreparation{}, fmt.Errorf("environment failure: cannot prepare TMPDIR/GOTMPDIR in %s: %w", normal, err)
	}
	return p, nil
}

// scratchRunPrefix starts every run folder prepareScratch makes, and the
// sweep matches only folders that carry it: the RAM disk is shared scratch
// space (AGENTS §10), so a sweep built from anything else could take another
// tool's folder.
const scratchRunPrefix = "sdlc-run-"

// runDirRE names a run folder: sdlc-run-<check>-<pid>-<random>. A check name
// may hold hyphens (localchecks.go's name rule), so the pid is the
// second-to-last field.
var runDirRE = regexp.MustCompile(`^` + regexp.QuoteMeta(scratchRunPrefix) + `.+-(\d+)-\d+$`)

// sweepStaleTemp removes, directly under dir, go-build folders older than a
// day and, when dir holds run folders, those whose process is gone (or older
// than a day, against a reused pid). A live go build keeps writing into its folder and suites time
// out well within a day, so age never takes a folder in use. Each removal is
// logged to w before it starts and when it ends (#631), so a setup stall names
// its target. A root excludeScratchSweep registered is left alone.
func sweepStaleTemp(dir string, runDirs bool, now time.Time, alive func(int) bool, w io.Writer) {
	if _, skip := scratchSweepExcluded.Load(sweepRootKey(dir)); skip {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		reason := staleTempReason(e.Name(), info.ModTime(), runDirs, now, alive)
		if reason == "" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if runnerArtifactProtected(path) {
			fmt.Fprintf(w, "[scratch sweep] kept %s (%s): a runner artifact owner protects it\n", path, reason)
			continue
		}
		fmt.Fprintf(w, "[scratch sweep] removing %s (%s)\n", path, reason)
		start := time.Now()
		if err := removeTree(path); err != nil {
			fmt.Fprintf(w, "[scratch sweep] could not remove %s after %s: %v\n", path, time.Since(start).Round(time.Millisecond), err)
			continue
		}
		fmt.Fprintf(w, "[scratch sweep] removed %s in %s\n", path, time.Since(start).Round(time.Millisecond))
	}
}

// staleTempReason says why the sweep may take the folder name, or "" to keep
// it: only go-build folders and our own run folders are ever candidates.
func staleTempReason(name string, mtime time.Time, runDirs bool, now time.Time, alive func(int) bool) string {
	old := now.Sub(mtime) > staleTempAge
	if m := runDirRE.FindStringSubmatch(name); m != nil && runDirs {
		pid, _ := strconv.Atoi(m[1])
		switch {
		case !alive(pid):
			return fmt.Sprintf("owner pid %d exited", pid)
		case old:
			return fmt.Sprintf("older than %s, owner pid %d reused", staleTempAge, pid)
		}
		return ""
	}
	if strings.HasPrefix(name, "go-build") && old {
		return fmt.Sprintf("go-build older than %s", staleTempAge)
	}
	return ""
}

// removeTree is os.RemoveAll that also takes a tree with read-only
// directories, such as a Go module cache a test left under its TMPDIR (#631):
// RemoveAll alone cannot unlink inside them, so a run folder holding one
// survived its own cleanup and failed every later sweep again.
func removeTree(path string) error {
	if err := os.RemoveAll(path); err == nil {
		return nil
	}
	if err := makeTreeWritable(path); err != nil {
		return err
	}
	return os.RemoveAll(path)
}

// makeTreeWritable gives the owner write and search on every directory under
// root. WalkDir never follows a symlink, and the change uses
// AT_SYMLINK_NOFOLLOW, so a symlink swapped in after the walk looked is not
// followed out of the tree.
func makeTreeWritable(root string) error {
	return filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		perm := info.Mode().Perm()
		if perm&0o700 == 0o700 {
			return nil
		}
		// Not through an opened handle: opening needs the read bit a 0o311 or
		// 0o000 directory lacks. No path-based fallback: a kernel without
		// fchmodat2 (Linux before 6.6) returns EOPNOTSUPP, and the removal fails
		// and is logged rather than risk following a swapped-in symlink.
		return unix.Fchmodat(unix.AT_FDCWD, p, uint32(perm|0o700), unix.AT_SYMLINK_NOFOLLOW)
	})
}

// scratchSweepExcluded holds the roots no sweep in this process may touch:
// TestMain registers the machine's shared temp roots, so the hundreds of test
// runs that reach prepareScratch never rescan them (#631).
var scratchSweepExcluded sync.Map

func excludeScratchSweep(roots ...string) {
	for _, r := range roots {
		scratchSweepExcluded.Store(sweepRootKey(r), true)
	}
}

// sweepRootKey names a root by its resolved path, so /tmp and /private/tmp are one.
func sweepRootKey(root string) string {
	if p, err := filepath.EvalSymlinks(root); err == nil {
		return p
	}
	return filepath.Clean(root)
}

// ssdTempDirs are the SSD temp folders a scratched run does not own: the
// system's per-user temp folder and the agents' scratch volume.
func ssdTempDirs() []string {
	var out []string
	if d, err := scratchdisk.UserTempDir(); err == nil && d != "" {
		out = append(out, d)
	}
	return append(out, "/Volumes/AgentScratch/tmp")
}

// listTemp names the entries directly under each folder.
func listTemp(dirs []string) map[string]bool {
	out := map[string]bool{}
	for _, d := range dirs {
		entries, _ := os.ReadDir(d)
		for _, e := range entries {
			out[filepath.Join(d, e.Name())] = true
		}
	}
	return out
}

// tempDirTestRE is a t.TempDir folder: os.MkdirTemp("", pattern), where the
// pattern is the test's full name, at most 64 bytes, with characters outside
// letters, digits and a few punctuation marks (subtests' slashes among them)
// dropped, and the random part is digits.
var tempDirTestRE = regexp.MustCompile(`^Test.*\d$`)

// scratchLeaks splits the entries that appeared during a scratched run into
// those attributable to it (a t.TempDir folder of a test this worktree
// declares, grouped by the test's package) and a count of the rest, which any
// process on the machine may have made.
func scratchLeaks(before, after map[string]bool, tests map[string]string) (byPkg map[string][]string, unattributed int) {
	byPkg = map[string][]string{}
	for p := range after {
		if before[p] {
			continue
		}
		if pkg, ok := attributeTempDir(filepath.Base(p), tests); ok {
			byPkg[pkg] = append(byPkg[pkg], p)
			continue
		}
		unattributed++
	}
	for _, ps := range byPkg {
		sort.Strings(ps)
	}
	return byPkg, unattributed
}

// attributeTempDir finds the declared test whose name (as t.TempDir truncates
// it) begins the folder's name; the longest wins, so TestFooBar beats TestFoo.
func attributeTempDir(name string, tests map[string]string) (string, bool) {
	if !tempDirTestRE.MatchString(name) {
		return "", false
	}
	best := ""
	for t := range tests {
		if strings.HasPrefix(name, t[:min(len(t), 64)]) && len(t) > len(best) {
			best = t
		}
	}
	if best == "" {
		return "", false
	}
	return tests[best], true
}

// declaredTestRE is a top-level Go test function.
var declaredTestRE = regexp.MustCompile(`(?m)^func (Test[A-Za-z0-9_]*)\(`)

// declaredTests maps each Go test function under wt to its package folder,
// relative to wt.
func declaredTests(wt string) map[string]string {
	out := map[string]string{}
	_ = filepath.WalkDir(wt, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if n := d.Name(); path != wt && (strings.HasPrefix(n, ".") || n == "vendor" || n == "testdata" || n == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(wt, filepath.Dir(path))
		for _, m := range declaredTestRE.FindAllSubmatch(src, -1) {
			out[string(m[1])] = rel
		}
		return nil
	})
	return out
}

// leakListCap bounds the names listed per package; the count and size stay whole.
const leakListCap = 20

// treeSizes measures each path's tree in bytes (best effort: an entry removed
// meanwhile counts what was read).
func treeSizes(paths map[string][]string) map[string]int64 {
	out := map[string]int64{}
	for _, ps := range paths {
		for _, p := range ps {
			var n int64
			_ = filepath.WalkDir(p, func(_ string, d os.DirEntry, err error) error {
				if err == nil && !d.IsDir() {
					if info, err := d.Info(); err == nil {
						n += info.Size()
					}
				}
				return nil
			})
			out[p] = n
		}
	}
	return out
}

// reportScratchLeaks warns about temp folders a scratched run left on the SSD
// (#468): per package their count, size and names (the first leakListCap),
// then the unattributed count. It measures, never fails: the run's outcome
// stands.
func reportScratchLeaks(w io.Writer, check string, byPkg map[string][]string, sizes map[string]int64, unattributed int) {
	if len(byPkg) == 0 && unattributed == 0 {
		return
	}
	n, total := 0, int64(0)
	pkgs := make([]string, 0, len(byPkg))
	for pkg, ps := range byPkg {
		pkgs = append(pkgs, pkg)
		n += len(ps)
		for _, p := range ps {
			total += sizes[p]
		}
	}
	sort.Strings(pkgs)
	fmt.Fprintf(w, "[!] local check %s: SSD temp folders gained %d t.TempDir folder(s), %s, from this worktree's tests despite the RAM disk; unattributed: %d (other processes write there too)\n", check, n, humanBytes(total), unattributed)
	for _, pkg := range pkgs {
		ps := byPkg[pkg]
		var size int64
		for _, p := range ps {
			size += sizes[p]
		}
		fmt.Fprintf(w, "    %s: %d, %s\n", pkg, len(ps), humanBytes(size))
		for i, p := range ps {
			if i == leakListCap {
				fmt.Fprintf(w, "      … %d more\n", len(ps)-leakListCap)
				break
			}
			fmt.Fprintf(w, "      %s (%s)\n", p, humanBytes(sizes[p]))
		}
	}
}

func humanBytes(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MiB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KiB", float64(n)/(1<<10))
	}
	return fmt.Sprintf("%d B", n)
}
