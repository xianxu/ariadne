// issue.go — `sdlc issue` command group: CRUD/authoring of the issue
// *record*, complementing the flat checkpoint guards that defend workflow
// *transitions* (ariadne#56).
//
// M1 wires the parent group + `issue new`. `set-status` moves in and
// `list`/`show` arrive in M2; `fetch` folds into `issue new --from-github`.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/helptext"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// NewIssueCmd returns the `sdlc issue` parent command. Long is a
// placeholder; main.go overrides with helptext.MustGet("issue").
func NewIssueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "issue",
		Short:         "Create and manage workshop issues",
		Long:          "Placeholder — replaced by helptext.MustGet(\"issue\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newIssueNewCmd())

	// set-status moved under `issue` (#56 M2). The transition guards live
	// in applyStatus / checkTransitionGuards (returned errors, unit-tested)
	// — only the cobra wiring relocates. main.go keeps a hidden deprecated
	// flat `sdlc set-status` alias for one cycle.
	setStatus := NewSetStatusCmd()
	setStatus.Long = helptext.MustGet("set-status")
	cmd.AddCommand(setStatus)

	cmd.AddCommand(newIssueListCmd())
	cmd.AddCommand(newIssueShowCmd())
	return cmd
}

// issueNewFlags holds the parsed flags for `sdlc issue new`.
type issueNewFlags struct {
	Slug       string
	FromGitHub int
	Deps       []string
	Target     string
	DryRun     bool
	IssuesDir  string
	HistoryDir string
}

func newIssueNewCmd() *cobra.Command {
	f := issueNewFlags{}
	cmd := &cobra.Command{
		Use:   "new <title>",
		Short: "Create a new workshop issue from the canonical template",
		Long: `Create workshop/issues/NNNNNN-<slug>.md from the canonical template
(see ` + "`sdlc issue --help`" + ` for the field/section contract). Allocates the
next 6-digit ID by scanning issues/ + history/ — the deterministic step the
agent must not do by hand under parallel workstreams — and prints the path.

  sdlc issue new "Some title"              # blank issue
  sdlc issue new "x" --target my-target    # with a target: slug
  sdlc issue new --from-github 42          # title/body from a GitHub issue

With --from-github the title is taken from the GitHub issue (a positional
title overrides it) and the issue body is seeded under ## Problem.`,
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIssueNew(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f, args)
		},
	}
	cmd.Flags().StringVar(&f.Slug, "slug", "", "override the auto-derived slug")
	cmd.Flags().IntVar(&f.FromGitHub, "from-github", 0, "derive title + body from this GitHub issue number")
	cmd.Flags().StringSliceVar(&f.Deps, "deps", nil, "dependency refs, e.g. --deps repo#1,repo#2")
	cmd.Flags().StringVar(&f.Target, "target", "", "target: frontmatter slug")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print would-be path + body; do not write")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	cmd.Flags().StringVar(&f.HistoryDir, "history-dir", envOr("WF_HISTORY_DIR", "workshop/history"), "directory holding archived issues")
	return cmd
}

// runIssueNew is the entry point for `sdlc issue new`. Hard guardrail
// failures call die(); the happy path prints the created path to stdout.
func runIssueNew(stdout, stderr io.Writer, f *issueNewFlags, args []string) error {
	title := ""
	if len(args) > 0 {
		title = args[0]
	}

	var ghNum, problemBody string
	if f.FromGitHub > 0 {
		repo, err := detectRepo()
		if err != nil {
			die(stderr, err.Error())
		}
		ghNum = strconv.Itoa(f.FromGitHub)
		ghTitle, ghBody, err := ghClient.TitleAndBody(repo, ghNum)
		if err != nil {
			die(stderr, fmt.Sprintf("fetch GitHub issue %s: %v", ghNum, err))
		}
		if title == "" {
			title = ghTitle
		}
		problemBody = ghBody
	}

	if strings.TrimSpace(title) == "" {
		die(stderr, "a title is required (positional arg, or --from-github N to derive it)")
	}

	slug := f.Slug
	if slug == "" {
		slug = issue.Slugify(title)
	}
	if slug == "" {
		die(stderr, fmt.Sprintf("title %q produced an empty slug; pass --slug", title))
	}

	nextID, err := issue.NextID(f.IssuesDir, f.HistoryDir)
	if err != nil {
		die(stderr, err.Error())
	}

	today := time.Now().Format("2006-01-02")
	dest := filepath.Join(f.IssuesDir, fmt.Sprintf("%s-%s.md", nextID, slug))
	if _, err := os.Stat(dest); err == nil {
		die(stderr, fmt.Sprintf("issue file already exists: %s", dest))
	}

	rendered := issue.Render(issue.ScaffoldSpec{
		ID:          nextID,
		Title:       title,
		Today:       today,
		GithubIssue: ghNum,
		ProblemBody: problemBody,
		Deps:        f.Deps,
		Target:      f.Target,
	})

	if f.DryRun {
		cinfo(stderr, "dry-run — no files written")
		fmt.Fprintf(stdout, "Would create: %s\n", dest)
		fmt.Fprintln(stdout, "─── body ───")
		fmt.Fprint(stdout, rendered)
		return nil
	}

	if err := os.MkdirAll(f.IssuesDir, 0o755); err != nil {
		die(stderr, fmt.Sprintf("mkdir %s: %v", f.IssuesDir, err))
	}
	if err := os.WriteFile(dest, []byte(rendered), 0o644); err != nil {
		die(stderr, fmt.Sprintf("write %s: %v", dest, err))
	}

	created := fmt.Sprintf("Created %s", dest)
	if ghNum != "" {
		created += fmt.Sprintf(" (GitHub #%s)", ghNum)
	}
	cok(stderr, created)
	fmt.Fprintln(stdout, dest)
	return nil
}

