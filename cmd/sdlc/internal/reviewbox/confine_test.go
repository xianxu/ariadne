package reviewbox

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestDiffReportsNewRepositoriesAndHomeEntries(t *testing.T) {
	before := OutsideSnapshot{
		Repos: map[string]bool{"/h": false, "/h/w": false, "/": false},
		Home:  "/h", Names: map[string]bool{"workspace": true},
	}
	same := OutsideSnapshot{Repos: before.Repos, Home: "/h", Names: map[string]bool{"workspace": true}}
	if d := Diff(before, same); len(d) != 0 {
		t.Fatalf("unchanged snapshot reported %v", d)
	}
	after := OutsideSnapshot{
		Repos: map[string]bool{"/h": true, "/h/w": false, "/": false},
		Home:  "/h", Names: map[string]bool{"workspace": true, "scripts": true},
	}
	want := []string{"new entry /h/scripts", "new repository /h/.git"}
	if d := Diff(before, after); !reflect.DeepEqual(d, want) {
		t.Fatalf("Diff = %v, want %v", d, want)
	}
}

func TestTakeSnapshotIgnoresHiddenHomeEntries(t *testing.T) {
	home := t.TempDir()
	before := snap(t, home)
	write(t, filepath.Join(home, ".zsh_history"), "x")
	if d := Diff(before, snap(t, home)); len(d) != 0 {
		t.Fatalf("a new dotfile was reported: %v", d)
	}
	if err := os.Mkdir(filepath.Join(home, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if d := Diff(before, snap(t, home)); len(d) != 1 || !strings.Contains(d[0], "new repository") {
		t.Fatalf("a new home .git was not reported: %v", d)
	}
}

func snap(t *testing.T, home string) OutsideSnapshot {
	t.Helper()
	s, err := TakeSnapshot([]string{home}, home)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestTakeSnapshotFailsWhenHomeCannotBeListed(t *testing.T) {
	if _, err := TakeSnapshot(nil, filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("an unlistable home passed silently")
	}
	if _, err := TakeSnapshot(nil, ""); err == nil {
		t.Fatal("an unknown home passed silently")
	}
}

// shimEnv renders the shim over a fake real git that records its argv, and
// returns a runner for it.
func shimEnv(t *testing.T, allowed ...string) func(dir string, args ...string) (string, error) {
	t.Helper()
	bin := t.TempDir()
	log := filepath.Join(bin, "ran")
	real := filepath.Join(bin, "real-git")
	write(t, real, "#!/bin/sh\necho \"$*\" >> "+shQuote(log)+"\n")
	if err := os.Chmod(real, 0o755); err != nil {
		t.Fatal(err)
	}
	shim := filepath.Join(bin, "git")
	write(t, shim, GitShim(real, allowed))
	return func(dir string, args ...string) (string, error) {
		os.Remove(log)
		cmd := exec.Command("sh", append([]string{shim}, args...)...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		ran, _ := os.ReadFile(log)
		return string(out) + string(ran), err
	}
}

func TestGitShimRefusesRepositoriesOutside(t *testing.T) {
	root, err := canonicalDir(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	outside, err := canonicalDir(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(outside, "src")
	if err := os.Mkdir(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "a"), 0o755); err != nil {
		t.Fatal(err)
	}
	git := shimEnv(t, root)
	refused := [][]string{
		{outside, "init"},
		{outside, "-c", "a=b", "init", "-q"},
		{root, "-C", outside, "init"},
		{root, "init", filepath.Join(outside, "new")},
		{root, "init", "--separate-git-dir=" + filepath.Join(outside, "g")},
		{root, "clone", src, filepath.Join(outside, "dst")},
		{outside, "clone", src},
		// Relative operands and git dirs are resolved too (#339 close review).
		{filepath.Join(root, "a"), "init", relTo(t, filepath.Join(root, "a"), filepath.Join(outside, "rel"))},
		{root, "init", "sub/new/../../../escape"},
		{root, "--git-dir=" + filepath.Join(outside, "gd"), "init"},
		{root, "--git-dir", filepath.Join(outside, "gd"), "init"},
		{root, "--work-tree=" + outside, "init"},
		{root, "init", "--separate-git-dir=../gd"},
	}
	for _, c := range refused {
		out, err := git(c[0], c[1:]...)
		if err == nil || !strings.Contains(out, "#339") {
			t.Errorf("git %v in %s: want refused, got err=%v out=%q", c[1:], c[0], err, out)
		}
	}
	allowed := [][]string{
		{root, "init", "-q"},
		{root, "init", filepath.Join(root, "sub", "new")},
		{root, "clone", src, filepath.Join(root, "dst")},
		{outside, "status"},
		{outside, "-C", root, "init"},
		{root, "init", "-b", "main", "sub/x"},
		{filepath.Join(root, "a"), "init", "../b"},
		{root, "clone", "--depth=1", src, "dst2"},
	}
	for _, c := range allowed {
		out, err := git(c[0], c[1:]...)
		if err != nil || !strings.Contains(out, strings.Join(c[1:], " ")) {
			t.Errorf("git %v in %s: want passed to real git, got err=%v out=%q", c[1:], c[0], err, out)
		}
	}
}

func relTo(t *testing.T, from, to string) string {
	t.Helper()
	r, err := filepath.Rel(from, to)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestGitShimChecksGitDirEnv(t *testing.T) {
	root, _ := canonicalDir(t.TempDir())
	outside, _ := canonicalDir(t.TempDir())
	git := shimEnv(t, root)
	t.Setenv("GIT_DIR", filepath.Join(outside, "gd"))
	if out, err := git(root, "init"); err == nil || !strings.Contains(out, "#339") {
		t.Fatalf("GIT_DIR outside: err=%v out=%q", err, out)
	}
}

func TestConfineInsideCheckout(t *testing.T) {
	src, _, h2 := sourceRepo(t)
	co, err := CreateContext(t.Context(), t.TempDir(), src, h2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer co.Close()
	c, err := Confine(t.Context(), co.Dir, true)
	if err != nil {
		t.Fatal(err)
	}
	dir, _ := canonicalDir(co.Dir)
	if c.Base != filepath.Join(dir, confineBaseName) {
		t.Fatalf("base %s is not inside the checkout %s", c.Base, dir)
	}
	write(t, filepath.Join(c.Tmp(), "scratch"), "x")
	if st := testfix.Capture(t, co.Dir, "status", "--porcelain"); strings.TrimSpace(st) != "" {
		t.Fatalf("the base shows in the checkout's status: %q", st)
	}
	env := strings.Join(c.Env(), "\n")
	for _, want := range []string{"TMPDIR=" + c.Tmp(), "CLAUDE_CODE_TMPDIR=" + c.Tmp(), "GIT_CEILING_DIRECTORIES=" + filepath.Dir(dir) + ":" + c.Base} {
		if !strings.Contains(env, want) {
			t.Errorf("env lacks %s:\n%s", want, env)
		}
	}
	// Claude Code's Bash rebuilds PATH from the shell snapshot; the shims reach
	// it only through the env file it sources before each command.
	out, err := exec.Command("sh", "-c", `PATH=/usr/bin:/bin; . "$1"; command -v git`, "sh", c.envFile()).Output()
	if err != nil || strings.TrimSpace(string(out)) != filepath.Join(c.BinDir(), "git") {
		t.Errorf("after the env file, git is %q (%v), want the shim", out, err)
	}
	if !strings.Contains(env, "CLAUDE_ENV_FILE="+c.envFile()) {
		t.Errorf("env lacks CLAUDE_ENV_FILE:\n%s", env)
	}
	// A retry resets the checkout, which deletes the base; Prepare restores it.
	if err := co.ResetContext(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := c.Prepare(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(c.BinDir(), "git")); err != nil {
		t.Fatalf("shim not restored: %v", err)
	}
}

func TestConfineFailsWhenTempDirCannotBeCreated(t *testing.T) {
	parent := t.TempDir()
	if err := os.Chmod(parent, 0o500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(parent, 0o700)
	t.Setenv("TMPDIR", parent)
	c, err := Confine(t.Context(), t.TempDir(), false)
	if err == nil {
		c.Close()
		t.Fatal("Confine succeeded with an unwritable temp parent")
	}
	if !strings.Contains(err.Error(), "temp dir") {
		t.Fatalf("error does not name the temp dir: %v", err)
	}
	if entries, _ := os.ReadDir(parent); len(entries) != 0 {
		t.Fatalf("Confine left %d entries behind", len(entries))
	}
}

func TestConfineCheckCatchesOutsideWrites(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	work := filepath.Join(home, "w")
	if err := os.MkdirAll(filepath.Join(work, "repo"), 0o755); err != nil {
		t.Fatal(err)
	}
	c, err := Confine(t.Context(), filepath.Join(work, "repo"), false)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := c.Check(); err != nil {
		t.Fatalf("clean run reported: %v", err)
	}
	write(t, filepath.Join(c.Tmp(), "fine"), "x")
	write(t, filepath.Join(c.Root, "fine"), "x")
	if err := c.Check(); err != nil {
		t.Fatalf("writes inside were reported: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(home, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(work, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	err = c.Check()
	if err == nil || !strings.Contains(err.Error(), filepath.Join(c.homeForTest(), "scripts")) || !strings.Contains(err.Error(), "new repository") {
		t.Fatalf("Check = %v, want both outside writes named", err)
	}
	base := c.Base
	c.Close()
	if _, err := os.Stat(base); !os.IsNotExist(err) {
		t.Fatalf("Close left the base: %v", err)
	}
}

func (c *Confinement) homeForTest() string { return c.before.Home }

func TestMktempShimNamesATemplateInTMPDIR(t *testing.T) {
	bin := t.TempDir()
	real := filepath.Join(bin, "real-mktemp")
	write(t, real, "#!/bin/sh\necho \"$*\"\n")
	if err := os.Chmod(real, 0o755); err != nil {
		t.Fatal(err)
	}
	shim := filepath.Join(bin, "mktemp")
	write(t, shim, MktempShim(real))
	cases := map[string][]string{
		"/r/tmp.XXXXXXXXXX":       nil,
		"-d /r/tmp.XXXXXXXXXX":    {"-d"},
		"-d -q /r/foo.XXXXXXXXXX": {"-d", "-q", "-t", "foo"},
		"/p/tmp.XXXXXXXXXX":       {"-p", "/p"},
		"-d x.XXXX":               {"-d", "x.XXXX"}, // a template runs unchanged
		"-p /p -d rel.XXXX":       {"-p", "/p", "-d", "rel.XXXX"},
	}
	for want, args := range cases {
		cmd := exec.Command("sh", append([]string{shim}, args...)...)
		cmd.Env = append(os.Environ(), "TMPDIR=/r")
		out, err := cmd.Output()
		if err != nil || strings.TrimSpace(string(out)) != want {
			t.Errorf("mktemp %v ran %q (err %v), want %q", args, strings.TrimSpace(string(out)), err, want)
		}
	}
}

func TestClaudeEnvFileSourcesAnInheritedOne(t *testing.T) {
	dir := t.TempDir()
	inherited := filepath.Join(dir, "parent.sh")
	write(t, inherited, "PARENT=yes; TMPDIR=/inherited; export PARENT TMPDIR\n")
	ours := filepath.Join(dir, "env.sh")
	write(t, ours, claudeEnvFile("/shim bin", inherited, []string{"TMPDIR=/ours", "CLAUDE_ENV_FILE=" + ours}))
	out, err := exec.Command("sh", "-c", `PATH=/usr/bin:/bin; . "$1"; echo "$PARENT $TMPDIR $PATH"`, "sh", ours).Output()
	if err != nil || strings.TrimSpace(string(out)) != "yes /ours /shim bin:/usr/bin:/bin" {
		t.Fatalf("sourced env = %q (%v)", out, err)
	}
}
