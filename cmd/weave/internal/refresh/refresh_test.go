package refresh

import (
	"bytes"
	"context"
	"errors"
	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1", "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com", "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com")
	v, e := c.CombinedOutput()
	if e != nil {
		t.Fatalf("git %v: %v %s", args, e, v)
	}
	return strings.TrimSuffix(string(v), "\n")
}
func write(t *testing.T, p, s string) {
	t.Helper()
	if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(p, []byte(s), 0644); e != nil {
		t.Fatal(e)
	}
}

type fixture struct{ root, seed, remote string }

func repo(t *testing.T, base, name string) fixture {
	t.Helper()
	var err error
	base, err = filepath.EvalSymlinks(base)
	if err != nil {
		t.Fatal(err)
	}
	f := fixture{filepath.Join(base, name), filepath.Join(base, name+"-seed"), filepath.Join(base, name+".git")}
	git(t, base, "init", "--bare", "--initial-branch=main", f.remote)
	git(t, base, "clone", f.remote, f.seed)
	write(t, filepath.Join(f.seed, "construct/base.manifest"), "# fixture\n")
	write(t, filepath.Join(f.seed, "file"), "base\n")
	git(t, f.seed, "add", ".")
	git(t, f.seed, "commit", "-m", "base")
	git(t, f.seed, "push", "origin", "main")
	git(t, base, "clone", f.remote, f.root)
	git(t, f.root, "config", "user.name", "Test")
	git(t, f.root, "config", "user.email", "test@example.com")
	return f
}
func commit(t *testing.T, dir, p, s string) string {
	write(t, filepath.Join(dir, p), s)
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-m", "change")
	return git(t, dir, "rev-parse", "HEAD")
}
func remoteAdvance(t *testing.T, f fixture) string {
	s := commit(t, f.seed, "new", "remote\n")
	git(t, f.seed, "push", "origin", "main")
	return s
}

// gitModel is a stateful real-Git model. The repository owns graph, refs, index,
// files and operation state; hooks inject external effects at deterministic IO
// boundaries. It shares Git's semantics rather than duplicating its algorithms.
type gitModel struct {
	base  acquire.GitRunner
	after func(string, []string, string, error) (string, error)
}

