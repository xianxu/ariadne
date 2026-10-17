package plan

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// Refuse every mutation of the protected file or its previously displaced copy.
// Atomic publication is covered by its destination rename, not just WriteFile.
type deferredDenyFS struct {
	weavefs.OSFS
	paths    map[string]bool
	attempts *int
}

func (f deferredDenyFS) deny(path string) error {
	if f.paths[path] {
		*f.attempts++
		return &os.PathError{Op: "protected write", Path: path, Err: os.ErrPermission}
	}
	return nil
}
func (f deferredDenyFS) WriteFile(path string, b []byte) error {
	if err := f.deny(path); err != nil {
		return err
	}
	return f.OSFS.WriteFile(path, b)
}
func (f deferredDenyFS) Rename(from, to string) error {
	if err := f.deny(from); err != nil {
		return err
	}
	if err := f.deny(to); err != nil {
		return err
	}
	return f.OSFS.Rename(from, to)
}
func (f deferredDenyFS) Remove(path string) error {
	if err := f.deny(path); err != nil {
		return err
	}
	return f.OSFS.Remove(path)
}
func (f deferredDenyFS) RemoveAll(path string) error {
	if err := f.deny(path); err != nil {
		return err
	}
	return f.OSFS.RemoveAll(path)
}
func (f deferredDenyFS) Chmod(path string, mode os.FileMode) error {
	if err := f.deny(path); err != nil {
		return err
	}
	return f.OSFS.Chmod(path, mode)
}
func (f deferredDenyFS) MkdirAll(path string) error {
	if err := f.deny(path); err != nil {
		return err
	}
	return f.OSFS.MkdirAll(path)
}
func (f deferredDenyFS) Symlink(target, path string) error {
	if err := f.deny(path); err != nil {
		return err
	}
	return f.OSFS.Symlink(target, path)
}

func TestManagedDeferredOutputPreservesEvidenceAndBackup(t *testing.T) {
	const settings = ".claude/settings.json"
	for _, state := range []string{"unchanged", "edited", "missing", "unrecorded missing"} {
		t.Run(state, func(t *testing.T) {
			root := t.TempDir()
			if state != "unrecorded missing" {
				managedApply(t, root, []Action{WriteFile{Path: settings, Content: "original"}}, ScopeArtifacts)
			}
			managedWrite(t, root, DisplacedPath(settings), "prior displaced bytes")
			if state == "edited" {
				managedWrite(t, root, settings, "operator edit")
			}
			if state == "missing" {
				if err := os.Remove(filepath.Join(root, settings)); err != nil {
					t.Fatal(err)
				}
			}
			before, err := readInventory(weavefs.OSFS{}, root)
			if err != nil {
				t.Fatal(err)
			}
			attempts := 0
			fs := deferredDenyFS{paths: map[string]bool{filepath.Join(root, settings): true, filepath.Join(root, DisplacedPath(settings)): true}, attempts: &attempts}
			for range 2 {
				report, err := ApplyManaged(fs, root, []Action{DeferredOutput{Path: settings}, WriteFile{Path: "other", Content: "new"}}, ScopeArtifacts)
				if err != nil {
					t.Fatal(err)
				}
				if len(report.Displaced)+len(report.Abandoned)+len(report.Retired) != 0 {
					t.Fatalf("unexpected report: %+v", report)
				}
				ids, err := readInventory(fs, root)
				if err != nil {
					t.Fatal(err)
				}
				var settingsIDs []outputIdentity
				for _, id := range ids {
					if id.Path == settings {
						settingsIDs = append(settingsIDs, id)
					}
				}
				if !reflect.DeepEqual(settingsIDs, before) {
					t.Fatalf("ownership changed: %+v -> %+v", before, settingsIDs)
				}
			}
			if attempts != 0 {
				t.Fatalf("attempted %d protected mutations", attempts)
			}
			if managedRead(t, root, DisplacedPath(settings)) != "prior displaced bytes" {
				t.Fatal("backup changed")
			}
			if managedRead(t, root, "other") != "new" {
				t.Fatal("unrelated output not updated")
			}
			switch state {
			case "unchanged":
				if managedRead(t, root, settings) != "original" {
					t.Fatal("settings changed")
				}
			case "edited":
				if managedRead(t, root, settings) != "operator edit" {
					t.Fatal("edit changed")
				}
			default:
				managedAbsent(t, root, settings)
			}
			if len(before) > 0 && !strings.Contains(managedRead(t, root, ".gitignore"), "/.claude/settings.json\n") {
				t.Fatal("settings ignore lost")
			}
			// The operator's normal publication still converges, preserving an edit,
			// and later removal retires the output and backup together.
			managedApply(t, root, []Action{WriteFile{Path: settings, Content: "refreshed"}}, ScopeArtifacts)
			if managedRead(t, root, settings) != "refreshed" {
				t.Fatal("did not converge")
			}
			if state == "edited" && managedRead(t, root, DisplacedPath(settings)) != "operator edit" {
				t.Fatal("normal convergence lost edit")
			}
			managedApply(t, root, nil, ScopeArtifacts)
			managedAbsent(t, root, settings)
			managedAbsent(t, root, DisplacedPath(settings))
		})
	}
}

func TestDeferredOutputApplyAndProducedPath(t *testing.T) {
	root := t.TempDir()
	actions := []Action{DeferredOutput{Path: ".claude/settings.json"}}
	if !ProducedPathSet(actions)[".claude/settings.json"] {
		t.Fatal("deferred output not produced")
	}
	if err := Apply(weavefs.OSFS{}, root, actions); err != nil {
		t.Fatal(err)
	}
	managedAbsent(t, root, ".claude")
}

func TestManagedDeferredOutputRefusesUnsafePaths(t *testing.T) {
	for _, path := range []string{"../outside", InventoryPath, "/absolute"} {
		if _, err := ApplyManaged(weavefs.OSFS{}, t.TempDir(), []Action{DeferredOutput{Path: path}}, ScopeArtifacts); err == nil {
			t.Fatalf("accepted %s", path)
		}
	}
	root := t.TempDir()
	if err := os.Symlink(t.TempDir(), filepath.Join(root, ".claude")); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyManaged(weavefs.OSFS{}, root, []Action{DeferredOutput{Path: ".claude/settings.json"}}, ScopeArtifacts); err == nil {
		t.Fatal("accepted symlink parent")
	}
}

func TestManagedDeferredOutputCannotOverlapWritesOrAnotherScope(t *testing.T) {
	const path = ".claude/settings.json"
	for _, writing := range []Action{
		WriteFile{Path: path, Content: "new"},
		WriteFile{Path: path + "/nested", Content: "new"},
		Symlink{Src: t.TempDir(), Dst: ".claude"},
	} {
		root := t.TempDir()
		if _, err := ApplyManaged(weavefs.OSFS{}, root, []Action{DeferredOutput{Path: path}, writing}, ScopeArtifacts); err == nil {
			t.Fatalf("accepted overlapping %T", writing)
		}
		managedAbsent(t, root, ".claude")
	}
	root := t.TempDir()
	managedApply(t, root, []Action{WriteFile{Path: path, Content: "data"}}, ScopeData)
	if _, err := ApplyManaged(weavefs.OSFS{}, root, []Action{DeferredOutput{Path: path}}, ScopeArtifacts); err == nil {
		t.Fatal("accepted another scope's output")
	}
	if managedRead(t, root, path) != "data" {
		t.Fatal("other scope changed")
	}
}
