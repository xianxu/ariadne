// push.go — `sdlc push` subcommand. Ports the `push:` Make target.
//
// The direct-on-main ship workflow. Run from main, refuses anything else.
// Sequence (Makefile.workflow ~lines 281-348):
//
//  1. branch == main check
//  2. untracked-files refusal
//  3. auto-commit tracked changes (commit subject synthesized from
//     touched workshop/issues/*.md titles, fallback "auto-commit
//     before push")
//  4. pre-merge judges (plan + specs + lessons by default — same
//     categories the shell `make pre-merge` runs via parallel-checks.sh).
//     Skippable with --no-judge.
//  5. not-done issue warn: scan touched issue files vs origin/main, warn
//     if any are still in working/open/blocked. Skippable with --yes.
//  6. git push
//  7. archive done/wontfix/punt issue files into history/. For status=done
//     with a github_issue: frontmatter, close the GitHub issue first.
//     Commit + push if any moved.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// pushFlags holds the parsed flag values for the push subcommand.
type pushFlags struct {
	Yes        bool
	NoJudge    bool
	DryRun     bool
	IssuesDir  string
	HistoryDir string
}

// pushRunner is the package-level runner for push (test seam). Type lives
// in runner.go.
var pushRunner gitRunner = execGitRunner{}

// NewPushCmd returns the cobra command for `sdlc push`.
func NewPushCmd() *cobra.Command {
	f := pushFlags{}
	cmd := &cobra.Command{
		Use:           "push",
		Short:         "Ship from main: auto-commit, run pre-merge judges, push, archive done issues",
		Long:          "Placeholder — replaced by helptext.MustGet(\"push\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPush(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	}
	cmd.Flags().BoolVar(&f.Yes, "yes", false, "skip the not-done-issue warn prompt")
	cmd.Flags().BoolVar(&f.NoJudge, "no-judge", false, "skip pre-merge judges (emergency-only)")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print would-be operations; do not commit/push/archive")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	cmd.Flags().StringVar(&f.HistoryDir, "history-dir", envOr("WF_HISTORY_DIR", "workshop/history"), "directory for archived issues")
	return cmd
}

// runPush dispatches the push workflow. Hard guard failures call die()
// directly (red prefix + os.Exit). Soft errors return through cobra.
func runPush(stdout, stderr io.Writer, f *pushFlags) error {
	// ── 1. Branch == main ───────────────────────────────────────────────────
	branch := gitx.Capture("branch", "--show-current")
	if branch != "main" {
		die(stderr, fmt.Sprintf("sdlc push must be run from main (current branch: %s)", valueOr(branch, "(detached)")))
	}

	if recovered, err := recoverInterruptedArchive(stdout, stderr, f); err != nil {
		die(stderr, err.Error())
	} else if recovered {
		cok(stderr, "Done.")
		return nil
	}

	// ── 2. No untracked files ───────────────────────────────────────────────
	untrackedOut, err := pushRunner.Git("ls-files", "--others", "--exclude-standard")
	if err != nil {
		die(stderr, fmt.Sprintf("git ls-files: %v\n%s", err, untrackedOut))
	}
	untracked := splitNonEmptyLines(string(untrackedOut))
	if len(untracked) > 0 {
		fmt.Fprintf(stderr, "  %s[x]%s Untracked files found — add or .gitignore them first\n", ansiRed, ansiReset)
		for _, u := range untracked {
			fmt.Fprintf(stderr, "       %s\n", u)
		}
		os.Exit(1)
	}

	// ── 3. Auto-commit tracked changes ──────────────────────────────────────
	dirty := gitx.Capture("status", "--porcelain")
	if dirty != "" {
		msg := buildPushCommitMessage(f.IssuesDir, pushRunner)
		cinfo(stderr, "Auto-committing tracked changes...")
		if f.DryRun {
			fmt.Fprintf(stdout, "Would: git commit -a -m %q\n", msg)
		} else {
			if out, gerr := pushRunner.Git("commit", "-a", "-m", msg); gerr != nil {
				die(stderr, fmt.Sprintf("git commit failed: %v\n%s", gerr, out))
			}
		}
	}

	// ── 4. Pre-merge judges ─────────────────────────────────────────────────
	if !f.NoJudge {
		preOpts := preflightOptions{
			Categories: []judge.Category{judge.Plan, judge.Specs, judge.Lessons},
			IssuesDir:  f.IssuesDir,
			HistoryDir: f.HistoryDir,
			DryRun:     f.DryRun,
			Stdout:     stdout,
			Stderr:     stderr,
		}
		if err := runPreflightJudges(preOpts); err != nil {
			die(stderr, fmt.Sprintf("pre-merge judges failed: %v", err))
		}
	} else {
		cwarn(stderr, "--no-judge: skipping pre-merge judges")
	}

	// ── 5. Not-done issue warn ──────────────────────────────────────────────
	notDone, err := touchedIssuesNotDone("origin/main", f.IssuesDir, pushRunner)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("not-done scan skipped: %v", err))
	}
	if len(notDone) > 0 && !f.Yes && !f.DryRun {
		fmt.Fprintf(stderr, "  %s[!]%s Touched issue files that are NOT done:\n", ansiYellow, ansiReset)
		for _, p := range notDone {
			fmt.Fprintf(stderr, "       %s\n", p)
		}
		fmt.Fprintf(stderr, "Continue anyway? [y/N] ")
		var answer string
		_, _ = fmt.Fscanln(os.Stdin, &answer)
		if answer != "y" && answer != "Y" {
			die(stderr, "aborted by operator")
		}
	} else if len(notDone) > 0 && f.Yes {
		cwarn(stderr, fmt.Sprintf("--yes: continuing past %d not-done issue(s)", len(notDone)))
	}

	// ── 6. git push ─────────────────────────────────────────────────────────
	if f.DryRun {
		cinfo(stderr, "dry-run — skipping git push + archive")
		return nil
	}
	cinfo(stderr, "Pushing to origin/main...")
	if out, gerr := pushRunner.Git("push"); gerr != nil {
		die(stderr, fmt.Sprintf("git push failed: %v\n%s", gerr, out))
	}

	// ── 7. Archive done/wontfix/punt issues ─────────────────────────────────
	repo, repoErr := detectRepo()
	if repoErr != nil {
		// Archive can still proceed; we just can't close GitHub issues.
		cwarn(stderr, fmt.Sprintf("repo detection failed: %v (skipping GitHub issue closes)", repoErr))
		repo = ""
	}
	moves, err := archiveDoneIssues(stderr, repo, f.IssuesDir, f.HistoryDir)
	if err != nil {
		die(stderr, err.Error())
	}
	if len(moves) > 0 {
		cinfo(stderr, "Committing archived history...")
		if out, gerr := pushRunner.Git(archiveAddArgs(moves)...); gerr != nil {
			die(stderr, fmt.Sprintf("git add archived paths: %v\n%s", gerr, out))
		}
		if out, gerr := pushRunner.Git("commit", "-m", "archive completed issues to history"); gerr != nil {
			die(stderr, fmt.Sprintf("commit archive failed: %v\n%s", gerr, out))
		}
		if out, gerr := pushRunner.Git("push"); gerr != nil {
			die(stderr, fmt.Sprintf("push archive failed: %v\n%s", gerr, out))
		}
		cok(stderr, fmt.Sprintf("archived %d issue file(s) to %s/", len(moves), f.HistoryDir))
	}

	cok(stderr, "Done.")
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

