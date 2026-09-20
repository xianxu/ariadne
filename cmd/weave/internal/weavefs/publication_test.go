package weavefs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
)

func TestPublishRejectsFinalAndStorageSymlinkParents(t *testing.T) {
	for _, name := range []string{"final", "storage"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			peer := t.TempDir()
			target := filepath.Join(root, "nested", "out")
			link := filepath.Join(root, "nested")
			if name == "storage" {
				link = filepath.Join(root, "construct")
			}
			if err := os.Symlink(peer, link); err != nil {
				t.Fatal(err)
			}
			if err := Publish(OSFS{}, root, target, []byte("new"), nil); err == nil {
				t.Fatal("published through symlink parent")
			}
			entries, err := os.ReadDir(peer)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Fatal("publication wrote into peer", entries)
			}
		})
	}
}
func TestPublishReplacesSymlinkWithoutChangingTarget(t *testing.T) {
	root := t.TempDir()
	peer := filepath.Join(t.TempDir(), "authored")
	if err := os.WriteFile(peer, []byte("mine"), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "out")
	if err := os.Symlink(peer, path); err != nil {
		t.Fatal(err)
	}
	if err := Publish(OSFS{}, root, path, []byte("generated"), nil); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(peer)
	if err != nil || string(data) != "mine" {
		t.Fatalf("target changed: %q %v", data, err)
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		t.Fatalf("not materialized: %v %v", info, err)
	}
}
func TestPublishPreservesExistingOrdinaryMode(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "out")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Publish(OSFS{}, root, path, []byte("new"), nil); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("ordinary mode changed: %o", info.Mode().Perm())
	}
	entries, err := os.ReadDir(filepath.Join(root, staging.RootRel))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatal("successful publication left a stage", entries)
	}
}
