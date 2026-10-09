// setstatus.go — `sdlc issue set-status --issue N <status>` subcommand
// (#56 M2: relocated under `issue`; flat `sdlc set-status` kept as a hidden
// deprecated alias built from the same NewSetStatusCmd).
//
// New verb (no Makefile equivalent today). Flips an issue file's
// status: frontmatter field with transition guards that match the
// xx-issues skill's contract:
//
//   - status → done routes to `sdlc close` (refused here so the
//     close-issue contract — ACTUAL + VERIFIED + atlas check — runs)
//   - done → anything-not-done (reopen) requires a fresh Log entry
//     dated today
//
// (#113) → working no longer requires estimate_hours — the estimate gate
// moved to `sdlc change-code`, so claiming early stays cheap.
//
// Each guard is bypassable with --force; the rationale belongs in
// the operator's commit message / log entry.
package main

import (
	"context"
	"fmt"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// The status SET and CATEGORIES now live in the vocabulary model
// (construct/vocabulary/issue.cue) — read them via vocab.Issue() (#122). The
// specific-state guards below keep literal status names on purpose: they encode
// ONE state's policy (done's close gate, the reopen edge), not category membership.

// setStatusFlags holds the parsed flag values for the set-status subcommand.
type setStatusFlags struct {
	Issue     int
	Status    string // positional arg
	Force     bool
	DryRun    bool
	IssuesDir string
}

// NewSetStatusCmd builds the set-status cobra command. Called twice:
// once under the `issue` group (`sdlc issue set-status`) and once as the
// hidden deprecated flat alias (`sdlc set-status`) — fresh instances, so
// no shared-pointer aliasing. Use is "set-status" with a dash.
func NewSetStatusCmd() *cobra.Command {
	f := setStatusFlags{}
	cmd := markMutatingCommand(&cobra.Command{
		Use:           "set-status <status>",
		Short:         "Flip an issue's status: with transition guards",
		Long:          "Placeholder — replaced by helptext.MustGet(\"set-status\") in main.go.",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f.Status = args[0]
			return runSetStatus(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "ariadne workshop issue ID (required)")
	cmd.Flags().BoolVar(&f.Force, "force", false, "bypass transition guards")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print what would change; do not write")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue details")
	return cmd
}

// runSetStatus changes the card's status through the transition guards (#252).
// The guards read the card's current status and, for a reopen, this checkout's
// details Log. codecomplete and done stay owned by close and merge.
func runSetStatus(ctx context.Context, stdout, stderr io.Writer, f *setStatusFlags) error {
	if tracked, err := repositoryTracked(ctx, f.IssuesDir); err != nil {
		return err
	} else if !tracked {
		return runLegacySetStatus(stdout, stderr, f)
	}
	var prev string
	decide := func(env *trackerEnv, card tracker.Record, body string) ([]byte, error) {
		// #277: entering working records who holds the work, as claim does —
		// resolved from the setter's own tracker environment.
		var me *issue.Claimant
		if f.Status == "working" {
			c, err := claimantIdentity(env)
			if err != nil {
				return nil, err
			}
			me = &c
		}
		next, p, err := statusDecision(card.Raw, body, f.Status, f.Force, time.Now().Format("2006-01-02"), startedClock(), me)
		prev = p
		return next, err
	}
	// #286: reopening abandoned work restores its branch first; the card
	// write below then clears the record, and the archive ref goes last.
	restored, err := reopenAbandoned(ctx, stderr, f)
	if err != nil {
		return err
	}
	if err := runCardUpdate(ctx, stdout, stderr, f.IssuesDir, f.Issue, "status → "+f.Status, f.DryRun, decide); err != nil {
		return err
	}
	if restored != nil {
		finishReopen(restored, stderr, f.IssuesDir)
	}
	// #122 M4: when --force masked the lifecycle gate on an illegal transition,
	// log the override — the escape hatch is explicit and recorded, not silent.
	if f.Force && prev != "" && prev != f.Status && !vocab.Issue().CanTransition(prev, f.Status) {
		cwarn(stderr, fmt.Sprintf("--force: overrode illegal transition %s → %s (not in the lifecycle)", prev, f.Status))
	}
	return nil
}

// statusDecision is set-status's pure core: guard the transition (unless
// forced), then set status/updated and stamp `started` on open → working
// without ever moving an existing stamp (#116). Entering working with an
// identity (#277) records it as the claimant — refused when another workspace
// owns the card, even under --force: reassignment is `sdlc reclaim` (#278).
// The `start` edge is guarded `owned` (#283): an unowned open card is claimed
// first, so starting never skips claim's readiness check (--force waives it,
// as it does every named guard; a foreign owner is never waivable).
func statusDecision(card []byte, detailsBody, next string, force bool, today, started string, me *issue.Claimant) ([]byte, string, error) {
	if !isValidStatus(next) {
		return nil, "", fmt.Errorf("invalid status %q (valid: %s)", next, strings.Join(vocab.Issue().AllStatuses(), ", "))
	}
	fm, _, err := issue.Parse(string(card))
	if err != nil {
		return nil, "", err
	}
	prev, _ := issue.GetField(fm, "status")
	// #286: started work ends through `sdlc abandon`, which keeps it under an
	// archive ref; --force cannot skip that (the work would be lost).
	if _, ending := abandonEvent[next]; ending && !vocab.Issue().IsOpen(prev) && !vocab.Issue().IsTerminal(prev) {
		return nil, prev, fmt.Errorf("refusing %s → %s: started work ends with `sdlc abandon --issue N --as %s --reason …`, which keeps it under an archive ref (#286); --force does not apply", prev, next, next)
	}
	if !force {
		if err := checkTransitionGuards(prev, next, fm, detailsBody); err != nil {
			return nil, prev, err
		}
	}
	// #286 BR-13: abandoned work with a kept branch leaves its terminal status
	// only by the reopen that restores it; anything else would orphan the ref.
	rec, abandoned, err := issue.CardAbandoned(card)
	if err != nil {
		return nil, prev, err
	}
	if abandoned && rec.Started() && vocab.Issue().IsTerminal(prev) && next != prev && next != "working" {
		return nil, prev, fmt.Errorf("refusing %s → %s: the work is kept at %s; reopen with `set-status working`, which restores it (--force does not apply)", prev, next, rec.Ref)
	}
	out, err := issue.SetCardField(card, "status", next)
	if err == nil && abandoned && vocab.Issue().IsTerminal(prev) && next == "working" {
		out, err = issue.SetCardAbandoned(out, nil) // the reopen ends the abandon record
	}
	if err == nil {
		out, err = issue.SetCardField(out, "updated", today)
	}
	if err == nil && vocab.Issue().IsOpen(prev) && next == "working" {
		if cur, _ := issue.GetField(fm, "started"); strings.TrimSpace(cur) == "" {
			out, err = issue.SetCardField(out, "started", started)
		}
	}
	if err == nil && next == "working" && me != nil {
		recorded, has, cerr := issue.CardClaimant(card)
		if cerr != nil {
			return nil, prev, cerr
		}
		if has && issue.MatchClaimant(&recorded, *me) == issue.OwnershipForeign {
			return nil, prev, fmt.Errorf("owned by %s; entering working from another workspace is a takeover — operator-directed `sdlc reclaim`, not set-status", describeClaimant(recorded))
		}
		if !has && prev == "working" {
			return nil, prev, fmt.Errorf("already working with no owner; `sdlc claim` takes it over, not set-status")
		}
		// #283: the model's `start` edge is guarded `owned`. An unowned open card
		// is claimed first — claim checks the details are on main — then started.
		if tr := vocab.Issue().TransitionFor(prev, next); !force && !has && tr != nil && tr.Event == "start" {
			return nil, prev, fmt.Errorf("unowned; `sdlc claim --issue N` takes the lock first, then `sdlc start-plan` (or this set-status) starts it")
		}
		out, err = issue.SetCardClaimant(out, *me)
	}
	return out, prev, err
}

// locateIssueFile resolves issue <id> to its single workshop/issues file,
// erroring on zero or multiple matches. Shared by set-status and claim so
// the NNNNNN-*.md glob convention lives in one place.
func locateIssueFile(issuesDir string, issueID int) (string, error) {
	if issueID <= 0 {
		return "", fmt.Errorf("--issue is required and must be positive (got %d)", issueID)
	}
	id := fmt.Sprintf("%06d", issueID)
	// Root-anchored: a relative issuesDir globbed against the process cwd, so
	// `claim --issue N` died from any subdirectory (#207 BR-16). Resolved here
	// rather than threaded, because set-status's call chain is outside this
	// issue's scope — the one deliberate exception to resolve-at-dispatch.
	dir := issuesDir
	if !filepath.IsAbs(dir) {
		if root, rerr := gitx.RepoTopLevel(); rerr == nil {
			dir = filepath.Join(root, dir)
		}
	}
	matches, err := filepath.Glob(filepath.Join(dir, id+"-*.md"))
	if err != nil {
		return "", fmt.Errorf("glob: %v", err)
	}
	sort.Strings(matches)
	if len(matches) == 0 {
		return "", fmt.Errorf("no issue file matches %s/%s-*.md", issuesDir, id)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("multiple issue files match: %v", matches)
	}
	return matches[0], nil
}

// issueStatus returns the current status: value of issue <id> (empty when
// unset). The read-only peek `sdlc claim` uses to gate its auto start-flip
// on the open→working transition only.
func issueStatus(issuesDir string, issueID int) (string, error) {
	path, err := locateIssueFile(issuesDir, issueID)
	if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %v", path, err)
	}
	fm, _, err := issue.Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parse frontmatter from %s: %v", path, err)
	}
	s, _ := issue.GetField(fm, "status")
	return s, nil
}

