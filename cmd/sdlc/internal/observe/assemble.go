package observe

import (
	"path"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// Inputs are the raw facts a collector read, each with its own read error.
// Assemble derives every judgment from them and performs no IO.
type Inputs struct {
	Issue      string // six-digit id
	ObservedAt time.Time

	// Tracked is false when the repository's publication remote has no issue
	// tracker: its details are the whole record and there is no card.
	Tracked       bool
	TrackerRef    string // the tracker commit the card was read at
	TrackerStale  bool   // read from the last fetch because the fresh one failed
	TrackerErr    error  // the fetch failure (stale), or the read failure
	TrackerRefErr error  // the cards were read but their commit could not be named

	Card     []byte // nil: no card for this id
	CardPath string
	CardBlob string

	// Me is the queried checkout's identity — what "this workspace" means.
	Me    *issue.Claimant
	MeErr error

	// Worktrees are every local worktree of the queried repository.
	Worktrees    []LocalWorktree
	WorktreesErr error

	// MainArchive is the issue's archived details path on main ("" when it is
	// not archived there).
	MainArchive    string
	MainArchiveErr error

	Branch   BranchFacts
	Holding  []HoldingFacts // local worktrees holding the issue branch
	Evidence Evidence       // where checkpoints are read
}

type LocalWorktree struct{ Path, Branch string }

// Assemble derives the observation. Every section's state is its read quality;
// values live in their own fields.
func Assemble(in Inputs) Observation {
	o := Observation{
		SchemaVersion: SchemaVersion,
		Issue:         in.Issue,
		ObservedAt:    in.ObservedAt.UTC().Format(time.RFC3339),
	}
	o.Tracker = assembleTracker(in)
	o.Card = assembleCard(in, o.Tracker)
	o.Assignment = assembleAssignment(in, o.Card)
	o.Workspaces = assembleWorkspaces(in, o.Assignment)
	o.Branch = assembleBranch(in)
	o.Checkpoints = assembleCheckpoints(in, o.Card)
	o.Completion = assembleCompletion(in, o.Card)
	o.Landing = assembleLanding(in, o.Card, o.Completion)
	return o
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func assembleTracker(in Inputs) Tracker {
	switch {
	case !in.Tracked && in.TrackerErr == nil:
		return Tracker{Read: Read{State: Absent, Source: "publication remote has no issue tracker; the details are the record"}}
	case in.TrackerStale:
		return Tracker{Read: Read{State: Stale, Error: "tracker unreachable, answered from the last fetch: " + errText(in.TrackerErr)}, Ref: in.TrackerRef, RefError: errText(in.TrackerRefErr)}
	case in.TrackerErr != nil:
		return Tracker{Read: Read{State: Unknown, Error: errText(in.TrackerErr)}}
	}
	return Tracker{Read: Read{State: Present}, Ref: in.TrackerRef, RefError: errText(in.TrackerRefErr)}
}

// tracked derives a tracker-authority section's read from the tracker read:
// absent stays absent, unknown stays unknown, stale stays stale (with its
// reason), present stays present.
func tracked(t Tracker, source string) Read {
	r := Read{State: t.State, Error: t.Error}
	if t.State == Present || t.State == Stale {
		r.Source = source
	}
	return r
}

func assembleCard(in Inputs, t Tracker) Card {
	c := Card{Authority: AuthorityTracker}
	if t.State == Absent || t.State == Unknown {
		c.Read = tracked(t, "")
		return c
	}
	if in.Card == nil {
		c.Read = Read{State: Absent, Source: "issue-tracker"}
		if t.State == Stale {
			c.Read = Read{State: Stale, Source: "issue-tracker", Error: t.Error + " (no card for this id as last fetched)"}
		}
		return c
	}
	card, err := issue.ParseCard(in.Card)
	if err != nil {
		c.Read = Read{State: Unknown, Source: "issue-tracker:" + in.CardPath, Error: "unreadable card: " + err.Error()}
		return c
	}
	c.Read = tracked(t, "issue-tracker:"+in.CardPath)
	c.Revision, c.Title = in.CardBlob, card.Title
	c.Status, _ = issue.GetField(card.Frontmatter, "status")
	c.Started, _ = issue.GetField(card.Frontmatter, "started")
	c.Updated, _ = issue.GetField(card.Frontmatter, "updated")
	return c
}

// valued reports whether a read yielded data (present, or stale-as-last-fetched).
func valued(r Read) bool { return r.State == Present || r.State == Stale }

// cardHasValue is true only when the card section holds a parsed card.
func cardHasValue(c Card) bool { return valued(c.Read) && c.Status != "" }

func assembleAssignment(in Inputs, c Card) Assignment {
	a := Assignment{Authority: AuthorityTracker, Read: c.Read}
	if !cardHasValue(c) {
		if valued(a.Read) { // stale with no card as last fetched
			a.Relation = RelationUnknown
		}
		return a
	}
	recorded, has, err := issue.CardClaimant(in.Card)
	if err != nil {
		a.Read = Read{State: Unknown, Source: c.Source, Error: "unreadable claimant: " + err.Error()}
		return a
	}
	if !has {
		a.Relation = RelationUnattributed
		return a
	}
	a.Claimant = &Claimant{Operator: recorded.Operator, Machine: recorded.Machine, MachineName: recorded.MachineName,
		Workspace: recorded.Workspace, Worktree: recorded.Worktree, Repository: recorded.Repository}
	if in.Me == nil {
		a.Relation, a.ClaimantWorktree = RelationUnknown, FateUnknown
		a.RelationError = "this checkout's identity is unavailable: " + errText(in.MeErr)
		return a
	}
	switch issue.MatchClaimant(&recorded, *in.Me) {
	case issue.OwnershipMine:
		a.Relation = RelationThisWorkspace
	default:
		a.Relation = RelationOtherWorkspace
	}
	a.ClaimantWorktree = worktreeFate(in, recorded)
	return a
}

// worktreeFate places the recorded owner's worktree: another machine's is
// never probed; on this machine it is judged from the repository's worktree
// list alone.
func worktreeFate(in Inputs, recorded issue.Claimant) WorktreeFate {
	if recorded.Machine != in.Me.Machine {
		return FateOtherMachine
	}
	if in.WorktreesErr != nil {
		return FateUnknown
	}
	branch := strings.TrimSuffix(path.Base(in.CardPath), ".md")
	for _, w := range in.Worktrees {
		if w.Path == recorded.Worktree {
			if w.Branch == branch {
				return FateHoldsBranch
			}
			return FateElsewhere
		}
	}
	return FateMissing
}

func assembleCompletion(in Inputs, c Card) Completion {
	out := Completion{Authority: AuthorityTracker, Read: c.Read}
	if !cardHasValue(c) {
		return out
	}
	b, ok, err := issue.CardCompletion(in.Card)
	switch {
	case err != nil:
		out.Read = Read{State: Unknown, Source: c.Source, Error: "unreadable completion binding: " + err.Error()}
	case !ok && c.State == Present:
		out.Read = Read{State: Absent, Source: c.Source}
	case ok:
		out.Token, out.EvidenceCommit, out.ReviewedHead, out.LandedCommit = b.Token, b.EvidenceCommit, b.ReviewedHEAD, b.LandedCommit
	}
	return out
}

func assembleLanding(in Inputs, c Card, comp Completion) Landing {
	l := Landing{Authority: AuthorityTracker, Read: c.Read}
	if !cardHasValue(c) {
		if valued(l.Read) {
			l.Outcome = OutcomeNotLanded // stale, no card as last fetched
		}
		return l
	}
	l.Outcome = OutcomeNotLanded
	if c.Status == "done" {
		l.Outcome = OutcomeLanded
		l.LandedCommit = comp.LandedCommit
	}
	if in.MainArchiveErr != nil {
		l.ArchiveError = in.MainArchiveErr.Error()
	} else {
		l.Archived = in.MainArchive
	}
	return l
}
