package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinkRepositoryClonesAndRecordsSource(t *testing.T) {
	workspace := t.TempDir()
	origin := filepath.Join(workspace, "origins", "base.git")
	source := filepath.Join(workspace, "source")
	os.MkdirAll(source, 0755)
	mkfile(t, filepath.Join(source, "construct", "base.manifest"), "# base\n")
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "base"}} {
		c := exec.Command("git", args...)
		c.Dir = source
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git: %s %v", out, err)
		}
	}
	c := exec.Command("git", "clone", "--bare", source, origin)
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git clone: %s %v", out, err)
	}
	leaf := filepath.Join(workspace, "leaf")
	os.MkdirAll(leaf, 0755)
	var out bytes.Buffer
	for range 2 {
		if err := linkRepository(context.Background(), leaf, "file://"+origin, &out); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(filepath.Join(leaf, "construct", "deps"))
	if err != nil {
		t.Fatal(err)
	}
	want := "substrate ../base file://" + origin + "\n"
	if string(data) != want {
		t.Fatalf("got %q, want %q", data, want)
	}
	if _, err := os.Stat(filepath.Join(workspace, "base", "construct", "base.manifest")); err != nil {
		t.Fatal(err)
	}
	// Linking the local checkout also records its origin, without another row.
	if err := linkRepository(context.Background(), leaf, "../base", &out); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(leaf, "construct", "deps"))
	if string(after) != string(data) {
		t.Fatal(string(after))
	}
}

func TestLinkRepositoryRepairsLegacySource(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, "base")
	leaf := filepath.Join(root, "leaf")
	os.MkdirAll(base, 0755)
	os.MkdirAll(leaf, 0755)
	mkfile(t, filepath.Join(base, "construct", "base.manifest"), "# layer\n")
	for _, args := range [][]string{{"init", "-q"}, {"remote", "add", "origin", "https://github.com/example/base.git"}} {
		c := exec.Command("git", args...)
		c.Dir = base
		if err := c.Run(); err != nil {
			t.Fatal(err)
		}
	}
	mkfile(t, filepath.Join(leaf, "construct", "deps"), "# own comment\nsubstrate ../base # base layer\n")
	if err := linkRepository(context.Background(), leaf, "../base", &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(leaf, "construct", "deps"))
	if !strings.Contains(string(data), "substrate ../base https://github.com/example/base.git # base layer") {
		t.Fatal(string(data))
	}
}

func TestLinkRepositoryLocalWithoutOrigin(t *testing.T) {
	root := t.TempDir()
	base, leaf := filepath.Join(root, "base"), filepath.Join(root, "leaf")
	mkfile(t, filepath.Join(base, "construct", "base.manifest"), "# local layer\n")
	if err := os.MkdirAll(leaf, 0755); err != nil {
		t.Fatal(err)
	}
	if err := linkRepository(context.Background(), leaf, "../base", &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(leaf, "construct", "deps"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "substrate ../base\n" {
		t.Fatal(string(data))
	}
}
