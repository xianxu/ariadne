package reviewpolicy

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/pkg/agentenv"
)

// Choice is a review's resolved reviewers: the candidates in order, the ones
// skipped before launch, and what the choice was made from.
type Choice struct {
	Candidates []Candidate
	Skips      []Skip
	Authors    []Family
	Forced     bool     // an explicit agent or AGENT_CMD named the reviewer
	Degraded   []string // inputs the choice could not read (the window's authors), recorded with the review
}

// Resolve is every reviewer site's choice: an explicit agent or AGENT_CMD is
// one forced candidate; otherwise the policy plans the candidates from the
// authors' families and the change kind.
func Resolve(explicit string, explicitSet bool, env agentenv.Env, policy Policy, authors []Family, kind ChangeKind, available Availability) Choice {
	choice := Choice{Authors: authors}
	forced := ""
	switch {
	case explicitSet:
		forced = explicit
	case env.AgentCmd != "":
		forced = env.AgentCmd
	}
	if forced != "" {
		a := Agent(forced)
		choice.Forced = true
		choice.Candidates = []Candidate{{Agent: a, Model: policy.Models[a], Family: FamilyOf(forced)}}
		return choice
	}
	choice.Candidates, choice.Skips = Plan(policy, Agent(env.Main()), authors, kind, available)
	return choice
}

// First is the reviewer a dry run shows and a single-reviewer caller uses.
// ok is false when no candidate is left; a caller then reports a refusal
// (sdlc: judge.Refusal; doc-review: a NoReviewerError of its skips), never
// dispatching the zero Candidate.
func (c Choice) First() (Candidate, bool) {
	if len(c.Candidates) == 0 {
		return Candidate{}, false
	}
	return c.Candidates[0], true
}

// ErrNoReviewer is a review with no reviewer that can run.
var ErrNoReviewer = errors.New("no reviewer available")

// NoReviewerError is a review that could not run because every candidate was
// unavailable: it names each one and why, never quietly reviewing with the
// authors' own family instead.
type NoReviewerError struct{ Skips []Skip }

func (e *NoReviewerError) Error() string {
	var parts []string
	for _, s := range e.Skips {
		parts = append(parts, fmt.Sprintf("%s: %s", s.Agent, s.Reason))
	}
	transient, policy := false, len(parts) == 0
	for _, s := range e.Skips {
		if s.Policy {
			policy = true
		} else {
			transient = true
		}
	}
	if len(parts) == 0 {
		parts = append(parts, "the policy leaves no reviewer of another family")
	}
	// The remedy follows the skips: waiting cures an unavailable CLI, not the
	// policy (#441).
	var remedies []string
	if transient {
		remedies = append(remedies, "retry later")
	}
	if policy {
		remedies = append(remedies, "pair this agent with another family in reviewer_pairs")
	}
	remedy := "name one explicitly (recorded as degraded)"
	if len(remedies) > 0 {
		remedy = strings.Join(remedies, ", ") + ", or " + remedy
	}
	return "review did not run: no reviewer available (" + strings.Join(parts, "; ") + "); " + remedy
}
func (e *NoReviewerError) Unwrap() error { return ErrNoReviewer }

// Record is what a finished review records about its reviewer: sdlc's
// `Review-Reviewer:` trailer, its ledger and sidecar, doc-review's report.
//
//	codex/gpt-6-astra (openai; author anthropic)
//	gemini (google; author anthropic,openai); degraded: codex unavailable: usage limit
//	claude (anthropic; author anthropic); degraded: same-family by choice
func Record(used Candidate, choice Choice, skips []Skip) string {
	who := string(used.Agent)
	if used.Model != "" {
		who += "/" + used.Model
	}
	s := fmt.Sprintf("%s (%s; author %s)", who, orUnknown(string(used.Family), "unknown"), orUnknown(familyList(choice.Authors), "unknown"))
	degraded := append([]string(nil), choice.Degraded...)
	for _, k := range skips {
		degraded = append(degraded, fmt.Sprintf("%s unavailable: %s", k.Agent, oneLine(k.Reason)))
	}
	if choice.Forced && used.Family != "" && containsFamily(choice.Authors, used.Family) {
		degraded = append(degraded, "same-family by choice")
	}
	if len(degraded) > 0 {
		s += "; degraded: " + strings.Join(degraded, "; ")
	}
	return s
}

func familyList(fs []Family) string {
	seen := map[Family]bool{}
	var s []string
	for _, f := range fs {
		if f != "" && !seen[f] {
			seen[f] = true
			s = append(s, string(f))
		}
	}
	sort.Strings(s)
	return strings.Join(s, ",")
}

func containsFamily(fs []Family, f Family) bool {
	for _, x := range fs {
		if x == f {
			return true
		}
	}
	return false
}

// oneLine keeps a reason to one short line for a record.
func oneLine(s string) string {
	s, _, _ = strings.Cut(strings.TrimSpace(s), "\n")
	if len(s) > 100 {
		s = s[:100] + "…"
	}
	return s
}
