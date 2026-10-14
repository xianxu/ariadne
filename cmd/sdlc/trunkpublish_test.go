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
		"workshop/parley/a\nb.md", "workshop/parley/a\rb.md",
		"workshop/parley/a\tb.md", "workshop/parley/a\x00b.md",
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
	if got, err := os.ReadFile(filepath.Join(r.root, rel)); err != nil || string(got) != "selected prose\n" {
		t.Fatalf("selected local file changed: %q, %v", got, err)
	}
	if got := r.git("status", "--porcelain", "--untracked-files=all"); !strings.Contains(got, "unrelated.go") || !strings.Contains(got, rel) {
		t.Fatalf("wrong worktree retention: %s", got)
	}
	if !strings.Contains(errs, trunkRestoreCommand(rel)) {
		t.Fatalf("missing exact-path cleanup guidance: %s", errs)
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

func TestTrunkPublishRetryAfterPushInterruption(t *testing.T) {
	r, _ := claimSetRepo(t)
	r.git("switch", "-q", "-c", "000009-s09")
	rel := "workshop/pensive/retry.md"
	writeRepoFile(t, r.root, rel, "published\n")
	before := r.originMain()
	old := trunkPublishAfterPush
	trunkPublishAfterPush = func() error {
		return errors.New("injected interruption after push")
	}
	var out, errs bytes.Buffer
	err := runTrunkPublish(context.Background(), &out, &errs, rel)
	trunkPublishAfterPush = old
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
	if got, err := os.ReadFile(filepath.Join(r.root, rel)); err != nil || string(got) != "published\n" {
		t.Fatalf("retry changed local copy: %q, %v", got, err)
	}
	if !strings.Contains(errs.String(), trunkRestoreCommand(rel)) {
		t.Fatalf("retry omitted cleanup guidance: %s", errs.String())
	}
}

func TestTrunkPublishPreservesOpenDescriptorWriteAfterPush(t *testing.T) {
	r, _ := claimSetRepo(t)
	r.git("switch", "-q", "-c", "000009-s09")
	rel := "workshop/pensive/open-editor.md"
	writeRepoFile(t, r.root, rel, "published\n")
	file, err := os.OpenFile(filepath.Join(r.root, rel), os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	old := trunkPublishAfterPush
	trunkPublishAfterPush = func() error {
		_, err := file.WriteString("later edit\n")
		return err
	}
	defer func() { trunkPublishAfterPush = old }()
	var out, errs bytes.Buffer
	if err := runTrunkPublish(context.Background(), &out, &errs, rel); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if got := r.git("show", r.originMain()+":"+rel); got != "published" {
		t.Fatalf("main has %q", got)
	}
	if got, err := os.ReadFile(filepath.Join(r.root, rel)); err != nil || string(got) != "published\nlater edit\n" {
		t.Fatalf("open-descriptor edit lost: %q, %v", got, err)
	}
	if strings.Contains(errs.String(), "git restore --") || !strings.Contains(errs.String(), "changed while publishing") {
		t.Fatalf("unsafe cleanup guidance after later edit: %s", errs.String())
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
	if got, err := os.ReadFile(filepath.Join(r.root, rel)); err != nil || string(got) != "revised\n" {
		t.Fatalf("local tracked copy changed: %q, %v", got, err)
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
	if got, err := os.ReadFile(filepath.Join(r.root, selected)); err != nil || string(got) != "selected edit\n" {
		t.Fatalf("selected path changed: %q, %v", got, err)
	}
	if !strings.Contains(errs.String(), trunkRestoreCommand(selected)) {
		t.Fatalf("wildcard filename missing literal cleanup command: %s", errs.String())
	}
}

func TestTrunkRestoreCommandQuotesShellAndGitMetacharacters(t *testing.T) {
	got := trunkRestoreCommand("workshop/parley/a'b*.md")
	want := "git restore -- ':(literal)workshop/parley/a'\\''b*.md'"
	if got != want {
		t.Fatalf("restore command = %q, want %q", got, want)
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
