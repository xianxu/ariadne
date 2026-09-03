// close.go — `sdlc close` subcommand. Ports scripts/close-issue.py.
//
// Same posture as the Python source:
//   - Validates inputs (ISSUE required; ACTUAL + VERIFIED required unless
//     --no-actual / --no-verified / --force).
//   - Emits the semantic warmup once per shell session (#69).
//   - On a standalone full-issue close, auto-dispatches the one binary-owned
//     boundary review on the whole-issue window (#69); milestone-close reviews
//     each milestone slice instead. See runCloseWithReview.
//   - Locates the issue file under workshop/issues/.
//   - Checks atlas/ was touched in the issue's commit window (bypassable with
//     --no-atlas / --force).
//   - Each guard has a per-gate --no-<gate> bypass; --force waives all of them
//     at once (#67). A bypass logs an audit line; the rationale goes in --verified.
//   - Mutates the issue file (milestone tick OR status flip + log line).
//   - Mutates the matching brain-side project file (task row tick + detail-block field upsert).
//   - --dry-run prints what would change and exits 0.
//
// Semantics preserved byte-for-byte where it matters. Deviations noted in
// inline comments.
package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// Plan-section regexes moved to internal/issue/plan.go so cmd/sdlc/state
// can share the source of truth (M2 review I5). Per-issue-tag patterns
// like the milestone-tick regex stay per-call because they interpolate
// f.Milestone.

// closeFlags holds the parsed flag values for the close subcommand.
type closeFlags struct {
	Issue     int
	Milestone string
	Actual    string
	Verified  string
	Force     bool
	DryRun    bool
	BrainDir  string
	IssuesDir string
	// PlansDir holds durable plans and the gate sidecars beside them. Added in #187 as a
	// FIELD rather than a third inline envOr: the boundary review already resolved it twice
	// inline, and the cost report needs it a third time to find the plan-gate ledger
	// (ARCH-DRY).
	PlansDir      string
	Agent         string // agent CLI for the issue boundary-review dispatch (#69)
	AgentExplicit bool
	Mode          string // optional supervision mode (supervised|delegated) for the calibration ledger (#117)

	// Per-gate bypass flags (#67). Each waives exactly ONE of the close
	// guards (checked in computeClose); --force waives them all. The flag is an explicit acknowledgment
	// that the gate doesn't apply here (the rationale belongs in --verified) —
	// not a way to forget it. See skip().
	NoActual    bool
	NoVerified  bool
	NoReclose   bool
	NoAtlas     bool
	NoVerdict   bool
	NoLedger    bool
	NoPlanCheck bool
	NoProject   bool
	NoJudge     bool // skip the issue boundary review (#69)
}

// resolvePlansDir is the SINGLE resolution of the durable-plans directory: the explicit
// value when set, else WF_PLANS_DIR, else the convention.
//
// The fallback is not belt-and-braces. closeFlags is constructed directly in production —
// milestoneCloseFlags.closeFlags() translates one struct into the other — so a field that
// only cobra's flag default populates is empty on the milestone-close path. An empty plans
// dir resolves to the REPO ROOT, which fails silently in the worst way: every issue reads
// as "no plan-gate ledger" (zeroing the metrics this feature exists to produce) and review
// sidecars land beside the Makefile.
func resolvePlansDir(explicit string) string {
	if explicit != "" {
		return explicit
	}
	return envOr("WF_PLANS_DIR", "workshop/plans")
}

// plansDir resolves this close's durable-plans directory. See resolvePlansDir.
func (f *closeFlags) plansDir() string { return resolvePlansDir(f.PlansDir) }

// skip reports whether the named gate should be bypassed — either the blanket
// --force or that gate's specific --no-<gate> flag (#67). Single source of
// truth so every gate site reads `!f.skip("<gate>")` uniformly.
func (f *closeFlags) skip(gate string) bool {
	if f.Force {
		return true
	}
	switch gate {
	case "actual":
		return f.NoActual
	case "verified":
		return f.NoVerified
	case "reclose":
		return f.NoReclose
	case "atlas":
		return f.NoAtlas
	case "verdict":
		return f.NoVerdict
	case "ledger":
		return f.NoLedger
	case "plan":
		return f.NoPlanCheck
	case "project":
		return f.NoProject
	case "judge":
		return f.NoJudge
	}
	return false
}

