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

// externalMerges judges every started, unfinished card against mainTip. A
// branch counts as merged when its pushed tip (the local branch when it was
// never pushed) is on main but off main's first-parent line: the GitHub
// button merges with a merge commit, while a fresh branch with no work of its
// own sits on that line and is not mistaken for a merge.
func externalMerges(env *trackerEnv, mainTip string) ([]externalMerge, error) {
	snap, err := env.repo.Snapshot()
	if err != nil {
		return nil, err
	}
	var firstParent map[string]bool
	var out []externalMerge
	for _, rec := range snap.Records() {
		status, _ := issue.GetField(rec.Card.Frontmatter, "status")
		if vocab.Issue().IsOpen(status) || vocab.Issue().IsTerminal(status) {
			continue
		}
		f := mergeFacts{}
		if b, ok, err := issue.CardCompletion(rec.Raw); err != nil {
			return nil, fmt.Errorf("card #%s: %w", rec.ID, err)
		} else if ok && status == "codecomplete" && b.Repository == env.target.Repository {
			if f.EvidenceOnMain, err = env.closeAncestorOf(b.EvidenceCommit, mainTip); err != nil {
				return nil, err
			}
		}
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
		if f.Owner, _, _, err = ownership(env, rec); err != nil {
			return nil, err
		}
		if v := externalMergeVerdict(f); v != mergeNone {
			out = append(out, externalMerge{ID: rec.ID, CardPath: rec.Path, Branch: branch, Tip: tip, Verdict: v})
		}
	}
	return out, nil
}

// branchTip is the issue branch's pushed tip, else its local one, else "".
func branchTip(env *trackerEnv, branch string) (string, error) {
	for _, ref := range []string{gitx.RemoteTrackingRef(env.target.Remote, branch), "refs/heads/" + branch} {
		tip, err := env.git("for-each-ref", "--format=%(objectname)", ref)
		if err != nil || tip != "" {
			return tip, err
		}
	}
	return "", nil
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
		msg = fmt.Sprintf("merged outside sdlc with its close on main — finish the bookkeeping (done, archive, branch) with `sdlc issue recovery reconcile --issue %s`", issueStr)
	case mergeCloseNeeded:
		msg = fmt.Sprintf("%s was merged into main outside sdlc without a close — from the owner's slot, `sdlc close --issue %s` on a branch from main, then `sdlc pr` and `sdlc merge`", m.Branch, issueStr)
	case mergeClaimNeeded:
		msg = fmt.Sprintf("%s was merged into main outside sdlc, and #%s has no owner — `sdlc claim --issue %s`, then `sdlc close --issue %s`", m.Branch, issueStr, issueStr, issueStr)
	}
	return DriftFinding{Severity: "warn", Issue: m.ID, Message: msg}
}

// externalMergeFindings is state's read of outside merges in a tracked
// repository; anything that keeps it from judging is reported, never fatal.
func externalMergeFindings(ctx context.Context, root string) []DriftFinding {
	if tracked, err := repositoryTracked(ctx, root); err != nil || !tracked {
		return nil
	}
	env, err := openTrackerAt(ctx, root)
	if err != nil {
		return []DriftFinding{{Severity: "info", Message: "merges outside sdlc not checked: " + err.Error()}}
	}
	view, err := env.main.Snapshot()
	if err != nil {
		return []DriftFinding{{Severity: "info", Message: "merges outside sdlc not checked: " + err.Error()}}
	}
	merges, err := externalMerges(env, view.Ref())
	if err != nil {
		return []DriftFinding{{Severity: "info", Message: "merges outside sdlc not checked: " + err.Error()}}
	}
	out := make([]DriftFinding, 0, len(merges))
	for _, m := range merges {
		out = append(out, externalMergeFinding(m))
	}
	return out
}

// finishLandedLeftovers finishes the bookkeeping a merge done outside sdlc
// left (#287): every done card whose details are still live on main gets them
// (and its plans) archived in a narrow main commit, and its issue branch is
// deleted on the remote once main contains it. A landing archives its own
// issues first, so what remains is exactly what landed elsewhere. Convergent:
// a rerun finds nothing live. It returns the issues it archived.
func finishLandedLeftovers(env *trackerEnv, stderr io.Writer, issuesDir string) ([]string, error) {
	snap, err := env.repo.Snapshot()
	if err != nil {
		return nil, err
	}
	view, err := env.main.Snapshot()
	if err != nil {
		return nil, err
	}
	live := map[string]bool{}
	files, err := view.Files(issuesDir)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		live[path.Base(f.Path)] = true
	}
	var finished []string
	for _, rec := range snap.Records() {
		status, _ := issue.GetField(rec.Card.Frontmatter, "status")
		base := path.Base(rec.Path)
		if status != "done" || !live[base] {
			continue
		}
		spec := mainArchive{
			Message: fmt.Sprintf("#%s: issue: archive after a merge outside sdlc", issue.CLIRef(rec.ID)),
			Allow: func(c tracker.Record) error { // the card is the authority: archive done work only
				if s, _ := issue.GetField(c.Card.Frontmatter, "status"); s != "done" {
					return fmt.Errorf("#%s is %s, no longer done; not archiving", issue.CLIRef(c.ID), s)
				}
				return nil
			},
		}
		if err := archiveIssueOnMain(env, stderr, rec.ID, path.Join(issuesDir, base), spec); err != nil {
			return finished, fmt.Errorf("archive #%s: %w", issue.CLIRef(rec.ID), err)
		}
		finished = append(finished, rec.ID)
		branch := issue.BranchName(rec.Path)
		tip, err := remoteRefTip(env.git, env.target.Remote, "refs/heads/"+branch)
		if err != nil || tip == "" {
			continue
		}
		if have, err := env.gitTest("cat-file", "-e", tip+"^{commit}"); err != nil || !have {
			continue // never fetched here: leave it rather than guess
		}
		if merged, err := env.ancestorOf(tip, view.Ref()); err == nil && merged {
			if err := deleteRemoteBranch(env.git, env.target.Remote, branch, tip); err != nil {
				cwarn(stderr, fmt.Sprintf("#%s archived, but %s was not deleted on %s: %v", issue.CLIRef(rec.ID), branch, env.target.Remote, err))
			} else {
				cok(stderr, fmt.Sprintf("%s (merged into main) deleted on %s", branch, env.target.Remote))
			}
		}
	}
	return finished, nil
}

// reportUnclosedMerge prints reconcile's next action for an issue merged
// outside sdlc without its close: there is no bookkeeping to finish until the
// owner closes it, so nothing changes.
func reportUnclosedMerge(env *trackerEnv, stderr io.Writer, id string) {
	view, err := env.main.Snapshot()
	if err != nil {
		return
	}
	merges, err := externalMerges(env, view.Ref())
	if err != nil {
		cwarn(stderr, fmt.Sprintf("merges outside sdlc not checked: %v", err))
		return
	}
	for _, m := range merges {
		if m.ID == id && m.Verdict != mergeSettle {
			cwarn(stderr, externalMergeFinding(m).Message)
		}
	}
}
