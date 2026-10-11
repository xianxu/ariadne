// transferguard.go — optimistic concurrency on published details (#252, #285).
// Once an issue's details are on main, a landing may change them only when no
// other workspace owns the issue (the card's claimant; commits don't carry it,
// and an unclaimed issue has no owner to protect, #274) and only from a branch
// based on main's latest version of the file. Otherwise the prospective merge
// must leave the file exactly as main has it. Branch names play no part: a
// renamed or reopened issue branch, or another branch of the owner's, lands.
package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// guardTransferredDetailsFn is the gates' seam: publish-gate unit fixtures model
// anchor logic in remote-less repositories and stub it; the guard itself is
// exercised against real remotes in transferguard_test.go.
var guardTransferredDetailsFn = guardTransferredDetails

// guardTransferredDetails refuses a landing (PR, push, merge) whose
// prospective merge into fresh main changes published details it may not.
func guardTransferredDetails(ctx context.Context) error {
	// Decided before opening the tracker environment, which needs main's
	// upstream: a legacy repository has no published-details protection.
	if tracked, err := repositoryTracked(ctx, "."); err != nil || !tracked {
		return err
	}
	env, err := openTracker(ctx)
	if err != nil {
		return err
	}
	view, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	changed, err := changedDetails(env, view.Ref())
	if err != nil {
		return err
	}
	var refusals []error
	for _, c := range changed {
		refusals = append(refusals, detailsRefusal(env, c))
	}
	return errors.Join(refusals...)
}

// changedDetail is one published details file a landing may not change as it
// stands, with the verdict that says why.
type changedDetail struct {
	ID, Path string
	Verdict  verdict
	Own      issue.Ownership // this workspace's relation to the card
	Owner    issue.Claimant  // the card's claimant, zero when unclaimed
}

// changedDetails judges every issue details file that the prospective merge
// of mainTip (fresh main) and HEAD changes, returning those it refuses
// (sorted by path). guardTransferredDetails refuses them; `issue restore`
// reverts them.
func changedDetails(env *trackerEnv, mainTip string) ([]changedDetail, error) {
	snap, err := env.repo.Snapshot()
	if err != nil {
		return nil, err
	}
	paths, err := prospectiveChanges(env, mainTip)
	if err != nil {
		return nil, err
	}
	// A details file is named after its card. An unreadable card's details
	// cannot be judged (its owner is unknown): fail closed on those only.
	cards := map[string]tracker.Record{}
	for _, rec := range snap.Records() {
		cards[path.Base(rec.Path)] = rec
	}
	unreadable := map[string]tracker.UnreadableCard{}
	for _, u := range snap.Unreadable() {
		unreadable[path.Base(u.Path)] = u
	}
	var out []changedDetail
	for _, p := range paths {
		name := path.Base(p)
		if u, bad := unreadable[name]; bad {
			return nil, fmt.Errorf("landing changes %s, but tracker card #%s is malformed: %w\n"+
				"      the landing is refused until the card is repaired", p, u.ID, u.Err)
		}
		rec, ok := cards[name]
		if !ok {
			continue // not an issue's details
		}
		last, err := env.git("rev-list", "-1", mainTip, "--", p)
		if err != nil {
			return nil, err
		}
		facts := detailsFacts{Published: last != ""}
		var own issue.Ownership
		var owner issue.Claimant
		if facts.Published {
			if own, owner, _, err = ownership(env, rec); err != nil {
				return nil, err
			}
			facts.Foreign = mayNotEditDetails(own)
			if facts.Based, err = env.ancestorOf(last, "HEAD"); err != nil {
				return nil, err
			}
		}
		if v := detailsVerdict(facts); v != verdictAccept {
			out = append(out, changedDetail{ID: rec.ID, Path: p, Verdict: v, Own: own, Owner: owner})
		}
	}
	return out, nil
}

// prospectiveChanges lists the paths whose content or presence in the merge
// of mainTip and HEAD differs from mainTip. A conflicted path is among them:
// merge-tree (exit 1) writes it with conflict markers.
func prospectiveChanges(env *trackerEnv, mainTip string) ([]string, error) {
	raw, err := env.gitRaw(nil, "merge-tree", "--write-tree", "--name-only", "--no-messages", "-z", mainTip, "HEAD")
	var exit *exec.ExitError
	if err != nil && !(errors.As(err, &exit) && exit.ExitCode() == 1) {
		return nil, fmt.Errorf("compute the prospective merge with main: %w", err)
	}
	tree, _, _ := strings.Cut(string(raw), "\x00")
	if tree == "" {
		return nil, fmt.Errorf("compute the prospective merge with main: no tree (%v)", err)
	}
	diff, err := env.gitRaw(nil, "diff", "--name-only", "--no-renames", "-z", mainTip, tree)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, p := range strings.Split(string(diff), "\x00") {
		if p != "" {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)
	return paths, nil
}

var errTransferredDetails = errors.New("landing would change published issue details")

// mayNotEditDetails is the one policy both writers of published details apply
// (this guard and `issue publish`, #274): another workspace's issue is theirs;
// this workspace's own and an unclaimed one are editable.
func mayNotEditDetails(own issue.Ownership) bool {
	return own == issue.OwnershipForeign
}

// detailsRefusal names the next action for one refused change. Another
// workspace's issue: hand the edit to its owner, or drop it here. Behind main:
// merge main and resolve the prose; an unclaimed issue's stale copy may also be
// dropped, but restore is never offered over this workspace's own record.
func detailsRefusal(env *trackerEnv, c changedDetail) error {
	issueStr := issue.CLIRef(c.ID)
	if c.Verdict == verdictForeign {
		return fmt.Errorf("%w: %s, but #%s is owned by %s.\n"+
			"      Hand the edit to its owner (reassignment is the operator-directed `sdlc reclaim --issue %s`), "+
			"or drop it here with `sdlc issue restore --issue %s`",
			errTransferredDetails, c.Path, issueStr, describeClaimant(c.Owner), issueStr, issueStr)
	}
	drop := ""
	if mayRestore(c) {
		drop = fmt.Sprintf("; or drop this branch's copy with `sdlc issue restore --issue %s`", issueStr)
	}
	return fmt.Errorf("%w: %s, from a branch not based on main's latest version of it.\n"+
		"      Merge main (`git fetch %s && git merge %s/main`), resolve #%s's details, and rerun%s",
		errTransferredDetails, c.Path, env.target.Remote, env.target.Remote, issueStr, drop)
}

// mayRestore says whether `issue restore` may drop a refused change: anything
// but this workspace's own issue, whose edit is a record merging main keeps.
func mayRestore(c changedDetail) bool {
	return c.Own != issue.OwnershipMine
}
