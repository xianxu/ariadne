// startplan.go — `sdlc start-plan`: the workflow's planning-entry transition
// (#75 M2). The flow had `claim` (start work) and `change-code` (the plan-quality
// *review* gate, which is too late — the design is already made) but no marker
// for "I'm now designing". start-plan fills it: it delivers architecture.md's
// `at-plan` lens to the agent's main thread so the design accounts for the
// architectural principles from the start (the highest-leverage injection —
// architecture is decided here). It's the *forward* counterpart to change-code's
// plan-quality judge (the *backward* check), both consuming the one registry.
//
// Re-run it each time a new design begins: agents don't reread a static doc, so
// re-delivering keeps the principles live in attention.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// NewStartPlanCmd returns the cobra command for `sdlc start-plan`.
func NewStartPlanCmd() *cobra.Command {
	var issue int
	cmd := &cobra.Command{
		Use:           "start-plan",
		Short:         "Enter planning: deliver the architecture principles to design against (#75)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"start-plan\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			guardSpineRepo(cmd.ErrOrStderr()) // #176 lifecycle guard
			if issue > 0 {
				// Same resolution as guardSpineRepo: env override cwd-relative,
				// default anchored at repo top (correct from any subdirectory).
				issuesDir := os.Getenv("WF_ISSUES_DIR")
				if issuesDir == "" {
					if repoTop, err := gitx.RepoTopLevel(); err == nil {
						issuesDir = filepath.Join(repoTop, "workshop", "issues")
					} else {
						issuesDir = "workshop/issues"
					}
				}
				if path, err := locateIssueFile(issuesDir, issue); err == nil {
					guardIssueNotDone(cmd.ErrOrStderr(), path, strconv.Itoa(issue)) // #176 done-issue guard
				}
			}
			runStartPlan(cmd.OutOrStdout(), issue)
			return nil
		},
	}
	cmd.Flags().IntVar(&issue, "issue", 0, "issue being planned (optional, for the label)")
	return cmd
}

// runStartPlan emits the planning framing + the at-plan architecture lens.
func runStartPlan(stdout io.Writer, issue int) {
	label := "this issue"
	if issue > 0 {
		label = fmt.Sprintf("#%d", issue)
	}
	cinfo(stdout, fmt.Sprintf("Entering planning for %s. Design with these architectural", label))
	fmt.Fprintln(stdout, "    principles in mind — the plan-quality gate (`sdlc change-code`) checks the")
	fmt.Fprintln(stdout, "    plan against them, and the boundary review checks the code. Cite ARCH-* in")
	fmt.Fprintln(stdout, "    your plan where a principle shaped a decision.")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, judge.ArchitectureBlock("at-plan"))

	// #72: HOW/WHERE to capture the design — the durable-plan pointer. Sits
	// between the architecture block (WHAT to design against) and the contention
	// read (the environmental epilogue): framing + principles + this line form one
	// "how to plan" unit. Continuation lines indent 4 to align under cinfo's `==> `.
	fmt.Fprintln(stdout)
	cinfo(stdout, planPointer(issue))

	// #113: a non-blocking estimate nudge. The estimate gate moved
	// claim → change-code, and start-plan is where it's naturally set (post-
	// design, scope knowable). Remind the operator to set estimate_hours now so
	// change-code's gate passes; acknowledge it when already present. Best-effort
	// — a missing/unreadable issue file just skips the line.
	if issue > 0 {
		if est, err := issueEstimate(issue); err == nil {
			fmt.Fprintln(stdout)
			cinfo(stdout, estimateNudge(est))
		}
	}

	// #134: point at the estimator SOURCE, not just the field. The nudge says SET
	// estimate_hours; this says DERIVE it against the calibrated source (named +
	// status-tagged) so the agent doesn't satisfy the block grammar from memory.
	// Best-effort + warn-and-continue: resolves the default brain layout (or the
	// $WF_ESTIMATOR_SRC override) and renders for every status, MISSING included —
	// a brain-less downstream repo gets the pointer, never a break.
	fmt.Fprintln(stdout)
	brainAbs, _ := filepath.Abs("../brain")
	src := estimateSourceStatus(brainAbs, estimate.CurrentModel(), os.Getenv("WF_ESTIMATOR_SRC"))
	cinfo(stdout, estimate.SourceLine(src))

	// #82 M3 / #83: a non-blocking heads-up on the DEPENDENCY PATH. The symlink
	// model means a repo reads ALL its transitive upstreams' working trees live,
	// so the "moving ground" you plan against is every repo this one depends on
	// (declared in construct/deps), not a single base. Walk that chain and report
	// each upstream's contention. The root (ariadne — no upstream) has an empty
	// chain, so it reports its own concurrent work (the one case #82 M3 got right).
	if root, err := gitx.RepoTopLevel(); err == nil {
		chain := substrateChain(root)
		if len(chain) == 0 {
			chain = []string{root}
		}
		fmt.Fprintln(stdout)
		for _, up := range chain {
			c := gatherBaseContention(up, issue)
			if c.Clean() {
				cok(stdout, baseContentionSummary(c))
			} else {
				cwarn(stdout, baseContentionSummary(c))
			}
		}
	}
}

