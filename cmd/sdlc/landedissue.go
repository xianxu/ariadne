// landedissue.go — an issue whose own branch already landed (#483). One issue
// keeps one branch (#272), so a landed branch is not continued: the issue
// closes with what is verified, from a follow-up branch cut from main, and the
// rest becomes a follow-up issue.
package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// landedPR picks the merged PR that landed an issue branch: one MERGED and no
// OPEN PR for the name (an open one means the branch is still in flight).
// Pure.
func landedPR(prs []landingPR) (landingPR, bool) {
	var merged landingPR
	found := false
	for _, pr := range prs {
		switch pr.State {
		case "OPEN":
			return landingPR{}, false
		case "MERGED":
			if !found {
				merged, found = pr, true
			}
		}
	}
	return merged, found
}

// issueBranchLanded reports whether the issue branch name has landed: GitHub
// holds a merged PR for it (a squash merge included), and a local copy of the
// branch, if any, holds nothing beyond that PR's head or main. Any failure to
// read the evidence reads as not landed, which keeps the verbs' older paths.
func issueBranchLanded(env *trackerEnv, name string) (landingPR, bool) {
	gh, ok := ghClient.(landingGH)
	if !ok {
		return landingPR{}, false
	}
	raw, err := env.git("config", "--get", "remote."+env.target.Remote+".url")
	if err != nil {
		return landingPR{}, false
	}
	repo, err := landingRepoURL(raw)
	if err != nil {
		return landingPR{}, false
	}
	prs, err := gh.LandingPRs(env.ctx, repo, name)
	if err != nil {
		return landingPR{}, false
	}
	pr, ok := landedPR(prs)
	if !ok {
		return landingPR{}, false
	}
	tip, err := env.git("for-each-ref", "--format=%(objectname)", "refs/heads/"+name)
	if err != nil {
		return landingPR{}, false
	}
	if tip == "" {
		return pr, true
	}
	for _, base := range []string{pr.HeadOID, gitx.RemoteTrackingRef(env.target.Remote, "main")} {
		if within, err := env.gitTest("merge-base", "--is-ancestor", tip, base); err == nil && within {
			return pr, true
		}
	}
	return landingPR{}, false
}

// landedRecovery is the next action for a landed issue, shared by every verb
// that meets one.
func landedRecovery(id, name, remote string, pr landingPR) string {
	ref := issue.CLIRef(id)
	return fmt.Sprintf("#%s's branch %s already landed (PR #%d); one issue keeps one branch (#272), so it is not continued.\n"+
		"      Close #%s with what is verified, from a follow-up branch cut from main:\n"+
		"        git switch -c %s-<suffix> %s/main && sdlc close --issue %s, then `sdlc pr` and `sdlc merge`\n"+
		"      File what is left as a follow-up: sdlc issue new '<title>' --deps %s",
		ref, name, pr.Number, ref, strings.SplitN(name, "-", 2)[0], remote, ref, ref)
}

// refuseLandedIssue refuses to continue an issue whose branch landed.
func refuseLandedIssue(env *trackerEnv, id, name string) error {
	if pr, landed := issueBranchLanded(env, name); landed {
		return errors.New(landedRecovery(id, name, env.target.Remote, pr))
	}
	return nil
}
