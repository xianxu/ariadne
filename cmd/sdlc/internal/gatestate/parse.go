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
	for _, f := range rr.New {
		// A titleless finding is unusable in the ledger and in the next round's prompt,
		// and an unmodeled severity would reach Decide with no defined blocking behavior.
		if !m.IsSeverity(f.Severity) || strings.TrimSpace(f.Title) == "" {
			return RoundReport{}, false
		}
	}
	for _, d := range rr.Dispositions {
		if !m.IsDisposition(d.State) || strings.TrimSpace(d.ID) == "" {
			return RoundReport{}, false
		}
	}
	return rr, true
}
