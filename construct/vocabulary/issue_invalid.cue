// Deliberately broken model — proves the M1 vet gate bites. `open` has an empty
// `when`, so the documented-value law (`when[s] & !=""`) evaluates to ⊥ and
// `cue vet` must fail. Separate package so it never unifies with issue.cue.
package issueinvalid

when: {
	open:    "" // ← violation: empty semantics
	working: "actively in progress"
}

laws: {
	"documented-value": {
		for s in ["open", "working"] {
			(s): when[s] & !=""
		}
	}
}