type preparedArchiveMove struct {
	IssuePath   string
	HistoryPath string
}

// archiveAddArgs builds the precise `git add` argument list that stages exactly
// the paths an archive touched — each moved issue's deleted source and created
// history file — and nothing else. It is the exactly-moved-paths counterpart to
// the broad `git add <issuesDir>/ <historyDir>/`, which also sweeps unrelated
// untracked tracker files (in-progress WIP for unclaimed issues) onto main (#80).
// The leading `--` guards against any path being parsed as a flag. Pure: callers
// (merge in the main worktree, push in cwd) feed the result to their own runner.
func archiveAddArgs(moves []preparedArchiveMove) []string {
	args := []string{"add", "--"}
	for _, m := range moves {
		args = append(args, m.IssuePath, m.HistoryPath)
	}
	return args
}

// recoverInterruptedArchive handles the state left by an interrupted archive
// step: issue files have already moved to history/, but the archive commit did
// not land. That state contains untracked history files, so it must be handled
// before the general untracked-file guard.
func recoverInterruptedArchive(stdout, stderr io.Writer, f *pushFlags) (bool, error) {
	statusOut, err := pushRunner.Git("status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return false, fmt.Errorf("git status: %v\n%s", err, statusOut)
	}
	moves, other, err := preparedArchiveMoves(string(statusOut), f.IssuesDir, f.HistoryDir)
	if err != nil {
		return false, err
	}
	if len(moves) == 0 {
		return false, nil
	}
	if len(other) > 0 {
		return false, fmt.Errorf("interrupted archive recovery found unrelated worktree changes:\n  %s\n"+
			"Commit/stash those unrelated changes, then re-run `sdlc push --yes` so it can finish the prepared archive move.",
			strings.Join(other, "\n  "))
	}
	cwarn(stderr, fmt.Sprintf("resuming interrupted archive: %d prepared move(s)", len(moves)))
	for _, m := range moves {
		fmt.Fprintf(stderr, "       %s → %s\n", m.IssuePath, m.HistoryPath)
	}
	if f.DryRun {
		fmt.Fprintf(stdout, "Would: git %s\n", strings.Join(archiveAddArgs(moves), " "))
		fmt.Fprintf(stdout, "Would: git commit -m %q\n", "archive completed issues to history")
		fmt.Fprintln(stdout, "Would: git push")
		return true, nil
	}
	if out, gerr := pushRunner.Git(archiveAddArgs(moves)...); gerr != nil {
		return false, fmt.Errorf("git add archived paths: %v\n%s", gerr, out)
	}
	if out, gerr := pushRunner.Git("commit", "-m", "archive completed issues to history"); gerr != nil {
		return false, fmt.Errorf("commit archive failed: %v\n%s", gerr, out)
	}
	if out, gerr := pushRunner.Git("push"); gerr != nil {
		return false, fmt.Errorf("push archive failed: %v\n%s", gerr, out)
	}
	cok(stderr, fmt.Sprintf("archived %d issue file(s) to %s/", len(moves), f.HistoryDir))
	return true, nil
}

