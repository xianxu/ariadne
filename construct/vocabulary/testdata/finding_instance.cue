// A VALID #Finding instance, including the ariadne#194 `family` key. Proves the schema
// accepts what the binary emits — the positive half of the closed-schema claim.
//
// Self-contained rather than importing finding.cue: `cue vet` on a single file needs the
// definition present, and duplicating the shape here would defeat the point, so the
// definition is inlined from the source and a drift test is not attempted. What this
// file exists to prove is the CLOSED-ness (see finding_instance_invalid.cue) — that an
// unmodeled key is rejected, which is the claim #194 M3 made and did not enforce.
package findinginstance

#Severity: "Critical" | "Important" | "Minor"

#Finding: {
	id:       string
	severity: #Severity
	title:    string
	detail?:  string
	family?:  string
}

instance: #Finding & {
	id:       "new"
	severity: "Important"
	title:    "the oracle cannot see the thing it certifies"
	detail:   "raw-notation check asks isPronunciation to grade its own output"
	family:   "oracle-blind-direction"
}
