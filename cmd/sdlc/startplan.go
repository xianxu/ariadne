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

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
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

	// #82 M3: a non-blocking base-contention heads-up. The symlink model means
	// every derivative reads the base's working tree live, so planning against a
	// branched / dirty / concurrently-worked base is the failure mode worth
	// surfacing at the commitment point. Emitted ONLY when cwd is the base repo
	// (you plan a base change while cd'd into it); silent in a derivative.
	if root, err := gitx.RepoTopLevel(); err == nil && isBaseRepo(root) {
		c := gatherBaseContention(root, issue)
		fmt.Fprintln(stdout)
		if c.Clean() {
			cok(stdout, baseContentionSummary(c))
		} else {
			cwarn(stdout, baseContentionSummary(c))
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

// issueRef renders a zero-padded id as "#83" (strips leading zeros; falls back
// to the raw id for a non-numeric / unreadable one).
func issueRef(id string) string {
	if n, err := strconv.Atoi(id); err == nil {
		return fmt.Sprintf("#%d", n)
	}
	return "#" + id
}

// isBaseRepo reports whether root is the base-layer source repo (ariadne)
// rather than a derivative. In the base, `construct/` is a real directory; in a
// derivative it's a symlink into the base (per construct/setup.sh). So a real,
// non-symlink `construct/` is the signal — the cwd==base assumption for #82 M3.
func isBaseRepo(root string) bool {
	info, err := os.Lstat(filepath.Join(root, "construct"))
	if err != nil {
		return false
	}
	return info.IsDir() && info.Mode()&os.ModeSymlink == 0
}

// gatherBaseContention is the thin IO seam that builds baseContention from the
// live repo: branch (gitx), dirty CODE count (worktreeDirty→assessDirty's
// Blocking bucket, so a dirty issue file is NOT counted, #82 M2), and other
// status:working base issues (listIssues, excluding the one being planned).
func gatherBaseContention(root string, thisIssue int) baseContention {
	issuesDir := envOr("WF_ISSUES_DIR", "workshop/issues")
	historyDir := envOr("WF_HISTORY_DIR", "workshop/history")
	c := baseContention{
		Repo:   filepath.Base(root),
		Branch: gitx.Capture("branch", "--show-current"),
	}
	if dirty, err := worktreeDirty(mergeRunner); err == nil {
		c.DirtyCode = len(assessDirty(dirty, issuesDir, historyDir).Blocking)
	}
	thisID := fmt.Sprintf("%06d", thisIssue)
	if issues, err := listIssues(issuesDir); err == nil {
		for _, is := range issues {
			if is.Status == "working" && is.ID != thisID {
				c.Others = append(c.Others, inFlightIssue{ID: is.ID, Title: is.Title})
			}
		}
	}
	return c
}