func preparedArchiveMoves(statusText, issuesDir, historyDir string) ([]preparedArchiveMove, []string, error) {
	type half struct {
		issueDeleted bool
		historyAdded bool
		issuePath    string
		historyPath  string
	}
	byBase := map[string]*half{}
	var other []string
	for _, line := range strings.Split(statusText, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		status, path, dest := parsePorcelainStatus(line)
		if dest != "" {
			if isIssuePath(path, issuesDir) && isHistoryPath(dest, historyDir) && filepath.Base(path) == filepath.Base(dest) {
				if ok, err := historyFileIsTerminal(dest); err != nil {
					return nil, nil, err
				} else if !ok {
					other = append(other, line)
					continue
				}
				h := byBase[filepath.Base(path)]
				if h == nil {
					h = &half{}
					byBase[filepath.Base(path)] = h
				}
				h.issueDeleted = true
				h.historyAdded = true
				h.issuePath = path
				h.historyPath = dest
				continue
			}
			other = append(other, line)
			continue
		}
		switch {
		case isIssuePath(path, issuesDir) && strings.Contains(status, "D"):
			h := byBase[filepath.Base(path)]
			if h == nil {
				h = &half{}
				byBase[filepath.Base(path)] = h
			}
			h.issueDeleted = true
			h.issuePath = path
		case isHistoryPath(path, historyDir) && (strings.Contains(status, "A") || status == "??"):
			if ok, err := historyFileIsTerminal(path); err != nil {
				return nil, nil, err
			} else if !ok {
				other = append(other, line)
				continue
			}
			h := byBase[filepath.Base(path)]
			if h == nil {
				h = &half{}
				byBase[filepath.Base(path)] = h
			}
			h.historyAdded = true
			h.historyPath = path
		default:
			other = append(other, line)
		}
	}
	var moves []preparedArchiveMove
	for _, h := range byBase {
		if h.issueDeleted && h.historyAdded {
			moves = append(moves, preparedArchiveMove{IssuePath: h.issuePath, HistoryPath: h.historyPath})
			continue
		}
		other = append(other, valueOr(h.issuePath, h.historyPath))
	}
	sort.Slice(moves, func(i, j int) bool { return moves[i].IssuePath < moves[j].IssuePath })
	sort.Strings(other)
	return moves, other, nil
}

func parsePorcelainStatus(line string) (status, path, dest string) {
	if len(line) < 4 {
		return strings.TrimSpace(line), "", ""
	}
	status = strings.TrimSpace(line[:2])
	path = strings.TrimSpace(line[3:])
	if strings.Contains(path, " -> ") {
		parts := strings.SplitN(path, " -> ", 2)
		path, dest = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return status, path, dest
}

func isIssuePath(path, issuesDir string) bool {
	return filepath.Dir(path) == filepath.Clean(issuesDir) && issueFilename(filepath.Base(path))
}

func isHistoryPath(path, historyDir string) bool {
	return filepath.Dir(path) == filepath.Clean(historyDir) && issueFilename(filepath.Base(path))
}

func issueFilename(name string) bool {
	matched, _ := filepath.Match("[0-9][0-9][0-9][0-9][0-9][0-9]-*.md", name)
	return matched
}

func historyFileIsTerminal(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read archive candidate %s: %v", path, err)
	}
	fm, _, perr := issue.Parse(string(data))
	if perr != nil {
		return false, nil
	}
	st, _ := issue.GetField(fm, "status")
	return vocab.Issue().IsTerminal(st), nil
}

