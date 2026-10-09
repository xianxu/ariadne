// externalmerge.go — merges done outside sdlc (#287). An issue branch merged
// into main by the GitHub button (not `sdlc merge`) leaves its card short of
// done. sdlc notices from git alone — the pushed branch tip (#286 keeps it
// current) or the close's evidence commit is on main — reports it read-only
// in `sdlc state`, and finishes the bookkeeping in reconcile and the next
// merge. Squash and rebase merges leave neither on main and are not seen.
package main

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

type mergeVerdict int

const (
	mergeNone        mergeVerdict = iota
	mergeSettle                   // the close is on main: done, archive, delete the branch
	mergeCloseNeeded              // merged without a close: the owner closes it
	mergeClaimNeeded              // merged, unowned: claim, then close
)

// mergeFacts are what one started issue is judged on.
type mergeFacts struct {
	EvidenceOnMain bool // the completion binding's evidence commit is on main
	BranchMerged   bool // the pushed issue branch's tip is on main
	Owner          issue.Ownership
}

// externalMergeVerdict classifies an issue against main. Landed evidence
// binds the close, so it settles whoever owns the card.
func externalMergeVerdict(f mergeFacts) mergeVerdict {
	switch {
	case f.EvidenceOnMain:
		return mergeSettle
	case !f.BranchMerged:
		return mergeNone
	case f.Owner == issue.OwnershipUnknown:
		return mergeClaimNeeded
	}
	return mergeCloseNeeded
}

// externalMerge is one started issue judged against main.
type externalMerge struct {
	ID, CardPath, Branch, Tip string
	Verdict                   mergeVerdict
}

// externalMerges judges every started, unfinished card in rs against mainTip.
// Evidence on main is ownedCompletions' own test (one source with the settle
// that finishes it). A branch counts as merged when its tip is on main but off
// main's first-parent line: the GitHub button merges with a merge commit,
// while a fresh branch with no work of its own sits on that line.
func externalMerges(env *trackerEnv, rs tracker.Records, mainTip string) ([]externalMerge, error) {
	landed, err := ownedCompletions(env, rs, mainTip, "", false)
	if err != nil {
		return nil, err
	}
	settled := map[string]bool{}
	for _, oc := range landed {
		settled[oc.ID] = true
	}
	var firstParent map[string]bool
	var out []externalMerge
	for _, ir := range rs.All() {
		rec := ir.Card
		if rec == nil || ir.Duplicate {
			continue
		}
		status := ir.Status()
		if vocab.Issue().IsOpen(status) || vocab.Issue().IsTerminal(status) {
			continue
		}
		f := mergeFacts{EvidenceOnMain: settled[rec.ID]}
		branch := issue.BranchName(rec.Path)
		tip, err := branchTip(env, branch)
		if err != nil {
			return nil, err
		}
		if tip != "" {
			if onMain, err := env.ancestorOf(tip, mainTip); err != nil {
				return nil, err
			} else if onMain {
				if firstParent == nil {
					if firstParent, err = firstParentSet(env, mainTip); err != nil {
						return nil, err
					}
				}
				f.BranchMerged = !firstParent[tip]
			}
		}
		if f.Owner, _, _, err = ownership(env, *rec); err != nil {
			return nil, err
		}
		if v := externalMergeVerdict(f); v != mergeNone {
			out = append(out, externalMerge{ID: rec.ID, CardPath: rec.Path, Branch: branch, Tip: tip, Verdict: v})
		}
	}
	return out, nil
}

// branchTip is the issue branch's newest known tip: the local branch when it
// contains the remote-tracking one (work not pushed yet, or a tracking ref
// left stale by work elsewhere), else the remote-tracking one; "" for none.
func branchTip(env *trackerEnv, branch string) (string, error) {
	remote, err := env.git("for-each-ref", "--format=%(objectname)", gitx.RemoteTrackingRef(env.target.Remote, branch))
	if err != nil {
		return "", err
	}
	local, err := env.git("for-each-ref", "--format=%(objectname)", "refs/heads/"+branch)
	if err != nil || local == "" || remote == "" {
		return valueOr(remote, local), err
	}
	if ahead, err := env.ancestorOf(remote, local); err != nil || ahead {
		return local, err
	}
	return remote, nil
}

// firstParentSet is main's first-parent line, read once per judgement.
func firstParentSet(env *trackerEnv, mainTip string) (map[string]bool, error) {
	out, err := env.git("rev-list", "--first-parent", mainTip)
	if err != nil {
		return nil, err
	}
	set := map[string]bool{}
	for _, c := range strings.Fields(out) {
		set[c] = true
	}
	return set, nil
}

// externalMergeFinding is `sdlc state`'s report for one verdict: what
// happened and the exact next action. state never acts on it.
func externalMergeFinding(m externalMerge) DriftFinding {
	issueStr := issue.CLIRef(m.ID)
	msg := ""
	switch m.Verdict {
	case mergeSettle:
		msg = fmt.Sprintf("its close is on main but the card is not done (merged outside sdlc, or an interrupted merge or push) — `sdlc issue recovery reconcile --issue %s` finishes it (for an interrupted `sdlc merge`, rerunning `sdlc merge --branch %s --yes` also does)", issueStr, m.Branch)
	case mergeCloseNeeded:
		msg = fmt.Sprintf("%s was merged into main outside sdlc without a close — from the owner's slot, `sdlc close --issue %s` on a branch from main, then `sdlc pr` and `sdlc merge`", m.Branch, issueStr)
	case mergeClaimNeeded:
		msg = fmt.Sprintf("%s was merged into main outside sdlc, and #%s has no owner — `sdlc claim --issue %s`, then `sdlc close --issue %s`", m.Branch, issueStr, issueStr, issueStr)
	}
	return DriftFinding{Severity: "warn", Issue: m.ID, Message: msg, kind: driftExternalMerge}
}

