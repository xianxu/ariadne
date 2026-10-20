package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionMoveRejectNestedIgnorePolicy(t *testing.T) {
	for _, rel := range []string{
		"old.side/.gitignore",
		"old.side/child/.gitignore",
		"old.side/child/deeper/.gitignore",
	} {
		t.Run(rel, func(t *testing.T) {
			s := sessionMoveSnapshot{Old: "old.md", Files: map[string]moveSource{
				rel: {Bytes: []byte("!secret.md\n")},
			}}
			_, err := planSessionMove(s)
			if err == nil || !strings.Contains(err.Error(), rel) {
				t.Fatalf("must refuse ignore policy before effects and name %s: %v", rel, err)
			}
		})
	}
}

func TestSessionMoveRejectRetryIgnorePolicy(t *testing.T) {
	for _, rel := range []string{
		"session.side/.gitignore",
		"session.side/child/.gitignore",
		"session.side/child/deeper/.gitignore",
	} {
		t.Run(rel, func(t *testing.T) {
			s := sessionMoveSnapshot{Old: "old.md", Dest: map[string]moveSource{
				rel: {Bytes: []byte("!secret.md\n")},
			}}
			_, err := planSessionMove(s)
			if err == nil || !strings.Contains(err.Error(), rel) {
				t.Fatalf("must refuse retry ignore policy before effects and name %s: %v", rel, err)
			}
		})
	}
}

func TestSessionMoveAcceptOrdinaryDotfiles(t *testing.T) {
	s := sessionMoveSnapshot{Old: "old.md", Files: map[string]moveSource{
		"old.side/.gitignore.backup": {Bytes: []byte("!secret.md\n")},
		"old.side/deeper/.notes":     {Bytes: []byte("private notes\n")},
	}, Dest: map[string]moveSource{
		"session.side/deeper/.gitignore.backup": {Bytes: []byte("!secret.md\n")},
	}}
	if _, err := planSessionMove(s); err != nil {
		t.Fatalf("ordinary dotfiles must remain supported: %v", err)
	}
}

func TestSessionMoveRejectLateIgnorePolicy(t *testing.T) {
	for _, rel := range []string{"session.side/.gitignore", "session.side/child/.gitignore"} {
		t.Run(rel, func(t *testing.T) {
			root, old, f := sessionMoveFixture(t)
			snap, err := readSessionMove(deps{ss: f}, root, old)
			if err != nil {
				t.Fatal(err)
			}
			plan, err := planSessionMove(snap)
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(root, ".gitignore"), string(plan.Ignores))
			// The snapshot predates an independently added destination policy.
			writeFile(t, filepath.Join(root, rel), "!secret.md\n")
			err = verifyMoveIgnores(root, snap, plan)
			if err == nil || !strings.Contains(err.Error(), rel) {
				t.Fatalf("late destination policy must be refused by name: %v", err)
			}
		})
	}
}

func TestSessionMoveIgnorePolicyRefusalBeforeEffects(t *testing.T) {
	for _, where := range []string{"source", "retry"} {
		t.Run(where, func(t *testing.T) {
			root, old, f := sessionMoveFixture(t)
			rel := "teams/dev/tl-status-op.side/child/.gitignore"
			if where == "retry" {
				rel = "session.side/child/.gitignore"
			}
			writeFile(t, filepath.Join(root, rel), "!secret.md\n")
			before, err := os.ReadFile(filepath.Join(root, old))
			if err != nil {
				t.Fatal(err)
			}
			err = moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, nil)
			if err == nil || !strings.Contains(err.Error(), rel) {
				t.Fatalf("must refuse named policy before effects: %v", err)
			}
			after, err := os.ReadFile(filepath.Join(root, old))
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("source changed: %v", err)
			}
			for _, newFile := range []string{".gitignore", "session.md", "session.side/01-a.md"} {
				if _, err := os.Stat(filepath.Join(root, newFile)); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("effect before refusal: %s: %v", newFile, err)
				}
			}
			if !f.registered[filepath.Join(root, old)] || f.registered[filepath.Join(root, "session.md")] {
				t.Fatal("registration switched before refusal")
			}
		})
	}
}
