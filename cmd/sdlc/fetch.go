// fetch.go — `sdlc fetch --github-issue N`, a hidden deprecated alias for
// `sdlc issue new --from-github N` (#56 M2). runFetch is now a thin shim
// over runIssueNew; the ID allocation + canonical renderer live in
// internal/issue (scaffold.go). detectRepo / originRE stay here as the
// origin-remote → owner/repo helper that runIssueNew calls for --from-github.
package main

import (
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// fetchFlags holds the parsed flag values for the fetch subcommand.
type fetchFlags struct {
	GitHubIssue int
	IssuesDir   string
	HistoryDir  string
	DryRun      bool
}

// NewFetchCmd returns the cobra command for `sdlc fetch`. Long is a
// placeholder; main.go overrides with helptext.MustGet("fetch").
func NewFetchCmd() *cobra.Command {
	f := fetchFlags{}
	cmd := &cobra.Command{
		Use:           "fetch",
		Short:         "Fetch a GitHub issue and create a local workshop/issues/ file",
		Long:          "Placeholder — replaced by helptext.MustGet(\"fetch\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFetch(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	}
	cmd.Flags().IntVar(&f.GitHubIssue, "github-issue", 0, "GitHub issue number to fetch (required)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	cmd.Flags().StringVar(&f.HistoryDir, "history-dir", envOr("WF_HISTORY_DIR", "workshop/history"), "directory holding archived issues")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print would-be path + body; do not write")
	return cmd
}

// ghClient and the ghCaller interface live in ghclient.go (shared with
// pr.go and merge.go). M5 promoted them out of fetch.go once they had
// three+ consumers.

// runFetch is the entry point for the cobra RunE. Since #56 M2 it is a
// thin alias for `sdlc issue new --from-github N`: it shares the canonical
// renderer + ID allocation in runIssueNew, so a fetched issue gets the same
// template as a blank one (GH body seeded under ## Problem). The retained
// `--github-issue` flag keeps old callers working.
func runFetch(stdout, stderr io.Writer, f *fetchFlags) error {
	if f.GitHubIssue <= 0 {
		die(stderr, fmt.Sprintf("--github-issue is required and must be positive (got %d)", f.GitHubIssue))
	}
	return runIssueNew(stdout, stderr, &issueNewFlags{
		FromGitHub: f.GitHubIssue,
		IssuesDir:  f.IssuesDir,
		HistoryDir: f.HistoryDir,
		DryRun:     f.DryRun,
	}, nil)
}

// ── helpers ──────────────────────────────────────────────────────────────────

// originRE captures owner/repo from either git@github.com:owner/repo.git
// or https://github.com/owner/repo[.git]. The Makefile target uses sed
// with two patterns; this single Go regex covers both shapes.
//
// Lazy capture so deeper-path hosts (rare) still parse owner/...rest;
// the shell's sed pipeline does the same via greedy `.*\.git` stripping.
// Review M4 I1.
var originRE = regexp.MustCompile(`github\.com[:/]([^/].*?)(?:\.git)?(?:\n|$)`)

// detectRepo returns the "owner/repo" slug for the current repo's
// `origin` remote. Errors if origin is not configured or doesn't look
// like a github.com URL.
func detectRepo() (string, error) {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return "", fmt.Errorf("git remote get-url origin: %w", err)
	}
	url := strings.TrimSpace(string(out)) + "\n"
	m := originRE.FindStringSubmatch(url)
	if m == nil {
		return "", fmt.Errorf("origin URL %q does not look like a github.com URL", strings.TrimSpace(url))
	}
	return m[1], nil
}

