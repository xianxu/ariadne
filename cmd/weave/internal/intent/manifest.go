package intent

import "strings"

// kindByVerb maps each manifest verb to its Kind. The file-op verbs are
// ported verbatim from setup.sh's walk_manifest `case` (ARCH-DRY — weave must
// dispatch the same verbs the shell does); `prose`/`skill` are the new
// semantic verbs weave adds. The retired `copy` verb is deliberately absent so
// it falls through to the unknown-action skip, mirroring walk_manifest's
// `copy)` warn-and-ignore (see construct/setup.sh:336).
var kindByVerb = map[string]Kind{
	"symlink":  Symlink,
	"seed":     Seed,
	"scaffold": Scaffold,
	"touch":    Touch,
	"merge":    Merge,
	"tool":     Tool,
	"prose":    Prose,
	"skill":    Skill,
}

// ParseManifest parses a base.manifest's text into typed Intents, in file
// order. The line grammar is ported from setup.sh:walk_manifest (ARCH-DRY):
//
//   - A whole-line comment (`^[[:space:]]*#`) or a blank/whitespace-only line
//     is skipped. NOTE: walk_manifest skips only WHOLE-line comments — it does
//     NOT strip trailing comments (that is lib-deps' grammar, in deps.go, not
//     this one), so a '#' mid-line is part of a field here.
//   - The remaining line is whitespace-split into `action source target`
//     (the shell's `read -r action source target`); a Target column omitted
//     leaves Target defaulting to Source (`target="${target:-$source}"`).
//   - A row with no source (verb only) is skipped — there is nothing to act on.
//   - An unrecognized verb is skipped (warn-and-ignore), mirroring
//     walk_manifest's `*)` case and its retired-`copy` handling. A stale row
//     must not abort the compile, so ParseManifest returns no error for it;
//     the error stays in the signature so a future strict mode is non-breaking.
//
// Pure: takes the manifest content, never the file (reading it is an IO-seam
// concern, ARCH-PURE).
func ParseManifest(content string) ([]Intent, error) {
	var intents []Intent
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue // blank or whole-line comment (walk_manifest's two guards)
		}
		fields := strings.Fields(trimmed) // `read -r action source target`
		if len(fields) < 2 {
			continue // verb with no source — nothing to act on
		}
		kind, ok := kindByVerb[fields[0]]
		if !ok {
			continue // unknown action — warn-and-skip (walk_manifest `*)`)
		}
		source := fields[1]
		target := source // `target="${target:-$source}"`
		if len(fields) >= 3 {
			target = fields[2]
		}
		intents = append(intents, Intent{Kind: kind, Source: source, Target: target})
	}
	return intents, nil
}
