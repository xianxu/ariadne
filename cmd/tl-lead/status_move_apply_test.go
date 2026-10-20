package main

import (
	"bytes"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sessionMoveFixture(t *testing.T) (string, string, *fakeSessionSync) {
	t.Helper()
	root := t.TempDir()
	root, _ = filepath.EvalSymlinks(root)
	gitInit(t, root)
	writeFile(t, filepath.Join(root, ".hub"), "ariadne\n")
	old := "teams/dev/tl-status-op.md"
	text := "---\ntype: session-sync\ncouch-owner: ops:0\noperator: Op\ndiscussions:\n  - doc: tl-status-op.side/01-a.md\n---\n\n## Needs you\n\n[x](tl-status-op.side/01-a.md)\n"
	writeFile(t, filepath.Join(root, old), text)
	writeFile(t, filepath.Join(root, "teams/dev/tl-status-op.side/01-a.md"), "---\ntype: side-doc\nparent: ../tl-status-op.md\n---\n[parent](../tl-status-op.md)\n")
	writeFile(t, filepath.Join(root, "teams/dev/tl-status-op.side/01-a.side/02-b.md"), "---\ntype: side-doc\nparent: ../01-a.md\n---\n[session](../../tl-status-op.md)\n")
	writeFile(t, filepath.Join(root, "teams/dev/tl-status-op.side/image.bin"), "\x00asset\xff")
	writeFile(t, seenPath(filepath.Join(root, old)), `["event"]`)
	f := newFake(map[string]string{"Needs you": "unread operator edit"})
	f.file = filepath.Join(root, old)
	f.owed = []string{"event"}
	f.registered = map[string]bool{f.file: true, filepath.Join(root, "teams/dev/tl-status-op.side/01-a.md"): true}
	f.seen()
	f.sections["Needs you"] = "new unread edit"
	return root, old, f
}

func TestSessionMoveApplyAndDryRun(t *testing.T) {
	root, old, f := sessionMoveFixture(t)
	d := deps{ss: f}
	var out bytes.Buffer
	if err := moveSession(d, &out, root, old, false, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "session.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry run wrote destination: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".gitignore")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry run wrote ignore: %v", err)
	}
	before, _ := f.LookupRegistration(filepath.Join(root, old))
	if err := moveSession(d, &out, root, old, true, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, old)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("old remains: %v", err)
	}
	newFile := filepath.Join(root, "session.md")
	after, _ := f.LookupRegistration(newFile)
	if after == nil || after.ID != before.ID {
		t.Fatalf("identity %+v -> %+v", before, after)
	}
	if f.seenAll || len(f.owed) != 1 {
		t.Fatalf("consumed unread/owed state: %+v", f)
	}
	for file, want := range map[string]string{
		"session.md":                     "[x](session.side/01-a.md)",
		"session.side/01-a.md":           "parent: ../session.md",
		"session.side/01-a.side/02-b.md": "[session](../../session.md)",
		"session.side/image.bin":         "\x00asset\xff",
		".session.md.ss-seen":            `["event"]`,
	} {
		b, err := os.ReadFile(filepath.Join(root, file))
		if err != nil || !strings.Contains(string(b), want) {
			t.Errorf("%s: %q %v", file, b, err)
		}
	}
	if err := moveSession(d, &out, root, old, true, nil); err != nil {
		t.Fatalf("repeat completed: %v", err)
	}
	if got, err := git(root, "check-ignore", "session.md", "session.side/01-a.md", ".session.md.ss-seen"); err != nil || len(strings.Fields(got)) != 3 {
		t.Fatalf("ignores: %s %v", got, err)
	}
}

func TestSessionMoveRefusals(t *testing.T) {
	for _, kind := range []string{"lock", "side-lock", "collision", "symlink", "unregistered", "clobbered", "unknown-destination-registration"} {
		t.Run(kind, func(t *testing.T) {
			root, old, f := sessionMoveFixture(t)
			switch kind {
			case "lock":
				writeFile(t, filepath.Join(root, old)+".lock", "holder: operator\n")
			case "side-lock":
				writeFile(t, filepath.Join(root, "teams/dev/tl-status-op.side/01-a.md.lock"), "malformed")
			case "collision":
				writeFile(t, filepath.Join(root, "session.md"), "somebody else's draft")
			case "symlink":
				if err := os.Symlink("teams/dev/tl-status-op.side", filepath.Join(root, "session.side")); err != nil {
					t.Fatal(err)
				}
			case "unregistered":
				delete(f.registered, filepath.Join(root, old))
			case "clobbered":
				writeFile(t, filepath.Join(root, old)+".clobbered", "recover me")
			case "unknown-destination-registration":
				f.registered[filepath.Join(root, "session.md")] = true
			}
			var out bytes.Buffer
			err := moveSession(deps{ss: f}, &out, root, old, true, nil)
			if err == nil {
				t.Fatal("accepted conflict")
			}
			if strings.Contains(kind, "lock") && exitCode(err) != 3 {
				t.Fatalf("lock exit: %v", err)
			}
			if kind == "unregistered" && !strings.Contains(err.Error(), "couch --session-sync register") {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(root, old)); err != nil {
				t.Fatalf("lost old: %v", err)
			}
			if _, err := os.Stat(filepath.Join(root, ".gitignore")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("refusal wrote ignore: %v", err)
			}
		})
	}
}

