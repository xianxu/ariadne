package main

import (
	"strings"
	"testing"
)

func TestRenderSkill_ByteStable(t *testing.T) {
	nouns := []string{"issue", "task"}
	if renderSkill(nouns) != renderSkill(nouns) {
		t.Fatal("renderSkill must be byte-stable for the same input")
	}
}

func TestRenderSkill_CarriesInstructionAndNouns(t *testing.T) {
	out := renderSkill([]string{"issue"})
	for _, want := range []string{
		"name: vocabulary",
		"construct/vocabulary/<noun>.cue",
		"Before editing a noun's lifecycle",
		"Defined nouns: issue",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered skill missing %q", want)
		}
	}
}
