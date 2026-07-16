package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/vocab"
)

type projectSetStatusFlags struct {
	Slug, To, Reality, Coverage, ProjectsDir string
	Force                                    bool
}

var projectGuardsFn = projectdoc.Guards

func newProjectSetStatusCmd() *cobra.Command {
	f := projectSetStatusFlags{}
	cmd := markMutatingCommand(&cobra.Command{Use: "set-status", Short: "Move a project through its guarded lifecycle", Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProjectSetStatus(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		}})
	cmd.Flags().StringVar(&f.Slug, "slug", "", "project slug")
	cmd.Flags().StringVar(&f.To, "to", "", "target project status")
	cmd.Flags().StringVar(&f.Reality, "reality", "", "reality-check evidence")
	cmd.Flags().StringVar(&f.Coverage, "coverage", "", "issues-cover-PRD evidence")
	cmd.Flags().BoolVar(&f.Force, "force", false, "waive named transition guards (not lifecycle legality)")
	cmd.Flags().StringVar(&f.ProjectsDir, "projects-dir", defaultProjectsDir(), "directory holding project files")
	_ = cmd.MarkFlagRequired("slug")
	_ = cmd.MarkFlagRequired("to")
	return cmd
}

func runProjectSetStatus(stdout, _ io.Writer, f *projectSetStatusFlags) error {
	path := filepath.Join(f.ProjectsDir, f.Slug+".md")
	ctx := projectdoc.GuardCtx{Today: projectTodayFn(), Evidence: map[string]string{"reality-check": f.Reality, "issues-cover-prd": f.Coverage}}
	prev, changed, err := applyProjectStatus(path, f.To, f.Force, ctx)
	if err != nil {
		return err
	}
	if changed {
		fmt.Fprintf(stdout, "%s: status %s → %s\n", path, prev, f.To)
	}
	return nil
}

func applyProjectStatus(path, to string, force bool, ctx projectdoc.GuardCtx) (prev string, changed bool, err error) {
	m := vocab.Project()
	valid := false
	for _, status := range m.AllStatuses() {
		if status == to {
			valid = true
			break
		}
	}
	if !valid {
		return "", false, fmt.Errorf("invalid status %q (valid: %s)", to, strings.Join(m.AllStatuses(), ", "))
	}
	d, err := readProject(path)
	if err != nil {
		return "", false, err
	}
	prev = d.FM("status")
	if prev != to && !m.CanTransition(prev, to) {
		return prev, false, fmt.Errorf("illegal transition %s → %s; legal from %q: %s", prev, to, prev, strings.Join(m.LegalTransitions(prev), ", "))
	}
	if to == "done" {
		return prev, false, fmt.Errorf("refusing → done; use `sdlc project close`")
	}
	if tr := m.TransitionFor(prev, to); tr != nil {
		guards := projectGuardsFn()
		for _, name := range tr.Guards {
			guard, ok := guards[name]
			if !ok {
				return prev, false, fmt.Errorf("unknown project guard %q named by the vocabulary", name)
			}
			if !force {
				if err := guard(d, ctx); err != nil {
					return prev, false, fmt.Errorf("guard %s: %w", name, err)
				}
			}
		}
		var evidence []string
		for _, name := range tr.Guards {
			if value := strings.TrimSpace(ctx.Evidence[name]); value != "" {
				evidence = append(evidence, fmt.Sprintf("- %s: %s", name, value))
			}
		}
		if len(evidence) > 0 {
			block := fmt.Sprintf("### %s — transition evidence\n\n%s", ctx.Today, strings.Join(evidence, "\n"))
			if err := d.AppendToSection("Log", block); err != nil {
				return prev, false, err
			}
		}
	}
	d.SetFM("status", to)
	d.SetFM("updated", ctx.Today)
	raw, err := os.ReadFile(path)
	if err != nil {
		return prev, false, err
	}
	next := d.Render()
	changed = next != string(raw)
	if changed {
		if err := os.WriteFile(path, []byte(next), 0o644); err != nil {
			return prev, false, err
		}
	}
	return prev, changed, nil
}
