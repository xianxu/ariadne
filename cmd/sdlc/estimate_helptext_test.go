package main

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/helptext"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

// TestEstimateHelptextDocumentsVocab is the ARCH-DRY drift guard (#117 M2): the
// closed primitive vocabulary is canonical in internal/estimate/vocab.go;
// helptext/estimate.md must document every slug. This catches the common drift —
// adding a slug to vocab.go without documenting it — so the two can't silently
// diverge.
func TestEstimateHelptextDocumentsVocab(t *testing.T) {
	doc, ok := helptext.Get("estimate")
	if !ok {
		t.Fatal("helptext/estimate.md is not embedded")
	}
	for _, slug := range estimate.Primitives() {
		if !strings.Contains(doc, slug) {
			t.Errorf("helptext/estimate.md does not document vocab slug %q (drift from vocab.go)", slug)
		}
	}
}
