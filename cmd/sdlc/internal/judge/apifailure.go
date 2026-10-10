// apifailure.go — a reviewer whose run failed before it could review is
// "review did not run", not a verdict (#300, D2 of the 2026-10-09 evidence).
// Judged on the run's channel, never on the review's prose (a review may quote
// these very strings), and never over a verdict the run already gave: like a
// postscript, a late failure must not erase it.
package judge

import (
	"errors"
	"regexp"
	"strings"
)

// ErrReviewDidNotRun is a dispatch that failed before the reviewer reviewed.
var ErrReviewDidNotRun = errors.New("review did not run")

// ErrAPIUnreachable is the failure whose cause is the network: the reviewer
// could not reach its API (in an agent sandbox, its host is not allowed).
var ErrAPIUnreachable = errors.New("review did not run: the reviewer could not reach its API")

// apiFailureRE are the unreachable-API signatures seen in the evidence.
var apiFailureRE = regexp.MustCompile(`API Error|ERR_PROXY_TUNNEL|Connection blocked by network allowlist|ECONNREFUSED|ENOTFOUND`)

// apiHosts is each agent's API host, for the sandbox remedy.
var apiHosts = map[AgentCLI]string{
	AgentClaude: "api.anthropic.com",
	AgentCodex:  "api.openai.com",
	AgentGemini: "generativelanguage.googleapis.com",
}

// RunFailure reports whether a run failed before reviewing, and why. A run
// that carries a verdict never failed. Otherwise the stream's terminal result
// says so directly (is_error); without a stream, a non-zero exit with a
// failure signature on stderr or in the output does. The error is
// ErrAPIUnreachable, naming the agent's host to allow, when the cause carries
// an unreachable-API signature, else ErrReviewDidNotRun with the cause.
func RunFailure(agent AgentCLI, run AgentRun, stderr []byte, nonZeroExit bool) error {
	if HasVerdict(run.Text()) {
		return nil
	}
	cause := ""
	switch {
	case run.Result != nil && run.Result.IsError:
		cause = firstLine(valueOrDefault(run.Result.Text, "the run reported an error"))
	case nonZeroExit && apiFailureRE.Match(stderr):
		cause = lineAt(string(stderr), apiFailureRE.FindIndex(stderr)[0])
	case nonZeroExit && !run.Stream && apiFailureRE.MatchString(run.Tail):
		cause = lineAt(run.Tail, apiFailureRE.FindStringIndex(run.Tail)[0])
	default:
		return nil
	}
	if apiFailureRE.MatchString(cause) || apiFailureRE.Match(stderr) {
		host := apiHosts[agent]
		if host == "" {
			host = apiHosts[AgentClaude]
		}
		return errorWithCause(ErrAPIUnreachable, cause, "in an agent sandbox, allow "+host+" for this command, then rerun")
	}
	return errorWithCause(ErrReviewDidNotRun, cause, "rerun the review")
}

func errorWithCause(kind error, cause, next string) error {
	return &runFailure{kind: kind, msg: kind.Error() + " (" + cause + "); " + next}
}

type runFailure struct {
	kind error
	msg  string
}

func (e *runFailure) Error() string { return e.msg }
func (e *runFailure) Unwrap() error { return e.kind }

func lineAt(s string, i int) string {
	start := strings.LastIndex(s[:i], "\n") + 1
	return firstLine(s[start:])
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(s), "\n")
	return line
}

func valueOrDefault(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}
