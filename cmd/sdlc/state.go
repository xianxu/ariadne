// state.go — `sdlc state` subcommand. Read-only inspection of SDLC
// workflow state for the current repo. The compaction-recovery surface:
// after a session resume, an agent runs `sdlc state` instead of re-
// inferring from issue files.
//
// Scope (per workshop/issues/000031 M2):
//   - Issues in workshop/issues/ with status + plan progress
//   - Active git worktrees
//   - Recent commits on the current branch (main..HEAD)
//   - Structural drift checks (warn-only)
//
// No mutations. All mutating verbs (close, set-status, milestone-close)
// live elsewhere and funnel through internal/issue. This separation is
// the funnel-mutations-through-binary discipline that lets state surface
// drift instead of obscuring it.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// ── flag struct ──────────────────────────────────────────────────────────────

type stateFlags struct {
	JSON            bool
	IssuesDir       string
	HistoryDir      string
	IssuesExplicit  bool
	HistoryExplicit bool
}

// ── public types (also the JSON schema) ──────────────────────────────────────

// IssueState is one row in `sdlc state`'s issues section. Field names are
// the JSON keys; rename with care, downstream tooling may grep them.
type IssueState struct {
	ID         string `json:"id"`
	Path       string `json:"path"`
	Status     string `json:"status"`
	Title      string `json:"title,omitempty"`
	PlanTotal  int    `json:"plan_total"`
	PlanTicked int    `json:"plan_ticked"`
	Updated    string `json:"updated,omitempty"`
	// CardOnly: the card exists but this checkout has no details (#252) — the
	// issue is still being created, or its details live on another branch.
	CardOnly bool `json:"card_only,omitempty"`
	// Unreadable says why an "unreadable" status could not be read, when the
	// reader knows (#288: an unparseable tracker card).
	Unreadable string `json:"unreadable,omitempty"`
}

// WorktreeState describes one entry from `git worktree list --porcelain -z`.
type WorktreeState struct {
	Path   string `json:"path"`
	Branch string `json:"branch"`
}

// CommitState is one entry from `git log main..HEAD` on the current
// branch. Subjects are surfaced (not bodies) to keep state output tight.
type CommitState struct {
	SHA     string `json:"sha"`
	Subject string `json:"subject"`
}

// DriftFinding is a single structural-consistency observation surfaced
// by state. Severity is advisory; state never refuses, only reports.
type DriftFinding struct {
	Severity string `json:"severity"` // "info" or "warn"
	Issue    string `json:"issue,omitempty"`
	Message  string `json:"message"`
}

// State is the full snapshot — the root JSON object when --json is set.
type State struct {
	Workspace workspace.Identity `json:"workspace"`
	Repo      string             `json:"repo"`
	Branch    string             `json:"branch"`
	Issues    []IssueState       `json:"issues"`
	Worktrees []WorktreeState    `json:"worktrees"`
	Recent    []CommitState      `json:"recent_commits"`
	Drift     []DriftFinding     `json:"drift"`
	// TrackerStale: the tracker could not be fetched; cards are as last fetched.
	TrackerStale bool `json:"tracker_stale,omitempty"`
}

// ── command constructor ─────────────────────────────────────────────────────

func NewStateCmd() *cobra.Command {
	f := stateFlags{}
	cmd := &cobra.Command{
		Use:           "state",
		Short:         "Inspect SDLC workflow state (read-only, JSON optional)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"state\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f.IssuesExplicit = cmd.Flags().Changed("issues-dir")
			f.HistoryExplicit = cmd.Flags().Changed("history-dir")
			return runState(cmd.Context(), cmd.OutOrStdout(), &f)
		},
	}
	cmd.Flags().BoolVar(&f.JSON, "json", false, "emit JSON instead of human-readable prose")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", "workshop/issues", "directory holding issue files")
	cmd.Flags().StringVar(&f.HistoryDir, "history-dir", "workshop/history", "directory holding archived issues")
	return cmd
}

// ── main flow ───────────────────────────────────────────────────────────────