// inFlightIssue is a base issue another session has claimed (status: working).
type inFlightIssue struct {
	ID    string // zero-padded, e.g. "000083"
	Title string
}

// baseContention is the pure input to the start-plan heads-up: the base repo's
// name, its current branch, how many uncommitted CODE files are dirty (tracker
// files excluded — they're not contention, #82 M2), and the other in-flight base
// issues. Built by the thin gatherBaseContention seam; consumed by the pure
// baseContentionSummary so the wording is table-testable without git/IO.
type baseContention struct {
	Repo      string
	Branch    string
	DirtyCode int
	Others    []inFlightIssue
}

// Clean reports whether the base is a calm place to plan against: on `main`, no
// dirty code, no other claimed base issues.
func (c baseContention) Clean() bool {
	return c.Branch == "main" && c.DirtyCode == 0 && len(c.Others) == 0
}

// baseContentionSummary renders the one-line heads-up. Pure.
func baseContentionSummary(c baseContention) string {
	if c.Clean() {
		return fmt.Sprintf("base (%s): clean main, no other base issues in flight — clear to plan.", c.Repo)
	}
	var parts []string
	switch {
	case c.Branch == "":
		parts = append(parts, "detached HEAD")
	case c.Branch != "main":
		parts = append(parts, fmt.Sprintf("on branch `%s`", c.Branch))
	}
	if c.DirtyCode > 0 {
		parts = append(parts, fmt.Sprintf("%d uncommitted code file(s)", c.DirtyCode))
	}
	if n := len(c.Others); n > 0 {
		refs := make([]string, 0, n)
		for _, o := range c.Others {
			refs = append(refs, issueRef(o.ID))
		}
		parts = append(parts, fmt.Sprintf("%d other issue(s) in-flight (%s)", n, strings.Join(refs, ", ")))
	}
	return fmt.Sprintf("base (%s): %s — planning against a moving base.", c.Repo, strings.Join(parts, "; "))
}

// planPointer renders the durable-plan reminder the planning gate prints (#72):
// the canonical plan lands in workshop/plans/ (version-controlled), authored via
// the superpowers-writing-plans skill — NOT the harness builtin's ephemeral
// ~/.claude/plans/ file. Pure: the only input is the issue id (for the slug), so
// the wording is table-testable without IO. Stays agent-agnostic — it names the
// skill + repo location, never teaching the binary the Claude-specific path.
// The two continuation lines indent 4 to align under cinfo's `==> ` prefix.
func planPointer(issue int) string {
	slug := "NNNNNN-slug"
	if issue > 0 {
		slug = fmt.Sprintf("%06d-slug", issue)
	}
	return fmt.Sprintf("Capture the plan via the superpowers-writing-plans skill →\n"+
		"    workshop/plans/%s-plan.md (version-controlled). The builtin plan-mode\n"+
		"    file (~/.claude/plans/…) is ephemeral — NOT the record.", slug)
}

// estimateNudge renders the start-plan reminder about estimate_hours (#113, retimed by
// #187): when absent, it tells you NOT to derive it yet — change-code asks only after the
// plan clears plan-quality — or a one-line acknowledgment when already present. Pure: the only input
// is the current estimate value (empty when unset), so the wording is
// table-testable without IO. Continuation lines indent 4 to align under
// cinfo's `==> ` prefix.
func estimateNudge(estimate string) string {
	est := strings.TrimSpace(estimate)
	if est == "" {
		return "Don't derive `estimate_hours:` yet — `sdlc change-code` runs plan-quality\n" +
			"    FIRST and asks for the estimate only after the plan clears (#187). Costing a\n" +
			"    plan nobody has accepted just gets recomputed on the next revision."
	}
	return fmt.Sprintf("estimate_hours: %s already set — change-code's estimate gate will pass.", est)
}

// issueEstimate reads the current estimate_hours value of issue <id> (empty when
// unset). The thin IO seam behind start-plan's estimate nudge; mirrors
// issueStatus's read-parse-getfield shape.
func issueEstimate(issueID int) (string, error) {
	issuesDir := envOr("WF_ISSUES_DIR", "workshop/issues")
	path, err := locateIssueFile(issuesDir, issueID)
	if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	fm, _, err := issue.Parse(string(raw))
	if err != nil {
		return "", err
	}
	v, _ := issue.GetField(fm, "estimate_hours")
	return v, nil
}

