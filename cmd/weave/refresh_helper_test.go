package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
)

type suiteHelperFixture struct{ host, environment, supplier, publisher, remote, before, target string }

func newSuiteHelperFixture(t *testing.T) suiteHelperFixture {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	_, environment, host := numberedSetupFixture(t)
	supplier := filepath.Join(environment, "ariadne")
	mkfile(t, filepath.Join(supplier, "construct/base.manifest"), "# supplier\n")
	mkfile(t, filepath.Join(supplier, "Makefile.workflow"), "# workflow\n")
	mkfile(t, filepath.Join(supplier, "scripts/test-lock.py"), "# old helper\n")
	setupGit(t, supplier, "init", "-q", "-b", "main")
	setupGit(t, supplier, "add", ".")
	setupGit(t, supplier, "commit", "-qm", "supplier")
	before := setupGit(t, supplier, "rev-parse", "HEAD")
	remote, publisher := refreshRemote(t, supplier)
	mkfile(t, filepath.Join(publisher, "scripts/test-lock.py"), "# new helper\n")
	setupGit(t, publisher, "add", ".")
	setupGit(t, publisher, "commit", "-qm", "new helper")
	setupGit(t, publisher, "push", "origin", "main")
	target := setupGit(t, publisher, "rev-parse", "HEAD")
	if err := os.Symlink("../ariadne/Makefile.workflow", filepath.Join(host, "Makefile.workflow")); err != nil {
		t.Fatal(err)
	}
	// Deliberately dirty consumer on an issue branch: none of it is a refresh subject.
	setupGit(t, host, "switch", "-c", "issue")
	mkfile(t, filepath.Join(host, "operator-work"), "keep\n")
	return suiteHelperFixture{host, environment, supplier, publisher, remote, before, target}
}

func TestSuiteHelperRefreshUpdatesOnlyCleanMainSupplier(t *testing.T) {
	f := newSuiteHelperFixture(t)
	head, index := setupGit(t, f.host, "rev-parse", "HEAD"), setupGit(t, f.host, "write-tree")
	t.Chdir(f.host)
	out, err := refreshCommand(t, "--suite-helper")
	if err != nil {
		t.Fatalf("helper refresh: %v\n%s", err, out)
	}
	if got := setupGit(t, f.supplier, "rev-parse", "HEAD"); got != f.target {
		t.Fatalf("supplier did not advance: %s\n%s", got, out)
	}
	if setupGit(t, f.host, "rev-parse", "HEAD") != head || setupGit(t, f.host, "write-tree") != index {
		t.Fatal("consumer moved")
	}
	if b, err := os.ReadFile(filepath.Join(f.host, "operator-work")); err != nil || string(b) != "keep\n" {
		t.Fatalf("lost work: %q %v", b, err)
	}
	if _, err := os.Stat(filepath.Join(f.host, ".weave")); !os.IsNotExist(err) {
		t.Fatalf("compiled consumer: %v", err)
	}
	for _, want := range []string{f.supplier, f.before, f.target, "refreshed"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q: %s", want, out)
		}
	}
}

func TestSuiteHelperRefreshSkipsUnsafeSupplier(t *testing.T) {
	for _, kind := range []string{"dirty", "off-main", "detached", "diverged", "busy"} {
		t.Run(kind, func(t *testing.T) {
			f := newSuiteHelperFixture(t)
			switch kind {
			case "dirty":
				mkfile(t, filepath.Join(f.supplier, "untracked"), "keep\n")
			case "off-main":
				setupGit(t, f.supplier, "switch", "-c", "issue")
			case "detached":
				setupGit(t, f.supplier, "checkout", "--detach")
			case "diverged":
				mkfile(t, filepath.Join(f.supplier, "local-work"), "keep\n")
				setupGit(t, f.supplier, "add", ".")
				setupGit(t, f.supplier, "commit", "-qm", "local")
			case "busy":
				lease, err := staging.AcquireSetup(f.environment)
				if err != nil {
					t.Fatal(err)
				}
				defer lease.Close()
			}
			before := setupGit(t, f.supplier, "rev-parse", "HEAD")
			t.Chdir(f.host)
			out, err := refreshCommand(t, "--suite-helper")
			if err != nil {
				t.Fatalf("optional refresh blocks: %v\n%s", err, out)
			}
			if setupGit(t, f.supplier, "rev-parse", "HEAD") != before {
				t.Fatal("unsafe supplier moved")
			}
			if strings.Count(strings.TrimSpace(out), "\n") != 0 || !strings.Contains(out, f.supplier) || !strings.Contains(out, "skipped") || !strings.Contains(out, "repair:") {
				t.Fatalf("want one path/reason/repair line: %s", out)
			}
		})
	}
}

func TestSuiteHelperRefreshRejectsFlagsBeforeSetup(t *testing.T) {
	for _, flag := range []string{"--substrate", "--rebase", "--planning", "--defer-settings"} {
		t.Run(flag, func(t *testing.T) {
			f := newSuiteHelperFixture(t)
			t.Chdir(f.host)
			if _, err := refreshCommand(t, "--suite-helper", flag); err == nil {
				t.Fatal("incompatible flags admitted")
			}
			if _, err := os.Lstat(filepath.Join(f.environment, ".weave-setup.lock")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("setup effect before flag validation: %v", err)
			}
		})
	}
}

func TestSuiteHelperRefreshRejectsHelperSymlinkEscape(t *testing.T) {
	f := newSuiteHelperFixture(t)
	setupGit(t, f.supplier, "pull", "--ff-only", "origin", "main")
	external := filepath.Join(t.TempDir(), "helper.py")
	mkfile(t, external, "# externally supplied helper\n")
	helper := filepath.Join(f.supplier, "scripts/test-lock.py")
	if err := os.Remove(helper); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, helper); err != nil {
		t.Fatal(err)
	}
	setupGit(t, f.supplier, "add", ".")
	setupGit(t, f.supplier, "commit", "-qm", "helper supplied elsewhere")
	setupGit(t, f.supplier, "push", "origin", "main")
	setupGit(t, f.publisher, "pull", "--ff-only")
	mkfile(t, filepath.Join(f.publisher, "new-file"), "update\n")
	setupGit(t, f.publisher, "add", ".")
	setupGit(t, f.publisher, "commit", "-qm", "advance")
	setupGit(t, f.publisher, "push", "origin", "main")
	before := setupGit(t, f.supplier, "rev-parse", "HEAD")
	t.Chdir(f.host)
	out, err := refreshCommand(t, "--suite-helper")
	if err != nil || !strings.Contains(out, "skipped") || setupGit(t, f.supplier, "rev-parse", "HEAD") != before {
		t.Fatalf("refreshed a workflow directory that does not supply the helper: %v %s", err, out)
	}
}

func TestSuiteHelperRefreshLeavesPrimarySupplierAlone(t *testing.T) {
	f := newSuiteHelperFixture(t)
	link := filepath.Join(f.host, "Makefile.workflow")
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(f.publisher, "Makefile.workflow"), link); err != nil {
		t.Fatal(err)
	}
	before := setupGit(t, f.publisher, "rev-parse", "HEAD")
	t.Chdir(f.host)
	out, err := refreshCommand(t, "--suite-helper")
	if err != nil || !strings.Contains(out, "primary fallback") || setupGit(t, f.publisher, "rev-parse", "HEAD") != before || setupGit(t, f.supplier, "rev-parse", "HEAD") != f.before {
		t.Fatalf("fallback touched: %v %s", err, out)
	}
}
