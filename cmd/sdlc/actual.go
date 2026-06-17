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
)

// transcriptsRoot is ~/.claude/projects (Claude Code's per-cwd transcript store).
// A package var so tests can point it at a fixture layout.
var transcriptsRoot = defaultTranscriptsRoot()

func defaultTranscriptsRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "projects")
}

// pathEncoder mirrors Claude Code's cwd→folder encoding: every '/' and '.'
// becomes '-'. e.g. /Users/x/workspace/nous → -Users-x-workspace-nous.
var pathEncoder = strings.NewReplacer("/", "-", ".", "-")

func cwdToTranscriptDir(absPath string) string { return pathEncoder.Replace(absPath) }

// selectActualDirs returns the transcript dirs to feed the engine: brain (always)
// + the repo itself, keeping only folders that actually exist under
// transcriptsRoot. Deliberately NOT every folder — unrelated repos with
// concurrent activity inflate the count (#68).
func selectActualDirs(repoTop, brainAbs string) []string {
	var out []string
	seen := map[string]bool{}
	for _, p := range []string{brainAbs, repoTop} {
		if p == "" {
			continue
		}
		dir := filepath.Join(transcriptsRoot, cwdToTranscriptDir(p))
		if seen[dir] {
			continue
		}
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			out = append(out, dir)
			seen[dir] = true
		}
	}
	return out
}

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
	Status actualStatus
	Hours  float64
	Issue  string
	Peers  []string
	Dirs   []string
	Window string // "<shortSHA> → HEAD"
	Detail string // diagnostic for the error path
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
	res.Peers, _ = gitx.DiscoverWindowIssues(firstISO, lastISO, issueNum)
	res.Dirs = selectActualDirs(repoTop, brainAbs)
	if len(res.Dirs) == 0 {
		res.Status, res.Detail = actualTelemetryGap, "no brain/repo transcript dirs found under "+transcriptsRoot
		return res
	}

	out, err := activetime.Compute(activetime.Options{
		Dirs:             res.Dirs,
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
	return res
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
		fmt.Fprintf(w, "  → close with:  --actual %.2f\n", res.Hours)
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