// NewCloseCmd returns the cobra command for `sdlc close`. The main session
// is responsible for registering it on the root command and for supplying
// the rich Long text via the embedded helptext package.
func NewCloseCmd() *cobra.Command {
	var f closeFlags

	cmd := markManualLockCommand(&cobra.Command{
		Use:   "close",
		Short: "Close an issue or milestone (records ACTUAL + VERIFIED, mutates issue + project files)",
		Long: "Performs AGENTS.md §5's mechanical closing steps for an issue or " +
			"milestone: enforces ACTUAL + VERIFIED, checks atlas/ was touched in " +
			"the commit window, ticks the issue's ## Plan, flips status, appends a " +
			"verification log line, and updates the matching brain-side project " +
			"file (task tick + detail-block field upsert). Does NOT commit — the " +
			"agent commits, usually bundling the close with other work. The main " +
			"session's helptext package replaces this Long text with the full " +
			"checkpoint contract once wired up.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			guardSpineRepo(cmd.ErrOrStderr()) // #176 lifecycle guard
			f.AgentExplicit = cmd.Flags().Changed("agent")
			return runCloseWithReviewLocked(cmd, cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})

	cmd.Flags().IntVar(&f.Issue, "issue", 0, "issue ID (numeric, required)")
	cmd.Flags().StringVar(&f.Milestone, "milestone", "", "DEPRECATED (#146) — use `sdlc milestone-close`; passing this to `close` refuses")
	_ = cmd.Flags().MarkHidden("milestone") // #146: milestone closing moved fully to `sdlc milestone-close`
	cmd.Flags().StringVar(&f.Actual, "actual", "", "focused dev-hours (measured + adopted when omitted at issue close; see `sdlc actual`)")
	cmd.Flags().StringVar(&f.Mode, "mode", "", "optional supervision mode for the calibration ledger: supervised | delegated (#117)")
	cmd.Flags().StringVar(&f.Verified, "verified", "", "one-line evidence the work meets done-when")
	cmd.Flags().BoolVar(&f.Force, "force", false, "bypass ALL gates (≡ every --no-* flag); record the reason in --verified")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print what would change; do not write")
	cmd.Flags().StringVar(&f.BrainDir, "brain-dir", "../brain", "path to the brain repo (for the calibration ledger; project files are now discovered across the fleet, #171)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", "workshop/issues", "directory holding issue files")
	cmd.Flags().StringVar(&f.PlansDir, "plans-dir", envOr("WF_PLANS_DIR", "workshop/plans"), "directory holding durable plans + gate sidecars (the plan-gate ledger the cost report reads, #187)")
	// Per-gate bypasses (#67) — each waives one guard; --force waives all.
	cmd.Flags().BoolVar(&f.NoActual, "no-actual", false, "record actual_hours: N/A and skip velocity calibration")
	cmd.Flags().BoolVar(&f.NoVerified, "no-verified", false, "bypass the VERIFIED-evidence requirement (only if there's no behavior to verify)")
	cmd.Flags().BoolVar(&f.NoReclose, "no-reclose-guard", false, "bypass the already-done refusal (intentionally re-close)")
	cmd.Flags().BoolVar(&f.NoAtlas, "no-atlas", false, "bypass the atlas/ change check (acknowledge: no new architectural surface)")
	cmd.Flags().BoolVar(&f.NoLedger, "no-ledger", false, "bypass the boundary gate ledger's open-findings refusal (#194)")
	cmd.Flags().BoolVar(&f.NoVerdict, "no-verdict", false, "bypass the milestone Review-Verdict trailer check")
	cmd.Flags().BoolVar(&f.NoPlanCheck, "no-plan-check", false, "bypass the unchecked-## Plan-items refusal")
	cmd.Flags().BoolVar(&f.NoProject, "no-project", false, "bypass the project detail-block update requirement")
	cmd.Flags().BoolVar(&f.NoJudge, "no-judge", false, "skip the issue boundary review auto-dispatched on full-issue close (#69)")
	cmd.Flags().StringVar(&f.Agent, "agent", "", "agent CLI for the boundary-review dispatch (claude | codex | gemini)")
	// Don't use MarkFlagRequired("issue"): cobra emits an uncolored,
	// differently-formatted error that conflicts with die()'s red prefix.
	// Validation lives in computeClose so all error formatting flows through
	// one path. SilenceErrors keeps cobra from printing on top of us.
	cmd.SilenceErrors = true
	return cmd
}

// Terminal helpers (cinfo / cok / cwarn / die) + ANSI constants live in
// term.go alongside the other package-shared helpers. Moved out of
// close.go in M4 to deduplicate across the now-7 subcommands.

// ── warmup ───────────────────────────────────────────────────────────────────

// warmupThreshold = 1: deliver the close-issue contract explainer ONCE per shell
// session, not twice. Re-delivering didn't help — an agent that read it the first
// time doesn't re-read it (the same "agents don't reread a static doc" premise
// behind #75's per-session re-injection); the second copy was just noise (#69).
const warmupThreshold = 1

func warmupStatePath() string {
	// Process group ID is stable across subshells of the same controlling
	// shell and resets on new shell / new Claude Code session. Matches
	// close-issue.py's os.getpgrp() identity.
	pgid := syscall.Getpgrp()
	if pgid < 0 {
		pgid = 0
	}
	// Hardcoded /tmp to match close-issue.py exactly. macOS's per-user
	// $TMPDIR would isolate Go/Python state, masking the warmup-count
	// during a transition period where both binaries co-exist. /tmp is
	// world-writable on every Unix; if it's not, the WriteFile in
	// warmupIncrement swallows the error silently (best-effort).
	return filepath.Join("/tmp", fmt.Sprintf("close-issue-warmup-%d", pgid))
}

func warmupCount() int {
	data, err := os.ReadFile(warmupStatePath())
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	return n
}

func warmupIncrement() {
	// Best-effort. /tmp may not be writable in some sandboxes; close-issue.py
	// swallows OSError here, we swallow likewise.
	_ = os.WriteFile(warmupStatePath(), []byte(strconv.Itoa(warmupCount()+1)), 0o644)
}

func printSemanticWarmup(w io.Writer) {
	n := warmupCount()
	if n >= warmupThreshold {
		return
	}
	lines := []string{
		fmt.Sprintf("%s── close-issue contract ── (warmup %d/%d)%s", ansiCyan, n+1, warmupThreshold, ansiReset),
		"",
		"  Closing an issue records two values that feed into velocity",
		"  calibration. Both must be earned, not guessed:",
		"",
		fmt.Sprintf("  %sACTUAL%s   = focused dev-hours on this issue (not wall-clock).", ansiCyan, ansiReset),
		"             sdlc measures it — omit --actual and close ADOPTS the measured",
		"             value (#178), or run `sdlc actual --issue N` to preview.",
		"             (method: 42shots/velocity/baseline-v3.md)",
		"             Pass --no-actual (or --force) only if there's genuinely nothing",
		"             to measure; close records actual_hours: N/A and skips calibration.",
		"",
		fmt.Sprintf("  %sVERIFIED%s = one-line evidence of behavior matching done-when.", ansiCyan, ansiReset),
		"             'tests pass' beats 'code written'. See AGENTS.md §5.",
		"",
		fmt.Sprintf("  This warmup auto-suppresses after %d invocations per shell session.", warmupThreshold),
		"",
	}
	fmt.Fprintln(w, strings.Join(lines, "\n"))
	warmupIncrement()
}

// logLineDateRE pulls the leading ISO date off a log line (`- YYYY-MM-DD: …`)
// so insertLogLine can file it under a matching `### YYYY-MM-DD` day header.
var logLineDateRE = regexp.MustCompile(`^- (\d{4}-\d{2}-\d{2}):`)

// insertLogLine inserts logLine into the `## Log` section.
//
// Placement:
//
//   - If logLine carries a leading date (`- YYYY-MM-DD: …`) and the Log
//     section already has a matching `### YYYY-MM-DD` day header, the line is
//     filed directly *beneath* that header (top of the day's group) — so a
//     dated close line sits inside its day rather than orphaned above the
//     header (#66). Top-of-group, not bottom, preserves the newest-first
//     convention insertLogLine already uses at the section level.
//
//   - Otherwise it goes at the top of the `## Log` section, mirroring
//     close-issue.py's one-shot:
//
//     re.sub(r"(^## Log\s*\n)(\s*\n)?", rf"\1\n{log_line}\n", body, count=1, MULTILINE)
//
//     Behavior preserved byte-for-byte from Python: `\s*\n` is greedy and
//     includes newlines, so group 1 consumes "## Log\n" plus any trailing
//     blank line(s) up to the next non-blank. The output is `<group1>\n<log>\n`
//     followed by whatever came after the match — so a blank line after
//     `## Log` yields one extra blank line (Python emits
//     "## Log\n\n\n<log>\n- existing\n"; surprising, but it's what the source does).
//
// If `## Log` is absent, we append a new section at the bottom of body.
//
// Anchor: the real `## Log` section, located by the fence-aware scanner (#211).
// Both the heading AND the section's end are resolved that way, and both halves
// matter: anchoring the heading alone still let the `### <date>` search run to
// EOF, so a quoted day header in a LATER section would capture the insert.
//
// This replaced a last-match heuristic added by #66 — first-match had filed a
// close line into #66's own fenced Problem-section example. Last-match fixed
// that case and only that case: it breaks when a quoted `## Log` sits after the
// real one. Both are the same defect FenceSpans now answers properly.
func insertLogLine(body, logLine string) string {
	// #211 M2: located by the fence-aware section scanner. The last-match
	// heuristic this replaces was added by #66 for exactly one reason — a
	// first-match version filed the close line into #66's own fenced example —
	// which is the same defect FenceSpans now solves properly. Last-match was
	// also only accidentally right: it fails when a quoted `## Log` sits AFTER
	// the real one, and 1 of 406 corpus files already has that shape.
	logStart, ok := issue.SectionHeadingByteOffset(body, "Log", issue.UnterminatedIsProse)
	if !ok {
		return strings.TrimRight(body, "\n\r\t ") + "\n\n## Log\n\n" + logLine + "\n"
	}
	// Bound the search to the real Log section. Taking body[logStart:] would let
	// the `### <date>` lookup below run past the section's end into a LATER
	// section's fenced example — the same class of bug one level down.
	_, logEnd, _ := issue.SectionByteBounds(body, "Log", issue.UnterminatedIsProse)
	section := body[logStart:logEnd]

	// Prefer the matching `### <date>` day header within the real Log section.
	// The match anchors on the date *prefix* and allows an optional ` — suffix`
	// (`([ \t].*)?$`), because the log convention routinely suffixes day headers
	// (`### 2026-05-30 — session summary`); a bare-date-only matcher orphaned the
	// line at the top of ## Log (#73). Still bounded to one line: `.` doesn't
	// match newline in RE2 (no `(?s)`) and `$` under `(?m)` sits before `\n`, so
	// the insert offset stays at the header line's end and no blank lines are
	// eaten the way `\s*$` would. The required `[ \t]` separator rejects
	// `### <date>x` (a date can't prefix-match a longer token).
	if m := logLineDateRE.FindStringSubmatch(logLine); m != nil {
		// Line-wise and fence-aware: bounding to the Log section is not enough,
		// because the section can quote its OWN format (#211 M2 review BR-11).
		dayRE := regexp.MustCompile(`^### ` + regexp.QuoteMeta(m[1]) + `([ \t].*)?$`)
		if _, dEnd, found := issue.FindLineOutsideFences(section, dayRE); found {
			pos := logStart + dEnd // end of the day-header line text (before its newline)
			return body[:pos] + "\n" + logLine + body[pos:]
		}
	}
	// Fallback: top of the real `## Log` section. Same shape as close-issue.py's
	// regex, but run on `section` so it anchors to the fence-aware header; for
	// the common single-`## Log` body this is byte-for-byte identical.
	insertRE := regexp.MustCompile(`(?m)(^## Log\s*\n)(\s*\n)?`)
	loc := insertRE.FindStringSubmatchIndex(section)
	if loc == nil {
		// The section was located but insertRE didn't match — shouldn't happen
		// in practice (the patterns are equivalent up to trailing content),
		// but fall through to append-mode rather than panic.
		return strings.TrimRight(body, "\n\r\t ") + "\n\n## Log\n\n" + logLine + "\n"
	}
	// loc[*] are relative to `section`; offset by logStart. group1 = loc[2..3].
	group1 := section[loc[2]:loc[3]]
	return body[:logStart+loc[0]] + group1 + "\n" + logLine + "\n" + body[logStart+loc[1]:]
}

// ── main entry point ─────────────────────────────────────────────────────────

// projectEdit is one project-file mutation the close performs. The close gate
// discovers every project across the fleet that references the closing issue
// (multiple matches are legitimate membership — #171), so a close can carry
// several. repoDir is retained for M3's safe peer-write commit decision.
type projectEdit struct {
	path    string // absolute path to the project .md
	repoDir string // absolute repo root that owns it
	oldText string // pre-edit content (TOCTOU guard + review snapshot)
	newText string // post-edit content to write
}

// closeResult bundles everything applyClose needs, computed by computeClose
// WITHOUT any writes — so the boundary review can run against the un-mutated
// working tree and the writes fire only after a finalizing verdict (#139).
type closeResult struct {
	issuePath    string
	issueText    string // original, for the "changed?" guard
	newIssueText string
	projectEdits []projectEdit
	repoTop      string // closing repo's top-level — the peer-write "current repo" anchor (M3)
	// calibration-ledger inputs (read from the ORIGINAL issue):
	fm, body, repoName, issueStr, today string
	// success messages that describe WRITES — emitted by applyClose (post-finalize),
	// so a REWORK never prints "flipped → codecomplete" for a write that didn't happen.
	appliedMsgs []string
}

// computeClose runs every close gate and composes the new issue/project text in
// memory, returning a closeResult — it writes NOTHING. Gate failures still die() /
// exitWithCode(1) fast, before any review (#139: extracted from runClose).
func computeClose(stderr io.Writer, f *closeFlags) closeResult {
	printSemanticWarmup(stderr)
	var applied []string

	if f.Issue <= 0 {
		die(stderr, fmt.Sprintf("--issue is required and must be positive (got %d)", f.Issue))
	}
	issueStr := strconv.Itoa(f.Issue)
	mode := "issue"
	if f.Milestone != "" {
		mode = "milestone"
	}
	// #178: the omit-path MEASURES AND ADOPTS. The gate's purpose is preventing
	// GUESSED hours; a value sdlc measured itself can't be a guess, so the old
	// compute-then-ask refusal ("→ close with: --actual N", copied verbatim
	// ~45/48 times — the spine's second-largest refusal volume, #172) is now an
	// adoption with a loud info line. Unmeasurable statuses keep the refusal.
	adopted := false
	if f.Actual == "" && !f.skip("actual") {
		if res, ok := adoptOmittedActual(stderr, f, issueStr, mode); ok {
			adopted = true
		} else {
			explainActual(stderr, issueStr, mode, f.Milestone, res)
			exitWithCode(1)
		}
	}

	if f.Actual != "" {
		v, err := strconv.ParseFloat(f.Actual, 64)
		if err != nil {
			die(stderr, fmt.Sprintf("ACTUAL must be a number, got '%s'", f.Actual))
		}
		// #87: sanity-check a PASSED --actual against the active-time-v3
		// measurement — the pass-path used to trust the value blindly, which let
		// a fabricated 13.5 (measured 0.30) pollute velocity calibration. A
		// wildly-off value is refused; a moderately-off one warns. --force/
		// --no-actual bypasses (rationale in --verified). Skips silently when
		// the engine can't measure — and for an ADOPTED value (#178): comparing
		// the measurement against itself would just re-run the engine.
		if !adopted && !f.skip("actual") {
			if derr := checkActualDeviation(stderr, issueStr, v, mode); derr != nil {
				die(stderr, derr.Error())
			}
		}
	}

	if f.Actual == "" {
		// Reachable only via --no-actual/--force (the omit-path above either
		// adopted or exited): genuinely nothing to measure.
		cwarn(stderr, fmt.Sprintf("--no-actual (or --force): closing with actual_hours: %s — velocity calibration skipped", issue.ActualNotApplicableSentinel))
	}
	if f.Verified == "" {
		if !f.skip("verified") {
			explainVerified(stderr, issueStr, mode, f.Milestone, f.Actual)
			exitWithCode(1)
		}
		cwarn(stderr, "--no-verified (or --force): closing with NO verification evidence — no behavior recorded as checked")
	}

	today := time.Now().Format("2006-01-02")

	// ── Locate issue file ───────────────────────────────────────────────────
	issuePath, err := issueFilePath(f.IssuesDir, f.Issue)
	if err != nil {
		die(stderr, err.Error())
	}
	issueBytes, err := os.ReadFile(issuePath)
	if err != nil {
		die(stderr, fmt.Sprintf("read %s: %v", issuePath, err))
	}
	issueText := string(issueBytes)

	repoTop, err := gitx.RepoTopLevel()
	if err != nil {
		die(stderr, err.Error())
	}
	repoName := filepath.Base(repoTop)

	fm, body, err := issue.Parse(issueText)
	if err != nil {
		die(stderr, fmt.Sprintf("no YAML frontmatter in %s", issuePath))
	}

	// #122 carve-out: re-close guard keys on "done" specifically (the verified-complete
	// state), not IsTerminal — re-closing a done issue is the case to guard.
	if currentStatus, _ := issue.GetField(fm, "status"); mode == "issue" && currentStatus == "done" {
		if !f.skip("reclose") {
			// #174: append-only after the pinned span — gatesig's RefusalPat and
			// the frozen codex golden fixtures both key on the head text.
			die(stderr, fmt.Sprintf("%s#%s is already status: done — pass --no-reclose-guard (or --force) to re-close intentionally.\n"+
				"  Post-publish follow-up work is a new issue (a `side-quest:` commit or `sdlc issue new`), not a re-close.", repoName, issueStr))
		}
		cwarn(stderr, fmt.Sprintf("--no-reclose-guard (or --force): re-closing %s#%s (already done)", repoName, issueStr))
	}

	// ── Commit window + atlas check ─────────────────────────────────────────
	// One window source (ARCH-DRY, #58): boundaryWindowBase gives the same base
	// the boundary review uses — the prior review boundary for a milestone close,
	// the branch start otherwise — so the atlas-coverage check and the review
	// cover exactly the same commits, including inter-milestone side-quests/fixes.
	windowBase := boundaryWindowBase(issueStr, f.Milestone, issuePath)
	// #194 M1 review: pin the window's upper end the same way the review does, so
	// "the same commits" is structural rather than a consequence of both running
	// under one lock. Falls back to the literal when rev-parse cannot answer.
	windowHead := "HEAD"
	if sha := gitx.Capture("rev-parse", "HEAD"); sha != "" {
		windowHead = sha
	}
	if windowBase != "" {
		cinfo(stderr, fmt.Sprintf("commit window: %s → %s", shortSHA(windowBase), abbrevSHA(windowHead)))
	} else {
		cwarn(stderr, fmt.Sprintf("no commits reference '#%s' on this branch", issueStr))
	}

	if windowBase != "" {
		diffFiles, derr := gitx.DiffNames(windowBase, windowHead)
		if derr != nil {
			// #177 review Important #1: a swallowed diff error used to inherit the
			// refusal (fail-closed); with the auto-satisfy arm, nil files would
			// fail OPEN ("0 doc files — auto-satisfied"). A broken window diff is
			// a hard stop, consistent with the other git failures in this path.
			die(stderr, fmt.Sprintf("window diff %s..HEAD failed: %v", shortSHA(windowBase), derr))
		}
		var atlasChanged, nonAtlas []string
		for _, p := range diffFiles {
			if strings.HasPrefix(p, "atlas/") {
				atlasChanged = append(atlasChanged, p)
			} else {
				nonAtlas = append(nonAtlas, p)
			}
		}
		if len(atlasChanged) == 0 {
			switch {
			case !hasCodePath(nonAtlas):
				// #177: no code surface in the window — nothing to map. Demanding
				// an atlas delta (or a --no-atlas acknowledgment) on a docs/
				// workshop-only close is incoherent; auto-satisfy loudly. Also
				// covers the empty window (a bookkeeping-only re-close).
				cinfo(stderr, atlasAutoSatisfyLine(len(nonAtlas)))
			case !f.skip("atlas"):
				explainNoAtlas(stderr, shortSHA(windowBase), nonAtlas)
				exitWithCode(1)
			default:
				cwarn(stderr, "--no-atlas (or --force): skipping atlas/ change check — rationale in --verified")
			}
		}
	}

	// ── Milestone-review verdict check (issue close only) ──────────────────
	//
	// Every milestone in the plan must carry a Review-Verdict: trailer on
	// its close commit (AGENTS.md §3 fresh-eyes review evidence). The
	// check is bypassable with --force; the rationale belongs in --verified.
	if mode == "issue" {
		ordered, missing, err := findMilestonesMissingVerdict(body, issueStr, issuePath)
		if err != nil {
			cwarn(stderr, fmt.Sprintf("milestone-verdict check skipped: %v", err))
		} else if len(missing) > 0 {
			midstream, trailing := partitionMissingVerdicts(ordered, missing)
			switch {
			case f.skip("verdict"):
				cwarn(stderr, fmt.Sprintf("--no-verdict (or --force): skipping Review-Verdict check for %d milestone(s): %s",
					len(missing), strings.Join(missing, ", ")))
			case len(midstream) > 0:
				explainMissingVerdicts(stderr, issueStr, midstream)
				exitWithCode(1)
			case f.skip("judge"):
				// Trailing-only, but --no-judge skips the very review that
				// would cover them — the #175 acceptance premise is gone.
				die(stderr, formatTrailingNeedsJudge(issueStr, trailing))
				exitWithCode(1)
			default:
				cinfo(stderr, formatTrailingVerdictAccepted(trailing))
			}
		}
	}

	// ── Edit issue file ─────────────────────────────────────────────────────
	newFM, newBody := fm, body

	if mode == "milestone" {
		// Scoped to the real Plan section and skipping fenced lines (#211 M2
		// review BR-12). This used to ReplaceAll over the WHOLE body, so a
		// `- [ ] M1` inside a quoted example anywhere in the issue was ticked —
		// the write-side twin of the read-side bug this issue fixes. An earlier
		// comment claimed the writer "rewrites a line it already matched"; it
		// does not, it rewrites every matching row in the document.
		pat := regexp.MustCompile(`(?m)^(- )\[[ .]\]( ` + regexp.QuoteMeta(f.Milestone) + `\b)`)
		n := 0
		if start, end, ok := issue.SectionByteBounds(newBody, "Plan", issue.UnterminatedIsProse); ok {
			planSrc := newBody[start:end]
			lines := strings.Split(planSrc, "\n")
			inside := issue.FenceSpans(lines, issue.UnterminatedIsProse)
			for i, line := range lines {
				if inside[i] {
					continue
				}
				if pat.MatchString(line) {
					lines[i] = pat.ReplaceAllString(line, "${1}[x]${2}")
					n++
				}
			}
			if n > 0 {
				newBody = newBody[:start] + strings.Join(lines, "\n") + newBody[end:]
			}
		}
		if n > 0 {
			applied = append(applied, fmt.Sprintf("ticked %s in %s ## Plan", f.Milestone, filepath.Base(issuePath)))
		} else {
			cwarn(stderr, fmt.Sprintf("no '- [ ] %s' in %s (project-tracked issue?)", f.Milestone, filepath.Base(issuePath)))
		}
	} else { // issue close
		if planBody, ok := issue.PlanItemsBody(newBody); ok {
			unchecked := issue.PlanUncheckedRE.FindAllString(planBody, -1)
			if len(unchecked) > 0 {
				if !f.skip("plan") {
					die(stderr, fmt.Sprintf(
						"%s ## Plan has %d unchecked item(s):\n  %s\n  (pass --no-plan-check, or --force, to close anyway)",
						filepath.Base(issuePath), len(unchecked), strings.Join(unchecked, "\n  ")))
				}
				cwarn(stderr, fmt.Sprintf("--no-plan-check (or --force): closing %s with %d unchecked ## Plan item(s)",
					filepath.Base(issuePath), len(unchecked)))
			}
		}
		// #160: close is the LOCAL acceptance gate — it flips to `codecomplete`, NOT
		// `done`. `merge`/`push` (the deterministic publish gate) flip codecomplete→done
		// after the reviewed-HEAD-unchanged invariant. close is the SOLE writer of
		// codecomplete (set-status refuses it), which is what makes the commit carrying
		// it a trustworthy anchor for that invariant. (#122 carve-out: value-specific
		// write, a literal like claim's "working" — not a category test.)
		newFM = issue.SetField(newFM, "status", "codecomplete")
		if f.Actual != "" {
			newFM = issue.SetField(newFM, "actual_hours", f.Actual)
		} else if f.skip("actual") {
			newFM = issue.SetField(newFM, "actual_hours", issue.ActualNotApplicableSentinel)
		}
		newFM = issue.SetField(newFM, "updated", today)
		msg := fmt.Sprintf("flipped %s → status: codecomplete", filepath.Base(issuePath))
		if f.Actual != "" {
			msg += fmt.Sprintf(", actual_hours: %s", f.Actual)
		} else if f.skip("actual") {
			msg += fmt.Sprintf(", actual_hours: %s", issue.ActualNotApplicableSentinel)
		}
		applied = append(applied, msg)
	}

	if f.Verified != "" {
		logLine := fmt.Sprintf("- %s: closed", today)
		if f.Milestone != "" {
			logLine += " " + f.Milestone
		}
		logLine += " — " + f.Verified
		newBody = insertLogLine(newBody, logLine)
		applied = append(applied, "appended verification line to ## Log")
	}

	newIssueText := issue.Compose(newFM, newBody)

	// ── Locate + edit project file(s) across the fleet ───────────────────────
	// Every project referencing this issue is updated (multiple matches are
	// legitimate membership, not ambiguity — #171). ActiveOnly so an archived
	// `done` project is never re-ticked; the peer-write commit decision is M3.
	var projectEdits []projectEdit

	matches, derr := project.DiscoverByIssueRef(filepath.Dir(repoTop), repoName, issueStr, project.ActiveOnly)
	if derr != nil {
		cwarn(stderr, derr.Error()+" — skipping project update")
	} else if len(matches) == 0 {
		cwarn(stderr, fmt.Sprintf("no project across the fleet references %s#%s — skipping project update", repoName, issueStr))
	}
	for _, m := range matches {
		// Label messages by repo/file so same-named projects in different repos
		// are distinguishable under fleet-wide multi-match (#171 M2 review).
		label := m.Repo + "/" + filepath.Base(m.Path)
		if m.Legacy {
			cwarn(stderr, fmt.Sprintf("project %s is in the deprecated brain/data/project home — migrate it to <repo>/workshop/projects (ariadne#171)", label))
		}
		projBytes, rerr := os.ReadFile(m.Path)
		if rerr != nil {
			die(stderr, fmt.Sprintf("read %s: %v", m.Path, rerr))
		}
		pt := string(projBytes)
		newPT := pt

		if mode == "milestone" {
			tickedPT, n := project.TickMilestoneTaskRow(newPT, repoName, issueStr, f.Milestone)
			newPT = tickedPT
			if n > 0 {
				applied = append(applied, fmt.Sprintf("ticked [%s#%s %s] in %s", repoName, issueStr, f.Milestone, label))
			} else {
				cwarn(stderr, fmt.Sprintf("no task line for [%s#%s %s] in %s", repoName, issueStr, f.Milestone, label))
			}

			anchor := project.AnchorFor(repoName, issueStr, f.Milestone)
			// Order matches close-issue.py: fm_set('actual') then fm_set('closed').
			// Slice (not map) so iteration order is deterministic.
			var fields []project.Field
			if f.Actual != "" {
				fields = append(fields, project.Field{Name: "actual", Value: f.Actual + "h"})
			}
			fields = append(fields, project.Field{Name: "closed", Value: today})
			updated, found := project.UpsertDetailBlockFields(newPT, anchor, fields)
			if !found {
				if !f.skip("project") {
					title := project.FindTaskTitle(newPT, repoName, issueStr, f.Milestone)
					est, _ := issue.GetField(fm, "estimate_hours")
					refLabel := fmt.Sprintf("%s#%s %s", repoName, issueStr, f.Milestone)
					actualOut := f.Actual + "h"
					skel, refDef := project.Skeleton{
						Anchor:    anchor,
						RefLabel:  refLabel,
						Title:     title,
						Est:       est,
						Actual:    actualOut,
						ClosedISO: today,
					}.Render()
					die(stderr, fmt.Sprintf(
						"no detail block <a id=\"%s\"> in %s (§5 step 4).\n"+
							"  Author one before closing — the prose paragraph is load-bearing\n"+
							"  for future calibration. Insert this skeleton inside ## details:\n\n"+
							"%s\n"+
							"  And add this reference definition at the file bottom:\n"+
							"    %s\n\n"+
							"  Then re-run. (--no-project, or --force, if it's a track-only milestone with nothing worth recording.)",
						anchor, label, skel, refDef))
				}
				cwarn(stderr, fmt.Sprintf("--no-project (or --force): skipping detail-block update for <a id=\"%s\"> in %s", anchor, label))
			}
			if found {
				newPT = updated
				applied = append(applied, fmt.Sprintf("updated detail block <a id=\"%s\"> in %s", anchor, label))
			}
		} else { // issue close
			tickedPT, n := project.TickAllTaskRowsForIssue(newPT, repoName, issueStr)
			newPT = tickedPT
			if n > 0 {
				applied = append(applied, fmt.Sprintf("ticked %d remaining task line(s) for %s#%s in %s", n, repoName, issueStr, label))
			}
			if n > 1 {
				cwarn(stderr, fmt.Sprintf("multiple %s#%s task rows ticked at once — confirm individual milestones were genuinely closed (§5 step 1)", repoName, issueStr))
			}
		}

		if newPT != pt {
			projectEdits = append(projectEdits, projectEdit{path: m.Path, repoDir: m.RepoDir, oldText: pt, newText: newPT})
		}
		if shouldNudgeProjectRetro(newPT, today, f.skip("project")) {
			cwarn(stderr, fmt.Sprintf("project retro is absent or older than 7 days in %s — consider `sdlc project retro`", label))
		}
	}

	return closeResult{
		issuePath:    issuePath,
		issueText:    issueText,
		newIssueText: newIssueText,
		projectEdits: projectEdits,
		repoTop:      repoTop,
		fm:           fm,
		body:         body,
		repoName:     repoName,
		issueStr:     issueStr,
		today:        today,
		appliedMsgs:  applied,
	}
}

func shouldNudgeProjectRetro(text, today string, skip bool) bool {
	if skip {
		return false
	}
	d, err := project.ParseDoc(text)
	if err != nil {
		return false
	}
	metadata, err := d.Metadata()
	if err != nil || !vocab.Project().IsExecuting(metadata.Status) {
		return false
	}
	return project.RetroStale(d, today, 7)
}

// printCloseDryRun prints what a close WOULD change, writing nothing (#139).
func printCloseDryRun(stderr io.Writer, r closeResult) {
	cinfo(stderr, "DRY=1 — no files written")
	fmt.Fprintf(os.Stdout, "Would update: %s\n", r.issuePath)
	for _, e := range r.projectEdits {
		fmt.Fprintf(os.Stdout, "Would update: %s\n", e.path)
	}
}

// closeRunner is the close path's gitRunner — the seam M3's peer-write commit
// shells through. Production is execGitRunner; tests may substitute.
var closeRunner gitRunner = execGitRunner{}

// applyClose performs the close's writes — issue + project files + the #117
// calibration ledger — then emits the success messages computeClose deferred
// (so "flipped → codecomplete" prints only when the flip actually happened). Called
// only after a finalizing verdict, or on the eager non-review path (#139).
// Peer-repo project edits then go through the safe peer-write commit decision
// (#171 M3): scoped commit when the peer is on main with a clean index,
// report-only otherwise — never failing the close either way.
func applyClose(stdout, stderr io.Writer, r gitRunner, f *closeFlags, res closeResult) {
	// ── Peer-write state snapshot (#171 M3) — BEFORE any file writes ─────────
	// The current repo's project edit rides the close commit; each PEER repo's
	// edit is committed there only when git state makes it unambiguous. The
	// snapshot must precede the writes below so pre-existing dirt on the target
	// files (another session's uncommitted edits) flips the peer to report-only
	// instead of being silently absorbed into the scoped commit.
	peerEdits := map[string][]string{}
	for _, e := range res.projectEdits {
		if e.repoDir == "" || e.repoDir == res.repoTop {
			continue
		}
		rel, rerr := filepath.Rel(e.repoDir, e.path)
		if rerr != nil {
			rel = e.path
		}
		peerEdits[e.repoDir] = append(peerEdits[e.repoDir], rel)
	}
	states := map[string]RepoGitState{}
	for repoDir, files := range peerEdits {
		states[repoDir] = readRepoGitState(r, repoDir, files)
	}

	if res.newIssueText != res.issueText {
		if err := os.WriteFile(res.issuePath, []byte(res.newIssueText), 0o644); err != nil {
			die(stderr, fmt.Sprintf("write %s: %v", res.issuePath, err))
		}
	}
	for _, e := range res.projectEdits {
		if err := os.WriteFile(e.path, []byte(e.newText), 0o644); err != nil {
			die(stderr, fmt.Sprintf("write %s: %v", e.path, err))
		}
	}
	for _, m := range res.appliedMsgs {
		cok(stderr, m)
	}

	if len(peerEdits) > 0 {
		decisions := planPeerWrites(peerEdits, states, res.repoTop, res.repoName+"#"+res.issueStr)
		applyPeerWrites(r, decisions, stdout, stderr)
	}

	// ── The cost report (#187 D1–D3) ─────────────────────────────────────────
	// UNCONDITIONAL, and that is the contract worth stating: unlike the ledger row
	// below — gated by #117's calibration-integrity rule, and skipped again inside
	// appendCalibrationRow when no brain/ledger dir resolves — this is diagnostic
	// output the operator always gets. Putting it inside appendCalibrationRow would
	// mean no cost report on a milestone close, under --no-actual, or in any
	// downstream repo without a sibling brain/, against a Done-when that says plainly
	// that `sdlc close` prints it.
	m := closeMetrics(stderr, f, res)
	cok(stderr, m.ChurnLine())
	cok(stderr, m.GateLine())

	// ── Close the loop (#117 mechanism 3) ────────────────────────────────────
	// On a full-issue close with a measured actual, append the estimate↔actual
	// data point to the calibration ledger. Milestone closes carry a partial
	// actual, so only the whole-issue close yields a clean row. The metrics are
	// PASSED IN rather than recomputed (ARCH-DRY) — two measurements of the same
	// window that could disagree would make the ledger unauditable against the
	// line the operator just read.
	if shouldLogCalibration(f) {
		appendCalibrationRow(stderr, f, res.fm, res.body, res.repoName, res.issueStr, res.today, m)
	}
	cok(stderr, "done — review with `git diff`, then commit")
}

// runClose is the eager compute→dry-run-or-apply wrapper. Since #139 the two
// production close paths finalize AFTER their review (compute → review → apply),
// so neither calls runClose: runCloseWithReview (whole-issue) and runMilestoneClose
// both drive computeClose/applyClose directly. #146 then removed the last
// production caller (the `sdlc close --milestone` short-circuit — runCloseWithReview
// now refuses it). runClose survives as the test-only convenience that bundles the
// mechanical close without a review (close_test.go / close_ledger_test.go).
func runClose(stdout, stderr io.Writer, f *closeFlags) error {
	r := computeClose(stderr, f)
	if f.DryRun {
		printCloseDryRun(stderr, r)
		return nil
	}
	applyClose(stdout, stderr, closeRunner, f, r)
	return nil
}

// shouldLogCalibration reports whether this close should append a calibration
// row. Only a FULL-issue close (not a milestone, which carries a partial actual)
// with a measured actual yields a clean estimate↔actual data point. This is the
// ledger's core integrity contract (#117): milestone closes must NOT pollute the
// calibration ledger with partial-actual rows.
func shouldLogCalibration(f *closeFlags) bool {
	return f.Milestone == "" && f.Actual != "" && !issue.IsActualNotApplicable(f.Actual)
}

// appendCalibrationRow writes one estimate↔actual data point to the calibration
// ledger at full-issue close (#117 mechanism 3). It is the IO seam — the row math
// (FormatRow / DriftVerdict / ParseRows) stays pure in internal/estimate.
//
// Graceful degradation (M2 plan-quality finding #1): `sdlc` is base-layer and
// propagates to downstream repos that may have no sibling brain/. When
// WF_CALIB_LEDGER is unset AND the resolved ledger dir is absent, it skips with a
// warning and returns — a missing ledger must NEVER break `sdlc close`.
func appendCalibrationRow(stderr io.Writer, f *closeFlags, fm, body, repoName, issueStr, today string, m closeCostMetrics) {
	ledgerPath := os.Getenv("WF_CALIB_LEDGER")
	usingOverride := ledgerPath != ""
	if !usingOverride {
		if f.BrainDir == "" {
			cwarn(stderr, "calibration ledger skipped (no brain dir resolved)")
			return
		}
		ledgerPath = estimate.VelocityPath(f.BrainDir, "calibration-ledger.tsv")
	}
	ledgerDir := filepath.Dir(ledgerPath)
	if _, err := os.Stat(ledgerDir); err != nil {
		if !usingOverride {
			cwarn(stderr, fmt.Sprintf("calibration ledger skipped (no ledger dir %s)", ledgerDir))
			return
		}
		if mkErr := os.MkdirAll(ledgerDir, 0o755); mkErr != nil {
			cwarn(stderr, fmt.Sprintf("calibration ledger skipped (mkdir %s: %v)", ledgerDir, mkErr))
			return
		}
	}

	estStr, _ := issue.GetField(fm, "estimate_hours")
	est, _ := strconv.ParseFloat(strings.TrimSpace(estStr), 64)
	actual, _ := strconv.ParseFloat(strings.TrimSpace(f.Actual), 64)

	var design, impl float64
	model := ""
	if section, ok := issue.EstimateSection(body); ok {
		if block, err := estimate.ParseBlock(section); err == nil {
			model = block.Model
			for _, it := range block.Items {
				design += it.Design
				impl += it.Impl
			}
		}
	}
	started, _ := issue.GetField(fm, "started")
	row := estimate.LedgerRow{
		Issue:     repoName + "#" + issueStr,
		Estimate:  est,
		EstDesign: design,
		EstImpl:   impl,
		Actual:    actual,
		Model:     model,
		Mode:      f.Mode,
		// #116 stamps `started:` at claim; until then every row is untrusted-window
		// and excluded from drift stats. Auto-upgrades once started: is present.
		WindowTrusted: strings.TrimSpace(started) != "",
		Date:          today,

		// #187 D1–D3: the cost columns, so "which gates earn their cost" is a query
		// over history rather than a recollection.
		ChurnProd:     m.Churn.Final.CodeProd,
		ChurnTest:     m.Churn.Final.CodeTest,
		ChurnAtlas:    m.Churn.Final.Atlas,
		ChurnWorkshop: m.Churn.Final.Workshop,
		Rework:        m.Churn.Rework,
		GateRounds:    m.Rounds,
		GateForced:    m.Forced,
		GateAddressed: m.Addressed,
		GateWithdrawn: m.Withdrawn,
		GateOpen:      m.Open,
	}

	existing, rerr := os.ReadFile(ledgerPath)
	if rerr != nil && !os.IsNotExist(rerr) {
		// Exists but transiently unreadable — do NOT clobber prior rows.
		cwarn(stderr, fmt.Sprintf("calibration ledger unreadable (%v) — not overwriting", rerr))
		return
	}
	var buf strings.Builder
	if len(existing) == 0 {
		buf.WriteString(estimate.Header() + "\n")
	} else {
		// Upgrade a legacy header before appending (#187). The column APPEND protects the
		// reader, but the header is written only on creation — so without this, every
		// pre-existing ledger would carry 20-column rows under a 10-column header: the
		// churn and gate data present but unlabeled, which for a file read by humans and
		// scripts by column name is worse than absent.
		text := string(existing)
		if upgraded, changed := estimate.UpgradeHeader(text); changed {
			text = upgraded
			cok(stderr, "calibration ledger: header upgraded to the #187 column set")
		}
		buf.WriteString(text)
		if !strings.HasSuffix(text, "\n") {
			buf.WriteString("\n")
		}
	}
	buf.WriteString(estimate.FormatRow(row) + "\n")
	if err := os.WriteFile(ledgerPath, []byte(buf.String()), 0o644); err != nil {
		cwarn(stderr, fmt.Sprintf("calibration ledger append failed: %v", err))
		return
	}

	trust := "untrusted-window"
	if row.WindowTrusted {
		trust = "trusted-window"
	}
	cok(stderr, fmt.Sprintf("calibration ledger: %s est %.2f / actual %.2f (ratio %.1f×, %s)",
		row.Issue, row.Estimate, row.Actual, row.Ratio(), trust))

	if warn, msg := estimate.DriftVerdict(estimate.ParseRows(buf.String()), 5); warn {
		cwarn(stderr, msg)
	}
}

// runCloseWithReview runs the mechanical close, then — for a standalone whole-issue
// close — auto-dispatches the one binary-owned boundary review (#69) on the
// whole-issue window (branch-point..HEAD), emits its Review-Verdict trailer, and
// mirrors the verdict into the close log line.
//
// milestone-close does NOT route through here: it calls computeClose directly
// (then reviews + finalizes), dispatching its own per-milestone review, so a
// milestone is never reviewed twice. The guard is structural — only a full-issue close (`f.Milestone == ""`)
// reaches the dispatch — and is the load-bearing invariant for "exactly one
// review per boundary". For a no-milestone issue this is the single review the
// boundary gets (previously it got none from the binary); for a multi-milestone
// issue it is the end-of-issue integration review (each milestone already
// reviewed its own slice).
func runCloseWithReview(stdout, stderr io.Writer, f *closeFlags) error {
	// #146: `sdlc close --milestone` was a redundant no-review milestone close — it
	// ran the mechanical close but skipped the boundary review milestone-close
	// dispatches, with no signal. Removed from the public surface (flag hidden):
	// refuse with a redirect. Returnable error, NOT die() — die()→os.Exit would kill
	// the test binary; runCloseWithReview returns error under SilenceErrors, so
	// main.go prints it. The mechanical milestone close still lives in
	// computeClose (which milestone-close calls directly, then reviews + finalizes).
	if f.Milestone != "" {
		return fmt.Errorf(
			"`sdlc close` no longer closes milestones (it would skip the boundary review).\n"+
				"  reviewed:       sdlc milestone-close --issue %d --milestone %s\n"+
				"  explicit skip:  sdlc milestone-close --issue %d --milestone %s --no-judge",
			f.Issue, f.Milestone, f.Issue, f.Milestone)
	}

	// Whole-issue close (#139): COMPUTE the close but write nothing yet — the
	// boundary review runs against the un-mutated working tree, and applyClose
	// fires only after a finalizing verdict.
	r := computeClose(stderr, f)

	// Window spans the whole branch (milestone "" → branch start), so issuePath
	// isn't consulted for the base — pass "" (#58).
	base, baseLong, head := resolveReviewWindow(strconv.Itoa(f.Issue), "", "")
	switch {
	case f.skip("judge"):
		// Explicit operator skip → finalize (this runs BEFORE dispatch, so only a
		// dispatch-ERROR VerdictNotRun ever reaches closeVerdictOutcome's halt).
		cinfo(stderr, "skipping issue boundary review per --no-judge (or --force)")
		applyClose(stdout, stderr, closeRunner, f, r)
		return finishBoundaryReview(stdout, stderr, f,
			reviewResult{Verdict: judge.VerdictNotRun, Reason: "--no-judge", Base: base, Head: head, BaseLong: baseLong})
	case f.DryRun:
		cinfo(stderr, "dry-run — would dispatch the issue boundary review")
		printCloseDryRun(stderr, r)
		return printBoundaryReviewDryRun(stdout, stderr, boundaryReviewParams{
			Label:         "#" + strconv.Itoa(f.Issue),
			Base:          base,
			BaseLong:      baseLong,
			Head:          head,
			IssuesDir:     f.IssuesDir,
			Agent:         f.Agent,
			AgentExplicit: f.AgentExplicit,
			IssueNum:      f.Issue, // #137: the dry-run prompt orientation needs this too
			Milestone:     "",
		})
	}

	return reviewThenFinalize(stdout, stderr, f, r, boundaryReviewParams{
		Label:         "#" + strconv.Itoa(f.Issue),
		Base:          base,
		BaseLong:      baseLong,
		Head:          head,
		IssuesDir:     f.IssuesDir,
		Agent:         f.Agent,
		AgentExplicit: f.AgentExplicit,
		IssueNum:      f.Issue,
		Milestone:     "",
		PlansDir:      f.plansDir(),
	})
}

func runCloseWithReviewLocked(cmd *cobra.Command, stdout, stderr io.Writer, f *closeFlags) error {
	if f.Milestone != "" || f.skip("judge") || f.DryRun {
		return withRequiredRepoTransactionLock(cmd, func() error {
			return runCloseWithReview(stdout, stderr, f)
		})
	}

	var r closeResult
	var base, baseLong, head string
	var snapshot closeReviewSnapshot
	var prior string
	if err := withRequiredRepoTransactionLock(cmd, func() error {
		r = computeClose(stderr, f)
		base, baseLong, head = resolveReviewWindow(strconv.Itoa(f.Issue), "", "")
		captured, captureErr := captureCloseReviewSnapshot(r, head, "", f.plansDir())
		if captureErr != nil {
			return captureErr
		}
		snapshot = captured
		prior = boundaryPriorFindings(stderr, boundaryReviewParams{
			IssuesDir: f.IssuesDir, IssueNum: f.Issue, Milestone: "", PlansDir: f.plansDir(),
		})
		return nil
	}); err != nil {
		return err
	}

	return reviewThenFinalizeLocked(cmd, stdout, stderr, f, r, boundaryReviewParams{
		Label:         "#" + strconv.Itoa(f.Issue),
		Base:          base,
		BaseLong:      baseLong,
		Head:          head,
		IssuesDir:     f.IssuesDir,
		Agent:         f.Agent,
		AgentExplicit: f.AgentExplicit,
		IssueNum:      f.Issue,
		Milestone:     "",
		PlansDir:      f.plansDir(),
		PriorFindings: prior,
	}, snapshot)
}

// closeOutcome is what the boundary verdict tells close to do (#139).
type closeOutcome int

const (
	closeFinalize closeOutcome = iota // apply the close
	closeRework                       // leave working; fix + re-run
	closeHalt                         // unexpected verdict; stop, consult a human
)

// closeVerdictOutcome maps a boundary verdict to a close outcome — DERIVED from
// the #147 verdict single-source (vocab.Verdict()), not a hardcoded switch, so a
// new token in verdict.cue flows here automatically. Only a finalizing verdict
// finalizes; REWORK reworks; everything else (unknown, a dispatch-error not-run)
// halts rather than papering over an ambiguous gate (#139).
func closeVerdictOutcome(v judge.Verdict) closeOutcome {
	switch t := string(v); {
	case vocab.Verdict().IsFinalizing(t):
		return closeFinalize
	case vocab.Verdict().IsBlocking(t):
		return closeRework
	default:
		return closeHalt
	}
}

// closeVerb returns the sdlc verb that owns a close of this shape — the milestone
// verb when a milestone tag is set, else the whole-issue close. Single source of
// the mode→verb mapping (#146), reused by the re-run hints (explainActual /
// explainVerified) so a gate refusal never suggests the removed `close --milestone`
// bypass path.
func closeVerb(milestone string) string {
	if milestone != "" {
		return "sdlc milestone-close"
	}
	return "sdlc close"
}

// rerunCmd builds the "Then re-run:" command line printed by a close gate refusal
// (explainActual / explainVerified). It picks the verb via closeVerb(milestone),
// so a milestone refusal points at `sdlc milestone-close` — never the removed
// `close --milestone` bypass (#146). actualArg is the pre-formatted " --actual X"
// segment (a concrete value or the " --actual <hours>" placeholder). Pure.
func rerunCmd(issueStr, milestone, actualArg string) string {
	ms := ""
	if milestone != "" {
		ms = " --milestone " + milestone
	}
	return fmt.Sprintf("%s --issue %s%s%s --verified '<evidence>'", closeVerb(milestone), issueStr, ms, actualArg)
}

// reviewThenFinalize dispatches the boundary review for an already-computed close
// and finalizes ONLY on a finalizing verdict (#139). Shared by full-issue close
// and milestone-close (annotateLogLineWithVerdict keys on f.Milestone). On REWORK
// or an unexpected verdict it writes NOTHING (issue stays `working`), emits the
// trailer for the record, and returns a non-nil error.
func reviewThenFinalize(stdout, stderr io.Writer, f *closeFlags, r closeResult, p boundaryReviewParams) error {
	review := dispatchBoundaryReview(stdout, stderr, p)
	return finalizeBoundaryReview(stdout, stderr, f, r, review, p, nil)
}

func reviewThenFinalizeLocked(cmd *cobra.Command, stdout, stderr io.Writer, f *closeFlags, r closeResult, p boundaryReviewParams, snapshot closeReviewSnapshot) error {
	dispatchParams := p
	dispatchParams.ReviewPlansDir = p.PlansDir
	dispatchParams.PlansDir = "" // sidecar is a repo write; persist it after reacquiring the lock.
	review := dispatchBoundaryReview(stdout, stderr, dispatchParams)
	return withRequiredRepoTransactionLock(cmd, func() error {
		return finalizeBoundaryReview(stdout, stderr, f, r, review, p, snapshot.validate)
	})
}

func finalizeBoundaryReview(stdout, stderr io.Writer, f *closeFlags, r closeResult, review reviewResult, p boundaryReviewParams, validate func() (string, error)) error {
	kind := "close"
	if f.Milestone != "" {
		kind = "milestone-close"
	}
	if review.Output != "" && review.SidecarPath == "" && p.PlansDir != "" {
		p.Agent = review.Agent
		if path, werr := writeReviewSidecar(p, string(review.Verdict), review.Output, nowRFC3339()); werr != nil {
			cwarn(stderr, fmt.Sprintf("review sidecar not written: %v", werr))
		} else {
			review.SidecarPath = path
			cok(stderr, "review sidecar: "+path)
		}
	}
	verb := closeVerb(f.Milestone)

	// #194 M2 (D4): the verdict and the ledger are BOTH gates, and finalizing requires
	// both to clear — an AND, not a fallback. A SHIP verdict carrying an undisposed
	// Important finding means the reviewer contradicted itself; taking the token at face
	// value there is exactly the pre-#187 posture the ledger exists to replace. REWORK
	// with everything disposed still reworks, because the verdict is still blocking.
	//
	// The round is persisted BEFORE the switch so a REWORK round lands in the ledger
	// too: the next round must be able to see what this one said.
	// Stamp the waiver onto the round the gate is about to write, or a bypassed refusal
	// reads as a clean pass in the one durable record of what this gate did (#194 close
	// review BR-39: the field existed but was set at zero call sites — inert, and worse
	// than absent because its comment claimed otherwise).
	if f.skip("ledger") {
		p.ForcedRationale = "--no-ledger (or --force): " + orPlaceholder(f.Verified, "no rationale given")
	}
	ledger := persistBoundaryRound(stderr, p, review, nowRFC3339())

	switch closeVerdictOutcome(review.Verdict) {
	case closeFinalize:
		if ledger.Block && f.skip("ledger") {
			cwarn(stderr, "--no-ledger (or --force): skipping the gate-ledger open-findings refusal")
		} else if ledger.Block {
			emitTrailerBlock(stdout, review, kind)
			if len(ledger.OpenBlocking) == 0 {
				// The ledger itself was unusable (blockOnLedgerFailure) — there are no
				// findings to name, and saying "open blocking finding(s)" would send the
				// operator looking for findings that do not exist.
				cwarn(stderr, fmt.Sprintf("%s NOT finalized: %s", kind, ledger.Reason))
				return fmt.Errorf("boundary gate: %s", ledger.Reason)
			}
			cwarn(stderr, fmt.Sprintf("boundary review: verdict %s, but the gate ledger still has open blocking finding(s) — %s NOT finalized", review.Verdict, kind))
			for _, fnd := range ledger.OpenBlocking {
				cwarn(stderr, fmt.Sprintf("  [%s] %s — %s", fnd.ID, fnd.Severity, fnd.Title))
			}
			cwarn(stderr, fmt.Sprintf("address them, or have the review dispose them explicitly, then re-run `%s`.\n"+
				"  %s\n"+
				"  Or pass --no-ledger (or --force); record why in --verified.", verb, fixTheClassLine()))
			return fmt.Errorf("boundary gate: %d open blocking finding(s) despite verdict %s", len(ledger.OpenBlocking), review.Verdict)
		}
		if validate != nil {
			note, err := validate()
			if err != nil {
				emitTrailerBlock(stdout, review, kind)
				cwarn(stderr, fmt.Sprintf("boundary review: reviewed state changed while the lock was released — close NOT finalized: %v", err))
				// #194 M1 review I3: formatAnchorRefusal carries its own re-run
				// instruction, but it is reached ONLY on the anchor branches — the
				// issue-file, project-file and git-error branches would otherwise
				// surface with no next step, and AGENTS.md §5 makes "errors are
				// next-action specs" a property of this gate.
				if !strings.Contains(err.Error(), "re-run `") {
					cwarn(stderr, fmt.Sprintf("re-run `%s` so the review covers the current repo state", verb))
				}
				return fmt.Errorf("boundary review stale: %w", err)
			}
			// #194: say what the gate decided when it let a delta through — silence
			// would read as "nothing happened", which is not what occurred.
			if note != "" {
				cinfo(stderr, note)
			}
		}
		applyClose(stdout, stderr, closeRunner, f, r)
		emitTrailerBlock(stdout, review, kind)
		if err := annotateLogLineWithVerdict(f.IssuesDir, f.Issue, f.Milestone, review.Verdict); err != nil {
			cwarn(stderr, fmt.Sprintf("log-line verdict annotation skipped: %v", err))
		}
		if review.Verdict == judge.VerdictFixThenShip {
			// #174: state the post-FIX-THEN-SHIP protocol at the moment of
			// ambiguity — before the lessons reminder, so bookkeeping lands
			// inside the same pre-commit window the protocol describes.
			cwarn(stderr, formatFixThenShipProtocol(verb))
		}
		if f.Milestone == "" { // #160 Q4: lessons ping only at the whole-issue close boundary
			emitLessonsReminder(stdout)
		}
		return nil
	case closeRework:
		emitTrailerBlock(stdout, review, kind)
		cwarn(stderr, "boundary review: REWORK — close NOT finalized; issue left at status: working")
		cwarn(stderr, fmt.Sprintf("address the findings, then re-run `%s` (no --no-reclose-guard needed)%s", verb, fixTheClassNote()))
		return fmt.Errorf("boundary review verdict REWORK — close not finalized")
	default: // closeHalt
		emitTrailerBlock(stdout, review, kind)
		cwarn(stderr, fmt.Sprintf("boundary review verdict %q is UNEXPECTED — close NOT finalized; issue left at status: working", review.Verdict))
		cwarn(stderr, "the review produced no clear SHIP/FIX-THEN-SHIP/REWORK verdict (a gate/prompt bug?).")
		cwarn(stderr, "STOP: investigate the review output (sidecar) and consult a human before re-running.")
		return fmt.Errorf("boundary review verdict %q — unexpected; close not finalized, consult a human", review.Verdict)
	}
}

// closeReviewSnapshot pins the repo state the boundary review is about to read, so
// finalization can tell whether that state still holds when the review returns ~20
// minutes later.
type closeReviewSnapshot struct {
	// reviewed is the CONCRETE SHA the review read (#194) — supplied by the caller,
	// which resolved it under the same lock that captured this snapshot, rather than
	// re-`rev-parse`d here. That identity is what makes the three users of the value
	// (the dispatched diff, the durable record, this check) provably agree.
	reviewed string
	// milestone distinguishes a milestone close from a whole-issue close, so a refusal
	// names the right re-run verb via closeVerb (ARCH-DRY).
	milestone string
	artifacts []closeReviewArtifact
}

type closeReviewArtifact struct {
	path    string
	present bool
	text    string
}

func captureCloseReviewSnapshot(r closeResult, reviewedSHA, milestone, plansDir string) (closeReviewSnapshot, error) {
	s := closeReviewSnapshot{
		reviewed:  reviewedSHA,
		milestone: milestone,
	}
	if r.issuePath != "" {
		s.artifacts = append(s.artifacts, closeReviewArtifact{path: r.issuePath, present: true, text: r.issueText})
	}
	for _, project := range r.projectEdits {
		s.artifacts = append(s.artifacts, closeReviewArtifact{path: project.path, present: true, text: project.oldText})
	}
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return closeReviewSnapshot{}, fmt.Errorf("resolve review repository root for snapshot: %w", err)
	}
	_, planPath := reviewPlanPaths(root, plansDir, r.issuePath)
	if planPath != "" {
		plan, err := captureCloseReviewArtifact(planPath)
		if err != nil {
			return closeReviewSnapshot{}, err
		}
		s.artifacts = append(s.artifacts, plan)
	}
	return s, nil
}

