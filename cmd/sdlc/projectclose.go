package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/vocab"
)

var (
	projectCloseTransitionFn = func(from, event string) *vocab.Transition { return vocab.Project().TransitionForEvent(from, event) }
	projectCloseStageFileFn  = stageProjectCloseFile
	projectCloseRenameFn     = os.Rename
)

type projectCloseFlags struct {
	Slug, ProjectsDir, HistoryDir string
	Verified                      string
	Drop, NoRetro, Force, DryRun  bool
}

// projectCloseTerminalFn is the close's MVP lookup; tests stub it.
var projectCloseTerminalFn = projectIssueTerminalFn

func newProjectCloseCmd() *cobra.Command {
	f := projectCloseFlags{}
	cmd := markMutatingCommand(&cobra.Command{Use: "close", Short: "Close or drop a project and archive its record", Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			guardSpineRepo(cmd.ErrOrStderr()) // #176 lifecycle guard; must precede verb validation
			if strings.TrimSpace(f.Slug) == "" {
				return fmt.Errorf("--slug is required")
			}
			return runProjectClose(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		}})
	cmd.Flags().StringVar(&f.Slug, "slug", "", "project slug")
	cmd.Flags().StringVar(&f.ProjectsDir, "projects-dir", defaultProjectsDir(), "directory holding live project files")
	cmd.Flags().StringVar(&f.HistoryDir, "history-dir", vocab.Project().Discovery().Archive, "history archive root")
	cmd.Flags().StringVar(&f.Verified, "verified", "", "evidence the done_when audit passed (required to close; recorded in the Log)")
	cmd.Flags().BoolVar(&f.Drop, "drop", false, "close through the lifecycle's dropped transition")
	cmd.Flags().BoolVar(&f.NoRetro, "no-retro", false, "waive the retrospective gate")
	cmd.Flags().BoolVar(&f.Force, "force", false, "waive the retro and MVP gates")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "evaluate every gate and print the archive move; write nothing")
	return cmd
}

func runProjectClose(ctx context.Context, stdout, stderr io.Writer, f *projectCloseFlags) error {
	path, err := projectdoc.ResolvePath(f.ProjectsDir, f.Slug)
	if err != nil {
		return err
	}
	d, err := readProject(path)
	if err != nil {
		return err
	}
	metadata, err := d.Metadata()
	if err != nil {
		return err
	}
	today := projectTodayFn()
	from := metadata.Status
	event := "close"
	m := vocab.Project()
	if f.Drop {
		event = "drop"
		if !m.IsExecuting(from) {
			return fmt.Errorf("project close --drop requires a live project (%s; current %q); drop a defined project with `sdlc project set-status --to dropped`", strings.Join(m.Categories["executing"], " or "), from)
		}
	}

	// Every gate is evaluated, in order, so a dry run reports them all; a real
	// close refuses on the first.
	var problems []error
	tr := projectCloseTransitionFn(from, event)
	guarded := tr
	if tr == nil {
		required := m.FirstTransitionForEvent(event)
		guarded = required
		switch {
		case !f.Drop && m.IsExecuting(from) && required != nil:
			problems = append(problems, fmt.Errorf("project is %s; wrap it first with `sdlc project set-status --to %s` (every mvp_scope issue landed)", from, required.From))
		case required != nil:
			problems = append(problems, fmt.Errorf("project close requires status %s (current %q); advance it through the project lifecycle funnel first", required.From, from))
		default:
			return fmt.Errorf("no modeled %s transition from project status %q", event, from)
		}
	}
	handlers := map[string]func() error{
		"retro-recorded": func() error {
			if f.NoRetro || f.Force {
				cwarn(stderr, "--no-retro (or --force): closing without a recorded project retro")
				return nil
			}
			if err := projectdoc.Guards()["retro-recorded"](d, projectdoc.GuardCtx{Today: today}); err != nil {
				return fmt.Errorf("retro gate: %w; run `sdlc project retro`, or pass --no-retro (or --force) when it is not applicable", err)
			}
			return nil
		},
	}
	for _, name := range guarded.Guards {
		handler, ok := handlers[name]
		if !ok {
			return fmt.Errorf("unknown project close guard %q named by the vocabulary", name)
		}
		if err := handler(); err != nil {
			problems = append(problems, err)
		}
	}
	if !f.Drop {
		// #373: a close records its evidence and re-checks the MVP, since scope
		// can grow while a project sits in closing.
		if strings.TrimSpace(f.Verified) == "" {
			problems = append(problems, fmt.Errorf("--verified '<evidence>' is required: say how the done_when audit passed"))
		}
		if err := projectCloseMVPGate(ctx, path, metadata, f.Force, stderr); err != nil {
			problems = append(problems, err)
		}
	}

	archiveDir := vocab.ArchiveSubdir(f.HistoryDir, vocab.ArchiveProjects)
	dest := filepath.Join(archiveDir, filepath.Base(path))
	if f.DryRun {
		return printProjectCloseDryRun(stdout, path, dest, from, guarded.To, problems)
	}
	if len(problems) > 0 {
		return problems[0]
	}

	closeEntry := fmt.Sprintf("### %s — %s\n\n- status: %s → %s", today, event, from, tr.To)
	if v := strings.TrimSpace(f.Verified); v != "" && !f.Drop {
		closeEntry += "\n- verified: " + v
	}
	if err := d.AppendToSection("Log", closeEntry); err != nil {
		return err
	}
	d.SetFM("status", tr.To)
	d.SetFM("updated", today)
	if err := commitProjectClose(path, dest, d.Render()); err != nil {
		return err
	}
	fmt.Fprintln(stdout, dest)
	return nil
}

