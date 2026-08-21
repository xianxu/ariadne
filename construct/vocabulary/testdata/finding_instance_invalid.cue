// Proves #Finding is CLOSED — ariadne#194 M3 asserted this in a code comment, a commit
// message and the issue Log, and nothing enforced it. An unmodeled key must make
// `cue vet` fail; if this file ever vets, the closed-schema rationale is false and the
// `family` key could have been added in Go alone without the model ever noticing.
package findinginstanceinvalid

#Severity: "Critical" | "Important" | "Minor"

#Finding: {
	id:       string
	severity: #Severity
	title:    string
	detail?:  string
	family?:  string
}

instance: #Finding & {
	id:        "new"
	severity:  "Important"
	title:     "a finding with a key the model never declared"
	taxonomy:  "not-a-modeled-field" // ← violation: #Finding is closed
}
