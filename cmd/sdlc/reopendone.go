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
	if env.onRest() {
		if warn := fastForwardRest(env, view.Ref()); warn != "" {
			return false, fmt.Errorf("#%s: %s", issueStr, warn)
		}
	}
	// The details are wherever HEAD has them: archived until step 2 commits.
	detailRel := path.Join(issuesDir, base)
	if archived, err := env.has("HEAD", archivedRel); err != nil {
		return false, err
	} else if archived {
		detailRel = archivedRel
	}
	entries, err := env.statusEntries(env.root, "--untracked-files=no")
	if err != nil {
		return false, err
	}
	if blocking := planningDirtyBlocking(entries, detailRel); len(blocking) > 0 {
		return false, fmt.Errorf("#%s: reopening cuts a branch and needs a clean tree besides its details (%s); commit or discard:\n  %s", issueStr, detailRel, strings.Join(blocking, "\n  "))
	}
	if !force {
		raw, err := os.ReadFile(filepath.Join(env.root, filepath.FromSlash(detailRel)))
		if err != nil {
			return false, err
		}
		_, body, err := issue.Parse(string(raw))
		if err != nil {
			return false, fmt.Errorf("%s is malformed (%v); repair it before reopening", detailRel, err)
		}
		if err := checkTransitionGuards("done", "working", "", body); err != nil {
			return false, fmt.Errorf("%w\n  Write it in %s; the reopen moves the details back with it. Nothing was changed", err, detailRel)
		}
	}
	// 1: the issue branch, from main. A local one main contains is the landed
	// branch (nothing to lose); any other is this reopen's, on a rerun.
	if env.branch != name {
		tip, err := env.git("for-each-ref", "--format=%(objectname)", "refs/heads/"+name)
		if err != nil {
			return false, err
		}
		landed := tip == ""
		if !landed {
			if landed, err = env.ancestorOf(tip, view.Ref()); err != nil {
				return false, err
			}
		}
		if landed {
			_, err = env.git("switch", "-q", "--no-track", "-C", name, view.Ref())
		} else {
			_, err = env.git("switch", "-q", name)
		}
		if err != nil {
			return false, err
		}
		env.branch = name
	}
	archiveCommit, err := env.git("rev-list", "-1", view.Ref(), "--", archivedRel)
	if err != nil {
		return false, err
	}
	if merged, err := env.ancestorOf(archiveCommit, "HEAD"); err != nil {
		return false, err
	} else if !merged {
		return false, fmt.Errorf("#%s: %s predates main's archive of its details; merge main into it, then rerun", issueStr, name)
	}
	// 2: the details and plans come back out of the archive.
	if err := unarchiveIssue(env, base, issuesDir, fmt.Sprintf("#%s: reopen: un-archive", issueStr)); err != nil {
		return false, err
	}
	// 3: the branch on the remote, before the card says it is back.
	if err := pushIssueBranch(env, name); err != nil {
		return false, fmt.Errorf("#%s: pushing %s failed: %w — the card is unchanged; rerun", issueStr, name, err)
	}
	cok(stderr, fmt.Sprintf("#%s's details un-archived on %s and pushed", issueStr, name))
	return true, nil
}
