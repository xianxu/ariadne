// projecttick.go — the close-time edit of one project file (#308): tick the
// closing issue's (or milestone's) task rows and, for a milestone, upsert its
// detail block. Pure: the same edit is previewed on the local copy at compute
// time and re-derived on the fetched copy when a peer repository's project is
// published to its origin (peerproject.go), so the two can never diverge.
package main

import (
	"fmt"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/project"
)

// projectEditInput is what the edit reads besides the project text.
type projectEditInput struct {
	Milestone          string // "" for an issue close
	RepoName, IssueStr string
	Actual, Today, Est string
	Label              string // repo/file, for messages
	WaiveDetailBlock   bool   // --force: a missing milestone detail block is a warning
}

// errMissingDetailBlock refuses a milestone close whose project has no detail
// block for it; Error() is the full authoring instruction, skeleton included.
type errMissingDetailBlock struct{ msg string }

func (e errMissingDetailBlock) Error() string { return e.msg }

// projectCloseEdit returns the edited text, what it applied, and warnings.
func projectCloseEdit(pt string, in projectEditInput) (newPT string, applied, warns []string, err error) {
	newPT = pt
	if in.Milestone == "" {
		ticked, n := project.TickAllTaskRowsForIssue(newPT, in.RepoName, in.IssueStr)
		newPT = ticked
		if n > 0 {
			applied = append(applied, fmt.Sprintf("ticked %d remaining task line(s) for %s#%s in %s", n, in.RepoName, in.IssueStr, in.Label))
		}
		if n > 1 {
			warns = append(warns, fmt.Sprintf("multiple %s#%s task rows ticked at once — confirm individual milestones were genuinely closed (§5 step 1)", in.RepoName, in.IssueStr))
		}
		return newPT, applied, warns, nil
	}
	ticked, n := project.TickMilestoneTaskRow(newPT, in.RepoName, in.IssueStr, in.Milestone)
	newPT = ticked
	if n > 0 {
		applied = append(applied, fmt.Sprintf("ticked [%s#%s %s] in %s", in.RepoName, in.IssueStr, in.Milestone, in.Label))
	} else {
		warns = append(warns, fmt.Sprintf("no task line for [%s#%s %s] in %s", in.RepoName, in.IssueStr, in.Milestone, in.Label))
	}
	anchor := project.AnchorFor(in.RepoName, in.IssueStr, in.Milestone)
	// Order matches close-issue.py: fm_set('actual') then fm_set('closed').
	// Slice (not map) so iteration order is deterministic.
	var fields []project.Field
	if in.Actual != "" {
		fields = append(fields, project.Field{Name: "actual", Value: in.Actual + "h"})
	}
	fields = append(fields, project.Field{Name: "closed", Value: in.Today})
	updated, found := project.UpsertDetailBlockFields(newPT, anchor, fields)
	if found {
		applied = append(applied, fmt.Sprintf("updated detail block <a id=\"%s\"> in %s", anchor, in.Label))
		return updated, applied, warns, nil
	}
	if in.WaiveDetailBlock {
		warns = append(warns, fmt.Sprintf("--force: skipping detail-block update for <a id=\"%s\"> in %s", anchor, in.Label))
		return newPT, applied, warns, nil
	}
	skel, refDef := project.Skeleton{
		Anchor:    anchor,
		RefLabel:  fmt.Sprintf("%s#%s %s", in.RepoName, in.IssueStr, in.Milestone),
		Title:     project.FindTaskTitle(newPT, in.RepoName, in.IssueStr, in.Milestone),
		Est:       in.Est,
		Actual:    in.Actual + "h",
		ClosedISO: in.Today,
	}.Render()
	return pt, nil, nil, errMissingDetailBlock{fmt.Sprintf(
		"no detail block <a id=\"%s\"> in %s (§5 step 4).\n"+
			"  Author one before closing — the prose paragraph is load-bearing\n"+
			"  for future calibration. Insert this skeleton inside ## details:\n\n"+
			"%s\n"+
			"  And add this reference definition at the file bottom:\n"+
			"    %s\n\n"+
			"  Then re-run. A track-only milestone with nothing worth recording authors a one-line\n"+
			"  block; `--no-project` skips the project update entirely (the tick included).",
		anchor, in.Label, skel, refDef)}
}
