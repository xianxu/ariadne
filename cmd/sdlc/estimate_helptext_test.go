package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/helptext"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

// TestEstimateHelptextMatchesVocab is the ARCH-DRY drift guard (#117 M2): the
// closed primitive vocabulary is canonical in internal/estimate/vocab.go;
// helptext/estimate.md must document EXACTLY that set — both directions:
//   - forward: every canonical slug is documented (else authors lack a reference);
//   - reverse: no documented slug is absent from vocab.go (else an author copies a
//     blessed-but-unknown slug and hits "unknown primitive" at the gate).
func TestEstimateHelptextMatchesVocab(t *testing.T) {
	doc, ok := helptext.Get("estimate")
	if !ok {
		t.Fatal("helptext/estimate.md is not embedded")
	}
	// Isolate the vocabulary block (between its header and the next section).
	start := strings.Index(doc, "CLOSED PRIMITIVE VOCABULARY")
	if start < 0 {
		t.Fatal("vocabulary section header not found in helptext/estimate.md")
	}
	block := doc[start:]
	if end := strings.Index(block, "UNIT NOTE"); end >= 0 {
		block = block[:end]
	}

	slugRE := regexp.MustCompile(`^[a-z][a-z0-9-]+$`)
	docSlugs := map[string]bool{}
	for _, line := range strings.Split(block, "\n") {
		f := strings.Fields(line)
		if len(f) > 0 && slugRE.MatchString(f[0]) {
			docSlugs[f[0]] = true
		}
	}

	want := map[string]bool{}
	for _, s := range estimate.Primitives() {
		want[s] = true
		if !docSlugs[s] {
			t.Errorf("vocab slug %q missing from helptext/estimate.md (drift: added to vocab.go, undocumented)", s)
		}
	}
	for s := range docSlugs {
		if !want[s] {
			t.Errorf("helptext/estimate.md documents %q which is NOT in vocab.go (stray slug — gate would reject it)", s)
		}
	}
}
