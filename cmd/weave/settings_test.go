package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
	"github.com/xianxu/ariadne/pkg/weaveownership"
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

func TestDeferredSettingsProtectsRemovedDeclarationAndBackup(t *testing.T) {
	for _, state := range []string{"owned", "edited", "missing", "backup only", "unmanaged"} {
		t.Run(state, func(t *testing.T) {
			root := t.TempDir()
			mkfile(t, filepath.Join(root, "construct/base.manifest"), "")
			const target = ".claude/settings.json"
			if state != "backup only" && state != "unmanaged" {
				if _, err := plan.ApplyManaged(weavefs.OSFS{}, root, []plan.Action{plan.WriteFile{Path: target, Content: "owned"}}, plan.ScopeArtifacts); err != nil {
					t.Fatal(err)
				}
			}
			if state == "edited" || state == "unmanaged" {
				mkfile(t, filepath.Join(root, target), "operator text")
			}
			if state == "missing" {
				if err := os.Remove(filepath.Join(root, target)); err != nil {
					t.Fatal(err)
				}
			}
			backup := filepath.Join(root, plan.DisplacedPath(target))
			if state != "unmanaged" {
				mkfile(t, backup, "prior backup")
			}
			attempts := 0
			fs := protectedSettingsFS{targets: map[string]bool{filepath.Join(root, target): true, backup: true}, attempts: &attempts}
			var out bytes.Buffer
			for range 2 {
				out.Reset()
				actions, err := deferredClaudeSettings(fs, root, nil, &out)
				if err != nil {
					t.Fatal(err)
				}
				want := 1
				if state == "unmanaged" {
					want = 0
				}
				if n := strings.Count(out.String(), "settings.json is stale:"); n != want {
					t.Fatalf("notice=%d want=%d: %s", n, want, &out)
				}
				if _, err := plan.ApplyManaged(fs, root, actions, plan.ScopeArtifacts); err != nil {
					t.Fatal(err)
				}
			}
			if attempts != 0 {
				t.Fatalf("attempted %d protected mutations", attempts)
			}
			if state != "unmanaged" {
				if b, err := os.ReadFile(backup); err != nil || string(b) != "prior backup" {
					t.Fatalf("backup changed: %q %v", b, err)
				}
			}
			// Run the actual ordinary compiler with no declaration: it owns retirement
			// and backup cleanup, including backup-only residue with no inventory ID.
			out.Reset()
			if err := run(weavefs.OSFS{}, root, plan.TargetAll, false, &out); err != nil {
				t.Fatal(err, &out)
			}
			out.Reset()
			if _, err := deferredClaudeSettings(weavefs.OSFS{}, root, nil, &out); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(out.String(), "is stale") {
				t.Fatalf("operator cleanup did not clear notice: %s", &out)
			}
			if state != "unmanaged" {
				if _, err := os.Stat(backup); !os.IsNotExist(err) {
					t.Fatalf("backup not cleaned: %v", err)
				}
			}
		})
	}
}

type protectedSettingsFS struct {
	weavefs.OSFS
	targets  map[string]bool
	attempts *int
}

func (f protectedSettingsFS) deny(path string) error {
	if f.targets[path] {
		*f.attempts++
		return &os.PathError{Op: "protected mutation", Path: path, Err: os.ErrPermission}
	}
	return nil
}
func (f protectedSettingsFS) Rename(a, b string) error {
	if err := f.deny(a); err != nil {
		return err
	}
	if err := f.deny(b); err != nil {
		return err
	}
	return f.OSFS.Rename(a, b)
}
func (f protectedSettingsFS) Remove(p string) error {
	if err := f.deny(p); err != nil {
		return err
	}
	return f.OSFS.Remove(p)
}
func (f protectedSettingsFS) RemoveAll(p string) error {
	if err := f.deny(p); err != nil {
		return err
	}
	return f.OSFS.RemoveAll(p)
}
func (f protectedSettingsFS) WriteFile(p string, b []byte) error {
	if err := f.deny(p); err != nil {
		return err
	}
	return f.OSFS.WriteFile(p, b)
}
func (f protectedSettingsFS) Chmod(p string, m os.FileMode) error {
	if err := f.deny(p); err != nil {
		return err
	}
	return f.OSFS.Chmod(p, m)
}
func (f protectedSettingsFS) MkdirAll(p string) error {
	if err := f.deny(p); err != nil {
		return err
	}
	return f.OSFS.MkdirAll(p)
}
func (f protectedSettingsFS) Symlink(a, p string) error {
	if err := f.deny(p); err != nil {
		return err
	}
	return f.OSFS.Symlink(a, p)
}

