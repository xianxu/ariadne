// actual.go — `sdlc actual`: compute an issue's focused dev-hours by running
// active-time-v3 with the RIGHT transcript dirs, so the number is measured, not
// hand-typed (#68 M2). Lifts the manual "run this 6-line python command"
// explainer prose into the binary — nobody ran the command, so every actual was
// a guess. The engine (computeActual) is shared: `sdlc actual` exposes it, and
// close's missing-`--actual` explainer calls it to print a suggestion inline.
//
// Dir-selection is the crux (#68 diagnosis): events come only from transcript
// `.jsonl` folders, and the work spans multiple cwds. The validated heuristic is
// **brain + the issue's own repo** — it matched human-recorded numbers within
// ~5%, while throwing in an unrelated concurrently-edited repo inflated them. So
// we feed exactly those two, never "all folders."
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

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

// selectActualDirs returns the transcript dirs to feed v3: brain (always) + the
// repo itself, keeping only folders that actually exist under transcriptsRoot.
// Deliberately NOT every folder — unrelated repos with concurrent activity
// inflate the count (#68).
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

// v3PrimaryHoursRE-style parse: pull issueNum's hours from v3's
// "  #<N>: <h> hr  (<m> min)" per-issue total line on stdout.
func parseV3PrimaryHours(stdout, issueNum string) (float64, bool) {
	re := regexp.MustCompile(`(?m)^\s*#` + regexp.QuoteMeta(issueNum) + `:\s+([0-9]+(?:\.[0-9]+)?)\s+hr\b`)
	m := re.FindStringSubmatch(stdout)
	if m == nil {
		return 0, false
	}
	h, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false
	}
	return h, true
}

// actualStatus is the outcome of an actual computation.
type actualStatus int

const (
	actualMeasured     actualStatus = iota // v3 produced a number
	actualTelemetryGap                     // commits exist but 0 events (v3 exit 3) — judgment
	actualEmptyWindow                      // no commits/events to measure
	actualNoWindow                         // no commits reference the issue yet
	actualNoScript                         // active-time-v3.py / python3 unavailable — fall back
)

type actualResult struct {
	Status actualStatus
	Hours  float64
	Issue  string
	Peers  []string
	Dirs   []string
	Window string // "<shortSHA> → HEAD"
	Detail string // diagnostic for the fall-back/error paths
}

// v3Runner runs active-time-v3.py and returns (stdout, exitCode). Package var
// so tests can stub it. exitCode is -1 when the process couldn't be started.
var v3Runner = func(scriptPath string, args []string) (string, int, error) {
	cmd := exec.Command("python3", append([]string{scriptPath}, args...)...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = io.Discard // v3's diagnostics go to stderr; we key off exit code
	err := cmd.Run()
	if err == nil {
		return out.String(), 0, nil
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return out.String(), ee.ExitCode(), nil
	}
	return out.String(), -1, err // python3 missing, etc.
}

// computeActual is the engine: resolve the commit window + peer issues + the
// brain/repo transcript dirs, run v3, and classify the result. Runs git via the
// cwd (gitx.CommitWindow), so the caller should be inside repoTop.
func computeActual(repoTop, brainAbs, issueNum string) actualResult {
	res := actualResult{Issue: issueNum}

	script := filepath.Join(repoTop, "construct", "local", "issues", "active-time-v3.py")
	if _, err := os.Stat(script); err != nil {
		res.Status, res.Detail = actualNoScript, "active-time-v3.py not found under construct/local/issues/"
		return res
	}
	if _, err := exec.LookPath("python3"); err != nil {
		res.Status, res.Detail = actualNoScript, "python3 not on PATH"
		return res
	}

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

	args := make([]string, 0, len(res.Dirs)*2+len(res.Peers)*2+10)
	for _, d := range res.Dirs {
		args = append(args, "--dir", d)
	}
	args = append(args, "--git-repo", repoTop, "--since", firstISO, "--until", lastISO)
	for _, p := range res.Peers {
		args = append(args, "--issue", p)
	}
	args = append(args, "--commit-weight", "1.0", "--threshold-min", "15", "--include-assistant")

	stdout, code, err := v3Runner(script, args)
	if err != nil {
		res.Status, res.Detail = actualNoScript, err.Error()
		return res
	}
	res.Status, res.Hours = classifyV3(code, stdout, issueNum)
	if res.Status == actualNoScript {
		res.Detail = fmt.Sprintf("active-time-v3 exited %d", code)
	}
	return res
}

// classifyV3 maps active-time-v3's (exit code, stdout) to an outcome for
// issueNum. Pure — split out from computeActual's subprocess machinery so the
// exit-code contract (the integration surface M2 depends on) is unit-testable
// without git or python. 3=telemetry-gap, 0+parseable=measured, 0+unparseable=
// empty window, anything else (incl. 2 misinvoke, unreachable here) → fall back.
func classifyV3(exitCode int, stdout, issueNum string) (actualStatus, float64) {
	switch exitCode {
	case 3: // TELEMETRY UNAVAILABLE (commits but 0 events)
		return actualTelemetryGap, 0
	case 0:
		if h, ok := parseV3PrimaryHours(stdout, issueNum); ok {
			return actualMeasured, h
		}
		return actualEmptyWindow, 0
	default: // 2 (misinvoke — shouldn't happen, we always pass --dir) or unexpected
		return actualNoScript, 0
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
		cwarn(w, fmt.Sprintf("v3 found no measurable activity for #%s — use a labeled judgment estimate.", res.Issue))
	case actualNoWindow:
		cwarn(w, fmt.Sprintf("no commits reference #%s yet — commit first, or use a judgment estimate.", res.Issue))
	case actualNoScript:
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
		Short:         "Compute an issue's focused dev-hours (runs active-time-v3 with the right transcript dirs)",
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
