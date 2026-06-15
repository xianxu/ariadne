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

// visByToken maps the OPTIONAL leading visibility token to its Visibility. A row
// may begin with `export` or `internal` before the type word; absent, visibility
// defaults to Export (see ParseManifest). The token set is disjoint from
// kindByVerb (no type is named `export`/`internal`), so a leading visibility word
// is unambiguous.
var visByToken = map[string]Visibility{
	"export":   Export,
	"internal": Internal,
}

// ParseManifest parses a base.manifest's text into typed Intents, in file
// order. The line grammar is ported from setup.sh:walk_manifest (ARCH-DRY),
// extended with the OPTIONAL leading visibility token (ariadne#99):
//
//   - A whole-line comment (`^[[:space:]]*#`) or a blank/whitespace-only line
//     is skipped. NOTE: walk_manifest skips only WHOLE-line comments — it does
//     NOT strip trailing comments (that is lib-deps' grammar, in deps.go, not
//     this one), so a '#' mid-line is part of a field here.
//   - The remaining line is whitespace-split. The FIRST field may be a
//     visibility token (`export`|`internal`); when present it is consumed and
//     sets the Intent's Visibility, leaving the rest as `action source target`.
//     Absent ⇒ Visibility defaults to Export (the algebra's default — every
//     pre-visibility row is unchanged). The token set is disjoint from the verb
//     set, so a leading `export`/`internal` is unambiguous.
//   - The remaining fields parse as `action source target` (the shell's
//     `read -r action source target`); a Target column omitted leaves Target
//     defaulting to Source (`target="${target:-$source}"`).
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
		fields := strings.Fields(trimmed) // [vis] action source target

		// Optional leading visibility token: consume it when present, else
		// default Export. Unambiguous — `export`/`internal` are not verbs.
		visibility := Export
		if v, isVis := visByToken[fields[0]]; isVis {
			visibility = v
			fields = fields[1:]
		}

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
			// Take only fields[2] as the target, silently ignoring a 4th+
			// column. This deliberately diverges from `read -r action source
			// target`, which folds all trailing words into `target` — benign
			// because no manifest row has >3 columns (a target is a single
			// path), so the two semantics coincide on every real input.
			target = fields[2]
		}
		intents = append(intents, Intent{Kind: kind, Visibility: visibility, Source: source, Target: target})
	}
	return intents, nil
}
