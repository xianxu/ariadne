// close.go — `sdlc close` subcommand. Ports scripts/close-issue.py.
//
// Same posture as the Python source:
//   - Validates inputs (ISSUE required; ACTUAL + VERIFIED required unless
//     --no-actual / --no-verified / --force).
//   - Emits the semantic warmup on the first 2 invocations per shell session.
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
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/project"
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

	// Per-gate bypass flags (#67). Each waives exactly ONE of runClose's
	// guards; --force waives them all. The flag is an explicit acknowledgment
	// that the gate doesn't apply here (the rationale belongs in --verified) —
	// not a way to forget it. See skip().
	NoActual    bool
	NoVerified  bool
	NoReclose   bool
	NoAtlas     bool
	NoVerdict   bool
	NoPlanCheck bool
	NoProject   bool
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
	}
	return false
}

// NewCloseCmd returns the cobra command for `sdlc close`. The main session
// is responsible for registering it on the root command and for supplying
// the rich Long text via the embedded helptext package.
func NewCloseCmd() *cobra.Command {
	var f closeFlags

	cmd := &cobra.Command{
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
			return runClose(cmd.OutOrStderr(), &f)
		},
	}

	cmd.Flags().IntVar(&f.Issue, "issue", 0, "issue ID (numeric, required)")
	cmd.Flags().StringVar(&f.Milestone, "milestone", "", "milestone tag (e.g. M1, M4b); omit for full issue close")
	cmd.Flags().StringVar(&f.Actual, "actual", "", "focused dev-hours via v3 procedure")
	cmd.Flags().StringVar(&f.Verified, "verified", "", "one-line evidence the work meets done-when")
	cmd.Flags().BoolVar(&f.Force, "force", false, "bypass ALL gates (≡ every --no-* flag); record the reason in --verified")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print what would change; do not write")
	cmd.Flags().StringVar(&f.BrainDir, "brain-dir", "../brain", "path to the brain repo (for project-file lookup)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", "workshop/issues", "directory holding issue files")
	// Per-gate bypasses (#67) — each waives one guard; --force waives all.
	cmd.Flags().BoolVar(&f.NoActual, "no-actual", false, "bypass the ACTUAL-hours requirement (weakens velocity calibration)")
	cmd.Flags().BoolVar(&f.NoVerified, "no-verified", false, "bypass the VERIFIED-evidence requirement (only if there's no behavior to verify)")
	cmd.Flags().BoolVar(&f.NoReclose, "no-reclose-guard", false, "bypass the already-done refusal (intentionally re-close)")
	cmd.Flags().BoolVar(&f.NoAtlas, "no-atlas", false, "bypass the atlas/ change check (acknowledge: no new architectural surface)")
	cmd.Flags().BoolVar(&f.NoVerdict, "no-verdict", false, "bypass the milestone Review-Verdict trailer check")
	cmd.Flags().BoolVar(&f.NoPlanCheck, "no-plan-check", false, "bypass the unchecked-## Plan-items refusal")
	cmd.Flags().BoolVar(&f.NoProject, "no-project", false, "bypass the project detail-block update requirement")
	// Don't use MarkFlagRequired("issue"): cobra emits an uncolored,
	// differently-formatted error that conflicts with die()'s red prefix.
	// Validation lives in runClose so all error formatting flows through
	// one path. SilenceErrors keeps cobra from printing on top of us.
	cmd.SilenceErrors = true
	return cmd
}

// Terminal helpers (cinfo / cok / cwarn / die) + ANSI constants live in
// term.go alongside the other package-shared helpers. Moved out of
// close.go in M4 to deduplicate across the now-7 subcommands.

// ── warmup ───────────────────────────────────────────────────────────────────

