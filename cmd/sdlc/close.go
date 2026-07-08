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
	Issue         int
	Milestone     string
	Actual        string
	Verified      string
	Force         bool
	DryRun        bool
	BrainDir      string
	IssuesDir     string
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
	NoPlanCheck bool
	NoProject   bool
	NoJudge     bool // skip the issue boundary review (#69)
}

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
			f.AgentExplicit = cmd.Flags().Changed("agent")
			return runCloseWithReviewLocked(cmd, cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})

	cmd.Flags().IntVar(&f.Issue, "issue", 0, "issue ID (numeric, required)")
	cmd.Flags().StringVar(&f.Milestone, "milestone", "", "DEPRECATED (#146) — use `sdlc milestone-close`; passing this to `close` refuses")
	_ = cmd.Flags().MarkHidden("milestone") // #146: milestone closing moved fully to `sdlc milestone-close`
	cmd.Flags().StringVar(&f.Actual, "actual", "", "focused dev-hours (sdlc computes it; see `sdlc actual`)")
	cmd.Flags().StringVar(&f.Mode, "mode", "", "optional supervision mode for the calibration ledger: supervised | delegated (#117)")
	cmd.Flags().StringVar(&f.Verified, "verified", "", "one-line evidence the work meets done-when")
	cmd.Flags().BoolVar(&f.Force, "force", false, "bypass ALL gates (≡ every --no-* flag); record the reason in --verified")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print what would change; do not write")
	cmd.Flags().StringVar(&f.BrainDir, "brain-dir", "../brain", "path to the brain repo (for project-file lookup)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", "workshop/issues", "directory holding issue files")
	// Per-gate bypasses (#67) — each waives one guard; --force waives all.
	cmd.Flags().BoolVar(&f.NoActual, "no-actual", false, "record actual_hours: N/A and skip velocity calibration")
	cmd.Flags().BoolVar(&f.NoVerified, "no-verified", false, "bypass the VERIFIED-evidence requirement (only if there's no behavior to verify)")
	cmd.Flags().BoolVar(&f.NoReclose, "no-reclose-guard", false, "bypass the already-done refusal (intentionally re-close)")
	cmd.Flags().BoolVar(&f.NoAtlas, "no-atlas", false, "bypass the atlas/ change check (acknowledge: no new architectural surface)")
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
		"             sdlc computes it — close suggests a number, or run",
		"             `sdlc actual --issue N`. (method: 42shots/velocity/baseline-v3.md)",
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
// Anchor: the **last** `## Log` header, not the first. The real Log section is
// conventionally the final `##` section, and a meta-issue (like #66 itself) can
// quote `## Log` / `### <date>` inside a fenced code block in an earlier
// section — first-match would then file the line into that prose. (Found by
// dogfooding: closing #66 with the first-match version filed the close line
// into its own Problem-section example.) All offsets below are taken relative
// to that last header so both the day-header and fallback inserts target the
// real section.
func insertLogLine(body, logLine string) string {
	logHeaderRE := regexp.MustCompile(`(?m)^## Log\s*$`)
	all := logHeaderRE.FindAllStringIndex(body, -1)
	if all == nil {
		return strings.TrimRight(body, "\n\r\t ") + "\n\n## Log\n\n" + logLine + "\n"
	}
	logStart := all[len(all)-1][0] // start of the LAST `## Log` header
	section := body[logStart:]     // the real Log section + anything after it

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
		dayRE := regexp.MustCompile(`(?m)^### ` + regexp.QuoteMeta(m[1]) + `([ \t].*)?$`)
		if d := dayRE.FindStringIndex(section); d != nil {
			pos := logStart + d[1] // end of the day-header line text (before its newline)
			return body[:pos] + "\n" + logLine + body[pos:]
		}
	}
	// Fallback: top of the real `## Log` section. Same shape as close-issue.py's
	// regex, but run on `section` so it anchors to the last header; for the
	// common single-`## Log` body this is byte-for-byte identical to the original.
	insertRE := regexp.MustCompile(`(?m)(^## Log\s*\n)(\s*\n)?`)
	loc := insertRE.FindStringSubmatchIndex(section)
	if loc == nil {
		// Header matched logHeaderRE but not insertRE — shouldn't happen
		// in practice (the patterns are equivalent up to trailing content),
		// but fall through to append-mode rather than panic.
		return strings.TrimRight(body, "\n\r\t ") + "\n\n## Log\n\n" + logLine + "\n"
	}
	// loc[*] are relative to `section`; offset by logStart. group1 = loc[2..3].
	group1 := section[loc[2]:loc[3]]
	return body[:logStart+loc[0]] + group1 + "\n" + logLine + "\n" + body[logStart+loc[1]:]
}

