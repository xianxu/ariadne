package main

import (
	"errors"
	"flag"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/pkg/sidedoc"
)

var updateGolden = flag.Bool("update", false, "rewrite testdata/migrate/after from the migration's output")

// The fixture's mtime for every file, so the test can tell kept mtimes
// (#455 decision 4) from fresh ones.
var fixtureTime = time.Date(2026, 10, 16, 9, 0, 0, 0, time.Local)

// migrateFixture copies testdata/migrate/before into a temp root and returns
// the root, the session file, and a fake holding the session file with the
// two registered docs.
func migrateFixture(t *testing.T) (string, string, *fakeSessionSync) {
	t.Helper()
	root := t.TempDir()
	src := filepath.Join("testdata", "migrate", "before")
	err := filepath.WalkDir(src, func(p string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		dst := filepath.Join(root, rel)
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			return err
		}
		return os.Chtimes(dst, fixtureTime, fixtureTime)
	})
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "tl-status-op.md")
	f := newFake(nil)
	f.file = file
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.load(string(b)); err != nil {
		t.Fatal(err)
	}
	if f.text() != string(b) {
		t.Fatalf("the fake does not round-trip the fixture's session file:\n%s", f.text())
	}
	f.registered = map[string]bool{
		file: true,
		filepath.Join(root, "qa/op/2026-10-17-01-s4.md"):             true,
		filepath.Join(root, "qa/op/2026-10-17-02-ci.side/01-why.md"): true,
		filepath.Join(root, "tl-status-op.side/01-existing.md"):      true,
	}
	return root, file, f
}

// tree reads every file under root, the session file taken from the fake (the
// file couch would hold).
func tree(t *testing.T, root string, f *fakeSessionSync) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		b, err := os.ReadFile(p)
		out[filepath.ToSlash(rel)] = string(b)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	out["tl-status-op.md"] = f.text()
	return out
}

