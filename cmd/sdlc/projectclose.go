package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/vocab"
)

const projectLedgerRel = "data/life/42shots/velocity/estimate-logic-project-v1.md"

var projectPhaseARE = regexp.MustCompile(`(?m)^\*\*phase-a:\*\*\s+((?:\d+(?:\.\d+)?|\.\d+))h\s*$`)

type projectCloseFlags struct {
	Slug, ProjectsDir, HistoryDir, BrainDir string
	Drop, NoRetro, NoLedger, Force          bool
}

func newProjectCloseCmd() *cobra.Command {
	f := projectCloseFlags{}
	cmd := markMutatingCommand(&cobra.Command{Use: "close", Short: "Close or drop a project and archive its record", Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			guardSpineRepo(cmd.ErrOrStderr()) // #176 lifecycle guard; must precede verb validation
			if strings.TrimSpace(f.Slug) == "" {
				return fmt.Errorf("--slug is required")
			}
			return runProjectClose(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		}})
	cmd.Flags().StringVar(&f.Slug, "slug", "", "project slug")
	cmd.Flags().StringVar(&f.ProjectsDir, "projects-dir", defaultProjectsDir(), "directory holding live project files")
	cmd.Flags().StringVar(&f.HistoryDir, "history-dir", vocab.Project().Discovery().Archive, "history archive root")
	cmd.Flags().StringVar(&f.BrainDir, "brain-dir", "../brain", "brain repository directory")
	cmd.Flags().BoolVar(&f.Drop, "drop", false, "close through the lifecycle's dropped transition")
	cmd.Flags().BoolVar(&f.NoRetro, "no-retro", false, "waive the retrospective gate")
	cmd.Flags().BoolVar(&f.NoLedger, "no-ledger", false, "skip the Phase-A fog-factor ledger")
	cmd.Flags().BoolVar(&f.Force, "force", false, "waive the retro and ledger gates")
	return cmd
}

func runProjectClose(stdout, stderr io.Writer, f *projectCloseFlags) error {
	path, err := projectdoc.ResolvePath(f.ProjectsDir, f.Slug)
	if err != nil {
		return err
	}
	d, err := readProject(path)
	if err != nil {
		return err
	}
	from := d.FM("status")
	to := "done"
	if f.Drop {
		to = "dropped"
		if from != "executing" && from != "paused" {
			return fmt.Errorf("project close --drop requires status executing or paused (current %q); use the lifecycle's ordinary status transition for a pre-execution drop", from)
		}
	} else if from == "paused" {
		return fmt.Errorf("project is paused; resume first with `sdlc project set-status --to executing`")
	} else if from != "executing" {
		return fmt.Errorf("project close requires status executing (current %q); advance it through the project lifecycle funnel first", from)
	}
	tr := vocab.Project().TransitionFor(from, to)
	if tr == nil {
		return fmt.Errorf("illegal project transition %s → %s; legal from %q: %s", from, to, from, strings.Join(vocab.Project().LegalTransitions(from), ", "))
	}

	skipRetro := f.NoRetro || f.Force
	if transitionHasGuard(tr.Guards, "retro-recorded") {
		if skipRetro {
			cwarn(stderr, "--no-retro (or --force): closing without a recorded project retro")
		} else if err := projectdoc.Guards()["retro-recorded"](d, projectdoc.GuardCtx{Today: projectTodayFn()}); err != nil {
			return fmt.Errorf("retro gate: %w; run `sdlc project retro`, or pass --no-retro (or --force) when it is not applicable", err)
		}
	}

	var ledgerPath, ledgerNext string
	phaseA, hasPhaseA := projectPhaseA(d)
	actuals := 0.0
	if !f.Drop && hasPhaseA {
		root, err := gitx.RepoTopLevel()
		if err != nil {
			return err
		}
		for _, ref := range parseInlineRefs(d.FM("mvp_scope")) {
			meta, lookupErr := projectIssueLookupFn(ref, root)
			if lookupErr != nil || meta.ActualHours <= 0 {
				detail := "missing or N/A actual_hours"
				if lookupErr != nil {
					detail = lookupErr.Error()
				}
				cwarn(stderr, fmt.Sprintf("fog factor: skipping %s (%s)", ref, detail))
				continue
			}
			actuals += meta.ActualHours
		}
		if f.NoLedger || f.Force {
			cwarn(stderr, "--no-ledger (or --force): skipping fog-factor ledger")
		} else {
			ledgerPath = filepath.Join(f.BrainDir, filepath.FromSlash(projectLedgerRel))
			ledgerNext, err = prepareProjectLedgerRow(ledgerPath, d.FM("name"), phaseA, actuals, projectTodayFn())
			if err != nil {
				return err
			}
		}
	} else if !f.Drop {
		cwarn(stderr, "fog factor: missing **phase-a:** <N>h in ## Estimate; recording fog: n/a and skipping ledger")
	}

	closeEntry := renderProjectCloseEntry(projectTodayFn(), phaseA, actuals, hasPhaseA)
	if err := d.AppendToSection("Log", closeEntry); err != nil {
		return err
	}
	d.SetFM("status", to)
	d.SetFM("updated", projectTodayFn())
	archiveDir := vocab.ArchiveSubdir(f.HistoryDir, vocab.ArchiveProjects)
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		return err
	}
	dest := filepath.Join(archiveDir, filepath.Base(path))
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("archived project already exists: %s", dest)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.WriteFile(path, []byte(d.Render()), 0o644); err != nil {
		return err
	}
	if err := os.Rename(path, dest); err != nil {
		return err
	}
	if ledgerPath != "" {
		if err := os.WriteFile(ledgerPath, []byte(ledgerNext), 0o644); err != nil {
			return fmt.Errorf("project archived but fog ledger write failed: %w", err)
		}
	}
	fmt.Fprintln(stdout, dest)
	return nil
}