func TestDeferredSettingsExcludesAlternateWritersAndClears(t *testing.T) {
	for _, kind := range []string{"write", "symlink", "seed", "seed once", "touch", "mkdir"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			const target = ".claude/settings.json"
			source := filepath.Join(t.TempDir(), "source")
			mkfile(t, source, "-include Makefile.workflow\n")
			var action plan.Action
			switch kind {
			case "write":
				action = plan.WriteFile{Path: target, Content: "new"}
			case "symlink":
				action = plan.Symlink{Dst: target, Src: source}
			case "seed":
				action = plan.Seed{Dst: target, Src: source}
			case "seed once":
				action = plan.SeedOnce{Dst: target, Src: source}
			case "touch":
				action = plan.Touch{Path: target}
			case "mkdir":
				action = plan.Mkdir{Path: target}
			}
			attempts := 0
			fs := protectedSettingsFS{targets: map[string]bool{filepath.Join(root, target): true}, attempts: &attempts}
			check := func(stale bool) {
				t.Helper()
				var out bytes.Buffer
				actions, err := deferredClaudeSettings(fs, root, []plan.Action{plan.Mkdir{Path: ".claude"}, action}, &out)
				if err != nil {
					t.Fatal(err)
				}
				if n := strings.Count(out.String(), "settings.json is stale:"); (stale && n != 1) || (!stale && n != 0) {
					t.Fatalf("stale=%v: %s", stale, &out)
				}
				if _, err := plan.ApplyManaged(fs, root, actions, plan.ScopeArtifacts); err != nil {
					t.Fatal(err)
				}
			}
			check(true)
			if attempts != 0 {
				t.Fatalf("attempted %d protected writes", attempts)
			}
			if _, err := plan.ApplyManaged(weavefs.OSFS{}, root, []plan.Action{action}, plan.ScopeArtifacts); err != nil {
				t.Fatal(err)
			}
			check(false)
			if attempts != 0 {
				t.Fatalf("attempted %d protected writes", attempts)
			}
		})
	}
}

func TestDeferredSettingsMissingSeedSourceDoesNotInventState(t *testing.T) {
	for _, existing := range []bool{false, true} {
		root := t.TempDir()
		const target = ".claude/settings.json"
		if existing {
			mkfile(t, filepath.Join(root, target), "hand owned")
		}
		for _, a := range []plan.Action{plan.Seed{Dst: target, Src: filepath.Join(root, "absent")}, plan.SeedOnce{Dst: target, Src: filepath.Join(root, "absent")}} {
			var out bytes.Buffer
			actions, err := deferredClaudeSettings(weavefs.OSFS{}, root, []plan.Action{a}, &out)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(out.String(), "is stale") {
				t.Fatalf("missing seed source is an ordinary no-op: %s", &out)
			}
			if _, err := plan.ApplyManaged(weavefs.OSFS{}, root, actions, plan.ScopeArtifacts); err != nil {
				t.Fatal(err)
			}
			ids, err := weaveownership.Read(weavefs.OSFS{}, root)
			if err != nil {
				t.Fatal(err)
			}
			for _, id := range ids {
				if id.Path == target {
					t.Fatalf("invented ownership: %+v", id)
				}
			}
		}
	}
}
