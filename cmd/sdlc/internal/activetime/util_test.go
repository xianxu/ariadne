package activetime

import (
	"strings"
	"testing"
)

// The MENTION path (ariadne#190): transcript prose naming another repo's issue must not
// count as this repo's. Every `pair#127` in #187's replay prose was counted as ariadne#127.
func TestParseEventMentionsExcludesForeignRefs(t *testing.T) {
	sc := newMentionScope("ariadne", []string{"187", "127"})
	got := parseEventMentions("working #187; replaying pair#127; more #187", sc)
	if got["127"] != 0 {
		t.Errorf("pair#127 counted as local 127: %v", got)
	}
	if got["187"] != 2 {
		t.Errorf("mentions[187] = %d, want 2", got["187"])
	}
	// Our own qualifier still counts.
	if m := parseEventMentions("see ariadne#187", sc); m["187"] != 1 {
		t.Errorf("a self-qualified ref must count as local: %v", m)
	}
}

// The zero scope matches nothing — the contract the old `nil *regexp` carried, which Compute
// relies on when no issues are tracked.
func TestMentionScopeZeroMatchesNothing(t *testing.T) {
	if m := parseEventMentions("#8 and #10", mentionScope{}); len(m) != 0 {
		t.Errorf("the zero scope must match nothing, got %v", m)
	}
	if m := parseEventMentions("#8", newMentionScope("ariadne", nil)); len(m) != 0 {
		t.Errorf("an empty issue list must match nothing, got %v", m)
	}
}

// The exclusion must be OBSERVABLE. A silently-dropped foreign ref reads identically to one
// that was never there — which is how the original defect survived: the numbers looked
// plausible (ariadne#190).
func TestForeignRefWarningsNameTheDroppedRefs(t *testing.T) {
	commits := []Commit{
		{Subject: "#187 M2: pair#127 replay harness + round 1 evidence"},
		{Subject: "#187 M2: more pair#127 work"},
		{Subject: "#187 M2: churn — four-bucket classification"},
		{Subject: "ariadne#180: self-qualified, not foreign"},
	}
	ws := foreignRefWarnings(commits, "ariadne")
	if len(ws) != 1 {
		t.Fatalf("want exactly one foreign ref reported, got %d: %+v", len(ws), ws)
	}
	if ws[0].Issue != "pair#127" {
		t.Errorf("Issue = %q, want pair#127 (qualified, so the operator knows WHERE it went)", ws[0].Issue)
	}
	if !strings.Contains(ws[0].Reason, "×2") {
		t.Errorf("Reason should carry the occurrence count, got %q", ws[0].Reason)
	}
	if ws[0].Active != 0 {
		t.Errorf("a dropped ref holds no minutes, got %v", ws[0].Active)
	}
	// Our own qualifier is not foreign.
	if len(foreignRefWarnings([]Commit{{Subject: "ariadne#180: x"}}, "ariadne")) != 0 {
		t.Error("a self-qualified ref must not be reported as foreign")
	}
	// Unknown self-repo: every qualified ref is foreign, which is the honest reading.
	if len(foreignRefWarnings([]Commit{{Subject: "ariadne#180: x"}}, "")) != 1 {
		t.Error(`with selfRepo "" a qualified ref cannot be confirmed local, so it must be reported`)
	}
}
