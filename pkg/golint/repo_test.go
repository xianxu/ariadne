package golint

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestTrackedGoFilesPassTheCatalog runs every catalogued rule over every
// tracked Go file (#407). A finding is fixed, or suppressed on its line with
// the reason it does not apply.
func TestTrackedGoFilesPassTheCatalog(t *testing.T) {
	root := moduleRoot(t)
	out, err := exec.Command("git", "-C", root, "ls-files", "-z", "--", "*.go").Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	start := time.Now()
	scanned := 0
	for _, rel := range strings.Split(strings.TrimRight(string(out), "\x00"), "\x00") {
		if rel == "" || strings.HasPrefix(rel, "workshop/") || strings.Contains("/"+rel, "/testdata/") {
			continue // archived prose and fixtures are not the module's code
		}
		src, err := os.ReadFile(filepath.Join(root, rel))
		if os.IsNotExist(err) {
			continue // deleted in the working tree
		}
		if err != nil {
			t.Fatal(err)
		}
		scanned++
		findings, err := Check(rel, src, Catalog)
		if err != nil {
			t.Errorf("%s: %v", rel, err)
			continue
		}
		for _, f := range findings {
			t.Errorf("%s:%s", rel, f)
		}
	}
	if scanned == 0 {
		t.Fatal("no tracked Go files found; the file filter is broken")
	}
	t.Logf("checked %d Go files against %d rules in %s", scanned, len(Catalog), time.Since(start).Round(time.Millisecond))
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no go.mod above the test directory")
		}
		dir = parent
	}
}
