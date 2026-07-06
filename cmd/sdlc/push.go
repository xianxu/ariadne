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
//  4. pre-push PUBLISH GATE (#160) — the deterministic reviewed-HEAD-unchanged
//     invariant (push is "merge without a PR", Q3); NO LLM. Skippable --no-judge.
//  5. not-done issue warn: scan touched issue files vs origin/main, warn
//     if any are still in working/open/blocked (codecomplete is NOT flagged —
//     it's about to be published). Skippable with --yes.
//  6. git push
//     6.5 publish flip (#160): codecomplete → done (before archiving)
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

	"github.com/xianxu/ariadne/pkg/vocab"
)

// pushFlags holds the parsed flag values for the push subcommand.
type pushFlags struct {
	Yes        bool
	NoJudge    bool
	NoValidate bool
	DryRun     bool
	IssuesDir  string
	HistoryDir string
	PlansDir   string
}

// pushRunner is the package-level runner for push (test seam). Type lives
// in runner.go.
var pushRunner gitRunner = execGitRunner{}

// NewPushCmd returns the cobra command for `sdlc push`.
func NewPushCmd() *cobra.Command {
	f := pushFlags{}
	cmd := markMutatingCommand(&cobra.Command{
		Use:           "push",
		Short:         "Ship from main: auto-commit, run the publish gate, push, flip codecomplete→done, archive",
		Long:          "Placeholder — replaced by helptext.MustGet(\"push\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPush(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})
	cmd.Flags().BoolVar(&f.Yes, "yes", false, "skip the not-done-issue warn prompt")
	cmd.Flags().BoolVar(&f.NoJudge, "no-judge", false, "skip the pre-push publish gate — #160 reviewed-HEAD-unchanged invariant (emergency-only)")
	cmd.Flags().BoolVar(&f.NoValidate, "no-validate", false, "skip the #124 instance-conformance gate (escape hatch — announced loudly)")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print would-be operations; do not commit/push/archive")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	cmd.Flags().StringVar(&f.HistoryDir, "history-dir", envOr("WF_HISTORY_DIR", "workshop/history"), "directory for archived issues")
	cmd.Flags().StringVar(&f.PlansDir, "plans-dir", envOr("WF_PLANS_DIR", "workshop/plans"), "directory holding durable plans + review sidecars (archived with the issue, #143)")
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
		exitWithCode(1)
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

	// ── 3.5 Instance-conformance gate (#124) ────────────────────────────────
	// Deterministic, separate from the judges (so --no-judge keeps it, and
	// --no-validate keeps the judges). Same window the judges use.
	if !f.NoValidate {
		if err := validateChangedIssuesFn(gitx.DiffBase(), "", f.IssuesDir, stdout, stderr); err != nil {
			die(stderr, err.Error())
		}
	} else {
		cwarn(stderr, "⚠️  --no-validate: SKIPPING the instance-conformance gate (#124) — issue frontmatter/sections NOT verified before main. Escape hatch: say why in your commit/log.")
	}

	// ── 4. Pre-push publish gate (#160) — deterministic, NO LLM ──────────────
	// Push is "merge without a PR": the same reviewed-HEAD-unchanged invariant.
	// All LLM review is close-time; the publish gate carries no judge.
	if !f.NoJudge {
		if err := runPublishGate(gitx.DiffBase(), f.IssuesDir, stderr); err != nil {
			if f.DryRun {
				cwarn(stderr, fmt.Sprintf("dry-run: publish gate WOULD refuse: %v", err))
			} else {
				die(stderr, err.Error())
			}
		}
	} else {
		cwarn(stderr, "--no-judge: skipping the pre-push publish gate (#160 reviewed-HEAD-unchanged invariant)")
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

	// ── 6.5 Publish flip (#160): codecomplete → done before archiving ────────
	// Direct-to-main publish (Q3): the just-pushed codecomplete issues become done
	// (the deterministic flip) before the archive scan, which keys on IsTerminal.
	// The flip is bundled into the archive commit + push below.
	if flipped, ferr := publishCodecompleteIssues(f.IssuesDir); ferr != nil {
		die(stderr, fmt.Sprintf("publish flip (codecomplete → done): %v", ferr))
	} else if len(flipped) > 0 {
		cinfo(stderr, fmt.Sprintf("Published %d issue(s): codecomplete → done", len(flipped)))
	}

	// ── 7. Archive done/wontfix/punt issues ─────────────────────────────────
	repo, repoErr := detectRepo()
	if repoErr != nil {
		// Archive can still proceed; we just can't close GitHub issues.
		cwarn(stderr, fmt.Sprintf("repo detection failed: %v (skipping GitHub issue closes)", repoErr))
		repo = ""
	}
	moves, err := archiveDoneIssues(stderr, repo, f.IssuesDir, f.HistoryDir, f.PlansDir)
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
	// SourceUntracked marks a move whose source was untracked at archive time
	// (a not-yet-committed review sidecar, #154). After the rename its old path
	// has no worktree file AND no index entry, so `git add <IssuePath>` would die
	// with "pathspec did not match" — stage only HistoryPath for these. Default
	// false = tracked source (issue files, durable plans, recovery moves): stage
	// the source deletion + the history addition, exactly as before.
	SourceUntracked bool
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
		// The source path stages a deletion — only meaningful when the source was
		// tracked. An untracked source (#154) simply vanished at the rename; adding
		// its pre-move path would fail "pathspec did not match". Stage the moved
		// file at its new location either way.
		if !m.SourceUntracked {
			args = append(args, m.IssuePath)
		}
		args = append(args, m.HistoryPath)
	}
	return args
}

// issueIDPrefix returns the leading 6-digit id of an issue/plan filename
// (e.g. "000143" from "000143-x.md"), or "" when the name doesn't match the
// NNNNNN- convention. The single source for "which plan artifacts belong to
// this issue" — the glob key is id+"-*" (#143).
func issueIDPrefix(name string) string {
	base := filepath.Base(name)
	if len(base) < 7 || base[6] != '-' {
		return ""
	}
	for i := 0; i < 6; i++ {
		if base[i] < '0' || base[i] > '9' {
			return ""
		}
	}
	return base[:6]
}

// archivePlanArtifacts moves every workshop/plans/NNNNNN-* artifact (the durable
// plan + every boundary-review sidecar, #136) that shares the archived issue's id
// prefix into history, and returns the moves. plansFull/historyFull are the
// source/dest dirs used for the rename; recPlansDir/recHistoryDir are the dirs
// recorded in the returned preparedArchiveMove for the git-add/commit step (they
// differ from *Full only on the merge path, which renames under mainPath but
// records mainPath-relative paths). An issue with no plan → zero moves, no error
// (the glob simply matches nothing). One mover, both archive callers (ARCH-DRY).
//
// srcUntracked is the injected IO seam (ARCH-PURE): given a move's recorded
// (git-relative) source path, it reports whether that path was untracked at
// archive time — a review sidecar `sdlc close` created but no commit staged
// reaches here untracked (#154). The caller backs it with `git ls-files` in the
// right worktree (cwd for push, mainPath for merge); a nil probe means "assume
// tracked" (the pre-#154 behavior). The probe is consulted before the rename so
// it observes the source at its original path.
func archivePlanArtifacts(issueBase, plansFull, historyFull, recPlansDir, recHistoryDir string, srcUntracked func(recPath string) bool) ([]preparedArchiveMove, error) {
	id := issueIDPrefix(issueBase)
	if id == "" {
		return nil, nil
	}
	matches, _ := filepath.Glob(filepath.Join(plansFull, id+"-*"))
	if len(matches) == 0 {
		return nil, nil
	}
	sort.Strings(matches)
	if err := os.MkdirAll(historyFull, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %v", historyFull, err)
	}
	var moves []preparedArchiveMove
	for _, p := range matches {
		base := filepath.Base(p)
		dest := filepath.Join(historyFull, base)
		recSrc := filepath.Join(recPlansDir, base)
		untracked := srcUntracked != nil && srcUntracked(recSrc)
		if err := os.Rename(p, dest); err != nil {
			return moves, fmt.Errorf("mv %s → %s: %v", p, dest, err)
		}
		moves = append(moves, preparedArchiveMove{
			IssuePath:       recSrc,
			HistoryPath:     filepath.Join(recHistoryDir, base),
			SourceUntracked: untracked,
		})
	}
	return moves, nil
}

// gitSrcUntracked builds the archivePlanArtifacts source-trackedness probe (#154)
// from a git invoker (pushRunner.Git in cwd, or a mergeRunner.GitInDir(mainPath,…)
// closure). It reports a recorded source path as untracked iff `git ls-files`
// cleanly returns no index entry for it (empty output, no error). On any git
// error it returns false — treat the source as tracked and stage its deletion,
// preserving the pre-#154 behavior rather than risk dropping a real deletion.
func gitSrcUntracked(git func(args ...string) ([]byte, error)) func(string) bool {
	return func(recPath string) bool {
		out, err := git("ls-files", "--", recPath)
		return err == nil && strings.TrimSpace(string(out)) == ""
	}
}

// isPlanPath reports whether path is a plan artifact directly under plansDir
// (the plans-dir counterpart to isIssuePath/isHistoryPath; reuses issueFilename).
func isPlanPath(path, plansDir string) bool {
	return filepath.Dir(path) == filepath.Clean(plansDir) && issueFilename(filepath.Base(path))
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
	moves, other, err := preparedArchiveMoves(string(statusOut), f.IssuesDir, f.HistoryDir, f.PlansDir)
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

func preparedArchiveMoves(statusText, issuesDir, historyDir, plansDir string) ([]preparedArchiveMove, []string, error) {
	// A half is one side of a src→history archive move. srcIsPlan marks a plan
	// artifact (workshop/plans/NNNNNN-*, #143), which — unlike an issue — carries
	// no terminal frontmatter, so its id-prefixed plans-dir source is the
	// membership proof instead of the terminal gate.
	type half struct {
		srcDeleted   bool
		srcIsPlan    bool
		historyAdded bool
		srcPath      string
		historyPath  string
	}
	byBase := map[string]*half{}
	get := func(base string) *half {
		if h := byBase[base]; h != nil {
			return h
		}
		h := &half{}
		byBase[base] = h
		return h
	}
	var other []string
	for _, line := range strings.Split(statusText, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		status, path, dest := parsePorcelainStatus(line)
		if dest != "" {
			// A staged rename of an issue OR plan artifact, src → history, same base.
			if isHistoryPath(dest, historyDir) && filepath.Base(path) == filepath.Base(dest) &&
				(isIssuePath(path, issuesDir) || isPlanPath(path, plansDir)) {
				h := get(filepath.Base(path))
				h.srcDeleted, h.historyAdded = true, true
				h.srcIsPlan = isPlanPath(path, plansDir)
				h.srcPath, h.historyPath = path, dest
				continue
			}
			other = append(other, line)
			continue
		}
		switch {
		case isIssuePath(path, issuesDir) && strings.Contains(status, "D"):
			h := get(filepath.Base(path))
			h.srcDeleted, h.srcPath = true, path
		case isPlanPath(path, plansDir) && strings.Contains(status, "D"):
			h := get(filepath.Base(path))
			h.srcDeleted, h.srcIsPlan, h.srcPath = true, true, path
		case isHistoryPath(path, historyDir) && (strings.Contains(status, "A") || status == "??"):
			// Defer the terminal-frontmatter decision to finalization: a history
			// addition's issue-vs-plan nature is only known once its paired deletion
			// is seen. Plan artifacts (no frontmatter) would otherwise be rejected.
			h := get(filepath.Base(path))
			h.historyAdded, h.historyPath = true, path
		default:
			other = append(other, line)
		}
	}
	var moves []preparedArchiveMove
	for _, h := range byBase {
		if h.srcDeleted && h.historyAdded {
			// Issue moves keep the terminal-frontmatter gate; plan moves rely on the
			// id-prefixed plans-dir source as the membership proof instead.
			if !h.srcIsPlan {
				ok, err := historyFileIsTerminal(h.historyPath)
				if err != nil {
					return nil, nil, err
				}
				if !ok {
					// Looks like an archive but the issue isn't terminal — refuse
					// both halves (a half-moved non-done issue is suspicious).
					other = append(other, h.srcPath, h.historyPath)
					continue
				}
			}
			moves = append(moves, preparedArchiveMove{IssuePath: h.srcPath, HistoryPath: h.historyPath})
			continue
		}
		other = append(other, valueOr(h.srcPath, h.historyPath))
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
		// #160: `codecomplete` is the normal pre-publish state — the publish gate is
		// about to flip it to done — so it is NOT "not done" (else every merge/push
		// would trip this warn). Only open/working/blocked are genuinely not-done.
		if !vocab.Issue().IsTerminal(st) && st != "codecomplete" {
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
func archiveDoneIssues(stderr io.Writer, repo, issuesDir, historyDir, plansDir string) ([]preparedArchiveMove, error) {
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
		// Sweep the issue's durable plan + review sidecars to history too (#143).
		// An untracked sidecar (#154) stages only its history dest, not a vanished
		// source path — probe via `git ls-files` in cwd.
		planMoves, perr := archivePlanArtifacts(filepath.Base(p), plansDir, historyDir, plansDir, historyDir, gitSrcUntracked(pushRunner.Git))
		if perr != nil {
			return moves, perr
		}
		moves = append(moves, planMoves...)
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
