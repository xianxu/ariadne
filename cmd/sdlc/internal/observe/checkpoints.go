package observe

import (
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flow"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// BranchFacts is the issue branch as read (Ref "" when it exists nowhere here).
type BranchFacts struct {
	Ref, Head, LastCommitAt string
	AheadOfMain             int
	Err                     error
}

// HoldingFacts is one local worktree that has the issue branch checked out.
type HoldingFacts struct {
	Path, Address, Branch, Head string
	DirtyCount, Ahead, Behind   int
	Err                         error // its activity facts could not be read
}

// ArtifactFacts is one review boundary's committed evidence at the evidence
// location: its prose sidecar (the verdict) and its gate ledger's open
// blocking-severity findings for that boundary.
type ArtifactFacts struct {
	Sidecar      string // "" when Found is false
	Found        bool
	SidecarErr   error
	OpenBlocking int
	LedgerErr    error
}

// Evidence is where checkpoints are read: the issue branch's committed tree
// before landing, main's archive after (artifacts move with the issue, #143),
// so squash merges and deleted branches don't strand them.
type Evidence struct {
	Source     string // e.g. "refs/heads/000279-x:workshop/plans"
	Err        error  // the location could not be read
	Details    []byte // the details file there (nil: not there, or unreadable — see DetailsErr)
	DetailsErr error  // the details file is there but could not be read
	Artifacts  map[string]ArtifactFacts
	PlanGate   *ArtifactFacts // the plan-quality ledger, if the issue has one
}

func assembleBranch(in Inputs) Branch {
	b := Branch{Authority: AuthorityCommitted}
	switch {
	case in.Branch.Err != nil:
		b.Read = Read{State: Unknown, Error: in.Branch.Err.Error()}
	case in.Branch.Ref == "":
		b.Read = Read{State: Absent, Source: "no issue branch in this repository"}
	default:
		b.Read = Read{State: Present, Source: in.Branch.Ref}
		b.Ref, b.Head, b.LastCommitAt, b.CommitsAheadOfMain = in.Branch.Ref, in.Branch.Head, in.Branch.LastCommitAt, in.Branch.AheadOfMain
	}
	return b
}

func assembleWorkspaces(in Inputs, a Assignment) Workspaces {
	w := Workspaces{Authority: AuthorityWorktree, Holding: []Workspace{}}
	if in.WorktreesErr != nil {
		w.Read = Read{State: Unknown, Error: in.WorktreesErr.Error()}
		return w
	}
	for _, h := range in.Holding {
		ws := Workspace{Path: h.Path, Address: h.Address, Branch: h.Branch, Head: h.Head,
			DirtyCount: h.DirtyCount, Ahead: h.Ahead, Behind: h.Behind,
			IsClaimant: a.Claimant != nil && a.Claimant.Worktree == h.Path, Error: errText(h.Err)}
		w.Holding = append(w.Holding, ws)
	}
	w.Read = Read{State: Present, Source: "git worktree list"}
	if len(w.Holding) == 0 {
		w.Read.State = Absent
	}
	return w
}

// expectedBoundaries are the review boundaries the issue's own record says
// happened: ticked milestones in the plan, and a whole-issue close once the
// card is codecomplete or done. Their evidence missing is unknown, not absent.
func expectedBoundaries(ticked map[string]bool, status string) map[string]bool {
	want := map[string]bool{}
	for tag := range ticked {
		want[tag] = true
	}
	if status == "codecomplete" || status == "done" {
		want["close"] = true
	}
	return want
}

var sidecarRowRE = regexp.MustCompile(`(?m)^\|\s*(verdict|window|timestamp)\s*\|\s*(.*?)\s*\|\s*$`)

// sidecarRows reads the metadata table a review sidecar opens with: the first
// row of each key wins, so a table quoted later in the review body cannot
// override it.
func sidecarRows(text string) map[string]string {
	rows := map[string]string{}
	for _, m := range sidecarRowRE.FindAllStringSubmatch(text, -1) {
		if _, seen := rows[m[1]]; !seen {
			rows[m[1]] = m[2]
		}
	}
	return rows
}

func assembleCheckpoints(in Inputs, c Card) Checkpoints {
	cp := Checkpoints{Authority: AuthorityCommitted, Reviews: []Review{}}
	ev := in.Evidence
	switch {
	case ev.Err != nil:
		cp.Read = Read{State: Unknown, Source: ev.Source, Error: ev.Err.Error()}
		return cp
	case ev.Source == "":
		cp.Read = Read{State: Absent, Source: "no issue branch and nothing archived on main"}
		if c.Status == "codecomplete" || c.Status == "done" {
			cp.Read = Read{State: Unknown, Error: "the card records a close, but no branch or archive here holds its evidence"}
		}
		return cp
	}
	cp.Read = Read{State: Present, Source: ev.Source}
	// Flow, plan and the ticked milestones are read from the details. A read
	// that fails degrades the section to unknown with its reason — never a
	// zero value standing in for an answer. (No Plan section at all is a real
	// absence: 0/0.)
	plan := "" // fence-filtered, as every milestone reader requires
	degrade := func(why string) { cp.Read = Read{State: Unknown, Source: ev.Source, Error: why} }
	switch {
	case ev.DetailsErr != nil:
		degrade("the details at the evidence location are unreadable: " + ev.DetailsErr.Error())
	case ev.Details == nil:
		degrade("the evidence location holds this issue's review artifacts but not its details")
	default:
		fm, body, err := issue.Parse(string(ev.Details))
		if err != nil {
			degrade("the details at the evidence location do not parse: " + err.Error())
			break
		}
		if f, recorded, ferr := flow.FromFrontmatter(fm); ferr != nil {
			degrade("the details' flow record is malformed: " + ferr.Error())
		} else if recorded {
			cp.Flow = &Flow{Kind: string(f.Kind()), Provenance: string(f.Provenance())}
		}
		cp.Plan.Total, cp.Plan.Ticked = issue.CountPlanItems(body)
		if items, ok := issue.PlanItemsBody(body); ok {
			plan = items
		}
	}
	if ev.PlanGate != nil {
		cp.Reviews = append(cp.Reviews, review("plan", *ev.PlanGate, false, ev.Source))
	}
	expected := expectedBoundaries(issue.TickedMilestones(plan), c.Status)
	order := issue.MilestonesInPlanOrder(plan)
	order = append(order, "close")
	for _, boundary := range order {
		art, has := ev.Artifacts[boundary]
		if !has && !expected[boundary] {
			continue // not reached: no evidence expected, none present
		}
		cp.Reviews = append(cp.Reviews, review(boundary, art, expected[boundary], ev.Source))
	}
	return cp
}

// review reads one boundary's evidence. A boundary the record says happened,
// whose artifact is not found, is unknown — evidence that should exist.
func review(boundary string, a ArtifactFacts, expected bool, source string) Review {
	r := Review{Boundary: boundary, OpenBlocking: a.OpenBlocking}
	switch {
	case a.SidecarErr != nil:
		r.Read = Read{State: Unknown, Source: source, Error: "unreadable review artifact: " + a.SidecarErr.Error()}
	case a.LedgerErr != nil:
		r.Read = Read{State: Unknown, Source: source, Error: "unreadable gate ledger: " + a.LedgerErr.Error()}
	case boundary == "plan":
		r.Read = Read{State: Present, Source: source} // plan-quality keeps a ledger, not a verdict sidecar
	case !a.Found && expected:
		r.Read = Read{State: Unknown, Source: source, Error: "recorded as closed, but its review artifact is not at " + source}
	case !a.Found:
		r.Read = Read{State: Absent, Source: source}
	default:
		rows := sidecarRows(a.Sidecar)
		r.Read = Read{State: Present, Source: source}
		r.Window, r.At = rows["window"], rows["timestamp"]
		switch v := strings.TrimSpace(rows["verdict"]); {
		case v == "":
			r.Read = Read{State: Unknown, Source: source, Error: "review artifact has no verdict row"}
		case !vocab.Verdict().IsEmitted(v):
			r.Read = Read{State: Unknown, Source: source, Error: "review artifact records " + v + ", not a review verdict (the review produced none)"}
		default:
			r.Verdict = v
		}
	}
	return r
}
