// actual.go — `sdlc actual`: compute an issue's focused dev-hours so the number
// is MEASURED, not hand-typed (#68 M2). The engine is the native Go
// `internal/activetime` package (#110): computeActual calls activetime.Compute
// in-process — no python3 subprocess, no stdout-regex, no script resolution. The
// engine is shared: `sdlc actual` exposes it, and close's missing-`--actual`
// explainer calls it to print a suggestion inline.
//
// Dir-selection is the crux (#68 diagnosis): events come only from transcript
// `.jsonl` folders, and the work spans multiple cwds. The validated heuristic is
// **brain + the issue's own repo** — it matched human-recorded numbers within
// ~5%, while throwing in an unrelated concurrently-edited repo inflated them. So
// we feed exactly those two, never "all folders."
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
)

// nonEmpty returns the non-empty arguments in order — the brain+repo cwds fed to
// the transcript-harness registry (a "" brain/repo path is dropped, not queried).
func nonEmpty(ss ...string) []string {
	var out []string
	for _, s := range ss {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// Transcript-source selection lives in internal/transcripts (#134): a Harness
// per agent CLI (Claude, Codex), a Registry, and a pure Select aggregator over
// brain+repo cwds. computeActual feeds transcripts.Select into the engine; adding
// a harness never touches this file.

// actualStatus is the outcome of an actual computation.
type actualStatus int

const (
	actualMeasured     actualStatus = iota // engine produced a number
	actualTelemetryGap                     // commits exist but 0 events — judgment
	actualEmptyWindow                      // no commits/events to measure
	actualNoWindow                         // no commits reference the issue yet
	actualError                            // engine/IO error — fall back to judgment
)

type actualResult struct {
	Status   actualStatus
	Hours    float64
	Issue    string
	Peers    []string
	Dirs     []string
	Window   string // "<shortSHA> → HEAD"
	Detail   string // diagnostic for the error path
	Warnings []string
}

// computeActual is the engine glue: resolve the commit window + peer issues + the
// brain/repo transcript dirs, run activetime.Compute, and classify the result.
// Runs git via the cwd (gitx.CommitWindow), so the caller should be inside
// repoTop.
func computeActual(repoTop, brainAbs, issueNum string) actualResult {
	res := actualResult{Issue: issueNum}

	firstSHA, firstISO, lastISO, _ := gitx.CommitWindow(issueNum)
	if firstSHA == "" {
		res.Status = actualNoWindow
		return res
	}
	res.Window = firstSHA[:8] + " → HEAD"

	// #113: pull the window-start back to the claim (working-transition) commit
	// when it's earlier than the parent-of-first-#N anchor, so DESIGN attention
	// after the claim (brainstorm / spec / plan / reviews) lands in-window.
	// Best-effort — a locate/parse miss just keeps the commit-based start.
	// Widening the start also widens DiscoverWindowIssues' peer membership (a
	// deliberate attribution change, not just this issue's minutes), so Peers is
	// derived AFTER the override.
	if id, err := strconv.Atoi(issueNum); err == nil {
		issuesDir := envOr("WF_ISSUES_DIR", "workshop/issues")
		if path, err := locateIssueFile(filepath.Join(repoTop, issuesDir), id); err == nil {
			wtISO, _ := gitx.WorkingTransitionISO(path)
			firstISO = resolveWindowStart(firstISO, startedAnchor(path), wtISO)
		}
	}

	res.Peers, _ = gitx.DiscoverWindowIssues(firstISO, lastISO, issueNum)
	src := transcripts.Select(nonEmpty(brainAbs, repoTop), transcripts.DefaultHarnesses())
	res.Dirs = src.Dirs
	if len(src.Dirs) == 0 && len(src.Files) == 0 {
		res.Status, res.Detail = actualTelemetryGap, "no brain/repo transcript sources found across the harness registry (Claude ~/.claude/projects, Codex ~/.codex/sessions)"
		return res
	}

	out, err := activetime.Compute(activetime.Options{
		Dirs:             src.Dirs,
		Files:            src.Files,
		GitRepo:          repoTop,
		SinceISO:         firstISO,
		UntilISO:         lastISO,
		Issues:           res.Peers,
		CommitWeight:     1.0,
		ThresholdMin:     15,
		IncludeAssistant: true,
	})
	if err != nil {
		res.Status, res.Detail = actualError, err.Error()
		return res
	}
	res.Status, res.Hours = statusFromResult(out, issueNum)
	for _, w := range out.Warnings {
		res.Warnings = append(res.Warnings, formatAttributionWarning(w))
	}
	return res
}

// startedAnchor reads the explicit `started:` engagement stamp (#116) from an
// issue file's frontmatter — the robust window anchor that supersedes the
// WorkingTransitionISO git-log heuristic. Returns "" if absent/unreadable (the
// legacy fallback path then takes over). The one IO boundary; the picking logic
// is the pure resolveWindowStart.
func startedAnchor(path string) string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	fm, _, err := issue.Parse(string(raw))
	if err != nil {
		return ""
	}
	v, _ := issue.GetField(fm, "started")
	return strings.TrimSpace(v)
}

// resolveWindowStart picks the active-time window's left edge from three anchor
// candidates in robustness order: the explicit `started:` stamp (#116) when
// present, else the WorkingTransitionISO claim heuristic (#113), both delegated
// to windowStart against the commit-parent default. Pure — the IO (file read +
// git) stays in computeActual's glue, so both anchor paths are unit-testable
// without fakes (ARCH-PURE, per the #116 plan-quality review).
//
// Caveat (inherited from windowStart): the ISO strings are compared lexically, so
// they must share a UTC offset to sort chronologically. started: (local RFC3339)
// and the git %aI anchors are same-machine/same-offset in the realistic case; a
// DST boundary inside a long issue could skew the compare by the offset delta, but
// gap-truncation bounds the blast radius to minutes.
func resolveWindowStart(parentISO, startedISO, wtISO string) string {
	anchor := startedISO
	if anchor == "" {
		anchor = wtISO
	}
	return windowStart(parentISO, anchor)
}

// windowStart picks the active-time window's left edge from the two candidate
// anchors: parentISO (parent-of-first-#N-commit, CommitWindow's default) and
// wtISO (the claim's working-transition commit, #113). Returns the EARLIER of
// the two non-empty ISOs — claim-early ⇒ wt is earlier ⇒ design attention is
// captured; a late claim ⇒ parent is earlier ⇒ no regression; either empty ⇒
// the other. Pure (ISO-8601 strings sort chronologically given a stable offset,
// matching CommitWindow's own WindowCapDays compare).
func windowStart(parentISO, wtISO string) string {
	switch {
	case wtISO == "":
		return parentISO
	case parentISO == "":
		return wtISO
	case wtISO < parentISO:
		return wtISO
	default:
		return parentISO
	}
}

// statusFromResult maps an activetime.Result to the actual outcome for issueNum.
// Pure — the integration contract, unit-testable without git/files. Mirrors the
// old classifyV3 exit-code mapping: telemetry-gap stays a judgment signal;
// measured-with-the-issue yields hours; everything else is an empty window.
func statusFromResult(out activetime.Result, issueNum string) (actualStatus, float64) {
	switch out.Status {
	case activetime.TelemetryGap:
		return actualTelemetryGap, 0
	case activetime.Measured:
		if mins, ok := out.PerIssue[issueNum]; ok {
			return actualMeasured, mins / 60
		}
		return actualEmptyWindow, 0
	default: // EmptyWindow (or, defensively, an unset status)
		return actualEmptyWindow, 0
	}
}

// printActual renders the engine's result for a human/agent: the suggested
// --actual value, or the judgment-fallback guidance.
func printActual(w io.Writer, res actualResult) {
	switch res.Status {
	case actualMeasured:
		cok(w, fmt.Sprintf("measured actual for #%s: %.2fh   (window %s)", res.Issue, res.Hours, res.Window))
		for _, warning := range res.Warnings {
			fmt.Fprintf(w, "  attribution warning: %s\n", warning)
		}
		fmt.Fprintf(w, "  → issue close ADOPTS this when --actual is omitted (#178); milestone close: pass the per-milestone increment\n")
		if len(res.Peers) > 1 {
			fmt.Fprintf(w, "  (attributed across window issues: %s)\n", strings.Join(prefixHash(res.Peers), ", "))
		}
	case actualTelemetryGap:
		cwarn(w, fmt.Sprintf("telemetry unavailable for #%s (window has commits but no transcript events).", res.Issue))
		fmt.Fprintln(w, "  The work's transcripts aren't under brain/repo (peer cwds / worktrees) or aged out.")
		fmt.Fprintln(w, "  → record a LABELED judgment estimate: --actual <h> (note 'judgment') or --no-actual.")
		if res.Detail != "" {
			fmt.Fprintf(w, "  (%s)\n", res.Detail)
		}
	case actualEmptyWindow:
		cwarn(w, fmt.Sprintf("found no measurable activity for #%s — use a labeled judgment estimate.", res.Issue))
	case actualNoWindow:
		cwarn(w, fmt.Sprintf("no commits reference #%s yet — commit first, or use a judgment estimate.", res.Issue))
	case actualError:
		cwarn(w, fmt.Sprintf("can't auto-compute (%s) — fall back to a judgment estimate.", res.Detail))
	}
}

func prefixHash(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = "#" + s
	}
	return out
}

// NewActualCmd returns the cobra command for `sdlc actual`.
func NewActualCmd() *cobra.Command {
	var issue int
	var brainDir string
	cmd := &cobra.Command{
		Use:           "actual",
		Short:         "Compute an issue's focused dev-hours (runs the activetime engine with the right transcript dirs)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"actual\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			stderr := cmd.ErrOrStderr()
			if issue <= 0 {
				die(stderr, fmt.Sprintf("--issue is required and must be positive (got %d)", issue))
			}
			repoTop, err := gitx.RepoTopLevel()
			if err != nil {
				die(stderr, err.Error())
			}
			brainAbs, _ := filepath.Abs(brainDir)
			res := computeActual(repoTop, brainAbs, strconv.Itoa(issue))
			printActual(stderr, res)
			return nil
		},
	}
	cmd.Flags().IntVar(&issue, "issue", 0, "issue ID to measure (numeric, required)")
	cmd.Flags().StringVar(&brainDir, "brain-dir", "../brain", "path to the brain repo (transcript dir always included)")
	return cmd
}
