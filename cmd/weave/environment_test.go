package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func setupGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com", "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func numberedSetupFixture(t *testing.T) (string, string, string) {
	t.Helper()
	return numberedSetupFixtureAt(t, t.TempDir())
}

func numberedSetupFixtureAt(t *testing.T, fleetPath string) (string, string, string) {
	t.Helper()
	if err := os.MkdirAll(fleetPath, 0755); err != nil {
		t.Fatal(err)
	}
	fleet, err := filepath.EvalSymlinks(fleetPath)
	if err != nil {
		t.Fatal(err)
	}
	primary := filepath.Join(fleet, "host")
	mkfile(t, filepath.Join(primary, "construct/base.manifest"), "# host\n")
	setupGit(t, primary, "init", "-q", "-b", "main")
	setupGit(t, primary, "add", ".")
	setupGit(t, primary, "commit", "-qm", "initial")
	env := filepath.Join(fleet, "worktree", "host-slot1")
	host := filepath.Join(env, "host")
	setupGit(t, primary, "worktree", "add", "-qb", "main-slot1", host)
	return fleet, env, host
}

func executeSetup(t *testing.T, verb string, dry bool) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := buildRoot()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	args := []string{verb}
	if dry {
		args = append(args, "--dry-run")
	}
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestNumberedSetupContentionAndDryRun(t *testing.T) {
	for _, verb := range []string{"compile", "dependencies"} {
		t.Run(verb, func(t *testing.T) {
			_, env, host := numberedSetupFixture(t)
			t.Chdir(host)
			if out, err := executeSetup(t, verb, true); err != nil {
				t.Fatal(err, out)
			}
			if _, err := os.Lstat(filepath.Join(env, ".weave-setup.lock")); !os.IsNotExist(err) {
				t.Fatal("dry-run created lease", err)
			}
			lease, err := staging.AcquireSetup(env)
			if err != nil {
				t.Fatal(err)
			}
			defer lease.Close()
			if out, err := executeSetup(t, verb, true); err != nil {
				t.Fatal("dry-run contended", err, out)
			}
			if _, err := executeSetup(t, verb, false); !errors.Is(err, staging.ErrSetupInUse) {
				t.Fatalf("setup failed to exclude concurrent operation: %v", err)
			}
			if _, err := os.Lstat(filepath.Join(host, ".weave")); !os.IsNotExist(err) {
				t.Fatal("contending call wrote artifacts", err)
			}
		})
	}
}

