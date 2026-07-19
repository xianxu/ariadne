package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/vocab"
)

type projectSetStatusFlags struct {
	Slug, To, Reality, Coverage, ProjectsDir string
	PlannedFinish, BrainDir                  string
	Force                                    bool
}

// plannedFinishDecision carries the two candidate planned-finish dates into
// applyProjectStatus, which owns the doc write and applies the precedence:
// Manual (explicit --planned-finish) > a pre-existing frontmatter value >
// Derived (the forecast's projected finish). Empty strings mean "not offered".
type plannedFinishDecision struct {
	Derived string // the forecast projection (committed transition)
	Manual  string // explicit --planned-finish
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
	cmd.Flags().StringVar(&f.Reality, "reality", "", "reality-check evidence (legacy fallback when no throughput baseline is blessed)")
	cmd.Flags().StringVar(&f.Coverage, "coverage", "", "issues-cover-PRD evidence")
	cmd.Flags().StringVar(&f.PlannedFinish, "planned-finish", "", "override the forecast-derived planned_finish (recorded as a manual override)")
	cmd.Flags().BoolVar(&f.Force, "force", false, "waive named transition guards (not lifecycle legality)")
	cmd.Flags().StringVar(&f.ProjectsDir, "projects-dir", defaultProjectsDir(), "directory holding project files")
	cmd.Flags().StringVar(&f.BrainDir, "brain-dir", "../brain", "brain root holding the throughput baseline (for the commit forecast)")
	_ = cmd.MarkFlagRequired("slug")
	_ = cmd.MarkFlagRequired("to")
	return cmd
}

func runProjectSetStatus(stdout, stderr io.Writer, f *projectSetStatusFlags) error {
	path, err := projectdoc.ResolvePath(f.ProjectsDir, f.Slug)
	if err != nil {
		return err
	}
	today := projectTodayFn()
	ctx := projectdoc.GuardCtx{Today: today, Evidence: map[string]string{"reality-check": f.Reality, "issues-cover-prd": f.Coverage}}
	decision := plannedFinishDecision{Manual: f.PlannedFinish}

	// #182: the commit transition COMPUTES a calendar forecast and INFORMS —
	// it never refuses on the answer. The rendered statement becomes the
	// reality-check evidence (so the existing evidenceGuard passes on having
	// computed), and planned_finish derives from it. With no blessed baseline,
	// fall back to the legacy --reality prose (the guard still requires it).
	if f.To == "committed" {
		if forecastErr := computeCommitForecast(stdout, stderr, path, f, &ctx, &decision, today); forecastErr != nil {
			return forecastErr
		}
	}

	prev, changed, err := applyProjectStatus(path, f.To, f.Force, ctx, decision)
	if err != nil {
		return err
	}
	if changed {
		fmt.Fprintf(stdout, "%s: status %s → %s\n", path, prev, f.To)
	}
	return nil
}

// computeCommitForecast runs the calendar forecast for a →committed transition.
// On success it prints the statement, injects it as reality-check evidence
// (only when the operator gave no --reality), and sets the derived
// planned_finish. On errNoBaseline it leaves --reality handling untouched and,
// if the operator also gave no --reality, returns a guard-shaped refusal with
// the bless hint. A compute error other than errNoBaseline is surfaced.
func computeCommitForecast(stdout, stderr io.Writer, path string, f *projectSetStatusFlags, ctx *projectdoc.GuardCtx, decision *plannedFinishDecision, today string) error {
	d, err := readProject(path)
	if err != nil {
		return err
	}
	// The forecast (and its no-baseline refusal) applies only to the LEGAL
	// defined→committed transition. Skip it for a re-run on an already-committed
	// project (idempotent no-op) or an illegal source — applyProjectStatus owns
	// that legality, and computing here would emit a spurious forecast or a
	// misleading no-baseline error on a transition that won't happen.
	if d.FM("status") != "defined" {
		return nil
	}
	forecast, deadline, ferr := forecastForProject(d, path, f.BrainDir, today)
	if ferr == errNoBaseline {
		if strings.TrimSpace(f.Reality) == "" && !f.Force {
			return fmt.Errorf("guard reality-check: no throughput baseline blessed — bless one (`sdlc project throughput --bless <FROM>..<TO>`) so the commit forecast computes, or pass --reality '<why it fits>'")
		}
		return nil // legacy path: --reality prose carries the evidence
	}
	if ferr != nil {
		cwarn(stderr, "commit forecast unavailable: "+ferr.Error())
		return nil // informational — never block the transition on a compute error
	}
	statement := projectdoc.RenderForecast(forecast, deadline)
	cinfo(stdout, statement)
	if strings.TrimSpace(f.Reality) == "" {
		ctx.Evidence["reality-check"] = "computed: " + statement
	}
	decision.Derived = forecast.ProjectedFinish
	return nil
}

func applyProjectStatus(path, to string, force bool, ctx projectdoc.GuardCtx, plannedFinish plannedFinishDecision) (prev string, changed bool, err error) {
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

	// Apply the derived/manual planned_finish BEFORE the guards run: the
	// baseline-set guard requires planned_finish to be present, and #182
	// DERIVES it at the commit transition — so the derived value must be in
	// the doc before that guard checks it. On a guard failure nothing is
	// written (the write is at the end), so this in-memory mutation is safe.
	plannedFinishNote := applyPlannedFinish(d, plannedFinish, ctx.Today)

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
		// The planned_finish provenance note is recorded only when the
		// transition actually proceeds past its guards (mirrors the evidence
		// block); the FM itself was set above so baseline-set could see it.
		if plannedFinishNote != "" {
			if err := d.AppendToSection("Log", plannedFinishNote); err != nil {
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

// applyPlannedFinish sets planned_finish per the #182 precedence and returns a
// Log note describing which path was taken (empty when nothing was offered).
// Manual (explicit flag) wins; else a pre-existing value is kept; else the
// forecast-derived date is set.
func applyPlannedFinish(d *projectdoc.Doc, pf plannedFinishDecision, today string) string {
	switch {
	case pf.Manual != "":
		d.SetFM("planned_finish", pf.Manual)
		note := fmt.Sprintf("### %s — planned_finish\n\n- planned_finish set manually: %s", today, pf.Manual)
		if pf.Derived != "" {
			note += fmt.Sprintf(" (forecast projected %s)", pf.Derived)
		}
		return note
	case strings.TrimSpace(d.FM("planned_finish")) != "":
		if pf.Derived == "" {
			return "" // nothing computed, existing value stands — no note needed
		}
		return fmt.Sprintf("### %s — planned_finish\n\n- manual planned_finish kept: %s (forecast projected %s)", today, strings.TrimSpace(d.FM("planned_finish")), pf.Derived)
	case pf.Derived != "":
		d.SetFM("planned_finish", pf.Derived)
		return fmt.Sprintf("### %s — planned_finish\n\n- planned_finish derived from the commit forecast: %s", today, pf.Derived)
	}
	return ""
}
