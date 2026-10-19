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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
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
	actualTelemetryGap                     // commits exist but 0 events — recover transcript evidence
	actualEmptyWindow                      // no commits/events to measure
	actualNoWindow                         // no commits reference the issue yet
	actualError                            // engine/IO error — recover required evidence
)

type actualResult struct {
	Status        actualStatus
	Hours         float64
	Issue         string
	Peers         []string
	Dirs          []string
	Window        string // "<start ISO> → HEAD": the resolved left edge (#254)
	Detail        string // diagnostic for the error path
	Warnings      []string
	QueueBlockers []string // complete queue refusal diagnostics, when present
	// Split divides Hours at the boundary reviews (#234); nil unless measured.
	Split *actualSplit
	// Inputs is the exact admitted evidence retained by close/correction.
	Inputs *gatestate.MeasurementInputs
	// Segments are transient replay evidence, never persisted in the close card.
	Segments      []activetime.Segment
	CountedPhases []activetime.Phase // exact phases that contributed to adopted Hours
	Exclusions    []activetime.Exclusion
}

// computeActual is the engine glue: resolve the commit window + peer issues + the
// brain/repo transcript dirs, run activetime.Compute, and classify the result.
// Runs git via the cwd (gitx.CommitWindow), so the caller should be inside
// repoTop.
func computeActual(ctx context.Context, repoTop, brainAbs, issueNum string) actualResult {
	return computeActualFor(ctx, repoTop, brainAbs, issueNum, "", "")
}

// computeActualFor measures the whole issue (milestone "") or one milestone's
// increment: the same window and attribution, kept to that boundary's phases
// (#234 E1). Hours is build + converge in both modes; the verification tail
// after each finalized boundary's last commit is not counted (#234 D2). plans
// is the plans dir holding the boundary gate ledger ("" for the default); a
// close passes its own --plans-dir, so the split reads the rounds it persists.
func computeActualFor(ctx context.Context, repoTop, brainAbs, issueNum, milestone, plans string) actualResult {
	return computeActualForPurpose(ctx, repoTop, brainAbs, issueNum, milestone, plans, actualHistorical)
}

func computeActualForPurpose(ctx context.Context, repoTop, brainAbs, issueNum, milestone, plans string, purpose actualInputPurpose) actualResult {
	owned, err := resolveOwnedActualInputsFor(ctx, repoTop, brainAbs, issueNum, plans, purpose, milestone)
	if err != nil {
		return actualErrorResult(issueNum, err)
	}
	if owned != nil {
		if !owned.Replay {
			protected, err := loadAdoptedExclusions(ctx, repoTop, issueNum, plans, owned)
			if err != nil {
				return actualErrorResult(issueNum, err)
			}
			owned.Protected = protected
		}
		return computeOwnedActualFor(ctx, repoTop, issueNum, milestone, owned)
	}
	result := computeLegacyActualFor(ctx, repoTop, brainAbs, issueNum, milestone, plans)
	result.Warnings = append(result.Warnings, "legacy no-tracker attribution heuristic: repository cwd plus brain; ownership provenance unavailable")
	return result
}

func actualErrorResult(issueNum string, err error) actualResult {
	res := actualResult{Status: actualError, Issue: issueNum, Detail: err.Error()}
	var queueErr *queueEvidenceError
	if errors.As(err, &queueErr) {
		res.QueueBlockers = append([]string(nil), queueErr.Blockers...)
	}
	return res
}

