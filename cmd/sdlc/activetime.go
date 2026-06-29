// activetime.go — `sdlc active-time`: the standalone CLI over the in-binary
// activetime engine (#110). It is the manual-inspection sibling of `sdlc actual`
// — same v3 attribution, but it prints the full per-segment table so a human can
// eyeball how a window was split. Replaces the old `python3 active-time-v3.py`
// invocation; the #68 loud-fail exit-code contract (2 misinvoke / 3 telemetry-gap
// / 0 measured-or-empty) is preserved.
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
)

// runActiveTime executes the computation, writes the table to out and
// diagnostics to errOut, and RETURNS the process exit code (0 measured/empty · 2
// misinvoke · 3 telemetry-gap). No os.Exit here, so it is unit-testable with
// bytes.Buffers; the cobra RunE wrapper calls os.Exit(runActiveTime(...)).
func runActiveTime(opts activetime.Options, out, errOut io.Writer) int {
	// Misinvoke guards (#68): without a transcript source or tracked issues the
	// result is a meaningless 0 — fail loudly with exit 2, never a silent 0.
	if len(opts.Dirs) == 0 {
		fmt.Fprintln(errOut, "error: no --dir given — active-time has no transcript source, so it would report 0 events.")
		fmt.Fprintln(errOut, "  Pass --dir ~/.claude/projects/<slug> or ~/.codex/sessions/YYYY/MM/DD for each source to inspect.")
		return 2
	}
	if len(opts.Issues) == 0 {
		fmt.Fprintln(errOut, "error: --issue required (at least one)")
		return 2
	}
	if opts.GitRepo == "" {
		fmt.Fprintln(errOut, "error: --git-repo required")
		return 2
	}

	res, err := activetime.Compute(opts)
	if err != nil {
		fmt.Fprintf(errOut, "error: %v\n", err)
		return 1
	}

	// Header (to stdout) — mirrors active-time-v3.py's preamble.
	fmt.Fprintln(out, "# v3 global-boundary attribution")
	if opts.PrefixWeight != nil && *opts.PrefixWeight != opts.CommitWeight {
		fmt.Fprintf(out, "# commit-weight: %g (prefix: %g)  •  threshold: %d min\n", opts.CommitWeight, *opts.PrefixWeight, opts.ThresholdMin)
	} else {
		fmt.Fprintf(out, "# commit-weight: %g  •  threshold: %d min\n", opts.CommitWeight, opts.ThresholdMin)
	}
	fmt.Fprintf(out, "# issues: %s\n", strings.Join(prefixHash(opts.Issues), ", "))
	fmt.Fprintf(out, "# events in window: %d  •  commits in window: %d\n\n", res.NumEvents, res.NumCommits)

	switch res.Status {
	case activetime.TelemetryGap:
		fmt.Fprintln(errOut, "# TELEMETRY UNAVAILABLE: window has commits but 0 transcript events.")
		fmt.Fprintln(errOut, "# The work's transcripts are likely in cwds not passed via --dir (peer repos / worktrees), or aged out.")
		fmt.Fprintln(errOut, "# Do NOT record 0 as measured — add the missing --dir folders, or use a labeled judgment estimate.")
		return 3
	case activetime.EmptyWindow:
		fmt.Fprintln(errOut, "# no events and no commits in window — nothing to measure")
		return 0
	}

	// Measured. Either a per-segment table (commits in window) or the
	// whole-window mention fallback (no commits).
	if len(res.Segments) == 0 {
		fmt.Fprintln(errOut, "# no commits in window — whole-window mention attribution")
	} else {
		renderSegmentTable(out, res.Segments)
		fmt.Fprintln(out)
		fmt.Fprintf(out, "# total active in window: %.1f min  (%.2f hr)\n", res.TotalActive, res.TotalActive/60)
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "# per-issue totals")
	for _, iss := range sortIssueKeys(res.PerIssue) {
		mins := res.PerIssue[iss]
		fmt.Fprintf(out, "  %s: %.2f hr  (%.1f min)\n", displayIssue(iss), mins/60, mins)
	}
	renderAttributionWarnings(out, res.Warnings)
	return 0
}

func renderAttributionWarnings(out io.Writer, warnings []activetime.AttributionWarning) {
	if len(warnings) == 0 {
		return
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "# attribution warnings")
	for _, w := range warnings {
		fmt.Fprintf(out, "  attribution warning: %s\n", formatAttributionWarning(w))
	}
}

func formatAttributionWarning(w activetime.AttributionWarning) string {
	return fmt.Sprintf("%s %.1fm/%.0f%% %s (%s → %s)",
		displayIssue(w.Issue), w.Active, w.Share*100, w.Reason, w.Start, w.End)
}

