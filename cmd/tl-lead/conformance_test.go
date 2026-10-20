package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/sidedoc"
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

	// The frontmatter scope the index uses (#426, pair#643): a read and a
	// write mark only the frontmatter seen, so an operator's section edit
	// stays in the diff; a stale hash is exit 6; the write keeps every body
	// byte. Against this scratch file only, never a real session file.
	if _, err := ss.Diff(file); err != nil {
		t.Fatal(err)
	}
	if err := ss.MarkAllSeen(file); err != nil {
		t.Fatal(err)
	}
	before, err := ss.Peek(file)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(before, "\nb\n", "\nb edited by the operator\n", 1)
	if err := os.WriteFile(file, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	block, fmHash, err := ss.ReadFrontmatter(file)
	if want, _, _ := sidedoc.FrontBlock(edited); err != nil || fmHash == "" || block != want {
		t.Fatalf("read --frontmatter: hash=%q err=%v\n%q\nwant %q", fmHash, err, block, want)
	}
	if diff, err := ss.Diff(file); err != nil || !strings.Contains(diff, "b edited by the operator") {
		t.Errorf("read --frontmatter consumed the operator's section edit: diff %q, %v", diff, err)
	}
	fm := filepath.Join(dir, "fm.md")
	next := strings.Replace(block, "type: session-sync\n", "type: session-sync\ndiscussions:\n  - doc: s.side/01-x.md\n", 1)
	if err := os.WriteFile(fm, []byte(next+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ss.WriteFrontmatter(file, fm, "0000000000000000"); !errors.Is(err, ErrStale) {
		t.Errorf("write --frontmatter with a wrong hash: err = %v, want ErrStale (exit 6)", err)
	}
	if err := ss.WriteFrontmatter(file, fm, fmHash); err != nil {
		t.Fatalf("write --frontmatter naming the read hash: %v", err)
	}
	after, _ := ss.Peek(file)
	if _, body, _ := sidedoc.FrontBlock(after); !strings.HasPrefix(after, next) || !strings.HasSuffix(edited, body) {
		t.Errorf("write --frontmatter changed more than the block:\n%s", after)
	}
	if diff, err := ss.Diff(file); err != nil || !strings.Contains(diff, "b edited by the operator") || strings.Contains(diff, "01-x.md") {
		t.Errorf("after write --frontmatter: diff %q, %v; want only the operator's section edit", diff, err)
	}

	if err := os.WriteFile(file+".lock", []byte("holder: operator\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, fmHash, err = ss.ReadFrontmatter(file); err != nil {
		t.Fatal(err)
	}
	if err := ss.WriteFrontmatter(file, fm, fmHash); !errors.Is(err, ErrLocked) {
		t.Errorf("write --frontmatter under the operator's lock: err = %v, want ErrLocked (exit 3)", err)
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

// TestSessionSyncRegistryConformance pins what the migration (#455) relies
// on in the real couch: a whole-file hash is fileHash of the file's bytes;
// apply is all or none and stale on a moved hash; rename needs the new file
// to exist, moves the registration, and succeeds again once done. It uses a
// scratch file registered to a dummy owner and forgets it at the end.
// Opt-in, like TestSessionSyncConformance.
func TestSessionSyncRegistryConformance(t *testing.T) {
	if os.Getenv("TL_LEAD_LIVE") != "1" {
		t.Skip("set TL_LEAD_LIVE=1 to run against the real couch")
	}
	bin, err := exec.LookPath("couch")
	if err != nil {
		t.Skip("couch is not on PATH")
	}
	ss := couchSessionSync{bin: bin}
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.md"), filepath.Join(dir, "b.md")
	if err := os.WriteFile(a, []byte("---\ntype: session-sync\ncouch-owner: tl-lead-conformance:0\n---\n\n## A\n\nsee [x](qa/x.md)\nand [y](qa/x.md)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ss.run("register", a, "--owner", "tl-lead-conformance:0", "--json"); err != nil {
		t.Fatalf("register: %v", err)
	}
	t.Cleanup(func() {
		for _, p := range []string{b, a} {
			_, _ = ss.run("forget", p, "--json")
		}
	})

	raw, err := ss.run("read", "--all", a, "--json")
	if err != nil {
		t.Fatal(err)
	}
	r, err := decodeReply("read", raw)
	text, _ := ss.Peek(a)
	if err != nil || r.Read == nil || r.Read.Hash != fileHash(text) {
		t.Fatalf("read --all hash %v, fileHash %s, err %v: the formula drifted", r.Read, fileHash(text), err)
	}
	edits := func(recs string) string {
		p := filepath.Join(dir, "edits.json")
		if err := os.WriteFile(p, []byte(recs), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	// occurrence 2 counts from the end of the first match, as applyRecords
	// does; the migration's records can name any occurrence.
	good := `[{"old":"](qa/x.md)","occurrence":2,"new":"](s.side/y.md)","explain":"conformance"},{"old":"](qa/x.md)","occurrence":1,"new":"](s.side/x.md)","explain":"conformance"}]`
	bad := `[{"old":"](qa/x.md)","occurrence":1,"new":"](s.side/x.md)","explain":"ok"},{"old":"absent text","occurrence":1,"new":"y","explain":"bad"}]`
	if err := ss.Apply(a, edits(bad), fileHash(text)); err == nil {
		t.Error("an apply with an unresolvable record succeeded")
	}
	if after, _ := ss.Peek(a); after != text {
		t.Errorf("a refused apply changed the file:\n%s", after)
	}
	if err := ss.Apply(a, edits(good), "0000000000000000"); !errors.Is(err, ErrStale) {
		t.Errorf("a moved hash: err = %v, want ErrStale", err)
	}
	if err := ss.Apply(a, edits(good), fileHash(text)); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if after, _ := ss.Peek(a); !strings.Contains(after, "see [x](s.side/x.md)\nand [y](s.side/y.md)") {
		t.Errorf("apply did not land:\n%s", after)
	}

	if err := ss.Rename(a, b); err == nil {
		t.Error("a rename to a path with no file succeeded")
	}
	if err := os.Rename(a, b); err != nil {
		t.Fatal(err)
	}
	if err := ss.Rename(a, b); err != nil {
		t.Fatalf("rename: %v", err)
	}
	regs, err := ss.Registrations()
	if err != nil {
		t.Fatal(err)
	}
	has := func(p string) bool {
		for _, r := range regs {
			if r == p {
				return true
			}
		}
		return false
	}
	// couch canonicalises paths; compare against the resolved temp dir.
	rb, _ := filepath.EvalSymlinks(b)
	ra := filepath.Join(filepath.Dir(rb), "a.md")
	if !(has(b) || has(rb)) || has(a) || has(ra) {
		t.Errorf("after rename, registrations = %v", regs)
	}
	if err := ss.Rename(a, b); err != nil {
		t.Errorf("a rename already done: %v, want success", err)
	}
}

// Exercise the production migration with real registry state, using only a
// disposable hub and dummy owner. No live session file is read or changed.
func TestSessionSyncMoveConformance(t *testing.T) {
	if os.Getenv("TL_LEAD_LIVE") != "1" {
		t.Skip("set TL_LEAD_LIVE=1 to run against real couch")
	}
	bin, err := exec.LookPath("couch")
	if err != nil {
		t.Skip("couch is not on PATH")
	}
	ss := couchSessionSync{bin: bin}
	root, old, _ := sessionMoveFixture(t)
	oldAbs := filepath.Join(root, old)
	sideOld := filepath.Join(root, "teams/dev/tl-status-op.side/01-a.md")
	raw, err := os.ReadFile(oldAbs)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, oldAbs, strings.Replace(string(raw), "couch-owner: ops:0", "couch-owner: tl-lead-conformance:0", 1))
	for _, p := range []string{oldAbs, sideOld} {
		if _, err := ss.run("register", p, "--owner", "tl-lead-conformance:0", "--json"); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		for _, p := range []string{oldAbs, sideOld, filepath.Join(root, "session.md"), filepath.Join(root, "session.side/01-a.md")} {
			if _, err := ss.run("forget", p, "--json"); err != nil {
				t.Errorf("forget scratch %s: %v", p, err)
			}
		}
	})
	if err := ss.MarkAllSeen(oldAbs); err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(oldAbs)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, oldAbs, string(text)+"\nunread-operator-edit\n")
	before, err := ss.LookupRegistration(oldAbs)
	if err != nil || before == nil {
		t.Fatalf("lookup before: %+v %v", before, err)
	}
	var out strings.Builder
	// A real published rename followed by a lost response must permit cleanup
	// after the operator has already edited the now-live destination.
	err = moveSession(deps{ss: ss}, &out, root, old, true, func(event string) error {
		if event == "rename session.md" {
			live, err := os.ReadFile(filepath.Join(root, "session.md"))
			if err != nil {
				return err
			}
			writeFile(t, filepath.Join(root, "session.md"), string(live)+"\npost-switch-operator-edit\n")
			return errInterrupted
		}
		return nil
	})
	if !errors.Is(err, errInterrupted) {
		t.Fatalf("expected post-switch interruption: %v", err)
	}
	if err := moveSession(deps{ss: ss}, &out, root, old, true, nil); err != nil {
		t.Fatalf("cleanup retry: %v\n%s", err, out.String())
	}
	after, err := ss.LookupRegistration(filepath.Join(root, "session.md"))
	if err != nil || after == nil || after.ID != before.ID {
		t.Fatalf("lookup after: %+v %v", after, err)
	}
	diff, err := ss.Diff(filepath.Join(root, "session.md"))
	if err != nil || (!strings.Contains(diff, "unread-operator-edit") || !strings.Contains(diff, "post-switch-operator-edit")) {
		t.Fatalf("unread edit lost: %q %v", diff, err)
	}
	if err := moveSession(deps{ss: ss}, &out, root, old, true, nil); err != nil {
		t.Fatalf("repeat: %v", err)
	}
}
