package workspace_test

import (
	"encoding/json"
	"fmt"
	"github.com/xianxu/ariadne/pkg/workspace"
	"github.com/xianxu/ariadne/pkg/workspace/workspacetest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

type realGit struct{ reads int }

func (r *realGit) GitInDir(dir string, args ...string) ([]byte, error) {
	r.reads++
	c := exec.Command("git", append([]string{"-C", dir}, args...)...)
	c.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_TERMINAL_PROMPT=0", "GIT_CONFIG_COUNT=1", "GIT_CONFIG_KEY_0=core.hooksPath", "GIT_CONFIG_VALUE_0=/dev/null")
	return c.CombinedOutput()
}
func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	r := &realGit{}
	b, e := r.GitInDir(dir, args...)
	if e != nil {
		t.Fatalf("git %v: %v %s", args, e, b)
	}
	return strings.TrimSuffix(string(b), "\n")
}

type fixture struct {
	primary, slot, ordinary, root string
	real                          *realGit
	fake                          *workspacetest.Fake
}

func setup(t *testing.T) fixture {
	t.Helper()
	root, e := workspace.CanonicalPath(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	p := filepath.Join(root, "repo")
	if e := os.Mkdir(p, 0755); e != nil {
		t.Fatal(e)
	}
	git(t, p, "init", "-b", "main")
	git(t, p, "-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "core.hooksPath=/dev/null", "commit", "--allow-empty", "-m", "initial")
	s := filepath.Join(root, "worktree", "repo-slot1")
	git(t, p, "worktree", "add", "-b", "main-slot1", s)
	o := filepath.Join(root, "ordinary")
	git(t, p, "worktree", "add", "-b", "ordinary", o)
	trees, e := workspace.ParseWorktrees([]byte(git(t, p, "worktree", "list", "--porcelain", "-z")))
	if e != nil {
		t.Fatal(e)
	}
	oid := git(t, p, "rev-parse", "HEAD")
	f := &workspacetest.Fake{Repositories: []*workspacetest.Repository{{CommonDir: filepath.Join(p, ".git"), Worktrees: trees, Refs: map[string]string{"main": oid, "main-slot1": oid, "ordinary": oid}}}}
	return fixture{p, s, o, root, &realGit{}, f}
}
func TestResolveConformance(t *testing.T) {
	f := setup(t)
	fakeBefore, _ := json.Marshal(f.fake.Repositories)
	before := git(t, f.primary, "show-ref") + git(t, f.primary, "worktree", "list", "--porcelain", "-z") + git(t, f.slot, "status", "--porcelain")
	for _, tc := range []struct{ dir, address, kind, addr string }{{f.primary, "", "primary", "repo:0"}, {f.slot, "", "slot", "repo:1"}, {f.ordinary, "", "worktree", ""}, {f.slot, "repo", "primary", "repo:0"}, {f.primary, ":1", "slot", "repo:1"}, {f.ordinary, "repo:1", "slot", "repo:1"}} {
		t.Run(tc.kind+tc.address, func(t *testing.T) {
			a, e := workspace.Resolve(f.real, tc.dir, tc.address)
			if e != nil {
				t.Fatal(e)
			}
			b, e := workspace.Resolve(f.fake, tc.dir, tc.address)
			if e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(a, b) || a.Kind != tc.kind {
				t.Fatalf("real %+v fake %+v", a, b)
			}
			if tc.addr != "" && (a.Address == nil || *a.Address != tc.addr) {
				t.Fatal(a)
			}
		})
	}
	after := git(t, f.primary, "show-ref") + git(t, f.primary, "worktree", "list", "--porcelain", "-z") + git(t, f.slot, "status", "--porcelain")
	if before != after {
		t.Fatal("resolver mutated Git")
	}
	fakeAfter, _ := json.Marshal(f.fake.Repositories)
	if string(fakeBefore) != string(fakeAfter) {
		t.Fatal("resolver mutated fake Git")
	}
	git(t, f.slot, "checkout", "-b", "issue")
	f.fake.Repositories[0].Worktrees[slotIndex(f)].Branch = "issue"
	for _, r := range []workspace.GitReader{f.real, f.fake} {
		id, e := workspace.Resolve(r, f.slot, "")
		if e != nil || *id.Branch != "issue" {
			t.Fatal(id, e)
		}
	}
	git(t, f.primary, "branch", "-D", "main-slot1")
	delete(f.fake.Repositories[0].Refs, "main-slot1")
	for _, r := range []workspace.GitReader{f.real, f.fake} {
		if _, e := workspace.Resolve(r, f.slot, ""); e == nil {
			t.Fatal("missing resting ref accepted")
		}
		if _, e := workspace.NormalizeVantage(r, f.slot); e != nil {
			t.Fatal("topology must survive invalid slot", e)
		}
	}
}
func TestResolveReadInterleavings(t *testing.T) {
	f := setup(t)
	f.fake.Reads = 0
	if _, e := workspace.Resolve(f.fake, f.slot, ""); e != nil {
		t.Fatal(e)
	}
	count := f.fake.Reads
	for ordinal := 1; ordinal <= count; ordinal++ {
		for _, after := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/after=%v", ordinal, after), func(t *testing.T) {
				f.fake.Reads = 0
				f.fake.Before = nil
				f.fake.After = nil
				r := f.fake.Repositories[0]
				saved := r.Worktrees[slotIndex(f)]
				defer func() { r.Worktrees[slotIndex(f)] = saved }()
				hook := func(_ *workspacetest.Fake, n int) {
					if n == ordinal {
						r.Worktrees[slotIndex(f)].HEAD = strings.Repeat("a", 40)
						r.Worktrees[slotIndex(f)].Branch = "changed"
					}
				}
				if after {
					f.fake.After = hook
				} else {
					f.fake.Before = hook
				}
				id, e := workspace.Resolve(f.fake, f.slot, "")
				conflict := ordinal > 2 && ordinal < count-1 || ordinal == 2 && after || ordinal == count-1 && !after
				if conflict && e == nil {
					t.Fatalf("accepted conflicting read %d after=%v: %+v", ordinal, after, id)
				}
				if e == nil && id.Head != nil && *id.Head != saved.HEAD && *id.Head != strings.Repeat("a", 40) {
					t.Fatal("mixed snapshot", id)
				} // A change outside the observation interval may legitimately produce either snapshot.
			})
		}
	}
}
func TestResolveRejectsEvidenceFailures(t *testing.T) {
	f := setup(t)
	for _, address := range []string{":2", "repo:01", "repo:-1", "rep:1", "../repo:1"} {
		for _, r := range []workspace.GitReader{f.real, f.fake} {
			if _, e := workspace.Resolve(r, f.primary, address); e == nil {
				t.Fatal("accepted", address)
			}
		}
	}
	f.fake.Reads = 0
	if _, e := workspace.Resolve(f.fake, f.slot, ""); e != nil {
		t.Fatal(e)
	}
	count := f.fake.Reads
	for n := 1; n <= count; n++ {
		f.fake.Reads = 0
		f.fake.Faults = map[int]error{n: fmt.Errorf("inaccessible")}
		if _, e := workspace.Resolve(f.fake, f.slot, ""); e == nil {
			t.Fatalf("accepted fault at %d", n)
		}
	}
}
func TestResolveScale(t *testing.T) {
	if testing.Short() {
		t.Skip("real Git scale samples")
	}
	for _, size := range []int{1, 10, 100} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			f := setup(t)
			for n := 3; n < size; n++ {
				git(t, f.primary, "worktree", "add", "--detach", filepath.Join(f.root, fmt.Sprintf("tree%d", n)))
			}
			selected := f.slot
			if size == 1 {
				git(t, f.primary, "worktree", "remove", f.slot)
				git(t, f.primary, "worktree", "remove", f.ordinary)
				selected = f.primary
			}
			f.real.reads = 0
			start := time.Now()
			if _, e := workspace.Resolve(f.real, selected, ""); e != nil {
				t.Fatal(e)
			}
			t.Logf("sample=%d registered=%d duration=%s queries=%d", size, size, time.Since(start), f.real.reads)
		})
	}
}

