package activetime

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueref"
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

// mentionScope decides which `#N` refs in text count as mentions of a tracked issue: the
// tracked set, plus the name of the repo whose qualifier is LOCAL.
//
// It replaces the compiled `*regexp.Regexp` that used to thread through this package
// (ariadne#190). Two reasons the set beats a pattern: building a regex from the tracked
// issues was the indirection that let three copies of the `#N` grammar drift apart, and a
// pattern cannot express "a qualified ref belongs to another repo" — which is the whole bug.
//
// The zero value matches nothing, preserving the old `nil pat` contract exactly.
type mentionScope struct {
	selfRepo string
	tracked  map[string]bool
}

// newMentionScope builds the scope for a set of tracked issue numbers.
func newMentionScope(selfRepo string, issues []string) mentionScope {
	if len(issues) == 0 {
		return mentionScope{selfRepo: selfRepo}
	}
	tracked := make(map[string]bool, len(issues))
	for _, iss := range issues {
		tracked[iss] = true
	}
	return mentionScope{selfRepo: selfRepo, tracked: tracked}
}

// parseEventMentions counts tracked-issue #N refs in text, excluding refs qualified with
// ANOTHER repo's name. Pure — split out so the mention logic is unit-testable without files.
// An empty scope or empty text yields no mentions.
func parseEventMentions(text string, sc mentionScope) map[string]int {
	return issueref.CountLocal(text, sc.selfRepo, sc.tracked)
}
