package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// Execute real Cobra gate code in a child with its own cwd and environment.
func TestBenchPrepareGateHelper(t *testing.T) {
	if os.Getenv("ARIADNE_BENCH_PREPARE_TEST_GATE") != "1" {
		return
	}
	args := os.Args
	for len(args) > 0 && args[0] != "--" {
		args = args[1:]
	}
	if len(args) == 0 {
		os.Exit(91)
	}
	root := buildRoot()
	root.SetArgs(args[1:])
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
func prepareTestGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	c.Env = benchPrepareEnv()
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}
func prepareTestWrite(t *testing.T, root, name, content string, mode os.FileMode) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}
func preparedFixtureTask(t *testing.T) bench.Task {
	t.Helper()
	task := benchFixtureTask(t)
	source := task.SourceRepo
	prepareTestGit(t, source, "init", "-q", "-b", "main")
	prepareTestGit(t, source, "config", "user.name", "Bench Fixture")
	prepareTestGit(t, source, "config", "user.email", "bench@example.invalid")
	prepareTestWrite(t, source, "src/main.txt", "broken\n", 0644)
	prepareTestWrite(t, source, "script.sh", "#!/bin/sh\nexit 0\n", 0755)
	prepareTestWrite(t, source, "workshop/issues/000001-old.md", "---\nid: 000001\nstatus: open\ncreated: 2026-10-23\n---\n\n# Original\n\n## Spec\n\n"+task.Spec+"\n\n## Done when\n\n"+task.DoneWhen+"\n\n## Plan\n\nSOLUTION SECRET\n\n## Log\n\nSOLUTION SECRET\n", 0644)
	prepareTestWrite(t, source, "workshop/history/issues/000001-archived.md", "SOLUTION SECRET", 0644)
	prepareTestWrite(t, source, "workshop/plans/000001-plan.md", "SOLUTION SECRET", 0644)
	prepareTestWrite(t, source, "workshop/history/issues/000099-sibling-hint.md", "SIBLING SOLUTION HINT", 0644)
	if err := os.Symlink("src/main.txt", filepath.Join(source, "safe-link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("src/generated.txt", filepath.Join(source, "generated-link")); err != nil {
		t.Fatal(err)
	}
	prepareTestGit(t, source, "add", "-A")
	prepareTestGit(t, source, "commit", "-qm", "parent")
	task.ParentOID = prepareTestGit(t, source, "rev-parse", "HEAD")
	task.OriginalIssueOID = prepareTestGit(t, source, "rev-parse", "HEAD:workshop/issues/000001-old.md")
	prepareTestWrite(t, source, "tests/acceptance_test.go", "package tests\n", 0644)
	prepareTestGit(t, source, "add", "tests/acceptance_test.go")
	patch := prepareTestGit(t, source, "diff", "--cached", "--", "tests/acceptance_test.go") + "\n"
	prepareTestWrite(t, source, "src/main.txt", "fixed SOLUTION SECRET\n", 0644)
	prepareTestGit(t, source, "add", "-A")
	prepareTestGit(t, source, "commit", "-qm", "landed solution")
	task.LandingOID = prepareTestGit(t, source, "rev-parse", "HEAD")
	task.Acceptance.ParentOID = task.ParentOID
	task.Acceptance.LandingOID = task.LandingOID
	task.Acceptance.Patch = filepath.Join(t.TempDir(), "hidden.patch")
	sum := sha256.Sum256([]byte(patch))
	task.Acceptance.PatchSHA256 = hex.EncodeToString(sum[:])
	if err := os.WriteFile(task.Acceptance.Patch, []byte(patch), 0600); err != nil {
		t.Fatal(err)
	}
	return task
}

func prepareFixture(t *testing.T, task bench.Task) (*benchStore, error) {
	t.Helper()
	s, err := createBenchStore(filepath.Join(t.TempDir(), "run"), task)
	if err != nil {
		t.Fatal(err)
	}
	d := benchPrepareDeps{GateArgv: []string{os.Args[0], "-test.run=^TestBenchPrepareGateHelper$", "--"}, GateEnv: []string{"ARIADNE_BENCH_PREPARE_TEST_GATE=1"}}
	return s, prepareBenchRun(context.Background(), s, d)
}

func TestBenchPrepareSanitizedIndependentAndRealTracker(t *testing.T) {
	task := preparedFixtureTask(t)
	head := prepareTestGit(t, task.SourceRepo, "rev-parse", "HEAD")
	status := prepareTestGit(t, task.SourceRepo, "status", "--porcelain")
	objects := prepareTestGit(t, task.SourceRepo, "count-objects", "-v")
	s, err := prepareFixture(t, task)
	if err != nil {
		t.Fatal(err)
	}
	record, state, err := s.read()
	if err != nil || state.Phase() != bench.PhasePrepared {
		t.Fatalf("phase: %s %v", state.Phase(), err)
	}
	work := filepath.Join(s.root, "public", "work")
	origin := filepath.Join(s.root, "public", "origin.git")
	if got := prepareTestGit(t, work, "remote", "-v"); strings.Contains(got, task.SourceRepo) || !strings.Contains(got, origin) {
		t.Fatal(got)
	}
	if got := prepareTestGit(t, work, "rev-list", "--max-parents=0", "main"); strings.Contains(got, task.ParentOID) || strings.Contains(got, task.LandingOID) {
		t.Fatal("source commit retained")
	}
	for _, root := range []string{work, origin} {
		c := exec.Command("git", "cat-file", "-e", task.LandingOID+"^{commit}")
		c.Dir = root
		c.Env = benchPrepareEnv()
		if c.Run() == nil {
			t.Fatal("landed object exists", root)
		}
	}
	if _, err := os.Lstat(filepath.Join(work, ".git", "objects", "info", "alternates")); !os.IsNotExist(err) {
		t.Fatal("alternate object storage retained")
	}
	if _, err := os.Stat(filepath.Join(work, "workshop", "history", "issues", "000001-archived.md")); !os.IsNotExist(err) {
		t.Fatal("archived answer retained")
	}
	if _, err := os.Stat(filepath.Join(work, "workshop", "history", "issues", "000099-sibling-hint.md")); !os.IsNotExist(err) {
		t.Fatal("sibling archived hint retained")
	}
	issueBytes, err := os.ReadFile(filepath.Join(work, "workshop", "issues", "000001-bench-replay.md"))
	if err != nil || !strings.Contains(string(issueBytes), task.Spec) || strings.Contains(string(issueBytes), "SOLUTION SECRET") {
		t.Fatalf("issue projection: %v\n%s", err, issueBytes)
	}
	if _, present, err := tracker.ReadCutoverMarker(work); err != nil || !present {
		t.Fatal("real migration missing", err)
	}
	card := prepareTestGit(t, origin, "show", "issue-tracker:"+tracker.CardPath("000001", "bench-replay"))
	if !strings.Contains(card, "status: open") || strings.Contains(card, "claimant:") {
		t.Fatal("target not unclaimed/open", card)
	}
	if data, err := os.ReadFile(filepath.Join(work, "src", "main.txt")); err != nil || string(data) != "broken\n" {
		t.Fatal("parent code changed", err, string(data))
	}
	if info, err := os.Stat(filepath.Join(work, "script.sh")); err != nil || info.Mode()&0111 == 0 {
		t.Fatal("executable mode lost")
	}
	if target, err := os.Readlink(filepath.Join(work, "safe-link")); err != nil || target != "src/main.txt" {
		t.Fatal("safe symlink lost", err)
	}
	for _, root := range []string{work, filepath.Join(s.root, "public", "hub")} {
		raw, err := os.ReadFile(filepath.Join(root, benchPublicContextPath))
		if err != nil {
			t.Fatal(err)
		}
		for _, secret := range []string{task.SourceRepo, task.ParentOID, task.LandingOID, task.Acceptance.Patch, "observer"} {
			if strings.Contains(string(raw), secret) {
				t.Fatal("private input leaked to context", secret)
			}
		}
		if !strings.Contains(string(raw), record.ID) {
			t.Fatal("context lacks run identity")
		}
	}
	if prepareTestGit(t, task.SourceRepo, "rev-parse", "HEAD") != head || prepareTestGit(t, task.SourceRepo, "status", "--porcelain") != status || prepareTestGit(t, task.SourceRepo, "count-objects", "-v") != objects {
		t.Fatal("preparation changed source")
	}
}

func TestBenchPrepareRefusesOverlayAncestryCollision(t *testing.T) {
	task := preparedFixtureTask(t)
	if err := os.Symlink("src", filepath.Join(task.SourceRepo, ".sdlc")); err != nil {
		t.Fatal(err)
	}
	prepareTestGit(t, task.SourceRepo, "add", ".sdlc")
	prepareTestGit(t, task.SourceRepo, "commit", "-qm", "overlay collision")
	task.ParentOID = prepareTestGit(t, task.SourceRepo, "rev-parse", "HEAD")
	task.Acceptance.ParentOID = task.ParentOID
	if _, err := prepareFixture(t, task); err == nil {
		t.Fatal("bench context written through parent symlink")
	}
}

func TestBenchPrepareTreeRejectsPortablePathCollisions(t *testing.T) {
	for _, files := range [][]gitx.TreeFile{
		{{Path: "Upper/a", Mode: "100644"}, {Path: "upper/b", Mode: "100644"}},
		{{Path: "workshop", Mode: "120000", Content: []byte("src")}},
		{{Path: ".sdlc", Mode: "120000", Content: []byte("src")}},
		{{Path: ".GIT/config", Mode: "100644"}},
		{{Path: "module", Mode: "160000"}},
		{{Path: "escape", Mode: "120000", Content: []byte("../../outside")}},
		{{Path: "a/anchor", Mode: "120000", Content: []byte("..")}, {Path: "escape", Mode: "120000", Content: []byte("a/anchor/../outside")}},
	} {
		if err := validateBenchTree(files); err == nil {
			t.Fatalf("unsafe tree accepted: %+v", files)
		}
	}
}

func TestBenchPrepareRefusesAttributeTransformedParentBytes(t *testing.T) {
	task := preparedFixtureTask(t)
	prepareTestWrite(t, task.SourceRepo, "src/raw.txt", "raw\r\n", 0644)
	prepareTestWrite(t, task.SourceRepo, ".gitattributes", "*.txt text eol=lf\n", 0644)
	prepareTestGit(t, task.SourceRepo, "add", ".gitattributes")
	blob := prepareTestGit(t, task.SourceRepo, "hash-object", "-w", "--no-filters", "--", "src/raw.txt")
	prepareTestGit(t, task.SourceRepo, "update-index", "--add", "--cacheinfo", "100644,"+blob+",src/raw.txt")
	prepareTestGit(t, task.SourceRepo, "commit", "-qm", "exact committed CRLF bytes")
	task.ParentOID = prepareTestGit(t, task.SourceRepo, "rev-parse", "HEAD")
	task.Acceptance.ParentOID = task.ParentOID
	if _, err := prepareFixture(t, task); err == nil {
		t.Fatal("Git attributes silently changed the committed parent bytes")
	}
}

func TestBenchPrepareRefusesMismatchedIssuePatchAndEscapingSymlinks(t *testing.T) {
	for _, kind := range []string{"issue", "patch-digest", "patch-path", "symlink", "missing-landing", "tree-parent"} {
		t.Run(kind, func(t *testing.T) {
			task := preparedFixtureTask(t)
			switch kind {
			case "issue":
				task.Spec = "Different contract"
			case "patch-digest":
				task.Acceptance.PatchSHA256 = strings.Repeat("0", 64)
			case "patch-path":
				task.Acceptance.Paths = []string{"tests/other_test.go"}
			case "missing-landing":
				task.LandingOID = strings.Repeat("e", 40)
				task.Acceptance.LandingOID = task.LandingOID
			case "tree-parent":
				task.ParentOID = prepareTestGit(t, task.SourceRepo, "rev-parse", task.ParentOID+"^{tree}")
				task.Acceptance.ParentOID = task.ParentOID
			case "symlink":
				if err := os.Symlink("../outside", filepath.Join(task.SourceRepo, "escape")); err != nil {
					t.Fatal(err)
				}
				prepareTestGit(t, task.SourceRepo, "add", "escape")
				prepareTestGit(t, task.SourceRepo, "commit", "-qm", "hostile parent")
				task.ParentOID = prepareTestGit(t, task.SourceRepo, "rev-parse", "HEAD")
				task.Acceptance.ParentOID = task.ParentOID
			}
			if _, err := prepareFixture(t, task); err == nil {
				t.Fatal("unsafe preparation accepted")
			}
		})
	}
}
