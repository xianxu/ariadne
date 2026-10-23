package gitx

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestReadLocalTreeIgnoresReplacementAndExportFilters(t *testing.T) {
	root := t.TempDir()
	testfix.Git(t, root, "init", "-b", "main")
	testfix.Git(t, root, "config", "user.name", "Fixture")
	testfix.Git(t, root, "config", "user.email", "fixture@example.invalid")
	for path, value := range map[string]string{"task.txt": "parent\n", ".gitattributes": "task.txt export-ignore export-subst\n"} {
		if err := os.WriteFile(filepath.Join(root, path), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	testfix.Git(t, root, "add", ".")
	testfix.Git(t, root, "commit", "-m", "parent")
	parent := strings.TrimSpace(testfix.Git(t, root, "rev-parse", "HEAD"))
	if err := os.WriteFile(filepath.Join(root, "task.txt"), []byte("landed solution\n"), 0600); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, root, "commit", "-am", "later solution")
	later := strings.TrimSpace(testfix.Git(t, root, "rev-parse", "HEAD"))
	testfix.Git(t, root, "replace", parent, later)
	files, err := ReadLocalTree(context.Background(), root, parent, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if file.Path == "task.txt" {
			if string(file.Content) != "parent\n" {
				t.Fatalf("replacement or export filter changed parent: %q", file.Content)
			}
			return
		}
	}
	t.Fatal("export-ignore removed a required file")
}

func TestReadLocalTreeMissingPromisorBlobNeverFetches(t *testing.T) {
	root, _ := trunkFixture(t, "seed\n")
	head := strings.TrimSpace(testfix.Git(t, root, "rev-parse", "HEAD"))
	blob := strings.TrimSpace(testfix.Git(t, root, "rev-parse", "HEAD:note.md"))
	// Point the promisor at a sentinel helper: any attempt to repair missing
	// source evidence is an observable violation, even if the helper fails.
	bin := t.TempDir()
	marker := filepath.Join(bin, "contacted")
	if err := os.WriteFile(filepath.Join(bin, "git-remote-benchprobe"), []byte("#!/bin/sh\n: > \"$BENCH_FETCH_MARKER\"\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("BENCH_FETCH_MARKER", marker)
	testfix.Git(t, root, "config", "extensions.partialClone", "origin")
	testfix.Git(t, root, "config", "remote.origin.promisor", "true")
	testfix.Git(t, root, "config", "remote.origin.url", "benchprobe::unavailable")
	if err := os.Remove(filepath.Join(root, ".git", "objects", blob[:2], blob[2:])); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadLocalTree(context.Background(), root, head, nil); err == nil {
		t.Fatal("missing blob became invented source evidence")
	}
	if _, err := ReadLocalBlob(context.Background(), root, blob); err == nil {
		t.Fatal("missing original issue blob became invented evidence")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("source promisor contacted: %v", err)
	}
}

func TestReadLocalBlobUsesExactUnfilteredBytes(t *testing.T) {
	root, _ := trunkFixture(t, "original issue\n")
	blob := strings.TrimSpace(testfix.Git(t, root, "rev-parse", "HEAD:note.md"))
	raw, err := ReadLocalBlob(context.Background(), root, blob)
	if err != nil || string(raw) != "original issue\n" {
		t.Fatalf("immutable original issue: %q %v", raw, err)
	}
	if _, err := ReadLocalBlob(context.Background(), root, "HEAD:note.md"); err == nil {
		t.Fatal("accepted moving object expression in place of a full pin")
	}
}

func TestReadLocalTreeRefusesAmbientObjectRedirection(t *testing.T) {
	root, _ := trunkFixture(t, "source\n")
	head := strings.TrimSpace(testfix.Git(t, root, "rev-parse", "HEAD"))
	// An inherited object store can make an otherwise pinned read observe a
	// different repository. Refuse that caller environment before any read.
	t.Setenv("GIT_OBJECT_DIRECTORY", t.TempDir())
	if _, err := ReadLocalTree(context.Background(), root, head, nil); err == nil || !strings.Contains(err.Error(), "GIT_OBJECT_DIRECTORY") {
		t.Fatalf("ambient source redirection was accepted: %v", err)
	}
}
