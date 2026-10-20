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

// landedPRs returns the merged PRs that landed an issue branch, or none when
// an OPEN PR for the name says the branch is still in flight. Pure.
func landedPRs(prs []landingPR) []landingPR {
	var merged []landingPR
	for _, pr := range prs {
		switch pr.State {
		case "OPEN":
			return nil
		case "MERGED":
			merged = append(merged, pr)
		}
	}
	return merged
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
	merged := landedPRs(prs)
	if len(merged) == 0 {
		return landingPR{}, false
	}
	// The branch as it stands, local or else remote (#483 BR-1): work pushed
	// after the landing, from any slot, keeps the issue in flight.
	tip, err := issueTip(env, name)
	if err != nil {
		return landingPR{}, false
	}
	if tip == "" {
		return merged[0], true
	}
	bases := []string{gitx.RemoteTrackingRef(env.target.Remote, "main")}
	for _, pr := range merged {
		bases = append(bases, pr.HeadOID)
	}
	for _, base := range bases {
		if within, err := env.gitTest("merge-base", "--is-ancestor", tip, base); err == nil && within {
			return merged[0], true
		}
	}
	return landingPR{}, false
}

// landedRecovery is the next action for a landed issue, shared by every verb
// that meets one.
func landedRecovery(id, name, remote string, pr landingPR) string {
	return fmt.Sprintf("#%s's branch %s already landed (PR #%d); one issue keeps one branch (#272), so it is not continued.\n"+
		"      Close #%s with what is verified, from a follow-up branch cut from main:\n"+
		"        git switch -c %s-<suffix> %s/main && sdlc close --issue %s, then `sdlc pr` and `sdlc merge`\n"+
		"      File what is left as a follow-up: sdlc issue new '<title>' --deps %s",
		issue.CLIRef(id), name, pr.Number, issue.CLIRef(id), strings.SplitN(name, "-", 2)[0], remote, issue.CLIRef(id), issue.CLIRef(id))
}

// refuseLandedIssue refuses to continue an issue whose branch landed.
func refuseLandedIssue(env *trackerEnv, id, name string) error {
	if pr, landed := issueBranchLanded(env, name); landed {
		return errors.New(landedRecovery(id, name, env.target.Remote, pr))
	}
	return nil
}