// applyStatus locates issue <id>, enforces the transition guards (unless
// force), and rewrites its status: + updated: frontmatter. On dryRun it
// computes the change without writing. Returns the file path, the previous
// status, and whether the content would change.
//
// Extracted from runSetStatus so `sdlc claim` can fold the → working flip
// into its sync (AGENTS.md §0: one command to claim + start work) without
// duplicating the guard logic or the frontmatter rewrite. Returns errors
// rather than die()-ing so callers compose it; runSetStatus translates the
// error back into a top-level die().
// startedClock is the injectable clock for the #116 `started:` stamp. New code
// injects the clock instead of calling time.Now() directly (controllable-time).
// Local-offset RFC3339 matches git's %aI author-date format that windowStart
// compares against lexically; tests override it for determinism.
var startedClock = func() string { return time.Now().Format(time.RFC3339) }

// ── transition guards ────────────────────────────────────────────────────────

// checkTransitionGuards enforces the xx-issues skill's status-transition
// contract. Returns nil if the transition is allowed (or --force is the
// caller's responsibility). Returns an error describing the refusal
// otherwise — message is the exact text presented to the operator.
func checkTransitionGuards(current, next, fm, body string) error {
	// Guard 0 (#122 M4): the lifecycle graph. Refuse a transition the model
	// (construct/vocabulary/issue.cue) doesn't declare — operator chose to enforce
	// (decision b). Skip the no-op (current==next) and an unset/initial status
	// (current=="") — neither is a real transition. --force bypasses, since this
	// whole function runs only when !force. Runs first so open→done reads as
	// "illegal" (it is) rather than the →done close-routing below.
	if current != "" && current != next && !vocab.Issue().CanTransition(current, next) {
		legal := vocab.Issue().LegalTransitions(current)
		if len(legal) == 0 {
			return fmt.Errorf("illegal transition %s → %s: %q is a dead-end in the lifecycle. Pass --force to override (logged).", current, next, current)
		}
		return fmt.Errorf("illegal transition %s → %s; legal from %q: %s. Pass --force to override (logged).",
			current, next, current, strings.Join(legal, ", "))
	}

	// Guard 1: → done routes through the publish flow. Always refused (done is
	// reached by `sdlc close` → codecomplete → `sdlc merge`/`push` → done, #160;
	// the flip carries ACTUAL + VERIFIED + atlas + the reviewed-HEAD-unchanged
	// invariant — none of which set-status enforces). Reached only from
	// `codecomplete` (the sole legal `→ done` edge); `working → done` is refused
	// earlier by Guard 0 as illegal. #122 carve-out: literal "done" value-specific.
	if next == "done" {
		return fmt.Errorf(
			"refusing to flip → done directly; done is set by the publish flow:\n" +
				"  sdlc close --issue <N> --verified '<evidence>'   # → codecomplete (local acceptance review)\n" +
				"  sdlc merge   (or sdlc push)                      # codecomplete → done (deterministic publish)\n" +
				"(#160: close runs the local review, merge/push do the reviewed-HEAD-unchanged publish flip)")
	}

	// Guard 1b (#160): → codecomplete routes to `sdlc close`. Always refused —
	// `close` is the ONLY writer of `codecomplete` (it runs the boundary review),
	// which is exactly what makes the commit carrying `status: codecomplete` a
	// trustworthy anchor for merge's reviewed-HEAD-unchanged invariant. Letting
	// set-status write it would forge that anchor. Value-specific, like `done`.
	if next == "codecomplete" {
		return fmt.Errorf(
			"refusing to flip → codecomplete directly; use:\n" +
				"  sdlc close --issue <N> --verified '<evidence>'\n" +
				"(#160: codecomplete is written ONLY by `sdlc close` after its boundary review —\n" +
				" that's what makes the merge-time reviewed-HEAD-unchanged invariant trustworthy)")
	}

	// (#113) No estimate guard here. `→ working` used to require
	// estimate_hours, which made starting work early demand a premature number.
	// The estimate gate moved to `sdlc change-code` (issue.CheckEstimate), the
	// universal implementation gate. Starting work — start-plan, or
	// `set-status working` — is estimate-free; claim never moves status (#283).

	// Guard 2: reopen (done → not-done) requires a fresh Log entry
	// dated today. The xx-issues skill puts the reason for reopening
	// in that entry. (#122 carve-out: "done" value-specific — the reopen-from-done edge.)
	if current == "done" && next != "done" {
		today := time.Now().Format("2006-01-02")
		if !logHasEntryToday(body, today) {
			return fmt.Errorf(
				"refusing to reopen (done → %s) without a fresh Log entry.\n"+
					"  Add an entry dated %s under ## Log explaining the reopen:\n"+
					"    - %s: reopened — <reason>\n"+
					"  (Reopens carry a rationale; the log is where it lands.)",
				next, today, today)
		}
	}

	return nil
}