func runState(ctx context.Context, stdout io.Writer, f *stateFlags) error {
	identity, err := resolveWorkspace(".")
	if err != nil {
		return err
	}
	recent, baseRef := recentCommits()
	worktrees, err := listWorktrees()
	if err != nil {
		return err
	}
	s := State{
		Workspace: identity,
		Repo:      gitx.Capture("rev-parse", "--show-toplevel"),
		Branch:    gitx.Capture("branch", "--show-current"),
		Worktrees: worktrees,
		Recent:    recent,
	}

	issuesDir := defaultWorkspacePath(identity.WorktreeRoot, f.IssuesDir, f.IssuesExplicit)
	historyDir := defaultWorkspacePath(identity.WorktreeRoot, f.HistoryDir, f.HistoryExplicit)
	issues, stale, err := listIssueStates(ctx, issuesDir)
	if err != nil {
		return fmt.Errorf("list issues: %w", err)
	}
	s.Issues, s.TrackerStale = issues, stale
	// gitx.ShippedWorkOnMain is the production ship probe; state_test fakes it.
	s.Drift = detectDrift(issues, historyDir, gitx.ShippedWorkOnMain)
	if baseRef == "" {
		s.Drift = append(s.Drift, DriftFinding{
			Severity: "info",
			Message:  "no main/origin/main detected — recent-commits unavailable",
		})
	}

	if f.JSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(s)
	}
	return renderProse(stdout, s)
}

// listWorktrees returns the parsed `git worktree list --porcelain -z` entries.
func listWorktrees() ([]WorktreeState, error) {
	porcelain, err := gitx.RunGit("worktree", "list", "--porcelain", "-z")
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}
	worktrees, err := gitx.ParseWorktrees(porcelain)
	if err != nil {
		return nil, fmt.Errorf("parse git worktree list: %w", err)
	}
	return worktreeStates(worktrees), nil
}

// worktreeStates deliberately maps rich porcelain records onto state's stable
// {path,branch} JSON contract. State does not expose parser-only attributes.
func worktreeStates(worktrees []gitx.Worktree) []WorktreeState {
	states := make([]WorktreeState, 0, len(worktrees))
	for _, worktree := range worktrees {
		branch := worktree.Branch
		switch {
		case worktree.Detached:
			branch = "(detached)"
		case worktree.Bare:
			branch = "(bare)"
		}
		states = append(states, WorktreeState{Path: worktree.Path, Branch: branch})
	}
	return states
}

// recentCommits returns the subjects of commits on the current branch
// since it diverged from origin/main (falls back to main if no upstream).
// Returns (commits, baseRef) where baseRef is the ref that was used; empty
// if neither origin/main nor main exists (fresh repo or master-only).
// Caller can surface a drift finding when baseRef is empty. Cap at 20
// entries to keep prose tight.
func recentCommits() ([]CommitState, string) {
	base := gitx.MainRef()
	if base == "" {
		return nil, ""
	}
	out := gitx.Capture("log", base+"..HEAD", "--pretty=%H%x00%s")
	if out == "" {
		return nil, base
	}
	var cs []CommitState
	for _, line := range strings.Split(out, "\n") {
		parts := strings.SplitN(line, "\x00", 2)
		if len(parts) != 2 {
			continue
		}
		cs = append(cs, CommitState{SHA: parts[0], Subject: parts[1]})
		if len(cs) >= 20 {
			break
		}
	}
	return cs, base
}

// ── issue parsing ───────────────────────────────────────────────────────────

// titleRE matches the first `# Title` heading after the frontmatter.
var titleRE = regexp.MustCompile(`(?m)^# (.+)$`)

// listIssues composes the active issues of the current repository: every
// details file in issuesDir, plus
// open cards whose details are not in this checkout (#252: a card-only issue is
// still being created). Card-owned fields come from the card. Sorted by ID.
// listIssueStates also reports whether the cards came from a stale tracker read.
// The repository is the one containing issuesDir (a dependency's, for
// start-plan's contention check).
func listIssueStates(ctx context.Context, issuesDir string) ([]IssueState, bool, error) {
	rs, err := loadIssueRecords(ctx, issuesDir, tracker.PreferFresh)
	if err != nil {
		return nil, false, err
	}
	var out []IssueState
	for _, rec := range rs.All() {
		if rec.CardErr != nil {
			// The card cannot be read (#288): its status is unknown, never the
			// mirror's; surfaced so it is not mistaken for absent.
			out = append(out, IssueState{ID: rec.ID, Path: rec.DetailPath, Status: "unreadable", CardOnly: rec.DetailPath == "", Unreadable: rec.CardErr.Error()})
			continue
		}
		if rec.DetailPath == "" {
			// Card only: archived (terminal) cards are history, not active work.
			if vocab.Issue().IsTerminal(rec.Status()) {
				continue
			}
			updated, _ := rec.Field("updated")
			out = append(out, IssueState{ID: rec.ID, Status: rec.Status(), Title: rec.Title(), Updated: updated, CardOnly: true})
			continue
		}
		if rec.DetailErr != nil {
			// Don't drop silently — surface as an unreadable entry so
			// detectDrift can warn (M2 review C2); a malformed file keeps an
			// empty status.
			status := ""
			if rec.DetailUnreadable {
				status = "unreadable"
			}
			out = append(out, IssueState{ID: rec.ID, Path: rec.DetailPath, Status: status})
			continue
		}
		updated, _ := rec.Field("updated")
		total, ticked := issue.CountPlanItems(rec.DetailBody)
		out = append(out, IssueState{
			ID:         rec.ID,
			Path:       rec.DetailPath,
			Status:     rec.Status(),
			Title:      rec.Title(),
			PlanTotal:  total,
			PlanTicked: ticked,
			Updated:    updated,
		})
	}
	return out, rs.Stale, nil
}

