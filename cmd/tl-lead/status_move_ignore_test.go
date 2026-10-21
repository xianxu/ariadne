package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionMoveEffectiveIgnores(t *testing.T) {
	for _, tc := range []struct{ name, rules, exposed string }{
		{"session", "/session.md\n!/session.md\n", "session.md"},
		{"side", "/session.side/\n!/session.side/\n", "session.side/01-a.md"},
		{"seen", "/.session.md.ss-seen\n!/.session.md.ss-seen\n", ".session.md.ss-seen"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, old, f := sessionMoveFixture(t)
			writeFile(t, filepath.Join(root, ".gitignore"), tc.rules)
			before, err := readSessionMove(deps{ss: f}, root, old)
			if err != nil {
				t.Fatal(err)
			}
			err = moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, nil)
			if err == nil || !strings.Contains(err.Error(), "ignore") {
				t.Fatalf("accepted exposed %s: %v", tc.exposed, err)
			}
			for rel, src := range before.Files {
				got, err := os.ReadFile(filepath.Join(root, rel))
				if err != nil || !bytes.Equal(got, src.Bytes) {
					t.Fatalf("source changed %s: %v", rel, err)
				}
				to, _ := moveTarget(old, rel)
				if _, err := os.Stat(filepath.Join(root, to)); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("copied %s before ignore refusal: %v", to, err)
				}
			}
		})
	}
}

func TestSessionMoveIgnoreExistingRetryDestination(t *testing.T) {
	root, old, f := sessionMoveFixture(t)
	err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, func(event string) error {
		if event == "copy session.md" {
			return errInterrupted
		}
		return nil
	})
	if !errors.Is(err, errInterrupted) {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "session.side/retry.bin"), "keep existing destination")
	rules, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, ".gitignore"), string(rules)+"!/session.side/\n/session.side/*\n!/session.side/retry.bin\n")
	err = moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, nil)
	if err == nil || !strings.Contains(err.Error(), "ignore") {
		t.Fatalf("accepted exposed retry destination: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, old)); err != nil {
		t.Fatalf("cleaned old session: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(root, "session.side/retry.bin")); err != nil || string(got) != "keep existing destination" {
		t.Fatalf("existing destination changed: %q %v", got, err)
	}
}

func TestSessionMoveIgnoreRulesAppended(t *testing.T) {
	root, old, f := sessionMoveFixture(t)
	writeFile(t, filepath.Join(root, ".gitignore"), "# keep\n")
	if err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, nil); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"session.md", "session.side/01-a.md", "session.side/01-a.side/02-b.md", "session.side/image.bin", ".session.md.ss-seen"} {
		if _, err := git(root, "check-ignore", "--quiet", "--", rel); err != nil {
			t.Fatalf("%s not ignored: %v", rel, err)
		}
	}
	rules, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil || !strings.HasPrefix(string(rules), "# keep\n/session.md\n") {
		t.Fatalf("missing appended anchored rules: %q %v", rules, err)
	}
}

func TestSessionMoveIgnoreCheckFailurePreservesSources(t *testing.T) {
	root, old, f := sessionMoveFixture(t)
	err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, func(event string) error {
		if event == "ignore" {
			return os.Rename(filepath.Join(root, ".git"), filepath.Join(root, ".git-off"))
		}
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "ignore") {
		t.Fatalf("accepted failed git ignore check: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, old)); err != nil {
		t.Fatalf("cleaned old session: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "session.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("copied before ignore check: %v", err)
	}
}
