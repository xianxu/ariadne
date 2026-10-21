package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionMoveAllowsLiveEditsAndLocks(t *testing.T) {
	root, old, f := sessionMoveFixture(t)
	var events []string
	err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, func(event string) error {
		events = append(events, event)
		if event == "rename session.md" {
			writeFile(t, filepath.Join(root, "session.md"), "live operator text")
			writeFile(t, filepath.Join(root, "session.md.lock"), "operator")
			writeFile(t, filepath.Join(root, "session.side/01-a.md"), "live side text")
			writeFile(t, filepath.Join(root, "session.side/01-a.md.lock"), "operator")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("live writes blocked cleanup: %v (%v)", err, events)
	}
	if _, err := os.Stat(filepath.Join(root, old)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("old remains: %v", err)
	}
	for p, want := range map[string]string{"session.md": "live operator text", "session.side/01-a.md": "live side text"} {
		b, err := os.ReadFile(filepath.Join(root, p))
		if err != nil || string(b) != want {
			t.Fatalf("live file overwritten: %s %q %v", p, b, err)
		}
	}
}

func TestSessionMoveRetriesCleanupOfChangedLiveTree(t *testing.T) {
	root, old, f := sessionMoveFixture(t)
	err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, func(event string) error {
		if strings.HasPrefix(event, "remove ") {
			return errInterrupted
		}
		return nil
	})
	if !errors.Is(err, errInterrupted) {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "session.md"), "live new root")
	writeFile(t, filepath.Join(root, "session.md.lock"), "operator")
	writeFile(t, filepath.Join(root, "session.side/01-a.md"), "live new side")
	writeFile(t, filepath.Join(root, "session.side/01-a.md.lock"), "operator")
	var preview bytes.Buffer
	if err := moveSession(deps{ss: f}, &preview, root, old, false, nil); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(preview.String(), "copy ") || strings.Contains(preview.String(), "update .gitignore") || !strings.Contains(preview.String(), "remove unchanged") {
		t.Fatalf("cleanup preview promised writes to live tree: %s", preview.String())
	}
	if err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, old)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("old remains: %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(root, "session.md"))
	if string(b) != "live new root" {
		t.Fatalf("clobbered live text %q", b)
	}
}

type countingMoveSync struct {
	*fakeSessionSync
	lists, renames int
}

func (c *countingMoveSync) LookupRegistration(string) (*syncRegistration, error) {
	return nil, fmt.Errorf("per-document lookup forbidden")
}
func (c *countingMoveSync) Status(string) (SyncStatus, error) {
	return SyncStatus{}, fmt.Errorf("per-document status forbidden")
}
func (c *countingMoveSync) RegistrationSnapshot() (map[string]*syncRegistration, error) {
	c.lists++
	return c.fakeSessionSync.RegistrationSnapshot()
}
func (c *countingMoveSync) Rename(a, b string) error {
	c.renames++
	return c.fakeSessionSync.Rename(a, b)
}
func TestSessionMoveBatchesRegistrationObservations(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			root, old, f := sessionMoveFixture(t)
			for i := 0; i < n; i++ {
				p := filepath.Join(root, "teams/dev/tl-status-op.side", fmt.Sprintf("doc-%d.md", i))
				writeFile(t, p, "ordinary document")
				f.registered[p] = true
			}
			c := &countingMoveSync{fakeSessionSync: f}
			gitCount := countMoveGitCommands(t)
			if err := moveSession(deps{ss: c}, &bytes.Buffer{}, root, old, true, nil); err != nil {
				t.Fatal(err)
			}
			if c.lists > 3 || c.lists < 2 {
				t.Fatalf("list processes at N=%d: %d", n, c.lists)
			}
			calls, err := os.ReadFile(gitCount)
			if err != nil || len(calls) > n+12 {
				t.Fatalf("Git process growth at N=%d: %d %v", n, len(calls), err)
			}
			if c.renames != n+2 {
				t.Fatalf("rename count %d", c.renames)
			}
		})
	}
}

func TestSessionMoveRejectsUnrelatedCleanupRegistration(t *testing.T) {
	root, old, f := sessionMoveFixture(t)
	delete(f.registered, filepath.Join(root, old))
	f.registered[filepath.Join(root, "session.md")] = true
	writeFile(t, filepath.Join(root, "session.md"), "unrelated live file")
	if err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, nil); err == nil {
		t.Fatal("unrelated registration authorized cleanup")
	}
	if _, err := os.Stat(filepath.Join(root, old)); err != nil {
		t.Fatal("lost source")
	}
	if _, err := os.Stat(filepath.Join(root, ".gitignore")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("cleanup candidate modified ignore policy")
	}
}

func TestSessionMovePreservesChangedSourceAfterSwitch(t *testing.T) {
	root, old, f := sessionMoveFixture(t)
	err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, func(event string) error {
		if strings.HasPrefix(event, "remove ") {
			writeFile(t, filepath.Join(root, old), "late old edit")
		}
		return nil
	})
	if err == nil {
		t.Fatal("deleted changed old source")
	}
	b, _ := os.ReadFile(filepath.Join(root, old))
	if string(b) != "late old edit" {
		t.Fatalf("lost old edit %q", b)
	}
}
