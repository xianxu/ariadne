package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/helptext"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// headSHA resolves the commit a freeze pins as the task's immutable base. Seam
// for tests.
var headSHA = func() (string, error) {
	s := gitx.Capture("rev-parse", "HEAD")
	if s == "" {
		return "", fmt.Errorf("could not resolve HEAD")
	}
	return s, nil
}

// NewBenchCmd is the `sdlc bench` parent: the multi-agent benchmark harness (#119).
func NewBenchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "bench",
		Short:         "Benchmark coding agents on the same frozen task",
		Long:          "Placeholder — replaced by helptext.MustGet(\"bench\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE:          func(c *cobra.Command, _ []string) error { return c.Help() },
	}
	cmd.AddCommand(newBenchFreezeCmd())
	return cmd
}

type benchFreezeFlags struct {
	Issue int
	Repo  string
	Root  string // repo root override (tests); "" → cwd
}

func newBenchFreezeCmd() *cobra.Command {
	var f benchFreezeFlags
	cmd := &cobra.Command{
		Use:           "freeze",
		Short:         "Snapshot a live issue into an immutable benchmark task",
		Long:          helptext.MustGet("bench-freeze"),
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, _ []string) error {
			return runBenchFreeze(c.OutOrStdout(), c.ErrOrStderr(), f)
		},
	}
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "live issue to freeze (required)")
	cmd.Flags().StringVar(&f.Repo, "repo", "ariadne", "repo name recorded on the task")
	return cmd
}

var benchH2RE = regexp.MustCompile(`(?m)^## (.+?)\s*$`)
var benchSpecHeadingRE = regexp.MustCompile(`(?m)^## Spec[ \t]*$`)

var canonicalIssueSections = map[string]bool{
	"problem": true, "spec": true, "done when": true, "estimate": true,
	"plan": true, "log": true, "side quests": true, "revisions": true,
}

// specHasEmbeddedHeading reports whether the issue's ## Spec section is
// terminated by a NON-canonical ## heading — i.e. the author put a `## ` inside
// the Spec, which issue.SectionBody (and thus freeze) silently truncates. The
// fix is the author's (use ### inside a spec); freeze just warns.
func specHasEmbeddedHeading(body string) bool {
	loc := benchSpecHeadingRE.FindStringIndex(body) // anchored — not "## Specification"
	if loc == nil {
		return false
	}
	m := benchH2RE.FindStringSubmatch(body[loc[1]:])
	if m == nil {
		return false
	}
	return !canonicalIssueSections[strings.ToLower(strings.TrimSpace(m[1]))]
}

func runBenchFreeze(stdout, stderr io.Writer, f benchFreezeFlags) error {
	if f.Issue == 0 {
		return fmt.Errorf("--issue is required")
	}
	root := f.Root
	if root == "" {
		root = "."
	}
	path, slug, err := findIssueFile(root, f.Issue)
	if err != nil {
		return err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, body, err := issue.Parse(string(b))
	if err != nil {
		return err
	}
	spec, ok := issue.SectionBody(body, "Spec")
	if !ok {
		return fmt.Errorf("issue %d has no ## Spec to freeze", f.Issue)
	}
	spec = strings.TrimSpace(spec)
	if specHasEmbeddedHeading(body) {
		fmt.Fprintf(stderr, "warning: issue %d's ## Spec contains a non-canonical `## ` heading; freeze truncates the spec there — use ### subheadings inside specs meant to be frozen\n", f.Issue)
	}
	sha, err := headSHA()
	if err != nil {
		return err
	}
	task := bench.Task{
		ID:          fmt.Sprintf("%d-%s", f.Issue, slug),
		Repo:        f.Repo,
		SourceIssue: fmt.Sprintf("%d", f.Issue),
		BaseSHA:     sha,
		Created:     time.Now().Format("2006-01-02"),
		Spec:        spec,
		Setup:       []string{"go build ./..."},
		Rubric:      bench.DefaultRubric(),
	}
	s := bench.NewStore(filepath.Join(root, "workshop", "benchmarks"))
	if err := s.WriteTask(task); err != nil {
		return err
	}
	short := sha
	if len(short) > 8 {
		short = short[:8]
	}
	fmt.Fprintf(stdout, "froze issue %d → task %s (base %s)\n", f.Issue, task.ID, short)
	return nil
}

// findIssueFile locates workshop/issues/NNNNNN-<slug>.md under root and returns
// its path + the <slug> (the segment after the zero-padded id). The NNNNNN-*.md
// glob convention lives in one place — locateIssueFile (ARCH-DRY); this only adds
// the slug derivation.
func findIssueFile(root string, issueNum int) (path, slug string, err error) {
	path, err = locateIssueFile(filepath.Join(root, "workshop", "issues"), issueNum)
	if err != nil {
		return "", "", err
	}
	id := fmt.Sprintf("%06d", issueNum)
	base := strings.TrimSuffix(filepath.Base(path), ".md")
	return path, strings.TrimPrefix(base, id+"-"), nil
}
