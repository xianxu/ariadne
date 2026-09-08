package gitx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// trunkFixture builds a repo with a local bare origin, seeds `queue.md` on main,
// and pushes. Returns (repo, origin).
//
// ARCH-MOCK: git is the external binary and a real throwaway repo with a real
// bare origin is its portable stateful fake. A function-call mock cannot produce
// a genuine non-fast-forward rejection, which is the behavior that matters most
// in this file.
func trunkFixture(t *testing.T, seed string) (string, string) {
	t.Helper()
	repo := testfix.Repo(t, testfix.InitialCommit())
	origin := filepath.Join(t.TempDir(), "origin.git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	testfix.Git(t, repo, "remote", "add", "origin", origin)
	if err := os.WriteFile(filepath.Join(repo, "queue.md"), []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, repo, "add", "queue.md")
	testfix.Git(t, repo, "commit", "-q", "-m", "seed")
	testfix.Git(t, repo, "push", "-q", "-u", "origin", "main")
	return repo, origin
}

// showTrunk reads a path out of the bare origin's main.
func showTrunk(t *testing.T, origin, path string) string {
	t.Helper()
	return testfix.Capture(t, origin, "show", "main:"+path)
}

// TestNewTrunkFile_RefusesEmptyDir is written BEFORE the feature it guards.
//
// The dangerous failure mode for this type is silent operation on the process
// cwd: gitx's other git shim (`run`, window.go) carries no Dir, so a TrunkFile
// that forgot to pass one would fetch from — and push `<commit>:main` to — the
// real ariadne origin during `go test`. A guard makes that impossible rather
// than merely avoided by convention.
func TestNewTrunkFile_RefusesEmptyDir(t *testing.T) {
	if _, err := NewTrunkFile("", "origin", "main"); err == nil {
		t.Fatal("empty dir must be refused — otherwise git runs against the process cwd")
	}
}

// TestTrunkFile_ReadsFromTrunkNotWorktree pins the whole point of the type: the
// trunk is the base, never the checkout.
//
// Deliberately does NOT testfix.Chdir(): `dir` carries the scoping, so a
// cwd-dependent implementation fails here instead of passing in CI while
// destroying a developer's checkout.
func TestTrunkFile_ReadsFromTrunkNotWorktree(t *testing.T) {
	repo, _ := trunkFixture(t, "from main\n")
	testfix.Git(t, repo, "checkout", "-q", "-b", "feature")
	if err := os.WriteFile(filepath.Join(repo, "queue.md"), []byte("from branch\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tf, err := NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	got, err := tf.Read("queue.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "from main\n" {
		t.Errorf("Read = %q, want the TRUNK copy %q", got, "from main\n")
	}
}

// A path absent from the trunk is not an error — it is an empty file. Without
// this the first-ever write to a new queue would have to special-case a git
// error string, which is the kind of thing that breaks on a git upgrade.
func TestTrunkFile_ReadMissingPathIsEmptyNotError(t *testing.T) {
	repo, _ := trunkFixture(t, "x\n")
	tf, err := NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	got, err := tf.Read("nope.md")
	if err != nil {
		t.Fatalf("missing path must not error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Read of a missing path = %q, want empty", got)
	}
}

// runGitIn is the seam the capability audit added. These assert the three
// parameters gitx's existing `run` shim does NOT carry, each of which the
// plumbing needs: Dir, Env, and stderr on failure.
func TestRunGitIn_CarriesDirEnvAndStderr(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit())

	// Dir: rev-parse resolves against the named repo, not the process cwd.
	out, err := runGitIn(repo, nil, "rev-parse", "--show-toplevel")
	if err != nil {
		t.Fatalf("Dir: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(string(out)); !strings.HasSuffix(got, filepath.Base(repo)) {
		t.Errorf("Dir not honored: toplevel = %q, want it under %q", got, repo)
	}

	// Env: GIT_INDEX_FILE reaches the child.
	idx := filepath.Join(t.TempDir(), "theindex")
	if out, err := runGitIn(repo, []string{"GIT_INDEX_FILE=" + idx}, "read-tree", "HEAD"); err != nil {
		t.Fatalf("Env: %v\n%s", err, out)
	}
	if _, err := os.Stat(idx); err != nil {
		t.Errorf("GIT_INDEX_FILE did not reach the child: %v", err)
	}

	// stderr: git's own message survives a failure. `run` (.Output()) drops it,
	// and three call sites render it — the retry exhaustion message, the offline
	// warning, and the refused-intent report.
	out, err = runGitIn(repo, nil, "cat-file", "blob", "main:definitely-absent")
	if err == nil {
		t.Fatal("expected failure")
	}
	if !strings.Contains(string(out), "definitely-absent") {
		t.Errorf("stderr dropped — got %q, want git's own message", out)
	}
}