func computeLegacyActualFor(ctx context.Context, repoTop, brainAbs, issueNum, milestone, plans string) actualResult {
	res := actualResult{Issue: issueNum}
	identity, err := resolveWorkspace(repoTop)
	if err != nil {
		res.Status, res.Detail = actualError, err.Error()
		return res
	}
	repoTop = identity.WorktreeRoot

	// #252: the tracker's card commits (claim, close) are this issue's activity
	// too, and its card holds the authoritative `started` stamp.
	trackerRefs, cardStarted, carded, trackerWarning := actualTrackerInputs(ctx, repoTop, issueNum)
	if trackerWarning != "" {
		res.Warnings = append(res.Warnings, trackerWarning)
	}
	firstSHA, firstISO, lastISO, _ := gitx.CommitWindow(issueNum, trackerRefs...)
	if firstSHA == "" {
		res.Status = actualNoWindow
		return res
	}

	// #113/#254: the window starts at the claim when there is one, so design
	// attention after it lands in-window and nothing before it does.
	// Best-effort — a locate/parse miss just keeps the commit-based start.
	// Peers are derived AFTER the override, from the same window.
	if carded {
		firstISO = resolveWindowStart(firstISO, cardStarted, "")
	} else if id, err := strconv.Atoi(issueNum); err == nil {
		issuesDir := envOr("WF_ISSUES_DIR", "workshop/issues")
		if path, err := locateIssueFile(filepath.Join(repoTop, issuesDir), id); err == nil {
			wtISO, _ := gitx.WorkingTransitionISO(path)
			firstISO = resolveWindowStart(firstISO, startedAnchor(path), wtISO)
		}
	}

	// The label names the resolved start, not the first #N commit: since the
	// claim anchor (#254, via #270) that commit is often the filing, outside
	// the window.
	res.Window = firstISO + " → HEAD"

	// #270: on an issue branch, only the branch's own commits and commits naming
	// this issue bound segments, so integrating main can't hand this session's
	// time to the issues main's commits name. Peers come from the same commits.
	scope := actualScope(issueNum)
	peers, err := activetime.WindowIssues(repoTop, firstISO, lastISO, scope, trackerRefs...)
	if err != nil {
		res.Status, res.Detail = actualError, err.Error()
		return res
	}
	res.Peers = peers
	src, err := transcripts.Select(ctx, nonEmpty(brainAbs, repoTop), transcripts.DefaultHarnesses())
	if err != nil {
		res.Status, res.Detail = actualError, err.Error()
		return res
	}
	if err := ctx.Err(); err != nil {
		res.Status, res.Detail = actualError, err.Error()
		return res
	}
	res.Warnings = append(res.Warnings, src.Warnings...)
	res.Dirs = src.Dirs
	if len(src.Dirs) == 0 && len(src.Files) == 0 {
		res.Status, res.Detail = actualTelemetryGap, "no brain/repo transcript sources found across the harness registry (Claude ~/.claude/projects, Codex ~/.codex/sessions, Qoder ~/.qoder/projects)"
		return res
	}

	out, err := activetime.Compute(activetime.Options{
		Dirs:             src.Dirs,
		Files:            src.Files,
		GitRepo:          repoTop,
		ExtraRefs:        trackerRefs,
		Scope:            scope,
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
	if res.Status == actualMeasured {
		id, _ := strconv.Atoi(issueNum)
		ledger, _, lerr := issueBoundaryLedger(repoTop, plans, id)
		if lerr != nil {
			res.Status, res.Detail = actualError, "boundary gate ledger: "+lerr.Error()
			return res
		}
		split, serr := splitActual(out, issueNum, milestone, splitRounds(ledger), firstISO, lastISO)
		if serr != nil {
			res.Status, res.Detail = actualError, serr.Error()
			return res
		}
		res.Hours, res.Split = split.Build+split.Converge, &split
		if milestone != "" {
			res.Window += " (" + milestone + "'s phases)"
		}
	}
	for _, w := range out.Warnings {
		res.Warnings = append(res.Warnings, formatAttributionWarning(w))
	}
	return res
}

// actualScope is the boundary scope for measuring issueNum from the cwd's HEAD
// (#270): scoped at the branch point when HEAD has diverged from main, and also
// on the issue's own branch before its first commit, where the branch point is
// HEAD itself. On main (or any branch that hasn't diverged) it stays unscoped.
func actualScope(issueNum string) activetime.Scope {
	bp := gitx.MergeBaseWithMain()
	if bp == "" && onIssueBranch(issueNum) {
		bp = gitx.BranchPoint()
	}
	return activetime.Scope{BranchPoint: bp, Issue: issueNum}
}

// onIssueBranch reports whether HEAD is the issue's branch (`NNNNNN-slug`).
func onIssueBranch(issueNum string) bool {
	id, err := strconv.Atoi(issueNum)
	if err != nil {
		return false
	}
	return strings.HasPrefix(gitx.Capture("branch", "--show-current")+"-", fmt.Sprintf("%06d-", id))
}

// actualTrackerInputs reads the tracker for active-time (#252): the tracking ref
// whose history holds the claim/close commits, and the card's `started` stamp.
// carded is false without a tracker (or card), leaving the legacy anchors.
func actualTrackerInputs(ctx context.Context, repoTop, issueNum string) (refs []string, started string, carded bool, warning string) {
	id, err := strconv.Atoi(issueNum)
	if err != nil {
		return nil, "", false, ""
	}
	repo, err := recordsRepository(ctx, repoTop)
	if err != nil {
		return nil, "", false, "tracker unavailable, measured without its claim/close commits: " + err.Error()
	}
	if repo == nil {
		return nil, "", false, ""
	}
	rs, err := actualInputRecords(ctx, repo, repoTop, filepath.Join(repoTop, envOr("WF_ISSUES_DIR", "workshop/issues")))
	if err != nil {
		return nil, "", false, "tracker unreadable, measured without its claim/close commits: " + err.Error()
	}
	if !rs.Tracker {
		return nil, "", false, ""
	}
	if rs.Stale {
		warning = staleTrackerNote + "; measured against its last-fetched claim/close commits"
	}
	rec, ok, cerr := rs.Require(fmt.Sprintf("%06d", id))
	if cerr != nil {
		return []string{repo.TrackingRef()}, "", false, "card unreadable, measured without its started stamp: " + cerr.Error()
	}
	if !ok || rec.Card == nil {
		return []string{repo.TrackingRef()}, "", false, warning
	}
	s, _ := rec.Field("started")
	return []string{repo.TrackingRef()}, strings.TrimSpace(s), true, warning
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

// resolveWindowStart picks the active-time window's left edge: the explicit
// `started:` claim stamp (#116), else the WorkingTransitionISO claim heuristic
// (#113), else the commit-parent default. Pure — the IO (file read + git) stays
// in computeActual's glue (ARCH-PURE, per the #116 plan-quality review).
func resolveWindowStart(parentISO, startedISO, wtISO string) string {
	anchor := startedISO
	if anchor == "" {
		anchor = wtISO
	}
	return windowStart(parentISO, anchor)
}

// windowStart starts the window at the claim when there is one, else at
// parentISO (the parent of the first `#N` commit). The claim wins even when
// the parent is earlier (#254, folded into #270): the parent moves whenever
// main is integrated, and an issue filed days before its claim would otherwise
// absorb all the slot's work in between — which, with boundaries scoped to the
// branch, its own commits would claim. Design attention after the claim is
// still in-window, which is what #113 widened the start for.
func windowStart(parentISO, claimISO string) string {
	if claimISO != "" {
		return claimISO
	}
	return parentISO
}

// statusFromResult maps an activetime.Result to the actual outcome for issueNum.
// Pure — the integration contract, unit-testable without git/files. Mirrors the
// old classifyV3 exit-code mapping: telemetry-gap stays unmeasurable;
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

// printActualMeasurement preserves the measurement window, phase cuts and
// bounded provenance diagnostics wherever a successful result is presented.
func printActualMeasurement(w io.Writer, res actualResult) {
	cok(w, fmt.Sprintf("measured actual for #%s: %.2fh   (window %s)", res.Issue, res.Hours, res.Window))
	if s := res.Split; s != nil {
		fmt.Fprintf(w, "  build %.2fh · converge %.2fh · uncounted tail %.2fh (cuts: %s)\n", s.Build, s.Converge, s.Tail, s.Source)
	}
	for _, warning := range res.Warnings {
		fmt.Fprintf(w, "  attribution warning: %s\n", warning)
	}
}

// legacyQueueAttemptIDs returns IDs only when every queue blocker is a
// controlled legacy warning. A mixed window must retain recovery guidance.
func legacyQueueAttemptIDs(blockers []string) []string {
	if len(blockers) == 0 {
		return nil
	}
	ids := make([]string, 0, len(blockers))
	for _, blocker := range blockers {
		rest, ok := strings.CutPrefix(blocker, legacyQueueWarningPrefix)
		if !ok {
			return nil
		}
		id, suffix, ok := strings.Cut(rest, " ")
		if !ok || len(id) < 2 || len(id) > 64 || !suiteHex.MatchString(id) || !strings.HasPrefix(suffix, "lacks exact queue endpoints;") {
			return nil
		}
		ids = append(ids, id)
	}
	return ids
}

// printActual renders measured evidence or the action needed to recover it.
// The milestone selects the exact close verb for an actual preview or refusal.
func printActual(w io.Writer, res actualResult, milestone string) {
	issueStr := issue.CLIRef(res.Issue)
	switch res.Status {
	case actualMeasured:
		printActualMeasurement(w, res)
		fmt.Fprintf(w, "  → close (or milestone-close, for a --milestone measurement) ADOPTS this when --actual is omitted (#178, #234)\n")
		if len(res.Peers) > 1 {
			fmt.Fprintf(w, "  (attributed across window issues: %s)\n", strings.Join(prefixHash(res.Peers), ", "))
		}
	case actualTelemetryGap:
		cwarn(w, fmt.Sprintf("telemetry unavailable for #%s (window has commits but no transcript events).", issueStr))
		fmt.Fprintf(w, "  → ask the TL/hub to restore the owning slot's transcripts, then re-run sdlc actual --issue %s.\n", issueStr)
		if res.Detail != "" {
			fmt.Fprintf(w, "  (%s)\n", res.Detail)
		}
	case actualEmptyWindow:
		if len(res.Exclusions) > 0 {
			cwarn(w, fmt.Sprintf("found no uncredited activity for #%s; its parked claim overlapped already adopted time.", issueStr))
			for _, exclusion := range res.Exclusions {
				fmt.Fprintf(w, "  excluded %.2f minutes during %s → %s: already adopted by issue #%s in slot %s\n", exclusion.Minutes, exclusion.Start.Format(time.RFC3339Nano), exclusion.End.Format(time.RFC3339Nano), exclusion.Issue, exclusion.Slot)
			}
			fmt.Fprintf(w, "  → close with %s and record this exclusion.\n", rerunCmd(issueStr, milestone, " --no-actual"))
		} else {
			cwarn(w, fmt.Sprintf("found no measurable activity for #%s; ask the TL/hub to restore its ownership/source evidence, then re-run sdlc actual --issue %s.", issueStr, issueStr))
		}
	case actualNoWindow:
		cwarn(w, fmt.Sprintf("no commits reference #%s yet; commit the issue's work with #%s in the subject, then re-run sdlc actual --issue %s.", issueStr, issueStr, issueStr))
	case actualError:
		cwarn(w, fmt.Sprintf("can't auto-compute (%s).", res.Detail))
		if ids := legacyQueueAttemptIDs(res.QueueBlockers); len(ids) > 0 {
			subject := "legacy suite attempt " + ids[0] + " lacks"
			if len(ids) > 1 {
				subject = "legacy suite attempts " + strings.Join(ids, ", ") + " lack"
			}
			fmt.Fprintf(w, "  → %s exact queue endpoints; those timestamps were never recorded and cannot be restored. Exclude this actual from calibration with:\n    %s\n  Include the attempt ID(s) and reason with the verification evidence.\n", subject, rerunCmd(issueStr, milestone, " --no-actual"))
		} else if strings.Contains(res.Detail, "retained measurement code head differs") {
			fmt.Fprintf(w, "  → run sdlc issue correct-actual --issue %s --reason 'repair inconsistent retained measurement' --dry-run; it names the exact target/exclusion action if numeric repair is unavailable.\n", issueStr)
		} else {
			fmt.Fprintf(w, "  → ask the TL/hub to restore the named measurement evidence, then re-run sdlc actual --issue %s.\n", issueStr)
		}
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
	var brainDir, milestone string
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
			res := computeActualFor(cmd.Context(), repoTop, brainAbs, strconv.Itoa(issue), milestone, "")
			printActual(stderr, res, milestone)
			return nil
		},
	}
	cmd.Flags().IntVar(&issue, "issue", 0, "issue ID to measure (numeric, required)")
	cmd.Flags().StringVar(&brainDir, "brain-dir", "../brain", "path to the brain repo (transcript dir always included)")
	cmd.Flags().StringVar(&milestone, "milestone", "", "measure one milestone's increment, Mx (#234): what milestone-close adopts")
	return cmd
}