func TestSessionMoveInterrupted(t *testing.T) {
	root, old, f := sessionMoveFixture(t)
	steps := 0
	if err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, func(string) error { steps++; return nil }); err != nil {
		t.Fatal(err)
	}
	for stop := 1; stop <= steps; stop++ {
		t.Run(string(rune('A'+stop)), func(t *testing.T) {
			root, old, f := sessionMoveFixture(t)
			n := 0
			err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, func(string) error {
				n++
				if n == stop {
					return errInterrupted
				}
				return nil
			})
			if !errors.Is(err, errInterrupted) {
				t.Fatalf("stop %d: %v", stop, err)
			}
			if err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, nil); err != nil {
				t.Fatalf("resume %d: %v", stop, err)
			}
			if _, err := os.Stat(filepath.Join(root, old)); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("old remains")
			}
			if !f.registered[filepath.Join(root, "session.md")] {
				t.Fatal("new registration missing")
			}
		})
	}
}

type uncertainMoveSync struct {
	*fakeSessionSync
	uncertain bool
}

func (f *uncertainMoveSync) Rename(old, new string) error {
	if err := f.fakeSessionSync.Rename(old, new); err != nil {
		return err
	}
	if f.uncertain {
		f.uncertain = false
		return errors.New("registry-uncertain: reply lost after publication")
	}
	return nil
}

func TestSessionMoveUncertainRename(t *testing.T) {
	root, old, f := sessionMoveFixture(t)
	ss := &uncertainMoveSync{fakeSessionSync: f, uncertain: true}
	err := moveSession(deps{ss: ss}, &bytes.Buffer{}, root, old, true, nil)
	if err == nil || !strings.Contains(err.Error(), "switch unconfirmed") {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, old)); err != nil {
		t.Fatal("old deleted after unknown outcome")
	}
	if err := moveSession(deps{ss: ss}, &bytes.Buffer{}, root, old, true, nil); err != nil {
		t.Fatal(err)
	}
}

func TestSessionMoveInterveningEdits(t *testing.T) {
	for _, kind := range []string{"source", "destination", "lock", "cleanup-destination", "late-file"} {
		t.Run(kind, func(t *testing.T) {
			root, old, f := sessionMoveFixture(t)
			triggered := false
			err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, func(event string) error {
				if triggered {
					return nil
				}
				if kind == "cleanup-destination" && strings.HasPrefix(event, "remove ") {
					triggered = true
					writeFile(t, filepath.Join(root, "session.md"), "operator's new draft")
					return nil
				}
				if kind == "late-file" && strings.HasPrefix(event, "copy ") {
					triggered = true
					writeFile(t, filepath.Join(root, "teams/dev/tl-status-op.side/late.md"), "new draft")
					return nil
				}
				if kind == "cleanup-destination" {
					return nil
				}
				if event != "copy session.md" {
					return nil
				}
				triggered = true
				switch kind {
				case "source":
					writeFile(t, filepath.Join(root, old), "operator's source edit")
				case "destination":
					writeFile(t, filepath.Join(root, "session.md"), "operator's destination edit")
				case "lock":
					writeFile(t, filepath.Join(root, old)+".lock", "holder: operator\n")
				}
				return nil
			})
			if kind == "late-file" {
				if err == nil {
					t.Fatal("accepted newly arrived source doc outside snapshot")
				}
				if b, _ := os.ReadFile(filepath.Join(root, "teams/dev/tl-status-op.side/late.md")); string(b) != "new draft" {
					t.Fatal("lost late doc")
				}
			} else if err == nil {
				t.Fatal("accepted an intervening edit/lock")
			}
			if _, err := os.Stat(filepath.Join(root, old)); err != nil {
				t.Fatalf("lost original session: %v", err)
			}
		})
	}
}