// ── main entry point ─────────────────────────────────────────────────────────

// closeResult bundles everything applyClose needs, computed by computeClose
// WITHOUT any writes — so the boundary review can run against the un-mutated
// working tree and the writes fire only after a finalizing verdict (#139).
type closeResult struct {
	issuePath       string
	issueText       string // original, for the "changed?" guard
	newIssueText    string
	projectEditPath string
	projectEditText string
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
	if f.Actual != "" {
		v, err := strconv.ParseFloat(f.Actual, 64)
		if err != nil {
			die(stderr, fmt.Sprintf("ACTUAL must be a number, got '%s'", f.Actual))
		}
		// #87: sanity-check a PASSED --actual against the active-time-v3
		// measurement. The omit-path (explainActual) already measures+suggests;
		// the pass-path used to trust the value blindly, which let a fabricated
		// 13.5 (measured 0.30) pollute velocity calibration. A wildly-off value
		// is refused; a moderately-off one warns. --force/--no-actual bypasses
		// (rationale in --verified). Skips silently when the engine can't measure.
		if !f.skip("actual") {
			if derr := checkActualDeviation(stderr, issueStr, v); derr != nil {
				die(stderr, derr.Error())
			}
		}
	}

	if f.Actual == "" {
		if !f.skip("actual") {
			explainActual(stderr, issueStr, mode, f.Milestone)
			exitWithCode(1)
		}
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
			die(stderr, fmt.Sprintf("%s#%s is already status: done — pass --no-reclose-guard (or --force) to re-close intentionally", repoName, issueStr))
		}
		cwarn(stderr, fmt.Sprintf("--no-reclose-guard (or --force): re-closing %s#%s (already done)", repoName, issueStr))
	}

	// ── Commit window + atlas check ─────────────────────────────────────────
	// One window source (ARCH-DRY, #58): boundaryWindowBase gives the same base
	// the boundary review uses — the prior review boundary for a milestone close,
	// the branch start otherwise — so the atlas-coverage check and the review
	// cover exactly the same commits, including inter-milestone side-quests/fixes.
	windowBase := boundaryWindowBase(issueStr, f.Milestone, issuePath)
	if windowBase != "" {
		cinfo(stderr, fmt.Sprintf("commit window: %s → HEAD", shortSHA(windowBase)))
	} else {
		cwarn(stderr, fmt.Sprintf("no commits reference '#%s' on this branch", issueStr))
	}

	if windowBase != "" {
		diffFiles, _ := gitx.DiffNames(windowBase, "HEAD")
		var atlasChanged, nonAtlas []string
		for _, p := range diffFiles {
			if strings.HasPrefix(p, "atlas/") {
				atlasChanged = append(atlasChanged, p)
			} else {
				nonAtlas = append(nonAtlas, p)
			}
		}
		if len(atlasChanged) == 0 {
			if !f.skip("atlas") {
				explainNoAtlas(stderr, shortSHA(windowBase), nonAtlas)
				exitWithCode(1)
			}
			cwarn(stderr, "--no-atlas (or --force): skipping atlas/ change check — rationale in --verified")
		}
	}

	// ── Milestone-review verdict check (issue close only) ──────────────────
	//
	// Every milestone in the plan must carry a Review-Verdict: trailer on
	// its close commit (AGENTS.md §3 fresh-eyes review evidence). The
	// check is bypassable with --force; the rationale belongs in --verified.
	if mode == "issue" {
		missing, err := findMilestonesMissingVerdict(body, issueStr, issuePath)
		if err != nil {
			cwarn(stderr, fmt.Sprintf("milestone-verdict check skipped: %v", err))
		} else if len(missing) > 0 {
			if !f.skip("verdict") {
				explainMissingVerdicts(stderr, issueStr, missing)
				exitWithCode(1)
			}
			cwarn(stderr, fmt.Sprintf("--no-verdict (or --force): skipping Review-Verdict check for %d milestone(s): %s",
				len(missing), strings.Join(missing, ", ")))
		}
	}

	// ── Edit issue file ─────────────────────────────────────────────────────
	newFM, newBody := fm, body

	if mode == "milestone" {
		pat := regexp.MustCompile(`(?m)^(- )\[[ .]\]( ` + regexp.QuoteMeta(f.Milestone) + `\b)`)
		n := len(pat.FindAllStringIndex(newBody, -1))
		if n > 0 {
			newBody = pat.ReplaceAllString(newBody, "${1}[x]${2}")
			applied = append(applied, fmt.Sprintf("ticked %s in %s ## Plan", f.Milestone, filepath.Base(issuePath)))
		} else {
			cwarn(stderr, fmt.Sprintf("no '- [ ] %s' in %s (project-tracked issue?)", f.Milestone, filepath.Base(issuePath)))
		}
	} else { // issue close
		if m := issue.PlanSectionRE.FindStringSubmatchIndex(newBody); m != nil {
			planBody := newBody[m[2]:m[3]]
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

	// ── Locate + edit project file ──────────────────────────────────────────
	var projectEditPath string
	var projectEditText string

	projPath, err := project.FindByIssueRef(f.BrainDir, repoName, issueStr)
	if err != nil {
		cwarn(stderr, err.Error()+" — skipping project update")
	} else if projPath == "" {
		cwarn(stderr, fmt.Sprintf("no project in %s/data/project/*.md references %s#%s — skipping project update",
			f.BrainDir, repoName, issueStr))
	} else {
		projBytes, err := os.ReadFile(projPath)
		if err != nil {
			die(stderr, fmt.Sprintf("read %s: %v", projPath, err))
		}
		pt := string(projBytes)
		newPT := pt

		if mode == "milestone" {
			tickedPT, n := project.TickMilestoneTaskRow(newPT, repoName, issueStr, f.Milestone)
			newPT = tickedPT
			if n > 0 {
				applied = append(applied, fmt.Sprintf("ticked [%s#%s %s] in %s", repoName, issueStr, f.Milestone, filepath.Base(projPath)))
			} else {
				cwarn(stderr, fmt.Sprintf("no task line for [%s#%s %s] in %s", repoName, issueStr, f.Milestone, filepath.Base(projPath)))
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
						anchor, filepath.Base(projPath), skel, refDef))
				}
				cwarn(stderr, fmt.Sprintf("--no-project (or --force): skipping detail-block update for <a id=\"%s\"> in %s", anchor, filepath.Base(projPath)))
			}
			if found {
				newPT = updated
				applied = append(applied, fmt.Sprintf("updated detail block <a id=\"%s\"> in %s", anchor, filepath.Base(projPath)))
			}
		} else { // issue close
			tickedPT, n := project.TickAllTaskRowsForIssue(newPT, repoName, issueStr)
			newPT = tickedPT
			if n > 0 {
				applied = append(applied, fmt.Sprintf("ticked %d remaining task line(s) for %s#%s in %s", n, repoName, issueStr, filepath.Base(projPath)))
			}
			if n > 1 {
				cwarn(stderr, fmt.Sprintf("multiple %s#%s task rows ticked at once — confirm individual milestones were genuinely closed (§5 step 1)", repoName, issueStr))
			}
		}

		if newPT != pt {
			projectEditPath = projPath
			projectEditText = newPT
		}
	}

	return closeResult{
		issuePath:       issuePath,
		issueText:       issueText,
		newIssueText:    newIssueText,
		projectEditPath: projectEditPath,
		projectEditText: projectEditText,
		fm:              fm,
		body:            body,
		repoName:        repoName,
		issueStr:        issueStr,
		today:           today,
		appliedMsgs:     applied,
	}
}