func captureCloseReviewArtifact(path string) (closeReviewArtifact, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return closeReviewArtifact{path: path, present: true, text: string(data)}, nil
	}
	if os.IsNotExist(err) {
		return closeReviewArtifact{path: path}, nil
	}
	if info, statErr := os.Stat(path); statErr == nil && info.IsDir() {
		return closeReviewArtifact{path: path}, nil
	}
	return closeReviewArtifact{}, fmt.Errorf("read review artifact %s: %w", path, err)
}

// validate reports whether finalization may proceed. It returns a note the caller
// surfaces when the gate allowed a delta through, so the operator learns what it
// decided rather than only that nothing blocked.
//
// #194: the HEAD question is now "does the delta touch code?", not "is HEAD identical?".
// Mutable artifact checks stay STRICT — the reviewer reads issue/project prose and the
// optional canonical plan, so a mid-review content or presence change is a genuine
// invalidation. Only the HEAD check was ever stricter than its own purpose.
func (s closeReviewSnapshot) validate() (string, error) {
	note := ""
	if s.reviewed != "" {
		d, err := gatherReviewAnchorDelta(s.reviewed)
		if err != nil {
			return "", err // fail closed on a git error, as the publish gate does
		}
		if d.Reviewed == "" {
			// The window head never resolved, so there is no anchor to classify
			// against. Say so — silence would read as "checked, nothing moved".
			return "boundary review: no resolved anchor for this window — the mid-review " +
				"delta check did not run (#194)", nil
		}
		switch outcome := classifyReviewAnchor(d); outcome {
		case anchorDocsOnly:
			note = formatAnchorDocsOnly(d)
		case anchorCodeDelta, anchorDiverged:
			return "", fmt.Errorf("%s", formatAnchorRefusal(d, outcome, closeVerb(s.milestone)))
		}
	}
	for _, artifact := range s.artifacts {
		current, err := captureCloseReviewArtifact(artifact.path)
		if err != nil {
			return "", err
		}
		if current.present != artifact.present {
			if current.present {
				return "", fmt.Errorf("%s appeared", artifact.path)
			}
			return "", fmt.Errorf("%s disappeared", artifact.path)
		}
		if current.present && current.text != artifact.text {
			return "", fmt.Errorf("%s changed", artifact.path)
		}
	}
	return note, nil
}

