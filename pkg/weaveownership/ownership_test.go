package weaveownership

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func put(t *testing.T, root, path, content string) {
	t.Helper()
	name := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
func record(t *testing.T, root string, ids []Identity) {
	t.Helper()
	b, err := json.Marshal(Inventory{Version: Version, Outputs: ids})
	if err != nil {
		t.Fatal(err)
	}
	put(t, root, InventoryPath, string(b))
}
func file(path, content string) Identity {
	return Identity{Path: path, Scope: ScopeArtifacts, Kind: "file", Value: Digest([]byte(content))}
}

func TestMatchingPathsUsesCurrentFileLinkAndModeProof(t *testing.T) {
	root := t.TempDir()
	put(t, root, "good", "generated")
	put(t, root, "edited", "authored")
	put(t, root, "chmodded", "generated")
	if err := os.Chmod(filepath.Join(root, "chmodded"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing-source", filepath.Join(root, "data")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("authored", filepath.Join(root, "changed-link")); err != nil {
		t.Fatal(err)
	}
	mode := os.FileMode(0644)
	changedMode := file("chmodded", "generated")
	changedMode.Mode = &mode
	record(t, root, []Identity{file("good", "generated"), file("good", "generated"), file("edited", "generated"), file("missing", "generated"), changedMode, {Path: "data", Scope: ScopeData, Kind: "link", Value: "missing-source"}, {Path: "changed-link", Scope: ScopeArtifacts, Kind: "link", Value: "generated-target"}})
	paths, err := MatchingPaths(OSReader{}, root)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"data", "good"}; !reflect.DeepEqual(paths, want) {
		t.Fatalf("paths = %v, want %v", paths, want)
	}
	// Old mode-less inventory remains content-based, as before extraction.
	record(t, root, []Identity{file("chmodded", "generated")})
	paths, err = MatchingPaths(OSReader{}, root)
	if err != nil || !reflect.DeepEqual(paths, []string{"chmodded"}) {
		t.Fatalf("legacy proof changed: %v %v", paths, err)
	}
}
func TestMatchingPathsMissingInventoryDoesNotCreateState(t *testing.T) {
	root := t.TempDir()
	paths, err := MatchingPaths(OSReader{}, root)
	if err != nil || len(paths) != 0 {
		t.Fatalf("%v %v", paths, err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("reader mutated tree: %v %v", entries, err)
	}
}
func TestMatchingPathsRejectsMalformedInventory(t *testing.T) {
	for _, body := range []string{`{broken`, `{"version":2}`, `{"version":1,"outputs":[{"path":"../outside","kind":"link","scope":"data","value":"x"}]}`, `{"version":1,"outputs":[{"path":"out","kind":"file","scope":"artifacts","value":"not-sha256"}]}`, `{"version":1,"outputs":[{"path":"out","kind":"link","scope":"data","value":"x","mode":420}]}`} {
		root := t.TempDir()
		put(t, root, InventoryPath, body)
		if _, err := MatchingPaths(OSReader{}, root); err == nil {
			t.Fatalf("accepted %s", body)
		}
	}
}
func TestMatchingPathsRejectsSymlinkParents(t *testing.T) {
	for _, part := range []string{"inventory", "output"} {
		t.Run(part, func(t *testing.T) {
			root := t.TempDir()
			peer := t.TempDir()
			if part == "inventory" {
				if err := os.Symlink(peer, filepath.Join(root, "construct")); err != nil {
					t.Fatal(err)
				}
			} else {
				put(t, peer, "file", "generated")
				record(t, root, []Identity{file("nested/file", "generated")})
				if err := os.Symlink(peer, filepath.Join(root, "nested")); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := MatchingPaths(OSReader{}, root); err == nil {
				t.Fatal("followed authored parent symlink")
			}
		})
	}
}

type failedReader struct {
	OSReader
	path string
}

func (f failedReader) ReadFile(path string) ([]byte, error) {
	if path == f.path {
		return nil, errors.New("injected read failure")
	}
	return f.OSReader.ReadFile(path)
}
func TestMatchingPathsSurfacesReadErrors(t *testing.T) {
	for _, path := range []string{InventoryPath, "out"} {
		t.Run(path, func(t *testing.T) {
			root := t.TempDir()
			put(t, root, "out", "generated")
			record(t, root, []Identity{file("out", "generated")})
			if _, err := MatchingPaths(failedReader{path: filepath.Join(root, path)}, root); err == nil {
				t.Fatal("read failure hidden")
			}
		})
	}
}
