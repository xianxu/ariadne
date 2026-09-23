package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

type workspaceFixture struct{ Fleet, Primary, Slot, Ordinary string }

func newWorkspaceFixture(t *testing.T) workspaceFixture {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	f := workspaceFixture{Fleet: t.TempDir()}
	f.Fleet, _ = filepath.EvalSymlinks(f.Fleet)
	f.Primary = testfix.Repo(t, testfix.At(f.Fleet, "sample"), testfix.InitialCommit())
	f.Slot = filepath.Join(f.Fleet, "worktree", "sample-slot1")
	f.Ordinary = filepath.Join(f.Fleet, "feature")
	testfix.Git(t, f.Primary, "worktree", "add", "-b", "main-slot1", f.Slot)
	testfix.Git(t, f.Slot, "checkout", "-b", "issue-slot")
	testfix.Git(t, f.Primary, "worktree", "add", "-b", "feature", f.Ordinary)
	for _, root := range []string{f.Primary, f.Slot} {
		p := filepath.Join(root, "workshop", "issues", "000001-local.md")
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		body := "---\nid: 000001\nstatus: open\n---\n# " + filepath.Base(root) + "\n## Plan\n- [ ] local\n"
		if err := os.WriteFile(p, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return f
}
func workspaceCommand(t *testing.T, args ...string) (map[string]any, error, string) {
	t.Helper()
	c := buildRoot()
	var b bytes.Buffer
	c.SetOut(&b)
	c.SetErr(&b)
	c.SetArgs(args)
	err := c.Execute()
	var out map[string]any
	if err == nil {
		if e := json.Unmarshal(b.Bytes(), &out); e != nil {
			t.Fatalf("JSON: %v: %s", e, b.String())
		}
	}
	return out, err, b.String()
}
func TestWorkspaceCommandIdentityAndState(t *testing.T) {
	f := newWorkspaceFixture(t)
	nested := filepath.Join(f.Slot, "nested")
	if err := os.Mkdir(nested, 0755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(nested)
	before := testfix.Git(t, f.Primary, "worktree", "list", "--porcelain")
	got, err, _ := workspaceCommand(t, "workspace", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if got["repo"] != "sample" || got["address"] != "sample:1" || got["branch"] != "issue-slot" || got["resting_branch"] != "main-slot1" || got["schema_version"] != float64(1) {
		t.Fatalf("identity: %#v", got)
	}
	for _, address := range []string{":0", "sample", "sample:0"} {
		v, e, _ := workspaceCommand(t, "workspace", address, "--json")
		if e != nil || v["address"] != "sample:0" {
			t.Fatalf("%s: %v %#v", address, e, v)
		}
	}
	state, e, _ := workspaceCommand(t, "state", "--json")
	if e != nil {
		t.Fatal(e)
	}
	ws, ok := state["workspace"].(map[string]any)
	if !ok || ws["address"] != "sample:1" {
		t.Fatalf("state workspace: %#v", state)
	}
	issues, ok := state["issues"].([]any)
	if !ok || len(issues) != 1 || issues[0].(map[string]any)["title"] != "sample-slot1" {
		t.Fatalf("state read wrong checkout: %#v", state)
	}
	if after := testfix.Git(t, f.Primary, "worktree", "list", "--porcelain"); before != after {
		t.Fatal("read changed worktrees")
	}
	if _, err := os.Stat(filepath.Join(f.Primary, ".git", "sdlc.lock")); !os.IsNotExist(err) {
		t.Fatalf("read-only command lock: %v", err)
	}
	t.Chdir(f.Ordinary)
	got, err, _ = workspaceCommand(t, "workspace", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if got["kind"] != "worktree" || got["address"] != nil || got["slot"] != nil || got["resting_branch"] != nil {
		t.Fatalf("ordinary: %#v", got)
	}
}
func TestWorkspaceCommandRefusesMissingSlot(t *testing.T) {
	f := newWorkspaceFixture(t)
	t.Chdir(f.Primary)
	_, err, out := workspaceCommand(t, "workspace", ":9", "--json")
	if err == nil {
		t.Fatal("missing slot accepted")
	}
	if strings.Contains(out, `"schema_version"`) {
		t.Fatalf("partial success: %s", out)
	}
}
func TestWorkspaceArtifactConsumers(t *testing.T) {
	f := newWorkspaceFixture(t)
	t.Chdir(f.Slot)
	peer := testfix.Repo(t, testfix.At(f.Fleet, "peer"), testfix.InitialCommit())
	for _, ref := range []string{"#1", "sample#1"} {
		artifacts, _, err := resolveArtifacts(ref, f.Slot)
		if err != nil {
			t.Fatal(err)
		}
		if len(artifacts) != 1 || !strings.HasPrefix(artifacts[0].Path, f.Slot+string(filepath.Separator)) {
			t.Fatalf("%s: %#v", ref, artifacts)
		}
	}
	if got, err := resolveRepoDir(ArtifactRef{Repo: "peer"}, f.Slot); err != nil || got != peer {
		t.Fatalf("peer: %s %v", got, err)
	}
	name, root := repoNameAndRoot()
	if name != "sample" || root != f.Slot {
		t.Fatalf("name/root: %q %q", name, root)
	}
	if canonicalRepoIssueIdentity(f.Slot, 1) != canonicalRepoIssueIdentity(f.Primary, 1) {
		t.Fatal("same repo issue identities differ")
	}
}

func TestWorkspaceBranchPlacement(t *testing.T) {
	f := newWorkspaceFixture(t)
	t.Chdir(f.Slot)
	var out, diag bytes.Buffer
	got, err := createWorktreeBranch(&out, &diag, "second-issue", execGitRunner{})
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(f.Fleet, "worktree", "sample", "second-issue")
	if got != want {
		t.Fatalf("created %q, want %q", got, want)
	}
	b, err := os.ReadFile(filepath.Join(f.Slot, ".goto"))
	if err != nil || string(b) != want {
		t.Fatalf("goto: %s %v", b, err)
	}
}

func TestWorkspaceMigrationRejectsSameGitRepository(t *testing.T) {
	f := newWorkspaceFixture(t)
	file := filepath.Join(f.Slot, "move.md")
	if err := os.WriteFile(file, []byte("document\n"), 0644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, f.Slot, "add", "move.md")
	testfix.Git(t, f.Slot, "commit", "-qm", "doc")
	t.Chdir(f.Slot)
	var diag bytes.Buffer
	msg, died := expectDie(t, func() {
		_ = runMigrate(&migrateOpts{file: "move.md", destDir: f.Primary, noCleanCheck: true, stderr: &diag})
	})
	if !died || !strings.Contains(msg, "same repo") {
		t.Fatalf("same repository migration accepted: died=%v msg=%s diag=%s", died, msg, diag.String())
	}
	if _, err := os.Stat(file); err != nil {
		t.Fatal("source changed before refusal", err)
	}
	if _, err := os.Stat(filepath.Join(f.Primary, "move.md")); !os.IsNotExist(err) {
		t.Fatal("destination changed before refusal", err)
	}
}

func TestWorkspacePropagationUsesFleetButRetainsSource(t *testing.T) {
	f := newWorkspaceFixture(t)
	peer := testfix.Repo(t, testfix.At(f.Fleet, "consumer"), testfix.InitialCommit())
	if err := os.MkdirAll(filepath.Join(peer, "construct"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(peer, "construct", "deps"), []byte("substrate ../worktree/sample-slot1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := runPropagateBase(f.Slot, "sample#1", true, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "consumer") {
		t.Fatalf("slot source lost fleet dependent: %s", out.String())
	}
}

func TestWorkspaceCloseKeepsPrimaryProjectUntouched(t *testing.T) {
	parent, issuesDir := fleetCloseFixture(t)
	primary := filepath.Join(parent, "ariadne")
	rel := seedPeerProject(t, primary, "ariadne#42")
	primaryBefore, err := os.ReadFile(filepath.Join(primary, rel))
	if err != nil {
		t.Fatal(err)
	}
	slot := filepath.Join(parent, "worktree", "ariadne-slot1")
	testfix.Git(t, primary, "worktree", "add", "-b", "main-slot1", slot)
	testfix.Git(t, slot, "checkout", "-b", "slot-issue")
	peer := initFleetRepo(t, parent, "peer")
	peerRel := seedPeerProject(t, peer, "ariadne#42")
	primaryHead := testfix.Git(t, primary, "rev-parse", "HEAD")
	t.Chdir(slot)
	var out bytes.Buffer
	if err := runClose(&out, &out, closeFlagsFor42(issuesDir)); err != nil {
		t.Fatalf("close: %v %s", err, out.String())
	}
	primaryAfter, err := os.ReadFile(filepath.Join(primary, rel))
	if err != nil || !bytes.Equal(primaryBefore, primaryAfter) {
		t.Fatal("close changed primary project", err)
	}
	if testfix.Git(t, primary, "rev-parse", "HEAD") != primaryHead {
		t.Fatal("close committed into primary")
	}
	for _, path := range []string{filepath.Join(slot, rel), filepath.Join(peer, peerRel)} {
		b, err := os.ReadFile(path)
		if err != nil || !bytes.Contains(b, []byte("[x] [ariadne#42]")) {
			t.Fatalf("project not updated %s: %s %v", path, b, err)
		}
	}
	if subject := testfix.Git(t, slot, "log", "-1", "--pretty=%s"); strings.Contains(subject, "close-time update") {
		t.Fatal("slot auto-committed as peer")
	}
}