// buildPushCommitMessage synthesizes a commit message by extracting the
// `# Title` of every workshop/issues/NNNNNN-*.md that has unstaged or
// staged changes. Falls back to "auto-commit before push" if none found
// (matches the shell target's else branch).
//
// Multiple touched issues → newline-joined titles. Single → just the title.
func buildPushCommitMessage(issuesDir string, r gitRunner) string {
	matches, _ := filepath.Glob(filepath.Join(issuesDir, "[0-9][0-9][0-9][0-9][0-9][0-9]-*.md"))
	sort.Strings(matches)
	var titles []string
	for _, f := range matches {
		// Has any change relative to HEAD?
		out1, err1 := r.Git("diff", "--quiet", "--", f)
		out2, err2 := r.Git("diff", "--cached", "--quiet", "--", f)
		_ = out1
		_ = out2
		if err1 == nil && err2 == nil {
			continue // both quiet → unchanged
		}
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		t := extractFirstTitle(string(data))
		if t != "" {
			titles = append(titles, t)
		}
	}
	if len(titles) == 0 {
		return "auto-commit before push"
	}
	return strings.Join(titles, "\n")
}

// extractFirstTitle returns the first `# Title` line in body (with leading
// "# " stripped), or "" if none. Matches the shell's `grep -m1 '^# '`.
func extractFirstTitle(body string) string {
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

// touchedIssuesNotDone diffs `origin/main..HEAD` for issue files and
// returns the ones whose status is NOT in {done, wontfix, punt}. Used
// by push's not-done warn step. Mirrors check_undone_issues in
// Makefile.workflow.
func touchedIssuesNotDone(baseRef, issuesDir string, r gitRunner) ([]string, error) {
	out, err := r.Git("diff", "--name-only", baseRef+"..HEAD", "--", issuesDir+"/*.md")
	if err != nil {
		return nil, fmt.Errorf("git diff %s..HEAD: %v\n%s", baseRef, err, out)
	}
	touched := splitNonEmptyLines(string(out))
	var notDone []string
	for _, p := range touched {
		// Read from the working tree — the file is on disk at p relative
		// to repo top. Matches the shell `[ -f "$target" ]` guard.
		data, derr := os.ReadFile(p)
		if derr != nil {
			continue
		}
		fm, _, perr := issue.Parse(string(data))
		if perr != nil {
			continue
		}
		st, _ := issue.GetField(fm, "status")
		if !vocab.Issue().IsTerminal(st) {
			notDone = append(notDone, fmt.Sprintf("%s (status: %s)", p, valueOr(st, "unset")))
		}
	}
	return notDone, nil
}

// archiveDoneIssues scans issuesDir for NNNNNN-*.md with terminal status
// and moves them to historyDir. For status=done with a github_issue:
// frontmatter, calls gh issue close (best-effort — failure warns but does
// not abort). Returns the moves it made (deleted issue path + created history
// path, repo-relative) so the caller can stage exactly those paths (#80).
func archiveDoneIssues(stderr io.Writer, repo, issuesDir, historyDir string) ([]preparedArchiveMove, error) {
	matches, _ := filepath.Glob(filepath.Join(issuesDir, "[0-9][0-9][0-9][0-9][0-9][0-9]-*.md"))
	sort.Strings(matches)
	var moves []preparedArchiveMove
	for _, p := range matches {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		fm, _, perr := issue.Parse(string(data))
		if perr != nil {
			continue
		}
		st, _ := issue.GetField(fm, "status")
		if !vocab.Issue().IsTerminal(st) {
			continue
		}
		// status=done + github_issue: → close GitHub issue first. (#122 carve-out:
		// literal "done" is value-specific — only done has a GitHub issue to close —
		// not a category test, so it stays a literal, not vocab.Issue().IsTerminal.)
		if st == "done" && repo != "" {
			if ghNum, ok := issue.GetField(fm, "github_issue"); ok && ghNum != "" {
				cinfo(stderr, fmt.Sprintf("Closing GitHub issue #%s...", ghNum))
				if cerr := ghClient.IssueClose(repo, ghNum, "Fixed on main."); cerr != nil {
					cwarn(stderr, fmt.Sprintf("gh issue close %s failed: %v (continuing)", ghNum, cerr))
				}
			}
		}
		if err := os.MkdirAll(historyDir, 0o755); err != nil {
			return moves, fmt.Errorf("mkdir %s: %v", historyDir, err)
		}
		dest := filepath.Join(historyDir, filepath.Base(p))
		cinfo(stderr, fmt.Sprintf("Archiving %s to %s/", p, historyDir))
		if err := os.Rename(p, dest); err != nil {
			return moves, fmt.Errorf("mv %s → %s: %v", p, dest, err)
		}
		moves = append(moves, preparedArchiveMove{IssuePath: p, HistoryPath: dest})
	}
	return moves, nil
}

// splitNonEmptyLines splits text on newlines and drops empties. Used to
// turn `git diff --name-only` and `git ls-files` output into clean slices.
func splitNonEmptyLines(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	var out []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}
