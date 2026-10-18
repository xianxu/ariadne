// reviewchoice.go — the one place every review site picks and runs its
// reviewer (#409): resolve the candidates once, try them in order, move on
// only from a reviewer that could not review at all, and say so.
package judge

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

// ErrReviewerUnavailable is a candidate that could not review at all (see
// Unavailable): trying the next one duplicates no review.
var ErrReviewerUnavailable = errors.New("reviewer unavailable")

// UnavailableError names the candidate and the evidence.
type UnavailableError struct {
	Agent  AgentCLI
	Reason string
}

func (e *UnavailableError) Error() string {
	return fmt.Sprintf("%s: %s: %s", ErrReviewerUnavailable, e.Agent, e.Reason)
}
func (e *UnavailableError) Unwrap() error { return ErrReviewerUnavailable }

// ReviewChoice is a review's resolved reviewers: the candidates in order, the
// ones skipped before launch, and what the choice was made from.
type ReviewChoice struct {
	Candidates []Candidate
	Skips      []Skip
	Authors    []Family
	Forced     bool // --agent or AGENT_CMD named the reviewer
}

// ResolveReview is ResolveReviewerCLI for every review site: an explicit
// --agent or AGENT_CMD is one forced candidate; otherwise the policy plans the
// candidates from the authors' families and the change kind.
func ResolveReview(explicit string, explicitSet bool, env AgentDefaultEnv, policy ReviewerPolicy, authors []Family, kind ChangeKind, available Availability) ReviewChoice {
	choice := ReviewChoice{Authors: authors}
	forced := ""
	switch {
	case explicitSet:
		forced = explicit
	case env.AgentCmd != "":
		forced = env.AgentCmd
	}
	if forced != "" {
		a := AgentCLI(forced)
		choice.Forced = true
		choice.Candidates = []Candidate{{Agent: a, Model: policy.Models[a], Family: FamilyOf(forced)}}
		return choice
	}
	choice.Candidates, choice.Skips = ReviewPlan(policy, MainAgent(env), authors, kind, available)
	return choice
}

// First is the reviewer a dry run shows and a single-reviewer caller uses.
func (c ReviewChoice) First() Candidate {
	if len(c.Candidates) == 0 {
		return Candidate{}
	}
	return c.Candidates[0]
}

// NoReviewerError is a review that could not run because every candidate was
// unavailable: it names each one and why, never quietly reviewing with the
// authors' own family instead.
type NoReviewerError struct{ Skips []Skip }

func (e *NoReviewerError) Error() string {
	var parts []string
	for _, s := range e.Skips {
		parts = append(parts, fmt.Sprintf("%s: %s", s.Agent, s.Reason))
	}
	if len(parts) == 0 {
		parts = append(parts, "the policy leaves no reviewer of another family")
	}
	return "review did not run: no reviewer available (" + strings.Join(parts, "; ") +
		"); retry later, or name one with --agent (recorded as degraded)"
}
func (e *NoReviewerError) Unwrap() error { return ErrReviewDidNotRun }

// DispatchChoice runs the review with the first candidate that can review,
// announcing each one passed over. It returns the output, the reviewer that
// produced it and every candidate skipped on the way (with choice.Skips
// first). A failure other than unavailability ends the review as before.
func DispatchChoice(ctx context.Context, opts DispatchOptions, choice ReviewChoice) (string, Candidate, []Skip, error) {
	skips := append([]Skip(nil), choice.Skips...)
	for _, s := range choice.Skips {
		announceSkip(opts.Stderr, s)
	}
	for _, c := range choice.Candidates {
		opts.Agent, opts.Model = c.Agent, c.Model
		out, err := Dispatch(ctx, opts)
		var unavailable *UnavailableError
		if errors.As(err, &unavailable) {
			s := Skip{Agent: c.Agent, Reason: unavailable.Reason}
			skips = append(skips, s)
			announceSkip(opts.Stderr, s)
			continue
		}
		return out, c, skips, err
	}
	return "", Candidate{}, skips, &NoReviewerError{Skips: skips}
}

func announceSkip(w io.Writer, s Skip) {
	if w != nil {
		fmt.Fprintf(w, "  [!] reviewer %s (%s) unavailable: %s; trying the next\n", s.Agent, valueOrDefault(string(FamilyOf(string(s.Agent))), "unknown family"), s.Reason)
	}
}

// ReviewerRecord is what a finished review records about its reviewer, as the
// `Review-Reviewer:` trailer and the ledger round's reviewer field:
//
//	codex/gpt-6-astra (openai; author anthropic)
//	gemini (google; author anthropic,openai); degraded: codex unavailable: usage limit
//	claude (anthropic; author anthropic); degraded: same-family by choice
func ReviewerRecord(used Candidate, choice ReviewChoice, skips []Skip) string {
	who := string(used.Agent)
	if used.Model != "" {
		who += "/" + used.Model
	}
	authors := familyList(choice.Authors)
	s := fmt.Sprintf("%s (%s; author %s)", who, valueOrDefault(string(used.Family), "unknown"), valueOrDefault(authors, "unknown"))
	var degraded []string
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

// ReviewerEntry is one parsed Review-Reviewer trailer.
type ReviewerEntry struct {
	Agent, Model string
	Family       Family
	Authors      []Family
	Degraded     bool
}

// ParseReviewerRecord reads a Review-Reviewer value back; ok is false for a
// value not in ReviewerRecord's form.
func ParseReviewerRecord(v string) (ReviewerEntry, bool) {
	head, _, degraded := strings.Cut(v, "; degraded:")
	open := strings.Index(head, " (")
	if open < 0 || !strings.HasSuffix(head, ")") {
		return ReviewerEntry{}, false
	}
	who, inner := head[:open], head[open+2:len(head)-1]
	fam, authors, ok := strings.Cut(inner, "; author ")
	if !ok {
		return ReviewerEntry{}, false
	}
	e := ReviewerEntry{Family: Family(fam), Degraded: degraded}
	e.Agent, e.Model, _ = strings.Cut(who, "/")
	if authors != "unknown" {
		for _, a := range strings.Split(authors, ",") {
			e.Authors = append(e.Authors, Family(a))
		}
	}
	return e, e.Agent != ""
}

// TallyReviewers counts Review-Reviewer values by "author → reviewer" family
// pair, for a week's retro; values not in the form are counted as "unparsed".
func TallyReviewers(values []string) map[string]int {
	out := map[string]int{}
	for _, v := range values {
		e, ok := ParseReviewerRecord(strings.TrimSpace(v))
		if !ok {
			out["unparsed"]++
			continue
		}
		key := valueOrDefault(familyList(e.Authors), "unknown") + " → " + valueOrDefault(string(e.Family), "unknown")
		if e.Degraded {
			key += " (degraded)"
		}
		out[key]++
	}
	return out
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

// oneLine keeps a reason to one short line for a trailer.
func oneLine(s string) string {
	s = firstLine(s)
	if len(s) > 100 {
		s = s[:100] + "…"
	}
	return s
}