func (m *gitModel) Run(c context.Context, d string, a ...string) (string, error) {
	s, e := m.base.Run(c, d, a...)
	if m.after != nil {
		return m.after(d, a, s, e)
	}
	return s, e
}
func (m *gitModel) RunOwned(c context.Context, d, p string, a ...string) (string, error) {
	return m.base.RunOwned(c, d, p, a...)
}
func client() acquire.Client { return acquire.Client{Git: acquire.ExecGit{Raw: true}} }
func run(t *testing.T, f fixture, c acquire.Client, rebase bool) (string, error, int) {
	t.Helper()
	var b bytes.Buffer
	n := 0
	e := Run(context.Background(), f.root, c, rebase, &b, func() error { n++; return nil })
	return b.String(), e, n
}
func TestGitConformance(t *testing.T) {
	for _, model := range []bool{false, true} {
		t.Run(map[bool]string{false: "live", true: "stateful"}[model], func(t *testing.T) {
			f := repo(t, t.TempDir(), "host")
			git(t, f.root, "switch", "-c", "main-slot1")
			target := remoteAdvance(t, f)
			c := client()
			if model {
				c.Git = &gitModel{base: c.Git}
			}
			_, e, n := run(t, f, c, false)
			if e != nil || n != 1 {
				t.Fatalf("%v compile=%d", e, n)
			}
			if git(t, f.root, "rev-parse", "HEAD") != target || git(t, f.root, "branch", "--show-current") != "main-slot1" {
				t.Fatal("wrong branch/target")
			}
			write(t, filepath.Join(f.root, "untracked"), "keep")
			_, e, n = run(t, f, c, false)
			if e == nil || n != 0 {
				t.Fatal("dirty state accepted")
			}

			if e := os.Remove(filepath.Join(f.root, "untracked")); e != nil {
				t.Fatal(e)
			}
			local := commit(t, f.root, "file", "local conflict\n")
			_, e, n = run(t, f, c, false)
			if e == nil || n != 0 || git(t, f.root, "rev-parse", "HEAD") != local {
				t.Fatal("ahead refusal mismatch")
			}
			commit(t, f.seed, "file", "remote conflict\n")
			git(t, f.seed, "push", "origin", "main")
			_, e, n = run(t, f, c, true)
			if e == nil || n != 0 {
				t.Fatal("conflict semantics mismatch")
			}
			if _, e := os.Stat(filepath.Join(f.root, ".git/rebase-merge")); e != nil {
				t.Fatalf("conflict state not preserved: %v", e)
			}
			_, e, n = run(t, f, c, true)
			if e == nil || n != 0 {
				t.Fatal("active operation accepted")
			}
		})
	}
}
func TestRefreshAheadRefused(t *testing.T) {
	f := repo(t, t.TempDir(), "host")
	before := commit(t, f.root, "local", "local")
	_, e, n := run(t, f, client(), false)
	if e == nil || n != 0 || git(t, f.root, "rev-parse", "HEAD") != before {
		t.Fatal("ahead accepted")
	}
}
func TestRefreshMissingCheckout(t *testing.T) {
	f := repo(t, t.TempDir(), "host")
	commit(t, f.root, "construct/deps", "substrate ../missing ../missing.git\n")
	_, e, n := run(t, f, client(), false)
	if e == nil || n != 0 {
		t.Fatal("missing accepted")
	}
}
func TestRefreshAllReadyBeforeFetch(t *testing.T) {
	base := t.TempDir()
	dep := repo(t, base, "dep")
	f := repo(t, base, "host")
	commit(t, f.seed, "construct/deps", "substrate ../dep\n")
	git(t, f.seed, "push", "origin", "main")
	git(t, f.root, "pull", "--ff-only")
	old := git(t, dep.root, "rev-parse", "HEAD")
	remoteAdvance(t, dep)
	write(t, filepath.Join(f.root, "dirty"), "keep")
	_, e, _ := run(t, f, client(), false)
	if e == nil || git(t, dep.root, "rev-parse", "HEAD") != old || git(t, dep.root, "rev-parse", "origin/main") != old {
		t.Fatal("later dirty subject failed all-ready gate")
	}
}
func TestCapturedTarget(t *testing.T) {
	f := repo(t, t.TempDir(), "host")
	old := git(t, f.root, "rev-parse", "HEAD")
	target := remoteAdvance(t, f)
	m := &gitModel{base: client().Git}
	m.after = func(d string, a []string, s string, e error) (string, error) {
		if strings.Join(a, " ") == "rev-parse --verify refs/remotes/origin/main^{commit}" {
			git(t, d, "update-ref", "refs/remotes/origin/main", old)
		}
		return s, e
	}
	_, e, _ := run(t, f, acquire.Client{Git: m}, false)
	if e != nil || git(t, f.root, "rev-parse", "HEAD") != target {
		t.Fatalf("target moved: %v", e)
	}
}
func TestChangedStartingState(t *testing.T) {
	f := repo(t, t.TempDir(), "host")
	remoteAdvance(t, f)
	m := &gitModel{base: client().Git}
	changed := false
	m.after = func(d string, a []string, s string, e error) (string, error) {
		if !changed && len(a) > 0 && a[0] == "fetch" {
			changed = true
			commit(t, d, "local", "racing")
		}
		return s, e
	}
	_, e, n := run(t, f, acquire.Client{Git: m}, false)
	if e == nil || n != 0 {
		t.Fatal("changed HEAD accepted")
	}
}
func TestUncertainApplication(t *testing.T) {
	f := repo(t, t.TempDir(), "host")
	target := remoteAdvance(t, f)
	m := &gitModel{base: client().Git}
	m.after = func(d string, a []string, s string, e error) (string, error) {
		if strings.Contains(strings.Join(a, " "), "merge --ff-only") && e == nil {
			return "", errors.New("lost response")
		}
		return s, e
	}
	out, e, n := run(t, f, acquire.Client{Git: m}, false)
	if e == nil || n != 0 || git(t, f.root, "rev-parse", "HEAD") != target || !strings.Contains(out, "uncertain") {
		t.Fatalf("%v %s", e, out)
	}
}
func TestDeclarationDrift(t *testing.T) {
	f := repo(t, t.TempDir(), "host")
	before := git(t, f.root, "rev-parse", "HEAD")
	commit(t, f.seed, "construct/deps", "substrate ../new\n")
	git(t, f.seed, "push", "origin", "main")
	_, e, n := run(t, f, client(), false)
	if e == nil || n != 0 || git(t, f.root, "rev-parse", "HEAD") != before {
		t.Fatal("graph drift accepted")
	}
}
func TestCompileRetry(t *testing.T) {
	f := repo(t, t.TempDir(), "host")
	target := remoteAdvance(t, f)
	e := Run(context.Background(), f.root, client(), false, &bytes.Buffer{}, func() error { return errors.New("compile fails") })
	if e == nil || git(t, f.root, "rev-parse", "HEAD") != target {
		t.Fatal("compile failure lost update")
	}
	_, e, n := run(t, f, client(), false)
	if e != nil || n != 1 {
		t.Fatalf("retry: %v", e)
	}
}
func TestRebaseConflictRetry(t *testing.T) {
	f := repo(t, t.TempDir(), "host")
	commit(t, f.root, "file", "local\n")
	commit(t, f.seed, "file", "remote\n")
	git(t, f.seed, "push", "origin", "main")
	_, e, n := run(t, f, client(), true)
	if e == nil || n != 0 {
		t.Fatal("conflict accepted")
	}
	_, e, n = run(t, f, client(), true)
	if e == nil || n != 0 {
		t.Fatal("active rebase accepted")
	}
	git(t, f.root, "rebase", "--abort")
	write(t, filepath.Join(f.root, "file"), "remote\n")
	git(t, f.root, "add", "file")
	git(t, f.root, "commit", "--amend", "--no-edit")
	_, e, n = run(t, f, client(), true)
	if e != nil || n != 1 {
		t.Fatalf("retry %v", e)
	}
}
func TestIgnoredCollision(t *testing.T) {
	f := repo(t, t.TempDir(), "host")
	commit(t, f.root, ".gitignore", "new\n")
	write(t, filepath.Join(f.root, "new"), "precious")
	remoteAdvance(t, f)
	_, e, n := run(t, f, client(), true)
	b, _ := os.ReadFile(filepath.Join(f.root, "new"))
	if e == nil || n != 0 || string(b) != "precious" {
		t.Fatalf("ignored file lost: %v %q", e, b)
	}
}
func TestPartialProgress(t *testing.T) {
	base := t.TempDir()
	dep := repo(t, base, "dep")
	f := repo(t, base, "host")
	commit(t, f.seed, "construct/deps", "substrate ../dep\n")
	git(t, f.seed, "push", "origin", "main")
	git(t, f.root, "pull", "--ff-only")
	target := remoteAdvance(t, dep)
	old := git(t, f.root, "rev-parse", "HEAD")
	remoteAdvance(t, f)
	m := &gitModel{base: client().Git}
	m.after = func(d string, a []string, s string, e error) (string, error) {
		if d == dep.root && strings.Contains(strings.Join(a, " "), "merge --ff-only") {
			write(t, filepath.Join(f.root, "dirty"), "preserve")
		}
		return s, e
	}
	out, e, n := run(t, f, acquire.Client{Git: m}, false)
	if e == nil || n != 0 || git(t, dep.root, "rev-parse", "HEAD") != target || git(t, f.root, "rev-parse", "HEAD") != old || !strings.Contains(out, "1 repository updates confirmed") {
		t.Fatalf("%v %s", e, out)
	}
}
func TestDataExcluded(t *testing.T) {
	base := t.TempDir()
	data := repo(t, base, "data")
	f := repo(t, base, "host")
	commit(t, f.seed, "construct/deps", "data ../data.git content\n")
	git(t, f.seed, "push", "origin", "main")
	git(t, f.root, "pull", "--ff-only")
	old := git(t, data.root, "rev-parse", "HEAD")
	remoteAdvance(t, data)
	remoteAdvance(t, f)
	_, e, n := run(t, f, client(), false)
	if e != nil || n != 1 || git(t, data.root, "rev-parse", "HEAD") != old || git(t, data.root, "rev-parse", "origin/main") != old {
		t.Fatalf("data refreshed %v", e)
	}
}
func TestChangedOrigin(t *testing.T) {
	f := repo(t, t.TempDir(), "host")
	remoteAdvance(t, f)
	m := &gitModel{base: client().Git}
	m.after = func(d string, a []string, s string, e error) (string, error) {
		if len(a) > 0 && a[0] == "fetch" {
			git(t, d, "remote", "set-url", "origin", f.seed)
		}
		return s, e
	}
	_, e, n := run(t, f, acquire.Client{Git: m}, false)
	if e == nil || n != 0 {
		t.Fatal("changed origin accepted")
	}
}
func TestObserveRefusals(t *testing.T) {
	for _, kind := range []string{"detached", "active", "symlink-deps", "unborn"} {
		t.Run(kind, func(t *testing.T) {
			f := repo(t, t.TempDir(), "host")
			switch kind {
			case "detached":
				git(t, f.root, "checkout", "--detach")
			case "active":
				write(t, filepath.Join(f.root, ".git/MERGE_HEAD"), git(t, f.root, "rev-parse", "HEAD")+"\n")
			case "symlink-deps":
				if e := os.Symlink("base.manifest", filepath.Join(f.root, "construct/deps")); e != nil {
					t.Fatal(e)
				}
				git(t, f.root, "add", ".")
				git(t, f.root, "commit", "-m", "symlink")
			case "unborn":
				git(t, f.root, "switch", "--orphan", "empty")
			}
			_, e := observe(context.Background(), client().Git, f.root)
			if e == nil {
				t.Fatalf("%s accepted", kind)
			}
		})
	}
}
func TestIgnoredRebaseIntermediateCollision(t *testing.T) {
	f := repo(t, t.TempDir(), "host")
	commit(t, f.root, "temporary", "committed then deleted")
	git(t, f.root, "rm", "temporary")
	write(t, filepath.Join(f.root, ".gitignore"), "temporary\n")
	git(t, f.root, "add", ".")
	git(t, f.root, "commit", "-m", "delete temporary")
	write(t, filepath.Join(f.root, "temporary"), "precious")
	remoteAdvance(t, f)
	before := git(t, f.root, "rev-parse", "HEAD")
	_, e, n := run(t, f, client(), true)
	if git(t, f.root, "rev-parse", "HEAD") != before {
		t.Fatal("intermediate collision was not refused before mutation")
	}
	b, _ := os.ReadFile(filepath.Join(f.root, "temporary"))
	if e == nil || n != 0 || string(b) != "precious" {
		t.Fatalf("intermediate replay can overwrite ignored data: %v %q", e, b)
	}
}
func TestGraphChangedDuringBaseline(t *testing.T) {
	base := t.TempDir()
	_ = repo(t, base, "dep")
	f := repo(t, base, "host")
	m := &gitModel{base: client().Git}
	changed := false
	var before string
	m.after = func(d string, a []string, s string, e error) (string, error) {
		if !changed && d == f.root && strings.Join(a, " ") == "rev-parse --show-toplevel" {
			changed = true
			before = commit(t, f.root, "construct/deps", "substrate ../dep\n")
			git(t, f.root, "push", "origin", "main")
		}
		return s, e
	}
	_, e, n := run(t, f, acquire.Client{Git: m}, false)
	if e == nil || n != 0 || git(t, f.root, "rev-parse", "HEAD") != before {
		t.Fatalf("baseline graph drift accepted %v", e)
	}
}
func TestWhitespacePath(t *testing.T) {
	base := t.TempDir()
	f := repo(t, base, "host")
	moved := filepath.Join(filepath.Dir(f.root), " host ")
	if e := os.Rename(f.root, moved); e != nil {
		t.Fatal(e)
	}
	f.root = moved
	target := remoteAdvance(t, f)
	_, e, n := run(t, f, client(), false)
	if e != nil || n != 1 || git(t, f.root, "rev-parse", "HEAD") != target {
		t.Fatalf("path whitespace lost: %v", e)
	}
}