// finishBoundaryReview emits the close trailer and mirrors the verdict into the
// issue's close log line — for BOTH the dispatched and the --no-judge/not-run
// paths, matching milestone-close's behavior (which annotates even when the judge
// didn't run). #69 M2 review I1: the no-judge path used to skip the annotation.
func finishBoundaryReview(stdout, stderr io.Writer, f *closeFlags, result reviewResult) error {
	emitTrailerBlock(stdout, result, "close")
	if err := annotateLogLineWithVerdict(f.IssuesDir, f.Issue, "", result.Verdict); err != nil {
		cwarn(stderr, fmt.Sprintf("log-line verdict annotation skipped: %v", err))
	}
	emitLessonsReminder(stdout) // #160 Q4: whole-issue close (this path is milestone-less by construction)
	return nil
}

// emitLessonsReminder prints the no-LLM lessons ping at a whole-issue close (#160
// Q4). It moved here from the publish gate (merge/push) so it fires while the agent
// is engaged and the boundary-review findings are fresh — when non-obvious patterns
// are easiest to capture into workshop/lessons.md (constitution §4).
func emitLessonsReminder(stdout io.Writer) {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, judge.LessonsReminder)
}

// ── explainers ───────────────────────────────────────────────────────────────

// computeActualForCloseFn is the measurement seam for the omit-path (#178) —
// a package var so tests can stub the engine (the file's validateChangedInstancesFn
// pattern). Production resolves roots and runs the same engine as `sdlc actual`.
var computeActualForCloseFn = func(issueStr string) actualResult {
	repoTop, brainAbs := resolveActualRoots()
	return computeActual(repoTop, brainAbs, issueStr)
}

