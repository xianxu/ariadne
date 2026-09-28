// pr.go — `sdlc pr` subcommand. Ports the `pull-request:` Make target.
//
// Run from a worktree branch (refuses main). Pushes the branch with
// upstream tracking, scans touched issue files since branch point for
// github_issue: frontmatter, formats them as "Fixes #1, #2, #3", and
// opens a PR via `gh pr create`. Mirrors Makefile.workflow ~lines 350-388.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// prFlags holds the parsed flag values for the pr subcommand.
type prFlags struct {
	Context   context.Context
	DryRun    bool
	IssuesDir string
}

// prRunner is the package-level runner for pr (test seam). Type lives
// in runner.go.
var prRunner gitRunner = execGitRunner{}

// NewPRCmd returns the cobra command for `sdlc pr`.
func NewPRCmd() *cobra.Command {
	f := prFlags{}
	cmd := markMutatingCommand(&cobra.Command{
		Use:           "pr",
		Short:         "Open a PR for the current worktree branch (scans touched issues for fixes)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"pr\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f.Context = cmd.Context()
			return runPR(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print would-be PR body + gh command; do not push or create PR")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	return cmd
}

// runPR dispatches the pr workflow.
func runPR(stdout, stderr io.Writer, f *prFlags) error {
	ctx := commandContext(f.Context)
	// #252: a PR must not carry a change to handed-off details (checked again at
	// merge against the then-current main).
	if err := guardTransferredDetailsFn(commandContext(f.Context)); err != nil {
		return err
	}
	target, targetErr := resolveLandingTarget(prRunner)
	if targetErr != nil {
		return targetErr
	}
	if target != nil {
		return runDurablePR(stdout, stderr, f, *target)
	}

	// ── 1. Refuse if on main / detached ─────────────────────────────────────
	branch := gitx.Capture("branch", "--show-current")
	if branch == "" || branch == "main" {
		die(stderr, fmt.Sprintf("sdlc pr must be run from a worktree branch (current: %s)", valueOr(branch, "(detached)")))
	}

	repo, err := detectRepo()
	if err != nil {
		die(stderr, err.Error())
	}

	// ── 2. Compute merge base ───────────────────────────────────────────────
	base := gitx.BranchPoint()
	if base == "" {
		base = "main"
	}

	// ── 3. Collect touched issues + github_issue numbers ────────────────────
	touched, err := touchedIssueFiles(base, f.IssuesDir, prRunner)
	if err != nil {
		die(stderr, fmt.Sprintf("scan touched issues: %v", err))
	}
	ghNums, lerr := collectGitHubIssueNumbers(ctx, touched)
	if lerr != nil {
		cwarn(stderr, fmt.Sprintf("PR body lacks Fixes lines: %v", lerr))
	}

	// ── 4. Build commits + fixes body ───────────────────────────────────────
	commits := gitCommitsSince(base, prRunner)
	fixes := formatFixes(ghNums)
	body := combineBody(commits, fixes)

	// ── 5. Push branch with upstream ────────────────────────────────────────
	if f.DryRun {
		cinfo(stderr, "dry-run — no push or PR creation")
		fmt.Fprintf(stdout, "Would: git push -u origin %s\n", branch)
		fmt.Fprintf(stdout, "Would: gh pr create --repo %s --base main --head %s\n", repo, branch)
		if fixes != "" {
			fmt.Fprintln(stdout, "── fixes line ──")
			fmt.Fprintln(stdout, fixes)
		}
		if body != "" {
			fmt.Fprintln(stdout, "── PR body ──")
			fmt.Fprintln(stdout, body)
		}
		return nil
	}

	cinfo(stderr, fmt.Sprintf("Pushing %s with upstream tracking...", branch))
	if out, gerr := prRunner.Git("push", "-u", "origin", branch); gerr != nil {
		die(stderr, fmt.Sprintf("git push -u origin %s: %v\n%s", branch, gerr, out))
	}

	// ── 6. Open PR ──────────────────────────────────────────────────────────
	if fixes != "" {
		cinfo(stderr, fmt.Sprintf("Including in PR body: %s", fixes))
	}
	cinfo(stderr, fmt.Sprintf("Creating PR (base=main head=%s)...", branch))
	url, err := ghClient.PRCreate(repo, "main", branch, body)
	if err != nil {
		die(stderr, err.Error())
	}
	if url != "" {
		fmt.Fprintln(stdout, url)
	}
	cok(stderr, "PR created.")
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

// touchedIssueFiles returns workshop/issues/*.md paths changed since
// baseRef. Empty slice if none. Used by pr.go to find linkable issues.
func touchedIssueFiles(baseRef, issuesDir string, r gitRunner) ([]string, error) {
	out, err := r.Git("diff", "--name-only", baseRef+"..HEAD", "--", issuesDir+"/*.md")
	if err != nil {
		return nil, fmt.Errorf("git diff: %v\n%s", err, out)
	}
	return splitNonEmptyLines(string(out)), nil
}

// collectGitHubIssueNumbers reads each path's frontmatter and pulls the
// `github_issue:` value if present + non-empty. Returns unique numbers
// in ascending numeric order (matches the shell's `sort -u`).
//
// Missing files are skipped silently — the shell target uses `[ -f ]`.
func collectGitHubIssueNumbers(ctx context.Context, paths []string) ([]string, error) {
	seen := map[string]struct{}{}
	records := map[string]tracker.Records{} // per details directory
	for _, p := range paths {
		id, _, ok := issue.ParseFilename(filepath.Base(p))
		if !ok {
			continue
		}
		dir := filepath.Dir(p)
		rs, loaded := records[dir]
		if !loaded {
			var err error
			// Fresh: the links are published in the PR body; a stale card could drop one.
			if rs, err = loadIssueRecords(ctx, dir, tracker.Fresh); err != nil {
				return nil, fmt.Errorf("read the GitHub links of %s: %w", dir, err)
			}
			records[dir] = rs
		}
		rec, ok := rs.Get(id)
		if !ok || rec.DetailPath == "" {
			continue // the shell target skips missing files: a touched path must exist
		}
		num, ok := rec.Field("github_issue") // the card's link, where a tracker exists
		if !ok || num == "" {
			continue
		}
		seen[num] = struct{}{}
	}
	var out []string
	for k := range seen {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool {
		ai, _ := strconv.Atoi(out[i])
		aj, _ := strconv.Atoi(out[j])
		if ai == aj {
			return out[i] < out[j]
		}
		return ai < aj
	})
	return out, nil
}

// formatFixes returns the "Fixes ..." line for the given github_issue
// references. Returns "" if numbers is empty (matches the shell's empty
// `$fixes` branch which falls through to `gh pr create --fill`).
//
// A bare number ("42") is a same-repo reference → "Fixes #42". A
// fully-qualified value ("owner/repo#3") is a cross-repo reference and must be
// used verbatim — GitHub's closing syntax is "Fixes owner/repo#3" with NO
// leading "#". Prepending "#" unconditionally produced the malformed
// "Fixes #owner/repo#3", which GitHub does not parse as a closing reference.
func formatFixes(numbers []string) string {
	if len(numbers) == 0 {
		return ""
	}
	refs := make([]string, len(numbers))
	for i, n := range numbers {
		refs[i] = fixesRef(n)
	}
	return "Fixes " + strings.Join(refs, ", ")
}

// fixesRef renders one github_issue value as a GitHub closing reference. A
// value already carrying "#" or "/" is a complete reference (cross-repo
// owner/repo#N, or already-hashed) and is returned unchanged; a bare number
// gets the leading "#".
func fixesRef(n string) string {
	if strings.ContainsAny(n, "#/") {
		return n
	}
	return "#" + n
}

// gitCommitsSince returns "- <subject>\n- <subject>" lines for every
// commit in `base..HEAD` (base: the branch point from the published trunk).
// Empty if none.
func gitCommitsSince(base string, r gitRunner) string {
	out, err := r.Git("log", base+"..HEAD", "--pretty=format:- %s")
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(out), "\n")
}

// combineBody assembles the final PR body from the commits list + the
// fixes line. Matches the shell pull-request target's logic:
//
//	if both → "<commits>\n\n<fixes>"
//	if only commits → "<commits>"
//	if only fixes → "<fixes>"
//	if neither → ""
func combineBody(commits, fixes string) string {
	commits = strings.TrimSpace(commits)
	fixes = strings.TrimSpace(fixes)
	switch {
	case commits != "" && fixes != "":
		return commits + "\n\n" + fixes
	case commits != "":
		return commits
	case fixes != "":
		return fixes
	}
	return ""
}

// readFile is a small indirection so tests can stub the file-read path.
// Production reads from disk directly; tests substitute fixtures.
var readFile = os.ReadFile
