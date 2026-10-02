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
	fmt.Fprintln(w, "  (activity is not progress: a working card proves a claim, not execution)")
}
