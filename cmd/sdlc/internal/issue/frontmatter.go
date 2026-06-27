// Package issue provides YAML-frontmatter parse + edit helpers shared across
// the sdlc subcommands (close, set-status, milestone-close).
//
// Ported from scripts/close-issue.py — same regex-based posture (no YAML
// library), so semantics match the Python source byte-for-byte where it
// matters (status flips, field upserts, ordering).
package issue

import (
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/pkg/frontmatter"
)

// ActualNotApplicableSentinel is the single accepted issue-frontmatter spelling
// for a close whose measured actual hours do not apply and must not enter
// velocity calibration.
const ActualNotApplicableSentinel = "N/A"

// IsActualNotApplicable reports whether an actual_hours value is the exact
// not-applicable sentinel, ignoring surrounding frontmatter whitespace.
func IsActualNotApplicable(s string) bool {
	return strings.TrimSpace(s) == ActualNotApplicableSentinel
}

// Parse splits a markdown document into its YAML frontmatter and body.
// Returns an error if the document doesn't start with a "---\n...---\n"
// fence. Delegates to pkg/frontmatter.Split — the one source for the split
// (#124 ARCH-DRY; cmd/vocabulary needs the same parse but can't import this
// internal package), preserving the missing-fence error contract.
func Parse(text string) (fm, body string, err error) {
	return frontmatter.Split(text)
}

// Compose reassembles a frontmatter + body back into the full document.
// Matches the exact spacing the close-issue.py source emits:
// "---\n<fm>\n---\n<body>" (no trailing newline added beyond what body
// carries).
func Compose(fm, body string) string {
	return "---\n" + fm + "\n---\n" + body
}

// GetField returns the value of `name:` in the frontmatter, trimmed.
// ok=false if the field is absent.
//
// The inter-token gap is `[ \t]*`, NOT `\s*`: Go's `\s` includes `\n`, so on
// an empty field (`github_issue:\n`) `\s*` would span the newline and capture
// the *next* line's value. Empty fields (github_issue, estimate_hours) are
// routine, so the gap must stay within the line.
func GetField(fm, name string) (value string, ok bool) {
	re, err := regexp.Compile(`(?m)^` + regexp.QuoteMeta(name) + `:[ \t]*(.*)$`)
	if err != nil {
		return "", false
	}
	m := re.FindStringSubmatch(fm)
	if m == nil {
		return "", false
	}
	return strings.TrimSpace(m[1]), true
}

// SetField sets `name: value` in the frontmatter. If the field exists,
// its line is replaced in place (preserving field order). If absent,
// it's appended at the end of the frontmatter block (after any trailing
// whitespace is trimmed, then a newline + the new field added).
//
// Mirrors close-issue.py's fm_set semantics exactly.
func SetField(fm, name, value string) string {
	re, err := regexp.Compile(`(?m)^` + regexp.QuoteMeta(name) + `:.*$`)
	if err != nil {
		return fm
	}
	if re.MatchString(fm) {
		return re.ReplaceAllString(fm, name+": "+value)
	}
	return strings.TrimRight(fm, "\n\r\t ") + "\n" + name + ": " + value
}
