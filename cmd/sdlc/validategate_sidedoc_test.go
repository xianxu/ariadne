package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

func TestValidateSideDocsUseTheirOwnType(t *testing.T) {
	name := "workshop/issues/000440-owner.side/2026-10/draft.md"
	doc := "---\ntype: side-doc\nparent: ../../000440-owner.md\nmarker: question\nfuture: {nested: [kept]}\n---\n# draft\n"
	defer stubGate(t, []gitx.FileChange{{Status: "A", Path: name}}, true, nil, map[string]string{name: doc})()
	var nouns []string
	validateFrontmatterFn = func(noun, file string) (string, bool, error) {
		nouns = append(nouns, noun)
		if noun != "side-doc" {
			return "must use side-doc conformance", false, nil
		}
		return "", true, nil
	}
	var out bytes.Buffer
	err := validateChangedInstances("", "BASE", "HEAD", nounGates("workshop/issues"), &out, io.Discard)
	if err != nil {
		t.Fatalf("valid side doc refused: %v; %s", err, out.String())
	}
	if len(nouns) != 1 || nouns[0] != "side-doc" {
		t.Fatalf("validator calls %v", nouns)
	}
}

func TestUntypedSideEvidenceRemainsValid(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"notes.md", "# Legacy evidence without YAML\n"},
		{"metadata.md", "---\nsource: fixture\nparent: absent-is-allowed-here\n---\n# evidence\n"},
		{"record.json", "{\"evidence\":true}\n"},
		{"record.log", "terminal capture\n"},
		{"malformed-untyped.md", "---\nsource: [legacy evidence\n---\n"},
		{"nested-flow-type.md", "---\n{metadata: {type: [invalid}}\n---\n"},
		{"literal-type.md", "---\nsource: |\n  type: [evidence\nother: [invalid\n---\n"},
		{"unterminated-untyped.md", "---\nsource: legacy evidence\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			name := "workshop/issues/000858-owner.side/evidence/" + tc.name
			defer stubGate(t, []gitx.FileChange{{Status: "A", Path: name}}, true, nil, map[string]string{name: tc.body})()
			calls := 0
			validateFrontmatterFn = func(_, _ string) (string, bool, error) { calls++; return "unexpected validator", false, nil }
			var out bytes.Buffer
			err := validateChangedInstances("", "BASE", "HEAD", nounGates("workshop/issues"), &out, io.Discard)
			if err != nil || calls != 0 {
				t.Fatalf("legacy attachment newly gated: calls%d, err%v, output%s", calls, err, out.String())
			}
		})
	}
}

func TestTypedSideDocRefusalDoesNotDemandIssueSections(t *testing.T) {
	name := "workshop/issues/000440-owner.side/draft.md"
	for _, body := range []string{"---\ntype: side-doc\n---\n# no parent\n", "---\ntype: side-doc\nparent: ''\n---\n# blank parent\n"} {
		t.Run(fmt.Sprint(len(body)), func(t *testing.T) {
			defer stubGate(t, []gitx.FileChange{{Status: "A", Path: name}}, false, nil, map[string]string{name: body})()
			validateFrontmatterFn = func(noun, _ string) (string, bool, error) {
				return noun + ": parent must be a nonempty string", false, nil
			}
			var out bytes.Buffer
			err := validateChangedInstances("", "BASE", "HEAD", nounGates("workshop/issues"), &out, io.Discard)
			if err == nil || !strings.Contains(err.Error(), "side-doc") || strings.Contains(err.Error(), "(section)") {
				t.Fatalf("wrong side-doc refusal: %v", err)
			}
		})
	}
}

func TestSideDocGateConfiguredRoots(t *testing.T) {
	gates := nounGates("tickets", "designs")
	for _, tc := range []struct{ path, noun string }{
		{"tickets/000001-owner.side/nested/typed.md", "side-doc"},
		{"designs/000001-owner-plan.side/typed.md", "side-doc"},
		{"tickets/000001-owner.md", "issue"},
		{"designs/000001-owner-plan.md", ""},
		{"workshop/plans/000001-owner-plan.side/typed.md", ""},
		{"designs/000001-owner-plan.side/evidence.json", ""},
	} {
		gate, ok := gateForPath(tc.path, gates)
		if ok != (tc.noun != "") || (ok && gate.noun != tc.noun) {
			t.Errorf("%s: noun%q ok%v; want%q", tc.path, gate.noun, ok, tc.noun)
		}
	}
}

func TestSideDocExplicitTypeSelection(t *testing.T) {
	for _, tc := range []struct {
		body  string
		typed bool
	}{
		{"---\ntype: side-doc\nparent: owner.md\n---\n", true},
		{"---\n\"type\": 123\n---\n", true},
		{"---\nparent: owner.md\ntype: [invalid\n---\n", true},
		{"---\nmetadata: {type: evidence}\n---\n", false},
		{"# notes\ntype: side-doc\n", false},
	} {
		if got, _ := classifySideDoc([]byte(tc.body)); got != tc.typed {
			t.Errorf("selection for %q: got%v want%v", tc.body, got, tc.typed)
		}
	}
}

func TestMalformedTypedSideDocsRefuse(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"quoted", "---\n\"type\": [invalid\n---\n"},
		{"single quoted", "---\n'type': [invalid\n---\n"},
		{"spaced", "---\ntype : [invalid\n---\n"},
		{"escaped key", "---\n\"ty\\u0070e\": [invalid\n---\n"},
		{"explicit key", "---\n? type\n: [invalid\n---\n"},
		{"flow", "---\n{source: kept, type: [invalid}\n---\n"},
		{"indented block", "---\n  type: [invalid\n---\n"},
		{"indented flow", "---\n  {source: kept, type: [invalid}\n---\n"},
		{"unterminated", "---\ntype: side-doc\nparent: owner.md\n"},
		{"unterminated CRLF", "---\r\n\"type\": side-doc\r\nparent: owner.md\r\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			name := "workshop/issues/000623-owner.side/evidence.md"
			defer stubGate(t, []gitx.FileChange{{Status: "A", Path: name}}, true, nil, map[string]string{name: tc.body})()
			var out bytes.Buffer
			err := validateChangedInstances("", "BASE", "HEAD", nounGates("workshop/issues"), &out, io.Discard)
			if err == nil || !strings.Contains(err.Error(), "side-doc") || !strings.Contains(err.Error(), name) {
				t.Fatalf("malformed typed side doc escaped: %v; %s", err, out.String())
			}
		})
	}
}

func TestSideTypeDeclarationLinearKeys(t *testing.T) {
	for _, n := range []int{8, 16} {
		for _, flow := range []bool{false, true} {
			keys := make([]string, n)
			for i := range keys {
				keys[i] = fmt.Sprintf("source%d: kept", i)
			}
			header := strings.Join(keys, "\n")
			if flow {
				header = "{" + strings.Join(keys, ", ") + "}"
			}
			visits, keyBytes := 0, 0
			found := declaresSideType(header, func(key string) bool {
				visits++
				keyBytes += len(key)
				return sideTypeKey(key)
			})
			if found || visits > 2*n || keyBytes > len(header) {
				t.Fatalf("N%d flow%v: found%v key visits%d bytes%d/header%d", n, flow, found, visits, keyBytes, len(header))
			}
		}
	}
}
