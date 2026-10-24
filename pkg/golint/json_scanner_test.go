package golint

import (
	"os"
	"strings"
	"testing"
)

func scannerCaptureWithoutAllow(raw []byte) []byte {
	var lines []string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "//lint:allow json-recursive-scanner ") {
			continue
		}
		lines = append(lines, line)
	}
	return []byte(strings.Join(lines, "\n"))
}

// These two roles need human-reviewed per-site reasons. The rule deliberately
// does not try to infer materializer value flow or differential independence.
func TestJSONScannerCapturedRolesRequireReasonedSiteAllows(t *testing.T) {
	for _, tc := range []struct{ fixture, path string }{
		{"oracle.go.txt", "cmd/internal/strictjson/decode_equivalence_test.go"},
		{"materializer.go.txt", "examples/mock-to-real/internal/jsonvalue/value.go"},
	} {
		t.Run(tc.fixture, func(t *testing.T) {
			raw, err := os.ReadFile("testdata/json-scanner-roles/" + tc.fixture)
			if err != nil {
				t.Fatal(err)
			}
			for _, annotated := range []bool{false, true} {
				src := scannerCaptureWithoutAllow(raw)
				want := 1
				if annotated {
					src = raw
					want = 0
				}
				rules := []Rule{ruleOf(t, "json-recursive-scanner")}
				got, err := Check(tc.path, src, rules)
				if err != nil || len(got) != want {
					t.Errorf("Check annotated=%v findings=%v err=%v want%d", annotated, got, err, want)
				}
				if want == 1 && len(got) == 1 && got[0].Rule != "json-recursive-scanner" {
					t.Errorf("wrong finding: %v", got)
				}
				all, err := CheckSources("fixture/module", map[string][]byte{tc.path: src}, rules)
				if err != nil || len(all[tc.path]) != want {
					t.Errorf("CheckSources annotated=%v findings=%v err=%v want%d", annotated, all, err, want)
				}
				if want == 1 && len(all[tc.path]) == 1 && all[tc.path][0].Rule != "json-recursive-scanner" {
					t.Errorf("wrong CheckSources finding: %v", all)
				}
			}
		})
	}
}

func TestJSONScannerDiagnosticNamesConsumerAuthorityAndSiteException(t *testing.T) {
	message := ruleOf(t, "json-recursive-scanner").Message
	for _, want := range []string{"this module", "strict", "per-site", "materializer", "oracle", "lint:allow"} {
		if !strings.Contains(message, want) {
			t.Errorf("missing %q in %q", want, message)
		}
	}
	if strings.Contains(message, "cmd/sdlc/internal") {
		t.Errorf("downstream cannot import Ariadne's internal authority: %s", message)
	}
}
