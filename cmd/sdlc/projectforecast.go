// projectforecast.go — the IO seams feeding the pure train forecast (#329):
// each car's issue card (status, estimate_hours) through the shared cross-repo
// lookup, and the calibration ledger through estimate.CalibrateForecast. The
// pure core (internal/project/trainforecast.go) never touches the filesystem.
// `sdlc project forecast` prints the statement; show/status print it
// best-effort through forecastLine.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/vocab"
	"github.com/xianxu/ariadne/pkg/workspace"
)

type projectForecastFlags struct{ Slug, ProjectsDir, BrainDir string }

func newProjectForecastCmd() *cobra.Command {
	f := projectForecastFlags{}
	cmd := &cobra.Command{Use: "forecast [file]", Short: "Forecast a project's remaining focused hours from its trains (read-only)", Args: cobra.MaximumNArgs(1), SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) == 1 {
				path = args[0]
			}
			return runProjectForecast(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), &f, path)
		}}
	cmd.Flags().StringVar(&f.Slug, "slug", "", "project slug (or pass the project file)")
	cmd.Flags().StringVar(&f.ProjectsDir, "projects-dir", defaultProjectsDir(), "directory holding project files")
	cmd.Flags().StringVar(&f.BrainDir, "brain-dir", "../brain", "brain root holding the calibration ledger")
	return cmd
}

func runProjectForecast(ctx context.Context, stdout, stderr io.Writer, f *projectForecastFlags, path string) error {
	if (path == "") == (f.Slug == "") {
		return fmt.Errorf("choose one of <file> or --slug")
	}
	if path == "" {
		var err error
		if path, err = projectdoc.ResolvePath(f.ProjectsDir, f.Slug); err != nil {
			return err
		}
	}
	d, err := readProject(path)
	if err != nil {
		return err
	}
	fc, plan, err := forecastForProject(ctx, d, path, f.BrainDir)
	if err != nil {
		return err
	}
	for _, w := range plan.Warnings {
		cwarn(stderr, w)
	}
	fmt.Fprintln(stdout, projectdoc.RenderTrainForecast(fc))
	return nil
}

// calibrationLedgerPath resolves the calibration ledger, honoring
// WF_CALIB_LEDGER first (test override), else the brain velocity path.
func calibrationLedgerPath(brainDir string) string {
	if p := os.Getenv("WF_CALIB_LEDGER"); p != "" {
		return p
	}
	return estimate.VelocityPath(brainDir, "calibration-ledger.tsv")
}

// loadForecastCalibration reads the ledger and calibrates. An unreadable or
// row-less ledger is an error: a forecast without calibration would be a
// guess with a range drawn on it.
func loadForecastCalibration(brainDir string) (estimate.ForecastCalibration, error) {
	brainAbs, _ := filepath.Abs(brainDir)
	path := calibrationLedgerPath(brainAbs)
	data, err := os.ReadFile(path)
	if err != nil {
		return estimate.ForecastCalibration{}, fmt.Errorf("calibration ledger %s: %w", path, err)
	}
	rows, err := estimate.ParseRowsChecked(string(data))
	if err == nil {
		rows, err = estimate.ResolveCorrections(rows)
	}
	if err != nil {
		return estimate.ForecastCalibration{}, fmt.Errorf("calibration ledger %s: %w", path, err)
	}
	cal, ok := estimate.CalibrateForecast(rows)
	if !ok {
		return estimate.ForecastCalibration{}, fmt.Errorf("calibration ledger %s has no row with a build/converge split and an estimate (ariadne#234)", path)
	}
	return cal, nil
}

// forecastForProject assembles the inputs and runs the pure forecast. The
// issue-lookup vantage is the project's own repo root, derived from the
// project path resolved to absolute (a relative path would resolve to "." and
// fail every cross-repo lookup); its repo name is the default for a bare #N.
func forecastForProject(ctx context.Context, d *projectdoc.Doc, projectPath, brainDir string) (projectdoc.TrainForecast, projectdoc.TrainPlan, error) {
	identity, err := projectCheckout(projectPath)
	if err != nil {
		return projectdoc.TrainForecast{}, projectdoc.TrainPlan{}, err
	}
	plan := d.Trains(identity.Repo)
	if len(plan.Trains) == 0 {
		return projectdoc.TrainForecast{}, plan, fmt.Errorf("%s has no trains to forecast (## Trains holds no `| Train | Cars, in order | Slot |` rows)", projectPath)
	}
	cal, err := loadForecastCalibration(brainDir)
	if err != nil {
		return projectdoc.TrainForecast{}, plan, err
	}
	issueModel := vocab.Issue()
	cars := map[projectdoc.IssueRef]projectdoc.CarEstimate{}
	for _, t := range plan.Trains {
		for _, c := range t.Cars {
			if _, done := cars[c]; done {
				continue
			}
			meta, lerr := projectIssueLookupFn(ctx, c.String(), identity.WorktreeRoot)
			if lerr != nil {
				cars[c] = projectdoc.CarEstimate{LookupErr: strings.TrimSpace(lerr.Error())}
				continue
			}
			cars[c] = projectdoc.CarEstimate{
				Estimate:    meta.EstimateHours,
				Terminal:    issueModel.IsTerminal(meta.Status),
				Unestimated: meta.EstimateHours <= 0,
			}
		}
	}
	fc, err := projectdoc.ForecastTrains(plan, cars, cal)
	return fc, plan, err
}

// projectCheckout resolves the workspace holding a project file, from its
// absolute path (a relative path would resolve to "." and fail every
// cross-repo lookup). Shared by the forecast and the mvp-landed lookup.
func projectCheckout(projectPath string) (workspace.Identity, error) {
	abs, err := filepath.Abs(projectPath)
	if err != nil {
		return workspace.Identity{}, err
	}
	return resolveWorkspace(filepath.Dir(abs))
}

// forecastLine renders the forecast for the read verbs (show/status),
// best-effort: a project with no trains prints nothing, any other failure a
// short unavailable line, and neither ever fails the read.
func forecastLine(ctx context.Context, path, brainDir string) string {
	d, err := readProject(path)
	if err != nil || len(d.Trains("").Trains) == 0 {
		return ""
	}
	fc, _, ferr := forecastForProject(ctx, d, path, brainDir)
	if ferr != nil {
		return "forecast: unavailable (" + ferr.Error() + ")"
	}
	return projectdoc.RenderTrainForecast(fc)
}
