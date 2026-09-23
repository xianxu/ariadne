package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

type privateFleet struct{ fleet, host, slot, env, dep, canonical string }

func privateEnvironment(t *testing.T) privateFleet {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	f := privateFleet{fleet: t.TempDir()}
	f.fleet, _ = filepath.EvalSymlinks(f.fleet)
	f.host = testfix.Repo(t, testfix.At(f.fleet, "app"), testfix.InitialCommit())
	f.env = filepath.Join(f.fleet, "worktree", "app-slot1")
	f.slot = filepath.Join(f.env, "app")
	testfix.Git(t, f.host, "worktree", "add", "-b", "main-slot1", f.slot)
	f.dep = testfix.Repo(t, testfix.At(f.env, "base"), testfix.InitialCommit())
	f.canonical = testfix.Repo(t, testfix.At(f.fleet, "base"), testfix.InitialCommit())
	return f
}
func privateWrite(t *testing.T, root, path, body string) {
	t.Helper()
	p := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}
func TestPrivateEnvironmentPeerSelection(t *testing.T) {
	f := privateEnvironment(t)
	localOnly := testfix.Repo(t, testfix.At(f.env, "local-only"), testfix.InitialCommit())
	testfix.Repo(t, testfix.At(f.fleet, "base-extra"), testfix.InitialCommit())
	for _, cwd := range []string{f.slot, f.dep} {
		for name, want := range map[string]string{"base": f.dep, "app": f.slot, "local-only": localOnly} {
			got, err := resolveRepoDir(ArtifactRef{Repo: name}, cwd)
			if err != nil || got != want {
				t.Errorf("%s %s => %s, %v; want %s", cwd, name, got, err, want)
			}
		}
		if _, err := resolveRepoDir(ArtifactRef{Repo: "ba"}, cwd); err == nil || !strings.Contains(err.Error(), "ambiguous") {
			t.Errorf("prefix: %v", err)
		}
	}
	privateWrite(t, f.canonical, "workshop/issues/000001-only-global.md", "---\nid: 000001\nstatus: open\n---\n# global\n")
	if artifacts, _, err := resolveArtifacts("base#1", f.slot); err == nil && len(artifacts) > 0 {
		t.Fatalf("fell back to global: %#v", artifacts)
	}
	if canonicalRepoIssueIdentity(f.dep, 1) == canonicalRepoIssueIdentity(f.canonical, 1) {
		t.Fatal("independent clones merged identity")
	}
}
func TestPrivateEnvironmentProjectsAndCalibration(t *testing.T) {
	f := privateEnvironment(t)
	localOnly := testfix.Repo(t, testfix.At(f.env, "local-only"), testfix.InitialCommit())
	other := testfix.Repo(t, testfix.At(f.fleet, "other"), testfix.InitialCommit())
	body := "---\ntype: project\nstatus: working\n---\n# project\n- [ ] [base#1] task\n"
	for _, root := range []string{f.host, f.slot, f.dep, f.canonical, localOnly, other} {
		privateWrite(t, root, "workshop/projects/test.md", body)
	}
	matches, _, err := discoverProjectsForRef("base#1", f.dep)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, m := range matches {
		got[m.RepoDir] = true
	}
	for _, r := range []string{f.slot, f.dep, localOnly, other} {
		if !got[r] {
			t.Errorf("missing %s: %#v", r, matches)
		}
	}
	for _, r := range []string{f.host, f.canonical} {
		if got[r] {
			t.Errorf("shadowed primary selected: %s", r)
		}
	}
	id, err := resolveWorkspace(f.dep)
	if err != nil {
		t.Fatal(err)
	}
	if id.FleetRoot != f.fleet {
		t.Errorf("calibration fleet = %s", id.FleetRoot)
	}
}
func TestPrivateEnvironmentFeatureWorktreeAndPropagation(t *testing.T) {
	f := privateEnvironment(t)
	privateWrite(t, f.slot, "construct/deps", "substrate ../base\n")
	privateWrite(t, f.dep, "construct/base.manifest", "")
	global := testfix.Repo(t, testfix.At(f.fleet, "global-consumer"), testfix.InitialCommit())
	privateWrite(t, global, "construct/deps", "substrate "+f.dep+"\n")
	var out bytes.Buffer
	if err := runPropagateBase(f.dep, "base#1", true, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "app") || strings.Contains(out.String(), "global-consumer") {
		t.Fatalf("scope: %s", out.String())
	}
	t.Chdir(f.dep)
	path, err := createWorktreeBranch(&out, &out, "feature", execGitRunner{})
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(f.env, ".worktrees", "base", "feature")
	if path != want {
		t.Fatalf("worktree %s want %s", path, want)
	}
	if got, err := resolveRepoDir(ArtifactRef{Repo: "app"}, path); err != nil || got != f.slot {
		t.Fatalf("inherited peer %s %v", got, err)
	}
}