func TestNumberedSetupKeepsPrivateSourceAndExternalSupplier(t *testing.T) {
	for _, verb := range []string{"compile", "dependencies"} {
		t.Run(verb, func(t *testing.T) {
			fleet, env, host := numberedSetupFixture(t)
			t.Chdir(host)
			home := t.TempDir()
			t.Setenv("HOME", home)
			supplier := filepath.Join(fleet, "base")
			mkfile(t, filepath.Join(supplier, "sentinel"), "external supplier\n")
			mkfile(t, filepath.Join(home, ".zshrc"), "authored shell rc\n")
			dep := filepath.Join(env, "base")
			mkfile(t, filepath.Join(dep, "construct/base.manifest"), "prose AGENTS.local.md\n")
			mkfile(t, filepath.Join(dep, "AGENTS.local.md"), "private feature prose\n")
			mkfile(t, filepath.Join(dep, "Brewfile"), "# fixture\n")
			mkfile(t, filepath.Join(dep, "Makefile"), "tools:\n\tprintf built > built\nsdlc-install:\n\tfalse\n")
			setupGit(t, dep, "init", "-q", "-b", "feature")
			setupGit(t, dep, "add", ".")
			setupGit(t, dep, "commit", "-qm", "feature")
			head := setupGit(t, dep, "rev-parse", "HEAD")
			mkfile(t, filepath.Join(dep, "untracked"), "keep me\n")
			mkfile(t, filepath.Join(host, "construct/deps"), "substrate ../base\n")
			bin := t.TempDir()
			mkfile(t, filepath.Join(bin, "brew"), "#!/bin/sh\nprintf '%s\\n' \"$PWD\" >> \"$HOME/brew.log\"\n")
			if err := os.Chmod(filepath.Join(bin, "brew"), 0755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			for i := 0; i < 2; i++ {
				if out, err := executeSetup(t, verb, false); err != nil {
					t.Fatal(err, out)
				}
			}
			if got := setupGit(t, dep, "rev-parse", "HEAD"); got != head {
				t.Fatal("feature revision changed")
			}
			for path, want := range map[string]string{filepath.Join(dep, "untracked"): "keep me\n", filepath.Join(supplier, "sentinel"): "external supplier\n", filepath.Join(home, ".zshrc"): "authored shell rc\n", filepath.Join(home, "brew.log"): dep + "\n" + dep + "\n"} {
				got, err := os.ReadFile(path)
				if err != nil || string(got) != want {
					t.Fatalf("%s = %q, %v", path, got, err)
				}
			}
			_, buildErr := os.Stat(filepath.Join(dep, "built"))
			if verb == "compile" && buildErr != nil {
				t.Fatal(buildErr)
			}
			if verb == "dependencies" && !os.IsNotExist(buildErr) {
				t.Fatal("dependencies ran builds")
			}
			lease, err := staging.AcquireSetup(env)
			if err != nil {
				t.Fatal("setup leaked lease", err)
			}
			lease.Close()
		})
	}
}

func TestArchiveCompileNeedsNoGitExecutable(t *testing.T) {
	root := t.TempDir()
	mkfile(t, filepath.Join(root, "construct/base.manifest"), "prose AGENTS.local.md\n")
	mkfile(t, filepath.Join(root, "AGENTS.local.md"), "archive prose\n")
	t.Chdir(root)
	t.Setenv("PATH", t.TempDir())
	if out, err := executeSetup(t, "compile", false); err != nil {
		t.Fatal(err, out)
	}
}

func TestDependencyFeatureCompileCannotFallBackToGenericAcquisition(t *testing.T) {
	_, env, _ := numberedSetupFixture(t)
	dep := filepath.Join(env, "base")
	mkfile(t, filepath.Join(dep, "construct/base.manifest"), "# base\n")
	mkfile(t, filepath.Join(dep, "construct/deps"), "substrate ../other https://example.com/other.git\n")
	setupGit(t, dep, "init", "-q", "-b", "main")
	setupGit(t, dep, "add", ".")
	setupGit(t, dep, "commit", "-qm", "base")
	feature := filepath.Join(env, ".worktrees", "base", "feature")
	setupGit(t, dep, "worktree", "add", "-qb", "feature", feature)
	t.Chdir(feature)
	for _, verb := range []string{"compile", "dependencies"} {
		if _, err := executeSetup(t, verb, true); err == nil || !strings.Contains(err.Error(), "compose from the dependency primary") {
			t.Fatalf("%s unsupported layout error = %v", verb, err)
		}
	}
}

type setupGuardFS struct {
	weavefs.OSFS
	t            *testing.T
	environment  string
	publications int
}

func (fs *setupGuardFS) Rename(oldpath, newpath string) error {
	fs.t.Helper()
	lease, err := staging.AcquireSetup(fs.environment)
	if lease != nil {
		lease.Close()
	}
	if !errors.Is(err, staging.ErrSetupInUse) {
		fs.t.Fatalf("publication %s lacks setup exclusion: %v", newpath, err)
	}
	fs.publications++
	return fs.OSFS.Rename(oldpath, newpath)
}

func TestCompileSetupLeaseCoversFinalPublication(t *testing.T) {
	_, env, host := numberedSetupFixture(t)
	mkfile(t, filepath.Join(host, "construct/base.manifest"), "prose AGENTS.local.md\n")
	mkfile(t, filepath.Join(host, "AGENTS.local.md"), "host prose\n")
	fs := &setupGuardFS{t: t, environment: env}
	var out bytes.Buffer
	if err := run(fs, host, plan.TargetAll, false, &out); err != nil {
		t.Fatal(err, out.String())
	}
	if fs.publications == 0 {
		t.Fatal("fixture did not publish final files")
	}
	lease, err := staging.AcquireSetup(env)
	if err != nil {
		t.Fatal(err)
	}
	lease.Close()
}

func TestSetupDiscoveryPreservesPathOwnedWhitespace(t *testing.T) {
	for _, suffix := range []string{" ", "\r"} {
		t.Run(fmt.Sprintf("suffix=%q", suffix), func(t *testing.T) {
			_, env, _ := numberedSetupFixtureAt(t, filepath.Join(t.TempDir(), "fleet"+suffix))
			root := filepath.Join(env, "base"+suffix)
			mkfile(t, filepath.Join(root, "construct/base.manifest"), "# base\n")
			setupGit(t, root, "init", "-q", "-b", "main")
			setupGit(t, root, "add", ".")
			setupGit(t, root, "commit", "-qm", "base")
			var out bytes.Buffer
			client, _, closeSetup, err := prepareSetup(context.Background(), root, true, &out, &out)
			if err != nil {
				t.Fatal(err)
			}
			defer closeSetup()
			if client.Policy == nil || client.Policy.EnvironmentRoot != env {
				t.Fatal("lost numbered policy")
			}
			if err := run(weavefs.OSFS{}, root, plan.TargetAll, true, &out); err != nil {
				t.Fatal(err, out.String())
			}
		})
	}
}

func TestCompileRejectsRedirectedEnvironmentBeforeCanonicalizing(t *testing.T) {
	_, env, _ := numberedSetupFixture(t)
	outside := t.TempDir()
	mkfile(t, filepath.Join(outside, "construct/base.manifest"), "# archive\n")
	alias := filepath.Join(env, "base")
	if err := os.Symlink(outside, alias); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, alias, plan.TargetAll, false, &out); err == nil || !strings.Contains(err.Error(), "redirected environment") {
		t.Fatalf("compile discarded lexical environment evidence: %v", err)
	}
}
