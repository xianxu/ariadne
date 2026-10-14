package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadReviewPreparationPinned(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init")
	git(t, root, "config", "user.email", "test@example.com")
	git(t, root, "config", "user.name", "Test")
	git(t, root, "commit", "--allow-empty", "-m", "empty")
	absent := git(t, root, "rev-parse", "HEAD")
	if got, err := readReviewPreparation(root, absent); err != nil || got != nil {
		t.Fatalf("absent=%+v,%v", got, err)
	}
	if err := os.Mkdir(filepath.Join(root, ".sdlc"), 0755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(root, ".sdlc", "review-prepare.json")
	write := func(s string) {
		t.Helper()
		if err := os.WriteFile(p, []byte(s), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"version":1,"command":["committed"]}`)
	git(t, root, "add", ".")
	git(t, root, "commit", "-m", "declaration")
	head := git(t, root, "rev-parse", "HEAD")
	write(`{"version":1,"command":["mutable"]}`)
	got, err := readReviewPreparation(root, head)
	if err != nil || got == nil || got.Command[0] != "committed" {
		t.Fatalf("pinned=%+v,%v", got, err)
	}
	if got, err := readReviewPreparation(root, absent); err != nil || got != nil {
		t.Fatalf("untracked should not override absent=%+v,%v", got, err)
	}
	if _, err := readReviewPreparation(root, "missing-revision"); err == nil {
		t.Fatal("bad revision accepted")
	}
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target", p); err != nil {
		t.Fatal(err)
	}
	git(t, root, "add", ".")
	git(t, root, "commit", "-m", "symlink")
	if _, err := readReviewPreparation(root, git(t, root, "rev-parse", "HEAD")); err == nil {
		t.Fatal("symlink accepted")
	}
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(p, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "child"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	git(t, root, "add", ".")
	git(t, root, "commit", "-m", "directory")
	if _, err := readReviewPreparation(root, git(t, root, "rev-parse", "HEAD")); err == nil {
		t.Fatal("tree accepted")
	}
}