func golden(t *testing.T, got map[string]string) {
	t.Helper()
	dir := filepath.Join("testdata", "migrate", "after")
	if *updateGolden {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
		for rel, text := range got {
			p := filepath.Join(dir, rel)
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	want := map[string]string{}
	_ = filepath.WalkDir(dir, func(p string, e fs.DirEntry, err error) error {
		if err == nil && !e.IsDir() {
			rel, _ := filepath.Rel(dir, p)
			b, _ := os.ReadFile(p)
			want[filepath.ToSlash(rel)] = string(b)
		}
		return nil
	})
	if g, w := slices.Sorted(maps.Keys(got)), slices.Sorted(maps.Keys(want)); !slices.Equal(g, w) {
		t.Fatalf("files:\n got %v\nwant %v", g, w)
	}
	for rel, w := range want {
		if got[rel] != w {
			t.Errorf("%s:\n--- got\n%s\n--- want\n%s", rel, got[rel], w)
		}
	}
}

// The whole migration over the fixture, against the golden tree: every
// moved doc's frontmatter, backlink, marker (by the bullet, and by the title
// for the doc whose bullet is gone) and links; the nested side folder and its
// image; the attachment; the session file's links and index; the outside
// files; the harness's dot folder left in place, so qa/op stays. Couch's
// registrations follow the moves, and every moved file keeps its mtime.
func TestMigrateAnswersGolden(t *testing.T) {
	root, file, f := migrateFixture(t)
	var out strings.Builder
	if err := migrateAnswers(deps{ss: f}, &out, file, "", false, noStep); err != nil {
		t.Fatalf("dry run: %v\n%s", err, out.String())
	}
	if got := tree(t, root, f); got["qa/op/2026-10-17-01-s4.md"] == "" || got["tl-status-op.side/2026-10-17-01-s4.md"] != "" {
		t.Fatal("the dry run wrote")
	}
	for _, want := range []string{
		"move     qa/op/2026-10-17-01-s4.md → tl-status-op.side/2026-10-17-01-s4.md  [registered, marker: bullet]",
		"move     qa/op/2026-10-16-01-old.md → tl-status-op.side/2026-10-16-01-old.md  [marker: title]",
		"left     qa/op/.claude ",
		"mention  tl-status-op.side/2026-10-17-01-s4.md:15: The sampler's log is `qa/op/log.jsonl`.",
		"mention  projects/p/project.md:3: sources: [ops:teams/x/qa/op/2026-10-17-01-s4.md]",
		"dangling qa/op/2026-10-16-01-old.md: projects/p/project.md",
		"summary: 6 moves (0 resumed), 1 left, 0 conflicts, 4 session edits, 2 outside files",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("dry run lacks %q:\n%s", want, out.String())
		}
	}

	out.Reset()
	if err := migrateAnswers(deps{ss: f}, &out, file, "", true, noStep); err != nil {
		t.Fatalf("apply: %v\n%s", err, out.String())
	}
	golden(t, tree(t, root, f))
	for _, p := range []string{"tl-status-op.side/2026-10-17-01-s4.md", "tl-status-op.side/2026-10-17-02-ci.side/assets/a.png", "tl-status-op.side/2026-10-16-01-old.md"} {
		if fi, err := os.Stat(filepath.Join(root, p)); err != nil || !fi.ModTime().Equal(fixtureTime) {
			t.Errorf("%s: mtime not kept (%v)", p, err)
		}
	}
	want := []string{file, filepath.Join(root, "tl-status-op.side/01-existing.md"),
		filepath.Join(root, "tl-status-op.side/2026-10-17-01-s4.md"), filepath.Join(root, "tl-status-op.side/2026-10-17-02-ci.side/01-why.md")}
	if got, _ := f.Registrations(); !slices.Equal(got, slices.Sorted(slices.Values(want))) {
		t.Errorf("registrations:\n%v", got)
	}
	if f.applies != 2 {
		t.Errorf("applies = %d, want 2 (the session file, the registered outside side doc)", f.applies)
	}
	// A rerun finds nothing left to do.
	out.Reset()
	if err := migrateAnswers(deps{ss: f}, &out, file, "", true, noStep); err != nil || !strings.Contains(out.String(), "summary: 0 moves") {
		t.Errorf("rerun: %v\n%s", err, out.String())
	}
}

func noStep() error { return nil }

// A stop at every step boundary, then a rerun, reaches the same tree as the
// uninterrupted run (the golden), with the same registrations: no step
// leaves a state the next run can't finish.
func TestMigrateAnswersResumesFromEveryStep(t *testing.T) {
	for k := 1; ; k++ {
		root, file, f := migrateFixture(t)
		n := 0
		stop := func() error {
			if n++; n == k {
				return errInterrupted
			}
			return nil
		}
		err := migrateAnswers(deps{ss: f}, &strings.Builder{}, file, "", true, stop)
		if err == nil {
			if k == 1 {
				t.Fatal("the run has no steps")
			}
			return // k passed the last step: every boundary was covered
		}
		if !errors.Is(err, errInterrupted) {
			t.Fatalf("stop %d: %v", k, err)
		}
		var out strings.Builder
		if err := migrateAnswers(deps{ss: f}, &out, file, "", true, noStep); err != nil {
			t.Fatalf("rerun after stop %d: %v\n%s", k, err, out.String())
		}
		golden(t, tree(t, root, f))
		if got, _ := f.Registrations(); len(got) != 4 || slices.ContainsFunc(got, func(p string) bool { return strings.Contains(p, "/qa/") }) {
			t.Errorf("stop %d: registrations %v", k, got)
		}
		// Every moved registered doc's baseline moved too, whichever run
		// renamed it (BR-1): otherwise couch shows the migration's own
		// edits to the TL as the operator's.
		for _, p := range []string{"tl-status-op.side/2026-10-17-01-s4.md", "tl-status-op.side/2026-10-17-02-ci.side/01-why.md"} {
			if !slices.Contains(f.seenFiles, filepath.Join(root, p)) {
				t.Errorf("stop %d: %s never marked seen (seen: %v)", k, p, f.seenFiles)
			}
		}
		if t.Failed() {
			t.Fatalf("stop %d diverged", k)
		}
	}
}

// Each kind of entry under qa/ has one treatment (the plan's PQ-1 table).
func TestClassifyEntry(t *testing.T) {
	for _, c := range []struct {
		rel     string
		regular bool
		want    entryClass
	}{
		{"2026-10-17-01-a.md", true, classDoc},
		{"x.side/01-b.md", true, classDoc},
		{"log.jsonl", true, classAttachment},
		{"x.side/assets/i.png", true, classAttachment},
		{".claude", false, classForeign},
		{".claude/.cc-writes/x", true, classForeign},
		{".a.md.shadow", true, classForeign},
		{"x.side/.a.sync", true, classForeign},
		{"2026-10-17-01-a.md.lock", true, classLock},
		{"2026-10-17-01-a.md.clobbered", true, classClobbered},
		{"link.md", false, classOdd},
	} {
		if got := classifyEntry(c.rel, c.regular); got != c.want {
			t.Errorf("classifyEntry(%s, %v) = %d, want %d", c.rel, c.regular, got, c.want)
		}
	}
}

// Every conflict refuses the whole run before any write, and names the path
// and what to do.
func TestMigrateAnswersConflicts(t *testing.T) {
	for _, c := range []struct {
		name, want string
		setup      func(t *testing.T, root string, f *fakeSessionSync)
	}{
		{"a different file at a destination", "already exists with other content", func(t *testing.T, root string, _ *fakeSessionSync) {
			writeFile(t, filepath.Join(root, "tl-status-op.side/2026-10-17-02-ci.md"), "someone else's\n")
		}},
		{"an operator lock file", "the operator holds this doc's lock", func(t *testing.T, root string, _ *fakeSessionSync) {
			writeFile(t, filepath.Join(root, "qa/op/2026-10-17-02-ci.md.lock"), "holder: operator\n")
		}},
		{"clobber evidence", "clobber evidence", func(t *testing.T, root string, _ *fakeSessionSync) {
			writeFile(t, filepath.Join(root, "qa/op/2026-10-17-02-ci.md.clobbered"), "x\n")
		}},
		{"a symlink", "not a regular file", func(t *testing.T, root string, _ *fakeSessionSync) {
			if err := os.Symlink("2026-10-17-02-ci.md", filepath.Join(root, "qa/op/2026-10-17-09-link.md")); err != nil {
				t.Fatal(err)
			}
		}},
		{"a registered doc the operator holds", "the operator holds this doc's lock", func(t *testing.T, root string, f *fakeSessionSync) {
			f.lockedFiles = map[string]bool{filepath.Join(root, "qa/op/2026-10-17-01-s4.md"): true}
		}},
		{"the session file locked", "the session file's lock", func(_ *testing.T, root string, f *fakeSessionSync) {
			f.lockedFiles = map[string]bool{filepath.Join(root, "tl-status-op.md"): true}
		}},
		{"an unclosed frontmatter", "not closed", func(t *testing.T, root string, _ *fakeSessionSync) {
			writeFile(t, filepath.Join(root, "qa/op/2026-10-17-03-broken.md"), "---\ntype: x\n# no end\n")
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			root, file, f := migrateFixture(t)
			c.setup(t, root, f)
			before := tree(t, root, f)
			var out strings.Builder
			err := migrateAnswers(deps{ss: f}, &out, file, "", true, noStep)
			if exitCode(err) != 1 || !strings.Contains(out.String(), c.want) || !strings.Contains(out.String(), "nothing was written") {
				t.Fatalf("exit %d, want 1 naming %q:\n%s", exitCode(err), c.want, out.String())
			}
			if after := tree(t, root, f); !maps.Equal(before, after) || f.applies != 0 {
				t.Error("a refused run wrote")
			}
		})
	}
}

// The operator deletes the bullet of a doc a stopped run already created:
// the rerun keeps that doc's marker (pinned from the destination) instead of
// reading the destination as someone else's file.
func TestMigrateAnswersResumeKeepsMarker(t *testing.T) {
	root, file, f := migrateFixture(t)
	n := 0
	stop := func() error {
		if n++; n == 3 { // after creating the s4 answer (the second move)
			return errInterrupted
		}
		return nil
	}
	if err := migrateAnswers(deps{ss: f}, &strings.Builder{}, file, "", true, stop); !errors.Is(err, errInterrupted) {
		t.Fatalf("stop: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "tl-status-op.side/2026-10-17-01-s4.md")); err != nil {
		t.Fatalf("the stop came before s4 was created: %v", err)
	}
	f.sections[instructionsHeading] = strings.Replace(f.sections[instructionsHeading], "1. what does S4 cost? → [answer](qa/op/2026-10-17-01-s4.md)\n   a follow-up line\n", "", 1)
	var out strings.Builder
	if err := migrateAnswers(deps{ss: f}, &out, file, "", true, noStep); err != nil {
		t.Fatalf("rerun: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "resume   qa/op/2026-10-17-01-s4.md → tl-status-op.side/2026-10-17-01-s4.md  [registered, marker: kept]") {
		t.Errorf("rerun output:\n%s", out.String())
	}
	b, _ := os.ReadFile(filepath.Join(root, "tl-status-op.side/2026-10-17-01-s4.md"))
	if !strings.Contains(string(b), "\nmarker: \"what does S4 cost?\"\n") {
		t.Errorf("the marker changed:\n%s", b)
	}
}

func writeFile(t *testing.T, p, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// The operator takes the session file's lock while the files move: its one
// apply refuses with the reason, the links wait, and a rerun after their
// submit finishes to the golden tree.
func TestMigrateAnswersLockMidRun(t *testing.T) {
	root, file, f := migrateFixture(t)
	n := 0
	lock := func() error {
		if n++; n == 2 {
			f.lockedFiles = map[string]bool{file: true}
		}
		return nil
	}
	err := migrateAnswers(deps{ss: f}, &strings.Builder{}, file, "", true, lock)
	if err == nil || !strings.Contains(err.Error(), "the operator holds its lock") {
		t.Fatalf("err = %v, want the session file's lock named", err)
	}
	if strings.Contains(f.text(), "](tl-status-op.side/") {
		t.Error("the session file was written under the operator's lock")
	}
	f.lockedFiles = nil
	if err := migrateAnswers(deps{ss: f}, &strings.Builder{}, file, "", true, noStep); err != nil {
		t.Fatalf("rerun: %v", err)
	}
	golden(t, tree(t, root, f))
}

// A stopped run already moved a doc's registration, and the operator now
// holds the doc at its new path: the rerun refuses rather than write past
// the lock (BR-1). The doc is the nested side doc, which never names qa/, so
// only the side-folder clause of the lock read can find it.
func TestMigrateAnswersLockAtDestination(t *testing.T) {
	root, file, f := migrateFixture(t)
	moved := filepath.Join(root, "tl-status-op.side/2026-10-17-02-ci.side/01-why.md")
	stop := func() error {
		if f.registered[moved] {
			return errInterrupted // right after its rename
		}
		return nil
	}
	if err := migrateAnswers(deps{ss: f}, &strings.Builder{}, file, "", true, stop); !errors.Is(err, errInterrupted) {
		t.Fatalf("stop: %v", err)
	}
	if b, _ := os.ReadFile(moved); strings.Contains(string(b), "qa/op") {
		t.Fatal("the fixture doc names qa/; it can't isolate the side-folder clause")
	}
	f.lockedFiles = map[string]bool{moved: true}
	var out strings.Builder
	if err := migrateAnswers(deps{ss: f}, &out, file, "", true, noStep); exitCode(err) != 1 || !strings.Contains(out.String(), "conflict qa/op/2026-10-17-02-ci.side/01-why.md: the operator holds this doc's lock") {
		t.Fatalf("exit %d:\n%s", exitCode(err), out.String())
	}
}

// Entries are relative to the file that holds them, like its links: an
// outside file's ../../qa/... entry is rebased from its own folder.
func TestRebaseIndexIsFileRelative(t *testing.T) {
	moved := func(p string) (string, bool) { return sidedoc.MoveUnder(p, "qa/op", "s.side") }
	in := "---\ndiscussions:\n  - doc: ../../qa/op/a.md\n  - doc: design.side/01-x.md\n---\nbody\n"
	want := "---\ndiscussions:\n  - doc: ../../s.side/a.md\n  - doc: design.side/01-x.md\n---\nbody\n"
	if got, err := sidedoc.RebaseIndex(in, "projects/p/design.md", "projects/p/design.md", moved); err != nil || got != want {
		t.Errorf("got\n%s\nerror: %v", got, err)
	}
	if got, err := sidedoc.RebaseIndex("---\ndiscussions:\n  - doc: qa/op/a.md\n---\n", "s.md", "s.md", moved); err != nil || !strings.Contains(got, "- doc: s.side/a.md") {
		t.Errorf("the session file's entry: %s; error: %v", got, err)
	}
}

// --from stays inside the session file's directory.
func TestMigrateAnswersFromStaysInside(t *testing.T) {
	_, file, f := migrateFixture(t)
	for _, from := range []string{"../elsewhere", "/tmp/qa", "..", "."} {
		if err := migrateAnswers(deps{ss: f}, &strings.Builder{}, file, from, false, noStep); err == nil || !strings.Contains(err.Error(), "must be a folder inside") {
			t.Errorf("--from %s: %v", from, err)
		}
	}
}

// Where git ignores qa/ but would track the side folder, the plan warns; once
// both are ignored it doesn't.
func TestMigrateAnswersGitignoreFollows(t *testing.T) {
	root, file, f := migrateFixture(t)
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	writeFile(t, filepath.Join(root, ".gitignore"), "qa/\n")
	var out strings.Builder
	if err := migrateAnswers(deps{ss: f}, &out, file, "", false, noStep); err != nil || !strings.Contains(out.String(), "warning  git ignores qa/op/ but would track tl-status-op.side/") {
		t.Fatalf("err %v:\n%s", err, out.String())
	}
	writeFile(t, filepath.Join(root, ".gitignore"), "qa/\ntl-status-*.side/\n")
	out.Reset()
	if err := migrateAnswers(deps{ss: f}, &out, file, "", false, noStep); err != nil || strings.Contains(out.String(), "warning") {
		t.Errorf("both ignored, still warns:\n%s", out.String())
	}
}

// A run killed mid-write leaves its temporary file; the next run removes it.
func TestMigrateAnswersSweepsItsTemps(t *testing.T) {
	root, file, f := migrateFixture(t)
	stale := filepath.Join(root, "tl-status-op.side", tempPrefix+"123")
	writeFile(t, stale, "half")
	if err := migrateAnswers(deps{ss: f}, &strings.Builder{}, file, "", true, noStep); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("the stale temp survived: %v", err)
	}
	golden(t, tree(t, root, f))
}

// #576: answerFrontmatter writes no mode key, and keeps every legacy one a
// doc already has, verbatim and in its order.
func TestAnswerFrontmatterWritesNoModeKeys(t *testing.T) {
	legacy := "---\ncollaboration-mode: discussion\npresentation-mode: highlight\nprocess-instruction: \"[free-form] x\"\nowner: op\n---\n# Q\n\nA.\n"
	got, err := answerFrontmatter(legacy, "../s.md", `"q"`, "S")
	if err != nil {
		t.Fatal(err)
	}
	want := "---\ntype: side-doc\nparent: ../s.md\nmarker: \"q\"\ncollaboration-mode: discussion\npresentation-mode: highlight\nprocess-instruction: \"[free-form] x\"\nowner: op\n---\n"
	if !strings.HasPrefix(got, want) {
		t.Errorf("legacy doc:\n got %q\nwant prefix %q", got, want)
	}
	fresh, err := answerFrontmatter("# Q\n\nA.\n", "../s.md", `"q"`, "S")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(fresh, "---\ntype: side-doc\nparent: ../s.md\nmarker: \"q\"\n---\n") {
		t.Errorf("a doc without frontmatter gains more than type/parent/marker:\n%s", fresh)
	}
}
