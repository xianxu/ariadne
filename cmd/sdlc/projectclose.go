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
	Drop, NoRetro, Force          bool
}

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
	cmd.Flags().BoolVar(&f.Drop, "drop", false, "close through the lifecycle's dropped transition")
	cmd.Flags().BoolVar(&f.NoRetro, "no-retro", false, "waive the retrospective gate")
	cmd.Flags().BoolVar(&f.Force, "force", false, "waive the retro gate")
	return cmd
}

func runProjectClose(_ context.Context, stdout, stderr io.Writer, f *projectCloseFlags) error {
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
	tr := projectCloseTransitionFn(from, event)
	if tr == nil {
		required := m.FirstTransitionForEvent(event)
		if !f.Drop && m.IsExecuting(from) && required != nil {
			return fmt.Errorf("project is %s; wrap it first with `sdlc project set-status --to %s` (every mvp_scope issue landed)", from, required.From)
		}
		if required != nil {
			return fmt.Errorf("project close requires status %s (current %q); advance it through the project lifecycle funnel first", required.From, from)
		}
		return fmt.Errorf("no modeled %s transition from project status %q", event, from)
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
	for _, name := range tr.Guards {
		handler, ok := handlers[name]
		if !ok {
			return fmt.Errorf("unknown project close guard %q named by the vocabulary", name)
		}
		if err := handler(); err != nil {
			return err
		}
	}

	closeEntry := fmt.Sprintf("### %s — %s\n\n- status: %s → %s", today, event, from, tr.To)
	if err := d.AppendToSection("Log", closeEntry); err != nil {
		return err
	}
	d.SetFM("status", tr.To)
	d.SetFM("updated", today)
	archiveDir := vocab.ArchiveSubdir(f.HistoryDir, vocab.ArchiveProjects)
	dest := filepath.Join(archiveDir, filepath.Base(path))
	if err := commitProjectClose(path, dest, d.Render()); err != nil {
		return err
	}
	fmt.Fprintln(stdout, dest)
	return nil
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
