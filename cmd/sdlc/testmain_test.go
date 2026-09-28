package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/repolock"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Test hermeticity guard (#149/#165). cmd/sdlc tests run with cwd inside the REAL
// ariadne repo, and sdlc code resolves "the repo" from cwd — so a test that fails
// to isolate (no chdir into a temp git repo) can grab the real .git/sdlc.lock
// (#149) or mutate the real working tree (#165, which corrupted `main` this
// session). TestMain snapshots the real repo before and after the whole package
// run and FAILS a passing run that left durable damage — the backstop against both
// classes. The individual fix is isolating offenders into a t.TempDir() git repo.

// repoSnapshot captures the parts of the REAL repo a non-hermetic test could
// durably change. Plain data; snapshotDiff is the pure decision over two of these.
type repoSnapshot struct {
	head      string          // git rev-parse HEAD
	branch    string          // git branch --show-current
	porcelain map[string]bool // set of `git status --porcelain` lines
	lockOwned bool            // .git/sdlc.lock present and possibly this run's (lockMayBeOurs)
	resolved  bool            // false ⇒ real repo couldn't be resolved (guard skips)
}

// snapshotDiff reports what CHANGED from before→after, NEW mutations only: a
// porcelain line present in BOTH is pre-existing (e.g. the session's untracked
// notes) and is ignored. Empty ⇒ hermetic. Pure.
//
// This is a NET-STATE check across the whole run: a test that dirties then cleanly
// reverts the real repo within m.Run() is intentionally NOT flagged — net durable
// damage is the signal that matters (and is what the #165 incident left behind).
// The lockFile check catches a LEAKED lock (left behind), not #149's transient
// lock contention (released by the time m.Run() returns) — that's fixed by
// isolating the offenders so they never touch the real lock at all.
func snapshotDiff(before, after repoSnapshot) []string {
	if !before.resolved || !after.resolved {
		return nil // not in a resolvable git repo (e.g. CI tarball) — skip the guard
	}
	var out []string
	if before.head != after.head {
		out = append(out, fmt.Sprintf("HEAD moved: %s → %s", abbrevSHA(before.head), abbrevSHA(after.head)))
	}
	if before.branch != after.branch {
		out = append(out, fmt.Sprintf("branch switched: %q → %q", before.branch, after.branch))
	}
	var fresh []string
	for line := range after.porcelain {
		if !before.porcelain[line] {
			fresh = append(fresh, line)
		}
	}
	sort.Strings(fresh)
	for _, p := range fresh {
		out = append(out, "new working-tree change: "+p)
	}
	if !before.lockOwned && after.lockOwned {
		out = append(out, "leaked .git/sdlc.lock (acquired and not released)")
	}
	return out
}

// guardVerdict decides the process exit code from the test run's own code and the
// before/after snapshots. A PASSING run (code 0) that mutated the real repo FAILS
// (1) — the backstop. A run that ALREADY failed is not overridden (don't mask the
// real failure); the caller still surfaces the mutation so it isn't lost. Pure.
func guardVerdict(before, after repoSnapshot, code int) (exit int, mutations []string) {
	mutations = snapshotDiff(before, after)
	if code == 0 && len(mutations) > 0 {
		return 1, mutations
	}
	return code, mutations
}

// lockMayBeOurs decides whether a present .git/sdlc.lock could have been leaked
// by this test process. The lock lives in the git common dir, which every linked
// worktree (every slot) shares, so another slot's running sdlc verb holding it
// is not a leak (#253). It is ours when this process holds it (an in-process
// command, #149) or when its holder is gone (a test's child sdlc, or a crash).
// Unreadable metadata means an acquirer is still initializing it: a live
// process, so not counted. Pure.
func lockMayBeOurs(m repolock.Metadata, metaErr error, self int, host string, alive func(int) bool) bool {
	if metaErr != nil || m.Hostname != host {
		return false
	}
	return m.PID == self || !alive(m.PID)
}

// realRepoRoot resolves the real repo top-level from the INITIAL cwd — called by
// TestMain before any test os.Chdir, so a test's stray chdir can't move it.
func realRepoRoot() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// readSnapshot is the thin git IO — always queried with `-C root` so a test's
// chdir can't fool it into reading a temp repo.
func readSnapshot(root string) repoSnapshot {
	if root == "" {
		return repoSnapshot{}
	}
	git := func(args ...string) string {
		out, err := exec.Command("git", append([]string{"-C", root}, args...)...).Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	head := git("rev-parse", "HEAD")
	porc := map[string]bool{}
	for _, l := range strings.Split(git("status", "--porcelain"), "\n") {
		if strings.TrimSpace(l) != "" {
			porc[l] = true
		}
	}
	// Resolve the lock via --git-common-dir (mirrors repoLockGitCommonDir), not a
	// hardcoded <root>/.git — so leaked-lock detection also works in a linked
	// worktree, where <root>/.git is a file and the shared lock lives elsewhere.
	common := git("rev-parse", "--git-common-dir")
	if common == "" {
		common = ".git"
	}
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}
	lockDir := filepath.Join(common, "sdlc.lock")
	lockOwned := false
	if _, err := os.Stat(lockDir); err == nil {
		host, _ := os.Hostname()
		meta, metaErr := repolock.ReadMetadata(lockDir)
		lockOwned = lockMayBeOurs(meta, metaErr, os.Getpid(), host, processAlive)
	}
	return repoSnapshot{
		head:      head,
		branch:    git("branch", "--show-current"),
		porcelain: porc,
		lockOwned: lockOwned,
		resolved:  head != "",
	}
}

func TestMain(m *testing.M) {
	testfix.PreferRealGit()
	root := realRepoRoot()
	before := readSnapshot(root)
	code := m.Run()
	after := readSnapshot(root)
	exit, mutations := guardVerdict(before, after, code)
	if len(mutations) > 0 {
		fmt.Fprintf(os.Stderr, "\n*** cmd/sdlc TEST HERMETICITY GUARD (#149/#165) ***\n")
		fmt.Fprintf(os.Stderr, "a test durably changed the REAL repo (%s):\n", root)
		for _, mu := range mutations {
			fmt.Fprintf(os.Stderr, "  - %s\n", mu)
		}
		if code == 0 {
			fmt.Fprintf(os.Stderr, "FAILING the run: a hermetic test must not touch the real repo — "+
				"isolate it into a t.TempDir() git repo (chdir in, mirror closereview_test.go).\n")
		} else {
			fmt.Fprintf(os.Stderr, "(the run already failed; not overriding its exit — but fix the mutation above too)\n")
		}
	}
	os.Exit(exit)
}
