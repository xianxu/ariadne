package observe

import (
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flow"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
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
	Source    string // e.g. "refs/heads/000279-x:workshop/plans"
	Err       error  // the location could not be read
	Details   []byte // the details file there (nil: not there)
	Artifacts map[string]ArtifactFacts
	PlanGate  *ArtifactFacts // the plan-quality ledger, if the issue has one
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

var tickedMilestoneRE = regexp.MustCompile(`(?m)^- \[x\] \*{0,2}(M\d+[a-z]?)\b`)

// expectedBoundaries are the review boundaries the issue's own record says
// happened: ticked milestones in the plan, and a whole-issue close once the
// card is codecomplete or done. Their evidence missing is unknown, not absent.
func expectedBoundaries(details []byte, status string) map[string]bool {
	want := map[string]bool{}
	if _, body, err := issue.Parse(string(details)); err == nil {
		if plan, ok := issue.SectionBody(body, "Plan"); ok {
			for _, m := range tickedMilestoneRE.FindAllStringSubmatch(plan, -1) {
				want[m[1]] = true
			}
		}
	}
	if status == "codecomplete" || status == "done" {
		want["close"] = true
	}
	return want
}

var sidecarRowRE = regexp.MustCompile(`(?m)^\|\s*(verdict|window|timestamp)\s*\|\s*(.*?)\s*\|\s*$`)

// sidecarRows reads the verdict table a review sidecar opens with.
func sidecarRows(text string) map[string]string {
	rows := map[string]string{}
	for _, m := range sidecarRowRE.FindAllStringSubmatch(text, -1) {
		rows[m[1]] = m[2]
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
	if ev.Details != nil {
		if fm, body, err := issue.Parse(string(ev.Details)); err == nil {
			if f, recorded, ferr := flow.FromFrontmatter(fm); ferr == nil && recorded {
				cp.Flow = &Flow{Kind: string(f.Kind()), Provenance: string(f.Provenance())}
			}
			cp.Plan.Total, cp.Plan.Ticked = issue.CountPlanItems(body)
		}
	}
	if ev.PlanGate != nil {
		cp.Reviews = append(cp.Reviews, review("plan", *ev.PlanGate, false, ev.Source))
	}
	expected := expectedBoundaries(ev.Details, c.Status)
	order := issue.MilestonesInPlanOrder(planOf(ev.Details))
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

func planOf(details []byte) string {
	if _, body, err := issue.Parse(string(details)); err == nil {
		if plan, ok := issue.SectionBody(body, "Plan"); ok {
			return plan
		}
	}
	return ""
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
		r.Verdict, r.Window, r.At = rows["verdict"], rows["window"], rows["timestamp"]
		if strings.TrimSpace(r.Verdict) == "" {
			r.Read = Read{State: Unknown, Source: source, Error: "review artifact has no verdict row"}
		}
	}
	return r
}
