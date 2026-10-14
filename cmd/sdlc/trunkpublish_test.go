package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTrunkProsePath(t *testing.T) {
	for _, name := range []string{
		"workshop/parley/note.md", "workshop/pensive/a/b.md",
		"workshop/targets/one.md", "workshop/projects/roadmap/202610/a.md",
		"workshop/lessons.md",
	} {
		if err := trunkProsePath(name); err != nil {
			t.Errorf("valid %q: %v", name, err)
		}
	}
	for _, name := range []string{
		"", "../workshop/parley/a.md", "/workshop/parley/a.md",
		"workshop/parley/../issues/a.md", "workshop/parley/a.txt",
		"workshop/issues/a.md", "workshop/plans/a.md", "atlas/workflow/a.md",
		".claude/settings.ariadne.json", "workshop/lessons.md/child.md",
	} {
		if err := trunkProsePath(name); err == nil {
			t.Errorf("accepted ineligible %q", name)
		}
	}
}

func TestProsePublishDecision(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		base, main, local       string
		baseOK, mainOK, localOK bool
		want                    publishVerdict
	}{
		{"new", "", "", "new", false, false, true, publishWrite},
		{"edit", "old", "old", "new", true, true, true, publishWrite},
		{"already", "old", "new", "new", true, true, true, publishNothing},
		{"moved", "old", "other", "new", true, true, true, publishRefuse},
		{"new-collision", "", "other", "new", false, true, true, publishRefuse},
		{"absent", "", "", "", false, false, false, publishNothing},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := prosePublishDecision([]byte(tc.local), tc.localOK, []byte(tc.base), tc.baseOK, []byte(tc.main), tc.mainOK)
			if got != tc.want {
				t.Fatalf("decision = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTrunkPublishFromDirtyIssueBranch(t *testing.T) {
	r, _ := claimSetRepo(t)
	r.git("switch", "-q", "-c", "000009-s09")
	rel := "workshop/parley/trunk-note.md"
	if err := os.MkdirAll(filepath.Join(r.root, "workshop/parley"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(r.root, rel), []byte("selected prose\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeRepoFile(t, r.root, "unrelated.go", "package unrelated\n")
	head := r.git("rev-parse", "HEAD")
	_, errs, err := executeSDLCTestCommand("trunk", "publish", "--path", rel)
	if err != nil {
		t.Fatalf("publish: %v\n%s", err, errs)
	}
	if got := r.git("show", r.originMain()+":"+rel); got != "selected prose" {
		t.Fatalf("main has %q", got)
	}
	if got := r.git("rev-parse", "HEAD"); got != head {
		t.Fatalf("issue branch moved: %s to %s", head, got)
	}
	if _, err := os.Stat(filepath.Join(r.root, rel)); !os.IsNotExist(err) {
		t.Fatalf("selected untracked file remains: %v", err)
	}
	if got := r.git("status", "--porcelain"); !strings.Contains(got, "unrelated.go") || strings.Contains(got, rel) {
		t.Fatalf("wrong worktree cleanup: %s", got)
	}
}

func TestTrunkPublishRefusesMovedMainPath(t *testing.T) {
	r, _ := claimSetRepo(t)
	rel := "workshop/parley/shared.md"
	writeRepoFile(t, r.root, rel, "base\n")
	r.git("add", "--", rel)
	r.git("commit", "-q", "-m", "seed shared prose")
	r.git("push", "-q", "origin", "HEAD:main")
	r.git("switch", "-q", "-c", "000009-s09")
	writeRepoFile(t, r.root, rel, "local\n")
	main := t.TempDir()
	r.git("clone", "-q", r.origin, main)
	writeRepoFile(t, main, rel, "peer\n")
	gitInTest(t, main, "add", "--", rel)
	gitInTest(t, main, "commit", "-q", "-m", "peer edit")
	gitInTest(t, main, "push", "-q", "origin", "HEAD:main")
	before := r.originMain()
	var out, errs bytes.Buffer
	err := runTrunkPublish(context.Background(), &out, &errs, rel)
	if err == nil || !strings.Contains(err.Error(), "changed on main") {
		t.Fatalf("moved main accepted: %v\n%s", err, errs.String())
	}
	if got := r.originMain(); got != before {
		t.Fatalf("refused publish changed main: %s to %s", before, got)
	}
	if got := r.git("status", "--porcelain", "--", rel); got == "" {
		t.Fatal("refusal lost the local edit")
	}
}

func TestTrunkPublishDoesNotRestoreAfterHeadMoves(t *testing.T) {
	r, _ := claimSetRepo(t)
	rel := "workshop/parley/shared.md"
	writeRepoFile(t, r.root, rel, "base\n")
	r.git("add", "--", rel)
	r.git("commit", "-q", "-m", "seed shared prose")
	r.git("push", "-q", "origin", "HEAD:main")
	r.git("switch", "-q", "-c", "000009-s09")
	writeRepoFile(t, r.root, rel, "published\n")
	pinned := r.git("rev-parse", "HEAD")
	// A new HEAD after preflight must not become the restore source.
	writeRepoFile(t, r.root, "unrelated.txt", "new\n")
	r.git("add", "--", "unrelated.txt")
	r.git("commit", "-q", "-m", "move head")
	if err := finishTrunkPublish(context.Background(), r.root, rel, []byte("published\n"), true, pinned); err == nil || !strings.Contains(err.Error(), "HEAD moved") {
		t.Fatalf("cleanup used a moved HEAD: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(r.root, rel)); err != nil || string(got) != "published\n" {
		t.Fatalf("cleanup clobbered local bytes: %q, %v", got, err)
	}
}

func TestReadLocalProseDistinguishesFinalComponent(t *testing.T) {
	root := t.TempDir()
	rel := "workshop/projects/note.md/note.md"
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte("nested prose\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, ok, err := readLocalProse(root, rel)
	if err != nil || !ok || string(b) != "nested prose\n" {
		t.Fatalf("read nested prose: %q, %v, %v", b, ok, err)
	}
}

func TestTrunkPublishRetryAfterCleanupInterruption(t *testing.T) {
	r, _ := claimSetRepo(t)
	r.git("switch", "-q", "-c", "000009-s09")
	rel := "workshop/pensive/retry.md"
	writeRepoFile(t, r.root, rel, "published\n")
	before := r.originMain()
	old := finishPublishedProse
	finishPublishedProse = func(context.Context, string, string, []byte, bool, string) error {
		return errors.New("injected cleanup interruption")
	}
	var out, errs bytes.Buffer
	err := runTrunkPublish(context.Background(), &out, &errs, rel)
	finishPublishedProse = old
	if err == nil || !strings.Contains(err.Error(), "on main") {
		t.Fatalf("lost confirmed publication: %v", err)
	}
	landed := r.originMain()
	if landed == before {
		t.Fatal("test did not publish before cleanup interruption")
	}
	if err := runTrunkPublish(context.Background(), &out, &errs, rel); err != nil {
		t.Fatalf("retry: %v, %s", err, out.String())
	}
	if got := r.originMain(); got != landed {
		t.Fatalf("retry republished: %s to %s", landed, got)
	}
	if _, err := os.Stat(filepath.Join(r.root, rel)); !os.IsNotExist(err) {
		t.Fatalf("retry did not clean local copy: %v", err)
	}
}

func TestTrunkPublishPreservesTrackedBranchAndIndex(t *testing.T) {
	r, _ := claimSetRepo(t)
	rel := "workshop/targets/stable.md"
	writeRepoFile(t, r.root, rel, "base\n")
	r.git("add", "--", rel)
	r.git("commit", "-q", "-m", "seed target")
	r.git("push", "-q", "origin", "HEAD:main")
	r.git("switch", "-q", "-c", "000009-s09")
	writeRepoFile(t, r.root, rel, "revised\n")
	writeRepoFile(t, r.root, "unrelated.txt", "staged\n")
	r.git("add", "--", "unrelated.txt")
	beforeIndex := r.git("write-tree")
	head := r.git("rev-parse", "HEAD")
	var out, errs bytes.Buffer
	if err := runTrunkPublish(context.Background(), &out, &errs, rel); err != nil {
		t.Fatalf("publish tracked: %v\n%s", err, errs.String())
	}
	if got := r.git("show", r.originMain()+":"+rel); got != "revised" {
		t.Fatalf("main has %q", got)
	}
	if got := r.git("rev-parse", "HEAD"); got != head {
		t.Fatal("issue HEAD moved")
	}
	if got := r.git("write-tree"); got != beforeIndex {
		t.Fatal("unrelated staged index changed")
	}
	if got, err := os.ReadFile(filepath.Join(r.root, rel)); err != nil || string(got) != "base\n" {
		t.Fatalf("local tracked copy not restored: %q, %v", got, err)
	}
}

func TestTrunkPublishRefusesStagedAndCommittedPaths(t *testing.T) {
	for _, kind := range []string{"staged", "committed"} {
		t.Run(kind, func(t *testing.T) {
			r, _ := claimSetRepo(t)
			r.git("switch", "-q", "-c", "000009-s09")
			rel := "workshop/parley/owned.md"
			writeRepoFile(t, r.root, rel, "owned\n")
			r.git("add", "--", rel)
			if kind == "committed" {
				r.git("commit", "-q", "-m", "branch owns prose")
				writeRepoFile(t, r.root, rel, "more\n")
			}
			before := r.originMain()
			var out, errs bytes.Buffer
			err := runTrunkPublish(context.Background(), &out, &errs, rel)
			if err == nil || !strings.Contains(err.Error(), kind) {
				t.Fatalf("%s path accepted: %v", kind, err)
			}
			if r.originMain() != before {
				t.Fatal("refusal moved main")
			}
		})
	}
}

func TestTrunkPublishRefusesSymlink(t *testing.T) {
	r, _ := claimSetRepo(t)
	r.git("switch", "-q", "-c", "000009-s09")
	rel := "workshop/parley/link.md"
	writeRepoFile(t, r.root, "real.md", "secret\n")
	if err := os.MkdirAll(filepath.Dir(filepath.Join(r.root, rel)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(r.root, "real.md"), filepath.Join(r.root, rel)); err != nil {
		t.Fatal(err)
	}
	before := r.originMain()
	var out, errs bytes.Buffer
	if err := runTrunkPublish(context.Background(), &out, &errs, rel); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("symlink accepted: %v", err)
	}
	if r.originMain() != before {
		t.Fatal("symlink refusal moved main")
	}
}

func TestFinishTrunkPublishPreservesLaterLocalEdit(t *testing.T) {
	r, _ := claimSetRepo(t)
	rel := "workshop/parley/later.md"
	writeRepoFile(t, r.root, rel, "later edit\n")
	head := r.git("rev-parse", "HEAD")
	err := finishTrunkPublish(context.Background(), r.root, rel, []byte("published\n"), false, head)
	if err == nil || !strings.Contains(err.Error(), "local bytes changed") {
		t.Fatalf("later edit not detected: %v", err)
	}
	if got, readErr := os.ReadFile(filepath.Join(r.root, rel)); readErr != nil || string(got) != "later edit\n" {
		t.Fatalf("later edit clobbered: %q, %v", got, readErr)
	}
}

func TestTrunkPublishTreatsSelectedWildcardFilenameLiterally(t *testing.T) {
	r, _ := claimSetRepo(t)
	selected := "workshop/parley/*.md"
	other := "workshop/parley/other.md"
	writeRepoFile(t, r.root, selected, "selected base\n")
	writeRepoFile(t, r.root, other, "other base\n")
	r.git("add", "--", ":(literal)"+selected, ":(literal)"+other)
	r.git("commit", "-q", "-m", "seed literal filenames")
	r.git("push", "-q", "origin", "HEAD:main")
	r.git("switch", "-q", "-c", "000009-s09")
	writeRepoFile(t, r.root, selected, "selected edit\n")
	writeRepoFile(t, r.root, other, "other edit\n")
	var out, errs bytes.Buffer
	if err := runTrunkPublish(context.Background(), &out, &errs, selected); err != nil {
		t.Fatalf("publish literal path: %v\n%s", err, errs.String())
	}
	if got, err := os.ReadFile(filepath.Join(r.root, other)); err != nil || string(got) != "other edit\n" {
		t.Fatalf("unselected wildcard match was changed: %q, %v", got, err)
	}
	if got, err := os.ReadFile(filepath.Join(r.root, selected)); err != nil || string(got) != "selected base\n" {
		t.Fatalf("selected path not restored: %q, %v", got, err)
	}
}

func TestFinishTrunkPublishPreservesEditDuringCleanup(t *testing.T) {
	for _, tracked := range []bool{false, true} {
		name := "untracked"
		if tracked {
			name = "tracked"
		}
		t.Run(name, func(t *testing.T) {
			r, _ := claimSetRepo(t)
			rel := "workshop/parley/concurrent.md"
			if tracked {
				writeRepoFile(t, r.root, rel, "base\n")
				r.git("add", "--", rel)
				r.git("commit", "-q", "-m", "seed")
			}
			writeRepoFile(t, r.root, rel, "published\n")
			head := r.git("rev-parse", "HEAD")
			old := trunkProseAfterMove
			trunkProseAfterMove = func(_, backup string) {
				if err := os.WriteFile(backup, []byte("later edit\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			defer func() { trunkProseAfterMove = old }()
			err := finishTrunkPublish(context.Background(), r.root, rel, []byte("published\n"), tracked, head)
			if err == nil || !strings.Contains(err.Error(), "changed") {
				t.Fatalf("intervening edit not detected: %v", err)
			}
			if got, readErr := os.ReadFile(filepath.Join(r.root, rel)); readErr != nil || string(got) != "later edit\n" {
				t.Fatalf("intervening edit lost: %q, %v", got, readErr)
			}
		})
	}
}

func TestFinishTrunkPublishPreservesNewPathDuringCleanup(t *testing.T) {
	for _, tracked := range []bool{false, true} {
		name := "untracked"
		if tracked {
			name = "tracked"
		}
		t.Run(name, func(t *testing.T) {
			r, _ := claimSetRepo(t)
			rel := "workshop/parley/replaced.md"
			if tracked {
				writeRepoFile(t, r.root, rel, "base\n")
				r.git("add", "--", rel)
				r.git("commit", "-q", "-m", "seed")
			}
			writeRepoFile(t, r.root, rel, "published\n")
			head := r.git("rev-parse", "HEAD")
			old := trunkProseAfterMove
			trunkProseAfterMove = func(original, _ string) {
				if err := os.WriteFile(original, []byte("new writer\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			defer func() { trunkProseAfterMove = old }()
			if err := finishTrunkPublish(context.Background(), r.root, rel, []byte("published\n"), tracked, head); err == nil || !strings.Contains(err.Error(), "another writer") {
				t.Fatalf("new path not detected: %v", err)
			}
			if got, err := os.ReadFile(filepath.Join(r.root, rel)); err != nil || string(got) != "new writer\n" {
				t.Fatalf("new writer lost: %q, %v", got, err)
			}
		})
	}
}

func TestFinishTrunkPublishChecksHeadAndIndexAfterMove(t *testing.T) {
	for _, kind := range []string{"HEAD", "index"} {
		t.Run(kind, func(t *testing.T) {
			r, _ := claimSetRepo(t)
			rel := "workshop/parley/ordering.md"
			writeRepoFile(t, r.root, rel, "base\n")
			r.git("add", "--", rel)
			r.git("commit", "-q", "-m", "seed")
			writeRepoFile(t, r.root, rel, "published\n")
			head := r.git("rev-parse", "HEAD")
			old := trunkProseAfterMove
			trunkProseAfterMove = func(_, _ string) {
				if kind == "HEAD" {
					writeRepoFile(t, r.root, "later.txt", "later\n")
					r.git("add", "--", "later.txt")
					r.git("commit", "-q", "-m", "move HEAD")
				} else {
					r.git("add", "-u", "--", rel)
				}
			}
			defer func() { trunkProseAfterMove = old }()
			err := finishTrunkPublish(context.Background(), r.root, rel, []byte("published\n"), true, head)
			if err == nil || !strings.Contains(err.Error(), kind+" moved") && !strings.Contains(err.Error(), kind+" changed") {
				t.Fatalf("%s mutation not detected: %v", kind, err)
			}
			if got, err := os.ReadFile(filepath.Join(r.root, rel)); err != nil || string(got) != "published\n" {
				t.Fatalf("local bytes lost: %q, %v", got, err)
			}
		})
	}
}

func gitInTest(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := append([]string{"-C", dir}, args...)
	out, err := gitForProse(context.Background(), dir, args...)
	if err != nil {
		t.Fatalf("git %v: %v", cmd, err)
	}
	return string(out)
}