// externalMergeFindings is state's read of outside merges. It adds no fetch:
// the cards are the records state already read, and main is as last fetched.
// Anything that keeps it from judging is reported, never fatal.
func externalMergeFindings(ctx context.Context, root, issuesDir string) []DriftFinding {
	if tracked, err := repositoryTracked(ctx, root); err != nil || !tracked {
		return nil
	}
	notChecked := func(err error) []DriftFinding {
		return []DriftFinding{{Severity: "info", Message: "merges outside sdlc not checked: " + err.Error()}}
	}
	rs, err := loadIssueRecords(ctx, issuesDir, tracker.PreferFresh)
	if err != nil {
		return notChecked(err)
	}
	env, err := openTrackerAt(ctx, root)
	if err != nil {
		return notChecked(err)
	}
	view, ok, err := env.main.LocalView()
	if err != nil || !ok {
		if err == nil {
			err = fmt.Errorf("main was never fetched")
		}
		return notChecked(err)
	}
	merges, err := externalMerges(env, rs, view.Ref())
	if err != nil {
		return notChecked(err)
	}
	out := make([]DriftFinding, 0, len(merges))
	for _, m := range merges {
		f := externalMergeFinding(m)
		if rs.Stale { // judged on the cards as last fetched
			f.Message += " (" + staleTrackerNote + ")"
		}
		out = append(out, f)
	}
	return out
}

// finishLandedLeftovers finishes the bookkeeping a merge done outside sdlc
// left (#287), in two convergent sweeps. Every done card whose details are
// still live on main gets them (and its plans) archived in a narrow main
// commit — a landing archives its own issues first, so what remains landed
// elsewhere. Then every done card's issue branch still on the remote is
// deleted once main contains it, from one listing of the remote's heads, so
// a delete that failed earlier is retried by the next run.
func finishLandedLeftovers(env *trackerEnv, stderr io.Writer, dirs archiveDirs) error {
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	view, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	live := map[string]bool{}
	files, err := view.Files(dirs.Issues)
	if err != nil {
		return err
	}
	for _, f := range files {
		live[path.Base(f.Path)] = true
	}
	var done []tracker.Record
	for _, rec := range snap.Records() {
		if status, _ := issue.GetField(rec.Card.Frontmatter, "status"); status == "done" {
			done = append(done, rec)
		}
	}
	for _, rec := range done {
		base := path.Base(rec.Path)
		if !live[base] {
			continue
		}
		spec := mainArchive{
			Message: fmt.Sprintf("#%s: issue: archive after a merge outside sdlc", issue.CLIRef(rec.ID)),
			Dirs:    dirs,
			Allow: func(c tracker.Record) error { // the card is the authority: archive done work only
				if s, _ := issue.GetField(c.Card.Frontmatter, "status"); s != "done" {
					return fmt.Errorf("#%s is %s, no longer done; not archiving", issue.CLIRef(c.ID), s)
				}
				return nil
			},
		}
		if err := archiveIssueOnMain(env, stderr, rec.ID, path.Join(dirs.Issues, base), spec); err != nil {
			return fmt.Errorf("archive #%s: %w", issue.CLIRef(rec.ID), err)
		}
	}
	heads, err := env.git("ls-remote", "--heads", env.target.Remote)
	if err != nil {
		return err
	}
	remote := map[string]string{}
	for _, line := range strings.Split(heads, "\n") {
		if f := strings.Fields(line); len(f) == 2 {
			remote[strings.TrimPrefix(f[1], "refs/heads/")] = f[0]
		}
	}
	for _, rec := range done {
		branch := issue.BranchName(rec.Path)
		tip := remote[branch]
		if tip == "" {
			continue
		}
		if have, err := env.gitTest("cat-file", "-e", tip+"^{commit}"); err != nil || !have {
			continue // never fetched here: leave it rather than guess
		}
		if merged, err := env.ancestorOf(tip, view.Ref()); err != nil || !merged {
			continue
		}
		if err := deleteRemoteBranch(env.git, env.target.Remote, branch, tip); err != nil {
			cwarn(stderr, fmt.Sprintf("#%s is done, but %s was not deleted on %s: %v — the next reconcile or merge retries", issue.CLIRef(rec.ID), branch, env.target.Remote, err))
			continue
		}
		cok(stderr, fmt.Sprintf("%s (merged into main) deleted on %s", branch, env.target.Remote))
	}
	return nil
}

// reportUnclosedMerge prints reconcile's next action for an issue merged
// outside sdlc without its close: there is no bookkeeping to finish until the
// owner closes it, so nothing changes.
func reportUnclosedMerge(ctx context.Context, env *trackerEnv, stderr io.Writer, issuesDir, id string) {
	notChecked := func(err error) { cwarn(stderr, fmt.Sprintf("merges outside sdlc not checked: %v", err)) }
	rs, err := loadIssueRecords(ctx, issuesDir, tracker.Fresh)
	if err != nil {
		notChecked(err)
		return
	}
	view, err := env.main.Snapshot()
	if err != nil {
		notChecked(err)
		return
	}
	merges, err := externalMerges(env, rs, view.Ref())
	if err != nil {
		notChecked(err)
		return
	}
	for _, m := range merges {
		if m.ID == id && m.Verdict != mergeSettle {
			cwarn(stderr, externalMergeFinding(m).Message)
		}
	}
}
