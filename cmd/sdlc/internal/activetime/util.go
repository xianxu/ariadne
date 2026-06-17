package activetime

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// parseISO parses an ISO-8601 timestamp. Mirrors active-time-v3.py parse_iso:
// Go's RFC3339 layout accepts a trailing "Z" and numeric offsets natively, plus
// fractional seconds (the transcript timestamps carry milliseconds).
func parseISO(ts string) (time.Time, error) {
	return time.Parse(time.RFC3339, ts)
}

// expandUser expands a leading "~" to the user's home dir (Python's
// Path(d).expanduser()). cobra does not expand "~", so the standalone
// `sdlc active-time --dir ~/...` path needs this. Non-"~" paths pass through.
func expandUser(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(strings.TrimPrefix(p, "~"), "/"))
		}
	}
	return p
}

// issuePattern compiles the tracked-issue matcher: #(<n1>|<n2>|…)\b, capture
// group 1 is the bare number. Mirrors active-time-v3.py's issue_pat. RE2
// supports \b. Returns nil when no issues are tracked (callers treat nil as
// "match nothing", same as the Python guard `if issue_pat is not None`).
func issuePattern(issues []string) *regexp.Regexp {
	if len(issues) == 0 {
		return nil
	}
	alts := make([]string, len(issues))
	for i, iss := range issues {
		alts[i] = regexp.QuoteMeta(iss)
	}
	return regexp.MustCompile(`#(` + strings.Join(alts, "|") + `)\b`)
}

// parseEventMentions counts tracked-issue #N refs in text. Pure — split out so
// the mention logic is unit-testable without files. nil pattern or empty text →
// no mentions.
func parseEventMentions(text string, pat *regexp.Regexp) map[string]int {
	mentions := map[string]int{}
	if pat == nil || text == "" {
		return mentions
	}
	for _, m := range pat.FindAllStringSubmatch(text, -1) {
		mentions[m[1]]++
	}
	return mentions
}
