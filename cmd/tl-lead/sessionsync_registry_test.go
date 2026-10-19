package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// The fake's whole-file apply and registry, as TestSessionSyncRegistryConformance
// pins couch's (#455): apply is all or none, stale on a moved hash, refused
// under the lock; rename needs the new file and succeeds again once done.
func TestFakeApplyAndRename(t *testing.T) {
	f := newFake(map[string]string{"A": "see [x](qa/x.md) and [y](qa/x.md)"})
	dir := t.TempDir()
	edits := func(recs string) string {
		p := filepath.Join(dir, "e.json")
		if err := os.WriteFile(p, []byte(recs), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	before := f.text()
	bad := `[{"old":"](qa/x.md)","occurrence":2,"new":"](s/x.md)"},{"old":"nowhere","occurrence":1,"new":"z"}]`
	if err := f.Apply("s.md", edits(bad), fileHash(before)); err == nil || f.text() != before {
		t.Errorf("a bad record: err %v, file changed %v", err, f.text() != before)
	}
	good := `[{"old":"](qa/x.md)","occurrence":2,"new":"](s/x.md)"},{"old":"](qa/x.md)","occurrence":1,"new":"](s/x.md)"}]`
	if err := f.Apply("s.md", edits(good), "0000000000000000"); !errors.Is(err, ErrStale) {
		t.Errorf("a moved hash: %v", err)
	}
	f.holder = "operator"
	if err := f.Apply("s.md", edits(good), fileHash(before)); !errors.Is(err, ErrLocked) {
		t.Errorf("under the lock: %v", err)
	}
	f.holder = ""
	if err := f.Apply("s.md", edits(good), fileHash(before)); err != nil || f.sections["A"] != "see [x](s/x.md) and [y](s/x.md)" || f.applies != 1 {
		t.Errorf("apply: err %v, A %q", err, f.sections["A"])
	}

	a, b := filepath.Join(dir, "a.md"), filepath.Join(dir, "b.md")
	f.registered = map[string]bool{a: true}
	if err := f.Rename(a, b); err == nil {
		t.Error("rename to a missing file succeeded")
	}
	if err := os.WriteFile(b, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := f.Rename(a, b); err != nil || !f.registered[b] || f.registered[a] {
		t.Errorf("rename: %v %v", err, f.registered)
	}
	if err := f.Rename(a, b); err != nil {
		t.Errorf("a rename already done: %v", err)
	}
	if err := f.Rename(filepath.Join(dir, "c.md"), b+"x"); err == nil {
		t.Error("rename of an unregistered file succeeded")
	}
}