// renderSegmentTable prints the per-segment breakdown (one row per segment).
func renderSegmentTable(out io.Writer, segs []Segment) {
	fmt.Fprintf(out, "%3s  %-19s  %-19s  %5s  %-30s  %-11s  %-22s  %s\n",
		"#", "start", "end", "min", "commit", "issues", "mentions", "alloc")
	for n, sr := range segs {
		commitStr := "(no anchor)"
		issStr := ""
		if sr.Commit != nil {
			subj := sr.Commit.Subject
			if len(subj) > 30 {
				subj = subj[:30]
			}
			commitStr = sr.Commit.SHA + " " + subj
			issStr = strings.Join(prefixHash(sr.Commit.Issues), ",")
		}
		if sr.IsPrefix {
			commitStr = "[prefix] " + commitStr
		}
		fmt.Fprintf(out, "%3d  %-19s  %-19s  %5.1f  %-30s  %-11s  %-22s  %s\n",
			n+1,
			sr.Start.Local().Format("2006-01-02 15:04"),
			sr.End.Local().Format("2006-01-02 15:04"),
			sr.Active,
			commitStr,
			issStr,
			joinCounts(sr.Mentions, func(k string, v int) string { return fmt.Sprintf("%s=%d", displayIssue(k), v) }),
			joinCounts(sr.Alloc, func(k string, v float64) string { return fmt.Sprintf("%s=%.1fm", displayIssue(k), v) }),
		)
	}
}

// Segment is an alias to the engine's Segment so the renderer reads cleanly.
type Segment = activetime.Segment

// displayIssue renders a per-issue key: "#N" for a numeric issue, "unattributed"
// for the engine's sentinel bucket (the intentional cosmetic fix vs the Python's
// "##unattributed").
func displayIssue(key string) string {
	if key == activetime.UnattributedKey {
		return "unattributed"
	}
	return "#" + key
}

// sortIssueKeys orders per-issue keys: numeric issues ascending, the
// unattributed bucket (and any non-numeric key) last. Go map iteration is
// unordered, so output must be sorted explicitly.
func sortIssueKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ai, aerr := strconv.Atoi(keys[i])
		bj, berr := strconv.Atoi(keys[j])
		switch {
		case aerr == nil && berr == nil:
			return ai < bj
		case aerr == nil:
			return true // numeric before non-numeric
		case berr == nil:
			return false
		default:
			return keys[i] < keys[j]
		}
	})
	return keys
}

// joinCounts renders a sorted "k=v,k=v" string from a count map via fmt.
func joinCounts[V any](m map[string]V, fmtFn func(string, V) string) string {
	parts := make([]string, 0, len(m))
	for _, k := range sortIssueKeys(m) {
		parts = append(parts, fmtFn(k, m[k]))
	}
	return strings.Join(parts, ",")
}

// NewActiveTimeCmd returns the cobra command for `sdlc active-time`.
func NewActiveTimeCmd() *cobra.Command {
	var dirs, issues []string
	var gitRepo, since, until string
	var commitWeight, prefixWeight float64
	var thresholdMin int
	var includeAssistant bool

	cmd := &cobra.Command{
		Use:           "active-time",
		Short:         "Per-issue active-time attribution table (the v3 engine, standalone)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"active-time\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := activetime.Options{
				Dirs:             dirs,
				GitRepo:          gitRepo,
				SinceISO:         since,
				UntilISO:         until,
				Issues:           issues,
				CommitWeight:     commitWeight,
				ThresholdMin:     thresholdMin,
				IncludeAssistant: includeAssistant,
			}
			// PrefixWeight is set only when the flag was given, so an explicit 0
			// is honored (nil = fall back to commit-weight).
			if cmd.Flags().Changed("prefix-commit-weight") {
				opts.PrefixWeight = &prefixWeight
			}
			os.Exit(runActiveTime(opts, cmd.OutOrStdout(), cmd.ErrOrStderr()))
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&dirs, "dir", nil, "transcript directory (Claude cwd dir or Codex date dir); repeatable")
	cmd.Flags().StringVar(&gitRepo, "git-repo", "", "repo to read commits from (required)")
	cmd.Flags().StringVar(&since, "since", "", "ISO timestamp; events/commits before are skipped")
	cmd.Flags().StringVar(&until, "until", "", "ISO timestamp; events/commits after are skipped")
	cmd.Flags().StringArrayVar(&issues, "issue", nil, "issue number to track (without #); repeatable")
	cmd.Flags().Float64Var(&commitWeight, "commit-weight", 1.0, "fraction of a segment's active time attributed by commit refs")
	cmd.Flags().Float64Var(&prefixWeight, "prefix-commit-weight", 0, "commit-weight for the pre-first-commit prefix segment (defaults to --commit-weight)")
	cmd.Flags().IntVar(&thresholdMin, "threshold-min", 15, "gap-truncation threshold in minutes")
	cmd.Flags().BoolVar(&includeAssistant, "include-assistant", false, "include assistant messages in the active-time stream")
	return cmd
}
