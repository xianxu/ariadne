package issue

import (
	"strings"
	"testing"
)

// TestUnterminatedPolicies_DisagreeOnPurpose pins the fork itself. These two
// consumers reach OPPOSITE conclusions about the same input, and that is the
// design — a test exists so a later reader "unifying" them has to argue with it.
//
//	stripCodeFences  a word-count gate: an unclosed fence is prose, so the gate
//	                 still sees the tail and can refuse on real content.
//	SplitFences      a file rewriter (#179 migrate): an unclosed tail stays
//	                 fenced, because editing inside maybe-code is unrecoverable.
func TestUnterminatedPolicies_DisagreeOnPurpose(t *testing.T) {
	const in = "prose here\n```\nnever closed but still words\n"

	if got := stripCodeFences(in); !strings.Contains(got, "never closed but still words") {
		t.Errorf("stripCodeFences must treat an unclosed tail as PROSE so the word count sees it; got %q", got)
	}
	var fencedTail bool
	for _, seg := range SplitFences(in) {
		if seg.Fenced && strings.Contains(seg.Text, "never closed") {
			fencedTail = true
		}
	}
	if !fencedTail {
		t.Error("SplitFences must treat an unclosed tail as FENCED so migrate won't rewrite inside it")
	}
}