// resolveOmittedActual is the pure omit-path decision (#178): adopt a measured
// value (format pinned to %.2f — the ledger value must equal what the info line
// shows) or refuse. Only actualMeasured adopts; telemetry-gap/no-window/empty/
// error keep the explainActual refusal (judgment paths).
func resolveOmittedActual(res actualResult) (string, bool) {
	if res.Status != actualMeasured {
		return "", false
	}
	return strconv.FormatFloat(res.Hours, 'f', 2, 64), true
}

// formatAdoptLine renders the adoption info line (issue mode only — milestone
// closes keep the suggest flow, see adoptOmittedActual). Must never match a
// GateCatalog ACK/refusal pattern (TestAdoptLineNoGatesigCollision).
func formatAdoptLine(res actualResult) string {
	line := fmt.Sprintf("using measured actual: %.2fh (window %s", res.Hours, res.Window)
	if len(res.Peers) > 1 {
		line += "; attributed across window issues: " + strings.Join(prefixHash(res.Peers), ", ")
	}
	return line + ")"
}

// adoptOmittedActual is the omit-path wiring (#178): measure once via the seam,
// adopt into f.Actual on success (deviation check is then skipped — it would
// compare the measurement with itself), print the info line. On unmeasurable
// statuses it returns ok=false with NO side effects — the caller explains
// (reusing the same measurement) and exits.
func adoptOmittedActual(stderr io.Writer, f *closeFlags, issueStr, mode string) (actualResult, bool) {
	res := computeActualForCloseFn(issueStr)
	if mode == "milestone" {
		// #178 close-review Important #1: per-milestone project detail blocks
		// record per-milestone hours, but computeActual's window is ISSUE-scoped
		// (claim → HEAD) — auto-adopting the cumulative value at M2+ would
		// double-count across blocks (lessons.md increments rule). Milestones
		// keep the suggest-then-decide flow (the agent applies the increment
		// arithmetic) until a windowed per-milestone measurement lands.
		return res, false
	}
	v, ok := resolveOmittedActual(res)
	if !ok {
		return res, false
	}
	f.Actual = v
	cinfo(stderr, formatAdoptLine(res))
	for _, w := range res.Warnings {
		cwarn(stderr, w)
	}
	return res, true
}

