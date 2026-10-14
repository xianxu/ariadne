// reopendone.go — reopening a done issue un-archives its details (#305), the
// counterpart of #286's abandon reopen for an issue with no kept branch.
package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// restoreDone is a reopen's work for a done issue whose details main has
// archived, run before set-status's card write: the issue branch is cut from
// main, the details and plans move back out of the archive (with the reopen's
// Log entry, written in the archived copy), and the branch is pushed. false
// when main has not archived them. Each step is recognised on rerun.
func restoreDone(env *trackerEnv, stderr io.Writer, card tracker.Record, issuesDir string, force bool) (bool, error) {
	issueStr := issue.CLIRef(card.ID)
	base, name := path.Base(card.Path), issue.BranchName(card.Path)
	archivedRel := archiveDestination(historyDir(), vocab.ArchiveIssues, base)
	view, err := env.main.Snapshot()
	if err != nil {
		return false, err
	}
	if archived, err := view.Exists(archivedRel); err != nil || !archived {
		return false, err
	}
	if live, err := view.Exists(path.Join(issuesDir, base)); err != nil || live {
		return false, err
	}
	if own, recorded, _, err := ownership(env, card); err != nil {
		return false, err
	} else if own == issue.OwnershipForeign {
		return false, fmt.Errorf("#%s is owned by %s; only that workspace reopens it (reassignment is `sdlc reclaim`)", issueStr, describeClaimant(recorded))
	}
	if env.branch != name && !env.onRest() {
		return false, fmt.Errorf("#%s: reopen from the resting branch (%s) or from %s, not %s", issueStr, env.resting, name, valueOr(env.branch, "a detached HEAD"))
	}
	// Every refusal is decided before the first change (BR-3, BR-5). The issue
	// branch must hold main's archive commit; a local one that lacks it is the
	// landed branch (main contains it: reset to main from rest, nothing to
	// lose) or stale work (merge main into it first).
	archiveCommit, err := env.git("rev-list", "-1", view.Ref(), "--", archivedRel)
	if err != nil {
		return false, err
	}
	tip := "HEAD"
	if env.branch != name {
		if tip, err = env.git("for-each-ref", "--format=%(objectname)", "refs/heads/"+name); err != nil {
			return false, err
		}
	}
	reset := tip == ""
	if !reset {
		holds, err := env.ancestorOf(archiveCommit, tip)
		if err != nil {
			return false, err
		}
		landed, err := env.ancestorOf(tip, view.Ref())
		if err != nil {
			return false, err
		}
		switch {
		case holds:
		case !landed:
			return false, fmt.Errorf("#%s: %s predates main's archive of its details; merge main into it, then rerun", issueStr, name)
		case env.branch == name:
			return false, fmt.Errorf("#%s: %s is the landed branch, without main's archive of its details; switch to %s and rerun, which cuts it again from main", issueStr, name, env.resting)
		default:
			reset = true
		}
	}
	// The lease, decided before any change: origin's branch must be main's
	// or this reopen's (the branch the reopen builds on).
	ours := view.Ref()
	if !reset {
		ours = tip
	}
	// Pin the checkout target and check both the archive source and its future
	// live asset paths before either the resting fast-forward or branch switch.
	materializedTip, err := env.git("rev-parse", ours+"^{commit}")
	if err != nil {
		return false, err
	}
	materialized := env.main.ViewOf(materializedTip)
	if err := verifyAssetTransition(env, materializedTip, issuesDir); err != nil {
		return false, err
	}
	if archived, err := materialized.Exists(archivedRel); err != nil {
		return false, err
	} else if archived {
		raw, err := materialized.Read(archivedRel)
		if err != nil {
			return false, err
		}
		bundle, err := assetFilesFromView(materialized, archivedRel, card.ID, raw)
		if err != nil {
			return false, err
		}
		for member := range bundle.Files {
			for _, ref := range []string{"", "HEAD", materializedTip, view.Ref()} {
				if err := verifyAssetAttributesAt(env.ctx, env.root, path.Join(issuesDir, member), ref); err != nil {
					return false, err
				}
			}
		}
	}
	if err := leaseOnRemoteBranch(env, name, view.Ref(), ours); err != nil {
		return false, err
	}
	// A dirty tree refuses before the fast-forward; the reopen's own details
	// edit (archived or live copy) is the one change allowed.
	entries, err := env.statusEntries(env.root, "--untracked-files=no")
	if err != nil {
		return false, err
	}
	liveRel := path.Join(issuesDir, base)
	if blocking := dirtyBesides(entries, archivedRel, liveRel); len(blocking) > 0 {
		return false, fmt.Errorf("#%s: reopening cuts a branch and needs a clean tree besides its details; commit or discard:\n  %s", issueStr, strings.Join(blocking, "\n  "))
	}
	// The archived copy needs rest at main. Every refusal after this
	// fast-forward names it (BR-6), through after.
	moved := ""
	if env.onRest() {
		before, err := env.git("rev-parse", "HEAD")
		if err != nil {
			return false, err
		}
		if warn := fastForwardRest(env, view.Ref()); warn != "" {
			return false, fmt.Errorf("#%s: %s", issueStr, warn)
		}
		if before != view.Ref() {
			moved = fmt.Sprintf("\n  (%s was fast-forwarded to main, which the archived copy needs; nothing else was changed)", env.resting)
		}
	}
	after := func(err error) (bool, error) { return false, fmt.Errorf("%w%s", err, moved) }
	// The details are wherever HEAD has them: archived until step 2 commits.
	detailRel := liveRel
	if archived, err := env.has("HEAD", archivedRel); err != nil {
		return after(err)
	} else if archived {
		detailRel = archivedRel
	}
	if !force {
		raw, err := os.ReadFile(filepath.Join(env.root, filepath.FromSlash(detailRel)))
		if err != nil {
			return after(err)
		}
		_, body, err := issue.Parse(string(raw))
		if err != nil {
			return after(fmt.Errorf("%s is malformed (%v); repair it before reopening", detailRel, err))
		}
		if err := checkTransitionGuards("done", "working", "", body); err != nil {
			return after(fmt.Errorf("%w\n  Write it in %s; the reopen moves the details back with it", err, detailRel))
		}
	}
	// 1: the issue branch.
	if env.branch != name {
		if reset {
			_, err = env.git("switch", "-q", "--no-track", "-C", name, view.Ref())
		} else {
			_, err = env.git("switch", "-q", name)
		}
		if err != nil {
			return false, err
		}
		env.branch = name
	}
	if err := verifyMaterializedAssets(env.root, materialized, issuesDir); err != nil {
		return false, fmt.Errorf("#%s: checkout left invalid evidence: %w; restore the named assets from %s before retrying reopen", issueStr, err, materializedTip)
	}
	// 2: the details and plans come back out of the archive.
	if err := unarchiveIssue(env, base, issuesDir, unarchiveSubject(card.ID)); err != nil {
		return false, err
	}
	// The postcondition, before anything is published (BR-5, BR-7).
	if err := checkUnarchived(env, archiveCommit, liveRel, archivedRel); err != nil {
		return false, fmt.Errorf("#%s: %s %w; nothing was pushed", issueStr, name, err)
	}
	// 3: the branch on the remote, before the card says it is back.
	if err := pushIssueBranch(env, name); err != nil {
		return false, fmt.Errorf("#%s: pushing %s failed: %w — the card is unchanged; rerun", issueStr, name, err)
	}
	cok(stderr, fmt.Sprintf("#%s's details un-archived on %s and pushed", issueStr, name))
	return true, nil
}

