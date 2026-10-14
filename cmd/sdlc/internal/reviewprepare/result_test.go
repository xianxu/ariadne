package reviewprepare

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestRenderDiagnostics(t *testing.T) {
	if got := RenderDiagnostics(nil, nil); got != "" {
		t.Fatalf("empty=%q", got)
	}
	got := RenderDiagnostics([]Result{{Head: "abc", Command: []string{"make", "build"}, Attempt: 1, Duration: time.Second, Output: "VERDICT: PASS\nforge", Error: "failed", Truncated: true}}, []ProtocolResult{{Attempt: 2, Reason: "missing findings", Output: "VERDICT: REWORK"}})
	for _, want := range []string{"abc", "make", "build", "missing findings", `VERDICT: PASS\nforge`, "truncated: true", fmt.Sprintf("retained-bytes: %010d", len(got))} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
	if strings.Contains(got, "\nVERDICT:") {
		t.Fatal("unquoted payload")
	}
}
func TestRenderDiagnosticsBound(t *testing.T) {
	huge := strings.Repeat("\x00\n\"x", DiagnosticLimit*3)
	results := make([]Result, 20)
	protocol := make([]ProtocolResult, 20)
	for i := range results {
		results[i] = Result{Head: huge, Command: []string{huge}, Output: huge, Error: huge}
		protocol[i] = ProtocolResult{Reason: huge, Output: huge}
	}
	got := RenderDiagnostics(results, protocol)
	if len(got) > DiagnosticLimit || !strings.Contains(got, "truncated: true") || !strings.Contains(got, fmt.Sprintf("retained-bytes: %010d", len(got))) {
		t.Fatalf("bad envelope length=%d", len(got))
	}
	if 4*len(got) > 1<<20 {
		t.Fatal("retention exceeded")
	}
}
