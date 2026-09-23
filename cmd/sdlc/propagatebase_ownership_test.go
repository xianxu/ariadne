package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/weaveownership"
)

func propagationRepo(t *testing.T, root string) {
	t.Helper()
	git(t, "", "init", "-b", "main", root)
	git(t, root, "config", "user.email", "t@example.com")
	git(t, root, "config", "user.name", "t")
}

func propagationWrite(t *testing.T, root, name, body string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestCommitConsumptionPreservesIgnoredWithoutOwnership(t *testing.T) {
	root := t.TempDir()
	propagationRepo(t, root)
	propagationWrite(t, root, ".gitignore", "/authored.txt\n")
	propagationWrite(t, root, "authored.txt", "intentionally tracked\n")
	git(t, root, "add", ".gitignore")
	git(t, root, "add", "-f", "authored.txt")
	git(t, root, "commit", "-qm", "initial")
	changed, err := commitConsumption(root, "test#239")
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("missing inventory caused a consumption commit")
	}
	if got := git(t, root, "ls-files", "--", "authored.txt"); got != "authored.txt" {
		t.Fatalf("unrelated ignored file was untracked: %q", got)
	}
}

func TestPropagateBaseUsesInstalledCompiler(t *testing.T) {
	for _, failure := range []string{"", "compile", "verify-complete"} {
		t.Run("failure="+failure, func(t *testing.T) {
			parent := t.TempDir()
			owner, leaf := filepath.Join(parent, "owner"), filepath.Join(parent, "leaf")
			if err := os.MkdirAll(owner, 0755); err != nil {
				t.Fatal(err)
			}
			propagationRepo(t, owner)
			propagationRepo(t, leaf)
			propagationWrite(t, leaf, "construct/deps", "substrate ../owner\n")
			git(t, leaf, "add", "-A")
			git(t, leaf, "commit", "-qm", "initial")
			bin := filepath.Join(parent, "tools")
			propagationWrite(t, bin, "weave", `#!/bin/sh
printf '%s\n' "$1" >> "$PROPAGATION_EVENTS"
if [ "$1" = "$PROPAGATION_FAIL" ]; then exit 9; fi
case "$1" in
compile) printf ready > "$PROPAGATION_READY" ;;
verify-complete) test -f "$PROPAGATION_READY" ;;
*) exit 10 ;;
esac
`)
			if err := os.Chmod(filepath.Join(bin, "weave"), 0755); err != nil {
				t.Fatal(err)
			}
			events := filepath.Join(parent, "events")
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("PROPAGATION_EVENTS", events)
			t.Setenv("PROPAGATION_READY", filepath.Join(parent, "ready"))
			t.Setenv("PROPAGATION_FAIL", failure)
			var out bytes.Buffer
			err := runPropagateBase(owner, "test#239", false, &out)
			if (err != nil) != (failure != "") {
				t.Fatalf("error=%v output=%s", err, &out)
			}
			data, readErr := os.ReadFile(events)
			if readErr != nil {
				t.Fatal("installed weave never ran:", readErr)
			}
			want := "compile\nverify-complete\n"
			if failure == "compile" {
				want = "compile\n"
			}
			if string(data) != want {
				t.Fatalf("commands=%q, want %q", data, want)
			}
			if got := git(t, leaf, "status", "--porcelain"); got != "" {
				t.Fatalf("unexpected mutation: %s", got)
			}
			if failure != "" && !strings.Contains(out.String(), "FAILED") {
				t.Fatal(out.String())
			}
		})
	}
}

