package layergraph

import (
	"path/filepath"
	"testing"
)

// writeFile is the shared helper declared in walk_test.go (same package).

// TestMergeByName_LeafWinsAndSkipsSubdirs is the shadow-policy gate: later dirs
// shadow earlier same-named files; missing dirs are skipped; non-matching files
// and subdirectories (e.g. testdata/) are ignored.
func TestMergeByName_LeafWinsAndSkipsSubdirs(t *testing.T) {
	root := t.TempDir()
	foundation := filepath.Join(root, "a")
	leaf := filepath.Join(root, "b")

	writeFile(t, filepath.Join(foundation, "issue.cue"), "// foundation issue")
	writeFile(t, filepath.Join(foundation, "task.cue"), "// foundation task")
	writeFile(t, filepath.Join(leaf, "issue.cue"), "// leaf issue (wins)")
	// noise that must be ignored:
	writeFile(t, filepath.Join(leaf, "README.md"), "not a cue file")
	writeFile(t, filepath.Join(leaf, "testdata", "issue_invalid.cue"), "// fixture, in a subdir")

	dirs := []string{foundation, leaf, filepath.Join(root, "missing")}
	got, err := MergeByName(OSFS{}, dirs, ".cue")
	if err != nil {
		t.Fatalf("MergeByName: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("want 2 entries (issue, task), got %d: %v", len(got), got)
	}
	if want := filepath.Join(leaf, "issue.cue"); got["issue"] != want {
		t.Errorf("issue: leaf should win\n got %q\nwant %q", got["issue"], want)
	}
	if want := filepath.Join(foundation, "task.cue"); got["task"] != want {
		t.Errorf("task: %q != %q", got["task"], want)
	}
	if _, ok := got["issue_invalid"]; ok {
		t.Errorf("testdata/ fixture must not be merged: %v", got)
	}
}

// TestMergeByName_RealReadDirErrorPropagates: a path that exists as a FILE (not
// a dir) is not os.IsNotExist, so it must surface rather than be swallowed.
func TestMergeByName_RealReadDirErrorPropagates(t *testing.T) {
	root := t.TempDir()
	asFile := filepath.Join(root, "notadir")
	writeFile(t, asFile, "x")
	if _, err := MergeByName(OSFS{}, []string{asFile}, ".cue"); err == nil {
		t.Fatal("want error reading a file-as-dir, got nil")
	}
}
