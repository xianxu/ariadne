package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/pkg/agentenv"
	"github.com/xianxu/ariadne/pkg/reviewpolicy"
)

// lookPath finds a reviewer CLI; tests replace it so no result depends on the
// machine.
var lookPath = exec.LookPath

// gitIn runs git in dir; tests use a real temporary repository.
var gitIn = func(dir string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	return string(out), err
}

// docPolicy is the reviewer policy of the repository holding the document
// (#441, sharing #409's): its workshop/review.json, or the default outside a
// repository or without one.
func docPolicy(dir string) (reviewpolicy.Policy, error) {
	root, err := gitIn(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return reviewpolicy.Policy{}, nil
	}
	return reviewpolicy.Load(strings.TrimSpace(root))
}

// docAuthors is every family that wrote the document: the Agent: trailers of
// the commits that touched it, and always this session's. When its history
// cannot be read, unread says why, recorded as degraded.
func docAuthors(file string, env agentenv.Env) (authors []reviewpolicy.Family, unread string) {
	session := reviewpolicy.FamilyOf(env.Main())
	out, err := gitIn(filepath.Dir(file), "log", "--format="+agentenv.LogFormat, "--", filepath.Base(file))
	if err != nil {
		return []reviewpolicy.Family{session}, "authors unread: no git history for the document"
	}
	return reviewpolicy.AuthorFamilies(session, agentenv.ParseLog(out)), ""
}

// chooseReviewers is doc-review's choice under #409's policy. doc-review exists
// to get a second vendor, so it never reviews a document with its author's own
// family: a repo's `same` (or the default) counts as `other`, and any other
// candidate of an author family (an `explicit` pair that maps back to it, a
// session with no pair) is moved to the skips, so an empty choice refuses. An
// agent named on the command line wins, recorded as degraded when it shares a
// family. AGENT_CMD is sdlc's judge override, not doc-review's, so it is ignored.
func chooseReviewers(policy reviewpolicy.Policy, explicit string, explicitSet bool, env agentenv.Env, authors []reviewpolicy.Family, available reviewpolicy.Availability) reviewpolicy.Choice {
	if policy.EffectiveMode() == reviewpolicy.Same {
		policy.Mode = reviewpolicy.Other
	}
	env.AgentCmd = ""
	choice := reviewpolicy.Resolve(explicit, explicitSet, env, policy, authors, reviewpolicy.DesignHeavy, available)
	if choice.Forced {
		return choice
	}
	authored := map[reviewpolicy.Family]bool{}
	for _, f := range authors {
		authored[f] = true
	}
	var kept []reviewpolicy.Candidate
	for _, c := range choice.Candidates {
		if authored[c.Family] {
			choice.Skips = append(choice.Skips, reviewpolicy.Skip{Agent: c.Agent, Reason: "shares the author family " + string(c.Family) + " (doc-review never self-reviews)"})
			continue
		}
		kept = append(kept, c)
	}
	choice.Candidates = kept
	return choice
}

// onPath is the pre-launch availability: the reviewer's CLI is installed.
func onPath(a reviewpolicy.Agent) (string, bool) {
	if _, err := lookPath(string(a)); err != nil {
		return string(a) + " is not on PATH", false
	}
	return "", true
}

// sessionEnv reads this session's agent signals.
func sessionEnv() agentenv.Env { return agentenv.EnvFrom(os.Getenv) }
