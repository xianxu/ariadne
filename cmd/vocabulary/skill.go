package main

import "strings"

// renderSkill produces the served SKILL.md body for the vocabulary skill — the
// always-loaded breadcrumb (#122). The `description:` is the eager matching
// surface; the body carries the touch-time instruction (read the .cue source
// before changing a lifecycle) plus the dynamically-enumerated noun list. Pure
// (string in, string out): unit-tested for byte-stability so two weave compiles
// produce identical bytes.
func renderSkill(nouns []string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: vocabulary\n")
	b.WriteString("description: The system's nouns and their lifecycles are formally modeled in construct/vocabulary/*.cue — the single source consumers derive from. Read the relevant .cue before editing a noun's status set or lifecycle.\n")
	b.WriteString("---\n\n")
	b.WriteString("# vocabulary\n\n")
	b.WriteString("The formal vocabulary layer (ariadne#122). Each noun's data shape + lifecycle + laws is a CUE model in `construct/vocabulary/<noun>.cue` — the single source of truth. Code consumers read the exported JSON at `construct/generated/vocabulary/<noun>.json`; you read the `.cue` source directly.\n\n")
	b.WriteString("**Before editing a noun's lifecycle or status set, read `construct/vocabulary/<noun>.cue`.** Its legal status values and transitions are defined there — don't hand-edit a status into an artifact out of band.\n\n")
	if len(nouns) > 0 {
		b.WriteString("Defined nouns: " + strings.Join(nouns, ", ") + "\n\n")
	}
	b.WriteString("Validate: `vocabulary vet`. Regenerate: `make weave`. Freshness: `vocabulary check --output construct/generated/vocabulary`.\n")
	return b.String()
}
