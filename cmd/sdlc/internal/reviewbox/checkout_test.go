package reviewbox

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// sourceRepo is a repo with two commits (H1, H2), an uncommitted tracked edit
// and an untracked file — the operator's live state a reviewer must not touch.
func sourceRepo(t *testing.T) (dir, h1, h2 string) {
	t.Helper()
	dir = testfix.Repo(t, testfix.InitialCommit())
	write(t, filepath.Join(dir, "a.txt"), "one\n")
	testfix.Git(t, dir, "add", "a.txt")
	testfix.Git(t, dir, "commit", "-q", "-m", "one")
	h1 = strings.TrimSpace(testfix.Capture(t, dir, "rev-parse", "HEAD"))
	write(t, filepath.Join(dir, "a.txt"), "two\n")
	testfix.Git(t, dir, "commit", "-q", "-am", "two")
	h2 = strings.TrimSpace(testfix.Capture(t, dir, "rev-parse", "HEAD"))
	write(t, filepath.Join(dir, "README"), "operator wip\n")
	write(t, filepath.Join(dir, "untracked.txt"), "operator scratch\n")
	return dir, h1, h2
}

// snapshot is every observable piece of the source a reviewer could damage:
// tracked and untracked file bytes, refs, config, index.
func snapshot(t *testing.T, dir string) string {
	t.Helper()
	var parts []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		if d.IsDir() {
			if rel == filepath.Join(".git", "objects") || rel == filepath.Join(".git", "logs") {
				return filepath.SkipDir // objects only grow by content address; logs mirror refs
			}
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		parts = append(parts, rel+"\x00"+string(data))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(parts)
	return strings.Join(parts, "\x01") + testfix.Capture(t, dir, "for-each-ref")
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// reviewerGit runs git in the checkout the way a reviewer would; it may fail.
func reviewerGit(dir string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-c", "user.email=r@r", "-c", "user.name=r", "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	return cmd.Run()
}

func TestCheckoutMutationLeavesSourceUntouched(t *testing.T) {
	src, _, h2 := sourceRepo(t)
	before := snapshot(t, src)
	co, err := Create(t.TempDir(), src, h2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer co.Close()

	write(t, filepath.Join(co.Dir, "a.txt"), "reverted fix\n")
	write(t, filepath.Join(co.Dir, "residue.bin"), "built\n")
	for _, args := range [][]string{
		{"commit", "-q", "-am", "mutation"},
		{"config", "x.y", "z"},
		{"branch", "evil"},
		{"checkout", "-q", h2, "--", "."},
	} {
		if err := reviewerGit(co.Dir, args...); err != nil {
			t.Fatalf("reviewer git %v: %v", args, err)
		}
	}
	if err := reviewerGit(co.Dir, "push", "origin", "HEAD:refs/heads/pushed"); err == nil {
		t.Fatal("push from the checkout succeeded; origin should be removed")
	}
	if after := snapshot(t, src); after != before {
		t.Fatal("reviewer activity in the checkout changed the source checkout")
	}
}

func TestCheckoutIsPristinePerRound(t *testing.T) {
	src, h1, h2 := sourceRepo(t)
	root := t.TempDir()
	one, err := Create(root, src, h1, nil)
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(one.Dir, "residue"), "round one\n")
	if got := readFile(t, filepath.Join(one.Dir, "a.txt")); got != "one\n" {
		t.Fatalf("round one a.txt = %q, want Head H1's", got)
	}
	if err := one.Close(); err != nil {
		t.Fatal(err)
	}

	two, err := Create(root, src, h2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer two.Close()
	if two.Dir == one.Dir {
		t.Fatal("round two reused round one's directory")
	}
	if _, err := os.Stat(filepath.Join(two.Dir, "residue")); !os.IsNotExist(err) {
		t.Fatal("round one's residue reached round two")
	}
	if got := strings.TrimSpace(testfix.Capture(t, two.Dir, "rev-parse", "HEAD")); got != h2 {
		t.Fatalf("round two HEAD = %s, want %s", got, h2)
	}
	if st := testfix.Capture(t, two.Dir, "status", "--porcelain"); st != "" {
		t.Fatalf("round two not pristine:\n%s", st)
	}
}

func TestCheckoutCloseRemovesAndIsIdempotent(t *testing.T) {
	src, _, h2 := sourceRepo(t)
	root := t.TempDir()
	co, err := Create(root, src, h2, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := co.Close(); err != nil {
		t.Fatal(err)
	}
	if err := co.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if entries, _ := os.ReadDir(root); len(entries) != 0 {
		t.Fatalf("Close left %d entries under root", len(entries))
	}
}

func TestSweepReclaimsDeadOwnerAndSparesLiveOne(t *testing.T) {
	src, _, h2 := sourceRepo(t)
	root := t.TempDir()
	live, err := Create(root, src, h2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()

	// A crashed owner: lock file and checkout present, nobody holding the lock.
	// Its tree has a read-only directory, as a Go module cache would.
	dead := filepath.Join(root, "deadbeef")
	if err := os.MkdirAll(filepath.Join(dead, "ro", "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dead, "ro", "pkg", "f"), "x")
	if err := os.Chmod(filepath.Join(dead, "ro"), 0o555); err != nil {
		t.Fatal(err)
	}
	write(t, dead+lockSuffix, "")
	// Debris: a checkout whose lock is gone, and a create killed before rename.
	if err := os.MkdirAll(filepath.Join(root, "orphan"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "cafe"+tmpLockSuffix), "")
	// An entry the sweep cannot handle must not stop it or a later Create.
	if err := os.MkdirAll(filepath.Join(root, "bad"+lockSuffix), 0o755); err != nil {
		t.Fatal(err)
	}

	removed, problems := Sweep(root)
	if removed != 3 || len(problems) != 1 || !strings.Contains(problems[0].Error(), "bad"+lockSuffix) {
		t.Fatalf("Sweep removed %d, problems %v; want 3 removed and one problem naming bad.lock", removed, problems)
	}
	for _, gone := range []string{dead, dead + lockSuffix, filepath.Join(root, "orphan"), filepath.Join(root, "cafe"+tmpLockSuffix)} {
		if _, err := os.Stat(gone); !os.IsNotExist(err) {
			t.Errorf("%s survived the sweep", gone)
		}
	}
	if _, err := os.Stat(filepath.Join(live.Dir, "a.txt")); err != nil {
		t.Fatalf("the sweep removed a live review's checkout: %v", err)
	}
	next, err := Create(root, src, h2, nil)
	if err != nil {
		t.Fatalf("Create after a sweep problem: %v", err)
	}
	next.Close()
}

// A checkout Close cannot remove keeps its lock file, so the next sweep finds
// an unheld lock and finishes the job instead of leaving lockless debris.
func TestCloseFailureLeavesLockForSweep(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("uses chflags uchg to make a file unremovable")
	}
	src, _, h2 := sourceRepo(t)
	root := t.TempDir()
	co, err := Create(root, src, h2, nil)
	if err != nil {
		t.Fatal(err)
	}
	pinned := filepath.Join(co.Dir, "a.txt")
	if out, err := exec.Command("chflags", "uchg", pinned).CombinedOutput(); err != nil {
		t.Skipf("chflags unavailable: %v %s", err, out)
	}
	t.Cleanup(func() { exec.Command("chflags", "nouchg", pinned).Run() })
	if err := co.Close(); err == nil {
		t.Fatal("Close reported success with an unremovable file")
	}
	if _, err := os.Stat(co.Dir + lockSuffix); err != nil {
		t.Fatalf("Close dropped the lock file of a checkout it could not remove: %v", err)
	}
	exec.Command("chflags", "nouchg", pinned).Run()
	if removed, problems := Sweep(root); removed != 1 || len(problems) != 0 {
		t.Fatalf("Sweep removed %d (%v); want the leftover reclaimed", removed, problems)
	}
	if entries, _ := os.ReadDir(root); len(entries) != 0 {
		t.Fatalf("%d entries remain", len(entries))
	}
}

func TestResetRestoresPristineCheckout(t *testing.T) {
	src, _, h2 := sourceRepo(t)
	co, err := Create(t.TempDir(), src, h2, []string{"README"})
	if err != nil {
		t.Fatal(err)
	}
	defer co.Close()
	write(t, filepath.Join(co.Dir, "residue"), "attempt one")
	write(t, filepath.Join(co.Dir, "README"), "attempt one edit")
	if err := reviewerGit(co.Dir, "commit", "-q", "-am", "attempt one"); err != nil {
		t.Fatal(err)
	}
	if err := co.Reset(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(co.Dir, "residue")); !os.IsNotExist(err) {
		t.Error("residue survived Reset")
	}
	if got := strings.TrimSpace(testfix.Capture(t, co.Dir, "rev-parse", "HEAD")); got != h2 {
		t.Errorf("HEAD after Reset = %s, want %s", got, h2)
	}
	if got := readFile(t, filepath.Join(co.Dir, "README")); got != "operator wip\n" {
		t.Errorf("overlay after Reset = %q, want the live text", got)
	}
}

func TestCheckoutOverlaysLiveFiles(t *testing.T) {
	src, _, h2 := sourceRepo(t)
	co, err := Create(t.TempDir(), src, h2, []string{"README", "missing/plan.md"})
	if err != nil {
		t.Fatal(err)
	}
	defer co.Close()
	if got := readFile(t, filepath.Join(co.Dir, "README")); got != "operator wip\n" {
		t.Fatalf("overlay README = %q, want the live text", got)
	}
}

func TestCreateFailureLeavesNothing(t *testing.T) {
	src, _, _ := sourceRepo(t)
	root := t.TempDir()
	if _, err := Create(root, src, strings.Repeat("0", 40), nil); err == nil {
		t.Fatal("Create at a missing commit succeeded")
	}
	if _, err := Create(root, src, "HEAD", []string{"../escape"}); err == nil {
		t.Fatal("Create accepted an overlay outside the repository")
	}
	if entries, _ := os.ReadDir(root); len(entries) != 0 {
		t.Fatalf("failed Create left %d entries", len(entries))
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