// leaseOnRemoteBranch points the push lease at what origin holds now for the
// landed name (BR-1): the tracking ref this checkout kept from before the
// landing names a branch origin has since deleted or that main contains, and
// leasing on it would refuse every rerun. A tip that is neither main's nor
// this reopen's is someone's work; refuse.
func leaseOnRemoteBranch(env *trackerEnv, name, mainTip, ours string) error {
	tracking := gitx.RemoteTrackingRef(env.target.Remote, name)
	remote, err := remoteRefTip(env.git, env.target.Remote, "refs/heads/"+name)
	if err != nil {
		return err
	}
	if remote == "" {
		_, err = env.git("update-ref", "-d", tracking)
		return err
	}
	if _, err := env.git("fetch", "-q", env.target.Remote, "+refs/heads/"+name+":"+tracking); err != nil {
		return err
	}
	for _, into := range []string{mainTip, ours} {
		if ok, err := env.ancestorOf(remote, into); err != nil || ok {
			return err
		}
	}
	return fmt.Errorf("%s's %s (%s) holds work neither main nor this reopen has; inspect it before reopening", env.target.Remote, name, shortOID(remote))
}

func unarchiveSubject(id string) string {
	return fmt.Sprintf("#%s: reopen: un-archive", issue.CLIRef(id))
}

// unfinishedDoneReopen recognises a done reopen whose card write landed but
// whose finish (the mirror commit and push) did not (BR-2): this checkout is on
// the issue branch with nothing after the un-archive commit but that finish's
// own mirror commit. nil otherwise, so later work is never swept into it.
func unfinishedDoneReopen(env *trackerEnv, card tracker.Record) (*reopenedIssue, error) {
	if env.branch != issue.BranchName(card.Path) {
		return nil, nil
	}
	out, err := env.git("log", "-2", "--format=%s")
	if err != nil {
		return nil, err
	}
	subjects := strings.Split(out, "\n")
	if subjects[0] == mirrorSubject(card.ID) {
		subjects = subjects[1:]
	}
	if len(subjects) == 0 || subjects[0] != unarchiveSubject(card.ID) {
		return nil, nil
	}
	return &reopenedIssue{env: env, id: card.ID, cardPath: card.Path}, nil
}

// dirtyBesides is the dirty tracked paths that block a reopen: those that
// would block a plain edit of either copy of its details. Pure.
func dirtyBesides(entries []gitx.StatusEntry, archivedRel, liveRel string) []string {
	live := map[string]bool{}
	for _, p := range planningDirtyBlocking(entries, liveRel) {
		live[p] = true
	}
	var out []string
	for _, p := range planningDirtyBlocking(entries, archivedRel) {
		if live[p] {
			out = append(out, p)
		}
	}
	return out
}

// checkUnarchived is a done reopen's postcondition (BR-7): HEAD descends from
// main's archive commit — which a branch from before the landing does not,
// though it has the live details too — and holds the details live, not
// archived.
func checkUnarchived(env *trackerEnv, archiveCommit, liveRel, archivedRel string) error {
	if holds, err := env.ancestorOf(archiveCommit, "HEAD"); err != nil {
		return err
	} else if !holds {
		return fmt.Errorf("does not descend from main's archive of its details (%s)", shortOID(archiveCommit))
	}
	if live, err := env.has("HEAD", liveRel); err != nil {
		return err
	} else if archived, err := env.has("HEAD", archivedRel); err != nil {
		return err
	} else if !live || archived {
		return fmt.Errorf("does not hold the un-archived details (%s)", liveRel)
	}
	return nil
}
