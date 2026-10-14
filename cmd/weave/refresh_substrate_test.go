package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func TestRefreshSubstrateIssueBranchKeepsConsumerAndOtherSlots(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	fleet, env, host := numberedSetupFixture(t)
	dep := filepath.Join(env, "base")
	mkfile(t, filepath.Join(dep, "construct/base.manifest"), "prose AGENTS.local.md\n")
	mkfile(t, filepath.Join(dep, "AGENTS.local.md"), "OLD PRIVATE BASE\n")
	setupGit(t, dep, "init", "-q", "-b", "main")
	setupGit(t, dep, "add", ".")
	setupGit(t, dep, "add", "-f", "AGENTS.local.md")
	setupGit(t, dep, "commit", "-qm", "base")
	depBefore := setupGit(t, dep, "rev-parse", "HEAD")
	remote, pubDep := refreshRemote(t, dep)
	shared := filepath.Join(fleet, "base")
	other := filepath.Join(fleet, "worktree", "host-slot2", "base")
	setupGit(t, fleet, "clone", remote, shared)
	setupGit(t, fleet, "clone", remote, other)
	mkfile(t, filepath.Join(host, "construct/deps"), "substrate ../base\n")
	mkfile(t, filepath.Join(host, ".gitignore"), ".weave/\n.agents/\n.claude/\nAGENTS.md\nCLAUDE.md\nGEMINI.md\n")
	log := filepath.Join(t.TempDir(), "builds")
	t.Setenv("REFRESH_TEST_LOG", log)
	mkfile(t, filepath.Join(host, "Makefile"), "tools:\n\t@printf 'build\\n' >> \"$$REFRESH_TEST_LOG\"\n")
	var initial bytes.Buffer
	if err := runCompile(context.Background(), weavefs.OSFS{}, host, plan.TargetAll, false, &initial); err != nil {
		t.Fatal(err)
	}
	setupGit(t, host, "add", ".")
	setupGit(t, host, "commit", "-qm", "prepared")
	bare, pubHost := refreshRemote(t, host)
	setupGit(t, host, "push", bare, "HEAD:refs/heads/main")
	setupGit(t, host, "fetch", "origin")
	setupGit(t, pubHost, "fetch", "origin")
	setupGit(t, pubHost, "checkout", "-B", "main", "origin/main")
	setupGit(t, host, "switch", "-c", "000001-issue")
	mkfile(t, filepath.Join(host, "issue-source"), "committed issue work\n")
	setupGit(t, host, "add", "issue-source")
	setupGit(t, host, "commit", "-qm", "issue work")
	head := setupGit(t, host, "rev-parse", "HEAD")
	index := setupGit(t, host, "write-tree")
	rootRemote := setupGit(t, host, "rev-parse", "origin/main")
	mkfile(t, filepath.Join(pubHost, "remote-only"), "remote work\n")
	setupGit(t, pubHost, "add", ".")
	setupGit(t, pubHost, "commit", "-qm", "remote consumer work")
	setupGit(t, pubHost, "push", "origin", "main")
	mkfile(t, filepath.Join(pubDep, "AGENTS.local.md"), "UPDATED PRIVATE BASE\n")
	setupGit(t, pubDep, "add", ".")
	setupGit(t, pubDep, "add", "-f", "AGENTS.local.md")
	setupGit(t, pubDep, "commit", "-qm", "substrate update")
	setupGit(t, pubDep, "push", "origin", "main")
	want := setupGit(t, pubDep, "rev-parse", "HEAD")
	if err := os.Remove(log); err != nil {
		t.Fatal(err)
	}
	t.Chdir(host)
	if out, err := refreshCommand(t, "--substrate"); err != nil {
		t.Fatalf("substrate refresh on divergent issue branch: %v\n%s", err, out)
	}
	for _, tc := range []struct{ dir, ref, want string }{
		{host, "HEAD", head}, {host, "origin/main", rootRemote}, {dep, "HEAD", want},
		{shared, "HEAD", depBefore}, {other, "HEAD", depBefore},
	} {
		if got := setupGit(t, tc.dir, "rev-parse", tc.ref); got != tc.want {
			t.Fatalf("%s %s changed unexpectedly: %s != %s", tc.dir, tc.ref, got, tc.want)
		}
	}
	if got := setupGit(t, host, "branch", "--show-current"); got != "000001-issue" {
		t.Fatalf("consumer branch moved: %s", got)
	}
	if got := setupGit(t, host, "write-tree"); got != index {
		t.Fatal("consumer index changed")
	}
	if got, err := os.ReadFile(filepath.Join(host, "issue-source")); err != nil || string(got) != "committed issue work\n" {
		t.Fatalf("consumer source changed: %q %v", got, err)
	}
	if got, err := os.ReadFile(filepath.Join(host, "AGENTS.md")); err != nil || !strings.Contains(string(got), "UPDATED PRIVATE BASE") {
		t.Fatalf("new substrate was not composed: %q %v", got, err)
	}
	if got, err := os.ReadFile(log); err != nil || string(got) != "build\n" {
		t.Fatalf("compile count: %q %v", got, err)
	}
}

func TestRefreshSubstrateRejectsRebaseBeforeEffects(t *testing.T) {
	_, env, host := numberedSetupFixture(t)
	t.Chdir(host)
	if out, err := refreshCommand(t, "--substrate", "--rebase"); err == nil {
		t.Fatalf("incompatible refresh modes accepted: %s", out)
	}
	if _, err := os.Lstat(filepath.Join(env, ".weave-setup.lock")); !os.IsNotExist(err) {
		t.Fatalf("invalid flags acquired setup lease: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(host, ".weave")); !os.IsNotExist(err) {
		t.Fatalf("invalid flags compiled: %v", err)
	}
}

func TestRefreshSubstrateRejectsSharedSupplierEscape(t *testing.T) {
	fleet, env, host := numberedSetupFixture(t)
	shared := filepath.Join(fleet, "base")
	mkfile(t, filepath.Join(shared, "construct/base.manifest"), "# shared supplier\n")
	setupGit(t, shared, "init", "-q", "-b", "main")
	setupGit(t, shared, "add", ".")
	setupGit(t, shared, "commit", "-qm", "shared")
	before := setupGit(t, shared, "rev-parse", "HEAD")
	if err := os.Symlink(shared, filepath.Join(env, "base")); err != nil {
		t.Fatal(err)
	}
	mkfile(t, filepath.Join(host, "construct/deps"), "substrate ../base\n")
	setupGit(t, host, "add", ".")
	setupGit(t, host, "commit", "-qm", "declared supplier")
	t.Chdir(host)
	if out, err := refreshCommand(t, "--substrate"); err == nil || !strings.Contains(err.Error(), "ordinary direct child") {
		t.Fatalf("shared supplier was admitted: %v %s", err, out)
	}
	if setupGit(t, shared, "rev-parse", "HEAD") != before {
		t.Fatal("shared supplier changed")
	}
}