func slotIndex(f fixture) int {
	for i, w := range f.fake.Repositories[0].Worktrees {
		if w.Path == f.slot {
			return i
		}
	}
	panic("missing slot")
}

func TestResolvePrimaryUnbornAndMaster(t *testing.T) {
	for _, committed := range []bool{false, true} {
		t.Run(fmt.Sprint(committed), func(t *testing.T) {
			root, _ := workspace.CanonicalPath(t.TempDir())
			git(t, root, "init", "-b", "master")
			if committed {
				git(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "core.hooksPath=/dev/null", "commit", "--allow-empty", "-m", "initial")
			}
			id, e := workspace.Resolve(&realGit{}, root, "")
			if e != nil || id.Kind != "primary" || *id.RestingBranch != "main" || (id.Head != nil) != committed {
				t.Fatal(id, e)
			}
		})
	}
}
func TestResolvePathAndMembership(t *testing.T) {
	f := setup(t)
	nested := filepath.Join(f.slot, "a", "b")
	if e := os.MkdirAll(nested, 0755); e != nil {
		t.Fatal(e)
	}
	alias := filepath.Join(f.root, "alias")
	if e := os.Symlink(f.slot, alias); e != nil {
		t.Fatal(e)
	}
	for _, dir := range []string{nested, alias} {
		for _, r := range []workspace.GitReader{f.real, f.fake} {
			id, e := workspace.Resolve(r, dir, "")
			if e != nil || id.WorktreeRoot != f.slot || id.Kind != "slot" {
				t.Fatal(id, e)
			}
		}
	}
	// Duplicate canonical registration is ambiguous even when spellings differ.
	repo := f.fake.Repositories[0]
	duplicate := repo.Worktrees[slotIndex(f)]
	duplicate.Path = alias
	repo.Worktrees = append(repo.Worktrees, duplicate)
	if _, e := workspace.Resolve(f.fake, f.slot, ""); e == nil {
		t.Fatal("canonical duplicate accepted")
	}
	repo.Worktrees = repo.Worktrees[:len(repo.Worktrees)-1]
	// A slot-looking worktree elsewhere has no numbered ownership.
	off := filepath.Join(f.root, "outside", "repo-slot2")
	git(t, f.primary, "worktree", "add", "-b", "main-slot2", off)
	id, e := workspace.Resolve(f.real, off, "")
	if e != nil || id.Kind != "worktree" || id.Address != nil {
		t.Fatal(id, e)
	}
	target := filepath.Join(f.root, "worktree", "repo-slot2")
	if e := os.Symlink(off, target); e != nil {
		t.Fatal(e)
	}
	if _, e := workspace.Resolve(f.real, f.primary, ":2"); e == nil {
		t.Fatal("redirected canonical path accepted")
	}
}
func TestResolveDetachedFault(t *testing.T) {
	f := setup(t)
	git(t, f.slot, "checkout", "--detach")
	w := &f.fake.Repositories[0].Worktrees[slotIndex(f)]
	w.Branch = ""
	w.Detached = true
	f.fake.Reads = 0
	a, e := workspace.Resolve(f.real, f.slot, "")
	if e != nil {
		t.Fatal(e)
	}
	b, e := workspace.Resolve(f.fake, f.slot, "")
	if e != nil || !reflect.DeepEqual(a, b) {
		t.Fatal(a, b, e)
	}
	n := f.fake.Reads
	f.fake.Reads = 0
	f.fake.Faults = map[int]error{n - 1: fmt.Errorf("permission denied")}
	if _, e := workspace.Resolve(f.fake, f.slot, ""); e == nil {
		t.Fatal("detached inaccessible final branch evidence accepted")
	}
}

