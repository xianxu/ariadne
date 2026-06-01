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

	cok(stderr, fmt.Sprintf("Created %s", dest))
	fmt.Fprintln(stdout, dest)
	return nil
}
