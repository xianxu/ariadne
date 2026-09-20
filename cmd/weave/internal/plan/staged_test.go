package plan

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func stagedFixture(t *testing.T, root string) (string, []Action) {
	t.Helper()
	stage, err := NewGenerationStage(weavefs.OSFS{}, root)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(stage, "vocabulary")
	for _, name := range []string{"SKILL.md", "issue.json", ".source-sha"} {
		managedWrite(t, dir, name, "generated "+name)
	}
	actions, err := StagedActions(weavefs.OSFS{}, root, dir, "construct/generated/vocabulary")
	if err != nil {
		t.Fatal(err)
	}
	return stage, actions
}
func TestStagedOutputsPublishAndRetireAllFiles(t *testing.T) {
	root := t.TempDir()
	stage, actions := stagedFixture(t, root)
	managedAbsent(t, root, "construct/generated/vocabulary")
	if err := RemoveGenerationStage(stage); err != nil {
		t.Fatal(err)
	}
	managedApply(t, root, actions, ScopeArtifacts)
	gi := managedRead(t, root, ".gitignore")
	for _, name := range []string{"SKILL.md", "issue.json", ".source-sha"} {
		if !strings.Contains(gi, "/construct/generated/vocabulary/"+name+"\n") {
			t.Fatal(gi)
		}
	}
	managedApply(t, root, actions, ScopeArtifacts)
	managedApply(t, root, nil, ScopeArtifacts)
	managedAbsent(t, root, "construct/generated/vocabulary")
}
func TestStagedOutputsProtectEditedAndUnownedDestinations(t *testing.T) {
	for _, kind := range []string{"file", "link", "unowned"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			stage, actions := stagedFixture(t, root)
			defer RemoveGenerationStage(stage)
			path := "construct/generated/vocabulary/SKILL.md"
			if kind != "unowned" {
				managedApply(t, root, actions, ScopeArtifacts)
			}
			if kind == "link" {
				if err := os.Remove(filepath.Join(root, path)); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("authored", filepath.Join(root, path)); err != nil {
					t.Fatal(err)
				}
			} else {
				managedWrite(t, root, path, "authored")
			}
			if _, err := ApplyManaged(weavefs.OSFS{}, root, actions, ScopeArtifacts); err == nil {
				t.Fatal("overwrote authored output")
			}
			if kind == "link" {
				if target, err := os.Readlink(filepath.Join(root, path)); err != nil || target != "authored" {
					t.Fatalf("%q %v", target, err)
				}
			} else if managedRead(t, root, path) != "authored" {
				t.Fatal("edited bytes changed")
			}
		})
	}
}
func TestStagedOutputRequiresRegularNonemptySkill(t *testing.T) {
	for _, kind := range []string{"missing", "empty", "directory", "link"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			stage := t.TempDir()
			switch kind {
			case "empty":
				managedWrite(t, stage, "SKILL.md", "")
			case "directory":
				if err := os.Mkdir(filepath.Join(stage, "SKILL.md"), 0755); err != nil {
					t.Fatal(err)
				}
			case "link":
				if err := os.Symlink("outside", filepath.Join(stage, "SKILL.md")); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := StagedActions(weavefs.OSFS{}, root, stage, "construct/generated/demo"); err == nil {
				t.Fatal("accepted invalid staged SKILL.md")
			}
		})
	}
}
func TestStagedLinksRelocateWithoutStageReferences(t *testing.T) {
	root := t.TempDir()
	stage := t.TempDir()
	managedWrite(t, stage, "SKILL.md", "generated")
	managedWrite(t, stage, "nested/target.json", "json")
	if err := os.Symlink("nested/target.json", filepath.Join(stage, "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(stage, "nested/target.json"), filepath.Join(stage, "absolute")); err != nil {
		t.Fatal(err)
	}
	actions, err := StagedActions(weavefs.OSFS{}, root, stage, "construct/generated/demo")
	if err != nil {
		t.Fatal(err)
	}
	managedApply(t, root, actions, ScopeArtifacts)
	for _, name := range []string{"link", "absolute"} {
		target, err := os.Readlink(filepath.Join(root, "construct/generated/demo", name))
		if err != nil {
			t.Fatal(err)
		}
		if target != "nested/target.json" {
			t.Fatalf("stage reference or wrong target: %q", target)
		}
	}
}
func TestGenerationStageRecoveryOnlyReclaimsDeadOwner(t *testing.T) {
	root := t.TempDir()
	live, err := NewGenerationStage(weavefs.OSFS{}, root)
	if err != nil {
		t.Fatal(err)
	}
	defer RemoveGenerationStage(live)
	dead := exec.Command("true")
	if err := dead.Run(); err != nil {
		t.Fatal(err)
	}
	deadStage, err := NewGenerationStage(weavefs.OSFS{}, root)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(deadStage, "owner.json"))
	if err != nil {
		t.Fatal(err)
	}
	var owner staging.Owner
	if err = json.Unmarshal(b, &owner); err != nil {
		t.Fatal(err)
	}
	owner.PID = dead.Process.Pid
	b, _ = json.Marshal(owner)
	managedWrite(t, deadStage, "owner.json", string(b))
	managedWrite(t, deadStage, "vocabulary/partial.json", "partial")
	if err := ReclaimGenerationStages(weavefs.OSFS{}, root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(deadStage); !os.IsNotExist(err) {
		t.Fatalf("dead stage remains: %v", err)
	}
	if _, err := os.Stat(live); err != nil {
		t.Fatal("live stage removed", err)
	}
	stage, actions := stagedFixture(t, root)
	defer RemoveGenerationStage(stage)
	managedApply(t, root, actions, ScopeArtifacts)
	managedApply(t, root, nil, ScopeArtifacts)
	managedAbsent(t, root, "construct/generated/vocabulary")
}

func TestStagedActionsExposeProducedPathsAndIgnores(t *testing.T) {
	root := t.TempDir()
	stage, actions := stagedFixture(t, root)
	defer RemoveGenerationStage(stage)
	paths := ProducedPathSet(actions)
	entries := strings.Join(GeneratedGitignoreEntries(actions), "\n")
	for _, name := range []string{"SKILL.md", "issue.json", ".source-sha"} {
		path := "construct/generated/vocabulary/" + name
		if !paths[path] || !strings.Contains(entries, "/"+path) {
			t.Fatalf("generated action missing from derived paths/ignores: %s", path)
		}
	}
}
func TestStagedPartialApplyRemainsRetirableAfterStageRemoved(t *testing.T) {
	root := t.TempDir()
	stage, actions := stagedFixture(t, root)
	if err := RemoveGenerationStage(stage); err != nil {
		t.Fatal(err)
	}
	fs := managedFailFS{path: filepath.Join(root, "construct/generated/vocabulary/SKILL.md")}
	if _, err := ApplyManaged(fs, root, actions, ScopeArtifacts); err == nil {
		t.Fatal("expected partial apply failure")
	}
	if managedRead(t, root, "construct/generated/vocabulary/.source-sha") == "" {
		t.Fatal("no partial progress")
	}
	managedApply(t, root, nil, ScopeArtifacts)
	managedAbsent(t, root, "construct/generated/vocabulary")
}
func TestStagedReadFailureDoesNotPublish(t *testing.T) {
	root := t.TempDir()
	stage, err := NewGenerationStage(weavefs.OSFS{}, root)
	if err != nil {
		t.Fatal(err)
	}
	defer RemoveGenerationStage(stage)
	output := filepath.Join(stage, "demo")
	managedWrite(t, output, "SKILL.md", "generated")
	if _, err := StagedActions(managedFailFS{path: filepath.Join(output, "SKILL.md"), reads: true}, root, output, "construct/generated/demo"); err == nil {
		t.Fatal("ignored staged read error")
	}
	managedAbsent(t, root, "construct/generated/demo")
}