// explainActual is now the UNMEASURABLE explainer (#178): the omit-path adopts
// a successful measurement, so this renders only when the engine couldn't
// produce a number — res (already computed by the caller) says why, and the
// agent supplies a labeled judgment value.
func explainActual(stderr io.Writer, issueStr, mode, milestone string, res actualResult) {
	var head []string
	head = append(head, fmt.Sprintf("%sACTUAL=<hours> required for %s close (§5 step 3) — measurement unavailable for this window.%s", ansiRed, mode, ansiReset), "")
	head = append(head, fmt.Sprintf("  %sSemantic:%s  focused dev-hours on this %s (#%s) — not wall-clock.", ansiCyan, ansiReset, mode, issueStr))
	head = append(head, "             (a measured value is adopted automatically; you only land here", "             when the engine can't measure — the status below says why)", "")
	fmt.Fprintln(stderr, strings.Join(head, "\n"))

	// #68 M2: render the engine's diagnosis (telemetry gap / no window / error)
	// with its next-action guidance. Same engine as `sdlc actual`; the caller
	// already ran it — no re-measure.
	printActual(stderr, res)

	var tail []string
	tail = append(tail, "", fmt.Sprintf("  %sThen re-run:%s", ansiCyan, ansiReset))
	tail = append(tail, "    "+rerunCmd(issueStr, milestone, " --actual <hours>"), "")
	tail = append(tail, fmt.Sprintf("  (Re-measure anytime: sdlc actual --issue %s)", issueStr))
	tail = append(tail, "  Pass --no-actual (or --force) only when measurement is not applicable; close records actual_hours: N/A and skips calibration.")
	fmt.Fprintln(stderr, strings.Join(tail, "\n"))
}

