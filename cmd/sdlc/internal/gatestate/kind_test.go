package gatestate

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestFindingKindRoundTrip(t *testing.T) {
	for _, kind := range []string{"", "defect", "unverifiable"} {
		field := ""
		if kind != "" {
			field = "  kind: " + kind + "\n"
		}
		rr, ok := ParseFindingsBlock("```findings\nfindings:\n- severity: Important\n" + field + "  title: blocked measurement\n  detail: run host check\n```")
		if !ok || rr.New[0].Kind != kind {
			t.Fatalf("kind %q lost: %+v", kind, rr)
		}
		ledger := Ledger{IDPrefix: "BR"}
		ledger = Apply(ledger, AssignIDsAt(ledger, rr, 1, "now", "judge", "M1"))
		ledger.Disputes = []Dispute{{ID: "BR-1", After: 1, Rationale: "sandbox only", Evidence: "host test passed"}}
		rendered := Render(ledger, "test")
		parsed, err := ParseSidecar(rendered)
		if err != nil || !reflect.DeepEqual(ledger, parsed) {
			t.Fatalf("roundtrip kind %q: %v\n%+v", kind, err, parsed)
		}
		if Render(parsed, "test") != rendered {
			t.Fatal("nondeterministic render")
		}
		prompt := RenderPriorFindingsScoped(FilterBoundary(parsed, "M1"), parsed)
		for _, want := range []string{"BR-1", "host test passed"} {
			if !strings.Contains(prompt, want) {
				t.Fatalf("missing %q", want)
			}
		}
		if kind != "" && !strings.Contains(prompt, "kind: "+kind) {
			t.Fatal("kind absent from prompt")
		}
		if kind == "" && strings.Contains(rendered, "kind:") {
			t.Fatal("legacy ledger changed")
		}
	}
}

func FuzzFindingKind(f *testing.F) {
	for _, s := range []string{"", "defect", "unverifiable", "unknown", "[", "DEFECT"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, kind string) {
		// Quote as YAML-compatible JSON string so fuzzing varies the enum, not YAML structure.
		output := fmt.Sprintf("```findings\nfindings:\n- severity: Important\n  title: test\n  kind: %q\n```", kind)
		rr, ok := ParseFindingsBlock(output)
		if ok && kind != "" && kind != "defect" && kind != "unverifiable" {
			t.Fatalf("unmodeled kind admitted: %q (%+v)", kind, rr)
		}
	})
}

func TestUnverifiableFindingRequiresDetail(t *testing.T) {
	for _, kind := range []string{"", "defect", "unverifiable"} {
		for _, severity := range []string{"Minor", "Important", "Critical"} {
			for _, detail := range []string{"", "   ", "go test failed: sandbox denied process lookup; run on host"} {
				output := fmt.Sprintf("```findings\nfindings:\n- severity: %s\n  kind: %q\n  title: Cannot verify\n  detail: %q\n```", severity, kind, detail)
				_, ok := ParseFindingsBlock(output)
				want := kind != "unverifiable" || strings.TrimSpace(detail) != ""
				if ok != want {
					t.Errorf("kind=%q severity=%s detail=%q accepted=%v, want %v", kind, severity, detail, ok, want)
				}
			}
		}
	}
}
