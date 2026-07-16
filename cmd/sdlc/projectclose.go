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

var (
	projectCloseTransitionFn = func(from, to string) *vocab.Transition { return vocab.Project().TransitionFor(from, to) }
	projectCloseStageFileFn  = stageProjectCloseFile
)

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
	tr := projectCloseTransitionFn(from, to)
	if tr == nil {
		return fmt.Errorf("illegal project transition %s → %s; legal from %q: %s", from, to, from, strings.Join(vocab.Project().LegalTransitions(from), ", "))
	}

	var ledgerPath, ledgerNext string
	phaseA, hasPhaseA := projectPhaseA(d)
	actuals, actualsComplete := 0.0, false
	handlers := map[string]func() error{
		"retro-recorded": func() error {
			if f.NoRetro || f.Force {
				cwarn(stderr, "--no-retro (or --force): closing without a recorded project retro")
				return nil
			}
			if err := projectdoc.Guards()["retro-recorded"](d, projectdoc.GuardCtx{Today: projectTodayFn()}); err != nil {
				return fmt.Errorf("retro gate: %w; run `sdlc project retro`, or pass --no-retro (or --force) when it is not applicable", err)
			}
			return nil
		},
		"fog-factor-recorded": func() error {
			if !hasPhaseA {
				cwarn(stderr, "fog factor: missing **phase-a:** <N>h in ## Estimate; recording fog: n/a and skipping ledger")
				return nil
			}
			root, err := gitx.RepoTopLevel()
			if err != nil {
				return err
			}
			var unavailable []string
			actuals, unavailable = rollupProjectActuals(d, root, stderr)
			actualsComplete = len(unavailable) == 0
			if f.NoLedger || f.Force {
				cwarn(stderr, "--no-ledger (or --force): skipping fog-factor ledger")
				return nil
			}
			if !actualsComplete {
				return fmt.Errorf("incomplete MVP actuals (%s); resolve every measured actual_hours or pass --no-ledger (or --force) to close without calibration", strings.Join(unavailable, ", "))
			}
			ledgerPath = filepath.Join(f.BrainDir, filepath.FromSlash(projectLedgerRel))
			ledgerNext, err = prepareProjectLedgerRow(ledgerPath, d.FM("name"), phaseA, actuals, projectTodayFn())
			return err
		},
	}
	for _, name := range tr.Guards {
		handler, ok := handlers[name]
		if !ok {
			return fmt.Errorf("unknown project close guard %q named by the vocabulary", name)
		}
		if err := handler(); err != nil {
			return err
		}
	}

	closeEntry := renderProjectCloseEntry(projectTodayFn(), phaseA, actuals, hasPhaseA, actualsComplete)
	if err := d.AppendToSection("Log", closeEntry); err != nil {
		return err
	}
	d.SetFM("status", to)
	d.SetFM("updated", projectTodayFn())
	archiveDir := vocab.ArchiveSubdir(f.HistoryDir, vocab.ArchiveProjects)
	dest := filepath.Join(archiveDir, filepath.Base(path))
	if err := commitProjectClose(path, dest, d.Render(), ledgerPath, ledgerNext); err != nil {
		return err
	}
	fmt.Fprintln(stdout, dest)
	return nil
}

func rollupProjectActuals(d *projectdoc.Doc, root string, stderr io.Writer) (float64, []string) {
	refs := parseInlineRefs(d.FM("mvp_scope"))
	if len(refs) == 0 {
		return 0, []string{"mvp_scope is empty"}
	}
	actuals := 0.0
	var unavailable []string
	for _, ref := range refs {
		meta, err := projectIssueLookupFn(ref, root)
		reason := ""
		switch {
		case err != nil:
			reason = err.Error()
		case meta.ActualNA:
			reason = "actual_hours is N/A"
		case !meta.ActualAvailable:
			reason = "actual_hours is missing"
		case meta.ActualHours <= 0:
			reason = "actual_hours must be positive"
		default:
			actuals += meta.ActualHours
		}
		if reason != "" {
			unavailable = append(unavailable, fmt.Sprintf("%s: %s", ref, reason))
			cwarn(stderr, fmt.Sprintf("fog factor: %s unavailable (%s)", ref, reason))
		}
	}
	return actuals, unavailable
}

func projectPhaseA(d *projectdoc.Doc) (float64, bool) {
	m := projectPhaseARE.FindStringSubmatch(d.SectionBody("Estimate"))
	if m == nil {
		return 0, false
	}
	hours, err := strconv.ParseFloat(m[1], 64)
	return hours, err == nil && hours > 0
}

