package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/sidedoc"
)

const indexOps = "1. what does S4 cost?\n2. how big is the backlog?\n3. why did CI fail?"

// indexFake is a status file with the operator's section and Needs you, in
// couch's whole-file shape, and the answer docs on disk beside it.
func indexFake(t *testing.T) (*fakeSessionSync, string) {
	t.Helper()
	f := newFake(map[string]string{instructionsHeading: indexOps, "Needs you": "- nothing yet"})
	f.front = "type: session-sync"
	f.order = []string{instructionsHeading, "Needs you"}
	file, _ := statusFile(t, "")
	side := filepath.Join(filepath.Dir(file), "tl-status-op.side")
	if err := os.MkdirAll(side, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"2026-10-17-01-s4.md", "2026-10-17-02-backlog.md", "2026-10-17-03-why.md"} {
		answerDoc(t, file, n)
	}
	if err := os.WriteFile(filepath.Join(side, "01-why.md"), []byte("answer\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return f, file
}

func entries(t *testing.T, f *fakeSessionSync) []string {
	t.Helper()
	e, err := sidedoc.Entries(f.text())
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func link(t *testing.T, f *fakeSessionSync, file, pos, target string) (string, error) {
	t.Helper()
	return run(t, deps{ss: f}, "status", "instructions", "link", file, pos, target, "--hash", fakeHash(f.sections[instructionsHeading]))
}

// The one-way rule over a sequence of TL and operator events (#424): link
// adds an entry; the operator's removal of an entry is reviewed and never
// undone; the operator's deletion of a bullet drops its entry at submit.
func TestIndexOneWaySequence(t *testing.T) {
	f, file := indexFake(t)
	s4, backlog := "tl-status-op.side/2026-10-17-01-s4.md", "tl-status-op.side/2026-10-17-02-backlog.md"
	if out, err := link(t, f, file, "1", s4); err != nil || !slices.Equal(entries(t, f), []string{s4}) {
		t.Fatalf("link 1: %v %v\n%s", err, entries(t, f), out)
	}
	if _, err := link(t, f, file, "2", backlog); err != nil || !slices.Equal(entries(t, f), []string{s4, backlog}) {
		t.Fatalf("link 2 appends: %v %v", err, entries(t, f))
	}
	// The operator reviews s4: they remove its entry, and the link stays.
	f.front = strings.Replace(f.front, "  - doc: "+s4+"\n", "", 1)
	if out, err := run(t, deps{ss: f}, "status", "submit", file); err != nil || !slices.Equal(entries(t, f), []string{backlog}) {
		t.Fatalf("a reviewed entry came back at submit: %v %v\n%s", err, entries(t, f), out)
	}
	// The operator deletes bullet 2 (and with it its link).
	f.sections[instructionsHeading] = "1. what does S4 cost? → [answer](" + s4 + ")\n3. why did CI fail?"
	out, err := run(t, deps{ss: f}, "status", "submit", file)
	if err != nil || len(entries(t, f)) != 0 || !strings.Contains(out, "index: dropped "+backlog) {
		t.Fatalf("a deleted bullet kept its entry: %v %v\n%s", err, entries(t, f), out)
	}
	// A later link of another bullet doesn't resurrect the reviewed one.
	if _, err := link(t, f, file, "2", "tl-status-op.side/2026-10-17-03-why.md"); err != nil || !slices.Equal(entries(t, f), []string{"tl-status-op.side/2026-10-17-03-why.md"}) {
		t.Fatalf("link 3: %v %v", err, entries(t, f))
	}
}

// The two writes of link split: the entry lands and the link's write goes
// stale. Rerunning link completes it without a second entry; never rerunning
// lets the next reconcile drop the entry, so nothing lingers.
func TestIndexLinkSplit(t *testing.T) {
	for _, rerun := range []bool{true, false} {
		f, file := indexFake(t)
		s4 := "tl-status-op.side/2026-10-17-01-s4.md"
		f.editBefore[1] = map[string]string{instructionsHeading: indexOps + "\n4. a new bullet"}
		if _, err := link(t, f, file, "1", s4); exitCode(err) != 6 || !slices.Equal(entries(t, f), []string{s4}) {
			t.Fatalf("split: exit %d, entries %v", exitCode(err), entries(t, f))
		}
		if rerun {
			if _, err := link(t, f, file, "1", s4); err != nil || !slices.Equal(entries(t, f), []string{s4}) || !strings.Contains(f.sections[instructionsHeading], "→ [answer]("+s4+")") {
				t.Fatalf("rerun: %v %v\n%s", err, entries(t, f), f.sections[instructionsHeading])
			}
			continue
		}
		if _, err := run(t, deps{ss: f}, "status", "index", file); err != nil || len(entries(t, f)) != 0 {
			t.Fatalf("an entry without a link lingered: %v %v", err, entries(t, f))
		}
	}
}

// --add lists a discussion side doc linked in another section; a missing
// file is named, not dropped; the lock and a racing section write behave as
// couch's guards do; an index the reconcile can't read refuses.
func TestIndexCommand(t *testing.T) {
	f, file := indexFake(t)
	why := "tl-status-op.side/01-why.md"
	f.sections["Needs you"] = "- ship #412 now? → [discussion](" + why + ") → [discussion](tl-status-op.side/09-gone.md)"
	out, err := run(t, deps{ss: f}, "status", "index", file, "--add", why, "--add", "tl-status-op.side/09-gone.md")
	if err != nil || !slices.Equal(entries(t, f), []string{why, "tl-status-op.side/09-gone.md"}) || !strings.Contains(out, "09-gone.md names a missing file") {
		t.Fatalf("--add: %v %v\n%s", err, entries(t, f), out)
	}
	// The operator deletes 09's link, and a frontmatter edit races the
	// frontmatter write that drops its entry: stale, then recomputed on the
	// new frontmatter, which keeps the racing edit.
	f.sections["Needs you"] = "- ship #412 now? → [discussion](" + why + ")"
	f.editFrontBefore = map[int]string{f.frontWriteCalls + 1: f.front + "\nnote: racing"}
	if out, err := run(t, deps{ss: f}, "status", "index", file); err != nil || !slices.Equal(entries(t, f), []string{why}) || f.frontWriteCalls != 3 || !strings.Contains(f.front, "note: racing") {
		t.Fatalf("a race: %v %v calls %d\n%s\n%s", err, entries(t, f), f.frontWriteCalls, f.front, out)
	}
	writes := f.frontWrites
	f.holder = "operator"
	f.sections["Needs you"] = "- nothing"
	if _, err := run(t, deps{ss: f}, "status", "index", file); exitCode(err) != 3 || f.frontWrites != writes {
		t.Errorf("under the lock: exit %d, writes %d → %d", exitCode(err), writes, f.frontWrites)
	}
	f.holder = ""
	f.front = "type: session-sync\ndiscussions:\n  - " + why
	if _, err := run(t, deps{ss: f}, "status", "index", file); err == nil || !strings.Contains(err.Error(), "not a `- doc: PATH` item") {
		t.Errorf("an unreadable index: %v", err)
	}
}

// The index never consumes the operator's unread section edits (#426): a
// reconcile with nothing to write reads nothing through couch, and a write
// marks only the frontmatter seen, so the edit stays in the diff, and nothing
// about it is printed.
func TestIndexKeepsUnreadEdits(t *testing.T) {
	f, file := indexFake(t)
	f.seen()
	f.sections["Needs you"] = "- the operator's new line"
	if _, err := run(t, deps{ss: f}, "status", "index", file); err != nil || f.frontWriteCalls != 0 {
		t.Fatalf("a no-op reconcile wrote: %v, calls %d", err, f.frontWriteCalls)
	}
	if diff, _ := f.Diff(file); !strings.Contains(diff, "## Needs you changed") {
		t.Fatalf("a no-op reconcile consumed the operator's edit; diff %q", diff)
	}
	out, err := link(t, f, file, "1", "tl-status-op.side/2026-10-17-01-s4.md")
	if err != nil || f.frontWrites != 1 || strings.Contains(out, "Needs you") {
		t.Fatalf("link: %v, writes %d\n%s", err, f.frontWrites, out)
	}
	if diff, _ := f.Diff(file); diff != "## Needs you changed\n" {
		t.Fatalf("the index write consumed the operator's edit, or left its own; diff %q", diff)
	}
}

// submit prints couch's diff once; the index write it triggers prints only
// its own report.
func TestSubmitPrintsUnreadOnce(t *testing.T) {
	f, file := indexFake(t)
	s4 := "tl-status-op.side/2026-10-17-01-s4.md"
	if _, err := link(t, f, file, "1", s4); err != nil {
		t.Fatal(err)
	}
	f.seen()
	f.sections[instructionsHeading] = "2. how big is the backlog?" // bullet 1 and its link deleted
	out, err := run(t, deps{ss: f}, "status", "submit", file)
	if err != nil || strings.Count(out, "## Operator instructions changed") != 1 || !strings.Contains(out, "index: dropped "+s4) {
		t.Fatalf("submit: %v\n%s", err, out)
	}
}

// A frontmatter line couch reads as the block's end but the reconcile does
// not (an indented "  ---") would leave part of the new index outside the
// write: syncIndex refuses instead of reporting an entry it never wrote.
func TestIndexRefusesADelimiterCouchReadsDifferently(t *testing.T) {
	f, file := indexFake(t)
	f.front = "type: session-sync\nnote: |\n  ---\n  text"
	_, err := link(t, f, file, "1", "tl-status-op.side/2026-10-17-01-s4.md")
	if err == nil || !strings.Contains(err.Error(), "couch reads as its end") || f.frontWrites != 0 {
		t.Fatalf("err %v, writes %d", err, f.frontWrites)
	}
}
