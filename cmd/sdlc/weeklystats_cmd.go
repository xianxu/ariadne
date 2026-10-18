package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/pkg/workspace"
)

type weeklySnapshotFlags struct {
	End       string
	BrainRoot string
}

type weeklySnapshotDeps struct {
	now      func() time.Time
	identity func() (workspace.Identity, error)
	discover func(workspace.Identity) ([]string, error)
	read     func(context.Context, string, string) weeklyRepoEvidence
	publish  func(string, time.Time, []byte) (string, error)
}

func defaultWeeklySnapshotDeps() weeklySnapshotDeps {
	return weeklySnapshotDeps{
		now:      time.Now,
		identity: func() (workspace.Identity, error) { return workspace.Resolve(execGitRunner{}, ".", "") },
		discover: discoverWeeklyRepos,
		read:     readWeeklyEvidence,
		publish:  publishBrainSnapshot,
	}
}

func newWeeklySnapshotCmd() *cobra.Command {
	flags := weeklySnapshotFlags{}
	cmd := &cobra.Command{
		Use: "snapshot", Short: "publish an immutable seven-day test-economy snapshot in brain",
		Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runWeeklySnapshot(cmd.Context(), cmd.OutOrStdout(), flags, defaultWeeklySnapshotDeps())
		},
	}
	cmd.Flags().StringVar(&flags.End, "end", "", "UTC midnight ending the seven-day window (RFC3339; default latest completed Monday 00:00 UTC)")
	cmd.Flags().StringVar(&flags.BrainRoot, "brain-root", "", "brain capture repository root (default sibling brain)")
	return cmd
}

func weeklyDefaultEnd(now time.Time) time.Time {
	day := now.UTC().Truncate(24 * time.Hour)
	return day.AddDate(0, 0, -((int(day.Weekday()) + 6) % 7))
}

func runWeeklySnapshot(ctx context.Context, out io.Writer, flags weeklySnapshotFlags, deps weeklySnapshotDeps) error {
	if deps.now == nil || deps.identity == nil || deps.discover == nil || deps.read == nil || deps.publish == nil {
		return fmt.Errorf("test-stats snapshot: missing command dependency")
	}
	now := deps.now().UTC()
	end := weeklyDefaultEnd(now)
	if flags.End != "" {
		parsed, err := time.Parse(time.RFC3339, flags.End)
		if err != nil {
			return fmt.Errorf("test-stats snapshot: --end: %w", err)
		}
		end = parsed
	}
	window, err := newWeeklyWindow(end, now)
	if err != nil {
		return fmt.Errorf("test-stats snapshot: %w", err)
	}
	if window.End.Hour() != 0 || window.End.Minute() != 0 || window.End.Second() != 0 || window.End.Nanosecond() != 0 {
		return fmt.Errorf("test-stats snapshot: --end must be midnight UTC so its dated artifact is unambiguous")
	}
	id, err := deps.identity()
	if err != nil {
		return fmt.Errorf("test-stats snapshot: workspace: %w", err)
	}
	brainRoot := flags.BrainRoot
	if brainRoot == "" {
		brainRoot = filepath.Join(id.FleetRoot, "brain")
	}
	if _, err := os.Stat(filepath.Join(brainRoot, ".brain", "config.md")); err != nil {
		return fmt.Errorf("test-stats snapshot: %s is not a brain capture repository: %w", brainRoot, err)
	}
	paths, err := deps.discover(id)
	if err != nil {
		return fmt.Errorf("test-stats snapshot: discover fleet: %w", err)
	}
	sort.Strings(paths)
	snapshot := WeeklySnapshot{SchemaVersion: weeklySnapshotSchemaVersion, Window: window, Repos: []WeeklyRepoStats{}}
	for _, root := range paths {
		repo := filepath.Base(root)
		evidence := deps.read(ctx, repo, root)
		stats := aggregateWeeklyRepo(repo, evidence.Attempts, window)
		landings, invalid := validateLandingEvidence(evidence, window)
		unknown := append(append([]string{}, evidence.CoverageUnknown...), invalid...)
		attribution := attributeWeeklyAttempts(evidence.Attempts, landings, window, unknown)
		stats.Source = WeeklySource{MainRef: evidence.MainRef, TrackerRef: evidence.TrackerRef, AttemptRecords: len(evidence.Attempts), LandingEvents: len(evidence.Events), Warnings: evidence.Warnings}
		stats.LandedIssues = attribution.LandedIssues
		stats.AttributedAttempts = attribution.AttributedAttempts
		stats.AttributedPassed = attribution.AttributedPassed
		stats.AttributedFailed = attribution.AttributedFailed
		stats.AttributedRepeats = attribution.AttributedRepeats
		stats.OutsideWeekAttributed = attribution.OutsideWeekAttributed
		stats.InWeekOutsideCohort = attribution.InWeekOutsideCohort
		stats.PreEnqueue = attribution.PreEnqueue
		stats.UnassignableAttempts = attribution.UnassignableAttempts
		stats.UnknownAttempts = attribution.UnknownAttempts
		stats.CoverageUnknown = attribution.CoverageUnknown
		stats.AttemptsPerLandedIssue = attribution.AttemptsPerLandedIssue
		snapshot.Repos = append(snapshot.Repos, stats)
	}
	if len(snapshot.Repos) == 0 {
		return fmt.Errorf("test-stats snapshot: no SDLC repositories found in fleet")
	}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	path, err := deps.publish(brainRoot, window.End, append(data, '\n'))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, path)
	return err
}
