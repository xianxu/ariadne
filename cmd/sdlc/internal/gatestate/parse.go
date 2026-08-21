package gatestate

import (
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// findingsBlockRE extracts the body of a fenced ```findings … ``` block — the structured
// handoff the plan judge emits. Same shape as judge.verdictBlockRE, one level of structure
// richer. (?s) so `.` spans newlines; the lazy body means the FIRST closing fence ends a
// block, and FindAll then walks to the last complete one.
var findingsBlockRE = regexp.MustCompile("(?s)```+[ \t]*findings[ \t]*\r?\n(.*?)\r?\n```+")

// ParseFindingsBlock extracts the LAST fenced ```findings block from an agent's output and
// validates every severity and disposition against the `finding` model.
//
// This is the AUTHORITATIVE structured handoff (the agent-binary-handoff-schema target):
// it does NOT parse prose, so a missing or model-invalid block is a genuine protocol miss
// (ok=false), never a heuristic read of the judge's paragraphs. The caller surfaces that
// as a protocol error and records a findings-less round; it does not guess.
//
// LAST block wins, mirroring ParseVerdictBlock: a judge that shows an example block before
// its real one must not hand us the example. Pure.
func ParseFindingsBlock(output string) (RoundReport, bool) {
	ms := findingsBlockRE.FindAllStringSubmatch(output, -1)
	if len(ms) == 0 {
		return RoundReport{}, false
	}
	var rr RoundReport
	if err := yaml.Unmarshal([]byte(ms[len(ms)-1][1]), &rr); err != nil {
		return RoundReport{}, false
	}
	m := vocab.Finding()
	for i, f := range rr.New {
		// A titleless finding is unusable in the ledger and in the next round's prompt,
		// and an unmodeled severity would reach Decide with no defined blocking behavior.
		if !m.IsSeverity(f.Severity) || strings.TrimSpace(f.Title) == "" {
			return RoundReport{}, false
		}
		rr.New[i].Title = normalizeText(f.Title)
		rr.New[i].Detail = normalizeText(f.Detail)
		rr.New[i].Family = normalizeText(rr.New[i].Family)
	}
	for i, d := range rr.Dispositions {
		if !m.IsDisposition(d.State) || strings.TrimSpace(d.ID) == "" {
			return RoundReport{}, false
		}
		rr.Dispositions[i].ID = strings.TrimSpace(d.ID)
		rr.Dispositions[i].Note = normalizeText(d.Note)
	}
	return rr, true
}

// normalizeText canonicalizes one field of agent-authored prose. Surrounding whitespace in
// a finding's title or detail carries no meaning — the judge wrote prose, not data — so the
// schema boundary is the right place to strip it.
//
// It is also load-bearing for durability. go.yaml.in/yaml/v3 mis-emits a string with a
// LEADING newline: it writes an explicit block-scalar indentation indicator (`|4-`) that
// disagrees with the 8-space indentation it actually produces, so its own parser rejects
// the result. A finding whose detail began with a newline would therefore render a gate
// ledger that could never be read back — permanently destroying the gate's memory for that
// issue. Found by FuzzRenderParseRoundTrip within a second of its first run.
func normalizeText(s string) string { return strings.TrimSpace(s) }
