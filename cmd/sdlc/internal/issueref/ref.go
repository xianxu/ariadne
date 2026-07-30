// Package issueref answers one question: does this `#N` name an issue in THIS repo?
//
// It exists because `#(\d+)\b` — copied into three separate files with no left boundary —
// matched `pair#127` as local issue 127. ariadne#127 exists (an archived issue about
// recalibrating estimates), so 46 minutes of ariadne#187's work were charged to it,
// corrupting the very calibration data that issue had produced (ariadne#190).
//
// The rule needs no lookbehind. The qualifier alternative CONSUMES the repo-name characters
// itself, so one pass recognizes both `#187` and `pair#127` and the caller can tell them
// apart. `\b`'s word class is [0-9A-Za-z_], which is how repo names end — and it deliberately
// excludes `-` and `.`, so a range like `#174-#176` stays two local refs rather than reading
// the second as `174-`-qualified.
//
// # Division of labor with parseRef
//
// `parseRef` (cmd/sdlc/resolve.go) remains the CANONICAL ref grammar and the only VALIDATOR —
// `[repo]#id [Mx]`, `gh#id`, documented in helptext/resolve.md. This package owns the SCAN
// half only, and exists as a separate package solely because `parseRef` lives in package
// `main`, which internal packages cannot import. Every consumer that needs validation still
// filters candidates through `parseRef`; this package's job is finding them and deciding
// locality.
//
// Pure: no IO, no clock, no git. The repo's own name arrives as a parameter — deliberately,
// since deriving it internally would tie the answer to the process's cwd rather than to the
// repo whose text is being scanned.
package issueref

import "regexp"

// QualifiedIDPattern is the un-anchored qualifier+id grammar, exported as a STRING so callers
// needing a different anchoring can compose it. A compiled *regexp cannot be re-anchored, so
// the fragment — not ScanRE — is the shareable unit: cmd/sdlc/migrate.go needs this grammar
// both un-anchored (candidate scanning) and anchored to a whole span with a trailing
// milestone group (its span discriminator). Exporting the fragment is what collapses five
// encodings of this grammar to one instead of two.
//
// The `[0-9]{1,6}` bound is inherited from parseRef, not invented here: workshop ids are
// 6-digit, and RE2's trailing `\b` makes a 7+-digit run match NOTHING rather than a truncated
// prefix. Widening it to `\d+` would fork the grammar in the same breath as consolidating it.
const QualifiedIDPattern = `([A-Za-z0-9][A-Za-z0-9_.-]*)?#([0-9]{1,6})`

// ScanRE is QualifiedIDPattern plus the trailing word boundary — the scanner every consumer
// uses. Group 1 is the qualifier (possibly empty), group 2 the id.
var ScanRE = regexp.MustCompile(QualifiedIDPattern + `\b`)

// Ref is one parsed `#N`. Qualifier is the repo name directly attached before the `#`, or ""
// for a bare ref.
//
// The qualifier is RETAINED rather than discarded once locality is decided: it is the field a
// cross-repo attribution row (`pair#127` as its own line in the active-time table) would need,
// and dropping it here would foreclose that. Nothing reads it beyond IsLocal today; that is a
// deliberate widening point, not dead weight.
type Ref struct {
	Qualifier string
	Num       string
}

// Find returns every `#N` in text, in order, each tagged with its qualifier.
func Find(text string) []Ref {
	var out []Ref
	for _, m := range ScanRE.FindAllStringSubmatch(text, -1) {
		out = append(out, Ref{Qualifier: m[1], Num: m[2]})
	}
	return out
}

// IsLocal reports whether r names an issue in the repo called selfRepo. A bare ref is always
// local; a qualified one is local only when the qualifier IS this repo (`ariadne#180` inside
// ariadne). selfRepo "" means "unknown", so only bare refs count.
//
// The match is EXACT, deliberately unlike resolveRepoDir's exact-then-unique-prefix
// resolution. Prefix matching there is a navigation convenience; here it would be a
// correctness bug — `brain` would claim `brain-family`'s refs, re-introducing exactly the
// cross-repo bleed this package removes.
func (r Ref) IsLocal(selfRepo string) bool {
	if r.Qualifier == "" {
		return true
	}
	return selfRepo != "" && r.Qualifier == selfRepo
}

// LocalNums returns the local issue numbers in text, deduped, preserving first-seen order.
// That ordering contract is inherited from the uniqueRefs it replaces, whose callers pass the
// result straight into commit attribution.
func LocalNums(text, selfRepo string) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range Find(text) {
		if !r.IsLocal(selfRepo) || seen[r.Num] {
			continue
		}
		seen[r.Num] = true
		out = append(out, r.Num)
	}
	return out
}

// CountLocal counts local mentions of TRACKED issues only.
//
// An empty tracked set yields an empty map — the contract activetime.Compute depends on,
// previously expressed as a nil *regexp meaning "match nothing". Keeping it as a set rather
// than a compiled pattern is the point: the tracked issues are data, and building a regex
// from them was the indirection that let three copies of this grammar drift apart.
func CountLocal(text, selfRepo string, tracked map[string]bool) map[string]int {
	counts := map[string]int{}
	if len(tracked) == 0 || text == "" {
		return counts
	}
	for _, r := range Find(text) {
		if r.IsLocal(selfRepo) && tracked[r.Num] {
			counts[r.Num]++
		}
	}
	return counts
}