// logHasEntryToday returns true if the body's ## Log section contains a
// line that starts with today's date (loose check: "- 2026-05-25: ...",
// "### 2026-05-25", or simply containing today's date string after the
// ## Log header line).
//
// Fence-aware since #211 M2. This used to take the FIRST `^## Log` and run to
// the next `\n## `, both of which matched inside fenced code blocks — so on an
// issue that quotes a log format (which issues here do; the deliverable is often
// a markdown document) the guard read a quoted example instead of the real Log.
// That was live: workshop/history/issues/000066-*.md has its first `## Log` at
// line 22 inside a fence and the real one at line 68.
func logHasEntryToday(body, today string) bool {
	section, ok := issue.SectionBody(body, "Log")
	if !ok {
		return false
	}
	// StripFenced because the real Log section can quote its own format — a
	// bounded search is not a fence-aware one (#211 M2 review BR-11).
	return strings.Contains(issue.StripFenced(section), today)
}

// isValidStatus returns whether s is a recognized status value (the set is the
// vocabulary model's, not a local list — #122).
func isValidStatus(s string) bool {
	for _, v := range vocab.Issue().AllStatuses() {
		if s == v {
			return true
		}
	}
	return false
}

// reopenedAbandon is a restore that ran before set-status's card write.
type reopenedAbandon struct {
	env          *trackerEnv
	rec          issue.Abandoned
	id, cardPath string
}