// issueRef renders a zero-padded id as "#83" (strips leading zeros; falls back
// to the raw id for a non-numeric / unreadable one).
func issueRef(id string) string {
	if n, err := strconv.Atoi(id); err == nil {
		return fmt.Sprintf("#%d", n)
	}
	return "#" + id
}

// parseSubstrateTargets extracts the `substrate <path>` targets from a
// construct/deps file's content, as written (raw, unresolved). Pure.
//
// CANONICAL GRAMMAR: construct/scripts/lib-deps.sh `deps_substrate_targets`
// (and its bootstrap.sh inline twin) — keep this in lockstep with that parse:
// strip a `#` comment to end-of-line, whitespace word-split, keep rows whose
// first field is `substrate` and that have ≥2 fields (blank / comment /
// malformed / `data` rows are dropped). #60 made construct/deps the sole
// substrate-graph carrier; this is its third reader (Go can't source the shell),
// so the edge cases below are mirrored in TestParseSubstrateTargets.
func parseSubstrateTargets(depsContent string) []string {
	var out []string
	for _, line := range strings.Split(depsContent, "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "substrate" {
			out = append(out, fields[1])
		}
	}
	return out
}

// substrateChain walks construct/deps transitively from root and returns the
// ordered, deduped list of upstream repo roots that exist on disk — the
// dependency path the caller reads live. root itself is excluded. Targets are
// resolved relative to their DECLARING repo root (the exact resolution #82 M3
// got wrong), so a 2-hop chain resolves each hop against the right base. Absent
// peers are skipped (present-walker semantics, like lib-deps.sh — don't abort on
// a missing sibling). The seen-set also guards against dependency cycles.
func substrateChain(root string) []string {
	var order []string
	seen := map[string]bool{}
	// Seed the cycle-guard with the root's CANONICAL key (EvalSymlinks, matching
	// how upstream entries are keyed below) so a cycle resolving back to root via
	// a symlinked path still matches and can't re-enter walk(root).
	rootKey := filepath.Clean(root)
	if real, err := filepath.EvalSymlinks(rootKey); err == nil {
		rootKey = real
	} else if abs, err := filepath.Abs(rootKey); err == nil {
		rootKey = abs
	}
	seen[rootKey] = true
	var walk func(declRoot string)
	walk = func(declRoot string) {
		data, err := os.ReadFile(filepath.Join(declRoot, "construct", "deps"))
		if err != nil {
			return
		}
		for _, target := range parseSubstrateTargets(string(data)) {
			resolved := target
			if !filepath.IsAbs(resolved) {
				resolved = filepath.Join(declRoot, target)
			}
			resolved = filepath.Clean(resolved)
			if real, err := filepath.EvalSymlinks(resolved); err == nil {
				resolved = real // canonicalize when the peer exists
			}
			if seen[resolved] {
				continue
			}
			if info, err := os.Stat(resolved); err != nil || !info.IsDir() {
				continue // absent peer → skip, keep walking the rest
			}
			seen[resolved] = true
			order = append(order, resolved)
			walk(resolved)
		}
	}
	walk(root)
	return order
}

// gatherBaseContention is the thin IO seam that builds baseContention for ONE
// repo root: branch + dirty CODE count (status --porcelain → assessDirty's
// Blocking bucket, so a dirty tracker file is NOT counted, #82 M2) read via
// `git -C root`, plus other status:working issues in that root's tracker
// (excluding the one being planned). Run per repo on the dependency path.
func gatherBaseContention(root string, excludeIssue int) baseContention {
	issuesDir := envOr("WF_ISSUES_DIR", "workshop/issues")
	historyDir := envOr("WF_HISTORY_DIR", "workshop/history")
	c := baseContention{Repo: filepath.Base(root)}
	// GitInDir output carries a trailing newline (unlike gitx.Capture, which
	// trims) — TrimSpace, or Branch never equals "main" and Clean() never fires.
	if out, err := mergeRunner.GitInDir(root, "branch", "--show-current"); err == nil {
		c.Branch = strings.TrimSpace(string(out))
	}
	if out, err := mergeRunner.GitInDir(root, "status", "--porcelain"); err == nil {
		c.DirtyCode = len(assessDirty(strings.TrimSpace(string(out)), issuesDir, historyDir).Blocking)
	}
	excludeID := fmt.Sprintf("%06d", excludeIssue)
	if issues, err := listIssues(filepath.Join(root, issuesDir)); err == nil {
		for _, is := range issues {
			// #122 carve-out: in-flight = "working" specifically (the contention warning
			// is about actively-worked peers; blocked is waiting) — not a category test.
			if is.Status == "working" && is.ID != excludeID {
				c.Others = append(c.Others, inFlightIssue{ID: is.ID, Title: is.Title})
			}
		}
	}
	return c
}
