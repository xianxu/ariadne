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
// read returns a hash, a write naming it succeeds, a stale hash is exit 6, a
// fold of an ID not owed refuses, and the operator's lock is exit 3. Opt-in (TL_LEAD_LIVE=1, couch on PATH): run it
// at each close of this binary's work and whenever pair changes session-sync.
// The file names a dummy owner, so no slot's owed events are involved, and the
// only fold names an ID that isn't owed (a refusal, which changes nothing).
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
	// status check tells a missing section (no finding) from a failed read by
	// this text.
	if _, _, err := ss.ReadSection(file, "Missing"); err == nil || !strings.Contains(err.Error(), missingSection) {
		t.Errorf("a missing section: err = %v, want it to contain %q", err, missingSection)
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
	// Folding an ID that isn't owed refuses and changes nothing, as the fake
	// models it. The dummy owner has no owed events, so no real one is touched.
	_, hash, err = ss.ReadSection(file, "A")
	if err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(dir, "other.md")
	if err := os.WriteFile(other, []byte("other\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ss.WriteSection(file, "A", other, hash, []string{"00000000-not-owed"}); err == nil || errors.Is(err, ErrStale) || errors.Is(err, ErrLocked) {
		t.Errorf("a fold of an ID not owed: err = %v, want a refusal", err)
	}
	if got, _, _ := ss.ReadSection(file, "A"); strings.TrimSpace(got) != "new" {
		t.Errorf("a refused fold changed A to %q", got)
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

// TestCouchActorsConformance checks decodeActors against the running Couch:
// the reply decodes and names at least one slot with an agent. Opt-in like
// TestSessionSyncConformance; it needs the Couch socket (unsandboxed).
func TestCouchActorsConformance(t *testing.T) {
	if os.Getenv("TL_LEAD_LIVE") != "1" {
		t.Skip("set TL_LEAD_LIVE=1 to run against the real couch")
	}
	bin, err := exec.LookPath("couch")
	if err != nil {
		t.Skip("couch is not on PATH")
	}
	actors, err := execActors{bin: bin}.Actors()
	if err != nil {
		t.Fatal(err)
	}
	for slot, a := range actors {
		if !strings.Contains(slot, ":") || a.Agent == "" {
			t.Errorf("actor %q = %+v: want a repo:N slot and an agent", slot, a)
		}
	}
	if len(actors) == 0 {
		t.Error("no actors from a running Couch")
	}
}
