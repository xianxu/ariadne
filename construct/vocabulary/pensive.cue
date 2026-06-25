// pensive — the vocabulary of a pensive note. The SECOND datatype in the
// instance-conformance loop (#124 M3): it exists to prove the validator
// generalizes past `issue` — the ONLY per-datatype addition is this `.cue`;
// `vocabulary validate-instance --type pensive <file>` reuses the same engine.
// Mirrors construct/datatype/pensive.md's frontmatter shape. A pensive has no
// lifecycle (it's a timestamped note, not a tracked work item) — frontmatter only.
package pensive

// #Mode: the kind of thinking captured (the documented enum).
#Mode: or(["ideas", "eureka", "thoughts"])

// #Pensive: the frontmatter shape of a pensive note. CLOSED (no `...`), unlike the
// OPEN #Issue — deliberate: #Issue is open because it sits behind a fail-closed merge
// gate where a false positive on an organically-grown field would train --no-validate;
// pensive is validated only on-demand (not gated), so closed is safe and makes the
// rejection sharper. If pensive frontmatter starts growing fields, revisit (open it).
#Pensive: {
	type:        "pensive" // the literal discriminator
	date:        string     // ISO YYYY-MM-DD (cue's YAML loader keeps it a string)
	topic:       string
	mode:        #Mode
	description: string
	references?: [...string]
}
