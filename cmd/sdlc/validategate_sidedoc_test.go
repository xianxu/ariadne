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
