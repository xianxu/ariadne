package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSettingsCheckIsReadOnlyAndClearsAfterRefresh(t *testing.T) {
	root := t.TempDir()
	mkfile(t, filepath.Join(root, "construct/base.manifest"), "merge .claude/settings.base.json .claude/settings.json\n")
	mkfile(t, filepath.Join(root, ".claude/settings.base.json"), `{"new":true}`)
	target := filepath.Join(root, ".claude/settings.json")
	mkfile(t, target, `{"old":true}`)
	t.Chdir(root)
	check := func(wantStale bool) {
		t.Helper()
		var out bytes.Buffer
		cmd := buildRoot()
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs([]string{"settings"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("settings check failed: %v: %s", err, &out)
		}
		n := strings.Count(out.String(), ".claude/settings.json is stale: run weave refresh --substrate from your shell")
		if (wantStale && n != 1) || (!wantStale && n != 0) {
			t.Fatalf("stale=%v: %q", wantStale, out.String())
		}
	}
	check(true)
	if b, _ := os.ReadFile(target); string(b) != `{"old":true}` {
		t.Fatalf("read-only check changed settings: %q", b)
	}
	if _, err := os.Stat(filepath.Join(root, "construct/generated/weave/ownership.json")); !os.IsNotExist(err) {
		t.Fatalf("check created ownership state: %v", err)
	}
	mkfile(t, target, "{\n \"new\": true\n}\n") // operator refresh, formatting-independent comparison
	check(false)
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	check(true)
}

func TestRefreshCanDeferProtectedSettings(t *testing.T) {
	root := t.TempDir()
	mkfile(t, filepath.Join(root, "construct/base.manifest"), "merge .claude/settings.base.json .claude/settings.json\n")
	mkfile(t, filepath.Join(root, ".claude/settings.base.json"), `{"new":true}`)
	mkfile(t, filepath.Join(root, ".claude/settings.json"), `{"old":true}`)
	setupGit(t, root, "init", "-q", "-b", "main")
	setupGit(t, root, "add", ".")
	setupGit(t, root, "commit", "-qm", "initial")
	refreshRemote(t, root)
	t.Chdir(root)
	out, err := refreshCommand(t, "--substrate", "--defer-settings")
	if err != nil {
		t.Fatalf("deferred refresh failed: %v: %s", err, out)
	}
	if n := strings.Count(out, ".claude/settings.json is stale: run weave refresh --substrate from your shell"); n != 1 {
		t.Fatalf("notice count %d: %s", n, out)
	}
	if b, _ := os.ReadFile(filepath.Join(root, ".claude/settings.json")); string(b) != `{"old":true}` {
		t.Fatalf("deferred refresh wrote settings: %q", b)
	}
	// Checkpoint newly generated ignore rules, then simulate normal issue work.
	setupGit(t, root, "add", ".")
	setupGit(t, root, "commit", "-qm", "composition")
	setupGit(t, root, "switch", "-qc", "issue-work")
	mkfile(t, filepath.Join(root, "feature"), "issue work")
	setupGit(t, root, "add", "feature")
	setupGit(t, root, "commit", "-qm", "issue work")
	head := setupGit(t, root, "rev-parse", "HEAD")
	if out, err := refreshCommand(t, "--substrate"); err != nil {
		t.Fatalf("operator recovery on issue branch: %v: %s", err, out)
	}
	if setupGit(t, root, "branch", "--show-current") != "issue-work" || setupGit(t, root, "rev-parse", "HEAD") != head {
		t.Fatal("settings recovery changed issue branch or HEAD")
	}
	var check bytes.Buffer
	cmd := buildRoot()
	cmd.SetOut(&check)
	cmd.SetErr(&check)
	cmd.SetArgs([]string{"settings"})
	if err := cmd.Execute(); err != nil || strings.Contains(check.String(), "is stale") {
		t.Fatalf("operator recovery did not clear notice: %v: %s", err, &check)
	}
}
