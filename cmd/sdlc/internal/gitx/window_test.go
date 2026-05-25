package gitx

import (
	"reflect"
	"regexp"
	"testing"
)

// TestIssueRefRE_DiscoveryParsing exercises the regex used by
// DiscoverWindowIssues, in isolation from git.
func TestIssueRefRE_DiscoveryParsing(t *testing.T) {
	tests := []struct {
		subject string
		want    []string
	}{
		{"#15: subject text", []string{"15"}},
		{"close #31 M4: foo", []string{"31"}},
		{"chore: bump (refs #1, #2, #3)", []string{"1", "2", "3"}},
		{"no issue ref here", nil},
		{"#42abc not a real ref", nil}, // word boundary blocks #42 from "42abc"
		{"prefix#42 suffix", []string{"42"}},
		{"#1 and #11 distinct", []string{"1", "11"}},
	}
	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			matches := issueRefRE.FindAllStringSubmatch(tt.subject, -1)
			var got []string
			for _, m := range matches {
				got = append(got, m[1])
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v want %v", got, tt.want)
			}
		})
	}
}

// TestSubjectAnchorRE verifies the regex used inside CommitWindow to
// filter --grep candidates down to true subject-anchored matches.
// We rebuild the same pattern here (CommitWindow compiles it inline from
// the issue number); equivalent to close-issue.py's
//
//	^(close\s+)?#NN(?!\d)
//
// Go's RE2 doesn't support negative lookahead, so we render (?!\d) as
// ($|[^0-9]). Equivalent behavior for the close-issue use case.
func TestSubjectAnchorRE(t *testing.T) {
	subjectRE := regexp.MustCompile(`^(close\s+)?#15($|[^0-9])`)
	tests := []struct {
		subject string
		match   bool
	}{
		{"#15: subject", true},
		{"close #15 done", true},
		{"close   #15: tabby", true},
		{"#15", true},
		{"#150: different issue", false},
		{"chore: see #15 in body", false},
		{"prefix #15 not anchored", false},
	}
	for _, tt := range tests {
		got := subjectRE.MatchString(tt.subject)
		if got != tt.match {
			t.Errorf("MatchString(%q) = %v, want %v", tt.subject, got, tt.match)
		}
	}
}