// resolveActualRoots returns the repo top (cwd fallback) and the sibling brain
// dir — the two roots computeActual needs. Shared by explainActual (omit-path)
// and checkActualDeviation (pass-path) so the resolution lives in one place.
func resolveActualRoots() (repoTop, brainAbs string) {
	repoTop, err := gitx.RepoTopLevel()
	if err != nil || repoTop == "" {
		cwd, _ := os.Getwd()
		repoTop, _ = filepath.Abs(cwd)
	}
	brainAbs, _ = filepath.Abs(filepath.Join(repoTop, "..", "brain"))
	return repoTop, brainAbs
}

// #87: backstop for a hand-passed --actual that doesn't match reality.
//
// Thresholds (recorded per the design step): we compare a passed value to the
// active-time-v3 measurement and gate on the RATIO, with an absolute-difference
// FLOOR so small legitimate gaps never trip (e.g. 0.05 vs 0.15 is 3× but only
// 0.10h apart — noise, not fabrication).
const (
	actualDevAbsFloor    = 0.5  // hours; deviations smaller than this are ignored
	actualDevWarnRatio   = 3.0  // ≥ this → warn
	actualDevRefuseRatio = 10.0 // ≥ this → refuse (looks fabricated/fat-fingered)
)

type devVerdict int

const (
	devOK     devVerdict = iota // within tolerance — close silently
	devWarn                     // moderately off — warn but proceed
	devRefuse                   // wildly off — refuse without an explicit override
)

// actualDeviation is the pure comparator (ARCH-PURE: no IO; unit-tested
// directly). Returns the verdict and the ratio (max/min, ≥1) for the message.
func actualDeviation(passed, measured float64) (devVerdict, float64) {
	if math.Abs(passed-measured) < actualDevAbsFloor {
		return devOK, 1
	}
	hi, lo := passed, measured
	if measured > passed {
		hi, lo = measured, passed
	}
	if lo <= 0 {
		lo = 0.01 // avoid div-by-zero / inf when measured is ~0
	}
	ratio := hi / lo
	switch {
	case ratio >= actualDevRefuseRatio:
		return devRefuse, ratio
	case ratio >= actualDevWarnRatio:
		return devWarn, ratio
	default:
		return devOK, ratio
	}
}

// checkActualDeviation is the thin IO glue for issue-close values: measure via
// the shared engine, run the pure comparator, and warn (to stderr) or return a
// refusal error. Milestone values are increments but the available measurement
// is cumulative claim→HEAD, so they are deliberately skipped until a windowed
// milestone measurement exists. Unavailable issue measurements also never gate.
func checkActualDeviation(stderr io.Writer, issueStr string, passed float64, mode string) error {
	if mode == "milestone" {
		return nil
	}
	res := computeActualForCloseFn(issueStr)
	if res.Status != actualMeasured {
		return nil // can't measure → don't block (judgment path owns this)
	}
	verdict, ratio := actualDeviation(passed, res.Hours)
	switch verdict {
	case devRefuse:
		return fmt.Errorf("--actual %.2f is %.0f× the active-time-v3 measurement (%.2fh) — looks fabricated or fat-fingered.\n"+
			"  Run `sdlc actual --issue %s` and pass the measured value; or --force (reason in --verified) if the\n"+
			"  measurement is genuinely wrong (a bad actual pollutes velocity calibration — the gate exists to catch this).",
			passed, ratio, res.Hours, issueStr)
	case devWarn:
		cwarn(stderr, fmt.Sprintf("--actual %.2f is %.1f× the active-time-v3 measurement (%.2fh) — confirm it's right, or re-run `sdlc actual --issue %s`",
			passed, ratio, res.Hours, issueStr))
	}
	return nil
}

func explainVerified(stderr io.Writer, issueStr, mode, milestone, actual string) {
	var lines []string
	lines = append(lines, fmt.Sprintf("%sVERIFIED=\"<one-line evidence>\" required for %s close (§5 step 1).%s", ansiRed, mode, ansiReset), "")
	lines = append(lines, fmt.Sprintf("  %sSemantic:%s  one-line evidence the work meets the issue's done-when.", ansiCyan, ansiReset))
	lines = append(lines, "             Behavior, not artifacts: 'tests pass' beats 'code written'.", "")
	lines = append(lines, fmt.Sprintf("  %sExamples:%s", ansiCyan, ansiReset))
	lines = append(lines, "    VERIFIED='ran make test, all green'")
	lines = append(lines, "    VERIFIED='e2e flow X→Y verified manually'")
	lines = append(lines, "    VERIFIED='code-review subagent, all Important addressed in <sha>'")
	lines = append(lines, "    VERIFIED='ran make nous-test-bootstrap, ROUND-TRIP-OK in 2:34'", "")
	actualArg := " --actual <hours>"
	if actual != "" {
		actualArg = " --actual " + actual
	}
	lines = append(lines, fmt.Sprintf("  %sThen re-run:%s", ansiCyan, ansiReset))
	lines = append(lines, "    "+rerunCmd(issueStr, milestone, actualArg), "")
	lines = append(lines, "  Pass --no-verified (or --force) only if there's genuinely no behavior to verify.")
	fmt.Fprintln(stderr, strings.Join(lines, "\n"))
}

// hasCodePath reports whether any window path is code surface — the single
// docs classifier (#177, aligned with the #172 windowstat study): *.md anywhere,
// or anything under workshop/, atlas/, docs/, is documentation; EVERYTHING else
// (Makefile, .gitignore, extensionless files) conservatively counts as code —
// build files are architectural surface, so they keep the atlas refusal.
func hasCodePath(paths []string) bool {
	for _, p := range paths {
		if strings.HasSuffix(p, ".md") ||
			strings.HasPrefix(p, "workshop/") ||
			strings.HasPrefix(p, "atlas/") ||
			strings.HasPrefix(p, "docs/") {
			continue
		}
		return true
	}
	return false
}

// atlasAutoSatisfyLine renders the #177 info line. Must never match a
// GateCatalog ACK/refusal pattern (TestAtlasAutoSatisfyLineNoGatesigCollision) —
// the friction instrument must not count auto-satisfactions as gate events.
func atlasAutoSatisfyLine(nDocs int) string {
	return fmt.Sprintf("atlas gate: no code surface in window (%d doc/workshop file(s)) — auto-satisfied", nDocs)
}

func explainNoAtlas(stderr io.Writer, windowBaseShort string, nonAtlas []string) {
	atlasFiles, _ := filepath.Glob("atlas/*.md")
	sort.Strings(atlasFiles)

	// Count top-level path frequencies: "split at most 2", join first 2 parts.
	counts := map[string]int{}
	for _, p := range nonAtlas {
		parts := strings.SplitN(p, "/", 3)
		var key string
		if len(parts) >= 2 {
			key = parts[0] + "/" + parts[1]
		} else {
			key = parts[0]
		}
		counts[key]++
	}
	type kv struct {
		k string
		v int
	}
	var ranked []kv
	for k, v := range counts {
		ranked = append(ranked, kv{k, v})
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].v != ranked[j].v {
			return ranked[i].v > ranked[j].v
		}
		return ranked[i].k < ranked[j].k
	})
	if len(ranked) > 10 {
		ranked = ranked[:10]
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("no atlas/ changes in %s..HEAD (§5 step 5).", windowBaseShort))
	if len(atlasFiles) > 0 {
		lines = append(lines, "  Existing atlas files (pick the one matching new surface):")
		for _, a := range atlasFiles {
			lines = append(lines, "    "+a)
		}
	}
	if len(ranked) > 0 {
		lines = append(lines, "  Code paths changed in this window:")
		for _, r := range ranked {
			plural := "s"
			if r.v == 1 {
				plural = ""
			}
			lines = append(lines, fmt.Sprintf("    %s (%d file%s)", r.k, r.v, plural))
		}
	}
	lines = append(lines, "  Update atlas where this work introduces architectural surface,")
	lines = append(lines, "  or pass --no-atlas (or --force) with the rationale in --verified")
	lines = append(lines, "  (e.g., 'pure bugfix, no new surface').")
	die(stderr, strings.Join(lines, "\n"))
}

// ── milestone-verdict guard ──────────────────────────────────────────────────

// milestonePlanRE matches a ticked-or-unticked milestone bullet at the
// start of a plan-section line:
//
//   - [x] **M1 — scaffold …
//   - [ ] **M4b — port milestone-close
//   - [.] **M5 — wip
//
// Captures the milestone tag (group 1, e.g. "M1" or "M4b"). The bold
// asterisks are typical but not strictly required — we accept both the
// emphasized and plain forms so the regex doesn't drift away from
// existing issue files that vary the formatting.
var milestonePlanRE = regexp.MustCompile(`(?m)^- \[[ x.]\] \*{0,2}(M\d+[a-z]?)\b`)