// projectCloseMVPGate refuses a close while an mvp_scope issue is open or
// unreadable (MVPProgress, the guard's own definition). --force waives it.
func projectCloseMVPGate(ctx context.Context, path string, meta projectdoc.Metadata, force bool, stderr io.Writer) error {
	if len(meta.MVPScope) == 0 {
		return nil
	}
	mvp := projectdoc.MVPProgress(meta.MVPScope, projectCloseTerminalFn(ctx, path))
	if mvp.Done() {
		return nil
	}
	var why []string
	if len(mvp.Open) > 0 {
		why = append(why, "not landed: "+strings.Join(mvp.Open, ", "))
	}
	for _, err := range mvp.Errors {
		why = append(why, err.Error())
	}
	if force {
		cwarn(stderr, "--force: closing with the MVP incomplete ("+strings.Join(why, "; ")+")")
		return nil
	}
	return fmt.Errorf("MVP gate: %s; land them, move them to explicitly_out, or pass --force", strings.Join(why, "; "))
}

// printProjectCloseDryRun reports the close without writing: the transition,
// every refusal, and where the record would go.
func printProjectCloseDryRun(stdout io.Writer, path, dest, from, to string, problems []error) error {
	fmt.Fprintf(stdout, "dry run: %s → %s\n", from, to)
	for _, p := range problems {
		fmt.Fprintf(stdout, "  refuses: %v\n", p)
	}
	if real, linked := projectLinkTarget(path); linked {
		fmt.Fprintf(stdout, "  would update %s in place and move the link %s → %s\n", real, path, dest)
	} else {
		fmt.Fprintf(stdout, "  would archive %s → %s\n", path, dest)
	}
	if len(problems) > 0 {
		return fmt.Errorf("dry run: the close would refuse (%d gate(s))", len(problems))
	}
	return nil
}

// projectLinkTarget resolves a live record that is a symlink (the ops layout:
// workshop/projects/x.md → teams/…/x/project.md).
func projectLinkTarget(path string) (string, bool) {
	fi, err := os.Lstat(path)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		return "", false
	}
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", false
	}
	return real, true
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

// commitProjectClose stages the archived record before moving the live one,
// and restores the live record if the archive rename fails.
func commitProjectClose(livePath, dest, projectNext string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	if _, err := os.Lstat(dest); err == nil {
		return fmt.Errorf("archived project already exists: %s", dest)
	} else if !os.IsNotExist(err) {
		return err
	}
	if real, linked := projectLinkTarget(livePath); linked {
		return commitLinkedProjectClose(livePath, real, dest, projectNext)
	}
	projectStage, err := projectCloseStageFileFn(filepath.Dir(dest), ".sdlc-project-close-project-*", []byte(projectNext))
	if err != nil {
		return err
	}
	defer os.Remove(projectStage)

	liveBackup, err := reserveProjectCloseBackup(filepath.Dir(livePath), ".sdlc-project-close-live-backup-*")
	if err != nil {
		return err
	}
	if err := projectCloseRenameFn(livePath, liveBackup); err != nil {
		return err
	}
	if err := projectCloseRenameFn(projectStage, dest); err != nil {
		if rollbackErr := projectCloseRenameFn(liveBackup, livePath); rollbackErr != nil {
			return fmt.Errorf("archive commit failed: %v; original remains at %s because rollback failed: %w", err, liveBackup, rollbackErr)
		}
		return err
	}
	if err := os.Remove(liveBackup); err != nil {
		return fmt.Errorf("project close committed but original backup cleanup failed at %s: %w", liveBackup, err)
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

// commitLinkedProjectClose closes a linked record (#373): the real file, which
// sits in a team folder beside the project's handoffs and reviews, is updated
// in place, and the link moves to the archive, retargeted. Moving the link
// alone used to orphan the real file as a live-looking record.
func commitLinkedProjectClose(livePath, real, dest, projectNext string) error {
	destDir, err := filepath.EvalSymlinks(filepath.Dir(dest)) // same spelling as real (/var vs /private/var)
	if err != nil {
		return err
	}
	target, err := filepath.Rel(destDir, real)
	if err != nil {
		return err
	}
	if err := os.Symlink(target, dest); err != nil {
		return err
	}
	stage, err := projectCloseStageFileFn(filepath.Dir(real), ".sdlc-project-close-project-*", []byte(projectNext))
	if err != nil {
		_ = os.Remove(dest)
		return err
	}
	defer os.Remove(stage)
	if err := projectCloseRenameFn(stage, real); err != nil {
		_ = os.Remove(dest)
		return err
	}
	return os.Remove(livePath)
}