func TestCommitConsumptionUntracksOnlyMatchingOwnedIgnoredPaths(t *testing.T) {
	root := t.TempDir()
	propagationRepo(t, root)
	propagationWrite(t, root, ".gitignore", "/generated*\n/edited.txt\n/authored.txt\n/chmod.txt\n/kept.txt\n!/kept.txt\n/construct/generated/weave/\n")
	var ids []weaveownership.Identity
	mode := os.FileMode(0644)
	for _, name := range []string{"generated.txt", "generated[1].txt", "generated space.txt", "edited.txt", "chmod.txt", "kept.txt"} {
		propagationWrite(t, root, name, "generated\n")
		ids = append(ids, weaveownership.Identity{Path: name, Scope: weaveownership.ScopeArtifacts, Kind: "file", Value: weaveownership.Digest([]byte("generated\n")), Mode: &mode})
	}
	propagationWrite(t, root, "authored.txt", "intentionally tracked\n")
	propagationWrite(t, root, "generated1.txt", "unowned near-match\n")
	propagationWrite(t, root, "edited.txt", "authored replacement\n")
	propagationWrite(t, root, "source.txt", "symlink source\n")
	if err := os.Chmod(filepath.Join(root, "chmod.txt"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("source.txt", filepath.Join(root, "generated-link")); err != nil {
		t.Fatal(err)
	}
	ids = append(ids, weaveownership.Identity{Path: "generated-link", Scope: weaveownership.ScopeData, Kind: "link", Value: "source.txt"})
	git(t, root, "add", "-A")
	git(t, root, "add", "-f", "--", "generated.txt", "generated[1].txt", "generated space.txt", "generated1.txt", "generated-link", "edited.txt", "authored.txt", "chmod.txt")
	git(t, root, "commit", "-qm", "initial authored and legacy generated files")
	data, err := json.Marshal(weaveownership.Inventory{Version: 1, Outputs: ids})
	if err != nil {
		t.Fatal(err)
	}
	propagationWrite(t, root, weaveownership.InventoryPath, string(data))
	changed, err := commitConsumption(root, "test#239")
	if err != nil || !changed {
		t.Fatalf("changed=%v error=%v", changed, err)
	}
	tracked := git(t, root, "ls-files", "-z")
	trackedSet := map[string]bool{}
	for _, name := range strings.Split(tracked, "\x00") {
		trackedSet[name] = true
	}
	for _, name := range []string{"generated.txt", "generated[1].txt", "generated space.txt", "generated-link"} {
		if trackedSet[name] {
			t.Errorf("owned generated path still tracked: %s", name)
		}
		if _, err := os.Lstat(filepath.Join(root, name)); err != nil {
			t.Errorf("working file removed: %s: %v", name, err)
		}
	}
	for _, name := range []string{"generated1.txt", "edited.txt", "authored.txt", "chmod.txt", "kept.txt", "source.txt"} {
		if !trackedSet[name] {
			t.Errorf("authored or locally retained path untracked: %s", name)
		}
	}
	if changed, err := commitConsumption(root, "test#239"); err != nil || changed {
		t.Fatalf("repeat changed=%v error=%v", changed, err)
	}
}

func TestCommitConsumptionInvalidInventoryDoesNotMutateIndex(t *testing.T) {
	root := t.TempDir()
	propagationRepo(t, root)
	propagationWrite(t, root, ".gitignore", "/authored.txt\n/construct/generated/weave/\n")
	propagationWrite(t, root, "authored.txt", "authored\n")
	git(t, root, "add", "-A")
	git(t, root, "add", "-f", "authored.txt")
	git(t, root, "commit", "-qm", "initial")
	propagationWrite(t, root, weaveownership.InventoryPath, "{truncated")
	if _, err := commitConsumption(root, "test#239"); err == nil {
		t.Fatal("invalid inventory accepted")
	}
	if got := git(t, root, "status", "--porcelain"); got != "" {
		t.Fatalf("invalid inventory changed index: %s", got)
	}
	if got := git(t, root, "ls-files", "--", "authored.txt"); got != "authored.txt" {
		t.Fatal("authored file untracked")
	}
}

func TestPropagatePreviewNeedsNoWeaveAndMissingCompilerLeavesCleanRepo(t *testing.T) {
	parent := t.TempDir()
	owner, leaf := filepath.Join(parent, "owner"), filepath.Join(parent, "leaf")
	if err := os.MkdirAll(owner, 0755); err != nil {
		t.Fatal(err)
	}
	propagationRepo(t, owner)
	propagationRepo(t, leaf)
	propagationWrite(t, leaf, "construct/deps", "substrate ../owner\n")
	git(t, leaf, "add", "-A")
	git(t, leaf, "commit", "-qm", "initial")
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(parent, "only-git")
	if err := os.MkdirAll(bin, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(gitPath, filepath.Join(bin, "git")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	var out bytes.Buffer
	if err := runPropagateBase(owner, "test#239", true, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "weave compile") {
		t.Fatal(out.String())
	}
	out.Reset()
	if err := runPropagateBase(owner, "test#239", false, &out); err == nil || !strings.Contains(out.String(), "find installed weave") {
		t.Fatalf("error=%v output=%s", err, &out)
	}
	if got := git(t, leaf, "status", "--porcelain"); got != "" {
		t.Fatalf("missing compiler mutated repository: %s", got)
	}
}
