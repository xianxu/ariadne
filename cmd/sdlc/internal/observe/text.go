package observe

import (
	"fmt"
	"io"
	"strings"
)

// RenderText writes the observation's human view: one line per section, each
// with its read quality, so a stale or unknown read is never mistaken for an
// answer.
func RenderText(w io.Writer, o Observation) {
	short := func(oid string) string {
		if len(oid) > 12 {
			return oid[:12]
		}
		return oid
	}
	quality := func(r Read) string {
		switch r.State {
		case Present:
			return ""
		case Absent:
			return " (none recorded)"
		default:
			return fmt.Sprintf(" [%s: %s]", r.State, r.Error)
		}
	}
	fmt.Fprintf(w, "observations (as of %s, tracker %s%s):\n", o.ObservedAt, short(o.Tracker.Ref), quality(o.Tracker.Read))
	a := o.Assignment
	owner := string(a.Relation)
	if a.Claimant != nil {
		where := a.Claimant.Worktree
		if a.Claimant.Workspace != "" {
			where = a.Claimant.Workspace + " (" + a.Claimant.Worktree + ")"
		}
		owner = fmt.Sprintf("%s on %s at %s — %s, its worktree %s", a.Claimant.Operator, a.Claimant.MachineName, where, a.Relation, a.ClaimantWorktree)
	}
	if a.RelationError != "" {
		owner += " (" + a.RelationError + ")"
	}
	fmt.Fprintf(w, "  owner:       %s%s\n", owner, quality(a.Read))
	fmt.Fprintf(w, "  status:      %s%s\n", o.Card.Status, quality(o.Card.Read))
	c := o.Completion
	if c.EvidenceCommit != "" {
		fmt.Fprintf(w, "  completion:  close evidence %s (reviewed %s)%s\n", short(c.EvidenceCommit), short(c.ReviewedHead), quality(c.Read))
	} else {
		fmt.Fprintf(w, "  completion:  %s\n", strings.TrimSpace("no close recorded"+map[bool]string{true: "", false: quality(c.Read)}[c.State == Absent]))
	}
	l := o.Landing
	landing := string(l.Outcome)
	if l.LandedCommit != "" {
		landing += " at " + short(l.LandedCommit)
	}
	if l.Archived != "" {
		landing += ", archived at " + l.Archived
	}
	if l.ArchiveError != "" {
		landing += ", archive unknown: " + l.ArchiveError
	}
	fmt.Fprintf(w, "  landing:     %s%s\n", strings.TrimSpace(landing), quality(l.Read))
	b := o.Branch
	if b.Ref != "" {
		fmt.Fprintf(w, "  branch:      %s at %s, %d ahead of main, last commit %s\n", b.Ref, short(b.Head), b.CommitsAheadOfMain, b.LastCommitAt)
	} else {
		fmt.Fprintf(w, "  branch:      none here%s\n", quality(b.Read))
	}
	ws := o.Workspaces
	if len(ws.Holding) == 0 {
		fmt.Fprintf(w, "  worktrees:   none holds the branch%s\n", quality(ws.Read))
	}
	for _, h := range ws.Holding {
		label := h.Path
		if h.Address != "" {
			label = h.Address + " (" + h.Path + ")"
		}
		owner := ""
		if h.IsClaimant {
			owner = ", the owner's"
		}
		fmt.Fprintf(w, "  worktree:    %s%s — %d dirty, %d ahead / %d behind main%s\n", label, owner, h.DirtyCount, h.Ahead, h.Behind,
			map[bool]string{true: " [unreadable: " + h.Error + "]", false: ""}[h.Error != ""])
	}
	cp := o.Checkpoints
	flowText := "no flow recorded"
	if cp.Flow != nil {
		flowText = cp.Flow.Kind + " flow (" + cp.Flow.Provenance + ")"
	}
	fmt.Fprintf(w, "  checkpoints: %s, plan %d/%d ticked%s\n", flowText, cp.Plan.Ticked, cp.Plan.Total, quality(cp.Read))
	for _, r := range cp.Reviews {
		verdict := r.Verdict
		if r.Boundary == "plan" {
			verdict = "plan-quality"
		}
		fmt.Fprintf(w, "    %-7s %s, %d open blocking%s\n", r.Boundary+":", verdict, r.OpenBlocking, quality(r.Read))
	}
	fmt.Fprintln(w, "  (worktree activity is not progress; a working card proves a claim, not execution)")
}