func TestResolveRefAndMembershipInterleavings(t *testing.T) {
	for _, mutation := range []string{"ref", "membership"} {
		t.Run(mutation, func(t *testing.T) {
			f := setup(t)
			_, e := workspace.Resolve(f.fake, f.slot, "")
			if e != nil {
				t.Fatal(e)
			}
			count := f.fake.Reads
			for n := 1; n <= count; n++ {
				for _, after := range []bool{false, true} {
					t.Run(fmt.Sprintf("%d/after=%v", n, after), func(t *testing.T) {
						r := f.fake.Repositories[0]
						saved := append([]workspace.Worktree(nil), r.Worktrees...)
						savedRef := r.Refs["main-slot1"]
						defer func() { r.Worktrees = saved; r.Refs["main-slot1"] = savedRef }()
						f.fake.Reads = 0
						f.fake.Before = nil
						f.fake.After = nil
						hook := func(_ *workspacetest.Fake, ordinal int) {
							if ordinal != n {
								return
							}
							if mutation == "ref" {
								r.Refs["main-slot1"] = strings.Repeat("b", 40)
							} else {
								idx := slotIndex(f)
								r.Worktrees = append(r.Worktrees[:idx], r.Worktrees[idx+1:]...)
							}
						}
						if after {
							f.fake.After = hook
						} else {
							f.fake.Before = hook
						}
						_, e := workspace.Resolve(f.fake, f.slot, "")
						if mutation == "ref" {
							conflict := n > 4 && n < count || n == 4 && after || n == count && !after
							if conflict && e == nil {
								t.Fatal("accepted conflicting resting ref")
							}
						} else if !(n == count && after) && e == nil {
							t.Fatal("accepted removed membership")
						}
					})
				}
			}
		})
	}
}
func TestResolveStaleBaselineAndOccupiedRefConformance(t *testing.T) {
	f := setup(t)
	git(t, f.slot, "checkout", "-b", "issue")
	f.fake.Repositories[0].Worktrees[slotIndex(f)].Branch = "issue"
	git(t, f.primary, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-m", "advance")
	oid := git(t, f.primary, "rev-parse", "HEAD")
	f.fake.Repositories[0].Worktrees[0].HEAD = oid
	f.fake.Repositories[0].Refs["main"] = oid
	for _, r := range []workspace.GitReader{f.real, f.fake} {
		if _, e := workspace.Resolve(r, f.slot, ""); e != nil {
			t.Fatal("stale resting ref refused", e)
		}
	}
	git(t, f.ordinary, "checkout", "main-slot1")
	for i := range f.fake.Repositories[0].Worktrees {
		if f.fake.Repositories[0].Worktrees[i].Path == f.ordinary {
			f.fake.Repositories[0].Worktrees[i].Branch = "main-slot1"
		}
	}
	for _, r := range []workspace.GitReader{f.real, f.fake} {
		if _, e := workspace.Resolve(r, f.slot, ""); e == nil {
			t.Fatal("occupied resting ref accepted")
		}
	}
}

func TestResolveWrongReservedBranchConformance(t *testing.T) {
	f := setup(t)
	git(t, f.slot, "checkout", "--ignore-other-worktrees", "main")
	f.fake.Repositories[0].Worktrees[slotIndex(f)].Branch = "main"
	for _, r := range []workspace.GitReader{f.real, f.fake} {
		if _, e := workspace.Resolve(r, f.slot, ""); e == nil {
			t.Fatal("mismatched reserved branch accepted")
		}
	}
}
func TestResolveForeignRepositoryAtSlotPath(t *testing.T) {
	f := setup(t)
	foreign := filepath.Join(f.root, "worktree", "repo-slot2")
	if e := os.Mkdir(foreign, 0755); e != nil {
		t.Fatal(e)
	}
	git(t, foreign, "init", "-b", "main")
	git(t, foreign, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-m", "foreign")
	git(t, foreign, "branch", "main-slot2")
	if _, e := workspace.Resolve(f.real, f.primary, ":2"); e == nil {
		t.Fatal("foreign repository accepted as slot")
	}
}

func TestResolveUnbornBranchWithExistingDescendant(t *testing.T) {
	f := setup(t)
	git(t, f.primary, "branch", "new/child")
	git(t, f.primary, "symbolic-ref", "HEAD", "refs/heads/new")
	id, e := workspace.Resolve(f.real, f.primary, "")
	if e != nil || id.Head != nil || id.Branch == nil || *id.Branch != "new" {
		t.Fatal(id, e)
	}
}