// printCloseDryRun prints what a close WOULD change, writing nothing (#139).
func printCloseDryRun(stderr io.Writer, r closeResult) {
	cinfo(stderr, "DRY=1 — no files written")
	fmt.Fprintf(os.Stdout, "Would update: %s\n", r.issuePath)
	if r.projectEditPath != "" {
		fmt.Fprintf(os.Stdout, "Would update: %s\n", r.projectEditPath)
	}
}

// applyClose performs the close's writes — issue + project files + the #117
// calibration ledger — then emits the success messages computeClose deferred
// (so "flipped → codecomplete" prints only when the flip actually happened). Called
// only after a finalizing verdict, or on the eager non-review path (#139).
func applyClose(stderr io.Writer, f *closeFlags, r closeResult) {
	if r.newIssueText != r.issueText {
		if err := os.WriteFile(r.issuePath, []byte(r.newIssueText), 0o644); err != nil {
			die(stderr, fmt.Sprintf("write %s: %v", r.issuePath, err))
		}
	}
	if r.projectEditPath != "" {
		if err := os.WriteFile(r.projectEditPath, []byte(r.projectEditText), 0o644); err != nil {
			die(stderr, fmt.Sprintf("write %s: %v", r.projectEditPath, err))
		}
	}
	for _, m := range r.appliedMsgs {
		cok(stderr, m)
	}
	// ── Close the loop (#117 mechanism 3) ────────────────────────────────────
	// On a full-issue close with a measured actual, append the estimate↔actual
	// data point to the calibration ledger. Milestone closes carry a partial
	// actual, so only the whole-issue close yields a clean row.
	if shouldLogCalibration(f) {
		appendCalibrationRow(stderr, f, r.fm, r.body, r.repoName, r.issueStr, r.today)
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
func runClose(stderr io.Writer, f *closeFlags) error {
	r := computeClose(stderr, f)
	if f.DryRun {
		printCloseDryRun(stderr, r)
		return nil
	}
	applyClose(stderr, f, r)
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
func appendCalibrationRow(stderr io.Writer, f *closeFlags, fm, body, repoName, issueStr, today string) {
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
		buf.Write(existing)
		if !strings.HasSuffix(string(existing), "\n") {
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
		applyClose(stderr, f, r)
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
		PlansDir:      envOr("WF_PLANS_DIR", "workshop/plans"),
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
	if err := withRequiredRepoTransactionLock(cmd, func() error {
		r = computeClose(stderr, f)
		base, baseLong, head = resolveReviewWindow(strconv.Itoa(f.Issue), "", "")
		snapshot = captureCloseReviewSnapshot(r)
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
		PlansDir:      envOr("WF_PLANS_DIR", "workshop/plans"),
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
	dispatchParams.PlansDir = "" // sidecar is a repo write; persist it after reacquiring the lock.
	review := dispatchBoundaryReview(stdout, stderr, dispatchParams)
	return withRequiredRepoTransactionLock(cmd, func() error {
		return finalizeBoundaryReview(stdout, stderr, f, r, review, p, snapshot.validate)
	})
}

func finalizeBoundaryReview(stdout, stderr io.Writer, f *closeFlags, r closeResult, review reviewResult, p boundaryReviewParams, validate func() error) error {
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
	switch closeVerdictOutcome(review.Verdict) {
	case closeFinalize:
		if validate != nil {
			if err := validate(); err != nil {
				emitTrailerBlock(stdout, review, kind)
				cwarn(stderr, fmt.Sprintf("boundary review: reviewed state changed while the lock was released — close NOT finalized: %v", err))
				cwarn(stderr, fmt.Sprintf("re-run `%s` so the review covers the current repo state", verb))
				return fmt.Errorf("boundary review stale: %w", err)
			}
		}
		applyClose(stderr, f, r)
		emitTrailerBlock(stdout, review, kind)
		if err := annotateLogLineWithVerdict(f.IssuesDir, f.Issue, f.Milestone, review.Verdict); err != nil {
			cwarn(stderr, fmt.Sprintf("log-line verdict annotation skipped: %v", err))
		}
		if f.Milestone == "" { // #160 Q4: lessons ping only at the whole-issue close boundary
			emitLessonsReminder(stdout)
		}
		return nil
	case closeRework:
		emitTrailerBlock(stdout, review, kind)
		cwarn(stderr, "boundary review: REWORK — close NOT finalized; issue left at status: working")
		cwarn(stderr, fmt.Sprintf("address the findings, then re-run `%s` (no --no-reclose-guard needed)", verb))
		return fmt.Errorf("boundary review verdict REWORK — close not finalized")
	default: // closeHalt
		emitTrailerBlock(stdout, review, kind)
		cwarn(stderr, fmt.Sprintf("boundary review verdict %q is UNEXPECTED — close NOT finalized; issue left at status: working", review.Verdict))
		cwarn(stderr, "the review produced no clear SHIP/FIX-THEN-SHIP/REWORK verdict (a gate/prompt bug?).")
		cwarn(stderr, "STOP: investigate the review output (sidecar) and consult a human before re-running.")
		return fmt.Errorf("boundary review verdict %q — unexpected; close not finalized, consult a human", review.Verdict)
	}
}

type closeReviewSnapshot struct {
	head      string
	issuePath string
	issueText string
}

func captureCloseReviewSnapshot(r closeResult) closeReviewSnapshot {
	return closeReviewSnapshot{
		head:      strings.TrimSpace(gitx.Capture("rev-parse", "HEAD")),
		issuePath: r.issuePath,
		issueText: r.issueText,
	}
}

func (s closeReviewSnapshot) validate() error {
	if s.head != "" {
		currentHead := strings.TrimSpace(gitx.Capture("rev-parse", "HEAD"))
		if currentHead == "" {
			return fmt.Errorf("cannot resolve HEAD")
		}
		if currentHead != s.head {
			return fmt.Errorf("HEAD changed from %s to %s", shortSHA(s.head), shortSHA(currentHead))
		}
	}
	if s.issuePath != "" {
		data, err := os.ReadFile(s.issuePath)
		if err != nil {
			return fmt.Errorf("read %s: %w", s.issuePath, err)
		}
		if string(data) != s.issueText {
			return fmt.Errorf("%s changed", s.issuePath)
		}
	}
	return nil
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

func explainActual(stderr io.Writer, issueStr, mode, milestone string) {
	repoTop, brainAbs := resolveActualRoots()

	var head []string
	head = append(head, fmt.Sprintf("%sACTUAL=<hours> required for %s close (§5 step 3).%s", ansiRed, mode, ansiReset), "")
	head = append(head, fmt.Sprintf("  %sSemantic:%s  focused dev-hours on this %s (#%s) — not wall-clock.", ansiCyan, ansiReset, mode, issueStr))
	head = append(head, "             sdlc computes it (below); method: 42shots/velocity/baseline-v3.md.", "")
	fmt.Fprintln(stderr, strings.Join(head, "\n"))

	// #68 M2: run v3 ourselves (brain + repo transcript dirs, window + peers) and
	// print the measured suggestion — the old "run this python command yourself"
	// prose, lifted into the binary. Same engine as `sdlc actual`.
	printActual(stderr, computeActual(repoTop, brainAbs, issueStr))

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

// checkActualDeviation is the thin IO glue: measure via the shared engine, run
// the pure comparator, and warn (to stderr) or return a refusal error. Returns
// nil — never blocks — when the engine can't measure (no window / telemetry gap
// / no script): an unavailable measurement must not gate a legitimate close.
func checkActualDeviation(stderr io.Writer, issueStr string, passed float64) error {
	repoTop, brainAbs := resolveActualRoots()
	res := computeActual(repoTop, brainAbs, issueStr)
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

// findMilestonesMissingVerdict enumerates milestones in the issue body's
// `## Plan` section and returns the tags of any whose close commit lacks
// a `Review-Verdict:` trailer.
//
// "Close commit" for milestone Mx = a commit whose subject opens with
// `#<issue> Mx:` AND whose message body contains a `Review-Verdict:`
// trailer line. The conjunctive `--all-match` over both --grep patterns
// matches the task spec exactly.
//
// Returns ([], nil) when every milestone has evidence. Returns ([], err)
// only on hard failures (issue body unparseable, git unavailable). A
// milestone whose subject doesn't match any commit is treated the same
// as one whose commit lacks the trailer — both are "no review evidence."
func findMilestonesMissingVerdict(body, issueStr, issuePath string) ([]string, error) {
	m := issue.PlanSectionRE.FindStringSubmatchIndex(body)
	if m == nil {
		// No plan section → no milestones to check. Treat as "fine":
		// the operator may be closing an issue that never had milestones.
		return nil, nil
	}
	planBody := body[m[2]:m[3]]
	matches := milestonePlanRE.FindAllStringSubmatch(planBody, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	// Preserve plan order; de-duplicate (a milestone may appear in the
	// plan more than once if revised).
	var ordered []string
	seen := map[string]bool{}
	for _, mm := range matches {
		tag := mm[1]
		if seen[tag] {
			continue
		}
		seen[tag] = true
		ordered = append(ordered, tag)
	}
	var missing []string
	for _, tag := range ordered {
		ok, err := milestoneHasVerdictCommit(issueStr, tag, issuePath)
		if err != nil {
			return nil, err
		}
		if !ok {
			missing = append(missing, tag)
		}
	}
	return missing, nil
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

// formatMissingVerdicts builds the next-action error message naming the
// milestones that lack Review-Verdict trailers. Pure: no IO, no exit.
// Lives next to explainMissingVerdicts so tests can assert the contract
// without subprocessing or os.Exit gymnastics.
func formatMissingVerdicts(issueStr string, missing []string) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("%smilestones %s lack Review-Verdict trailer in close commits (AGENTS.md §3).%s",
		ansiRed, strings.Join(missing, ", "), ansiReset))
	lines = append(lines, "")
	lines = append(lines, "  Each milestone close must carry a fresh-eyes review verdict in")
	lines = append(lines, "  the commit message. Without it, there's no evidence the work")
	lines = append(lines, "  was reviewed before the next milestone began.")
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %sNext actions:%s", ansiCyan, ansiReset))
	for _, tag := range missing {
		lines = append(lines, fmt.Sprintf("    sdlc judge milestone-review --issue %s --milestone %s", issueStr, tag))
	}
	lines = append(lines, "    # then amend the milestone-close commit (or land a new commit)")
	lines = append(lines, "    # with these trailers:")
	lines = append(lines, "    #   Review-Verdict: SHIP")
	lines = append(lines, "    #   Review-Window: <base>..<head>")
	lines = append(lines, "")
	lines = append(lines, "  Or pass --no-verdict (or --force); record the reason in --verified.")
	return strings.Join(lines, "\n")
}

func explainMissingVerdicts(stderr io.Writer, issueStr string, missing []string) {
	die(stderr, formatMissingVerdicts(issueStr, missing))
}