// ── drift detection ─────────────────────────────────────────────────────────

// shipProbe reports whether implementation work for an issue has landed on
// main — the IO seam detectDrift depends on for its close-off check. Returns
// (firstWorkSHA, itsSubject, shipped). Production wires gitx.ShippedWorkOnMain;
// tests pass a fake so the drift logic is exercised without touching git.
type shipProbe func(issueNum string) (sha, subject string, shipped bool)

// detectDrift surfaces structural inconsistencies. Warn-only — state
// reports drift but never refuses (refusal lives on mutating verbs).
//
// Checks:
//  1. Issue with status=done is still in workshop/issues/ (should be
//     archived to workshop/history/).
//  2. Issue with status=working has zero plan items ticked — likely
//     stale state (the no-progress end of the spectrum).
//  3. Issue file with no frontmatter — broken state.
//  4. Close-off candidate (#76): an open/working issue whose plan is all (or
//     all-but-one) ticked AND whose work has shipped to main — done work that
//     never got formally closed. The all-progress-no-close end, the inverse of
//     check 2. Warn-only: surface for a human glance, never auto-close (closing
//     carries actual/verified judgment a heuristic can't supply).
func detectDrift(issues []IssueState, historyDir string, shipped shipProbe) []DriftFinding {
	archivedIssuesDir := vocab.ArchiveSubdir(historyDir, vocab.ArchiveIssues) // #181 layout in the drift hint
	var out []DriftFinding
	for _, i := range issues {
		// Tagless switch so the terminal arm can read the model's category (#122);
		// "" / "unreadable" are sentinel non-statuses (reader markers), not model values.
		switch {
		case i.Status == "":
			out = append(out, DriftFinding{
				Severity: "warn",
				Issue:    i.ID,
				Message:  "no frontmatter or missing status: field",
			})
		case i.Status == "unreadable":
			msg := fmt.Sprintf("could not read %s — check permissions / symlinks", i.Path)
			if i.Unreadable != "" {
				msg = "card unreadable: " + i.Unreadable + " — repair it on the tracker"
			}
			out = append(out, DriftFinding{Severity: "warn", Issue: i.ID, Message: msg})
		case vocab.Issue().IsTerminal(i.Status):
			out = append(out, DriftFinding{
				Severity: "warn",
				Issue:    i.ID,
				Message:  fmt.Sprintf("status=%s but still in workshop/issues/ — move to %s/ (`sdlc push`/`merge` archives it)", i.Status, archivedIssuesDir),
			})
		case i.Status == "working": // #122 carve-out: working-specific (blocked is waiting; IsActive too broad)
			if i.PlanTotal > 0 && i.PlanTicked == 0 {
				out = append(out, DriftFinding{
					Severity: "info",
					Issue:    i.ID,
					Message:  fmt.Sprintf("working with %d plan item(s), none ticked yet", i.PlanTotal),
				})
			}
		}
		if f, ok := closeOffFinding(i, shipped); ok {
			out = append(out, f)
		}
	}
	return out
}

// closeOffFinding returns the close-off candidate finding for i, if it
// qualifies. Pre-filters cheaply on status + plan completion before paying for
// the ship probe; the `PlanTicked >= 1` floor keeps this disjoint from
// detectDrift's "working, none ticked" info finding (no contradictory
// double-flag) and rejects the degenerate 1-item/0-ticked case.
func closeOffFinding(i IssueState, shipped shipProbe) (DriftFinding, bool) {
	// #122: close-off candidates are open or actively-working (not blocked/terminal);
	// "open" reads from the model, "working" stays literal (blocked is waiting, not done).
	if !vocab.Issue().IsOpen(i.Status) && i.Status != "working" {
		return DriftFinding{}, false
	}
	if i.PlanTicked < 1 || i.PlanTotal-i.PlanTicked > 1 {
		return DriftFinding{}, false
	}
	// Commit subjects reference the *unpadded* number (`#82`, §12), so the probe
	// (and the close hint) take unpadID, while the finding keeps the padded ID
	// for display parity with the other findings.
	num := unpadID(i.ID)
	sha, subj, ok := shipped(num)
	if !ok {
		return DriftFinding{}, false
	}
	return DriftFinding{
		Severity: "warn",
		Issue:    i.ID,
		Message: fmt.Sprintf("looks done — plan %d/%d + shipped work on main (%s %q) — close it? (sdlc close --issue %s)",
			i.PlanTicked, i.PlanTotal, abbrevSHA(sha), truncate(subj, 50), num),
	}, true
}