// partitionMissingVerdicts splits the missing-verdict milestones by plan
// position relative to the LAST verdict-carrying milestone (#175). Missing
// rows before it are "midstream" — a later boundary was crossed with no
// review evidence for them, a genuine §3 violation. Missing rows after it
// (or every row, when nothing carries a verdict — the single-pass case)
// are "trailing" — no reviewed boundary follows them, so the imminent
// issue-close boundary review (window branch-point→HEAD) is their review.
// Pure; plan order is preserved in both outputs.
func partitionMissingVerdicts(ordered, missing []string) (midstream, trailing []string) {
	missingSet := make(map[string]bool, len(missing))
	for _, tag := range missing {
		missingSet[tag] = true
	}
	last := -1 // index of the last verdict-carrying milestone
	for i, tag := range ordered {
		if !missingSet[tag] {
			last = i
		}
	}
	for i, tag := range ordered {
		if !missingSet[tag] {
			continue
		}
		if i < last {
			midstream = append(midstream, tag)
		} else {
			trailing = append(trailing, tag)
		}
	}
	return midstream, trailing
}

// findMilestonesMissingVerdict enumerates milestones in the issue body's
// `## Plan` section and returns them in plan order (ordered), plus the
// tags of any whose close commit lacks a `Review-Verdict:` trailer
// (missing). The caller partitions missing against ordered to tell
// midstream from trailing misses (#175, partitionMissingVerdicts).
//
// "Close commit" for milestone Mx = a commit whose subject opens with
// `#<issue> Mx:` AND whose message body contains a `Review-Verdict:`
// trailer line. The conjunctive `--all-match` over both --grep patterns
// matches the task spec exactly.
//
// Returns (ordered, [], nil) when every milestone has evidence. Returns
// (nil, nil, err) only on hard failures (issue body unparseable, git
// unavailable). A milestone whose subject doesn't match any commit is
// treated the same as one whose commit lacks the trailer — both are "no
// review evidence."
func findMilestonesMissingVerdict(body, issueStr, issuePath string) (ordered, missing []string, err error) {
	planBody, ok := issue.PlanItemsBody(body)
	if !ok {
		// No plan section → no milestones to check. Treat as "fine":
		// the operator may be closing an issue that never had milestones.
		return nil, nil, nil
	}
	matches := milestonePlanRE.FindAllStringSubmatch(planBody, -1)
	if len(matches) == 0 {
		return nil, nil, nil
	}
	// Preserve plan order; de-duplicate (a milestone may appear in the
	// plan more than once if revised).
	seen := map[string]bool{}
	for _, mm := range matches {
		tag := mm[1]
		if seen[tag] {
			continue
		}
		seen[tag] = true
		ordered = append(ordered, tag)
	}
	for _, tag := range ordered {
		ok, err := milestoneHasVerdictCommit(issueStr, tag, issuePath)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			missing = append(missing, tag)
		}
	}
	return ordered, missing, nil
}

// milestoneHasVerdictCommit reports whether `git log` finds a commit
// matching both the subject anchor `#<issue> <milestone>:` and the
// trailer presence `Review-Verdict:`, scoped to commits that touched
// the issue file (so unrelated history grepping the same string can't
// satisfy the check).
//
// Uses `--all-match -1 -F` semantics via gitx so the patterns are
// treated as fixed strings rather than regex (the colon and braces in
// commit subjects don't bite us).
func milestoneHasVerdictCommit(issueStr, milestone, issuePath string) (bool, error) {
	// `Mx` may be followed by a colon (`#56 M1: …`) or more subject words
	// before the colon (`#56 M1 close: …`) — both are natural milestone-close
	// subjects. Anchor on `Mx` + a colon-or-space boundary so neither form is
	// missed, while `M1` still can't match `M10` (next char would be a digit).
	subjectGrep := fmt.Sprintf("^#%s %s[: ]", issueStr, milestone)
	args := []string{
		"log",
		"--grep=" + subjectGrep,
		"--grep=Review-Verdict:",
		"--all-match",
		"-E", // ERE for the subject `^` anchor; the verdict grep matches the literal colon either way
		"--max-count=1",
		"--pretty=format:%H",
		"--", issuePath,
	}
	out, err := gitx.RunGit(args...)
	if err != nil {
		// git log failed (not a repo, etc.). Surface as a hard error so
		// the caller's `cwarn → skip` branch fires rather than silently
		// passing every milestone.
		return false, fmt.Errorf("git log: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// verdictBypassClosingLine is the shared last line of every no-verdict
// refusal. internal/processmanual/gatesig.go regex-pins it for friction
// attribution (#172) — reword there and here together, never one side.
const verdictBypassClosingLine = "  Or pass --no-verdict (or --force); record the reason in --verified."

// verdictNextActionLines renders the per-milestone retroactive-review
// commands + trailer shape shared by both no-verdict refusals (ARCH-DRY).
// Pure.
func verdictNextActionLines(issueStr string, tags []string) []string {
	lines := []string{fmt.Sprintf("  %sNext actions:%s", ansiCyan, ansiReset)}
	for _, tag := range tags {
		lines = append(lines, fmt.Sprintf("    sdlc judge milestone-review --issue %s --milestone %s", issueStr, tag))
	}
	lines = append(lines,
		"    # then amend the milestone-close commit (or land a new commit)",
		"    # with these trailers:",
		"    #   Review-Verdict: SHIP",
		"    #   Review-Window: <base>..<head>")
	return lines
}

// formatMissingVerdicts builds the next-action error message naming the
// milestones that lack Review-Verdict trailers. Pure: no IO, no exit.
// Lives next to explainMissingVerdicts so tests can assert the contract
// without subprocessing or os.Exit gymnastics.
//
// Since #175 this fires only for MIDSTREAM misses (a later milestone
// closed with review evidence while these rows have none); trailing
// misses are accepted by the issue-close boundary review instead.
func formatMissingVerdicts(issueStr string, missing []string) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("%smilestones %s lack Review-Verdict trailer in close commits (AGENTS.md §3).%s",
		ansiRed, strings.Join(missing, ", "), ansiReset))
	lines = append(lines, "")
	lines = append(lines, "  Each milestone close must carry a fresh-eyes review verdict in")
	lines = append(lines, "  the commit message. Without it, there's no evidence the work")
	lines = append(lines, "  was reviewed before the next milestone began.")
	lines = append(lines, "")
	lines = append(lines, "  An `Mx` tag in ## Plan is a review boundary, not a task label")
	lines = append(lines, "  (AGENTS.md §3) — single-pass work should use plain checkboxes.")
	lines = append(lines, "  If this plan was over-split, fold the never-closed Mx rows into")
	lines = append(lines, "  plain checkboxes (append a ## Revisions note saying why);")
	lines = append(lines, "  otherwise land the per-row review evidence:")
	lines = append(lines, "")
	lines = append(lines, verdictNextActionLines(issueStr, missing)...)
	lines = append(lines, "")
	lines = append(lines, verdictBypassClosingLine)
	return strings.Join(lines, "\n")
}

// formatFixThenShipProtocol builds the post-FIX-THEN-SHIP next-action block
// (#174). FIX-THEN-SHIP finalizes the close but the findings still need
// fixing — and nothing else states what to do with them, which is how the
// re-close loop and the publish-gate --no-judge bypasses started (#172).
// verb is closeVerb(f.Milestone) — the escape hatch names the boundary verb
// that was actually run (same threading as the REWORK arm). Pure.
func formatFixThenShipProtocol(verb string) string {
	var lines []string
	lines = append(lines, "FIX-THEN-SHIP protocol (#174):")
	// One append, not two: the routing line must share a statement with the
	// message it routes, or a guard that attributes per statement cannot tell
	// which line it credits (#203 BR-7 shape H).
	lines = append(lines, "      1. Fix the findings NOW, before committing this close.",
		"         "+fixTheClassLine())
	lines = append(lines, "      2. Bundle the fixes + the issue-file close mutations (+ lessons/bookkeeping)")
	lines = append(lines, "         into ONE commit (or amend), so the publish gate's reviewed anchor is HEAD.")
	lines = append(lines, fmt.Sprintf("      3. Do NOT re-run `%s` — this verdict already sanctions shipping after", verb))
	lines = append(lines, "         the fixes; a second review of the same boundary is the #172 re-close loop.")
	if verb == "sdlc close" {
		// Anchor semantics are issue-close/publish-gate territory; a milestone
		// close writes no codecomplete anchor (close review #174 M-finding).
		lines = append(lines, fmt.Sprintf("      Only if fixes must land AFTER the close commit: re-run `%s` so the", verb))
		lines = append(lines, "      delta is re-reviewed and the anchor advances (doc-only deltas pass on their own).")
	} else {
		lines = append(lines, "      Fixes that land after this milestone-close commit are covered by the")
		lines = append(lines, "      NEXT boundary review — its window starts at this boundary (#58).")
	}
	return strings.Join(lines, "\n")
}

// formatTrailingVerdictAccepted builds the loud acceptance line for trailing
// unclosed milestones (#175): the issue-close boundary review's window is
// branch-point→HEAD, so their work is inside the diff this close is about to
// review — the close IS their review boundary. Pure.
func formatTrailingVerdictAccepted(trailing []string) string {
	return fmt.Sprintf("milestones %s never had their own milestone-close; accepted — the "+
		"issue-close boundary review (window branch-point→HEAD) covers their work (#175). "+
		"Next time: single-pass work takes plain checkboxes, not Mx tags (AGENTS.md §3).",
		strings.Join(trailing, ", "))
}

// formatTrailingNeedsJudge builds the refusal for trailing unclosed
// milestones when --no-judge skips the issue-close review — the coverage
// premise behind the #175 acceptance is gone. Pure. Ends with the same
// closing line gatesig.go pins for the no-verdict gate.
func formatTrailingNeedsJudge(issueStr string, trailing []string) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("%smilestones %s lack Review-Verdict trailers, and --no-judge skips the issue-close review that would cover them (#175).%s",
		ansiRed, strings.Join(trailing, ", "), ansiReset))
	lines = append(lines, "")
	lines = append(lines, "  Trailing milestones without their own milestone-close are normally")
	lines = append(lines, "  accepted because the issue-close boundary review covers their work.")
	lines = append(lines, "  With --no-judge that review never runs, so the coverage is fictional.")
	lines = append(lines, "")
	lines = append(lines, "  Drop --no-judge (let the close review run), or review each row:")
	lines = append(lines, "")
	lines = append(lines, verdictNextActionLines(issueStr, trailing)...)
	lines = append(lines, "")
	lines = append(lines, verdictBypassClosingLine)
	return strings.Join(lines, "\n")
}

func explainMissingVerdicts(stderr io.Writer, issueStr string, missing []string) {
	die(stderr, formatMissingVerdicts(issueStr, missing))
}

// orPlaceholder returns s, or fallback when s is blank — used where an operator-supplied
// rationale is recorded durably and an empty string would read as "no waiver happened".
func orPlaceholder(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}