func renderProjectCloseEntry(today string, phaseA, actuals float64, hasPhaseA, actualsComplete bool) string {
	if !hasPhaseA {
		return fmt.Sprintf("### %s — close\n\n- fog: n/a", today)
	}
	if !actualsComplete {
		return fmt.Sprintf("### %s — close\n\n- phase-a: %gh\n- actuals: incomplete\n- fog: n/a", today, phaseA)
	}
	return fmt.Sprintf("### %s — close\n\n- phase-a: %gh\n- actuals: %gh\n- fog: %.2f", today, phaseA, actuals, actuals/phaseA)
}

func stageProjectCloseFile(dir, pattern string, data []byte) (path string, err error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	path = f.Name()
	defer func() {
		if closeErr := f.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
		if err != nil {
			_ = os.Remove(path)
		}
	}()
	if err = f.Chmod(0o644); err != nil {
		return "", err
	}
	_, err = f.Write([]byte(data))
	return path, err
}

// commitProjectClose stages both records before changing either durable path.
// Renames are compensated in reverse order on failure, so a reported error
// leaves the live project and ledger at their pre-command contents.
func commitProjectClose(livePath, dest, projectNext, ledgerPath, ledgerNext string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("archived project already exists: %s", dest)
	} else if !os.IsNotExist(err) {
		return err
	}
	projectStage, err := projectCloseStageFileFn(filepath.Dir(dest), ".sdlc-project-close-project-*", []byte(projectNext))
	if err != nil {
		return err
	}
	defer os.Remove(projectStage)

	ledgerStage := ""
	if ledgerPath != "" {
		ledgerStage, err = projectCloseStageFileFn(filepath.Dir(ledgerPath), ".sdlc-project-close-ledger-*", []byte(ledgerNext))
		if err != nil {
			return err
		}
		defer os.Remove(ledgerStage)
	}

	liveBackup, err := reserveProjectCloseBackup(filepath.Dir(livePath), ".sdlc-project-close-live-backup-*")
	if err != nil {
		return err
	}
	ledgerBackup := ""
	ledgerBackupActive := false
	if ledgerPath != "" {
		ledgerBackup, err = reserveProjectCloseBackup(filepath.Dir(ledgerPath), ".sdlc-project-close-ledger-backup-*")
		if err != nil {
			return err
		}
		if err := os.Rename(ledgerPath, ledgerBackup); err != nil {
			return err
		}
		ledgerBackupActive = true
		if err := os.Rename(ledgerStage, ledgerPath); err != nil {
			if rollbackErr := os.Rename(ledgerBackup, ledgerPath); rollbackErr != nil {
				return fmt.Errorf("ledger commit failed: %v; original remains at %s because rollback failed: %w", err, ledgerBackup, rollbackErr)
			}
			ledgerBackupActive = false
			return err
		}
	}
	rollbackLedger := func() error {
		if ledgerPath == "" {
			return nil
		}
		if err := os.Remove(ledgerPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := os.Rename(ledgerBackup, ledgerPath); err != nil {
			return err
		}
		ledgerBackupActive = false
		return nil
	}
	if err := os.Rename(livePath, liveBackup); err != nil {
		if rollbackErr := rollbackLedger(); rollbackErr != nil {
			return fmt.Errorf("archive preparation failed: %v; ledger rollback failed: %w", err, rollbackErr)
		}
		return err
	}
	liveBackupActive := true
	if err := os.Rename(projectStage, dest); err != nil {
		projectRollbackErr := os.Rename(liveBackup, livePath)
		if projectRollbackErr == nil {
			liveBackupActive = false
		}
		ledgerRollbackErr := rollbackLedger()
		if projectRollbackErr != nil || ledgerRollbackErr != nil {
			return fmt.Errorf("archive commit failed: %v; project rollback: %v; ledger rollback: %v", err, projectRollbackErr, ledgerRollbackErr)
		}
		return err
	}
	if liveBackupActive {
		if err := os.Remove(liveBackup); err != nil {
			return fmt.Errorf("project close committed but original backup cleanup failed at %s: %w", liveBackup, err)
		}
	}
	if ledgerBackupActive {
		if err := os.Remove(ledgerBackup); err != nil {
			return fmt.Errorf("project close committed but ledger backup cleanup failed at %s: %w", ledgerBackup, err)
		}
	}
	return nil
}

func reserveProjectCloseBackup(dir, pattern string) (string, error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	path := f.Name()
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	if err := os.Remove(path); err != nil {
		return "", err
	}
	return path, nil
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