func TestSessionMoveDiscovery(t *testing.T) {
	root, old, f := sessionMoveFixture(t)
	if err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, "", false, nil); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "teams/other/tl-status-op.md"), "---\ntype: session-sync\n---\n")
	if err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, "", false, nil); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatal(err)
	}
	if err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, "../outside.md", true, nil); err == nil {
		t.Fatal("accepted traversal")
	}
	if _, err := os.Stat(filepath.Join(root, old)); err != nil {
		t.Fatal(err)
	}
}

func TestSessionMoveNestedSessionRegistration(t *testing.T) {
	for _, value := range []string{"session-sync", `"session-sync"`, "'session-sync'", "!!str session-sync"} {
		t.Run(value, func(t *testing.T) {
			root, old, f := sessionMoveFixture(t)
			child := filepath.Join(root, "teams/dev/tl-status-op.side/nested.md")
			writeFile(t, child, "---\ntype: "+value+"\ncouch-owner: ops:0\n---\n")
			err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, nil)
			if err == nil || !strings.Contains(err.Error(), "register") {
				t.Fatalf("unregistered nested session accepted: %v", err)
			}
			if _, err := os.Stat(filepath.Join(root, "session.md")); !os.IsNotExist(err) {
				t.Fatal("refusal wrote destination")
			}
		})
	}
}

func TestSessionMoveParentRefusals(t *testing.T) {
	for _, parent := range []string{
		"parent: '../tl-status-op.md'\n",
		"parent: \"../tl-status-op.md\"\n",
		"parent: ../tl-status-op.md # backlink\n",
		"parent: !!str ../tl-status-op.md\n",
		"ref: &ref ../tl-status-op.md\nparent: *ref\n",
		"parent: |-\n  ../tl-status-op.md\n",
		"'parent': ../tl-status-op.md\n",
	} {
		t.Run(parent, func(t *testing.T) {
			root, old, f := sessionMoveFixture(t)
			rel := "teams/dev/tl-status-op.side/01-a.md"
			writeFile(t, filepath.Join(root, rel), "---\ntype: side-doc\n"+parent+"---\nbody\n")
			err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, nil)
			if err == nil || !strings.Contains(err.Error(), rel) || !strings.Contains(err.Error(), "parent") {
				t.Fatalf("unsupported parent accepted or unnamed: %v", err)
			}
			if _, err := os.Stat(filepath.Join(root, ".gitignore")); !os.IsNotExist(err) {
				t.Fatal("parent refusal changed ignore file")
			}
			if _, err := os.Stat(filepath.Join(root, "session.md")); !os.IsNotExist(err) {
				t.Fatal("parent refusal published copies")
			}
		})
	}
}

func TestSessionMovePreservesOrdinaryOwner(t *testing.T) {
	root, old, f := sessionMoveFixture(t)
	writeFile(t, filepath.Join(root, "teams/dev/tl-status-op.side/01-a.md"), "---\ntype: side-doc\nparent: ../tl-status-op.md\nowner: {team: devinfra}\noperator: [one, two]\n---\nbody\n")
	if err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, nil); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "session.side/01-a.md"))
	if err != nil || !strings.Contains(string(b), "owner: {team: devinfra}\noperator: [one, two]") {
		t.Fatalf("business metadata changed: %s %v", b, err)
	}
}

func TestSessionMoveCorpus(t *testing.T) {
	root := t.TempDir()
	root, _ = filepath.EvalSymlinks(root)
	if err := os.CopyFS(root, os.DirFS("testdata/status-move/before")); err != nil {
		t.Fatal(err)
	}
	gitInit(t, root)
	old := "teams/dev/tl-status-op.md"
	f := newFake(nil)
	f.registered = map[string]bool{filepath.Join(root, old): true}
	collect := func(root string) map[string]string {
		out := map[string]string{}
		if err := filepath.WalkDir(root, func(p string, e fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if e.IsDir() {
				if e.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			rel, _ := filepath.Rel(root, p)
			b, err := os.ReadFile(p)
			out[filepath.ToSlash(rel)] = string(b)
			return err
		}); err != nil {
			t.Fatal(err)
		}
		return out
	}
	before := collect(root)
	if err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, false, nil); err != nil {
		t.Fatal(err)
	}
	if !maps.Equal(before, collect(root)) {
		t.Fatal("dry run changed corpus")
	}
	if err := moveSession(deps{ss: f}, &bytes.Buffer{}, root, old, true, nil); err != nil {
		t.Fatal(err)
	}
	want, got := collect("testdata/status-move/after"), collect(root)
	if !maps.Equal(want, got) {
		for p, b := range want {
			if got[p] != b {
				t.Errorf("%s got %q want %q", p, got[p], b)
			}
		}
		for p := range got {
			if _, ok := want[p]; !ok {
				t.Errorf("unexpected %s", p)
			}
		}
	}
}
