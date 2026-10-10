// apifailure.go — a reviewer that never reached its API is "review did not
// run", not a verdict (#300, D2 of the 2026-10-09 evidence). Judged on the
// run's channel, never on the review's prose: a review may quote these very
// strings (this issue's own diff does).
package judge

import (
	"errors"
	"regexp"
	"strings"
)

// ErrAPIUnreachable is a dispatch that failed before the reviewer could work.
var ErrAPIUnreachable = errors.New("review did not run: the reviewer could not reach its API")

// apiFailureRE are the failure signatures seen in the evidence transcripts.
var apiFailureRE = regexp.MustCompile(`API Error|ERR_PROXY_TUNNEL|Connection blocked by network allowlist|ECONNREFUSED|ENOTFOUND`)

// APIFailure reports whether a run failed to reach its API, and the cause.
// The stream's terminal result says so directly (is_error). Without a stream
// (codex, gemini, a claude stream that didn't parse), it takes a run with no
// verdict whose stderr — or, for a non-zero exit, whose output — carries a
// failure signature.
func APIFailure(run AgentRun, stderr []byte, nonZeroExit bool) (string, bool) {
	if run.Result != nil && run.Result.IsError {
		return firstLine(valueOrDefault(run.Result.Text, "the run reported an error")), true
	}
	if HasVerdict(run.Text()) {
		return "", false
	}
	if m := apiFailureRE.FindIndex(stderr); m != nil {
		return lineAt(string(stderr), m[0]), true
	}
	if nonZeroExit && !run.Stream {
		if m := apiFailureRE.FindStringIndex(run.Tail); m != nil {
			return lineAt(run.Tail, m[0]), true
		}
	}
	return "", false
}

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
