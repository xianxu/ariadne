package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestSessionSyncConformance checks the fake's model against the real couch:
// read returns a hash, a write naming it succeeds, a stale hash is exit 6, and
// the operator's lock is exit 3. Opt-in (TL_LEAD_LIVE=1, couch on PATH): run it
// at each close of this binary's work and whenever pair changes session-sync.
// The file names a dummy owner, so no slot's owed events are involved, and the
// test never folds.
func TestSessionSyncConformance(t *testing.T) {
	if os.Getenv("TL_LEAD_LIVE") != "1" {
		t.Skip("set TL_LEAD_LIVE=1 to run against the real couch")
	}
	bin, err := exec.LookPath("couch")
	if err != nil {
		t.Skip("couch is not on PATH")
	}
	ss := couchSessionSync{bin: bin}
	dir := t.TempDir()
	file := filepath.Join(dir, "s.md")
	if err := os.WriteFile(file, []byte("---\ntype: session-sync\ncouch-owner: tl-lead-conformance:0\n---\nUpdated 2026-01-01 00:00\n\n## A\n\nold\n\n## B\n\nb\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	body := filepath.Join(dir, "body.md")
	if err := os.WriteFile(body, []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, hash, err := ss.ReadSection(file, "A")
	if err != nil || hash == "" || strings.TrimSpace(got) != "old" {
		t.Fatalf("read: body=%q hash=%q err=%v", got, hash, err)
	}
	if err := ss.WriteSection(file, "A", body, "0000000000000000", nil); !errors.Is(err, ErrStale) {
		t.Errorf("a wrong hash: err = %v, want ErrStale (exit 6)", err)
	}
	if err := ss.WriteSection(file, "A", body, hash, nil); err != nil {
		t.Fatalf("write naming the read hash: %v", err)
	}
	if got, _, _ := ss.ReadSection(file, "A"); strings.TrimSpace(got) != "new" {
		t.Errorf("after the write, A = %q", got)
	}
	if _, err := ss.Status(file); err != nil {
		t.Errorf("status: %v", err)
	}

	if err := os.WriteFile(file+".lock", []byte("holder: operator\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, hash, err = ss.ReadSection(file, "B")
	if err != nil {
		t.Fatal(err)
	}
	if err := ss.WriteSection(file, "B", body, hash, nil); !errors.Is(err, ErrLocked) {
		t.Errorf("under the operator's lock: err = %v, want ErrLocked (exit 3)", err)
	}
}
