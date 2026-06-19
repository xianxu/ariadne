package bench

import (
	"strings"
	"testing"
)

func TestTaskRoundTrip(t *testing.T) {
	in := Task{
		ID: "119-demo", Repo: "ariadne", SourceIssue: "119",
		BaseSHA: "abc123", Created: "2026-06-19",
		// Spec deliberately embeds a ```json fence — the config extraction must
		// NOT pick this up (regression guard for the first-block bug).
		Spec:   "Solve the thing.\n\n```json\n{\"example\": true}\n```\n\nDetails here.",
		Setup:  []string{"go build ./..."},
		Rubric: DefaultRubric(),
	}
	got, err := ParseTask(RenderTask(in))
	if err != nil {
		t.Fatalf("ParseTask: %v", err)
	}
	if got.ID != in.ID || got.BaseSHA != in.BaseSHA || got.Spec != in.Spec {
		t.Errorf("scalar/spec mismatch:\n got %+v\nwant %+v", got, in)
	}
	if len(got.Setup) != 1 || got.Setup[0] != "go build ./..." {
		t.Fatalf("config leaked from spec's json fence: setup=%v", got.Setup)
	}
	if len(got.Rubric.Subjective) != len(in.Rubric.Subjective) {
		t.Errorf("rubric not round-tripped")
	}
}

func TestParseTaskMissingConfig(t *testing.T) {
	text := "---\ntype: benchmark-task\nid: x\n---\n\n# x\n\n## Spec\n\nhi\n"
	if _, err := ParseTask(text); err == nil {
		t.Fatal("expected error for missing ## Config")
	}
}

func TestExtractJSONBlock(t *testing.T) {
	body := "preamble\n\n```json\n{\"a\":1}\n```\ntrailer"
	got, ok := extractJSONBlock(body)
	if !ok || !strings.Contains(got, `"a":1`) {
		t.Fatalf("extractJSONBlock = %q, %v", got, ok)
	}
	if _, ok := extractJSONBlock("no fence here"); ok {
		t.Fatal("expected no block")
	}
}
