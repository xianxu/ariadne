// Command datatype is the DAG-aware datatype subsystem (#115): it renders the
// per-repo datatype SKILL.md with a LIVE datatype-noun list — the union of
// prototypes across the repo's layer DAG (local/leaf shadows shared) — injected
// into the description, so an agent triggers on "make a continuation" without
// first discovering that continuation is a datatype (the motivating discovery
// miss). It also answers apply-time queries: `datatype list` / `datatype show
// <name>`.
//
// It descends from #111's dynamic-skill generator: the datatype skill package
// carries an executable `.dynamic-skill` that runs the binary at `weave compile`
// time. The prose is a `go:embed`ed verbatim copy of the authored SKILL.md
// (SKILL.md.tmpl) with ONE placeholder — the description tail — replaced by the
// sorted datatype nouns. Everything else is byte-identical, so a real change
// surfaces as a reviewable git diff.
//
// ARCH-PURE: mergeTypes + renderSkill + formatList + findRepoRoot are pure
// (over the injected layergraph.FS / a path-walk), unit-tested without IO; main
// is the thin filesystem shell.
package main

import (
	_ "embed"
	"strings"
)

// skillTemplate is the authored datatype SKILL.md prose, verbatim, with the
// description tail carrying the placeholder token (see datatypeNamesPlaceholder).
// It is the single source of truth for the generated SKILL.md body — only this
// file is hand-edited; the lowered construct/local/datatype/SKILL.md is committed
// codegen produced from it.
//
//go:embed SKILL.md.tmpl
var skillTemplate string

// datatypeNamesPlaceholder is a plain string token (NOT a template action) the
// description tail carries, so rendering is a single string-replace — COLLISION-
// PROOF against the prose's literal braces (the body's `awk '/^---$/{c++; next}'`
// recipe contains a `{`, which any `{{ }}` template engine could trip over). One
// token, one replace: no template grammar, no escaping.
const datatypeNamesPlaceholder = "__DATATYPE_NAMES__"

// renderSkill produces the SKILL.md body by replacing the description-tail
// placeholder with the names joined by ", ". Pure and deterministic: the same
// sorted names always yield byte-identical output (the byte-stable guard
// depends on it). The names come from mergeTypes (the DAG-merged set), so the
// filename — NOT the prototype's `type:` frontmatter — is the authoritative type
// name (product.md carries `type: type, name: product`).
func renderSkill(names []string) string {
	return strings.Replace(skillTemplate, datatypeNamesPlaceholder, strings.Join(names, ", "), 1)
}
