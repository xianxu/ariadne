package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommitSelection(t *testing.T) {
	binary := buildFleetE2EBinary(t)
	repo, origin := syncRepo(t)
	git(t, repo, "switch", "-c", "selected-docs")
	for _, p := range []string{issuePath206, "workshop/plans/000206-plan.md", "workshop/projects/group.md"} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(repo, p)), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(repo, p), []byte("chosen document\n"), 0644); err != nil {
			t.Fatal(err)
		}
		git(t, repo, "add", "--", p)
	}
	git(t, repo, "commit", "-m", "chosen package")
	source := strings.TrimSpace(git(t, repo, "rev-parse", "HEAD"))
	cmd := exec.Command(binary, "issue", "publish", "--commit", source)
	cmd.Dir = repo
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("publish selected documentation: %v\n%s", err, out)
	}
	if got := git(t, origin, "show", "main:workshop/plans/000206-plan.md"); !strings.Contains(got, "chosen document") {
		t.Fatal(got)
	}
	if got := strings.TrimSpace(git(t, repo, "rev-parse", "HEAD")); got != source {
		t.Fatalf("publication moved source HEAD: %s", got)
	}
	for _, tc := range []struct {
		name, path string
		symlink    bool
	}{
		{"code", "implementation.go", false}, {"atlas", "atlas/coupled.md", false}, {"symlink", "workshop/plans/link.md", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(repo, tc.path)
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			var err error
			if tc.symlink {
				err = os.Symlink("000206-plan.md", path)
			} else {
				err = os.WriteFile(path, []byte("ineligible\n"), 0644)
			}
			if err != nil {
				t.Fatal(err)
			}
			git(t, repo, "add", "--", tc.path)
			git(t, repo, "commit", "-m", "mixed source")
			before := git(t, origin, "rev-parse", "main")
			cmd := exec.Command(binary, "issue", "publish", "--commit", "HEAD")
			cmd.Dir = repo
			if out, err := cmd.CombinedOutput(); err == nil {
				t.Fatalf("accepted ineligible %s: %s", tc.path, out)
			}
			if after := git(t, origin, "rev-parse", "main"); after != before {
				t.Fatal("refusal changed origin")
			}
		})
	}
}

func TestCommitPublicationSlots(t *testing.T) {
	binary := buildFleetE2EBinary(t)
	for _, kind := range []string{"primary-0", "linked-1", "linked-2", "private-dependency-clone"} {
		t.Run(kind, func(t *testing.T) {
			repo, origin := syncRepo(t)
			dir := repo
			if strings.HasPrefix(kind, "linked") {
				dir = filepath.Join(t.TempDir(), kind)
				git(t, repo, "worktree", "add", "-b", kind, dir)
			} else if kind == "private-dependency-clone" {
				dir = filepath.Join(t.TempDir(), ".deps", "private")
				git(t, "", "clone", origin, dir)
				git(t, dir, "config", "user.name", "Test")
				git(t, dir, "config", "user.email", "test@example.com")
			}
			if err := os.WriteFile(filepath.Join(dir, "unselected.go"), []byte("package private\n"), 0644); err != nil {
				t.Fatal(err)
			}
			git(t, dir, "add", "--", "unselected.go")
			git(t, dir, "commit", "-m", "unselected code ancestor")
			writeSyncIssue(t, dir, filepath.Base(issuePath206), "selected doc\n")
			git(t, dir, "add", "--", issuePath206)
			git(t, dir, "commit", "-m", "selected doc")
			source := strings.TrimSpace(git(t, dir, "rev-parse", "HEAD"))
			stageForeignFile(t, dir, "staged.go")
			writeSyncIssue(t, dir, filepath.Base(issuePath206), "dirty future document\n")
			before := git(t, dir, "status", "--porcelain=v1")
			index := git(t, dir, "write-tree")
			cmd := exec.Command(binary, "issue", "publish", "--commit", source)
			cmd.Dir = dir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("publish in %s: %v %s", kind, err, out)
			}
			if got := strings.TrimSpace(git(t, dir, "rev-parse", "HEAD")); got != source {
				t.Fatal("publication moved branch")
			}
			if got := git(t, dir, "status", "--porcelain=v1"); got != before {
				t.Fatalf("working state changed: %s vs %s", got, before)
			}
			if got := git(t, dir, "write-tree"); got != index {
				t.Fatal("publication changed index")
			}
			if got := git(t, origin, "ls-tree", "--name-only", "main", "--", "unselected.go", "staged.go"); strings.TrimSpace(got) != "" {
				t.Fatalf("unselected code leaked: %s", got)
			}
			if got := git(t, origin, "show", "main:"+issuePath206); got != "selected doc" {
				t.Fatalf("published working copy instead of commit: %q", got)
			}
		})
	}
}
