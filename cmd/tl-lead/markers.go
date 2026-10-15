package main

import (
	"regexp"
	"strings"
)

// markerRE matches an operator marker opening (🤖[ 🤖< 🤖~ 🤖{), per xx-fix.
var markerRE = regexp.MustCompile(`🤖[\[<~{]`)

// MarkerLine is one line of the file holding an operator marker.
type MarkerLine struct {
	Line int
	Text string
}

// markerLines lists the lines carrying a marker, skipping the file's own
// "To comment:" legend line.
func markerLines(text string) []MarkerLine {
	var out []MarkerLine
	for i, l := range strings.Split(text, "\n") {
		if markerRE.MatchString(l) && !strings.Contains(l, "To comment:") {
			out = append(out, MarkerLine{Line: i + 1, Text: strings.TrimSpace(l)})
		}
	}
	return out
}