func TestPrivateEnvironmentInboundReferences(t *testing.T) {
	f := privateEnvironment(t)
	for _, root := range []string{f.slot, f.dep, f.canonical, f.host} {
		privateWrite(t, root, "ref.txt", "old.md\n")
		testfix.Git(t, root, "add", "ref.txt")
		testfix.Git(t, root, "commit", "-qm", "reference")
	}
	var out bytes.Buffer
	reportInboundRefs(&out, f.dep, f.slot, "old.md", "new.md")
	// Both selected repositories carry the reference. Shadowed copies must not
	// produce additional identical reports or expose their distinct marker.
	if strings.Count(out.String(), "ref.txt:1:old.md") != 2 {
		t.Fatalf("scope: %s", out.String())
	}
	privateWrite(t, f.canonical, "canonical-only.txt", "old.md\n")
	testfix.Git(t, f.canonical, "add", "canonical-only.txt")
	testfix.Git(t, f.canonical, "commit", "-qm", "global reference")
	out.Reset()
	reportInboundRefs(&out, f.dep, f.slot, "old.md", "new.md")
	if strings.Contains(out.String(), "canonical-only") {
		t.Fatalf("shadowed inbound: %s", out.String())
	}
}
func TestPrivateEnvironmentRejectsBrokenLocalEvidence(t *testing.T) {
	f := privateEnvironment(t)
	privateWrite(t, f.env, "broken/.git", "gitdir: /missing/private-git\n")
	if _, err := resolveRepoDir(ArtifactRef{Repo: "base"}, f.slot); err == nil {
		t.Fatal("broken local evidence silently ignored")
	}
	id, err := resolveWorkspace(f.slot)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := projectWorkspaceRoots(id); err == nil {
		t.Fatal("project fallback ignored broken local evidence")
	}
}

func TestPrivateEnvironmentCloseWritesOnlySelectedProjects(t *testing.T) {
	parent, issuesDir := fleetCloseFixture(t)
	primary := filepath.Join(parent, "ariadne")
	host := initFleetRepo(t, parent, "app")
	env := filepath.Join(parent, "worktree", "app-slot1")
	slot := filepath.Join(env, "app")
	testfix.Git(t, host, "worktree", "add", "-b", "main-slot1", slot)
	dep := filepath.Join(env, "ariadne")
	testfix.Git(t, parent, "clone", "--no-local", primary, dep)
	// The private clone's issue is the same source revision but independent Git
	// authority. Close should edit the selected host project, not canonical app.
	rel := seedPeerProject(t, host, "ariadne#42")
	seedPeerProject(t, slot, "ariadne#42")
	before := testfix.Git(t, host, "rev-parse", "HEAD")
	t.Chdir(dep)
	var out bytes.Buffer
	if err := runClose(&out, &out, closeFlagsFor42(issuesDir)); err != nil {
		t.Fatalf("close %v: %s", err, out.String())
	}
	global, err := os.ReadFile(filepath.Join(host, rel))
	if err != nil {
		t.Fatal(err)
	}
	local, err := os.ReadFile(filepath.Join(slot, rel))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(global, []byte("[x]")) || !bytes.Contains(local, []byte("[x] [ariadne#42]")) {
		t.Fatalf("wrong project writes: global=%s local=%s", global, local)
	}
	if testfix.Git(t, host, "rev-parse", "HEAD") != before {
		t.Fatal("canonical host branch changed")
	}
}

func TestPrivateEnvironmentHostPropagation(t *testing.T) {
	f := privateEnvironment(t)
	privateWrite(t, f.dep, "construct/deps", "substrate ../app\n")
	privateWrite(t, f.slot, "construct/base.manifest", "")
	var out bytes.Buffer
	if err := runPropagateBase(f.slot, "app#1", true, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "base") {
		t.Fatalf("lost host source: %s", out.String())
	}
}
func TestSelectWorkspaceRepo(t *testing.T) {
	current := workspaceRepo{"app", "/private/app"}
	repos := []workspaceRepo{current, {"base", "/private/base"}, {"base-extra", "/global/base-extra"}, {"Local", "/private/Local"}}
	for _, tc := range []struct{ token, want string }{{"", current.Root}, {"app", current.Root}, {"base", "/private/base"}, {"loc", "/private/Local"}} {
		got, err := selectWorkspaceRepo(tc.token, current, repos)
		if err != nil || got.Root != tc.want {
			t.Fatalf("%s => %v %v", tc.token, got, err)
		}
	}
	for _, bad := range []string{"ba", "missing"} {
		if _, err := selectWorkspaceRepo(bad, current, repos); err == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
}

func TestPrivateEnvironmentDoesNotApplyFleetNameHeuristics(t *testing.T) {
	f := privateEnvironment(t)
	privateWrite(t, f.env, ".weave-setup.lock", "")
	for _, name := range []string{".private-base", "base.bak", "worktree"} {
		root := testfix.Repo(t, testfix.At(f.env, name), testfix.InitialCommit())
		got, err := resolveRepoDir(ArtifactRef{Repo: name}, f.slot)
		if err != nil || got != root {
			t.Errorf("accepted sibling clone %s invisible: %s %v", name, got, err)
		}
	}
	if err := os.Symlink(f.canonical, filepath.Join(f.env, "redirected")); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveRepoDir(ArtifactRef{Repo: "base"}, f.slot); err == nil {
		t.Fatal("redirected local Git evidence ignored")
	}
}
