package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/sidedoc"
	"go.yaml.in/yaml/v3"
)

// Exercise the production snapshot, plan, copy and cleanup together. Checking
// resolved targets catches a valid-looking relative path left in old coordinates.
func TestSessionMovePreservesReferenceTargets(t *testing.T) {
	for _, source := range []string{"session", "side-document"} {
		for _, target := range []string{"session", "side-root", "directory", "empty-directory", "file", "stationary-file"} {
			t.Run(source+"/"+target, func(t *testing.T) {
				root, old, f := sessionMoveFixture(t)
				oldSide := sidedoc.SideDir(old)
				writeFile(t, filepath.Join(root, oldSide, "assets/picture.bin"), "picture")
				if err := os.MkdirAll(filepath.Join(root, oldSide, "empty"), 0755); err != nil {
					t.Fatal(err)
				}
				writeFile(t, filepath.Join(root, "docs/stationary.md"), "stationary")
				targets := map[string][2]string{
					"session":         {old, "session.md"},
					"side-root":       {oldSide, "session.side"},
					"directory":       {oldSide + "/assets", "session.side/assets"},
					"empty-directory": {oldSide + "/empty", "session.side/empty"},
					"file":            {oldSide + "/assets/picture.bin", "session.side/assets/picture.bin"},
					"stationary-file": {"docs/stationary.md", "docs/stationary.md"},
				}
				from, to := old, "session.md"
				metadata := "type: session-sync\ncouch-owner: ops:0\noperator: Op\n"
				if source == "side-document" {
					from, to = oldSide+"/01-a.md", "session.side/01-a.md"
					metadata = "type: side-doc\n"
				}
				pair := targets[target]
				rel := sidedoc.RelPath(path.Dir(from), pair[0])
				writeFile(t, filepath.Join(root, from), targetReferenceDocument(metadata, rel))
				if err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, nil); err != nil {
					t.Fatal(err)
				}
				assertMovedReferenceTargets(t, root, to, pair[1])
			})
		}
	}
}

func TestSessionMovePreservesEmptySideRoot(t *testing.T) {
	root, old, f := sessionMoveFixture(t)
	oldSide := sidedoc.SideDir(old)
	if err := os.RemoveAll(filepath.Join(root, oldSide)); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, oldSide), 0755); err != nil {
		t.Fatal(err)
	}
	delete(f.registered, filepath.Join(root, oldSide, "01-a.md"))
	writeFile(t, filepath.Join(root, old), targetReferenceDocument("type: session-sync\ncouch-owner: ops:0\noperator: Op\n", path.Base(oldSide)))
	if err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, nil); err != nil {
		t.Fatal(err)
	}
	assertMovedReferenceTargets(t, root, "session.md", "session.side")
}

func TestSessionMoveRejectsEmptyDirectoryFileCollision(t *testing.T) {
	root, old, f := sessionMoveFixture(t)
	if err := os.Mkdir(filepath.Join(root, sidedoc.SideDir(old), "empty"), 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "session.side/empty"), "destination file")
	before, err := os.ReadFile(filepath.Join(root, old))
	if err != nil {
		t.Fatal(err)
	}
	if err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, nil); err == nil {
		t.Fatal("accepted a regular file at an empty directory destination")
	}
	after, err := os.ReadFile(filepath.Join(root, old))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("source changed: %v", err)
	}
	if f.file != filepath.Join(root, old) {
		t.Fatalf("registration switched to %s", f.file)
	}
	if _, err := os.Stat(filepath.Join(root, ".gitignore")); !os.IsNotExist(err) {
		t.Fatalf("collision performed ignore effects: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "session.md")); !os.IsNotExist(err) {
		t.Fatalf("collision copied session: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(root, "session.side/empty")); err != nil || string(got) != "destination file" {
		t.Fatalf("collision clobbered destination: %q %v", got, err)
	}
}

func targetReferenceDocument(metadata, target string) string {
	return fmt.Sprintf("---\n%sparent: %s\ndiscussions:\n  - doc: %s\n---\n[inline](%s)\n[reference][target]\n[target]: %s\n![image](%s)\n", metadata, target, target, target, target, target)
}

func assertMovedReferenceTargets(t *testing.T, root, document, want string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, document))
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.SplitN(string(b), "---\n", 3)
	if len(parts) != 3 {
		t.Fatalf("%s has no frontmatter", document)
	}
	var metadata struct {
		Parent      string `yaml:"parent"`
		Discussions []struct {
			Doc string `yaml:"doc"`
		} `yaml:"discussions"`
	}
	if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
		t.Fatal(err)
	}
	if len(metadata.Discussions) != 1 {
		t.Fatalf("discussions: %+v", metadata.Discussions)
	}
	targets := map[string]string{"parent": metadata.Parent, "discussions": metadata.Discussions[0].Doc}
	for name, pattern := range map[string]string{
		"inline":    `\[inline\]\(([^)]+)\)`,
		"reference": `(?m)^\[target\]: (\S+)`,
		"image":     `!\[image\]\(([^)]+)\)`,
	} {
		match := regexp.MustCompile(pattern).FindStringSubmatch(parts[2])
		if len(match) != 2 {
			t.Fatalf("missing %s consumer in %s", name, document)
		}
		targets[name] = match[1]
	}
	for consumer, target := range targets {
		resolved := path.Join(path.Dir(document), target)
		if resolved != want {
			t.Errorf("%s in %s resolves to %s, want %s", consumer, document, resolved, want)
		}
		if _, err := os.Stat(filepath.Join(root, resolved)); err != nil {
			t.Errorf("%s final target %s does not exist: %v", consumer, resolved, err)
		}
	}
}