const warmupThreshold = 2

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
		fmt.Sprintf("  %sACTUAL%s   = focused dev-hours, derived via the v3 procedure.", ansiCyan, ansiReset),
		"             Run active-time-v3.py over the issue's commit window",
		"             with --commit-weight 1.0; read the per-issue total.",
		"             See brain/data/life/42shots/velocity/baseline-v3.md.",
		"             Pass --no-actual (or --force) only if you genuinely cannot run the script",
		"             (e.g., wontfix issue with no commits) — record the reason.",
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
//   - If logLine carries a leading date (`- YYYY-MM-DD: …`) and the Log
//     section already has a matching `### YYYY-MM-DD` day header, the line is
//     filed directly *beneath* that header (top of the day's group) — so a
//     dated close line sits inside its day rather than orphaned above the
//     header (#66). Top-of-group, not bottom, preserves the newest-first
//     convention insertLogLine already uses at the section level.
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

func runClose(stderr io.Writer, f *closeFlags) error {
	printSemanticWarmup(stderr)

	if f.Issue <= 0 {
		die(stderr, fmt.Sprintf("--issue is required and must be positive (got %d)", f.Issue))
	}
	if f.Actual != "" {
		if _, err := strconv.ParseFloat(f.Actual, 64); err != nil {
			die(stderr, fmt.Sprintf("ACTUAL must be a number, got '%s'", f.Actual))
		}
	}
	issueStr := strconv.Itoa(f.Issue)
	issueID := fmt.Sprintf("%06d", f.Issue)
	mode := "issue"
	if f.Milestone != "" {
		mode = "milestone"
	}

	if f.Actual == "" {
		if !f.skip("actual") {
			explainActual(stderr, issueStr, mode, f.Milestone)
			os.Exit(1)
		}
		cwarn(stderr, "--no-actual (or --force): closing with NO actual_hours — this issue won't feed velocity calibration")
	}
	if f.Verified == "" {
		if !f.skip("verified") {
			explainVerified(stderr, issueStr, mode, f.Milestone, f.Actual)
			os.Exit(1)
		}
		cwarn(stderr, "--no-verified (or --force): closing with NO verification evidence — no behavior recorded as checked")
	}

	today := time.Now().Format("2006-01-02")

	// ── Locate issue file ───────────────────────────────────────────────────
	pattern := filepath.Join(f.IssuesDir, issueID+"-*.md")
	candidates, err := filepath.Glob(pattern)
	if err != nil {
		die(stderr, fmt.Sprintf("glob %s: %v", pattern, err))
	}
	sort.Strings(candidates)
	if len(candidates) == 0 {
		die(stderr, fmt.Sprintf("no issue file matches %s", pattern))
	}
	if len(candidates) > 1 {
		die(stderr, fmt.Sprintf("multiple issue files match: %v", candidates))
	}
	issuePath := candidates[0]
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

	if currentStatus, _ := issue.GetField(fm, "status"); mode == "issue" && currentStatus == "done" {
		if !f.skip("reclose") {
			die(stderr, fmt.Sprintf("%s#%s is already status: done — pass --no-reclose-guard (or --force) to re-close intentionally", repoName, issueStr))
		}
		cwarn(stderr, fmt.Sprintf("--no-reclose-guard (or --force): re-closing %s#%s (already done)", repoName, issueStr))
	}

	// ── Commit window + atlas check ─────────────────────────────────────────
	refSubject := "#" + issueStr
	if f.Milestone != "" {
		refSubject += " " + f.Milestone
	}
	entries, err := gitx.LogReverse()
	if err != nil {
		// `git log` failure is non-fatal — close-issue.py raises here too,
		// but we treat it as "no commit window" (warn) to match the
		// behavior when there's no history yet.
		cwarn(stderr, fmt.Sprintf("git log failed: %v", err))
		entries = nil
	}
	var firstSHA string
	matchingCount := 0
	for _, e := range entries {
		if strings.Contains(e.SHA+" "+e.Date+" "+e.Subject, refSubject) {
			if firstSHA == "" {
				firstSHA = e.SHA
			}
			matchingCount++
		}
	}
	if firstSHA != "" {
		plural := "s"
		if matchingCount == 1 {
			plural = ""
		}
		cinfo(stderr, fmt.Sprintf("commit window: %s → HEAD (%d commit%s reference '%s')",
			firstSHA[:8], matchingCount, plural, refSubject))
	} else {
		cwarn(stderr, fmt.Sprintf("no commits reference '%s' on this branch", refSubject))
	}

	if firstSHA != "" {
		diffFiles, _ := gitx.DiffNames(firstSHA+"^", "HEAD")
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
				explainNoAtlas(stderr, firstSHA, nonAtlas)
				os.Exit(1)
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
				os.Exit(1)
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
			cok(stderr, fmt.Sprintf("ticked %s in %s ## Plan", f.Milestone, filepath.Base(issuePath)))
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
		newFM = issue.SetField(newFM, "status", "done")
		if f.Actual != "" {
			newFM = issue.SetField(newFM, "actual_hours", f.Actual)
		}
		newFM = issue.SetField(newFM, "updated", today)
		msg := fmt.Sprintf("flipped %s → status: done", filepath.Base(issuePath))
		if f.Actual != "" {
			msg += fmt.Sprintf(", actual_hours: %s", f.Actual)
		}
		cok(stderr, msg)
	}

	if f.Verified != "" {
		logLine := fmt.Sprintf("- %s: closed", today)
		if f.Milestone != "" {
			logLine += " " + f.Milestone
		}
		logLine += " — " + f.Verified
		newBody = insertLogLine(newBody, logLine)
		cok(stderr, "appended verification line to ## Log")
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
				cok(stderr, fmt.Sprintf("ticked [%s#%s %s] in %s", repoName, issueStr, f.Milestone, filepath.Base(projPath)))
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
				cok(stderr, fmt.Sprintf("updated detail block <a id=\"%s\"> in %s", anchor, filepath.Base(projPath)))
			}
		} else { // issue close
			tickedPT, n := project.TickAllTaskRowsForIssue(newPT, repoName, issueStr)
			newPT = tickedPT
			if n > 0 {
				cok(stderr, fmt.Sprintf("ticked %d remaining task line(s) for %s#%s in %s", n, repoName, issueStr, filepath.Base(projPath)))
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

	// ── Write ───────────────────────────────────────────────────────────────
	if f.DryRun {
		cinfo(stderr, "DRY=1 — no files written")
		fmt.Fprintf(os.Stdout, "Would update: %s\n", issuePath)
		if projectEditPath != "" {
			fmt.Fprintf(os.Stdout, "Would update: %s\n", projectEditPath)
		}
		return nil
	}

	if newIssueText != issueText {
		if err := os.WriteFile(issuePath, []byte(newIssueText), 0o644); err != nil {
			die(stderr, fmt.Sprintf("write %s: %v", issuePath, err))
		}
	}
	if projectEditPath != "" {
		if err := os.WriteFile(projectEditPath, []byte(projectEditText), 0o644); err != nil {
			die(stderr, fmt.Sprintf("write %s: %v", projectEditPath, err))
		}
	}

	cok(stderr, "done — review with `git diff`, then commit")
	return nil
}

// ── explainers ───────────────────────────────────────────────────────────────

func explainActual(stderr io.Writer, issueStr, mode, milestone string) {
	cwd, _ := os.Getwd()
	repoDir, err := filepath.Abs(cwd)
	if err != nil {
		repoDir = cwd
	}
	repoSlug := filepath.Base(repoDir)
	transcriptSlugRepo := "-Users-xianxu-workspace-" + repoSlug
	transcriptSlugBrain := "-Users-xianxu-workspace-brain"

	firstSHA, firstTS, lastTS, _ := gitx.CommitWindow(issueStr)
	_ = firstSHA // not used in explainer

	var lines []string
	lines = append(lines, fmt.Sprintf("%sACTUAL=<hours> required for %s close (§5 step 3).%s", ansiRed, mode, ansiReset), "")
	lines = append(lines, fmt.Sprintf("  %sSemantic:%s  focused dev-hours spent on this %s (#%s).", ansiCyan, ansiReset, mode, issueStr))
	lines = append(lines, "             Not wall-clock; not 'hours since I created the issue.'")
	lines = append(lines, "             Method: v3 commit-anchored segment-local attribution.")
	lines = append(lines, "             See brain/data/life/42shots/velocity/baseline-v3.md.", "")

	if firstTS != "" && lastTS != "" {
		windowIssues, _ := gitx.DiscoverWindowIssues(firstTS, lastTS, issueStr)
		var issueFlags []string
		for _, n := range windowIssues {
			issueFlags = append(issueFlags, "--issue "+n)
		}
		lines = append(lines, fmt.Sprintf("  %sCompute via:%s", ansiCyan, ansiReset))
		lines = append(lines, "    python3 construct/local/issues/active-time-v3.py \\")
		lines = append(lines, fmt.Sprintf("      --dir ~/.claude/projects/%s \\", transcriptSlugRepo))
		lines = append(lines, fmt.Sprintf("      --dir ~/.claude/projects/%s \\", transcriptSlugBrain))
		lines = append(lines, fmt.Sprintf("      --git-repo %s \\", repoDir))
		lines = append(lines, fmt.Sprintf("      --since %s --until %s \\", firstTS, lastTS))
		lines = append(lines, fmt.Sprintf("      %s \\", strings.Join(issueFlags, " ")))
		lines = append(lines, "      --commit-weight 1.0 --threshold-min 15 --include-assistant", "")

		var peers []string
		for _, n := range windowIssues {
			if n != issueStr {
				peers = append(peers, n)
			}
		}
		if len(peers) > 0 {
			lines = append(lines, fmt.Sprintf("  Issues auto-discovered from #refs in window subjects: #%s + peers #%s.",
				issueStr, strings.Join(peers, ", #")))
			lines = append(lines, "  Why all of them: v3 anchors segments by commit-subject issue ref;")
			lines = append(lines, fmt.Sprintf("  unrecognized refs fall back to mention-fallback, inflating #%s by 3-10x.", issueStr))
			lines = append(lines, "  If a discovered peer looks unrelated to real work, drop its --issue flag.", "")
		}
		lines = append(lines, fmt.Sprintf("  The 'per-issue totals' line for #%s in the output is your ACTUAL.", issueStr))
		lines = append(lines, "  (Round to nearest 0.5; under 1 hr keep one decimal: 0.45 → 0.5.)")
	} else {
		lines = append(lines, fmt.Sprintf("  %sNo commits matching #%s found — compute hours by judgment%s", ansiYellow, issueStr, ansiReset))
		lines = append(lines, fmt.Sprintf("  %sor wait until commits land. Pass --no-actual (or --force) to bypass.%s", ansiYellow, ansiReset))
	}
	lines = append(lines, "")
	extra := ""
	if milestone != "" {
		extra = " --milestone " + milestone
	}
	lines = append(lines, fmt.Sprintf("  %sThen re-run:%s", ansiCyan, ansiReset))
	lines = append(lines, fmt.Sprintf("    sdlc close --issue %s%s --actual <hours> --verified '<evidence>'", issueStr, extra), "")
	lines = append(lines, "  Pass --no-actual (or --force) to bypass this requirement (record the reason in --verified).")
	fmt.Fprintln(stderr, strings.Join(lines, "\n"))
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
	extra := ""
	if milestone != "" {
		extra = " --milestone " + milestone
	}
	actualArg := " --actual <hours>"
	if actual != "" {
		actualArg = " --actual " + actual
	}
	lines = append(lines, fmt.Sprintf("  %sThen re-run:%s", ansiCyan, ansiReset))
	lines = append(lines, fmt.Sprintf("    sdlc close --issue %s%s%s --verified '<evidence>'", issueStr, extra, actualArg), "")
	lines = append(lines, "  Pass --no-verified (or --force) only if there's genuinely no behavior to verify.")
	fmt.Fprintln(stderr, strings.Join(lines, "\n"))
}

func explainNoAtlas(stderr io.Writer, firstSHA string, nonAtlas []string) {
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
	lines = append(lines, fmt.Sprintf("no atlas/ changes in %s..HEAD (§5 step 5).", firstSHA[:8]))
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
//	- [x] **M1 — scaffold …
//	- [ ] **M4b — port milestone-close
//	- [.] **M5 — wip
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

