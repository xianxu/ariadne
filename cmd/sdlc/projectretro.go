package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
)

type projectRetroFlags struct {
	Slug, ProjectsDir string
	DryRun            bool
}

func newProjectRetroCmd() *cobra.Command {
	f := projectRetroFlags{}
	cmd := markMutatingCommand(&cobra.Command{Use: "retro", Short: "Append a project retrospective checkpoint", Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProjectRetro(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		}})
	cmd.Flags().StringVar(&f.Slug, "slug", "", "project slug")
	cmd.Flags().StringVar(&f.ProjectsDir, "projects-dir", defaultProjectsDir(), "directory holding project files")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print the retro without writing")
	_ = cmd.MarkFlagRequired("slug")
	return cmd
}

func runProjectRetro(stdout, _ io.Writer, f *projectRetroFlags) error {
	path, err := projectdoc.ResolvePath(f.ProjectsDir, f.Slug)
	if err != nil {
		return err
	}
	d, err := readProject(path)
	if err != nil {
		return err
	}
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return err
	}
	b, err := computeBoard(d, func(ref string) (issueMeta, error) { return projectIssueLookupFn(ref, root) })
	if err != nil {
		return err
	}
	entry := renderRetroEntry(b, projectTodayFn())
	if f.DryRun {
		fmt.Fprint(stdout, entry)
		return nil
	}
	if err := d.AppendToSection("Log", entry); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(d.Render()), 0o644); err != nil {
		return err
	}
	fmt.Fprintln(stdout, path)
	return nil
}

func renderRetroEntry(b board, today string) string {
	days := 0
	if deadline, err := time.Parse("2006-01-02", b.Deadline); err == nil {
		if now, err := time.Parse("2006-01-02", today); err == nil {
			days = int(deadline.Sub(now).Hours() / 24)
		}
	}
	deadline := valueOr(b.Deadline, "-")
	if b.Deadline != "" {
		deadline = fmt.Sprintf("%s (%d days)", b.Deadline, days)
	}
	frontier := valueOr(strings.Join(b.Frontier, ", "), "-")
	return fmt.Sprintf("### %s — retro\n\n**board:** %d/%d done · Σ remaining ≈ %gh · deadline %s · frontier: %s\n\n<where we are + what changed + new forecast — replace this line>", today, b.Done, b.Total, b.RemainingHours, deadline, frontier)
}
