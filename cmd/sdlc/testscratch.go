// testscratch.go — suites write their temp files to the TestScratch RAM disk
// when it is mounted (#468), sparing the shared SSD the fsync storm a suite's
// temp churn causes. runUnderSlot is the one seam: every recorded suite and
// known-flake rerun starts there.
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/scratchdisk"
)

// staleTempAge is how old a go-build or run folder must be before a sweep
// calls it abandoned.
const staleTempAge = 24 * time.Hour

// prepareScratch gives a run its own temp folder on the RAM disk: the env
// entries that point TMPDIR and GOTMPDIR there, and a cleanup that removes it.
// It returns no env when the run should use the normal temp folder. Either
// way it first sweeps what crashed runs and old builds left behind.
func prepareScratch(stderr io.Writer, check string) (env []string, dir string, cleanup func()) {
	cleanup = func() {}
	sweepStaleTemp(os.TempDir(), false, time.Now(), processAlive)
	root, note := scratchdisk.Use()
	if note != "" {
		fmt.Fprintf(stderr, "[!] %s\n", note)
	}
	if root == "" {
		return nil, "", cleanup
	}
	base := filepath.Join(root, "tmp")
	sweepStaleTemp(base, true, time.Now(), processAlive)
	if err := os.MkdirAll(base, 0o755); err == nil {
		dir, err = os.MkdirTemp(base, check+"-"+strconv.Itoa(os.Getpid())+"-")
	}
	if dir == "" {
		fmt.Fprintf(stderr, "[!] test scratch %s not writable; using the normal temp folder\n", base)
		return nil, "", cleanup
	}
	return []string{"TMPDIR=" + dir, "GOTMPDIR=" + dir}, dir, func() { _ = os.RemoveAll(dir) }
}

// runDirRE names a run folder prepareScratch made: <check>-<pid>-<random>.
var runDirRE = regexp.MustCompile(`^[A-Za-z0-9_.]+-(\d+)-\d+$`)

// sweepStaleTemp removes, directly under dir, go-build folders older than a
// day and, when dir holds run folders, those whose process is gone (or older
// than a day, against a reused pid). A live go build keeps writing into its folder and suites time
// out well within a day, so age never takes a folder in use.
func sweepStaleTemp(dir string, runDirs bool, now time.Time, alive func(int) bool) {
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
		old := now.Sub(info.ModTime()) > staleTempAge
		name := e.Name()
		stale := strings.HasPrefix(name, "go-build") && old
		if m := runDirRE.FindStringSubmatch(name); m != nil && runDirs {
			pid, _ := strconv.Atoi(m[1])
			stale = old || !alive(pid)
		}
		if stale {
			_ = os.RemoveAll(filepath.Join(dir, name))
		}
	}
}

// ssdTempDirs are the SSD temp folders a scratched run does not own: the
// system's per-user temp folder and the agents' scratch volume.
func ssdTempDirs() []string {
	var out []string
	if d, err := userTempDir(); err == nil && d != "" {
		out = append(out, d)
	}
	return append(out, "/Volumes/AgentScratch/tmp")
}

// userTempDir is macOS's per-user temp folder (/var/folders/…/T), where a
// template-less mktemp and Go's os.TempDir land when TMPDIR is unset.
func userTempDir() (string, error) {
	if runtime.GOOS != "darwin" {
		return "", nil
	}
	out, err := exec.Command("getconf", "DARWIN_USER_TEMP_DIR").Output()
	return strings.TrimRight(strings.TrimSpace(string(out)), "/"), err
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

// reportScratchLeaks warns about temp files a scratched run left on the SSD
// (#468). It measures, never fails: the run's outcome stands.
func reportScratchLeaks(w io.Writer, check string, byPkg map[string][]string, unattributed int) {
	if len(byPkg) == 0 && unattributed == 0 {
		return
	}
	n := 0
	pkgs := make([]string, 0, len(byPkg))
	for pkg, ps := range byPkg {
		pkgs = append(pkgs, pkg)
		n += len(ps)
	}
	sort.Strings(pkgs)
	fmt.Fprintf(w, "[!] local check %s: SSD temp folders gained %d t.TempDir folder(s) from this worktree's tests despite the RAM disk; unattributed: %d (other processes write there too)\n", check, n, unattributed)
	for _, pkg := range pkgs {
		ps := byPkg[pkg]
		fmt.Fprintf(w, "    %s: %d, e.g. %s\n", pkg, len(ps), ps[0])
	}
}
