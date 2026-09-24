package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func refreshCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := buildRoot()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(append([]string{"refresh"}, args...))
	err := cmd.ExecuteContext(context.Background())
	return out.String(), err
}

func refreshRemote(t *testing.T, repo string) (string, string) {
	t.Helper()
	area := t.TempDir()
	bare, publisher := filepath.Join(area, "remote.git"), filepath.Join(area, "publisher")
	setupGit(t, repo, "clone", "--bare", repo, bare)
	setupGit(t, repo, "remote", "add", "origin", bare)
	setupGit(t, repo, "clone", bare, publisher)
	return bare, publisher
}

func TestRefreshCLICompilesAndRetries(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	_, base, host := buildFixture(t)
	var initial bytes.Buffer
	if err := runCompile(context.Background(), weavefs.OSFS{}, host, plan.TargetAll, false, &initial); err != nil {
		t.Fatal(err)
	}
	for _, repo := range []string{base, host} {
		setupGit(t, repo, "init", "-q", "-b", "main")
		setupGit(t, repo, "add", ".")
		setupGit(t, repo, "add", "-f", "AGENTS.local.md")
		setupGit(t, repo, "commit", "-qm", "initial")
	}
	_, pubBase := refreshRemote(t, base)
	_, pubHost := refreshRemote(t, host)
	setupGit(t, host, "branch", "-m", "main-slot1")
	signal := filepath.Join(t.TempDir(), "fail")
	log := filepath.Join(t.TempDir(), "builds")
	t.Setenv("REFRESH_TEST_FAIL", signal)
	t.Setenv("REFRESH_TEST_LOG", log)
	mkfile(t, signal, "fail")
	mkfile(t, filepath.Join(pubBase, "AGENTS.local.md"), "UPDATED BASE")
	mkfile(t, filepath.Join(pubHost, "Makefile"), "tools:\n\t@printf 'build\\n' >> \"$$REFRESH_TEST_LOG\"\n\t@test ! -e \"$$REFRESH_TEST_FAIL\"\n")
	for _, repo := range []string{pubBase, pubHost} {
		setupGit(t, repo, "add", "-f", ".")
		setupGit(t, repo, "commit", "-qm", "update")
		setupGit(t, repo, "push", "origin", "main")
	}
	wantBase := setupGit(t, pubBase, "rev-parse", "HEAD")
	wantHost := setupGit(t, pubHost, "rev-parse", "HEAD")
	t.Chdir(host)
	if out, err := refreshCommand(t); err == nil || !strings.Contains(err.Error(), "tools") {
		t.Fatalf("want compile failure, got %v: %s", err, out)
	}
	if got := setupGit(t, base, "rev-parse", "HEAD"); got != wantBase {
		t.Fatalf("base update lost: %s", got)
	}
	if got := setupGit(t, host, "rev-parse", "HEAD"); got != wantHost {
		t.Fatalf("host update lost: %s", got)
	}
	if err := os.Remove(signal); err != nil {
		t.Fatal(err)
	}
	if out, err := refreshCommand(t); err != nil {
		t.Fatalf("retry: %v: %s", err, out)
	}
	got, err := os.ReadFile(filepath.Join(host, "AGENTS.md"))
	if err != nil || !strings.Contains(string(got), "UPDATED BASE") {
		t.Fatalf("composition: %s %v", got, err)
	}
	builds, err := os.ReadFile(log)
	if err != nil || string(builds) != "build\nbuild\n" {
		t.Fatalf("compile retry count: %q %v", builds, err)
	}
	if got := setupGit(t, host, "branch", "--show-current"); got != "main-slot1" {
		t.Fatalf("branch changed: %s", got)
	}
	mkfile(t, filepath.Join(host, "local.txt"), "local work")
	setupGit(t, host, "add", "local.txt")
	setupGit(t, host, "commit", "-qm", "local work")
	localHead := setupGit(t, host, "rev-parse", "HEAD")
	if out, err := refreshCommand(t); err == nil || !strings.Contains(err.Error(), "--rebase") {
		t.Fatalf("ahead branch default: %v %s", err, out)
	}
	if out, err := refreshCommand(t, "--rebase"); err != nil {
		t.Fatalf("explicit rebase: %v %s", err, out)
	}
	if got := setupGit(t, host, "rev-parse", "HEAD"); got != localHead {
		t.Fatalf("unnecessary replay changed local head: %s", got)
	}
}

func TestRefreshCLIHelpAndArguments(t *testing.T) {
	out, err := refreshCommand(t, "--help")
	if err != nil || !strings.Contains(out, "--rebase") || !strings.Contains(out, "origin/main") {
		t.Fatalf("help: %s %v", out, err)
	}
	if _, err := refreshCommand(t, "unexpected"); err == nil {
		t.Fatal("accepted positional arguments")
	}
}

func TestRefreshNumberedLease(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	_, env, host := numberedSetupFixture(t)
	// Prepare existing generated metadata before putting the starting point on main.
	var initial bytes.Buffer
	if err := runCompile(context.Background(), weavefs.OSFS{}, host, plan.TargetAll, false, &initial); err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	mark := filepath.Join(t.TempDir(), "lease-probed")
	t.Setenv("REFRESH_TEST_EXE", exe)
	t.Setenv("REFRESH_TEST_ENV", env)
	t.Setenv("REFRESH_TEST_MARK", mark)
	mkfile(t, filepath.Join(host, "Makefile"), "tools:\n\t@\"$$REFRESH_TEST_EXE\" -test.run=^TestRefreshLeaseProbe$\n")
	setupGit(t, host, "add", ".")
	setupGit(t, host, "add", "-f", "Makefile")
	setupGit(t, host, "commit", "--allow-empty", "-qm", "prepared")
	area := t.TempDir()
	bare := filepath.Join(area, "remote.git")
	setupGit(t, host, "clone", "--bare", host, bare)
	setupGit(t, host, "push", bare, "HEAD:refs/heads/main")
	setupGit(t, host, "remote", "add", "origin", bare)
	t.Chdir(host)
	lease, err := staging.AcquireSetup(env)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := refreshCommand(t); !errors.Is(err, staging.ErrSetupInUse) {
		t.Fatalf("lost setup exclusion: %v", err)
	}
	if err := lease.Close(); err != nil {
		t.Fatal(err)
	}
	if out, err := refreshCommand(t); err != nil {
		t.Fatalf("refresh reacquired its own setup lease: %v: %s", err, out)
	}
	if _, err := os.Stat(mark); err != nil {
		t.Fatalf("compile did not probe held lease: %v", err)
	}
}

// Invoked as a real compile child: setup exclusion must still be held during tools.
func TestRefreshLeaseProbe(t *testing.T) {
	env := os.Getenv("REFRESH_TEST_ENV")
	if env == "" {
		return
	}
	lease, err := staging.AcquireSetup(env)
	if lease != nil {
		lease.Close()
	}
	if !errors.Is(err, staging.ErrSetupInUse) {
		t.Fatalf("compile ran without setup lease: %v", err)
	}
	if err := os.WriteFile(os.Getenv("REFRESH_TEST_MARK"), []byte("held"), 0600); err != nil {
		t.Fatal(err)
	}
}