func transitionHasGuard(guards []string, want string) bool {
	for _, guard := range guards {
		if guard == want {
			return true
		}
	}
	return false
}

func projectPhaseA(d *projectdoc.Doc) (float64, bool) {
	m := projectPhaseARE.FindStringSubmatch(d.SectionBody("Estimate"))
	if m == nil {
		return 0, false
	}
	hours, err := strconv.ParseFloat(m[1], 64)
	return hours, err == nil && hours > 0
}

func renderProjectCloseEntry(today string, phaseA, actuals float64, hasPhaseA bool) string {
	if !hasPhaseA {
		return fmt.Sprintf("### %s — close\n\n- fog: n/a", today)
	}
	return fmt.Sprintf("### %s — close\n\n- phase-a: %gh\n- actuals: %gh\n- fog: %.2f", today, phaseA, actuals, actuals/phaseA)
}

func prepareProjectLedgerRow(path, name string, phaseA, actuals float64, today string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("fog ledger unavailable: %w; add `## Fog ledger` and its markdown table, or pass --no-ledger (or --force)", err)
	}
	text := string(b)
	heading := strings.Index(text, "## Fog ledger")
	if heading < 0 {
		return "", fmt.Errorf("fog ledger missing exact heading `## Fog ledger`; add the heading and its markdown table, or pass --no-ledger (or --force)")
	}
	tail := text[heading:]
	lines := strings.Split(tail, "\n")
	tableEnd := -1
	seenHeader, seenDivider := false, false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i > 0 && strings.HasPrefix(trimmed, "## ") {
			break
		}
		if strings.HasPrefix(trimmed, "|") {
			if !seenHeader {
				seenHeader = true
			} else if !seenDivider {
				seenDivider = strings.Contains(trimmed, "---")
			}
			tableEnd = i
		}
	}
	if !seenHeader || !seenDivider || tableEnd < 0 {
		return "", fmt.Errorf("fog ledger `## Fog ledger` is missing its markdown table; add it, or pass --no-ledger (or --force)")
	}
	row := fmt.Sprintf("| %s | %gh | %gh | %.2f | %s |", name, phaseA, actuals, actuals/phaseA, today)
	insertAt := heading
	for i := 0; i <= tableEnd; i++ {
		insertAt += len(lines[i]) + 1
	}
	return text[:insertAt] + row + "\n" + text[insertAt:], nil
}
