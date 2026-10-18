// family.go — who reviews whom (#409), judge's side: the policy and the plan
// are pkg/reviewpolicy's (#441); here are the run-shaped checks the dispatch
// loop uses (what counts as unavailable, what can launch here).
package judge

import (
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/pkg/reviewpolicy"
)

// Families, change kinds, candidates and the plan are pkg/reviewpolicy's
// (#441); judge keeps their historical names and the run-shaped parts below.
type (
	Family       = reviewpolicy.Family
	ChangeKind   = reviewpolicy.ChangeKind
	Candidate    = reviewpolicy.Candidate
	Skip         = reviewpolicy.Skip
	Availability = reviewpolicy.Availability
)

const (
	FamilyAnthropic = reviewpolicy.Anthropic
	FamilyOpenAI    = reviewpolicy.OpenAI
	FamilyGoogle    = reviewpolicy.Google
	KindDesignHeavy = reviewpolicy.DesignHeavy
	KindBounded     = reviewpolicy.Bounded
)

// FamilyOf is an agent's vendor ("" when unknown).
func FamilyOf(agent string) Family { return reviewpolicy.FamilyOf(agent) }

// ReviewPlan is reviewpolicy.Plan.
func ReviewPlan(policy ReviewerPolicy, main AgentCLI, authors []Family, kind ChangeKind, available Availability) ([]Candidate, []Skip) {
	return reviewpolicy.Plan(policy, main, authors, kind, available)
}

// unavailableRE are the causes that mean the reviewer could not review at all:
// its API unreachable (apiFailureRE) or erroring, its quota spent, its sign-in
// missing, or the CLI itself refused (gemini's IneligibleTierError, measured
// 2026-10-18). A bad model or request (HTTP 400) is a misconfiguration, not
// unavailability, so it halts rather than moving on.
var unavailableRE = regexp.MustCompile(`(?i)` + apiFailureRE.String() +
	`|rate.?limit|usage limit|quota|insufficient_quota|credit balance|\b429\b|unauthori[sz]ed|\b401\b|\b403\b|"status":\s*5\d\d|invalid api key|not logged in|please (log|sign) in|authentication (failed|required|error)|IneligibleTierError|no longer supported`)

// Unavailable reports whether a run is evidence that the reviewer could not
// review at all, so trying another reviewer duplicates no review:
//   - a failed launch;
//   - claude's stream that failed before any assistant or tool event
//     (AgentRun.Began) with an unavailable cause in its result or stderr;
//   - codex or gemini (plain output) that exited non-zero having printed
//     nothing at all on stdout, where a review's text would be, with an
//     unavailable cause on stderr: for codex only its `ERROR:` lines, since
//     codex echoes the prompt (which may contain any word) to stderr, and
//     only when stderr shows none of codex's activity blocks.
//
// Anything else (a run that began, an unparseable stream, a cause outside the
// list) is not evidence of absence: RunFailure decides it and it halts as
// before.
func Unavailable(agent AgentCLI, run AgentRun, stderr []byte, nonZeroExit bool, launchErr error) (string, bool) {
	if launchErr != nil {
		return "could not launch: " + firstLine(launchErr.Error()), true
	}
	if HasVerdict(run.Text()) {
		return "", false
	}
	var causes []string
	switch {
	case run.Stream:
		if run.Began || !nonZeroExit && (run.Result == nil || !run.Result.IsError) {
			return "", false
		}
		causes = []string{resultText(run), string(stderr)}
	case nonZeroExit && strings.TrimSpace(run.Tail) == "":
		if agent == AgentCodex {
			if codexBegan(string(stderr)) {
				return "", false
			}
			causes = errorLines(string(stderr))
		} else {
			causes = []string{string(stderr)}
		}
	default:
		return "", false
	}
	for _, text := range causes {
		if loc := unavailableRE.FindStringIndex(text); loc != nil {
			return lineAt(text, loc[0]), true
		}
	}
	return "", false
}

// codexBegan reports codex's activity blocks on stderr: a message ("codex"),
// a command ("exec") or reasoning ("thinking") each open with a line of their
// own after the echoed prompt, so any of them means the review began.
func codexBegan(stderr string) bool {
	for _, line := range strings.Split(stderr, "\n") {
		switch strings.TrimSpace(line) {
		case "codex", "exec", "thinking":
			return true
		}
	}
	return false
}

// errorLines are codex's own error reports on stderr.
func errorLines(stderr string) []string {
	var out []string
	for _, line := range strings.Split(stderr, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "ERROR:") {
			out = append(out, line)
		}
	}
	return out
}

func resultText(run AgentRun) string {
	if run.Result == nil {
		return ""
	}
	return run.Result.Text
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
