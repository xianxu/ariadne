package main

import (
	"regexp"
	"strings"
	"time"
)

// A section's freshness stamp is its first line, `_updated YYYY-MM-DD HH:MM_`
// in local time. writeSection owns it: it strips the stamp before computing a
// body and puts a fresh one on top of what it writes. The operator may delete
// it; the section then has unknown age until the next write.
const stampLayout = "2006-01-02 15:04"

var stampRE = regexp.MustCompile(`^_updated (\d{4}-\d\d-\d\d \d\d:\d\d)_$`)

// splitStamp returns the time of the stamp on body's first non-blank line
// (zero when there is none or it doesn't parse) and the body without it.
func splitStamp(body string) (time.Time, string) {
	rest := strings.TrimLeft(body, "\n")
	first, after, _ := strings.Cut(rest, "\n")
	m := stampRE.FindStringSubmatch(strings.TrimSpace(first))
	if m == nil {
		return time.Time{}, body
	}
	t, err := time.ParseInLocation(stampLayout, m[1], time.Local)
	if err != nil {
		return time.Time{}, body
	}
	return t, strings.TrimLeft(after, "\n")
}

// stamped puts a fresh stamp on top of body, replacing any stamp it has.
func stamped(body string, now time.Time) string {
	_, rest := splitStamp(body)
	return "_updated " + now.Format(stampLayout) + "_\n\n" + strings.Trim(rest, "\n")
}

// stampedSection reports whether writes to heading carry a stamp: every
// section but the operator's own.
func stampedSection(heading string) bool { return heading != instructionsHeading }