// abbrevSHA trims an already-resolved commit hash to 8 chars for display —
// pure, unlike the IO-doing shortSHA in milestoneclose.go (the probe hands us a
// full SHA, so no `git rev-parse --short` round-trip is needed).
//
// #194 gave it a second caller where the purity is load-bearing rather than
// merely cheap: the boundary review's Review-Window trailer. shortSHA RESOLVES
// its argument, so shortSHA("HEAD") returns the ambient repo's HEAD — and the
// review window's head degrades to the literal "HEAD" when rev-parse failed.
// Rendering that through shortSHA would print a commit the review never read;
// through abbrevSHA it stays "HEAD", visibly degraded rather than silently wrong.
func abbrevSHA(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

// unpadID strips the 6-digit zero padding for the `sdlc close --issue N` hint
// (issues are addressed by bare number). Falls back to the padded form if it
// would otherwise empty out.
func unpadID(id string) string { return issue.CLIRef(id) }

// ── prose rendering ─────────────────────────────────────────────────────────

func renderProse(w io.Writer, s State) error {
	fmt.Fprintf(w, "Repo:    %s\n", s.Repo)
	fmt.Fprintf(w, "Branch:  %s\n", s.Branch)
	if s.Workspace.SchemaVersion != 0 {
		fmt.Fprintf(w, "Workspace: %s\nResting branch: %s\n", workspaceText(s.Workspace.Address, "(ordinary worktree)"), workspaceText(s.Workspace.RestingBranch, "(none)"))
	}
	fmt.Fprintln(w)

	if s.TrackerStale {
		fmt.Fprintln(w, "Issues (tracker unreachable — card fields as last fetched):")
	} else {
		fmt.Fprintln(w, "Issues:")
	}
	if len(s.Issues) == 0 {
		fmt.Fprintln(w, "  (none)")
	}
	for _, i := range s.Issues {
		// "#000031  status: working  3/8 ticked  — title"
		if i.CardOnly {
			fmt.Fprintf(w, "  #%s  status: %-8s  card only", i.ID, valueOr(i.Status, "?"))
		} else {
			fmt.Fprintf(w, "  #%s  status: %-8s  %d/%d ticked", i.ID, valueOr(i.Status, "?"), i.PlanTicked, i.PlanTotal)
		}
		if i.Title != "" {
			fmt.Fprintf(w, "  — %s", truncate(i.Title, 60))
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Worktrees:")
	if len(s.Worktrees) == 0 {
		fmt.Fprintln(w, "  (none)")
	}
	for _, wt := range s.Worktrees {
		fmt.Fprintf(w, "  %s  (%s)\n", wt.Path, valueOr(wt.Branch, "(detached)"))
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Recent commits (main..HEAD):")
	if len(s.Recent) == 0 {
		fmt.Fprintln(w, "  (branch is at base)")
	}
	for _, c := range s.Recent {
		fmt.Fprintf(w, "  %s  %s\n", c.SHA[:8], truncate(c.Subject, 80))
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Drift:")
	if len(s.Drift) == 0 {
		fmt.Fprintln(w, "  (no inconsistencies detected)")
	}
	for _, d := range s.Drift {
		tag := fmt.Sprintf("[%s]", d.Severity)
		if d.Issue != "" {
			fmt.Fprintf(w, "  %s #%s — %s\n", tag, d.Issue, d.Message)
		} else {
			fmt.Fprintf(w, "  %s %s\n", tag, d.Message)
		}
	}

	// Footer: timestamp so output is reproducible-ish for logging.
	fmt.Fprintf(w, "\n(captured at %s)\n", time.Now().Format(time.RFC3339))
	return nil
}

// valueOr and truncate live in term.go (shared across the sdlc verbs).
// M2 review I1's rune-aware truncate is preserved verbatim there.
