package tracker

import (
	"reflect"
	"testing"
)

func TestSelectLegacyCloseCanonicalGeneration(t *testing.T) {
	evidence := []LegacyCloseEvidence{
		{Position: 0, Commit: "terminal", Status: "done", Actual: "0.8", Started: "start"},
		{Position: 1, Commit: "metadata", Status: "codecomplete", Actual: "0.8", Started: "start"},
		{Position: 2, Commit: "latest", Status: "codecomplete", Actual: "0.8", Started: "start", HasReview: true},
		{Position: 3, Commit: "first", Status: "codecomplete", Actual: "0.7", Started: "start", HasReview: true},
		{Position: 4, Commit: "working", Status: "working", Started: "start"},
	}
	want, err := SelectLegacyClose(evidence, "0.8", "start")
	if err != nil {
		t.Fatal(err)
	}
	if want.TerminalCommit != "terminal" || want.CloseCommit != "latest" {
		t.Fatalf("wrong generation: %+v", want)
	}
	for i := range evidence {
		for j := range evidence {
			swapped := append([]LegacyCloseEvidence(nil), evidence...)
			swapped[i], swapped[j] = swapped[j], swapped[i]
			got, err := SelectLegacyClose(swapped, "0.8", "start")
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("input order changed selection: %+v %v", got, err)
			}
		}
	}
}
func TestSelectLegacyCloseRefusesConflictingGeneration(t *testing.T) {
	base := []LegacyCloseEvidence{{Position: 0, Commit: "terminal", Status: "done", Actual: "0.8", Started: "start"}, {Position: 1, Commit: "close", Status: "codecomplete", Actual: "0.8", Started: "start", HasReview: true}}
	for _, kind := range []string{"reopened", "actual", "started", "position", "older terminal", "missing terminal", "terminal actual"} {
		t.Run(kind, func(t *testing.T) {
			e := append([]LegacyCloseEvidence(nil), base...)
			switch kind {
			case "reopened":
				e = append(e, LegacyCloseEvidence{Position: 2, Commit: "earlier-close", Status: "codecomplete", HasReview: true})
				e[1].Status = "working"
				e[1].HasReview = false
			case "terminal actual":
				e[0].Actual = "0.7"
			case "actual":
				e[1].Actual = "0.7"
			case "started":
				e[1].Started = "different"
			case "position":
				e[1].Position = 0
			case "older terminal":
				e = append(e, LegacyCloseEvidence{Position: 2, Commit: "old-terminal", Status: "done"})
			case "missing terminal":
				e = e[1:]
			}
			if got, err := SelectLegacyClose(e, "0.8", "start"); err == nil {
				t.Fatalf("accepted %s: %+v", kind, got)
			}
		})
	}
	got, err := SelectLegacyClose([]LegacyCloseEvidence{{Position: 0, Commit: "born", Status: "done"}}, "", "")
	if err != nil || got.CloseCommit != "" || got.TerminalCommit != "born" {
		t.Fatalf("fabricated absent close: %+v %v", got, err)
	}
}
