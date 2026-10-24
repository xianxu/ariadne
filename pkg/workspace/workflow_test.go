package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkflowSuiteHelperResolvesBothLinksAndAnchor(t *testing.T) {
	supplier := t.TempDir()
	consumer := t.TempDir()
	contents := t.TempDir()
	if err := os.WriteFile(filepath.Join(supplier, "Makefile.workflow"), []byte("# fixture\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(supplier, "scripts"), 0700); err != nil {
		t.Fatal(err)
	}
	actual := filepath.Join(t.TempDir(), "test-lock.py")
	if err := os.WriteFile(actual, []byte("# actual\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(actual, filepath.Join(supplier, "scripts", "test-lock.py")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(supplier, "Makefile.workflow"), filepath.Join(consumer, "Makefile.workflow")); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ root, anchor string }{{consumer, ""}, {contents, consumer}} {
		got, err := SuiteHelper(c.root, c.anchor)
		if err != nil || got != actual {
			t.Fatalf("wrong actual helper: %s %v", got, err)
		}
	}
	if _, err := SuiteHelper(t.TempDir(), ""); err == nil {
		t.Fatal("missing workflow hidden")
	}
}
