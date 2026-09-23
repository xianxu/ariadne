package workspace_test

import (
	"github.com/xianxu/ariadne/pkg/workspace"
	"github.com/xianxu/ariadne/pkg/workspace/workspacetest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func addIndependent(t *testing.T, f fixture, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
	git(t, path, "init", "-b", "main")
	git(t, path, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-m", "independent")
	syncFake(t, f, path)
}
func syncFake(t *testing.T, f fixture, path string) {
	t.Helper()
	trees, e := workspace.ParseWorktrees([]byte(git(t, path, "worktree", "list", "--porcelain", "-z")))
	if e != nil {
		t.Fatal(e)
	}
	common := git(t, path, "rev-parse", "--git-common-dir")
	if !filepath.IsAbs(common) {
		common = filepath.Join(path, common)
	}
	for _, r := range f.fake.Repositories {
		if r.CommonDir == common {
			r.Worktrees = trees
			return
		}
	}
	f.fake.Repositories = append(f.fake.Repositories, &workspacetest.Repository{CommonDir: common, Worktrees: trees, Refs: map[string]string{"main": git(t, path, "rev-parse", "HEAD")}})
}
func TestEnvironmentIndependentConformance(t *testing.T) {
	f := setup(t)
	env := filepath.Dir(f.slot)
	dep := filepath.Join(env, "dep")
	addIndependent(t, f, dep)
	canonical := filepath.Join(f.root, "dep")
	addIndependent(t, f, canonical)
	feature := filepath.Join(env, ".worktrees", "dep", "feature")
	git(t, dep, "worktree", "add", "-b", "feature", feature)
	syncFake(t, f, dep)
	for _, dir := range []string{f.slot, dep, feature} {
		var prior workspace.Identity
		for i, r := range []workspace.GitReader{f.real, f.fake} {
			id, e := workspace.Resolve(r, dir, "")
			if e != nil {
				t.Fatal(e)
			}
			if i > 0 && !reflect.DeepEqual(prior, id) {
				t.Fatalf("fake/real differ: %+v %+v", prior, id)
			}
			prior = id
			if id.SchemaVersion != 2 || id.EnvironmentRoot != env || id.FleetRoot != f.root || id.EnvironmentHost == nil || id.EnvironmentHost.WorktreeRoot != f.slot {
				t.Fatalf("context %+v", id)
			}
			if dir != f.slot {
				if id.PrimaryRoot != dep || id.RepoIdentity != filepath.Join(dep, ".git") || id.Address != nil || id.Slot != nil || id.RestingBranch != nil {
					t.Fatalf("clone identity lost: %+v", id)
				}
				want := "dependency"
				if dir == feature {
					want = "worktree"
				}
				if id.Kind != want {
					t.Fatal(id)
				}
				if _, e := workspace.Resolve(r, dir, ":0"); e == nil {
					t.Fatal("contextual clone address accepted")
				}
				explicit, e := workspace.Resolve(r, dir, "dep:0")
				if e != nil || explicit.PrimaryRoot != canonical {
					t.Fatal(explicit, e)
				}
			}
			context, e := workspace.DiscoverEnvironment(r, dir)
			if e != nil || context == nil || context.Root != env {
				t.Fatal(context, e)
			}
		}
	}
	// Dependency context remains usable while host resting readiness is broken.
	git(t, f.slot, "checkout", "-b", "issue")
	git(t, f.primary, "branch", "-D", "main-slot1")
	if _, e := workspace.Resolve(f.real, dep, ""); e != nil {
		t.Fatal(e)
	}
}
func TestEnvironmentCandidateFailsClosed(t *testing.T) {
	f := setup(t)
	bad := filepath.Join(f.root, "worktree", "repo-slot2", "dep")
	addIndependent(t, f, bad)
	for _, r := range []workspace.GitReader{f.real, f.fake} {
		if _, e := workspace.NormalizeVantage(r, bad); e == nil {
			t.Fatal("missing host accepted")
		}
		if _, e := workspace.DiscoverEnvironment(r, bad); e == nil {
			t.Fatal("missing host discovery accepted")
		}
	}
	linked := filepath.Join(filepath.Dir(f.slot), "linked")
	git(t, f.primary, "worktree", "add", "-b", "linked-dep", linked)
	syncFake(t, f, f.primary)
	for _, r := range []workspace.GitReader{f.real, f.fake} {
		if _, e := workspace.NormalizeVantage(r, linked); e == nil {
			t.Fatal("linked dependency accepted")
		}
	}
}
func TestEnvironmentLegacyAndArchive(t *testing.T) {
	f := setup(t)
	legacy := filepath.Join(f.root, "worktree", "repo-slot3")
	git(t, f.primary, "worktree", "add", "-b", "legacy", legacy)
	id, e := workspace.Resolve(f.real, legacy, "")
	if e != nil || id.Kind != "worktree" || id.Address != nil || id.EnvironmentHost != nil {
		t.Fatal(id, e)
	}
	archive := t.TempDir()
	c, e := workspace.DiscoverEnvironment(nil, archive)
	if e != nil || c != nil {
		t.Fatal(c, e)
	}
}
func TestFeatureWorktreePath(t *testing.T) {
	for _, dep := range []bool{false, true} {
		id := workspace.Identity{Repo: "dep", FleetRoot: "/fleet", EnvironmentRoot: "/fleet/worktree/host-slot1", PrimaryRoot: "/fleet/dep"}
		want := "/fleet/worktree/dep/feature/topic"
		if dep {
			id.EnvironmentHost = &workspace.EnvironmentHost{RepoIdentity: "/fleet/host/.git"}
			id.RepoIdentity = "/fleet/worktree/host-slot1/dep/.git"
			want = "/fleet/worktree/host-slot1/.worktrees/dep/feature/topic"
		}
		p, e := workspace.FeatureWorktreePath(id, "feature/topic")
		if e != nil || p != want {
			t.Fatal(p, e)
		}
		for _, branch := range []string{"", "../escape", "/absolute", "a/../b"} {
			if _, e := workspace.FeatureWorktreePath(id, branch); e == nil {
				t.Fatal("unsafe branch", branch)
			}
		}
	}
}

func TestEnvironmentRejectsSymlinkDependency(t *testing.T) {
	f := setup(t)
	outside := filepath.Join(f.root, "dep")
	addIndependent(t, f, outside)
	alias := filepath.Join(filepath.Dir(f.slot), "dep")
	if e := os.Symlink(outside, alias); e != nil {
		t.Fatal(e)
	}
	for _, r := range []workspace.GitReader{f.real, f.fake} {
		if _, e := workspace.NormalizeVantage(r, alias); e == nil {
			t.Fatal("symlink clone redirects authority")
		}
		if _, e := workspace.DiscoverEnvironment(r, alias); e == nil {
			t.Fatal("symlink discovery redirects authority")
		}
	}
}

func TestEnvironmentRejectsConflictingPrimaryRegistration(t *testing.T) {
	f := setup(t)
	// Host answers still claim membership, but its canonical primary no longer agrees.
	r := &conflictingHostReader{GitReader: f.fake, primary: f.primary, host: f.slot}
	if _, e := workspace.NormalizeVantage(r, f.slot); e == nil {
		t.Fatal("host registration not corroborated by primary")
	}
}

type conflictingHostReader struct {
	workspace.GitReader
	primary, host string
}

func (r *conflictingHostReader) GitInDir(dir string, args ...string) ([]byte, error) {
	if dir == r.primary && len(args) > 0 && args[0] == "worktree" {
		b, e := r.GitReader.GitInDir(dir, args...)
		if e != nil {
			return b, e
		}
		trees, e := workspace.ParseWorktrees(b)
		if e != nil {
			return nil, e
		}
		var kept []workspace.Worktree
		for _, w := range trees {
			if w.Path != r.host {
				kept = append(kept, w)
			}
		}
		return workspacetest.Porcelain(kept), nil
	}
	return r.GitReader.GitInDir(dir, args...)
}

func TestEnvironmentMalformedCandidates(t *testing.T) {
	f := setup(t)
	for _, name := range []string{"repo-slot01", "repo-slot0", "repo-slot-1", "repo-slot999999999999999999999999"} {
		dir := filepath.Join(f.root, "worktree", name, "dep")
		if e := os.MkdirAll(dir, 0755); e != nil {
			t.Fatal(e)
		}
		if _, e := workspace.DiscoverEnvironment(nil, dir); e == nil {
			t.Fatalf("malformed candidate %s silently downgraded", name)
		}
	}
}

func TestEnvironmentDiscoveryFeatureOutside(t *testing.T) {
	f := setup(t)
	dep := filepath.Join(filepath.Dir(f.slot), "dep")
	addIndependent(t, f, dep)
	feature := filepath.Join(f.root, "unrelated-location")
	git(t, dep, "worktree", "add", "-b", "outside", feature)
	syncFake(t, f, dep)
	for _, r := range []workspace.GitReader{f.real, f.fake} {
		got, e := workspace.DiscoverEnvironment(r, feature)
		if e != nil || got == nil || got.Root != filepath.Dir(f.slot) {
			t.Fatal(got, e)
		}
	}
}

func TestEnvironmentFeatureRequiresPrimaryRegistration(t *testing.T) {
	f := setup(t)
	dep := filepath.Join(filepath.Dir(f.slot), "dep")
	addIndependent(t, f, dep)
	feature := filepath.Join(f.root, "feature-dep")
	git(t, dep, "worktree", "add", "-b", "feature", feature)
	syncFake(t, f, dep)
	r := &conflictingHostReader{GitReader: f.fake, primary: dep, host: feature}
	if _, e := workspace.NormalizeVantage(r, feature); e == nil {
		t.Fatal("feature inherited context without primary corroboration")
	}
}

func TestEnvironmentCandidateCannotBorrowAncestorGit(t *testing.T) {
	f := setup(t)
	dir := filepath.Join(f.primary, "worktree", "host-slot1", "dep")
	if e := os.MkdirAll(dir, 0755); e != nil {
		t.Fatal(e)
	}
	if _, e := workspace.DiscoverEnvironment(f.real, dir); e == nil {
		t.Fatal("candidate silently borrowed outer repository")
	}
}
