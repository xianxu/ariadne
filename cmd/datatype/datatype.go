// Command datatype is a dynamic-skill generator (#111): it writes the datatype
// SKILL.md with a LIVE datatype-noun list injected into the description, so an
// agent triggers on "make a continuation" without first discovering that
// continuation is a datatype (the motivating discovery miss).
//
// It is the first consumer of weave's dynamic-skill mechanism (#111 M1): the
// datatype skill package carries an executable `.dynamic-skill` that runs
// `go run ../../../cmd/datatype --output .`, regenerating its committed SKILL.md
// at `weave compile` time. The prose is a `go:embed`ed verbatim copy of the
// authored SKILL.md (SKILL.md.tmpl) with ONE placeholder — the description tail —
// replaced by the sorted datatype nouns. Everything else is byte-identical, so a
// real change surfaces as a reviewable git diff (the CI drift guard keeps it
// honest).
//
// ARCH-PURE: typeNames + renderSkill are pure (sorted-names string → output
// string), unit-tested without IO; main is the thin filesystem shell.
package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

// typeNames lists the datatype noun for every prototype under datatypeDir: the
// filename without `.md`, sorted ascending. The filename — NOT the prototype's
// `type:` frontmatter — is the authoritative type name (the SKILL.md "filename
// without .md is the type name" convention; e.g. product.md carries `type: type,
// name: product`, so reading `type:` would mislabel it). Non-`.md` entries and
// subdirectories are ignored. Pure over the directory listing (the only IO is the
// ReadDir at the edge); deterministic sorted output makes the drift guard meaningful.
func typeNames(datatypeDir string) ([]string, error) {
	ents, err := os.ReadDir(datatypeDir)
	if err != nil {
		return nil, fmt.Errorf("read datatype dir %s: %w", datatypeDir, err)
	}
	var names []string
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		names = append(names, strings.TrimSuffix(name, ".md"))
	}
	sort.Strings(names)
	return names, nil
}

// renderSkill produces the SKILL.md body by replacing the description-tail
// placeholder with the names joined by ", ". Pure and deterministic: the same
// sorted names always yield byte-identical output (the drift guard depends on it).
func renderSkill(names []string) string {
	return strings.Replace(skillTemplate, datatypeNamesPlaceholder, strings.Join(names, ", "), 1)
}

// writeSkill enumerates datatypeDir, renders the SKILL.md, and writes it to
// outputDir/SKILL.md. The thin IO shell over the two pure functions.
func writeSkill(datatypeDir, outputDir string) error {
	names, err := typeNames(datatypeDir)
	if err != nil {
		return err
	}
	out := renderSkill(names)
	dst := filepath.Join(outputDir, "SKILL.md")
	if err := os.WriteFile(dst, []byte(out), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}
	return nil
}