// ── issue list ───────────────────────────────────────────────────────────────

type issueListFlags struct {
	Status    string
	IssuesDir string
}

func newIssueListCmd() *cobra.Command {
	f := issueListFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workshop issues (ID, status, title)",
		Long: `List issues in workshop/issues/ as "ID  STATUS  TITLE", sorted by ID.
Filter with --status. Broader than 'sdlc state', which surfaces only the
working set + drift.`,
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runIssueList(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	}
	cmd.Flags().StringVar(&f.Status, "status", "", "filter to this status (open|working|blocked|done|wontfix|punt)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	return cmd
}

// runIssueList reuses state.go's listIssues (which reads + sorts by ID)
// rather than re-deriving the scan/sort.
func runIssueList(stdout, stderr io.Writer, f *issueListFlags) error {
	if f.Status != "" && !isValidStatus(f.Status) {
		die(stderr, fmt.Sprintf("invalid status %q (valid: %s)", f.Status, strings.Join(validStatuses, ", ")))
	}
	issues, err := listIssues(f.IssuesDir)
	if err != nil {
		die(stderr, fmt.Sprintf("list issues: %v", err))
	}
	n := 0
	for _, is := range issues {
		if f.Status != "" && is.Status != f.Status {
			continue
		}
		// width 10 fits the longest real status + the "unreadable"
		// sentinel listIssues emits for broken files.
		fmt.Fprintf(stdout, "%s  %-10s  %s\n", is.ID, valueOr(is.Status, "?"), is.Title)
		n++
	}
	if n == 0 {
		cinfo(stderr, "no issues match")
	}
	return nil
}

// ── issue show ───────────────────────────────────────────────────────────────

type issueShowFlags struct {
	IssuesDir string
}

func newIssueShowCmd() *cobra.Command {
	f := issueShowFlags{}
	cmd := &cobra.Command{
		Use:   "show <N>",
		Short: "Show an issue's frontmatter + section headers (no bodies)",
		Long: `Print issue <N>'s frontmatter and its body section headers (# / ## lines)
without the section contents — a structured peek for orienting on an issue
without loading the whole file.`,
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIssueShow(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f, args[0])
		},
	}
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	return cmd
}

func runIssueShow(stdout, stderr io.Writer, f *issueShowFlags, arg string) error {
	id, err := strconv.Atoi(arg)
	if err != nil || id <= 0 {
		die(stderr, fmt.Sprintf("invalid issue id %q (want a positive number, e.g. 56)", arg))
	}
	path, err := locateIssueFile(f.IssuesDir, id)
	if err != nil {
		die(stderr, err.Error())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		die(stderr, fmt.Sprintf("read %s: %v", path, err))
	}
	fm, body, err := issue.Parse(string(data))
	if err != nil {
		die(stderr, fmt.Sprintf("parse %s: %v", path, err))
	}
	fmt.Fprintf(stdout, "%s\n---\n%s---\n", filepath.Base(path), ensureTrailingNewline(fm))
	// Title + section headers only (`# ` / `## `). Deeper headers like
	// `### YYYY-MM-DD` Log entries are intentionally omitted — this is a
	// structure peek, not a content dump.
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") {
			fmt.Fprintln(stdout, line)
		}
	}
	return nil
}

// ensureTrailingNewline returns s with exactly one terminating newline.
func ensureTrailingNewline(s string) string {
	return strings.TrimRight(s, "\n") + "\n"
}
