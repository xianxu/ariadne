// branchcreate.go — branch-creation helpers shared by `sdlc change-code`
// (and historically `sdlc start`, now removed). Factored out so the
// worktree-or-in-place choice has one source of truth.
//
// Also houses the issue-name resolution previously living in start.go,
// since change-code is now the sole consumer.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// nameFlags is the subset of changeCodeFlags that name-resolution
// needs. Kept as a separate struct so the function signature stays
// independent of the calling verb's wider flag set.
type nameFlags struct {
	Issue     int
	Name      string
	IssuesDir string
}

// resolveBranchName resolves the branch + worktree name from --name
// or --issue. Name-resolution priority:
//
//  1. --name explicit          → use as-is, no untracked detection
//  2. --issue N                → look up workshop/issues/NNNNNN-*.md,
//     derive name from basename. Returns it
//     as untrackedFile *only if* git
//     reports it as untracked.
//  3. neither                  → scan untracked files in issues-dir;
//     if exactly one NNNNNN-*.md, use that.
//     Multiple / zero → error.
//
// Returns (name, untrackedFile, err). untrackedFile is the path that
// should be committed before branch creation; empty if no commit is
// needed (--name was given, or the --issue file is already tracked).
func resolveBranchName(f *nameFlags, r gitRunner) (name, untrackedFile string, err error) {
	if f.Name != "" && f.Issue > 0 {
		return "", "", fmt.Errorf("--name and --issue are mutually exclusive")
	}

	if f.Name != "" {
		return f.Name, "", nil
	}

	untracked, err := listUntrackedIssues(f.IssuesDir, r)
	if err != nil {
		return "", "", err
	}

	if f.Issue > 0 {
		id := fmt.Sprintf("%06d", f.Issue)
		matches, _ := filepath.Glob(filepath.Join(f.IssuesDir, id+"-*.md"))
		if len(matches) == 0 {
			return "", "", fmt.Errorf("no issue file matches %s/%s-*.md", f.IssuesDir, id)
		}
		if len(matches) > 1 {
			return "", "", fmt.Errorf("multiple issue files match %s/%s-*.md: %v", f.IssuesDir, id, matches)
		}
		if info, err := os.Stat(matches[0]); err != nil || !info.Mode().IsRegular() {
			return "", "", fmt.Errorf("issue file %s exists in glob but is not a readable regular file", matches[0])
		}
		base := strings.TrimSuffix(filepath.Base(matches[0]), ".md")
		for _, u := range untracked {
			if filepath.Base(u) == filepath.Base(matches[0]) {
				return base, matches[0], nil
			}
		}
		return base, "", nil
	}

	switch len(untracked) {
	case 0:
		return "", "", fmt.Errorf("no untracked issue file found in %s; pass --name or --issue", f.IssuesDir)
	case 1:
		base := strings.TrimSuffix(filepath.Base(untracked[0]), ".md")
		return base, untracked[0], nil
	default:
		return "", "", fmt.Errorf("multiple untracked issue files found:\n  %s\npass --name or --issue to disambiguate",
			strings.Join(untracked, "\n  "))
	}
}

// listUntrackedIssues returns paths to NNNNNN-<slug>.md files reported
// as untracked by `git ls-files --others --exclude-standard`. Filters
// to the issuesDir prefix + 6-digit prefix shape. Empty slice + nil
// error if none.
func listUntrackedIssues(issuesDir string, r gitRunner) ([]string, error) {
	out, err := r.Git("ls-files", "--others", "--exclude-standard", "--", issuesDir+"/")
	if err != nil {
		return nil, fmt.Errorf("git ls-files: %v\n%s", err, out)
	}
	text := strings.TrimSpace(string(out))
	if text == "" {
		return nil, nil
	}
	var matches []string
	for _, line := range strings.Split(text, "\n") {
		base := filepath.Base(line)
		if issueIDRE.MatchString(base) {
			matches = append(matches, line)
		}
	}
	return matches, nil
}

// issueIDRE matches NNNNNN-<slug>.md filenames (6-digit prefix, dash,
// any slug, .md).
var issueIDRE = regexp.MustCompile(`^\d{6}-.*\.md$`)

// commitUntrackedIssueFile commits + pushes one untracked file before
// branch creation, so the new branch starts from a tracked state.
// Push failures are warnings, not fatal — same posture as start.go's
// pre-#39 behavior and the legacy Makefile target.
func commitUntrackedIssueFile(stderr io.Writer, untrackedFile string, r gitRunner) error {
	if untrackedFile == "" {
		return nil
	}
	cinfo(stderr, fmt.Sprintf("Committing %s before branch creation...", untrackedFile))
	if out, err := r.Git("add", untrackedFile); err != nil {
		return fmt.Errorf("git add %s: %v\n%s", untrackedFile, err, out)
	}
	if out, err := r.Git("commit", "-m", "committing issue file before branch creation"); err != nil {
		return fmt.Errorf("git commit: %v\n%s", err, out)
	}
	if out, err := r.Git("push"); err != nil {
		cwarn(stderr, fmt.Sprintf("push failed, continuing with branch creation: %v\n%s", err, out))
	}
	return nil
}

// createWorktreeBranch creates a fresh git worktree on a new branch
// under ../worktree/<repo-dir-name>/<name>/, and writes the worktree
// path to <repo-root>/.goto so the `g` shell alias can cd there.
//
// Returns the worktree path on success.
func createWorktreeBranch(stdout, stderr io.Writer, name string, r gitRunner) (string, error) {
	repoTop, err := gitx.RepoTopLevel()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %v", err)
	}
	repoDir := filepath.Base(repoTop)
	wtRoot := filepath.Join(filepath.Dir(repoTop), "worktree", repoDir)
	wtPath := filepath.Join(wtRoot, name)

	if err := r.MkdirAll(wtRoot); err != nil {
		return "", fmt.Errorf("mkdir %s: %v", wtRoot, err)
	}
	if out, err := r.Git("worktree", "add", "-b", name, wtPath, "HEAD"); err != nil {
		return "", fmt.Errorf("git worktree add: %v\n%s", err, out)
	}
	cok(stderr, fmt.Sprintf("Worktree created at %s on branch %s", wtPath, name))

	gotoPath := filepath.Join(repoTop, ".goto")
	if err := r.WriteFile(gotoPath, []byte(wtPath)); err != nil {
		cwarn(stderr, fmt.Sprintf(".goto write failed: %v", err))
	} else {
		cok(stderr, "Run: g (to cd into worktree)")
	}
	fmt.Fprintln(stdout, wtPath)
	return wtPath, nil
}

// createInPlaceBranch creates a new branch on the current worktree.
// The working tree (including any uncommitted plan edits) carries
// forward to the new branch — that's the whole point of "in-place":
// the operator stays put and starts coding.
//
// Returns the branch name on success (same as input, for symmetry
// with createWorktreeBranch's return-the-location pattern).
func createInPlaceBranch(stdout, stderr io.Writer, name string, r gitRunner) (string, error) {
	if out, err := r.Git("checkout", "-b", name); err != nil {
		return "", fmt.Errorf("git checkout -b %s: %v\n%s", name, err, out)
	}
	cok(stderr, fmt.Sprintf("Branch %s created in place (working tree carried forward)", name))
	fmt.Fprintln(stdout, name)
	return name, nil
}
