package layergraph

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// #289: the one construct/deps reader: an ordinary file within the bound is
// read; a symlink, a FIFO (which would block) or an oversized file is refused;
// a missing file is fs.ErrNotExist.
func TestReadDeclaration(t *testing.T) {
	dir := t.TempDir()
	ordinary := filepath.Join(dir, "deps")
	if err := os.WriteFile(ordinary, []byte("substrate ../ariadne\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if b, err := ReadDeclaration(ordinary, DeclarationLimit); err != nil || string(b) != "substrate ../ariadne\n" {
		t.Fatalf("ordinary: %q %v", b, err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(ordinary, link); err != nil {
		t.Fatal(err)
	}
	fifo := filepath.Join(dir, "fifo")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Fatal(err)
	}
	for name, path := range map[string]string{"symlink": link, "fifo": fifo} {
		if _, err := ReadDeclaration(path, DeclarationLimit); err == nil || !strings.Contains(err.Error(), "ordinary file") {
			t.Errorf("%s: %v, want refused as not an ordinary file", name, err)
		}
	}
	if _, err := ReadDeclaration(ordinary, 4); err == nil || !strings.Contains(err.Error(), "byte limit") {
		t.Errorf("oversized: %v", err)
	}
	if _, err := ReadDeclaration(filepath.Join(dir, "absent"), DeclarationLimit); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("missing: %v", err)
	}
}
