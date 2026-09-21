package plan

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func managedWrite(t *testing.T, root, path, text string) {
	t.Helper()
	p := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
}
func managedRead(t *testing.T, root, path string) string {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(root, path))
	if e != nil {
		t.Fatal(e)
	}
	return string(b)
}
func managedApply(t *testing.T, root string, actions []Action, scope OwnershipScope) {
	t.Helper()
	if _, e := ApplyManaged(weavefs.OSFS{}, root, actions, scope); e != nil {
		t.Fatal(e)
	}
}
func managedAbsent(t *testing.T, root, path string) {
	t.Helper()
	if _, e := os.Lstat(filepath.Join(root, path)); !os.IsNotExist(e) {
		t.Fatalf("%s remains: %v", path, e)
	}
}

func TestManagedRetiresWithoutSourceAndPreservesReplacement(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(t.TempDir(), "gone")
	managedApply(t, root, []Action{WriteFile{Path: "AGENTS.md", Content: "generated"}, Symlink{Src: source, Dst: "helper"}, WriteFile{Path: "edited", Content: "generated"}}, ScopeArtifacts)
	managedWrite(t, root, "edited", "authored")
	managedApply(t, root, nil, ScopeArtifacts)
	managedAbsent(t, root, "AGENTS.md")
	managedAbsent(t, root, "helper")
	if got := managedRead(t, root, "edited"); got != "authored" {
		t.Fatal(got)
	}
	if got := managedRead(t, root, ".gitignore"); strings.Contains(got, "/edited\n") || strings.Contains(got, "/AGENTS.md\n") {
		t.Fatal(got)
	}
}
func TestManagedScopesPreserveOtherInventoryAndIgnores(t *testing.T) {
	root := t.TempDir()
	source := t.TempDir()
	managedApply(t, root, []Action{WriteFile{Path: "AGENTS.md", Content: "generated"}}, ScopeArtifacts)
	managedApply(t, root, []Action{Symlink{Src: source, Dst: "data/first"}, Symlink{Src: source, Dst: "data/second"}}, ScopeData)
	managedApply(t, root, nil, ScopeData)
	managedAbsent(t, root, "data/first")
	managedAbsent(t, root, "data/second")
	if managedRead(t, root, "AGENTS.md") != "generated" {
		t.Fatal("artifact changed")
	}
	if !strings.Contains(managedRead(t, root, ".gitignore"), "/AGENTS.md\n") {
		t.Fatal("artifact ignore removed")
	}
	managedApply(t, root, nil, ScopeArtifacts)
	managedAbsent(t, root, "AGENTS.md")
}
func TestManagedDataRejectsAuthoredDestinations(t *testing.T) {
	for _, kind := range []string{"file", "directory", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			dest := filepath.Join(root, "data")
			switch kind {
			case "file":
				managedWrite(t, root, "data", "mine")
			case "directory":
				if e := os.Mkdir(dest, 0755); e != nil {
					t.Fatal(e)
				}
			case "symlink":
				if e := os.Symlink("authored", dest); e != nil {
					t.Fatal(e)
				}
			}
			if _, e := ApplyManaged(weavefs.OSFS{}, root, []Action{Symlink{Src: t.TempDir(), Dst: "data"}}, ScopeData); e == nil {
				t.Fatal("accepted authored destination")
			}
			if _, e := os.Lstat(dest); e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestManagedRetainedEntrypointsAndEditedLink(t *testing.T) {
	root := t.TempDir()
	source := t.TempDir()
	managedApply(t, root, []Action{WriteFile{Path: "keep", Content: "generated"}, Symlink{Src: source, Dst: "link"}}, ScopeArtifacts)
	if e := os.Remove(filepath.Join(root, "link")); e != nil {
		t.Fatal(e)
	}
	if e := os.Symlink("authored", filepath.Join(root, "link")); e != nil {
		t.Fatal(e)
	}
	managedApply(t, root, []Action{Touch{Path: "keep"}, Mkdir{Path: "scaffold"}}, ScopeArtifacts)
	if managedRead(t, root, "keep") != "generated" {
		t.Fatal("retained path changed")
	}
	if s, e := os.Readlink(filepath.Join(root, "link")); e != nil || s != "authored" {
		t.Fatalf("link %q %v", s, e)
	}
	gi := managedRead(t, root, ".gitignore")
	if strings.Contains(gi, "/keep\n") || strings.Contains(gi, "/scaffold") {
		t.Fatal(gi)
	}
}

type managedFailFS struct {
	weavefs.OSFS
	path   string
	reads  bool
	atomic bool
}

func (f managedFailFS) WriteFile(p string, b []byte) error {
	if f.path != "" && filepath.Base(p) == filepath.Base(f.path) {
		return errors.New("injected write failure")
	}
	return f.OSFS.WriteFile(p, b)
}
func (f managedFailFS) ReadFile(p string) ([]byte, error) {
	if f.reads && p == f.path {
		return nil, errors.New("injected read failure")
	}
	return f.OSFS.ReadFile(p)
}
func (f managedFailFS) Rename(old, p string) error {
	if f.atomic {
		return errors.New("injected atomic failure")
	}
	return f.OSFS.Rename(old, p)
}

func TestManagedPartialApplyRetainsOldAndNewIdentities(t *testing.T) {
	root := t.TempDir()
	managedApply(t, root, []Action{WriteFile{Path: "first", Content: "old"}, WriteFile{Path: "second", Content: "old"}}, ScopeArtifacts)
	fs := managedFailFS{path: filepath.Join(root, "second")}
	if _, e := ApplyManaged(fs, root, []Action{WriteFile{Path: "first", Content: "new"}, WriteFile{Path: "second", Content: "new"}}, ScopeArtifacts); e == nil {
		t.Fatal("expected injected failure")
	}
	if managedRead(t, root, "first") != "new" || managedRead(t, root, "second") != "old" {
		t.Fatal("wrong partial state")
	}
	managedApply(t, root, nil, ScopeArtifacts)
	managedAbsent(t, root, "first")
	managedAbsent(t, root, "second")
}
func TestManagedPrepareFailureDoesNotApply(t *testing.T) {
	root := t.TempDir()
	if _, e := ApplyManaged(managedFailFS{atomic: true}, root, []Action{WriteFile{Path: "out", Content: "x"}}, ScopeArtifacts); e == nil {
		t.Fatal("expected failure")
	}
	managedAbsent(t, root, "out")
}
func TestManagedReadErrorsAndMalformedStateFailBeforeApply(t *testing.T) {
	for _, state := range []string{"inventory", "ignore", "inventory-read", "ignore-read"} {
		t.Run(state, func(t *testing.T) {
			root := t.TempDir()
			fs := managedFailFS{}
			switch state {
			case "inventory":
				managedWrite(t, root, InventoryPath, "{broken")
			case "ignore":
				managedWrite(t, root, ".gitignore", "# BEGIN weave generated\nunterminated\n")
			case "inventory-read":
				fs.reads = true
				fs.path = filepath.Join(root, InventoryPath)
			case "ignore-read":
				fs.reads = true
				fs.path = filepath.Join(root, ".gitignore")
			}
			if _, e := ApplyManaged(fs, root, []Action{WriteFile{Path: "out", Content: "x"}}, ScopeArtifacts); e == nil {
				t.Fatal("expected malformed/read failure")
			}
			managedAbsent(t, root, "out")
		})
	}
}
func TestManagedIgnoreMigrationAndLocalNegations(t *testing.T) {
	root := t.TempDir()
	managedWrite(t, root, ".gitignore", "# local\n/AGENTS.md\n/.agents/skills/\n/construct/generated/\n/custom\n!/AGENTS.md\n")
	managedApply(t, root, []Action{WriteFile{Path: "AGENTS.md", Content: "x"}, Symlink{Src: t.TempDir(), Dst: ".agents/skills/one"}, Mkdir{Path: ".agents/skills/local"}}, ScopeArtifacts)
	got := managedRead(t, root, ".gitignore")
	if !strings.HasSuffix(got, "# local\n/custom\n!/AGENTS.md\n") {
		t.Fatal(got)
	}
	if strings.Contains(got, "/.agents/skills/\n") || strings.Contains(got, "/construct/generated/\n") || strings.Contains(got, "/.agents/skills/local") {
		t.Fatal(got)
	}
	if !strings.Contains(got, "/.agents/skills/one\n") {
		t.Fatal(got)
	}
	managedApply(t, root, []Action{WriteFile{Path: "AGENTS.md", Content: "x"}, Symlink{Src: t.TempDir(), Dst: ".agents/skills/one"}}, ScopeArtifacts)
	managedApply(t, root, nil, ScopeArtifacts)
	if got := managedRead(t, root, ".gitignore"); !strings.Contains(got, "!/AGENTS.md\n") || strings.Contains(got, "/.agents/skills/one\n") {
		t.Fatal(got)
	}
}
func TestManagedMergeIdentityAndDynamicFiles(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(t.TempDir(), "settings.json")
	if e := os.WriteFile(source, []byte(`{"a":1}`), 0644); e != nil {
		t.Fatal(e)
	}
	stage := t.TempDir()
	managedWrite(t, stage, "SKILL.md", "generated skill")
	actions, e := StagedActions(weavefs.OSFS{}, root, stage, "construct/generated/demo")
	if e != nil {
		t.Fatal(e)
	}
	actions = append(actions, MergeSettings{Sources: []string{source}, Target: ".claude/settings.json"})
	managedApply(t, root, actions, ScopeArtifacts)
	managedWrite(t, root, "construct/generated/demo/authored", "mine")
	managedApply(t, root, nil, ScopeArtifacts)
	managedAbsent(t, root, ".claude/settings.json")
	managedAbsent(t, root, "construct/generated/demo/SKILL.md")
	if managedRead(t, root, "construct/generated/demo/authored") != "mine" {
		t.Fatal("authored file lost")
	}
}

func TestManagedProtectsEditedActiveOutput(t *testing.T) {
	root := t.TempDir()
	managedApply(t, root, []Action{WriteFile{Path: "out", Content: "old"}}, ScopeArtifacts)
	managedWrite(t, root, "out", "authored")
	if _, e := ApplyManaged(weavefs.OSFS{}, root, []Action{WriteFile{Path: "out", Content: "new"}}, ScopeArtifacts); e == nil {
		t.Fatal("overwrote authored replacement")
	}
	if managedRead(t, root, "out") != "authored" {
		t.Fatal("edited output changed")
	}
}
func TestManagedRejectsEscapingAndReservedInventoryPaths(t *testing.T) {
	for _, path := range []string{"../outside", "/absolute", "construct/generated/weave", "construct/generated", ".gitignore"} {
		t.Run(path, func(t *testing.T) {
			root := t.TempDir()
			if _, e := ApplyManaged(weavefs.OSFS{}, root, []Action{Symlink{Src: t.TempDir(), Dst: path}}, ScopeArtifacts); e == nil {
				t.Fatal("accepted unsafe path")
			}
		})
	}
}
func TestManagedRejectsSymlinkParent(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	managedWrite(t, outside, "mine", "authored")
	if e := os.Symlink(outside, filepath.Join(root, "nested")); e != nil {
		t.Fatal(e)
	}
	if _, e := ApplyManaged(weavefs.OSFS{}, root, []Action{WriteFile{Path: "nested/mine", Content: "new"}}, ScopeArtifacts); e == nil {
		t.Fatal("followed parent symlink")
	}
	if managedRead(t, outside, "mine") != "authored" {
		t.Fatal("outside changed")
	}
}
func TestManagedIgnoreMalformedBlocks(t *testing.T) {
	for _, input := range []string{ignoreEnd + "\n", ignoreBegin + "\n" + ignoreBegin + "\n" + ignoreEnd + "\n", ignoreBegin + "\n" + ignoreEnd + "\n" + ignoreBegin + "\n" + ignoreEnd + "\n"} {
		if _, e := managedIgnoreText(input, nil); e == nil {
			t.Fatalf("accepted %q", input)
		}
	}
}
func TestPruneGeneratedPreservesInventory(t *testing.T) {
	root := t.TempDir()
	managedWrite(t, root, InventoryPath, "inventory")
	if _, e := PruneGenerated(weavefs.OSFS{}, root, nil); e != nil {
		t.Fatal(e)
	}
	if managedRead(t, root, InventoryPath) != "inventory" {
		t.Fatal("inventory removed")
	}
}

func TestManagedInventoryRejectsInvalidEntries(t *testing.T) {
	for _, body := range []string{
		`{"version":2,"outputs":[]}`,
		`{"version":1,"outputs":[{"path":"../outside","scope":"artifacts","kind":"link","value":"target"}]}`,
		`{"version":1,"outputs":[{"path":"out","scope":"unknown","kind":"link","value":"target"}]}`,
		`{"version":1,"outputs":[{"path":"out","scope":"data","kind":"file","value":"bad-digest"}]}`,
	} {
		root := t.TempDir()
		managedWrite(t, root, InventoryPath, body)
		if _, e := ApplyManaged(weavefs.OSFS{}, root, nil, ScopeArtifacts); e == nil {
			t.Fatalf("accepted %s", body)
		}
		if managedRead(t, root, InventoryPath) != body {
			t.Fatal("rewrote invalid inventory")
		}
	}
}

func TestManagedGeneratedRetirementRemovesOnlyEmptyDirs(t *testing.T) {
	root := t.TempDir()
	managedApply(t, root, []Action{WriteFile{Path: "construct/generated/gone/SKILL.md", Content: "generated"}}, ScopeArtifacts)
	managedApply(t, root, nil, ScopeArtifacts)
	managedAbsent(t, root, "construct/generated/gone")
}
func TestGeneratedGitignoreEntriesAreExact(t *testing.T) {
	entries := GeneratedGitignoreEntries([]Action{Symlink{Dst: "skills/a*[x]"}, WriteFile{Path: "AGENTS.md"}, MergeSettings{Target: "settings.json"}, Seed{Dst: "Makefile"}, Touch{Path: "notes"}, Mkdir{Path: "skills"}})
	got := strings.Join(entries, "\n")
	if !strings.Contains(got, "/skills/a\\*\\[x\\]") || !strings.Contains(got, "/settings.json") || strings.Contains(got, "Makefile") || strings.Contains(got, "notes") || strings.Contains(got, "/skills\n") {
		t.Fatal(got)
	}
}

type managedAtomicStepFS struct {
	weavefs.OSFS
	calls int
	fail  int
}

func (f *managedAtomicStepFS) Rename(old, path string) error {
	f.calls++
	if f.calls == f.fail {
		return errors.New("injected final inventory failure")
	}
	return f.OSFS.Rename(old, path)
}
func TestManagedFinalSaveFailureLeavesRecoverableInventory(t *testing.T) {
	root := t.TempDir()
	fs := &managedAtomicStepFS{fail: 4}
	if _, e := ApplyManaged(fs, root, []Action{WriteFile{Path: "out", Content: "generated"}}, ScopeArtifacts); e == nil {
		t.Fatal("expected final save failure")
	}
	if managedRead(t, root, "out") != "generated" {
		t.Fatal("output missing")
	}
	managedApply(t, root, nil, ScopeArtifacts)
	managedAbsent(t, root, "out")
}
func TestManagedRejectsCrossScopeClaimOfAbsentOutput(t *testing.T) {
	root := t.TempDir()
	managedApply(t, root, []Action{WriteFile{Path: "out", Content: "generated"}}, ScopeArtifacts)
	if e := os.Remove(filepath.Join(root, "out")); e != nil {
		t.Fatal(e)
	}
	if _, e := ApplyManaged(weavefs.OSFS{}, root, []Action{Symlink{Dst: "out", Src: t.TempDir()}}, ScopeData); e == nil {
		t.Fatal("data claimed artifact scope output")
	}
	managedAbsent(t, root, "out")
}
