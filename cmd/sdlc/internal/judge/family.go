// family.go — who reviews whom (#409): a reviewer is chosen by the vendor
// family of the change's authors and by the kind of change, within a repo's
// #310 policy. Everything here is pure; the dispatch loop supplies
// availability and records what was skipped.
package judge

import (
	"fmt"
	"regexp"
	"strings"
)

// Family is a model vendor. Two agents of one family share blind spots, which
// is what cross-family review is for.
type Family string

const (
	FamilyAnthropic Family = "anthropic"
	FamilyOpenAI    Family = "openai"
	FamilyGoogle    Family = "google"
)

// FamilyOf is an agent's vendor, or "" for one sdlc does not know (a human,
// an unrecognized CLI): an unknown author is never guessed into a family.
func FamilyOf(agent string) Family {
	switch AgentCLI(agent) {
	case AgentClaude:
		return FamilyAnthropic
	case AgentCodex:
		return FamilyOpenAI
	case AgentGemini:
		return FamilyGoogle
	}
	return ""
}

// ChangeKind is what the review judges: a design (a new mechanism, state or
// invariant) or a bounded change.
type ChangeKind string

const (
	KindDesignHeavy ChangeKind = "design-heavy"
	KindBounded     ChangeKind = "bounded"
)

// defaultOrders is the preference among candidate reviewers per kind, before
// the authors' families are removed. Designs go first to Claude, whose reviews
// caught Codex designs over-engineering; bounded code goes first to Codex,
// whose reviews caught bugs in Claude code (#409's evidence). Claude and Codex
// lead both, so #310's `other` (claude and codex review each other) holds;
// Gemini is the fallback when neither can.
var defaultOrders = map[ChangeKind][]AgentCLI{
	KindDesignHeavy: {AgentClaude, AgentCodex, AgentGemini},
	KindBounded:     {AgentCodex, AgentClaude, AgentGemini},
}

// Candidate is one reviewer to try, with the model to ask for ("" is the
// CLI's own default).
type Candidate struct {
	Agent  AgentCLI
	Model  string
	Family Family
}

func (c Candidate) String() string {
	s := string(c.Agent)
	if c.Model != "" {
		s += "/" + c.Model
	}
	return fmt.Sprintf("%s (%s)", s, valueOrDefault(string(c.Family), "unknown family"))
}

// Skip is a candidate passed over before or instead of reviewing, and why.
type Skip struct {
	Agent  AgentCLI
	Reason string
}

// Availability reports, before launch, whether an agent can review here, and
// why not when it cannot (its CLI missing, its sandbox known to fail).
type Availability func(AgentCLI) (reason string, ok bool)

// ReviewPlan is the ordered reviewers to try for a change by authors of the
// given families, under policy. `same` and `explicit` are the #310 floor and
// name one reviewer; the dynamic `other` lists every agent whose family is
// not among the authors, in the kind's order. Candidates available rejects
// are returned as skips; an empty plan means none can review here.
func ReviewPlan(policy ReviewerPolicy, main AgentCLI, authors []Family, kind ChangeKind, available Availability) ([]Candidate, []Skip) {
	var agents []AgentCLI
	switch policy.mode() {
	case ReviewerOther:
		excluded := map[Family]bool{}
		for _, f := range authors {
			excluded[f] = true
		}
		for _, a := range policy.order(kind) {
			if !excluded[FamilyOf(string(a))] {
				agents = append(agents, a)
			}
		}
	case ReviewerExplicit:
		agents = []AgentCLI{policy.Reviewer(main)}
	default:
		agents = []AgentCLI{main}
	}
	var plan []Candidate
	var skips []Skip
	for _, a := range agents {
		if available != nil {
			if reason, ok := available(a); !ok {
				skips = append(skips, Skip{Agent: a, Reason: reason})
				continue
			}
		}
		plan = append(plan, Candidate{Agent: a, Model: policy.Models[a], Family: FamilyOf(string(a))})
	}
	return plan, skips
}

// unavailableRE are the causes that mean the reviewer could not review at all:
// its API unreachable (apiFailureRE), its quota spent, or its sign-in missing.
var unavailableRE = regexp.MustCompile(`(?i)` + apiFailureRE.String() +
	`|rate.?limit|usage limit|quota|insufficient_quota|credit balance|\b429\b|unauthori[sz]ed|\b401\b|invalid api key|not logged in|please (log|sign) in|authentication (failed|required|error)`)

// Unavailable reports whether a run is evidence that the reviewer could not
// review at all, so trying another reviewer duplicates no review. Only a
// failed launch, or a stream that failed before any assistant or tool event
// with an unreachable, quota or sign-in cause, qualifies. A run whose output
// is not a parseable stream, or that began reviewing, is not evidence of
// absence: it is left to RunFailure and halts as before.
func Unavailable(run AgentRun, stderr []byte, nonZeroExit bool, launchErr error) (string, bool) {
	if launchErr != nil {
		return "could not launch: " + firstLine(launchErr.Error()), true
	}
	if HasVerdict(run.Text()) || !run.Stream || run.Began {
		return "", false
	}
	failed := nonZeroExit || run.Result != nil && run.Result.IsError
	if !failed {
		return "", false
	}
	for _, text := range []string{resultText(run), string(stderr), run.Tail} {
		if loc := unavailableRE.FindStringIndex(text); loc != nil {
			return lineAt(text, loc[0]), true
		}
	}
	return "", false
}

func resultText(run AgentRun) string {
	if run.Result == nil {
		return ""
	}
	return run.Result.Text
}

// trimReason keeps a skip reason to one short line for a trailer.
func trimReason(s string) string {
	s = firstLine(s)
	if len(s) > 120 {
		s = s[:120] + "…"
	}
	return strings.TrimSpace(s)
}

// LocalAvailability is the pre-launch check the dispatch loop uses: an agent
// whose CLI is not on PATH cannot review, nor can codex under checkout
// containment on macOS, where its own Seatbelt sandbox fails as the fence
// probe did (#310) so every command it runs would be refused. lookPath is
// exec.LookPath in production.
func LocalAvailability(containment Containment, goos string, lookPath func(string) (string, error)) Availability {
	return func(a AgentCLI) (string, bool) {
		if _, err := lookPath(string(a)); err != nil {
			return string(a) + " is not on PATH", false
		}
		if a == AgentCodex && containment == ContainCheckout && goos == "darwin" {
			return "codex's sandbox cannot apply under checkout containment on macOS", false
		}
		return "", true
	}
}