// reopenAbandoned restores an abandoned issue's branch when set-status
// reopens it (→ working from wontfix/punt with a kept branch); nil otherwise.
func reopenAbandoned(ctx context.Context, stderr io.Writer, f *setStatusFlags) (*reopenedAbandon, error) {
	if f.Status != "working" || f.DryRun {
		return nil, nil
	}
	env, err := openTracker(ctx)
	if err != nil {
		return nil, err
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		return nil, err
	}
	card, err := snap.Require(fmt.Sprintf("%06d", f.Issue))
	if err != nil {
		return nil, err
	}
	status, _ := issue.GetField(card.Card.Frontmatter, "status")
	if status == "working" {
		// BR-12: a rerun after the reopen's card write (a lost response, an
		// interruption) finishes it: on the restored branch, with the archive
		// ref still holding work the branch contains.
		return unfinishedReopen(env, card)
	}
	rec, ok, err := issue.CardAbandoned(card.Raw)
	if err != nil || !ok || !rec.Started() || !vocab.Issue().IsTerminal(status) {
		return nil, err
	}
	if err := restoreAbandoned(env, stderr, card, rec, f.IssuesDir); err != nil {
		return nil, err
	}
	return &reopenedAbandon{env: env, rec: rec, id: card.ID, cardPath: card.Path}, nil
}

// unfinishedReopen recognises a reopen whose card write landed but whose
// finish did not: the archive ref still exists and this checkout's branch
// contains it. nil when there is nothing to finish.
func unfinishedReopen(env *trackerEnv, card tracker.Record) (*reopenedAbandon, error) {
	if env.branch == "" || env.onRest() {
		return nil, nil
	}
	ref := issue.AbandonedRef(card.ID)
	tip, err := remoteRefTip(env.git, env.target.Remote, ref)
	if err != nil || tip == "" {
		return nil, err
	}
	if have, err := env.gitTest("cat-file", "-e", tip+"^{commit}"); err != nil || !have {
		return nil, err
	}
	if kept, err := env.ancestorOf(tip, "HEAD"); err != nil || !kept {
		return nil, err
	}
	return &reopenedAbandon{env: env, rec: issue.Abandoned{Ref: ref, Branch: env.branch, Head: tip}, id: card.ID, cardPath: card.Path}, nil
}
